"""
LLM analysis module for visual comparison using OpenAI's vision API.
"""

import base64
import json
from pathlib import Path
from typing import Dict, Any
from openai import OpenAI


def encode_image_to_base64(image_path: str) -> str:
    """Encode image to base64 string."""
    with open(image_path, 'rb') as f:
        return base64.b64encode(f.read()).decode('utf-8')


def analyze_visual_diff(
    screenshot1_path: str,
    screenshot2_path: str,
    diff_path: str,
    css_data1: Dict[str, Any],
    css_data2: Dict[str, Any],
    diff_stats: Dict[str, Any],
    user_question: str,
    model: str = "gpt-4.1-mini"
) -> Dict[str, Any]:
    """
    Analyze visual differences using LLM with vision capabilities.
    
    Args:
        screenshot1_path: Path to first screenshot
        screenshot2_path: Path to second screenshot
        diff_path: Path to diff visualization
        css_data1: CSS data for first element
        css_data2: CSS data for second element
        diff_stats: Statistics from diff generation
        user_question: User's question about the comparison
        model: OpenAI model to use
    
    Returns:
        Dictionary with analysis results
    """
    client = OpenAI()
    
    # Encode images
    print("Encoding images for LLM analysis...")
    img1_b64 = encode_image_to_base64(screenshot1_path)
    img2_b64 = encode_image_to_base64(screenshot2_path)
    diff_b64 = encode_image_to_base64(diff_path)
    
    # Prepare CSS context
    css_context = f"""
## CSS Information

### URL 1: {css_data1['url']}
Selector: `{css_data1['selector']}`
Dimensions: {css_data1['bounding_box']}

Key Computed Styles:
{json.dumps({k: v for k, v in list(css_data1['computed_styles'].items())[:20]}, indent=2)}

Matching CSS Rules: {len(css_data1['matching_rules'])} rules found

### URL 2: {css_data2['url']}
Selector: `{css_data2['selector']}`
Dimensions: {css_data2['bounding_box']}

Key Computed Styles:
{json.dumps({k: v for k, v in list(css_data2['computed_styles'].items())[:20]}, indent=2)}

Matching CSS Rules: {len(css_data2['matching_rules'])} rules found

## Diff Statistics
- Total pixels: {diff_stats['total_pixels']:,}
- Changed pixels: {diff_stats['changed_pixels']:,}
- Change percentage: {diff_stats['change_percentage']:.2f}%
"""
    
    # Prepare system prompt
    system_prompt = """You are an expert web developer and visual designer specializing in CSS and UI/UX analysis. 
You are analyzing visual differences between two web page elements captured from different URLs.

Your task is to:
1. Carefully examine the three images provided (URL 1, URL 2, and the diff visualization)
2. Analyze the CSS information provided for both elements
3. Identify visual differences and their likely causes
4. Answer the user's specific question with detailed, actionable insights

Focus on:
- Visual appearance changes (colors, fonts, spacing, layout)
- CSS property differences that explain visual changes
- Potential causes (CSS rule changes, inheritance, specificity)
- Impact on user experience
- Recommendations for addressing differences if needed
"""
    
    # Prepare user prompt
    user_prompt = f"""
I'm comparing two web page elements from different URLs. Please analyze the visual differences and CSS information.

{css_context}

**User Question:** {user_question}

Please provide a comprehensive analysis addressing the question above.
"""
    
    print("Sending request to LLM...")
    
    # Make API call
    response = client.chat.completions.create(
        model=model,
        messages=[
            {
                "role": "system",
                "content": system_prompt
            },
            {
                "role": "user",
                "content": [
                    {
                        "type": "text",
                        "text": user_prompt
                    },
                    {
                        "type": "image_url",
                        "image_url": {
                            "url": f"data:image/png;base64,{img1_b64}",
                            "detail": "high"
                        }
                    },
                    {
                        "type": "image_url",
                        "image_url": {
                            "url": f"data:image/png;base64,{img2_b64}",
                            "detail": "high"
                        }
                    },
                    {
                        "type": "image_url",
                        "image_url": {
                            "url": f"data:image/png;base64,{diff_b64}",
                            "detail": "high"
                        }
                    }
                ]
            }
        ],
        max_tokens=2000,
        temperature=0.7
    )
    
    analysis = response.choices[0].message.content
    
    return {
        'analysis': analysis,
        'model': model,
        'tokens_used': {
            'prompt': response.usage.prompt_tokens,
            'completion': response.usage.completion_tokens,
            'total': response.usage.total_tokens
        }
    }


def save_analysis_report(
    analysis_result: Dict[str, Any],
    output_dir: Path
) -> str:
    """
    Save the analysis report to a file.
    
    Returns:
        Path to the saved report
    """
    report_path = output_dir / "analysis_report.md"
    
    with open(report_path, 'w') as f:
        f.write("# Visual Diff Analysis Report\n\n")
        f.write(f"**Model:** {analysis_result['model']}\n\n")
        f.write(f"**Tokens Used:** {analysis_result['tokens_used']['total']} ")
        f.write(f"(prompt: {analysis_result['tokens_used']['prompt']}, ")
        f.write(f"completion: {analysis_result['tokens_used']['completion']})\n\n")
        f.write("---\n\n")
        f.write(analysis_result['analysis'])
        f.write("\n")
    
    return str(report_path)
