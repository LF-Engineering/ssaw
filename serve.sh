#!/bin/bash
if [ -z "$1" ]
then
  echo "$0: you need to specify env as a 1st argument: prod|test"
  exit 1
fi
if [ "$1" = "local" ]
then
  env="secret.example"
else
  env="$1.secret"
fi
if [ -z "${SH_DB_ENDPOINT}" ]
then
  export SH_DB_ENDPOINT=`cat helm/ssaw/secrets/SH_DB_ENDPOINT.$env`
fi
if [ -z "${GITDM_SYNC_URL}" ]
then
  export GITDM_SYNC_URL=`cat helm/ssaw/secrets/GITDM_SYNC_URL.$env`
fi
if [ -z "${NOTIF_SVC_URL}" ]
then
  export NOTIF_SVC_URL=`cat helm/ssaw/secrets/NOTIF_SVC_URL.$env`
fi
if [ -z "${ORG_SVC_URL}" ]
then
  export ORG_SVC_URL=`cat helm/ssaw/secrets/ORG_SVC_URL.$env`
fi
if [ -z "${USER_SVC_URL}" ]
then
  export USER_SVC_URL=`cat helm/ssaw/secrets/USER_SVC_URL.$env`
fi
if [ -z "${AFF_API_URL}" ]
then
  export AFF_API_URL=`cat helm/ssaw/secrets/AFF_API_URL.$env`
fi
if [ -z "${AWS_REGION}" ]
then
  export AWS_REGION=`cat helm/ssaw/secrets/AWS_REGION.$env`
fi
if [ -z "${AWS_KEY}" ]
then
  export AWS_KEY=`cat helm/ssaw/secrets/AWS_KEY.$env`
fi
if [ -z "${AWS_SECRET}" ]
then
  export AWS_SECRET=`cat helm/ssaw/secrets/AWS_SECRET.$env`
fi
if [ -z "${AWS_TOPIC}" ]
then
  export AWS_TOPIC=`cat helm/ssaw/secrets/AWS_TOPIC.$env`
fi
if [ -z "${AUTH0_URL}" ]
then
  export AUTH0_URL=`cat helm/ssaw/secrets/AUTH0_URL.$env`
fi
if [ -z "${AUTH0_AUDIENCE}" ]
then
  export AUTH0_AUDIENCE=`cat helm/ssaw/secrets/AUTH0_AUDIENCE.$env`
fi
if [ -z "${AUTH0_CLIENT_ID}" ]
then
  export AUTH0_CLIENT_ID=`cat helm/ssaw/secrets/AUTH0_CLIENT_ID.$env`
fi
if [ -z "${AUTH0_CLIENT_SECRET}" ]
then
  export AUTH0_CLIENT_SECRET=`cat helm/ssaw/secrets/AUTH0_CLIENT_SECRET.$env`
fi
if [ -z "${BEARER_TOKEN}" ]
then
  export BEARER_TOKEN=`cat helm/ssaw/secrets/BEARER_TOKEN.$env`
fi
./ssaw
