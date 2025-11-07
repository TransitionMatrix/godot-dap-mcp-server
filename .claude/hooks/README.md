# Claude Code Hooks

This directory contains optional git hooks for security and development workflow support.

## Available Hooks

### pre-commit.sh

**Purpose**: Provides security checks and development workflow reminders.

**Security Checks (BLOCKING)**:
- 🔒 **Gitleaks secret detection**: Scans for API keys, credentials, private keys, and other secrets
  - Uses comprehensive pattern database (100+ secret types)
  - **Blocks commit** if secrets are detected
  - Provides clear remediation steps
  - Gracefully degrades if gitleaks not installed

**Development Reminders (NON-BLOCKING)**:
- 📝 **Memory sync reminders**: Suggests when to sync Serena memories
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

## Gitleaks Setup (Recommended for Security)

The pre-commit hook uses **gitleaks** for secret detection. While the hook works without it, installing gitleaks is **strongly recommended** for public repositories.

**Installation (macOS)**:
```bash
# Using Homebrew
brew install gitleaks

# Or download binary directly
curl -sSfL https://github.com/gitleaks/gitleaks/releases/latest/download/gitleaks-darwin-arm64 -o /usr/local/bin/gitleaks
chmod +x /usr/local/bin/gitleaks
```

**Installation (Linux)**:
```bash
# Download latest release
curl -sSfL https://github.com/gitleaks/gitleaks/releases/latest/download/gitleaks-linux-amd64 -o /usr/local/bin/gitleaks
chmod +x /usr/local/bin/gitleaks
```

**Verify installation**:
```bash
gitleaks version
```

**Configuration**:

The repository includes `.gitleaks.toml` with sensible defaults. To customize:

1. Edit `.gitleaks.toml` to add allowlists for false positives
2. Test configuration: `gitleaks detect --verbose --source .`
3. Scan specific files: `gitleaks protect --staged --verbose`

**Manual scanning** (without committing):
```bash
# Scan entire repository
gitleaks detect --verbose --source .

# Scan specific file
gitleaks detect --verbose --source . --path internal/tools/connect.go

# Scan unstaged changes
gitleaks protect --verbose
```

## Hook Philosophy

These hooks follow a **layered security** philosophy:

**Security (Blocking)**:
- ✅ Block commits that contain potential secrets
- ✅ Provide clear remediation guidance
- ✅ Allow bypass for emergencies (--no-verify)
- ✅ Gracefully degrade if tools not installed

**Development (Non-blocking)**:
- ✅ Provide helpful context-aware reminders
- ✅ Suggest relevant tools and workflows
- ✅ Always allow commits to proceed (exit 0)
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
