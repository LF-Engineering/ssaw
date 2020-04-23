#!/bin/bash
if [ -z "$1" ]
then
  echo "$0: you need to specify env as a 1st argument: prod|test"
  exit 1
fi
if [ -z "${SH_DB_ENDPOINT}" ]
then
  export SH_DB_ENDPOINT=`cat helm/sfdc-sh-sync/secrets/SH_DB_ENDPOINT.$1.secret`
fi
if [ -z "${AWS_REGION}" ]
then
  export AWS_REGION=`cat helm/sfdc-sh-sync/secrets/AWS_REGION.$1.secret`
fi
if [ -z "${AWS_KEY}" ]
then
  export AWS_KEY=`cat helm/sfdc-sh-sync/secrets/AWS_KEY.$1.secret`
fi
if [ -z "${AWS_SECRET}" ]
then
  export AWS_SECRET=`cat helm/sfdc-sh-sync/secrets/AWS_SECRET.$1.secret`
fi
if [ -z "${AWS_TOPIC}" ]
then
  export AWS_TOPIC=`cat helm/sfdc-sh-sync/secrets/AWS_TOPIC.$1.secret`
fi
if [ -z "${LF_AUTH}" ]
then
  export LF_AUTH=`cat helm/sfdc-sh-sync/secrets/LF_AUTH.$1.secret`
fi
./sfdc-sh-sync
