// Copyright (c) 2026 dido GmbH and others.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// https://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const aboutDescription = `Interactive demonstration of a Particle Cryptography System (PCS) pipeline: one-time-pad encryption, particle splitting, parity, stealth transport (punch / permute), and integrity hashes.

For demonstration and education only — not production cryptography.`

func setupAboutMenu(a fyne.App, win fyne.Window) {
	meta := a.Metadata()
	body := fmt.Sprintf("Version %s\n\n%s\n\nThird-party open-source components: see THIRD_PARTY_LICENSES.md in the repository.", meta.Version, aboutDescription)
	showAbout := func() {
		label := widget.NewLabel(body)
		label.Wrapping = fyne.TextWrapWord
		scroll := container.NewScroll(label)
		scroll.SetMinSize(fyne.NewSize(480, 200))
		d := dialog.NewCustom("PCS Demo", "OK", scroll, win)
		d.Resize(fyne.NewSize(520, 280))
		d.Show()
	}
	win.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu(meta.Name,
			fyne.NewMenuItem("About", showAbout),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Quit", func() { a.Quit() }),
		),
	))
}

// Run starts the PCS Demo GUI application.
func Run() {
	a := app.New()

	view := bindings()
	model, headingW, caption, viewW := initialize(&view)

	secretW := widget.NewEntry()
	win := a.NewWindow("PCS Demo")
	grid, pipeline := build(&caption, &headingW, secretW, &view, &viewW)
	win.SetContent(grid)

	pipeline.setVisible(true, true)
	size := grid.MinSize()
	win.Resize(size)
	win.SetFixedSize(true)
	pipeline.setVisible(false, false)

	secretW.OnChanged = func(secret string) {
		model.secret = []byte(secret)
		hasPunchOuts := update(&model, view)
		pipeline.setVisible(len(model.secret) > 0, hasPunchOuts)
	}

	setupAboutMenu(a, win)
	win.ShowAndRun()
}
