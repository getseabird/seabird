FROM mcr.microsoft.com/devcontainers/go:1-1.21-bookworm

RUN apt update && apt install -y cmake libfontconfig1-dev gcc pkg-config libwayland-dev libgles2-mesa-dev libegl1-mesa-dev libffi-dev libxcursor-dev libgtk-4-dev libgirepository1.0-dev libgraphene-1.0-dev dbus-x11 libgl1-mesa-dev libadwaita-1-dev libnotify-bin libgtksourceview-5-dev meson libsass-dev libdbus-1-dev glslc libgstreamer-plugins-bad1.0-dev libappstream-dev valac flatpak-builder

RUN git clone --depth 1 --branch 2.76.6 https://gitlab.gnome.org/GNOME/glib.git /tmp/glib && \
  cd /tmp/glib && meson setup _build && meson compile -C _build && sudo ninja -C _build install

RUN git clone --depth 1 --branch 1.4.2 https://gitlab.gnome.org/GNOME/libadwaita.git /tmp/libadwaita && \
  cd /tmp/libadwaita && meson setup _build && meson compile -C _build && sudo ninja -C _build install