#!/bin/sh
#go get github.com/elazarl/go-bindata-assetfs/...
[ -d assembly-assetfs ] || mkdir assembly-assetfs
go-bindata-assetfs -pkg assemblyAssetfs static/*
mv bindata_assetfs.go assembly-assetfs/assemblyAssetfs.go
