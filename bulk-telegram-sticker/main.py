import os
import asyncio
import json
from telethon import TelegramClient, events
from telethon.tl.types import InputPeerUser
from telethon.tl.functions.messages import SendMessageRequest
from pathlib import Path
from dotenv import load_dotenv

# Load environment variables from .env file
load_dotenv()

# Telegram API credentials
API_ID = os.getenv('API_ID')
API_HASH = os.getenv('API_HASH')
PHONE = os.getenv('PHONE')

# Sticker bot username
STICKER_BOT_USERNAME = "@Stickers"

print(f"API_ID: {API_ID}")
print(f"API_HASH: {API_HASH}")
print(f"PHONE: {PHONE}")

# Initialize the client
client = TelegramClient(PHONE, API_ID, API_HASH)

def load_config(config_path):
    """Load and validate configuration file"""
    try:
        with open(config_path, 'r', encoding='utf-8') as f:
            config = json.load(f)
        
        # Validate required fields
        if not all(key in config for key in ['title', 'short_name', 'stickers']):
            raise ValueError("Missing required fields in configuration")
        
        return config
    except Exception as e:
        print(f"Error loading configuration: {str(e)}")
        return None

async def create_sticker_pack(config):
    """
    Create a new sticker pack by interacting with the Sticker bot
    
    Args:
        config (dict): Configuration dictionary containing sticker pack details
    """
    try:
        # Start the client
        await client.start(phone=PHONE)
        
        # Get the sticker bot entity
        sticker_bot = await client.get_entity(STICKER_BOT_USERNAME)
        
        # Start new sticker pack creation
        await client(SendMessageRequest(
            peer=sticker_bot,
            message="/newpack"
        ))
        
        # Wait for bot to ask for pack name
        await asyncio.sleep(1)
        
        # Send pack name
        await client(SendMessageRequest(
            peer=sticker_bot,
            message=config['title']
        ))
        
        # Wait for bot to ask for short name
        await asyncio.sleep(1)
        
        # Send short name
        await client(SendMessageRequest(
            peer=sticker_bot,
            message=config['short_name']
        ))
        
        # Wait for bot to ask for username
        await asyncio.sleep(1)
        
        # Send username (using short name)
        await client(SendMessageRequest(
            peer=sticker_bot,
            message=config['short_name']
        ))
        
        output_dir = "output"
        
        # Upload stickers one by one
        for sticker in config['stickers']:
            sticker_path = os.path.join(output_dir, sticker['webp'])
            if not os.path.exists(sticker_path):
                print(f"Warning: Sticker file not found: {sticker_path}")
                continue
            
            # Send sticker file
            await client.send_file(
                sticker_bot,
                sticker_path,
                caption=sticker['emoji']
            )
            
            # Wait for bot to process the sticker
            await asyncio.sleep(2)
            
            # Skip any additional prompts
            await asyncio.sleep(1)
            await client(SendMessageRequest(
                peer=sticker_bot,
                message="/skip"
            ))
            await asyncio.sleep(1)
        
        # Send /publish command to finalize the pack
        await client(SendMessageRequest(
            peer=sticker_bot,
            message="/publish"
        ))
        
        # Wait for bot to ask for confirmation
        await asyncio.sleep(1)
        
        # Confirm publishing
        await client(SendMessageRequest(
            peer=sticker_bot,
            message="Yes"
        ))
        
        print(f"Sticker pack created successfully: {config['short_name']}")
        return config['short_name']
        
    except Exception as e:
        print(f"Error creating sticker pack: {str(e)}")
        print("Note: If the script fails, you may need to complete the process manually with the Sticker bot.")
        return None
    finally:
        await client.disconnect()

async def main():
    # Load configuration
    config_path = os.path.join("output", "sticker_config.json")
    if not os.path.exists(config_path):
        print("Configuration file not found! Please run prepare.py first.")
        return
    
    config = load_config(config_path)
    if not config:
        return
    
    await create_sticker_pack(config)

if __name__ == "__main__":
    asyncio.run(main())
