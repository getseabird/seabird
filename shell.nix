{ pkgs ? import <nixpkgs> { } }:
pkgs.mkShell {
  nativeBuildInputs = with pkgs; [
    gtk4
    libadwaita
    gtksourceview5
  ];
}
