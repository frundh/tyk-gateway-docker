FROM docker.tyk.io/tyk-gateway/tyk-gateway:v5.1.0
COPY ./tyk.standalone.conf /opt/tyk-gateway/tyk.conf
COPY ./apps /opt/tyk-gateway/apps
COPY ./middleware /opt/tyk-gateway/middleware