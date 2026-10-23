package core

import "runtime"

// Mobile is true on Android and iOS, where there is no window frame,
// no file manager to reveal files in, and no local Python to run
// models: engines there are API servers.
const Mobile = runtime.GOOS == "android" || runtime.GOOS == "ios"
