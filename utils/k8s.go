package kube

import (
	"context"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

func DynamicK8s(operating string, data []byte) error {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	// 1. Prepare a RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
	// 2. Prepare the dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}
	// 3. Decode YAML manifest into unstructured.Unstructured
	runtimeObject, gvk, err :=
		yaml.
			NewDecodingSerializer(unstructured.UnstructuredJSONScheme).
			Decode(data, nil, nil)
	if err != nil {
		return err
	}
	unstructuredObj := runtimeObject.(*unstructured.Unstructured)
	// 4. Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}
	// 5. Obtain REST interface for the GVR
	var resourceREST dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		resourceREST = dynamicClient.Resource(mapping.Resource).Namespace(unstructuredObj.GetNamespace())
	} else {
		// for cluster-wide resources
		resourceREST = dynamicClient.Resource(mapping.Resource)
	}
	switch operating {
	case "create":
		_, err = resourceREST.Create(context.TODO(), unstructuredObj, metav1.CreateOptions{})
	case "delete":
		deletePolicy := metav1.DeletePropagationForeground
		deleteOptions := metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		}
		err = resourceREST.Delete(context.TODO(), unstructuredObj.GetName(), deleteOptions)
	}

	return err
}
