# pcs-demo

Interactive Particle Cloud Security (PCS) demo: a Fyne GUI app (`pcs-demo`) and two
command-line tools (`pcs-encode`, `pcs-decode`) that read and write footer-v1 particle
files on local disk.

PCS math and the footer codec live in the shared module
[github.com/eclipse-pcs/pcs](https://github.com/eclipse-pcs/pcs). This repo adds
file layout, GUI, and CLI wiring.

## Project layout

```
pcs-demo/
  cmd/
    pcs-demo/     GUI application
    pcs-encode/   encode CLI
    pcs-decode/   decode CLI
  internal/
    store/        encode/decode and particle paths on disk
    gui/          Fyne GUI implementation
  go.mod
  FyneApp.toml
```

## Build

Build all three executables from the project root:

```bash
go build -o pcs-demo   ./cmd/pcs-demo
go build -o pcs-encode ./cmd/pcs-encode
go build -o pcs-decode ./cmd/pcs-decode
```

## Test

```bash
go test ./test/...
```

Integration tests write particle files under `test/_data/` (gitignored). When running
`pcs-encode` manually in that folder, use `-y` to skip delete prompts on re-runs.

## Particle file format (footer v1)

Each particle file has this layout:

```
[ particle payload ][ 64-byte footer ]
```

The footer carries logical length, a SHA-256 fingerprint shard, own-payload CRC,
partner cross-CRC, WriteID, and particle kind. Full spec:
[github.com/eclipse-pcs/pcs/design/footer-v1.md](https://github.com/eclipse-pcs/pcs/blob/main/design/footer-v1.md).

| Pair   | File A (storage) | File B (storage) |
|--------|------------------|------------------|
| Cypher | `storageA/*.ec`  | `storageB/*.oc`  |
| Noise  | `storageA/*.on`  | `storageB/*.en`  |
| Parity | `storageC/*.cp`  | `storageC/*.np`  |

Odd-length secrets no longer use a separate `.cp_` / `.np_` suffix; length is stored
in the footer. Empty input files are allowed (minimum file size is 64 bytes: 0-byte
payload + footer).

## pcs-encode

```bash
./pcs-encode Hello.txt
```

Reads the input file, encodes it with PCS, and writes six particle files into
`storageA/`, `storageB/`, and `storageC/` in the current working directory. Use `-y`
to skip the prompt to delete existing particle files before writing.

## pcs-decode

```bash
./pcs-decode Hello.txt
```

Reads particle files from `storageA/`, `storageB/`, and `storageC/`, reconstructs the
secret, and validates it. On success it writes `Hello_reconstructed.txt` for
`Hello.txt` in the current working directory.

### Validation order

1. **WriteID** — all six footers must share the same WriteID.
2. **Cross-CRC** — each particle's cross-CRC must match CRC32-IEEE of its partner's
   payload (e.g. `storageA/Hello.txt.ec` ↔ `storageB/Hello.txt.oc`). Mismatch
   triggers parity recovery when possible; `pcs-decode` logs which file was rebuilt.
3. **Own-payload CRC** — each payload must match its footer's PayloadCRC.
4. **SHA-256 fingerprint** — fingerprint shards in the footers must reconstruct to
   SHA-256(secret). Failure prompts **[a] Abort** or **[b] Save** as
   `*_reconstructed_but_corrupt*`.

### Hands-on examples

These walkthroughs use `Hello.txt` as the secret. Run all commands from the directory
that contains `Hello.txt` and the `storageA` / `storageB` / `storageC` folders after
encode. Use `-y` on encode and decode to skip overwrite prompts.

```bash
echo -n "Hello Freiburg" > Hello.txt
../../pcs-encode -y Hello.txt
```

After each decode, compare the result with the original:

```bash
diff Hello.txt Hello_reconstructed.txt
```

Each example changes or removes particle files on disk. **Example 1** uses the initial
encode above. **Before examples 2–4**, run `./pcs-encode -y Hello.txt` again so all
six particle files and both storage folders exist.

**1. Missing storage folder (`storageB`)**

```bash
rm -rf storageB
../../pcs-decode -y Hello.txt
```

Expected stderr (among other lines): missing `storageB`, missing
`storageB/Hello.txt.oc` and `storageB/Hello.txt.en`, then
`reconstructing missing particles using parity particles`, then
`successfully decoded Hello_reconstructed.txt`.

**2. Missing one particle file**

```bash
../../pcs-encode -y Hello.txt
rm storageA/Hello.txt.ec
../../pcs-decode -y Hello.txt
```

Expected: `missing particle: evenCypher (storageA/Hello.txt.ec)`, parity
reconstruction message, successful decode.

**3. Corrupt payload (first byte of `Hello.txt.oc`)**

```bash
../../pcs-encode -y Hello.txt
f=storageB/Hello.txt.oc
x=$(( $(od -An -tu1 -N1 "$f") ^ 255 ))
printf "\\$(printf '%03o' "$x")" | dd of="$f" bs=1 count=1 conv=notrunc 2>/dev/null
../../pcs-decode -y Hello.txt
```

Expected: `cross-CRC mismatch: reconstructed storageB/Hello.txt.oc using parity particles`,
then success.

**4. Corrupt cross-CRC in footer (last byte of `Hello.txt.oc`)**

```bash
../../pcs-encode -y Hello.txt
f=storageB/Hello.txt.oc
n=$(wc -c < "$f" | tr -d ' ')
x=$(( $(od -An -tu1 -j $((n-1)) -N1 "$f") ^ 255 ))
printf "\\$(printf '%03o' "$x")" | dd of="$f" bs=1 seek=$((n-1)) conv=notrunc 2>/dev/null
../../pcs-decode -y Hello.txt
```

Expected: cross-CRC recovery; the log may name the partner file
(`storageA/Hello.txt.ec`). Decode still succeeds and `diff` is quiet.

## License

pcs-demo is licensed under the [Eclipse Public License 2.0](LICENSE) (EPL-2.0).

Third-party components are listed in [THIRD_PARTY_LICENSES.md](THIRD_PARTY_LICENSES.md).
Contributions require the [Eclipse Contributor Agreement](https://www.eclipse.org/contribute/cla);
see [CONTRIBUTING.md](CONTRIBUTING.md).
