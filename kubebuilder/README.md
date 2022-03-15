# kubebuilder

## Important files 

### main.go
#### Note

init function needs to be changed in order to make the controller use other resources.

### Dockerfile

Dockerfile used to Dockerize the whole controller project.

### /chart/test 

The generetad helm chart by Helmify corresponding to this Kubebuilder project.

### /api/v1/groupversion_info.go 

The CRD group is based on this file.

### /api/v1/submaster_types.go 

The CRD is based on this file .

### /controllers/submaster_controller.go

This is our custom controller.

### /controllers/herlpers.go 

Provides some helpers for our controller.
