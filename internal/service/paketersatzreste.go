package service

// **Was ein Wechsel neben der vorigen Fassung liegen lässt** — Befunde N1
// bis N4 des Nachlaufs über Phase 6.5 (DOM-01,
// #1696).
//
// Ein Wechsel besteht aus Umbenennungen, und zwischen zweien kann der
// Prozess enden. **Was dann daneben liegt, ist keine Abfallware**: ein
// `.beiseite` ist die vorige Fassung eines abgebrochenen Ersetzens, ein
// `.tausch` die eines abgebrochenen Zurückholens. Sie werden geholt, nicht
// geräumt — und wo das nicht eindeutig geht, entscheidet ein Mensch.
//
// # Was es nicht hält
//
// Ein Abbruch **nach** dem Aufheben hinterlässt am Ort der vorigen Fassung
// die bisher aktuelle; welche von beiden gilt, lässt sich aus den Dateien
// nicht ablesen. Dann hält der nächste Wechsel an und nennt den Rest.

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// resteHolen legt einen liegengebliebenen Rest zurück an den Ort der
// vorigen Fassung — Befund N3: bisher wurde ein `.beiseite` geräumt, und
// mit ihm die einzige erreichbare vorige Fassung.
func (v *Paketverwalter) resteHolen(vorige string) error {
	for _, rest := range []string{vorige + ".beiseite", vorige + ".tausch"} {
		if !belegt(rest) {
			continue
		}

		if belegt(vorige) {
			return fmt.Errorf("%w: neben der vorigen Fassung %q liegt %q, der Rest eines abgebrochenen "+
				"Wechsels — welche von beiden gilt, lässt sich nicht ablesen; bitte ansehen und eine entfernen",
				domain.ErrConflict, vorige, rest)
		}

		if err := v.umbenennen(rest, vorige); err != nil {
			return fmt.Errorf("den Rest eines abgebrochenen Wechsels %q zurücklegen: %w", rest, err)
		}

		if !belegt(rest + ".vermerk") {
			continue
		}

		if err := v.umbenennen(rest+".vermerk", vorige+".vermerk"); err != nil {
			return fmt.Errorf("den Vermerk des Rests %q zurücklegen: %w", rest, err)
		}
	}

	return nil
}

// einpflegen liest die Pakete neu und gleicht den Katalog ab — **erst, wenn
// das gewechselte Paket lesbar ist** (Befund N1): der Abgleich ist
// vollständig, ein fehlendes Paket fiele aus dem Katalog, und der Rückweg
// brächte nur die Dateien zurück.
func (v *Paketverwalter) einpflegen(ctx context.Context, vermerk Vermerk, art, name string) (domain.Paket, error) {
	pakete, _ := v.lesen([]Vermerk{vermerk})
	for _, paket := range pakete {
		if paket.Art == art && paket.Name == name {
			return paket, v.katalog.Abgleichen(context.WithoutCancel(ctx), pakete)
		}
	}

	return domain.Paket{}, fmt.Errorf("%w: %q ist nach dem Wechsel nicht lesbar",
		domain.ErrValidation, filepath.Base(vermerk.Ort))
}

// mitRueckweg hängt an einen Fehler, was beim Zurücknehmen scheiterte —
// Befund N4: bisher ging es verloren, und die Meldung versprach einen Stand,
// den es nicht mehr gab. Der Ort sagt, wo nachzusehen ist.
func mitRueckweg(err error, ort string, rueckweg error) error {
	if rueckweg == nil {
		return err
	}

	return fmt.Errorf("%w — und der Rückweg scheiterte, bitte %q und die vorige Fassung daneben ansehen: %w",
		err, ort, rueckweg)
}

// beiseite hält eine vorige Fassung samt Vermerk fest, bis ein Wechsel
// gelungen ist — dann weicht sie, sonst kommt sie zurück.
type beiseite struct {
	v       *Paketverwalter
	vorige  string
	da      bool
	vermerk bool
}

// beiseiteLegen legt die vorige Fassung neben ihren Ort — gibt es keine,
// gibt es nichts zu halten.
func (v *Paketverwalter) beiseiteLegen(vorige string) (beiseite, error) {
	gehalten := beiseite{v: v, vorige: vorige}
	if !belegt(vorige) {
		return gehalten, nil
	}

	if err := v.umbenennen(vorige, vorige+".beiseite"); err != nil {
		return gehalten, fmt.Errorf("die vorige Fassung beiseitelegen: %w", err)
	}

	gehalten.da = true
	if !belegt(vorige + ".vermerk") {
		return gehalten, nil
	}

	// **Ohne ihren Vermerk ist die vorige Fassung keine** (Befund N2) — bisher
	// blieb er stehen, und das Aufheben überschrieb ihn.
	if err := v.umbenennen(vorige+".vermerk", vorige+".beiseite.vermerk"); err != nil {
		return gehalten, mitRueckweg(fmt.Errorf("den Vermerk der vorigen Fassung beiseitelegen: %w", err),
			vorige, v.umbenennen(vorige+".beiseite", vorige))
	}

	gehalten.vermerk = true

	return gehalten, nil
}

// zurueck stellt die gehaltene vorige Fassung wieder an ihren Ort und
// räumt den Vermerk, den das Aufheben inzwischen geschrieben hat.
func (b beiseite) zurueck() error {
	fehler := []error{}
	if err := os.Remove(b.vorige + ".vermerk"); err != nil && !errors.Is(err, fs.ErrNotExist) {
		fehler = append(fehler, err)
	}

	if b.da {
		fehler = append(fehler, b.v.umbenennen(b.vorige+".beiseite", b.vorige))
	}

	if b.vermerk {
		fehler = append(fehler, b.v.umbenennen(b.vorige+".beiseite.vermerk", b.vorige+".vermerk"))
	}

	return errors.Join(fehler...)
}

// verwerfen lässt die gehaltene vorige Fassung weichen. **Der neue Stand
// steht** (§7, Regel 3) — und was liegen bleibt, bleibt nicht still: der
// nächste Wechsel hält daran an und nennt es ([Paketverwalter.resteHolen]).
func (b beiseite) verwerfen() {
	_ = os.RemoveAll(b.vorige + ".beiseite")      //nolint:errcheck // siehe oben
	_ = os.Remove(b.vorige + ".beiseite.vermerk") //nolint:errcheck // dito
}
