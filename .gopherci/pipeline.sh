#!/bin/bash

# pipeline.sh - trusted pipeline script for gopherci for Luminos

source $GOPHER_INSTALLDIR/lib/gopherbot_v1.sh

if [ -n "$NOTIFY_USER" ]
then
    FailTask notify $NOTIFY_USER "Luminos build failed"
fi

# Get dependencies
AddTask localexec go get -v -t -d ./...

# Install required tools
AddTask localexec ./.gopherci/tools.sh

# Publish coverage results
#AddTask localexec goveralls -coverprofile=coverage.out -service=circle-ci -repotoken=$COVERALLS_TOKEN

# Do a full build for all platforms
AddTask localexec ./build.sh

# Publish archives to github
AddTask localexec ./.gopherci/publish.sh

# Notify of success
if [ -n "$NOTIFY_USER" ]
then
    AddTask notify $NOTIFY_USER "Successfully built and released latest Luminos"
fi
