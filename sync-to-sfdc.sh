#!/bin/bash
if [ -z "${SYNC_URL}" ]
then
  SYNC_URL='localhost:6060'
fi
(curl -s "${SYNC_URL}/sync-to-sfdc" |& tee output.txt | grep 'SYNC_OK') || ( cat output.txt; exit 1 )

