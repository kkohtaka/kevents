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
	"math/rand"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
//+kubebuilder:rbac:groups=test.kkohtaka.org,resources=children,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=test.kkohtaka.org,resources=children/status,verbs=get

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
	logger := log.FromContext(ctx).WithValues("parent", req.NamespacedName)

	var parent testv1.Parent
	if err := r.Get(ctx, req.NamespacedName, &parent); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "unable to fetch Parent")
		return ctrl.Result{}, err
	}

	if parent.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&parent, finalizerName) {
			controllerutil.AddFinalizer(&parent, finalizerName)
			if err := r.Update(ctx, &parent); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	} else {
		if controllerutil.ContainsFinalizer(&parent, finalizerName) {
			if err := r.deleteExternalResources(ctx, &parent); err != nil {
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(&parent, finalizerName)
			if err := r.Update(ctx, &parent); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if err := r.createExternalResources(ctx, &parent); err != nil {
		return ctrl.Result{}, err
	}

	return r.updateStatus(ctx, &parent)
}

func generateChildNamePrefix(p *testv1.Parent) string {
	return p.Name + "-"
}

func generateChildLabels(p *testv1.Parent) map[string]string {
	return map[string]string{
		ownerLabelKey: p.Name,
	}
}

func generateChildSpec(p *testv1.Parent) testv1.ChildSpec {
	return testv1.ChildSpec{}
}

func generateChildSelector(p *testv1.Parent) (labels.Selector, error) {
	req, err := labels.NewRequirement(ownerLabelKey, selection.Equals, []string{p.Name})
	if err != nil {
		return nil, fmt.Errorf("couldn't create a requirement of label selection: %w", err)
	}
	return labels.NewSelector().Add(*req), nil
}

func generateAnnotationsForChild() map[string]string {
	duration := time.Second * time.Duration(rand.Intn(30))
	return map[string]string{
		updatingUntilAnnotationKey: time.Now().Add(duration).Format(time.RFC3339),
	}
}

func (r *ParentReconciler) createExternalResources(ctx context.Context, p *testv1.Parent) error {
	sel, err := generateChildSelector(p)
	if err != nil {
		return fmt.Errorf("couldn't generate a children selector: %w", err)
	}
	var children testv1.ChildList
	if err := r.List(ctx, &children, &client.ListOptions{LabelSelector: sel}); err != nil {
		return fmt.Errorf("couldn't list children of parent %s/%s: %w", p.Namespace, p.Name, err)
	}

	if len(children.Items) > int(p.Spec.NumChildren) {
		for diff := len(children.Items) - int(p.Spec.NumChildren); diff > 0; diff-- {
			if err := r.Delete(ctx, &children.Items[diff-1]); err != nil {
				return fmt.Errorf("couldn't delete child %s/%s: %w",
					children.Items[diff-1].Namespace,
					children.Items[diff-1].Name,
					err,
				)
			}
		}
	} else {
		for diff := int(p.Spec.NumChildren) - len(children.Items); diff > 0; diff-- {
			child := testv1.Child{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: generateChildNamePrefix(p),
					Namespace:    p.Namespace,
					Labels:       generateChildLabels(p),
					Annotations:  generateAnnotationsForChild(),
				},
				Spec: generateChildSpec(p),
			}
			if err := controllerutil.SetControllerReference(p, &child, r.Scheme); err != nil {
				return fmt.Errorf("couldn't set a controller reference on a child: %w", err)
			}
			if err := r.Create(ctx, &child, &client.CreateOptions{}); err != nil {
				return fmt.Errorf("couldn't create a chilld: %w", err)
			}
		}
	}
	return nil
}

func (r *ParentReconciler) deleteExternalResources(ctx context.Context, p *testv1.Parent) error {
	sel, err := generateChildSelector(p)
	if err != nil {
		return fmt.Errorf("couldn't generate a children selector: %w", err)
	}
	if err := r.DeleteAllOf(ctx, &testv1.Child{}, &client.DeleteAllOfOptions{ListOptions: client.ListOptions{
		LabelSelector: sel,
	}}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("couldn't delete children of parent %s/%s: %w", p.Namespace, p.Name, err)
	}
	return nil
}

func (r *ParentReconciler) updateStatus(ctx context.Context, p *testv1.Parent) (ctrl.Result, error) {
	patch := client.MergeFrom(p.DeepCopy())
	if !p.DeletionTimestamp.IsZero() {
		p.Status.Phase = testv1.ParentPhaseDeleting
	} else {
		sel, err := generateChildSelector(p)
		if err != nil {
			return ctrl.Result{Requeue: true},
				fmt.Errorf("couldn't generate a children selector: %w", err)
		}
		var children testv1.ChildList
		if err := r.List(ctx, &children, &client.ListOptions{LabelSelector: sel}); err != nil {
			return ctrl.Result{Requeue: true},
				fmt.Errorf("couldn't list children of parent %s/%s: %w", p.Namespace, p.Name, err)
		}

		runningChildren := 0
		for i := range children.Items {
			if children.Items[i].Status.Phase == testv1.ChildPhaseRunning {
				runningChildren++
			}
		}

		p.Status.Phase = testv1.ParentPhaseRunning
		if runningChildren != int(p.Spec.NumChildren) {
			p.Status.Phase = testv1.ParentPhaseUpdating
		}
	}

	if err := r.Status().Patch(ctx, p, patch); err != nil {
		return ctrl.Result{Requeue: true}, fmt.Errorf("couldn't update status of parent %s/%s: %w", p.Namespace, p.Name, err)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ParentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&testv1.Parent{}).
		Owns(&testv1.Child{}).
		Complete(r)
}
