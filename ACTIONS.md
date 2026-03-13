# PicoClaw GitHub Actions — Reference Guide

This document covers every new GitHub Actions artefact in this fork:
the reusable **`action.yml`** composite action, the four **workflows**
(Blog Writer, Swarm, Research Swarm, Pages), and the **docs/** Jekyll blog that they publish to.

---

## Table of contents

1. [Overview](#overview)
2. [Quick start](#quick-start)
3. [Secrets and variables](#secrets-and-variables)
4. [OpenRouter setup and rate limits](#openrouter-setup-and-rate-limits)
5. [`action.yml` — PicoClaw AI Agent action](#actionyml--picoclaw-ai-agent-action)
6. [Workflow: Blog Writer](#workflow-blog-writer)
7. [Workflow: PicoClaw Swarm](#workflow-picoclaw-swarm)
8. [Workflow: Research Swarm](#workflow-research-swarm)
9. [Workflow: Deploy to GitHub Pages](#workflow-deploy-to-github-pages)
10. [Jekyll blog scaffold](#jekyll-blog-scaffold)
11. [Workspace persona](#workspace-persona)
12. [Customisation](#customisation)

---

## Overview

This fork repurposes PicoClaw as a **GitHub Actions–native AI blog system**.
No server, no self-hosted runner, no persistent process — the entire pipeline
runs on GitHub's free compute tier.

```
┌─────────────────────────────────────────────────────────────────┐
│  Blog Writer workflow (weekly schedule or workflow_dispatch)    │
│  ─────────────────────────────────────────────────────────────  │
│  matrix: [topic1, topic2, topic3]  ← parallel picoclaw agents  │
│  → docs/_posts/YYYY-MM-DD-slug.md                               │
│  → git push → Pages deploy                                      │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  Swarm workflow (workflow_dispatch)                             │
│  ─────────────────────────────────────────────────────────────  │
│  orchestrate → research[N] (parallel) → write → edit → commit  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  Research Swarm workflow (workflow_dispatch)                    │
│  ─────────────────────────────────────────────────────────────  │
│  setup → research[10 shards] (parallel) → reduce → report      │
│  → docs/research/YYYY-MM-DD-<slug>.md + artifact               │
└─────────────────────────────────────────────────────────────────┘
```

The reusable `action.yml` is the building block all workflows share.
Any workflow — in this repo or in another — can invoke a PicoClaw agent
with a single `uses:` step.

---

## Quick start

### 1. Fork and enable Pages

1. Fork **shelbeely/picoclaw**.
2. Go to **Settings → Pages**, set source to **Deploy from a branch**, branch
   `main`, folder `/docs`.  
   *(The `pages.yml` workflow handles this automatically via GitHub Actions
   once the Pages environment is created.)*

### 2. Add your API key

Go to **Settings → Secrets and variables → Actions → New repository secret**:

| Secret name | Value |
|---|---|
| `OPENROUTER_API_KEY` | Your OpenRouter API key (recommended) |
| `AI_API_KEY` | Any other provider key (fallback) |

> **Tip:** Only one of these is needed. If both are set, `OPENROUTER_API_KEY`
> takes precedence.

### 3. Run the Blog Writer

Go to **Actions → Blog Writer → Run workflow**.  
Accept the defaults to write three posts with the free Llama model on
OpenRouter.  Posts appear in `docs/_posts/` on the next commit and the
Pages site updates automatically.

---

## Secrets and variables

| Name | Kind | Required | Description |
|---|---|---|---|
| `OPENROUTER_API_KEY` | Secret | One of these two | OpenRouter API key |
| `AI_API_KEY` | Secret | One of these two | Any other provider key |
| `OPENROUTER_MODEL` | Variable | No | Default model for the Research Swarm (takes precedence over `AI_MODEL`). |
| `AI_MODEL` | Variable | No | Default model for Blog Writer and Swarm. Falls back to `openrouter/meta-llama/llama-3.1-8b-instruct:free` |

Set variables at **Settings → Secrets and variables → Actions → Variables**.

---

## OpenRouter setup and rate limits

[OpenRouter](https://openrouter.ai) proxies 200+ models behind a single
API key, including many **free-tier** variants (model IDs ending in `:free`).

### Getting an API key

1. Create an account at <https://openrouter.ai>.
2. Go to **Keys** and create a new key.
3. Add the key as `OPENROUTER_API_KEY` in your repo secrets.

### Model format

OpenRouter models follow the pattern `openrouter/<provider>/<model-id>[:free]`:

```
openrouter/meta-llama/llama-3.1-8b-instruct:free   ← free tier
openrouter/anthropic/claude-3-5-haiku              ← paid
openrouter/google/gemma-2-9b-it:free               ← free tier
openrouter/auto                                    ← auto-select cheapest
```

### Rate limits for free-tier models

OpenRouter enforces the following limits on models ending in `:free`:

| Condition | Requests / minute | Requests / day |
|---|---|---|
| Any free model | **20** | — |
| Account with < 10 credits | — | **50** |
| Account with ≥ 10 credits | — | **1 000** |

> One PicoClaw task typically makes **5–15 API calls** (each tool use counts
> as one request).  Plan your daily topic count accordingly:
>
> | Credits | Daily budget | Safe topics/day (est. 10 calls each) |
> |---|---|---|
> | < 10 | 50 | ~5 |
> | ≥ 10 | 1 000 | ~100 |

### How rate limiting is enforced in the workflows

All three workflows (Blog Writer, Swarm, Pages) share the same approach:

1. **Auto-detection** — any model whose ID contains `:free` is treated as
   free tier.  You can also set `free_tier: true` explicitly in
   `workflow_dispatch` inputs.

2. **`max-parallel`** — the matrix `max-parallel` strategy key is set
   dynamically:
   - Free tier → `1` (serial execution; one API call chain at a time)
   - Paid tier → `6` (full parallel)

3. **`rate_limit_delay`** — passed to `action.yml`, which sleeps the
   specified number of seconds *before* the first API call in each agent run:
   - Free tier → `3` seconds (safe headroom below 20 req/min)
   - Paid tier → `0` (no delay)

These are conservative defaults.  If your model usage per task is lower,
you can safely increase `max-parallel` or reduce `rate_limit_delay`.

---

## `action.yml` — PicoClaw AI Agent action

The composite action at the repository root builds the PicoClaw binary from
source and runs `picoclaw agent --message "..."` in one-shot mode.

### Inputs

| Input | Required | Default | Description |
|---|---|---|---|
| `message` | **Yes** | — | Task or prompt for the agent |
| `model` | No | `openrouter/meta-llama/llama-3.1-8b-instruct:free` | Full model ID (`provider/model-id[:free]`) |
| `api_key` | **Yes** | — | API key for the provider (pass via `secrets.*`) |
| `api_base` | No | *(provider default)* | Override the API base URL |
| `workspace_dir` | No | `$RUNNER_TEMP/picoclaw-cfg/workspace` | Agent workspace (memory, skills) |
| `session` | No | `actions:default` | Session key for conversation history |
| `rate_limit_delay` | No | `auto` | Seconds to sleep before the API call. `auto` applies 3 s for `:free` models, 0 s otherwise. Set an integer to override. |

### Outputs

| Output | Description |
|---|---|
| `response` | The agent's text response (ANSI codes and logo prefix stripped) |

### Usage examples

**Minimal — write a blog post:**

```yaml
- uses: shelbeely/picoclaw@main
  with:
    message: 'Write a short blog post about Rust vs Go.'
    api_key: ${{ secrets.OPENROUTER_API_KEY }}
```

**With explicit model and rate-limit delay:**

```yaml
- uses: shelbeely/picoclaw@main
  id: agent
  with:
    message:          'Summarise the top 5 AI papers from this week.'
    model:            'openrouter/google/gemma-2-9b-it:free'
    api_key:          ${{ secrets.OPENROUTER_API_KEY }}
    rate_limit_delay: '3'
- run: echo "${{ steps.agent.outputs.response }}"
```

**From a different repository:**

```yaml
- uses: shelbeely/picoclaw@main
  with:
    message: 'Review this PR and suggest improvements.'
    model:   'openrouter/anthropic/claude-3-5-haiku'
    api_key: ${{ secrets.OPENROUTER_API_KEY }}
```

### How the action works (internals)

The shell scripts in `action.yml` and the workflows use **`jq`** for all JSON
work — no Python, no Node.  This is intentional: PicoClaw is a Go project and
`jq` is the standard JSON tool for shell scripts.  It is pre-installed on
every GitHub-hosted Ubuntu and macOS runner.

1. **Setup Go** — installs the Go toolchain declared in `go.mod`, with module
   cache enabled so subsequent runs skip the download step.
2. **Build picoclaw** — compiles `cmd/picoclaw` to `$RUNNER_TEMP/picoclaw`.
3. **Write config** — generates a minimal `config.json` using `jq`.
   The provider default API base is inferred from the model prefix:

   | Prefix | Default base URL |
   |---|---|
   | `anthropic` | `https://api.anthropic.com/v1` |
   | `openai` | `https://api.openai.com/v1` |
   | `openrouter` | `https://openrouter.ai/api/v1` |
   | `deepseek` | `https://api.deepseek.com/v1` |
   | `groq` | `https://api.groq.com/openai/v1` |
   | `gemini` | `https://generativelanguage.googleapis.com/v1beta` |
   | `mistral` | `https://api.mistral.ai/v1` |

4. **Rate-limit delay** — sleeps for `rate_limit_delay` seconds before the
   API call.  In `auto` mode, detects `:free` models and applies 3 s.
5. **Run agent** — executes `picoclaw agent --message "..." --session "..."`,
   capturing stdout, stripping ANSI codes and the `🦞` logo prefix, and
   writing the clean text to `$GITHUB_OUTPUT` as `response`.

---

## Workflow: Blog Writer

**File:** `.github/workflows/blog-writer.yml`  
**Trigger:** Weekly schedule (Monday 09:00 UTC) + `workflow_dispatch`

Runs one PicoClaw agent per topic in parallel using a GitHub Actions matrix,
then commits all generated posts in a single push.

### Inputs (workflow_dispatch)

| Input | Default | Description |
|---|---|---|
| `topics` | 3 built-in defaults | JSON array of topics, e.g. `["Rust","AI ethics"]` |
| `model` | `AI_MODEL` variable | Model in `provider/model-id` format |
| `free_tier` | `false` | Set `true` to enable rate-limiting for `:free` models |

### Jobs

| Job | Description |
|---|---|
| `set-matrix` | Resolves topics, model, `max_parallel`, and `delay` |
| `generate` | Matrix job — one picoclaw per topic, parallel up to `max_parallel` |
| `commit` | Downloads all post artifacts, copies to `docs/_posts/`, commits once |

### Rate limiting

```
free_tier: true  →  max-parallel: 1  + rate_limit_delay: 3
free_tier: false →  max-parallel: 6  + rate_limit_delay: 0
```

Auto-detection: if the model ID contains `:free`, `free_tier` is treated as
`true` regardless of the input value.

---

## Workflow: PicoClaw Swarm

**File:** `.github/workflows/swarm.yml`  
**Trigger:** `workflow_dispatch` only

A four-stage multi-agent pipeline where each stage is a separate GitHub
Actions job.  All research agents run in **parallel** via matrix strategy.

### Inputs (workflow_dispatch)

| Input | Default | Description |
|---|---|---|
| `topic` | *required* | Blog post topic for the entire swarm |
| `research_depth` | `3` | Number of parallel research agents (1–6) |
| `model` | `AI_MODEL` variable | Model in `provider/model-id` format |
| `free_tier` | `false` | Enable rate-limiting for `:free` models |

### Pipeline stages

```
orchestrate
  Prompt:  "Plan a blog post on <topic>. Output JSON with title,
            outline, and research_questions[]."
  Output:  title, outline, questions[] (forwarded to research matrix)

research (matrix — N jobs in parallel)
  Prompt:  "Research this question: <question>."
  Output:  300–500 word Markdown notes per question

write
  Input:   title + outline + all research notes (via step outputs)
  Prompt:  "Write a full Jekyll post using this context."
  Output:  Jekyll-formatted draft

edit
  Input:   draft (via step output)
  Prompt:  "Review and polish this draft."
  Output:  Final post → committed to docs/_posts/
```

### Job inter-communication

Data between jobs is passed in two ways:

- **Small values** (title, outline, questions, model) — via `jobs.<job>.outputs`
  and `needs.<job>.outputs`.
- **Large values** (research notes, draft) — via `actions/upload-artifact@v4`
  and `actions/download-artifact@v4`.

Context strings that need to be passed into a `uses:` step's `with.message`
field are captured as multiline step outputs first (GitHub Actions does not
support shell command substitution inside `with:` blocks).

### JSON tooling

All JSON work — config generation, plan parsing, field extraction — uses
**`jq`**.  No Python is used anywhere in the workflows.  `jq` is pre-installed
on every GitHub-hosted Ubuntu and macOS runner and is consistent with the
project's Go/no-Python stance.  In `swarm.yml`, the orchestrator's response is
parsed with a two-step strategy:

1. Try `jq -e '.'` on the raw response (ideal path — model returned clean JSON).
2. Fall back to `grep -Pzo '(?s)\{.*\}'` to extract the first `{…}` block,
   then re-parse with `jq` (handles prose the model accidentally prepended).

### Rate limiting in the swarm

The research matrix is the only stage where multiple API calls could overlap.
With `free_tier: true`:

- `max-parallel: 1` — research jobs run serially.
- `rate_limit_delay: 3` — each agent sleeps 3 s before its first call.

This guarantees at most ~1 API call every 3 s, well under the 20 req/min cap.

---

## Workflow: Research Swarm

**File:** `.github/workflows/research-swarm.yml`  
**Trigger:** `workflow_dispatch` only

Launches 10 parallel PicoClaw agents, each investigating a different
angle of a single research goal, then merges and summarises all findings
into one final Markdown report.

### Inputs (workflow_dispatch)

| Input | Default | Description |
|---|---|---|
| `goal` | *required* | Root research goal or question for the entire swarm |
| `model` | `OPENROUTER_MODEL` or `AI_MODEL` variable | Model in `provider/model-id` format |
| `free_tier` | `false` | Enable rate-limiting for `:free` models |
| `max_parallel` | auto (2 free / 10 paid) | Override max parallel research agents (1–10) |

### Shards (10 fixed)

Each shard is a unique research angle assigned to one agent.  Agents
are instructed to stay within their shard and avoid overlap.

| Shard | Scope |
|---|---|
| `official-docs` | Official documentation, specifications, reference material |
| `github-repos` | Open-source repositories, code examples, GitHub projects |
| `blog-posts` | Expert blog posts, articles, commentary |
| `tutorials` | Step-by-step guides, how-tos, walkthroughs |
| `discussions` | Forum threads, Reddit, Hacker News, Stack Overflow |
| `benchmarks` | Performance data, comparisons, metrics, evaluations |
| `critical-opinions` | Critiques, limitations, known issues, sceptical viewpoints |
| `implementation-examples` | Real-world implementations, production case studies |
| `related-tools` | Adjacent tools, libraries, alternatives, integrations |
| `recent-updates` | News, changelog entries, developments from the past 6 months |

### Pipeline stages

```
setup
  Resolves model, rate-limit parameters, and the 10-element shard array.
  Outputs: model, max_parallel, delay, jitter_max, shards

research (matrix — 10 shards in parallel)
  Each agent:
  1. Sleeps a random jitter (1–5 s free / 1 s paid) to spread API calls.
  2. Searches the web for sources within its shard only.
  3. Returns 3–6 findings in structured Markdown.
  4. Saves output as two artifacts: <shard>.md and <shard>.json.

reduce
  1. Downloads all research-shard-* artifacts.
  2. Merges all Markdown files and a JSON manifest.
  3. Calls a picoclaw reducer agent to:
     - Remove duplicate findings.
     - Rank findings by usefulness.
     - Generate the final Markdown report.
  4. Saves the report to docs/research/YYYY-MM-DD-<slug>.md.
  5. Commits and pushes the report, and uploads it as the research-report artifact.
```

### Jobs

| Job | Description |
|---|---|
| `setup` | Resolves model, rate limits, and the 10-element shard matrix |
| `research` | Matrix job — one picoclaw per shard, up to 10 in parallel |
| `reduce` | Downloads all shard artifacts, merges, deduplicates, generates report |

### Rate limiting

```
free_tier: true  →  max-parallel: 2  +  rate_limit_delay: 3  +  jitter: 1–5 s
free_tier: false →  max-parallel: 10 +  rate_limit_delay: 0  +  jitter: 1 s
```

Auto-detection: if the model ID contains `:free`, `free_tier` is treated as
`true` regardless of the input value.

With free tier and 10 shards serialised to 2 at a time, total runtime is
approximately 5 × (agent time + 3 s delay + jitter).  With paid models all 10
run in parallel.

### Outputs

- **`research-report` artifact** — the final Markdown report, downloadable from
  the workflow run's Artifacts section.
- **`docs/research/YYYY-MM-DD-<slug>.md`** — the same report committed to the
  repository.
- **`research-shard-<shard>` artifacts** — raw per-shard Markdown and JSON
  (useful for debugging or re-running the reducer).

### Quick start

1. Add `OPENROUTER_API_KEY` to your repo secrets (see [Secrets and variables](#secrets-and-variables)).
2. Go to **Actions → Research Swarm → Run workflow**.
3. Enter your research goal, e.g. `"Best practices for building AI agent systems in 2025"`.
4. Accept the defaults (free Llama model, auto rate-limiting).
5. When the workflow completes, download the `research-report` artifact or read
   `docs/research/` in the repository.

### Example dispatch

```yaml
# Trigger from another workflow or the Actions UI
- uses: actions/github-script@v7
  with:
    script: |
      await github.rest.actions.createWorkflowDispatch({
        owner: context.repo.owner,
        repo:  context.repo.repo,
        workflow_id: 'research-swarm.yml',
        ref:   'main',
        inputs: {
          goal:       'Comprehensive overview of vector databases in 2025',
          model:      'openrouter/meta-llama/llama-3.1-8b-instruct:free',
          free_tier:  'true'
        }
      })
```

---

## Workflow: Deploy to GitHub Pages

**File:** `.github/workflows/pages.yml`  
**Trigger:** Push to `main` that touches `docs/**`, or `workflow_dispatch`

Builds the Jekyll site in `docs/` and deploys to GitHub Pages using the
standard `actions/jekyll-build-pages` + `actions/deploy-pages` stack.

### Required repository setting

Go to **Settings → Pages** and set:
- Source: **GitHub Actions** (not "Deploy from a branch")

### Jobs

| Job | Description |
|---|---|
| `build` | Runs Jekyll, uploads the `_site/` output as a Pages artifact |
| `deploy` | Deploys the artifact to GitHub Pages |

---

## Jekyll blog scaffold

All blog content lives under `docs/`:

```
docs/
├── _config.yml        ← Jekyll site config (title, theme, plugins)
├── _posts/            ← Generated blog posts (YYYY-MM-DD-slug.md)
│   └── .gitkeep
├── about.md           ← About page (layout: page)
└── index.md           ← Home page (layout: home)
```

### `docs/_config.yml` highlights

```yaml
title:    "PicoClaw Blog"
theme:    minima
minima:
  skin:   dark
plugins:
  - jekyll-feed        # RSS feed at /feed.xml
  - jekyll-seo-tag     # <meta> SEO tags
  - jekyll-sitemap     # /sitemap.xml
```

> **Required:** set `url` and `baseurl` in `docs/_config.yml` to match your
> repository before publishing.
>
> | Deployment type | `url` | `baseurl` |
> |---|---|---|
> | `https://<org>.github.io/<repo>/` (typical fork) | `https://<org>.github.io` | `/<repo>` |
> | Root (`<org>.github.io`) or custom domain | your domain | `""` |
>
> Jekyll uses these for the RSS feed, sitemap, SEO tags, and all absolute links.
> Leaving them empty means internal links work but the feed and sitemap will
> have broken absolute URLs.

### Post front matter template

Every agent-generated post must include:

```yaml
---
layout: post
title: "Your Title"
date: 2025-01-20 09:00:00 +0000   # UTC datetime
categories: [category1, category2]
tags: [tag1, tag2, tag3]
excerpt: "One-sentence hook (max 160 chars, no Markdown)."
---
```

---

## Workspace persona

The agent's personality and role are defined in `workspace/`:

| File | Purpose |
|---|---|
| `IDENTITY.md` | Persona name, swarm roles (Orchestrator / Researcher / Writer / Editor) |
| `AGENTS.md` | Detailed instructions for each role including exact output formats |
| `SOUL.md` | Personality and values |
| `skills/blog/SKILL.md` | Jekyll post format, style guide, swarm pipeline summary |

The `workspace/` directory is passed to the action via `workspace_dir`.  The
agent loads `IDENTITY.md`, `AGENTS.md`, and skills at startup to prime its
system prompt.

---

## Customisation

### Change the default topics (Blog Writer)

Edit the `DEFAULT_TOPICS` line in `.github/workflows/blog-writer.yml`:

```yaml
DEFAULT_TOPICS='["Your topic 1","Your topic 2","Your topic 3"]'
```

Or pass a custom JSON array at dispatch time via the `topics` input.

### Change the model

Set the `AI_MODEL` repository variable to any supported model string, e.g.:

```
openrouter/google/gemma-2-9b-it:free
openrouter/anthropic/claude-3-5-haiku
anthropic/claude-3-5-haiku-20241022
openai/gpt-4o-mini
```

Remember to set `free_tier: true` (or let auto-detection handle it) when
switching to a `:free` model.

### Change the blog theme

Edit `docs/_config.yml`:

```yaml
theme: minima          # or jekyll-theme-cayman, jekyll-theme-slate, etc.
minima:
  skin: auto           # auto | dark | solarized | solarized-dark
```

### Adjust the post schedule

Edit the cron expression in `.github/workflows/blog-writer.yml`:

```yaml
schedule:
  - cron: '0 9 * * 1'   # Every Monday 09:00 UTC
  # - cron: '0 9 * * *'  # Daily
  # - cron: '0 9 1 * *'  # First of the month
```

### Use a different provider

Swap the model prefix and set the matching secret:

| Provider | Model format | Secret |
|---|---|---|
| OpenRouter | `openrouter/<provider>/<model>` | `OPENROUTER_API_KEY` |
| Anthropic | `anthropic/claude-*` | `AI_API_KEY` |
| OpenAI | `openai/gpt-*` | `AI_API_KEY` |
| DeepSeek | `deepseek/deepseek-*` | `AI_API_KEY` |
| Groq | `groq/llama-*` | `AI_API_KEY` |
| Mistral | `mistral/mistral-*` | `AI_API_KEY` |
