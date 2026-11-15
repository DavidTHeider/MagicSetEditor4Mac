package storage_test

// **Das Journal neben dem Handel** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// # Was hier gemessen wird
//
// **Dass Kauf, Verkauf, Geschenk, Einfuhr und Korrektur einen Weg haben** —
// sonst wären fünf der sechs Anlässe eine Aufzählung ohne Verbraucher (§10).
//
// **Dass ein Tausch diesen Weg nicht nimmt.** Er bewegt den Bestand mit; wer
// ihn hier bucht, verschiebt das Journal gegen den Bestand.
//
// **Und dass die zweite Unterschrift eine eigene Tat ist** und nicht aus zwei
// Parteien gefolgert wird.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einKauf ist eine Zeile, die kein Handel bewirkt hat.
func einKauf() domain.Journalzeile {
	return domain.Journalzeile{
		Art: domain.JournalZugang, Anlass: domain.AnlassKauf,
		Printing: "bolt-m10", Menge: 1, Zustand: domain.ZustandNearMint,
		Sprache: "de", Finish: domain.FinishNonfoil, Lagerort: "Kiste 4",
	}
}

func TestEineBuchungOhneHandelStehtImJournal(t *testing.T) {
	probe := eineAbschlussprobe(t)
	ctx := probekontext(t)

	if fehler := probe.buch.Buchen(ctx, einKauf()); fehler != nil {
		t.Fatalf("buchen: %v", fehler)
	}

	// **Ohne Vorgang gelesen** — die Zeile gehört zu keinem Handel.
	zeilen, fehler := probe.buch.Journal(ctx, "")
	if fehler != nil {
		t.Fatalf("Journal: %v", fehler)
	}

	if len(zeilen) != 1 {
		t.Fatalf("das Journal führt %d Zeilen statt einer", len(zeilen))
	}

	if zeilen[0].Anlass != domain.AnlassKauf {
		t.Errorf("der Anlass ist %q", zeilen[0].Anlass)
	}

	// **Sie trägt eine Kennung** — daran hängt die Gegenbewegung.
	if zeilen[0].ID == 0 {
		t.Error("die Zeile trägt keine Kennung")
	}

	if zeilen[0].Gebucht.IsZero() {
		t.Error("die Zeile trägt keine Uhr")
	}

	// **Und die Gegenbewegung zu ihr nennt sie.**
	gegen, fehler := zeilen[0].Gegenbewegung("verzählt")
	if fehler != nil {
		t.Fatalf("Gegenbewegung: %v", fehler)
	}

	if gebucht := probe.buch.Buchen(ctx, gegen); gebucht != nil {
		t.Fatalf("die Gegenbewegung buchen: %v", gebucht)
	}

	nachher, fehler := probe.buch.Journal(ctx, "")
	if fehler != nil {
		t.Fatalf("Journal: %v", fehler)
	}

	if len(nachher) != 2 {
		t.Fatalf("das Journal führt %d Zeilen statt zweier", len(nachher))
	}

	// **Die erste steht unverändert da** — korrigiert wird daneben.
	if nachher[1].Bezug != zeilen[0].ID || nachher[1].Grund != "verzählt" {
		t.Errorf("die Korrektur ist %+v", nachher[1])
	}
}

// **Ein Tausch geht hier nicht durch.**
func TestEinTauschWirdNichtGebucht(t *testing.T) {
	probe := eineAbschlussprobe(t)

	zeile := einKauf()
	zeile.Anlass, zeile.Vorgang = domain.AnlassTausch, einHandel

	fehler := probe.buch.Buchen(probekontext(t), zeile)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Tausch gab %v statt einer Beanstandung", fehler)
	}
}

// **Eine unstimmige Zeile kommt nicht ins Journal.**
func TestEineUnstimmigeZeileKommtNichtInsJournal(t *testing.T) {
	probe := eineAbschlussprobe(t)

	zeile := einKauf()
	zeile.Menge = 0

	fehler := probe.buch.Buchen(probekontext(t), zeile)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("die Zeile gab %v statt einer Beanstandung", fehler)
	}
}

