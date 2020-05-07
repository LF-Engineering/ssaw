# ssaw

Self Serve Affiliations Workflow

# Running locally

- To compile locally use: `make`.
- To run local sync service: `./serve.sh test|prod` (it will use prod/test `*.secret` files for all environment variables.
- To do request to local service (check service used for reacting `sync-to-sfdc` request): `./sync-to-sfdc.sh`.

# DB triggers

- To test DB triggers on a local database execute: `mysql ... < ./sql/sh_sync_test.sql`.
- Use `` SH_DB_ENDPOINT="`cat ./helm/ssaw/secrets/SH_DB_ENDPOINT.secret.example`" ./serve.sh `` to use local SortingHat DB running from docker (using `LF-Engineering/dev-analytics-affiliation` repo).
- After starting against a local docker MariaDB run: `SH_USR=sortinghat SH_PASS=pwd SH_DB=sortinghat ./sh/mariadb_sortinghat_shell.sh`, then confirm that company with `src=exampleOrigin` was created.
- Execute: `select * from organizations where src = 'exampleOrigin'` (there will be company with name `exampleCompany`).

# Docker

- Build docker image: `DOCKER_USER=... docker/build_image.sh`.
- Run it: `DOCKER_USER=... docker/run.sh`. It will serve on 16060 instead of 6060 port.
- Shell into the container: `DOCKER_USER=... docker/shell.sh`.
- Test request, `SYNC_URL` must be provided to specify non-default 16060 port: `SYNC_URL='127.0.0.1:16060' ./sync-to-sfdc.sh`.

# Kubernetes/Helm

To deploy on Kubernetes

- Go to `helm/`, run (LF real world example): `./setup.sh prod`.
- Eventually adjust Helm chart to your needs, including `setup.sh` and `delete.sh` shell scripts.
- Run from repository root directory (test env): `` SYNC_URL="`cat helm/ssaw/secrets/SYNC_URL.test.secret`" ./sync-to-sfdc.sh ``.
- Run from repository root directory (prod env): `` SYNC_URL="`cat helm/ssaw/secrets/SYNC_URL.prod.secret`" ./sync-to-sfdc.sh ``.

# GitHub actions

- Add your sync URL (for example AWS ELB of ssaw service stored in `helm/ssaw/secrets/SYNC_URL.prod.secret`) in GitHub repository (Settings -> Secrets -> New secret: `SYNC_URL`).
- Configre actions in `.github/workflows/`, for example: `.github/workflows/sync-to-sfdc.yaml`.
