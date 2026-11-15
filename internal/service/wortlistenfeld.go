// Die Wortlisten eines Feldes
// (#1074).
//
// # Die Naht
//
// Drei Dinge treffen sich hier, und keines allein genügt:
//
//   - der **Wert** sagt, welche Stelle zu welcher Liste gehört —
//     `<word-list-class-en>Bear</word-list-class-en>`;
//   - das **Spielpaket** sagt, was in der Liste steht;
//   - die **Karte** sagt, was ein `script:` darin ausrechnet.
//
// Ohne das Dritte gäbe jede skriptgebaute Liste dasselbe, nämlich nichts —
// und im Pack haben 77 der 432 Wörter ein Skript.

package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Wortlistenbereich ist eine Stelle eines Feldwerts mit ihrem Menü.
type Wortlistenbereich struct {
	// Liste ist der Name der Wortliste — `type-en`, `class-en`, `land`.
	Liste string

	// Text ist, was an dieser Stelle steht, ohne Auszeichnung.
	Text string

	// Textanfang und Textende spannen den Inhalt im **rohen** Wert auf — das,
	// was eine Wahl ersetzt.
	//
	// **Sie stehen hier, damit [Planbauer.Wortwahl] nicht ein zweites Mal
	// suchen muss.** Der erste Anlauf tat das und zählte die zugeordneten
	// Stellen dabei noch einmal nach; der Zweig für »nicht mehr da« war von
	// keinem Aufruf zu betreten, denn zwischen Suchen und Schreiben liegt
	// nichts (§3).
	Textanfang int
	Textende   int

	Menue Wortmenue
}

// Wortlistenfeld gibt die Wortlistenbereiche eines Kartenfeldes.
//
// **Ohne Marken gibt es nichts, und keinen Befund.** Die allermeisten Felder
// haben keine Wortliste; eine Meldung je Feld ohne eine wäre eine Liste, die
// niemand liest (§8) — dieselbe Überlegung wie bei [Planbauer.Symbolpalette].
//
// **Eine Marke ohne Liste ist ein Befund und kein Abbruch**:
//
//	queue_message(MESSAGE_WARNING, _ERROR_1_("word list type not found", name));
//
// `TextValueEditor::findWordLists`, `src/gui/value/text.cpp`, Zeile 1372,
// Fassung f9d9356 vom 2021-07-06.
func (p *Planbauer) Wortlistenfeld(
	ctx context.Context, karte *domain.CardRow, feld string, vorlage Vorlage,
) ([]Wortlistenbereich, []domain.Befund, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	// **Die Umgebung zuerst, und das ist der ganze Punkt.**
	//
	// Die Marken stehen **nicht** im gespeicherten Wert, sondern im
	// **gerechneten**: `magic.mse-game/card_fields` gibt dem Feld `super type`
	// ein `script: super_type_filter(value, …)`, und erst dieses Skript legt
	// `<word-list-type-en>…</word-list-type-en>` um den Text.
	//
	// Nachgemessen am 2026-08-24 an `Vanguard.mse-set`: gespeichert steht dort
	// `Vanguard`, in der Umgebung `<word-list-type-en>Vanguard</…>` — und
	// **keines** der zehn kanonischen Sets führt die Marken in der Datei. Wer
	// den gespeicherten Wert nähme, hätte eine Auswahl, die auf keiner echten
	// Karte erscheint.
	//
	// **Ihre Befunde bleiben hier.** Sie gehören zur Karte und stehen an der
	// Vorschau; ein zweites Mal an jedem Feld gezeigt wären es dieselben
	// Meldungen sechzigmal.
	umgebung, gerechnet, _ := umgebungAus(ctx, karte, vorlage, p.schriften, p.bilder)

	roh, _ := gerechneterWert(gerechnet, feld)

	stellen := Wortlistenstellen(roh)
	if len(stellen) == 0 {
		return nil, nil, nil
	}

	var befunde []domain.Befund

	bereiche := make([]Wortlistenbereich, 0, len(stellen))

	for _, stelle := range stellen {
		liste, da := wortlisteMitNamen(vorlage.Wortlisten, stelle.Liste)
		if !da {
			befunde = append(befunde, domain.Befund{
				Ort:     feld,
				Meldung: "das Spiel kennt die Wortliste " + stelle.Liste + " nicht",
			})

			continue
		}

		menue := NeuesWortmenue(liste.Woerter, stelle.Text,
			skriptauswerter(ctx, umgebung, feld, &befunde))

		bereiche = append(bereiche, Wortlistenbereich{
			Liste:      stelle.Liste,
			Text:       stelle.Text,
			Textanfang: stelle.Textanfang,
			Textende:   stelle.Textende,
			Menue:      menue,
		})
	}

	return bereiche, befunde, nil
}

