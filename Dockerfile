FROM alpine

RUN mkdir -p /run/docker/plugins /var/log/LoggingDriverTest

COPY LoggingDriverTest /LoggingDriverTest

ARG git_commit=unknown
ARG version="2.9.0"
ARG descriptive_version=unknown

LABEL org.cyverse.git-ref="$git_commit"
LABEL org.cyverse.version="$version"
LABEL org.cyverse.descriptive-version="$descriptive_version"
LABEL org.label-schema.vcs-ref="$git_commit"
LABEL org.label-schema.vcs-url="https://github.com/JinHyukParkk/LoggingDriverTest"
LABEL org.label-schema.version="$descriptive_version"

CMD ["/LoggingDriverTest"]
