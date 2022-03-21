# About the API
## groupversion_info.go
Contains the CRD group specifications. 
In our case our group is named branch.kubernetrees.com
## submaster_types.go
Contains the CRD specifications.
### Spec 
Contains CRD specification used to create the Kubernetes object.
In our case since we only use object name to generate our corresponding Kubernetes object, it's empty.
### Status 
Contains CRD fields displayed as status.
In our case we have the branch IPv4 address, readyness status and nodes number.
