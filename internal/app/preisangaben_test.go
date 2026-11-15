package app_test

// **Was ein Preisstand am Fenster über sich sagt** — Werkstück A3 von Phase
// 5.3 (#1442).
//
// Aus `preise_test.go` herausgelöst, als es über 400 Zeilen ging (§2). Der
// Schnitt folgt der Frage: dort steht, **welche Bindung was tut**, hier, was
// eine einzelne Angabe bedeutet — der Tag der Quelle und der Preis, den
// niemand genannt hat.

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Tag, den die Quelle nennt, kommt bis ans Fenster** — A3 (#1442).
//
// MTGJSON führt seine Preise unter einem ISO-8601-Datum; ein Abruf von heute
// kann einen Preis von vorgestern bringen. Scryfall nennt keinen — und dann
// steht dort **nichts** statt eines erfundenen Tages.
func TestDasBeobachtungsdatumErreichtDasFenster(t *testing.T) {
	t.Parallel()

	beobachtet := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC)

	mitTag := probestand(417, 2)
	mitTag.Quelle = "mtgjson"
	mitTag.Beobachtet = beobachtet

	anwendung := app.NeueApp("")
	anwendung.MitUhr(festeUhr())
	anwendung.MitPreisen(preisnahtdoppel{
		staende: []domain.Preisstand{mitTag, probestand(400, 2)},
	})

	antwort := anwendung.Preise(t.Context(), "abc")
	if antwort.Fehler != nil {
		t.Fatalf("Preise: %+v", antwort.Fehler)
	}

	if len(antwort.Staende) != 2 {
		t.Fatalf("%d Stände", len(antwort.Staende))
	}

	if antwort.Staende[0].Beobachtet != "2026-08-25T00:00:00Z" {
		t.Errorf("beobachtet ist %q", antwort.Staende[0].Beobachtet)
	}

	// **Die Gegenprobe**: ohne genannten Tag steht dort nichts.
	if antwort.Staende[1].Beobachtet != "" {
		t.Errorf("ohne Tag steht %q", antwort.Staende[1].Beobachtet)
	}
}

// **Ein Stand, den niemand genannt hat, trägt keinen Preis** — A3 (#1442).
//
// Eine `"0.00"` sähe aus wie „umsonst"; bei einer Lage, die nicht `genannt`
// ist, steht null in der Zeile und bedeutet nichts.
func TestEinNichtGenannterStandZeigtKeinenWert(t *testing.T) {
	t.Parallel()

	weggefallen := probestand(0, 1)
	weggefallen.Lage = domain.PreisWeggefallen

	anwendung := app.NeueApp("")
	anwendung.MitUhr(festeUhr())
	anwendung.MitPreisen(preisnahtdoppel{
		staende: []domain.Preisstand{weggefallen, probestand(417, 1)},
	})

	antwort := anwendung.Preise(t.Context(), "abc")

	if antwort.Staende[0].Wert != "" || antwort.Staende[0].Lage != "weggefallen" {
		t.Errorf("der weggefallene Stand ist %+v", antwort.Staende[0])
	}

	// **Die Gegenprobe**: ein genannter trägt seinen Preis.
	if antwort.Staende[1].Wert != "4.17" {
		t.Errorf("der genannte Stand ist %+v", antwort.Staende[1])
	}
}
