FROM fedora@sha256:57d86e03971841e79585379f8346289ceb5a3e8ee06933fbd5064b4f004659d1

RUN dnf install -y python-pip && dnf clean all

RUN pip install ipmisim pycryptodome
