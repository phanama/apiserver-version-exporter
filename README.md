# k8s apiserver-version-exporter

Queries k8s apiserver `/version` path and transforms it into `kube_build_info` metrics.

Rationale:
k8s 1.17 apiserver had the `kube_build_info` metrics removed: https://github.com/kubernetes/kubernetes/issues/89685

This project can be used as a temporary solution for those needing the metric.
