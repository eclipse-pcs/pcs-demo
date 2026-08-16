# Contributing to pcs-demo

Thank you for your interest in the Particle Cryptography System (PCS) demo.

## License

Contributions to this repository are licensed under the
[Eclipse Public License 2.0](LICENSE) (EPL-2.0).

## Eclipse Contributor Agreement (required)

**You must sign the [Eclipse Contributor Agreement (ECA)](https://www.eclipse.org/contribute/cla)
before your contribution can be accepted.**

- Use the **same email address** on the ECA as in your git commit `Author` line.
- The ECA covers contributions to Eclipse Foundation projects under the project license.
- Committers may later be asked to sign additional Eclipse agreements during
  project provisioning.

Maintainers: see [ECA_CHECKLIST.md](ECA_CHECKLIST.md) for signing steps.

## Building and testing

From the repository root:

```bash
go build -o pcs-demo   ./cmd/pcs-demo
go build -o pcs-encode ./cmd/pcs-encode
go build -o pcs-decode ./cmd/pcs-decode
go test ./...
```

See [Readme.md](Readme.md) for hands-on encode/decode examples.

## Contribution workflow

1. Sign the ECA (if not already signed).
2. Fork or branch from `main`.
3. Make focused changes with tests where behavior changes.
4. Ensure `go test ./...` passes.
5. Open a merge/pull request with a clear description.

For the preparatory phase (private repo / community review), coordinate with
project leads before pushing large changes.

## Security

See [SECURITY.md](SECURITY.md) for vulnerability reporting.

## Eclipse onboarding

Formal IP log, git.eclipse.org migration, and committer provisioning are
tracked by the Eclipse Management Office. See [EMO_COORDINATION.md](EMO_COORDINATION.md).
