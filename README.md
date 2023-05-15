# albion-match-sell-order

CLI tool that captures the lowest price (minus 1) and copies it to your clipboard so you can easily paste it.

![](https://i.imgur.com/4lBqxTz.png)

# Pre-requisites

- Download [Tesseract](https://tesseract-ocr.github.io/tessdoc/Downloads)
- Use `1920x1080` Albion Online resolution
- Get one free API key from https://ocr.space/ocrapi/freekey
- Put API key in `config.json`.

# config.json

- **apikey**: "your api key goes here",
- **startX**: starting X coord (blue),
- **startY**: starting Y coord (blue),
- **sizeX**: horizontal size in pixels (orange),
- **sizeY**: vertical size in pixels (yellow),
- **deleteImage**: deletes image after the process

![](https://i.imgur.com/cR3C55W.png)

# How to

1. Open Albion Online
2. Go to market
3. Go to your sell orders tab  
	![](https://i.imgur.com/QnDnXEO.png)
4. Edit one of your orders  
	![](https://i.imgur.com/alXZ4eH.png)
5. Run `albion-match-sell-order.exe` from your 2nd screen
6. The cheapest price minus 1 will be inserted into your clipboard
7. Paste the price and update your item
8. Rinse and repeat
