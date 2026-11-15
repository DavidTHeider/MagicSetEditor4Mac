package app

// **Das Turniermatch an der Fassade** — Werkstück G7 von Phase 5.5
// (#1464).
//
// # Wogegen hier geprüft wird
//
//   - Der ganze Weg: **binden → melden → melden → übernehmen**, über einen
//     echten Bestand.
//   - **Ohne Bindung ist das kein Fehler** — die meisten Paarungen werden von
//     Hand eingetragen.
//   - Jeder Weg fragt nach **seiner** Wirkung, und jede Antwort trägt leere
//     Listen statt `null`.

import (
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// turniertisch ist, was ein Lauf über das Match braucht.
type turniertisch struct {
	anwendung *App

	turnierID string
	paarungID string
	partieID  string

	// zuordnung ordnet die beiden Teilnehmer den beiden Sitzen zu.
	zuordnung []MatchsitzDTO
}

// einTurniertisch gibt eine Anwendung, ein Turnier mit einer Paarung und eine
// laufende Partie mit zwei Sitzen.
func einTurniertisch(t *testing.T) turniertisch {
	t.Helper()

	db := bestandFuerDenRiegel(t)
	anwendung := MitPartienleitung(NeueApp(""), db)
	anwendung.profil.Store(1)

	anwendung.MitTurniernaht(service.NeueTurnierleitung(
		storage.NeueTurnierablage(db), storage.NeueDeckablage(db),
		storage.NeuerSetLeser(db)))

	anwendung.MitMatchnaht(service.NeueMatchleitung(
		storage.NeueTurnierablage(db), storage.NeuePartieablage(db),
		storage.NeueMatchablage(db)))

	ctx := riegelkontext(t)

	deckID := einDeckImBestand(t, db)

	turnier := anwendung.TurnierAnlegen(ctx, "Freitag", "magic", "", false, "")
	if turnier.Fehler != nil {
		t.Fatalf("TurnierAnlegen: %+v", turnier.Fehler)
	}

	turnierID := turnier.Turnier.ID

	teilnehmer := make([]string, 0, 2)

	for _, name := range []string{"Anna", "Bert"} {
		angemeldet := anwendung.TurnierAnmelden(ctx, turnierID, name, deckID)
		if angemeldet.Fehler != nil {
			t.Fatalf("TurnierAnmelden: %+v", angemeldet.Fehler)
		}

		teilnehmer = append(teilnehmer, angemeldet.Teilnehmer.ID)
	}

	runde := anwendung.TurnierRundeLegen(ctx, turnierID)
	if runde.Fehler != nil {
		t.Fatalf("TurnierRundeLegen: %+v", runde.Fehler)
	}

	paarungID := runde.Runde.Paarungen[0].ID

	partie := PartieflaecheUeber(db).PartieAusWuenschen(ctx, "magic", "anna",
		[]Sitzwunsch{
			{Kennung: "anna", Name: "Anna", DeckID: deckID},
			{Kennung: "bert", Name: "Bert", DeckID: deckID},
		})
	if partie.Fehler != nil {
		t.Fatalf("PartieAusWuenschen: %+v", partie.Fehler)
	}

	return turniertisch{
		anwendung: anwendung,
		turnierID: turnierID, paarungID: paarungID,
		partieID: partie.Partie.ID,
		zuordnung: []MatchsitzDTO{
			{TeilnehmerID: teilnehmer[0], Sitz: "anna"},
			{TeilnehmerID: teilnehmer[1], Sitz: "bert"},
		},
	}
}

// **Ohne Bindung sagt die Antwort es — und meldet keinen Fehler.**
func TestOhneBindungIstDasKeinFehler(t *testing.T) {
	tisch := einTurniertisch(t)

	antwort := tisch.anwendung.TurnierMatch(riegelkontext(t), tisch.paarungID)

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	case antwort.Gebunden:
		t.Error("es ist etwas gebunden")
	case antwort.Sitze == nil || antwort.Meldungen == nil:
		t.Error("die Listen sind nil statt leer")
	}
}

