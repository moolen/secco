#!/bin/bash

set -eux

sudo add-apt-repository ppa:longsleep/golang-backports
sudo apt update
sudo apt-get install -y golang-go "linux-headers-$(uname -r)"

# official bcc packages and gobpf is broken ATM
# see https://github.com/iovisor/gobpf/issues/201 and https://github.com/iovisor/gobpf/issues/214
# We need to build it from source at the moment
# the following is taken from https://github.com/iovisor/bcc/blob/master/INSTALL.md#ubuntu---source

# Trusty (14.04 LTS) and older
VER=trusty
echo "deb http://llvm.org/apt/$VER/ llvm-toolchain-$VER-3.7 main
deb-src http://llvm.org/apt/$VER/ llvm-toolchain-$VER-3.7 main" | \
  sudo tee /etc/apt/sources.list.d/llvm.list
wget -O - http://llvm.org/apt/llvm-snapshot.gpg.key | sudo apt-key add -
sudo apt-get update

# For Bionic (18.04 LTS)
sudo apt-get -y install bison build-essential cmake flex git libedit-dev \
  libllvm6.0 llvm-6.0-dev libclang-6.0-dev python zlib1g-dev libelf-dev python3-distutils

# recommended tools
sudo apt-get install -y arping netperf iperf

# For Lua support
sudo apt-get -y install luajit luajit-5.1-dev

git clone https://github.com/iovisor/bcc.git
mkdir bcc/build; cd bcc/build
cmake ..
make
sudo make install
cmake -DPYTHON_CMD=python3 .. # build python3 binding
pushd src/python/
make
sudo make install
popd

# install docker
sudo apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg-agent \
    software-properties-common

curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -

sudo add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable"

sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io
sudo docker run hello-world
