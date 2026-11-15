package service

// **Ersetzen und zurück** — DOM-01.AK-02 und AK-03
// (#1696):
// „Aktualisierung … und Rückkehr zur vorigen kompatiblen Version" sowie
// „eigene Kopien und geänderte Pakete werden nicht … überschrieben".
//
// # Eine vorige Fassung, nicht eine Geschichte
//
// **Ersetzen hebt die bisherige Fassung auf** — samt einem Vermerk über ihre
// Herkunft, ihre Prüfsumme und ihren Abdruck —, und zurück tauscht beide.
// Eine Ebene reicht, damit ein Ersetzen nie etwas unwiderruflich nimmt; eine
// Geschichte darüber hinaus wäre ein zweites Paketverzeichnis, das niemand
// aufräumt.
//
// **Ein geändertes Paket wird nur mit Bestätigung ersetzt** — und auch dann
// geht die Änderung nicht verloren: sie ist die vorige Fassung. Automatische
// Aktualisierungen gibt es nicht; jede geht von einem Menschen aus.
//
// Portierungsmarke **adapted**: MSE2 ersetzt beim Aktualisieren ohne
// Aufbewahrung und warnt bei `PACKAGE_CONFLICTS` (Aktualisierung **und**
// lokale Änderung, `src/data/installer.hpp`, Fassung f9d9356 vom
// 2021-07-06). Hier ist der Konflikt eine Bestätigung, und die Änderung wird
// aufgehoben.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Ersetzen legt eine neue Fassung eines vorhandenen Pakets ein.
func (v *Paketverwalter) Ersetzen(ctx context.Context, quelle string, geaendertBestaetigt bool) (domain.Paket, error) {
	if err := v.ortPruefen(); err != nil {
		return domain.Paket{}, err
	}

	neu, err := v.paketLesen(quelle)
	if err != nil {
		return domain.Paket{}, err
	}

	_, alt, err := v.vorhandenes(ctx, neu.Art, neu.Name)
	if err != nil {
		return domain.Paket{}, err
	}

	if pruefung := v.verwaltetPruefen(alt); pruefung != nil {
		return domain.Paket{}, pruefung
	}

	if alt.Geaendert && !geaendertBestaetigt {
		return domain.Paket{}, domain.NewFieldError("paket", domain.SchluesselPaketGeaendert,
			"%q weicht vom Stand beim Einbinden ab; wer es ersetzt, bestätigt es — "+
				"die Änderung bleibt als vorige Fassung erhalten.", alt.Dateiname())
	}

	// **Die bisherige vorige Fassung weicht erst, wenn alles gelang** —
	// Befund Z1 des Zweitblicks über Phase 6.5: sie wurde geräumt, bevor die
	// neue lag, und ein Scheitern danach nahm den Weg zurück zu ihr mit.
	vorige := v.vorigeOrt(alt)
	if reste := v.resteHolen(vorige); reste != nil {
		return domain.Paket{}, reste
	}

	beiseite, err := v.beiseiteLegen(vorige)
	if err != nil {
		return domain.Paket{}, err
	}

	if err := v.aufheben(alt, vorige); err != nil {
		return domain.Paket{}, mitRueckweg(err, alt.Pfad, beiseite.zurueck())
	}

	// **Kein halber Stand**: die neue Fassung geht, die alte und die vorige
	// kommen zurück — auch wenn erst der Katalog klemmt (Befund Z2).
	rueck := func() error {
		return errors.Join(os.RemoveAll(alt.Pfad), v.umbenennen(vorige, alt.Pfad), beiseite.zurueck())
	}

	if err := kopierePaket(quelle, alt.Pfad); err != nil {
		return domain.Paket{}, mitRueckweg(
			fmt.Errorf("die neue Fassung von %q einlegen: %w", alt.Dateiname(), err), alt.Pfad, rueck())
	}

	if _, err := v.einpflegen(ctx, Vermerk{Ort: alt.Pfad, Herkunft: quelle}, neu.Art, neu.Name); err != nil {
		return domain.Paket{}, mitRueckweg(err, alt.Pfad, rueck())
	}

	beiseite.verwerfen()

	neu.Pfad, neu.Herkunft = alt.Pfad, quelle

	return *neu, nil
}

