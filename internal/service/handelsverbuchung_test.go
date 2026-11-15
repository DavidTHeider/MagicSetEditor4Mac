package service_test

// **Die lokale Verbuchung eines Handels** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// # Was hier gemessen wird
//
// **Dass zweimal Verbuchen einmal bucht.** Ein Neustart, ein
// wiederhergestelltes Backup, eine zweite Nachricht — der Vorgang ist
// derselbe, und der Bestand bewegt sich einmal.
//
// **Dass ohne bestätigte Übergabe gar nicht gebucht wird**, und dass eine
// fehlende Gegenseite den Abgleich **offen** meldet, statt zu buchen oder zu
// scheitern.
//
// **Und dass die Unterschrift vor dem Bestand entsteht.** Sie hängt am
// Schlüsselbund; wer sie in der offenen Transaktion rechnete, ließe eine
// Datenbank auf einen Dialog warten.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// derVorgang ist der Handel dieser Läufe.
const derVorgang = "v-verbuchung"

// derUebergabestand ist ein Angebot, dessen Übergabe offen ist.
func derUebergabestand(ich peer.Identitaet) domain.Handelsangebot {
	return domain.Handelsangebot{
		Vertragsfassung: 1, Vorgang: derVorgang, Revision: 3,
		Eltern: "eine-zweite-revision", Art: domain.HandelTausch,
		Initiator: string(ich.Kennung()),
		Parteien:  []string{string(ich.Kennung()), "bgegenueber"},
		Zustand:   domain.HandelUebergabeOffen,
		Positionen: []domain.Handelsposition{
			{
				Von: string(ich.Kennung()), Printing: "bolt-m10", Menge: 2,
				Guete: string(domain.ZustandNearMint), Sprache: "de", Finish: "nonfoil",
			},
			{
				Von: "bgegenueber", Printing: "shivan-lea", Menge: 1,
				Guete: string(domain.ZustandNearMint), Sprache: "en", Finish: "nonfoil",
			},
		},
	}
}

// stelleEin legt Stand, Besitz und eine bestätigte Zusage hin.
func stelleEin(
	t *testing.T, stand *handelsprobe, bestand *verbuchungsprobe,
	uebergeben bool,
) {
	t.Helper()

	ich := derAusweis(t)
	stand.staende[derVorgang] = derUebergabestand(ich)

	bestand.stuecke[7] = domain.Exemplar{
		ID: 7, Printing: "bolt-m10", Menge: 4,
		Zustand: domain.ZustandNearMint, Sprache: "de",
		Finish: domain.FinishNonfoil, Erwerbspreis: 350,
		Erwerbswaehrung: "EUR", Lagerort: "Kiste 4", Fassung: 1,
	}

	bestand.zusagen = []domain.Reservierung{{
		Vorgang: derVorgang, Revision: 3, ExemplarID: 7, Menge: 2,
		Fassung: 1, Frist: time.Now().Add(time.Hour),
		Uebergeben: uebergeben,
	}}
}

