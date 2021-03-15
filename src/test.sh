#!/bin/bash
go clean
rm ./app -f
docker image rm waterme7on/opengauss-operator:test
GOOS=linux go build -o ./app .
docker build -t waterme7on/opengauss-operator:test .
docker image push waterme7on/opengauss-operator:test
# use deploy to test
kubectl delete -f ../manifests/opengauss-operator.yaml
kubectl apply -f ../manifests/opengauss-operator.yaml