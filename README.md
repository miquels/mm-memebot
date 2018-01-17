# Mattermost memebot

> /meme command for mattermost
> based on:
> - [memegen.link](https://memegen.link)
> - [A Meme Bot for Slack](https://github.com/nicolewhite/slack-meme)

## Help output

`/meme help`

Mattermost Meme Bot  
> Commands:  
  
`/meme memename top_row;bottom_row` generate a meme image  
(NOTE: memename can also be a URL to an image)  
`/meme list` List templates  
`/meme help` Shows this menu  
  
## Show meme image
`/meme buzz memes,memes; everywhere`  

<img src="https://user-images.githubusercontent.com/6455542/35054872-00809bbe-fbae-11e7-8569-fa3e46ddd2bf.jpg">


## Installation

To build the bot, simply:
```
go build
```

To install and run:
```
cp mm-memebot /usr/local/bin
cp mm-memebot.default /etc/default/memebot
# edit that file to check and maybe change settings.
vi /etc/default/mm-memebot
cp mm-memebot.service /lib/systemd/system
systemctl enable mm-memebot
systemctl start mm-memebot
```
The bot is now running and listening on localhost:5020.

## Add to mattermost
Go to the integration - slash commands menu
```
Title:                    Meme Generator
Description:              Displays a meme image with your text overlaid
Command Trigger Word:     meme
Request URL:              http://localhost:5020/meme
Request Method:           POST
Autocomplete:             [v]
Autocomplete hint:        memename top text; bottom text
Autocomplete description: Generates a meme image (try /meme help)
```

## Allow mattermost to connect to services on localhost
Go to System Console -> Advanced -> Developer. Set:
```
Allow untrusted internal connections to:   localhost
```
