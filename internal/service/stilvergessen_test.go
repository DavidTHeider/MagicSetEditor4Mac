package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Stilwechsel macht das Gedächtnis des Stilspeichers ungültig**
// (#693).
//
// # Was vorher geschah
//
// Der Speicher merkt sich die Vorlage **je Set** und räumte sie nie. Der
// Bestand nahm einen Wechsel an, und die Vorschau zeigte bis zum Neustart
// dieselbe Karte — Byte für Byte über alle vier Stile der Saat.
//
// Gefunden hat es der Abnahmeweg „ein Set umgestalten"
// (#680) —
// keine Bausteinebene für sich.
//
// # Warum die Zusage hier steht
//
// Sie ist die des **Wählers**: wer wechselt, sorgt dafür. Läge sie beim
// Aufrufer, hinge sie an jedem einzelnen — der MCP-Server oder ein Test
// vergäße sie.
func TestEinStilwechselVergisstDieGemerkteVorlage(t *testing.T) {
	t.Parallel()

	gemeldet := []int64{}

	waehle := waehler(t, domain.SetHead{ID: 1, Game: "probe", Style: "probe-bild"},
		&stilablage{}, &feldliste{})

	waehle.MitVergessen(func(setID int64) {
		gemeldet = append(gemeldet, setID)
	})

	if err := waehle.Wechseln(t.Context(), 1, "probe-schlicht"); err != nil {
		t.Fatalf("wechseln: %v", err)
	}

	if len(gemeldet) != 1 || gemeldet[0] != 1 {
		t.Errorf("gemeldet wurde %v, erwartet [1]", gemeldet)
	}
}

// **Ein gescheiterter Wechsel vergisst nichts** (§7, Regel 3).
//
// Was gemerkt ist, gilt so lange, wie der Bestand es sagt. Ein Vergessen vor
// einem gescheiterten Schreiben ließe den nächsten Aufruf neu lesen —
// richtig, aber unbegründet.
func TestEinGescheiterterWechselVergisstNichts(t *testing.T) {
	t.Parallel()

	gemeldet := 0

	waehle := waehler(t, domain.SetHead{ID: 1, Game: "probe", Style: "probe-bild"},
		&stilablage{}, &feldliste{})

	waehle.MitVergessen(func(int64) { gemeldet++ })

	if err := waehle.Wechseln(t.Context(), 1, "niegesehen"); err == nil {
		t.Fatal("erwartet wurde ein Fehler, es kam keiner")
	}

	if gemeldet != 0 {
		t.Errorf("es wurde %d-mal vergessen, erwartet keinmal", gemeldet)
	}
}

// **Und mit einem echten Speicher: der zweite Blick sieht den neuen Stil.**
//
// Die beiden Fälle darüber prüfen, **ob** gemeldet wird. Dieser prüft, dass
// das Melden auch wirkt — die eigentliche Zusage, und die einzige, die jemand
// am Bildschirm bemerkt.
func TestNachDemWechselStehtDerNeueStilDa(t *testing.T) {
	t.Parallel()

	koepfe := &koepfequelle{koepfe: []domain.SetHead{
		{ID: 1, Game: "probe", Style: "probe-schlicht"},
	}}

	speicher := service.NeuerStilspeicher(koepfe, festeSaatwurzel(t))

	vorlage, _, da := speicher.StilFuerSet(t.Context(), 1)
	if !da {
		t.Fatal("der erste Blick findet keinen Stil")
	}

	waehle := service.NeuerStilwaehler(
		koepfe, saatkatalog(), &stilablage{}, &feldliste{}, festeSaatwurzel(t),
	).MitSpeicher(speicher)

	// **Der Kopf ändert sich mit** — die Ablage ist ein Doppel, und ohne das
	// wäre der zweite Blick derselbe, ganz ohne Gedächtnis.
	koepfe.koepfe[0].Style = "probe-bild"

	if err := waehle.Wechseln(t.Context(), 1, "probe-bild"); err != nil {
		t.Fatalf("wechseln: %v", err)
	}

	nachher, _, da := speicher.StilFuerSet(t.Context(), 1)
	if !da {
		t.Fatal("der zweite Blick findet keinen Stil")
	}

	if nachher.Stil.Paket == vorlage.Stil.Paket {
		t.Errorf("nach dem Wechsel steht weiter %q da", nachher.Stil.Paket)
	}
}

// **Und ein gescheitertes Schreiben kommt als Fehler zurück** — nicht als
// stiller Erfolg mit geräumtem Gedächtnis.
func TestEinGescheitertesSchreibenMeldetSich(t *testing.T) {
	t.Parallel()

	wund := errors.New("die Platte ist voll")
	gemeldet := 0

	waehle := waehler(t, domain.SetHead{ID: 1, Game: "probe", Style: "probe-bild"},
		&stilablage{setzfehler: wund}, &feldliste{})

	waehle.MitVergessen(func(int64) { gemeldet++ })

	if err := waehle.Wechseln(t.Context(), 1, "probe-schlicht"); !errors.Is(err, wund) {
		t.Errorf("Wechseln = %v, erwartet %v", err, wund)
	}

	if gemeldet != 0 {
		t.Errorf("es wurde %d-mal vergessen, erwartet keinmal", gemeldet)
	}
}
