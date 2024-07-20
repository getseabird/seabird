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
				return gtk.NewImageFromIconName("box-symbolic")
			case "ConfigMap":
				return gtk.NewImageFromIconName("file-sliders-symbolic")
			case "Secret":
				return gtk.NewImageFromIconName("file-key-2-symbolic")
			case "Namespace":
				return gtk.NewImageFromIconName("orbit-symbolic")
			case "Service":
				return gtk.NewImageFromIconName("waypoints-symbolic")
			case "Node":
				return gtk.NewImageFromIconName("server-symbolic")
			case "PersistentVolume":
				return gtk.NewImageFromIconName("hard-drive-download-symbolic")
			case "PersistentVolumeClaim":
				return gtk.NewImageFromIconName("hard-drive-upload-symbolic")
			case "Event":
				return gtk.NewImageFromIconName("dialog-information-symbolic")
			case "Endpoints":
			case "LimitRange":
			case "PodTemplate":
			case "ResourceQuota":
			case "ReplicationController":
			case "ServiceAccount":
				return gtk.NewImageFromIconName("user-symbolic")
			}
		}
	case appsv1.GroupName:
		switch gvk.Kind {
		case "ReplicaSet":
			return gtk.NewImageFromIconName("layers-2-symbolic")
		case "Deployment":
			return gtk.NewImageFromIconName("layers-3-symbolic")
		case "StatefulSet":
			return gtk.NewImageFromIconName("database-symbolic")
		case "DaemonSet":
			return gtk.NewImageFromIconName("server-cog-symbolic")
		case "ControllerRevision":
		}
	case batchv1.GroupName:
		switch gvk.Kind {
		case "Job":
			return gtk.NewImageFromIconName("cloud-cog-symbolic")
		case "CronJob":
			return gtk.NewImageFromIconName("timer-reset-symbolic")
		}
	case networkingv1.GroupName:
		switch gvk.Kind {
		case "Ingress":
			return gtk.NewImageFromIconName("radio-tower-symbolic")
		case "IngressClass":
			return gtk.NewImageFromIconName("cast-symbolic")
		case "NetworkPolicy":
			return gtk.NewImageFromIconName("globe-lock-symbolic")
		}
	case eventsv1.GroupName:
		switch gvk.Kind {
		case "Event":
			return gtk.NewImageFromIconName("dialog-information-symbolic")
		}
	case apiextensionsv1.GroupName:
		switch gvk.Kind {
		case "CustomResourceDefinition":
			return gtk.NewImageFromIconName("toy-brick-symbolic")
		}
	case storagev1.GroupName:
		switch gvk.Kind {
		case "CSIDriver":
			return gtk.NewImageFromIconName("warehouse-symbolic")
		case "CSINode":
			return gtk.NewImageFromIconName("cylinder-symbolic")
		case "CSIStorageCapacity":
		case "StorageClass":
			return gtk.NewImageFromIconName("import-symbolic")
		case "VolumeAttachment":
		}
	case "helm.toolkit.fluxcd.io":
		switch gvk.Kind {
		case "HelmRelease":
			return gtk.NewImageFromIconName("package-open-symbolic")
		}
	case "source.toolkit.fluxcd.io":
		switch gvk.Kind {
		case "HelmChart":
			return gtk.NewImageFromIconName("map-symbolic")
		case "HelmRepository":
			return gtk.NewImageFromIconName("library-symbolic")
		case "GitRepository":
			return gtk.NewImageFromIconName("folder-git-symbolic")
		case "Bucket":
			return gtk.NewImageFromIconName("paint-bucket-symbolic")
		case "OCIRepository":
			return gtk.NewImageFromIconName("container-symbolic")
		}
	case "monitoring.coreos.com":
		switch gvk.Kind {
		case "AlertmanagerConfig":
		case "Alertmanager":
		case "PrometheusAgent":
		case "PodMonitor":
			return gtk.NewImageFromIconName("package-search-symbolic")
		case "Probe":
		case "PrometheusRole":
		case "Prometheus":
		case "ScrapingConfig":
		case "ServiceMonitor":
		case "ThanosRuler":
		}
	}

	return gtk.NewImageFromIconName("blocks-symbolic")
}
