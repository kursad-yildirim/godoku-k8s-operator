package controllers

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	tuffv1alpha1 "github.com/tuff/godoku-operator/api/v1alpha1"
)

type GodokuReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=tuff.tripko.local,resources=godokus,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tuff.tripko.local,resources=godokus/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=tuff.tripko.local,resources=godokus/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
func (r *GodokuReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	// Fetch the godoku cr instance. If it does not exist no reconcilation is required.
	godoku := &tuffv1alpha1.Godoku{}
	err := r.Get(ctx, req.NamespacedName, godoku)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("godoku resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get godoku")
		return ctrl.Result{}, err
	}
	// Create godoku deployment it does not exists already
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: godoku.Name, Namespace: godoku.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		dep, err := r.createDeployment(godoku)
		if err != nil {
			log.Error(err, "Failed to define new Deployment resource for Godoku")
			return ctrl.Result{}, err
		}
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		if err = r.Create(ctx, dep); err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}
	size := godoku.Spec.Replicas
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		if err = r.Update(ctx, found); err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			if err := r.Get(ctx, req.NamespacedName, godoku); err != nil {
				log.Error(err, "Failed to re-fetch godoku")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

// set environment
// reconciler does recreate deplyoment after deleted manually
// define separate functions for differente operations deployment.scale/create/update/delete service.create/update/delete route.create/update/delete
// create config map, service and route
// adjust naming and labels
func (r *GodokuReconciler) createDeployment(godoku *tuffv1alpha1.Godoku) (*appsv1.Deployment, error) {
	ls := map[string]string{
		"game": godoku.Spec.Name,
	}
	name := godoku.Spec.Name
	namespace := godoku.Spec.Namespace
	replicas := godoku.Spec.Replicas
	image := godoku.Spec.Image

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           image,
						Name:            godoku.Spec.Name,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Ports:           godoku.Spec.Ports,
					}},
				},
			},
		},
	}
	if err := ctrl.SetControllerReference(godoku, dep, r.Scheme); err != nil {
		return nil, err
	}
	return dep, nil
}

func (r *GodokuReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tuffv1alpha1.Godoku{}).
		Complete(r)
}
