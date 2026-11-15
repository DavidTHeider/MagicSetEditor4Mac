//go:build windows

package app

// prozesszeitMs ist auf Windows null — beschrieben und nicht gebaut (§12).
func prozesszeitMs() int64 { return 0 }
