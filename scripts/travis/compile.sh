#!/bin/bash
LDFLAGS="$(go run scripts/build/gen-ldflags.go)"
echo "LDFLAGS build release ID:>"
echo $LDFLAGS

# In the main.go :>
#go:generate echo Go Generate!
#go:generate ./scripts/build/bindata.sh
#go:generate ./scripts/build/bindata-assetfs.sh

go generate

echo "compile each binary"

if [ -z "$TRAVIS_BUILD_DIR" ]
then
	TRAVIS_BUILD_DIR=$PWD
fi

echo $TRAVIS_BUILD_DIR

for GOOS in darwin linux windows; do
  for GOARCH in 386 amd64; do
    echo "Building $GOOS-$GOARCH"
    export GOOS=$GOOS
    export GOARCH=$GOARCH
    export CGO_ENABLED=0
    go build -o $TRAVIS_BUILD_DIR/bin/goBloomsky-$GOOS-$GOARCH --ldflags "${LDFLAGS}"
  done
done
mv bin/goBloomsky-darwin-386 bin/goBloomsky-darwin-386.bin
mv bin/goBloomsky-darwin-amd64 bin/goBloomsky-darwin-amd64.bin
mv bin/goBloomsky-linux-386 bin/goBloomsky-linux-386.bin
mv bin/goBloomsky-linux-amd64 bin/goBloomsky-linux-amd64.bin
mv bin/goBloomsky-windows-386 bin/goBloomsky-windows-386.exe
mv bin/goBloomsky-windows-amd64 bin/goBloomsky-windows-amd64.exe
