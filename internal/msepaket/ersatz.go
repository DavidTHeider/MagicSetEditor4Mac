// **Ein Paket, in dem eine Datei aus dem Speicher kommt** — Werkstück 3 von
// Phase 3.3 (#1164).
//
// # Wofür
//
// Für die Vorschau einer **ungespeicherten** Änderung. §6 verlangt: „Eine
// Änderung an einem Feld ist in der Vorschau zu sehen, ohne dass jemand etwas
// bestätigt" — und Speichern ist eine Bestätigung. Eine Vorlage, die man erst
// schreiben müsste, um sie zu sehen, machte aus jedem Versuch eine Änderung an
// **allem**, was dieses Paket benutzt.
//
// # Warum ein Ersatz und keine Kopie
//
// Eine Kopie des Pakets kostete bei jedem Tastendruck alles, was darin liegt.
// Der Ersatz legt **eine** Datei über den Baum und lässt den Rest, wo er ist.

package msepaket

import (
	"errors"
	"io"
	"io/fs"
	"time"
)

// OeffneMitErsatz öffnet ein Paket, in dem eine Datei einen anderen Inhalt hat.
//
// **Auf der Platte ändert sich nichts.** Der Ersatz lebt so lange wie das
// geöffnete Paket.
func OeffneMitErsatz(pfad, name string, inhalt []byte) (*Paket, error) {
	sauber, err := paketpfad(name)
	if err != nil {
		return nil, err
	}

	paket := &Paket{Pfad: pfad, Art: ArtAusName(pfad)}

	dateien, schliesser, oeffnefehler := paketbaum(pfad)
	if oeffnefehler != nil {
		return nil, oeffnefehler
	}

	paket.dateien = ersatzbaum{darunter: dateien, name: sauber, inhalt: inhalt}
	paket.schliesser = schliesser

	if fehler := paket.kopfLesen(); fehler != nil {
		return nil, errors.Join(fehler, paket.Schliessen())
	}

	return paket, nil
}

// ersatzbaum gibt eine Datei aus dem Speicher und alles andere von darunter.
type ersatzbaum struct {
	darunter fs.FS
	name     string
	inhalt   []byte
}

func (e ersatzbaum) Open(name string) (fs.File, error) {
	if name != e.name {
		return e.darunter.Open(name)
	}

	return &ersatzdatei{name: name, inhalt: e.inhalt}, nil
}

// ersatzdatei ist die ersetzte Datei als [fs.File].
type ersatzdatei struct {
	name    string
	inhalt  []byte
	gelesen int
}

func (d *ersatzdatei) Stat() (fs.FileInfo, error) { return ersatzangabe{datei: d}, nil }

func (d *ersatzdatei) Read(nach []byte) (int, error) {
	if d.gelesen >= len(d.inhalt) {
		return 0, io.EOF
	}

	genommen := copy(nach, d.inhalt[d.gelesen:])
	d.gelesen += genommen

	return genommen, nil
}

func (d *ersatzdatei) Close() error { return nil }

// ersatzangabe beschreibt die ersetzte Datei.
//
// **Die Zeit ist die Nullzeit.** Sie wird von niemandem gelesen — `Paket`
// fragt nach Inhalt und Name —, und eine erfundene Uhrzeit wäre eine Angabe,
// die etwas behauptet.
type ersatzangabe struct{ datei *ersatzdatei }

func (a ersatzangabe) Name() string       { return a.datei.name }
func (a ersatzangabe) Size() int64        { return int64(len(a.datei.inhalt)) }
func (a ersatzangabe) Mode() fs.FileMode  { return 0 }
func (a ersatzangabe) ModTime() time.Time { return time.Time{} }
func (a ersatzangabe) IsDir() bool        { return false }
func (a ersatzangabe) Sys() any           { return nil }
