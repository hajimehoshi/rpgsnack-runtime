RPGSnack Runtime

# How to install and run (macOS)

1. Install [Go](https://golang.org/)
2. Run `go get -u github.com/hajimehoshi/rpgsnack-runtime/...`
3. Run `cd ~/go/src/github.com/hajimehoshi/rpgsnack-runtime`
4. Run `go run . /path/to/project`

## How to run on Android (for testing)

```sh
go run ~/go/src/github.com/hajimehoshi/rpgsnack-runtime/tools/testserver/main.go -http=:7800 -basepath=<project local location>
gomobile install -tags="gomobilebuild" -ldflags='-X github.com/hajimehoshi/rpgsnack-runtime/internal/game.injectedProjectLocation=http://<your machine IP>:7800' github.com/hajimehoshi/rpgsnack-runtime
```

## How to run on iOS (for testing)

```sh
go run ~/go/src/github.com/hajimehoshi/rpgsnack-runtime/tools/testserver/main.go -http=:7800 -basepath=<project local location>
gomobile build -target=ios -tags="gomobilebuild" -ldflags='-X github.com/hajimehoshi/rpgsnack-runtime/internal/game.injectedProjectLocation=http://<your machine IP>:7800' -work github.com/hajimehoshi/rpgsnack-runtime
```

Then, open the `WORK` directory as the last command shows, and the open `main.xcodeproj`. Specify the singing and run it.

## How to create .framework file for iOS

1. Install gomobile with `go get golang.org/x/mobile/cmd/...`
2. Run `gomobile bind -target ios -o ./RPGSnackRuntime.framework github.com/hajimehoshi/rpgsnack-runtime/mobile`

## How to create .aar file for Android

1. Install gomobile with `go get golang.org/x/mobile/cmd/...`
2. Run `gomobile bind -target android -javapkg net.rpgsnack.runtime -o ./rpgsnack_runtime.aar github.com/hajimehoshi/rpgsnack-runtime/mobile`

## How to test on browsers

```sh
gopherjs serve
go run ~/go/src/github.com/hajimehoshi/rpgsnack-runtime/tools/testserver/main.go -http=:7800 -basepath=<project local location>
# If basepath includes '~', '=' should be omitted like '-basepath ~/foo/bar'
open http://localhost:8080/github.com/hajimehoshi/rpgsnack-runtime?project_location=http://localhost:7800/
```

If you find some crashes, try

```
rm -rf ~/go/pkg/*_js
```
