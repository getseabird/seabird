package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
	"github.com/getseabird/seabird/api"
	"github.com/getseabird/seabird/internal/style"
	"github.com/getseabird/seabird/internal/ui/common"
	"github.com/getseabird/seabird/internal/ui/editor"
	"github.com/getseabird/seabird/internal/util"
	"github.com/getseabird/seabird/widget"
	"github.com/imkira/go-observer/v2"
	"github.com/zmwangx/debounce"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/reference"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Navigation struct {
	*adw.ToolbarView
	*common.ClusterState
	ctx                   context.Context
	resourceList          *gtk.ListBox
	pinList               *gtk.ListBox
	pinRows               []*gtk.ListBoxRow
	pinViews              []*adw.NavigationView
	favourites            []*gtk.ListBoxRow
	resources             []*gtk.ListBoxRow
	viewStack             *gtk.Stack
	editor                *editor.EditorWindow
	resourcesToggle       *gtk.ToggleButton
	pinsToggle            *gtk.ToggleButton
	search                *gtk.SearchEntry
	cancelFuncs           map[string]context.CancelFunc
	informerRegistrations map[informers.GenericInformer]cache.ResourceEventHandlerRegistration
}

func NewNavigation(ctx context.Context, state *common.ClusterState, viewStack *gtk.Stack, editor *editor.EditorWindow) *Navigation {
	n := &Navigation{
		ToolbarView:           adw.NewToolbarView(),
		ctx:                   ctx,
		ClusterState:          state,
		viewStack:             viewStack,
		editor:                editor,
		cancelFuncs:           map[string]context.CancelFunc{},
		informerRegistrations: map[informers.GenericInformer]cache.ResourceEventHandlerRegistration{},
	}
	n.SetVExpand(true)
	n.AddCSSClass("navigation-sidebar")

	header := adw.NewHeaderBar()
	title := gtk.NewLabel(n.ClusterPreferences.Value().Name)
	title.SetEllipsize(pango.EllipsizeEnd)
	title.AddCSSClass("heading")
	header.SetTitleWidget(title)
	header.SetShowEndTitleButtons(false)
	header.SetShowStartTitleButtons(style.Eq(style.Darwin))

	button := gtk.NewMenuButton()
	button.SetIconName("open-menu-symbolic")

	windowSection := gio.NewMenu()
	windowSection.Append("New Window", "win.newWindow")
	windowSection.Append("Disconnect", "win.disconnect")

	prefSection := gio.NewMenu()
	prefSection.Append("Preferences", "win.prefs")
	// prefSection.Append("Keyboard Shortcuts", "win.shortcuts")
	prefSection.Append("About", "win.about")

	m := gio.NewMenu()
	m.AppendSection("", windowSection)
	m.AppendSection("", prefSection)

	popover := gtk.NewPopoverMenuFromModel(m)
	button.SetPopover(popover)

	header.PackEnd(button)
	n.AddTopBar(header)

	content := gtk.NewBox(gtk.OrientationVertical, 4)
	n.SetContent(content)

	toggleBox := gtk.NewBox(gtk.OrientationHorizontal, 4)
	toggleBox.SetMarginStart(8)
	toggleBox.SetMarginEnd(8)
	content.Append(toggleBox)
	n.resourcesToggle = gtk.NewToggleButton()
	n.resourcesToggle.AddCSSClass("flat")
	n.resourcesToggle.SetIconName("view-list-symbolic")
	n.resourcesToggle.SetHExpand(true)
	n.resourcesToggle.SetActive(true)
	toggleBox.Append(n.resourcesToggle)
	n.pinsToggle = gtk.NewToggleButton()
	n.pinsToggle.AddCSSClass("flat")
	n.pinsToggle.SetIconName("view-pin-symbolic")
	n.pinsToggle.SetHExpand(true)
	toggleBox.Append(n.pinsToggle)

	navStack := gtk.NewStack()
	content.Append(navStack)

	resbin := adw.NewBin()
	resw := gtk.NewScrolledWindow()
	resw.SetChild(resbin)
	resw.SetVExpand(true)

	resBox := gtk.NewBox(gtk.OrientationVertical, 4)
	resBox.Append(resw)
	n.search = gtk.NewSearchEntry()
	n.search.SetVAlign(gtk.AlignEnd)
	n.search.SetObjectProperty("placeholder-text", "Filter...")
	n.search.ConnectSearchChanged(func() {
		resbin.SetChild(n.createResourceList(n.ClusterPreferences.Value()))
	})
	resBox.Append(n.search)

	navStack.AddChild(resBox)
	navStack.SetVisibleChild(resBox)

	n.pinList = gtk.NewListBox()
	n.pinList.AddCSSClass("navigation-sidebar")
	n.pinList.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		if row == nil {
			return
		}

		var ref corev1.ObjectReference
		if err := json.Unmarshal([]byte(row.Name()), &ref); err != nil {
			panic(err)
		}
		pages := n.viewStack.Pages()
		for i := 0; i < int(pages.NItems()); i++ {
			page := pages.Item(uint(i)).Cast().(*gtk.StackPage)
			if page.Name() == fmt.Sprintf("%s/%s", ref.Namespace, ref.Name) {
				n.viewStack.SetVisibleChild(page.Child())
				return
			}
		}
	})

	pinw := gtk.NewScrolledWindow()
	pinw.SetChild(n.pinList)
	pinw.SetVExpand(true)
	navStack.AddChild(pinw)

	n.resourcesToggle.ConnectClicked(func() {
		n.resourcesToggle.SetActive(true)
	})
	n.resourcesToggle.ConnectToggled(func() {
		if n.resourcesToggle.Active() {
			n.pinsToggle.SetActive(false)
			navStack.SetVisibleChild(resBox)
			if row := n.resourceList.SelectedRow(); row != nil {
				row.Activate()
			}
		}
	})
	n.pinsToggle.ConnectClicked(func() {
		n.pinsToggle.SetActive(true)
	})
	n.pinsToggle.ConnectToggled(func() {
		if n.pinsToggle.Active() {
			n.resourcesToggle.SetActive(false)
			navStack.SetVisibleChild(pinw)
			if row := n.pinList.SelectedRow(); row != nil {
				row.Activate()
			} else if len(n.pinRows) > 0 {
				n.pinList.SelectRow(n.pinRows[0])
				n.pinRows[0].Activate()
			}
		}
	})

	common.OnChange(ctx, n.ClusterPreferences, func(prefs api.ClusterPreferences) {
		resbin.SetChild(n.createResourceList(prefs))
		n.updatePins(prefs.Navigation.Pins)
	})

	resbin.SetChild(n.createResourceList(n.ClusterPreferences.Value()))
	if len(n.favourites) > 0 {
		n.resourceList.SelectRow(n.favourites[0])
	}

	return n
}

