// Copyright (c) 2026 dido GmbH and others.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// https://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

package gui

import (
	"crypto/md5"
	"encoding/hex"
)

func punch(text []byte) ([]byte, []byte) {
	l := len(text)
	punched := make([]byte, 0)
	punch := make([]byte, 0)

	for i := 0; i < l; i++ {
		if i == 10 || i == 20 || i == 30 || i == 40 || i == 50 || i == 60 {
			punch = append(punch, text[i])
		} else {
			punched = append(punched, text[i])
		}
	}
	return punched, punch
}

func permute(text []byte) []byte {
	l := len(text)
	permuted := make([]byte, 0)
	a := make([]byte, 0)
	b := make([]byte, 0)
	c := make([]byte, 0)

	for i := 0; i < l; i++ {
		if len(a) < 4 {
			a = append(a, text[i])
		} else if len(b) < 4 {
			b = append(b, text[i])
		} else if len(c) < 4 {
			c = append(c, text[i])
		} else {
			permuted = append(permuted, c...)
			permuted = append(permuted, a...)
			permuted = append(permuted, b...)
			a = make([]byte, 0)
			b = make([]byte, 0)
			c = make([]byte, 0)
		}
	}
	return permuted
}

func hash(text []byte) string {
	h := md5.Sum([]byte(text))
	return hex.EncodeToString(h[:])
}
