/*
Copyright 2025.

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

package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "github.com/blokovi/kubebuilder-configmap-sync-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigMapSyncReconciler reconciles a ConfigMapSync object
type ConfigMapSyncReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.blokovi.com,resources=configmapsyncs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.blokovi.com,resources=configmapsyncs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.blokovi.com,resources=configmapsyncs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ConfigMapSync object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile

// Reconcile se pozove svaki put kada se kreira/promeni/obrise bilo koj CR koji je tipa configMapSync
func (r *ConfigMapSyncReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx) //dal ovde da stavim log :=, bilo je _ = logf...., u primeru je log := r.Log.WithValues("configmapsync", req.NamespacedName)

	// TODO(user): your logic here

	configMapSync := &appsv1.ConfigMapSync{}
	log.Info("IVKE", "req.NamespacedName", req.NamespacedName)
	fmt.Printf("%+v\n", req.NamespacedName)
	// ovde se ucita CR tipa configMapSync koji je Kubernetes event loop prosledio kontroleru (neki koji je upravo kreiran/promenjen/obrisan)
	if err := r.Get(ctx, req.NamespacedName, configMapSync); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// Fetch the source ConfigMap
	sourceConfigMap := &corev1.ConfigMap{}
	sourceConfigMapName := types.NamespacedName{
		Namespace: configMapSync.Spec.SourceNamespace,
		Name:      configMapSync.Spec.ConfigMapName,
	}
	log.Info("IVKE2", "SOURCE Namespace", configMapSync.Spec.SourceNamespace)

	if err := r.Get(ctx, sourceConfigMapName, sourceConfigMap); err != nil {
		return ctrl.Result{}, err
	}
	// Create or Update the destination ConfigMap in the target namespace
	destinationConfigMap := &corev1.ConfigMap{}
	destinationConfigMapName := types.NamespacedName{
		Namespace: configMapSync.Spec.DestinationNamespace,
		Name:      configMapSync.Spec.ConfigMapName,
	}
	if err := r.Get(ctx, destinationConfigMapName, destinationConfigMap); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating ConfigMap in destination namespace", "Namespace", configMapSync.Spec.DestinationNamespace)
			destinationConfigMap = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      configMapSync.Spec.ConfigMapName,
					Namespace: configMapSync.Spec.DestinationNamespace,
				},
				Data: sourceConfigMap.Data, // Copy data from source to destination
			}
			if err := r.Create(ctx, destinationConfigMap); err != nil {
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, err
		}
	} else {
		log.Info("Updating ConfigMap in destination namespace", "Namespace", configMapSync.Spec.DestinationNamespace)
		destinationConfigMap.Data = sourceConfigMap.Data // Update data from source to destination
		if err := r.Update(ctx, destinationConfigMap); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigMapSyncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.ConfigMapSync{}).
		Named("configmapsync").
		Complete(r)
}
