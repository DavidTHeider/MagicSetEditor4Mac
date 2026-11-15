package storage_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func sprachpaket(kennung string) domain.Sprachpaket {
	return domain.Sprachpaket{
		Kennung: kennung, Name: "Português", Basis: "en", Stand: domain.StandEntwurf,
		Texte: map[string]domain.Uebersetzungstext{
			"allgemein.titel": {Text: "Conjuntos", Quelle: "Sets"},
		},
	}
}

// **Ein Sprachpaket geht in den Bestand und kommt wieder** — LOC-02
// (#1688).
func TestEinSprachpaketGehtInDenBestandUndKommtWieder(t *testing.T) {
	t.Parallel()

	ablage := storage.NeueSprachpaketablage(frischeDatenbank(t))
	ctx := probekontext(t)

	if err := ablage.SprachpaketAnlegen(ctx, sprachpaket("pt-BR")); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	// **Ohne einen Satz** steht ein leeres Objekt da und kein `null`.
	ohneSaetze := sprachpaket("da")
	ohneSaetze.Texte = nil

	if err := ablage.SprachpaketAnlegen(ctx, ohneSaetze); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	pakete, err := ablage.Sprachpakete(ctx)
	if err != nil || len(pakete) != 2 || pakete[0].Kennung != "da" || pakete[1].Kennung != "pt-BR" ||
		pakete[0].Texte == nil || len(pakete[0].Texte) != 0 {
		t.Fatalf("die Liste: %+v, %v", pakete, err)
	}

	gelesen, err := ablage.Sprachpaket(ctx, "pt-BR")
	if err != nil || gelesen.Name != "Português" || gelesen.Basis != "en" ||
		gelesen.Stand != domain.StandEntwurf || gelesen.Angelegt.IsZero() ||
		gelesen.Texte["allgemein.titel"].Quelle != "Sets" {
		t.Fatalf("gelesen: %+v, %v", gelesen, err)
	}

	// **Dieselbe Kennung zweimal ist doppelt**, und die Absage steht an der
	// Kennung.
	var feld *domain.FieldError
	if doppelt := ablage.SprachpaketAnlegen(ctx, sprachpaket("pt-BR")); !errors.Is(doppelt, domain.ErrDoppelt) ||
		!errors.As(doppelt, &feld) || feld.Field != "kennung" {
		t.Errorf("doppelt: %v", doppelt)
	}

	gelesen.Name, gelesen.Lizenz, gelesen.Herkunft = "Português do Brasil", "CC0-1.0", "Ana"
	gelesen.Texte = map[string]domain.Uebersetzungstext{}

	if schreibfehler := ablage.SprachpaketSchreiben(ctx, gelesen); schreibfehler != nil {
		t.Fatalf("schreiben: %v", schreibfehler)
	}

	neu, err := ablage.Sprachpaket(ctx, "pt-BR")
	if err != nil || neu.Name != "Português do Brasil" || neu.Lizenz != "CC0-1.0" ||
		neu.Herkunft != "Ana" || len(neu.Texte) != 0 {
		t.Errorf("geschrieben: %+v, %v", neu, err)
	}

	if err := ablage.SprachpaketLoeschen(ctx, "pt-BR"); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	for name, fehler := range map[string]error{
		"lesen":     func() error { _, e := ablage.Sprachpaket(ctx, "pt-BR"); return e }(),
		"schreiben": ablage.SprachpaketSchreiben(ctx, gelesen),
		"löschen":   ablage.SprachpaketLoeschen(ctx, "pt-BR"),
	} {
		if !errors.Is(fehler, domain.ErrNotFound) {
			t.Errorf("%s nach dem Löschen: %v", name, fehler)
		}
	}
}

// **Ein Fehler des Bestands kommt übersetzt an**, auf jedem Weg.
func TestDieSprachpaketablageMeldetEinenGeschlossenenBestand(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueSprachpaketablage(db)
	ctx := probekontext(t)

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := ablage.Sprachpakete(ctx); err == nil {
		t.Error("die Liste las aus einem geschlossenen Bestand")
	}

	if _, err := ablage.Sprachpaket(ctx, "da"); err == nil || errors.Is(err, domain.ErrNotFound) {
		t.Errorf("lesen: %v", err)
	}

	if err := ablage.SprachpaketAnlegen(ctx, sprachpaket("da")); err == nil {
		t.Error("anlegen in einen geschlossenen Bestand")
	}

	if err := ablage.SprachpaketSchreiben(ctx, sprachpaket("da")); err == nil {
		t.Error("schreiben in einen geschlossenen Bestand")
	}

	if err := ablage.SprachpaketLoeschen(ctx, "da"); err == nil {
		t.Error("löschen in einem geschlossenen Bestand")
	}

	if err := ablage.SprachpaketAendern(ctx, "da", func(*domain.Sprachpaket) error { return nil }); err == nil {
		t.Error("ändern in einem geschlossenen Bestand")
	}
}
