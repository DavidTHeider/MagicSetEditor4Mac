package storage_test

// **Ein fremder Wert vor der Annahme ist ein Konflikt** — U02b
// (#2179).
//
// Das Fenster schreibt einen Entwurf mit den Werten, die es von diesem Feld
// kennt: dem Ausgangswert, den selbst geschickten, den gesehenen. Steht im
// Bestand **ein anderer**, hat jemand anderes — MCP, CLI, REST — dazwischen
// geschrieben, und der Entwurf darüber wäre ein stilles Überschreiben. Geprüft
// wird **in derselben Transaktion** wie das Schreiben; ohne Bedingung bleibt
// jeder Schreiber, wie er ist.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// bedingterWeg ist ein Schreibweg, sein Lesen und seine Vorbereitung.
type bedingterWeg struct {
	name      string
	schreiben func(ctx context.Context, wert string) error
	lesen     func() string
}

func bedingteWege(t *testing.T) []bedingterWeg {
	t.Helper()

	kartenDB := frischeDatenbank(t)
	karteID := eineKarte(t, kartenDB, domain.Value{Field: "name", Text: "Ausgang"})
	karten := storage.NeuerKartenschreiber(kartenDB)
	leser := storage.NeuerSetLeser(kartenDB)

	setID, err := storage.NeuerSetSpeicher(kartenDB).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Setangaben", Game: "magic", Values: []domain.Value{{Field: "title", Text: "Ausgang"}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	wortDB, wortSet := mitWoertern(t)
	wort := woerter(t, wortDB, wortSet)[0]
	ablage := storage.NeueSchluesselwortablage(wortDB)

	if err := ablage.SchluesselwortSetzen(probekontext(t), wortSet, wort.ID, "erinnerung", "Ausgang"); err != nil {
		t.Fatal(err)
	}

	return []bedingterWeg{
		{
			"Kartenfeld",
			func(ctx context.Context, wert string) error {
				_, _, err := karten.FeldSetzenMitVorwert(ctx, karteID, "name", wert)

				return err
			},
			func() string {
				var gespeichert string
				if lesefehler := kartenDB.QueryRowContext(probekontext(t),
					`SELECT wert FROM card_values WHERE card_id = ? AND feld = 'name'`, karteID).
					Scan(&gespeichert); lesefehler != nil {
					t.Fatal(lesefehler)
				}

				return gespeichert
			},
		},
		{
			"Setangabe",
			func(ctx context.Context, wert string) error {
				_, _, err := karten.SetwertSetzenMitVorwert(ctx, setID, "title", wert)

				return err
			},
			func() string { return setwert(t, leser, setID, "title") },
		},
		{
			"Schlüsselwort",
			func(ctx context.Context, wert string) error {
				return ablage.SchluesselwortSetzen(ctx, wortSet, wort.ID, "erinnerung", wert)
			},
			func() string { return woerter(t, wortDB, wortSet)[0].Reminder },
		},
	}
}

func TestEinFremderWertVorDerAnnahmeIstEinKonflikt(t *testing.T) {
	for _, weg := range bedingteWege(t) {
		t.Run(weg.name, func(t *testing.T) {
			// **Jemand anderes schreibt zuerst** — ohne Bedingung, wie MCP.
			if err := weg.schreiben(probekontext(t), "Fremd"); err != nil {
				t.Fatalf("fremd schreiben: %v", err)
			}

			// **Dann das Fenster, das nur den Ausgangswert kennt.**
			bedingt := storage.MitBekanntenWerten(probekontext(t), []string{"Ausgang", "Entwurf"})
			fehler := weg.schreiben(bedingt, "Entwurf")

			if !errors.Is(fehler, domain.ErrConflict) {
				t.Fatalf("Entwurf über einen fremden Wert: %v, erwartet ein Konflikt", fehler)
			}

			var feldfehler *domain.FieldError
			if !errors.As(fehler, &feldfehler) || feldfehler.Key != domain.SchluesselFremderWert {
				t.Errorf("der Befund steht nicht am Feld: %#v", fehler)
			}

			if gespeichert := weg.lesen(); gespeichert != "Fremd" {
				t.Errorf("gespeichert ist %q — der fremde Wert wurde überschrieben", gespeichert)
			}

			// **Kennt das Fenster ihn, schreibt es** — nach „behalten" etwa.
			gesehen := storage.MitBekanntenWerten(probekontext(t), []string{"Ausgang", "Fremd", "Entwurf"})
			if err := weg.schreiben(gesehen, "Entwurf"); err != nil {
				t.Fatalf("mit gesehenem Wert: %v", err)
			}

			if gespeichert := weg.lesen(); gespeichert != "Entwurf" {
				t.Errorf("gespeichert ist %q, erwartet der Entwurf", gespeichert)
			}
		})
	}
}
