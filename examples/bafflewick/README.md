# Bafflewick example

```bash
go run ./cmd/gretrieve -corpus examples/bafflewick/corpus.json -caller examples/bafflewick/caller.json -query "insurer hotline"
```

The result cites `runbook.md`. Querying only for privileged legal material with this caller returns a governed refusal (exit 1), not the privileged passage.
