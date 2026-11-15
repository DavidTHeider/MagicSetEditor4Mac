package domain_test

import (
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Muss wiederholt werden, dann so selten wie möglich.**
//
// Der erste Anlauf suchte eine wiederholungsfreie Paarung und fiel, wenn es
// keine gab, auf das kurzsichtige Der-Reihe-nach zurück — und das nimmt mehr
// Wiederholungen in Kauf, als nötig sind (Befund 6 des Nachlaufs über Phase
// 4.7).
//
// **Zugesichert wird die Eigenschaft, nicht ein Plan.** Welche Paarung
// herauskommt, hängt an der Standreihenfolge, und die hängt an den
// Ergebnissen; ein festgeschriebener Plan prüfte die Rangliste mit. Der Test
// rechnet das Minimum deshalb selbst aus — über **alle** Paarungen derselben
// Standreihenfolge — und vergleicht.
func TestEineUnvermeidlicheWiederholungBleibtSoSeltenWieMoeglich(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name   string
		frisch [][2]int64
	}{
		{
			"drei freie Paare, die sich in die Quere kommen",
			[][2]int64{{2, 3}, {2, 4}, {1, 3}},
		},
		{
			"zwei freie Paare am selben Teilnehmer",
			[][2]int64{{1, 3}, {1, 4}, {2, 3}},
		},
		// **Der gemessene Fall** (Maske 2 aus dem Durchgang über alle 32.768
		// Verläufe): das Kurzsichtige gibt drei Wiederholungen, zwei reichen.
		{"ein einziges freies Paar", [][2]int64{{1, 3}}},
		{"gar keines", nil},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			frisch := map[[2]int64]bool{}
			for _, paar := range fall.frisch {
				frisch[paar] = true
				frisch[[2]int64{paar[1], paar[0]}] = true
			}

			turnier := sechsMitVerlauf(frisch)
			plan := domain.Paarungsplan(turnier)

			stand := make([]int64, 0, 6)
			for _, rang := range domain.Rangliste(turnier) {
				stand = append(stand, rang.TeilnehmerID)
			}

			moeglich := wenigsteMoeglich(stand, frisch)

			if gezaehlt := wiederholungen(plan, frisch); gezaehlt != moeglich {
				t.Errorf("%d Wiederholungen in %s, möglich sind %d",
					gezaehlt, alsText(plan), moeglich)
			}
		})
	}
}

// sechsMitVerlauf baut sechs Teilnehmer, gegen die alles gespielt ist außer
// den genannten Paaren.
func sechsMitVerlauf(frisch map[[2]int64]bool) domain.Turnier {
	turnier := domain.Turnier{Name: "Küchentisch", Spiel: "magic"}

	for nummer := int64(1); nummer <= 6; nummer++ {
		turnier.Teilnehmer = append(turnier.Teilnehmer, domain.Teilnehmer{
			ID: nummer, Name: fmt.Sprint(nummer), Stelle: int(nummer),
		})
	}

	var gespielte []domain.Paarung

	for a := int64(1); a <= 6; a++ {
		for b := a + 1; b <= 6; b++ {
			if frisch[[2]int64{a, b}] {
				continue
			}

			gespielte = append(gespielte, domain.Paarung{
				A: a, B: b, Ergebnis: &domain.Ergebnis{SpieleA: 2},
			})
		}
	}

	turnier.Runden = []domain.Runde{{Nummer: 1, Paarungen: gespielte}}

	return turnier
}

// wiederholungen zählt die Paarungen, die es schon einmal gab.
func wiederholungen(plan []domain.Paarung, frisch map[[2]int64]bool) int {
	var gezaehlt int

	for _, paarung := range plan {
		if !paarung.IstBye() && !frisch[[2]int64{paarung.A, paarung.B}] {
			gezaehlt++
		}
	}

	return gezaehlt
}

// wenigsteMoeglich rechnet das Minimum über alle Paarungen aus — roh und
// vollständig, weil sechs Teilnehmer fünfzehn Paarungen ergeben.
func wenigsteMoeglich(offen []int64, frisch map[[2]int64]bool) int {
	if len(offen) < 2 {
		return 0
	}

	bestes := len(offen)

	for stelle := 1; stelle < len(offen); stelle++ {
		rest := make([]int64, 0, len(offen)-2)
		rest = append(rest, offen[1:stelle]...)
		rest = append(rest, offen[stelle+1:]...)

		kosten := 0
		if !frisch[[2]int64{offen[0], offen[stelle]}] {
			kosten = 1
		}

		if zusammen := kosten + wenigsteMoeglich(rest, frisch); zusammen < bestes {
			bestes = zusammen
		}
	}

	return bestes
}

// alsText macht einen Plan lesbar.
func alsText(plan []domain.Paarung) string {
	satz := ""

	for _, paarung := range plan {
		satz += fmt.Sprintf(" %d-%d", paarung.A, paarung.B)
	}

	return satz
}
