package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"

	"github.com/bartekpacia/fhome/api"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var (
	botUsername   string
	allowedChatID int64
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
	allowedChatID, err = strconv.ParseInt(os.Getenv("TELEGRAM_ALLOWED_CHAT_ID"), 10, 64)
	if err != nil {
		slog.Error("TELEGRAM_ALLOWED_CHAT_ID is invalid", slog.Any("error", err))
		os.Exit(1)
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
	l := slog.With(slog.Int64("id", update.ID))

	l.Info("start processing update")
	msg := update.Message
	if msg == nil {
		l.Info("update has no message, ignoring")
		return
	}

	user := msg.From

	l.Info("update has message", slog.String("text", msg.Text), slog.String("from_user", msg.From.Username))

	// Check if the bot was added to the group. If yes, print group ID
	if msg.NewChatMembers != nil && len(msg.NewChatMembers) > 0 {
		if botUsername != msg.NewChatMembers[0].Username {
			slog.Error("user was added to group, but it's not me", slog.String("bot_username", botUsername), slog.String("new_username", msg.NewChatMembers[0].Username))
			// Some other user was added to our group. We don't care.
			return
		}

		l.Info(
			"added to group",
			slog.Group(
				"added_by",
				slog.String("username", user.Username),
				slog.String("name", user.FirstName+" "+user.LastName),
				slog.Int64("id", user.ID),
			),
			slog.Group(
				"added_to_group",
				slog.String("title", msg.Chat.Title),
				slog.Int64("id", msg.Chat.ID),
			),
		)

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   "Siema, jestem botem i dodał mnie do tej grupy" + user.Username + "!",
		})
	}

	// Only make this bot work in my family group. Possibly more in the future.
	if msg.Chat.ID != allowedChatID {
		l.Error("ignoring message from chat", slog.Int64("chat_id", msg.Chat.ID), slog.String("chat_title", msg.Chat.Title))
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   "Currently I don't work in this group, sorry",
		})
		return
	}

	l.Info("handle", slog.String("user", user.Username), slog.String("text", msg.Text))

	if msg.Text == "/enabled@FAndHomeBot" {
		resp, err := fhomeClient.GetSystemConfig()
		if err != nil {
			l.Error("error getting system config", slog.Any("error", err))
			return
		}

		cells := []struct {
			Name  string
			Value int
		}{}
		mdCells := resp.Response.MobileDisplayProperties.Cells
		for _, cell := range mdCells {
			if cell.DisplayType != api.Percentage && cell.MinValue == "0x6000" {
				continue
			}

			val, err := api.RemapLighting(cell.Step)
			if err != nil {
				l.Error(
					"error remapping lighting value",
					slog.Group("cell", slog.String("id", cell.ID), slog.String("desc", cell.Desc)),
					slog.Any("error", err),
				)
				continue
			}

			if val == 0 {
				continue
			}

			cells = append(cells, struct {
				Name  string
				Value int
			}{Name: cell.Desc, Value: val})
		}

		text := "Oto status oświetlenia:\n"
		for _, cell := range cells {
			text += fmt.Sprintf("• %s: %d%%\n", cell.Name, cell.Value)
		}

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   text,
		})
	}

	l.Info("end processing update")
}