func (n *Navigation) createResourceList(prefs api.ClusterPreferences) *gtk.ListBox {
	for inf, reg := range n.informerRegistrations {
		inf.Informer().RemoveEventHandler(reg)
		delete(n.informerRegistrations, inf)
	}

	n.resourceList = gtk.NewListBox()
	n.resourceList.AddCSSClass("navigation-sidebar")

	// TODO actions should be able to use "u" for uint but I can't get it to work
	actionGroup := gio.NewSimpleActionGroup()
	pin := gio.NewSimpleAction("pin", glib.NewVariantType("s"))
	pin.ConnectActivate(func(idx *glib.Variant) {
		id, _ := strconv.Atoi(idx.String())
		prefs := n.ClusterPreferences.Value()
		prefs.Navigation.Favourites = append(prefs.Navigation.Favourites, util.GVRForResource(&n.Resources[id]))
		n.ClusterPreferences.Update(prefs)
	})
	actionGroup.AddAction(pin)
	unpin := gio.NewSimpleAction("unpin", glib.NewVariantType("s"))
	unpin.ConnectActivate(func(idx *glib.Variant) {
		id, _ := strconv.Atoi(idx.String())
		prefs := n.ClusterPreferences.Value()
		for i, f := range prefs.Navigation.Favourites {
			if util.GVREquals(f, util.GVRForResource(&n.Resources[id])) {
				prefs.Navigation.Favourites = slices.Delete(prefs.Navigation.Favourites, i, i+1)
				n.ClusterPreferences.Update(prefs)
				break
			}
		}
	})
	actionGroup.AddAction(unpin)
	n.resourceList.InsertActionGroup("navigation", actionGroup)

	n.resourceList.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		if row == nil {
			return
		}

		var gvr schema.GroupVersionResource
		if err := json.Unmarshal([]byte(row.Name()), &gvr); err != nil {
			return
		}
		pages := n.viewStack.Pages()
		for i := 0; i < int(pages.NItems()); i++ {
			page := pages.Item(uint(i)).Cast().(*gtk.StackPage)
			if page.Name() == "list" {
				n.viewStack.SetVisibleChild(page.Child())
				break
			}
		}
	})

	n.resourceList.ConnectRowSelected(func(row *gtk.ListBoxRow) {
		if row == nil {
			return
		}

		var gvr schema.GroupVersionResource
		if err := json.Unmarshal([]byte(row.Name()), &gvr); err != nil {
			return
		}
		for _, res := range n.Resources {
			if util.GVREquals(util.GVRForResource(&res), gvr) && !util.ResourceEquals(n.SelectedResource.Value(), &res) {
				n.SelectedResource.Update(&res)
				break
			}
		}
	})

	n.favourites = nil
	n.resources = nil

	for i, resource := range n.Resources {
		if len(n.search.Text()) > 0 {
			if !strings.Contains(resource.Name, n.search.Text()) &&
				strutil.Similarity(resource.Name, n.search.Text(), metrics.NewLevenshtein()) < 0.7 &&
				!strings.Contains(resource.Group, n.search.Text()) &&
				strutil.Similarity(resource.Group, n.search.Text(), metrics.NewLevenshtein()) < 0.7 {
				continue
			}
		}

		var fav bool
		for _, f := range prefs.Navigation.Favourites {
			if util.GVREquals(f, util.GVRForResource(&resource)) {
				fav = true
			}
		}
		row := n.createResourceRow(&resource, i, fav)
		if fav {
			n.favourites = append(n.favourites, row)
		} else {
			n.resources = append(n.resources, row)
		}

		if selected := n.SelectedResource.Value(); selected != nil && util.ResourceEquals(selected, &resource) {
			n.resourceList.SelectRow(row)
		}
	}

	if len(n.favourites) > 0 {
		header := n.createHeaderRow("Favourites")
		n.resourceList.Append(header)
		for _, row := range n.favourites {
			n.resourceList.Append(row)
		}
	}

	if len(n.resources) > 0 {
		header := n.createHeaderRow("Resources")
		n.resourceList.Append(header)
		for _, row := range n.resources {
			n.resourceList.Append(row)
		}
	}

	return n.resourceList
}

