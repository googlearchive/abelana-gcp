set -e
export DEBIAN_FRONTEND=noninteractive

# install C compiler
sudo apt-get -q -y update
sudo apt-get -q -y upgrade
sudo apt-get -q -y install build-essential

# installing webp
cd
sudo apt-get -q -y install libjpeg-dev libpng-dev libtiff-dev libgif-dev
wget http://downloads.webmproject.org/releases/webp/libwebp-0.4.2.tar.gz
tar xvzf libwebp-0.4.2.tar.gz
cd libwebp-0.4.2
./configure
make
sudo make install

# install latest version of imagemagick
wget http://www.imagemagick.org/download/ImageMagick.tar.gz
tar xvzf ImageMagick.tar.gz
cd ImageMagick-*
./configure
make
sudo make install
sudo ldconfig /usr/local/lib
# not really needed
# make check

# install latest go version
cd
wget https://storage.googleapis.com/golang/go1.3.3.linux-amd64.tar.gz
tar xvzf go1.3.3.linux-amd64.tar.gz
sudo mv go /usr/local/go
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# obtain code and build
gsutil -m cp -R gs://abelana-code/imagemagick .
cd imagemagick
GOPATH=$PWD/Godeps/_workspace go build
touch ~/logs

# generate cert.pem and key.pem
export HOST=`grep Google /etc/hosts | awk '{print $2}'`
go run /usr/local/go/src/pkg/crypto/tls/generate_cert.go --host=$HOST

# start running
./imagemagick --debug=1 >> ~/logs 2>&1 &
