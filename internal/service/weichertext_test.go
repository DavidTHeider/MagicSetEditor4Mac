package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **`<sep>` kennt MSE gar nicht — und übergeht es deshalb.**
//
// Die Liste der gedeuteten Auszeichnungen in `TextElements::fromString`
// (`src/render/text/element.cpp`, Zeilen 78–130, Fassung f9d9356 vom
// 2021-07-06) führt `b`, `i`, `sym`, `line`, `soft-line`, `sep-soft`, `soft`,
// `li`, `atom-kwpph`, `code…`, `color`, `font`, `size`, `margin`, `align` —
// **`sep` steht nicht darin**. Eine unbekannte Auszeichnung wird übersprungen,
// ihr Inhalt bleibt Text.
//
// `combined_editor` klammert damit den Trenner zwischen Regel- und Flavortext:
// `<sep><line>\n</line></sep>`. Dieses Programm tat schon dasselbe — und
// **meldete es**, 367-mal über die 190 Stile von `M15-Magic-Pack`. Ein Befund
// für richtiges Verhalten verdeckt die echten
// (#345).
func TestSepIstKeinBefund(t *testing.T) {
	_, befunde := gesetzterText(t, "vorn<sep><line>\n</line></sep>hinten")

	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "sep") {
			t.Errorf("`<sep>` wird gemeldet: %s", befund.Meldung)
		}
	}
}

// **`<soft>` zählt bei der Ausrichtung nicht mit.**
//
//	bool soft : 1; ///< Is this a 'soft' character? soft characters are
//	               ///  ignored for alignment
//
// `CharInfo`, `src/render/text/element.hpp`, Zeilen 36–47. Die Breite einer
// Zeile endet am letzten **nicht weichen** Zeichen:
//
//	double width() const { return positions[end_or_soft-start] - positions.front(); }
//
// `src/render/text/viewer.cpp`, Zeilen 17–35 und 610.
//
// **Wozu:** `sub_type_filter` hängt an die Typzeile ein
// `"<soft>" + subtype_separator + "</soft>"` — einen Trenner für den Fall, dass
// noch etwas folgt. Zählte er mit, stünde jede mittig gesetzte Typzeile um
// seine halbe Breite zu weit links.
//
// Fassung f9d9356 vom 2021-07-06.
func TestWeicherTextWirdAlsWeichGefuehrt(t *testing.T) {
	ebenen, befunde := gesetzterText(t, "AB<soft>CD</soft>EF")

	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "soft") {
			t.Errorf("`<soft>` wird gemeldet: %s", befund.Meldung)
		}
	}

	if len(ebenen) < 2 {
		t.Fatalf("%d Ebenen", len(ebenen))
	}

	weich := map[string]bool{}
	for _, abschnitt := range ebenen[1].Abschnitte {
		weich[abschnitt.Text] = abschnitt.Weich
	}

	if !weich["CD"] {
		t.Errorf("der weiche Teil fehlt: %+v", ebenen[1].Abschnitte)
	}

	if weich["AB"] || weich["EF"] {
		t.Errorf("harter Text steht als weich: %+v", ebenen[1].Abschnitte)
	}
}

// gesetzterText baut einen Plan aus einem einzigen, mittig gesetzten Textfeld.
func gesetzterText(t *testing.T, text string) ([]render.Ebene, []domain.Befund) {
	t.Helper()

	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 200, Hoehe: 60,
			Felder: []msepaket.Feldstil{{
				Feld: "name", Links: "0", Oben: "0", Breite: "200", Hoehe: "40",
				Ausrichtung: "middle center",
				Schrift:     msepaket.Schrift{Name: "Helvetica", Groesse: "10"},
			}},
		},
		Kartenfelder: []domain.Field{{Name: "name", Kind: domain.KindText}},
	}

	karte := &domain.CardRow{ID: 1, Values: []domain.Value{{Field: "name", Text: text}}}

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	return plan.Ebenen, befunde
}