// Wortwahl gibt den **rohen** Feldwert nach einer Wahl.
//
// Der Weg zählt die Einträge von oben; der Bereich sagt, die wievielte
// Wortlistenstelle des Feldes gemeint ist. Beide kommen von außen und werden
// geprüft — ein Menü, das inzwischen ein anderes ist, darf keinen Wert
// verstümmeln.
//
// **Nur der Inhalt wird ersetzt, nicht die Marken.** So tut MSE es:
// `fixSelection` rückt die Auswahl über `index_to_cursor`/`cursor_to_index` auf
// die nächste Textstelle, womit `<word-list-…>` und `</word-list-…>` außen vor
// bleiben (`TextValueEditor::fixSelection`, `src/gui/value/text.cpp`, Zeilen
// 1104–1112, Fassung f9d9356 vom 2021-07-06). Ein Wert ohne seine Marken
// verlöre die Auswahl beim nächsten Öffnen.
func (p *Planbauer) Wortwahl(
	ctx context.Context, karte *domain.CardRow, feld string, vorlage Vorlage,
	bereich int, weg []int,
) (string, []domain.Befund, error) {
	bereiche, befunde, err := p.Wortlistenfeld(ctx, karte, feld, vorlage)
	if err != nil {
		return "", befunde, err
	}

	_, gerechnet, _ := umgebungAus(ctx, karte, vorlage, p.schriften, p.bilder)

	if bereich < 0 || bereich >= len(bereiche) {
		return "", befunde, fmt.Errorf("%w: das Feld %s hat keinen %d. Wortlistenbereich",
			domain.ErrValidation, feld, bereich+1)
	}

	neu, gewaehlt := bereiche[bereich].Menue.WaehlenAmWeg(weg)
	if !gewaehlt {
		return "", befunde, fmt.Errorf(
			"%w: der Weg %v führt in der Wortliste %s zu keinem Eintrag",
			domain.ErrValidation, weg, bereiche[bereich].Liste)
	}

	// **Die Grenzen kommen aus dem Bereich**, nicht aus einer zweiten Suche:
	// er ist aus **diesem** Wert gebaut, und dazwischen liegt kein Schreiben.
	//
	// **Und der Wert ist der gerechnete**, aus dem auch die Bereiche stammen.
	// Geschrieben wird er mitsamt seinen Marken — so tut MSE es: dort **ist**
	// der gerechnete Wert der Wert (`TextValue::value`), und `replaceSelection`
	// ändert genau ihn. Beim nächsten Lauf nimmt `super_type_filter` die Marken
	// mit `remove_tag(tag: "<word-list-")` wieder heraus und setzt sie neu; die
	// Runde ist damit geschlossen.
	roh, _ := gerechneterWert(gerechnet, feld)
	stelle := bereiche[bereich]

	return roh[:stelle.Textanfang] + geschuetzt(neu) + roh[stelle.Textende:], befunde, nil
}

