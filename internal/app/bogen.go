package app

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Bogenleger setzt Karten auf Druckseiten — die Naht zu `internal/service`.
type Bogenleger interface {
	Lege(ctx context.Context, setID int64, pfad string, masse service.Bogenmasse,
		auswahl domain.Kartenauswahl) (*service.Bogenlauf, error)

	// Probelauf rechnet denselben Bogen durch, ohne ihn zu zeichnen
	// (#2015).
	Probelauf(ctx context.Context, setID int64, masse service.Bogenmasse,
		auswahl domain.Kartenauswahl) (*service.Bogenprobelauf, error)
}

// BogenAntwort ist der Ausgang eines gelegten Bogens.
type BogenAntwort struct {
	Pfad   string `json:"pfad"`
	Seiten int    `json:"seiten"`
	Karten int    `json:"karten"`

	// Befunde sind die Stellen, an denen etwas nicht passte — eine Vorlage im
	// falschen Verhältnis, eine fehlende Schrift (§8).
	Befunde []string `json:"befunde"`

	// Abgebrochen sagt, dass niemand einen Ort gewählt hat.
	Abgebrochen bool `json:"abgebrochen"`

	Fehler *Fehler `json:"fehler"`
}

// papierformate sind die Bogengrößen, die die Oberfläche anbietet.
//
// **Als Name und nicht als zwei Zahlen**: „A4" ist etwas, das jemand auswählt;
// 210 × 297 ist etwas, das er nachschlägt. Die Zahlen stehen hier an einer
// Stelle, statt in jedem Aufruf.
var papierformate = map[string][2]float64{
	"a4":     {210, 297},
	"letter": {215.9, 279.4},
}

// bogenmasseAus macht aus den Wahlen der Oberfläche die Maße des Bogens.
//
// **Eine Stelle für den Lauf und den Probelauf** (#2015): was „Wie es wird"
// rechnet, rechnet der Bogen.
//
// **Ein unbekanntes Format fällt auf A4 zurück.** Die Auswahl der Oberfläche
// kennt nur die geführten; ein dritter Wert wäre ein Fehler, den niemand
// herbeiführen kann, und ein Bogen ohne Seitenmaß wäre die schlechtere Antwort
// darauf.
func bogenmasseAus(papier string, anschnitt float64, dpi int, marken bool) service.Bogenmasse {
	masse := service.StandardBogen()
	masse.Anschnitt = anschnitt
	masse.DPI = dpi
	masse.Schnittmarken = marken

	if format, gefuehrt := papierformate[papier]; gefuehrt {
		masse.Seitenbreite, masse.Seitenhoehe = format[0], format[1]
	}

	return masse
}

// BogenLegen schreibt die Karten eines Sets als Druckbogen.
//
// # Was die Oberfläche **nicht** einstellt
//
// Das Kartenmaß. 63 × 88 mm ist die Zusage aus
// #50 und keine
// Einstellung: eine Karte, die anders herauskommt, passt in keine Hülle. Wer
// ein anderes Spiel mit anderen Karten druckt, braucht dafür eine Entscheidung
// und kein Eingabefeld.
// # Und eine Teilmenge geht auch
//
// **`auswahl` leer heißt: das ganze Set** — Werkstück 4 von Phase 3.6
// (#1167).
// MSE2 stellt die Frage vor **vier** seiner sieben Ausgabewege — Drucken,
// Druckvorschau, HTML-Ausgabe und mehrere Bilder (`SetWindow::selectionChoices`,
// `src/gui/set/window.cpp`, Zeilen 383–389 und 656–712, Fassung f9d9356 vom
// 2021-07-06); hier stellt sie die Seite, und **entschieden wird im Dienst**.
func (a *App) BogenLegen(
	ctx context.Context, setID, pfad, papier string,
	anschnitt float64, dpi int, marken bool, auswahl []string,
) BogenAntwort {
	ctx = a.imProfil(ctx)

	leer := BogenAntwort{Befunde: []string{}}

	kennung, err := strconv.ParseInt(setID, 10, 64)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	masse := bogenmasseAus(papier, anschnitt, dpi, marken)

	gewaehlt, auswahlbefund := kartenauswahlAus(auswahl)
	if auswahlbefund != nil {
		leer.Fehler = auswahlbefund

		return leer
	}

	lauf, err := a.bogenleger.Lege(ctx, kennung, pfad, masse, gewaehlt)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	for _, befund := range lauf.Befunde {
		leer.Befunde = append(leer.Befunde, befund.String())
	}

	leer.Pfad, leer.Seiten, leer.Karten = lauf.Pfad, lauf.Seiten, lauf.Karten

	// **Ein Bogen ist eine Ausgabe** — wie MTG JSON und das Speichern; bis
	// FMT-03 sagte die Startseite danach „Noch nichts ausgegeben".
	a.arbeitVermerken(ctx, domain.ArbeitsartSet, setID, domain.VorgangAusgegeben)

	return leer
}

// MitBogenleger setzt die Naht zum Druckbogen.
func (a *App) MitBogenleger(leger Bogenleger) { a.bogenleger = leger }
