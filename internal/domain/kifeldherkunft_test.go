package domain_test

// **Woher ein übernommener Feldwert kommt** — Werkstück A11 von Phase 5.3
// (#1450).
//
// # Sie heißt **Feld**herkunft, und das ist kein Beiwerk
//
// **`domain.KIHerkunft` gibt es schon** (A6): das ist die Kette, die sagt,
// woher ein **Modell** stammt — Gewichte, Lizenz, Prüfsummen. Hier geht es
// um etwas anderes: woher ein **Feldwert** stammt. Zwei Fragen, zwei Namen.
//
// # Warum das mehr ist als die Marke »ki«
//
// **Seit A7 trägt ein übernommenes Feld die Herkunft `ki`.** Das sagt, dass
// es nicht getippt wurde, und sonst nichts — nicht **wer** geantwortet hat,
// nicht **worauf**, nicht **wann**.
//
// Das Issue verlangt fünf Dinge, und zwar **bei der ersten Übernahme** und
// nicht später: „Provider/Modell/Aufgabe/Zeit/Inhaltsbezug". Dazu einen
// Satz, der eine Grenze zieht: „**Rohprompt standardmäßig aus.**"

import (
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine vollständige Herkunft geht durch.**
func TestEineVollstaendigeKIFeldherkunftGehtDurch(t *testing.T) {
	t.Parallel()

	if befunde := eineHerkunft().Validate(); len(befunde) > 0 {
		t.Fatalf("die Befunde sind %v", befunde)
	}
}

// **Und jedes fehlende Stück wird benannt.**
//
// **Einzeln und nicht als »unvollständig«**: wer sie später liest, soll
// sehen, welches der fünf Dinge fehlt.
func TestJedesFehlendeStueckDerHerkunftWirdBenannt(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(*domain.KIFeldherkunft){
		"feld":     func(h *domain.KIFeldherkunft) { h.Feld = " " },
		"anbieter": func(h *domain.KIFeldherkunft) { h.Anbieter = "" },
		"modell":   func(h *domain.KIFeldherkunft) { h.Modell = "" },
		"aufgabe":  func(h *domain.KIFeldherkunft) { h.Aufgabe = "  " },
		"auftrag":  func(h *domain.KIFeldherkunft) { h.Auftrag = "" },
	}

	for feld, verbiegen := range faelle {
		t.Run("ohne "+feld, func(t *testing.T) {
			t.Parallel()

			herkunft := eineHerkunft()
			verbiegen(&herkunft)

			befunde := herkunft.Validate()
			if len(befunde) != 1 || befunde[0].Field != feld {
				t.Fatalf("die Befunde sind %v", befunde)
			}
		})
	}
}

// **Ohne Zeitpunkt ist sie keine.**
//
// **Weil »wann« eines der fünf Dinge ist.** Eine Herkunft ohne Zeit sagt
// nicht, ob sie von heute oder von einem Modellstand vor einem Jahr stammt.
func TestOhneZeitpunktIstDieHerkunftKeine(t *testing.T) {
	t.Parallel()

	herkunft := eineHerkunft()
	herkunft.Zeitpunkt = time.Time{}

	befunde := herkunft.Validate()
	if len(befunde) != 1 || befunde[0].Field != "zeitpunkt" {
		t.Fatalf("die Befunde sind %v", befunde)
	}
}

// **Der Inhaltsbezug entsteht aus dem Wert und nicht aus dem Vertrauen.**
//
// # Wofür er da ist
//
// **Damit sich später sagen lässt, ob das Feld noch das trägt, was das
// Modell vorgeschlagen hat.** Wer den Wert danach von Hand ändert, darf das
// (das Issue sagt es ausdrücklich) — und dann soll die Herkunft nicht
// behaupten, der Inhalt stamme von der KI.
//
// **Ein Abdruck und keine Kopie**: der Wert selbst steht schon im Feld; ihn
// hier zweitzuschreiben hieße, ihn an zwei Stellen zu pflegen.
func TestDerInhaltsbezugEntstehtAusDemWert(t *testing.T) {
	t.Parallel()

	einer := domain.KIInhaltsbezug([]string{"Blitzschlag", "3 Schaden"})

	if einer == "" || len(einer) < 16 {
		t.Fatalf("der Bezug ist %q", einer)
	}

	// **Derselbe Wert, derselbe Bezug.**
	if domain.KIInhaltsbezug([]string{"Blitzschlag", "3 Schaden"}) != einer {
		t.Error("zweimal derselbe Wert gibt zwei Bezüge")
	}

	// **Ein anderer Wert, ein anderer Bezug** — und die Grenze zwischen den
	// Teilen zählt mit: `["ab","c"]` ist nicht `["a","bc"]`.
	for _, anderer := range [][]string{
		{"Blitzschlag", "4 Schaden"},
		{"Blitzschlag3 Schaden"},
		{},
	} {
		if domain.KIInhaltsbezug(anderer) == einer {
			t.Errorf("%v gibt denselben Bezug", anderer)
		}
	}
}

// **Und der Rohprompt steht nicht darin.**
//
// # Der Satz, für den dieser Lauf da ist
//
// Das Issue: „**Rohprompt standardmäßig aus.**" Die **Aufgabe** ist, was ein
// Mensch getippt hat; der **Rohprompt** ist, was daraus wurde — Anweisung,
// Schema, Feldliste, mitgeschickte Bilder. Das eine gehört in die Herkunft,
// das andere nicht.
//
// **Geprüft wird an der Form und nicht an der Absicht**: es gibt kein Feld
// dafür. Ein Wächter, der eine Absicht prüft, prüft nichts (§10).
func TestDieHerkunftHatKeinFeldFuerDenRohprompt(t *testing.T) {
	t.Parallel()

	for _, name := range domain.KIFeldherkunftsfelder() {
		if strings.Contains(strings.ToLower(name), "prompt") ||
			strings.Contains(strings.ToLower(name), "anweisung") {
			t.Errorf("die Herkunft trägt ein Feld %q", name)
		}
	}
}

// eineHerkunft baut eine vollständige.
func eineHerkunft() domain.KIFeldherkunft {
	return domain.KIFeldherkunft{
		Feld: "name", Anbieter: "ollama", Modell: "gemma4:8b",
		Aufgabe: "Schreibe den Regeltext.", Auftrag: "ki-7-1",
		Zeitpunkt: time.Date(2026, 9, 12, 9, 0, 0, 0, time.UTC),
		Bezug:     domain.KIInhaltsbezug([]string{"Blitzschlag"}),
	}
}
