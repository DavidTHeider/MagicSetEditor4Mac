//go:build darwin

// **Das Geheimnis eines Anmeldeanbieters** —
// #1401,
// Befund 3 des Zweitblicks über Phase 4.12.
//
// # Warum es diese Bindungen gibt
//
// **GitHub verlangt ein `client_secret`**, und keine Bindung schrieb eines.
// Wer sich über GitHub anmelden wollte, musste die Schlüsselbundverwaltung
// öffnen — §6 sagt, ein leerer Zustand erklärt den nächsten Schritt, nicht:
// er verweist auf ein anderes Programm.

package app_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Liste sagt je Anbieter, was er braucht und was vorliegt.**
//
// **Und nie den Wert** — dieselbe Zusage wie bei `Zugaenge`: ein Weg, der ihn
// zurückgäbe, machte aus dem Schlüsselbund eine Durchreiche.
func TestDieAnbietergeheimnisseSagenWasFehlt(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitSchluesselbund(neuerBundspeicher())

	antwort := anwendung.Anbietergeheimnisse(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Anbieter) != 4 {
		t.Fatalf("es kamen %d Anbieter", len(antwort.Anbieter))
	}

	nach := map[string]app.AnbietergeheimnisDTO{}
	for _, eintrag := range antwort.Anbieter {
		nach[eintrag.Anbieter] = eintrag
	}

	// **Je Anbieter nachgelesen** (`internal/anmeldung/anbieter.go`): GitHub
	// führt es als *Required*, Microsoft verbietet es für öffentliche
	// Klienten, Google nennt es freigestellt.
	if !nach["github"].Pflicht {
		t.Error("GitHub verlangt eines, und die Liste sagt es nicht")
	}

	if !nach["microsoft"].Verboten {
		t.Error("Microsoft verbietet eines, und die Liste sagt es nicht")
	}

	if nach["google"].Pflicht || nach["google"].Verboten {
		t.Errorf("Google ist freigestellt: %+v", nach["google"])
	}

	for name, eintrag := range nach {
		if eintrag.Hinterlegt {
			t.Errorf("%s meldet ein Geheimnis, das nie geschrieben wurde", name)
		}
	}
}

// **Ein gesetztes Geheimnis liegt danach vor — und niemand liest es zurück.**
func TestEinGesetztesAnbietergeheimnisLiegtVor(t *testing.T) {
	t.Parallel()

	bund := neuerBundspeicher()

	anwendung := app.NeueApp("")
	anwendung.MitSchluesselbund(bund)

	if gesetzt := anwendung.AnbietergeheimnisSetzen(
		t.Context(), "github", "sehr geheim"); gesetzt.Fehler != nil {
		t.Fatalf("setzen: %+v", gesetzt.Fehler)
	}

	if bund.werte["github"] != "sehr geheim" {
		t.Errorf("im Bund steht %+v", bund.werte)
	}

	antwort := anwendung.Anbietergeheimnisse(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	for _, eintrag := range antwort.Anbieter {
		if eintrag.Anbieter == "github" && !eintrag.Hinterlegt {
			t.Error("das gesetzte Geheimnis wird nicht gemeldet")
		}
	}
}

// **Und es lässt sich wieder herausnehmen.**
func TestEinAnbietergeheimnisLaesstSichLoeschen(t *testing.T) {
	t.Parallel()

	bund := neuerBundspeicher()

	anwendung := app.NeueApp("")
	anwendung.MitSchluesselbund(bund)

	if gesetzt := anwendung.AnbietergeheimnisSetzen(
		t.Context(), "github", "sehr geheim"); gesetzt.Fehler != nil {
		t.Fatalf("setzen: %+v", gesetzt.Fehler)
	}

	if geloescht := anwendung.AnbietergeheimnisLoeschen(
		t.Context(), "github"); geloescht.Fehler != nil {
		t.Fatalf("löschen: %+v", geloescht.Fehler)
	}

	if _, steht := bund.werte["github"]; steht {
		t.Errorf("das Geheimnis steht noch im Bund: %+v", bund.werte)
	}
}

// **Ein unbekannter Anbieter wird abgewiesen** — sonst entstünde im Bund ein
// Eintrag, den niemand je liest (§8).
func TestEinUnbekannterAnbieterWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitSchluesselbund(neuerBundspeicher())

	if gesetzt := anwendung.AnbietergeheimnisSetzen(
		t.Context(), "gibtsnicht", "x"); gesetzt.Fehler == nil {
		t.Error("ein unbekannter Anbieter wurde angenommen")
	}

	if geloescht := anwendung.AnbietergeheimnisLoeschen(
		t.Context(), "gibtsnicht"); geloescht.Fehler == nil {
		t.Error("ein unbekannter Anbieter wurde angenommen")
	}
}

// **Und wer eines verbietet, bekommt keines** — Microsoft weist einen
// öffentlichen Klienten mit Geheimnis ab, und ein Eintrag, der die Anmeldung
// scheitern lässt, ist schlimmer als keiner.
func TestEinVerbotenesAnbietergeheimnisWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitSchluesselbund(neuerBundspeicher())

	if gesetzt := anwendung.AnbietergeheimnisSetzen(
		t.Context(), "microsoft", "x"); gesetzt.Fehler == nil {
		t.Error("Microsoft nahm ein Geheimnis an")
	}
}

// **Jeder Anbieter hat einen Bauplan** — die Zusage, die `anbietergeheimnis`
// ohne Wache auskommen lässt.
//
// **Zwei Listen, eine Wahrheit**: `domain.Anbieter()` sagt, welche es gibt,
// `anbieterbau` in `anmeldenaht.go`, wie jeder gebaut wird. Laufen sie
// auseinander, greift dieser Lauf — und nicht ein `if`, das im Betrieb nie
// betreten wird.
func TestJederAnbieterHatEinenBauplan(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitSchluesselbund(neuerBundspeicher())

	antwort := anwendung.Anbietergeheimnisse(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Anbieter) != len(domain.Anbieter()) {
		t.Errorf("es kamen %d Einträge für %d Anbieter",
			len(antwort.Anbieter), len(domain.Anbieter()))
	}
}

// **Und ein klemmender Schlüsselbund reißt die Frage ab, statt zu raten.**
//
// **Nicht „kein Geheimnis"**: eine Anwendung, die einen Lesefehler als
// „liegt nichts" ausgibt, schickt jemanden dazu, ein Geheimnis noch einmal
// einzutragen, das längst dort liegt (§8).
func TestEinKlemmenderBundReisstDieFrageAb(t *testing.T) {
	t.Parallel()

	bund := neuerBundspeicher()
	bund.fehler = errors.New("der Bund klemmt")

	anwendung := app.NeueApp("")
	anwendung.MitSchluesselbund(bund)

	if antwort := anwendung.Anbietergeheimnisse(t.Context()); antwort.Fehler == nil {
		t.Error("ein klemmender Bund gab eine Liste")
	}
}
