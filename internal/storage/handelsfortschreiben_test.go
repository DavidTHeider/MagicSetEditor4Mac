package storage_test

// **Zwei Schreiber, ein Handel** — TRADE-03.AK-01
// (#1717).
//
// # Was hier gemessen wird
//
// **Dass eine Änderung nicht still verloren geht.** Fenster und
// Kommandozeile, oder ein Schritt hier und eine Stimme, die gerade über die
// Leitung ankommt: wer von einem Stand aus schreibt, den es nicht mehr gibt,
// bekommt ein `ErrUeberholt` statt die Änderung des anderen zu überschreiben.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func einAngebotZumFortschreiben() domain.Handelsangebot {
	return domain.Handelsangebot{
		Vertragsfassung: 1, Vorgang: "v-zwei", Revision: 1,
		Art: domain.HandelTausch, Initiator: "aich", Parteien: []string{"aich", "bdu"},
		Zustand: domain.HandelAngeboten,
	}
}

func TestEineVeralteteRevisionWirdNichtGeschrieben(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueHandelsablage(db)
	ctx := probekontext(t)

	vorher := einAngebotZumFortschreiben()
	if fehler := ablage.Schreiben(ctx, vorher); fehler != nil {
		t.Fatal(fehler)
	}

	// **Bert stimmt zu** — von demselben Stand aus, und zuerst.
	mitBert := vorher
	mitBert.Zustimmungen = []string{"bdu"}

	if fehler := ablage.Fortschreiben(ctx, vorher, mitBert); fehler != nil {
		t.Fatalf("Bert: %v", fehler)
	}

	// **Anna bricht ab** — vom alten Stand, ohne Berts Stimme zu kennen.
	abgebrochen := vorher
	abgebrochen.Zustand = domain.HandelAbgebrochen

	if fehler := ablage.Fortschreiben(ctx, vorher, abgebrochen); !errors.Is(fehler, domain.ErrUeberholt) {
		t.Fatalf("Anna vom alten Stand: %v", fehler)
	}

	gelesen, fehler := ablage.Vorgang(ctx, "v-zwei")
	if fehler != nil || gelesen.Zustand != domain.HandelAngeboten || len(gelesen.Zustimmungen) != 1 {
		t.Fatalf("Berts Stimme ist fort: %+v, %v", gelesen, fehler)
	}

	// **Vom neuen Stand aus geht es.**
	if fehler := ablage.Fortschreiben(ctx, mitBert, abgebrochen); fehler != nil {
		t.Fatalf("Anna vom neuen Stand: %v", fehler)
	}
}

func TestEineNeueRevisionEntstehtEinmal(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueHandelsablage(db)
	ctx := probekontext(t)

	vorher := einAngebotZumFortschreiben()
	if fehler := ablage.Schreiben(ctx, vorher); fehler != nil {
		t.Fatal(fehler)
	}

	zweite := vorher
	zweite.Revision, zweite.Eltern, zweite.Zustand = 2, vorher.Abdruck(), domain.HandelVerhandlung

	if fehler := ablage.Fortschreiben(ctx, vorher, zweite); fehler != nil {
		t.Fatalf("die zweite Revision: %v", fehler)
	}

	// **Wer dieselbe zweite Revision noch einmal anlegen will**, kam vom
	// alten Stand — der andere war schneller.
	anders := zweite
	anders.Zustimmungen = []string{"aich"}

	if fehler := ablage.Fortschreiben(ctx, vorher, anders); !errors.Is(fehler, domain.ErrUeberholt) {
		t.Fatalf("dieselbe Revision zweimal: %v", fehler)
	}
}

func TestFortschreibenOhneProfil(t *testing.T) {
	ablage := storage.NeueHandelsablage(frischeDatenbank(t))
	vorher := einAngebotZumFortschreiben()

	if fehler := ablage.Fortschreiben(t.Context(), vorher, vorher); fehler == nil {
		t.Fatal("geschrieben ohne Profil")
	}
}
