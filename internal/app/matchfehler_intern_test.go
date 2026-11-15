package app

// **Die Fehlerwege des Turniermatches** — Werkstück G7 von Phase 5.5
// (#1464).
//
// **Neben `match_intern_test.go` und nicht darin**: dort steht der Weg, der
// trägt, hier steht, was geschieht, wenn er nicht trägt. Die erste stand mit
// beidem über der Zeilengrenze (§2).

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Jeder Fehlerweg des Matches wird betreten** (§3).
func TestJederFehlerwegDesMatchesWirdBetreten(t *testing.T) {
	tisch := einTurniertisch(t)
	ctx := riegelkontext(t)

	// **Krumme Kennungen** — an allen vier Stellen, an denen eine gedeutet
	// wird.
	krumm := map[string]MatchAntwort{
		"lesen": tisch.anwendung.TurnierMatch(ctx, "keine Zahl"),
		"binden ohne Turnier": tisch.anwendung.TurnierMatchBinden(
			ctx, "keine Zahl", tisch.paarungID, tisch.partieID,
			tisch.zuordnung),
		"binden ohne Paarung": tisch.anwendung.TurnierMatchBinden(
			ctx, tisch.turnierID, "keine Zahl", tisch.partieID,
			tisch.zuordnung),
		"binden mit krummem Teilnehmer": tisch.anwendung.TurnierMatchBinden(
			ctx, tisch.turnierID, tisch.paarungID, tisch.partieID,
			[]MatchsitzDTO{{TeilnehmerID: "keine Zahl", Sitz: "anna"}}),
		"melden": tisch.anwendung.TurnierMatchMelden(
			ctx, "keine Zahl", "anna", "glied", ErgebnisDTO{SpieleA: 1}),
		"übernehmen": tisch.anwendung.TurnierMatchUebernehmen(
			ctx, "keine Zahl", tisch.paarungID),
	}

	for name, antwort := range krumm {
		if antwort.Fehler == nil {
			t.Errorf("%s: die krumme Kennung kam durch", name)
		}

		if antwort.Sitze == nil || antwort.Meldungen == nil {
			t.Errorf("%s: die Listen sind nil statt leer", name)
		}
	}

	// **Melden ohne Bindung.**
	if antwort := tisch.anwendung.TurnierMatchMelden(
		ctx, tisch.paarungID, "anna", "glied", ErgebnisDTO{SpieleA: 1},
	); antwort.Fehler == nil {
		t.Error("die Meldung ohne Bindung kam durch")
	}

	// **Übernehmen ohne Bindung.**
	if antwort := tisch.anwendung.TurnierMatchUebernehmen(
		ctx, tisch.turnierID, tisch.paarungID); antwort.Fehler == nil {
		t.Error("die Übernahme ohne Bindung kam durch")
	}

	// **Binden auf eine Partie, die es nicht gibt.**
	if antwort := tisch.anwendung.TurnierMatchBinden(ctx, tisch.turnierID,
		tisch.paarungID, "gibtesnicht", tisch.zuordnung,
	); antwort.Fehler == nil {
		t.Error("die fehlende Partie kam durch")
	}
}

// **Eine tote Datenbank meldet sich beim Lesen.**
func TestEineToteDatenbankMeldetSichBeimMatch(t *testing.T) {
	tisch := einTurniertisch(t)
	ctx := riegelkontext(t)

	if antwort := tisch.anwendung.TurnierMatchBinden(ctx, tisch.turnierID,
		tisch.paarungID, tisch.partieID, tisch.zuordnung,
	); antwort.Fehler != nil {
		t.Fatalf("binden: %+v", antwort.Fehler)
	}

	if fehler := tisch.anwendung.db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if antwort := tisch.anwendung.TurnierMatch(
		ctx, tisch.paarungID); antwort.Fehler == nil {
		t.Fatal("die geschlossene Datenbank kam nicht an")
	}
}

// **Ein fremder Sitz meldet nicht** — die Bindung steht, die Meldung passt
// nicht.
func TestEinFremderSitzMeldetNicht(t *testing.T) {
	tisch := einTurniertisch(t)
	ctx := riegelkontext(t)

	if antwort := tisch.anwendung.TurnierMatchBinden(ctx, tisch.turnierID,
		tisch.paarungID, tisch.partieID, tisch.zuordnung,
	); antwort.Fehler != nil {
		t.Fatalf("binden: %+v", antwort.Fehler)
	}

	antwort := tisch.anwendung.TurnierMatchMelden(
		ctx, tisch.paarungID, "cem", "glied", ErgebnisDTO{SpieleA: 2})

	if antwort.Fehler == nil {
		t.Fatal("der fremde Sitz kam durch")
	}
}

// **Eine klagende Meldungsliste lässt eine leere Antwort zurück.**
//
// **Über ein Doppel**, weil der echte Weg sie nicht hergibt: wer eine Bindung
// lesen kann, kann auch ihre Meldungen lesen. Der Zweig steht trotzdem da —
// und ein Zweig, den kein Lauf betritt, ist eine Zusage ohne Prüfung (§3).
func TestEineKlagendeMeldungslisteAnDerFassade(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitMatchnaht(&meldungsklage{})

	antwort := anwendung.TurnierMatch(t.Context(), "7")

	switch {
	case antwort.Fehler == nil:
		t.Fatal("die Klage kam nicht an")
	case antwort.Sitze == nil || antwort.Meldungen == nil:
		t.Error("die Listen sind nil statt leer")
	}
}

// meldungsklage gibt eine Bindung her und klagt bei den Meldungen.
type meldungsklage struct{}

func (meldungsklage) Binden(
	context.Context, int64, int64, string, []service.Sitzzuordnung,
) (domain.Matchbindung, error) {
	return domain.Matchbindung{}, nil
}

func (meldungsklage) Melden(context.Context, domain.Matchmeldung) error {
	return nil
}

func (meldungsklage) AusDerPartie(context.Context, int64, string) error {
	return nil
}

func (meldungsklage) Uebernehmen(context.Context, int64, int64) (
	domain.Ergebnis, error,
) {
	return domain.Ergebnis{}, nil
}

func (meldungsklage) Bindung(context.Context, int64) (
	domain.Matchbindung, error,
) {
	return domain.Matchbindung{PaarungID: 7}, nil
}

func (meldungsklage) Meldungen(context.Context, int64) (
	[]domain.Matchmeldung, error,
) {
	return nil, domain.MitGrund(domain.ErrConflict, nil, "die Probe klagt")
}

// **Eine klagende Naht lässt auch beim Melden aus der Partie eine leere
// Antwort zurück** — M17 (#1618).
func TestEineKlagendeNahtBeimMeldenAusDerPartie(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitMatchnaht(&ausgangsklage{})

	antwort := anwendung.TurnierMatchAusDerPartie(t.Context(), "7", "anna")

	switch {
	case antwort.Fehler == nil:
		t.Fatal("die Klage kam nicht an")
	case antwort.Sitze == nil || antwort.Meldungen == nil:
		t.Error("die Listen sind nil statt leer")
	}
}

// ausgangsklage klagt **nur** beim Melden aus der Partie.
type ausgangsklage struct{ meldungsklage }

func (ausgangsklage) AusDerPartie(context.Context, int64, string) error {
	return domain.MitGrund(domain.ErrConflict, nil, "die Probe klagt")
}
