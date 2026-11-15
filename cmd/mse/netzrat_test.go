package main

// **Ein fremder Dienst, der nicht antwortet, bekommt seinen Rat** — D02
// (#2233).
//
// Dieselbe Gestalt wie beim belegten Bestand: der Fehler sagt, **was** war —
// welcher Fall, welcher Wirt, welche Frist —, der Rat, **was zu tun ist**, und
// der steht nicht im Fehler. Die Sätze entsprechen denen des deutschen
// Katalogs (`frontend/src/sprache/de/fehler.ts`).

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinFremderDienstZeigtSeinenRat(t *testing.T) {
	t.Parallel()

	grund := errors.New("dial tcp: lookup api.scryfall.com: no such host")

	for _, fall := range []struct {
		name   string
		fehler error
		rat    string
	}{
		{"Name", domain.NeuerGegenstellenfehler(domain.ErrNameUnaufloesbar, grund, "api.scryfall.com", 0), "Prüfe die Internetverbindung"},
		{"Port", domain.NeuerGegenstellenfehler(domain.ErrNichtErreichbar, grund, "127.0.0.1:9", 0), "Läuft der Dienst"},
		{"Kopf", domain.NeuerGegenstellenfehler(domain.ErrKeineAntwort, grund, "api.scryfall.com", 30*time.Second), "später noch einmal"},
		{"Rumpf", domain.NeuerGegenstellenfehler(domain.ErrAntwortStockt, grund, "mtgjson.com", time.Minute), "Prüfe die Verbindung"},
		{"Abriss", domain.NeuerGegenstellenfehler(domain.ErrVerbindungAbgerissen, grund, "api.scryfall.com", 0), "liegt es am Dienst"},
		{"Gegenüber", domain.MitGrund(domain.ErrPeerNichtErreichbar, grund, "die Gegenstelle ist nicht erreichbar"), "mse kontakt-adresse"},
		{"ohne Adresse", fmt.Errorf("%w (8749-8908)", domain.ErrPeerOhneAdresse), "Trage die Adresse mit mse kontakt-adresse ein"},
		{"Proxy", domain.NeuerGegenstellenfehler(domain.ErrProxyAnmeldung, grund, "proxy.firma:3128", 0), "HTTPS_PROXY"},
		{"429", fmt.Errorf("%w: https://api.scryfall.com/cards", domain.ErrGedrosselt), "Warte einen Moment"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			var raus, rfehler bytes.Buffer

			an := &schreiber{ziel: &raus, fehler: &rfehler}
			an.klageUeber(fmt.Errorf("Scryfall fragen: %w", fall.fehler))

			gesagt := rfehler.String()
			if !strings.Contains(gesagt, fall.fehler.Error()) || !strings.Contains(gesagt, fall.rat) {
				t.Errorf("die Ausgabe nennt Fehler oder Rat nicht: %q", gesagt)
			}

			if strings.Contains(gesagt, "no such host") {
				t.Errorf("der Grund steht in der Ausgabe: %q", gesagt)
			}
		})
	}
}
