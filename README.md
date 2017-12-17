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

# How to downgrade NDK

There is a known issue that Go doesn't work with NDK r16 (https://github.com/golang/go/issues/22766). Let's use r15 (or older) until the issue is fixed.

1. Download r15c version of NDK from https://developer.android.com/ndk/downloads/older_releases.html
2. Unzip `android-ndk-r15c-*-x86_64.zip`
3. `cd $ANDROID_HOME`
4. `mv ndk-bundle ndk-bundle.r16` for example.
5. Move the unziped directory to `$ANDROID_HOME/nkd-bundle`
6. `gomobile init`
