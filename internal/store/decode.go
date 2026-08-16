// Copyright (c) 2026 dido GmbH and others.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// https://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

package store

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/eclipse-pcs/pcs"
	"github.com/eclipse-pcs/pcs/footer"
)

type loadedParticle struct {
	payload []byte
	footer  *footer.Footer
}

// DecodeResult holds the outcome of decoding particle files from disk.
type DecodeResult struct {
	Secret              []byte
	UsedParityRecovery  bool
	CRCParityRecoveries []string
	FingerprintValid    bool
	Inventory           *Inventory
}

// DecodeFromStorage reconstructs secret from present particle files.
func DecodeFromStorage(baseDir, baseName string, inv *Inventory) (*DecodeResult, error) {
	if inv == nil {
		return nil, fmt.Errorf("particle inventory is required")
	}

	loaded, err := loadPresentParticles(baseDir, baseName, inv)
	if err != nil {
		return nil, err
	}

	footers := make(map[pcs.ParticleKind]*footer.Footer, len(loaded))
	particles := make(map[pcs.ParticleKind][]byte, len(loaded))
	for kind, lp := range loaded {
		footers[kind] = lp.footer
		particles[kind] = lp.payload
	}

	if err := footer.VerifyWriteIDs(footers); err != nil {
		return nil, fmt.Errorf("WriteID: %w", err)
	}

	logicalSize, err := deriveLogicalSize(loaded, inv)
	if err != nil {
		return nil, err
	}

	pcsInv, err := pcs.InventoryFromPresent(inv.Present)
	if err != nil {
		return nil, err
	}

	var crcRecoveries []string
	if err := runCrossCRCValidation(baseName, pcsInv, logicalSize, particles, footers, &crcRecoveries); err != nil {
		return nil, err
	}

	for kind, payload := range particles {
		if pcs.CRC32IEEE(payload) != footers[kind].PayloadCRC {
			return nil, fmt.Errorf("payload CRC mismatch on %s", kind)
		}
	}

	secret, usedParity, err := pcs.DecodeWithRecovery(pcsInv, particles, logicalSize)
	if err != nil {
		return nil, err
	}

	fingerprintValid, err := verifyFingerprint(secret, footers)
	if err != nil {
		return nil, fmt.Errorf("fingerprint: %w", err)
	}

	return &DecodeResult{
		Secret:              secret,
		UsedParityRecovery:  usedParity,
		CRCParityRecoveries: crcRecoveries,
		FingerprintValid:    fingerprintValid,
		Inventory:           inv,
	}, nil
}

func loadPresentParticles(baseDir, baseName string, inv *Inventory) (map[pcs.ParticleKind]loadedParticle, error) {
	out := make(map[pcs.ParticleKind]loadedParticle)
	for kind, present := range inv.Present {
		if !present {
			continue
		}
		path := ParticlePath(baseDir, baseName, kind)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read particle file %s: %w", path, err)
		}
		payload, f, err := splitParticleFile(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		out[kind] = loadedParticle{payload: payload, footer: f}
	}
	return out, nil
}

func splitParticleFile(data []byte) (payload []byte, f *footer.Footer, err error) {
	if len(data) < footer.Size {
		return nil, nil, fmt.Errorf("particle file too short: %d bytes, minimum %d", len(data), footer.Size)
	}
	payload = data[:len(data)-footer.Size]
	rawFooter := data[len(data)-footer.Size:]
	f, err = footer.Parse(rawFooter)
	if err != nil {
		return nil, nil, err
	}
	return payload, f, nil
}

func deriveLogicalSize(loaded map[pcs.ParticleKind]loadedParticle, inv *Inventory) (int64, error) {
	for _, lp := range loaded {
		return int64(lp.footer.Length), nil
	}
	return deriveLogicalSizeFromPayloadSizes(loaded, inv)
}

func deriveLogicalSizeFromPayloadSizes(loaded map[pcs.ParticleKind]loadedParticle, inv *Inventory) (int64, error) {
	var evenPayload, oddPayload int64 = -1, -1
	evenKinds := []pcs.ParticleKind{pcs.EvenCypher, pcs.EvenNoise, pcs.CypherParity, pcs.NoiseParity}
	oddKinds := []pcs.ParticleKind{pcs.OddCypher, pcs.OddNoise}
	for _, kind := range evenKinds {
		if lp, ok := loaded[kind]; ok {
			evenPayload = int64(len(lp.payload))
			break
		}
	}
	for _, kind := range oddKinds {
		if lp, ok := loaded[kind]; ok {
			oddPayload = int64(len(lp.payload))
			break
		}
	}
	if evenPayload >= 0 && oddPayload >= 0 {
		return footer.LogicalSizeFromPayloadSizes(evenPayload, oddPayload), nil
	}
	if evenPayload >= 0 && !inv.Present[pcs.OddCypher] && !inv.Present[pcs.OddNoise] {
		if lp, ok := loaded[pcs.CypherParity]; ok {
			return int64(lp.footer.Length), nil
		}
		if lp, ok := loaded[pcs.NoiseParity]; ok {
			return int64(lp.footer.Length), nil
		}
	}
	if evenPayload >= 0 {
		oddLen, evenLen := footer.LogicalSizeCandidates(evenPayload)
		_ = oddLen
		return evenLen, nil
	}
	return 0, fmt.Errorf("cannot derive logical object size from present particles")
}

