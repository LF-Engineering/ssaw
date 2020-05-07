#!/bin/bash
if [ -z "$1" ]
then
  echo "$0: you need to specify env as a 1st argument: prod|test"
  exit 1
fi
if [ -z "${SH_DB_ENDPOINT}" ]
then
  export SH_DB_ENDPOINT=`cat helm/ssaw/secrets/SH_DB_ENDPOINT.$1.secret`
fi
if [ -z "${ORG_SVC_URL}" ]
then
  export ORG_SVC_URL=`cat helm/ssaw/secrets/ORG_SVC_URL.$1.secret`
fi
if [ -z "${USER_SVC_URL}" ]
then
  export USER_SVC_URL=`cat helm/ssaw/secrets/USER_SVC_URL.$1.secret`
fi
if [ -z "${AWS_REGION}" ]
then
  export AWS_REGION=`cat helm/ssaw/secrets/AWS_REGION.$1.secret`
fi
if [ -z "${AWS_KEY}" ]
then
  export AWS_KEY=`cat helm/ssaw/secrets/AWS_KEY.$1.secret`
fi
if [ -z "${AWS_SECRET}" ]
then
  export AWS_SECRET=`cat helm/ssaw/secrets/AWS_SECRET.$1.secret`
fi
if [ -z "${AWS_TOPIC}" ]
then
  export AWS_TOPIC=`cat helm/ssaw/secrets/AWS_TOPIC.$1.secret`
fi
if [ -z "${LF_AUTH}" ]
then
  export LF_AUTH=`cat helm/ssaw/secrets/LF_AUTH.$1.secret`
fi
./ssaw
