# pod_deployment 

## Requirements

Having an existing Kubernetes cluster. 

All node that can be granted submaster role must use Docker as container engine and have the 6443 and 6444 port available.

## Usage

Apply the yaml file to the cluster

'''bash 
kubectl apply -f k3smaster.yml
'''

TODO rajouter le volume hostpath pour y mettre l'output 