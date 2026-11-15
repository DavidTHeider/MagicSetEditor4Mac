package storage_test

// **Der Briefkasten und der Ausgang** — Werkstück `T5` von Phase 5.6
// (#1470).
//
// # Was hier gemessen wird
//
// **Dass eine zweite Zustellung dieselbe Antwort bekommt** — und zwar die
// abgelegte und keine neu gerechnete. Das ist der Unterschied zwischen einer
// Wiederholung und einem zweiten Vorgang, der zufällig gleich aussieht.
//
// **Dass derselbe Schlüssel mit anderem Abdruck ein Konflikt ist.**
//
// **Und dass ein Eintrag zu einem Profil gehört** (0066): derselbe Vorgang in
// einem anderen Profil ist ein anderer.
//
// **Kein `t.Parallel()`**, aus dem Grund, der in `sammlung_test.go` steht.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// derSchluessel ist der Idempotenzschlüssel dieser Läufe.
const derSchluessel = "1\x00v-1\x00vorschlag\x002"

// einBriefkasten stellt die Ablage über einer frischen Datenbank.
func einBriefkasten(t *testing.T) *storage.Briefkastenablage {
	t.Helper()

	return storage.NeueBriefkastenablage(frischeDatenbank(t))
}

func TestEineZweiteZustellungBekommtDieselbeAntwort(t *testing.T) {
	ablage := einBriefkasten(t)
	ctx := probekontext(t)

	eintrag := domain.Inboxeintrag{
		Schluessel: derSchluessel, Nutzlastabdruck: "abdruck-1",
		Ergebnis: `{"angenommen":true}`,
	}

	if fehler := ablage.Vermerken(ctx, "v-1", eintrag); fehler != nil {
		t.Fatalf("Vermerken: %v", fehler)
	}

	gelesen, fehler := ablage.Nachschlagen(ctx, derSchluessel)
	if fehler != nil {
		t.Fatalf("Nachschlagen: %v", fehler)
	}

	if gelesen == nil {
		t.Fatal("der Eintrag kam nicht zurück")
	}

	if gelesen.Ergebnis != eintrag.Ergebnis || gelesen.Nutzlastabdruck != "abdruck-1" {
		t.Errorf("zurück kam %+v", gelesen)
	}

	// **Die eigene Uhr steht dran** — sie entsteht im Bestand und nicht im
	// Aufrufer, damit sie nicht die des Absenders sein kann.
	if gelesen.Behandelt.IsZero() {
		t.Error("der Eintrag trägt keine Uhr")
	}

	// **Und die Wiederholung fällt darauf**: dieselbe Nachricht noch einmal.
	if art := domain.Zustellung(gelesen, "abdruck-1"); art != domain.ZustellungBekannt {
		t.Errorf("es gilt %q", art)
	}

	if art := domain.Zustellung(gelesen, "abdruck-2"); art != domain.ZustellungKonflikt {
		t.Errorf("mit anderem Abdruck gilt %q", art)
	}
}

// **Ein unbekannter Schlüssel ist kein Fehler, sondern nichts.**
func TestEinUnbekannterSchluesselGibtNichts(t *testing.T) {
	gelesen, fehler := einBriefkasten(t).Nachschlagen(probekontext(t), "gibtesnicht")
	if fehler != nil {
		t.Fatalf("Nachschlagen: %v", fehler)
	}

	if gelesen != nil {
		t.Errorf("es kam %+v", gelesen)
	}
}

// **Ein zweites Vermerken desselben Schlüssels schreibt nicht um.**
//
// **Und das ist die Sache selbst**: wer den Eintrag überschriebe, verlöre die
// erste Antwort — und gäbe beim dritten Mal eine dritte.
func TestEinZweitesVermerkenSchreibtNichtUm(t *testing.T) {
	ablage := einBriefkasten(t)
	ctx := probekontext(t)

	erster := domain.Inboxeintrag{
		Schluessel: derSchluessel, Nutzlastabdruck: "abdruck-1",
		Ergebnis: "erste Antwort",
	}

	if fehler := ablage.Vermerken(ctx, "v-1", erster); fehler != nil {
		t.Fatalf("Vermerken: %v", fehler)
	}

	zweiter := erster
	zweiter.Nutzlastabdruck, zweiter.Ergebnis = "abdruck-2", "zweite Antwort"

	if fehler := ablage.Vermerken(ctx, "v-1", zweiter); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("das zweite Vermerken gab %v", fehler)
	}

	gelesen, fehler := ablage.Nachschlagen(ctx, derSchluessel)
	if fehler != nil {
		t.Fatalf("Nachschlagen: %v", fehler)
	}

	if gelesen.Ergebnis != "erste Antwort" {
		t.Errorf("es steht %q da", gelesen.Ergebnis)
	}
}

// **Ein Eintrag gehört einem Profil** (0066).
func TestEinEintragGehoertEinemProfil(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueBriefkastenablage(db)

	eintrag := domain.Inboxeintrag{
		Schluessel: derSchluessel, Nutzlastabdruck: "abdruck-1", Ergebnis: "ja",
	}

	if fehler := ablage.Vermerken(probekontext(t), "v-1", eintrag); fehler != nil {
		t.Fatalf("Vermerken: %v", fehler)
	}

	gelesen, fehler := ablage.Nachschlagen(
		storage.MitProfil(t.Context(), 2), derSchluessel)
	if fehler != nil {
		t.Fatalf("Nachschlagen: %v", fehler)
	}

	if gelesen != nil {
		t.Errorf("das zweite Profil sah %+v", gelesen)
	}
}

