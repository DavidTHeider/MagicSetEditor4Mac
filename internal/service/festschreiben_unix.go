//go:build unix

package service

import "syscall"

// leichtFestschreiben nimmt `fsync` — unter macOS **nicht** `F_FULLFSYNC`,
// das `(*os.File).Sync` dort nimmt: dieselbe Stufe wie SQLite mit
// `synchronous=FULL` (Vermerk 0174).
var leichtFestschreiben = func(d nachbardatei) error {
	return syscall.Fsync(int(d.Fd())) // #nosec G115 -- ein Dateigriff passt in int
}
