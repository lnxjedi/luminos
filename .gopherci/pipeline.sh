#!/bin/bash

# pipeline.sh - trusted pipeline script for gopherci for Luminos

source $GOPHER_INSTALLDIR/lib/gopherbot_v1.sh

if [ -n "$NOTIFY_USER" ]
then
    FailTask notify $NOTIFY_USER "Luminos build failed"
    # Email the job history if it fails
    FailCommand builtin-history "send history $GOPHER_JOB_NAME:$GOPHER_NAMESPACE_EXTENDED/$GOPHERCI_BRANCH $GOPHER_RUN_INDEX to user $NOTIFY_USER"
fi

# Install required tools
AddTask exec ./.gopherci/tools.sh

# Publish coverage results
#AddTask exec goveralls -coverprofile=coverage.out -service=circle-ci -repotoken=$COVERALLS_TOKEN

# Do a full build for all platforms
AddTask exec ./build.sh

# Publish archives to github
AddTask exec ./.gopherci/publish.sh

# Trigger Docker build
AddTask exec ./.gopherci/dockercloud.sh

# Notify of success
if [ -n "$NOTIFY_USER" ]
then
    AddTask notify $NOTIFY_USER "Successfully built and released latest Luminos"
fi
