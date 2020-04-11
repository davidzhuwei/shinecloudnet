FROM ubuntu:18.04
MAINTAINER Greg Szabo <greg@tendermint.com>

RUN apt update && \
    apt install curl jq file -y

VOLUME [ /scloud ]
WORKDIR /scloud
EXPOSE 26656 26657
ENTRYPOINT ["/usr/bin/wrapper.sh"]
CMD ["start"]
STOPSIGNAL SIGTERM

COPY wrapper.sh /usr/bin/wrapper.sh

