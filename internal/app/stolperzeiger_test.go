package app_test

import "context"

// stolperzeiger ist ein Zeiger, der nicht sagt, wer arbeitet.
//
// **Ohne Build-Tag**, weil ihn zwei Dateien brauchen und nur eine davon an
// darwin hängt: `leserverbindung_test.go` und `befehlausfuehren_test.go`.
// Bis zum 2026-09-09 stand er in der ersten, und `golangci-lint` für Windows
// fand ihn in der zweiten nicht — der Lauf für darwin war grün (§4).
type stolperzeiger struct{ fehler error }

func (z stolperzeiger) Nutzerzeiger(context.Context) (int64, error)  { return 0, z.fehler }
func (z stolperzeiger) Revision(context.Context) (int64, error)      { return 0, z.fehler }
func (z stolperzeiger) Arbeitsprofil(context.Context) (int64, error) { return 0, z.fehler }
