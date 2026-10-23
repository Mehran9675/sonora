//go:build !android && !ios

package paths

// mobileDataDir is only used on Android and iOS.
func mobileDataDir() string { return "" }
