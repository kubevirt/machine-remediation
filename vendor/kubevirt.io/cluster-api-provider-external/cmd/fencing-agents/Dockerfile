FROM centos:centos7
RUN yum install -y fence-agents-all
COPY fence_lockfile secret-handler.sh /bin/
ENTRYPOINT ["/bin/secret-handler.sh"]