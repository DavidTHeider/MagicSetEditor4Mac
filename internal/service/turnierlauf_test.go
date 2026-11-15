package service_test

// **Ein ganzes kleines Turnier** — die Abnahme aus
// #1294:
//
//	„Ein vollständiges kleines Turnier — acht Teilnehmer, drei Runden — lässt
//	sich von der Anlage bis zur Endrangliste durchführen."
//
// **Drei Runden bei acht Teilnehmern** ist die Zahl aus Appendix E der
// *Magic Tournament Rules* (5–8 Spieler → 3 Swiss-Runden), Fassung vom
// 2026-02-27.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Acht Teilnehmer, drei Runden, eine Endrangliste.**
func TestEinGanzesKleinesTurnier(t *testing.T) {
	t.Parallel()

	werk := neuesTurnierwerk(t)
	turnier := werk.turnierAnlegen(t)

	wer := werk.anmelden(t, turnier.ID,
		"Ada", "Bert", "Cem", "Dana", "Eike", "Ferry", "Gil", "Hana")

	for runde := 1; runde <= 3; runde++ {
		gelegt, err := werk.leitung.RundeLegen(t.Context(), turnier.ID)
		if err != nil {
			t.Fatalf("Runde %d legen: %v", runde, err)
		}

		if gelegt.Nummer != runde || len(gelegt.Paarungen) != 4 {
			t.Fatalf("Runde %d: %+v", runde, gelegt)
		}

		// **Der Erstgenannte gewinnt jedes Match 2–0.** Ein fester Ausgang,
		// damit die Endrangliste nachrechenbar ist.
		for _, paarung := range gelegt.Paarungen {
			if err := werk.leitung.ErgebnisSetzen(t.Context(), turnier.ID,
				werk.fassung(t, turnier.ID), paarung.ID, &domain.Ergebnis{SpieleA: 2}); err != nil {
				t.Fatalf("Ergebnis in Runde %d: %v", runde, err)
			}
		}
	}

	rangliste, err := werk.leitung.Rangliste(t.Context(), turnier.ID)
	if err != nil {
		t.Fatalf("Rangliste: %v", err)
	}

	if len(rangliste) != len(wer) {
		t.Fatalf("%d Ränge, erwartet %d", len(rangliste), len(wer))
	}

	// **Genau einer hat alles gewonnen.** Bei acht Teilnehmern und drei
	// Runden im Schweizer System bleibt einer ohne Niederlage.
	var ungeschlagen int

	for _, rang := range rangliste {
		if rang.Matchpunkte == 9 {
			ungeschlagen++
		}
	}

	if ungeschlagen != 1 {
		t.Errorf("%d Teilnehmer mit neun Punkten, erwartet einen", ungeschlagen)
	}

	// **Und die Plätze sind lückenlos vergeben.**
	for stelle, rang := range rangliste {
		if rang.Platz != stelle+1 {
			t.Errorf("Rang %d trägt Platz %d", stelle+1, rang.Platz)
		}
	}
}

// **Eine neue Runde entsteht nicht, solange die laufende offen ist.**
//
// Sie paarte sonst nach einem Stand, den es noch nicht gibt — und das merkt
// niemand, bis die Endrangliste nicht stimmt.
func TestKeineNeueRundeSolangeDieLaufendeOffenIst(t *testing.T) {
	t.Parallel()

	werk := neuesTurnierwerk(t)
	turnier := werk.turnierAnlegen(t)

	werk.anmelden(t, turnier.ID, "Ada", "Bert")

	if _, err := werk.leitung.RundeLegen(t.Context(), turnier.ID); err != nil {
		t.Fatalf("die erste Runde: %v", err)
	}

	if _, err := werk.leitung.RundeLegen(t.Context(), turnier.ID); !errors.Is(
		err, domain.ErrValidation,
	) {
		t.Errorf("Fehler ist %v, erwartet eine Validierung", err)
	}
}

// **Ein Bye hält die nächste Runde nicht auf.**
//
// Er *ist* ein Ergebnis (Appendix C, 2–0) und wartet auf nichts.
func TestEinByeHaeltDieNaechsteRundeNichtAuf(t *testing.T) {
	t.Parallel()

	werk := neuesTurnierwerk(t)
	turnier := werk.turnierAnlegen(t)

	werk.anmelden(t, turnier.ID, "Ada", "Bert", "Cem")

	erste, err := werk.leitung.RundeLegen(t.Context(), turnier.ID)
	if err != nil {
		t.Fatalf("die erste Runde: %v", err)
	}

	// **Nur das gespielte Match wird eingetragen**, der Bye bleibt, wie er ist.
	for _, paarung := range erste.Paarungen {
		if paarung.IstBye() {
			continue
		}

		if err := werk.leitung.ErgebnisSetzen(t.Context(), turnier.ID,
			werk.fassung(t, turnier.ID), paarung.ID, &domain.Ergebnis{SpieleA: 2}); err != nil {
			t.Fatalf("Ergebnis: %v", err)
		}
	}

	if _, err := werk.leitung.RundeLegen(t.Context(), turnier.ID); err != nil {
		t.Errorf("die zweite Runde: %v", err)
	}
}

