## How to build



```

GOOS=linux go build -o ./app .

docker build -t waterme7on/opengauss-operator .


```




If you are using a minikube cluster, you can just use the local image. otherwise you have to push it to a registry that you Kubernetes cluster can pull from.

```

# minikube

eval $(minikube docker-env)


# kubernetes

docker image push waterme7on/opengauss-operator


# run

kubectl run --rm -i demo --image=waterme7on/opengauss-operator

```



For testing:

```
go clean
rm ./app -f
kubectl delete po demo --force
docker image rm waterme7on/opengauss-operator:test
GOOS=linux go build -o ./app .
docker build -t waterme7on/opengauss-operator:test .
docker image push waterme7on/opengauss-operator:test

kubectl delete -f ../manifests/opengauss-operator.yaml
kubectl apply -f ../manifests/opengauss-operator.yaml


kubectl run --rm -i demo --image=waterme7on/opengauss-operator:test
```




```