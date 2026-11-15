package service

// **Die Rückseite ist ein Plan und keine Datei** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Zuerst der Rücken aus dem Spielpaket
//
// **Wenn das Spielpaket einen mitbringt, ist er es.** `magic.mse-game` führt
// `card-back.png`, und das ist der echte Rücken einer Magic-Karte;
// `space.mse-game` führt seinen eigenen. Beide liegen auf **diesem** Rechner,
// weil jemand das Paket bezogen hat — sie gehen denselben Weg wie jeder
// Rahmen und jedes Symbol, und dieses Programm liefert sie nicht mit. Genau
// das verlangt §11: die Lizenz deckt den **Code**, nicht fremde Grafik, und
// nichts davon kommt ins Repositorium.
//
// **Am Dateinamen und nicht am `icon:` des Pakets.** Gemessen am 2026-09-15
// über `Full-Magic-Pack`: von sieben Spielpaketen nennen **fünf** unter
// `icon:` eine `card-sample` — eine **Vorderseite**. Wer das Symbol nähme,
// legte auf jede verdeckte Archenemy-Karte eine offene.
//
// **Und er ist klein: 75 × 104 Punkte.** Die kleinste Güte hier ist 128
// breit, die größte 1536 — er wird also immer vergrößert, an der oberen Stufe
// um das Zwanzigfache. Das ist hinnehmbar und kein Versehen: ein Rücken soll
// nicht lesbar sein, sondern wiedererkennbar, und weich wird dabei das
// **Motiv**. Was scharf bleibt, ist der Rand — den zeichnet der Plan, in
// jeder Größe.
//
// # Und wo keiner liegt, wird einer gezeichnet
//
// **Ein eigenes Muster statt gar keines Bildes.** Ein Spiel aus `saat/` hat
// keinen Rücken, und eine verdeckte Karte muss trotzdem aussehen wie eine
// Karte. Gezeichnet bekommt sie **von selbst** alle fünf Güten, denselben
// Schlüssel und dieselbe Zwischenspeicherung wie jede andere Textur.
//
// # Und warum beide Wege durch denselben Schlüssel gehen
//
// **Weil der Rücken aus dem Paket sich ändern kann.** Sein Inhalt wandert
// über [Ressourcenquelle] in den Schlüssel; wer die Datei austauscht,
// bekommt ein neues Bild statt des alten.

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"

// Rueckenbreite, Rueckenhoehe und Rueckendpi sind die Maße jedes Rückens.
//
// **Fest und nicht aus einer Karte genommen.** Sie sind die Kartengröße von
// MSE — 375 × 523 Punkte bei 150 dpi, das Maß jeder `magic-m15`-Vorlage.
//
// **Der erste Anlauf nahm die Maße der Vorderseite**, damit eine verdeckte
// Karte genauso groß daliegt wie eine offene. Der Gedanke war richtig und
// die Ausführung falsch: damit trugen Breite, Höhe **und Schlüssel** die
// Maße der Vorlage darunter, und zwei verdeckte Karten unterschieden sich um
// 512 × 512 gegen 258 × 258 — gemessen am 2026-09-15. Das Issue verbietet es
// ausdrücklich: „fremde verdeckte Karten nicht über Texturmetadaten
// verraten."
//
// **Der Preis steht hier**: eine Karte im Sonderformat liegt verdeckt in
// Kartengröße da. Eine Ungenauigkeit in der Darstellung wiegt weniger als
// eine Auskunft über eine verdeckte Karte.
const (
	Rueckenbreite = 375.0
	Rueckenhoehe  = 523.0
	Rueckendpi    = 150
)

// Rueckseitenplan gibt den Plan des Kartenrückens.
//
// **Mit `bildpfad` den aus dem Paket, ohne ihn den gezeichneten** — die
// Gründe stehen im Kopf dieser Datei.
//
// # Jede Ebene trägt `Deckung: 1`
//
// **Ohne sie ist sie unsichtbar.** [render.Ebene.Deckung] geht von 0 bis 1,
// und der Nullwert von `float64` ist die Null: `Rasterer.punkt` steigt bei
// `anteil <= 0` aus, ohne etwas zu malen. Der erste Anlauf setzte sie
// nirgends — am Tisch standen achtzehn Rücken mit **durchweg
// durchsichtigen** Bildpunkten, und was man sah, war der CSS-Rahmen des
// Kästchens darum.
//
// **Gefunden hat es die Sichtprüfung am 2026-09-15**, und zwar erst, nachdem
// die Bilder überhaupt ankamen: `(0,0,0,0)` an jeder Stelle eines
// 512 × 714 großen PNG. Kein Lauf sah es — sie prüfen den Plan, den
// Schlüssel und die Maße, und alle drei waren richtig.
func Rueckseitenplan(bildpfad string) render.Plan {
	breite, hoehe, dpi := Rueckenbreite, Rueckenhoehe, Rueckendpi

	if bildpfad != "" {
		return render.Plan{
			Breite: breite, Hoehe: hoehe, DPI: dpi,
			Ebenen: []render.Ebene{{
				Art:     render.ArtBild,
				Rahmen:  render.Rechteck{Breite: breite, Hoehe: hoehe},
				Bild:    render.Bildbaum{Knotenart: "datei", Pfad: bildpfad},
				Deckung: 1,
			}},
		}
	}

	rand := breite / 25

	return render.Plan{
		Breite: breite, Hoehe: hoehe, DPI: dpi,
		Ebenen: []render.Ebene{
			// **Der Grund** — eine dunkle Fläche über die ganze Karte.
			{
				Art:     render.ArtFlaeche,
				Rahmen:  render.Rechteck{Breite: breite, Hoehe: hoehe},
				Farbe:   render.Farbe{R: 24, G: 26, B: 31, A: 255},
				Ecke:    rand,
				Deckung: 1,
			},
			// **Und ein Feld darin** — der Rand, den jede Karte zeigt.
			{
				Art: render.ArtFlaeche,
				Rahmen: render.Rechteck{
					Links: rand, Oben: rand,
					Breite: breite - 2*rand, Hoehe: hoehe - 2*rand,
				},
				Farbe:   render.Farbe{R: 58, G: 62, B: 74, A: 255},
				Ecke:    rand / 2,
				Deckung: 1,
			},
		},
	}
}
