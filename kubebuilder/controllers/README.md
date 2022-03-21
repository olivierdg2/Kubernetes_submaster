# About the controller
## submaster_controller.go
Contains the controller logic.
### Finalizer
#### Explanation
A finalizer is a mark that can be placed on a kubernetes object. 
If the object is deleted, the corresponding controller is alerted in order to perform necessary deletions before the object can be deleted.
Finalizers are often used to perform external dependencies deletion.
For further informations about finalizers: https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers/
#### About the code 
First the controller retrieve a submaster object representing the branch.
Then it checks if the submaster is being deleted.
If so, it performs the finalizer.
If not, it add the finalizer to the submaster if needed.
Our finalizer deletes the corresponding secret and kubefedcluster.
### Containerized master deployment 
### Explanation
When we create a submaster, it only represents a meaningless Kubernetes object.
I order to make it meaningful we have to create the corresponding components and event that represents our branch object.
In our case we need to create a deployment on the desired node that will create a pod containing the containerized master. Then we need to join the master to the existing federation.
### About the code
The controller create a deployment on the node specified by the submaster name.
It also create an ownership between the submaster and the deployment. This ownership is useful on deletion, as the submaster owns the deployment, if the submaster is deleted so is the deployment.
### Kubefederation join
### Explanation
As explained in the previous section, the generated submaster needs to join the existing federation. In order to do so, we need the Kubeconfig of both the trunk and the branch.
### About the code
The controller checks if the pod of the containerized master is running (so if the kubeconfig is generated). 
If so it creates 2 jobs. The first one will retrieve the kubeconfig of the submaster then create a secret containing it. The second one will perform a kubefedctl join command with the 2 kubeconfig.
## helpers.go 
Provides some helper functions for the controller.
### desiredDeployment
Create the deployment of the containerized master.
### desiredConfigJob
Create the job that will retrieve the Kubeconfig of the submaster.
### desiredKubefedJob
Create the jjob that will perform the kubefedctl join command.
