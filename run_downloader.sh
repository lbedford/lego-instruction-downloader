#!/bin/bash

grep -v ^# ids.txt  | grep -v ^$ | while read ID ; do 
  if [ ! -d ~/Documents/LEGO/$ID/ ] ; then
    ./lego-instruction-downloader --id $ID
    if [ ! -d ~/Documents/LEGO/$ID/ ] ; then
      echo "Failed to download $ID" >> failures.txt
    fi
  else
    echo "$ID is already downloaded"
  fi
done
