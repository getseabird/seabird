package widget

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	corev1 "k8s.io/api/core/v1"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var html = `
<!doctype html>
  <html>
    <head>
			<script src="https://cdn.jsdelivr.net/npm/xterm@5.3.0/lib/xterm.min.js"></script>
			<script src="https://cdn.jsdelivr.net/npm/xterm-addon-attach@0.9.0/lib/xterm-addon-attach.min.js"></script>
			<link href="https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.min.css" rel="stylesheet">
    </head>
    <body>
      <div id="terminal"></div>
      <script>
        var term = new Terminal();
        term.open(document.getElementById('terminal'));
        term.write('Hello from \x1B[1;3;31mxterm.js\x1B[0m $ ')
      </script>
    </body>
  </html>
`

type TerminalPage struct {
	*adw.NavigationPage
}

func NewTerminalPage(parent *gtk.Window, cluster *api.Cluster, pod *corev1.Pod, container string) *TerminalPage {
	return nil
}

func server() (int, error) {
	http.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			log.Printf(err.Error())
			return
		}
		defer c.CloseNow()

		ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
		defer cancel()

		var v interface{}
		err = wsjson.Read(ctx, c, &v)
		if err != nil {
			log.Printf(err.Error())
			return
		}

		log.Printf("received: %v", v)

		c.Close(websocket.StatusNormalClosure, "")
	}))

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}

	return listener.Addr().(*net.TCPAddr).Port, http.Serve(listener, nil)
}
