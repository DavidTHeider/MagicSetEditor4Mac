//go:build !unix

// Die Sperre, wo es `flock` nicht gibt.
//
// **Ohne Sperre bleibt der Stand vor Phase 3.2**: die Aufrufer prüfen
// weiterhin, was sie ohne Absprache prüfen können, und das Fenster zwischen
// Prüfung und Tausch bleibt offen. Das gehört in
// [`PORTIERUNG.md`](../../docs/PORTIERUNG.md), nicht in eine stille Zeile.

package dateisperre

import "os"

// teilen tut nichts und gelingt.
func teilen(*os.File) error { return nil }

// allein tut nichts und gelingt.
func allein(*os.File) error { return nil }
