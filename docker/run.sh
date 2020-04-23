#!/bin/bash
if [ -z "$1" ]
then
  echo "$0: you need to specify env as a 1st argument: prod|test"
  exit 1
fi
if [ -z "${DOCKER_USER}" ]
then
  echo "$0: you need to set docker user via DOCKER_USER=username"
  exit 1
fi
SH_DB_ENDPOINT=`cat helm/sfdc-sh-sync/sfdc-sh-sync/secrets/SH_DB_ENDPOINT.$1.secret`
AWS_REGION=`cat helm/sfdc-sh-sync/sfdc-sh-sync/secrets/AWS_REGION.$1.secret`
AWS_KEY=`cat helm/sfdc-sh-sync/sfdc-sh-sync/secrets/AWS_KEY.$1.secret`
AWS_SECRET=`cat helm/sfdc-sh-sync/sfdc-sh-sync/secrets/AWS_SECRET.$1.secret`
AWS_TOPIC=`cat helm/sfdc-sh-sync/sfdc-sh-sync/secrets/AWS_TOPIC.$1.secret`
LF_AUTH=`cat helm/sfdc-sh-sync/sfdc-sh-sync/secrets/LF_AUTH.$1.secret`
docker run -p 16060:6060 -e "SH_DB_ENDPOINT=${SH_DB_ENDPOINT}" -e "AWS_REGION=${AWS_REGION}" -e "AWS_KEY=${AWS_KEY}" -e "AWS_SECRET=${AWS_SECRET}" -e "AWS_TOPIC=${AWS_TOPIC}" -e "LF_AUTH=${LF_AUTH}" -it "${DOCKER_USER}/lf-sfdc-sh-sync" "/usr/bin/sfdc-sh-sync"
