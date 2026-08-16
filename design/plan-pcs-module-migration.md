# Plan: migrate pcs-demo to the shared `github.com/eclipse-pcs/pcs` module

Read first (canonical, do not paraphrase into this repo):

- `~/go/src/pcs/design/footer-v1.md` — the particle footer v1 format spec
- `~/go/src/pcs/design/plan-module.md` — the module's API surface

Prerequisite: the module in `~/go/src/pcs` exists and its tests pass.
pcs-demo is the **first consumer** to migrate (order: module → demo →
service → files-gateway → s3-gateway).

## Scope

- `cmd/pcs-encode` and `cmd/pcs-decode` fully working on the new format is
  the acceptance bar for this migration.
- The Fyne GUI (`cmd/pcs-demo`, `internal/gui`) only needs to **compile**;
  making it fully functional on the new format is a separate follow-up at the
  end of the whole migration series.
- No backward compatibility: old particle files (12-byte tail, `.e`/`.o`,
  `.cp_`/`.np_`) are abandoned, not read.

## Work items

1. **Wire the module.** Create `~/go/src/go.work` (if absent) with
   `use ./pcs` and `use ./pcs-demo` so
   `github.com/eclipse-pcs/pcs` resolves locally (GitHub org not yet
   populated). Add the requirement to `go.mod`.

2. **Replace `internal/pcs` math with module imports.** Delete from
   `internal/pcs`: `pcs.go` (Split/Encrypt/Parity/RandomNoise), `encode.go`,
   `decode.go`, `hash.go`, `crc.go`, `decode_crc.go`, `particles.go`,
   `inventory.go` — all superseded by module packages `pcs` and `footer`.
   Exception: keep `PrintableNoise` and `ParityLegacy` (used by
   `internal/gui/update.go`); move them into a small `internal/gui`-adjacent
   file (e.g. `internal/gui/noise.go`) so `internal/pcs` can be removed
   entirely.

3. **Rebuild the on-disk layer** (successor of `internal/pcs/storage.go`,
   e.g. new `internal/store` package). This stays in pcs-demo — the module
   deliberately excludes file I/O. It must:
   - use the unified suffixes and placement from the spec:
     `storageA/<base>.ec`, `storageA/<base>.on`, `storageB/<base>.oc`,
     `storageB/<base>.en`, `storageC/<base>.cp`, `storageC/<base>.np`
     (note: `.e` → `.ec`, `.o` → `.oc`; **no** `.cp_`/`.np_` variants)
   - write files as `payload + footer.Marshal()`; footers built with one
     shared WriteID (`footer.NewWriteID()`), SHA-256 fingerprint shards from
     `pcs.EncodeFingerprint`, own/cross CRCs per spec (partner pairs
     ec↔oc, on↔en, cp↔np)
   - scan inventory by file presence only (no underscore probing); derive the
     logical length from payload sizes per the spec's "Size and parity
     derivation" section, falling back to a parsed footer in the
     both-odd-missing case
   - on read, verify footers via the module (`footer.Parse`, CRC verdicts,
     fingerprint, `VerifyWriteIDs`)

4. **Update `cmd/pcs-encode`.** Same UX (overwrite prompt, `-y`), now:
   read file → `pcs.Encode` (buffered path is fine for the demo) →
   fingerprint + footers → write six files via the new store layer.

5. **Update `cmd/pcs-decode`.** Same UX, now: inventory scan → load
   particles → strip/parse footers → `pcs.DecodeWithRecovery` with the
   length from footers/sizes → verify fingerprint + WriteIDs + CRCs →
   write `<base>_reconstructed.<ext>` (keep the
   `_reconstructed_but_corrupt` path for verification failures, now driven
   by module verdicts).

6. **GUI compile fix only.** Adjust `internal/gui` imports so the project
   builds (`go build ./...`). If GUI code depended on deleted internals
   beyond `PrintableNoise`/`ParityLegacy`, stub or comment the affected demo
   flows with `// TODO(gui-migration)` markers rather than porting them now.

7. **Tests.**
   - Delete the old `internal/pcs` unit tests (they moved into the module).
   - Keep/adapt `test/encode_test.go`, `test/decode_test.go`,
     `test/cli_test.go` as **CLI integration tests** against the new format:
     encode a fixture, assert the six expected filenames and sizes
     (`payload+64`), delete particles per recovery scenario, decode, compare
     bytes. Include: even length, odd length, empty file, missing `.oc`,
     missing `.oc`+`.on` (ambiguous case), corrupted payload byte (must fail
     with integrity error).
   - Keep `internal/pcs/storage_test.go` equivalents for the new store
     package.

## Acceptance

- `go build ./...` and `go test ./...` green
- Manual round-trip: `pcs-encode <file>` then `pcs-decode <file>` reproduces
  the file bit-identically; works for odd and even lengths and after deleting
  any single recoverable particle
- Particle files on disk match the spec byte-for-byte (`xxd` spot-check:
  magic `50 43 53 00`, version `01 00` at offset 4)

## Out of scope

- Full GUI functionality (follow-up after all four projects migrate)
- Reading legacy particle files
