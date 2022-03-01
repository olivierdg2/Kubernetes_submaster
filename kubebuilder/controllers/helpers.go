package controllers

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	branch "kubernetrees.com/kubebuilder/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	batchv1 "k8s.io/api/batch/v1"
)

func (r *SubmasterReconciler) desiredDeployment(sub branch.Submaster) (appsv1.Deployment, error) {
	t := true
	var path corev1.HostPathVolumeSource
	path.Path = "."
	//TODO add output volume to access kubeconfig from host
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
					Labels: map[string]string{"submaster": sub.Name,"pod": sub.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:         "k3s",
							Image:        "rancher/k3s:latest",
							Args:         []string{"server","--disable-cloud-controller", "--kubelet-arg", "port=10350", "--kubelet-arg", "healthz-port=10348", "--kube-proxy-arg", "healthz-bind-address=0.0.0.0:10356", "--kube-proxy-arg", "metrics-bind-address=127.0.0.1:10349"},
							Env:          []corev1.EnvVar{{Name: "K3S_TOKEN", Value: "f7b5d6eab0aa000d7a8615065a40e40b"}, {Name: "K3S_KUBECONFIG_OUTPUT", Value: "/output/kubeconfig.yaml"}, {Name: "K3S_KUBECONFIG_MODE", Value: "666"}},
							VolumeMounts: []corev1.VolumeMount{{MountPath: "/var/lib/rancher/k3s", Name: "k3s-server"}},
							//Add mount to output
							SecurityContext: &corev1.SecurityContext{Privileged: &t},
						},
					},
					HostNetwork: true,
					//default = emptydir (will be deprecated)
					Volumes: []corev1.Volume{{Name: "k3s-server"}},
					//Add volume to ouput
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

func (r *SubmasterReconciler) desiredJob(pod corev1.Pod, sub branch.Submaster) (batchv1.Job, error) {
	a := int32(100)
	b := int32(4)
	job := batchv1.Job{
		TypeMeta: metav1.TypeMeta{APIVersion: batchv1.SchemeGroupVersion.String(), Kind: "Job"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "retrieve-kubeconfig-" + sub.Name,
			Namespace: sub.Namespace,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &a,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "kubefedctl-" + sub.Name,
				},
				Spec: corev1.PodSpec{
					RestartPolicy: "Never",
					Containers: []corev1.Container{
						{
							Name:         "kubefed",
							Image:        "olivierdg1/kubefedctl",
							Command:         []string{"/bin/sh","-c","kubectl exec -it " + pod.Name + " -- cat /etc/rancher/k3s/k3s.yaml > tmp.yaml && sed -i -e 's/^/    /' tmp.yaml && echo \"\" >> branch-config.yaml && cat tmp.yaml >> branch-config.yaml && sed -i -e \"s/subname/$NAME/g\" -e \"s/127.0.0.1/$IP/g\" -e \"s/sublabel/$NAME/g\" branch-config.yaml && cat branch-config.yaml && kubectl apply -f branch-config.yaml"},
							Env:          []corev1.EnvVar{{Name: "KUBECONFIG", Value: "config.yaml"}, {Name: "NAME", Value: sub.Name}, {Name: "IP", Value: pod.Status.PodIP}},
							VolumeMounts: []corev1.VolumeMount{{MountPath: "./config.yaml", Name: "kubeconfig", SubPath: "config.yaml"}},
						},
					},
					Volumes: []corev1.Volume{{Name: "kubeconfig", VolumeSource : corev1.VolumeSource{ ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference : corev1.LocalObjectReference{Name: "kubeconfig"}}}}},
				},
		
			},
			BackoffLimit: &b,
		},
	}

	if err := ctrl.SetControllerReference(&sub, &job, r.Scheme); err != nil {
		return job, err
	}

	return job, nil
}
