package icon

import (
	_ "embed"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func Kind(gvk schema.GroupVersionKind) *gtk.Image {
	switch gvk.Group {
	case corev1.GroupName:
		{
			switch gvk.Kind {
			case "Pod":
				return gtk.NewImageFromIconName("application-x-executable-symbolic")
			case "ConfigMap":
				return gtk.NewImageFromIconName("rich-text-symbolic")
			case "Secret":
				return gtk.NewImageFromIconName("key-symbolic")
			case "Namespace":
				return gtk.NewImageFromIconName("globe-symbolic")
			case "Service":
				return gtk.NewImageFromIconName("network-proxy-server-symbolic")
			case "Node":
				return gtk.NewImageFromIconName("network-server-symbolic")
			case "PersistentVolume":
				return gtk.NewImageFromIconName("harddisk-symbolic")
			case "PersistentVolumeClaim":
				return gtk.NewImageFromIconName("harddisk-inverted-symbolic")
			case "Event":
				return gtk.NewImageFromIconName("bullhorn-symbolic")
			case "Endpoints":
			case "LimitRange":
			case "PodTemplate":
			case "ResourceQuota":
			case "ReplicationController":
			case "ServiceAccount":
				return gtk.NewImageFromIconName("people-symbolic")
			}
		}
	case appsv1.GroupName:
		switch gvk.Kind {
		case "ReplicaSet":
			return gtk.NewImageFromIconName("grid-symbolic")
		case "Deployment":
			return gtk.NewImageFromIconName("grid-large-symbolic")
		case "StatefulSet":
			return gtk.NewImageFromIconName("raid-symbolic")
		case "DaemonSet":
			return gtk.NewImageFromIconName("display-with-window-symbolic")
		case "ControllerRevision":
		}
	case batchv1.GroupName:
		switch gvk.Kind {
		case "Job":
			return gtk.NewImageFromIconName("meeting-symbolic")
		case "CronJob":
			return gtk.NewImageFromIconName("clock-alt-symbolic")
		}
	case networkingv1.GroupName:
		switch gvk.Kind {
		case "Ingress":
			return gtk.NewImageFromIconName("network-transmit-receive-symbolic")
		case "IngressClass":
			return gtk.NewImageFromIconName("network-no-route-symbolic")
		case "NetworkPolicy":
			return gtk.NewImageFromIconName("network-error-symbolic")
		}
	case eventsv1.GroupName:
		switch gvk.Kind {
		case "Event":
			return gtk.NewImageFromIconName("bullhorn-symbolic")
		}
	case apiextensionsv1.GroupName:
		switch gvk.Kind {
		case "CustomResourceDefinition":
			return gtk.NewImageFromIconName("puzzle-piece-symbolic")
		}
	case storagev1.GroupName:
		switch gvk.Kind {
		case "CSIDriver":
		case "CSINode":
		case "CSIStorageCapacity":
		case "StorageClass":
		case "VolumeAttachment":
		}
	case "helm.toolkit.fluxcd.io":
		switch gvk.Kind {
		case "HelmRelease":
			return gtk.NewImageFromIconName("package-x-generic-symbolic")
		}
	case "source.toolkit.fluxcd.io":
		switch gvk.Kind {
		case "HelmChart":
			return gtk.NewImageFromIconName("map-symbolic")
		case "HelmRepository":
			return gtk.NewImageFromIconName("library-symbolic")
		case "GitRepository":
			return gtk.NewImageFromIconName("git-symbolic")
		case "Bucket":
			return gtk.NewImageFromIconName("fill-tool-symbolic")
		case "OCIRepository":
			return gtk.NewImageFromIconName("image-symbolic")
		}
	case "monitoring.coreos.com":
		switch gvk.Kind {
		case "AlertmanagerConfig":
		case "Alertmanager":
		case "PrometheusAgent":
		case "PodMonitor":
		case "Probe":
		case "PrometheusRole":
		case "Prometheus":
		case "ScrapingConfig":
		case "ServiceMonitor":
		case "ThanosRuler":
		}
	}

	return gtk.NewImageFromIconName("puzzle-piece-symbolic")
}
