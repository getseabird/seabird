# Seabird

Seabird is a native cross-platform Kubernetes desktop client that makes it super
easy to explore your cluster's resources. We aim to visualize all common
resource types in a simple, bloat-free user interface.

![Screenshot](https://getseabird.github.io/images/screenshot.png)

## Download

Downloads for all platforms are available under
[releases](https://github.com/getseabird/seabird/releases). On Linux, we
recommend using the Flatpak package.

<a href='https://flathub.org/apps/dev.skynomads.Seabird'>
  <img width='180' alt='Download on Flathub' src='https://flathub.org/api/badge?locale=en'/>
</a>

## Building From Source

Build dependencies

#### Fedora

```bash
sudo dnf install gtk4-devel gtksourceview5-devel libadwaita-devel gobject-introspection-devel glib2-devel golang
```

#### Debian

```bash
sudo apt install libgtk-4-dev libgtksourceview-5-dev libadwaita-1-dev libgirepository1.0-dev libglib2.0-dev-bin golang-go
```

Run go generate to create the embedded resource file:

```bash
go generate ./...
```

Then build with:

```bash
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
