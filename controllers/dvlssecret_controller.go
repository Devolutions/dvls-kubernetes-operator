/*
Copyright 2023.

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

package controllers

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	dvlsv1alpha1 "github.com/Devolutions/dvls-kubernetes-operator/api/v1alpha1"

	"github.com/Devolutions/go-dvls"
)

var (
	DvlsClient      dvls.Client
	RequeueDuration time.Duration
)

const (
	DefaultRequeueDuration time.Duration = time.Minute

	dvlsSecretType string = "devolutions.com/dvlssecret"

	statusAvailableDvlsSecret string = "Available"
	statusDegradedDvlsSecret  string = "Degraded"
)

// DvlsSecretReconciler reconciles a DvlsSecret object
type DvlsSecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=dvls.devolutions.com,resources=dvlssecrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dvls.devolutions.com,resources=dvlssecrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dvls.devolutions.com,resources=dvlssecrets/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DvlsSecret object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *DvlsSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	dvlsSecret := &dvlsv1alpha1.DvlsSecret{}
	err := r.Get(ctx, req.NamespacedName, dvlsSecret)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("DvlsSecret object not found. Ignoring event since object must be deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get DvlsSecret object, %w", err)
	}

	if dvlsSecret.Status.Conditions == nil || len(dvlsSecret.Status.Conditions) == 0 || dvlsSecret.Status.EntryModifiedDate.IsZero() {
		meta.SetStatusCondition(&dvlsSecret.Status.Conditions, v1.Condition{Type: statusAvailableDvlsSecret, Status: v1.ConditionUnknown, Reason: "Reconciling"})
		dvlsSecret.Status.EntryModifiedDate = v1.Date(0001, time.January, 1, 1, 1, 1, 1, time.UTC)
		if err := r.Status().Update(ctx, dvlsSecret); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update DvlsSecret status, %w", err)
		}

		err := r.Get(ctx, req.NamespacedName, dvlsSecret)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to get DvlsSecret object, %w", err)
		}
	}

	entry, err := DvlsClient.GetEntry(dvlsSecret.Spec.EntryID)
	if err != nil {
		log.Error(err, "unable to fetch dvls entry", "entryId", dvlsSecret.Spec.EntryID)
		meta.SetStatusCondition(&dvlsSecret.Status.Conditions, v1.Condition{Type: statusDegradedDvlsSecret, Status: v1.ConditionTrue, Reason: "Reconciling", Message: "Unable to fetch entry on DVLS instance"})
		if err := r.Status().Update(ctx, dvlsSecret); err != nil {
			log.Error(err, "Failed to update DvlsSecret status")
		}
		return ctrl.Result{}, nil
	}

	if entry.ConnectionType != dvls.ServerConnectionCredential || entry.ConnectionSubType != dvls.ServerConnectionSubTypeDefault {
		log.Error(err, "entry type not supported, only username/password entries are supported", "entryId", dvlsSecret.Spec.EntryID, "entryType", entry.ConnectionType, "entrySubType", entry.ConnectionSubType)
		meta.SetStatusCondition(&dvlsSecret.Status.Conditions, v1.Condition{Type: statusDegradedDvlsSecret, Status: v1.ConditionTrue, Reason: "Reconciling", Message: "Entry type not supported, only username/password entries are supported"})
		if err := r.Status().Update(ctx, dvlsSecret); err != nil {
			log.Error(err, "Failed to update DvlsSecret status")
		}
		return ctrl.Result{}, nil
	}

	kSecret := &corev1.Secret{}
	err = r.Get(ctx, req.NamespacedName, kSecret)
	if err != nil && !apierrors.IsNotFound(err) {
		return ctrl.Result{}, fmt.Errorf("failed to get kubernetes secret object, %w", err)
	}
	kSecretNotFound := apierrors.IsNotFound(err)

	var entryTime, secretTime time.Time
	if !dvlsSecret.Status.EntryModifiedDate.IsZero() && entry.ModifiedDate != nil {
		secretTime = dvlsSecret.Status.EntryModifiedDate.Time
		entryTime = entry.ModifiedDate.Time
	}

	if entryTime.Equal(secretTime) && !kSecretNotFound {
		return ctrl.Result{
			RequeueAfter: RequeueDuration,
		}, nil
	}

	secret, err := DvlsClient.GetEntryCredentialsPassword(entry)
	if err != nil {
		log.Error(err, "unable to fetch dvls secret", "entryId", dvlsSecret.Spec.EntryID)
		meta.SetStatusCondition(&dvlsSecret.Status.Conditions, v1.Condition{Type: statusDegradedDvlsSecret, Status: v1.ConditionTrue, Reason: "Reconciling", Message: "Unable to fetch secret on DVLS instance"})
		if err := r.Status().Update(ctx, dvlsSecret); err != nil {
			log.Error(err, "Failed to update DvlsSecret status")
		}
		return ctrl.Result{}, nil
	}
	secretMap := make(map[string]string)
	secretMap["entry-id"] = secret.ID
	secretMap["entry-name"] = secret.EntryName
	secretMap["username"] = secret.Credentials.Username
	if secret.Credentials.Password != nil {
		secretMap["password"] = *secret.Credentials.Password
	}

	if kSecretNotFound {
		log.Info("Kubernetes secret not found, creating")
		kSecret.ObjectMeta = v1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		}
		err := ctrl.SetControllerReference(dvlsSecret, kSecret, r.Scheme)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to set kubernetes secret owner, %w", err)
		}

		kSecret.Type = corev1.SecretType(dvlsSecretType)
		kSecret.StringData = secretMap

		err = r.Create(ctx, kSecret)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to create kubernetes secret, %w", err)
		}

		return ctrl.Result{}, nil
	}

	var owned bool
	kSecretOwner := kSecret.GetOwnerReferences()
	for _, v := range kSecretOwner {
		if v.UID == dvlsSecret.GetUID() {
			owned = true
		}
	}

	if kSecret.Type != corev1.SecretType(dvlsSecretType) || !owned {
		return ctrl.Result{}, fmt.Errorf("found existing kubernetes secret with name %s in namespace %s but is either not the correct type or not owned by the DvlsSecret resource. Either delete the existing secret or use a different name", kSecret.GetName(), kSecret.GetNamespace())
	}

	kSecret.StringData = secretMap
	err = r.Update(ctx, kSecret)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get update kubernetes secret object, %w", err)
	}
	log.Info("updated secret")
	err = r.Get(ctx, req.NamespacedName, dvlsSecret)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get DvlsSecret object, %w", err)
	}

	meta.SetStatusCondition(&dvlsSecret.Status.Conditions, v1.Condition{Type: statusAvailableDvlsSecret, Status: v1.ConditionTrue, Reason: "Reconciling"})
	meta.RemoveStatusCondition(&dvlsSecret.Status.Conditions, statusDegradedDvlsSecret)
	dvlsSecret.Status.EntryModifiedDate = v1.NewTime(entryTime)

	if err := r.Status().Update(ctx, dvlsSecret); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update DvlsSecret status, %w", err)
	}

	return ctrl.Result{
		RequeueAfter: RequeueDuration,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DvlsSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dvlsv1alpha1.DvlsSecret{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
