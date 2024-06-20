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
	return r.validateIngressHost(ctx, nil, obj.(*networkv1.Ingress))
}

func (r *IngressWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	klog.V(4).Infof("validate update ingress: %v", newObj)
	return r.validateIngressHost(ctx, oldObj.(*networkv1.Ingress), newObj.(*networkv1.Ingress))
}

func (r *IngressWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (r *IngressWebhook) validateIngressHost(ctx context.Context, oldIngress, newIngress *networkv1.Ingress) (admission.Warnings, error) {
	settingList := v1alpha2.ClusterIngressSettingList{}
	if err := r.List(ctx, &settingList); err != nil {
		return nil, err
	}

	var g glob.Glob
	var err error
	var domain string

	for ruleIndex, rule := range newIngress.Spec.Rules {
		domain = rule.Host
		for i := range settingList.Items {
			if settingList.Items[i].Spec.UniqueDomainPattern != "" {
				pattern := settingList.Items[i].Spec.UniqueDomainPattern
				g = glob.MustCompile(pattern, '.')
				if g.Match(domain) {
					// for all ingress in the cluster
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

					// for current ingress
					if oldIngress != nil {
						// for update
						// if the host has not changed or part of the host has been removed, the verification will be skipped.
						oldHosts := sets.New[string]()
						newHosts := sets.New[string]()
						for _, r := range oldIngress.Spec.Rules {
							oldHosts.Insert(r.Host)
						}
						for _, r := range newIngress.Spec.Rules {
							// cannot be duplicates in newHosts
							if newHosts.Has(r.Host) {
								return nil, fmt.Errorf("duplicate host %s in the current ingress", r.Host)
							}
							newHosts.Insert(r.Host)
						}
						if newHosts.IsSuperset(oldHosts) {
							return nil, nil
						}
					}

					for i, rule := range newIngress.Spec.Rules {
						if ruleIndex != i && g.Match(rule.Host) {
							return nil, fmt.Errorf("Restrict the use of %s, the current ingress", domain)
						}
					}
				}
			}
		}
	}

	klog.V(4).Infof("validate ingress host success: %v", newIngress)
	return nil, nil
}

func (r *IngressWebhook) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		WithValidator(r).
		For(&networkv1.Ingress{}).
		Complete()
}
