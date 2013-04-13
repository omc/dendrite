#!/bin/bash
cd `dirname $0`
rm -rf dist
mkdir -p dist

export GOROOT=~/workspace/golang-trunk
source ~/workspace/golang-crosscompile/crosscompile.bash

go-build-all dendrite.go

INCLUDES="cookbook LICENSE Readme.md VERSION"

for BINARY in `find . -type f -name "dendrite-*" -maxdepth 1`
do
  NAME=`echo $BINARY | sed 's/dendrite-//'`
  VERSION=`cat VERSION`
  DIST="dist/$NAME/$VERSION"
  mkdir -p $DIST
  
  TAGGED_NAME="${BINARY}-$VERSION"

  rm -rf $TAGGED_NAME
  mkdir -p $TAGGED_NAME
  mv $BINARY $TAGGED_NAME
  for F in $INCLUDES
  do
    cp -R $F $TAGGED_NAME
  done
  
  GZ="${TAGGED_NAME}.tar.gz"
  tar -zcvf $GZ $TAGGED_NAME
  md5sum $GZ > $DIST/$GZ.md5
  mv $GZ $DIST
  
  BZ="${TAGGED_NAME}.tar.bz2"
  tar -jcvf $BZ $TAGGED_NAME
  md5sum $BZ > $DIST/$BZ.md5
  mv $BZ $DIST
  
  ZIP="${TAGGED_NAME}.zip"
  zip -r $ZIP $TAGGED_NAME
  md5sum $ZIP > $DIST/$ZIP.md5
  mv $ZIP $DIST
  
  rm -rf $TAGGED_NAME
  
done

s3cmd sync dist/. s3://dendrite-binaries