func (n *Navigation) updatePins(pins []corev1.ObjectReference) *gtk.ListBox {
rows:
	for _, row := range n.pinRows {
		var ref corev1.ObjectReference
		if err := json.Unmarshal([]byte(row.Name()), &ref); err != nil {
			panic(err)
		}

		for _, pin := range pins {
			if pin.Name == ref.Name && pin.Namespace == ref.Namespace {
				continue rows
			}
		}

		defer n.removePin(ref)
	}

outer:
	for i, pin := range pins {
		for _, row := range n.pinRows {
			var ref corev1.ObjectReference
			if err := json.Unmarshal([]byte(row.Name()), &ref); err != nil {
				panic(err)
			}
			if pin.Name == ref.Name && pin.Namespace == ref.Namespace {
				continue outer
			}
		}
		object, err := n.GetReference(n.ctx, pin)
		if err != nil {
			if errors.IsNotFound(err) {
				prefs := n.ClusterPreferences.Value()
				prefs.Navigation.Pins = slices.Delete(pins, i, i+1)
				n.ClusterPreferences.Update(prefs)
			} else {
				klog.Infof("updatePins: %s %v", err, pin)
			}
			continue
		}
		n.createPin(object)
	}

	return n.pinList
}

func createObjectRow(ref corev1.ObjectReference) *gtk.ListBoxRow {
	row := gtk.NewListBoxRow()
	json, err := json.Marshal(ref)
	if err != nil {
		panic(err)
	}
	row.SetName(string(json))
	row.AddCSSClass(fmt.Sprintf("%s/%s", ref.Namespace, ref.Name))
	box := gtk.NewBox(gtk.OrientationHorizontal, 8)
	box.SetMarginTop(4)
	box.SetMarginBottom(4)
	box.Append(resourceImage(ref.GroupVersionKind()))
	row.SetChild(box)
	label := gtk.NewLabel(ref.Name)
	label.SetHAlign(gtk.AlignStart)
	label.SetEllipsize(pango.EllipsizeEnd)
	box.Append(label)

	return row
}

