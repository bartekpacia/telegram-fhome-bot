# telegram-fhome-bot

```console
go build
```

## Install

**macOS**

TBD

**ArchLinux**

Clone [my unofficial Arch User Repository (AUR)](https://github.com/bartekpacia/aur):

```console
git clone https://github.com/bartekpacia/aur
```

```
cd telegram-fhome-bot
```

```console
makepkg -si
```

## systemd setup

### as a systemd user-service

You can also install as a systemd user service.
The unit file would looks like this:

```
# /home/charlie/.config/systemd/user/telegram-fhome-bot.service
[Unit]
Description=Telegram bot that provides access to F&Home smart home system
Wants=network-online.target
After=network-online.target
#StartLimitIntervalSec=60
#StartLimitBurst=3

[Service]
Type=simple
ExecStart=%h/telegram-fhome-bot/telegram-fhome-bot
EnvironmentFile=%h/telegram-fhome-bot/.env
Restart=on-failure
RestartSec=10s

[Install]
WantedBy=multi-user.target
```

- install as user service, so [`%h` specifier can be used][link1]
- enable lingering (`loginctl enable-linger`) to avoid [death on logout][link2]

```console
cp telegram-fhome-bot.service ~/.config/systemd/user
```

```console
systemctl --user daemon-reload
```

Now you can see that systemd sees it:

```console
systemctl --user list-unit-files --type service
```

Don't forget to start/enable the service:

```console
systemctl --user enable telegram-fhome-bot.service
```

### view logs

```console
journalctl --user --unit telegram-fhome-bot.service --follow
```

## Metrics

The bot can expose Prometheus-compatible metrics on `/metrics`.

Set `METRICS_LISTEN_ADDR` (for example `127.0.0.1:9108`) in your environment file.
If unset, it defaults to `127.0.0.1:9108`.

Available toggle metrics:

- `telegram_fhome_bot_gate_toggle_events_total{result="success|error"}`
- `telegram_fhome_bot_gate_toggle_duration_seconds{result="success|error"}` (`_bucket`, `_sum`, `_count`)
- `telegram_fhome_bot_gate_toggle_last_unix_time_seconds{result="success|error"}`

[link1]: https://serverfault.com/a/997608/590260
[link2]: https://unix.stackexchange.com/q/521538/417321
