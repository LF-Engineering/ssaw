#!/bin/bash
# DRY=1 - dry run mode
# NS=ssaw - set namespace name, default ssaw
helm=helm
denv=test
if [ -z "$1" ]
then
  echo "$0: you should env: test, prod, using default helm"
else
  helm="${1}h.sh"
  denv="${1}"
fi
if [ -z "$NS" ]
then
  NS=ssaw
fi
if [ -z "$DRY" ]
then
  $helm install "${NS}-namespace" ./ssaw --set "namespace=$NS,skipSecrets=1,skipSync=1"
  change_namespace.sh $1 "$NS"
  $helm install "$NS" ./ssaw --set "namespace=$NS,deployEnv=$denv,skipNamespace=1"
  change_namespace.sh $1 default
else
  echo "Dry run mode"
  change_namespace.sh $1 "$NS"
  $helm install --debug --dry-run --generate-name ./ssaw --set "namespace=$NS,deployEnv=$denv,dryRun=1"
  change_namespace.sh $1 default
fi
