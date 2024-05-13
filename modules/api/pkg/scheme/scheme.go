package scheme

import (
	gatewayv1alpha2 "github.com/kubesphere-extensions/ingress-utils/pkg/api/gateway/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

// Scheme contains all types of custom Scheme and kubernetes client-go Scheme.
var Scheme = runtime.NewScheme()

func init() {
	// register common meta types into schemas.
	metav1.AddToGroupVersion(Scheme, metav1.SchemeGroupVersion)

	_ = clientgoscheme.AddToScheme(Scheme)
	_ = gatewayv1alpha2.AddToScheme(Scheme)
}