// **Ein Ergebnis für eine fremde Paarung wird abgewiesen.**
func TestEinErgebnisFuerEineFremdePaarungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	werk := neuesTurnierwerk(t)
	turnier := werk.turnierAnlegen(t)

	werk.anmelden(t, turnier.ID, "Ada", "Bert")

	if _, err := werk.leitung.RundeLegen(t.Context(), turnier.ID); err != nil {
		t.Fatalf("legen: %v", err)
	}

	vorher := werk.turniere.gesetzt

	if err := werk.leitung.ErgebnisSetzen(t.Context(), turnier.ID,
		werk.fassung(t, turnier.ID), 999,
		&domain.Ergebnis{SpieleA: 2}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", err)
	}

	// **Und der Speicher wurde gar nicht erst gefragt.**
	//
	// Ohne diese Zeile prüfte der Lauf nichts: das Doppel gibt für eine
	// unbekannte Paarung ebenfalls `ErrNotFound`, und der Fehler käme mit
	// **und ohne** die Prüfung im Dienst. Zwei Lesarten, und die Gegenprobe
	// blieb genau deshalb grün (§10).
	if werk.turniere.gesetzt != vorher {
		t.Error("der Speicher wurde gerufen — die Prüfung im Dienst greift nicht")
	}
}

// **Ein Einzelspielturnier nimmt kein Match über zwei Spiele.**
func TestEinEinzelspielturnierNimmtKeineZweiSpiele(t *testing.T) {
	t.Parallel()

	werk := neuesTurnierwerk(t)

	turnier, err := werk.leitung.Anlegen(t.Context(), domain.Turnier{
		Name: "Blitz", Spiel: "magic", Einzelspiel: true,
	})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	werk.anmelden(t, turnier.ID, "Ada", "Bert")

	gelegt, err := werk.leitung.RundeLegen(t.Context(), turnier.ID)
	if err != nil {
		t.Fatalf("legen: %v", err)
	}

	if err := werk.leitung.ErgebnisSetzen(t.Context(), turnier.ID,
		werk.fassung(t, turnier.ID), gelegt.Paarungen[0].ID, &domain.Ergebnis{SpieleA: 2}); !errors.Is(
		err, domain.ErrValidation,
	) {
		t.Errorf("Fehler ist %v, erwartet eine Validierung", err)
	}

	// **Und eines über ein Spiel geht durch.**
	if err := werk.leitung.ErgebnisSetzen(t.Context(), turnier.ID,
		werk.fassung(t, turnier.ID), gelegt.Paarungen[0].ID, &domain.Ergebnis{SpieleA: 1}); err != nil {
		t.Errorf("abgewiesen: %v", err)
	}
}

// **Ein Turnier ohne Spielbereite bekommt keine Runde.**
func TestEinTurnierOhneSpielbereiteBekommtKeineRunde(t *testing.T) {
	t.Parallel()

	werk := neuesTurnierwerk(t)
	turnier := werk.turnierAnlegen(t)

	if _, err := werk.leitung.RundeLegen(t.Context(), turnier.ID); !errors.Is(
		err, domain.ErrValidation,
	) {
		t.Errorf("Fehler ist %v, erwartet eine Validierung", err)
	}
}

// **Der Vorschlag schafft keine Tatsachen.**
//
// Ein Turnierleiter soll hinsehen können, bevor er eine Runde legt.
func TestDerVorschlagSchafftKeineTatsachen(t *testing.T) {
	t.Parallel()

	werk := neuesTurnierwerk(t)
	turnier := werk.turnierAnlegen(t)

	werk.anmelden(t, turnier.ID, "Ada", "Bert")

	plan, err := werk.leitung.Paarungsvorschlag(t.Context(), turnier.ID)
	if err != nil {
		t.Fatalf("vorschlagen: %v", err)
	}

	if len(plan) != 1 {
		t.Fatalf("%d Paarungen, erwartet eine", len(plan))
	}

	gelesen, err := werk.leitung.Turnier(t.Context(), turnier.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(gelesen.Runden) != 0 {
		t.Errorf("%d Runden entstanden — ein Vorschlag legt keine",
			len(gelesen.Runden))
	}
}
