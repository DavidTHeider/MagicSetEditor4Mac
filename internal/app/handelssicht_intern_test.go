package app

// **Die Sicht eines Gegenübers über die Leitung, an der Naht** — P11
// (#2379).
//
// Zwei Anwendungen über eine Schleife: Bert gibt Anna seine Sammlung zum
// Handeln frei, Anna fragt — und bekommt genau das, was Bert ihr zeigt. Nach
// dem Widerruf eine Absage.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// beantwortet lässt die Gegenseite annehmen und die Sitzung behandeln, wie
// es ihr Horchposten täte.
func beantwortet(t *testing.T, dort *seite) <-chan error {
	t.Helper()

	fertig := make(chan error, 1)

	go func() {
		griff, fehler := dort.anwendung.sitzungsverwaltung().Annehmen(dort.ctx, dort.leitung)
		if fehler == nil {
			dort.anwendung.sitzungAufnehmen(dort.ctx, griff)
		}

		fertig <- fehler
	}()

	return fertig
}

func TestEineSichtKommtUeberDieLeitung(t *testing.T) {
	t.Parallel()

	hier, dort := zweiAnwendungen(t)

	if _, fehler := storage.NeuerSammlungsspeicher(dort.anwendung.db).Anlegen(dort.ctx, &domain.Exemplar{
		Printing: "00000000-0000-4000-8000-0000000b1002", Menge: 2,
		Zustand: domain.ZustandNearMint, Sprache: "de",
	}); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if _, fehler := storage.NeuerSammlungsspeicher(dort.anwendung.db).Anlegen(dort.ctx, &domain.Exemplar{
		Printing: "00000000-0000-4000-8000-0000000a1001", Menge: 1,
		Zustand: domain.ZustandNearMint, Sprache: "de", Wunsch: true,
	}); fehler != nil {
		t.Fatalf("Wunsch anlegen: %v", fehler)
	}

	for _, umfang := range []string{"sammlung", "wuensche"} {
		if erteilt := dort.anwendung.SammlungFreigeben(dort.ctx, SammlungsfreigabeDTO{
			Empfaenger: string(hier.kennung), Umfang: umfang, Zweck: "handel",
		}); erteilt.Fehler != nil {
			t.Fatalf("freigeben: %+v", erteilt.Fehler)
		}
	}

	fertig := beantwortet(t, dort)
	antwort := hier.anwendung.HandelSichtAnfragen(hier.ctx, string(dort.kennung))

	if fehler := <-fertig; fehler != nil {
		t.Fatalf("die Gegenseite: %v", fehler)
	}

	if antwort.Fehler != nil || len(antwort.Zeilen) != 2 {
		t.Fatalf("Antwort: %+v", antwort)
	}

	// **Das Angebot und der Wunsch, und der Wunsch sagt, dass er einer ist.**
	for _, zeile := range antwort.Zeilen {
		angebot := zeile.Printing == "00000000-0000-4000-8000-0000000b1002" && zeile.Menge == 2 && !zeile.Wunsch
		gesucht := zeile.Printing == "00000000-0000-4000-8000-0000000a1001" && zeile.Wunsch

		if !angebot && !gesucht {
			t.Errorf("Zeile: %+v", zeile)
		}
	}

	// **Nach dem Widerruf** — dieselbe Leitung gibt es nicht mehr; eine neue.
	if widerrufen := dort.anwendung.SammlungsfreigabeWiderrufen(dort.ctx,
		string(hier.kennung), "handel", "sammlung"); widerrufen.Fehler != nil {
		t.Fatalf("widerrufen: %+v", widerrufen.Fehler)
	}
}

func TestOhneFreigabeIstDieSichtEineAbsage(t *testing.T) {
	t.Parallel()

	hier, dort := zweiAnwendungen(t)

	fertig := beantwortet(t, dort)
	antwort := hier.anwendung.HandelSichtAnfragen(hier.ctx, string(dort.kennung))

	if fehler := <-fertig; fehler != nil {
		t.Fatalf("annehmen: %v", fehler)
	}

	if antwort.Fehler == nil || antwort.Fehler.Code != CodeErlaubnis || antwort.Zeilen == nil || len(antwort.Zeilen) != 0 {
		t.Fatalf("ohne Freigabe: %+v", antwort)
	}

	// **Ein eigener Satz und nicht der über die Rolle** — gefunden am
	// nativen Fenster (`p11sicht09`): dort stand „Dazu fehlt dir die
	// Rolle", und das Gegenüber hatte nur nichts freigegeben.
	if antwort.Fehler.Schluessel != domain.SchluesselNichtFreigegeben {
		t.Errorf("der Schlüssel ist %q", antwort.Fehler.Schluessel)
	}
}

func TestEineSichtanfrageBrauchtWirkungUndKennung(t *testing.T) {
	t.Parallel()

	wache := &abweisendeWache{}
	if antwort := freigabeanwendung(t).MitWache(wache).HandelSichtAnfragen(t.Context(), "quatsch"); antwort.Fehler == nil ||
		wache.gefragt != domain.WirkungExternSenden {
		t.Errorf("ohne Wirkung: %+v, gefragt %q", antwort, wache.gefragt)
	}

	if antwort := freigabeanwendung(t).HandelSichtAnfragen(t.Context(), "quatsch"); antwort.Fehler == nil ||
		antwort.Fehler.Code != CodeValidierung {
		t.Errorf("ohne Peerkennung: %+v", antwort)
	}

	// **Ohne Kontakt kommt keine Leitung zustande** — und keine Zeile.
	jener, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if antwort := freigabeanwendung(t).HandelSichtAnfragen(t.Context(), string(jener.Kennung())); antwort.Fehler == nil ||
		antwort.Zeilen == nil || len(antwort.Zeilen) != 0 {
		t.Errorf("ohne Kontakt: %+v", antwort)
	}
}
