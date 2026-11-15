package domain_test

// **Ein Befund ist kein Auftrag** — Werkstück O8 von Phase 5.2
// (#1432).

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// pflegezeit ist der Augenblick, in dem diese Läufe stehen.
var pflegezeit = time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)

func befund(art string, klasse domain.Pflegeklasse, anzahl int, revision string) domain.Pflegebefund {
	return domain.Pflegebefund{Art: art, Klasse: klasse, Anzahl: anzahl, Revision: revision}
}

// wahl baut die Wahl aus Marken und Revisionen.
func wahl(paare ...string) []domain.Pflegewahl {
	gewaehlt := make([]domain.Pflegewahl, 0, len(paare)/2)
	for stelle := 0; stelle+1 < len(paare); stelle += 2 {
		gewaehlt = append(gewaehlt,
			domain.Pflegewahl{Art: paare[stelle], Revision: paare[stelle+1]})
	}

	return gewaehlt
}

// **Nur `sicher` läuft ohne Entscheidung.** Die anderen zwei sind der Grund,
// warum es diese Klassen gibt: was aussieht wie ein Fehler, ist manchmal keiner.
func TestNurSichereBefundeLaufenOhneEntscheidung(t *testing.T) {
	t.Parallel()

	for klasse, erwartet := range map[domain.Pflegeklasse]bool{
		domain.PflegeSicher:         true,
		domain.PflegeEntscheidung:   false,
		domain.PflegeNurInformieren: false,
	} {
		t.Run(string(klasse), func(t *testing.T) {
			t.Parallel()

			if machbar := befund("x", klasse, 1, "abc").OhneRueckfrage(); machbar != erwartet {
				t.Errorf("%q gibt %v, erwartet war %v", klasse, machbar, erwartet)
			}
		})
	}
}

// **`nurInformieren` lässt sich gar nicht ausführen** — auch nicht auf
// ausdrückliche Wahl. Das ist die Zusage aus dem Issue: echte Korruption wird
// nicht durch kosmetisches Löschen verdeckt.
func TestNurInformierenLaesstSichNichtAusfuehren(t *testing.T) {
	t.Parallel()

	zeilen, fehler := domain.Pflegeabgleich(
		pflegezeit,
		[]domain.Pflegebefund{befund(domain.PflegeDubletten, domain.PflegeNurInformieren, 4, "abc")},
		wahl(domain.PflegeDubletten, "abc"), pflegezeit,
	)
	if fehler != nil {
		t.Fatalf("abgleichen: %v", fehler)
	}

	if len(zeilen) != 1 || zeilen[0].Ausgang != domain.PflegeUebersprungen {
		t.Fatalf("die Zeile ist %+v", zeilen)
	}

	if zeilen[0].Grund != domain.PflegegrundNurInformieren {
		t.Errorf("der Grund ist %q", zeilen[0].Grund)
	}
}

// **Eine geänderte Revision stoppt die betroffene Maßnahme — und nur sie.**
func TestEineGeaenderteRevisionStopptNurIhreMassnahme(t *testing.T) {
	t.Parallel()

	zeilen, fehler := domain.Pflegeabgleich(
		pflegezeit,
		[]domain.Pflegebefund{
			befund("kartenwerteOhneKarte", domain.PflegeSicher, 5, "xyz"),
			befund("suchzeilenOhneKarte", domain.PflegeSicher, 2, "def"),
		},
		wahl("kartenwerteOhneKarte", "abc", "suchzeilenOhneKarte", "def"), pflegezeit,
	)
	if fehler != nil {
		t.Fatalf("abgleichen: %v", fehler)
	}

	ausgaenge := map[string]domain.Pflegeausgang{}
	for _, zeile := range zeilen {
		ausgaenge[zeile.Art] = zeile.Ausgang
	}

	if ausgaenge["kartenwerteOhneKarte"] != domain.PflegeKonflikt {
		t.Errorf("die geänderte Maßnahme ist %q", ausgaenge["kartenwerteOhneKarte"])
	}

	if ausgaenge["suchzeilenOhneKarte"] != domain.PflegeBereit {
		t.Errorf("die unveränderte Maßnahme ist %q", ausgaenge["suchzeilenOhneKarte"])
	}
}

// **Ein Befund, den es nicht mehr gibt, ist erledigt und kein Konflikt.**
//
// Wer ihn als Konflikt meldete, schickte jemanden nachsehen, wo nichts mehr
// ist.
func TestEinVerschwundenerBefundIstErledigt(t *testing.T) {
	t.Parallel()

	zeilen, fehler := domain.Pflegeabgleich(
		pflegezeit, nil, wahl("kartenwerteOhneKarte", "abc"), pflegezeit,
	)
	if fehler != nil {
		t.Fatalf("abgleichen: %v", fehler)
	}

	if len(zeilen) != 1 || zeilen[0].Ausgang != domain.PflegeSchonWeg {
		t.Errorf("die Zeile ist %+v", zeilen)
	}
}

