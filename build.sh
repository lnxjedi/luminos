#!/bin/bash -e

# mkdist.sh - create a distributable .zip file

usage(){
	cat <<EOF
Usage: build.sh (linux|darwin|windows)

Generate an executable for the given platform, or all platforms if none explicitly given.
EOF
	exit 0
}

if [ "$1" = "-h" -o "$1" = "--help" ]
then
	usage
fi

git status | grep -qE "nothing to commit, working directory|tree clean" || { echo "Your working directory isn't clean, aborting build"; exit 1; }
COMMIT=$(git rev-parse --short HEAD)

eval `go env`
PLATFORMS=${1:-linux darwin windows}
for BUILDOS in $PLATFORMS
do
	echo "Building luminos for $BUILDOS"
	if [ "$BUILDOS" = "windows" ]
	then
		GOOS=$BUILDOS go build -mod vendor -ldflags "-X main.Commit=$COMMIT" -o luminos-windows.exe
	elif [ "$BUILDOS" = "linux" ]
	then
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod vendor -ldflags "-X main.Commit=$COMMIT" -tags 'netgo osusergo static_build' -o luminos-linux
	else
		GOOS=$BUILDOS go build -mod vendor -ldflags "-X main.Commit=$COMMIT" -o luminos-$BUILDOS
	fi
done
