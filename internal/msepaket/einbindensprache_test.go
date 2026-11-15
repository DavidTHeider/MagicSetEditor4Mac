package msepaket_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **`include localized file:` wird aufgelöst** —
// #1065.
//
// # Der Befund, der dazu geführt hat
//
// `magic.mse-game/game` bindet seine **334** Schlüsselwörter so ein. Der
// Einbinder kannte nur `include file:`, und damit stand `keyword:` in keinem
// Knoten des Kopfes: der ganze Weg aus #1060 und #1061 war gebaut, geprüft,
// bei 100 % Abdeckung — und mit dem echten Pack passierte nichts.
//
// Gemessen: eine Karte aus `White.mse-set` mit dem Regeltext „Flying" nach
// SVG gezeichnet enthielt keinen Erinnerungstext.
//
// # Was der Quelltext sagt — und was nicht
//
// **Nichts.** `mse2-quelle` f9d9356 vom 2021-07-06 kennt den Schlüssel nicht;
// `Reader::unknownKey` behandelt `include_file` und sonst keinen. Das Pack
// trägt `mse version: 2.5.8`. Die Bedeutung ist deshalb **erschlossen** — aus
// den Dateinamen daneben — und in `docs/FORMATE.md` als solche gekennzeichnet.
func TestEinSprachverweisWirdAufgeloest(t *testing.T) {
	t.Parallel()

	quelle := dateien{
		"a.mse-game/keywords":    "keyword:\n\tkeyword: Grundfassung\n",
		"a.mse-game/keywords_de": "keyword:\n\tkeyword: Deutsch\n",
	}

	baum, err := msepaket.Zerteile("game", []byte("include localized file: keywords\n"))
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	// **Ohne gewählte Sprache gilt die Grundfassung.**
	neu := msepaket.NeuerEinbinder(quelle).ImPaket("a.mse-game").Einbinden(baum)

	if gebaut := string(msepaket.Schreibe(neu)); !strings.Contains(gebaut, "Grundfassung") {
		t.Errorf("ohne Sprache kam %q", gebaut)
	}

	// **Mit Sprache die Sprachfassung.**
	deutsch := msepaket.NeuerEinbinder(quelle).ImPaket("a.mse-game").
		InSprache("de").Einbinden(baum)

	if gebaut := string(msepaket.Schreibe(deutsch)); !strings.Contains(gebaut, "Deutsch") {
		t.Errorf("mit »de« kam %q", gebaut)
	}
}

// **Eine fehlende Sprachfassung fällt auf die Grundfassung zurück.**
//
// Fremde Daten sind der Normalfall (§8): ein Pack, das `keywords_de` nicht
// mitbringt, soll seine Schlüsselwörter trotzdem haben — englisch statt gar
// nicht.
func TestEineFehlendeSprachfassungFaelltZurueck(t *testing.T) {
	t.Parallel()

	quelle := dateien{"a.mse-game/keywords": "keyword:\n\tkeyword: Grundfassung\n"}

	baum, err := msepaket.Zerteile("game", []byte("include localized file: keywords\n"))
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	einbinder := msepaket.NeuerEinbinder(quelle).ImPaket("a.mse-game").InSprache("gibtsnicht")
	neu := einbinder.Einbinden(baum)

	if gebaut := string(msepaket.Schreibe(neu)); !strings.Contains(gebaut, "Grundfassung") {
		t.Errorf("gebaut %q", gebaut)
	}

	if befunde := einbinder.Befunde(); len(befunde) != 0 {
		t.Errorf("der Rückfall gab Befunde: %v", befunde)
	}
}

// **Und eine Datei, die es gar nicht gibt, ist ein Befund.**
func TestEinSprachverweisInsLeereWirdGemeldet(t *testing.T) {
	t.Parallel()

	baum, err := msepaket.Zerteile("game", []byte("include localized file: gibtsnicht\n"))
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	einbinder := msepaket.NeuerEinbinder(dateien{}).ImPaket("a.mse-game")
	einbinder.Einbinden(baum)

	if len(einbinder.Befunde()) == 0 {
		t.Error("der Verweis ins Leere blieb still")
	}
}
