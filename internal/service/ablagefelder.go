package service

// **Was ein Zielset beim Einfügen annimmt** —
// #1282.
//
// **Eine eigene Datei**, weil `kartenablage.go` mit ihr über 400 Zeilen ginge
// (§2). Der Schnitt folgt der Frage: dort steht, **wie Karten hin- und
// hergehen**, hier, **welche ihrer Werte das Ziel überhaupt speichert**.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// speicherbare lässt weg, was das Zielset nicht speichert.
//
// **Sonst scheitert jedes Einfügen einer echten Karte.** Ein `.mse-set`
// speichert `time created` und `time modified`; die Vorlage rechnet sie, und
// `FelderVonAussen` weist jeden Schreibversuch ab. Was blieb, war eine
// angelegte Karte ohne Werte — der halbe Stand, den §8 verbietet. Im
// laufenden Fenster nachgestellt am 2026-08-28.
//
// **Ohne Feldquelle geht alles durch** — dieselbe Wahlfreiheit wie bei den
// Bildern, und dieselbe Folge: dann gilt das alte Verhalten.
//
// **Ein Fehler beim Lesen der Felder hält nichts auf.** Er hieße, dass die
// Vorlage nicht lesbar ist; dann weiß niemand, was gespeichert wird, und der
// Schreiber entscheidet es wie bisher.
func (k *Kartenablage) speicherbare(
	ctx context.Context, setID int64, werte []domain.Value,
) (bleiben []domain.Value, uebergangen []string) {
	formen := k.formenVon(ctx, setID)
	if len(formen) == 0 {
		return werte, nil
	}

	gespeichert := make(map[string]bool, len(formen))
	for _, form := range formen {
		gespeichert[form.Name] = form.Gespeichert
	}

	bleiben = make([]domain.Value, 0, len(werte))

	for _, wert := range werte {
		// **Ein Feld, das die Vorlage gar nicht kennt, geht mit.** Es kommt aus
		// einem fremden Set, und der Schreiber entscheidet darüber — nicht
		// diese Stelle.
		if steht, kennt := gespeichert[wert.Field]; kennt && !steht {
			uebergangen = append(uebergangen, wert.Field)

			continue
		}

		bleiben = append(bleiben, wert)
	}

	return bleiben, uebergangen
}

// formenVon gibt die Feldbeschreibungen eines Sets — leer, wenn es keine
// Feldquelle gibt oder die Vorlage nicht lesbar ist.
//
// **Ein Lesefehler hält nichts auf.** Er hieße, dass die Vorlage nicht lesbar
// ist; dann weiß niemand, was gespeichert wird, und alles geht durch wie
// bisher.
func (k *Kartenablage) formenVon(ctx context.Context, setID int64) []domain.Field {
	if k.felder == nil {
		return nil
	}

	formen, err := k.felder.Kartenfelder(ctx, setID)
	if err != nil {
		return nil
	}

	return formen
}
