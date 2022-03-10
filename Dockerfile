FROM golang:1.17.8
ENV CGO_ENABLED 0
ADD . /src
WORKDIR /src
RUN go build -a --installsuffix cgo --ldflags="-s" -o rollingpin
RUN echo "nobody:x:65534:65534:nobody:/nonexistent:/usr/sbin/nologin" > /src/etc_passwd

FROM scratch
ENV GIN_MODE=release
EXPOSE 8080
COPY --from=0 /src/etc_passwd /etc/passwd
COPY --from=0 /src/rollingpin /rollingpin
USER nobody
ENTRYPOINT ["/rollingpin"]
