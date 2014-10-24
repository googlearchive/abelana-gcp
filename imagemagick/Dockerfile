FROM golang:1.3.1
MAINTAINER Francesc Campoy <campoy@google.com>

# Ignore APT warnings about not having a TTY
ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update && apt-get install -y wget imagemagick --no-install-recommends

ADD Godeps/_workspace/ /go/
ADD . /go/src/github.com/GoogleCloudPlatform/cloud-abelana-go/imagemagick
RUN go install github.com/GoogleCloudPlatform/cloud-abelana-go/imagemagick

WORKDIR /go/src/github.com/GoogleCloudPlatform/cloud-abelana-go/imagemagick
CMD /go/bin/imagemagick

# Document that the service listens on port 8080.
EXPOSE 10443