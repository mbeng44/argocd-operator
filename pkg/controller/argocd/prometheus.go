// Copyright 2019 ArgoCD Operator Developers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package argocd

import (
	"context"
	"fmt"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	argoproj "github.com/jmckind/argocd-operator/pkg/apis/argoproj/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// getPrometheusSize will return the size value for the Prometheus replica count.
func getPrometheusReplicas(cr *argoproj.ArgoCD) *int32 {
	replicas := ArgoCDDefaultPrometheusReplicas
	if cr.Spec.Prometheus.Size > replicas {
		replicas = cr.Spec.Prometheus.Size
	}
	return &replicas
}

// newPrometheus retuns a new Prometheus instance for the given ArgoCD.
func newPrometheus(cr *argoproj.ArgoCD) *monitoringv1.Prometheus {
	return &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labelsForCluster(cr),
		},
	}
}

// newServiceMonitor retuns a new ServiceMonitor instance.
func newServiceMonitor(cr *argoproj.ArgoCD) *monitoringv1.ServiceMonitor {
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labelsForCluster(cr),
		},
	}
}

// newServiceMonitorWithName retuns a new ServiceMonitor instance for the given ArgoCD using the given name.
func newServiceMonitorWithName(name string, cr *argoproj.ArgoCD) *monitoringv1.ServiceMonitor {
	svcmon := newServiceMonitor(cr)
	svcmon.ObjectMeta.Name = name

	lbls := svcmon.ObjectMeta.Labels
	lbls[ArgoCDKeyName] = name
	lbls[ArgoCDKeyRelease] = "prometheus-operator"
	svcmon.ObjectMeta.Labels = lbls

	return svcmon
}

// newServiceMonitorWithSuffix retuns a new ServiceMonitor instance for the given ArgoCD using the given suffix.
func newServiceMonitorWithSuffix(suffix string, cr *argoproj.ArgoCD) *monitoringv1.ServiceMonitor {
	return newServiceMonitorWithName(fmt.Sprintf("%s-%s", cr.Name, suffix), cr)
}

// reconcileMetricsServiceMonitor will ensure that the ServiceMonitor is present for the ArgoCD metrics Service.
func (r *ReconcileArgoCD) reconcileMetricsServiceMonitor(cr *argoproj.ArgoCD) error {
	sm := newServiceMonitorWithSuffix(ArgoCDKeyMetrics, cr)
	if r.isObjectFound(cr.Namespace, sm.Name, sm) {
		return nil // ServiceMonitor found, do nothing
	}

	sm.Spec.Selector = metav1.LabelSelector{
		MatchLabels: map[string]string{
			ArgoCDKeyName: nameWithSuffix(ArgoCDKeyMetrics, cr),
		},
	}
	sm.Spec.Endpoints = []monitoringv1.Endpoint{
		{
			Port: ArgoCDKeyMetrics,
		},
	}

	if err := controllerutil.SetControllerReference(cr, sm, r.scheme); err != nil {
		return err
	}
	return r.client.Create(context.TODO(), sm)
}

// reconcilePrometheus will ensure that Prometheus is present for ArgoCD metrics.
func (r *ReconcileArgoCD) reconcilePrometheus(cr *argoproj.ArgoCD) error {
	prometheus := newPrometheus(cr)
	if r.isObjectFound(cr.Namespace, prometheus.Name, prometheus) {
		return nil // Prometheus found, do nothing
	}

	prometheus.Spec.Replicas = getPrometheusReplicas(cr)
	prometheus.Spec.ServiceAccountName = "prometheus-k8s"
	prometheus.Spec.ServiceMonitorSelector = &metav1.LabelSelector{}

	if err := controllerutil.SetControllerReference(cr, prometheus, r.scheme); err != nil {
		return err
	}
	return r.client.Create(context.TODO(), prometheus)
}

// reconcileRepoServerServiceMonitor will ensure that the ServiceMonitor is present for the Repo Server metrics Service.
func (r *ReconcileArgoCD) reconcileRepoServerServiceMonitor(cr *argoproj.ArgoCD) error {
	sm := newServiceMonitorWithSuffix("repo-server-metrics", cr)
	if r.isObjectFound(cr.Namespace, sm.Name, sm) {
		return nil // ServiceMonitor found, do nothing
	}

	sm.Spec.Selector = metav1.LabelSelector{
		MatchLabels: map[string]string{
			ArgoCDKeyName: nameWithSuffix("repo-server", cr),
		},
	}
	sm.Spec.Endpoints = []monitoringv1.Endpoint{
		{
			Port: ArgoCDKeyMetrics,
		},
	}

	if err := controllerutil.SetControllerReference(cr, sm, r.scheme); err != nil {
		return err
	}
	return r.client.Create(context.TODO(), sm)
}

// reconcileServerMetricsServiceMonitor will ensure that the ServiceMonitor is present for the ArgoCD Server metrics Service.
func (r *ReconcileArgoCD) reconcileServerMetricsServiceMonitor(cr *argoproj.ArgoCD) error {
	sm := newServiceMonitorWithSuffix("server-metrics", cr)
	if r.isObjectFound(cr.Namespace, sm.Name, sm) {
		return nil // ServiceMonitor found, do nothing
	}

	sm.Spec.Selector = metav1.LabelSelector{
		MatchLabels: map[string]string{
			ArgoCDKeyName: nameWithSuffix("server-metrics", cr),
		},
	}
	sm.Spec.Endpoints = []monitoringv1.Endpoint{
		{
			Port: ArgoCDKeyMetrics,
		},
	}

	if err := controllerutil.SetControllerReference(cr, sm, r.scheme); err != nil {
		return err
	}
	return r.client.Create(context.TODO(), sm)
}
