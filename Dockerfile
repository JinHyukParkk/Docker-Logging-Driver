FROM localhost:5000/dockerLoggingDriver
MAINTAINER https://github.com/orgs/NAVER-CAMPUS-HACKDAY/teams/dockerloggingdriver

COPY . /go/src/github.com/JinHyukParkk/docker-log-driver
RUN cd /go/src/github.com/JinHyukParkk/docker-log-driver && go get && go build --ldflags '-extldflags "-static"' -o /usr/bin/dockerloggingdriver
RUN rm -rf /go /usr/local /usr/lib /usr/share
