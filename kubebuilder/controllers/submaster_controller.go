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
	kubefed "sigs.k8s.io/kubefed/pkg/apis/core/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	types "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

    branchFinalizer := "branch.finalizers.kubernetrees.com"

    if sub.ObjectMeta.DeletionTimestamp.IsZero() {
        // The object is not being deleted, so if it does not have our finalizer,
        // then lets add the finalizer and update the object.
		if !controllerutil.ContainsFinalizer(&sub, branchFinalizer) {
			controllerutil.AddFinalizer(&sub, branchFinalizer)
			if err := r.Update(ctx, &sub); err != nil {
				return ctrl.Result{}, err
			}
		}
        
    } else {
        // The object is being deleted
        if controllerutil.ContainsFinalizer(&sub, branchFinalizer) {
            // our finalizer is present, so lets handle any external dependency
            if err := deleteExternalResources(sub, ctx, r); err != nil {
                // if fail to delete the external dependency here, return with error
                // so that it can be retried
                return ctrl.Result{}, err
            }

            // remove our finalizer from the list and update it.
            controllerutil.RemoveFinalizer(&sub, branchFinalizer)
            if err := r.Update(ctx, &sub); err != nil {
                return ctrl.Result{}, err
            }
        }

        // Stop reconciliation as the item is being deleted
        return ctrl.Result{}, nil
    }

	listOptions := []client.ListOption{
		client.MatchingLabels(map[string]string{"submaster":sub.Name,"pod": sub.Name}),
		client.InNamespace(sub.Namespace),
	}

	var podList corev1.PodList
	if err := r.List(ctx, &podList, listOptions...); err != nil {
		return ctrl.Result{}, fmt.Errorf("%v", err)
	}

	if len(podList.Items) != 0 {
		var pod corev1.Pod
		pod = podList.Items[0]
		sub.Status.Status = pod.Status.Phase
		sub.Status.IP = pod.Status.PodIP
		if pod.Status.Phase == "Running"{
			configJob, err := r.desiredConfigJob(pod,sub)
			applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner("submaster")}
			if err = r.Patch(ctx, &configJob, client.Apply, applyOpts...); err != nil{
				return ctrl.Result{}, err
			}
			listOptions := []client.ListOption{
				client.MatchingLabels(map[string]string{"submaster":sub.Name,"configJob": sub.Name}),
				client.InNamespace(sub.Namespace),
			}
			var jobList batchv1.JobList
			if err := r.List(ctx, &jobList, listOptions...); err != nil {
				return ctrl.Result{}, fmt.Errorf("%v", err)
			}
			if len(jobList.Items) != 0 {
				kubefedJob, err := r.desiredKubefedJob(sub)
				applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner("submaster")}
				if err = r.Patch(ctx, &kubefedJob, client.Apply, applyOpts...); err != nil{
					return ctrl.Result{}, err
				}
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

func deleteExternalResources(sub branch.Submaster, ctx context.Context, r *SubmasterReconciler) error {
	var kubefedObject kubefed.KubeFedCluster
	var kubefedNamespacedName types.NamespacedName
	kubefedNamespacedName.Namespace = sub.Namespace
	kubefedNamespacedName.Name = "branch-" + sub.Name
	if err := r.Get(ctx, kubefedNamespacedName, &kubefedObject); err != nil {
		return fmt.Errorf("%v", err)
	}
	if err := r.Delete(ctx, &kubefedObject); err != nil{
		return fmt.Errorf("%v", err)
	}
	var secretObject corev1.Secret
	kubefedNamespacedName.Name = "kubeconfig-" + sub.Name
	if err := r.Get(ctx, kubefedNamespacedName, &secretObject); err != nil {
		return fmt.Errorf("%v", err)
	}
	if err := r.Delete(ctx, &secretObject); err != nil{
		return fmt.Errorf("%v", err)
	}
	return nil
}