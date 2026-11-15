package scryfall_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// bildserver stellt ein Bild bereit und zählt die Abrufe.
func bildserver(t *testing.T, rumpf []byte, status int) (server *httptest.Server, abrufe *int) {
	t.Helper()

	gezaehlt := 0
	gestellt := httptest.NewServer(http.HandlerFunc(
		func(schreiber http.ResponseWriter, anfrage *http.Request) {
			gezaehlt++

			if anfrage.Header.Get("User-Agent") == "" {
				t.Error("das Bild wurde ohne User-Agent geholt")
			}

			schreiber.WriteHeader(status)
			if _, fehler := schreiber.Write(rumpf); fehler != nil {
				t.Errorf("das Bild ließ sich nicht schreiben: %v", fehler)
			}
		}))
	t.Cleanup(gestellt.Close)

	return gestellt, &gezaehlt
}

func TestKlientBild(t *testing.T) {
	daten := []byte("\x89PNG\r\n\x1a\n und noch etwas")

	t.Run("holt die ganze Karte samt Nennung", func(t *testing.T) {
		server, abrufe := bildserver(t, daten, http.StatusOK)
		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		karte := &scryfall.Karte{
			Name: "Lightning Bolt", Kuenstler: "Christopher Rush",
			Bilder:  scryfall.Bilder{Normal: server.URL + "/normal.jpg"},
			Adresse: "https://scryfall.com/card/lea/161",
		}

		fund, fehler := klient.Bild(t.Context(), karte, scryfall.BildNormal)
		if fehler != nil {
			t.Fatalf("Bild: %v", fehler)
		}

		if !bytes.Equal(fund.Daten, daten) {
			t.Errorf("die Bytes sind verändert angekommen: %q", fund.Daten)
		}

		if *abrufe != 1 {
			t.Errorf("Abrufe = %d, erwartet 1", *abrufe)
		}

		if !strings.Contains(fund.Nennung, "Christopher Rush") {
			t.Errorf("die Nennung führt den Künstler nicht: %q", fund.Nennung)
		}

		if !strings.Contains(fund.Nennung, "Wizards of the Coast") {
			t.Errorf("die Nennung führt das Copyright nicht: %q", fund.Nennung)
		}

		if fund.Adresse != karte.Adresse {
			t.Errorf("Adresse = %q, erwartet die Kartenseite", fund.Adresse)
		}
	})

	t.Run("weist den Ausschnitt ohne Künstlernamen ab", func(t *testing.T) {
		server, abrufe := bildserver(t, daten, http.StatusOK)
		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		karte := &scryfall.Karte{
			Name:   "Ohne Nennung",
			Bilder: scryfall.Bilder{Ausschnitt: server.URL + "/crop.jpg"},
		}

		_, fehler := klient.Bild(t.Context(), karte, scryfall.BildAusschnitt)
		if !errors.Is(fehler, domain.ErrValidation) {
			t.Fatalf("Fehler = %v, erwartet ErrValidation", fehler)
		}

		// **Der Kern der Wache**: es wurde gar nicht erst geholt. Ein Bild, das
		// im Speicher liegt, findet später jemand und benutzt es.
		if *abrufe != 0 {
			t.Errorf("Abrufe = %d — das Bild wurde trotz der Wache geholt", *abrufe)
		}
	})

	t.Run("lässt den Ausschnitt mit Künstlernamen durch", func(t *testing.T) {
		server, _ := bildserver(t, daten, http.StatusOK)
		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		karte := &scryfall.Karte{
			Name: "Mit Nennung", Kuenstler: "Rebecca Guay",
			Bilder: scryfall.Bilder{Ausschnitt: server.URL + "/crop.jpg"},
		}

		fund, fehler := klient.Bild(t.Context(), karte, scryfall.BildAusschnitt)
		if fehler != nil {
			t.Fatalf("Bild: %v", fehler)
		}

		if fund.Art != scryfall.BildAusschnitt {
			t.Errorf("Art = %q", fund.Art)
		}

		if !strings.Contains(fund.Nennung, "Rebecca Guay") {
			t.Errorf("Nennung = %q", fund.Nennung)
		}
	})

	t.Run("findet das Bild einer zweiseitigen Karte auf der ersten Seite",
		func(t *testing.T) {
			server, abrufe := bildserver(t, daten, http.StatusOK)
			klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

			karte := &scryfall.Karte{
				Name: "Delver of Secrets // Insectile Aberration",
				Seiten: []scryfall.Seite{{
					Name: "Delver of Secrets", Kuenstler: "Nils Hamm",
					Bilder: scryfall.Bilder{Normal: server.URL + "/vorne.jpg"},
				}},
			}

			fund, fehler := klient.Bild(t.Context(), karte, scryfall.BildNormal)
			if fehler != nil {
				t.Fatalf("Bild: %v", fehler)
			}

			if *abrufe != 1 || len(fund.Daten) == 0 {
				t.Errorf("Abrufe = %d, Bytes = %d", *abrufe, len(fund.Daten))
			}
		})

	t.Run("meldet eine fehlende Art als nicht gefunden", func(t *testing.T) {
		klient := klientFuer(t, "http://unerreichbar.invalid", nil)

		karte := &scryfall.Karte{Name: "Ohne Bilder"}

		_, fehler := klient.Bild(t.Context(), karte, scryfall.BildKlein)
		if !errors.Is(fehler, domain.ErrNotFound) {
			t.Fatalf("Fehler = %v, erwartet ErrNotFound", fehler)
		}

		if !strings.Contains(fehler.Error(), "Ohne Bilder") {
			t.Errorf("der Fehler nennt die Karte nicht: %v", fehler)
		}
	})

	t.Run("weist eine unbekannte Art ab", func(t *testing.T) {
		klient := klientFuer(t, "http://unerreichbar.invalid", nil)

		karte := &scryfall.Karte{
			Name:   "Mit allen dreien",
			Bilder: scryfall.Bilder{Klein: "a", Normal: "b", Ausschnitt: "c"},
		}

		_, fehler := klient.Bild(t.Context(), karte, scryfall.Bildart("riesig"))
		if !errors.Is(fehler, domain.ErrNotFound) {
			t.Fatalf("Fehler = %v, erwartet ErrNotFound", fehler)
		}
	})

	t.Run("weist eine fehlende Karte ab", func(t *testing.T) {
		klient := klientFuer(t, "http://unerreichbar.invalid", nil)

		_, fehler := klient.Bild(t.Context(), nil, scryfall.BildNormal)
		if !errors.Is(fehler, domain.ErrValidation) {
			t.Fatalf("Fehler = %v, erwartet ErrValidation", fehler)
		}
	})

	t.Run("meldet einen Fehlerstatus des Bildservers", func(t *testing.T) {
		server, _ := bildserver(t, nil, http.StatusInternalServerError)
		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		karte := &scryfall.Karte{
			Name: "Kaputt", Bilder: scryfall.Bilder{Normal: server.URL + "/weg.jpg"},
		}

		_, fehler := klient.Bild(t.Context(), karte, scryfall.BildNormal)
		if !errors.Is(fehler, domain.ErrValidation) {
			t.Fatalf("Fehler = %v, erwartet ErrValidation", fehler)
		}

		if !strings.Contains(fehler.Error(), "500") {
			t.Errorf("der Fehler nennt den Status nicht: %v", fehler)
		}
	})

	t.Run("meldet einen abgebrochenen Kontext", func(t *testing.T) {
		server, _ := bildserver(t, daten, http.StatusOK)
		klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

		ctx, abbrechen := context.WithCancel(t.Context())
		abbrechen()

		karte := &scryfall.Karte{
			Name:   "Abgebrochen",
			Bilder: scryfall.Bilder{Normal: server.URL + "/nie.jpg"},
		}

		if _, fehler := klient.Bild(ctx, karte, scryfall.BildNormal); fehler == nil {
			t.Fatal("der Abbruch blieb ohne Fehler")
		}
	})
}
