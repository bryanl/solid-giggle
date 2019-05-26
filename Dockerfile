FROM debian:9 AS hugo-builder

LABEL description="Docker container for building keps static site"
LABEL maintainer="Bryan Liles <lilesb@vmware.com>"

WORKDIR /site

# knobs
ENV HUGO_VERSION=0.55.6
ENV NODE_VERSION=10.x
ENV HUGO_HOME=./site

# expert level knobs
ENV HUGO_ID=hugo_extended_${HUGO_VERSION}
ENV HUGO_PKG=${HUGO_ID}_Linux-64bit.deb

ENV HUGO_URL=https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/${HUGO_PKG}
ENV NODE_URL=https://deb.nodesource.com/setup_${NODE_VERSION}

ADD ${HUGO_URL} /tmp
ADD ${NODE_URL} /tmp/nodesource_setup.sh

RUN apt-get update \
    && bash /tmp/nodesource_setup.sh \
    && apt-get install -y -q \
        build-essential \
        nodejs \
    && dpkg -i /tmp/${HUGO_PKG}

ADD ${HUGO_HOME} /site
RUN cd /site/themes/kep \
    && npm i \
    && npm run build \
    && cd /site \
    && hugo

FROM nginx

LABEL description="Docker container for building keps static site"
LABEL maintainer="Bryan Liles <lilesb@vmware.com>"

COPY --from=hugo-builder /site/public /usr/share/nginx/html