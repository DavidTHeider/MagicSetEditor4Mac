package storage_test

// **Ein zweiter Abschluss bewegt nichts — bei jedem Zuschnitt** — C05
// (#2230).
//
// `TestEinZweiterAbschlussBewegtNichts` prüft eine Besitzzeile und einen
// Zugang. Hier bestimmt ein Samen, wie viele Besitzzeilen mit welchen Mengen
// weggehen — ganz oder teilweise — und wie viele Zugänge kommen. Zugesagt
// ist zweierlei:
//
//  1. Der erste Abschluss bucht genau, was im Beleg steht: Rest bleibt, eine
//     ganz abgegebene Zeile verschwindet, jeder Zugang steht da.
//  2. Der zweite — mit neuer Nachrichtenkennung — ist ein Konflikt, und
//     danach stehen Bestand und Journal Zeile für Zeile wie vorher.

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// bestandsbild ist die Sammlung als Zeilen „Druck Menge Fassung Lagerort".
func (p abschlussprobe) bestandsbild(t *testing.T) []string {
	t.Helper()

	liste, fehler := p.bestand.Liste(probekontext(t))
	if fehler != nil {
		t.Fatalf("Liste: %v", fehler)
	}

	bild := make([]string, 0, len(liste))
	for _, exemplar := range liste {
		bild = append(bild, fmt.Sprintf("%d %s %d %d %s",
			exemplar.ID, exemplar.Printing, exemplar.Menge, exemplar.Fassung, exemplar.Lagerort))
	}

	slices.Sort(bild)

	return bild
}

func TestEinZweiterAbschlussBewegtNichtsBeiJedemZuschnitt(t *testing.T) {
	for samen := range uint64(30) {
		t.Run(fmt.Sprintf("Samen %d", samen), func(t *testing.T) {
			zufall := rand.New(rand.NewPCG(samen, 2230)) //nolint:gosec // eine Folge für einen Test, kein Geheimnis
			probe := eineAbschlussprobe(t)
			ctx := probekontext(t)
			abschluss := einAbschluss(&domain.Exemplar{})
			abschluss.Beleg.Zeilen = nil

			rest := map[string]int{}

			for nummer := range 1 + zufall.IntN(4) {
				menge := 1 + zufall.IntN(5)
				weg := 1 + zufall.IntN(menge)
				druck := fmt.Sprintf("abgang-%d", nummer)

				kennung, fehler := probe.bestand.Anlegen(ctx, &domain.Exemplar{
					Printing: druck, Menge: menge, Zustand: domain.ZustandNearMint,
					Sprache: "de", Lagerort: "Kiste 4",
				})
				if fehler != nil {
					t.Fatalf("anlegen: %v", fehler)
				}

				if fehler := probe.zusagen.Festhalten(ctx, domain.Reservierung{
					Vorgang: einHandel, Revision: 1, ExemplarID: kennung, Menge: weg,
					Fassung: 1, Frist: time.Now().Add(time.Hour), Uebergeben: true,
				}, ueber(&domain.Exemplar{Printing: druck, Zustand: domain.ZustandNearMint, Sprache: "de"}, 99)); fehler != nil {
					t.Fatalf("festhalten: %v", fehler)
				}

				abschluss.Beleg.Zeilen = append(abschluss.Beleg.Zeilen, domain.Journalzeile{
					Art: domain.JournalAbgang, Anlass: domain.AnlassTausch, Vorgang: einHandel,
					Printing: druck, Menge: weg, Zustand: domain.ZustandNearMint, Sprache: "de",
					Finish: domain.FinishNonfoil, ExemplarID: kennung, Lagerort: "Kiste 4",
				})

				if menge > weg {
					rest[druck] = menge - weg
				}
			}

			for nummer := range zufall.IntN(4) {
				druck := fmt.Sprintf("zugang-%d", nummer)
				menge := 1 + zufall.IntN(3)

				abschluss.Beleg.Zeilen = append(abschluss.Beleg.Zeilen, domain.Journalzeile{
					Art: domain.JournalZugang, Anlass: domain.AnlassTausch, Vorgang: einHandel,
					Printing: druck, Menge: menge, Zustand: domain.ZustandNearMint, Sprache: "en",
					Finish: domain.FinishNonfoil,
				})
				rest[druck] = menge
			}

			if fehler := probe.buch.Abschliessen(ctx, abschluss); fehler != nil {
				t.Fatalf("erster Abschluss: %v", fehler)
			}

			// **1. Gebucht ist genau der Beleg.**
			liste, fehler := probe.bestand.Liste(ctx)
			if fehler != nil {
				t.Fatalf("Liste: %v", fehler)
			}

			gebucht := map[string]int{}
			for _, exemplar := range liste {
				gebucht[exemplar.Printing] += exemplar.Menge
			}

			for druck, menge := range rest {
				if gebucht[druck] != menge {
					t.Errorf("%s: %d Stück statt %d", druck, gebucht[druck], menge)
				}
			}

			if len(gebucht) != len(rest) {
				t.Errorf("im Bestand %v, erwartet %v", gebucht, rest)
			}

			vorher := probe.bestandsbild(t)

			zeilenVorher, fehler := probe.buch.Journal(ctx, einHandel)
			if fehler != nil {
				t.Fatalf("Journal: %v", fehler)
			}

			// **2. Der zweite Abschluss ist ein Konflikt und bewegt nichts.**
			abschluss.Vermerk = &domain.Inboxeintrag{
				Schluessel: fmt.Sprintf("neu-%d", samen), Nutzlastabdruck: "abdruck-neu",
				Ergebnis: `{"verbucht":true}`,
			}

			if zweiter := probe.buch.Abschliessen(ctx, abschluss); !errors.Is(zweiter, domain.ErrConflict) {
				t.Fatalf("der zweite Abschluss gab %v statt eines Konflikts", zweiter)
			}

			if nachher := probe.bestandsbild(t); !slices.Equal(nachher, vorher) {
				t.Errorf("der zweite Abschluss hat gebucht:\n%q\nvorher\n%q", nachher, vorher)
			}

			zeilenNachher, fehler := probe.buch.Journal(ctx, einHandel)
			if fehler != nil {
				t.Fatalf("Journal: %v", fehler)
			}

			// **Zeile für Zeile, jedes Feld** (Zweitblick über Phase C,
			// Befund 8): die Zahl allein ließe eine veränderte Zeile durch.
			if !reflect.DeepEqual(zeilenNachher, zeilenVorher) {
				t.Errorf("das Journal hat sich verändert:\n%+v\nvorher\n%+v", zeilenNachher, zeilenVorher)
			}
		})
	}
}
