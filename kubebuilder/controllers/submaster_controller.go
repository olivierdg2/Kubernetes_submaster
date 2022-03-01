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
	// "strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	branch "kubernetrees.com/kubebuilder/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	//"k8s.io/client-go/kubernetes"
	//"k8s.io/client-go/tools/clientcmd"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SubmasterReconciler reconciles a Submaster object
type SubmasterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=branch.kubernetrees.com,resources=submasters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=branch.kubernetrees.com,resources=submasters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=branch.kubernetrees.com,resources=submasters/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:resources=pods,verbs=get;watch;list

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
	var sub branch.Submaster
	if err := r.Get(ctx, req.NamespacedName, &sub); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	listOptions := []client.ListOption{
		client.MatchingLabels(map[string]string{"pod": sub.Name}),
		client.InNamespace(sub.Namespace),
	}

	var list corev1.PodList
	if err := r.List(ctx, &list, listOptions...); err != nil {
		return ctrl.Result{}, fmt.Errorf("%v", err)
	}

	if len(list.Items) != 0 {
		var pod corev1.Pod
		pod = list.Items[0]
		sub.Status.Status = pod.Status.Phase
		sub.Status.IP = pod.Status.PodIP
		if pod.Status.Phase == "Running"{
			job, err := r.desiredJob(pod,sub)
			applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner("submaster")}
			if err = r.Patch(ctx, &job, client.Apply, applyOpts...); err != nil{
				return ctrl.Result{}, err
			}
		}
	} else {
		sub.Status.Status = "No pod generated"
		sub.Status.IP = ""
	}

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
		For(&branch.Submaster{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
