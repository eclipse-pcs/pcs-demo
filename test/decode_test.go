// Copyright (c) 2026 dido GmbH and others.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// https://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

package pcs_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/eclipse-pcs/pcs"

	"github.com/eclipse-pcs/pcs-demo/internal/store"
)

func TestDecodeHelloTxt(t *testing.T) {
	runDecodeRoundTrip(t, "Hello.txt", "Hello Freiburg")
}

func TestDecodeOddLengthSecret(t *testing.T) {
	runDecodeRoundTrip(t, "Odd.txt", "Hello")
}

func TestDecodeEmptyFile(t *testing.T) {
	runDecodeRoundTrip(t, "Empty.txt", "")
}

func TestDecodeMissingEvenCypher(t *testing.T) {
	runParityRecoveryDecode(t, "Hello.txt", "Hello Freiburg", []string{
		store.ParticleRelPath("Hello.txt", pcs.EvenCypher),
	})
}

func TestDecodeMissingOddCypher(t *testing.T) {
	runParityRecoveryDecode(t, "Hello.txt", "Hello Freiburg", []string{
		store.ParticleRelPath("Hello.txt", pcs.OddCypher),
	})
}

func TestDecodeMissingOddNoise(t *testing.T) {
	runParityRecoveryDecode(t, "Hello.txt", "Hello Freiburg", []string{
		store.ParticleRelPath("Hello.txt", pcs.OddNoise),
	})
}

func TestDecodeMissingBothOddCores(t *testing.T) {
	runParityRecoveryDecode(t, "Odd.txt", "Hello", []string{
		store.ParticleRelPath("Odd.txt", pcs.OddCypher),
		store.ParticleRelPath("Odd.txt", pcs.OddNoise),
	})
}

func TestDecodeMissingStorageA(t *testing.T) {
	runParityRecoveryDecode(t, "Hello.txt", "Hello Freiburg", nil, pcs.StorageA)
}

func TestDecodeOddMissingStorageB(t *testing.T) {
	runParityRecoveryDecodeOdd(t, "Odd.txt", "Hello", nil, pcs.StorageB)
}

func TestDecodeUnrecoverableBothCypher(t *testing.T) {
	secret := []byte("Hello Freiburg")
	fileName := "Hello.txt"
	dir := setupEncodedEvenSecret(t, fileName, secret)

	_ = os.Remove(store.ParticlePath(dir, fileName, pcs.EvenCypher))
	_ = os.Remove(store.ParticlePath(dir, fileName, pcs.OddCypher))

	inv, err := store.ScanInventory(dir, fileName)
	if err != nil {
		t.Fatalf("scan particle inventory: %v", err)
	}

	_, err = store.DecodeFromStorage(dir, fileName, inv)
	if err == nil {
		t.Fatal("expected error when both cypher particles are missing")
	}
}

func runParityRecoveryDecodeOdd(t *testing.T, fileName, content string, removeFiles []string, removeStorage ...string) {
	t.Helper()
	dir := setupEncodedSecret(t, fileName, []byte(content))
	runParityRecoveryInDir(t, dir, fileName, []byte(content), removeFiles, removeStorage...)
}

func runParityRecoveryDecode(t *testing.T, fileName, content string, removeFiles []string, removeStorage ...string) {
	t.Helper()
	dir := setupEncodedEvenSecret(t, fileName, []byte(content))
	runParityRecoveryInDir(t, dir, fileName, []byte(content), removeFiles, removeStorage...)
}

func runParityRecoveryInDir(t *testing.T, dir, fileName string, secret []byte, removeFiles []string, removeStorage ...string) {
	t.Helper()

	for _, relPath := range removeFiles {
		if err := os.Remove(filepath.Join(dir, relPath)); err != nil {
			t.Fatalf("remove %s: %v", relPath, err)
		}
	}
	for _, storage := range removeStorage {
		if err := os.RemoveAll(filepath.Join(dir, storage)); err != nil {
			t.Fatalf("remove storage %s: %v", storage, err)
		}
	}

	inv, err := store.ScanInventory(dir, fileName)
	if err != nil {
		t.Fatalf("scan inventory: %v", err)
	}
	result, err := store.DecodeFromStorage(dir, fileName, inv)
	if err != nil {
		t.Fatalf("decode from storage: %v", err)
	}
	if !result.UsedParityRecovery {
		t.Fatal("expected parity recovery to be used")
	}
	if !result.FingerprintValid {
		t.Fatal("expected fingerprint validation to pass")
	}
	if !bytes.Equal(result.Secret, secret) {
		t.Fatalf("decoded secret = %q, want %q", result.Secret, secret)
	}
}

func setupEncodedEvenSecret(t *testing.T, fileName string, secret []byte) string {
	return setupEncodedSecret(t, fileName, secret)
}

func setupEncodedSecret(t *testing.T, fileName string, secret []byte) string {
	t.Helper()

	dataDir := "_data"
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatalf("create data directory: %v", err)
	}

	inputPath := filepath.Join(dataDir, fileName)
	if err := os.WriteFile(inputPath, secret, 0o644); err != nil {
		t.Fatalf("write %s: %v", fileName, err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(dataDir); err != nil {
		t.Fatalf("change to data directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})

	if err := store.SaveEncoded(".", fileName, secret); err != nil {
		t.Fatalf("save encoded file: %v", err)
	}
	return "."
}

func runDecodeRoundTrip(t *testing.T, fileName, content string) {
	t.Helper()

	dir := setupEncodedSecret(t, fileName, []byte(content))

	inv, err := store.ScanInventory(dir, fileName)
	if err != nil {
		t.Fatalf("scan inventory: %v", err)
	}
	result, err := store.DecodeFromStorage(dir, fileName, inv)
	if err != nil {
		t.Fatalf("decode from storage: %v", err)
	}
	if result.UsedParityRecovery {
		t.Fatal("did not expect parity recovery for full particle set")
	}
	if !result.FingerprintValid {
		t.Fatal("expected fingerprint validation to pass")
	}
	if !bytes.Equal(result.Secret, []byte(content)) {
		t.Fatalf("decoded secret = %q, want %q", result.Secret, content)
	}

	outputName := store.ReconstructedFileName(fileName)
	outputPath := filepath.Join(dir, outputName)
	if err := os.WriteFile(outputPath, result.Secret, 0o644); err != nil {
		t.Fatalf("write reconstructed file: %v", err)
	}

	written, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read reconstructed file: %v", err)
	}
	if !bytes.Equal(written, []byte(content)) {
		t.Fatalf("reconstructed file content = %q, want %q", written, content)
	}
}
