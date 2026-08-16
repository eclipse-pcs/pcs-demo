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
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const headingTopGap float32 = 12

func verticalGap(h float32) fyne.CanvasObject {
	r := canvas.NewRectangle(color.Transparent)
	r.SetMinSize(fyne.NewSize(0, h))
	return r
}

type pipelineUI struct {
	leftPipeline    fyne.CanvasObject
	rightPipeline   fyne.CanvasObject
	secretCharLabel *widget.Label
	secretCharValue *widget.Label
}

func (p *pipelineUI) setVisible(hasSecret, hasPunchOuts bool) {
	if hasSecret {
		p.leftPipeline.Show()
		p.secretCharLabel.Show()
		p.secretCharValue.Show()
	} else {
		p.leftPipeline.Hide()
		p.secretCharLabel.Hide()
		p.secretCharValue.Hide()
	}
	if hasSecret && hasPunchOuts {
		p.rightPipeline.Show()
	} else {
		p.rightPipeline.Hide()
	}
}

func build(caption *captionData, headingW *headingWidget, secretW *widget.Entry, view *viewData, viewW *viewWidget) (*fyne.Container, *pipelineUI) {
	secretCharLabel := widget.NewLabel(caption.char)
	secretForm := container.New(layout.NewFormLayout(),
		widget.NewLabel(caption.secret), wrapSecretEntry(secretW),
		secretCharLabel, viewW.secret,
	)

	secretHashForm := container.New(layout.NewFormLayout(),
		widget.NewLabel(caption.secretHash), viewW.secretHash,
	)

	encryptForm := container.New(layout.NewFormLayout(),
		widget.NewLabel(caption.noise), viewW.noise,
		widget.NewLabel(caption.enc), viewW.enc,
	)

	particleForm := container.New(layout.NewFormLayout(),
		widget.NewLabel(caption.noise1), viewW.noise1,
		widget.NewLabel(caption.noise2), viewW.noise2,
		widget.NewLabel(caption.enc1), viewW.enc1,
		widget.NewLabel(caption.enc2), viewW.enc2,
	)

	parityForm := container.New(layout.NewFormLayout(),
		widget.NewLabel(caption.noiseP), viewW.noiseP,
		widget.NewLabel(caption.encP), viewW.encP,
	)

	noisePunchGrid := container.NewGridWithColumns(2, viewW.noise1punch, viewW.noise2punch)
	encPunchGrid := container.NewGridWithColumns(2, viewW.enc1punch, viewW.enc2punch)

	punchForm := container.New(layout.NewFormLayout(),
		widget.NewLabel(caption.noise1punched), viewW.noise1punched,
		widget.NewLabel(caption.noise2punched), viewW.noise2punched,
		widget.NewLabel(caption.noisePunch), noisePunchGrid,

		widget.NewLabel(caption.enc1punched), viewW.enc1punched,
		widget.NewLabel(caption.enc2punched), viewW.enc2punched,
		widget.NewLabel(caption.encPunch), encPunchGrid,
	)

	permuteForm := container.New(layout.NewFormLayout(),
		widget.NewLabel(caption.noise1perm), viewW.noise1perm,
		widget.NewLabel(caption.noise2perm), viewW.noise2perm,
		widget.NewLabel(caption.enc1perm), viewW.enc1perm,
		widget.NewLabel(caption.enc2perm), viewW.enc2perm,
	)

	hashForm := container.New(layout.NewFormLayout(),
		widget.NewLabel(caption.noise1hash), viewW.noise1hash,
		widget.NewLabel(caption.noise2hash), viewW.noise2hash,
		widget.NewLabel(caption.enc1hash), viewW.enc1hash,
		widget.NewLabel(caption.enc2hash), viewW.enc2hash,
	)

	dummy := "                                "
	view.secretHash.Set(dummy)
	view.noise1hash.Set(dummy)
	view.noise2hash.Set(dummy)
	view.enc1hash.Set(dummy)
	view.enc2hash.Set(dummy)

	leftPipeline := wrapPipelineTheme(container.NewVBox(
		verticalGap(headingTopGap), headingW.secretHash, secretHashForm,
		verticalGap(headingTopGap), headingW.enc, encryptForm,
		verticalGap(headingTopGap), headingW.split, particleForm,
		verticalGap(headingTopGap), headingW.parity, parityForm,
	))
	rightPipeline := wrapPipelineTheme(container.NewVBox(
		verticalGap(headingTopGap), headingW.punch, punchForm,
		verticalGap(headingTopGap), headingW.permute, permuteForm,
		verticalGap(headingTopGap), headingW.particleHash, hashForm,
	))

	secretFormThemed := wrapPipelineTheme(secretForm)
	leftContainer := container.NewVBox(headingW.secret, secretFormThemed, leftPipeline)
	grid := container.NewGridWithColumns(2, leftContainer, rightPipeline)

	return grid, &pipelineUI{
		leftPipeline:    leftPipeline,
		rightPipeline:   rightPipeline,
		secretCharLabel: secretCharLabel,
		secretCharValue: viewW.secret,
	}
}
