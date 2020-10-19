FROM golang:1.13 as builder

ADD . /go/bhchain

WORKDIR /go/

RUN git clone --depth 1 https://github.com/hbtc-chain/tendermint.git
RUN git clone --depth 1 https://github.com/hbtc-chain/iavl.git
RUN git clone --depth 1 https://github.com/hbtc-chain/chainnode.git
RUN git clone --depth 1 https://github.com/hbtc-chain/gotron-sdk.git

ENV GOPROXY=https://goproxy.io

WORKDIR /go/bhchain
RUN CGO_ENABLED=0 go build -tags netgo -o /go/bin/hbtcd ./cmd/hbtcd
RUN CGO_ENABLED=0 go build -tags netgo -o /go/bin/hbtccli ./cmd/hbtccli


FROM alpine:latest

WORKDIR /go/
COPY --from=builder /go/bin/hbtcd /go/
COPY --from=builder /go/bin/hbtccli /go/
COPY --from=builder /go/bhchain/run.sh /go/

# p2p port
EXPOSE 26656
# RPC port
EXPOSE 26657
# p2p GRPC service port for bhsettle
EXPOSE 26659

VOLUME [ "/root/.hbtcd" ]

ENTRYPOINT [ "sh", "run.sh" ]
