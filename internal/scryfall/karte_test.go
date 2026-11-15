package scryfall_test

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// **Eine gewöhnliche Karte wird vollständig gelesen**
// (#1005).
//
// # Warum gegen eine eingecheckte Datei und nicht gegen den Dienst
//
// §11: hier liegt kein fremdes Material. Die Beispiele in `testdaten/` sind
// **selbst geschrieben** — dieselbe Gestalt, eigener Inhalt. Ein Lauf gegen
// den echten Dienst steht hinter einem Build-Tag und zählt nicht zur Abdeckung
// (§3, wie `fremdpakete`).
func TestEineGewoehnlicheKarteWirdGelesen(t *testing.T) {
	t.Parallel()

	karte, befunde, fehler := scryfall.LiesKarte(beispiel(t, "einfach.json"))
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(befunde) != 0 {
		t.Errorf("es gab Befunde: %v", befunde)
	}

	pruefe(t, map[string]string{
		"Name":       karte.Name,
		"Manakosten": karte.Manakosten,
		"Typzeile":   karte.Typzeile,
		"Seltenheit": karte.Seltenheit,
		"Set":        karte.Set,
		"Künstler":   karte.Kuenstler,
	}, map[string]string{
		"Name":       "Prüfstein",
		"Manakosten": "{2}{W}",
		"Typzeile":   "Kreatur — Prüfling",
		"Seltenheit": "rare",
		"Set":        "prb",
		"Künstler":   "Selbst Gemalt",
	})

	if karte.Bilder.Ausschnitt != "https://beispiel.invalid/ausschnitt.jpg" {
		t.Errorf("der Bildausschnitt fehlt: %q", karte.Bilder.Ausschnitt)
	}

	if len(karte.Seiten) != 0 {
		t.Errorf("eine einseitige Karte hat %d Seiten", len(karte.Seiten))
	}
}

// **Ein Pflichtfeld, das fehlt, bricht laut.**
//
// # Warum streng und nicht tolerant
//
// §16: „Pflichtfelder streng, Zusatzfelder tolerant. Ein neues Feld beim
// Anbieter bricht nichts; ein **verschwundenes Pflichtfeld** bricht laut."
//
// Ohne diese Strenge stünde nach einer Änderung bei Scryfall eine Karte ohne
// Namen im Bestand — und niemand wüsste, woher sie kommt.
func TestEinFehlendesPflichtfeldBrichtLaut(t *testing.T) {
	t.Parallel()

	fuer := map[string]string{
		"object": `{"id":"x","name":"y"}`,
		"id":     `{"object":"card","name":"y"}`,
		"name":   `{"object":"card","id":"x"}`,
	}

	for feld, roh := range fuer {
		t.Run("ohne "+feld, func(t *testing.T) {
			t.Parallel()

			_, _, fehler := scryfall.LiesKarte([]byte(roh))
			if fehler == nil {
				t.Fatalf("die Karte ohne %q ging durch", feld)
			}

			if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("Fehler %v ist keine Validierung", fehler)
			}

			// **Die Meldung nennt das Feld.** Ein „ungültige Antwort" ohne
			// Feldnamen schickt den Leser auf die Suche.
			if !strings.Contains(fehler.Error(), feld) {
				t.Errorf("die Meldung nennt %q nicht: %v", feld, fehler)
			}
		})
	}
}

// **Ein unbekanntes Feld bricht nichts.**
//
// Es wird **auch nicht gemeldet**: Scryfall darf Felder hinzufügen, und ein
// Befund je neuem Feld machte aus jeder Erweiterung dort eine Meldung hier.
func TestEinUnbekanntesFeldBrichtNichts(t *testing.T) {
	t.Parallel()

	karte, befunde, fehler := scryfall.LiesKarte(beispiel(t, "unbekanntes-feld.json"))
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if karte.Name != "Neuling" {
		t.Errorf("Name %q", karte.Name)
	}

	if len(befunde) != 0 {
		t.Errorf("ein unbekanntes Feld wurde gemeldet: %v", befunde)
	}
}

