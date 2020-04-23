# sfdc-sh-sync
Bidirectional sync between SFDC service and SortingHat database

# Running locally

- To compile locally use: `make`.
- To run local sync service: `./serve.sh`.
- To do request to local service (check service used for reacting `sync-to-sfdc` request): `PR=pr_number ./sync-to-sfdc.sh`.

# Docker

- Build docker image: `DOCKER_USER=... docker/build_image.sh`.
- Run it: `DOCKER_USER=... docker/run.sh`. It will serve on 16060 instead of 6060 port.
- Shell into the container: `DOCKER_USER=... docker/shell.sh`.
- Test request, `SYNC_URL` must be provided to specify non-default 16060 port: `SYNC_URL='127.0.0.1:16060' ./sync-to-sfdc.sh`.
