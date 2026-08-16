// Copyright (c) 2026 dido GmbH and others.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// https://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

type headingWidget struct {
	secret                                                     *widget.Label
	secretHash, enc, split, parity, punch, permute, particleHash *fyne.Container
}

type captionData struct {
	secret, secretHash, char, noise, enc, noise1, noise2, noiseP, enc1, enc2, encP, enc1punched, enc2punched, encPunch, noise1punched, noise2punched, noisePunch, enc1perm, enc2perm, noise1perm, noise2perm, noise1hash, noise2hash, enc1hash, enc2hash string
}

type modelData struct {
	secret, noise []byte
}

type viewData struct {
	secretHash, noise, secret, enc, noise1, noise2, noiseP, enc1, enc2, encP, enc1punched, enc1punch, enc2punched, enc2punch, noise1punched, noise1punch, noise2punched, noise2punch, enc1perm, enc2perm, noise1perm, noise2perm, noise1hash, noise2hash, enc1hash, enc2hash binding.String
}

type viewWidget struct {
	secretHash, noise, secret, enc, noise1, noise2, noiseP, enc1, enc2, encP, enc1punched, enc1punch, enc2punched, enc2punch, noise1punched, noise1punch, noise2punched, noise2punch, enc1perm, enc2perm, noise1perm, noise2perm, noise1hash, noise2hash, enc1hash, enc2hash *widget.Label
}

func bindings() viewData {
	var view viewData
	view.secretHash = binding.NewString()
	view.noise = binding.NewString()
	view.secret = binding.NewString()
	view.enc = binding.NewString()
	view.noise1 = binding.NewString()
	view.noise2 = binding.NewString()
	view.noiseP = binding.NewString()
	view.enc1 = binding.NewString()
	view.enc2 = binding.NewString()
	view.encP = binding.NewString()
	view.noise1punched = binding.NewString()
	view.noise1punch = binding.NewString()
	view.noise2punched = binding.NewString()
	view.noise2punch = binding.NewString()
	view.noise1perm = binding.NewString()
	view.noise2perm = binding.NewString()

	view.enc1punched = binding.NewString()
	view.enc1punch = binding.NewString()
	view.enc2punched = binding.NewString()
	view.enc2punch = binding.NewString()
	view.enc1perm = binding.NewString()
	view.enc2perm = binding.NewString()

	view.noise1hash = binding.NewString()
	view.noise2hash = binding.NewString()
	view.enc1hash = binding.NewString()
	view.enc2hash = binding.NewString()

	return view
}

func initialize(view *viewData) (modelData, headingWidget, captionData, viewWidget) {
	headingW := initHeading()
	caption := initCaption()
	viewW := initView(view)

	var model modelData
	model.noise = []byte("")

	return model, headingW, caption, viewW
}

func newStyledHeading(title, subtitle string) *fyne.Container {
	titleLabel := widget.NewLabel(title)
	titleLabel.TextStyle.Bold = true
	subLabel := widget.NewLabel(subtitle)
	subLabel.TextStyle.Bold = true
	subLabel.TextStyle.Italic = true
	return container.NewHBox(titleLabel, subLabel)
}

func initHeading() headingWidget {
	var headingW headingWidget

	headingW.secret = widget.NewLabel("Data to be kept secret")
	headingW.secret.TextStyle.Bold = true

	headingW.secretHash = newStyledHeading("PCS Integrity: ", "Hash of secret for integrity validation")
	headingW.enc = newStyledHeading("PCS Encrypter: ", "One-time pad, encryption with random noise")
	headingW.split = newStyledHeading("PCS Splitter: ", "secret sharing, split data into particles")
	headingW.parity = newStyledHeading("PCS Parity: ", "disaster recovery with parity particles")
	headingW.punch = newStyledHeading("PCS Stealth: ", "transport through public internet, punching data")
	headingW.permute = newStyledHeading("PCS Stealth: ", "transport through public internet, permutation of data sections")
	headingW.particleHash = newStyledHeading("PCS Integrity: ", "hashes of particles for cloud storage validation")

	return headingW
}

