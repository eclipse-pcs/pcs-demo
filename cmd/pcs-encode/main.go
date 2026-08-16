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
	yes := flag.Bool("y", false, "delete existing particle files without prompting")
	flag.BoolVar(yes, "yes", false, "delete existing particle files without prompting")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: pcs-encode [-y] <file>\n")
		os.Exit(1)
	}

	inputPath := args[0]
	secret, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pcs-encode: read input file: %v\n", err)
		os.Exit(1)
	}

	baseName := filepath.Base(inputPath)
	baseDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "pcs-encode: get working directory: %v\n", err)
		os.Exit(1)
	}

	existing := store.ExistingParticleFiles(baseDir, baseName)
	if len(existing) > 0 && !*yes {
		fmt.Fprintln(os.Stderr, "The following particle files already exist:")
		for _, path := range existing {
			fmt.Fprintf(os.Stderr, "  %s\n", path)
		}
		fmt.Fprint(os.Stderr, "Delete existing particle files? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		answer, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "pcs-encode: read confirmation: %v\n", err)
			os.Exit(1)
		}
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Fprintln(os.Stderr, "pcs-encode: aborted")
			os.Exit(1)
		}
	}

	if len(existing) > 0 {
		if err := store.DeleteParticleFiles(existing); err != nil {
			fmt.Fprintf(os.Stderr, "pcs-encode: delete existing files: %v\n", err)
			os.Exit(1)
		}
	}

	if err := store.SaveEncoded(baseDir, baseName, secret); err != nil {
		fmt.Fprintf(os.Stderr, "pcs-encode: save particles: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "pcs-encode: successfully encoded %s\n", baseName)
}
