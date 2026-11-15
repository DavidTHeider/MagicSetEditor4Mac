// **Die Kartenvorschau** — vom Plan zum Bild und zurück ins Fenster.
//
// Aus `karteneditor.go` herausgelöst, als es über 400 Zeilen ging (§2). Der
// Schnitt folgt der Frage: dort steht, **wie ein Feld geschrieben wird**, hier,
// **wie die Karte danach aussieht**.

package app

import (
	"bytes"
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Vorschau rastert die Karte, so wie sie **jetzt** im Bestand steht.
//
// # Warum aus dem Bestand und nicht aus dem, was das Frontend schickt
//
// Sonst zeigte die Vorschau, was jemand getippt hat, und nicht, was gespeichert
// wurde. Bei einem gescheiterten Schreiben liefen beide auseinander, und die
// Vorschau behauptete einen Stand, den es nicht gibt (§7, Regel 3).
//
// # Der vorige Lauf wird abgebrochen
//
// **Die Sonderregel dieses Projekts** (§7): eine Kartenvorschau wird
// abgebrochen, nicht abgewartet. Wer während des Tippens vier Änderungen macht,
// stößt vier Läufe an; ohne Abbruch streiten sie um Rechenzeit, Paketzugriffe
// und Schriftdateien, und der teuerste Teil ist nicht einmal das Rastern,
// sondern das Auswerten der Feldskripte im [service.Planbauer].
//
// **Das ist nicht dasselbe wie „die alte Antwort verwerfen".** Das Frontend tut
// das über `useLaden`, und damit ist §7 Regel 2 erfüllt — eine Antwort gehört
// zu dem Stand, aus dem sie stammt. Die überholte **Arbeit** lief trotzdem
// weiter; gefunden vom Zweitblick über #274/#275/#277/#278
// (#294).
//
// **Hier und nicht im Rasterer**, denn nur hier umfasst der Abbruch den ganzen
// Lauf: Lesen, Planbauen, Rastern. `render.Vorschau` hat genau das nicht
// gekonnt und ist deshalb entfallen.
func (a *App) Vorschau(ctx context.Context, karteID string, vergroesserung float64) VorschauAntwort {
	ctx = a.imProfil(ctx)

	ctx, fertig := a.vorschaulauf.beginne(ctx)
	defer fertig()

	kennung, befund := kennungAus(karteID)
	if befund != nil {
		return VorschauAntwort{Fehler: befund}
	}

	karte, err := a.kartenleser.KarteLesen(ctx, kennung)
	if err != nil {
		return VorschauAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return a.zeichneVorschau(ctx, karte, vergroesserung, a.planbauer, a.rasterer)
}

// Probevorschau zeichnet die Karte, als stünde in einem Feld schon der neue
// Wert — **ohne ihn zu schreiben**
// (#288).
//
// # Wofür
//
// AGENTS.md §6 verlangt: „Eine Änderung an einem Feld ist in der Vorschau zu
// sehen, ohne dass jemand etwas bestätigt." Bis hierher schrieb der Editor beim
// **Verlassen** des Feldes — und Verlassen ist eine Bestätigung.
//
// # Warum nicht einfach bei jedem Tastendruck schreiben
//
// Ein Schreiben legt einen Schritt im Verlauf an. Wer einen Kartennamen tippt,
// bekäme zwanzig davon, und ⌘Z nähme Buchstaben zurück statt Änderungen. MSE2
// trennt beides ebenso: `TextValueEditor` ändert den Wert im Arbeitsspeicher
// und zeichnet, die `TextValueAction` für den Verlauf entsteht gröber
// (`src/gui/value/text.cpp`, Fassung f9d9356 vom 2021-07-06).
//
// **Der Abbruch gilt hier genauso** (§7): der Aufruf teilt sich den Lauf mit
// [App.Vorschau], und ein neuer Tastendruck bricht den vorigen ab.
func (a *App) Probevorschau(
	ctx context.Context, karteID, feld, wert string, vergroesserung float64,
) VorschauAntwort {
	ctx = a.imProfil(ctx)

	ctx, fertig := a.vorschaulauf.beginne(ctx)
	defer fertig()

	kennung, befund := kennungAus(karteID)
	if befund != nil {
		return VorschauAntwort{Fehler: befund}
	}

	karte, err := a.kartenleser.KarteLesen(ctx, kennung)
	if err != nil {
		return VorschauAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return a.zeichneVorschau(ctx, mitWert(karte, feld, wert), vergroesserung,
		a.planbauer, a.rasterer)
}

// mitWert gibt die Karte mit einem eingesetzten Feldwert.
//
// **Eine Kopie und keine Änderung am Gelesenen.** Der Kartenleser darf einen
// Zeiger auf etwas Gemerktes geben; wer darin schriebe, änderte den Bestand
// über einen Weg, den niemand als Schreiben erkennt.
//
// **Ein Feld, das die Karte nicht führt, kommt hinzu.** Eine Karte aus einem
// fremden Set trägt nicht jedes Feld des Spiels; wer dort zu tippen anfängt,
// soll es trotzdem sehen.
func mitWert(karte *domain.CardRow, feld, wert string) *domain.CardRow {
	kopie := *karte
	kopie.Values = make([]domain.Value, 0, len(karte.Values)+1)

	gesetzt := false

	for _, vorhanden := range karte.Values {
		if vorhanden.Field == feld {
			vorhanden.Text = wert
			gesetzt = true
		}

		kopie.Values = append(kopie.Values, vorhanden)
	}

	if !gesetzt {
		kopie.Values = append(kopie.Values, domain.Value{Field: feld, Text: wert})
	}

	return &kopie
}

// zeichneVorschau ist der gemeinsame Teil von [App.Vorschau],
// [App.Probevorschau] und [App.Vorlagenvorschau]: planen, rastern, packen.
//
// **Planbauer und Rasterer stehen im Aufruf und nicht am `App`**
// (#1164).
// Die Vorlagenvorschau bringt eigene mit — einen Planbauer über einem
// Stilspeicher, der eine ungespeicherte Datei sieht.
//
// Der erste Anlauf baute dafür ein zweites `App` mit den zwei Feldern. Es
// stürzte ab: `nennungen` greift nach dem Kartenleser, und der fehlte. Ein
// halbes `App` ist eine Attrappe, die beim nächsten Feld wieder bricht — die
// beiden Teile gehören in die Aufrufliste, wo man sie sieht.
func (a *App) zeichneVorschau(
	ctx context.Context, karte *domain.CardRow, vergroesserung float64,
	planbauer Planbauer, rasterer Rasterer,
) VorschauAntwort {
	plan, planbefunde, err := planbauer.Kartenplan(ctx, karte)
	if err != nil {
		return VorschauAntwort{Fehler: Uebersetze(ctx, err)}
	}

	// **Noch die Vorgabe.** Die Güte kommt als Wahl der Nutzenden über die
	// Flächen herein (§16); bis dahin rastert dieser Weg wie bisher.
	bild, befunde, err := rasterer.Zeichne(ctx, plan, vergroesserung,
		render.GueteMittel)
	if err != nil {
		return VorschauAntwort{Fehler: Uebersetze(ctx, err)}
	}

	var gepackt bytes.Buffer
	if err := a.packen(&gepackt, bild); err != nil {
		return VorschauAntwort{Fehler: Uebersetze(ctx, err)}
	}

	// **Die Befunde des Plans zuerst.** Sie sagen, *warum* eine Karte behilflich
	// aussieht; die des Rasterers sagen, was beim Zeichnen fehlte. Wer den Grund
	// hinter die Folge stellte, ließe die Ursache im Anhang stehen.
	meldungen := make([]string, 0, len(planbefunde)+len(befunde))
	for _, gefunden := range planbefunde {
		meldungen = append(meldungen, gefunden.String())
	}

	for _, gefunden := range befunde {
		meldungen = append(meldungen, gefunden.String())
	}

	nennungen, befund := a.nennungen(ctx, karte, plan)
	if befund != "" {
		meldungen = append(meldungen, befund)
	}

	return VorschauAntwort{Vorschau: &VorschauDTO{
		Stil:       stilDerKarte(karte),
		Gezeichnet: plan.Stil,
		Nennungen:  nennungen,
		Flaechen:   feldflaechen(plan, vergroesserung),
		Vorgaben:   vorgabenDTO(plan),
		// **Eine Adresse und keine Data-URI** (#1270).
		//
		// 182 KB PNG werden als Base64 zu 242 KB und wandern als
		// JSON-Zeichenkette durch die Wails-Naht: gemessen **554 ms**,
		// während der ganze Go-Weg 53 ms braucht. Über die Assetleitung, die
		// Wails ohnehin betreibt, holt der Browser dieselben Bytes wie jede
		// andere Datei — mit welcher Messung, steht in [vorschauweg].
		Bild:    a.bildfach.Hinterlege(gepackt.Bytes()),
		Breite:  bild.Bounds().Dx(),
		Hoehe:   bild.Bounds().Dy(),
		DPI:     plan.DPI,
		Befunde: meldungen,
	}}
}

// MitEditor setzt die Nähte des Karteneditors — für Tests und den
// Composition Root.
func (a *App) MitEditor(
	schreiber Kartenschreiber, leser Kartenleser, bauer Planbauer, rasterer Rasterer,
) {
	a.schreiber = schreiber
	a.kartenleser = leser
	a.planbauer = bauer
	a.rasterer = rasterer
}

// feldflaechen gibt die Rechtecke der Felder, die auf der Karte liegen.
//
// **Aus [render.Plan.Feldflaechen] und nicht aus den Ebenen.** Ein leeres
// Textfeld ergibt keine Ebene — es gibt nichts zu malen —, hat seinen Kasten
// aber trotzdem; wer die Ebenen läse, böte auf einer frisch angelegten Karte
// kein einziges Feld an
// (#562).
//
// **Und in der Reihenfolge des Plans**, also von hinten nach vorn gezeichnet.
// Das Frontend legt sie in derselben Ordnung übereinander; damit gewinnt beim
// Klick die **oberste** — dieselbe Antwort, die das Auge gibt.
func feldflaechen(plan *render.Plan, vergroesserung float64) []FeldflaecheDTO {
	flaechen := make([]FeldflaecheDTO, 0, len(plan.Feldflaechen))

	for i := range plan.Feldflaechen {
		// **Nur, was auf der Karte liegt.** Ein Feldstil ohne `left:` trägt
		// MSEs Sentinel 1.000.000; daraus wurde eine Schaltfläche außerhalb
		// jedes Fensters, die in der Tastaturreihenfolge trotzdem mitzählte.
		// Dieselbe Frage stellt MSE vor dem Bedienen eines Feldes
		// (`src/gui/control/card_editor.cpp`, Zeilen 155–157, Fassung f9d9356
		// vom 2021-07-06), und sie wird deshalb dort beantwortet, wo sie schon
		// steht — im Renderer.
		if !render.Sichtbar(plan.Feldflaechen[i].Rahmen) {
			continue
		}

		rahmen := plan.Feldflaechen[i].Rahmen
		flaechen = append(flaechen, FeldflaecheDTO{
			Feld:   plan.Feldflaechen[i].Feld,
			Links:  rahmen.Links * vergroesserung,
			Oben:   rahmen.Oben * vergroesserung,
			Breite: rahmen.Breite * vergroesserung,
			Hoehe:  rahmen.Hoehe * vergroesserung,
		})
	}

	return flaechen
}

// vorgabenDTO übersetzt die gerechneten Werte des Plans.
//
// **Leer und nicht nil.** Ein `nil` käme als `null` an, und die Oberfläche
// prüfte darauf; eine leere Liste sagt dasselbe und braucht keinen zweiten
// Zweig.
func vorgabenDTO(plan *render.Plan) []VorgabeDTO {
	vorgaben := make([]VorgabeDTO, 0, len(plan.Vorgaben))

	for i := range plan.Vorgaben {
		vorgaben = append(vorgaben, VorgabeDTO{
			Feld: plan.Vorgaben[i].Feld,
			Wert: plan.Vorgaben[i].Wert,
		})
	}

	return vorgaben
}

// stilDerKarte ist der Wert `stylesheet` einer Karte, so wie das Fenster ihn
// liest (`Karteneditor`, „der eigene Stil der Karte") — getrimmt.
func stilDerKarte(karte *domain.CardRow) string {
	for _, wert := range karte.Values {
		if wert.Field == "stylesheet" {
			return strings.TrimSpace(wert.Text)
		}
	}

	return ""
}