// Zurueck tauscht die aktuelle und die vorige Fassung — und ist damit selbst
// umkehrbar.
func (v *Paketverwalter) Zurueck(ctx context.Context, art, name string) (domain.Paket, error) {
	_, alt, err := v.vorhandenes(ctx, art, name)
	if err != nil {
		return domain.Paket{}, err
	}

	if pruefung := v.verwaltetPruefen(alt); pruefung != nil {
		return domain.Paket{}, pruefung
	}

	vorige := v.vorigeOrt(alt)
	if reste := v.resteHolen(vorige); reste != nil {
		return domain.Paket{}, reste
	}

	vermerk, err := vermerkLesen(vorige)
	if err != nil {
		return domain.Paket{}, err
	}

	// **Ein Fehler nach dem Bereitlegen nimmt den Tausch zurück** — Befund Z2
	// des Zweitblicks über Phase 6.5. Sonst läge die vorige Fassung an ihrem
	// Ort, der Katalog nennte die bisherige, und der Rufer hörte „nichts
	// geändert". Der Vermerk der vorigen Fassung wird beim Aufheben
	// überschrieben und kommt aus dem Gelesenen zurück.
	tausch := vorige + ".tausch"
	alterVermerk, _ := json.Marshal(vermerk) //nolint:errcheck,errchkjson // Zeichenketten
	rueck := func(eingelegt bool) error {
		var fehler []error
		if eingelegt {
			fehler = append(fehler, v.umbenennen(alt.Pfad, tausch))
		}

		return errors.Join(append(fehler, v.umbenennen(vorige, alt.Pfad), v.umbenennen(tausch, vorige),
			schreibeMarke(vorige+".vermerk", alterVermerk))...)
	}

	if err = v.umbenennen(vorige, tausch); err != nil {
		return domain.Paket{}, fmt.Errorf("die vorige Fassung von %q bereitlegen: %w", alt.Dateiname(), err)
	}

	if err = v.aufheben(alt, vorige); err != nil {
		return domain.Paket{}, mitRueckweg(err, alt.Pfad, v.umbenennen(tausch, vorige))
	}

	if err = v.umbenennen(tausch, alt.Pfad); err != nil {
		return domain.Paket{}, mitRueckweg(
			fmt.Errorf("die vorige Fassung von %q einlegen: %w", alt.Dateiname(), err), alt.Pfad, rueck(false))
	}

	vermerk.Ort = alt.Pfad

	paket, err := v.einpflegen(ctx, vermerk, art, name)
	if err != nil {
		return domain.Paket{}, mitRueckweg(err, alt.Pfad, rueck(true))
	}

	return paket, nil
}

// aufheben legt ein Paket samt Vermerk als vorige Fassung ab.
func (v *Paketverwalter) aufheben(paket domain.Paket, vorige string) error {
	if err := os.MkdirAll(filepath.Dir(vorige), 0o750); err != nil {
		return fmt.Errorf("den Ort der vorigen Fassung anlegen: %w", err)
	}

	_ = os.RemoveAll(vorige) //nolint:errcheck // eine ältere vorige Fassung weicht; fehlt sie, gibt es nichts zu räumen

	if err := v.umbenennen(paket.Pfad, vorige); err != nil {
		return fmt.Errorf("die bisherige Fassung von %q aufheben: %w", paket.Dateiname(), err)
	}

	roh, _ := json.Marshal(Vermerk{ //nolint:errcheck,errchkjson // drei Zeichenketten
		Herkunft: paket.Herkunft, Pruefsumme: paket.Pruefsumme, Abdruck: paket.Abdruck,
	})

	if err := schreibeMarke(vorige+".vermerk", roh); err != nil {
		return mitRueckweg(fmt.Errorf("den Vermerk der vorigen Fassung schreiben: %w", err),
			paket.Pfad, v.umbenennen(vorige, paket.Pfad))
	}

	return nil
}

// vermerkLesen liest den Vermerk einer vorigen Fassung — **ohne ihn gibt es
// keine**, denn beides entsteht zusammen.
func vermerkLesen(vorige string) (Vermerk, error) {
	roh, err := msepaket.LeseDateiBegrenzt(vorige + ".vermerk")
	if err != nil {
		return Vermerk{}, fmt.Errorf("%w: zu %q gibt es keine vorige Fassung", domain.ErrNotFound, filepath.Base(vorige))
	}

	var vermerk Vermerk
	if err := json.Unmarshal(roh, &vermerk); err != nil {
		return Vermerk{}, fmt.Errorf("%w: der Vermerk zu %q ist unlesbar", domain.ErrValidation, filepath.Base(vorige))
	}

	return vermerk, nil
}

// verwaltetPruefen weist ein Paket außerhalb des verwalteten Verzeichnisses
// ab — dieselbe Grenze wie beim Entfernen (#58).
func (v *Paketverwalter) verwaltetPruefen(paket domain.Paket) error {
	if !imVerzeichnis(paket.Pfad, v.verwaltet) {
		return domain.NewFieldError("pfad", domain.SchluesselValidierung,
			"%q liegt nicht im verwalteten Paketverzeichnis und wird deshalb nicht verändert.", paket.Pfad)
	}

	return nil
}
