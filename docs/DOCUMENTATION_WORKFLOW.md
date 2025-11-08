# Documentation Workflow

**Last Updated**: 2025-11-07

This document explains how to maintain documentation in this project, particularly when completing implementation phases that involve debugging or discovering new patterns.

---

## Table of Contents

1. [Philosophy](#philosophy)
2. [Four-Document Pattern](#four-document-pattern)
3. [When to Create Lessons Learned](#when-to-create-lessons-learned)
4. [Workflow Example](#workflow-example)
5. [Cross-Reference Guidelines](#cross-reference-guidelines)
6. [Benefits](#benefits)

---

## Philosophy

This project uses a **hybrid documentation approach** that balances:
- **Narrative preservation** - The story of HOW we discovered patterns through debugging
- **Pattern discoverability** - Making reusable patterns easy to find when implementing features
- **Architectural clarity** - Explaining WHY decisions matter for the system
- **Quick troubleshooting** - Providing immediate answers WHEN you hit issues

The result is **strategic redundancy with purpose** - knowledge appears in multiple places, but each instance serves a different audience and use case.

---

## Four-Document Pattern

When you complete a phase that involves significant debugging or pattern discovery, update these four types of documents:

### 1. Phase-Specific Lessons Learned (e.g., `LESSONS_LEARNED_PHASE_3.md`)

**Purpose**: Preserve the debugging narrative and educational journey

**Contains**:
- Complete chronological story of issues encountered
- Dead-ends and failed approaches (learning opportunities)
- "Aha!" moments and critical discoveries
- Full code examples showing before/after
- Thought process behind solutions
- Future considerations and unanswered questions

**Audience**:
- Developers debugging similar issues
- Future maintainers wanting to understand WHY
- Anyone learning from the implementation journey

**Example**: `docs/LESSONS_LEARNED_PHASE_3.md`

**When to create**: After completing any phase involving:
- Multiple debugging iterations (3+ issue/solution cycles)
- Discovery of non-obvious patterns
- Integration challenges with external systems
- Solutions that required research or experimentation

### 2. Implementation Guide (`IMPLEMENTATION_GUIDE.md`)

**Purpose**: Provide reusable patterns and code templates (WHAT to use)

**Contains**:
- Complete working code patterns
- "When to use this pattern" guidance
- Common mistakes to avoid
- Template code that can be copied
- Integration testing patterns
- Step-by-step procedures

**Audience**:
- Developers implementing new features
- Contributors adding tools or components
- Anyone looking for "how do I do X" answers

**When to update**: Extract patterns from lessons learned that:
- Are reusable across multiple features
- Solve common problems
- Would benefit from a template
- Need to be discoverable by feature area

### 3. Architecture Documentation (`ARCHITECTURE.md`)

**Purpose**: Explain critical patterns and design decisions (WHY it matters)

**Contains**:
- System-wide architectural patterns
- Critical reliability patterns
- State machines and flow diagrams
- Design decision rationale
- Performance or safety-critical implementations
- Cross-cutting concerns

**Audience**:
- Developers making architectural decisions
- Reviewers evaluating design choices
- Anyone trying to understand system design

**When to update**: Extract from lessons learned when:
- Pattern is critical for system reliability
- Discovery changes architectural understanding
- Pattern affects multiple layers
- Design decision has non-obvious rationale

### 4. FAQ / Troubleshooting (`reference/GODOT_DAP_FAQ.md`)

**Purpose**: Provide quick answers for specific problems (WHEN you hit issues)

**Contains**:
- Q&A format for common issues
- Quick reference tables
- Known limitations and workarounds
- "Why isn't X working?" answers
- Platform-specific quirks

**Audience**:
- Developers hitting specific errors
- Users troubleshooting integrations
- Anyone searching for quick answers

**When to update**: Add entries for:
- Common error messages
- Known limitations discovered
- Platform-specific behaviors
- Frequently asked questions

---

## When to Create Lessons Learned

Create a phase-specific lessons learned document when:

### ✅ Must Create
- You encountered **3+ significant debugging iterations** to complete the phase
- You discovered **patterns not obvious from reading specs**
- You hit **dead-ends** that would help others avoid same mistakes
- The solution required **research or experimentation**
- The debugging story has **educational value**

### ⚠️ Consider Creating
- Phase involved integration with unfamiliar systems
- Multiple approaches were tried before finding solution
- Community would benefit from the implementation story
- Pattern discovery process was non-linear

### ❌ Skip (just update other docs)
- Straightforward implementation following existing patterns
- No significant debugging required
- No new patterns discovered
- Simple bug fixes

---

## Workflow Example

### Phase 3 Integration Testing Example

**Context**: Implementing automated integration tests revealed multiple issues with persistent sessions, protocol handshakes, and path handling.

**Step 1: Complete the implementation and capture notes**

During debugging, track:
- Each issue encountered
- Approaches tried (including failures)
- Final solutions
- Code snippets showing before/after

**Step 2: Create phase-specific lessons learned**

Created `docs/LESSONS_LEARNED_PHASE_3.md` with:
- 7 critical issues (coprocess, named pipes, ConfigurationDone, paths, JSON parsing, etc.)
- Complete narrative for each issue
- Code examples showing the evolution
- "Aha!" moments and discoveries

**Step 3: Extract to implementation guide**

Added to `docs/IMPLEMENTATION_GUIDE.md`:
- Pattern: Persistent Subprocess Communication (complete bash template)
- Pattern: Port Conflict Handling (graceful fallback)
- Pattern: Robust JSON Parsing in Bash (regex pattern)

Each pattern includes:
- Complete working code
- "Why this works" explanation
- "When to use" guidance
- "Common mistake" warnings

**Step 4: Update architecture documentation**

Enhanced `docs/ARCHITECTURE.md`:
- Expanded "DAP Protocol Handshake Pattern"
- Added state machine diagram
- Emphasized ConfigurationDone requirement
- Cross-reference to lessons learned for full story

**Step 5: Add to FAQ**

Updated `docs/reference/GODOT_DAP_FAQ.md`:
- Added initialization sequence with ConfigurationDone requirement
- Clarified res:// path limitation (Godot-specific)
- Added new section: "MCP Bridge Considerations"

**Step 6: Add cross-references**

In `LESSONS_LEARNED_PHASE_3.md` header:
```markdown
**Reusable Patterns → [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md#integration-testing-patterns)**
**Critical Implementation Details → [ARCHITECTURE.md](ARCHITECTURE.md#critical-implementation-patterns)**
**Known Quirks & FAQ → [GODOT_DAP_FAQ.md](reference/GODOT_DAP_FAQ.md)**
```

From other docs back to lessons learned:
```markdown
For the full debugging story, see [LESSONS_LEARNED_PHASE_3.md](LESSONS_LEARNED_PHASE_3.md).
```

---

## Cross-Reference Guidelines

### From Lessons Learned → Other Docs
- Add header section explaining where patterns were extracted
- Use clear labels: "Reusable Patterns →", "Critical Details →", "FAQ →"
- Keep the narrative intact in lessons learned

### From Other Docs → Lessons Learned
- Add brief cross-reference when mentioning discoveries
- Use: "For the full debugging story, see..."
- Link to specific sections when relevant (use anchors)

### Example Cross-References

**In IMPLEMENTATION_GUIDE.md:**
```markdown
### Pattern: Persistent Subprocess Communication

These patterns were discovered while implementing Phase 3 integration tests.
For the full debugging story, see [LESSONS_LEARNED_PHASE_3.md](LESSONS_LEARNED_PHASE_3.md).
```

**In ARCHITECTURE.md:**
```markdown
**Critical Discovery** (Phase 3 debugging):
- Without `ConfigurationDone()`, session remains in "initialized" state
- ...

For full debugging story, see [LESSONS_LEARNED_PHASE_3.md](LESSONS_LEARNED_PHASE_3.md#issue-2-dap-protocol-handshake-incomplete).
```

**In LESSONS_LEARNED_PHASE_3.md header:**
```markdown
## 📚 Documentation Organization (Hybrid Approach)

This document preserves the **debugging narrative**. The **reusable patterns**
have been extracted to appropriate reference documentation:

**Reusable Patterns → [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md#integration-testing-patterns)**:
- Pattern 1: Persistent Subprocess Communication
- Pattern 2: Port Conflict Handling
...
```

---

## Benefits

### For Feature Implementation
- **Quick pattern lookup** - Find templates in IMPLEMENTATION_GUIDE.md
- **Architectural context** - Understand WHY from ARCHITECTURE.md
- **Troubleshooting** - Get quick answers from FAQ

### For Debugging
- **Learn from similar issues** - Read lessons learned narratives
- **Avoid dead-ends** - See what didn't work
- **Understand evolution** - See how solutions were discovered

### For Maintenance
- **Centralized patterns** - Easy to update in one place
- **Historical context** - Understand decisions years later
- **Onboarding** - New contributors can learn the journey

### For the Project
- **Knowledge preservation** - Insights aren't lost to time
- **Community value** - Others benefit from debugging stories
- **Quality improvement** - Patterns become repeatable

---

## Document Naming Convention

### Phase-Specific Lessons Learned
Format: `LESSONS_LEARNED_PHASE_<N>.md`

Examples:
- `LESSONS_LEARNED_PHASE_3.md` - Integration testing patterns
- `LESSONS_LEARNED_PHASE_5.md` - Scene launching implementation
- `LESSONS_LEARNED_PHASE_7.md` - Error handling and polish

### Reference Documentation
Stable names that grow over time:
- `ARCHITECTURE.md` - System architecture (add patterns as discovered)
- `IMPLEMENTATION_GUIDE.md` - Implementation patterns (add sections per phase)
- `reference/GODOT_DAP_FAQ.md` - Q&A (add entries as issues arise)

---

## Quick Checklist

When completing a phase with significant debugging:

- [ ] Create `LESSONS_LEARNED_PHASE_<N>.md` with complete narrative
- [ ] Extract reusable patterns to `IMPLEMENTATION_GUIDE.md`
- [ ] Add critical patterns to `ARCHITECTURE.md`
- [ ] Add Q&A entries to relevant FAQ
- [ ] Add cross-references in both directions
- [ ] Update README.md if new doc types added
- [ ] Commit with message: "docs: Phase N lessons learned and pattern extraction"

---

## Related Documentation

- [CLAUDE.md](../CLAUDE.md) - AI-assisted development workflow (includes memory vs documentation strategy)
- [README.md](../README.md) - Main project documentation
- [PLAN.md](PLAN.md) - Implementation phases and timeline

---

**This workflow was established during Phase 3 (2025-11-07)** based on the integration testing debugging journey. It formalizes the hybrid approach for future phases.
