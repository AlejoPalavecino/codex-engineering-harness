---
name: chained-pr
description: "Trigger: PRs over 400 lines, stacked PRs, review slices. Split oversized changes into chained PRs that protect review focus."
license: Apache-2.0
metadata:
  author: gentleman-programming
  version: "1.0"
  upstream: "Gentle AI v4.4.0"
  chuli-runtime: "Codex"
  chuli-dependencies: "Git; publicar ramas o PR requiere autorización"
---

> Adaptación Chuli 1.0. La regla local es ~400 líneas authored (altas + bajas). La plantilla, los diagramas y las etiquetas específicas de Gentle AI sólo aplican si la política del destino los adopta. Hacé un plan de cortes; crear ramas remotas, publicar o aplicar etiquetas requiere autorización para ese destino.

## Activation Contract

Load this skill when a planned PR may exceed **400 changed lines**, an ODD feature's forecast or running authored changed-line count from work-unit commits exceeds about 400, or the user asks for chained/stacked PRs, review slices, or reviewer-load control.

## Hard Rules

- At about **400 authored changed lines**, try one cohesive split. Use a formal `size:exception` only when the destination policy actually requires it.
- The budget constrains how work is **sliced**, never the code itself. Never delete comments, blank lines, docs, or tests, and never compress or restyle code, to fit under the budget.
- Slicing is bounded: make **one** honest slicing pass. If no cohesive split brings every slice within budget, stop iterating, keep the best cohesive split, and report the final line count with an explanation of the overage.
- Keep each PR reviewable in about **≤60 minutes**.
- Use one deliverable work unit per PR; keep tests/docs with the unit they verify.
- State start, end, prior dependencies, follow-up work, and out-of-scope items in every chained PR.
- Where a chain diagram helps reviewers, include a dependency diagram marking the current PR with `📍`.
- In Feature Branch Chain, create a draft/no-merge tracker PR; child PR #1 targets the tracker branch, later children target the immediate parent branch.
- Treat polluted diffs as base bugs: retarget or rebase until only the current work unit appears.
- Do not mix chain strategies after the user chooses one.

## Decision Gates

| Condition | Action |
|---|---|
| PR ≤400 changed lines and focused | Keep single PR. |
| PR >400, each slice can land independently | Use Stacked PRs to main. |
| PR >400, feature must integrate before main | Use Feature Branch Chain with tracker. |
| Generated/vendor/migration diff cannot split cleanly | Report these lines separately and document the review boundary. |
| No cohesive split fits the budget after one slicing pass | Stop; deliver the best split, report the overage and why it cannot shrink further, and follow the destination policy for exceptions. |
| ODD provides `delivery_strategy` and `chain_strategy` | Follow them before the next work-unit commit or PR creation. |

## Execution Steps

1. Estimate changed lines and identify independent work units.
2. Ask for a chain strategy when none is cached and the budget is exceeded.
3. Plan the branches/PRs; create them only when authorized for that destination.
4. Add Chain Context to each PR without replacing the repo PR template.
5. Verify each PR independently: CI/tests/docs/manual checks, rollback scope, and clean diff.
6. Keep tracker PR draft/no-merge until all child PRs are reviewed and integrated.

## Output Contract

Return the chosen strategy, PR order, current PR boundary, dependency diagram, review budget (`additions + deletions`), verification plan, and any `size:exception` rationale.

## References

- [references/chaining-details.md](references/chaining-details.md) — strategy diagrams, PR body section, branch commands, and reviewer guidance.
