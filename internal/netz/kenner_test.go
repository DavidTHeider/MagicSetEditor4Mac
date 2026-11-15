package netz_test

// **Wie sich dieses Programm vorstellt** — Phase 4.8
// (#1321).

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Ein Dienst darf mehr als eine Kopfzeile verlangen.**
//
// **Nachgelesen am 2026-09-02**: Scrydex verlangt `X-Api-Key` **und**
// `X-Team-ID` — zwei Angaben, keine davon ein `Authorization`. Ein Kenner mit
// genau drei festen Feldern trüge das nicht, und ein zweiter Weg neben
// [netz.Hole] hätte die Hälfte seiner Zusagen.
func TestEinKennerTraegtZusaetzlicheKopfzeilen(t *testing.T) {
	t.Parallel()

	var gesehen http.Header

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiber http.ResponseWriter, gestellt *http.Request) {
			gesehen = gestellt.Header.Clone()
			schreiber.WriteHeader(http.StatusOK)
		}))
	defer server.Close()

	kenner := netz.Kenner{
		Name:    "MagicSetEditor4Mac",
		Annahme: "application/json",
		Zusatz: map[string]string{
			"X-Api-Key": "geheim",
			"X-Team-ID": "mannschaft",
		},
	}

	if _, err := netz.Hole(t.Context(), netz.KlientUeber(server.Client()), nil,
		kenner, server.URL, 1024); err != nil {
		t.Fatalf("holen: %v", err)
	}

	for kopf, erwartet := range map[string]string{
		"User-Agent": "MagicSetEditor4Mac",
		"Accept":     "application/json",
		"X-Api-Key":  "geheim",
		"X-Team-Id":  "mannschaft",
	} {
		if gelesen := gesehen.Get(kopf); gelesen != erwartet {
			t.Errorf("%s ist %q, erwartet %q", kopf, gelesen, erwartet)
		}
	}
}

// **Ein leerer Zusatz setzt keine Kopfzeile.**
//
// Ein leeres `X-Api-Key` ist etwas anderes als keines — manche Dienste weisen
// es ab, statt die Anfrage als unangemeldet zu behandeln. Dieselbe
// Überlegung wie bei [netz.Kenner.Marke].
func TestEinLeererZusatzSetztKeineKopfzeile(t *testing.T) {
	t.Parallel()

	var gesehen http.Header

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiber http.ResponseWriter, gestellt *http.Request) {
			gesehen = gestellt.Header.Clone()
			schreiber.WriteHeader(http.StatusOK)
		}))
	defer server.Close()

	kenner := netz.Kenner{
		Name:    "probe",
		Annahme: "*/*",
		Zusatz:  map[string]string{"X-Api-Key": ""},
	}

	if _, err := netz.Hole(t.Context(), netz.KlientUeber(server.Client()), nil,
		kenner, server.URL, 1024); err != nil {
		t.Fatalf("holen: %v", err)
	}

	if _, gesetzt := gesehen["X-Api-Key"]; gesetzt {
		t.Error("ein leerer Zusatz wurde als Kopfzeile gesetzt")
	}
}
