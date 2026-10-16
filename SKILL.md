# Governed Retrieval — Skill

## Use this when

Use `governed-retrieval` when the task is exactly this component's job:

> Provide a deliberately small Go command-line application that can retrieve only material the caller is allowed to see, return source citations, and refuse unsupported queries.

## Do not use this for

- prose generation;
- vector database operation;
- embedding generation or model hosting;
- document OCR;
- PDF parsing in V0.1;
- authorisation policy administration;

## Core rule

> **Filter before ranking. Cite what supports the answer. Refuse when the corpus does not.**

## Workflow

1. Read `PRD.md` first; do not infer requirements outside its boundary.
2. Read `EDD.md` for runtime, failure and security behaviour.
3. Run `go vet ./...` and `go test -race ./...` before changing code.
4. Run the Bafflewick acceptance case from `examples/bafflewick/README.md`.
5. Preserve the exit-code contract: `0` success, `1` governed failure, `2` usage, `3` execution/input failure.
6. Keep the core under `internal/` and the executable edge thin.
7. After code changes, regenerate `generated/EDD.md` from code evidence and update `generated/COMPARISON.md`.
8. Do not tag a release until `RELEASE.md` is satisfied.

## Constraint

Do not widen this repository merely because CRAB needs another feature. Add behaviour only when it belongs to this component's one reusable job.
