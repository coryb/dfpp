package main

import (
	"fmt"
	"github.com/andrew-d/go-termutil"
	"github.com/coryb/dfpp"
	"github.com/op/go-logging"
	"os"
)

var log = logging.MustGetLogger("dfpp")
var format = "%{color}%{time:2006-01-02T15:04:05.000Z07:00} %{level:-5s} [%{shortfile}]%{color:reset} %{message}"

func main() {
	if len(os.Args) == 1 && termutil.Isatty(os.Stdin.Fd()) {
		Usage()
		os.Exit(0)
	}

	var err error
	inputFile := os.Stdin
	if len(os.Args) > 1 {
		inputFile, err = os.Open(os.Args[1])
		if err != nil {
			fmt.Printf("Failed to read %s: %s\n", os.Args[1], err)
			os.Exit(1)
		}
	}

	logBackend := logging.NewLogBackend(os.Stderr, "", 0)
	logging.SetBackend(
		logging.NewBackendFormatter(
			logBackend,
			logging.MustStringFormatter(format),
		),
	)
	logging.SetLevel(logging.DEBUG, "")

	pp := dfpp.NewDfpp()
	pp.ProcessDockerfile(inputFile)
}

func Usage() {
	fmt.Print(`
NAME
    dfpp - Dockerfile preprocessor

SYNOPSIS
        $ dfpp Dockerfile.pre > Dockerfile

        # Dockerfile Syntax:
        INCLUDE ./Dockerfile.inc
        INCLUDE http://path/to/Dockerfile.inc
        INCLUDE ./Dockerfile.inc http://path/to/Dockerfile.inc

        INCLUDE MERGE a.inc b.inc

        # include only RUN instructions
        INCLUDE RUN a.inc b.inc

        # include only RUN and ENV instructions
        INCLUDE RUN ENV a.inc b.inc
   
        # include only RUN and ENV instructions but merge them
        INCLUDE MERGE RUN ENV a.inc b.inc

        # exclude FROM instructions
        INCLUDE -FROM a.inc b.inc

DESCRIPTION
    "dfpp" was written to allow simple pre-processing of Dockerfiles to add
    capabilities currently unsupported by docker build.

INSTRUCTIONS
  INCLUDE [MERGE] [FILTERS] [file|uri] ...
    This will inline a file or uri into the Dockerfile being generated.

   MERGE
    When including multiple Dockerfile snippets this will attempt to merge
    common instructions. Currently only ENV, LABEL and RUN are merged,
    otherwise multiple instructions will be repeated. RUN instructions are
    merged with "&&" while other instructions are merged with a space.

   FILTERS
   [-]ADD
    Include or Exclude ADD instructions from inlined Dockerfile snippets

   [-]CMD
    Include or Exclude CMD instructions from inlined Dockerfile snippets

   [-]COPY
    Include or Exclude COPY instructions from inlined Dockerfile snippets

   [-]ENTRYPOINT
    Include or Exclude ENTRYPOINT instructions from inlined Dockerfile
    snippets

   [-]ENV
    Include or Exclude ENV instructions from inlined Dockerfile snippets

   [-]EXPOSE
    Include or Exclude EXPOSE instructions from inlined Dockerfile snippets

   [-]FROM
    Include or Exclude FROM instructions from inlined Dockerfile snippets

   [-]INCLUDE
    Include or Exclude INCLUDE instructions from inlined Dockerfile snippets

   [-]LABEL
    Include or Exclude LABEL instructions from inlined Dockerfile snippets

   [-]MAINTAINER
    Include or Exclude MAINTAINER instructions from inlined Dockerfile
    snippets

   [-]ONBUILD
    Include or Exclude ONBUILD instructions from inlined Dockerfile snippets

   [-]RUN
    Include or Exclude RUN instructions from inlined Dockerfile snippets

   [-]USER
    Include or Exclude USER instructions from inlined Dockerfile snippets

   [-]VOLUME
    Include or Exclude VOLUME instructions from inlined Dockerfile snippets

   [-]WORKDIR
    Include or Exclude WORKDIR instructions from inlined Dockerfile snippets

AUTHOR
    2015, Cory Bennett <github@corybennett.org>

SOURCE
    The Source is available at github:
    https://github.com/coryb/dfpp

COPYRIGHT and LICENSE
    Copyright (c) 2015 Netflix Inc. All rights reserved. The copyrights to
    the contents of this file are licensed under the Apache License, Version 2.0
`)
}
