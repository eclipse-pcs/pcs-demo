// Copyright (c) 2026 dido GmbH and others.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// https://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

package pcs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eclipse-pcs/pcs"
	"github.com/eclipse-pcs/pcs/footer"

	"github.com/eclipse-pcs/pcs-demo/internal/store"
)

func TestEncodeHelloTxt(t *testing.T) {
	runEncodeTest(t, "Hello.txt", "Hello Freiburg")
}

func TestEncodeOddLengthSecret(t *testing.T) {
	runEncodeTest(t, "Odd.txt", "Hello")
}

func TestEncodeEmptyFile(t *testing.T) {
	runEncodeTest(t, "Empty.txt", "")
}

func runEncodeTest(t *testing.T, fileName, content string) {
	t.Helper()

	dataDir := "_data"
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatalf("create data directory: %v", err)
	}

	inputPath := filepath.Join(dataDir, fileName)
	if err := os.WriteFile(inputPath, []byte(content), 0o644); err != nil {
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

	secret := []byte(content)
	if err := store.SaveEncoded(".", fileName, secret); err != nil {
		t.Fatalf("save encoded file: %v", err)
	}

	for _, dir := range []string{pcs.StorageA, pcs.StorageB, pcs.StorageC} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("expected storage directory %s: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %s to be a directory", dir)
		}
	}

	for _, kind := range pcs.AllParticleKinds {
		path := store.ParticleRelPath(fileName, kind)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read particle file %s: %v", path, err)
		}
		if len(data) < footer.Size {
			t.Fatalf("particle file %s too short: %d bytes", path, len(data))
		}
		payloadLen := len(data) - footer.Size
		wantPayload := expectedPayloadLen(len(secret), kind)
		if payloadLen != wantPayload {
			t.Fatalf("%s payload len = %d, want %d", path, payloadLen, wantPayload)
		}
		if len(data) != wantPayload+footer.Size {
			t.Fatalf("%s total len = %d, want %d", path, len(data), wantPayload+footer.Size)
		}
	}
}

func expectedPayloadLen(n int, kind pcs.ParticleKind) int {
	switch kind {
	case pcs.EvenCypher, pcs.EvenNoise, pcs.CypherParity, pcs.NoiseParity:
		return (n + 1) / 2
	default:
		return n / 2
	}
}
