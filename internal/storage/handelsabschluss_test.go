package storage_test

// **Der Abschluss in einem Zug** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// # Was hier gemessen wird
//
// **Dass alles zusammen geschieht oder nichts.** Abgang, Zugang, Journal,
// Zusagenabschluss, Zustandsereignis und Beleg stehen hinterher entweder alle
// da oder alle nicht — dazwischen gibt es keinen Stand.
//
// **Dass eine Menge auf null die Besitzzeile löscht und das Journal stehen
// bleibt.** Das ist der Satz, für den es ein Journal gibt: die Zeile, die
// sagt, wohin das letzte Stück ging, darf nicht mit ihm verschwinden.
//
// **Und dass ein zweiter Abschluss nichts bewegt** — gleich mit welcher
// Nachrichtenkennung.
//
// **Kein `t.Parallel()`**, aus dem Grund, der in `sammlung_test.go` steht.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// einHandel ist der Vorgang, um den es in diesen Läufen geht.
const einHandel = "v-abschluss"

// abschlussprobe stellt Bestand, Zusagen und Abschluss über einer Datenbank.
type abschlussprobe struct {
	bestand     *storage.Sammlungsspeicher
	zusagen     *storage.Reservierungsablage
	buch        *storage.Abschlussablage
	briefkasten *storage.Briefkastenablage
	ausgang     *storage.Ausgangsablage
}

func eineAbschlussprobe(t *testing.T) abschlussprobe {
	t.Helper()

	db := frischeDatenbank(t)

	return abschlussprobe{
		bestand:     storage.NeuerSammlungsspeicher(db),
		zusagen:     storage.NeueReservierungsablage(db),
		buch:        storage.NeueAbschlussablage(db),
		briefkasten: storage.NeueBriefkastenablage(db),
		ausgang:     storage.NeueAusgangsablage(db),
	}
}

// vierStueck legt vier Stück hin und sagt zwei davon zu — übergeben.
func (p abschlussprobe) vierStueck(t *testing.T, uebergeben bool) *domain.Exemplar {
	t.Helper()

	exemplar := &domain.Exemplar{
		Printing: "bolt-m10", Menge: 4, Zustand: domain.ZustandNearMint,
		Sprache: "de", Erwerbspreis: 350, Erwerbswaehrung: "EUR",
		Lagerort: "Kiste 4",
	}

	kennung, fehler := p.bestand.Anlegen(probekontext(t), exemplar)
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	exemplar.ID, exemplar.Fassung = kennung, 1

	if fehler := p.zusagen.Festhalten(probekontext(t), domain.Reservierung{
		Vorgang: einHandel, Revision: 1, ExemplarID: kennung, Menge: 2,
		Fassung: 1, Frist: time.Now().Add(time.Hour), Uebergeben: uebergeben,
	}, ueber(exemplar, 99)); fehler != nil {
		t.Fatalf("festhalten: %v", fehler)
	}

	return exemplar
}

// einAbschluss baut den Abschluss zu einem Exemplar.
func einAbschluss(exemplar *domain.Exemplar) domain.Handelsabschluss {
	abgang := domain.Journalzeile{
		Art: domain.JournalAbgang, Anlass: domain.AnlassTausch,
		Vorgang: einHandel, Printing: "bolt-m10", Menge: 2,
		Zustand: domain.ZustandNearMint, Sprache: "de",
		Finish: domain.FinishNonfoil, ExemplarID: exemplar.ID,
		Lagerort: "Kiste 4",
	}

	zugang := domain.Journalzeile{
		Art: domain.JournalZugang, Anlass: domain.AnlassTausch,
		Vorgang: einHandel, Printing: "shivan-lea", Menge: 1,
		Zustand: domain.ZustandNearMint, Sprache: "en",
		Finish: domain.FinishNonfoil,
	}

	angebot := domain.Handelsangebot{
		Vertragsfassung: 1, Vorgang: einHandel, Revision: 2,
		Eltern: "eine-erste-revision",
		Art:    domain.HandelTausch, Initiator: "aich",
		Parteien: []string{"aich", "bdu"},
		Zustand:  domain.HandelLokalVerbucht,
	}

	return domain.Handelsabschluss{
		Angebot: angebot,
		Beleg: domain.Handelsbeleg{
			// **Der Abdruck der Revision und keine Zeichenkette daneben** —
			// er ist die Zusage, dass unterschrieben wurde, was gebucht wird.
			Vorgang: einHandel, Abdruck: angebot.Abdruck(),
			Parteien: []string{"aich", "bdu"},
			Zeilen:   []domain.Journalzeile{abgang, zugang},
		},
		Unterschrift: "unterschrift-vorher-gerechnet",
	}
}

// menge liest die Stückzahl eines Exemplars — oder sagt, dass es keines gibt.
func (p abschlussprobe) menge(t *testing.T, id int64) (int, bool) {
	t.Helper()

	exemplar, fehler := p.bestand.Einzeln(probekontext(t), id)
	if errors.Is(fehler, domain.ErrNotFound) {
		return 0, false
	}

	if fehler != nil {
		t.Fatalf("einzeln: %v", fehler)
	}

	return exemplar.Menge, true
}

