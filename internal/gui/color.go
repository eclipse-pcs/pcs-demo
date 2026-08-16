// Copyright (c) 2026 dido GmbH and others.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// https://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	colorSecretRed     = color.NRGBA{R: 0xC0, G: 0x28, B: 0x28, A: 0xFF}
	colorEncryptBlue   = color.NRGBA{R: 0x1A, G: 0x4D, B: 0xB8, A: 0xFF}
	colorParticleGreen = color.NRGBA{R: 0x2A, G: 0x8C, B: 0x3A, A: 0xFF}
)

type colorOverrideTheme struct {
	fyne.Theme
}

type secretEntryTheme struct {
	colorOverrideTheme
}

func (t *secretEntryTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameForeground {
		return colorSecretRed
	}
	return t.Theme.Color(name, variant)
}

type pipelineTheme struct {
	colorOverrideTheme
}

func (t *pipelineTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameError:
		return colorSecretRed
	case theme.ColorNamePrimary:
		return colorEncryptBlue
	case theme.ColorNameSuccess:
		return colorParticleGreen
	}
	return t.Theme.Color(name, variant)
}

func newSecretEntryTheme() fyne.Theme {
	return &secretEntryTheme{colorOverrideTheme{Theme: theme.Current()}}
}

func newPipelineTheme() fyne.Theme {
	return &pipelineTheme{colorOverrideTheme{Theme: theme.Current()}}
}

func styleSecretValue(l *widget.Label) {
	l.TextStyle = fyne.TextStyle{Monospace: true}
	l.Importance = widget.DangerImportance
}

func styleEncryptValue(l *widget.Label) {
	l.TextStyle = fyne.TextStyle{Monospace: true}
	l.Importance = widget.HighImportance
}

func styleParticleValue(l *widget.Label) {
	l.TextStyle = fyne.TextStyle{Monospace: true}
	l.Importance = widget.SuccessImportance
}

func wrapPipelineTheme(obj fyne.CanvasObject) fyne.CanvasObject {
	return container.NewThemeOverride(obj, newPipelineTheme())
}

func wrapSecretEntry(entry *widget.Entry) fyne.CanvasObject {
	entry.TextStyle = fyne.TextStyle{Monospace: true}
	return container.NewThemeOverride(entry, newSecretEntryTheme())
}
