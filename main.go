package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"

	"github.com/bartekpacia/fhome/api"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var (
	botUsername    string
	allowedChatIDs []int64
	allowedUserIDs []int64
)

var fhomeClient *api.Client

func main() {
	logOpts := slog.HandlerOptions{Level: slog.LevelDebug}
	logHandler := slog.NewTextHandler(os.Stdout, &logOpts)
	slog.SetDefault(slog.New(logHandler))
	log.SetFlags(log.Flags() &^ (log.Ldate))

	var err error
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		slog.Error("TELEGRAM_BOT_TOKEN is empty")
		os.Exit(1)
	}
	botUsername = os.Getenv("TELEGRAM_BOT_USERNAME")
	if botUsername == "" {
		slog.Error("TELEGRAM_BOT_USERNAME is empty")
		os.Exit(1)
	}
	allowedChatIDsStr := os.Getenv("TELEGRAM_ALLOWED_CHAT_IDS")
	for _, chatIDStr := range strings.Split(allowedChatIDsStr, ",") {
		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			slog.Error("Invalid chat ID", slog.Any("error", err))
			os.Exit(1)
		}
		allowedChatIDs = append(allowedChatIDs, chatID)
	}
	allowedUserIDsStr := os.Getenv("TELEGRAM_ALLOWED_USER_IDS")
	for _, userIDStr := range strings.Split(allowedUserIDsStr, ",") {
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			slog.Error("Invalid user ID", slog.Any("error", err))
			os.Exit(1)
		}
		allowedUserIDs = append(allowedUserIDs, userID)
	}

	fhomeClient, err = createFhomeClient()
	if err != nil {
		slog.Error("error creating fhome client", slog.Any("error", err))
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		panic(err)
	}

	b.Start(ctx)
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	l := slog.With(slog.Int64("update_id", update.ID))

	l.Info("start processing update")
	msg := update.Message
	if msg == nil {
		l.Info("update is not about a message, ignoring")
		return
	}

	user := msg.From
	l = l.With(
		slog.Int64("chat_id", msg.Chat.ID),
		slog.Group(
			"from_user",
			slog.String("username", msg.From.Username),
			slog.String("name", msg.From.FirstName+" "+msg.From.LastName),
			slog.Int64("id", msg.From.ID),
		))
	defer l.Info("end processing update")
	l.Info("update is a message", slog.String("text", msg.Text))

	// Check if the bot was added to the group. If yes, print group ID
	if len(msg.NewChatMembers) > 0 {
		if botUsername != msg.NewChatMembers[0].Username {
			l.Error("user was added to group, but it's not me", slog.String("bot_username", botUsername), slog.String("new_username", msg.NewChatMembers[0].Username))
			// Some other user was added to our group. We don't care.
			return
		}

		l.Info(
			"added to group",
			slog.Group(
				"group",
				slog.String("title", msg.Chat.Title),
				slog.Int64("id", msg.Chat.ID),
			),
		)

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   "Hej, jestem botem! Do tej grupy dodał mnie " + user.Username + "!",
		})
	}

	validGroupChat := msg.Chat.Type == "group" && slices.Contains(allowedChatIDs, msg.Chat.ID)
	validPrivateChat := msg.Chat.Type == "private" && slices.Contains(allowedUserIDs, user.ID)
	if !validPrivateChat && !validGroupChat {
		l.Error("ignoring message from chat", slog.Group(
			"chat",
			slog.Int64("id", msg.Chat.ID),
			slog.String("title", msg.Chat.Title),
		))
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   "Currently I don't work in this chat, sorry",
		})
		return
	}

	if strings.Contains(strings.ToLower(msg.Text), "brama") {
		const gateId = 260
		err := fhomeClient.SendEvent(gateId, api.ValueToggle)
		if err != nil {
			l.Error("error sending event", slog.Any("error", err))
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: msg.Chat.ID,
				Text:   "Nie udało się otworzyć/zamknąć bramy\n" + err.Error(),
			})
		} else {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: msg.Chat.ID,
				Text:   getConfirmationMessage(user.ID),
			})
		}
	} else {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   "Sory ale nie rozumiem. Na razie umiem tylko otwierać/zamykać bramę",
		})
	}
}

func getConfirmationMessage(userID int64) string {
	var name string
	var noun string
	if userID == 1028925187 { // tata
		name = "Tomaszu"
		noun = "Panie"
	} else if userID == 1174832124 { // mama
		name = "Elżbieto"
		noun = "Pani"
	} else if userID == 754149197 { // bartek
		name = "Bartłomieju"
		noun = "Panie"
	} else if userID == 910335851 { // ola
		name = "Aleksandro"
		noun = "Pani"
	}

	var confirmationMessages = []string{
		fmt.Sprintf("%s, Twoje życzenie jest dla mnie rozkazem! Otwieram/zamykam bramę :)", name),
		fmt.Sprintf("%s, niech mi się stanie według Słowa twego. Brama zostanie otwarta!", name),
		"Niechaj brama zostanie otwarta jak dusza Kordiana na szczycie Mont Blanc. Wstępuj!",
		"O bramo, bramo, uchyl (lub zamknij) swe wrota!",
		"Człowieku! Władza nad bramą w Twoich rękach! Otwieram/zamykam!",
		"Miej serce i patrzaj w bramę. Już otwarta/zamknięta!",
		fmt.Sprintf("Z rozkazu Twego, o %s, brama się poddaje!” Otwieram/zamykam!", noun),
		"Lepiej późno niż wcale! Ale spokojnie, brama już otwarta/zamknięta!",
		fmt.Sprintf("Jako iż wola Twoja, %s %s, brzmi potężnie, tak też brama otwarta/zamknięta zostanie!", noun, name),
		"Niech żywi nie tracą nadziei! Bo brama już otwarta/zamknięta!",
		"Tak jest! Wcale nie musisz być czarodziejem, aby otwierać bramy!",
		fmt.Sprintf("Robi się, %s %s", noun, name),
		"Zaklęcia zostały rzucone, a status bramy zostanie zmieniony!",
		fmt.Sprintf("%s, wiedz, że brama zawsze się Ciebie usłucha! Już otwieram/zamykam!", name),
	}

	msg := confirmationMessages[rand.IntN(len(confirmationMessages))]
	return msg
}
