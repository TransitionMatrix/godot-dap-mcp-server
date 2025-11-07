#!/bin/bash
#
# Pre-commit hook for godot-dap-mcp-server
#
# Security Checks (BLOCKING):
#   - Gitleaks: Scans for secrets, API keys, credentials
#
# Development Reminders (NON-BLOCKING):
#   - Memory sync reminders for significant code changes
#
# To enable this hook in your git repository:
#   ln -sf ../../.claude/hooks/pre-commit.sh .git/hooks/pre-commit
#

# Colors for output
RED='\033[1;31m'
YELLOW='\033[1;33m'
GREEN='\033[1;32m'
NC='\033[0m' # No Color

#######################################
# SECURITY CHECKS (BLOCKING)
#######################################

# Check for secrets using gitleaks
if command -v gitleaks &> /dev/null; then
    echo -e "${GREEN}🔒 Running gitleaks secret detection...${NC}"

    # Run gitleaks on staged files
    if ! gitleaks protect --staged --verbose --redact --no-banner 2>&1; then
        echo ""
        echo -e "${RED}❌ COMMIT BLOCKED: Gitleaks detected potential secrets!${NC}"
        echo ""
        echo "   Gitleaks found patterns that may be secrets, API keys, or credentials."
        echo ""
        echo "   Next steps:"
        echo "   1. Review the findings above carefully"
        echo "   2. Remove any actual secrets from your code"
        echo "   3. If this is a false positive, add to .gitleaks.toml allowlist"
        echo "   4. Consider using environment variables or config files for secrets"
        echo ""
        echo "   To bypass this check (NOT RECOMMENDED for public repos):"
        echo "   git commit --no-verify"
        echo ""
        exit 1
    fi
    echo -e "${GREEN}✓ No secrets detected${NC}"
    echo ""
else
    echo -e "${YELLOW}⚠️  Gitleaks not installed - skipping secret detection${NC}"
    echo "   Install with: brew install gitleaks"
    echo "   Or download from: https://github.com/gitleaks/gitleaks/releases"
    echo ""
fi

#######################################
# DEVELOPMENT REMINDERS (NON-BLOCKING)
#######################################

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
