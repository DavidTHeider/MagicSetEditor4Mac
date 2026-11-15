package service_test

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// sprachpaketdoppel hält Sprachpakete im Arbeitsspeicher.
type sprachpaketdoppel struct {
	pakete        map[string]domain.Sprachpaket
	fehler        error
	schreibfehler error
	geschrieben   int
}

func neuesSprachpaketdoppel() *sprachpaketdoppel {
	return &sprachpaketdoppel{pakete: map[string]domain.Sprachpaket{}}
}

func (d *sprachpaketdoppel) Sprachpakete(context.Context) ([]domain.Sprachpaket, error) {
	if d.fehler != nil {
		return nil, d.fehler
	}

	pakete := []domain.Sprachpaket{}
	for _, kennung := range slices.Sorted(maps.Keys(d.pakete)) {
		pakete = append(pakete, d.pakete[kennung])
	}

	return pakete, nil
}

func (d *sprachpaketdoppel) Sprachpaket(_ context.Context, kennung string) (domain.Sprachpaket, error) {
	if d.fehler != nil {
		return domain.Sprachpaket{}, d.fehler
	}

	paket, da := d.pakete[kennung]
	if !da {
		return domain.Sprachpaket{}, fmt.Errorf("%w: %s", domain.ErrNotFound, kennung)
	}

	paket.Texte = maps.Clone(paket.Texte)

	return paket, nil
}

func (d *sprachpaketdoppel) SprachpaketAnlegen(_ context.Context, paket domain.Sprachpaket) error {
	if d.schreibfehler != nil {
		return d.schreibfehler
	}

	if _, da := d.pakete[paket.Kennung]; da {
		return domain.ErrDoppelt
	}

	d.pakete[paket.Kennung] = paket
	d.geschrieben++

	return nil
}

func (d *sprachpaketdoppel) SprachpaketSchreiben(_ context.Context, paket domain.Sprachpaket) error {
	if d.schreibfehler != nil {
		return d.schreibfehler
	}

	d.pakete[paket.Kennung] = paket
	d.geschrieben++

	return nil
}

// SprachpaketAendern ist im Doppel ein Lesen, Ändern und Schreiben — einen
// zweiten Prozess gibt es hier nicht; das Rennen prüft die Ablage selbst.
func (d *sprachpaketdoppel) SprachpaketAendern(
	ctx context.Context, kennung string, aendere func(*domain.Sprachpaket) error,
) error {
	paket, err := d.Sprachpaket(ctx, kennung)
	if err != nil {
		return err
	}

	if err := aendere(&paket); err != nil {
		return err
	}

	return d.SprachpaketSchreiben(ctx, paket)
}

func (d *sprachpaketdoppel) SprachpaketLoeschen(_ context.Context, kennung string) error {
	if _, da := d.pakete[kennung]; !da {
		return fmt.Errorf("%w: %s", domain.ErrNotFound, kennung)
	}

	delete(d.pakete, kennung)

	return nil
}

// sprachvorlage ist ein kleiner Ausschnitt des deutschen Katalogs.
var sprachvorlage = domain.Katalogvorlage{
	"allgemein.titel":                     "Sets",
	"fehler.symbol":                       "Die Datei ist größer als {grenze}.",
	"allgemein.zaehlwoerter.karten.eine":  "Karte",
	"allgemein.zaehlwoerter.karten.viele": "Karten",
}

// sprachuhr steht still, damit eine Ausgabe dieselbe bleibt.
func sprachuhr() time.Time { return time.Date(2026, 9, 29, 23, 0, 0, 0, time.UTC) }

func sprachdienst(ablage *sprachpaketdoppel) *service.Sprachpakete {
	return service.NeueSprachpakete(ablage, sprachvorlage, "MagicSetEditor4Mac 0.6.0", sprachuhr)
}
