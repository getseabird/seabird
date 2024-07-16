package extension

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/util"
	"github.com/getseabird/seabird/widget"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/reference"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	Extensions = append(Extensions, func(cluster *api.Cluster) Extension {
		return &Batch{Cluster: cluster}
	})
}

type Batch struct {
	*api.Cluster
}

func (e *Batch) CreateColumns(ctx context.Context, resource *metav1.APIResource, columns []api.Column) []api.Column {
	switch util.GVRForResource(resource).String() {
	case batchv1.SchemeGroupVersion.WithResource("jobs").String():
		columns = append(columns,
			api.Column{
				Name:     "Status",
				Priority: 70,
				Bind: func(listitem *gtk.ColumnViewCell, object client.Object) {
					listitem.SetChild(widget.ObjectStatus(object).Icon())
				},
				Compare: widget.CompareObjectStatus,
			},
			api.Column{
				Name:     "Completions",
				Priority: 70,
				Bind: func(listitem *gtk.ColumnViewCell, object client.Object) {
					job := object.(*batchv1.Job)
					label := gtk.NewLabel(fmt.Sprintf("%d/%d", job.Status.Succeeded, ptr.Deref(job.Spec.Completions, 1)))
					label.SetHAlign(gtk.AlignStart)
					listitem.SetChild(label)
				},
			},
		)
	case batchv1.SchemeGroupVersion.WithResource("cronjobs").String():
		columns = append(columns,
			api.Column{
				Name:     "Last schedule",
				Priority: 70,
				Bind: func(listitem *gtk.ColumnViewCell, object client.Object) {
					cron := object.(*batchv1.CronJob)
					if cron.Status.LastScheduleTime != nil {
						duration := time.Since(cron.Status.LastScheduleTime.Time)
						label := gtk.NewLabel(util.HumanizeApproximateDuration(duration))
						label.SetHAlign(gtk.AlignStart)
						listitem.SetChild(label)
					}
				},
			},
		)
	}
	return columns
}

func (e *Batch) CreateObjectProperties(ctx context.Context, resource *metav1.APIResource, object client.Object, props []api.Property) []api.Property {
	switch object := object.(type) {
	case *batchv1.Job:
		var images []string
		for _, c := range object.Spec.Template.Spec.Containers {
			images = append(images, c.Image)
		}

		props = append(props, &api.GroupProperty{Name: "Job", Children: []api.Property{
			&api.TextProperty{
				Name:  "Images",
				Value: strings.Join(images, ", "),
			},
			&api.TextProperty{
				Name:  "Completions",
				Value: fmt.Sprintf("%d", ptr.Deref(object.Spec.Completions, 1)),
			},
			&api.TextProperty{
				Name:  "Parallelism",
				Value: fmt.Sprintf("%d", ptr.Deref(object.Spec.Parallelism, 1)),
			},
			&api.TextProperty{
				Name:  "Backoff limit",
				Value: fmt.Sprintf("%d", ptr.Deref(object.Spec.BackoffLimit, 6)),
			},
		}})

		prop := &api.GroupProperty{Name: "Pods"}
		var pods corev1.PodList
		e.List(ctx, &pods, client.InNamespace(object.Namespace), client.MatchingLabels(object.Spec.Selector.MatchLabels))
		for i, pod := range pods.Items {
			ref, _ := reference.GetReference(e.Scheme, &pod)
			prop.Children = append(prop.Children, &api.TextProperty{
				ID:        fmt.Sprintf("pods.%d", i),
				Reference: ref,
				Value:     pod.Name,
				Widget: func(w gtk.Widgetter, nv *adw.NavigationView) {
					switch row := w.(type) {
					case *adw.ActionRow:
						row.AddPrefix(widget.ObjectStatus(&pod).Icon())
					}
				},
			})
		}
		props = append(props, prop)
	}

	return props
}
