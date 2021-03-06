package controllers

import (
	"fmt"
	deploymentsv1alpha1 "github.com/devlup-labs/darkroom-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"net"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *DarkroomReconciler) desiredConfigMap(darkroom deploymentsv1alpha1.Darkroom) (corev1.ConfigMap, error) {
	cfg := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String(), Kind: "ConfigMap"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      darkroom.Name,
			Namespace: darkroom.Namespace,
		},
		Data: map[string]string{
			"DEBUG":                                 "true",
			"LOG_LEVEL":                             "debug",
			"APP_NAME":                              darkroom.Name,
			"APP_VERSION":                           darkroom.Spec.Version,
			"APP_DESCRIPTION":                       "A realtime image processing service",
			"SOURCE_KIND":                           fmt.Sprintf("%s", darkroom.Spec.Source.Kind),
			"SOURCE_BASEURL":                        darkroom.Spec.Source.BaseURL,
			"PORT":                                  "3000",
			"CACHE_TIME":                            "31536000",
			"SOURCE_HYSTRIX_COMMANDNAME":            fmt.Sprintf("%s_ADAPTER", darkroom.Spec.Source.Kind),
			"SOURCE_HYSTRIX_TIMEOUT":                "5000",
			"SOURCE_HYSTRIX_MAXCONCURRENTREQUESTS":  "100",
			"SOURCE_HYSTRIX_REQUESTVOLUMETHRESHOLD": "10",
			"SOURCE_HYSTRIX_SLEEPWINDOW":            "10",
			"SOURCE_HYSTRIX_ERRORPERCENTTHRESHOLD":  "25",
		},
	}

	if err := ctrl.SetControllerReference(&darkroom, &cfg, r.Scheme); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func (r *DarkroomReconciler) desiredDeployment(darkroom deploymentsv1alpha1.Darkroom, configMap corev1.ConfigMap) (appsv1.Deployment, error) {
	depl := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: appsv1.SchemeGroupVersion.String(), Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      darkroom.Name,
			Namespace: darkroom.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"darkroom": darkroom.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"darkroom": darkroom.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "darkroom",
							Image: "gojektech/darkroom:latest",
							EnvFrom: []corev1.EnvFromSource{
								{
									ConfigMapRef: &corev1.ConfigMapEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: configMap.Name,
										},
									},
								},
							},
							Ports: []corev1.ContainerPort{
								{ContainerPort: 3000, Name: "http", Protocol: "TCP"},
							},
						},
					},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(&darkroom, &depl, r.Scheme); err != nil {
		return depl, err
	}
	return depl, nil
}

func (r *DarkroomReconciler) desiredService(darkroom deploymentsv1alpha1.Darkroom) (corev1.Service, error) {
	svc := corev1.Service{
		TypeMeta: metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String(), Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      darkroom.Name,
			Namespace: darkroom.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: "http", Port: 8080, Protocol: "TCP", TargetPort: intstr.FromString("http")},
			},
			Selector: map[string]string{"darkroom": darkroom.Name},
			Type:     corev1.ServiceTypeNodePort,
		},
	}

	if err := ctrl.SetControllerReference(&darkroom, &svc, r.Scheme); err != nil {
		return svc, err
	}
	return svc, nil
}

func urlForService(svc corev1.Service, port int32) string {
	if len(svc.Status.LoadBalancer.Ingress) == 0 {
		return ""
	}

	host := svc.Status.LoadBalancer.Ingress[0].Hostname
	if host == "" {
		host = svc.Status.LoadBalancer.Ingress[0].IP
	}

	return fmt.Sprintf("http://%s", net.JoinHostPort(host, fmt.Sprintf("%v", port)))
}
