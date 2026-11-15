package app

// **Der Verlauf an der Naht** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
// **Dass diese Fassade nicht noch einmal filtert** — und dass sie auch nichts
// durchlässt, was der Dienst schon gefiltert hat. Sie bildet ab, mehr nicht;
// wer was sehen darf, entscheidet `service` über `domain.RueckschauFuer`.
//
// **Und dass durch die Naht keine `null` geht.** Eine Oberfläche, die über
// `null` liefe, bekäme einen Fehler statt einer leeren Anzeige — das ist die
// Klasse von Fehlern, die §3 unter „zwischen den Prüfebenen" führt.

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// verlaufsleitung gibt einen festen Strom — genau so, wie `service` ihn
// gefiltert hätte.
type verlaufsleitung struct {
	schweigsameLeitung

	strom []domain.Ereignis
	klage error
	fuer  string
}

func (v *verlaufsleitung) Verlauf(
	_ context.Context, _, fuer string,
) ([]domain.Ereignis, error) {
	v.fuer = fuer

	if v.klage != nil {
		return nil, v.klage
	}

	return v.strom, nil
}

func verlaufsanwendung(t *testing.T, leitung Partieleitung) *App {
	t.Helper()

	anwendung := NeueApp("")
	anwendung.db = bestandFuerDenRiegel(t)
	anwendung.profil.Store(1)

	return anwendung.MitPartienaht(leitung)
}

func TestDerVerlaufKommtAbgebildetAn(t *testing.T) {
	t.Parallel()

	leitung := &verlaufsleitung{strom: []domain.Ereignis{{
		Folge: 7, Art: "objekt.ziehen",
		Oeffentlich: map[string]any{"sitz": "a"},
		Privat: []domain.Ergaenzung{
			{Empfaenger: "a", Inhalt: map[string]any{"karte": "Wald"}},
		},
		Kette: "glied",
	}}}

	antwort := verlaufsanwendung(t, leitung).PartieVerlauf(
		t.Context(), "p1", "a")

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %v", antwort.Fehler)
	case len(antwort.Verlauf) != 1:
		t.Fatalf("%d Ereignisse, erwartet 1", len(antwort.Verlauf))
	case antwort.Verlauf[0].Folge != "7":
		t.Errorf("Folge %q", antwort.Verlauf[0].Folge)
	case antwort.Verlauf[0].Kette != "glied":
		t.Errorf("Kette %q", antwort.Verlauf[0].Kette)
	case antwort.Verlauf[0].Privat["karte"] != "Wald":
		t.Errorf("privat %v", antwort.Verlauf[0].Privat)
	}
}

// **Die Sicht wird durchgereicht und nicht abgeleitet.**
//
// Wer sie aus dem Profil nähme, ließe beim Self-Play einen Sitz den Verlauf
// des anderen sehen — und dort ist dasselbe Profil der Normalfall (G8).
func TestDieSichtWirdDurchgereicht(t *testing.T) {
	t.Parallel()

	leitung := &verlaufsleitung{}

	verlaufsanwendung(t, leitung).PartieVerlauf(t.Context(), "p1", "sitz-2")

	if leitung.fuer != "sitz-2" {
		t.Errorf("gefragt wurde für %q, erwartet sitz-2", leitung.fuer)
	}
}

// **Durch die Naht geht keine `null`.**
func TestEinLeererVerlaufGehtAlsListeDurchDieNaht(t *testing.T) {
	t.Parallel()

	antwort := verlaufsanwendung(t, &verlaufsleitung{}).PartieVerlauf(
		t.Context(), "p1", "a")

	roh, fehler := json.Marshal(antwort)
	if fehler != nil {
		t.Fatalf("verpacken: %v", fehler)
	}

	var gelesen struct {
		Verlauf []EreignisDTO `json:"verlauf"`
	}

	if fehler := json.Unmarshal(roh, &gelesen); fehler != nil {
		t.Fatalf("auspacken: %v", fehler)
	}

	if gelesen.Verlauf == nil {
		t.Error("durch die Naht kam null statt einer leeren Liste")
	}
}

