# telegram-fhome-bot

```console
go build
```

## systemd setup

### install

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

```console
systemctl --user enable telegram-fhome-bot.service
```

### view logs

```console
journalctl --user --unit telegram-fhome-bot.service --follow
```

[link1]: https://serverfault.com/a/997608/590260
[link2]: https://unix.stackexchange.com/q/521538/417321
