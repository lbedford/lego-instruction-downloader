#!/bin/bash

DIR=~/Documents/LEGO
grep -v ^# $DIR/ids.txt  | grep -v ^$ | while read ID ; do 
  if [ ! -d $DIR/$ID/ ] ; then
    go run lego-instruction-downloader.go --id $ID
    if [ ! -d $DIR/$ID/ ] ; then
      echo "Failed to download $ID" >> $DIR/failures.txt
    fi
  else
    echo "$ID is already downloaded"
  fi
done
