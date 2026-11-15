package app

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Vorschau, die eine Stiloption einsetzt, ohne sie zu schreiben** —
// #590.
//
// # Wofür
//
// AGENTS.md §6: „Eine Änderung an einem Feld ist in der Vorschau zu sehen, ohne
// dass jemand etwas bestätigt." Für ein **Kartenfeld** löst das
// [App.Probevorschau] seit #288
// ein. Für eine **Stiloption** galt es nur zur Hälfte: eine Auswahl und ein
// Kästchen schreiben beim Umstellen, ein **Textfeld** erst beim Verlassen — und
// Verlassen ist eine Bestätigung.
//
// Über die 376 Stilpakete des Packs sind **738 Stiloptionen** vom Typ `text`,
// die dritthäufigste Art (`docs/STILOPTIONEN.md`).
//
// # Warum über die Karte und nicht über das Set
//
// Weil MSE die Reihenfolge schon hat:
//
//	IndexMap<FieldP, ValueP>& Set::stylingDataFor(const StyleSheet& stylesheet) {
//	  return styling_data.get(stylesheet.name(), stylesheet.styling_fields);
//	}
//
// `src/data/set.cpp`, Zeilen 99–101, Fassung f9d9356 vom 2021-07-06. **Die
// Karte schlägt das Set, das Set schlägt die Vorgabe des Stils** — hier
// umgesetzt in `service/stilumgebung.go`. Eine Option auf der Probekarte zu
// überlagern ergibt dasselbe Bild, als hätte jemand sie setweit gesetzt, und es
// entsteht kein zweiter Weg in den Bestand.

// Stilprobevorschau zeichnet eine Karte mit einer eingesetzten Stiloption.
//
// **Sie schreibt nicht.** Ein Schreiben legte einen Schritt im Verlauf an; wer
// eine Zahl in ein Textfeld tippt, bekäme einen je Taste, und ⌘Z nähme Ziffern
// zurück statt Änderungen — dieselbe Begründung wie bei [App.Probevorschau].
//
// **Sie prüft nicht, ob der Stil die Option kennt.** [App.StiloptionSetzen] tut
// das, weil dort ein Wert in den Bestand geht, den ein Skript lesen soll. Hier
// geht nichts hinein: ein unbekannter Name ergäbe eine Zeile, die
// `kartenstilwerte` übergeht, und die Vorschau sähe aus wie ohne Änderung. Die
// Namen kommen ohnehin aus [App.Stiloptionen].
//
// **Und sie bringt eine Mehrfachwahl nicht in die Normalform.** `OptionSetzen`
// tut es beim Schreiben (`service/stiloptionen.go`, `gespeicherteForm`); für
// jede andere Art ist das die Kennzeichnung. Der Weg hier ist für das
// **Textfeld** gebaut — die einzige Art, die beim Tippen noch keinen Wert
// geschrieben hat.
func (a *App) Stilprobevorschau(
	ctx context.Context, karteID, name, wert string, vergroesserung float64,
) VorschauAntwort {
	ctx = a.imProfil(ctx)

	ctx, fertig := a.vorschaulauf.beginne(ctx)
	defer fertig()

	kennung, befund := kennungAus(karteID)
	if befund != nil {
		return VorschauAntwort{Fehler: befund}
	}

	if strings.TrimSpace(name) == "" {
		return VorschauAntwort{Fehler: Uebersetze(ctx, domain.NewFieldError(
			"name", domain.SchluesselValidierung,
			"Eine Stiloption ohne Namen gibt es nicht."))}
	}

	karte, err := a.kartenleser.KarteLesen(ctx, kennung)
	if err != nil {
		return VorschauAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return a.zeichneVorschau(ctx, mitStiloption(karte, name, wert), vergroesserung,
		a.planbauer, a.rasterer)
}

// mitStiloption gibt die Karte mit einer überlagerten Stiloption.
//
// **Eine Kopie und keine Änderung am Gelesenen** — dieselbe Regel wie bei
// [mitWert]: der Kartenleser darf einen Zeiger auf etwas Gemerktes geben, und
// wer darin schriebe, änderte den Bestand über einen Weg, den niemand als
// Schreiben erkennt.
//
// **Der Schalter wird mitgesetzt.** `kartenstilwerte` fragt `has styling`,
// statt zu schauen, ob Daten dastehen (`service/kartenstiloptionen.go`) — ohne
// ihn läse der Planbau die Zeile gar nicht.
func mitStiloption(karte *domain.CardRow, name, wert string) *domain.CardRow {
	kopie := *karte
	kopie.Values = make([]domain.Value, 0, len(karte.Values)+2)

	bisher := ""

	for _, vorhanden := range karte.Values {
		switch vorhanden.Field {
		case service.Stilwertefeld:
			bisher = vorhanden.Text

		case service.Stilschalterfeld:
			// **Fällt weg und kommt unten neu** — er steht danach auf `true`.

		default:
			kopie.Values = append(kopie.Values, vorhanden)
		}
	}

	kopie.Values = append(kopie.Values,
		domain.Value{Field: service.Stilschalterfeld, Text: "true"},
		domain.Value{Field: service.Stilwertefeld, Text: mitZeile(bisher, name, wert)})

	return &kopie
}

// mitZeile ersetzt die Zeile einer Option oder hängt sie an.
//
// **Die Reihenfolge bleibt.** Ein Stildatenblock ist das, was die Karte selbst
// trägt; wer ihn neu sortierte, zeigte in der Vorschau einen Stand, den der
// Bestand so nicht hat — und beim nächsten Schreiben stünde er drin.
//
// **Getrennt am ersten Doppelpunkt**, wie `kartenstilwerte` liest: `other
// options` führt Listen, und ein Wert darf selbst einen tragen.
func mitZeile(bisher, name, wert string) string {
	neue := name + ": " + wert

	if strings.TrimSpace(bisher) == "" {
		return neue
	}

	zeilen := strings.Split(bisher, "\n")
	gesetzt := false

	for stelle, zeile := range zeilen {
		vorhanden, _, getrennt := strings.Cut(zeile, ":")
		if !getrennt || strings.TrimSpace(vorhanden) != name {
			continue
		}

		zeilen[stelle] = neue
		gesetzt = true
	}

	if !gesetzt {
		zeilen = append(zeilen, neue)
	}

	return strings.Join(zeilen, "\n")
}
