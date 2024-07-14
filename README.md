# Seabird

Seabird is a Kubernetes IDE designed for the GNOME desktop. Explore and manage
your clusters with a simple and intuitive interface. Equipped with essential
features such as a terminal for executing commands, monitoring through logs and
metrics, and a resource editor that conveniently places the API reference at
your fingertips.

![Screenshot](https://getseabird.github.io/images/screenshot.png)

## Download

Downloads for all platforms are available under
[releases](https://github.com/getseabird/seabird/releases). On Linux, we
recommend using the Flatpak package.

<a href='https://flathub.org/apps/dev.skynomads.Seabird'>
  <img width='140' alt='Download on Flathub' src='https://flathub.org/api/badge?locale=en'/>
</a>

## Building From Source

Build dependencies

#### Fedora

```sh
sudo dnf install gtk4-devel gtksourceview5-devel libadwaita-devel gobject-introspection-devel glib2-devel vte291-gtk4-devel golang
```

#### Debian

```sh
sudo apt install libgtk-4-dev libgtksourceview-5-dev libadwaita-1-dev libgirepository1.0-dev libglib2.0-dev-bin libvte-2.91-gtk4-dev golang-go
```

Run go generate to create the embedded resource file:

```sh
go generate ./...
```

Then build with:

```sh
go build
```

## Reporting Issues

If you experience problems, please open an
[issue](github.com/getseabird/seabird/issues). Try to include as much
information as possible, such as version, operating system and reproduction
steps.

For feature suggestions, please create a
[discussion](https://github.com/getseabird/seabird/discussions). If you have a
concrete vision for the feature, open an issue instead and use the proposal
template.

## License

Seabird is available under the terms of the Mozilla Public License v2, a copy of
the license is distributed in the LICENSE file.

Note: This is paid software with an unlimited free trial.
