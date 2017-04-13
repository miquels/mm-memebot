# Mattermost memebot

> /meme command for mattermost
> based on:
> - [memegen.link](https://memegen.link)
> - [A Meme Bot for Slack](https://github.com/nicolewhite/slack-meme)

## Installation
- go build
- copy _mm-memebot_ to /usr/local/bin
- copy _mm-memebot.default_ to /etc/default/memebot
- edit /etc/default/mm-memebot
- copy _mm-memebot.service_ to /lib/systemd/system
- run: systemctl start mm-memebot

