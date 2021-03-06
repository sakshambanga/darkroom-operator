/*


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
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	deploymentsv1alpha1 "github.com/devlup-labs/darkroom-operator/api/v1alpha1"
)

// DarkroomReconciler reconciles a Darkroom object
type DarkroomReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=deployments.gojek.io,resources=darkrooms,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=deployments.gojek.io,resources=darkrooms/status,verbs=get;update;patch

func (r *DarkroomReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("darkroom", req.NamespacedName)

	var darkroom deploymentsv1alpha1.Darkroom
	if err := r.Get(ctx, req.NamespacedName, &darkroom); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	cfg, err := r.desiredConfigMap(darkroom)
	if err != nil {
		return ctrl.Result{}, err
	}
	deployment, err := r.desiredDeployment(darkroom, cfg)
	if err != nil {
		return ctrl.Result{}, err
	}
	svc, err := r.desiredService(darkroom)
	if err != nil {
		return ctrl.Result{}, err
	}
	
	igr, err := r.desiredIngress(darkroom)
	if err != nil {
		return ctrl.Result{}, err
	}

	applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner("darkroom-controller")}

	err = r.Patch(ctx, &cfg, client.Apply, applyOpts...)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.Patch(ctx, &deployment, client.Apply, applyOpts...)
	if err != nil {
		return ctrl.Result{}, err
	}
	
	err = r.Patch(ctx, &igr, client.Apply, applyOpts...)
	if err!= nil {
		return ctrl.Result{}, err
	}

	err = r.Patch(ctx, &svc, client.Apply, applyOpts...)
	if err != nil {
		return ctrl.Result{}, err
	}

	darkroom.Status.URL = urlForService(svc, 8080)
	err = r.Status().Update(ctx, &darkroom)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *DarkroomReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&deploymentsv1alpha1.Darkroom{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
