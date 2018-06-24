FROM centos

RUN yum -y install golang

RUN mkdir -p /opt/mm-memebot

COPY . /opt/mm-memebot

WORKDIR /opt/mm-memebot

RUN go build

ENTRYPOINT ["/opt/mm-memebot/mm-memebot"]
