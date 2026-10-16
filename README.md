# Governed Retrieval

[![test](https://github.com/jonathanselby/governed-retrieval/actions/workflows/test.yml/badge.svg)](https://github.com/jonathanselby/governed-retrieval/actions/workflows/test.yml)

**Current version:** 0.1.0

**Filter before ranking. Cite what supports the answer. Refuse when the corpus does not.**

`governed-retrieval` is a deliberately small Go command-line component for retrieving only material a caller is allowed to see, returning source citations and refusing unsupported queries. It is independently usable outside the Cyber Response Action Board (CRAB).

## Why it exists

> Can a small retrieval component make access control and evidential support part of retrieval itself rather than post-processing?

V0.1 answers that question with code, deterministic tests and a Bafflewick acceptance case, not architecture prose alone.

## What this does

- Markdown and DOCX corpus ingestion;
- caller incident-assignment and compartment filtering before ranking;
- deterministic keyword ranking with no model dependency;
- citations to source path and passage locator;
- explicit refusal when no authorised passage supports the query;
- bounded document paths under the configured corpus directory;
- stable JSON and CLI exit-code contracts;
- Go standard library only;

## What this does not do

- prose generation;
- vector database operation;
- embedding generation or model hosting;
- document OCR;
- PDF parsing in V0.1;
- authorisation policy administration;

## Install

Requires Go 1.26+.

```bash
go install github.com/jonathanselby/governed-retrieval/cmd/gretrieve@main
```

Check the executable:

```bash
gretrieve -version
```

Expected version: `0.1.0`.

## Run it

The checked-in Bafflewick corpus is a complete offline acceptance case:

```bash
gretrieve \
  -corpus examples/bafflewick/corpus.json \
  -caller examples/bafflewick/caller.json \
  -query "insurer hotline"
```

The result cites `runbook.md`. A query supported only by the privileged legal passage returns a governed refusal (exit `1`) rather than leaking the passage.

## Bafflewick acceptance

The full acceptance notes and fixtures are under [`examples/bafflewick/`](examples/bafflewick/). Continuous integration runs the example plus the material boundary tests in addition to the race-tested package suite.

## Stable exit codes

| Code | Meaning |
|---:|---|
| `0` | success |
| `1` | governed or deterministic contract failure |
| `2` | invalid CLI usage |
| `3` | malformed input, file I/O or execution failure |

## Design cross-check

This repository keeps intended design and implementation evidence separate. [`EDD.md`](EDD.md) states the intended component boundary. [`generated/EDD.md`](generated/EDD.md) is recovered from code, tests, configuration and operational files without using the hand-written PRD/EDD as runtime evidence. [`generated/COMPARISON.md`](generated/COMPARISON.md) compares the two and records mismatches rather than silently reconciling them.

## Engineering artefacts

- `PRD.md` — product boundary and testable acceptance criteria;
- `EDD.md` — hand-written design and production boundary;
- `architecture.txt` — plain-text runtime view;
- `internal/` — the one reusable Go contract;
- `cmd/` — thin executable edge;
- `examples/bafflewick/` — fictional acceptance case;
- `.github/workflows/test.yml` — vet, race tests, version contract, Bafflewick acceptance and `linux/arm64` cross-build;
- `generated/EDD.md` — code-derived design evidence pinned to a source commit;
- `generated/COMPARISON.md` — intended design versus recovered implementation;
- `SKILL.md` — compact adoption instructions;
- `RELEASE.md`, `CHANGELOG.md` and `VERSION` — release discipline.

## Repository shape

```text
governed-retrieval/
├── .github/workflows/test.yml
├── PRD.md
├── EDD.md
├── README.md
├── SKILL.md
├── RELEASE.md
├── VERSION
├── CHANGELOG.md
├── LICENSE
├── architecture.txt
├── go.mod
├── cmd/gretrieve/main.go
├── internal/retrieval/
├── examples/bafflewick/
└── generated/
    ├── EDD.md
    └── COMPARISON.md
```

## Design rule

> **Filter before ranking. Cite what supports the answer. Refuse when the corpus does not.**

## Licence

MIT. See [`LICENSE`](LICENSE).
