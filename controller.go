// controller/namespaced_controller.go
package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// NamespacedController watches namespaces with dynamic label/annotation values
type NamespacedController struct {
	client client.Client
}

// Reconcile handles namespace reconciliation
func (r *NamespacedController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Get the namespace
	ns := &v1.Namespace{}
	if err := r.client.Get(ctx, req.NamespacedName, ns); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get namespace: %w", err)
	}

	// Check if namespace matches our dynamic criteria
	predicate := NewDynamicNamespacePredicate(
		"app",
		[]string{"staging", "production"},
		"env",
		[]string{"staging", "production"},
	)
	if !predicate.CheckNamespace(ns) {
		return ctrl.Result{}, nil // Skip if doesn't match
	}

	// Real-world action: Reconcile resources in this namespace
	log.Info("Reconciling resources in namespace with dynamic values", "namespace", ns.Name)
	// In a real operator: Deploy resources, check health, etc.
	return ctrl.Result{}, nil
}

// SetupWithManager registers the controller
func (r *NamespacedController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Namespace{}).
		WithPredicates(predicate.Funcs{
			AddFunc: func(obj client.Object) bool {
				return true // We'll filter in Reconcile
			},
			UpdateFunc: func(oldObj, newObj client.Object) bool {
				return true // We'll filter in Reconcile
			},
		}).
		Complete(r)
}

// DynamicNamespacePredicate handles dynamic values for labels/annotations
type DynamicNamespacePredicate struct {
	LabelKey      string
	LabelValues   []string
	AnnotationKey string
	AnnotationValues []string
}

// NewDynamicNamespacePredicate creates a predicate that matches ANY value in the list
func NewDynamicNamespacePredicate(
	labelKey, annotationKey string,
	labelValues, annotationValues []string,
) *DynamicNamespacePredicate {
	return &DynamicNamespacePredicate{
		LabelKey:      labelKey,
		LabelValues:   labelValues,
		AnnotationKey: annotationKey,
		AnnotationValues: annotationValues,
	}
}

// CheckNamespace verifies if namespace matches ANY value in the list
func (p *DynamicNamespacePredicate) CheckNamespace(ns *v1.Namespace) bool {
	// Check labels
	if p.LabelKey != "" {
		labelValue := ns.Labels[p.LabelKey]
		for _, val := range p.LabelValues {
			if labelValue == val {
				return true
			}
		}
	}

	// Check annotations
	if p.AnnotationKey != "" {
		annotationValue := ns.Annotations[p.AnnotationKey]
		for _, val := range p.AnnotationValues {
			if annotationValue == val {
				return true
			}
		}
	}

	return false
}
