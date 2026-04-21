"""
Image diffing module for generating visual comparison of screenshots.
"""

from pathlib import Path
from PIL import Image, ImageChops, ImageDraw
import numpy as np


def create_diff_image(
    image1_path: str,
    image2_path: str,
    output_path: str,
    threshold: int = 30
) -> dict:
    """
    Create a visual diff image highlighting pixel differences.
    
    Args:
        image1_path: Path to first screenshot
        image2_path: Path to second screenshot
        output_path: Path to save diff image
        threshold: Pixel difference threshold (0-255)
    
    Returns:
        Dictionary with diff statistics
    """
    # Load images
    img1 = Image.open(image1_path).convert('RGB')
    img2 = Image.open(image2_path).convert('RGB')
    
    # Get dimensions
    size1 = img1.size
    size2 = img2.size
    
    # Resize to match if dimensions differ
    if size1 != size2:
        # Use the larger dimensions
        max_width = max(size1[0], size2[0])
        max_height = max(size1[1], size2[1])
        
        # Create new images with white background
        new_img1 = Image.new('RGB', (max_width, max_height), (255, 255, 255))
        new_img2 = Image.new('RGB', (max_width, max_height), (255, 255, 255))
        
        new_img1.paste(img1, (0, 0))
        new_img2.paste(img2, (0, 0))
        
        img1 = new_img1
        img2 = new_img2
    
    # Calculate pixel-by-pixel difference
    diff = ImageChops.difference(img1, img2)
    
    # Convert to numpy for analysis
    diff_array = np.array(diff)
    
    # Calculate difference magnitude
    diff_magnitude = np.sqrt(np.sum(diff_array ** 2, axis=2))
    
    # Create mask for changed pixels
    changed_mask = diff_magnitude > threshold
    
    # Calculate statistics
    total_pixels = diff_magnitude.size
    changed_pixels = np.sum(changed_mask)
    change_percentage = (changed_pixels / total_pixels) * 100
    
    # Create visual diff image
    # Start with a side-by-side comparison
    width, height = img1.size
    combined_width = width * 3  # img1, img2, diff
    combined_height = height
    
    combined = Image.new('RGB', (combined_width, combined_height), (255, 255, 255))
    
    # Paste images side by side
    combined.paste(img1, (0, 0))
    combined.paste(img2, (width, 0))
    
    # Create diff visualization (highlight differences in red)
    diff_visual = img2.copy()
    diff_pixels = diff_visual.load()
    
    for y in range(height):
        for x in range(width):
            if changed_mask[y, x]:
                # Highlight changed pixels in red
                diff_pixels[x, y] = (255, 0, 0)
    
    combined.paste(diff_visual, (width * 2, 0))
    
    # Add labels
    draw = ImageDraw.Draw(combined)
    label_height = 20
    
    # Add white background for labels
    draw.rectangle([(0, 0), (combined_width, label_height)], fill=(255, 255, 255))
    
    # Add text labels
    draw.text((10, 5), "URL 1", fill=(0, 0, 0))
    draw.text((width + 10, 5), "URL 2", fill=(0, 0, 0))
    draw.text((width * 2 + 10, 5), "Diff (red = changed)", fill=(0, 0, 0))
    
    # Save combined image
    combined.save(output_path)
    
    # Also save a standalone diff image
    diff_only_path = str(Path(output_path).parent / "diff_only.png")
    diff_visual.save(diff_only_path)
    
    return {
        'diff_path': output_path,
        'diff_only_path': diff_only_path,
        'total_pixels': int(total_pixels),
        'changed_pixels': int(changed_pixels),
        'change_percentage': float(change_percentage),
        'dimensions': {
            'url1': size1,
            'url2': size2,
            'normalized': (width, height)
        }
    }


def generate_diff_report(
    image1_path: str,
    image2_path: str,
    output_dir: Path,
    threshold: int = 30
) -> dict:
    """
    Generate a complete diff report with visualizations.
    
    Returns:
        Dictionary with diff statistics and file paths
    """
    output_path = output_dir / "diff_comparison.png"
    
    print("Generating visual diff...")
    diff_stats = create_diff_image(
        image1_path,
        image2_path,
        str(output_path),
        threshold
    )
    
    print(f"Diff complete: {diff_stats['change_percentage']:.2f}% pixels changed")
    
    return diff_stats
