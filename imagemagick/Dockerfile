FROM golang:1.3.1
MAINTAINER Francesc Campoy <campoy@google.com>

# Ignore APT warnings about not having a TTY
ENV DEBIAN_FRONTEND noninteractive

# install build essentials
RUN apt-get update && apt-get install -y wget build-essential pkg-config --no-install-recommends

# Install webp
RUN apt-get -q -y install libjpeg-dev libpng-dev libtiff-dev libgif-dev --no-install-recommends
RUN wget http://downloads.webmproject.org/releases/webp/libwebp-0.4.2.tar.gz
RUN tar xvzf libwebp-0.4.2.tar.gz
WORKDIR libwebp-0.4.2
RUN ./configure
RUN make
RUN make install

# install latest version of imagemagick
RUN cd
RUN wget http://www.imagemagick.org/download/ImageMagick.tar.gz
RUN tar xvzf ImageMagick.tar.gz
RUN cd ImageMagick-*
RUN ./configure
RUN make
RUN make install
RUN ldconfig /usr/local/lib

# ENV PKG_CONFIG_PATH /go/libwebp-0.4.2/ImageMagick-6.8.9-9/magick:/go/libwebp-0.4.2/ImageMagick-6.8.9-9/wand
# RUN cp /go/libwebp-0.4.2/ImageMagick-6.8.9-9/wand/MagickWand.pc.in /go/libwebp-0.4.2/ImageMagick-6.8.9-9/wand/MagickWand.pc
# ENV C_INCLUDE_PATH /go/libwebp-0.4.2/ImageMagick-6.8.9-9

ADD Godeps/_workspace/ /go/
ADD . /go/src/github.com/GoogleCloudPlatform/abelana-gcp/imagemagick
# Why is this failing? Missing some config here.
RUN go install github.com/GoogleCloudPlatform/abelana-gcp/imagemagick
RUN touch ~/logs
CMD /go/bin/imagemagick >> ~/logs 2>&1

# Document that the service listens on port 8080.
EXPOSE 8080