// **Der ganze Weg: binden, melden, melden, übernehmen.**
func TestEinMatchWirdGebundenGemeldetUndUebernommen(t *testing.T) {
	tisch := einTurniertisch(t)
	ctx := riegelkontext(t)

	gebunden := tisch.anwendung.TurnierMatchBinden(ctx, tisch.turnierID,
		tisch.paarungID, tisch.partieID, tisch.zuordnung)

	switch {
	case gebunden.Fehler != nil:
		t.Fatalf("binden: %+v", gebunden.Fehler)
	case !gebunden.Gebunden:
		t.Fatal("es wurde nichts gebunden")
	case len(gebunden.Sitze) != 2:
		t.Fatalf("es sind %d Sitze", len(gebunden.Sitze))
	case gebunden.Sitze[0].Deckabdruck == "":
		t.Error("der Abdruck fehlt")
	}

	stand := ErgebnisDTO{SpieleA: 2, SpieleB: 1}

	for _, von := range []string{"anna", "bert"} {
		gemeldet := tisch.anwendung.TurnierMatchMelden(
			ctx, tisch.paarungID, von, "letztesglied", stand)
		if gemeldet.Fehler != nil {
			t.Fatalf("melden (%s): %+v", von, gemeldet.Fehler)
		}
	}

	uebernommen := tisch.anwendung.TurnierMatchUebernehmen(
		ctx, tisch.turnierID, tisch.paarungID)

	switch {
	case uebernommen.Fehler != nil:
		t.Fatalf("übernehmen: %+v", uebernommen.Fehler)
	case uebernommen.Uebernommen == nil:
		t.Fatal("es wurde nichts übernommen")
	case uebernommen.Uebernommen.SpieleA != 2:
		t.Errorf("der Stand ist %+v", uebernommen.Uebernommen)
	}

	// **Und das Turnier trägt es.**
	turnier := tisch.anwendung.Turnier(ctx, tisch.turnierID)
	if turnier.Fehler != nil {
		t.Fatalf("Turnier: %+v", turnier.Fehler)
	}

	ergebnis := turnier.Turnier.Runden[0].Paarungen[0].Ergebnis
	if ergebnis == nil || ergebnis.SpieleA != 2 {
		t.Fatalf("das Ergebnis ist %+v", ergebnis)
	}

	// **Die Wiederholung ändert nichts** — und klagt nicht.
	if wieder := tisch.anwendung.TurnierMatchUebernehmen(
		ctx, tisch.turnierID, tisch.paarungID); wieder.Fehler != nil {
		t.Fatalf("die zweite Übernahme: %+v", wieder.Fehler)
	}
}

// **Jeder Weg fragt nach seiner Wirkung.**
func TestJederMatchwegFragtNachSeinerWirkung(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		fahre   func(*App) MatchAntwort
		wirkung domain.Wirkung
	}{
		"lesen": {
			fahre: func(a *App) MatchAntwort {
				return a.TurnierMatch(t.Context(), "7")
			},
			wirkung: domain.WirkungLesen,
		},
		"binden": {
			fahre: func(a *App) MatchAntwort {
				return a.TurnierMatchBinden(t.Context(), "1", "7", "p1", nil)
			},
			wirkung: domain.WirkungSchreiben,
		},
		"melden": {
			fahre: func(a *App) MatchAntwort {
				return a.TurnierMatchMelden(
					t.Context(), "7", "anna", "glied", ErgebnisDTO{})
			},
			wirkung: domain.WirkungSchreiben,
		},
		"aus der Partie": {
			fahre: func(a *App) MatchAntwort {
				return a.TurnierMatchAusDerPartie(t.Context(), "7", "anna")
			},
			wirkung: domain.WirkungSchreiben,
		},
		"übernehmen": {
			fahre: func(a *App) MatchAntwort {
				return a.TurnierMatchUebernehmen(t.Context(), "1", "7")
			},
			wirkung: domain.WirkungSchreiben,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wache := &abweisendeWache{}
			anwendung := NeueApp("")

			anwendung.MitWache(wache)

			antwort := fall.fahre(anwendung)

			switch {
			case antwort.Fehler == nil:
				t.Fatal("die Abweisung kam nicht an")
			case antwort.Sitze == nil || antwort.Meldungen == nil:
				t.Error("die Listen sind nil statt leer")
			case wache.gefragt != fall.wirkung:
				t.Errorf("gefragt wurde nach %q", wache.gefragt)
			}
		})
	}
}

