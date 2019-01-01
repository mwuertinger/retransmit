#!/bin/sh

set -e


VERSION="$(git rev-list -1 HEAD)"
if [ -n "$(git status --porcelain)" ]
then
	VERSION="${VERSION} (dirty)"
fi

go build -ldflags "-X 'main.buildVersion=$VERSION'" .
