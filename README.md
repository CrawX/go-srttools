# go-srttools
Small toolkit for timestamp manipulation in .srt subtitles written in go.

Currently supports stretching and shrinking.

# Motivation
I wrote this because I was unable to find a simple solution that allowed me to adjust the framerate of a subtitle. This occurs a lot when you try to use subtitles made for TV (23.976fps) to on BluRay versions of tvshows (25fps).

#Usage
Build via `go build`.

Run e.g. `./go-srttoools -in in.srt -infps 23.976 -out out.srt -outfps 25`