func (n *Navigation) createResourceRow(resource *metav1.APIResource, idx int, fav bool) *gtk.ListBoxRow {
	gvr := util.GVRForResource(resource)

	row := gtk.NewListBoxRow()
	json, err := json.Marshal(gvr)
	if err != nil {
		panic(err)
	}
	row.SetName(string(json))
	box := gtk.NewBox(gtk.OrientationHorizontal, 8)
	box.SetMarginTop(4)
	box.SetMarginBottom(4)
	box.Append(resourceImage(util.GVKForResource(resource)))
	vbox := gtk.NewBox(gtk.OrientationVertical, 2)
	vbox.SetVAlign(gtk.AlignCenter)
	box.Append(vbox)
	label := gtk.NewLabel(resource.Kind)
	label.SetHAlign(gtk.AlignStart)
	label.SetEllipsize(pango.EllipsizeEnd)
	vbox.Append(label)
	label = gtk.NewLabel(resource.Group)
	if resource.Group == "" {
		label.SetText("k8s.io")
	}
	label.SetHAlign(gtk.AlignStart)
	label.AddCSSClass("caption")
	label.AddCSSClass("dim-label")
	label.SetEllipsize(pango.EllipsizeEnd)
	vbox.Append(label)
	row.SetChild(box)

	statusBox := gtk.NewBox(gtk.OrientationHorizontal, 4)
	statusBox.SetHExpand(true)
	statusBox.SetHAlign(gtk.AlignEnd)
	statusBox.SetVAlign(gtk.AlignCenter)
	box.Append(statusBox)

	errorBox := gtk.NewBox(gtk.OrientationHorizontal, 4)
	errorBox.SetVisible(false)
	statusBox.Append(errorBox)
	errorBox.AddCSSClass("warning")
	errorBox.AddCSSClass("pill")
	// icon := gtk.NewImageFromIconName("dialog-warning-symbolic")
	// errorBox.Append(icon)
	errorLabel := gtk.NewLabel("")
	errorBox.Append(errorLabel)

	if fav && n.Scheme.IsGroupRegistered(resource.Group) && slices.Contains(resource.Verbs, "watch") {
		go func() {
			informer := n.Cluster.GetInformer(util.GVRForResource(resource))
			h := bindStatusCount(n.ctx, informer, func(m map[widget.StatusType]int) {
				glib.IdleAdd(func() {
					errors := m[widget.StatusError] + m[widget.StatusWarning]
					errorBox.SetVisible(errors > 0)
					errorLabel.SetMarkup(fmt.Sprintf("<b>%d</b>", errors))
				})
			})
			if h != nil {
				glib.IdleAdd(func() {
					n.informerRegistrations[informer] = h
				})
			}
		}()
	}

	gesture := gtk.NewGestureClick()
	gesture.SetButton(gdk.BUTTON_SECONDARY)
	gesture.ConnectPressed(func(nPress int, x, y float64) {
		menu := gio.NewMenu()
		if fav {
			menu.Append("Move to Resources", fmt.Sprintf("navigation.unpin('%d')", idx))
		} else {
			menu.Append("Move to Favourites", fmt.Sprintf("navigation.pin('%d')", idx))
		}
		popover := gtk.NewPopoverMenuFromModel(menu)
		popover.SetHasArrow(false)
		row.FirstChild().(*gtk.Box).Append(popover)
		popover.SetVisible(true)
	})
	row.AddController(gesture)

	return row
}

func (n *Navigation) createHeaderRow(label string) *gtk.ListBoxRow {
	box := gtk.NewBox(gtk.OrientationHorizontal, 4)
	box.SetHAlign(gtk.AlignFill)
	box.AddCSSClass("dim-label")
	lbl := gtk.NewLabel(label)
	box.Append(lbl)
	row := gtk.NewListBoxRow()
	row.SetChild(box)
	row.SetSelectable(false)
	return row
}

func (n *Navigation) createPin(object client.Object) *gtk.ListBoxRow {
	ref, err := reference.GetReference(n.Scheme, object)
	if err != nil {
		log.Print("createPin: %s", err)
		return nil
	}

	row := createObjectRow(*ref)
	n.pinRows = append(n.pinRows, row)
	n.pinList.Append(row)

	state := *n.ClusterState
	state.SelectedObject = observer.NewProperty[client.Object](object)
	navView := adw.NewNavigationView()
	navView.SetName(fmt.Sprintf("%s/%s", ref.Namespace, ref.Name))
	ctx, cancel := context.WithCancel(n.ctx)
	n.cancelFuncs[fmt.Sprintf("%s/%s", ref.Namespace, ref.Name)] = cancel
	navView.Push(NewObjectView(ctx, &state, n.editor, navView, n).NavigationPage)
	n.pinViews = append(n.pinViews, navView)

	page := n.viewStack.AddChild(navView)
	page.SetName(fmt.Sprintf("%s/%s", ref.Namespace, ref.Name))

	return row
}