// **Das Ergebnis erreicht das Turnier, ohne dass jemand es eintippt** — die
// dritte Abnahme von M17
// (#1618).
//
// **Der ganze Weg und nicht der Ausschnitt**: binden, spielen bis zum Ende,
// beide Seiten melden **aus der Partie**, übernehmen — und am Ende steht das
// Ergebnis in der Paarung. Kein Spielstand und kein Kettenabdruck geht dabei
// durch die Hände eines Aufrufers.
func TestDasErgebnisErreichtDasTurnierOhneEingabe(t *testing.T) {
	tisch := einTurniertisch(t)
	ctx := riegelkontext(t)

	gebunden := tisch.anwendung.TurnierMatchBinden(ctx, tisch.turnierID,
		tisch.paarungID, tisch.partieID, tisch.zuordnung)
	if gebunden.Fehler != nil {
		t.Fatalf("binden: %+v", gebunden.Fehler)
	}

	// **Bert gibt auf** — CR 104.3a, und damit gewinnt Anna (CR 104.2a).
	stand := tisch.anwendung.Partie(ctx, tisch.partieID, "anna")
	if stand.Fehler != nil {
		t.Fatalf("Partie: %+v", stand.Fehler)
	}

	fassung, fehler := strconv.ParseInt(stand.Partie.Fassung, 10, 64)
	if fehler != nil {
		t.Fatalf("Fassung: %v", fehler)
	}

	aufgegeben := tisch.anwendung.PartieBefehl(ctx, domain.Spielbefehl{
		PartieID: tisch.partieID, Kennung: "b1", Absender: "bert",
		Art: "zug.aufgeben", Erwartet: fassung,
	})
	if aufgegeben.Fehler != nil {
		t.Fatalf("aufgeben: %+v", aufgegeben.Fehler)
	}

	for _, von := range []string{"anna", "bert"} {
		gemeldet := tisch.anwendung.TurnierMatchAusDerPartie(
			ctx, tisch.paarungID, von)
		if gemeldet.Fehler != nil {
			t.Fatalf("melden (%s): %+v", von, gemeldet.Fehler)
		}
	}

	uebernommen := tisch.anwendung.TurnierMatchUebernehmen(
		ctx, tisch.turnierID, tisch.paarungID)
	if uebernommen.Fehler != nil {
		t.Fatalf("übernehmen: %+v", uebernommen.Fehler)
	}

	turnier := tisch.anwendung.Turnier(ctx, tisch.turnierID)
	if turnier.Fehler != nil {
		t.Fatalf("Turnier: %+v", turnier.Fehler)
	}

	ergebnis := turnier.Turnier.Runden[0].Paarungen[0].Ergebnis

	switch {
	case ergebnis == nil:
		t.Fatal("die Paarung trägt kein Ergebnis")
	case ergebnis.SpieleA != 1 || ergebnis.SpieleB != 0:
		t.Fatalf("das Ergebnis ist %+v", ergebnis)
	}
}

// **Eine Paarungskennung, die keine ist, wird abgewiesen.**
func TestEineKrummeKennungMeldetNichtAusDerPartie(t *testing.T) {
	t.Parallel()

	antwort := NeueApp("").TurnierMatchAusDerPartie(t.Context(), "sieben", "a")

	switch {
	case antwort.Fehler == nil:
		t.Fatal("die krumme Kennung kam durch")
	case antwort.Sitze == nil || antwort.Meldungen == nil:
		t.Error("die Listen sind nil statt leer")
	}
}
