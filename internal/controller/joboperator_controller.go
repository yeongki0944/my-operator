/*
Copyright 2026.

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
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	batchv1 "github.com/yeongki/my-operator/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// JobOperatorReconciler reconciles a JobOperator object
type JobOperatorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=batch.my.domain,resources=joboperators,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.my.domain,resources=joboperators/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch.my.domain,resources=joboperators/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete

func (r *JobOperatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// [Metrics] 시작 시간 측정
	startTime := time.Now()

	// Fetch the JobOperator instance
	jobOp := &batchv1.JobOperator{}
	if err := r.Get(ctx, req.NamespacedName, jobOp); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		// [Metrics] 조회 실패 기록 추가
		ReconcileErrors.WithLabelValues(req.Name, req.Namespace, "fetch_failed").Inc()
		ReconcileTotal.WithLabelValues(req.Name, req.Namespace, "error").Inc()
		return ctrl.Result{}, err
	}

	// Create or update StatefulSet
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobOp.Name + "-sts",
			Namespace: jobOp.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: jobOp.Name,
			Replicas:    jobOp.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": jobOp.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": jobOp.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "worker",
							Image: jobOp.Spec.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: jobOp.Spec.Port,
								},
							},
						},
					},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(jobOp, sts, r.Scheme); err != nil {
		// [Metrics] OwnerRef 설정 실패 기록 추가
		ReconcileErrors.WithLabelValues(req.Name, req.Namespace, "owner_ref_failed").Inc()
		return ctrl.Result{}, err
	}

	if err := r.Create(ctx, sts); err != nil && !apierrors.IsAlreadyExists(err) {
		// [Metrics] 생성 실패 기록 추가
		ReconcileErrors.WithLabelValues(req.Name, req.Namespace, "create_sts_failed").Inc()
		ReconcileTotal.WithLabelValues(req.Name, req.Namespace, "error").Inc()
		// [Metrics] 실패 시에도 소요 시간 기록
		ReconcileDurationSeconds.WithLabelValues(req.Name, req.Namespace, "error").Observe(time.Since(startTime).Seconds())
		
		return ctrl.Result{}, err
	}

	// [Metrics] 성공 기록
	ReconcileTotal.WithLabelValues(req.Name, req.Namespace, "success").Inc()
	ReconcileDurationSeconds.WithLabelValues(req.Name, req.Namespace, "success").Observe(time.Since(startTime).Seconds())

	log.Info("Reconciliation successful", "duration", time.Since(startTime).String())

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JobOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.JobOperator{}).
		Named("joboperator").
		Complete(r)
}