FROM fedora@sha256:9c78c69f748953ba8fdb6eb9982e1abefe281d9b931a13f251eb8aec988353de

# Install packages
RUN dnf install -y dnf-plugins-core && \
dnf copr enable -y vbatts/bazel && \
dnf -y install \
bazel-0.26.0-2.fc30 \
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

ENV GIMME_GO_VERSION=1.11.5
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
go install

COPY rsyncd.conf /etc/rsyncd.conf

COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT [ "/entrypoint.sh" ]
