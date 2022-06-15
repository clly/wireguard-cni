#!/usr/bin/env bash

set -eou pipefail
# shellcheck disable=SC2154
trap 's=${?:-1}; echo >&2 "$0: Error on line "$LINENO": $BASH_COMMAND"; exit $s' ERR


gitCommit=$(git rev-parse HEAD)
gitBranch=$(git rev-parse --abbrev-ref HEAD)
gitTagRef=$(git name-rev --name-only --tags "${gitCommit}")
gitTag=${gitTagRef#tags/}
case $(uname -m) in
    i386)   architecture="386" ;;
    i686)   architecture="386" ;;
    x86_64) architecture="amd64" ;;
	aarch64) architecture="arm64" ;;
esac

if grep -q -s -P "^ID=\S+$" /etc/os-release; then
    os=$(grep -P "^ID=\S+$" /etc/os-release|cut -f2 -d=)
else
    os=""
fi


echo "os: $os"
echo "gitCommit: $gitCommit"
echo "gitTag: $gitTag"
echo "architecture: $architecture"
echo "gitBranch: $gitBranch"