// **Die Klasse kommt aus der frischen Analyse und nicht aus der Wahl.**
//
// Wer sie mitschickte, könnte einen `nurInformieren`-Befund als `sicher`
// ausgeben — und die Zusage hinge daran, dass der Aufrufer die Wahrheit sagt.
func TestDieKlasseKommtAusDerAnalyseUndNichtAusDerWahl(t *testing.T) {
	t.Parallel()

	zeilen, fehler := domain.Pflegeabgleich(
		pflegezeit,
		[]domain.Pflegebefund{
			befund("offeneRueckkehren", domain.PflegeNurInformieren, 1, "abc"),
		},
		wahl("offeneRueckkehren", "abc"), pflegezeit,
	)
	if fehler != nil {
		t.Fatalf("abgleichen: %v", fehler)
	}

	if len(zeilen) != 1 || zeilen[0].Ausgang != domain.PflegeUebersprungen {
		t.Errorf("die Zeile ist %+v", zeilen)
	}
}

// **Ein abgelaufener Plan wird als Ganzes abgewiesen** — die Frist ist die
// Zusage, dass die Vorschau noch etwas beschreibt.
func TestEinAbgelaufenerPlanWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	_, fehler := domain.Pflegeabgleich(
		pflegezeit,
		[]domain.Pflegebefund{befund("kartenwerteOhneKarte", domain.PflegeSicher, 3, "abc")},
		wahl("kartenwerteOhneKarte", "abc"),
		pflegezeit.Add(domain.Pflegefrist).Add(time.Second),
	)

	if !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("der Fehler ist %v", fehler)
	}
}

// **Genau an der Frist gilt er noch.** Eine Grenze, die man nur von einer
// Seite prüft, ist eine halbe Zusage (§10).
func TestGenauAnDerFristGiltDerPlanNoch(t *testing.T) {
	t.Parallel()

	if _, fehler := domain.Pflegeabgleich(
		pflegezeit,
		[]domain.Pflegebefund{befund("kartenwerteOhneKarte", domain.PflegeSicher, 3, "abc")},
		wahl("kartenwerteOhneKarte", "abc"),
		pflegezeit.Add(domain.Pflegefrist),
	); fehler != nil {
		t.Errorf("genau an der Frist: %v", fehler)
	}
}

// **Ohne Wahl geschieht nichts.** Ein leerer Plan ist keine Erlaubnis für
// alles.
func TestOhneWahlGeschiehtNichts(t *testing.T) {
	t.Parallel()

	zeilen, fehler := domain.Pflegeabgleich(
		pflegezeit,
		[]domain.Pflegebefund{befund("kartenwerteOhneKarte", domain.PflegeSicher, 3, "abc")},
		nil, pflegezeit,
	)
	if fehler != nil {
		t.Fatalf("abgleichen: %v", fehler)
	}

	if len(zeilen) != 0 {
		t.Errorf("ohne Wahl entstanden Zeilen: %+v", zeilen)
	}
}

// **Eine erfundene Art wird abgewiesen und nicht als „schon weg" gemeldet.**
//
// Beides löscht nichts — aber „es gibt diesen Befund nicht mehr" und „es gibt
// ihn gar nicht" sind zwei Auskünfte, und wer die erste bekommt, sucht an der
// falschen Stelle.
func TestEineErfundeneArtWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	_, fehler := domain.Pflegeabgleich(
		pflegezeit, nil, wahl("alleKartenLoeschen", "abc"), pflegezeit,
	)

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("der Fehler ist %v", fehler)
	}
}

// **Dreizehn Arten, jede genau einmal.**
func TestEsGibtDreizehnPflegearten(t *testing.T) {
	t.Parallel()

	arten := domain.Pflegearten()
	if len(arten) != 13 {
		t.Fatalf("es sind %d Arten: %v", len(arten), arten)
	}

	gesehen := map[string]bool{}

	for _, art := range arten {
		if gesehen[art] {
			t.Errorf("%q steht zweimal in der Reihe", art)
		}

		gesehen[art] = true

		if !domain.IstPflegeart(art) {
			t.Errorf("%q gilt nicht als Art", art)
		}
	}

	if domain.IstPflegeart("schallplatten") {
		t.Error("eine erfundene Art gilt als Art")
	}
}
