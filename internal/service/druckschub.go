// **Viele Drucke in wenigen Anfragen** — Befund 6 des Zweitblicks über Phase
// 4.6, #1373.
//
// # Der Befund, und was er wirklich kostete
//
// Die Formatprüfung rief `KarteNachKennung` je Platz. Bei kaltem
// Zwischenspeicher sind das für ein Commander-Deck hundert Anfragen
// nacheinander, und die Kartenwege sind auf 600 ms gedrosselt (§16, Regel 4).
//
// **Gemessen am 2026-09-01 gegen `api.scryfall.com`**, hundert verschiedene
// Drucke aus `set:dmu`, kalter Zwischenspeicher:
//
//	Einzelweg    100 Anfragen    60,13 s
//	Sammelweg      2 Anfragen     0,89 s
//
// Der Zweitblick hatte 59,4 s **hergeleitet** und dazugeschrieben, dass es
// eine Rechnung ist und keine Messung. Sie war richtig — aber das wusste
// vorher niemand, und deshalb steht die Messung hier (§10: die Reichweite
// einer Messung ist die Reichweite ihrer Aussage). Nachzufahren ist sie über
// `TestHundertDruckeInWenigerAlsFuenfSekunden` hinter dem Tag `netzlauf`.
//
// # Eine eigene Datei
//
// Weil es eine andere Frage ist als die in `formatpruefung.go`: dort steht,
// **wie geurteilt wird**, hier, **wie die Drucke herkommen** (§2).

package service

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// druckantwort ist, was zu **einem** gefragten Druck herauskam.
//
// **Immer eines von beidem.** Ein Eintrag ohne Karte trägt den Grund, warum
// keine kam — und der ist verschieden: „Scryfall kennt ihn nicht" ist etwas
// anderes als „Scryfall hat nicht geantwortet", und die Oberfläche zeigt ihn
// wörtlich an.
type druckantwort struct {
	karte *scryfall.Karte
	grund string
}

// druckschluessel glättet eine Druckkennung, damit Frage und Antwort
// zueinanderfinden.
//
// **Dieselbe Glättung wie in `scryfall.druckadresse`** — dort, weil zwei
// Schreibweisen desselben Drucks sonst zwei Einträge im Zwischenspeicher
// ergäben; hier, weil ein Deck die Kennung groß geschrieben halten kann und
// Scryfall kanonisch klein antwortet. Ohne sie fände die Antwort ihre Frage
// nicht wieder, und jede Karte eines solchen Decks bliebe ungeprüft.
func druckschluessel(roh string) string {
	return strings.ToLower(strings.TrimSpace(roh))
}

// hole beantwortet alle Drucke eines Decks — gebündelt.
//
// **Jede gefragte Kennung bekommt einen Eintrag**, auch eine, zu der nichts
// kam. Sonst müsste der Aufrufer aus einem fehlenden Eintrag raten, warum er
// fehlt, und die drei Gründe sähen gleich aus.
func (f *Formatpruefung) hole(ctx context.Context, bezuege []druckbezug) (
	map[string]druckantwort, []domain.Befund,
) {
	var (
		antworten = map[string]druckantwort{}
		kennungen []scryfall.Kennung
	)

	for _, bezug := range bezuege {
		schluessel := druckschluessel(bezug.druck)
		if schluessel == "" {
			continue
		}

		// **Ein Druck wird einmal gefragt, nicht je Ausfertigung.** Vier
		// Wälder sind eine Kennung — und Scryfall zählt jede gegen die
		// Grenze von 75.
		if _, schon := antworten[schluessel]; schon {
			continue
		}

		antworten[schluessel] = druckantwort{grund: fmt.Sprintf(
			"Scryfall hat zum Druck %s nichts geschickt.", schluessel)}
		kennungen = append(kennungen, scryfall.Kennung{ID: schluessel})
	}

	var befunde []domain.Befund

	for schub := range slices.Chunk(kennungen, scryfall.Sammelgrenze) {
		befunde = append(befunde, f.holeSchub(ctx, schub, antworten)...)
	}

	return antworten, befunde
}

// holeSchub holt bis zu [scryfall.Sammelgrenze] Drucke in **einer** Anfrage.
func (f *Formatpruefung) holeSchub(
	ctx context.Context, schub []scryfall.Kennung, ziel map[string]druckantwort,
) []domain.Befund {
	antwort, _, befunde, err := f.drucke.Sammlung(ctx, schub)
	if err != nil {
		// **Der Einzelweg springt ein** (§16, Regel 5). Er liest den
		// Zwischenspeicher **je Druck**; der Sammelweg kennt nur den Schub
		// als Ganzes. Ohne diesen Rückfall wäre ein Deck, das gestern noch
		// geprüft werden konnte, ohne Netz heute **ganz** ungeprüft — eine
		// stille Verschlechterung gegenüber dem Weg, den es ersetzt.
		return append(befunde, f.holeEinzeln(ctx, schub, ziel)...)
	}

	for _, karte := range antwort.Karten {
		ziel[druckschluessel(karte.Kennung)] = druckantwort{karte: karte}
	}

	for _, fehlend := range antwort.NichtGefunden {
		schluessel := druckschluessel(fehlend.ID)
		grund := fmt.Sprintf("Scryfall kennt den Druck %s nicht.", schluessel)

		ziel[schluessel] = druckantwort{grund: grund}
		befunde = append(befunde, domain.Befund{
			Datei: schluessel, Art: domain.ArtUmgebung,
			Meldung: grund + " Diese Karte bleibt ungeprüft.",
		})
	}

	return befunde
}

// holeEinzeln fragt Druck für Druck — der Weg vor der Bündelung.
func (f *Formatpruefung) holeEinzeln(
	ctx context.Context, schub []scryfall.Kennung, ziel map[string]druckantwort,
) []domain.Befund {
	var befunde []domain.Befund

	for _, kennung := range schub {
		karte, _, weitere, err := f.drucke.KarteNachKennung(ctx, kennung.ID)
		befunde = append(befunde, weitere...)

		if err != nil {
			grund := fmt.Sprintf(
				"Scryfall antwortete für den Druck %s nicht: %v.", kennung.ID, err)

			ziel[kennung.ID] = druckantwort{grund: grund}
			befunde = append(befunde, domain.Befund{
				Datei: kennung.ID, Art: domain.ArtUmgebung,
				Meldung: grund + " Diese Karte bleibt ungeprüft.",
			})

			continue
		}

		ziel[druckschluessel(karte.Kennung)] = druckantwort{karte: karte}
	}

	return befunde
}
