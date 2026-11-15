// Die Naht zum Bilderkenner — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// **Paketintern**, weil die Naht unexportiert ist: sie wird beim Start
// gesetzt und von niemandem sonst gebaut. Was hier gemessen wird, ist der
// Satz, für den es sie gibt — **sie steht nach dem Start**, und sie führt an
// **dieselbe** Drucksuche wie eine getippte Kartenliste.
package app

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ohne Bestand sagt sie es — statt abzustürzen.**
//
// Der Fall ist kein Kunstgriff: die Naht hängt an der **laufenden**
// Verbindung, und die ist vor dem ersten Öffnen und nach einem Zurückspielen
// `nil`.
func TestDieBilderkennungsnahtOhneBestand(t *testing.T) {
	t.Parallel()

	naht := bilderkennungsnaht{holen: func() *sql.DB { return nil }}

	_, fehler := naht.Deuten(t.Context(), domain.Bilderkennung{})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Nach dem Start steht sie — und führt an die echte Drucksuche.**
//
// # Was dieser Lauf hält
//
// **Dass die Naht wirklich fragt.** Der Wächter über die Nähte sagt nur, dass
// das Feld gesetzt ist; dass dahinter eine Suche steht, sagt erst ein Server,
// der die Anfrage sieht. „Gebaut und nicht angeschlossen" sähe an beiden
// sonst gleich aus (#1539).
//
// **Und dass sie nicht nach draußen geht**: die Adresse ist umgelenkt und das
// Zuhause ein Wegwerfordner — beides zusammen, denn einzeln reicht keines
// (§3).
func TestDieBilderkennungsnahtFragtDieDrucksuche(t *testing.T) {
	gefragt := 0

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiber http.ResponseWriter, _ *http.Request) {
			gefragt++

			schreiber.Header().Set("Content-Type", "application/json")
			if _, fehler := schreiber.Write(
				[]byte(`{"object":"list","total_cards":0,` +
					`"has_more":false,"data":[]}`)); fehler != nil {
				t.Errorf("antworten: %v", fehler)
			}
		}))
	defer server.Close()

	t.Setenv("HOME", t.TempDir())

	vorher := MitScryfalladresse(server.URL)
	t.Cleanup(func() { MitScryfalladresse(vorher) })

	anwendung := NeueApp(t.TempDir() + "/bestand.db")
	if fehler := anwendung.Starten(t.Context()); fehler != nil {
		t.Fatalf("starten: %v", fehler)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	deuter, steht := anwendung.bilderkenner.(bilderkennungsnaht)
	if !steht {
		t.Fatalf("nach dem Start hängt %T am Bilderkenner", anwendung.bilderkenner)
	}

	deutung, fehler := deuter.Deuten(anwendung.imProfil(t.Context()),
		domain.Bilderkennung{Beobachtungen: []domain.Beobachtung{
			{Art: domain.EvidenzName, Wert: "Lightning Bolt", Lesbar: true},
		}})
	if fehler != nil {
		t.Fatalf("deuten: %v", fehler)
	}

	if gefragt == 0 {
		t.Error("die Naht hat niemanden gefragt — sie ist nicht angeschlossen")
	}

	if deutung.Drucke != 0 || len(deutung.Kandidaten) != 0 {
		t.Errorf("aus einer leeren Antwort wurde %+v", deutung)
	}
}

// **Ein Zuhause, das keines ist, meldet sich — statt abzustürzen.**
//
// # Warum dieser Zweig zählt
//
// **Weil [Scryfallklient] das Benutzerverzeichnis braucht** und es beim
// ersten Ruf anlegt. Wer kein schreibbares hat — ein gesperrtes Profil, eine
// volle Platte —, bekommt hier einen Satz und keine leere Kandidatenliste:
// „nichts gefunden" wäre eine Aussage über das Foto, und es ist eine über
// diesen Rechner.
func TestEinUnbrauchbaresZuhauseMeldetSich(t *testing.T) {
	// **Eine Datei statt eines Verzeichnisses.** `os.MkdirAll` scheitert
	// daran, und das ist derselbe Weg wie auf einer Platte, auf der niemand
	// schreiben darf — nur herstellbar.
	zuhause := filepath.Join(t.TempDir(), "keinordner")
	if fehler := os.WriteFile(zuhause, []byte("x"), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	t.Setenv("HOME", zuhause)

	// **Irgendeine offene Verbindung genügt** — die Naht kommt gar nicht
	// bis zur Einstellung, sie scheitert vorher am Verzeichnis.
	db, fehler := sql.Open("sqlite3", ":memory:")
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() {
		if fehler := db.Close(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	naht := bilderkennungsnaht{holen: func() *sql.DB { return db }}

	if _, fehler := naht.Deuten(t.Context(), domain.Bilderkennung{
		Beobachtungen: []domain.Beobachtung{
			{Art: domain.EvidenzName, Wert: "Lightning Bolt", Lesbar: true},
		},
	}); fehler == nil {
		t.Fatal("das unbrauchbare Zuhause kam durch")
	}
}
