---
description: Generate a weekly summary of repository activity — new issues, merged PRs, and notable changes
on:
  schedule: weekly
  workflow_dispatch:
permissions:
  contents: read
  issues: read
  pull-requests: read
tools:
  github:
    toolsets: [issues, pull_requests, repos]
safe-outputs:
  create-issue:
    max: 1
    title-prefix: "[Weekly Summary] "
    expires: 8d
    close-older-issues: true
---

# Weekly Repository Summary

Generate a weekly activity report for ${{ github.repository }}.

1. **Collect data** for the past 7 days:
   - Newly opened issues (count and titles)
   - Closed/resolved issues (count)
   - Merged pull requests (count, titles, and authors)
   - Open PRs still awaiting review

2. **Create a GitHub issue** titled `[Weekly Summary] Week of {current_date}` with this structure:

```markdown
## 📊 Weekly Activity — {week_start} to {week_end}

### 🐛 Issues
- **Opened**: {count}
- **Closed**: {count}

<details>
<summary>New Issues</summary>

{list of new issues with links}

</details>

### 🔀 Pull Requests
- **Merged**: {count}
- **Open**: {count}

<details>
<summary>Merged PRs</summary>

{list of merged PRs with authors and links}

</details>

### 📝 Highlights
{2–3 sentence narrative summary of notable activity this week}

---
*Generated automatically by the picoclaw weekly summary agent.*
```

**Important**: If there is no activity to report, call the `noop` safe-output with a brief explanation.
