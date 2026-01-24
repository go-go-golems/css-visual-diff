#!/usr/bin/env python3
"""
CSS Visual Diff - Compare web page elements visually with AI analysis.
"""

import sys
import json
from pathlib import Path
from datetime import datetime
import click

from browser_capture import capture_both_elements
from image_diff import generate_diff_report
from llm_analysis import analyze_visual_diff, save_analysis_report


@click.command()
@click.option('--url1', required=True, help='First URL to compare')
@click.option('--selector1', required=True, help='CSS selector for element in first URL')
@click.option('--url2', required=True, help='Second URL to compare')
@click.option('--selector2', required=True, help='CSS selector for element in second URL')
@click.option('--question', required=True, help='Question to ask the LLM about the comparison')
@click.option('--output-dir', default=None, help='Output directory (default: ./output_TIMESTAMP)')
@click.option('--threshold', default=30, type=int, help='Pixel difference threshold (0-255, default: 30)')
@click.option('--model', default='gpt-4.1-mini', help='LLM model to use (default: gpt-4.1-mini)')
@click.option('--no-analysis', is_flag=True, help='Skip LLM analysis (only capture and diff)')
def main(url1, selector1, url2, selector2, question, output_dir, threshold, model, no_analysis):
    """
    Compare web page elements visually with AI-powered analysis.
    
    This tool captures screenshots of specific elements from two URLs,
    generates a visual diff, extracts CSS information, and uses an LLM
    to analyze the differences.
    
    Example:
    
        css-visual-diff \\
            --url1 https://example.com/page1 \\
            --selector1 ".header-nav" \\
            --url2 https://example.com/page2 \\
            --selector2 ".navigation-header" \\
            --question "What are the main visual differences and what CSS changes caused them?"
    """
    try:
        # Setup output directory
        if output_dir is None:
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            output_dir = Path(f"./output_{timestamp}")
        else:
            output_dir = Path(output_dir)
        
        output_dir.mkdir(parents=True, exist_ok=True)
        
        print("=" * 80)
        print("CSS Visual Diff Tool")
        print("=" * 80)
        print(f"\nOutput directory: {output_dir.absolute()}\n")
        
        # Step 1: Capture elements
        print("Step 1: Capturing elements from both URLs...")
        print("-" * 80)
        element1_data, element2_data = capture_both_elements(
            url1, selector1,
            url2, selector2,
            output_dir
        )
        print(f"✓ Screenshots saved:")
        print(f"  - {element1_data['screenshot_path']}")
        print(f"  - {element2_data['screenshot_path']}")
        print(f"✓ CSS data saved:")
        print(f"  - {element1_data['css_path']}")
        print(f"  - {element2_data['css_path']}")
        print()
        
        # Step 2: Generate diff
        print("Step 2: Generating visual diff...")
        print("-" * 80)
        diff_stats = generate_diff_report(
            element1_data['screenshot_path'],
            element2_data['screenshot_path'],
            output_dir,
            threshold
        )
        print(f"✓ Diff images saved:")
        print(f"  - {diff_stats['diff_path']}")
        print(f"  - {diff_stats['diff_only_path']}")
        print(f"✓ Change statistics:")
        print(f"  - Total pixels: {diff_stats['total_pixels']:,}")
        print(f"  - Changed pixels: {diff_stats['changed_pixels']:,}")
        print(f"  - Change percentage: {diff_stats['change_percentage']:.2f}%")
        print()
        
        # Step 3: LLM Analysis
        if not no_analysis:
            print("Step 3: Analyzing with LLM...")
            print("-" * 80)
            analysis_result = analyze_visual_diff(
                element1_data['screenshot_path'],
                element2_data['screenshot_path'],
                diff_stats['diff_path'],
                element1_data['css_data'],
                element2_data['css_data'],
                diff_stats,
                question,
                model
            )
            
            report_path = save_analysis_report(analysis_result, output_dir)
            print(f"✓ Analysis complete (tokens used: {analysis_result['tokens_used']['total']})")
            print(f"✓ Report saved: {report_path}")
            print()
            
            # Display analysis
            print("=" * 80)
            print("ANALYSIS RESULTS")
            print("=" * 80)
            print()
            print(analysis_result['analysis'])
            print()
        else:
            print("Step 3: Skipping LLM analysis (--no-analysis flag set)")
            print()
        
        # Save summary
        summary = {
            'url1': url1,
            'selector1': selector1,
            'url2': url2,
            'selector2': selector2,
            'question': question,
            'output_dir': str(output_dir.absolute()),
            'diff_stats': diff_stats,
            'timestamp': datetime.now().isoformat()
        }
        
        if not no_analysis:
            summary['analysis'] = {
                'model': analysis_result['model'],
                'tokens_used': analysis_result['tokens_used']
            }
        
        summary_path = output_dir / 'summary.json'
        with open(summary_path, 'w') as f:
            json.dump(summary, f, indent=2)
        
        print("=" * 80)
        print(f"✓ All outputs saved to: {output_dir.absolute()}")
        print("=" * 80)
        
    except Exception as e:
        print(f"\n❌ Error: {str(e)}", file=sys.stderr)
        import traceback
        traceback.print_exc()
        sys.exit(1)


if __name__ == '__main__':
    main()
