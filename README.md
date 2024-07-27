# telegram-fhome-bot

```console
go build
```

## systemd setup

### install

install as user service, so [`%h` specifier can be used][so_link].

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

```
journalctl --user --unit telegram-fhome-bot.service --follow
```

[so_link]: https://serverfault.com/a/997608/590260
