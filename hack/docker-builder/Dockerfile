FROM alukiano/builder@sha256:5068f41cafc51af22ee2a2f2a1426b6461e925777d19e919068c68faabf3d4c5

ADD rsyncd.conf /etc/rsyncd.conf
ADD entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
