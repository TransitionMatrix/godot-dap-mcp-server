#!/bin/bash
# setup-skills.sh - Fetch required Claude skills for this project
#
# This script pulls specific skills from GitHub repositories into .claude/skills/
# These are external dependencies (like node_modules) and are gitignored.
# To add project-specific skills later, exclude them from .gitignore individually.

set -e  # Exit on error

# Check if git is installed (should be available on most systems)
if ! command -v git &> /dev/null; then
    echo "Error: git is not installed."
    echo "Please install git first."
    exit 1
fi

# Navigate to project root's .claude/skills directory
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
SKILLS_DIR="$PROJECT_ROOT/.claude/skills"

echo "Setting up Claude skills in: $SKILLS_DIR"

# Create skills directory if it doesn't exist
mkdir -p "$SKILLS_DIR"
cd "$SKILLS_DIR"

# Fetch mcp-builder skill
echo "Fetching mcp-builder skill..."
if [ -d "mcp-builder" ]; then
    echo "  (mcp-builder already exists, skipping)"
else
    # Use git sparse-checkout to fetch only the specific subdirectory
    TEMP_DIR=$(mktemp -d)
    git clone --depth 1 --filter=blob:none --sparse \
        https://github.com/sammcj/anthropic-skills-fork.git "$TEMP_DIR"
    cd "$TEMP_DIR"
    git sparse-checkout set mcp-builder
    cd "$SKILLS_DIR"
    mv "$TEMP_DIR/mcp-builder" .
    rm -rf "$TEMP_DIR"
    echo "  ✓ mcp-builder installed"
fi

# Add more skills here as needed:
# echo "Fetching other-skill..."
# if [ -d "other-skill" ]; then
#     echo "  (other-skill already exists, skipping)"
# else
#     TEMP_DIR=$(mktemp -d)
#     git clone --depth 1 --filter=blob:none --sparse \
#         https://github.com/other-org/other-repo.git "$TEMP_DIR"
#     cd "$TEMP_DIR"
#     git sparse-checkout set path/to/skill
#     cd "$SKILLS_DIR"
#     mv "$TEMP_DIR/path/to/skill" other-skill
#     rm -rf "$TEMP_DIR"
#     echo "  ✓ other-skill installed"
# fi

echo ""
echo "✓ Skills setup complete!"
echo ""
echo "Note: These are external skills (gitignored like dependencies)."
echo "To create project-specific skills, add them to .claude/skills/ and"
echo "update .gitignore to exclude them from the ignore pattern."
