FROM golang:1.7.1

ENV CEPH_VERSION jewel

RUN echo deb http://download.ceph.com/debian-$CEPH_VERSION/ jessie main | tee /etc/apt/sources.list.d/ceph-$CEPH_VERSION.list

RUN wget --no-check-certificate -q -O- 'https://ceph.com/git/?p=ceph.git;a=blob_plain;f=keys/release.asc' | apt-key add - \
    && apt-get update \
    && apt-get install -y --no-install-recommends \
        ceph \
        ceph-mds \
        librados-dev \
        librbd-dev \
        libcephfs-dev \
        uuid-runtime \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

VOLUME /go

COPY ./ci/micro-osd.sh /tmp/micro-osd.sh
COPY ./ci/entrypoint.sh /tmp/entrypoint.sh

#RUN bash /tmp/micro-osd.sh

ENTRYPOINT ["bash", "/tmp/micro-osd.sh" ]
