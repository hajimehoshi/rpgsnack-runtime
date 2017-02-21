RPGSnack Runtime

# How to install and run (macOS)

1. Install [Go](https://golang.org/)
2. Set the environment variable 'GOPATH' as e.g. `~/go`
3. Run `go get -u github.com/hajimehoshi/rpgsnack-runtime/...`
4. Run `cd $GOPATH/src/github.com/hajimehoshi/rpgsnack-runtime`
5. Run `go run main.go`

## How to specify a JSON file to run

Run `go run main.go /path/to/json/file`

## How to create .framework file for iOS

1. Install gomobile with `go get golang.org/x/mobile/cmd/...`
2. Run `gomobile bind -target ios -o ./RPGSnackRuntime.framework github.com/hajimehoshi/rpgsnack-runtime/mobile`

## How to create .aar file for Android

1. Install gomobile with `go get golang.org/x/mobile/cmd/...`
2. Run `gomobile bind -target android -javapkg net.rpgsnack.runtime -o ./rpgsnack_runtime.aar github.com/hajimehoshi/rpgsnack-runtime/mobile`

## How to pack resources
```
go generate ./...
```