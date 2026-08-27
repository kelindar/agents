---
name: to-domain
description: Create or update specs/DOMAIN.yaml from business-domain understanding. Use when domain modeling or grilling is intended to produce that file. Do not use for conceptual DDD explanation, machine-learning models, or schema-only design.
---

## Process

Capture agreed business language in `specs/DOMAIN.yaml`.

Inspect only the domain-bearing sources needed for the requested scope. Resolve evidence in this order: user-confirmed decisions and the existing domain file, accepted specs and domain documentation, tests and public behavior, then implementation names. Put conflicts and unsupported interpretations under `questions` instead of choosing one.

A bounded context is a boundary where terms or business rules have a distinct meaning. Do not turn packages, services, or technical layers into contexts. Put the same term in separate contexts when its business meaning differs.

Include only terms with domain-specific meaning that help state rules or distinguish concepts. Exclude implementation vocabulary such as handlers, repositories, DTOs, tables, transports, and workers.

If no context, term, or rule is supported, do not create an empty file. Ask for the smallest missing piece of domain information.

Create or update `specs/DOMAIN.yaml` as a YAML list of bounded contexts:

```yaml
- context: Ordering
  terms:
    Cart: Products selected before ordering.
    Order: A confirmed request to purchase products.
  rules:
    - An Order cannot return to a Cart.
  questions:
    - Does payment happen before or after an Order is created?
```

Each context may contain `terms`, `rules`, and `questions`. Omit empty sections.

Make focused updates. Preserve comments, ordering, context names, and uncontradicted entries. Never silently delete, rename, or merge existing language. When the user resolves a question, remove it and add only the confirmed term or rule.

When used during grilling, read `specs/DOMAIN.yaml` before asking questions. Write only decisions settled during the interview, and only after the user confirms the summarized shared understanding. Keep unresolved decisions under `questions`.

Parse the written YAML. The root must be a list. Every item must have one unique string `context`. When present, `terms` must map unique strings to strings, and `rules` and `questions` must be lists of strings. Every domain-specific term used in a definition or rule must be defined in that context or have a question asking for its meaning. Quote scalars when YAML could reinterpret them.

Only modify `specs/DOMAIN.yaml` unless the user asks for other changes.
