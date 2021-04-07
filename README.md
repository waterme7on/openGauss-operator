# openGauss-controller

The openGauss-controller uses the [client-go](https://github.com/kubernetes/client-go) library to develop a custom controller monitoring, scheduling and updating openGauss cluster in kubernetes.

## Structure

![](./docs/diagrams/operator.png)

<br>

## Develop

how the various components in the [client-go](https://github.com/kubernetes/client-go) library work and their interaction points with the custom controller code

![](./docs/diagrams/client-go-controller-interaction.jpeg)