// **`<soft-line>` ist kein weicher Text, sondern ein weicher Umbruch.**
//
// MSE prüft `<soft-line` **vor** `<soft`, weil `is_tag` auf einem Präfix
// arbeitet (`src/render/text/element.cpp`, Zeilen 86–92, Fassung f9d9356 vom
// 2021-07-06). Wer die Reihenfolge umdrehte, machte aus jedem weichen Umbruch
// eine weiche Stelle — und die zählte bei der Ausrichtung nicht mehr mit.
func TestSoftLineIstKeinWeicherText(t *testing.T) {
	ebenen, _ := gesetzterText(t, "AB<soft-line>CD</soft-line>")

	if len(ebenen) < 2 {
		t.Fatalf("%d Ebenen", len(ebenen))
	}

	for _, abschnitt := range ebenen[1].Abschnitte {
		if abschnitt.Weich {
			t.Errorf("`<soft-line>` macht weichen Text: %+v", ebenen[1].Abschnitte)
		}
	}
}

// **`<sep-soft>` ist beides — übergangen wird es nicht.**
//
// Es beginnt mit `sep`, ist aber in MSEs Liste **vor** `sep` geprüft und zählt
// als weich (Zeilen 88–89).
func TestSepSoftIstWeich(t *testing.T) {
	ebenen, _ := gesetzterText(t, "AB<sep-soft>CD</sep-soft>")

	if len(ebenen) < 2 {
		t.Fatalf("%d Ebenen", len(ebenen))
	}

	gefunden := false

	for _, abschnitt := range ebenen[1].Abschnitte {
		if abschnitt.Text == "CD" && abschnitt.Weich {
			gefunden = true
		}
	}

	if !gefunden {
		t.Errorf("`<sep-soft>` macht keinen weichen Text: %+v", ebenen[1].Abschnitte)
	}
}

// **`<atom-kwpph>` ist *nicht* weich — es wird gezeichnet.**
//
// # Der Fehler, den dieser Test festhält
//
// Hier stand bis zum 2026-08-21 „`<atom-kwpph>` zählt bei MSE ebenso als weich
// (Zeile 299)". Die genannte Zeile ist die **Schriftfarbe**:
//
//	(soft        > 0 ? FONT_SOFT        : FONT_NORMAL) |
//	(kwpph       > 0 ? FONT_SOFT        : FONT_NORMAL) |
//
// `src/render/text/element.cpp`, Zeilen 298–299, Fassung f9d9356 vom
// 2021-07-06. Weich im Sinne der **Ausrichtung** und des **Zeichnens** wird ein
// Stück fünf Zeilen darüber, und dort steht `kwpph` nicht:
//
//	DrawWhat what = soft > 0 ? DRAW_ACTIVE : DRAW_NORMAL;
//
// Zeile 234. Ein Platzhalter wird also **gezeichnet** — in Trennfarbe, ohne
// Schatten — und er zählt bei der Ausrichtung mit.
//
// # Warum das erst jetzt auffiel
//
// Weil es folgenlos war, solange „weich" nur die Ausrichtung betraf. Seit
// #725 heißt
// weich zusätzlich „wird nicht gezeichnet", und damit wäre der Platzhalter
// unsichtbar geworden.
//
// # Die Reichweite dieser Messung (§10)
//
// **In den zehn kanonischen Sets steht `<atom-kwpph>` null Mal**, und im ganzen
// `M15-Magic-Pack` ebenso wenig — er entsteht erst beim Aufschlüsseln eines
// Schlüsselworts mit leerem Parameter (`src/data/keyword.cpp`, Zeile 638). Was
// dieser Test verhindert, ist deshalb kein heutiger Ausfall, sondern ein
// stiller, sobald die Schlüsselwörter dazukommen.
func TestEinSchluesselwortplatzhalterIstNichtWeich(t *testing.T) {
	ebenen, _ := gesetzterText(t, "Verzaubere <atom-kwpph>Bleibendes</atom-kwpph>")

	if len(ebenen) < 2 {
		t.Fatalf("%d Ebenen", len(ebenen))
	}

	for _, abschnitt := range ebenen[1].Abschnitte {
		if strings.Contains(abschnitt.Text, "Bleibendes") && abschnitt.Weich {
			t.Error("der Platzhalter ist weich und würde nicht gezeichnet")
		}
	}
}