func verifyFingerprint(secret []byte, footers map[pcs.ParticleKind]*footer.Footer) (bool, error) {
	present := make(map[pcs.ParticleKind]bool)
	shards := make(map[pcs.ParticleKind][]byte)
	for kind, f := range footers {
		present[kind] = true
		shards[kind] = append([]byte(nil), f.FingerprintShard[:]...)
	}
	got, err := pcs.DecodeFingerprint(present, shards)
	if err != nil {
		return false, err
	}
	want := sha256.Sum256(secret)
	return bytes.Equal(got[:], want[:]), nil
}

type crossPair struct {
	name       string
	left       pcs.ParticleKind
	right      pcs.ParticleKind
	parityKind pcs.ParticleKind
	recover    bool
}

func coreCrossPairs() []crossPair {
	return []crossPair{
		{name: "cypher", left: pcs.EvenCypher, right: pcs.OddCypher, parityKind: pcs.CypherParity, recover: true},
		{name: "noise", left: pcs.OddNoise, right: pcs.EvenNoise, parityKind: pcs.NoiseParity, recover: true},
	}
}

func runCrossCRCValidation(
	baseName string,
	inv *pcs.ParticleInventory,
	logicalSize int64,
	particles map[pcs.ParticleKind][]byte,
	footers map[pcs.ParticleKind]*footer.Footer,
	crcRecoveries *[]string,
) error {
	if inv.NeedsParityRecovery() {
		if err := validateParityCrossCRC(baseName, inv, particles, footers); err != nil {
			return err
		}
	}
	for _, pair := range coreCrossPairs() {
		if err := validateCrossCRCPairWithRetry(baseName, inv, logicalSize, pair, particles, footers, crcRecoveries); err != nil {
			return err
		}
	}
	if err := validateParityPairCrossCRC(particles, footers); err != nil {
		return err
	}
	return nil
}

func validateParityCrossCRC(
	baseName string,
	inv *pcs.ParticleInventory,
	particles map[pcs.ParticleKind][]byte,
	footers map[pcs.ParticleKind]*footer.Footer,
) error {
	if !inv.Present[pcs.CypherParity] || !inv.Present[pcs.NoiseParity] {
		return nil
	}
	left := particles[pcs.CypherParity]
	right := particles[pcs.NoiseParity]
	v := footer.VerifyCrossCRC(
		left, footers[pcs.CypherParity].CrossCRC,
		right, footers[pcs.NoiseParity].CrossCRC,
	)
	if v == footer.CrossBothCorrupt {
		return fmt.Errorf("cross-CRC both checks failed for parity pair (%s)", ParticleRelPath(baseName, pcs.CypherParity))
	}
	if v != footer.CrossOK {
		return fmt.Errorf("cross-CRC mismatch for parity pair")
	}
	return nil
}

func validateParityPairCrossCRC(particles map[pcs.ParticleKind][]byte, footers map[pcs.ParticleKind]*footer.Footer) error {
	left := particles[pcs.CypherParity]
	right := particles[pcs.NoiseParity]
	if left == nil || right == nil {
		return nil
	}
	v := footer.VerifyCrossCRC(
		left, footers[pcs.CypherParity].CrossCRC,
		right, footers[pcs.NoiseParity].CrossCRC,
	)
	if v != footer.CrossOK {
		return fmt.Errorf("cross-CRC mismatch for parity pair: %s", v)
	}
	return nil
}

