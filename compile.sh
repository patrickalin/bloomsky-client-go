#!/bin/bash
echo "compile each binary"
for GOOS in darwin linux windows; do
  for GOARCH in 386 amd64; do
    echo "Building $GOOS-$GOARCH"
    export GOOS=$GOOS
    export GOARCH=$GOARCH
    go build -o bin/goBloomsky-$GOOS-$GOARCH
  done
done
mv bin/goBloomsky-darwin-386 bin/goBloomsky-darwin-386.bin
mv bin/goBloomsky-darwin-amd64 bin/goBloomsky-darwin-amd64.bin
mv bin/goBloomsky-linux-386 bin/goBloomsky-linux-386.bin
mv bin/goBloomsky-linux-amd64 bin/goBloomsky-linux-amd64.bin
mv bin/goBloomsky-windows-386 bin/goBloomsky-windows-386.exe
mv bin/goBloomsky-windows-amd64 bin/goBloomsky-windows-amd64.exe
