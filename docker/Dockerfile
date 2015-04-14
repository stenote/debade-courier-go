FROM ubuntu:latest
MAINTAINER rui.ma@geneegroup.com

# Install Basic Packages
RUN apt-get update && apt-get install -y pkg-config software-properties-common golang git build-essential && \
     apt-get update && \
     add-apt-repository -y ppa:chris-lea/zeromq && \
     apt-get update && \
     apt-get install -y libzmq3-dev libpgm-dev && \
     rm -rf /var/lib/apt/lists/*

EXPOSE 3333

ADD run.sh /run.sh

CMD /run.sh
