package storage_test

// **Der Ausgang** — Werkstück `T5` von Phase 5.6
// (#1470).
//
// # Was hier gemessen wird
//
// **Dass eine Nachricht liegen bleibt, bis sie bestätigt ist.** Ein Ausgang,
// der nach dem Senden vergisst, ist keiner: die Nachricht ist dann weg, und
// niemand weiß, ob sie angekommen ist.
//
// **Dass die Versuche gezählt werden** — ein totes Gegenüber soll nicht
// endlos angerufen werden.
//
// **Und dass zweimal dieselbe Kennung nicht zwei Zeilen ergibt.**
//
// **Kein `t.Parallel()`**, aus dem Grund, der in `sammlung_test.go` steht.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEineNachrichtBleibtLiegenBisSieBestaetigtIst(t *testing.T) {
	ablage := storage.NeueAusgangsablage(frischeDatenbank(t))
	ctx := probekontext(t)

	if fehler := ablage.Einreihen(ctx, storage.Ausgangszeile{
		NachrichtID: "n-1", Vorgang: "v-1",
		Empfaenger: "bgegenueber", Nachricht: `{"art":"vorschlag"}`,
	}); fehler != nil {
		t.Fatalf("Einreihen: %v", fehler)
	}

	offen, fehler := ablage.Offene(ctx, "bgegenueber")
	if fehler != nil {
		t.Fatalf("Offene: %v", fehler)
	}

	if len(offen) != 1 || offen[0].NachrichtID != "n-1" {
		t.Fatalf("offen ist %+v", offen)
	}

	if offen[0].Versuche != 0 {
		t.Errorf("die Zeile trägt %d Versuche", offen[0].Versuche)
	}

	// **Ein Versuch wird gezählt** — und die Zeile bleibt liegen.
	if zaehlfehler := ablage.Versucht(ctx, "n-1"); zaehlfehler != nil {
		t.Fatalf("Versucht: %v", zaehlfehler)
	}

	nachVersuch, fehler := ablage.Offene(ctx, "bgegenueber")
	if fehler != nil {
		t.Fatalf("Offene: %v", fehler)
	}

	if len(nachVersuch) != 1 || nachVersuch[0].Versuche != 1 {
		t.Fatalf("nach dem Versuch ist es %+v", nachVersuch)
	}

	// **Erst die Bestätigung räumt sie weg.**
	if raeumfehler := ablage.Bestaetigen(ctx, "n-1"); raeumfehler != nil {
		t.Fatalf("Bestaetigen: %v", raeumfehler)
	}

	danach, fehler := ablage.Offene(ctx, "bgegenueber")
	if fehler != nil {
		t.Fatalf("Offene: %v", fehler)
	}

	if len(danach) != 0 {
		t.Errorf("es liegt noch %+v", danach)
	}
}

// **Zweimal dieselbe Kennung ergibt nicht zwei Zeilen.**
//
// **Und das ist die Idempotenz des Ausgangs**: wer nach einem Neustart noch
// einmal einreiht, was schon dalag, soll es nicht doppelt schicken.
func TestZweimalDieselbeKennungErgibtEineZeile(t *testing.T) {
	ablage := storage.NeueAusgangsablage(frischeDatenbank(t))
	ctx := probekontext(t)

	zeile := storage.Ausgangszeile{
		NachrichtID: "n-1", Vorgang: "v-1",
		Empfaenger: "bgegenueber", Nachricht: `{"art":"vorschlag"}`,
	}

	if fehler := ablage.Einreihen(ctx, zeile); fehler != nil {
		t.Fatalf("Einreihen: %v", fehler)
	}

	if fehler := ablage.Einreihen(ctx, zeile); fehler != nil {
		t.Fatalf("das zweite Einreihen: %v", fehler)
	}

	offen, fehler := ablage.Offene(ctx, "bgegenueber")
	if fehler != nil {
		t.Fatalf("Offene: %v", fehler)
	}

	if len(offen) != 1 {
		t.Errorf("es liegen %d Zeilen", len(offen))
	}
}

// **Eine Bestätigung ins Leere ist ein Nein** — sonst hielte jemand eine
// Nachricht für zugestellt, die nie dalag.
func TestEineBestaetigungInsLeereIstEinNein(t *testing.T) {
	if fehler := storage.NeueAusgangsablage(frischeDatenbank(t)).
		Bestaetigen(probekontext(t), "gibtesnicht"); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Und ein Versuch ins Leere ebenso.**
func TestEinVersuchInsLeereIstEinNein(t *testing.T) {
	if fehler := storage.NeueAusgangsablage(frischeDatenbank(t)).
		Versucht(probekontext(t), "gibtesnicht"); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Der Ausgang gehört einem Profil** (0066).
func TestDerAusgangGehoertEinemProfil(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueAusgangsablage(db)

	if fehler := ablage.Einreihen(probekontext(t), storage.Ausgangszeile{
		NachrichtID: "n-1", Vorgang: "v-1",
		Empfaenger: "bgegenueber", Nachricht: "{}",
	}); fehler != nil {
		t.Fatalf("Einreihen: %v", fehler)
	}

	offen, fehler := ablage.Offene(storage.MitProfil(t.Context(), 2), "bgegenueber")
	if fehler != nil {
		t.Fatalf("Offene: %v", fehler)
	}

	if len(offen) != 0 {
		t.Errorf("das zweite Profil sah %+v", offen)
	}
}

// **Der Stand zählt und zeigt nichts.**
func TestDerPoststandZaehlt(t *testing.T) {
	ablage := storage.NeueAusgangsablage(frischeDatenbank(t))
	ctx := probekontext(t)

	// **Ohne Zeilen ist er null** — und nicht ein Lesefehler: `max` über null
	// Zeilen gibt `NULL`.
	leer, fehler := ablage.Stand(ctx, "bgegenueber")
	if fehler != nil {
		t.Fatalf("Stand: %v", fehler)
	}

	if leer.Offen != 0 || leer.HoechsteVersuche != 0 {
		t.Fatalf("der leere Stand ist %+v", leer)
	}

	for _, kennung := range []string{"n-1", "n-2"} {
		if reihfehler := ablage.Einreihen(ctx, storage.Ausgangszeile{
			NachrichtID: kennung, Vorgang: "v-1",
			Empfaenger: "bgegenueber", Nachricht: "{}",
		}); reihfehler != nil {
			t.Fatalf("Einreihen: %v", reihfehler)
		}
	}

	if zaehlfehler := ablage.Versucht(ctx, "n-2"); zaehlfehler != nil {
		t.Fatalf("Versucht: %v", zaehlfehler)
	}

	stand, fehler := ablage.Stand(ctx, "bgegenueber")
	if fehler != nil {
		t.Fatalf("Stand: %v", fehler)
	}

	// **Die höchste und nicht die Summe** — eine Zahl, die mit jeder
	// Nachricht wächst, sagt nichts darüber, ob eine davon klemmt.
	if stand.Offen != 2 || stand.HoechsteVersuche != 1 {
		t.Errorf("der Stand ist %+v", stand)
	}
}

// **Ohne Profil gibt es keinen Stand.**
func TestOhneProfilGibtEsKeinenPoststand(t *testing.T) {
	if _, fehler := storage.NeueAusgangsablage(frischeDatenbank(t)).
		Stand(context.Background(), "b"); fehler == nil {
		t.Error("der Stand kam ohne Profil durch")
	}
}
