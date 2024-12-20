FROM golang:1.20 as builder
WORKDIR /app
USER root
COPY . .
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download
RUN /bin/sh build.sh aliyun-clb-controller
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/aliyun-clb-controller .
CMD [ "./aliyun-clb-controller" ]