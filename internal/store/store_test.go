// Copyright (c) 2026 dido GmbH and others.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// https://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

package store_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eclipse-pcs/pcs"
	"github.com/eclipse-pcs/pcs/footer"

	"github.com/eclipse-pcs/pcs-demo/internal/store"
)

func TestSplitParticleFileTooShort(t *testing.T) {
	dir := t.TempDir()
	baseName := "Hello.txt"
	if err := store.SaveEncoded(dir, baseName, []byte("x")); err != nil {
		t.Fatalf("save: %v", err)
	}
	path := store.ParticlePath(dir, baseName, pcs.EvenCypher)
	if err := os.WriteFile(path, []byte{1, 2, 3}, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	inv, err := store.ScanInventory(dir, baseName)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	_, err = store.DecodeFromStorage(dir, baseName, inv)
	if err == nil {
		t.Fatal("expected error for short particle file")
	}
}

func TestDeleteParticleFilesRemovesExistingPaths(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "storageC", "Hello.txt.np")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := store.DeleteParticleFiles([]string{path}); err != nil {
		t.Fatalf("delete files: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected file to be deleted, stat err = %v", err)
	}
}

func TestExistingParticleFilesUnifiedSuffixes(t *testing.T) {
	dir := t.TempDir()
	baseName := "Hello.txt"
	touch := func(relPath string) {
		path := filepath.Join(dir, relPath)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatalf("write %s: %v", relPath, err)
		}
	}
	touch(store.ParticleRelPath(baseName, pcs.EvenCypher))
	got := store.ExistingParticleFiles(dir, baseName)
	if len(got) != 1 {
		t.Fatalf("existing files = %v, want one file", got)
	}
}

func TestSaveEncodedFooterMagic(t *testing.T) {
	dir := t.TempDir()
	secret := []byte("Hello Freiburg")
	if err := store.SaveEncoded(dir, "Hello.txt", secret); err != nil {
		t.Fatalf("save encoded: %v", err)
	}
	path := store.ParticlePath(dir, "Hello.txt", pcs.EvenCypher)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(data) < footer.Size {
		t.Fatalf("file too short")
	}
	rawFooter := data[len(data)-footer.Size:]
	if rawFooter[0] != 'P' || rawFooter[1] != 'C' || rawFooter[2] != 'S' || rawFooter[3] != 0 {
		t.Fatalf("bad footer magic: % x", rawFooter[:4])
	}
	if rawFooter[4] != 0x01 || rawFooter[5] != 0x00 {
		t.Fatalf("bad footer version: % x", rawFooter[4:6])
	}
}
