package msepaket_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein Verweis ohne Schrägstrich meint das Paket der Datei, in der er steht** —
// nicht das Paket, in dem der Lauf begonnen hat
// (#241,
// #240).
//
// # Woran das aufgefallen ist
//
// #241 stand als „ein Verzeichnis hinter `include file:` wird nicht aufgelöst"
// im Bestand, mit dieser Meldung aus dem laufenden Fenster:
//
//	magic-blends.mse-include/new-blends:10: magic.mse-game/masks ist nicht
//	lesbar: /magic.mse-game/masks: die Datei ist in keinem Paket zu finden
//
// **Die Deutung war falsch.** Nachgesehen am 2026-08-16 in `M15-Magic-Pack`:
// `masks` ist kein Verzeichnis, sondern eine gewöhnliche Textdatei mit 44
// Zeilen — und sie liegt in `magic-blends.mse-include`, dem Paket der Datei
// `new-blends`, die sie einbindet. In `magic.mse-game` gibt es sie überhaupt
// nicht.
//
// Der Weg dorthin ist zweistufig: `magic.mse-game/script` bindet
// `/magic-blends.mse-include/new-blends` ein, und **dessen** Zeile 10 sagt
// `include file: masks`. Der Einbinder trug das eigene Paket als *einen* Namen
// über den ganzen Lauf — also den des Ausgangspakets. Damit suchte er `masks`
// dort, wo der Lauf angefangen hatte.
//
// # Warum das mehr ist als eine Datei
//
// Ausgezählt über alle Pakete von `M15-Magic-Pack`: **49** Verweise ohne
// Schrägstrich stehen in `magic-modules.mse-include`, weitere 15 in
// `magic.mse-game`. Jeder von ihnen, aus einem Stil heraus erreicht, suchte
// bisher im Stilpaket.
func TestEinRelativerVerweisMeintDasPaketSeinerDatei(t *testing.T) {
	// Die Gestalt von `magic.mse-game` → `magic-blends.mse-include`, auf das
	// Nötige gekürzt.
	quelle := dateien{
		"blends.mse-include/new-blends": "include file: masks\n",
		"blends.mse-include/masks":      "id_mask: da\n",

		// **Die Falle, die den Fehler unsichtbar machte.** Läge im Ausgangspaket
		// eine gleichnamige Datei, gäbe der alte Weg einen Baum zurück — den
		// falschen, und ohne Befund. Genau deshalb steht sie hier: der Test
		// prüft, *welche* Datei ankommt, nicht bloß *ob* eine ankommt.
		"spiel.mse-game/masks": "id_mask: falsch\n",
	}

	baum, err := msepaket.Zerteile("script", []byte("include file: /blends.mse-include/new-blends\n"))
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	einbinder := msepaket.NeuerEinbinder(quelle).ImPaket("spiel.mse-game")
	gebaut := einbinder.Einbinden(baum)

	if befunde := einbinder.Befunde(); len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(gebaut.Wurzel) != 1 {
		t.Fatalf("%d Knoten, erwartet war einer", len(gebaut.Wurzel))
	}

	if gebaut.Wurzel[0].Wert != "da" {
		t.Errorf("%q — gelesen wurde die Datei des falschen Pakets", gebaut.Wurzel[0].Wert)
	}
}

// **Die Heimat kehrt zurück**, wenn der eingebundene Baum abgearbeitet ist.
//
// Ohne das Zurücksetzen wäre der Fehler nur verschoben: was **nach** einem
// tiefen `include` steht, suchte danach im zuletzt besuchten Paket. So steht es
// in `magic.mse-game/script`, wo auf `/magic-blends.mse-include/new-blends`
// noch weitere Zeilen folgen.
func TestNachEinemTiefenIncludeGiltWiederDasEigenePaket(t *testing.T) {
	quelle := dateien{
		"blends.mse-include/new-blends": "include file: masks\n",
		"blends.mse-include/masks":      "tief: ja\n",
		"spiel.mse-game/danach":         "danach: ja\n",
	}

	baum, err := msepaket.Zerteile("script",
		[]byte("include file: /blends.mse-include/new-blends\ninclude file: danach\n"))
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	einbinder := msepaket.NeuerEinbinder(quelle).ImPaket("spiel.mse-game")
	gebaut := einbinder.Einbinden(baum)

	if befunde := einbinder.Befunde(); len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	schluessel := make([]string, 0, len(gebaut.Wurzel))
	for _, knoten := range gebaut.Wurzel {
		schluessel = append(schluessel, knoten.Schluessel)
	}

	if verbunden := strings.Join(schluessel, ","); verbunden != "tief,danach" {
		t.Errorf("%q — nach dem tiefen Include galt das falsche Paket", verbunden)
	}
}

// **Ein Verweis ohne Schrägstrich in einem fremden Paket braucht keinen
// Ausgangsnamen.**
//
// Die Gegenprobe zu [msepaket.Einbinder.ImPaket]: ein Einbinder ohne
// Paketbezug weist einen eigenen Verweis ab (`einbinden_test.go`). Sobald der
// Lauf aber in einem Paket **angekommen** ist, ist der Bezug da — er steht im
// Verweis, der dorthin geführt hat. Ohne diesen Satz bliebe der häufigste Fall
// von #240 offen: eine Vorlagenliste, die ein Stilpaket ohne eigenen Namen
// einbindet.
func TestEinTiefesPaketGibtDenBezugAuchOhneAusgangsnamen(t *testing.T) {
	quelle := dateien{
		"blends.mse-include/new-blends": "include file: masks\n",
		"blends.mse-include/masks":      "tief: ja\n",
	}

	baum, err := msepaket.Zerteile("style",
		[]byte("include file: /blends.mse-include/new-blends\n"))
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	einbinder := msepaket.NeuerEinbinder(quelle)
	gebaut := einbinder.Einbinden(baum)

	if befunde := einbinder.Befunde(); len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(gebaut.Wurzel) != 1 || gebaut.Wurzel[0].Schluessel != "tief" {
		t.Errorf("%+v", gebaut.Wurzel)
	}
}