// **Vermerk und Wirkung stehen oder fallen zusammen** — Werkstück T5.
//
// **Ein Vermerk ohne Wirkung hieße „schon behandelt", obwohl nichts geschehen
// ist**: die Nachricht käme nie wieder, und der Tausch bliebe stehen. Eine
// Wirkung ohne Vermerk führte die nächste Zustellung noch einmal aus.
func TestVermerkUndWirkungStehenZusammen(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueBriefkastenablage(db)
	handel := storage.NeueHandelsablage(db)
	ctx := probekontext(t)

	angebot := einZugestelltesAngebot()

	eintrag := domain.Inboxeintrag{
		Schluessel: derSchluessel, Nutzlastabdruck: "abdruck-1",
		Ergebnis: `{"angenommen":true}`,
	}

	if fehler := ablage.Zustellen(ctx, eintrag, domain.Handelsangebot{}, angebot, domain.ZusagenBleiben); fehler != nil {
		t.Fatalf("Zustellen: %v", fehler)
	}

	// **Beides steht da.**
	vermerk, fehler := ablage.Nachschlagen(ctx, derSchluessel)
	if fehler != nil || vermerk == nil {
		t.Fatalf("Nachschlagen: %v, %+v", fehler, vermerk)
	}

	gelesen, fehler := handel.Vorgang(ctx, angebot.Vorgang)
	if fehler != nil {
		t.Fatalf("Vorgang: %v", fehler)
	}

	if gelesen.Revision != angebot.Revision {
		t.Errorf("im Bestand steht Revision %d", gelesen.Revision)
	}

	// **Und der zweite Zug scheitert am Schlüssel** — das ist der Wettlauf
	// zwischen Nachschlagen und Schreiben, und der Bestand fängt ihn.
	if fehler := ablage.Zustellen(ctx, eintrag, domain.Handelsangebot{}, angebot, domain.ZusagenBleiben); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Errorf("der zweite Zug gab %v", fehler)
	}
}

// **Ohne Profil geht keine Zustellung durch.**
func TestOhneProfilGehtKeineZustellungDurch(t *testing.T) {
	ablage := storage.NeueBriefkastenablage(frischeDatenbank(t))

	if fehler := ablage.Zustellen(context.Background(),
		domain.Inboxeintrag{Schluessel: derSchluessel}, domain.Handelsangebot{},
		einZugestelltesAngebot(), domain.ZusagenBleiben); fehler == nil {
		t.Error("die Zustellung kam ohne Profil durch")
	}
}

// einZugestelltesAngebot baut die Revision, die eine Nachricht bewirkt.
func einZugestelltesAngebot() domain.Handelsangebot {
	return domain.Handelsangebot{
		Vertragsfassung: 1, Vorgang: "v-1", Revision: 2,
		Eltern: "abdruck-der-ersten",
		Art:    domain.HandelTausch, Initiator: "bich",
		Parteien: []string{"bich", "bdu"},
		Zustand:  domain.HandelAngenommen,
	}
}

// **Ein geschlossener Bestand meldet sich an beiden Stellen** — beim Beginnen
// der Zustellung und beim Schreiben der Wirkung.
func TestEineZustellungAufGeschlossenemBestandMeldetSich(t *testing.T) {
	ablage := storage.NeueBriefkastenablage(geschlossen(t))

	eintrag := domain.Inboxeintrag{
		Schluessel: derSchluessel, Nutzlastabdruck: "abdruck-1",
	}

	if fehler := ablage.Zustellen(probekontext(t), eintrag, domain.Handelsangebot{},
		einZugestelltesAngebot(), domain.ZusagenBleiben); fehler == nil {
		t.Error("der geschlossene Bestand kam durch")
	}
}

// **Eine Wirkung, die scheitert, nimmt den Vermerk mit** — sie stehen oder
// fallen zusammen, und das ist hier die fallende Hälfte.
func TestEineGescheiterteWirkungNimmtDenVermerkMit(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueBriefkastenablage(db)
	ctx := probekontext(t)

	// **Eine zweite Revision ohne Eltern** — die Prüfung im Bestand weist sie
	// ab (0068), und damit scheitert die Wirkung **nach** dem Vermerk.
	krumm := einZugestelltesAngebot()
	krumm.Eltern = ""

	eintrag := domain.Inboxeintrag{
		Schluessel: derSchluessel, Nutzlastabdruck: "abdruck-1",
	}

	if fehler := ablage.Zustellen(ctx, eintrag, domain.Handelsangebot{}, krumm, domain.ZusagenBleiben); fehler == nil {
		t.Fatal("die krumme Revision kam durch")
	}

	// **Und der Vermerk steht nicht da.** Stünde er, hieße die Nachricht
	// „schon behandelt", obwohl nichts geschehen ist.
	vermerk, fehler := ablage.Nachschlagen(ctx, derSchluessel)
	if fehler != nil {
		t.Fatalf("Nachschlagen: %v", fehler)
	}

	if vermerk != nil {
		t.Errorf("der Vermerk blieb stehen: %+v", vermerk)
	}
}
