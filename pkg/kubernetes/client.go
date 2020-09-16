// Code generated by generate-client. DO NOT EDIT.

package kubernetes

import (
	admissionregistrationv1 "github.com/onosproject/helm-client/pkg/kubernetes/admissionregistration/v1"
	apiextensionsv1 "github.com/onosproject/helm-client/pkg/kubernetes/apiextensions/v1"
	apiextensionsv1beta1 "github.com/onosproject/helm-client/pkg/kubernetes/apiextensions/v1beta1"
	appsv1 "github.com/onosproject/helm-client/pkg/kubernetes/apps/v1"
	appsv1beta1 "github.com/onosproject/helm-client/pkg/kubernetes/apps/v1beta1"
	batchv1 "github.com/onosproject/helm-client/pkg/kubernetes/batch/v1"
	batchv1beta1 "github.com/onosproject/helm-client/pkg/kubernetes/batch/v1beta1"
	batchv2alpha1 "github.com/onosproject/helm-client/pkg/kubernetes/batch/v2alpha1"
	"github.com/onosproject/helm-client/pkg/kubernetes/config"
	corev1 "github.com/onosproject/helm-client/pkg/kubernetes/core/v1"
	extensionsv1beta1 "github.com/onosproject/helm-client/pkg/kubernetes/extensions/v1beta1"
	networkingv1beta1 "github.com/onosproject/helm-client/pkg/kubernetes/networking/v1beta1"
	policyv1beta1 "github.com/onosproject/helm-client/pkg/kubernetes/policy/v1beta1"
	rbacv1 "github.com/onosproject/helm-client/pkg/kubernetes/rbac/v1"
	"github.com/onosproject/helm-client/pkg/kubernetes/resource"
	storagev1 "github.com/onosproject/helm-client/pkg/kubernetes/storage/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// New returns a new Kubernetes client for the current namespace
func New() (Client, error) {
	return NewForNamespace(config.GetNamespaceFromEnv())
}

// NewOrDie returns a new Kubernetes client for the current namespace
func NewOrDie() Client {
	client, err := New()
	if err != nil {
		panic(err)
	}
	return client
}

// NewForNamespace returns a new Kubernetes client for the given namespace
func NewForNamespace(namespace string) (Client, error) {
	kubernetesConfig, err := config.GetRestConfig()
	if err != nil {
		return nil, err
	}
	kubernetesClient, err := kubernetes.NewForConfig(kubernetesConfig)
	if err != nil {
		return nil, err
	}
	return &client{
		namespace: namespace,
		config:    kubernetesConfig,
		client:    kubernetesClient,
		filter:    resource.NoFilter,
	}, nil
}

// NewForNamespaceOrDie returns a new Kubernetes client for the given namespace
func NewForNamespaceOrDie(namespace string) Client {
	client, err := NewForNamespace(namespace)
	if err != nil {
		panic(err)
	}
	return client
}

// Client is a Kubernetes client
type Client interface {
	// Namespace returns the client namespace
	Namespace() string

	// Config returns the Kubernetes REST client configuration
	Config() *rest.Config

	// Clientset returns the client's Clientset
	Clientset() *kubernetes.Clientset
	AdmissionregistrationV1() admissionregistrationv1.Client
	ApiextensionsV1() apiextensionsv1.Client
	ApiextensionsV1beta1() apiextensionsv1beta1.Client
	AppsV1() appsv1.Client
	AppsV1beta1() appsv1beta1.Client
	BatchV1() batchv1.Client
	BatchV1beta1() batchv1beta1.Client
	BatchV2alpha1() batchv2alpha1.Client
	ExtensionsV1beta1() extensionsv1beta1.Client
	NetworkingV1beta1() networkingv1beta1.Client
	PolicyV1beta1() policyv1beta1.Client
	RbacV1() rbacv1.Client
	StorageV1() storagev1.Client
	CoreV1() corev1.Client
}

// NewFiltered returns a new filtered Kubernetes client
func NewFiltered(namespace string, filter resource.Filter) (Client, error) {
	kubernetesConfig, err := config.GetRestConfig()
	if err != nil {
		return nil, err
	}
	kubernetesClient, err := kubernetes.NewForConfig(kubernetesConfig)
	if err != nil {
		return nil, err
	}
	return &client{
		namespace: namespace,
		config:    kubernetesConfig,
		client:    kubernetesClient,
		filter:    filter,
	}, nil
}

// NewFilteredOrDie returns a new filtered Kubernetes client
func NewFilteredOrDie(namespace string, filter resource.Filter) Client {
	client, err := NewFiltered(namespace, filter)
	if err != nil {
		panic(err)
	}
	return client
}

type client struct {
	namespace string
	config    *rest.Config
	client    *kubernetes.Clientset
	filter    resource.Filter
}

func (c *client) Namespace() string {
	return c.namespace
}

func (c *client) Config() *rest.Config {
	return c.config
}

func (c *client) Clientset() *kubernetes.Clientset {
	return c.client
}
func (c *client) AdmissionregistrationV1() admissionregistrationv1.Client {
	return admissionregistrationv1.NewClient(c, c.filter)
}

func (c *client) ApiextensionsV1() apiextensionsv1.Client {
	return apiextensionsv1.NewClient(c, c.filter)
}

func (c *client) ApiextensionsV1beta1() apiextensionsv1beta1.Client {
	return apiextensionsv1beta1.NewClient(c, c.filter)
}

func (c *client) AppsV1() appsv1.Client {
	return appsv1.NewClient(c, c.filter)
}

func (c *client) AppsV1beta1() appsv1beta1.Client {
	return appsv1beta1.NewClient(c, c.filter)
}

func (c *client) BatchV1() batchv1.Client {
	return batchv1.NewClient(c, c.filter)
}

func (c *client) BatchV1beta1() batchv1beta1.Client {
	return batchv1beta1.NewClient(c, c.filter)
}

func (c *client) BatchV2alpha1() batchv2alpha1.Client {
	return batchv2alpha1.NewClient(c, c.filter)
}

func (c *client) ExtensionsV1beta1() extensionsv1beta1.Client {
	return extensionsv1beta1.NewClient(c, c.filter)
}

func (c *client) NetworkingV1beta1() networkingv1beta1.Client {
	return networkingv1beta1.NewClient(c, c.filter)
}

func (c *client) PolicyV1beta1() policyv1beta1.Client {
	return policyv1beta1.NewClient(c, c.filter)
}

func (c *client) RbacV1() rbacv1.Client {
	return rbacv1.NewClient(c, c.filter)
}

func (c *client) StorageV1() storagev1.Client {
	return storagev1.NewClient(c, c.filter)
}

func (c *client) CoreV1() corev1.Client {
	return corev1.NewClient(c, c.filter)
}