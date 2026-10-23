//go:build android || ios

package paths

import "gioui.org/app"

// mobileDataDir is the app's private storage (Context.getFilesDir on
// Android).
func mobileDataDir() string {
	dir, _ := app.DataDir()
	return dir
}