func TestEineVerbuchungBuchtBeideRichtungenUndUnterschreibt(t *testing.T) {
	t.Parallel()

	stelle, stand, bestand := eineVerbuchung(t)
	stelleEin(t, stand, bestand, true)

	neu, fehler := stelle.Verbuchen(t.Context(), derVorgang, derFreibrief)
	if fehler != nil {
		t.Fatalf("verbuchen: %v", fehler)
	}

	if neu.Zustand != domain.HandelLokalVerbucht {
		t.Errorf("der Vorgang steht auf %q", neu.Zustand)
	}

	if len(bestand.gebucht) != 1 {
		t.Fatalf("es wurde %dmal gebucht", len(bestand.gebucht))
	}

	abschluss := bestand.gebucht[0]

	if len(abschluss.Beleg.Zeilen) != 2 {
		t.Fatalf("der Beleg führt %d Zeilen", len(abschluss.Beleg.Zeilen))
	}

	var abgang, zugang domain.Journalzeile

	for _, zeile := range abschluss.Beleg.Zeilen {
		if zeile.Art == domain.JournalAbgang {
			abgang = zeile
		} else {
			zugang = zeile
		}
	}

	// **Der Abgang nennt die Besitzzeile** — ohne sie wüsste der Bestand
	// nicht, wovon er abzieht.
	if abgang.ExemplarID != 7 || abgang.Menge != 2 {
		t.Errorf("der Abgang ist %+v", abgang)
	}

	if abgang.Lagerort != "Kiste 4" {
		t.Errorf("der Lagerort fehlt: %q", abgang.Lagerort)
	}

	// **Der Zugang kommt aus dem Angebot und bringt keine Besitzzeile mit.**
	if zugang.Printing != "shivan-lea" || zugang.Menge != 1 ||
		zugang.ExemplarID != 0 {
		t.Errorf("der Zugang ist %+v", zugang)
	}

	// **Die Unterschrift steht an der Form, die hinausgeht** — und sie hält.
	unterschrift, entschluesselt := hexAus(t, abschluss.Unterschrift)
	if !entschluesselt {
		return
	}

	if fehler := peer.Pruefe(derAusweis(t).Kennung(),
		abschluss.Beleg.Unterschriftsform(), unterschrift); fehler != nil {
		t.Errorf("die Unterschrift hält nicht: %v", fehler)
	}

	// **Und der Beleg ist einseitig** — die zweite Unterschrift ist eine
	// eigene Tat.
	if abschluss.Beleg.Beidseitig {
		t.Error("der Beleg gibt sich als beidseitig aus")
	}
}

// **Zweimal Verbuchen ist einmal.**
func TestZweimalVerbuchenBuchtEinmal(t *testing.T) {
	t.Parallel()

	stelle, stand, bestand := eineVerbuchung(t)
	stelleEin(t, stand, bestand, true)

	if _, fehler := stelle.Verbuchen(t.Context(), derVorgang, derFreibrief); fehler != nil {
		t.Fatalf("erste Verbuchung: %v", fehler)
	}

	// **Der Stand steht danach auf `lokalVerbucht`** — ein zweiter Ruf käme
	// aus diesem Zustand und fände den Beleg.
	neu, fehler := stelle.Verbuchen(t.Context(), derVorgang, derFreibrief)
	if fehler != nil {
		t.Fatalf("zweite Verbuchung: %v", fehler)
	}

	if neu.Zustand != domain.HandelLokalVerbucht {
		t.Errorf("der Vorgang steht auf %q", neu.Zustand)
	}

	if len(bestand.gebucht) != 1 {
		t.Errorf("es wurde %dmal gebucht", len(bestand.gebucht))
	}
}

// **Ohne bestätigte Übergabe wird nicht gebucht.**
func TestOhneBestaetigteUebergabeWirdNichtVerbucht(t *testing.T) {
	t.Parallel()

	stelle, stand, bestand := eineVerbuchung(t)
	stelleEin(t, stand, bestand, false)

	_, fehler := stelle.Verbuchen(t.Context(), derVorgang, derFreibrief)
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("die Verbuchung gab %v statt eines Konflikts", fehler)
	}

	if len(bestand.gebucht) != 0 {
		t.Errorf("es wurde %dmal gebucht", len(bestand.gebucht))
	}
}

// **Eine Zusage, die weniger deckt, als abgeht, ist keine Deckung.**
func TestEineZuKleineZusageBuchtNicht(t *testing.T) {
	t.Parallel()

	stelle, stand, bestand := eineVerbuchung(t)
	stelleEin(t, stand, bestand, true)
	bestand.zusagen[0].Menge = 1

	_, fehler := stelle.Verbuchen(t.Context(), derVorgang, derFreibrief)
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("die Verbuchung gab %v statt eines Konflikts", fehler)
	}

	if len(bestand.gebucht) != 0 {
		t.Errorf("es wurde %dmal gebucht", len(bestand.gebucht))
	}
}

// **Eine Zusage rechnet gegen eine Fassung** (§7, Regel 2).
func TestEineVeralteteZusageBuchtNichtImDienst(t *testing.T) {
	t.Parallel()

	stelle, stand, bestand := eineVerbuchung(t)
	stelleEin(t, stand, bestand, true)

	stueck := bestand.stuecke[7]
	stueck.Fassung = 2
	bestand.stuecke[7] = stueck

	_, fehler := stelle.Verbuchen(t.Context(), derVorgang, derFreibrief)
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("die Verbuchung gab %v statt eines Konflikts", fehler)
	}
}

