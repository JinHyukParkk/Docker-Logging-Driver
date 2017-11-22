FROM alpine

RUN mkdir -p /run/docker/plugins /var/log/de-docker-logging-plugin

COPY de-docker-logging-plugin de-docker-logging-plugin

ARG git_commit=unknown
ARG version="2.9.0"
ARG descriptive_version=unknown

LABEL org.cyverse.git-ref="$git_commit"
LABEL org.cyverse.version="$version"
LABEL org.cyverse.descriptive-version="$descriptive_version"
LABEL org.label-schema.vcs-ref="$git_commit"
LABEL org.label-schema.vcs-url="https://github.com/cyverse-de/de-docker-logging-plugin"
LABEL org.label-schema.version="$descriptive_version"

CMD ["/de-docker-logging-plugin"]
