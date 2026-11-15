package app

// **Die Wege von `Fremdkartenbild` und des Malers am Wirt** — P9
// (#2377).
// Der ganze Weg über zwei Anwendungen steht in `partiegastbild_intern_test.go`;
// hier jede Abzweigung für sich: der Grund des Wirts, ein Bild, das keines
// ist, die Frist, kein Tisch, ein Maler, der scheitert.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

// sitzenderGast gibt eine Anwendung, die an einem fremden Tisch sitzt.
func sitzenderGast(t *testing.T) *App {
	t.Helper()

	anwendung := freigabeanwendung(t)
	anwendung.spielbilder = vorschauweg.NeuesTexturfach(1 << 20)
	anwendung.beitrittMerken(eineAngenommene(t, domain.ZweckPartie), "p1")

	return anwendung
}

// mitAntwort ruft das Bild und liefert die Antwort des Wirts, bis es zurück
// ist — eine Antwort vor dem Wunsch fiele ins Leere, also immer wieder.
func mitAntwort(t *testing.T, anwendung *App, antwort domain.Bildantwort) TexturAntwort {
	t.Helper()

	ergebnis := make(chan TexturAntwort, 1)
	go func() {
		ergebnis <- anwendung.Fremdkartenbild(t.Context(), "p1", "o1", "b1", "vorne", "tisch")
	}()

	inhalt, fehler := domain.VerpackeKontrolle(domain.KontrolleBild, antwort)
	if fehler != nil {
		t.Fatalf("verpacken: %v", fehler)
	}

	nachricht := domain.Partienachricht{
		Art: domain.ArtKontrolle, Partie: "p1", Epoche: 1, Folge: 1, Von: "wirt", Inhalt: inhalt,
	}

	for {
		select {
		case fertig := <-ergebnis:
			return fertig
		case <-time.After(5 * time.Millisecond):
			gast, _, gefragt := anwendung.gastAmTisch(t.Context())
			if gefragt != nil {
				t.Fatalf("gastAmTisch: %v", gefragt)
			}

			_ = gast.Angekommen(t.Context(), nachricht) //nolint:errcheck // geprüft wird die Antwort der Bindung
		}
	}
}

var erfragt = domain.Bildwunsch{Objekt: "o1", Seite: "vorne", Guete: "tisch", Bild: "b1"}

func TestEinGrundDesWirtsStehtAlsBefund(t *testing.T) {
	t.Parallel()

	antwort := mitAntwort(t, sitzenderGast(t), domain.Bildantwort{Wunsch: erfragt, Grund: "verdeckt"})

	if antwort.Fehler != nil || antwort.Textur.Adresse != "" ||
		len(antwort.Textur.Befunde) != 1 || antwort.Textur.Befunde[0] != "verdeckt" {
		t.Errorf("die Antwort: %+v", antwort)
	}
}

func TestEinBildDasKeinesIstWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]domain.Bildantwort{
		"kein PNG":     {Wunsch: erfragt, Png: []byte("<svg/>"), Breite: 2, Hoehe: 3},
		"falsche Maße": {Wunsch: erfragt, Png: einKleinesPNG(t), Breite: 20, Hoehe: 30},
	}

	for name, bild := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anwendung := sitzenderGast(t)
			antwort := mitAntwort(t, anwendung, bild)

			if antwort.Fehler == nil || antwort.Fehler.Code != CodeValidierung {
				t.Errorf("die Antwort: %+v", antwort)
			}

			if anwendung.spielbilder.Belegt() != 0 {
				t.Error("das falsche Bild liegt im Fach")
			}
		})
	}
}

func TestOhneAntwortDesWirtsKommtNachDerFristEinFehler(t *testing.T) {
	t.Parallel()

	ctx, fertig := context.WithTimeout(t.Context(), 30*time.Millisecond)
	defer fertig()

	antwort := sitzenderGast(t).Fremdkartenbild(ctx, "p1", "o1", "b1", "vorne", "tisch")
	if antwort.Fehler == nil || antwort.Fehler.Code != CodeKonflikt {
		t.Errorf("die Antwort: %+v", antwort)
	}
}

