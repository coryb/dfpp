# NAME

dfpp - Dockerfile preprocessor

# SYNOPSIS

     $ dfpp Dockerfile.pre > Dockerfile

     # without installing run via docker:
     $ docker run --rm -i -v $(pwd):/root:ro coryb/dfpp Dockerfile.pre > Dockerfile

     # or alias dfpp to docker call
     $ alias dfpp='docker run --rm -i -v $(pwd):/root:ro coryb/dfpp'
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

# DESCRIPTION

`dfpp` was written to allow simple pre-processing of Dockerfiles to add
capabilities currently unsupported by docker build.

The INCLUDE instruction is primarily useful to make Dockerfiles composible if you have wide variation in images being produced but you dont want to copy & paste common instructions.  For example, if you are using docker for build isolation where you have some projects that are c++ and some that are java.  Typically you would create a specific `cpp-build` image with all the c++ build dependencies and a `java-build` image with all the java build dependencies.  Then you find yourself needing to build something that requires both c++ AND java, so you create a `cpp-java-build` image . Now you are maintaining the c++ build deps in the `cpp-build` and `cpp-java-build` image.  INCLUDE allows you to maintain the build deps in a single location and simply reference that single location for both `cpp-build` and `cpp-java-build` images.

# INSTRUCTIONS

## INCLUDE \[MERGE\] \[FILTERS\] \[file|uri\] ...

This will inline a file or uri into the Dockerfile being generated.

### MERGE

When including multiple Dockerfile snippets this will attempt to merge common instructions.  Currently only 
ENV, LABEL and RUN are merged, otherwise multiple instructions will be repeated.  RUN instructions are merged with "&&" while other
instructions are merged with a space.

### FILTERS

#### \[-\]ADD

Include or Exclude ADD instructions from inlined Dockerfile snippets

#### \[-\]CMD

Include or Exclude CMD instructions from inlined Dockerfile snippets

#### \[-\]COPY

Include or Exclude COPY instructions from inlined Dockerfile snippets

#### \[-\]ENTRYPOINT

Include or Exclude ENTRYPOINT instructions from inlined Dockerfile snippets

#### \[-\]ENV

Include or Exclude ENV instructions from inlined Dockerfile snippets

#### \[-\]EXPOSE

Include or Exclude EXPOSE instructions from inlined Dockerfile snippets

#### \[-\]FROM

Include or Exclude FROM instructions from inlined Dockerfile snippets

#### \[-\]INCLUDE

Include or Exclude INCLUDE instructions from inlined Dockerfile snippets

#### \[-\]LABEL

Include or Exclude LABEL instructions from inlined Dockerfile snippets

#### \[-\]MAINTAINER

Include or Exclude MAINTAINER instructions from inlined Dockerfile snippets

#### \[-\]ONBUILD

Include or Exclude ONBUILD instructions from inlined Dockerfile snippets

#### \[-\]RUN

Include or Exclude RUN instructions from inlined Dockerfile snippets

#### \[-\]USER

Include or Exclude USER instructions from inlined Dockerfile snippets

#### \[-\]VOLUME

Include or Exclude VOLUME instructions from inlined Dockerfile snippets

#### \[-\]WORKDIR

Include or Exclude WORKDIR instructions from inlined Dockerfile snippets

# AUTHOR

2015, Cory Bennett <github@corybennett.org>

# SOURCE

The Source is available at github: https://github.com/coryb/dfpp

# COPYRIGHT and LICENSE

Copyright (c) 2015 Netflix Inc. All rights reserved. The copyrights to the contents of this file are licensed under the Apache License, Version 2.0.
