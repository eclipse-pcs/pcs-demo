// Copyright (c) 2026 dido GmbH and others.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// https://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eclipse-pcs/pcs-demo/internal/store"
)

func main() {
	yes := flag.Bool("y", false, "skip prompts: overwrite output file; on fingerprint mismatch abort without prompting")
	flag.BoolVar(yes, "yes", false, "skip prompts: overwrite output file; on fingerprint mismatch abort without prompting")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: pcs-decode [-y] <file>\n")
		os.Exit(1)
	}

	baseName := filepath.Base(args[0])
	baseDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "pcs-decode: get working directory: %v\n", err)
		os.Exit(1)
	}

	outputName := store.ReconstructedFileName(baseName)
	outputPath := filepath.Join(baseDir, outputName)
	outputExists := fileExists(outputPath)
	if outputExists && !*yes {
		fmt.Fprintf(os.Stderr, "The following output file already exists:\n  %s\n", outputPath)
		fmt.Fprint(os.Stderr, "Delete existing output file? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		answer, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "pcs-decode: read confirmation: %v\n", err)
			os.Exit(1)
		}
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Fprintln(os.Stderr, "pcs-decode: aborted")
			os.Exit(1)
		}
	}

	inv, err := store.ScanInventory(baseDir, baseName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pcs-decode: %v\n", err)
		os.Exit(1)
	}

	for _, storage := range inv.MissingStorages {
		fmt.Fprintf(os.Stderr, "pcs-decode: missing storage folder: %s\n", storage)
	}

	for _, kind := range inv.MissingCoreParticles() {
		relPath := store.ParticleRelPath(baseName, kind)
		fmt.Fprintf(os.Stderr, "pcs-decode: missing particle: %s (%s)\n", kind, relPath)
	}

	if inv.NeedsParityRecovery() {
		fmt.Fprintf(os.Stderr, "pcs-decode: reconstructing missing particles using parity particles for %s\n", baseName)
	}

	result, err := store.DecodeFromStorage(baseDir, baseName, inv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pcs-decode: %v\n", err)
		os.Exit(1)
	}

	for _, relPath := range result.CRCParityRecoveries {
		fmt.Fprintf(os.Stderr, "pcs-decode: cross-CRC mismatch: reconstructed %s using parity particles\n", relPath)
	}

	if !result.FingerprintValid {
		fmt.Fprintln(os.Stderr, "pcs-decode: decoded secret could not be validated (fingerprint mismatch)")
		if *yes {
			fmt.Fprintln(os.Stderr, "pcs-decode: aborted")
			os.Exit(1)
		}
		switch promptValidationFailure() {
		case 'b', 'B':
			fmt.Fprintln(os.Stderr, "pcs-decode: saving result for inspection")
		default:
			fmt.Fprintln(os.Stderr, "pcs-decode: aborted")
			os.Exit(1)
		}
	}

	writeName := outputName
	writePath := outputPath
	if !result.FingerprintValid {
		writeName = store.ReconstructedButCorruptFileName(baseName)
		writePath = filepath.Join(baseDir, writeName)
	}

	writeExists := fileExists(writePath)
	if writeExists && !*yes && writePath != outputPath {
		fmt.Fprintf(os.Stderr, "The following output file already exists:\n  %s\n", writePath)
		fmt.Fprint(os.Stderr, "Delete existing output file? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		answer, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "pcs-decode: read confirmation: %v\n", err)
			os.Exit(1)
		}
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Fprintln(os.Stderr, "pcs-decode: aborted")
			os.Exit(1)
		}
	}

	if writeExists {
		if err := os.Remove(writePath); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "pcs-decode: delete output file: %v\n", err)
			os.Exit(1)
		}
	}

	if err := os.WriteFile(writePath, result.Secret, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "pcs-decode: write output file: %v\n", err)
		os.Exit(1)
	}

	if result.FingerprintValid {
		fmt.Fprintf(os.Stderr, "pcs-decode: successfully decoded %s\n", writeName)
	} else {
		fmt.Fprintf(os.Stderr, "pcs-decode: wrote %s (validation failed)\n", writeName)
	}
}

func promptValidationFailure() rune {
	fmt.Fprint(os.Stderr, "[a] Abort  [b] Save result for inspection: ")
	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return 'a'
	}
	answer = strings.TrimSpace(answer)
	if len(answer) == 0 {
		return 'a'
	}
	return rune(answer[0])
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