func validateCrossCRCPairWithRetry(
	baseName string,
	inv *pcs.ParticleInventory,
	logicalSize int64,
	pair crossPair,
	particles map[pcs.ParticleKind][]byte,
	footers map[pcs.ParticleKind]*footer.Footer,
	crcRecoveries *[]string,
) error {
	if !inv.Present[pair.left] || !inv.Present[pair.right] {
		return nil
	}

	tryValidate := func() footer.CrossCRCVerdict {
		return footer.VerifyCrossCRC(
			particles[pair.left], footers[pair.left].CrossCRC,
			particles[pair.right], footers[pair.right].CrossCRC,
		)
	}

	v := tryValidate()
	if v == footer.CrossOK {
		return nil
	}
	if v == footer.CrossBothCorrupt || !pair.recover {
		return fmt.Errorf("cross-CRC both checks failed for %s pair", pair.name)
	}

	recoverKind := pair.left
	if v == footer.CrossRightCorrupt {
		recoverKind = pair.right
	}

	if err := recoverCoreParticle(inv, particles, pair.left, pair.right, pair.parityKind, recoverKind, logicalSize); err != nil {
		return fmt.Errorf("cross-CRC recovery for %s pair (%s): %w", pair.name, ParticleRelPath(baseName, recoverKind), err)
	}
	refreshFooterCRCs(particles, footers, pair.left, pair.right)
	if crcRecoveries != nil {
		*crcRecoveries = append(*crcRecoveries, ParticleRelPath(baseName, recoverKind))
	}

	if tryValidate() != footer.CrossOK {
		return fmt.Errorf("cross-CRC still failed for %s pair after parity recovery", pair.name)
	}
	return nil
}

func recoverCoreParticle(
	inv *pcs.ParticleInventory,
	particles map[pcs.ParticleKind][]byte,
	evenKind, oddKind, parityKind, recoverKind pcs.ParticleKind,
	logicalSize int64,
) error {
	trial := cloneInventory(inv)
	trial.Present[recoverKind] = false
	oddLength := logicalSize%2 == 1
	even, odd, err := recoverPair(trial, particles, evenKind, oddKind, parityKind, oddLength)
	if err != nil {
		return err
	}
	particles[evenKind] = even
	particles[oddKind] = odd
	return nil
}

func recoverPair(
	inv *pcs.ParticleInventory,
	particles map[pcs.ParticleKind][]byte,
	evenKind, oddKind, parityKind pcs.ParticleKind,
	oddLength bool,
) (even, odd []byte, err error) {
	if oddLength {
		return recoverPairOdd(inv, particles, evenKind, oddKind, parityKind)
	}
	return recoverPairEven(inv, particles, evenKind, oddKind, parityKind)
}

func recoverPairEven(
	inv *pcs.ParticleInventory,
	particles map[pcs.ParticleKind][]byte,
	evenKind, oddKind, parityKind pcs.ParticleKind,
) (even, odd []byte, err error) {
	haveEven := inv.Present[evenKind]
	haveOdd := inv.Present[oddKind]
	switch {
	case haveEven && haveOdd:
		return particles[evenKind], particles[oddKind], nil
	case haveEven && !haveOdd:
		odd, err = pcs.ReconstructFromParityEven(particles[evenKind], particles[parityKind])
		return particles[evenKind], odd, err
	case !haveEven && haveOdd:
		even, err = pcs.ReconstructFromParityEven(particles[oddKind], particles[parityKind])
		return even, particles[oddKind], err
	default:
		return nil, nil, fmt.Errorf("cannot recover %s/%s pair from parity alone", evenKind, oddKind)
	}
}

func recoverPairOdd(
	inv *pcs.ParticleInventory,
	particles map[pcs.ParticleKind][]byte,
	evenKind, oddKind, parityKind pcs.ParticleKind,
) (even, odd []byte, err error) {
	haveEven := inv.Present[evenKind]
	haveOdd := inv.Present[oddKind]
	switch {
	case haveEven && haveOdd:
		return particles[evenKind], particles[oddKind], nil
	case haveEven && !haveOdd:
		odd, err = pcs.ReconstructOddFromParityOdd(particles[evenKind], particles[parityKind])
		return particles[evenKind], odd, err
	case !haveEven && haveOdd:
		even, err = pcs.ReconstructEvenFromParityOdd(particles[oddKind], particles[parityKind])
		return even, particles[oddKind], err
	default:
		return nil, nil, fmt.Errorf("cannot recover %s/%s pair from parity alone", evenKind, oddKind)
	}
}

func cloneInventory(inv *pcs.ParticleInventory) *pcs.ParticleInventory {
	cp := make(map[pcs.ParticleKind]bool, len(inv.Present))
	for k, v := range inv.Present {
		cp[k] = v
	}
	return &pcs.ParticleInventory{Present: cp}
}

func refreshFooterCRCs(particles map[pcs.ParticleKind][]byte, footers map[pcs.ParticleKind]*footer.Footer, left, right pcs.ParticleKind) {
	if f := footers[left]; f != nil {
		f.PayloadCRC = pcs.CRC32IEEE(particles[left])
		f.CrossCRC = pcs.CRC32IEEE(particles[right])
	}
	if f := footers[right]; f != nil {
		f.PayloadCRC = pcs.CRC32IEEE(particles[right])
		f.CrossCRC = pcs.CRC32IEEE(particles[left])
	}
}
