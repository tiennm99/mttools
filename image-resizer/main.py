import os
from pathlib import Path

from PIL import Image


def resize_image(input_path, output_path, scale_percent):
    """
    Resize an image to a specified percentage of its original size.

    Args:
        input_path: Path to the input image
        output_path: Path where the resized image will be saved
        scale_percent: Percentage to scale the image (e.g., 50 for 50%)
    """
    try:
        with Image.open(input_path) as img:
            # Calculate new dimensions
            width, height = img.size
            new_width = int(width * scale_percent / 100)
            new_height = int(height * scale_percent / 100)

            # Resize using high-quality resampling
            resized_img = img.resize((new_width, new_height), Image.Resampling.LANCZOS)

            # Save the resized image
            resized_img.save(output_path)
            print(f"Resized: {input_path} -> {output_path}")
            print(f"  Original: {width}x{height}, New: {new_width}x{new_height}")
    except Exception as e:
        print(f"Error processing {input_path}: {e}")


def resize_images_in_folder(
    folder_path, scale_percent, output_folder=None, overwrite=False
):
    """
    Resize all images in a folder and its subfolders.

    Args:
        folder_path: Path to the folder containing images
        scale_percent: Percentage to scale the images (e.g., 50 for 50%)
        output_folder: Optional output folder. If None, creates a 'resized' subfolder
        overwrite: If True, overwrites original images. If False, saves to output folder
    """
    # Supported image formats
    image_extensions = {".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp"}

    # Setup output path
    folder_path = Path(folder_path)
    if overwrite:
        output_base_path = folder_path
    else:
        if output_folder is None:
            output_folder = folder_path / "resized"
        else:
            output_folder = Path(output_folder)

    # Process all files in folder and subfolders
    for root, dirs, files in os.walk(folder_path):
        root_path = Path(root)

        # Create relative path structure for output
        if not overwrite:
            relative_path = root_path.relative_to(folder_path)
            current_output_path = output_folder / relative_path
            current_output_path.mkdir(parents=True, exist_ok=True)
        else:
            current_output_path = root_path

        # Process each image file
        for file in files:
            file_path = root_path / file
            file_ext = file_path.suffix.lower()

            if file_ext in image_extensions:
                if overwrite:
                    output_path = file_path
                else:
                    output_path = current_output_path / file

                resize_image(file_path, output_path, scale_percent)


if __name__ == "__main__":
    # Configuration
    FOLDER_PATH = "D:\\kvtm\\kvtm\\client\\res\\common\\ui\\EventTet2026"  # Current directory - change to your folder path
    SCALE_PERCENT = 50  # Resize to 50% of original size

    print(f"Image Resizer - Scaling to {SCALE_PERCENT}%")
    print(f"Source folder: {FOLDER_PATH}")
    print("Mode: Overwriting original images")
    print("-" * 50)

    resize_images_in_folder(FOLDER_PATH, SCALE_PERCENT, overwrite=True)

    print("-" * 50)
    print("Done!")
