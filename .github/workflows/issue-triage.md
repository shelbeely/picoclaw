---
description: Automatically triage new issues in the picoclaw repository by analyzing content and applying appropriate labels
on:
  issues:
    types: [opened, reopened]
  workflow_dispatch:
permissions:
  issues: read
  contents: read
tools:
  github:
    lockdown: true
    toolsets: [issues, labels]
safe-outputs:
  add-labels:
    allowed: [bug, enhancement, documentation, question, help-wanted, good-first-issue, feature, duplicate, invalid, wontfix]
  add-comment:
    max: 1
---

# Issue Triage Agent

Analyze the newly opened issue #${{ github.event.issue.number }} in ${{ github.repository }}.

Read the issue title and body carefully, then:

1. Classify the issue into exactly one of these categories:
   - `bug` — reports a defect, crash, unexpected behavior, or error
   - `feature` — requests a new capability or integration
   - `enhancement` — improvement to an existing feature
   - `documentation` — missing, incorrect, or unclear docs
   - `question` — asking for help or clarification
   - `help-wanted` — good candidates for external contributors
   - `good-first-issue` — simple issues suitable for first-time contributors

2. Apply the matching label to the issue.

3. Leave a brief, welcoming comment explaining the classification. Format:

```markdown
### 🏷️ Issue Triaged

Thanks for opening this issue!

I've classified this as **{label}** because: {one sentence rationale}.

{optional: next steps or pointers to relevant docs/code}

---
*Automatically triaged by the picoclaw issue triage agent.*
```

**Important**: If no action is needed (e.g. issue already has labels), call the `noop` safe-output with a brief explanation.
