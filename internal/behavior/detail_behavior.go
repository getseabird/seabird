package behavior

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/util"
	"github.com/imkira/go-observer/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DetailBehavior struct {
	*ClusterBehavior

	SelectedObject observer.Property[client.Object]
	Yaml           observer.Property[string]
	Properties     observer.Property[[]api.Property]
}

func (b *ClusterBehavior) NewRootDetailBehavior() *DetailBehavior {
	db := b.NewDetailBehavior()
	b.RootDetailBehavior = db
	return db
}

func (b *ClusterBehavior) NewDetailBehavior() *DetailBehavior {
	d := DetailBehavior{
		ClusterBehavior: b,
		SelectedObject:  observer.NewProperty[client.Object](nil),
		Yaml:            observer.NewProperty[string](""),
		Properties:      observer.NewProperty[[]api.Property](nil),
	}

	onChange(d.SelectedObject, d.onObjectChange)

	return &d
}

func (b *DetailBehavior) onObjectChange(object client.Object) {
	if object == nil {
		b.Properties.Update([]api.Property{})
		b.Yaml.Update("")
		return
	}

	codec := unstructured.NewJSONFallbackEncoder(serializer.NewCodecFactory(b.Scheme).LegacyCodec(b.Scheme.PreferredVersionAllGroups()...))
	objWithoutManagedFields := object.DeepCopyObject().(client.Object)
	objWithoutManagedFields.SetManagedFields(nil)
	encoded, err := runtime.Encode(codec, objWithoutManagedFields)
	if err != nil {
		b.Yaml.Update(fmt.Sprintf("error: %v", err))
	} else {
		yaml, err := util.JsonToYaml(encoded)
		if err != nil {
			b.Yaml.Update(fmt.Sprintf("error: %v", err))
		} else {
			b.Yaml.Update(string(yaml))
		}
	}

	var properties []api.Property

	var labels []api.Property
	for key, value := range object.GetLabels() {
		labels = append(labels, &api.TextProperty{Name: key, Value: value})
	}
	var annotations []api.Property
	for key, value := range object.GetAnnotations() {
		annotations = append(annotations, &api.TextProperty{Name: key, Value: value})
	}
	var owners []api.Property
	for _, ref := range object.GetOwnerReferences() {
		owners = append(owners, &api.TextProperty{Name: fmt.Sprintf("%s %s", ref.APIVersion, ref.Kind), Value: ref.Name})
	}

	properties = append(properties,
		&api.GroupProperty{
			Name: "Metadata",
			Children: []api.Property{
				&api.TextProperty{
					Name:  "Name",
					Value: object.GetName(),
				},
				&api.TextProperty{
					Name:  "Namespace",
					Value: object.GetNamespace(),
				},
				&api.GroupProperty{
					Name:     "Labels",
					Children: labels,
				},
				&api.GroupProperty{
					Name:     "Annotations",
					Children: annotations,
				},
				&api.GroupProperty{
					Name:     "Owners",
					Children: owners,
				},
			},
		})

	for _, ext := range b.extensions {
		properties = ext.CreateObjectProperties(object, properties)
	}

	events := &api.GroupProperty{Name: "Events"}
	for _, ev := range b.Events.For(object) {
		eventTime := ev.EventTime.Time
		if eventTime.IsZero() {
			eventTime = ev.CreationTimestamp.Time
		}
		events.Children = append(events.Children, &api.TextProperty{
			Name:  eventTime.Format(time.RFC3339),
			Value: ev.Note,
		})
	}
	if len(events.Children) > 0 {
		properties = append(properties, events)
	}

	b.Properties.Update(properties)
}

func (b *DetailBehavior) PodLogs(pod *corev1.Pod, container string) ([]byte, error) {
	req := b.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{Container: container})
	r, err := req.Stream(context.TODO())
	if err != nil {
		return nil, err
	}
	return io.ReadAll(r)
}

func (b *DetailBehavior) PodExec(ctx context.Context, pod *corev1.Pod, container string, command []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, sizeQueue remotecommand.TerminalSizeQueue) error {
	req := b.CoreV1().RESTClient().Post().Resource("pods").Name(pod.Name).Namespace(pod.Namespace).SubResource("exec")
	option := &corev1.PodExecOptions{
		Container: container,
		Command:   command,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)

	spdy, err := remotecommand.NewSPDYExecutor(b.Config, "POST", req.URL())
	if err != nil {
		return err
	}
	ws, err := remotecommand.NewWebSocketExecutor(b.Config, "GET", req.URL().String())
	if err != nil {
		return err
	}
	exec, err := remotecommand.NewFallbackExecutor(ws, spdy, httpstream.IsUpgradeFailure)
	if err != nil {
		return err
	}

	return exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             stdin,
		Stdout:            stdout,
		Stderr:            stderr,
		Tty:               true,
		TerminalSizeQueue: sizeQueue,
	})
}

func (b *DetailBehavior) UpdateObject(obj *unstructured.Unstructured) error {
	m, err := b.RESTMapper.RESTMapping(obj.GetObjectKind().GroupVersionKind().GroupKind(), obj.GetObjectKind().GroupVersionKind().Version)
	if err != nil {
		return err
	}
	var iface dynamic.ResourceInterface = b.DynamicClient.Resource(m.Resource)
	if len(obj.GetNamespace()) > 0 {
		iface = iface.(dynamic.NamespaceableResourceInterface).Namespace(obj.GetNamespace())
	}
	_, err = iface.Update(context.TODO(), obj, metav1.UpdateOptions{})
	return err
}

func (b *DetailBehavior) DeleteObject(obj client.Object) error {
	return b.Delete(context.TODO(), obj)
}
