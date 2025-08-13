package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/NdoleStudio/lemonsqueezy-go"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/ctxt"
	"github.com/getseabird/seabird/internal/pubsub"
	"github.com/getseabird/seabird/internal/ui/common"
	"github.com/getseabird/seabird/widget"
	"k8s.io/klog/v2"
)

type WelcomeWindow struct {
	*adw.ApplicationWindow
	*common.State
	ctx     context.Context
	content *adw.Bin
	nav     *adw.NavigationView
	toast   *adw.ToastOverlay
}

func NewWelcomeWindow(ctx context.Context, app *gtk.Application, state *common.State) *WelcomeWindow {
	window := adw.NewApplicationWindow(app)
	ctx = ctxt.With[*gtk.Window](ctx, &window.Window)
	w := WelcomeWindow{
		ctx:               ctx,
		ApplicationWindow: window,
		content:           adw.NewBin(),
		State:             state,
	}
	w.SetApplication(app)
	w.SetIconName("seabird")
	w.SetDefaultSize(600, 650)
	w.toast = adw.NewToastOverlay()
	w.toast.SetChild(w.content)
	w.SetContent(w.toast)
	w.content.SetChild(w.createContent(true))
	w.SetTitle(ApplicationName)

	go w.showUpdateNotification()

	var h glib.SignalHandle
	h = w.ConnectCloseRequest(func() bool {
		prefs := w.Preferences.Value()
		if err := prefs.Save(); err != nil {
			d := widget.ShowErrorDialog(ctx, "Could not save preferences", err)
			d.ConnectUnrealize(func() {
				w.Close()
			})
			w.HandlerDisconnect(h)
			return true
		}
		return false
	})

	return &w
}

func (w *WelcomeWindow) createContent(first bool) *adw.NavigationView {
	w.nav = adw.NewNavigationView()
	w.nav.ConnectPopped(func(page *adw.NavigationPage) {
		w.content.SetChild(w.createContent(false))
	})

	box := gtk.NewBox(gtk.OrientationVertical, 0)
	w.nav.Add(adw.NewNavigationPage(box, ApplicationName))

	header := gtk.NewHeaderBar()
	box.Append(header)

	page := adw.NewPreferencesPage()
	box.Append(page)

	if clusters := w.Preferences.Value().Clusters; len(clusters) > 0 {
		// if first && !style.Eq(style.Windows) && w.Preferences.Value().License == nil && rand.IntN(10) == 0 {
		// 	w.nav.Push(w.createPurchasePage())
		// }

		group := adw.NewPreferencesGroup()
		group.SetTitle("Connect to Cluster")
		page.Add(group)

		add := gtk.NewButton()
		add.AddCSSClass("flat")
		add.SetIconName("plus-symbolic")
		add.ConnectClicked(func() {
			pref := NewClusterPrefPage(w.ctx, w.State, pubsub.NewProperty(api.ClusterPreferences{}))
			w.nav.Push(pref.NavigationPage)
		})

		group.SetHeaderSuffix(add)

		for i, c := range w.Preferences.Value().Clusters {
			cluster := c
			row := adw.NewActionRow()
			row.SetTitle(cluster.Value().Name)
			row.SetActivatable(true)

			if kubeconfig := c.Value().Kubeconfig; kubeconfig != nil {
				label := gtk.NewLabel(kubeconfig.Path)
				label.AddCSSClass("dim-label")
				label.SetHAlign(gtk.AlignStart)
				row.AddSuffix(label)
			}

			spinner := widget.NewFallbackSpinner(gtk.NewImageFromIconName("go-next-symbolic"))
			row.AddSuffix(spinner)
			row.ConnectActivated(func() {
				if showClusterPrefsErrorDialog(w.ctx, cluster.Value()) {
					return
				}

				spinner.Start()
				go func() {
					state, err := w.NewClusterState(w.ctx, cluster)
					glib.IdleAdd(func() {
						spinner.Stop()
						if err != nil {
							widget.ShowErrorDialog(w.ctx, "Cluster connection failed", err)
							return
						}
						app := w.Application()
						w.Close()
						NewClusterWindow(w.ctx, app, state).Present()
					})
				}()
			})
			group.Add(row)
			if os.Getenv("SEABIRD_DEV") == "1" && i == 0 {
				defer row.Activate()
			}
		}
	} else {
		status := adw.NewStatusPage()
		status.SetIconName("seabird")
		status.SetTitle("No Clusters Found")
		status.SetDescription("Connect to a cluster to get started.")
		btn := gtk.NewButton()
		btn.ConnectClicked(func() {
			pref := NewClusterPrefPage(w.ctx, w.State, pubsub.NewProperty(api.ClusterPreferences{}))
			w.nav.Push(pref.NavigationPage)
		})
		btn.SetHAlign(gtk.AlignCenter)
		btn.SetLabel("New Cluster")
		btn.AddCSSClass("pill")
		btn.AddCSSClass("suggested-action")
		status.SetChild(btn)
		box.Append(status)
	}

	return w.nav
}

