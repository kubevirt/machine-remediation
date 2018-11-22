FROM fedora@sha256:57d86e03971841e79585379f8346289ceb5a3e8ee06933fbd5064b4f004659d1

RUN dnf -y install make \
    git \
    mercurial \
    sudo \
    gcc \
    findutils \
    gradle \
    rsync-daemon \
    rsync \
    protobuf-compiler && dnf -y clean all

# Add bazel repo
COPY bazel.repo /etc/yum.repos.d/

# Install bazel and all needed dependencies
RUN dnf -y install gcc-c++ glibc-devel bazel docker \
    && dnf -y clean all

ENV GIMME_GO_VERSION=1.10

RUN mkdir -p /gimme && curl -sL https://raw.githubusercontent.com/travis-ci/gimme/master/gimme | HOME=/gimme bash >> /etc/profile.d/gimme.sh

ENV GOPATH="/go" GOBIN="/usr/bin"

ADD rsyncd.conf /etc/rsyncd.conf

RUN \
    mkdir -p /go && \
    source /etc/profile.d/gimme.sh && \
    go get github.com/mattn/goveralls && \
    go get -u github.com/golang/dep/cmd/dep && \
    go get golang.org/x/tools/cmd/goimports && \
    git clone https://github.com/mvdan/sh.git $GOPATH/src/mvdan.cc/sh && \
    cd $GOPATH/src/mvdan.cc/sh/cmd/shfmt && \
    git checkout v2.5.0 && \
    go get mvdan.cc/sh/cmd/shfmt && \
    go install && \
    go get -u github.com/golang/mock/gomock && \
    go get -u github.com/golang/mock/mockgen && \
    go get -u github.com/onsi/ginkgo/ginkgo && \
    git clone https://github.com/kubernetes/code-generator -b release-1.11 $GOPATH/src/k8s.io/code-generator && \
    cd $GOPATH/src/k8s.io/code-generator && \
    go install ./cmd/{defaulter-gen,client-gen,lister-gen,informer-gen,deepcopy-gen} && \
    go get -u -d github.com/golang/protobuf/protoc-gen-go && \
    cd $GOPATH/src/github.com/golang/protobuf/protoc-gen-go && \
    git checkout 1643683e1b54a9e88ad26d98f81400c8c9d9f4f9 && \
    go install

ADD entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
