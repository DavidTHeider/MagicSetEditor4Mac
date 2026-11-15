//go:build darwin

// **Ein Leser schreibt kein Symbol auf die Platte** — DOM-02
// (#1697),
// nach [0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md).
//
// Einsetzen legt eine Kopie neben das Set, Nachziehen überschreibt die Kopien
// fremder Sets, Ausgeben schreibt, wohin es heißt — **alle drei vor jedem
// Wert im Bestand**, und die lesende Verbindung hält keine Datei auf. Gefragt
// wird deshalb vorher, wie beim Export.
//
// **Die Kennung gibt es im fremden Profil nicht**: ohne Riegel käme
// »nicht gefunden« zurück, mit ihm »nur lesend« — der Code unterscheidet die
// beiden Wege.

package app_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestEinLeserSchreibtKeinSymbol(t *testing.T) {
	anwendung, geoeffnet := alsLeserVorEinemSet(t)
	ziel := filepath.Join(t.TempDir(), "krone.mse-symbol")

	if liste := anwendung.Symbolbibliothek(t.Context(), ""); liste.Fehler != nil {
		t.Fatalf("der Leser sieht die Bibliothek nicht: %+v", liste.Fehler)
	}

	for name, fehler := range map[string]*app.Fehler{
		"Einsetzen":  anwendung.BibliothekssymbolEinsetzen(t.Context(), "1", geoeffnet.SetID).Fehler,
		"Nachziehen": anwendung.BibliothekssymbolNachziehen(t.Context(), "1").Fehler,
		"Ausgeben":   anwendung.BibliothekssymbolAusgabeSchreiben(t.Context(), "1", ziel).Fehler,
	} {
		if fehler == nil || fehler.Code != app.CodeNurLesend {
			t.Errorf("%s: %+v, erwartet %q", name, fehler, app.CodeNurLesend)
		}
	}

	if _, err := os.Stat(ziel); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("die Ausgabe liegt da: %v", err)
	}
}
