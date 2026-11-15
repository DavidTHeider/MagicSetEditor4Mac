package netz_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// schreibe schreibt in eine Antwort und meldet, wenn es nicht ging.
//
// **In einer Gegenstelle für Tests**, wo ein Schreibfehler nur eines heißen
// kann: der Test hat sich selbst im Weg gestanden.
func schreibe(t *testing.T, ziel io.Writer, inhalt []byte) {
	t.Helper()

	if _, err := ziel.Write(inhalt); err != nil {
		t.Errorf("die Gegenstelle konnte nicht schreiben: %v", err)
	}
}

// contextMitAbbruch gibt einen Kontext samt seinem Abbruch.
func contextMitAbbruch(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()

	ctx, abbrechen := context.WithCancel(t.Context())
	t.Cleanup(abbrechen)

	return ctx, abbrechen
}

// im bündelt Helfer, die dem Speicher etwas in den Weg legen.
var im imWeg

type imWeg struct{}

// Weg legt ein Verzeichnis dorthin, wo der Speicher eine Datei erwartet.
//
// **Der Pfad wird nachgerechnet**, weil er eine Prüfsumme ist — anders kommt
// das Verzeichnis nicht an die richtige Stelle.
func (imWeg) Weg(t *testing.T, ort, schluessel, endung string) {
	t.Helper()

	if err := os.MkdirAll(ort, 0o700); err != nil {
		t.Fatalf("den Ort anlegen: %v", err)
	}

	summe := sha256.Sum256([]byte(schluessel))
	pfad := filepath.Join(ort, hex.EncodeToString(summe[:])+".json"+endung)

	if err := os.Mkdir(pfad, 0o700); err != nil {
		t.Fatalf("das Hindernis anlegen: %v", err)
	}
}
