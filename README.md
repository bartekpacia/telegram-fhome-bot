# telegram-fhome-bot

```console
go build
```

## systemd setup

### install

```console
cp telegram-fhome-bot.service /etc/systemd/system
```

```console
sudo systemd daemon-reload
```

Now you can see that systemd sees it:

```console
sudo systemctl list-unit-files --type service
```

```console
sudo systemctl enable telegram-fhome-bot.service
```

### view logs

```
journalctl --unit telegram-fhome-bot.service
```
