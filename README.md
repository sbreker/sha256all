# sha256all

GoLang GoRoutine/channel/waitgroup example. main.go launches 2 types of go routines: 1 sender sends filenames to N receivers that calculate and display the sha256sum.

## Usage

    sha256all [options]... [path]

Examples:

    sha256all /home/user/data
    sha256all -buffer=true /home/user/data
    sha256all -cpuprofile=/tmp/prof1 /home/user/data

## CPU profiling

Enable CPU profiling with:

    go run . -cpuprofile=/tmp/prof1 [path]

Inspect the profile:

    go tool pprof -http=:12345 /tmp/prof1
