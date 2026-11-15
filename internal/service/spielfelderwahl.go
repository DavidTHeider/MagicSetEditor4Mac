package service

// **Welche Felder ein Spiel erklärt** — für Karten und für Sets.
//
// Eine eigene Datei, weil `stilwahl.go` mit ihr über 400 Zeilen ginge (§2) —
// und weil der Schnitt einer Frage folgt: dort steht, **welcher Stil** gilt,
// hier, **welche Felder** es gibt.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Spielskript gibt das `init script:` des Spiels, zu dem ein Set gehört —
// #1238.
//
// **Für die Kartenliste**, und dort einmal je Aufbau: die Feldskripte des
// Spiels rufen Funktionen, die es selbst anlegt (`card_full_name()`,
// `cmc(…)`). Ohne sie bleiben genau die Spalten leer, die man wählen möchte.
//
// **Dieselbe Quelle wie [Stilwaehler.Kartenfelder]** und derselbe
// Zwischenspeicher — das Spielmodell trägt sein Skript seit dem 2026-08-28
// mit, statt es ein zweites Mal von der Platte zu lesen.
func (w *Stilwaehler) Spielskript(ctx context.Context, setID int64) (string, error) {
	kopf, err := w.kopf(ctx, setID)
	if err != nil {
		return "", err
	}

	modell, err := w.spielmodell(kopf.Game)
	if err != nil {
		return "", err
	}

	return modell.Skript, nil
}

// Kartenfelder gibt die Feldbeschreibungen des Spiels, zu dem ein Set gehört.
//
// **Für die Oberfläche und nicht für den Plan.** Der Karteneditor braucht die
// Form jedes Feldes — Art, Mehrzeiligkeit, Wahlen, Beschreibung —, um das
// passende Steuerelement zu wählen. Ohne sie stellt er jedes Feld als
// einzeiliges Textfeld dar, auch den Regeltext und die Seltenheit
// (#560).
//
// **Getrennt von den Kartenwerten.** Die Werte ändern sich bei jedem
// Tastendruck, die Beschreibungen nicht; ein gemeinsamer Aufruf schickte die
// Erklärung jedes Feldes durch jede Antwort.
func (w *Stilwaehler) Kartenfelder(ctx context.Context, setID int64) ([]domain.Field, error) {
	kopf, err := w.kopf(ctx, setID)
	if err != nil {
		return nil, err
	}

	modell, err := w.spielmodell(kopf.Game)
	if err != nil {
		return nil, err
	}

	return modell.Fields, nil
}

// Setfelder gibt die `set field:`-Blöcke des Spiels, zu dem ein Set gehört —
// #567.
//
// # Warum aus dem Spiel und nicht aus dem Stil
//
//	void SetInfoEditor::onChangeSet() {
//	  setStyles(set->stylesheet, set->stylesheet->set_info_style);
//	  setData(set->data);
//	}
//
// `src/gui/control/native_look_editor.cpp`, Zeilen 187–190, Fassung f9d9356
// vom 2021-07-06. Der Stil **schmückt** die Felder; welche es gibt, steht im
// Spiel: `set_info_style.cloneFrom(game->default_set_style)`
// (`src/data/stylesheet.cpp`, Zeile 110).
//
// # Und ohne Sichtbarkeitsfilter
//
// `NativeLookEditor::resizeViewers` legt **jede** Ansicht untereinander und
// setzt ihren Kasten selbst (Zeilen 48–93). Anders als beim Karteneditor
// entscheidet der Stil hier nicht, was zu sehen ist. Über `magic.mse-game`
// sind das 89 Felder; `default set style:` nennt zwei davon, `set info style:`
// von `magic-m15` genau eines — der Rest steht mit Vorgabestil da.
func (w *Stilwaehler) Setfelder(ctx context.Context, setID int64) ([]domain.Field, error) {
	kopf, err := w.kopf(ctx, setID)
	if err != nil {
		return nil, err
	}

	modell, err := w.spielmodell(kopf.Game)
	if err != nil {
		return nil, err
	}

	return modell.SetFields, nil
}

// Autoersetzungen gibt die `auto replace:`-Regeln des Spiels eines Sets.
//
// **Über denselben Weg wie [Stilwaehler.Kartenfelder]** (#897): Setkopf, Spiel,
// Spielpaket. Ein eigener Weg daneben suchte das Paket ein viertes Mal.
//
// **Kein Fehler ohne Spielpaket.** Es fehlt auf einem frischen Rechner, und
// dann wird eben nichts ersetzt. Wer daraus einen Fehler machte, hielte das
// Tippen an, weil eine Vorlage nicht daliegt (§8).
func (w *Stilwaehler) Autoersetzungen(ctx context.Context, setID int64) (
	[]msepaket.AutoErsetzung, error,
) {
	kopf, err := w.kopf(ctx, setID)
	if err != nil {
		return nil, err
	}

	return w.spielregeln(kopf.Game), nil
}
