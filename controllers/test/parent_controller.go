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

package test

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	testv1 "github.com/kkohtaka/kevents/apis/test/v1"
)

// ParentReconciler reconciles a Parent object
type ParentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=test.kkohtaka.org,resources=parents,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=test.kkohtaka.org,resources=parents/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=test.kkohtaka.org,resources=parents/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Parent object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *ParentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var parent testv1.Parent
	if err := r.Client.Get(ctx, req.NamespacedName, &parent); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{Requeue: true}, fmt.Errorf("couldn't get Parent %s: %w", req.NamespacedName, err)
	}

	parent.Status.Phase = testv1.ParentPhaseRunning
	if err := r.Client.Status().Update(
		ctx, &parent, &client.UpdateOptions{}); err != nil {
		return ctrl.Result{Requeue: true},
			fmt.Errorf("couldn't update status of Parent %s: %w", req.NamespacedName, err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ParentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&testv1.Parent{}).
		Complete(r)
}
