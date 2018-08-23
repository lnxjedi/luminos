#!/bin/bash

# publish.sh - copy the install archive to a distribution point

VERSION=$(grep "const Version =" main.go)
VERSION=${VERSION##* }
VERSION=${VERSION//\"/}

eval `go env`
BRANCH=$(git rev-parse --abbrev-ref HEAD)
COMMIT=$(git rev-parse HEAD)
if [ $BRANCH = "master" ] || [[ $BRANCH = *-release ]]
then
    RELEASE=$VERSION
    if [[ $RELEASE = *-snapshot ]]
    then
        PRERELEASE="--prerelease"
        UPDATE="--update"
    fi
else
    RELEASE="$BRANCH-snapshot"
    PRERELEASE="--prerelease"
    UPDATE="--update"
fi

github-release $RELEASE luminos-* --github-repository lnxjedi/luminos --commit $COMMIT --target $BRANCH $PRERELEASE $UPDATE
