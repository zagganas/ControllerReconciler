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
	"strings"
	"time"

	"github.com/golang/gddo/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1alpha1 "github.com/zagganas/ControllerReconciler/api/config/v1alpha1"
	"github.com/zagganas/ControllerReconciler/internal/util"
)

const (
	// typeAvailableOnDemand represents the status of the OnDemand reconciliation
	typeAvailable = "Available"
	// typeProgressingOnDemand represents the status used when the OnDemand is being reconciled
	typeProgressing = "Progressing"
	// typeDegradedOnDemand represents the status used when the OnDemand has encountered an error
	typeDegraded = "Degraded"
	// typeReadydOnDemand represents the status used when the OnDemand has succeeded
	typeReady = "Ready"
)

var (
	generationPredicate = predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},

		// Allow delete events
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},

		// Allow generic events (e.g., external triggers)
		GenericFunc: func(e event.GenericEvent) bool {
			return true
		},
	}
)

// ControllerReconciler reconciles a OnDemand object
type ControllerReconciler struct {
	Manager           ctrl.Manager
	ManagerCancelFunc context.CancelFunc
	ReconcilerAddTime int64
	client.Client
	Scheme *runtime.Scheme
}

const (
	finalizerName  = "config.zagganas.com/controllerReconciler"
	controllerName = "Controller"
)

// +kubebuilder:rbac:groups=config.zagganas.com,resources=controllers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=config.zagganas.com,resources=controllers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=config.zagganas.com,resources=controllers/finalizers,verbs=update
// +kubebuilder:rbac:groups=config.zagganas.com,resources=states,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=config.zagganas.com,resources=states/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=config.zagganas.com,resources=states/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OnDemand object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.24.1/pkg/reconcile
func (r *ControllerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	var controller v1alpha1.Controller
	if err := r.Get(ctx, req.NamespacedName, &controller); err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then it usually means that it was deleted or not created
			// Stop the reconciliation
			log.Info(ctx, "Controller deleted. Stopping manager...")
			// return ctrl.Result{}, nil
			//maybe gracefully stop the manager here to force the operator to restart
			r.ManagerCancelFunc()

			return ctrl.Result{}, nil
		}
	}

	// State will be queried by the respective resource reconciler
	// If it the boolean active field is false, then
	// on-demand resource reconciliation will stop, because we assume that the
	// controller resource was deleted.
	// The state will always be created in the same namespace as the controller definition
	// and has the name of the reconciler that was just created.
	// This technique is used as a deleted controller & relevant watchers cannot be
	// stopped inside a running manager, once they have been registered.
	// Additionally, no custom resources will be registered until this reconciliation has
	// successfully finished

	stateName := req.NamespacedName.Name
	stateNamespace := req.NamespacedName.Namespace
	state := &unstructured.Unstructured{}

	// Get State
	stateFound, stateErr := getStateByName(ctx, r.Client, state, stateName, stateNamespace)
	// If unable to get the state due to an error, which is not apiNotFound, then reschedule
	if stateErr != nil {
		log.Error(ctx, "Unable to get state for controller", "controller", req.NamespacedName, "error", stateErr)
		return ctrl.Result{}, stateErr
	}

	// Clean up on deletion; no reason to create the state before
	deletion := !controller.ObjectMeta.DeletionTimestamp.IsZero()
	if deletion {
		// If state exists, remove finalizer and delete it
		// if stateFound {
		// 	//remove finalizer from controller
		// 	state.SetFinalizers([]string{})
		// 	if err := applyState(ctx, r.Client, state, stateName); err != nil {
		// 		log.Error(ctx, "Error creating/modifying the state", "error", err)
		// 		return ctrl.Result{}, err
		// 	}
		// 	if err := r.Delete(ctx, state); err != nil {
		// 		if !apierrors.IsNotFound(err) {
		// 			log.Error(ctx, "Unable to delete resource", "controller", req.NamespacedName, "error", err)
		// 			return ctrl.Result{}, err
		// 		}
		// 	}
		// }

		// Remove finalizer from the controller resource
		controllerutil.RemoveFinalizer(&controller, finalizerName)
		if err := r.Update(ctx, &controller); err != nil {
			log.Error(ctx, "Unable to remove finalizer from controller", "controller", req.NamespacedName, "error", err)
			return ctrl.Result{}, err
		}

		// Allow the object to be deleted before the context is canceled
		return ctrl.Result{RequeueAfter: (time.Millisecond * time.Duration(100))}, nil

	} else {
		// If not deleting, just apply the finalizer on the controller resource if it's not there
		if !controllerutil.ContainsFinalizer(&controller, finalizerName) {
			controllerutil.AddFinalizer(&controller, finalizerName)
			if err := r.Update(ctx, &controller); err != nil {
				log.Error(ctx, "Unable to add finalizer to Controller", "controller", req.NamespacedName, "error", err)
				return ctrl.Result{}, err
			}
		}
	}
	// Get last State transition time (if it exists)
	lastStateTransitionTime, transitionFound, _ := unstructured.NestedInt64(state.Object, "status", "lastTransitionTime")
	// If state does not exist or state is stale, create it with default values
	if !stateFound || (transitionFound && lastStateTransitionTime < r.ReconcilerAddTime) {
		log.Info(ctx, "Controller state stale or non-existent: updating", "controller", req.NamespacedName)
		state = constructStateUnstructured(stateName, stateNamespace, controller.Spec.GVK, &controller, r.Scheme, false, "")
		// Set nested field to created
		if err := applyState(ctx, r.Client, state, stateName); err != nil {
			log.Error(ctx, "Error creating/modifying the state", "error", err, "state", state)
		}
		if err := applyStateStatus(ctx, r.Client, state, controller.ObjectMeta.Name); err != nil {
			log.Error(ctx, "Error updating the state status", "error", err)
		}
	}

	// Check if controller spec was modified
	controllerSpecStr, err := controller.GetSpecBytes()
	if err != nil {
		log.Error(ctx, "Error calculating controller spec bytestring", "error", err)

		return ctrl.Result{}, err
	}
	controllerSpecHash := util.CalculateMd5(controllerSpecStr)

	stateSpecHash, found, err := unstructured.NestedString(state.Object, "spec", "hash")
	if !found || err != nil {
		log.Error(ctx, "Error getting hash field from state", "found", found, "error", err)
	}

	// If controller was modified, cancel the manager context and restart operator
	if shouldRestartManager(controllerSpecHash, stateSpecHash) {
		log.Info(ctx, "Controller modified. Stopping manager...")
		r.ManagerCancelFunc()

		return ctrl.Result{}, nil
	}

	// Register controller with manager
	if err := r.RegisterNewControllerWithManager(&controller); err != nil {
		// Is there a better way to check the error?
		if !strings.Contains(err.Error(), "already exists") {
			log.Error(ctx, "Error registering controller", "controller", req.NamespacedName)
		}
	}
	// Update the state to the correct active status
	state = constructStateUnstructured(stateName, stateNamespace, controller.Spec.GVK, &controller, r.Scheme, true, controllerSpecHash)
	if err := applyState(ctx, r.Client, state, controller.ObjectMeta.Name); err != nil {
		log.Error(ctx, "Error creating/modifying the state", "error", err)
	}
	if err := applyStateStatus(ctx, r.Client, state, controller.ObjectMeta.Name); err != nil {
		log.Error(ctx, "Error updating the state status", "error", err)
	}

	if updErr := r.updateResourceStatusConditions(ctx, &controller, typeAvailable, "Active", "Successfully registered controller"); updErr != nil {
		return ctrl.Result{}, updErr
	}

	return ctrl.Result{}, nil
}

