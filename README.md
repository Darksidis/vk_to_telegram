# vk_to_telegram

### About
-------

A program that transfers content from a VK public to a Telegram channel

### What did I use
-------

Golang, VK API, Telegram API

### How it works
--------

The program monitors updates in the VK public, and when new posts appear, it pulls out text and multimedia (except for music) using the VK-api method, transfers it to the designated telegram channel. Optional: adds links to the telegram channel in VK posts

### How it starts
--------

Small preparations

You will need
- the token of your group in VK (we get it here https://dev.vk.com/api/access-token/getting-started)
- your bot's token in telegram (get here https://t.me/BotFather)
- ID of your channel in telegram (send the post from the channel here https://t.me/getmyid_bot)
- URL telegram-group. Visible in the channel

- standalone token is required for editing posts. Optionally, if not necessary, then do nothing with *STANDALONE_TOKEN* in the env file. How to get described here (https://vk.com/@steadyschool-kak-sozdat-standalone-prilozhenie-i-poluchit-token)
The link that will need to be entered in the browser will have the following format
```
https://oauth.vk.com/authorize?client_id=** ID YOUR APP **&display=page&redirect_uri=https://oauth.vk.com/blank.html&scope=wall,offline&response_type=token&v=5.130
```


*You will also need to add the bot to your telegram channel, and grant admin rights*

Create an .env file with the following information.
- VK_TOKEN
- BOT_TOKEN
- ID_CHANNEL
- URL_TG_GROUP
- STANDALONE_TOKEN

You can run through a binary file
```
./main
```
