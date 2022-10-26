FROM alpine:3.14
RUN apk add --no-cache libc6-compat
RUN mkdir /tankerkoenig-mqtt
COPY ./tankerkoenig-mqtt /tankerkoenig-mqtt/
COPY ./config.yml /config.yml
RUN chmod 777 /tankerkoenig-mqtt/tankerkoenig-mqtt
ENTRYPOINT ["/tankerkoenig-mqtt/tankerkoenig-mqtt"]
