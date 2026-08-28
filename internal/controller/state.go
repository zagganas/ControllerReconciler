package controller

import (
	"context"
	"time"

	"github.com/zagganas/ControllerReconciler/api/config/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	stateSchema = schema.GroupVersionKind{
		Group:   "config.zagganas.com",
		Version: "v1alpha1",
		Kind:    "State",
	}
)

func getStateByName(
	ctx context.Context,
	client client.Client,
	state *unstructured.Unstructured,
	name string,
	namespace string,
) (bool, error) {

	state.SetGroupVersionKind(stateSchema)
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	if err := client.Get(ctx, namespacedName, state); err != nil {
		if apierrors.IsNotFound(err) {
			// log.Info(ctx, "Controller state not found", "controller", namespacedName.Name)
			return false, nil
		}
		return false, err
	}

	return true, nil

}

func constructStateUnstructured(
	name string,
	namespace string,
	gvk metav1.GroupVersionKind,
	owner *v1alpha1.Controller,
	scheme *runtime.Scheme,
	active bool,
	hash string,
) *unstructured.Unstructured {
	object := &unstructured.Unstructured{}

	// finalizers := []string{finalizer}
	fields := map[string]interface{}{
		"spec": map[string]interface{}{
			"group":   gvk.Group,
			"kind":    gvk.Kind,
			"version": gvk.Version,
			"active":  active,
			"hash":    hash,
		},
	}
	object.SetUnstructuredContent(fields)
	object.SetGroupVersionKind(stateSchema)
	object.SetName(name)
	object.SetNamespace(namespace)
	controllerutil.SetControllerReference(owner, object, scheme)

	return object
}

func applyState(ctx context.Context,
	c client.Client,
	state *unstructured.Unstructured,
	fieldManager string,
) error {
	// Set the managed fields list into nil
	// because otherwise, the APIserver is complaining
	state.SetManagedFields(nil)
	applyConfig := client.ApplyConfigurationFromUnstructured(state)
	applyOptions := &client.ApplyOptions{
		FieldManager: fieldManager,
		Force:        func() *bool { b := true; return &b }(),
	}

	if err := c.Apply(ctx, applyConfig, applyOptions); err != nil {
		return err
	}

	return nil
}

func applyStateStatus(ctx context.Context,
	c client.Client,
	state *unstructured.Unstructured,
	fieldManager string,
) error {
	object := &unstructured.Unstructured{}

	// finalizers := []string{finalizer}
	fields := map[string]interface{}{
		"status": map[string]interface{}{
			"lastTransitionTime": time.Now().UnixMicro(),
		},
	}
	object.SetUnstructuredContent(fields)
	object.SetGroupVersionKind(stateSchema)
	object.SetName(state.GetName())
	object.SetNamespace(state.GetNamespace())

	applyConfig := client.ApplyConfigurationFromUnstructured(object)
	applyOptions := &client.ApplyOptions{
		FieldManager: fieldManager,
		Force:        func() *bool { b := true; return &b }(),
	}
	subApplyOptions := &client.SubResourceApplyOptions{
		ApplyOptions: *applyOptions,
	}
	if err := c.Status().Apply(ctx, applyConfig, subApplyOptions); err != nil {
		return err
	}

	return nil
}
