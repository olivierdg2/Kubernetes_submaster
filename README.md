# Kubernetes_submaster

Kubernetes_submaster is a Kubernetes architecture project. The goal of this project is to make creation of Kubernetes cluster easier. To do so, a containerized submaster is deployed on an existing node. To simplify its usage, CRD and custom controller have been used.

### kubebuilder

The kubebuilder directory contains all the file related to the kubebuilder framework, used to generate CRD and use a custome controller.

### pod_deployment

The pod_deployment directory contains a yaml file that can be used to deploy a submaster without any CRD and custom controller. 



