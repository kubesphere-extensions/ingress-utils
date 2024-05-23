package controller

import (
	"context"
	"fmt"

	"github.com/gobwas/glob"
	"github.com/kubesphere-extensions/ingress-utils/pkg/api/gateway/v1alpha2"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/klog/v2"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ admission.CustomValidator = &IngressWebhook{}

type IngressWebhook struct {
	client.Client
}

func (r *IngressWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	klog.V(4).Infof("validate create ingress: %v", obj)
	return r.validateIngressHost(ctx, obj.(*networkv1.Ingress))
}

func (r *IngressWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	klog.V(4).Infof("validate update ingress: %v", newObj)
	// If the host has not changed or part of the host has been removed, the verification will be skipped.
	oldIngress := oldObj.(*networkv1.Ingress)
	newIngress := newObj.(*networkv1.Ingress)
	oldHosts := sets.New[string]()
	newHosts := sets.New[string]()
	for _, rule := range oldIngress.Spec.Rules {
		oldHosts.Insert(rule.Host)
	}
	for _, rule := range newIngress.Spec.Rules {
		// cannot be duplicates in newHosts
		if newHosts.Has(rule.Host) {
			return nil, fmt.Errorf("duplicate host %s in the current ingress", rule.Host)
		}
		newHosts.Insert(rule.Host)
	}
	if newHosts.IsSuperset(oldHosts) {
		return nil, nil
	}

	return r.validateIngressHost(ctx, newObj.(*networkv1.Ingress))
}

func (r *IngressWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (r *IngressWebhook) validateIngressHost(ctx context.Context, ingress *networkv1.Ingress) (admission.Warnings, error) {
	settingList := v1alpha2.ClusterIngressSettingList{}
	if err := r.List(ctx, &settingList); err != nil {
		return nil, err
	}

	var g glob.Glob
	var err error
	var domain string

	for _, rule := range ingress.Spec.Rules {
		domain = rule.Host
		for i := range settingList.Items {
			if settingList.Items[i].Spec.UniqueDomainPattern != "" {
				pattern := settingList.Items[i].Spec.UniqueDomainPattern
				g = glob.MustCompile(pattern, '.')
				if g.Match(domain) {
					ingressList := networkv1.IngressList{}
					if err = r.List(ctx, &ingressList); err != nil {
						return nil, err
					}

					for _, ingress := range ingressList.Items {
						for _, rule := range ingress.Spec.Rules {
							if g.Match(rule.Host) {
								return nil, fmt.Errorf("Restrict the use of %s, the existing ingress  %s/%s host name is %s", domain, ingress.Namespace, ingress.Name, rule.Host)
							}
						}
					}
				}
			}
		}
	}

	klog.V(4).Infof("validate ingress host success: %v", ingress)
	return nil, nil
}

func (r *IngressWebhook) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		WithValidator(r).
		For(&networkv1.Ingress{}).
		Complete()
}
