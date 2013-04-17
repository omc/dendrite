#!/bin/bash
# set -ex
cd `dirname $0`
rm -rf dist
mkdir -p dist

GOROOT=~/workspace/golang-trunk
PATH=~/workspace/golang-trunk/bin:$PATH
GOPATH=/tmp/gp
rm -rf $GOPATH
mkdir -p $GOPATH/src $GOPATH/bin $GOPATH/pkg
source ~/workspace/golang-crosscompile/crosscompile.bash

go-all get
go-build-all dendrite.go
git rev-parse HEAD > REVISION
INCLUDES="cookbook LICENSE Readme.md tutorial.md VERSION REVISION"
VERSION=`cat VERSION`
ROOT="https://s3.amazonaws.com/dendrite-binaries/"

touch downloads.md
echo "## $VERSION" > tmp.md
echo >> tmp.md

for BINARY in `find . -type f -name "dendrite-*" -maxdepth 1`
do
  NAME=`echo $BINARY | sed 's/dendrite-//' | xargs basename`
  DIST="dist/$NAME/$VERSION"
  mkdir -p $DIST
  
  TAGGED_NAME=`basename ${BINARY}-$VERSION`

  rm -rf $TAGGED_NAME
  mkdir -p $TAGGED_NAME
  mv $BINARY $TAGGED_NAME/dendrite
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

  WEB="$ROOT$NAME/$VERSION/"
  rm -rf $TAGGED_NAME
  echo "* $NAME -- [tar.gz]($WEB$GZ) [md5]($WEB$GZ.md5) | [tar.bz2]($WEB$BZ) [md5]($WEB$BZ.md5) | [zip]($WEB$ZIP) [md5]($WEB$ZIP.md5)" >> tmp.md
  
done

echo >> tmp.md
echo >> tmp.md

cp downloads.md tmp2.md
cat tmp.md tmp2.md > downloads.md
rm tmp*.md
markdown downloads.md > dist/index.html

