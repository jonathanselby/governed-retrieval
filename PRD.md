# Governed Retrieval — PRD

**Version:** 0.1

## Problem

Sensitive applications need retrieval that respects incident assignment and compartment boundaries before any candidate passage can influence ranking or a later model prompt. They also need evidence for what was retrieved and a deterministic refusal when the authorised corpus cannot support the query.

## Product objective

Provide a deliberately small Go command-line application that can retrieve only material the caller is allowed to see, return source citations, and refuse unsupported queries.

> **Filter before ranking. Cite what supports the answer. Refuse when the corpus does not.**

The product test for V0.1 is:

> Can a small retrieval component make access control and evidential support part of retrieval itself rather than post-processing?

## Users

Engineers embedding governed retrieval into incident, legal, risk or other sensitive applications.

## In scope

- Markdown and DOCX corpus ingestion;
- caller incident-assignment and compartment filtering before ranking;
- deterministic keyword ranking with no model dependency;
- citations to source path and passage locator;
- explicit refusal when no authorised passage supports the query;
- bounded document paths under the configured corpus directory;
- stable JSON and CLI exit-code contracts;
- Go standard library only;

## Out of scope

- prose generation;
- vector database operation;
- embedding generation or model hosting;
- document OCR;
- PDF parsing in V0.1;
- authorisation policy administration;

These exclusions are deliberate. A future application need should widen this component only when the same reusable job has genuinely changed; otherwise the capability belongs in the consuming application or another component.

## Acceptance criteria

- an unprivileged caller never receives a privileged passage even when it is the strongest lexical match;
- a caller not assigned to an incident never receives incident-scoped material;
- Markdown and DOCX documents are indexed with stable source locators;
- a document path cannot escape the configured corpus directory;
- when no authorised passage overlaps the query, the result is a governed refusal;
- keyword retrieval works with no model or network;
- the CLI distinguishes governed refusal from usage and execution errors;
- the repository builds and tests for `linux/arm64`;

## Definition of done for V0.1

V0.1 is complete when all acceptance criteria are represented by automated tests or executable acceptance checks, CI runs `go vet ./...` and `go test -race ./...`, the repository cross-builds for `linux/arm64`, the production boundary and known limitations are explicit, `generated/EDD.md` is recovered from the implementation, and `generated/COMPARISON.md` shows whether the intended one-job boundary survived implementation.
