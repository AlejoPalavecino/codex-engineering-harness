---
name: cognitive-doc-design
description: "Design docs that reduce cognitive load. Trigger: writing guides, READMEs, RFCs, onboarding, architecture, or review-facing docs."
license: Apache-2.0
metadata:
  author: gentleman-programming
  version: "1.0"
  upstream: "Gentle AI v4.4.0"
  chuli-runtime: "Codex"
  chuli-dependencies: "Documento y audiencia"
---

> Adaptación Chuli 1.0. Seguí la política y los permisos del repositorio de destino.

## When to Use

Load this skill when creating or editing documentation that people need to understand quickly, retain, or use during review.

Use it especially for:

- PR descriptions and review notes.
- Contributor or maintainer guides.
- Architecture, workflow, or onboarding docs.
- Any doc that currently feels long, dense, or hard to scan.

## Critical Patterns

| Pattern | Rule |
|---------|------|
| Lead with the answer | Put the decision, action, or outcome first. Context comes after. |
| Progressive disclosure | Start with the happy path, then add details, edge cases, and references. |
| Chunking | Group related information into small sections. Keep flat lists short. |
| Signposting | Use headings, labels, callouts, and summaries so readers know where they are. |
| Recognition over recall | Prefer tables, checklists, examples, and templates over prose that must be remembered. |
| Review empathy | Design docs so reviewers can verify intent without reconstructing the whole story. |

## Diagrams in Chuli

Use [the diagram guide](references/diagrams.md) when relationships, decisions or handoffs are clearer visually. Load the bundled `diagram-design` skill only for a diagram that earns its place. Preserve the project's brand, source evidence, accessible text equivalent and editable source; a table or paragraph is sufficient when it explains the same thing. Do not invent components or render every document as a diagram.

## Documentation Shape

Use this default structure unless the repo already provides a stronger template:

```markdown
# <Outcome-oriented title>

<One paragraph: what changed, who it helps, and why it matters.>

## Quick path

1. <First action>
2. <Second action>
3. <Verification or expected result>

## Details

| Topic | Decision |
|-------|----------|
| <area> | <concise explanation> |

## Checklist

- [ ] <Reader can confirm this>
- [ ] <Reader can confirm that>

## Next step

<Link or action that continues the workflow.>
```

## PR and Review Docs

When documenting a PR, reduce reviewer burnout by making the review path explicit:

- State what to review first.
- State what is intentionally out of scope.
- Link the previous and next PR when work is chained.
- Keep each section focused on one decision or unit of work.
- Use checklists for acceptance criteria and verification.

## Commands

```bash
# Check markdown files changed in the current branch
git diff --name-only -- '*.md'

# Inspect PR changed-line count for cognitive load
gh pr view <PR_NUMBER> --json additions,deletions,changedFiles
```
