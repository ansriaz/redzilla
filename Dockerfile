FROM golang AS builder

RUN go get -v github.com/golang/dep/cmd/dep

WORKDIR $GOPATH/src/github/ansriaz/redzilla
# COPY Gopkg.toml Gopkg.lock ./

# RUN dep ensure --vendor-only
COPY . .
RUN go get -d -v
RUN rm -rf $GOPATH/github.com/docker/docker/vendor/github.com/docker/go-connections
RUN CGO_ENABLED=0 GOOS=linux go build  -a -ldflags "-s -X main.Buildtimestamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.Githash=`git rev-parse HEAD`"

FROM scratch
ENV GIN_MODE=release
COPY --from=builder /go/src//src/github/ansriaz/redzilla/redzilla .

ENTRYPOINT [ "/redzilla" ]