func initCaption() captionData {
	return captionData{
		secret:        "Secret to be encrypted",
		secretHash:    "Hash of secret",
		char:          "Secret (printable ASCII only)",
		noise:         "Random noise key",
		enc:           "Ciphertext (secret XOR noise)",
		noise1:        "Even noise particle",
		noise2:        "Odd noise particle",
		enc1:          "Even cypher particle",
		enc2:          "Odd cypher particle",
		noiseP:        "Parity particle of noise",
		encP:          "Parity particle of cyphertext",
		noise1punched: "Punched even noise particle",
		noise2punched: "Punched odd noise particle",
		noisePunch:    "Punched-outs of even / odd noise particles",
		enc1punched:   "Punched even cyphertext particle",
		enc2punched:   "Punched odd cyphertext particle",
		encPunch:      "Punched-outs of even / odd cyphertext particles",
		noise1perm:    "Permuted punched even noise particle",
		noise2perm:    "Permuted punched odd noise particle",
		enc1perm:      "Permuted punched even cyphertext particle",
		enc2perm:      "Permuted punched odd cyphertext particle",
		noise1hash:    "Hash of permuted punched even noise",
		noise2hash:    "Hash of permuted punched odd noise",
		enc1hash:      "Hash of permuted punched even cyphertext",
		enc2hash:      "Hash of permuted punched odd cyphertext",
	}
}

func initView(view *viewData) viewWidget {
	var viewW viewWidget
	viewW.secret = widget.NewLabelWithData(view.secret)
	viewW.secretHash = widget.NewLabelWithData(view.secretHash)
	viewW.noise = widget.NewLabelWithData(view.noise)
	viewW.enc = widget.NewLabelWithData(view.enc)
	viewW.noise1 = widget.NewLabelWithData(view.noise1)
	viewW.noise2 = widget.NewLabelWithData(view.noise2)
	viewW.noiseP = widget.NewLabelWithData(view.noiseP)
	viewW.enc1 = widget.NewLabelWithData(view.enc1)
	viewW.enc2 = widget.NewLabelWithData(view.enc2)
	viewW.encP = widget.NewLabelWithData(view.encP)
	viewW.noise1punched = widget.NewLabelWithData(view.noise1punched)
	viewW.noise1punch = widget.NewLabelWithData(view.noise1punch)
	viewW.noise2punched = widget.NewLabelWithData(view.noise2punched)
	viewW.noise2punch = widget.NewLabelWithData(view.noise2punch)
	viewW.noise1perm = widget.NewLabelWithData(view.noise1perm)
	viewW.noise2perm = widget.NewLabelWithData(view.noise2perm)

	viewW.enc1punched = widget.NewLabelWithData(view.enc1punched)
	viewW.enc1punch = widget.NewLabelWithData(view.enc1punch)
	viewW.enc2punched = widget.NewLabelWithData(view.enc2punched)
	viewW.enc2punch = widget.NewLabelWithData(view.enc2punch)
	viewW.enc1perm = widget.NewLabelWithData(view.enc1perm)
	viewW.enc2perm = widget.NewLabelWithData(view.enc2perm)

	viewW.noise1hash = widget.NewLabelWithData(view.noise1hash)
	viewW.noise2hash = widget.NewLabelWithData(view.noise2hash)
	viewW.enc1hash = widget.NewLabelWithData(view.enc1hash)
	viewW.enc2hash = widget.NewLabelWithData(view.enc2hash)

	styleSecretValue(viewW.secret)
	styleSecretValue(viewW.secretHash)
	styleEncryptValue(viewW.noise)
	styleEncryptValue(viewW.enc)
	styleParticleValue(viewW.noise1)
	styleParticleValue(viewW.noise2)
	styleParticleValue(viewW.enc1)
	styleParticleValue(viewW.enc2)
	styleParticleValue(viewW.noiseP)
	styleParticleValue(viewW.encP)
	styleParticleValue(viewW.noise1punched)
	styleParticleValue(viewW.noise1punch)
	styleParticleValue(viewW.noise2punched)
	styleParticleValue(viewW.noise2punch)
	styleParticleValue(viewW.enc1punched)
	styleParticleValue(viewW.enc1punch)
	styleParticleValue(viewW.enc2punched)
	styleParticleValue(viewW.enc2punch)
	styleParticleValue(viewW.noise1perm)
	styleParticleValue(viewW.noise2perm)
	styleParticleValue(viewW.enc1perm)
	styleParticleValue(viewW.enc2perm)
	styleParticleValue(viewW.noise1hash)
	styleParticleValue(viewW.noise2hash)
	styleParticleValue(viewW.enc1hash)
	styleParticleValue(viewW.enc2hash)

	return viewW
}
