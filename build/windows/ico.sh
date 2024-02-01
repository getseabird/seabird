#!/bin/bash
for size in 16 32 48 128 256; do
    flatpak run org.inkscape.Inkscape -o $size.png -w $size -h $size ../../icon/seabird.svg
done
convert 16.png 32.png 48.png 128.png 256.png -colors 256 icon.ico
rm *.png