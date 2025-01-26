# Go module interfacing with X52 Pro Joystics

Logitech (formerly Saitek) has a series of gaming devices, primarily
a joystic called X52 Pro. This module can talk to it via the DirectOutput API included
in the official driver.

It only works on Windows since it uses the DirectOutput.dll provided
in the driver installation.

Much inspiration comes from [Trade-Dangerous](https://github.com/eyeonus/Trade-Dangerous/blob/release/v1/tradedangerous/mfd/saitek/)
but I would like something in Go and to be able to use it separatly.


## Disclaimer

It is very much a Work In Progress and contains bugs, missing features etc.
It might destroy everything on your machine (unlikely) and 
crash completely (somewhat likely). 
It will eventually be strict about semver versioning but for now, every 
change could be a breaking one.

My available time to work on this project will be limited but I will gladly accept
pull requests that I find reasonable.


## Usage
- Install the driver which can, at time of writing, be found at 
[logi.com](https://support.logi.com/hc/en-gb/articles/360024838173--Downloads-X52-Professional-Space-Flight-H-O-T-A-S)

The software needs DirectOutput.dll which normally can be found at
`C:\Program Files\Logitech\DirectOutput\DirectOutput.dll` after
driver installation.

Then run
`go run cmd\mfdcli`


## Development

Install the driver and you will find documentation and dlls in 
`C:\Program Files\Logitech\DirectOutput\SDK`.
Since [DirectOutput.h](contrib/saitek/DirectOutput.h) is under a licens
identical to MIT, it is included here under contrib/saitek.


## Resources
- https://github.com/eyeonus/Trade-Dangerous/blob/release/v1/tradedangerous/mfd/saitek/directoutput.py
- https://leandrofroes.github.io/posts/An-in-depth-look-at-Golang-Windows-calls/
- https://go.dev/wiki/WindowsDLLs
- https://anubissec.github.io/How-To-Call-Windows-APIs-In-Golang/#
