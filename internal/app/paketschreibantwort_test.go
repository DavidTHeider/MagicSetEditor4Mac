package app_test

// **Was ein Schreibweg über sein Paket zurückgibt** — C01
// (#2226).
//
// Der Vertragsrundgang am nativen Fenster fand es: `PaketAnlegen`,
// `PaketKopieren`, `PaketErsetzen` und `PaketZurueck` gaben ihr Paket ohne
// die Zeile der Übersicht zurück — `herkunftsart` leer und `gebrauchtVon`
// als `null`, wo das Fenster `Paketherkunftsart` und eine Liste erwartet
// (`frontend/src/pakete.ts`). Dasselbe Paket hieß in `Pakete` anders als in
// der Antwort, die es eben geschrieben hat.

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// paketschreibprobe gibt bei jedem Schreibweg dasselbe Paket zurück; ihre
// Übersicht kennt es als ein eigenes, das ein Stil braucht.
type paketschreibprobe struct {
	paketfuehrer

	geschrieben domain.Paket
}

func (p *paketschreibprobe) Anlegen(context.Context, service.Neuanlage) (domain.Paket, error) {
	return p.geschrieben, nil
}

func (p *paketschreibprobe) Kopieren(context.Context, string, string, string) (domain.Paket, error) {
	return p.geschrieben, nil
}

func (p *paketschreibprobe) Ersetzen(context.Context, string, bool) (domain.Paket, error) {
	return p.geschrieben, nil
}

func (p *paketschreibprobe) Zurueck(context.Context, string, string) (domain.Paket, error) {
	return p.geschrieben, nil
}

func neuePaketschreibprobe(listenfehler error) *paketschreibprobe {
	eigenes := domain.Paket{Art: "mse-game", Name: "eigenes", Pfad: "/p/eigenes.mse-game"}

	return &paketschreibprobe{
		paketfuehrer: paketfuehrer{
			pakete:       []domain.Paket{{Art: "mse-game", Name: "fremd", Pfad: "/p/fremd.mse-game"}, eigenes},
			zusatz:       service.Paketzeile{Herkunftsart: domain.PaketEigen, GebrauchtVon: []string{"stil.mse-style"}, Benutzt: 2},
			listenfehler: listenfehler,
		},
		geschrieben: eigenes,
	}
}

// schreibwege sind die vier Bindungen, die ein Paket schreiben und es
// zurückgeben.
var schreibwege = []struct {
	name string
	ruf  func(*app.App, context.Context) app.PaketAntwort
}{
	{"PaketAnlegen", func(a *app.App, ctx context.Context) app.PaketAntwort {
		return a.PaketAnlegen(ctx, "mse-game", "eigenes", "", "")
	}},
	{"PaketKopieren", func(a *app.App, ctx context.Context) app.PaketAntwort {
		return a.PaketKopieren(ctx, "mse-game", "fremd", "eigenes")
	}},
	{"PaketErsetzen", func(a *app.App, ctx context.Context) app.PaketAntwort {
		return a.PaketErsetzen(ctx, "/q/eigenes.mse-game", true)
	}},
	{"PaketZurueck", func(a *app.App, ctx context.Context) app.PaketAntwort {
		return a.PaketZurueck(ctx, "mse-game", "eigenes")
	}},
}

// alsJSON gibt die Gestalt, die das Fenster sieht — ein Fehler beim
// Umwandeln ist einer des Tests und keine leere Zeichenkette.
func alsJSON(t *testing.T, wert any) []byte {
	t.Helper()

	roh, fehler := json.Marshal(wert)
	if fehler != nil {
		t.Fatalf("json: %v", fehler)
	}

	return roh
}

func TestEinGeschriebenesPaketTraegtSeineZeileDerUebersicht(t *testing.T) {
	for _, weg := range schreibwege {
		t.Run(weg.name, func(t *testing.T) {
			anwendung := app.NeueApp("")
			anwendung.MitPaketen(neuePaketschreibprobe(nil))

			antwort := weg.ruf(anwendung, t.Context())
			if antwort.Fehler != nil || antwort.Paket == nil {
				t.Fatalf("die Antwort: %+v", antwort)
			}

			roh := alsJSON(t, antwort.Paket)
			for _, will := range []string{`"herkunftsart":"eigen"`, `"gebrauchtVon":["stil.mse-style"]`, `"benutzt":2`} {
				if !strings.Contains(string(roh), will) {
					t.Errorf("%s fehlt in %s", will, roh)
				}
			}
		})
	}
}

// **Und scheitert das Nachlesen, bleibt das Geschriebene geschrieben** (§7,
// Regel 3) — aber die Liste ist eine Liste und kein `null`.
func TestOhneUebersichtBleibtDieAntwortEinErfolgMitLeererListe(t *testing.T) {
	for _, weg := range schreibwege {
		t.Run(weg.name, func(t *testing.T) {
			anwendung := app.NeueApp("")
			anwendung.MitPaketen(neuePaketschreibprobe(errors.New("die Liste ist nicht lesbar")))

			antwort := weg.ruf(anwendung, t.Context())
			if antwort.Fehler != nil || antwort.Paket == nil {
				t.Fatalf("die Antwort: %+v", antwort)
			}

			roh := alsJSON(t, antwort.Paket)
			if !strings.Contains(string(roh), `"gebrauchtVon":[]`) {
				t.Errorf("gebrauchtVon ist keine leere Liste: %s", roh)
			}

			// **Und die Herkunftsart steht trotzdem da** — Zweitblick über
			// Phase B, Befund 2: das Fenster kennt nur vier Werte, und ein
			// leerer ist keiner davon. Sie hängt am Paket und am verwalteten
			// Verzeichnis, nicht an der Übersicht.
			if !strings.Contains(string(roh), `"herkunftsart":"eigen"`) {
				t.Errorf("die Herkunftsart fehlt: %s", roh)
			}
		})
	}
}

// **Und was übernommen wurde, auch** — dieselbe Klasse an `VorlagenUebernehmen`
// und `InstallerEinlesen`: jedes übernommene Paket ohne seine Zeile.
func TestEinUebernommenesPaketTraegtSeineZeileDerUebersicht(t *testing.T) {
	probe := neuePaketschreibprobe(nil)
	probe.zusatz.Herkunftsart = domain.PaketUebernommen
	probe.uebernahme = &service.Uebernahme{Uebernommen: []domain.Paket{probe.geschrieben}}

	anwendung := app.NeueApp("")
	anwendung.MitPaketen(probe)

	antwort := anwendung.VorlagenUebernehmen(t.Context(), "/mse2/data")
	if antwort.Fehler != nil || len(antwort.Uebernommen) != 1 {
		t.Fatalf("die Antwort: %+v", antwort)
	}

	roh := alsJSON(t, antwort.Uebernommen[0])
	for _, will := range []string{`"herkunftsart":"uebernommen"`, `"gebrauchtVon":["stil.mse-style"]`} {
		if !strings.Contains(string(roh), will) {
			t.Errorf("%s fehlt in %s", will, roh)
		}
	}
}
