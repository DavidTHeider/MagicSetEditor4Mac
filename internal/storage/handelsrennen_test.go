package storage_test

// **Auch der Briefkasten und der Abschluss schreiben nur gegen den gelesenen
// Stand** — Befund Z1 des Zweitblicks über Phase 6.8
// (#1669).
//
// # Der Befund
//
// `Fortschreiben` (TRADE-03.AK-01) schützte den lokalen Schritt. Briefkasten
// und Abschluss schrieben dieselbe Revision weiter über das unbedingte
// `ON CONFLICT … DO UPDATE` — eine Stimme, die gegen `angeboten` gelesen
// war, setzte einen inzwischen **angenommenen** Handel still zurück, und zwei
// Verbuchungen, die beide `uebergabeOffen` gelesen hatten, buchten beide.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// einRennhandel ist ein angebotener Tausch.
func einRennhandel() domain.Handelsangebot {
	return domain.Handelsangebot{
		Vertragsfassung: 1, Vorgang: "v-rennen", Revision: 1,
		Art: domain.HandelTausch, Initiator: "aich", Parteien: []string{"aich", "bdu"},
		Zustand: domain.HandelAngeboten,
	}
}

func TestEineZustellungUeberschreibtKeinenNeuerenStand(t *testing.T) {
	db := frischeDatenbank(t)
	briefkasten := storage.NeueBriefkastenablage(db)
	handel := storage.NeueHandelsablage(db)
	ctx := probekontext(t)

	angeboten := einRennhandel()
	if fehler := handel.Schreiben(ctx, angeboten); fehler != nil {
		t.Fatal(fehler)
	}

	// **Hier geht der Handel weiter**, während die Post `angeboten` gelesen hat.
	angenommen := angeboten
	angenommen.Zustand = domain.HandelAngenommen

	if fehler := handel.Fortschreiben(ctx, angeboten, angenommen); fehler != nil {
		t.Fatal(fehler)
	}

	mitStimme := angeboten
	mitStimme.Zustimmungen = []string{"bdu"}

	eintrag := domain.Inboxeintrag{Schluessel: "s-rennen", Nutzlastabdruck: "a", Ergebnis: "v-rennen/1"}
	if fehler := briefkasten.Zustellen(ctx, eintrag, angeboten, mitStimme,
		domain.ZusagenBleiben); !errors.Is(fehler, domain.ErrUeberholt) {
		t.Fatalf("Zustellen gegen einen alten Stand: %v", fehler)
	}

	// **Der neuere Stand bleibt, und kein Vermerk sagt „behandelt"** — sonst
	// käme die Nachricht nie wieder.
	if gelesen, fehler := handel.Vorgang(ctx, "v-rennen"); fehler != nil || gelesen.Zustand != domain.HandelAngenommen {
		t.Errorf("Stand: %+v (%v)", gelesen, fehler)
	}

	if vermerk, fehler := briefkasten.Nachschlagen(ctx, "s-rennen"); fehler != nil || vermerk != nil {
		t.Errorf("Vermerk: %+v (%v)", vermerk, fehler)
	}

	// **Gegen den richtigen Stand geht es** — dieselbe Stimme, neu gelesen.
	nachgezogen := angenommen
	nachgezogen.Zustimmungen = []string{"bdu"}

	if fehler := briefkasten.Zustellen(ctx, eintrag, angenommen, nachgezogen,
		domain.ZusagenBleiben); fehler != nil {
		t.Fatalf("Zustellen gegen den heutigen Stand: %v", fehler)
	}
}

func TestZweiErsteZustellungenLegenEinenVorgangAn(t *testing.T) {
	db := frischeDatenbank(t)
	briefkasten := storage.NeueBriefkastenablage(db)
	ctx := probekontext(t)

	erste := einRennhandel()

	if fehler := briefkasten.Zustellen(ctx, domain.Inboxeintrag{Schluessel: "s-1", Nutzlastabdruck: "a"},
		domain.Handelsangebot{}, erste, domain.ZusagenBleiben); fehler != nil {
		t.Fatal(fehler)
	}

	// **Eine zweite „erste" Revision** — unter anderem Schlüssel, etwa mit
	// anderen Positionen — findet den Vorgang schon vor.
	zweite := erste
	zweite.Positionen = []domain.Handelsposition{{Von: "bdu", Printing: "anders", Menge: 1}}

	if fehler := briefkasten.Zustellen(ctx, domain.Inboxeintrag{Schluessel: "s-2", Nutzlastabdruck: "b"},
		domain.Handelsangebot{}, zweite, domain.ZusagenBleiben); !errors.Is(fehler, domain.ErrUeberholt) {
		t.Errorf("die zweite erste Revision: %v", fehler)
	}
}

