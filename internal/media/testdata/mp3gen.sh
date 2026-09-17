#!/usr/bin/env bash
set -e

ffmpeg \
  -f lavfi \
  -i anullsrc=r=44100:cl=mono \
  -i cover.png \
  -t 3 \
  -map 0:a -map 1:v \
  -metadata title="Test Audio Track" \
  -metadata artist="Test Artist" \
  -metadata album_artist="Test Album Artist" \
  -metadata comment="Test Comment" \
  -codec:a libmp3lame \
  -codec:v mjpeg \
  -b:a 32k \
  -id3v2_version 3 \
  three_sec_tagged_with_cover.mp3