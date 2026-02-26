# Overview

This bot can listen to events created on a discord server, and post a notification to everyone as well as reminders for the event in a given channel.

# Configuration

The bot is configured using environment variables. The following variables exist:


| Variable                     | Required | Default  | Note |
|------------------------------|----------|----------|----------|
|DISCORD_BOT_TOKEN             | Yes      |          | The discord bot token |
|DISCORD_CHANNEL_ID            | Yes      |          | The discord channel ID for posting the event notifications |



# Installation

## Binary

Download a binary from the [Releases](https://github.com/patrickjane/event-bot/releases) for your corresponding architecture, and run the binary. Make sure to set the environment variables as per the above definitions.

Example:

```
$ ./eventbot.linux-amd64
EventBot 1.0.0
https://github.com/patrickjane/event-bot
Sync complete. 0 reminders in queue.
Bot is running. Monitoring events...
```

## Docker

The bot comes with 2 docker images, linux/amd64 and linux/arm64. Image path is `ghcr.io/patrickjane/event-bot:latest` (or substitute `latest` with an actual version.).

You can run it directly, or e.g. use a `docker compose` file:

`docker-compose.yml`

```
services:
  eventbot:
    container_name: event-bot
    image: ghcr.io/patrickjane/event-bot:latest
    restart: unless-stopped
    environment:
      - DISCORD_BOT_TOKEN=XXX
      - DISCORD_CHANNEL_ID=XXX
```

And then:

```
$ docker compose up
[+] up 1/1
 ✔ Container event-bot Recreated                                                                                                                                                                    0.1s
Attaching to event-bot

event-bot  | EventBot 1.0.0
event-bot  | https://github.com/patrickjane/event-bot
event-bot  | Sync complete. 0 reminders in queue.
event-bot  | Bot is running. Monitoring events...
```

#### Note
Use `docker compose up -d` to run the bot in background. Afterwards use `docker ps` to find the container, and `docker logs -f [ID]` to show the logs of the container.


# License
MIT License, see [LICENSE](LICENSE)