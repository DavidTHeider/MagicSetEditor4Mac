package app

// **Wie das Fenster aufgehen soll** — UX-14.AK-01
// (#1686).
//
// # Warum hier und nicht in `main.go`
//
// **Weil `main.go` keinen Bestand liest.** Es baut die Anwendung, ruft
// `Starten` und reicht die Angaben an die Hülle; was in den Einstellungen
// steht, weiß diese Schicht. `internal/fenster` übersetzt den Wert in Wails'
// Startzustand — das ist die Wails-Seite, und sie bleibt dort (ENG-02).
//
// # Und warum es keine Bindung ist
//
// **Weil das Fenster zu dem Zeitpunkt noch nicht steht.** Eine Bindung wird
// aus dem Fenster gerufen; dies wird **vor** ihm gerufen, aus `main.go`. Es
// taucht deshalb in keiner Flächenzahl auf und in keiner Fähigkeit — es ist
// kein Weg nach außen, sondern einer nach innen.

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Fensterstand ist, wie das Fenster aufgehen soll.
type Fensterstand struct {
	// Start ist einer der drei Werte aus [domain.EinstellungFensterstart].
	Start string

	// Maximiert sagt, ob das Fenster den Bildschirm füllen soll.
	//
	// **Als Antwort und nicht als Vergleich beim Aufrufer.** `main.go` darf
	// `internal/domain` nicht kennen (§2: zulässig sind `app` und
	// `mcpnaehte`), und ein `gewaehlt == "maximiert"` dort wäre das Wort der
	// Domäne als Zeichenkette abgeschrieben — dieselbe Regel an zwei Stellen,
	// von denen eine unter keinem Wächter steht.
	Maximiert bool

	// Breite und Hoehe sind die gemerkten Maße — **null, wenn es keine
	// gibt**. Dann nimmt `main.go` seine eingebaute Größe: wer „wie zuletzt"
	// wählt, bevor es ein Zuletzt gibt, bekommt kein Fenster von null
	// Punkten.
	Breite int
	Hoehe  int
}

// Fensterstand liest, wie das Fenster aufgehen soll.
//
// **Ohne offenen Bestand gilt die Vorgabe.** Genau dann gibt es keine
// Einstellung zu lesen — und genau dann öffnet das Fenster zur
// Wiederherstellung (#962). Ein Fenster, das dabei in einer merkwürdigen
// Größe aufginge, wäre der zweite Schreck nach dem ersten.
func (a *App) Fensterstand(ctx context.Context) Fensterstand {
	ctx = a.imProfil(ctx)

	stand := Fensterstand{
		Start:     domain.FensterstartMaximiert,
		Maximiert: true,
	}

	if a.einstellungen == nil {
		return stand
	}

	werte, err := a.einstellungen.Alle(ctx)
	if err != nil {
		return stand
	}

	if gewaehlt, steht := werte[domain.EinstellungFensterstart]; steht && gewaehlt != "" {
		stand.Start = gewaehlt
	}

	stand.Maximiert = stand.Start == domain.FensterstartMaximiert

	// **Die Maße gehören zu „wie zuletzt"** — „normal" ist die eingebaute
	// Größe (`domain.FensterstartNormal`), auch wenn eine gemerkt ist.
	if stand.Start == domain.FensterstartZuletzt {
		stand.Breite = zahlOderNull(werte[domain.EinstellungFensterbreite])
		stand.Hoehe = zahlOderNull(werte[domain.EinstellungFensterhoehe])
	}

	return stand
}

// FenstergroesseMerken schreibt zurück, wie groß das Fenster zuletzt war.
//
// **Ohne Antwort und ohne Fehlerweg.** Sie läuft beim Schließen; ein Satz,
// den niemand mehr liest, wäre kein Satz. Was hier scheitert, kostet die
// gemerkte Größe und sonst nichts — beim nächsten Start gilt dann die
// eingebaute.
//
// **Eine unsinnige Größe wird nicht gemerkt**: `domain.PruefeFenstermass`
// weist sie ab, und `Setzen` gibt den Fehler zurück, den wir hier verwerfen.
func (a *App) FenstergroesseMerken(ctx context.Context, breite, hoehe int) {
	ctx = a.imProfil(ctx)

	if a.einstellungen == nil {
		return
	}

	for schluessel, wert := range map[string]int{
		domain.EinstellungFensterbreite: breite,
		domain.EinstellungFensterhoehe:  hoehe,
	} {
		//nolint:errcheck // beim Schließen liest niemand mehr einen Satz
		_ = a.einstellungen.Setzen(ctx, schluessel, strconv.Itoa(wert))
	}
}

// zahlOderNull liest eine gemerkte Zahl — null, wo keine steht.
func zahlOderNull(wert string) int {
	zahl, err := strconv.Atoi(wert)
	if err != nil {
		return 0
	}

	return zahl
}
