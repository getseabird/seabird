# Seabird

Seabird is a native cross-platform Kubernetes desktop client that makes it super
easy to explore your cluster's resources. We aim to visualize all common
resource types in a simple, bloat-free user interface.

![Screenshot](https://getseabird.github.io/images/screenshot.png)

Builds are currently available for Linux and Windows. Note that this is
**alpha** quality software with missing features and bugs.

## Requirements

Seabird requires libadwaita (>1.4) and gtksourceview 5 to run. The Windows
builds include all dependencies.

#### Fedora

```bash
sudo dnf install libadwaita gtksourceview5
```

#### Debian

```bash
sudo apt install libadwaita-1 libgtksourceview-5
```

Note: Releases older than Debian Trixie or Ubuntu Mantic are not supported.
Please use the Flatpak instead.

## Building From Source

Build dependencies

#### Debian
```bash
sudo apt install libgtk-4-dev libgtksourceview-5-dev libadwaita-1-dev libgirepository1.0-dev libglib2.0-dev-bin golang-go
```

#### Fedora
```bash
sudo dnf install gtk4-devel gtksourceview5-devel libadwaita-devel gobject-introspection-devel glib2-devel golang
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
concrete vision for the feature and are prepared to implement it, open an issue
instead.

## License

Seabird is available under the terms of the Mozilla Public License v2, a copy of
the license is distributed in the LICENSE file.

Disclosure: We plan to distribute this application with an semi-optional yearly
subscription price to support development.
