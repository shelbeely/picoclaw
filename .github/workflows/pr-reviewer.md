---
description: Provide an AI-assisted code review for pull requests in the picoclaw repository
on:
  pull_request:
    types: [opened, synchronize, ready_for_review]
  workflow_dispatch:
permissions:
  contents: read
  pull-requests: read
tools:
  github:
    lockdown: true
    toolsets: [pull_requests, repos]
safe-outputs:
  add-comment:
    max: 2
---

# PR Review Assistant

Review pull request #${{ github.event.pull_request.number }} in ${{ github.repository }}.

The PR title is: **${{ github.event.pull_request.title }}**

Perform the following review steps:

1. **Understand the change**: Read the PR description and the diff of changed files.

2. **Check for common issues**:
   - Missing or incorrect error handling in Go code
   - Unprotected concurrent access (missing mutex/atomic where needed)
   - Context propagation — all long-running operations should accept `context.Context`
   - Proper use of the channel/bus/provider interfaces
   - New exported symbols that lack godoc comments
   - Test coverage for new or changed behavior

3. **Review security considerations**:
   - No hardcoded credentials or API keys
   - Proper input validation / sanitisation in channel handlers
   - Safe handling of user-supplied data before passing to tools

4. **Post a review comment** summarising your findings. Use this format:

```markdown
### 🤖 Automated Code Review

**Summary**: {brief one-liner of the change}

#### ✅ Looks Good
- {positive observation 1}

#### 💡 Suggestions
- {suggestion with file+line reference if possible}

#### ⚠️ Issues Found
- {blocking issue if any — otherwise omit this section}

---
*Review generated automatically. A human reviewer must approve before merging.*
```

If the PR is a draft, still review but note it at the top of the comment.

**Important**: If no action is needed (e.g. trivial change such as a typo fix), call the `noop` safe-output with a brief explanation.
