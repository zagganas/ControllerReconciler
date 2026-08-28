/*
Copyright 2026.

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

package controller

import (
	"context"

	"github.com/golang/gddo/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	util "github.com/zagganas/ControllerReconciler/internal/util"
)

// ResourceReconciler reconciles a OnDemand object
type ResourceReconciler struct {
	controllerName      *string
	controllerNamespace *string
	GVK                 metav1.GroupVersionKind
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=config.zagganas.com,resources=states,verbs=get;list;watch
// +kubebuilder:rbac:groups=core.dynamic.zagganas.com,resources=dynamicmessagejobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OnDemand object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.24.1/pkg/reconcile
func (r *ResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// Get state for resource reconciler
	// state := unstructured.Unstructured{}

	// _, stateErr := getStateByName(ctx, r.Client, &state, *r.controllerName, *r.controllerNamespace)
	// // If unable to get the state due to an error, which is not apiNotFound, then reschedule
	// if stateErr != nil {
	// 	log.Error(ctx, "Unable to get state for controller", "controller", *r.controllerName, "resource", req.NamespacedName, "error", stateErr)
	// 	return ctrl.Result{}, stateErr
	// }
	// active, found, err := unstructured.NestedBool(state.Object, "spec", "active")
	// log.Error(ctx, "printing", "active", active, "found", found, "err", err)

	// if active, found, err := unstructured.NestedBool(state.Object, "spec", "active"); err != nil || (!found) || (!active && found) {
	// 	log.Info(ctx, "Controller is marked as inactive. Reconciliation will not be performed", "controller", *r.controllerName, "resource", req.NamespacedName)
	// 	return ctrl.Result{}, nil
	// }

	// Start reconciliation
	log.Info(ctx, "Started controller reconciliation", "controller", *r.controllerName)

	resource := &unstructured.Unstructured{}
	if err := util.GetResourceWithGKV(ctx, r.Client, r.GVK, req, resource); err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then it usually means that it was deleted or not created
			// Stop the reconciliation
			log.Info(ctx, r.GVK.Kind+" resource "+req.Name+" not found. Ignoring since object must be deleted or not created")
			return ctrl.Result{}, nil
		}
	}

	// Do normal reconciliation stuff here (composite-style)
	log.Info(ctx, "Resource Found", "resource", resource.GetAPIVersion(), "namespacedname", req.NamespacedName)

	return ctrl.Result{}, nil
}

func (r *ResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// if err := mgr.GetFieldIndexer().IndexField(context.Background(), &batchv1.Job{}, ".metadata.labels", SetupFieldIndexer); err != nil {
	// 	return err
	// }
	primary := &unstructured.Unstructured{}
	primary.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   r.GVK.Group,
		Kind:    r.GVK.Kind,
		Version: r.GVK.Version,
	})
	return ctrl.NewControllerManagedBy(mgr).
		For(primary). // Watch the primary resource
		Named(*r.controllerName).
		Complete(r)
}
