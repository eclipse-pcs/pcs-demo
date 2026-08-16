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
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eclipse-pcs/pcs"
	"github.com/eclipse-pcs/pcs/footer"

	"github.com/eclipse-pcs/pcs-demo/internal/store"
)

func TestCLIEncodeDecodeRoundTrip(t *testing.T) {
	dir := t.TempDir()
	encodeBin := buildCLI(t, "pcs-encode")
	decodeBin := buildCLI(t, "pcs-decode")

	fileName := "Hello.txt"
	secret := []byte("Hello Freiburg")
	inputPath := filepath.Join(dir, fileName)
	if err := os.WriteFile(inputPath, secret, 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	runCLI(t, encodeBin, dir, "-y", fileName)

	for _, kind := range pcs.AllParticleKinds {
		rel := store.ParticleRelPath(fileName, kind)
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("expected particle file %s: %v", rel, err)
		}
	}

	output := runCLI(t, decodeBin, dir, "-y", fileName)
	if !strings.Contains(output, "successfully decoded Hello_reconstructed.txt") {
		t.Fatalf("decode output = %q, want success message", output)
	}

	got, err := os.ReadFile(filepath.Join(dir, "Hello_reconstructed.txt"))
	if err != nil {
		t.Fatalf("read reconstructed file: %v", err)
	}
	if !bytes.Equal(got, secret) {
		t.Fatalf("reconstructed content = %q, want %q", got, secret)
	}
}

