package extension

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/widget"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type PortForwarder struct {
	*api.Cluster
	forwarders map[types.NamespacedName]*portforward.PortForwarder
}

func (p *PortForwarder) New(ctx context.Context, name types.NamespacedName, ports []string) error {
	readyChan := make(chan struct{}, 1)
	errChan := make(chan error, 1)

	url := p.Clientset.CoreV1().RESTClient().Post().Resource("pods").Namespace(name.Namespace).Name(name.Name).SubResource("portforward").URL()
	transport, upgrader, err := spdy.RoundTripperFor(p.Config)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, url)
	tunnelingDialer, err := portforward.NewSPDYOverWebsocketDialer(url, p.Config)
	if err != nil {
		return err
	}
	dialer = portforward.NewFallbackDialer(tunnelingDialer, dialer, httpstream.IsUpgradeFailure)

	forwarder, err := portforward.NewOnAddresses(dialer, []string{"localhost"}, ports, ctx.Done(), readyChan, nil, os.Stderr)
	if err != nil {
		return err
	}
	p.forwarders[name] = forwarder

	go func() {
		if err := forwarder.ForwardPorts(); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-readyChan:
		return nil
	case err := <-errChan:
		return err
	case <-time.After(5 * time.Second):
		return errors.New("timeout")
	}
}

func (p *PortForwarder) GetPorts(name types.NamespacedName) ([]portforward.ForwardedPort, error) {
	if forwarder := p.forwarders[name]; forwarder != nil {
		return forwarder.GetPorts()
	} else {
		return nil, errors.New("not found")
	}
}

func (p *PortForwarder) Close(name types.NamespacedName) error {
	if forwarder := p.forwarders[name]; forwarder != nil {
		forwarder.Close()
		delete(p.forwarders, name)
		return nil
	} else {
		return errors.New("not found")
	}
}

func (p *PortForwarder) UpdateButton(ctx context.Context, btn *gtk.Button, name types.NamespacedName, ports []string) {
	var handle glib.SignalHandle
	if fwd, err := p.GetPorts(name); err != nil {
		btn.SetIconName("vertical-arrows-long-symbolic")
		btn.SetTooltipText("Forward port to localhost")
		btn.AddCSSClass("flat")
		handle = btn.ConnectClicked(func() {
			if err := p.New(ctx, name, ports); err != nil {
				widget.ShowErrorDialog(ctx, "Port forward error", err)
			} else {
				p.UpdateButton(ctx, btn, name, ports)
			}
			btn.HandlerDisconnect(handle)
		})
	} else {
		box := gtk.NewBox(gtk.OrientationHorizontal, 2)
		icon := gtk.NewImageFromIconName("cross-small-symbolic")
		icon.AddCSSClass("error")
		box.Append(icon)
		box.Append(gtk.NewLabel(fmt.Sprintf("%d", fwd[0].Local)))
		btn.SetChild(box)
		btn.RemoveCSSClass("flat")
		btn.SetTooltipText("Close forwarding port")
		handle = btn.ConnectClicked(func() {
			p.Close(name)
			p.UpdateButton(ctx, btn, name, ports)
			btn.HandlerDisconnect(handle)
		})
	}

}
