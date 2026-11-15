package app_test

// **Fremdschreiben vor der Annahme überschreibt nicht** — U02b, AK-3
// (#2179).
//
// Gemessen am Stand `04c3efc0b` (Codex-Probe): ein anderer Klient schrieb das
// Feld, das Fenster nahm danach seinen Entwurf an, und gespeichert war der
// Entwurf — der fremde Wert war still fort. Hier schreibt der Fremde über den
// unbedingten Weg, den MCP, CLI und REST nehmen, und das Fenster über seinen
// bedingten mit den Werten, die es kannte.

import (
	"path/filepath"
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// fremdweg ist ein Feld, sein unbedingtes und sein bedingtes Schreiben.
type fremdweg struct {
	name     string
	fremd    func(wert string) app.SchreibAntwort
	entwurf  func(wert string, bekannt []string) app.SchreibAntwort
	gelesen  func() string
	ausgangs string
}

func fremdwege(t *testing.T) []fremdweg {
	t.Helper()

	// **Ein eigener Pfad**, damit der Test die Setangabe im Bestand selbst
	// lesen kann — `Setangaben` braucht das Spielpaket, und das liegt hier
	// nicht.
	pfad := filepath.Join(t.TempDir(), "bestand.db")
	anwendung := app.NeueApp(pfad)

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() { _ = anwendung.Beenden() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	quelle := kopiereVerzeichnis(t, filepath.Join("..", "..", "testdaten", "probe.mse-set"))

	geoeffnet := anwendung.SetOeffnen(t.Context(), quelle)
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	setID := geoeffnet.SetID
	karteID := ersteKartenkennung(t, anwendung, setID)
	ctx := probekontext(t)

	wort := anwendung.SchluesselwortAnlegen(ctx, setID)
	if wort.Fehler != nil {
		t.Fatalf("Schlüsselwort anlegen: %+v", wort.Fehler)
	}

	kartenwert := func() string {
		for _, wert := range anwendung.Karte(t.Context(), karteID).Karte.Werte {
			if wert.Feld == "name" {
				return wert.Wert
			}
		}

		return ""
	}

	setwert := func() string {
		db, err := storage.Oeffnen(ctx, pfad)
		if err != nil {
			t.Fatalf("öffnen: %v", err)
		}

		defer func() { _ = db.Close() }() //nolint:errcheck // nur gelesen

		kennung, err := strconv.ParseInt(setID, 10, 64)
		if err != nil {
			t.Fatal(err)
		}

		werte, err := storage.NeuerSetLeser(db).Setwerte(ctx, kennung)
		if err != nil {
			t.Fatalf("Setwerte: %v", err)
		}

		for _, wert := range werte {
			if wert.Field == "title" {
				return wert.Text
			}
		}

		return ""
	}

	erinnerung := func() string {
		for _, eines := range anwendung.Schluesselwoerter(ctx, setID).Woerter {
			if eines.ID == wort.WortID {
				return eines.Erinnerung
			}
		}

		return ""
	}

	return []fremdweg{
		{
			name: "Kartenfeld",
			fremd: func(wert string) app.SchreibAntwort {
				return app.SchreibAntwort{Fehler: anwendung.FeldSetzen(ctx, setID, karteID, "name", wert).Fehler}
			},
			entwurf: func(wert string, bekannt []string) app.SchreibAntwort {
				return app.SchreibAntwort{
					Fehler: anwendung.FeldSetzenBedingt(ctx, setID, karteID, "name", wert, bekannt).Fehler,
				}
			},
			gelesen: kartenwert, ausgangs: kartenwert(),
		},
		{
			name: "Setangabe",
			fremd: func(wert string) app.SchreibAntwort {
				return anwendung.SetwertSetzen(ctx, setID, "title", wert)
			},
			entwurf: func(wert string, bekannt []string) app.SchreibAntwort {
				return anwendung.SetwertSetzenBedingt(ctx, setID, "title", wert, bekannt)
			},
			gelesen: setwert, ausgangs: setwert(),
		},
		{
			name: "Schlüsselwort",
			fremd: func(wert string) app.SchreibAntwort {
				return anwendung.SchluesselwortSetzen(ctx, setID, wort.WortID, "erinnerung", wert)
			},
			entwurf: func(wert string, bekannt []string) app.SchreibAntwort {
				return anwendung.SchluesselwortSetzenBedingt(ctx, setID, wort.WortID, "erinnerung", wert, bekannt)
			},
			gelesen: erinnerung, ausgangs: erinnerung(),
		},
	}
}

func TestFremdschreibenVorDerAnnahmeUeberschreibtNicht(t *testing.T) {
	for _, weg := range fremdwege(t) {
		t.Run(weg.name, func(t *testing.T) {
			if antwort := weg.fremd("Fremd"); antwort.Fehler != nil {
				t.Fatalf("fremd: %+v", antwort.Fehler)
			}

			antwort := weg.entwurf("Entwurf", []string{weg.ausgangs, "Entwurf"})
			if antwort.Fehler == nil || antwort.Fehler.Schluessel != domain.SchluesselFremderWert {
				t.Fatalf("Entwurf über den fremden Wert: %+v, erwartet %q", antwort.Fehler, domain.SchluesselFremderWert)
			}

			if antwort.Fehler.Code != app.CodeKonflikt {
				t.Errorf("Code %q, erwartet %q", antwort.Fehler.Code, app.CodeKonflikt)
			}

			if gespeichert := weg.gelesen(); gespeichert != "Fremd" {
				t.Errorf("gespeichert %q — der fremde Wert wurde überschrieben", gespeichert)
			}

			// **Nach „behalten" kennt das Fenster den fremden Wert** — und
			// schreibt.
			if antwort := weg.entwurf("Entwurf", []string{weg.ausgangs, "Fremd", "Entwurf"}); antwort.Fehler != nil {
				t.Fatalf("behalten: %+v", antwort.Fehler)
			}

			if gespeichert := weg.gelesen(); gespeichert != "Entwurf" {
				t.Errorf("gespeichert %q, erwartet der Entwurf", gespeichert)
			}

			// **Ohne bekannte Werte schreibt der Weg unbedingt**, wie bisher.
			if antwort := weg.entwurf("Ohne", nil); antwort.Fehler != nil {
				t.Fatalf("ohne Bedingung: %+v", antwort.Fehler)
			}
		})
	}
}
