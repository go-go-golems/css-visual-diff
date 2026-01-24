#!/bin/bash

# Installation script for CSS Visual Diff

set -e

echo "=========================================="
echo "CSS Visual Diff - Installation"
echo "=========================================="
echo ""

# Check Python version
echo "Checking Python version..."
python3 --version || { echo "Error: Python 3 is required"; exit 1; }
echo "✓ Python 3 found"
echo ""

# Install Python dependencies
echo "Installing Python dependencies..."
pip3 install -r requirements.txt
echo "✓ Python dependencies installed"
echo ""

# Install Playwright browsers
echo "Installing Playwright browsers (Chromium)..."
playwright install chromium
echo "✓ Playwright browsers installed"
echo ""

# Check for OpenAI API key
if [ -z "$OPENAI_API_KEY" ]; then
    echo "⚠️  Warning: OPENAI_API_KEY environment variable not set"
    echo "   You'll need to set it before running the tool:"
    echo "   export OPENAI_API_KEY='your-api-key-here'"
    echo ""
else
    echo "✓ OPENAI_API_KEY found"
    echo ""
fi

# Make scripts executable
chmod +x src/cli.py
chmod +x tests/run_tests.sh

echo "=========================================="
echo "Installation Complete!"
echo "=========================================="
echo ""
echo "Usage:"
echo "  python3 src/cli.py --help"
echo ""
echo "Run tests:"
echo "  cd tests && ./run_tests.sh"
echo ""
