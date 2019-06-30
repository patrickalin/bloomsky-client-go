#!/bin/sh
#go get -u github.com/jteeuwen/go-bindata/...
echo execute binddata.sh :> embedding binary data into a go program
go-bindata -pkg assembly -o assembly/assembly.go -ignore=lang/en-us.untranslated.json -ignore=lang/fr.untranslated.json -ignore=lang/merge.sh -ignore=lang/noTranslation.sh  ./tmpl/* lang/* test/*
ls -la assembly/assembly.go
echo end binddata.sh