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
SH_DB_ENDPOINT=`cat helm/ssaw/secrets/SH_DB_ENDPOINT.$1.secret`
GITDM_SYNC_URL=`cat helm/ssaw/secrets/GITDM_SYNC_URL.$1.secret`
NOTIF_SVC_URL=`cat helm/ssaw/secrets/NOTIF_SVC_URL.$1.secret`
ORG_SVC_URL=`cat helm/ssaw/secrets/ORG_SVC_URL.$1.secret`
USER_SVC_URL=`cat helm/ssaw/secrets/USER_SVC_URL.$1.secret`
AFF_API_URL=`cat helm/ssaw/secrets/AFF_API_URL.$1.secret`
AWS_REGION=`cat helm/ssaw/secrets/AWS_REGION.$1.secret`
AWS_KEY=`cat helm/ssaw/secrets/AWS_KEY.$1.secret`
AWS_SECRET=`cat helm/ssaw/secrets/AWS_SECRET.$1.secret`
AWS_TOPIC=`cat helm/ssaw/secrets/AWS_TOPIC.$1.secret`
AUTH0_URL=`cat helm/ssaw/secrets/AUTH0_URL.$1.secret`
AUTH0_AUDIENCE=`cat helm/ssaw/secrets/AUTH0_AUDIENCE.$1.secret`
AUTH0_CLIENT_ID=`cat helm/ssaw/secrets/AUTH0_CLIENT_ID.$1.secret`
AUTH0_CLIENT_SECRET=`cat helm/ssaw/secrets/AUTH0_CLIENT_SECRET.$1.secret`
BEARER_TOKEN=`cat helm/ssaw/secrets/BEARER_TOKEN.$1.secret`
docker run -p 16060:6060 -e "BEARER_TOKEN=${BEARER_TOKEN}" -e "SH_DB_ENDPOINT=${SH_DB_ENDPOINT}" -e "GITDM_SYNC_URL=${GITDM_SYNC_URL}" -e "NOTIF_SVC_URL=${NOTIF_SVC_URL}" -e "ORG_SVC_URL=${ORG_SVC_URL}" -e "USER_SVC_URL=${USER_SVC_URL}" -e "AFF_API_URL=${AFF_API_URL}" -e "AWS_REGION=${AWS_REGION}" -e "AWS_KEY=${AWS_KEY}" -e "AWS_SECRET=${AWS_SECRET}" -e "AWS_TOPIC=${AWS_TOPIC}" -e "AUTH0_URL=${AUTH0_URL}" -e "AUTH0_AUDIENCE=${AUTH0_AUDIENCE}" -e "AUTH0_CLIENT_ID=${AUTH0_CLIENT_ID}" -e "AUTH0_CLIENT_SECRET=${AUTH0_CLIENT_SECRET}" -it "${DOCKER_USER}/ssaw" "/bin/sh"
