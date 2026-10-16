# Governed Retrieval — Hand-Written vs Generated EDD

**Implementation evidence:** `e3c98124f88ef503da1514bbfc8fcce38dacfa5f`  
**Generated:** 2026-10-17

## Purpose

This comparison tests whether the V0.1 implementation of `governed-retrieval` communicates the architecture intended by the hand-written `EDD.md`. The generated EDD is derived from source, tests, configuration, examples and CI without using the PRD/hand-written EDD as runtime proof.

## Result

**Pass.** The material V0.1 runtime architecture is recoverable from code and the intended one-job boundary is preserved. Excluded capabilities have not silently appeared in the component.

## Comparison

| Design area | Hand-written intent | Recovered implementation | Assessment |
|---|---|---|---|
| one-job boundary | retrieve authorised passages with citations/refusal | one CLI, one retrieval package, no model/generation layer | MATCH |
| filter before ranking | inaccessible material must never enter ranking | `allowed` check occurs before document load and lexical scoring | MATCH |
| incident scope | incident-scoped documents require assignment | `allowed` checks `AssignedIncidents` | MATCH |
| compartments | compartmented documents require clearance | `allowed` checks caller compartments | MATCH |
| formats | V0.1 Markdown + DOCX | only `.md`/`.markdown` and `.docx` loaders exist | MATCH |
| source locators | results cite stable source/locator | Markdown line ranges and DOCX paragraph numbers returned | MATCH |
| unsupported query | refuse rather than guess | no hits returns governed refusal | MATCH |
| model/network independence | keyword fallback must work offline | core has no model or network dependency | MATCH |
| path boundary | documents cannot escape corpus root | absolute/parent paths rejected | MATCH |
| operational contract | vet, race, version, Bafflewick, ARM64 | all encoded in CI | MATCH |

## What code communicates well

1. **Access control is structurally before ranking, not a post-generation filter.**
2. **The only corpus loaders are visible and bounded; unsupported formats fail explicitly.**
3. **A refusal is a first-class result rather than an empty answer that a caller might reinterpret.**

## What implementation cannot establish by itself

- whether lexical retrieval quality is sufficient for a production corpus;
- future embedding/vector design and how it preserves the same pre-ranking access boundary;
- runtime memory/latency at the CRAB minimum-hardware envelope.

These unknowns are not treated as implementation defects unless the PRD/EDD says V0.1 owns them. The generated EDD marks uncertainty rather than inventing deployment or product intent.

## V0.1 architectural finding

The implementation communicates the intended security property directly: widening retrieval techniques later must not move authorisation behind ranking.