// **Eine Position, die dieses Programm nicht deuten kann, wird abgewiesen.**
//
// # Warum hier und nicht erst im Bestand
//
// **Weil die Positionen von der Gegenseite kommen** (§8: fremde Daten sind
// der Normalfall). Wer sie durchreichte, bekäme die Beanstandung aus der
// Tiefe einer offenen Transaktion — mit einem Satz über eine Spalte statt
// über den Handel, und nachdem der Abgang schon gerechnet war.
func TestEineUndeutbarePositionBuchtNicht(t *testing.T) {
	t.Parallel()

	stelle, stand, bestand := eineVerbuchung(t)
	stelleEin(t, stand, bestand, true)

	angebot := stand.staende[derVorgang]
	angebot.Positionen[1].Guete = "zerknittert"
	stand.staende[derVorgang] = angebot

	_, fehler := stelle.Verbuchen(t.Context(), derVorgang, derFreibrief)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("die Verbuchung gab %v statt einer Beanstandung", fehler)
	}

	if len(bestand.gebucht) != 0 {
		t.Errorf("es wurde %dmal gebucht", len(bestand.gebucht))
	}
}

// **Das Journal nennt die Karte und nicht ihre Kennung.**
//
// # Der Fund, für den dieser Lauf steht
//
// **Am Fenster stand „−2 × a1b2c3d4-0000-4000-8000-00000000000a".** Gesehen
// bei der Sichtprüfung am 2026-09-18; ein DOM-Lauf fängt das nicht, weil
// „steht die Menge da" auch dann zutrifft. Ein Journal, das niemand lesen
// kann, ist keins.
func TestDasJournalNenntDieKarteUndNichtIhreKennung(t *testing.T) {
	t.Parallel()

	stelle, stand, bestand := eineVerbuchung(t)
	stelleEin(t, stand, bestand, true)

	if _, fehler := stelle.Verbuchen(t.Context(), derVorgang, derFreibrief); fehler != nil {
		t.Fatalf("verbuchen: %v", fehler)
	}

	eintraege, fehler := stelle.Journal(t.Context(), derVorgang)
	if fehler != nil {
		t.Fatalf("Journal: %v", fehler)
	}

	if len(eintraege) != 2 {
		t.Fatalf("es kamen %d Einträge", len(eintraege))
	}

	namen := map[string]string{}
	for _, eintrag := range eintraege {
		namen[eintrag.Zeile.Printing] = eintrag.Bezeichnung
	}

	if namen["bolt-m10"] != "Lightning Bolt · M10 #146" {
		t.Errorf("der Abgang heißt %q", namen["bolt-m10"])
	}

	// **Ein Druck, den dieser Rechner nicht kennt, behält seine Kennung** —
	// ein empfangenes Stück nennt nur seinen Druck, und einen Namen zu
	// erfinden wäre schlimmer als keiner.
	if namen["shivan-lea"] != "shivan-lea" {
		t.Errorf("der Zugang heißt %q", namen["shivan-lea"])
	}
}

// **Und beide Lesewege reichen ihren Fehler durch.**
func TestJederWegDesJournalsImDienstReichtSeinenFehlerDurch(t *testing.T) {
	t.Parallel()

	klage := errors.New("es klagt")

	for name, verderben := range map[string]func(*verbuchungsprobe){
		"das Journal klagt": func(b *verbuchungsprobe) { b.journalklage = klage },
		"der Druck klagt":   func(b *verbuchungsprobe) { b.druckklage = klage },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			stelle, stand, bestand := eineVerbuchung(t)
			stelleEin(t, stand, bestand, true)

			if _, fehler := stelle.Verbuchen(t.Context(), derVorgang, derFreibrief); fehler != nil {
				t.Fatalf("verbuchen: %v", fehler)
			}

			verderben(bestand)

			if _, fehler := stelle.Journal(t.Context(), derVorgang); fehler == nil {
				t.Error("der Fehler kam nicht an")
			}
		})
	}
}
