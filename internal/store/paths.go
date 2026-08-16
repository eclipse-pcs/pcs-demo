// Copyright (c) 2026 dido GmbH and others.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// https://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

package store

import (
	"os"
	"path/filepath"

	"github.com/eclipse-pcs/pcs"
)

// MinParticleFileSize is the smallest valid on-disk particle (empty payload + footer).
const MinParticleFileSize = footerSize

const footerSize = 64

// ParticleRelPath returns the storage-relative path for a particle kind.
func ParticleRelPath(baseName string, kind pcs.ParticleKind) string {
	return filepath.Join(pcs.StorageForParticle(kind), pcs.ShardKey(baseName, kind))
}

// ParticlePath returns the absolute path for a particle kind.
func ParticlePath(baseDir, baseName string, kind pcs.ParticleKind) string {
	return filepath.Join(baseDir, ParticleRelPath(baseName, kind))
}

// AllParticleRelPaths returns all six expected particle paths.
func AllParticleRelPaths(baseName string) []string {
	out := make([]string, 0, len(pcs.AllParticleKinds))
	for _, kind := range pcs.AllParticleKinds {
		out = append(out, ParticleRelPath(baseName, kind))
	}
	return out
}

// ExistingParticleFiles lists particle files that already exist under baseDir.
func ExistingParticleFiles(baseDir, baseName string) []string {
	var existing []string
	for _, rel := range AllParticleRelPaths(baseName) {
		path := filepath.Join(baseDir, rel)
		if _, err := os.Stat(path); err == nil {
			existing = append(existing, path)
		}
	}
	return existing
}

// DeleteParticleFiles removes the given particle file paths.
func DeleteParticleFiles(paths []string) error {
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// EnsureStorageDirs creates storageA/B/C under baseDir.
func EnsureStorageDirs(baseDir string) error {
	for _, dir := range []string{pcs.StorageA, pcs.StorageB, pcs.StorageC} {
		if err := os.MkdirAll(filepath.Join(baseDir, dir), 0o755); err != nil {
			return err
		}
	}
	return nil
}

// ReconstructedFileName returns the default decode output name for baseName.
func ReconstructedFileName(baseName string) string {
	ext := filepath.Ext(baseName)
	if ext == "" {
		return baseName + "_reconstructed"
	}
	name := baseName[:len(baseName)-len(ext)]
	return name + "_reconstructed" + ext
}

// ReconstructedButCorruptFileName returns the output name when verification fails.
func ReconstructedButCorruptFileName(baseName string) string {
	ext := filepath.Ext(baseName)
	if ext == "" {
		return baseName + "_reconstructed_but_corrupt"
	}
	name := baseName[:len(baseName)-len(ext)]
	return name + "_reconstructed_but_corrupt" + ext
}
