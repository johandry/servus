# Build image
FROM golang:1.8.3 AS build

ARG PKG_NAME=servus
ARG PKG_BASE=github.com/johandry

ADD . /go/src/${PKG_BASE}/${PKG_NAME}

RUN echo \
  && cd /go/src/${PKG_BASE}/${PKG_NAME} \
  && make build4docker \
  && echo

# To do a manual build uncomment this line and comment out the following lines
# CMD [ "/bin/bash" ]

# Application image
FROM scratch AS application

COPY --from=build /servus .

EXPOSE 9000
EXPOSE 7946 7373 8080

CMD [ "./servus" ]
