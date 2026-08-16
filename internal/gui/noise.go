// Copyright (c) 2026 dido GmbH and others.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// https://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

package gui

import (
	"crypto/rand"
)

// PrintableNoise extends noise to length l with printable ASCII bytes (GUI demo only).
func PrintableNoise(noise []byte, l int) []byte {
	size := l - len(noise)
	if size > 0 {
		token := make([]byte, size)
		_, _ = rand.Read(token)
		for i, v := range token {
			if v < 33 {
				token[i] = v + 33
			} else if v > 127+33 {
				token[i] = v - 128
			} else if v > 127 {
				token[i] = v - 33
			} else {
				token[i] = v
			}
		}
		return append(noise, token...)
	}
	return noise[:l]
}

// ParityLegacy XORs equal-length prefixes of two particles (GUI demo only).
func ParityLegacy(p1, p2 []byte) []byte {
	l := min(len(p1), len(p2))
	parity := make([]byte, l)
	for i := 0; i < l; i++ {
		parity[i] = p1[i] ^ p2[i]
	}
	return parity
}
