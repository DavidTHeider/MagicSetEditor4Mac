package service

// **Viele Zeilen in einer Anfrage** — Befund 6 des Zweitblicks über Phase 4.5
// (#1348).
//
// # Was gemessen war
//
// [Listeneinfuhr.Deuten] legte gleiche Fragen zusammen und stellte den Rest
// **seriell**. Bei 5.000 verschiedenen Namen und 600 ms Abstand
// (`scryfall.kartenabstand`) sind das rund **50 Minuten** — und die Abnahme
// von #1292 verlangt einen Bestand „von mehreren tausend Exemplaren".
//
// # Warum die Aufteilung genau hier verläuft
//
// **Eine Zeile mit Setcode benennt ihren Druck; eine ohne tut es nicht.**
//
//   - *Mit* Setcode gibt `POST /cards/collection` genau eine Karte — die
//     Kennungsform `name,set` beziehungsweise `collector_number,set`, und
//     **75 Kennungen je Anfrage**. Eine ausgeführte Deckliste trägt den
//     Setcode; das sind in der Praxis fast alle Zeilen.
//   - *Ohne* Setcode ist die Zeile mehrdeutig: „Counterspell" hat 72 Drucke
//     (nachgelesen am 2026-08-30), und die Sammelanfrage gäbe davon **einen**
//     — Scryfalls neuesten, nicht den, den jemand besitzt. Solche Zeilen
//     brauchen weiter `unique=prints`, damit es etwas zu wählen gibt (§16,
//     Regel 6).
//
// **Damit kostet eine Liste aus 5.000 gesetzten Zeilen 67 Anfragen** — rund
// 40 Sekunden statt 50 Minuten.
//
// # Was es nicht tut
//
// Es lädt die Sammeldatei nicht. Die Scryfall-FAQ empfiehlt sie für große
// Mengen einfacher Nachschlagevorgänge, und `internal/scryfall/bulk.go` gibt
// es schon — aber sie wiegt 25 MB und wäre für eine Liste mit dreißig Zeilen
// der teurere Weg. Die Sammelanfrage braucht kein Vorwissen und keinen
// Download.

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sammelliste"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// Fremdsammelsuche fragt viele Karten auf einmal.
//
// **Eine eigene Schnittstelle neben [Fremddrucksuche]**, weil sie eine andere
// Frage stellt: *diese Drucke, gib mir die Karten* statt *dieser Name, welche
// Drucke gibt es?*. Beide zusammenzulegen hieße, einen Aufrufer zu zwingen,
// die falsche zu benutzen.
type Fremdsammelsuche interface {
	Sammlung(ctx context.Context, kennungen []scryfall.Kennung) (
		antwort scryfall.Sammelantwort, alter time.Duration,
		befunde []domain.Befund, fehler error)
}

// MitSammelsuche setzt den Weg für die Sammelanfragen.
//
// **Wahlfrei, und ohne ihn geht alles seriell.** Eine Fläche, die ihn nicht
// setzt, bekommt dasselbe Ergebnis — nur langsamer. Ein Pflichtfeld daraus zu
// machen hieße, jede Naht anzufassen, damit eine Beschleunigung ankommt.
func (l *Listeneinfuhr) MitSammelsuche(sammelsuche Fremdsammelsuche) *Listeneinfuhr {
	l.sammelsuche = sammelsuche

	return l
}

// istAdressiert sagt, ob eine Zeile ihren Druck benennt.
//
// **Der Setcode entscheidet, nicht die Nummer.** `e:m11` allein gibt bei
// Scryfall eine Karte je Name im Set; eine Nummer ohne Set gibt gar nichts,
// denn Sammlernummern sind je Set vergeben.
func istAdressiert(zeile sammelliste.Zeile) bool { return zeile.Setcode != "" }

// kennungFuer macht aus einer adressierten Zeile eine Scryfall-Kennung.
//
// **Name und Nummer schließen einander aus.** Scryfalls Kennungsschemata sind
// `name,set` **oder** `collector_number,set`; wer beides schickt, stellt eine
// Frage, die dort nicht steht. Die Nummer ist die genauere — sie trifft auch
// die richtige von zwei Ausgaben desselben Namens im selben Set.
func kennungFuer(zeile sammelliste.Zeile) scryfall.Kennung {
	if zeile.Nummer != "" {
		return scryfall.Kennung{Setcode: zeile.Setcode, Nummer: zeile.Nummer}
	}

	return scryfall.Kennung{Name: zeile.Name, Setcode: zeile.Setcode}
}

// sammelnd deutet die adressierten Zeilen in Schüben.
//
// **Es gibt eine Deutung je Frage zurück**, damit der Aufrufer sie wie eine
// serielle behandelt. Was nicht gefunden wurde, bekommt seinen Befund; was
// gefunden wurde, genau einen Treffer.
func (l *Listeneinfuhr) sammelnd(
	ctx context.Context, fragen []string, zeilen map[string]sammelliste.Zeile,
) (map[string]Zeilendeutung, time.Duration, error) {
	gedeutet := map[string]Zeilendeutung{}

	var aeltester time.Duration

	for anfang := 0; anfang < len(fragen); anfang += scryfall.Sammelgrenze {
		schub := fragen[anfang:min(anfang+scryfall.Sammelgrenze, len(fragen))]

		kennungen := make([]scryfall.Kennung, 0, len(schub))
		for _, frage := range schub {
			kennungen = append(kennungen, kennungFuer(zeilen[frage]))
		}

		antwort, alter, befunde, fehler := l.sammelsuche.Sammlung(ctx, kennungen)
		if fehler != nil {
			return nil, 0, fehler
		}

		if alter > aeltester {
			aeltester = alter
		}

		// **Zugeordnet wird über die Kennung, nicht über die Stelle.**
		// Scryfall sagt es ausdrücklich: „cards that aren't found will throw
		// off the mapping of request identifiers to results, so you should
		// not rely on positional index alone." Wer über den Index ginge,
		// hätte ab der ersten fehlenden Karte alles verschoben — still.
		nachKennung := map[scryfall.Kennung]*scryfall.Karte{}
		for _, karte := range antwort.Karten {
			nachKennung[scryfall.Kennung{Setcode: karte.Set, Nummer: karte.Nummer}] = karte
			nachKennung[scryfall.Kennung{Name: karte.Name, Setcode: karte.Set}] = karte
		}

		for stelle, frage := range schub {
			karte, gefunden := nachKennung[kennungen[stelle]]
			if !gefunden {
				gedeutet[frage] = Zeilendeutung{
					Befund: anhaengen("kein Druck heißt so", befunde),
				}

				continue
			}

			gedeutet[frage] = Zeilendeutung{
				Drucke:  1,
				Treffer: []Suchtreffer{alsSuchtreffer(karte)},
				Befund:  anhaengen("", befunde),
			}
		}
	}

	return gedeutet, aeltester, nil
}
