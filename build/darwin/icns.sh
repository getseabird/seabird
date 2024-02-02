#!/bin/bash
for size in 16 32 48 128 256; do
  inkscape -o $size.png -w $size -h $size ../../icon/seabird.svg
done
png2icns icon.icns 16.png 32.png 48.png 128.png 256.png
rm *.png