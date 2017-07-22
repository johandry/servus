FROM alpine

ARG VERSION

ADD ./pkg/v${VERSION}/linux/amd64/servus .

COPY ./tmp/serf /serf

EXPOSE 9000
EXPOSE 7946 7373 8080

CMD [ "./servus" ]
