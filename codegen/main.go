// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"

	gflags "github.com/jessevdk/go-flags"
)

type Flags struct {
	Git struct {
		Repo   string `long:"repo" description:"git repository url to use to fetch dataset" required:"true"`
		Branch string `long:"branch" description:"git branch to pull from" default:"gh-pages"`
	} `namespace:"git" group:"Git Options"`
	OutputPath string `long:"output" description:"directory to generate code" required:"true"`
	PkgPath    string `long:"pkg-path" description:"package location (e.g. git.example.com/user/your-pkg)" required:"true"`
}

var flags Flags

var logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

func main() {
	parser := gflags.NewParser(&flags, gflags.HelpFlag)
	_, err := parser.Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("initializing")
	var data map[string]*DataNode

	if data, err = fetchData(flags.Git.Repo, flags.Git.Branch); err != nil {
		logger.Fatalf("error fetching data: %v", err)
	}

	if err = gen(flags.OutputPath, flags.PkgPath, data); err != nil {
		logger.Fatalf("error generating code from data: %v", err)
	}
}
