#!/bin/bash
ffmpeg -framerate 30 -pattern_type glob -i 'frame*.jpg' -c:v libx264 -r 30 -pix_fmt yuv420p video.mp4
