#!/bin/bash
flatpak run org.inkscape.Inkscape -o 256.png -w 256 -h 256 ../../internal/icon/seabird.svg
convert -define icon:auto-resize=256,96,48,32,24,16 256.png -colors 256 icon.ico
rm *.png