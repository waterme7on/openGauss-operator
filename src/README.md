## How to build



```

# go version go1.16.2 linux/amd64

GOOS=linux go build -o ./app .

docker build -t waterme7on/opengauss-operator .


```


<br></br>


If you are using a minikube cluster, you can just use the local image, otherwise you have to push it to a registry that you Kubernetes cluster can pull from.

```

# minikube

eval $(minikube docker-env)


# kubernetes

docker image push waterme7on/opengauss-operator


# run

kubectl run --rm -i demo --image=waterme7on/opengauss-operator

```



For testing: `test.sh`


