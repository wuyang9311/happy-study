package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/wuyang9311/happy-study/pkg/version"
)

func main() {
	fmt.Println("🎉 Happy Study - Go Learning Project")
	fmt.Println()

	// Print version info
	info := version.Get()
	fmt.Printf("Version  : %s\n", info.Version)
	fmt.Printf("Commit   : %s\n", info.Commit)
	fmt.Printf("BuildTime: %s\n", info.BuildTime)
	fmt.Printf("Go       : %s\n", info.GoVersion)

	// Print system info
	fmt.Println()
	fmt.Printf("OS/Arch  : %s/%s\n", runtime.GOOS, runtime.GOARCH)

	os.Exit(0)
}
