// **Die Stände der bezogenen Bäume für den Dienst** — G04
// (#2251).
//
// Gerechnet wird in `service`, übersetzt in `app` ([app.Paketweg.Bezuege]);
// hier wird nur die Gestalt in die von `internal/rest` gebracht.

package main

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// PaketBezuege nennt die bezogenen Bäume.
func (p pakete) PaketBezuege(ctx context.Context, fragen bool) ([]rest.Bezugsstand, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return nil, err
	}

	bezuege, err := weg.Bezuege(ctx, fragen)
	if err != nil {
		return nil, err
	}

	staende := make([]rest.Bezugsstand, 0, len(bezuege))
	for _, bezug := range bezuege {
		stand := rest.Bezugsstand{
			Name: bezug.Name, Adresse: bezug.Adresse, Stand: bezug.Stand, Geprueft: bezug.Geprueft,
			Pakete: bezug.Pakete, Vorige: bezug.Vorige, VorigerStand: bezug.VorigerStand, Lage: bezug.Lage,
			Spitze: bezug.Spitze, SpitzeZeit: bezug.SpitzeZeit, Vorsprung: bezug.Vorsprung,
		}

		if bezug.Fehler != nil {
			stand.Grund = bezug.Fehler.Meldung
		}

		staende = append(staende, stand)
	}

	return staende, nil
}

// PaketBezugVorschau sagt, was eine Aktualisierung täte.
func (p pakete) PaketBezugVorschau(ctx context.Context, name, stand string) (rest.Bezugsvorschau, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return rest.Bezugsvorschau{}, err
	}

	vorschau, err := weg.BezugVorschau(ctx, name, stand)
	if err != nil {
		return rest.Bezugsvorschau{}, err
	}

	pakete := make([]rest.Paketaenderung, 0, len(vorschau.Pakete))
	for _, paket := range vorschau.Pakete {
		pakete = append(pakete, rest.Paketaenderung{
			Paket: paket.Paket, Neu: paket.Neu, Geaendert: paket.Geaendert, Entfernt: paket.Entfernt,
		})
	}

	return rest.Bezugsvorschau{
		Name: vorschau.Name, Von: vorschau.Von, Zu: vorschau.Zu, Lage: vorschau.Lage,
		Vorsprung: vorschau.Vorsprung, Rueckstand: vorschau.Rueckstand, Pakete: pakete,
		Uebrige: vorschau.Uebrige, Gekuerzt: vorschau.Gekuerzt, Sets: vorschau.Sets,
		BestaetigungNoetig: vorschau.BestaetigungNoetig, Gruende: vorschau.Gruende,
		VorigeErsetzt: vorschau.VorigeErsetzt,
	}, nil
}

// PaketBezugAktualisieren aktualisiert einen Baum.
func (p pakete) PaketBezugAktualisieren(ctx context.Context, name, stand string, bestaetigt bool) (rest.Bezug, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return rest.Bezug{}, err
	}

	ergebnis, err := weg.BezugAktualisieren(ctx, name, stand, bestaetigt)
	if ergebnis == nil {
		return rest.Bezug{}, err
	}

	return rest.Bezug{
		Ort: ergebnis.Ort, Stand: ergebnis.Stand, Pruefsumme: ergebnis.Pruefsumme,
		Uebergangen: append([]string{}, ergebnis.Uebergangen...),
	}, err
}

// PaketBezugZurueck tauscht einen Baum mit seinem aufgehobenen Stand.
func (p pakete) PaketBezugZurueck(ctx context.Context, name string) (rest.Bezug, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return rest.Bezug{}, err
	}

	rueckweg, err := weg.BezugZurueck(ctx, name)
	if rueckweg == nil {
		return rest.Bezug{}, err
	}

	return rest.Bezug{Ort: rueckweg.Ort, Stand: rueckweg.Stand, Pruefsumme: rueckweg.Pruefsumme, Uebergangen: []string{}}, err
}
