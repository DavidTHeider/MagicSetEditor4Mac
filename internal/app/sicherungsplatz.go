package app

// **Eine Sicherung, für die der Platz nicht reicht, heißt „voll"** — M02
// (#2346).
//
// **SQLite meldet einen vollen Datenträger beim Sichern nicht als `FULL`.**
// Gemessen mit der Hüllenprobe `wanderung09` am 2026-10-14: auf einem Abbild
// mit 2 MB frei scheitert `VACUUM INTO` für einen Bestand von 4 MB mit einem
// Fehler des Datenträgers, und nach dem Räumen der halben Sicherung sind
// wieder 2 MB frei. Die Regel des Starts — weniger als ein Mebibyte,
// [vollStattEA] — sieht das nicht; **eine Sicherung braucht, was der Bestand
// belegt** ([storage.Sicherungsbedarf]). Gefragt wird, wie dort, nur nach
// einem Fehler des Datenträgers.
//
// **Für jede Sicherung über diese Naht**: die vor der Wanderung, an jeder der
// vier Wurzeln, und die auf Wunsch — beide über [sichernUeber].

import (
	"context"
	"path/filepath"
)

// mitPlatzfrage gibt `sichern` mit dieser Einordnung.
func mitPlatzfrage(
	sichern func(context.Context, string) error,
	bedarf func(context.Context) int64,
	frei func(string) (int64, error),
) func(context.Context, string) error {
	return func(ctx context.Context, ziel string) error {
		fehler := sichern(ctx, ziel)

		return vollStattEABei(fehler, filepath.Dir(ziel), frei, max(bedarf(ctx), knapperPlatz))
	}
}