func TestEinAbschlussGegenEinenAnderenStandBuchtNichts(t *testing.T) {
	db := frischeDatenbank(t)
	probe := abschlussprobe{
		bestand: storage.NeuerSammlungsspeicher(db), zusagen: storage.NeueReservierungsablage(db),
		buch: storage.NeueAbschlussablage(db),
	}
	handel := storage.NeueHandelsablage(db)
	ctx := probekontext(t)

	exemplar := probe.vierStueck(t, true)
	abschluss := einAbschluss(exemplar)

	offen := abschluss.Angebot
	offen.Zustand = domain.HandelUebergabeOffen

	if fehler := handel.Schreiben(ctx, offen); fehler != nil {
		t.Fatal(fehler)
	}

	// **Inzwischen ging der Handel in die Klärung** — etwa, weil eine
	// widersprüchliche Nachricht kam.
	geklaert := offen
	geklaert.Zustand = domain.HandelKlaerung

	if fehler := handel.Fortschreiben(ctx, offen, geklaert); fehler != nil {
		t.Fatal(fehler)
	}

	abschluss.Vorher = offen

	if fehler := probe.buch.Abschliessen(ctx, abschluss); !errors.Is(fehler, domain.ErrUeberholt) {
		t.Fatalf("Abschliessen gegen einen alten Stand: %v", fehler)
	}

	// **Nichts ist gebucht** — der Bestand bewegte sich im selben Zug.
	if menge, da := probe.menge(t, exemplar.ID); !da || menge != 4 {
		t.Errorf("Menge %d (%v)", menge, da)
	}
}

// **Eine Folgerevision prüft ihren Vorgänger** — Befund N1 des Nachlaufs
// über Phase 6.8: eine neue Nummer belegte nur, dass die Zielzeile frei war,
// nicht, dass der gelesene Vorgänger noch galt.
func TestEineFolgerevisionEntstehtNurAufDemGelesenenVorgaenger(t *testing.T) {
	db := frischeDatenbank(t)
	handel := storage.NeueHandelsablage(db)
	ctx := probekontext(t)

	angeboten := einRennhandel()
	if fehler := handel.Schreiben(ctx, angeboten); fehler != nil {
		t.Fatal(fehler)
	}

	// **Inzwischen wurde zugestimmt** — und das Gegenangebot, das gegen
	// `angeboten` gerechnet war, kommt danach.
	angenommen := angeboten
	angenommen.Zustand = domain.HandelAngenommen

	if fehler := handel.Fortschreiben(ctx, angeboten, angenommen); fehler != nil {
		t.Fatal(fehler)
	}

	gegen := angeboten.Fortgeschrieben([]domain.Handelsposition{{Von: "bdu", Printing: "anders", Menge: 1}})

	if fehler := handel.Fortschreiben(ctx, angeboten, gegen); !errors.Is(fehler, domain.ErrUeberholt) {
		t.Fatalf("Folgerevision gegen einen alten Vorgänger: %v", fehler)
	}

	if gelesen, fehler := handel.Vorgang(ctx, "v-rennen"); fehler != nil || gelesen.Revision != 1 {
		t.Errorf("es steht Revision %d da (%v)", gelesen.Revision, fehler)
	}

	// **Gegen den heutigen Vorgänger geht sie** — hier ist die Einigung der
	// Stand, und aus ihm wird kein Gegenangebot; der Bestand sagt nur, dass
	// er nicht über einen überholten schreibt.
	if fehler := handel.Fortschreiben(ctx, angenommen, angenommen.Fortgeschrieben(
		[]domain.Handelsposition{{Von: "bdu", Printing: "anders", Menge: 1}})); fehler != nil {
		t.Errorf("gegen den heutigen Vorgänger: %v", fehler)
	}
}
