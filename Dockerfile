FROM golang:1.17

COPY ./ ./apisix-etcd

RUN if [ "$ENABLE_PROXY" = "true" ] ; then go env -w GOPROXY=https://goproxy.io,direct ; fi \
    && go env -w GO111MODULE=on \
    && cd apisix-etcd \
    && make build \
    &&  make install

WORKDIR /usr/local/apisix-etcd

ENV PATH=$PATH:/usr/local/apisix-etcd

ENV APISIX_ETCD_WORKDIR /usr/local/apisix-etcd

CMD [ "/usr/local/apisix-etcd/etcd-adapter" ]
