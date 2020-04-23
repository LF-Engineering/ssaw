FROM alpine
RUN apk --update add git less openssh && rm -rf /var/lib/apt/lists/* && rm /var/cache/apk/*
COPY sfdc-sh-sync /usr/bin/
CMD sfdc-sh-sync
