---
name: blog
description: "Write, format, and publish Jekyll blog posts for the PicoClaw GitHub Pages blog."
metadata: {"nanobot":{"emoji":"✍️"}}
---

# Blog Skill

This skill covers everything needed to produce a publication-ready Jekyll post
for the PicoClaw GitHub Pages blog hosted from `docs/`.

---

## Jekyll front matter template

Every post **must** start with this block (no blank lines before it):

```markdown
---
layout: post
title: "Your Engaging Title Here"
date: 2025-01-20 09:00:00 +0000
categories: [category1, category2]
tags: [tag1, tag2, tag3]
excerpt: "One sentence shown on the index page — make it a hook."
---
```

### Field rules

| Field | Rule |
|-------|------|
| `layout` | Always `post` |
| `title` | Title-cased, wrapped in double quotes |
| `date` | UTC datetime: `YYYY-MM-DD HH:MM:SS +0000` |
| `categories` | 1–3 broad categories in a YAML list |
| `tags` | 3–6 specific keywords in a YAML list |
| `excerpt` | Max 160 characters, no Markdown |

---

## Post structure

```
---
<front matter>
---

## Introduction
Hook the reader. State the problem or surprising fact. (~100 words)

## <Section 1 from outline>
...

## <Section 2 from outline>
...

## <Section N from outline>
...

## Conclusion
Restate the key insight. End with a call to action or open question.
```

---

## Style guide

- **Voice:** First person, direct, opinionated but evidence-based.
- **Length:** 600–1 200 words for a standard post.
- **Links:** Inline Markdown links `[text](url)` to authoritative sources.
- **Code:** Fenced code blocks with language tag (` ```python `, ` ```bash `).
- **Emphasis:** `**bold**` for key terms on first use; avoid over-bolding.
- **Lists:** Use sparingly — prefer prose for narrative, lists for reference.
- **Headings:** `##` for main sections, `###` for sub-sections; no `#` (reserved for title).

---

## Swarm pipeline summary

The blog post pipeline runs entirely on GitHub Actions:

```
workflow_dispatch (topic)
  └─ orchestrate   → title + outline + research_questions[]
       └─ research (matrix, parallel)  → one note per question
            └─ write   → full Jekyll draft
                 └─ edit   → polished final post → git commit → Pages
```

Each stage is a separate GitHub Actions job. The `action.yml` at the repo
root exposes `picoclaw agent` as a reusable composite action so any workflow
step can invoke an agent with a single `uses: shelbeely/picoclaw@main`.

---

## Checking the blog

Posts land in `docs/_posts/YYYY-MM-DD-slug.md`. GitHub Pages (Jekyll) picks
them up automatically on the next `push` to `main`.

View the live blog at: `https://<org>.github.io/<repo>/`