func TestFremdkartenbildWeistAbWasNichtGeht(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		anwendung            func() *App
		partie, objekt, bild string
		code                 Fehlercode
	}{
		"an keinem fremden Tisch": {func() *App {
			a := freigabeanwendung(t)
			a.spielbilder = vorschauweg.NeuesTexturfach(1 << 20)

			return a
		}, "p1", "o1", "b1", CodeKonflikt},
		"ohne Objekt":      {func() *App { return sitzenderGast(t) }, "p1", "", "b1", CodeValidierung},
		"ohne Bildkennung": {func() *App { return sitzenderGast(t) }, "p1", "o1", "", CodeValidierung},
		// **An einem anderen Tisch** (Zweitblick über Phase 7, Befund 4):
		// der Gast sitzt bei p1, das Bild gehört nach p2 — gefragt wird
		// niemand, sonst legte der Wirt von p1 sein Bild unter die
		// Kennung des anderen Tisches.
		"an einem anderen Tisch": {func() *App { return sitzenderGast(t) }, "p2", "o1", "b1", CodeKonflikt},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// **Ohne zu fragen** — wer fragte, wartete die Bildfrist ab und
			// bekäme danach denselben Code.
			vorher := time.Now()

			antwort := fall.anwendung().Fremdkartenbild(t.Context(), fall.partie, fall.objekt, fall.bild, "vorne", "tisch")
			if antwort.Fehler == nil || antwort.Fehler.Code != fall.code {
				t.Errorf("die Antwort: %+v", antwort)
			}

			if gewartet := time.Since(vorher); gewartet > Bildfrist/2 {
				t.Errorf("gefragt und %v gewartet", gewartet)
			}
		})
	}
}

// **Der Maler am Wirt gibt jeden Fehler durch** — der Wirt macht daraus
// einen Grund für den Gast.
func TestDerWirtsmalerGibtSeineFehlerDurch(t *testing.T) {
	t.Parallel()

	klage := errors.New("kaputt")

	faelle := map[string]func(*App){
		"die Karte ist nicht zu lesen": func(a *App) { a.kartenleser = &texturkartenleser{klagt: klage} },
		"das Werk zeichnet nicht":      func(a *App) { a.MitTexturwerk(&texturprobe{klagt: klage}) },
		"das Bild liegt nicht im Fach": func(*App) {},
	}

	for name, richte := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anwendung := NeueApp("")
			anwendung.kartenleser = &texturkartenleser{}
			anwendung.spielbilder = vorschauweg.NeuesTexturfach(1 << 20)
			anwendung.MitTexturwerk(&texturprobe{})
			richte(anwendung)

			if _, fehler := (eigenmaler{anwendung}).Kartenbild(t.Context(), 7, "vorne", "tisch"); fehler == nil {
				t.Error("der Maler kam durch")
			}
		})
	}
}

// **Und mit Bild gibt er Bytes und Maße.**
func TestDerWirtsmalerGibtDasBildAusDemFach(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.kartenleser = &texturkartenleser{}
	anwendung.spielbilder = vorschauweg.NeuesTexturfach(1 << 20)
	anwendung.MitTexturwerk(&bildtexturen{fach: anwendung.spielbilder, png: einKleinesPNG(t)})

	bild, fehler := (eigenmaler{anwendung}).Kartenbild(t.Context(), 7, "vorne", "tisch")
	if fehler != nil || len(bild.Png) == 0 || bild.Breite != 2 || bild.Hoehe != 3 {
		t.Errorf("das Bild: %d Bytes, %d×%d, %v", len(bild.Png), bild.Breite, bild.Hoehe, fehler)
	}

	var _ service.Kartenbildmaler = eigenmaler{}
}