func TestEinAbschlussBewegtBestandJournalUndZusageZusammen(t *testing.T) {
	probe := eineAbschlussprobe(t)
	ctx := probekontext(t)
	exemplar := probe.vierStueck(t, true)

	if fehler := probe.buch.Abschliessen(ctx, einAbschluss(exemplar)); fehler != nil {
		t.Fatalf("abschließen: %v", fehler)
	}

	// **Der Abgang ist abgezogen.**
	if menge, da := probe.menge(t, exemplar.ID); !da || menge != 2 {
		t.Errorf("es liegen %d Stück da (da: %v) statt 2", menge, da)
	}

	// **Der Zugang liegt da — und erbt keinen Preis.**
	zugegangen, fehler := probe.bestand.Fuer(ctx, "shivan-lea")
	if fehler != nil {
		t.Fatalf("für den Druck: %v", fehler)
	}

	if len(zugegangen) != 1 {
		t.Fatalf("es kamen %d Zeilen statt einer", len(zugegangen))
	}

	if zugegangen[0].Menge != 1 {
		t.Errorf("es kam %d Stück statt einem", zugegangen[0].Menge)
	}

	if zugegangen[0].Erwerbspreis != 0 || zugegangen[0].Erwerbswaehrung != "" {
		t.Errorf("das empfangene Stück trägt einen Preis: %d %q",
			zugegangen[0].Erwerbspreis, zugegangen[0].Erwerbswaehrung)
	}

	// **Die Zusage ist fort** — sie ist eingelöst und nicht mehr offen.
	offen, fehler := probe.zusagen.Reservierungen(ctx, exemplar.ID)
	if fehler != nil {
		t.Fatalf("Reservierungen: %v", fehler)
	}

	if len(offen) != 0 {
		t.Errorf("es stehen %d Zusagen offen", len(offen))
	}

	// **Das Journal führt beide Bewegungen.**
	zeilen, fehler := probe.buch.Journal(ctx, einHandel)
	if fehler != nil {
		t.Fatalf("Journal: %v", fehler)
	}

	if len(zeilen) != 2 {
		t.Fatalf("das Journal führt %d Zeilen statt zweier", len(zeilen))
	}

	// **Und der Beleg steht da — einseitig.**
	beleg, fehler := probe.buch.Beleg(ctx, einHandel)
	if fehler != nil {
		t.Fatalf("Beleg: %v", fehler)
	}

	if beleg == nil {
		t.Fatal("es steht kein Beleg da")
	}

	if beleg.Beidseitig {
		t.Error("der Beleg gibt sich als beidseitig aus — es ist erst einer")
	}

	if beleg.Unterschrift != "unterschrift-vorher-gerechnet" {
		t.Errorf("die Unterschrift lautet %q", beleg.Unterschrift)
	}
}

func TestMengeNullLoeschtDieBesitzzeileUndDasJournalBleibt(t *testing.T) {
	probe := eineAbschlussprobe(t)
	ctx := probekontext(t)
	exemplar := probe.vierStueck(t, true)

	abschluss := einAbschluss(exemplar)

	// **Alle vier gehen** — die Zusage wird auf vier gehoben.
	abschluss.Beleg.Zeilen[0].Menge = 4

	if fehler := probe.zusagen.Festhalten(ctx, domain.Reservierung{
		Vorgang: einHandel, Revision: 1, ExemplarID: exemplar.ID, Menge: 2,
		Fassung: 1, Frist: time.Now().Add(time.Hour), Uebergeben: true,
	}, ueber(exemplar, 99)); fehler != nil {
		t.Fatalf("zweite Zusage: %v", fehler)
	}

	if fehler := probe.buch.Abschliessen(ctx, abschluss); fehler != nil {
		t.Fatalf("abschließen: %v", fehler)
	}

	if _, da := probe.menge(t, exemplar.ID); da {
		t.Error("die Besitzzeile steht noch da")
	}

	zeilen, fehler := probe.buch.Journal(ctx, einHandel)
	if fehler != nil {
		t.Fatalf("Journal: %v", fehler)
	}

	if len(zeilen) != 2 {
		t.Fatalf("das Journal führt %d Zeilen statt zweier", len(zeilen))
	}

	// **Die gelöschte Zeile steht im Journal weiter mit ihrer Kennung** —
	// sie ist der einzige Weg zurück zu dem, was da war.
	var abgang domain.Journalzeile

	for _, zeile := range zeilen {
		if zeile.Art == domain.JournalAbgang {
			abgang = zeile
		}
	}

	if abgang.ExemplarID != exemplar.ID {
		t.Errorf("die Abgangszeile nennt %d statt %d",
			abgang.ExemplarID, exemplar.ID)
	}

	if abgang.Lagerort != "Kiste 4" {
		t.Errorf("der Lagerort ging verloren: %q", abgang.Lagerort)
	}
}

func TestEinZweiterAbschlussBewegtNichts(t *testing.T) {
	probe := eineAbschlussprobe(t)
	ctx := probekontext(t)
	exemplar := probe.vierStueck(t, true)

	abschluss := einAbschluss(exemplar)

	if fehler := probe.buch.Abschliessen(ctx, abschluss); fehler != nil {
		t.Fatalf("erster Abschluss: %v", fehler)
	}

	// **Eine neue Nachrichtenkennung ändert daran nichts** — der Vorgang ist
	// derselbe, und er ist verbucht.
	abschluss.Vermerk = &domain.Inboxeintrag{
		Schluessel: "ganz\x00neu", Nutzlastabdruck: "abdruck-neu",
		Ergebnis: `{"verbucht":true}`,
	}

	fehler := probe.buch.Abschliessen(ctx, abschluss)
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("der zweite Abschluss gab %v statt eines Konflikts", fehler)
	}

	if menge, da := probe.menge(t, exemplar.ID); !da || menge != 2 {
		t.Errorf("es liegen %d Stück da (da: %v) statt 2 — der zweite "+
			"Abschluss hat gebucht", menge, da)
	}

	zeilen, fehler := probe.buch.Journal(ctx, einHandel)
	if fehler != nil {
		t.Fatalf("Journal: %v", fehler)
	}

	if len(zeilen) != 2 {
		t.Errorf("das Journal führt %d Zeilen statt zweier", len(zeilen))
	}
}
