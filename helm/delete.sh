#!/bin/bash
# NS=ssaw - set namespace name, default ssaw
helm=helm
denv=test
if [ -z "$1" ]
then
  echo "$0: you should specify env: test, prod, using default helm"
else
  helm="${1}h.sh"
  denv="${1}"
fi
if [ -z "$NS" ]
then
  NS=ssaw
fi
change_namespace.sh $1 "$NS"
$helm delete "$NS"
change_namespace.sh $1 default
$helm delete "${NS}-namespace"