// **Und auch kein `null` in einem Ereignis.**
func TestEinEreignisOhneInhaltGehtAlsKarteDurchDieNaht(t *testing.T) {
	t.Parallel()

	leitung := &verlaufsleitung{strom: []domain.Ereignis{{Folge: 1, Art: "x"}}}

	antwort := verlaufsanwendung(t, leitung).PartieVerlauf(
		t.Context(), "p1", "a")

	switch {
	case antwort.Verlauf[0].Oeffentlich == nil:
		t.Error("Oeffentlich ist null")
	case antwort.Verlauf[0].Privat == nil:
		t.Error("Privat ist null")
	}
}

// **Was der Dienst klagt, wird übersetzt** (§8).
func TestEinKlagenderVerlaufWirdUebersetzt(t *testing.T) {
	t.Parallel()

	leitung := &verlaufsleitung{klage: errors.New("der Bestand ist zu")}

	antwort := verlaufsanwendung(t, leitung).PartieVerlauf(
		t.Context(), "p1", "a")

	switch {
	case antwort.Fehler == nil:
		t.Fatal("kein Fehler")
	case antwort.Verlauf == nil:
		t.Error("im Fehlerfall kam null statt einer leeren Liste")
	}
}

// **Auch die Rückschau fragt nach der Rolle** — und zwar lesend.
//
// **Ein eigener Lauf und nicht eine Zeile in `partierolle_intern_test.go`**:
// jene Datei steht an der Zeilengrenze aus §2, und der Schnitt folgt der
// Frage — dort geht es darum, was eine Partie **tut**, hier darum, was sie
// **war**. Der Grund für den Lauf selbst ist derselbe: eine Tabelle, die eine
// Rolle nennt, die nichts durchsetzt, verspricht mehr, als ihr Rumpf hält
// (§10, und N7 in Phase 5.4 war genau das).
func TestDerVerlaufFragtNachDerRolle(t *testing.T) {
	t.Parallel()

	wache := &zaehlendeWache{}
	gebunden := NeueBindungen(
		verlaufsanwendung(t, &verlaufsleitung{}).MitWache(wache))

	antwort := gebunden.PartieVerlauf(t.Context(), "p1", "a")

	switch {
	case antwort.Fehler == nil:
		t.Fatal("der Weg kam durch")
	case antwort.Verlauf == nil:
		t.Error("der Fehlerweg gibt nil statt einer leeren Liste")
	case len(wache.gefragt) == 0 ||
		wache.gefragt[0] != domain.WirkungLesen:
		t.Errorf("gefragt wurde nach %v statt nach %q",
			wache.gefragt, domain.WirkungLesen)
	}
}

// **Und die kopflose Fläche nimmt denselben Weg** (§16).
//
// **Hier und nicht nur in `internal/flaechen`**: dort fahren gebaute
// Programme, und was sie betreten, steht in ihrem eigenen Profil. Derselbe
// Satz steht über `partieflaechenweg_intern_test.go`.
func TestDieKopfloseFlaecheGibtDenVerlauf(t *testing.T) {
	t.Parallel()

	leitung := &verlaufsleitung{strom: []domain.Ereignis{{
		Folge: 3, Art: "objekt.ziehen", Kette: "glied",
	}}}

	antwort := (&Partieflaeche{app: verlaufsanwendung(t, leitung)}).
		PartieVerlauf(t.Context(), "p1", "a")

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %v", antwort.Fehler)
	case len(antwort.Verlauf) != 1:
		t.Fatalf("%d Ereignisse, erwartet 1", len(antwort.Verlauf))
	case antwort.Verlauf[0].Folge != "3":
		t.Errorf("Folge %q", antwort.Verlauf[0].Folge)
	}
}
