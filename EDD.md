# Governed Retrieval — Engineering Design Document

**Version:** 0.1

## 1. Purpose

Governed Retrieval is a small Go command-line component implementing one reusable job.

The core design rule is:

> **Filter before ranking. Cite what supports the answer. Refuse when the corpus does not.**

The V0.1 engineering question is: Can access control and evidential support be enforced inside retrieval itself rather than after ranking?

## 2. Non-goals

The component:

- does not generate prose answers;
- does not run a vector database or embedding model;
- does not perform OCR or PDF parsing in V0.1;
- does not administer authorisation policy;

These exclusions are part of the design. V0.1 is not a smaller version of a future platform; it is intended to remain independently understandable and reusable.

## 3. Component model

```text
CALLER / INPUT
      |
      v
+-------------------------+
| cmd/gretrieve           |
+------------+------------+
             | typed contract
             v
+-------------------------+
| internal/retrieval      |
+------------+------------+
             |
             v
     STRUCTURED RESULT
```

| Component | Responsibility |
|---|---|
| CLI (`cmd/gretrieve`) | parse configuration/input, invoke the core, print structured output and map exit codes |
| Core (`internal/retrieval`) | own the component's one reusable contract |
| `examples/bafflewick` | executable fictional inputs or a test-backed acceptance example |

## 4. Core model

The core contract is represented by real Go types rather than a second schema invented only for documentation.

```go
type Document struct {
    Path        string `json:"path"`
    Incident    string `json:"incident,omitempty"`
    Compartment string `json:"compartment,omitempty"`
}

type Caller struct {
    ID                string   `json:"id"`
    AssignedIncidents []string `json:"assigned_incidents,omitempty"`
    Compartments      []string `json:"compartments,omitempty"`
}

type Result struct {
    Refused   bool       `json:"refused"`
    Reason    string     `json:"reason,omitempty"`
    Citations []Citation `json:"citations,omitempty"`
}
```

Fields not shown above remain defined in the package source; `generated/EDD.md` is the implementation-derived evidence record.

## 5. Core behaviour

1. validate query and caller.
2. discard documents outside the caller’s incident assignments or compartments before loading/ranking them.
3. load only relative Markdown or DOCX paths under the corpus directory.
4. tokenise authorised passages and rank by deterministic lexical overlap.
5. return stable source locators for matching passages.
6. return a governed refusal when no authorised passage supports the query.

## 6. Contract and abstraction

The core package is the stable in-process boundary. The CLI is deliberately thin and contains no duplicate policy logic. Where the component crosses an external edge, that edge is isolated behind the smallest practical function or interface so unit tests can exercise behaviour without a live service.

The component does not import any other Jonno's Garage repository. Composition happens in the consuming application; reuse does not become runtime coupling.

## 7. Credential boundary

No credentials are accepted. Caller identity and access metadata are input data; authentication belongs to the consuming application.

## 8. Runtime flow

```text
INPUT
  |
  v
PARSE + VALIDATE
  |
  +-- malformed -----------------> EXECUTION ERROR
  |
  v
CORE CONTRACT
  |
  +-- governed refusal/failure --> STRUCTURED RESULT + EXIT 1
  |
  v
SUCCESS -------------------------> STRUCTURED RESULT + EXIT 0
```

There is no silent failover. A component may deliberately degrade only where that behaviour is part of its explicit contract.

## 9. Configuration and CLI contract

Primary invocation:

```text
`gretrieve -corpus <manifest.json> -caller <caller.json> -query <text> [-limit N]`
```

Every executable also supports:

```text
-version
```

Exit codes:

```text
0  success
1  governed / deterministic contract failure
2  invalid CLI usage
3  malformed input, file I/O or execution failure
```

This is a one-shot command-line component: governed contract refusal is visible to the caller, while malformed input, file or execution failure is kept distinct from that business result.

## 10. Logging and trace

Normal results are structured JSON on standard output. Diagnostics are written to standard error. The component emits no hidden telemetry.

Security-relevant or external operations expose outcome information in their explicit result/audit contract rather than relying on unstructured logs.

## 11. Failure handling

| Failure | Behaviour |
|---|---|
| empty/malformed request | reject before retrieval; usage/input error |
| path outside corpus directory | execution error; file is not read |
| unsupported document format | execution error |
| no authorised supporting passage | structured refusal, exit 1 |

Failure is visible. V0.1 does not implement retries unless a retry is part of the component's one defined job.

## 12. State and persistence

No persistent runtime state. The corpus manifest and documents are read-only inputs.

## 13. Test strategy

Package tests cover the reusable retrieval contract; CI then exercises the documented CLI acceptance case and ARM64 build.

| ID | Acceptance behaviour | Executable evidence |
|---|---|---|
| T-01 | privileged material is filtered before ranking | `TestPrivilegeFilteredBeforeRanking` |
| T-02 | incident assignment is enforced | `TestUnassignedRefused` |
| T-03 | DOCX produces stable paragraph locators | `TestDOCX` |
| T-04 | corpus paths cannot escape the configured root | `TestPathEscapeRejected` |
| T-05 | unsupported authorised queries refuse without citations | `TestUnsupportedQueryRefuses` |
| T-06 | Markdown keyword retrieval works without a model or network | Bafflewick CLI acceptance in CI |
| T-07 | version and `linux/arm64` build contracts hold | CI `Version contract` and `Linux arm64 build` |

The key security test is T-01: the privileged candidate is removed **before** ranking, so an inaccessible passage cannot become prompt or answer material merely because it is the strongest lexical match.

CI also runs `go vet ./...` and `go test -race ./...`. Tests do not claim OCR, PDF ingestion, vector retrieval or live model integration.

## 14. Security boundary

The decisive boundary is retrieval-time filtering: unauthorised documents are excluded before their text can contribute to ranking or appear in a result. The component does not authenticate callers or decide policy membership; it consumes those decisions as input.

Host hardening, filesystem permissions, transport security, secret storage and deployment identity remain deployment responsibilities unless this component explicitly owns them.

## 15. Operational behaviour

- one small Go executable;
- Go standard library only in V0.1;
- deterministic core where the job permits it;
- structured machine-readable output;
- stable exit codes;
- no background daemon unless the protocol itself requires a foreground loop;
- no hidden database, cache or second source of truth.

The repository's minimum-hardware proof is a `linux/arm64` build. Raspberry Pi runtime acceptance belongs to the end-to-end application unless a component-specific resource claim requires separate measurement.

## 16. Known limitations

- keyword overlap is intentionally simple and does not provide semantic retrieval;
- DOCX support reads paragraph text from `word/document.xml`; it does not interpret complex layout, headers, tracked changes or embedded objects;
- Markdown locators are line ranges and DOCX locators are paragraph numbers, not immutable document anchors;

These are explicit boundaries, not implied future features.

## 17. V0.1 design conclusion

The core stays small because access filtering, parsing and deterministic ranking are sufficient to prove the security boundary without importing a model, vector store or CRAB.