// **Ein unbekannter Aufzählungswert wird als Befund geführt, nicht
// verschluckt.**
//
// # Warum der Unterschied zum Feld
//
// Ein neues **Feld** ist eine Erweiterung; ein neuer **Aufzählungswert** ist
// eine Bedeutung, die dieses Programm nicht kennt. `layout: "eine_gestalt"`
// heißt: die Karte sieht anders aus, als wir annehmen — und wer das
// verschluckt, zeichnet sie falsch.
//
// **Er bricht trotzdem nicht.** Die Karte kommt an, mit einem benannten
// Befund daneben. Das ist dieselbe Haltung wie bei `msepaket` (§8): 300
// lesbare Karten und zwei Befunde, nicht null Karten.
func TestEinUnbekannterAufzaehlungswertWirdGemeldet(t *testing.T) {
	t.Parallel()

	karte, befunde, fehler := scryfall.LiesKarte(beispiel(t, "unbekannte-art.json"))
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if karte.Name != "Seltsam" {
		t.Errorf("Name %q — die Karte soll trotzdem ankommen", karte.Name)
	}

	if len(befunde) != 2 {
		t.Fatalf("%d Befunde, erwartet waren zwei (layout und rarity): %v", len(befunde), befunde)
	}

	gemeinsam := strings.Join(befundtexte(befunde), " | ")
	for _, erwartet := range []string{"layout", "rarity", "2027"} {
		if !strings.Contains(gemeinsam, erwartet) {
			t.Errorf("die Befunde nennen %q nicht: %s", erwartet, gemeinsam)
		}
	}
}

// **Eine mehrseitige Karte ist ein eigener Fall.**
//
// # Warum `card_faces` und `image_uris` unabhängig wahlfrei sind
//
// Eine gewöhnliche Karte hat `image_uris` und keine `card_faces`. Eine
// zweiseitige hat `card_faces` und **kein** `image_uris` auf der obersten
// Ebene — die Bilder hängen an den Seiten. Und eine geteilte Karte hat
// **beides**. Wer eines aus dem anderen schließt, liest ein Drittel der Karten
// falsch.
func TestEineMehrseitigeKarteIstEinEigenerFall(t *testing.T) {
	t.Parallel()

	karte, befunde, fehler := scryfall.LiesKarte(beispiel(t, "zweiseitig.json"))
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(befunde) != 0 {
		t.Errorf("es gab Befunde: %v", befunde)
	}

	if len(karte.Seiten) != 2 {
		t.Fatalf("%d Seiten, erwartet waren zwei", len(karte.Seiten))
	}

	if karte.Seiten[0].Name != "Vorderseite" || karte.Seiten[1].Name != "Rückseite" {
		t.Errorf("die Seiten heißen %q und %q", karte.Seiten[0].Name, karte.Seiten[1].Name)
	}

	// **Die Bilder hängen an den Seiten und nicht an der Karte.**
	if karte.Bilder.Normal != "" {
		t.Errorf("die zweiseitige Karte hat ein Bild auf der obersten Ebene: %q", karte.Bilder.Normal)
	}

	if karte.Seiten[0].Bilder.Normal != "https://beispiel.invalid/vorn.jpg" {
		t.Errorf("das Bild der Vorderseite fehlt: %q", karte.Seiten[0].Bilder.Normal)
	}

	// **Und die Rückseite hat keinen Ausschnitt** — das ist bei Scryfall
	// üblich und darf nichts brechen.
	if karte.Seiten[1].Bilder.Ausschnitt != "" {
		t.Errorf("die Rückseite hat unerwartet einen Ausschnitt: %q",
			karte.Seiten[1].Bilder.Ausschnitt)
	}
}

// **Etwas, das kein JSON ist, bricht mit einer brauchbaren Meldung.**
func TestKeinJSONBrichtMitEinerMeldung(t *testing.T) {
	t.Parallel()

	if _, _, fehler := scryfall.LiesKarte([]byte("das ist kein JSON")); fehler == nil {
		t.Error("etwas, das kein JSON ist, ging durch")
	}
}

// beispiel liest eine der eingecheckten, selbst geschriebenen Antworten.
func beispiel(t *testing.T, name string) []byte {
	t.Helper()

	roh, err := os.ReadFile(filepath.Join("testdaten", name)) // #nosec G304 -- ein fester Pfad im Repositorium
	if err != nil {
		t.Fatalf("%s lesen: %v", name, err)
	}

	return roh
}

