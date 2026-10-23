package manager

import "runtime"

// mobile is true on Android and iOS, where engines are API servers.
const mobile = runtime.GOOS == "android" || runtime.GOOS == "ios"
