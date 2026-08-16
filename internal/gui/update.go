// Copyright (c) 2026 dido GmbH and others.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// https://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

package gui

import (
	"github.com/eclipse-pcs/pcs"
)

const map2ascii = true

func allPunchOutsEmpty(noise1punch, noise2punch, enc1punch, enc2punch []byte) bool {
	return len(noise1punch)+len(noise2punch)+len(enc1punch)+len(enc2punch) == 0
}

func update(model *modelData, view viewData) (hasPunchOuts bool) {
	l := len(model.secret)

	model.noise = PrintableNoise(model.noise, l)
	noise1, noise2 := pcs.Split(model.noise)
	noiseP := ParityLegacy(noise1, noise2)

	enc := pcs.Encrypt(model.secret, model.noise)
	enc1, enc2 := pcs.Split(enc)
	encP := ParityLegacy(enc1, enc2)

	noise1punched, noise1punch := punch(noise1)
	noise2punched, noise2punch := punch(noise2)
	noise1perm := permute(noise1punched)
	noise2perm := permute(noise2punched)
	noise1hash := hash(noise1perm)
	noise2hash := hash(noise2perm)

	enc1punched, enc1punch := punch(enc1)
	enc2punched, enc2punch := punch(enc2)
	enc1perm := permute(enc1punched)
	enc2perm := permute(enc2punched)
	enc1hash := hash(enc1perm)
	enc2hash := hash(enc2perm)

	view.secret.Set(byte2string(model.secret, map2ascii))
	view.noise.Set(byte2string(model.noise, map2ascii))
	view.noise1.Set(byte2string(noise1, map2ascii))
	view.noise2.Set(byte2string(noise2, map2ascii))
	view.enc.Set(byte2string(enc, map2ascii))
	view.enc1.Set(byte2string(enc1, map2ascii))
	view.enc2.Set(byte2string(enc2, map2ascii))
	view.noiseP.Set(byte2string(noiseP, map2ascii))
	view.encP.Set(byte2string(encP, map2ascii))

	view.noise1punched.Set(byte2string(noise1punched, map2ascii))
	view.noise1punch.Set(byte2string(noise1punch, map2ascii))
	view.noise2punched.Set(byte2string(noise2punched, map2ascii))
	view.noise2punch.Set(byte2string(noise2punch, map2ascii))

	view.enc1punched.Set(byte2string(enc1punched, map2ascii))
	view.enc1punch.Set(byte2string(enc1punch, map2ascii))
	view.enc2punched.Set(byte2string(enc2punched, map2ascii))
	view.enc2punch.Set(byte2string(enc2punch, map2ascii))

	view.noise1perm.Set(byte2string(noise1perm, map2ascii))
	view.noise2perm.Set(byte2string(noise2perm, map2ascii))
	view.enc1perm.Set(byte2string(enc1perm, map2ascii))
	view.enc2perm.Set(byte2string(enc2perm, map2ascii))

	dummy := "                                "
	if l > 0 {
		view.secretHash.Set(hash(model.secret))
	} else {
		view.secretHash.Set(dummy)
	}
	if len(noise1perm) > 0 && len(noise2perm) > 0 {
		view.noise1hash.Set(noise1hash)
		view.noise2hash.Set(noise2hash)
		view.enc1hash.Set(enc1hash)
		view.enc2hash.Set(enc2hash)
	} else {
		view.noise1hash.Set(dummy)
		view.noise2hash.Set(dummy)
		view.enc1hash.Set(dummy)
		view.enc2hash.Set(dummy)
	}

	return !allPunchOutsEmpty(noise1punch, noise2punch, enc1punch, enc2punch)
}

func byte2string(contents []byte, map2ascii bool) string {
	out := ""
	for i := 0; i < len(contents); i++ {
		cur := contents[i]
		if map2ascii && (cur < 32 || cur > 127 || cur == 127 || cur == 129 || cur == 141 || cur == 143 || cur == 144 || cur == 157 || cur == 160 || cur == 173 || cur == 135 || cur == 131) {
			cur = '_'
		}
		out += string(cur)
	}
	return out
}
