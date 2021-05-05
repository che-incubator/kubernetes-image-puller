//
// Copyright (c) 2021 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//

package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

func displayUsage(out io.Writer) int {
	fmt.Fprintf(out, "Usage: %s <duration>\n", os.Args[0])
	fmt.Fprintln(out, "See https://pkg.go.dev/time#ParseDuration")
	return -1
}

func main() {
	exitCode := entryPoint(os.Stderr)
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

func entryPoint(out io.Writer) int {
	if len(os.Args) != 2 {
		return displayUsage(out)
	}
	duration, err := time.ParseDuration(os.Args[1])
	if err != nil {
		fmt.Fprintf(out, "Invalid duration: %s\n", err)
		return displayUsage(out)
	}
	time.Sleep(duration)
	return 0
}
