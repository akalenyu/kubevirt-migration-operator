package controller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	migrationsv1alpha1 "kubevirt.io/kubevirt-migration-operator/api/v1alpha1"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// watch registers specific watches
func (r *MigControllerReconciler) watch() error {
	if err := r.reconciler.WatchResourceTypes(&corev1.ConfigMap{}); err != nil {
		return err
	}

	return nil
}

// preCreate creates the operator config map
func (r *MigControllerReconciler) preCreate(cr client.Object) error {
	// claim the configmap
	if err := r.createOperatorConfig(cr); err != nil {
		return err
	}
	return nil
}

// checkSanity verifies whether config map exists and is in proper relation with the cr
func (r *MigControllerReconciler) checkSanity(cr client.Object, reqLogger logr.Logger) (*reconcile.Result, error) {
	configMap, err := r.getConfigMap()
	if err != nil {
		return &reconcile.Result{}, err
	}
	if !metav1.IsControlledBy(configMap, cr) {
		ownerDeleted, err := r.configMapOwnerDeleted(configMap)
		if err != nil {
			return &reconcile.Result{}, err
		}

		if ownerDeleted || configMap.DeletionTimestamp != nil {
			reqLogger.Info("Waiting for kubevirt-migration-controller-config to be deleted before reconciling", "MigController", cr.GetName())
			return &reconcile.Result{RequeueAfter: time.Second}, nil
		}

		reqLogger.Info("Reconciling to error state, unwanted MigController object")
		result, err := r.reconciler.ReconcileError(cr, "Reconciling to error state, unwanted MigController object")
		return &result, err
	}
	return nil, nil
}

func (r *MigControllerReconciler) configMapOwnerDeleted(cm *corev1.ConfigMap) (bool, error) {
	ownerRef := metav1.GetControllerOf(cm)
	if ownerRef != nil {
		if ownerRef.Kind != "MigController" {
			return false, fmt.Errorf("unexpected configmap owner kind %q", ownerRef.Kind)
		}

		owner := &migrationsv1alpha1.MigController{}
		if err := r.Client.Get(context.TODO(), client.ObjectKey{Name: ownerRef.Name}, owner); err != nil {
			if errors.IsNotFound(err) {
				return true, nil
			}

			return false, err
		}

		if owner.DeletionTimestamp == nil && owner.UID == ownerRef.UID {
			return false, nil
		}
	}

	return true, nil
}

func (r *MigControllerReconciler) registerHooks() {
	r.reconciler.
		WithPreCreateHook(r.preCreate).
		WithWatchRegistrator(r.watch).
		WithSanityChecker(r.checkSanity)
}
