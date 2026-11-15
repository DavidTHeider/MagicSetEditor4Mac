package domain_test

// **Was schon da war, kommt nicht zweimal** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
// Das Issue nennt vier Dinge, die blockieren statt weiterzuspielen:
// „Doppelte/geänderte IDs, Lücken, falsche Hashes und alte Epochs blockieren
// statt weiterzuspielen." Drei davon entscheidet dieser Stand; der vierte —
// der Hash — steht in der Kette.
//
// **Und einen fünften nennt es beiläufig**: „Host- und Client-Sleep/Wake,
// Netzwechsel und App-Neustart während Command/Transfer erzeugen keinen
// doppelten Draw". Ein doppelter Draw ist genau das, was eine zweimal
// zugestellte Nachricht anrichtet.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDerErsteStandNimmtDieErsteFolge(t *testing.T) {
	t.Parallel()

	stand := domain.NeuerEmpfaengerstand(1)

	urteil := stand.Pruefe(1, 1)
	if urteil != domain.FolgeNeu {
		t.Errorf("%v, erwartet %v", urteil, domain.FolgeNeu)
	}
}

func TestEineFolgeWirdNurEinmalGenommen(t *testing.T) {
	t.Parallel()

	stand := domain.NeuerEmpfaengerstand(1)
	stand.Nimm(1, 1)

	if urteil := stand.Pruefe(1, 1); urteil != domain.FolgeSchonDa {
		t.Errorf("%v, erwartet %v", urteil, domain.FolgeSchonDa)
	}
}

// **Eine Lücke blockiert** — und wird nicht etwa übersprungen.
//
// Wer zwischen 1 und 3 die 2 ausließe, hätte einen Stand, den niemand
// nachrechnen kann. Das Issue sagt „blockieren statt weiterzuspielen".
func TestEineLueckeBlockiert(t *testing.T) {
	t.Parallel()

	stand := domain.NeuerEmpfaengerstand(1)
	stand.Nimm(1, 1)

	if urteil := stand.Pruefe(1, 3); urteil != domain.FolgeLuecke {
		t.Errorf("%v, erwartet %v", urteil, domain.FolgeLuecke)
	}
}

// **Eine alte Epoche gehört zu einem Lauf, den es nicht mehr gibt.**
func TestEineAlteEpocheWirdVerworfen(t *testing.T) {
	t.Parallel()

	stand := domain.NeuerEmpfaengerstand(2)

	if urteil := stand.Pruefe(1, 1); urteil != domain.FolgeAlteEpoche {
		t.Errorf("%v, erwartet %v", urteil, domain.FolgeAlteEpoche)
	}
}

// **Eine neue Epoche fängt von vorn an** — und der Stand zieht mit.
//
// Sie entsteht, wenn der Wirt neu anfängt; die Folge beginnt dann wieder bei
// eins, und wer die alte Folge weiterzählte, sähe eine Lücke, wo keine ist.
func TestEineNeueEpocheFaengtVonVornAn(t *testing.T) {
	t.Parallel()

	stand := domain.NeuerEmpfaengerstand(1)
	stand.Nimm(1, 7)

	if urteil := stand.Pruefe(2, 1); urteil != domain.FolgeNeu {
		t.Errorf("%v, erwartet %v", urteil, domain.FolgeNeu)
	}

	stand.Nimm(2, 1)
	if stand.Epoche() != 2 || stand.Folge() != 1 {
		t.Errorf("Epoche %d, Folge %d — erwartet 2 und 1",
			stand.Epoche(), stand.Folge())
	}
}

// **Was der Stand nicht nimmt, sagt er mit einem Urteil** — und das Urteil
// macht daraus einen Fehler, wo einer gebraucht wird.
func TestNimmMeldetWasNichtGeht(t *testing.T) {
	t.Parallel()

	stand := domain.NeuerEmpfaengerstand(1)
	stand.Nimm(1, 1)

	faelle := map[string]struct {
		epoche, folge uint64
	}{
		"schon da":    {1, 1},
		"Lücke":       {1, 5},
		"alte Epoche": {0, 1},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			urteil := stand.Nimm(fall.epoche, fall.folge)
			if urteil == domain.FolgeNeu {
				t.Fatal("es kam durch")
			}

			if fehler := urteil.Fehler(fall.epoche, fall.folge); !errors.Is(
				fehler, domain.ErrConflict,
			) {
				t.Errorf("es kam %v, erwartet ein Konflikt", fehler)
			}
		})
	}
}

// **Und was durchgeht, ist kein Fehler.**
func TestEinNeuesUrteilIstKeinFehler(t *testing.T) {
	t.Parallel()

	if fehler := domain.FolgeNeu.Fehler(1, 1); fehler != nil {
		t.Errorf("es kam %v, erwartet keiner", fehler)
	}
}

// **Und was geht, ändert den Stand wirklich.**
func TestNimmSchreibtDenStandFort(t *testing.T) {
	t.Parallel()

	stand := domain.NeuerEmpfaengerstand(1)

	for folge := uint64(1); folge <= 3; folge++ {
		if urteil := stand.Nimm(1, folge); urteil != domain.FolgeNeu {
			t.Fatalf("Folge %d: %v", folge, urteil)
		}
	}

	if stand.Folge() != 3 {
		t.Errorf("Folge %d, erwartet 3", stand.Folge())
	}
}

// **Ein Urteil hat einen Namen** — für Meldungen und Protokolle.
func TestJedesUrteilHatEinenNamen(t *testing.T) {
	t.Parallel()

	urteile := []domain.Folgeurteil{
		domain.FolgeNeu, domain.FolgeSchonDa,
		domain.FolgeLuecke, domain.FolgeAlteEpoche,
	}

	gesehen := map[string]bool{}

	for _, urteil := range urteile {
		name := urteil.String()
		switch {
		case name == "":
			t.Errorf("%d hat keinen Namen", urteil)
		case gesehen[name]:
			t.Errorf("%q kommt zweimal vor", name)
		}

		gesehen[name] = true
	}
}

// **Eine neue Epoche, die nicht bei eins anfängt, ist eine Lücke** — der
// Wirt hat neu angefangen, und der Anfang fehlt.
func TestEineNeueEpocheMussBeiEinsAnfangen(t *testing.T) {
	t.Parallel()

	stand := domain.NeuerEmpfaengerstand(1)

	if urteil := stand.Pruefe(2, 5); urteil != domain.FolgeLuecke {
		t.Errorf("%v, erwartet %v", urteil, domain.FolgeLuecke)
	}
}

// **Und was kein Urteil ist, sagt das auch.**
func TestEinFremdesUrteilHeisstUnbekannt(t *testing.T) {
	t.Parallel()

	if domain.Folgeurteil(99).String() != "unbekannt" {
		t.Error("eine fremde Zahl gibt keinen Namen her")
	}
}