// geschuetzt macht aus einem `<` im gewählten Wort MSEs `ESCAPED_LANGLE`.
//
//	tve.replaceSelection(escape(new_value), …)
//
// `DropDownWordList::select`, `src/gui/value/text.cpp`, Zeile 300, Fassung
// f9d9356 vom 2021-07-06. Ohne das begänne beim nächsten Lesen eine Marke, und
// der Wert bräche ab.
func geschuetzt(wort string) string {
	return strings.ReplaceAll(wort, "<", "\x01")
}

// skriptauswerter gibt den Auswerter, den ein Wort mit `script:` braucht.
//
// **Ein gescheitertes Skript ist ein Befund und kein Abbruch** (§8): das Wort
// bleibt aus, die übrigen stehen. Wer die ganze Liste verwürfe, nähme dem
// Bearbeitenden auch die Einträge, die dastehen.
func skriptauswerter(
	ctx context.Context, umgebung *msescript.Umgebung, feld string, befunde *[]domain.Befund,
) func(string) string {
	return func(quelle string) string {
		wert, fehler := msescript.AuswertenQuelle(ctx, quelle, umgebung)
		if fehler != nil {
			*befunde = append(*befunde, domain.Befund{
				Ort:     feld,
				Meldung: "die Wortliste rechnet " + quelle + " nicht aus: " + fehler.Error(),
			})

			return ""
		}

		return wert.AlsText()
	}
}

// wortlisteMitNamen sucht eine Liste des Spiels.
func wortlisteMitNamen(listen []msepaket.Wortliste, name string) (msepaket.Wortliste, bool) {
	for _, liste := range listen {
		if liste.Name == name {
			return liste, true
		}
	}

	return msepaket.Wortliste{}, false
}

// Kartenwortlisten sucht den Stil der Karte und gibt die Bereiche eines Feldes.
//
// **Dieselbe Auflösung wie [Planbauer.Kartenpalette]**, und aus demselben Grund
// getrennt: die Vorschau, die Palette und die Wortliste sprechen über dieselbe
// Karte, und drei Wege zum Stil liefen beim ersten Sonderfall auseinander.
//
// **Ohne Behelf.** Die Listen stehen im **Spielpaket**, und ohne Stil ist auch
// das Spiel nicht bestimmt; ein Behelfsplan zeichnet Feldwerte als Zeilen und
// kennt keine Wortlisten.
func (p *Planbauer) Kartenwortlisten(
	ctx context.Context, karte *domain.CardRow, feld string,
) ([]Wortlistenbereich, []domain.Befund, error) {
	vorlage, befunde, da := p.kartenvorlage(ctx, karte)
	if !da {
		return nil, befunde, nil
	}

	bereiche, weitere, err := p.Wortlistenfeld(ctx, karte, feld, vorlage)

	return bereiche, append(befunde, weitere...), err
}

// Kartenwortwahl gibt den neuen rohen Feldwert nach einer Wahl.
func (p *Planbauer) Kartenwortwahl(
	ctx context.Context, karte *domain.CardRow, feld string, bereich int, weg []int,
) (string, []domain.Befund, error) {
	vorlage, befunde, _ := p.kartenvorlage(ctx, karte)

	// **Ohne Stil wird trotzdem gefragt**, und die Antwort ist dann »diesen
	// Bereich gibt es nicht«. Das ist dieselbe Meldung wie bei einem Klick auf
	// ein Menü, das inzwischen ein anderes ist — und sie ist richtig: ohne
	// Spielpaket ist keine Liste zugeordnet.
	neu, weitere, err := p.Wortwahl(ctx, karte, feld, vorlage, bereich, weg)

	return neu, append(befunde, weitere...), err
}

// kartenvorlage gibt den Stil der Karte — leer, wenn es keinen gibt.
func (p *Planbauer) kartenvorlage(ctx context.Context, karte *domain.CardRow) (
	Vorlage, []domain.Befund, bool,
) {
	if p.stile == nil {
		return Vorlage{}, nil, false
	}

	return p.stile.StilFuerKarte(ctx, karte)
}