func TestCLIDecodeWithParityRecovery(t *testing.T) {
	dir := t.TempDir()
	encodeBin := buildCLI(t, "pcs-encode")
	decodeBin := buildCLI(t, "pcs-decode")

	fileName := "Hello.txt"
	secret := []byte("Hello Freiburg")
	inputPath := filepath.Join(dir, fileName)
	if err := os.WriteFile(inputPath, secret, 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	runCLI(t, encodeBin, dir, "-y", fileName)
	if err := os.Remove(store.ParticlePath(dir, fileName, pcs.EvenCypher)); err != nil {
		t.Fatalf("remove even cypher: %v", err)
	}

	output := runCLI(t, decodeBin, dir, "-y", fileName)
	if !strings.Contains(output, "reconstructing missing particles") {
		t.Fatalf("decode output = %q, want parity recovery message", output)
	}

	got, err := os.ReadFile(filepath.Join(dir, "Hello_reconstructed.txt"))
	if err != nil {
		t.Fatalf("read reconstructed file: %v", err)
	}
	if !bytes.Equal(got, secret) {
		t.Fatalf("reconstructed content = %q, want %q", got, secret)
	}
}

func TestCLIDecodeFingerprintMismatchYAborts(t *testing.T) {
	dir := t.TempDir()
	encodeBin := buildCLI(t, "pcs-encode")
	decodeBin := buildCLI(t, "pcs-decode")
	fileName := "Hello.txt"
	secret := []byte("Hello Freiburg")
	if err := os.WriteFile(filepath.Join(dir, fileName), secret, 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}
	runCLI(t, encodeBin, dir, "-y", fileName)
	corruptFingerprintShard(t, dir, fileName)

	output, err := runCLIExpectError(t, decodeBin, dir, "-y", fileName)
	if err == nil {
		t.Fatalf("expected abort, output: %s", output)
	}
	if !strings.Contains(output, "could not be validated") || !strings.Contains(output, "aborted") {
		t.Fatalf("output = %q, want fingerprint mismatch and abort", output)
	}
	if _, err := os.Stat(filepath.Join(dir, "Hello_reconstructed.txt")); !os.IsNotExist(err) {
		t.Fatal("expected no reconstructed file")
	}
}

func TestCLIDecodeCorruptPayloadReportsParityRecovery(t *testing.T) {
	dir := t.TempDir()
	encodeBin := buildCLI(t, "pcs-encode")
	decodeBin := buildCLI(t, "pcs-decode")
	fileName := "Hello.txt"
	secret := []byte("Hello Freiburg")
	if err := os.WriteFile(filepath.Join(dir, fileName), secret, 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}
	runCLI(t, encodeBin, dir, "-y", fileName)
	corruptFileByte(t, store.ParticlePath(dir, fileName, pcs.EvenCypher), 0)

	output := runCLI(t, decodeBin, dir, "-y", fileName)
	want := "cross-CRC mismatch: reconstructed " + store.ParticleRelPath(fileName, pcs.EvenCypher)
	if !strings.Contains(output, want) {
		t.Fatalf("output = %q, want cross-CRC parity recovery message", output)
	}
	if !strings.Contains(output, "successfully decoded Hello_reconstructed.txt") {
		t.Fatalf("output = %q, want success message", output)
	}
}

func TestCLIDecodeCorruptFooterSucceeds(t *testing.T) {
	dir := t.TempDir()
	encodeBin := buildCLI(t, "pcs-encode")
	decodeBin := buildCLI(t, "pcs-decode")
	fileName := "Hello.txt"
	secret := []byte("Hello Freiburg")
	if err := os.WriteFile(filepath.Join(dir, fileName), secret, 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}
	runCLI(t, encodeBin, dir, "-y", fileName)
	corruptFileCrossCRC(t, store.ParticlePath(dir, fileName, pcs.OddCypher))

	output := runCLI(t, decodeBin, dir, "-y", fileName)
	want := "cross-CRC mismatch: reconstructed " + store.ParticleRelPath(fileName, pcs.EvenCypher)
	if !strings.Contains(output, want) {
		t.Fatalf("output = %q, want cross-CRC parity recovery message for .ec", output)
	}
	if !strings.Contains(output, "successfully decoded Hello_reconstructed.txt") {
		t.Fatalf("output = %q, want success message", output)
	}
	got, err := os.ReadFile(filepath.Join(dir, "Hello_reconstructed.txt"))
	if err != nil {
		t.Fatalf("read reconstructed: %v", err)
	}
	if !bytes.Equal(got, secret) {
		t.Fatalf("reconstructed = %q, want %q", got, secret)
	}
}

func TestCLIDecodeCrossCRCBothFail(t *testing.T) {
	dir := t.TempDir()
	encodeBin := buildCLI(t, "pcs-encode")
	decodeBin := buildCLI(t, "pcs-decode")
	fileName := "Hello.txt"
	if err := os.WriteFile(filepath.Join(dir, fileName), []byte("Hello Freiburg"), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}
	runCLI(t, encodeBin, dir, "-y", fileName)
	corruptFileByte(t, store.ParticlePath(dir, fileName, pcs.EvenCypher), 0)
	corruptFileByte(t, store.ParticlePath(dir, fileName, pcs.OddCypher), 0)

	output, err := runCLIExpectError(t, decodeBin, dir, "-y", fileName)
	if err == nil {
		t.Fatalf("expected decode to fail, output: %s", output)
	}
	if !strings.Contains(output, "both checks failed") {
		t.Fatalf("output = %q, want both checks failed", output)
	}
}

func TestCLIEncodeDecodeOddLength(t *testing.T) {
	dir := t.TempDir()
	encodeBin := buildCLI(t, "pcs-encode")
	decodeBin := buildCLI(t, "pcs-decode")
	fileName := "Odd.txt"
	secret := []byte("Hello")
	if err := os.WriteFile(filepath.Join(dir, fileName), secret, 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}
	runCLI(t, encodeBin, dir, "-y", fileName)
	runCLI(t, decodeBin, dir, "-y", fileName)
	got, err := os.ReadFile(filepath.Join(dir, "Odd_reconstructed.txt"))
	if err != nil {
		t.Fatalf("read reconstructed: %v", err)
	}
	if !bytes.Equal(got, secret) {
		t.Fatalf("reconstructed = %q, want %q", got, secret)
	}
}

func buildCLI(t *testing.T, name string) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), name)
	cmd := exec.Command("go", "build", "-o", out, "./cmd/"+name)
	cmd.Dir = moduleRoot(t)
	if outBytes, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build %s in %s: %v\n%s", name, cmd.Dir, err, outBytes)
	}
	return out
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func runCLI(t *testing.T, bin, dir string, args ...string) string {
	t.Helper()
	output, err := runCLIExpectError(t, bin, dir, args...)
	if err != nil {
		t.Fatalf("run %s: %v\n%s", filepath.Base(bin), err, output)
	}
	return output
}

func runCLIExpectError(t *testing.T, bin, dir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func corruptFingerprintShard(t *testing.T, dir, fileName string) {
	t.Helper()
	path := store.ParticlePath(dir, fileName, pcs.EvenCypher)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if len(data) < footer.Size {
		t.Fatalf("file too short")
	}
	// Corrupt fingerprint shard bytes in footer without fixing cross-CRC.
	data[len(data)-footer.Size+16] ^= 0xff
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func corruptFileCrossCRC(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if len(data) < footer.Size {
		t.Fatalf("file %s too short", path)
	}
	// Flip stored cross-CRC field (footer offset 36) while keeping magic/version valid.
	idx := len(data) - footer.Size + 36
	data[idx] ^= 0xff
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func corruptFileByte(t *testing.T, path string, index int) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if index >= len(data) {
		t.Fatalf("index %d out of range for %s", index, path)
	}
	data[index] ^= 0xff
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
