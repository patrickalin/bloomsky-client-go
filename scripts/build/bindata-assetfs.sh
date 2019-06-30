#!/bin/sh
#go get github.com/elazarl/go-bindata-assetfs/...
echo execute binddata-assetfs.sh Serve embedded files with net/http
[ -d assembly-assetfs ] || mkdir assembly-assetfs
go-bindata-assetfs -pkg assemblyAssetfs static/*
mv bindata.go assembly-assetfs/assemblyAssetfs.go
echo end binddata-assetfs.sh