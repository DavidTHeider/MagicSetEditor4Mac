//go:build windows

package app

// prozessLebt sagt unter Windows **immer ja** — beschrieben, nicht gebaut
// (§12): ein falscher Abbruch wäre schlimmer als ein ausgefallener Hinweis.
func prozessLebt(int) bool { return true }
