// Copyright (c) 2026 dido GmbH and others.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// https://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

package store

import (
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/eclipse-pcs/pcs"
	"github.com/eclipse-pcs/pcs/footer"
)

// SaveEncoded writes six footer-v1 particle files for secret under baseDir.
func SaveEncoded(baseDir, baseName string, secret []byte) error {
	result, err := pcs.Encode(secret)
	if err != nil {
		return fmt.Errorf("encode secret: %w", err)
	}

	sum := sha256.Sum256(secret)
	fpResult, err := pcs.EncodeFingerprint(sum)
	if err != nil {
		return fmt.Errorf("encode fingerprint: %w", err)
	}

	writeID, err := footer.NewWriteID()
	if err != nil {
		return fmt.Errorf("generate WriteID: %w", err)
	}

	payloads := encodePayloads(result)
	footers := buildFooters(payloads, fpResult, writeID, uint64(len(secret)))

	if err := EnsureStorageDirs(baseDir); err != nil {
		return err
	}

	for _, kind := range pcs.AllParticleKinds {
		path := ParticlePath(baseDir, baseName, kind)
		rawFooter := footers[kind].Marshal()
		raw := append(append([]byte(nil), payloads[kind]...), rawFooter[:]...)
		if err := os.WriteFile(path, raw, 0o644); err != nil {
			return fmt.Errorf("write particle file %s: %w", path, err)
		}
	}
	return nil
}

func encodePayloads(result *pcs.EncodeResult) map[pcs.ParticleKind][]byte {
	out := make(map[pcs.ParticleKind][]byte, len(pcs.AllParticleKinds))
	for _, kind := range pcs.AllParticleKinds {
		out[kind] = pcs.EncodeResultShard(result, kind)
	}
	return out
}

func buildFooters(
	payloads map[pcs.ParticleKind][]byte,
	fpResult *pcs.EncodeResult,
	writeID uint64,
	logicalLen uint64,
) map[pcs.ParticleKind]*footer.Footer {
	out := make(map[pcs.ParticleKind]*footer.Footer, len(pcs.AllParticleKinds))
	for _, kind := range pcs.AllParticleKinds {
		partner := footer.PartnerKind(kind)
		var shard [16]byte
		footer.CopyFingerprintShard(&shard, pcs.EncodeResultShard(fpResult, kind))
		out[kind] = &footer.Footer{
			Version:          footer.Version,
			Kind:             kind,
			Length:           logicalLen,
			FingerprintShard: shard,
			PayloadCRC:       pcs.CRC32IEEE(payloads[kind]),
			CrossCRC:         pcs.CRC32IEEE(payloads[partner]),
			WriteID:          writeID,
		}
	}
	return out
}
