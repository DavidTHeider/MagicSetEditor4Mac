package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **`stylesheet:` trägt die Kurzform, das Paket den Namen mit Spiel davor.**
//
// # Der Beleg
//
// `magic-set-generator` schreibt `.mse-set`-Dateien und ist die einzige
// **belegte** Quelle dafür, was in einem echten Set steht — ein solches liegt
// hier nicht vor. Es schreibt in `src/mse.rs`:
//
//	("stylesheet", Data::from(if game == "magic" { "m15-altered" } else { "standard" })),
//
// und benennt den Stylingblock dazu `magic-m15-altered`; die Zusammensetzung
// steht eine Funktion weiter:
//
//	let prefixed_stylesheet = format!("{}-{}", mse_game, stylesheet.render());
//
// Das Paket heißt also `<spiel>-<stylesheet>.mse-style`, und der Wert im Set
// ist die **Kurzform**.
//
// # Warum das mehr als eine Namensfrage ist
//
// Bis zum 2026-08-16 suchte der Stilspeicher den Wert **roh**. Ein echtes Set
// mit `game: magic` und `stylesheet: m15-altered` fand damit kein Paket, und
// die Anwendung zeigte einen **Behelfsplan** — eine Fläche mit den rohen
// Feldwerten darauf. Aufgefallen ist das bei
// #292, als
// ein Prüfset zum ersten Mal die Kurzform trug.
func TestEinStylesheetInKurzformFindetSeinPaket(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{
		{ID: 7, Style: "schlicht", Game: "probe"},
	}}

	speicher := service.NeuerStilspeicher(quelle, saatwurzel())

	vorlage, befunde, da := speicher.StilFuerSet(t.Context(), 7)
	if !da {
		t.Fatalf("kein Stil gefunden, Befunde: %v", befunde)
	}

	if len(vorlage.Stil.Felder) == 0 {
		t.Error("der gefundene Stil führt keine Felder")
	}
}

// **Der volle Name findet sein Paket weiterhin.**
//
// # Warum beide Formen gelten
//
// Die Kurzform ist die belegte; ob **jedes** Werkzeug sie schreibt, ist damit
// nicht gesagt, und §8 macht fremde Daten zum Normalfall. Ein Set, das den
// vollen Paketnamen trägt, abzuweisen brächte niemandem etwas — es wäre eine
// Strenge ohne Gegenwert.
//
// Die zusammengesetzte Form wird **zuerst** gesucht: sie ist die belegte, und
// bei `game: probe` mit `stylesheet: probe-stil` wäre sonst nicht zu
// entscheiden, ob `probe-probe-stil` oder `probe-stil` gemeint ist.
func TestEinStylesheetMitVollemNamenFindetSeinPaketWeiterhin(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{
		{ID: 7, Style: "probe-schlicht", Game: "probe"},
	}}

	speicher := service.NeuerStilspeicher(quelle, saatwurzel())

	vorlage, befunde, da := speicher.StilFuerSet(t.Context(), 7)
	if !da {
		t.Fatalf("kein Stil gefunden, Befunde: %v", befunde)
	}

	if len(vorlage.Stil.Felder) == 0 {
		t.Error("der gefundene Stil führt keine Felder")
	}
}

// **Ein Set ohne Spielangabe nimmt den Wert, wie er dasteht.**
//
// Ohne Spiel gibt es keine Kurzform, die sich auflösen ließe. Ein
// vorangestellter Bindestrich wäre ein Name, den kein Paket trägt.
func TestOhneSpielBleibtDerStilnameWieErIst(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{
		{ID: 7, Style: "probe-schlicht"},
	}}

	speicher := service.NeuerStilspeicher(quelle, saatwurzel())

	vorlage, befunde, da := speicher.StilFuerSet(t.Context(), 7)
	if !da {
		t.Fatalf("kein Stil gefunden, Befunde: %v", befunde)
	}

	if len(vorlage.Stil.Felder) == 0 {
		t.Error("der gefundene Stil führt keine Felder")
	}
}
