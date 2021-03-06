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
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	testv1 "github.com/kkohtaka/kevents/apis/test/v1"
)

// ChildReconciler reconciles a Child object
type ChildReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=test.kkohtaka.org,resources=children,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=test.kkohtaka.org,resources=children/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=test.kkohtaka.org,resources=children/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Child object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *ChildReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("child", req.NamespacedName)

	var child testv1.Child
	if err := r.Client.Get(ctx, req.NamespacedName, &child); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "unable to fetch Child")
		return ctrl.Result{}, err
	}

	if child.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&child, finalizerName) {
			controllerutil.AddFinalizer(&child, finalizerName)
			if err := r.Update(ctx, &child); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	} else {
		if controllerutil.ContainsFinalizer(&child, finalizerName) {
			if err := r.deleteExternalResources(ctx, &child); err != nil {
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(&child, finalizerName)
			if err := r.Update(ctx, &child); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if err := r.createExternalResources(ctx, &child); err != nil {
		return ctrl.Result{}, err
	}

	return r.updateStatus(ctx, &child)
}

func (r *ChildReconciler) createExternalResources(ctx context.Context, c *testv1.Child) error {
	return nil
}

func (r *ChildReconciler) deleteExternalResources(ctx context.Context, c *testv1.Child) error {
	return nil
}

func (r *ChildReconciler) updateStatus(ctx context.Context, c *testv1.Child) (ctrl.Result, error) {
	patch := client.MergeFrom(c.DeepCopy())
	if !c.DeletionTimestamp.IsZero() {
		c.Status.Phase = testv1.ChildPhaseDeleting
	} else {
		c.Status.Phase = testv1.ChildPhaseRunning

		updatingUntilStr, ok := c.Annotations[updatingUntilAnnotationKey]
		if ok {
			updatingUntil, err := time.Parse(time.RFC3339, updatingUntilStr)
			if err == nil {
				if time.Now().Before(updatingUntil) {
					c.Status.Phase = testv1.ChildPhaseUpdating
				}
			}
		}
	}

	if err := r.Status().Patch(ctx, c, patch); err != nil {
		return ctrl.Result{Requeue: true},
			fmt.Errorf("couldn't update status of child %s/%s: %w", c.Namespace, c.Name, err)
	}
	if c.Status.Phase == testv1.ChildPhaseUpdating {
		updatingUntilStr, ok := c.Annotations[updatingUntilAnnotationKey]
		if ok {
			updatingUntil, err := time.Parse(time.RFC3339, updatingUntilStr)
			if err == nil {
				return ctrl.Result{
					RequeueAfter: time.Until(updatingUntil),
				}, nil
			}
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ChildReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&testv1.Child{}).
		Complete(r)
}
