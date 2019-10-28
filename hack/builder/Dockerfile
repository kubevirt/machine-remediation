FROM fedora@sha256:8a91dbd4b9d283ca1edc2de5dbeef9267b68bb5dae2335ef64d2db77ddf3aa68

# Install packages
RUN dnf install -y dnf-plugins-core && \
dnf copr enable -y vbatts/bazel && \
dnf -y install \
bazel \
cpio \
patch \
make \
git \
mercurial \
sudo \
gcc \
gcc-c++ \
glibc-devel \
rsync-daemon \
rsync \
findutils && \
dnf -y clean all

ENV GIMME_GO_VERSION=1.12.8
ENV GOPATH="/go" GOBIN="/usr/bin"

RUN mkdir -p /gimme && curl -sL https://raw.githubusercontent.com/travis-ci/gimme/master/gimme | HOME=/gimme bash >> /etc/profile.d/gimme.sh

# Install persisten go packages
RUN \
mkdir -p /go && \
source /etc/profile.d/gimme.sh && \
# Install mvdan/sh
git clone https://github.com/mvdan/sh.git $GOPATH/src/mvdan.cc/sh && \
cd $GOPATH/src/mvdan.cc/sh/cmd/shfmt && \
git checkout v2.5.0 && \
go get mvdan.cc/sh/cmd/shfmt && \
go install && \
# install ginkgo
go get -u github.com/onsi/ginkgo/ginkgo

COPY rsyncd.conf /etc/rsyncd.conf

COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT [ "/entrypoint.sh" ]
