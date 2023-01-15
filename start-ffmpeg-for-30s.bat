@echo off
tasklist /nh|findstr /i "ffmpeg.exe"
if ERRORLEVEL 1 (F:\tools\ffmpeg\ffmpeg -i "rtsp://192.168.31.96:8554/home" -y -f image2 -r 4/1 -update 1   -s 960x540  -vf format=gray  -t  30  F:\tools\ffmpeg\rtsp.jpg) else (echo "ffmpeg already exist")