FROM golang:alpine
COPY . /go/src
RUN cd /go/src && go mod download && go build -o /entrypoint /go/src/cmd/ci/*.go
CMD /entrypoint
