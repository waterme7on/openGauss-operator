FROM golang
WORKDIR $GOPATH/src/openGauss-operator
ADD . $GOPATH/src/openGauss-operator
ENV GO111MODULE=on 
ENV GOPROXY="https://goproxy.io"
ENV KUBERNETES_SERVICE_HOST="kubernetes.default.svc"
ENV KUBERNETES_SERVICE_PORT=443
RUN go build -o controller .
EXPOSE 17173
ENTRYPOINT ["./controller"]