func (r *ControllerReconciler) updateResourceStatusConditions(
	ctx context.Context,
	controller *v1alpha1.Controller,
	conditionType string,
	reason string,
	message string,
) error {
	// Set status condition
	meta.SetStatusCondition(&controller.Status.Conditions, metav1.Condition{
		Type:    conditionType,
		Status:  metav1.ConditionTrue,
		Reason:  reason,
		Message: message,
	})
	// Update the controller resource status
	if statusUpdateErr := r.Status().Update(ctx, controller); statusUpdateErr != nil {
		log.Error(ctx, "Unable to update Controller status conditions", "error", statusUpdateErr, "controller", controller.ObjectMeta.Name, "reason", reason)
		return statusUpdateErr
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ControllerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Controller{}).
		Named(controllerName).
		Watches(
			// Watch state resources:
			// Since states have the same name as the controller,
			// The same name and namespace can be used to set up a new request
			&v1alpha1.State{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				return []reconcile.Request{
					{
						NamespacedName: types.NamespacedName{
							Namespace: obj.GetNamespace(),
							Name:      obj.GetName(),
						},
					},
				}
			}),
			builder.WithPredicates(generationPredicate),
		).
		Complete(r)
}

func (r *ControllerReconciler) RegisterNewControllerWithManager(
	controller *v1alpha1.Controller,
) error {
	err := (&ResourceReconciler{
		controllerName:      &controller.ObjectMeta.Name,
		controllerNamespace: &controller.ObjectMeta.Namespace,
		GVK: metav1.GroupVersionKind{
			Group:   controller.Spec.GVK.Group,
			Version: controller.Spec.GVK.Version,
			Kind:    controller.Spec.GVK.Kind,
		},
		Client: r.Manager.GetClient(),
		Scheme: r.Manager.GetScheme(),
	}).SetupWithManager(r.Manager)

	return err
}

func shouldRestartManager(controllerHash string, stateHash string) bool {
	if stateHash != "" && controllerHash != stateHash {
		return true
	}

	return false
}
