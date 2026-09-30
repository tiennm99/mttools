import os
import json
from PIL import Image
import glob
from pathlib import Path

def convert_to_webp(input_path, output_path, size=(512, 512)):
    """Convert image to WebP format with specified size"""
    try:
        with Image.open(input_path) as img:
            # Convert to RGBA if necessary
            if img.mode != 'RGBA':
                img = img.convert('RGBA')
            
            # Resize image maintaining aspect ratio
            img.thumbnail(size, Image.Resampling.LANCZOS)
            
            # Create a new image with padding if necessary
            new_img = Image.new('RGBA', size, (0, 0, 0, 0))
            offset = ((size[0] - img.size[0]) // 2, (size[1] - img.size[1]) // 2)
            new_img.paste(img, offset)
            
            # Save as WebP
            new_img.save(output_path, 'WEBP', quality=90)
            return True
    except Exception as e:
        print(f"Error converting {input_path}: {str(e)}")
        return False

def process_images(input_dir, output_dir):
    """Process all images in input directory and convert to WebP"""
    # Create output directory if it doesn't exist
    Path(output_dir).mkdir(parents=True, exist_ok=True)
    
    # Supported input formats
    input_formats = ['*.png', '*.jpg', '*.jpeg']
    processed_files = []
    
    # Process each image
    for format in input_formats:
        for input_path in glob.glob(os.path.join(input_dir, format)):
            filename = os.path.basename(input_path)
            name_without_ext = os.path.splitext(filename)[0]
            output_path = os.path.join(output_dir, f"{name_without_ext}.webp")
            
            if convert_to_webp(input_path, output_path):
                processed_files.append({
                    "original": filename,
                    "webp": f"{name_without_ext}.webp",
                    "emoji": "😀"  # Default emoji
                })
    
    return processed_files

def generate_config(processed_files, output_dir):
    """Generate configuration file"""
    config = {
        "title": "My Awesome Stickers",
        "short_name": "my_awesome_stickers",
        "stickers": processed_files
    }
    
    # Save config file
    config_path = os.path.join(output_dir, "sticker_config.json")
    with open(config_path, 'w', encoding='utf-8') as f:
        json.dump(config, f, indent=4, ensure_ascii=False)
    
    return config_path

def main():
    # Setup directories
    input_dir = "input"
    output_dir = "output"
    
    # Create input directory if it doesn't exist
    Path(input_dir).mkdir(parents=True, exist_ok=True)
    
    print("Processing images...")
    processed_files = process_images(input_dir, output_dir)
    
    if not processed_files:
        print("No images found in input directory!")
        return
    
    print(f"Processed {len(processed_files)} images")
    
    # Generate configuration file
    config_path = generate_config(processed_files, output_dir)
    print(f"\nConfiguration file generated: {config_path}")
    print("\nYou can now edit the configuration file to:")
    print("1. Change the sticker pack title and short name")
    print("2. Modify emojis for each sticker")
    print("3. Add or remove stickers from the pack")
    print("\nAfter editing, run main.py to create the sticker pack")

if __name__ == "__main__":
    main() 