func (n *Navigation) removePin(ref corev1.ObjectReference) {
outer:
	for i, row := range n.pinRows {
		for _, c := range row.CSSClasses() {
			if c == fmt.Sprintf("%s/%s", ref.Namespace, ref.Name) {
				n.pinList.Remove(row)
				n.pinRows = slices.Delete(n.pinRows, i, i+1)
				break outer
			}
		}
	}

	for i, v := range n.pinViews {
		if v.Name() == fmt.Sprintf("%s/%s", ref.Namespace, ref.Name) {
			n.viewStack.Remove(v)
			n.pinViews = slices.Delete(n.pinViews, i, i+1)
			break
		}
	}

	if cancel, ok := n.cancelFuncs[fmt.Sprintf("%s/%s", ref.Namespace, ref.Name)]; ok {
		cancel()
		delete(n.cancelFuncs, fmt.Sprintf("%s/%s", ref.Namespace, ref.Name))
	}
}

func (n *Navigation) AddPin(object client.Object) {
	ref, err := reference.GetReference(n.Scheme, object)
	if err != nil {
		log.Print(err.Error())
		return
	}
	prefs := n.ClusterPreferences.Value()
	for _, p := range prefs.Navigation.Pins {
		if p.Name == ref.Name && p.Namespace == ref.Namespace {
			return
		}
	}
	prefs.Navigation.Pins = append(prefs.Navigation.Pins, *ref)
	n.ClusterPreferences.Update(prefs)
	n.updatePins(prefs.Navigation.Pins)

pins:
	for _, row := range n.pinRows {
		for _, c := range row.CSSClasses() {
			if c == fmt.Sprintf("%s/%s", object.GetNamespace(), object.GetName()) {
				n.pinList.SelectRow(row)
				break pins
			}
		}
	}

	n.pinsToggle.Activate()
}

func (n *Navigation) RemovePin(object client.Object) {
	ref, err := reference.GetReference(n.Scheme, object)
	if err != nil {
		log.Print(err.Error())
		return
	}

	n.removePin(*ref)

	prefs := n.ClusterPreferences.Value()
	for i, p := range prefs.Navigation.Pins {
		if p.Name == object.GetName() && p.Namespace == object.GetNamespace() {
			prefs.Navigation.Pins = slices.Delete(prefs.Navigation.Pins, i, i+1)
			break
		}
	}
	n.ClusterPreferences.Update(prefs)

	if len(n.pinRows) == 0 {
		n.resourcesToggle.SetActive(true)
	} else if n.pinList.SelectedRow() == nil {
		n.pinList.SelectRow(n.pinRows[0])
		n.pinList.SelectedRow().Activate()
	}
}

func resourceImage(gvk schema.GroupVersionKind) *gtk.Image {
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
		}
	case batchv1.GroupName:
		switch gvk.Kind {
		case "Job":
			return gtk.NewImageFromIconName("briefcase-symbolic")
		case "CronJob":
			return gtk.NewImageFromIconName("timer-reset-symbolic")
		}
	case networkingv1.GroupName:
		switch gvk.Kind {
		case "Ingress":
			return gtk.NewImageFromIconName("radio-tower-symbolic")
		}
	}

	return gtk.NewImageFromIconName("blocks")
}

func bindStatusCount(ctx context.Context, informer informers.GenericInformer, callback func(map[widget.StatusType]int)) cache.ResourceEventHandlerRegistration {
	if !cache.WaitForCacheSync(ctx.Done(), informer.Informer().HasSynced) {
		return nil
	}

	updateLabels, _ := debounce.Debounce(func() {
		var objects []client.Object
		err := cache.ListAll(informer.Informer().GetIndexer(), labels.Everything(), func(m interface{}) {
			objects = append(objects, m.(client.Object))
		})
		if err != nil {
			klog.Warning("list for status: %v", err)
			return
		}

		statuses := map[widget.StatusType]int{}
		for _, o := range objects {
			status := widget.ObjectStatus(o)
			statuses[status.Type]++
		}
		callback(statuses)
	}, time.Second)

	reg, _ := informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			updateLabels()
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			updateLabels()
		},
	})

	return reg
}
