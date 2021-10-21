package controllers

import (
	k3smasterv1 "example.com/k3smaster/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *SubmasterReconciler) desiredDeployment(sub k3smasterv1.Submaster) (appsv1.Deployment, error) {
	t := true
	var path corev1.HostPathVolumeSource
	path.Path = "."
	depl := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: appsv1.SchemeGroupVersion.String(), Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      sub.Name,
			Namespace: sub.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"submaster": sub.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"submaster": sub.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "k3s",
							Image:           "rancher/k3s:latest",
							Args:            []string{"server", "--kubelet-arg", "port=10350", "--kubelet-arg", "healthz-port=10348", "--kube-proxy-arg", "healthz-bind-address=0.0.0.0:10356", "--kube-proxy-arg", "metrics-bind-address=127.0.0.1:10349", "--kube-scheduler-arg", "port=10351"},
							Env:             []corev1.EnvVar{{Name: "K3S_TOKEN", Value: "f7b5d6eab0aa000d7a8615065a40e40b"}, {Name: "K3S_KUBECONFIG_OUTPUT", Value: "/output/kubeconfig.yaml"}, {Name: "K3S_KUBECONFIG_MODE", Value: "666"}},
							VolumeMounts:    []corev1.VolumeMount{{MountPath: "/var/lib/rancher/k3s", Name: "k3s-server"}},
							SecurityContext: &corev1.SecurityContext{Privileged: &t},
						},
					},
					HostNetwork: true,
					//default = emptydir (will be deprecated)
					Volumes:  []corev1.Volume{{Name: "k3s-server"}},
					NodeName: sub.Name,
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(&sub, &depl, r.Scheme); err != nil {
		return depl, err
	}

	return depl, nil
}
