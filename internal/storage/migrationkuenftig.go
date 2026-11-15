package storage

// **Eine künftige Fassung, für Läufe** — die heutigen Wanderungen und weitere
// dazu. Eigene Datei, weil sie eine eigene Frage beantwortet: wie sieht die
// Fassung nach dieser aus? `migrator.go` beantwortet, wie gewandert wird.

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"slices"
	"strings"
)

// MigriereMit ist [Migriere] für eine **künftige** Fassung: die Wanderungen
// dieser und dazu `weitere`, unter `migrationen/`.
//
// **Für Läufe, die eine künftige Fassung brauchen** — ein abbrechendes
// Upgrade, eine neuere Fassung im Wettlauf. Der Betrieb nimmt [Migriere].
//
// **Dazu und nicht statt.** Hier stand bis M01
// (#2345)
// „mit anderen Wanderungen": solange es nach der Grundlinie keine gab, war
// beides dasselbe. Mit `0098` sah jede so gebaute „künftige" Fassung einen
// heutigen Bestand als neueren — sechs Läufe auf einmal. Eine künftige Fassung
// trägt die heutigen Wanderungen mit.
func MigriereMit(ctx context.Context, db *sql.DB, weitere fs.FS, vorbereiten Vorbereitung) error {
	return migriereAusMitVorbereitung(ctx, db, ueberlagert{heute: migrationen, weitere: weitere},
		PruefeBestand, vorbereiten, grundlinie)
}

// ueberlagert legt die Wanderungen einer künftigen Fassung über die heutigen.
type ueberlagert struct {
	heute, weitere fs.FS
}

// Open nimmt die weitere Datei vor der heutigen — **und nur, was dort
// fehlt, kommt von heute** (Zweitblick über Phase 2, Befund 12): eine
// vorhandene, die sich nicht lesen ließ, ersetzte sonst still die heutige.
func (u ueberlagert) Open(name string) (fs.File, error) {
	datei, fehler := u.weitere.Open(name)
	if !errors.Is(fehler, fs.ErrNotExist) {
		return datei, fehler
	}

	return u.heute.Open(name)
}

// ReadDir nennt die Einträge beider, je Name einmal und geordnet.
func (u ueberlagert) ReadDir(name string) ([]fs.DirEntry, error) {
	heute, heutefehler := fs.ReadDir(u.heute, name)
	weitere, weiterfehler := fs.ReadDir(u.weitere, name)

	if fehler := errors.Join(heutefehler, weiterfehler); fehler != nil {
		return nil, fehler
	}

	je := make(map[string]fs.DirEntry, len(heute)+len(weitere))
	for _, eintrag := range append(heute, weitere...) {
		je[eintrag.Name()] = eintrag
	}

	eintraege := make([]fs.DirEntry, 0, len(je))
	for _, eintrag := range je {
		eintraege = append(eintraege, eintrag)
	}

	slices.SortFunc(eintraege, func(a, b fs.DirEntry) int { return strings.Compare(a.Name(), b.Name()) })

	return eintraege, nil
}