// pruefe vergleicht zwei Zuordnungen Feld für Feld.
func pruefe(t *testing.T, gelesen, erwartet map[string]string) {
	t.Helper()

	for name, soll := range erwartet {
		if ist := gelesen[name]; ist != soll {
			t.Errorf("%s = %q, erwartet war %q", name, ist, soll)
		}
	}
}

// befundtexte gibt die Meldungen der Befunde.
func befundtexte(befunde []domain.Befund) []string {
	texte := make([]string, 0, len(befunde))
	for _, befund := range befunde {
		texte = append(texte, befund.String())
	}

	return texte
}

// schreibe schreibt in eine Antwort und meldet, wenn es nicht ging.
func schreibe(t *testing.T, ziel io.Writer, inhalt []byte) {
	t.Helper()

	if _, err := ziel.Write(inhalt); err != nil {
		t.Errorf("die Gegenstelle konnte nicht schreiben: %v", err)
	}
}

// **Die Orakelkennung kommt über die Naht** —
// #1375.
//
// Sie ist das, was zwei englische Namen derselben Karte verbindet: „Universes
// Beyond" und „Universes Within" teilen sich eine `oracle_id`, und CR 100.2a
// zählt sie deshalb zusammen. Ohne diesen Weg stünde sie in der Antwort und
// nirgends sonst — zwei grüne Seiten einer Naht belegen die Naht nicht.
func TestDieOrakelkennungKommtUeberDieNaht(t *testing.T) {
	t.Parallel()

	karte, _, err := scryfall.LiesKarte(beispiel(t, "einfach.json"))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if karte.Orakelkennung != "00000000-0000-4000-8000-0000000000a1" {
		t.Errorf("die Orakelkennung ist %q", karte.Orakelkennung)
	}
}

// **Und ihr Fehlen ist kein Fehler** (§16: Zusatzfelder tolerant).
//
// Eine Antwort ohne `oracle_id` deshalb abzuweisen, ließe eine Deckprüfung an
// einer Kennung scheitern, die niemand braucht, solange der Name da ist.
func TestEineKarteOhneOrakelkennungIstKeinFehler(t *testing.T) {
	t.Parallel()

	karte, _, err := scryfall.LiesKarte([]byte(
		`{"object":"card","id":"1","name":"Ohne Orakel"}`))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if karte.Orakelkennung != "" {
		t.Errorf("die Orakelkennung ist %q, erwartet war leer", karte.Orakelkennung)
	}
}

// **Farbe und Farbidentität werden getrennt gelesen.**
//
// **Der Fall, für den es diesen Lauf gibt**, ist der dritte: eine farblose
// Karte mit einem Manasymbol im Regeltext. Sie ist rot-identisch und nicht
// rot — wer beides aus einem Feld läse, machte jede Kommandeur-Frage falsch
// (Regel 903.4, A15).
func TestFarbeUndFarbidentitaetWerdenGetrenntGelesen(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		roh                    string
		farben, farbidentitaet []string
	}{
		"eine rote Karte": {
			`"colors":["R"],"color_identity":["R"],`,
			[]string{"R"},
			[]string{"R"},
		},
		"farblos und rot-identisch": {
			`"colors":[],"color_identity":["R"],`,
			[]string{},
			[]string{"R"},
		},
		"gar keine Angabe": {"", nil, nil},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			karte, _, fehler := scryfall.LiesKarte([]byte(`{
				"object":"card","id":"a","name":"Probe","lang":"en",
				"layout":"normal","type_line":"Instant","cmc":1,
				"set":"tst","set_name":"Test","collector_number":"1",
				"rarity":"common",` + fall.roh + `
				"scryfall_uri":"https://beispiel.invalid/a"}`))
			if fehler != nil {
				t.Fatalf("LiesKarte: %v", fehler)
			}

			if !slices.Equal(karte.Farben, fall.farben) ||
				!slices.Equal(karte.Farbidentitaet, fall.farbidentitaet) {
				t.Fatalf("gelesen wurden %v / %v, erwartet %v / %v",
					karte.Farben, karte.Farbidentitaet,
					fall.farben, fall.farbidentitaet)
			}
		})
	}
}
