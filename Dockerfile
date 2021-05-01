FROM golang:1.16 as build

RUN apt-get update && apt-get install -y ninja-build

WORKDIR /go/src
RUN git clone https://github.com/jn-lp/se-lab21
WORKDIR /go/src/se-lab21
RUN go get -u ./build/cmd/bood

WORKDIR /go/src/se-lab22
COPY . .

RUN CGO_ENABLED=0 bood

# ==== Final image ====
FROM alpine:3.11
WORKDIR /opt/se-lab22
COPY entry.sh ./
COPY --from=build /go/src/se-lab22/out/bin/* ./
ENTRYPOINT ["/opt/se-lab22/entry.sh"]
CMD ["server"]
