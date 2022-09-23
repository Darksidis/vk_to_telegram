# vk_to_telegram

### About
-------

A program that transfers content from a VK public to a Telegram channel

### What did I use
-------

Golang, VK API, Telegram API

### How it works
--------

The program monitors updates in the VK public, and when new posts appear, it pulls out text and multimedia (except for music) using the VK-api method, transfers it to the designated telegram channel.

### How it starts
--------

Small preparations

You will need
- the token of your group in VK (we get it here https://dev.vk.com/api/access-token/getting-started)
- your bot's token in telegram (get here https://t.me/BotFather)
- ID of your channel in telegram (send the post from the channel here https://t.me/getmyid_bot)

Create an .env file with the following information.
- VK_TOKEN
- BOT_TOKEN
- ID_CHANNEL

You can run through a binary file
```
./vk_to_telegram_parser.exe
```

