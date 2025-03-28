# Bulk Telegram Sticker Pack Creator

This tool helps you create and upload sticker packs to Telegram using a two-step process:
1. Prepare and process your images
2. Create and upload the sticker pack

## Setup

1. Install the required dependencies:
```bash
pip install -r requirements.txt
```

2. Get your Telegram API credentials:
   - Go to https://my.telegram.org/auth
   - Log in with your phone number
   - Go to 'API development tools'
   - Create a new application
   - Copy the `API_ID` and `API_HASH`

3. Create a `.env` file in the project root with your credentials:
```
API_ID=your_api_id
API_HASH=your_api_hash
PHONE=your_phone_number
```

## Usage

### Step 1: Prepare Images

1. Create an `input` directory in the project root
2. Place your sticker images in the `input` directory (supported formats: PNG, JPG, JPEG)
3. Run the preparation script:
```bash
python prepare.py
```

This will:
- Convert all images to WebP format (512x512 pixels)
- Create an `output` directory with the processed images
- Generate a `sticker_config.json` file in the `output` directory

4. Edit the `output/sticker_config.json` file to:
   - Change the sticker pack title and short name
   - Modify emojis for each sticker
   - Add or remove stickers from the pack

Example configuration:
```json
{
    "sticker_pack": {
        "title": "My Awesome Stickers",
        "short_name": "my_awesome_stickers",
        "stickers": [
            {
                "original": "sticker1.png",
                "webp": "sticker1.webp",
                "emoji": "😀"
            },
            {
                "original": "sticker2.png",
                "webp": "sticker2.webp",
                "emoji": "😎"
            }
        ]
    }
}
```

### Step 2: Create Sticker Pack

1. Run the main script to create and upload the sticker pack:
```bash
python main.py
```

2. The first time you run the script, you'll need to authenticate with your phone number
3. The script will create the sticker pack and upload all stickers according to the configuration

## Important Notes

- The `short_name` must be unique across all Telegram sticker packs
- Input images will be automatically resized to 512x512 pixels while maintaining aspect ratio
- Images are converted to WebP format with transparency support
- Make sure your input images are clear and of good quality
- The first time you run the script, you'll need to authenticate with your phone number

## Troubleshooting

Due to the nature of interacting with Telegram's Sticker bot, some issues may occur:

1. **Manual Intervention Required**: Sometimes the script may fail to complete the process automatically. In such cases:
   - The script will indicate that manual completion is needed
   - You can continue the process manually by messaging @Stickers
   - Follow the bot's prompts to complete the sticker pack creation

2. **Common Issues**:
   - Bot response delays: The script includes delays between actions, but sometimes the bot may need more time
   - Network issues: Ensure you have a stable internet connection
   - Rate limiting: If you create multiple packs, Telegram may temporarily limit your actions

3. **If the Script Fails**:
   - Check the error message for specific details
   - You can start a new conversation with @Stickers and complete the process manually
   - The bot will guide you through the remaining steps

Remember that while the script automates most of the process, you may need to complete some steps manually if unexpected issues arise.
