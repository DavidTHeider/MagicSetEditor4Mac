package cardtrader_test

// **Die Wege des Klienten, die kein Erfolgsfall betritt** — A4, #1443.
//
// # Warum sie eine eigene Datei haben
//
// **Weil sie eine andere Frage beantworten.** Die übrigen Läufe fragen, ob
// eine Antwort richtig gelesen wird; diese, was geschieht, wenn keine kommt.

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/cardtrader"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Ohne Adresse gilt die Vorgabe des Pakets.**
//
// **Die Gegenprobe steht daneben**: ohne diesen Lauf wäre die Vorgabe eine
// Zeile, die niemand betritt — und niemand sähe, wenn sie falsch wäre.
func TestOhneAdresseGiltDieVorgabe(t *testing.T) {
	t.Parallel()

	klient := cardtrader.NeuerKlient("", netz.NeuerKlient(netz.UeblicheFristen()),
		cardtrader.Zugang{Marke: "geheim"})

	// **Gefragt wird nicht** — geprüft wird die Adresse über die Meldung, die
	// ein unbekannter Umfang erzeugt, ohne dass eine Anfrage hinausgeht.
	_, _, fehler := klient.Blueprints(t.Context(), 0)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Eine Drossel wartet vor der Anfrage.**
func TestDieDrosselWartetVorDerAnfrage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			_, _ = schreiben.Write([]byte("[]")) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	var gewartet time.Duration

	drossel := netz.NeueDrossel(cardtrader.Abstand, time.Now,
		func(dauer time.Duration) { gewartet += dauer })

	klient := cardtrader.NeuerKlient(server.URL,
		netz.KlientUeber(server.Client()), cardtrader.Zugang{Marke: "geheim"}).
		MitDrossel(drossel)

	for range 2 {
		if _, fehler := klient.Spiele(t.Context()); fehler != nil {
			t.Fatalf("Spiele: %v", fehler)
		}
	}

	if gewartet <= 0 {
		t.Error("zwischen zwei Anfragen wurde nicht gewartet")
	}
}

// **Ein Status, der nicht 200 ist, ist ein Fehler und keine leere Liste.**
func TestEinFremderStatusIstEinFehler(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusUnauthorized)
		}))
	t.Cleanup(server.Close)

	_, fehler := klientFuer(server).Spiele(t.Context())
	if fehler == nil || !strings.Contains(fehler.Error(), "401") {
		t.Fatalf("der Fehler ist %v, erwartet einen mit 401", fehler)
	}
}

// **Eine Adresse, die sich nicht bauen lässt, ist ein Fehler.**
//
// **Ein Steuerzeichen in der Grundadresse** — `url.JoinPath` weist es ab. Der
// Fall ist nicht erfunden: die Adresse kommt aus einer Umgebungsvariablen
// ([app.CardTraderUmgebung]), und dort steht, was jemand hineinschreibt.
func TestEineUnbaubareAdresseIstEinFehler(t *testing.T) {
	t.Parallel()

	klient := cardtrader.NeuerKlient("://kaputt\x7f",
		netz.NeuerKlient(netz.UeblicheFristen()), cardtrader.Zugang{Marke: "geheim"})

	if _, fehler := klient.Spiele(t.Context()); fehler == nil {
		t.Fatal("eine unbaubare Adresse kam als Erfolg zurück")
	}
}

// **Ein Fehler der Erweiterungen reist weiter.**
func TestEinFehlerDerErweiterungenReistWeiter(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusInternalServerError)
		}))
	t.Cleanup(server.Close)

	if _, fehler := klientFuer(server).Erweiterungen(t.Context()); fehler == nil {
		t.Fatal("ein 500 kam als Erfolg zurück")
	}
}

// **Nach `etched` wird nicht gefiltert.**
//
// CardTrader kennt `mtg_foil` als Ja-Nein und kein `etched` (0095). Ein
// `foil=false` wäre falsch — geätzte Karten sind Foils —, ein `foil=true`
// ebenso ungenau. Gefragt wird deshalb nach allem.
func TestNachEtchedWirdNichtGefiltert(t *testing.T) {
	t.Parallel()

	gesehen := make(chan string, 1)

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			gesehen <- gestellt.URL.RawQuery
			_, _ = schreiben.Write([]byte("{}")) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	if _, _, fehler := klientFuer(server).Angebote(t.Context(), 10050,
		cardtrader.Angebotsfilter{Finish: domain.FinishEtched}); fehler != nil {
		t.Fatalf("Angebote: %v", fehler)
	}

	if abfrage := <-gesehen; abfrage != "blueprint_id=10050" {
		t.Errorf("die Abfrage ist %q — nach etched wurde gefiltert", abfrage)
	}
}

// **Und eine kaputte Marktantwort ist ein Fehler.**
func TestEineKaputteMarktantwortIstEinFehler(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			_, _ = schreiben.Write([]byte("[]")) //nolint:errcheck // ein Array, kein Objekt
		}))
	t.Cleanup(server.Close)

	if _, _, fehler := klientFuer(server).Angebote(t.Context(), 10050,
		cardtrader.Angebotsfilter{}); fehler == nil {
		t.Fatal("ein Array kam als Marktantwort durch")
	}
}

// **Und eine Gegenstelle, die nicht antwortet, ist ein Fehler.**
//
// **Nicht dasselbe wie eine unbaubare Adresse**: dort scheitert das Bauen,
// hier der Weg. Ohne diesen Lauf bliebe der Zweig unbetreten, in dem ein
// Verbindungsfehler weiterreist — und das ist der häufigste von allen.
func TestEineSchweigendeGegenstelleIstEinFehler(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(http.ResponseWriter, *http.Request) {}))

	adresse := server.URL

	// **Zugesperrt, bevor gefragt wird** — der Port ist dann frei, und der
	// Verbindungsversuch scheitert sofort statt in eine Frist zu laufen.
	server.Close()

	klient := cardtrader.NeuerKlient(adresse,
		netz.NeuerKlient(netz.UeblicheFristen()), cardtrader.Zugang{Marke: "geheim"})

	if _, fehler := klient.Spiele(t.Context()); fehler == nil {
		t.Fatal("eine geschlossene Gegenstelle kam als Erfolg zurück")
	}
}
