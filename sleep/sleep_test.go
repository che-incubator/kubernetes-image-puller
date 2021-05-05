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
	"bytes"
	"flag"
	"os"
	"regexp"
	"testing"
)

func TestArguments(T *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	cases := []struct {
		Name           string
		Args           []string
		ExpectedExit   int
		ExpectedOutput string
	}{
		{"valid second duration", []string{"1s"}, 0, ""},
		{"valid ns duration", []string{"1000ns"}, 0, ""},
		{"invalid day duration", []string{"10d"}, -1, "Invalid duration: time: unknown unit .* in duration .*10d.*"},
		{"too many args", []string{"a", "b"}, -1, "Usage: too many args <duration>\nSee https://pkg.go.dev/time#ParseDuration\n"},
	}
	for _, tc := range cases {
		flag.CommandLine = flag.NewFlagSet(tc.Name, flag.ExitOnError)
		os.Args = append([]string{tc.Name}, tc.Args...)
		var buf bytes.Buffer
		actualExit := entryPoint(&buf)
		if tc.ExpectedExit != actualExit {
			T.Errorf("Wrong exit code for args: %v, expected: %v, got: %v",
				tc.Args, tc.ExpectedExit, actualExit)
		}

		actualOutput := buf.String()
		r, _ := regexp.Compile(tc.ExpectedOutput)
		if !r.MatchString(actualOutput) {
			T.Errorf("Wrong output for args: %v, expected %v, got: %v",
				tc.Args, tc.ExpectedOutput, actualOutput)
		}

		// check main entrypoint
		flag.CommandLine = flag.NewFlagSet("100ms", 0)
		os.Args = append([]string{"100ms"}, "100ms")
		main()
	}
}
