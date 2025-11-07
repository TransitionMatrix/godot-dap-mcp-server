#!/bin/bash
#
# Pre-commit hook reminder for memory sync
# This hook provides a reminder to sync memories when significant code changes are made.
# It is NON-BLOCKING - it only reminds, never prevents commits.
#
# To enable this hook in your git repository:
#   ln -sf ../../.claude/hooks/pre-commit.sh .git/hooks/pre-commit
#

# Colors for output
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if this is a significant change that might require memory sync
significant_changes=false

# Check for changes in internal/ or cmd/ directories (excluding tests)
if git diff --cached --name-only | grep -E "(internal/|cmd/)" | grep -v "_test.go" | wc -l | grep -v "^0" > /dev/null; then
    significant_changes=true
fi

# Check for changes to documentation structure
if git diff --cached --name-only | grep -E "(docs/.*\.md|CLAUDE\.md)" | wc -l | grep -v "^0" > /dev/null; then
    doc_changes=true
fi

# If significant changes detected, provide reminder
if [ "$significant_changes" = true ]; then
    echo ""
    echo -e "${YELLOW}📝 Memory Sync Reminder${NC}"
    echo "   Detected changes in internal/ or cmd/ directories"
    echo ""
    echo "   Consider syncing Serena memories if:"
    echo "   ✅ Architecture patterns changed"
    echo "   ✅ New conventions introduced"
    echo "   ✅ Critical patterns discovered"
    echo "   ✅ Phase completion or major refactoring"
    echo ""
    echo "   To sync: /memory-sync"
    echo ""
fi

# Check if .serena/memories/ has changes without corresponding code changes
if git diff --cached --name-only | grep -E "^\.serena/memories/" > /dev/null; then
    memory_changes=true

    if [ "$significant_changes" = false ]; then
        echo ""
        echo -e "${YELLOW}📝 Memory Update Detected${NC}"
        echo "   Memories updated without corresponding code changes"
        echo ""
        echo "   Consider if documentation also needs updates:"
        echo "   - docs/ARCHITECTURE.md"
        echo "   - docs/PLAN.md"
        echo "   - docs/reference/CONVENTIONS.md"
        echo ""
    fi
fi

# Always allow the commit to proceed
exit 0
