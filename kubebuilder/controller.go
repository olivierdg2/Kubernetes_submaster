/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"strconv"

	k3smasterv1 "example.com/k3smaster/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	
)

// SubmasterReconciler reconciles a Submaster object
type SubmasterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=k3smaster.example.com,resources=submasters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=k3smaster.example.com,resources=submasters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=k3smaster.example.com,resources=submasters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Submaster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *SubmasterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	// your logic here
	var sub k3smasterv1.Submaster
	if err := r.Get(ctx, req.NamespacedName, &sub); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	listOptions := []client.ListOption{
		client.MatchingLabels(map[string]string{"submaster": "a-virtualbox"}),
		client.InNamespace("default"),
	}

	var list corev1.PodList
	if err := r.List(ctx, &list, listOptions...); err != nil {
		return ctrl.Result{}, fmt.Errorf("blabla %v", err)
	}
	var pod corev1.Pod
	if len(list.Items) != 0{
		pod = list.Items[0]
	        sub.Status.Status = pod.Status.Phase
	        sub.Status.IP = pod.Status.PodIP
	}else {
	        sub.Status.Status = "No pod generated"
	        sub.Status.IP = ""
	}
	
	//Pour l'instant marche, mais attention: n'update que quand on lance le controller -> doit mettre un watch ou quoi
	//Ne gère pes les erreurs, si le fichier n'est pas trouvé, que le server n'est pas joignable, que la config n'est pas bonne, le controller crash 
	//Question, comment gérer les erreurs + comment gérer la kubeconfig -> le pod est créé -> envoyer sa kubeconfig au bigmaster
	config, err := clientcmd.BuildConfigFromFlags("", "/etc/rancher/submaster/a-virtualbox.yaml")
	if err != nil {
	         panic(err.Error())
	}
	
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
                 panic(err.Error())
	}
	
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
        if err != nil {
	         panic(err.Error())
	}
	sub.Status.Nodes = strconv.Itoa(len(nodes.Items))


	deployment, err := r.desiredDeployment(sub)
	if err != nil {
		return ctrl.Result{}, err
	}

	applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner("submaster")}
	err = r.Patch(ctx, &deployment, client.Apply, applyOpts...)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.Status().Update(ctx, &sub); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SubmasterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k3smasterv1.Submaster{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
