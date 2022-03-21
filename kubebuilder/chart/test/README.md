# Tools used
## Kubebuilder : https://github.com/kubernetes-sigs/kubebuilder
We used Kubebuilder to generate the CRDs and operator.
## Helmify : https://github.com/arttor/helmify
We used Helmify to generate a base of chart based on our Kubebuilder project.
# The Kubebuilder project : https://github.com/olivierdg2/Kubernetes_submaster/tree/main/kubebuilder
# The controller image : https://hub.docker.com/repository/docker/olivierdg1/submaster
# Notes
## TODO
replace the scheduling on master node by scheduling on roots
## templates/deployment.yaml
```
    spec:
      nodeSelector:
        dedicated: master
      tolerations:
      - key: dedicated
        operator: Equal
        value: master
        effect: NoSchedule
```
Those lines where added to deploy the controller on the master node. To make it work the master node needs to be tainted this way : 
```
kubectl label nodes name_of_your_node dedicated=master
```
The controller must be deployed on the master otherwise the deployed containerized submaster will interfer with it leading to CrashLoopBackOff.
## templates/manager-rbac.yaml
```
rules:
- apiGroups:
  - branch.kubernetrees.com
  resources:
  - submasters
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - branch.kubernetrees.com
  resources:
  - submasters/finalizers
  verbs:
  - update
- apiGroups:
  - branch.kubernetrees.com
  resources:
  - submasters/status
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - apps
  resources:
  - deployments
  verbs: 
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get 
  - watch 
  - list
- apiGroups:
  - batch
  resources:
  - jobs
  verbs: 
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups: 
  - ""
  resources:
  - secrets
  verbs:
  - get 
  - watch 
  - list
  - delete
- apiGroups: 
  - "core.kubefed.io"
  resources:
  - kubefedclusters
  verbs:
  - get 
  - watch 
  - list
  - delete
```
Those lines where added to grant the controller access to the necessary resources. The pods are needed to be watched in order to generate the corresponding deployment to it. Finalizers, secret and kubefedclusters are needed in order to perform the necessary deletion when the branch is deleted.
