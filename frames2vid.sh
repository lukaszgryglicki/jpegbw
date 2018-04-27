#!/bin/bash
ffmpeg -framerate 10 -pattern_type glob -i 'frame*.jpg' -c:v libx264 -r 10 -pix_fmt yuv420p video.mp4
