FROM alpine
RUN apk --update add git less openssh && rm -rf /var/lib/apt/lists/* && rm /var/cache/apk/*
COPY ssaw /usr/bin/
CMD ssaw
