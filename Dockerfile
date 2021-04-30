FROM golang:1.16 as build

RUN apt-get update && apt-get install -y ninja-build

# TODO: Змініть на власну реалізацію системи збірки
RUN go get -u github.com/jn-lp/se-lab22/build/cmd/bood

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
