/*
Copyright (c) 2022 RaptorML authors.

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

package operator

// +kubebuilder:rbac:groups=k8s.raptor.ml,resources=dataconnectors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=k8s.raptor.ml,resources=dataconnectors/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=k8s.raptor.ml,resources=dataconnectors/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

import (
	"context"
	"fmt"
	"time"

	"github.com/raptor-ml/raptor/api"
	raptorApi "github.com/raptor-ml/raptor/api/v1alpha1"
	"github.com/raptor-ml/raptor/pkg/plugins"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// DataConnectorReconciler reconciles a DataConnector object
type DataConnectorReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	CoreAddr string
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *DataConnectorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the Feature definition from the Kubernetes API.
	conn := new(raptorApi.DataConnector)
	err := r.Get(ctx, req.NamespacedName, conn)
	if err != nil {
		logger.Error(err, "Failed to get DataConnector")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if conn.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(conn, finalizerName) {
			controllerutil.AddFinalizer(conn, finalizerName)
			if err := r.Update(ctx, conn); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(conn, finalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if len(conn.Status.Features) > 0 {
				// return with error so that it can be retried
				return ctrl.Result{}, fmt.Errorf("cannot delete DataConnector with associated Features")
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(conn, finalizerName)
			if err := r.Update(ctx, conn); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	if p := plugins.DataConnectorReconciler.Get(conn.Spec.Kind); p != nil {
		if changed, err := p(ctx, r.reconcileRequest(conn)); err != nil {
			return ctrl.Result{}, err
		} else if changed {
			// Ask to requeue after 1 minute in order to give enough time for the
			// pods be created on the cluster side and the operand be able
			// to do the next update step accurately.
			return ctrl.Result{RequeueAfter: time.Minute}, nil
		}
	}

	// Todo change status to ready

	return ctrl.Result{}, nil
}

func (r *DataConnectorReconciler) reconcileRequest(conn *raptorApi.DataConnector) api.ReconcileRequest {
	return api.ReconcileRequest{
		DataConnector: conn,
		Client:        r.Client,
		Scheme:        r.Scheme,
		CoreAddress:   r.CoreAddr,
	}
}

// SetupWithManager sets up the controller with the Controller Manager.
func (r *DataConnectorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(new(raptorApi.DataConnector)).
		Owns(new(appsv1.Deployment)).
		Complete(r)
}
