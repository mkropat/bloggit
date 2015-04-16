# Build image with:
#
#     docker build -t bloggit .
#
# Run with:
#
#     docker run -p 80:8000 bloggit

FROM golang:1.3.1-onbuild
MAINTAINER Michael Kropat <mail@michael.kropat.name>

# https://github.com/docker-library/node/blob/013858ac35afb9ca7b10102956427b629e7708da/0.10/Dockerfile

RUN gpg --keyserver pgp.mit.edu --recv-keys 7937DFD2AB06298B2293C3187D33FF9D0246406D

ENV NODE_VERSION 0.10.33
ENV NPM_VERSION 2.1.8

RUN curl -SLO "http://nodejs.org/dist/v$NODE_VERSION/node-v$NODE_VERSION-linux-x64.tar.gz" \
	&& curl -SLO "http://nodejs.org/dist/v$NODE_VERSION/SHASUMS256.txt.asc" \
	&& gpg --verify SHASUMS256.txt.asc \
	&& grep " node-v$NODE_VERSION-linux-x64.tar.gz\$" SHASUMS256.txt.asc | sha256sum -c - \
	&& tar -xzf "node-v$NODE_VERSION-linux-x64.tar.gz" -C /usr/local --strip-components=1 \
	&& rm "node-v$NODE_VERSION-linux-x64.tar.gz" SHASUMS256.txt.asc \
	&& npm install -g npm@"$NPM_VERSION" \
	&& npm cache clear

RUN npm install -g bower gulp

RUN useradd -d "$PWD" bloggit && chown -R bloggit .
USER bloggit

RUN npm install && bower install && gulp

EXPOSE 8000
CMD ["/go/bin/bloggit"]
