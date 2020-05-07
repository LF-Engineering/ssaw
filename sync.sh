#!/bin/bash
if [ -z "$1" ]
then
  echo "$0: you need to specify origin as a 1st arument"
  exit 1
fi
if [ -z "${SYNC_URL}" ]
then
  SYNC_URL='localhost:6060'
fi
(curl -s "${SYNC_URL}/sync/$1" |& tee output.txt | grep 'SYNC_OK') || ( cat output.txt; exit 1 )