func (w *WelcomeWindow) createPurchasePage() *adw.NavigationPage {
	body := gtk.NewBox(gtk.OrientationVertical, 0)
	navPage := adw.NewNavigationPage(body, "Purchase Seabird")

	header := adw.NewHeaderBar()
	header.SetShowBackButton(false)
	body.Append(header)

	clamp := adw.NewClamp()
	clamp.SetMaximumSize(650)
	body.Append(clamp)

	status := adw.NewStatusPage()
	status.SetIconName("seabird")
	status.SetTitle("This Bird Needs Your Help")
	status.SetDescription("Seabird is free software with no limitations. To maintain free and open access, we need your support.")
	clamp.SetChild(status)

	content := gtk.NewBox(gtk.OrientationVertical, 24)
	status.SetChild(content)

	benefits := gtk.NewGrid()
	benefits.SetColumnSpacing(8)
	benefits.SetRowSpacing(8)
	content.Append(benefits)
	for i, benefit := range []string{"Get direct email support", "Influence our roadmap", "No vendor lock-in", "No enterprise-only features", "Auditable code under MPL 2.0 license", "Contribute to open-source ecosystem"} {
		box := gtk.NewBox(gtk.OrientationHorizontal, 4)
		icon := gtk.NewImageFromIconName("emblem-ok-symbolic")
		icon.AddCSSClass("success")
		box.Append(icon)
		box.Append(gtk.NewLabel(benefit))
		box.SetHExpand(true)
		benefits.Attach(box, i%2, i/2, 1, 1)
	}
	later := gtk.NewButton()
	later.ConnectClicked(func() {
		w.nav.Pop()
	})
	later.SetHAlign(gtk.AlignCenter)
	later.SetLabel("Remind Me Later")
	later.AddCSSClass("pill")
	purchase := gtk.NewButton()
	purchase.ConnectClicked(func() {
		gtk.ShowURI(&w.Window, "https://seabird.lemonsqueezy.com/checkout/buy/7cbd80a0-701b-46cc-b61f-c46cc339dca5", gdk.CURRENT_TIME)
	})
	purchase.SetHAlign(gtk.AlignCenter)
	purchase.SetLabel("Purchase Now")
	purchase.AddCSSClass("pill")
	purchase.AddCSSClass("suggested-action")
	actions := gtk.NewBox(gtk.OrientationHorizontal, 16)
	actions.SetHAlign(gtk.AlignCenter)
	content.Append(actions)
	actions.Append(later)
	actions.Append(purchase)

	// label = gtk.NewLabel("Did you know that nearly 60% of open-source maintainers have either quit or contemplated quitting their roles? By supporting this project financially, you can help ensure its long-term sustainability.")
	// label.SetWrap(true)
	// content.Append(label)

	group := adw.NewPreferencesGroup()
	group.SetMarginTop(16)
	content.Append(group)

	entry := adw.NewEntryRow()
	entry.SetTitle("License key")
	entry.SetShowApplyButton(true)
	entry.ConnectApply(func() {
		res, raw, err := lemonsqueezy.New().Licenses.Activate(w.ctx, strings.TrimSpace(entry.Text()), "Seabird")
		switch {
		case err != nil:
			klog.Infof("%v", err)
			err = errors.New(http.StatusText(raw.HTTPResponse.StatusCode))
			widget.ShowErrorDialog(w.ctx, "Could not activate license", err)
		case res.Activated:
			prefs := w.Preferences.Value()
			prefs.License = &api.License{
				ID:        res.Instance.ID,
				Key:       res.LicenseKey.Key,
				ExpiresAt: res.LicenseKey.ExpiresAt,
			}
			w.Preferences.Pub(prefs)
			w.toast.AddToast(adw.NewToast("License activated. Thank you!"))
			w.nav.Pop()
		default:
			widget.ShowErrorDialog(w.ctx, "Could not activate license", errors.New(res.Error))
		}
	})
	group.Add(entry)

	return navPage
}

func (w *WelcomeWindow) showUpdateNotification() {
	if strings.Contains(Version, "dev") {
		return
	}

	res, err := http.Get("https://api.github.com/repos/getseabird/seabird/releases")
	if err != nil {
		return
	}

	type Release struct {
		TagName     string    `json:"tag_name"`
		PublishedAt time.Time `json:"published_at"`
		Draft       bool      `json:"draft"`
		Prerelease  bool      `json:"prerelease"`
	}
	var releases []Release
	json.NewDecoder(res.Body).Decode(&releases)

	var release *Release
	for _, r := range releases {
		if r.Draft || r.Prerelease {
			continue
		}
		release = &r
		break
	}
	if release == nil {
		return
	}

	if strings.Contains(Version, strings.TrimPrefix(release.TagName, "v")) {
		return
	}

	// wait a bit for stores to propagate updates
	if time.Now().Add(24 * time.Hour).Before(release.PublishedAt) {
		return
	}

	glib.IdleAdd(func() {
		group := gio.NewSimpleActionGroup()
		action := gio.NewSimpleAction("releases", nil)
		action.ConnectActivate(func(idx *glib.Variant) {
			gtk.ShowURI(&w.Window, "https://github.com/getseabird/seabird/releases", gdk.CURRENT_TIME)
		})
		group.AddAction(action)
		w.InsertActionGroup("welcome", group)

		toast := adw.NewToast(fmt.Sprintf("Version %s is available.", strings.TrimPrefix(release.TagName, "v")))
		toast.SetActionName("welcome.releases")
		toast.SetButtonLabel("Update")
		toast.SetPriority(adw.ToastPriorityNormal)
		w.toast.AddToast(toast)
	})
}
