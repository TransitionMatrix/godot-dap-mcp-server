# Claude Code Hooks

This directory contains optional git hooks that provide helpful reminders during development.

## Available Hooks

### pre-commit.sh

**Purpose**: Reminds you to sync Serena memories when significant code changes are detected.

**Behavior**:
- **NON-BLOCKING**: Never prevents commits, only provides reminders
- Triggers on changes to `internal/` or `cmd/` directories
- Suggests when to use `/memory-sync` skill
- Reminds about doc updates when memories are changed

**Installation**:

To enable this hook in your local repository:

```bash
# From project root
ln -sf ../../.claude/hooks/pre-commit.sh .git/hooks/pre-commit
```

**To verify it's working:**

```bash
# Check if hook is linked
ls -la .git/hooks/pre-commit

# Should show: .git/hooks/pre-commit -> ../../.claude/hooks/pre-commit.sh
```

**To disable:**

```bash
# Remove the symlink
rm .git/hooks/pre-commit
```

## Hook Philosophy

These hooks follow a **reminder-only** philosophy:
- ✅ Provide helpful context-aware reminders
- ✅ Suggest relevant tools and workflows
- ✅ Always allow commits to proceed (exit 0)
- ❌ Never block or prevent commits
- ❌ Never run long-running operations
- ❌ Never modify files automatically

## Adding New Hooks

If you create additional hooks:
1. Place the script in `.claude/hooks/`
2. Make it executable: `chmod +x .claude/hooks/your-hook.sh`
3. Document it in this README
4. Ensure it's non-blocking (always `exit 0`)
5. Follow the naming convention: `<hook-name>.sh`

## Hook Triggers Reference

**pre-commit** - Before a commit is created
- Use for: Reminders about memory sync, linting suggestions, test reminders
- Install: `ln -sf ../../.claude/hooks/pre-commit.sh .git/hooks/pre-commit`

**post-commit** - After a commit is created
- Use for: Success messages, next-step suggestions
- Install: `ln -sf ../../.claude/hooks/post-commit.sh .git/hooks/post-commit`

**pre-push** - Before pushing to remote
- Use for: Reminders about documentation, release notes
- Install: `ln -sf ../../.claude/hooks/pre-push.sh .git/hooks/pre-push`

## Notes

- Hooks are **opt-in**: Users must explicitly install them
- Hooks are **not tracked in .git/hooks/**: Only `.claude/hooks/` is version-controlled
- This ensures hooks don't interfere with other git workflows
- Each developer can choose which hooks to enable
