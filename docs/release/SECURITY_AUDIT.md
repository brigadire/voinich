# Security audit

**Date:** 2026-08-23
**Publishable scope:** clean-root public candidate, not the private source
repository or its historical commits.

## Task72 pattern scan

Task72 pattern-scanned the then-current tracked tree and 91 reachable private
commits for common API keys, token formats, private-key headers, and
assignment-style passwords. It found no credential-shaped material. That scan
was heuristic and is retained as a separate earlier result.

## Task72b dedicated gitleaks scan

Tool: `gitleaks 8.30.1` (official Linux x64 release archive SHA-256
`551f6fc83ea457d62a0d98237cbad105af8d557003051f41f3e7ca7b3f2470eb`).

Commands, run from the clean-root public candidate:

```bash
gitleaks dir --no-banner --redact --exit-code 1 .
gitleaks git --no-banner --redact --exit-code 1 .
```

The initial dedicated scan reported four `generic-api-key` matches. Each was
reviewed and retained in the audit rather than hidden:

- two lines in `docs/audit/REPOSITORY_INVENTORY.tsv`: scientific package
  names beginning with `token...`;
- one metric selector in
  `experiments/mechanism-space-v1/VOYNICH_TARGET_MANIFEST.tsv` containing an
  uppercase token-unit label and percentage threshold followed by a numeric
  scientific value;
- the Go source that constructs that same metric selector in
  `research/phase1/mechanism-space-analyze/targets.go`.

None is a credential, endpoint, authentication assignment, or secret value.
Their exact directory and root-commit fingerprints are narrowly listed in
`.gitleaksignore`; no file, path, or detector rule is globally allowlisted.

After review, both commands were repeated with fail-on-finding against the
final replacement candidate and reported `no leaks found`. The full tree pass
scanned approximately 3.43 GB and the full sanitized-history pass approximately
3.55 GB. Later audit-only commit deltas were separately scanned with the same
fail-on-finding setting, so all reachable history is covered. No credential
needs rotation or revocation.

## History and privacy boundary

The public candidate starts from a reviewed clean root. Its reachable history
never contains the removed Astafiev files, vendored IVTT files, or `tasks/`.
The original private repository at starting commit
`19badafdcccd40120e74cc1fa349ac331df200e7` was not rewritten. Its dangling
objects and private commits are not part of the publishable scope.

No repository visibility, remote, credential, or external account was changed.
