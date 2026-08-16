# Third-Party Licenses

**PCS Demo (pcs-demo) application source** is licensed under the
[Eclipse Public License 2.0](LICENSE) (EPL-2.0). See [NOTICE](NOTICE).

This document lists **bundled third-party open source software** included
when building or running the application. The licenses below apply **only**
to those components, not to pcs-demo application code.

The list summarizes the main dependencies pulled in via Go modules (`go.mod`). Transitive dependencies are recorded in `go.sum`.

## Direct dependency

| Component | Version | License (typical) |
|-----------|---------|-------------------|
| [Fyne](https://fyne.io/) (`fyne.io/fyne/v2`) | v2.7.4 | BSD-3-Clause |

Fyne copyright and full license text: see the
[Fyne LICENSE](https://github.com/fyne-io/fyne/blob/master/LICENSE) in the
upstream repository.

## Notable transitive dependencies

| Component | License (typical) |
|-----------|-------------------|
| `fyne.io/systray` | MIT |
| `github.com/go-gl/glfw` | zlib |
| `github.com/go-text/typesetting` | BSD-3-Clause |
| `github.com/fsnotify/fsnotify` | BSD-3-Clause |
| `github.com/BurntSushi/toml` | MIT |
| `github.com/godbus/dbus/v5` | BSD-2-Clause |
| `github.com/stretchr/testify` | MIT (tests / tooling only) |
| `golang.org/x/image`, `golang.org/x/net`, `golang.org/x/sys`, `golang.org/x/text` | BSD-3-Clause |
| `gopkg.in/yaml.v3` | MIT and Apache-2.0 |

## Go toolchain

The application is built with the [Go programming language](https://go.dev/).
Go and the Go standard library are distributed under the
[Go license](https://go.dev/LICENSE) (BSD-style).

---

If you distribute a binary of PCS Demo, ensure you comply with the license
terms of all included third-party components, including attribution and
notice requirements where applicable.