// **Die zweite Unterschrift ist eine eigene Tat.**
func TestDieGegenzeichnungIstEineEigeneTat(t *testing.T) {
	probe := eineAbschlussprobe(t)
	ctx := probekontext(t)
	exemplar := probe.vierStueck(t, true)

	if fehler := probe.buch.Abschliessen(
		ctx, einAbschluss(exemplar)); fehler != nil {
		t.Fatalf("abschließen: %v", fehler)
	}

	vorher, fehler := probe.buch.Beleg(ctx, einHandel)
	if fehler != nil {
		t.Fatalf("Beleg: %v", fehler)
	}

	if vorher.Beidseitig {
		t.Fatal("der Beleg war von Anfang an beidseitig")
	}

	if gezeichnet := probe.buch.Gegenzeichnen(ctx,
		domain.Fremdbeleg{Vorgang: einHandel, Von: "bert", Abdruck: "a1"}); gezeichnet != nil {
		t.Fatalf("gegenzeichnen: %v", gezeichnet)
	}

	nachher, fehler := probe.buch.Beleg(ctx, einHandel)
	if fehler != nil {
		t.Fatalf("Beleg: %v", fehler)
	}

	if !nachher.Beidseitig {
		t.Error("der Beleg ist nicht beidseitig")
	}

	// **Und die Form ist dieselbe geblieben.** Was unterschrieben wurde,
	// steht fest; hinzu kommt allein, dass ein Zweiter es getan hat.
	if nachher.Form != vorher.Form {
		t.Error("die unterschriebene Form hat sich geändert")
	}

	if nachher.Unterschrift != vorher.Unterschrift {
		t.Error("die Unterschrift hat sich geändert")
	}
}

// **Ein Beleg, den es nicht gibt, lässt sich nicht gegenzeichnen.**
func TestOhneBelegGibtEsKeineGegenzeichnung(t *testing.T) {
	probe := eineAbschlussprobe(t)

	fehler := probe.buch.Gegenzeichnen(probekontext(t),
		domain.Fremdbeleg{Vorgang: "v-gibtesnicht", Von: "bert"})
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("die Gegenzeichnung gab %v statt „gibt es nicht\"", fehler)
	}
}

// **Aus einem Wunsch geht nichts ab.**
func TestAusEinemWunschGehtNichtsAb(t *testing.T) {
	probe := eineAbschlussprobe(t)
	ctx := probekontext(t)
	exemplar := probe.vierStueck(t, true)

	wunsch := &domain.Exemplar{
		Printing: "shivan-lea", Menge: 1, Zustand: domain.ZustandNearMint,
		Sprache: "en", Wunsch: true,
	}

	kennung, fehler := probe.bestand.Anlegen(ctx, wunsch)
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	abschluss := einAbschluss(exemplar)
	abschluss.Beleg.Zeilen[0].ExemplarID = kennung
	abschluss.Beleg.Zeilen[0].Printing = "shivan-lea"

	if fehler := probe.buch.Abschliessen(ctx, abschluss); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Fatalf("der Abschluss gab %v statt eines Konflikts", fehler)
	}
}

// **Vermerk und Belegnachricht gehören zum selben Zug.**
//
// Das Issue verlangt sie ausdrücklich in derselben Transaktion: ein Vermerk,
// der ohne die Buchung dasteht, hieße „schon behandelt", während nichts
// geschehen ist; eine Belegnachricht ohne Buchung ginge hinaus und bezöge sich
// auf nichts.
func TestVermerkUndBelegnachrichtGehenMit(t *testing.T) {
	probe := eineAbschlussprobe(t)
	ctx := probekontext(t)
	exemplar := probe.vierStueck(t, true)

	abschluss := einAbschluss(exemplar)
	abschluss.Vermerk = &domain.Inboxeintrag{
		Schluessel:      "1\x00" + einHandel + "\x00abschluss\x001",
		Nutzlastabdruck: "abdruck-der-nachricht",
		Ergebnis:        `{"verbucht":true}`,
	}
	abschluss.Ausgang = &domain.Belegnachricht{
		NachrichtID: "n-beleg-1", Empfaenger: "bdu",
		Nachricht: `{"art":"beleg"}`,
	}

	if fehler := probe.buch.Abschliessen(ctx, abschluss); fehler != nil {
		t.Fatalf("abschließen: %v", fehler)
	}

	gelesen, fehler := probe.briefkasten.Nachschlagen(
		ctx, abschluss.Vermerk.Schluessel)
	if fehler != nil {
		t.Fatalf("Nachschlagen: %v", fehler)
	}

	if gelesen == nil || gelesen.Ergebnis != abschluss.Vermerk.Ergebnis {
		t.Errorf("der Vermerk ist %+v", gelesen)
	}

	offen, fehler := probe.ausgang.Offene(ctx, "bdu")
	if fehler != nil {
		t.Fatalf("Offene: %v", fehler)
	}

	if len(offen) != 1 || offen[0].NachrichtID != "n-beleg-1" {
		t.Fatalf("im Ausgang liegen %d Nachrichten: %+v", len(offen), offen)
	}
}
