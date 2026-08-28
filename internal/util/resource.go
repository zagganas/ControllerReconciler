package util

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetResourceWithGKV(ctx context.Context, client client.Client, gkv metav1.GroupVersionKind, req ctrl.Request, resource *unstructured.Unstructured) error {

	resource.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   gkv.Group,
		Kind:    gkv.Kind,
		Version: gkv.Version,
	})

	if err := client.Get(context.Background(), req.NamespacedName, resource); err != nil {
		return err
	}

	return nil

}
