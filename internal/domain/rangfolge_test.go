package domain_test

// **Die Rangfolge über ein ganzes Turnier** — Werkstück 2 von Phase 4.7
// (#1294).
//
// **Eine eigene Datei**: `rangliste_test.go` prüft **eine** Bilanz,
// `gegnerquote_test.go` das Mittel über mehrere — hier steht, was aus einem
// **Verlauf** wird.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// turnierMit baut ein Turnier mit n durchnummerierten Teilnehmern.
func turnierMit(anzahl int) domain.Turnier {
	turnier := domain.Turnier{Name: "Küchentisch", Spiel: "magic"}

	for nummer := 1; nummer <= anzahl; nummer++ {
		turnier.Teilnehmer = append(turnier.Teilnehmer, domain.Teilnehmer{
			ID: int64(nummer), Name: string(rune('A' + nummer - 1)),
			Stelle: nummer,
		})
	}

	return turnier
}

// runde hängt eine Runde mit den gegebenen Paarungen an.
func runde(turnier *domain.Turnier, paarungen ...domain.Paarung) {
	turnier.Runden = append(turnier.Runden, domain.Runde{
		Nummer: len(turnier.Runden) + 1, Paarungen: paarungen,
	})
}

// spiel gibt eine gespielte Paarung.
func spiel(a, b int64, spieleA, spieleB int) domain.Paarung {
	return domain.Paarung{
		A: a, B: b,
		Ergebnis: &domain.Ergebnis{SpieleA: spieleA, SpieleB: spieleB},
	}
}

// **Matchpunkte entscheiden zuerst.**
func TestDieRangfolgeBeginntBeiDenMatchpunkten(t *testing.T) {
	t.Parallel()

	turnier := turnierMit(4)

	runde(&turnier, spiel(1, 2, 2, 0), spiel(3, 4, 2, 1))
	runde(&turnier, spiel(1, 3, 2, 0), spiel(2, 4, 2, 0))

	rangliste := domain.Rangliste(turnier)

	if len(rangliste) != 4 {
		t.Fatalf("%d Ränge, erwartet 4", len(rangliste))
	}

	// A hat zwei Siege (6), B und C je einen (3), D keinen (0).
	if rangliste[0].TeilnehmerID != 1 || rangliste[0].Matchpunkte != 6 {
		t.Errorf("Platz 1 ist %+v", rangliste[0])
	}

	if rangliste[3].TeilnehmerID != 4 || rangliste[3].Matchpunkte != 0 {
		t.Errorf("Platz 4 ist %+v", rangliste[3])
	}

	for stelle, rang := range rangliste {
		if rang.Platz != stelle+1 {
			t.Errorf("Rang %d trägt Platz %d", stelle+1, rang.Platz)
		}
	}
}

// **Bei gleichen Matchpunkten entscheidet die Gegnermatchquote — vor der
// eigenen Spielquote.**
//
// # Warum dieser Fall so umständlich gebaut ist
//
// Weil ein einfacher nichts beweist. Der erste Anlauf stellte B und C
// gleichauf und gab B die stärkeren Gegner — B hatte damit auch die bessere
// Spielquote, und die Rangfolge kam bei **jeder** Reihenfolge der Tiebreaker
// gleich heraus. Die Gegenprobe (2 und 3 vertauschen) blieb grün: der Test
// enthielt seinen Fall nicht.
//
// **Hier widersprechen sich die beiden.** B hat die besseren Gegner und die
// **schlechtere** Spielquote, C umgekehrt. MTR §3.1 stellt B vor C; wer die
// Reihenfolge vertauscht, bekommt C vor B.
//
//	B  gewinnt gegen A 2–1, verliert gegen D 0–2   → Spielquote 0,40
//	C  gewinnt gegen F 2–0, verliert gegen E 1–2   → Spielquote 0,60
//	A  ist 1–1, D ist 2–0, E ist 1–1, F ist 0–2
//	   → Bs Gegnerquote (0,5 + 1,0)/2  = 0,750
//	   → Cs Gegnerquote (0,33 + 0,5)/2 = 0,415
func TestBeiGleichstandEntscheidetDieGegnermatchquote(t *testing.T) {
	t.Parallel()

	turnier := turnierMit(6)

	runde(&turnier, spiel(2, 1, 2, 1), spiel(3, 6, 2, 0), spiel(4, 5, 2, 0))
	runde(&turnier, spiel(4, 2, 2, 0), spiel(5, 3, 2, 1), spiel(1, 6, 2, 0))

	nachID := map[int64]domain.Rang{}
	for _, rang := range domain.Rangliste(turnier) {
		nachID[rang.TeilnehmerID] = rang
	}

	b, c := nachID[2], nachID[3]

	// **Erst die drei Bedingungen, die den Fall ausmachen.** Ohne sie prüfte
	// der Test eine Rangfolge, die aus einem anderen Grund stimmt.
	if b.Matchpunkte != c.Matchpunkte {
		t.Fatalf("B hat %d Matchpunkte, C hat %d — der Fall trägt nicht",
			b.Matchpunkte, c.Matchpunkte)
	}

	if b.Gegnermatchquote <= c.Gegnermatchquote {
		t.Fatalf("B hat die Gegnerquote %.4f, C %.4f — der Fall trägt nicht",
			b.Gegnermatchquote, c.Gegnermatchquote)
	}

	if b.Spielquote >= c.Spielquote {
		t.Fatalf("B hat die Spielquote %.4f, C %.4f — sie müssen sich "+
			"widersprechen, sonst beweist der Fall nichts",
			b.Spielquote, c.Spielquote)
	}

	// **Und dann die Zusage.**
	if b.Platz >= c.Platz {
		t.Errorf("B steht auf Platz %d, C auf %d — die Gegnermatchquote "+
			"entscheidet vor der eigenen Spielquote (MTR §3.1)",
			b.Platz, c.Platz)
	}
}

// **Ein Bye ist ein Sieg, und der Bye-Gegner zählt nicht mit.**
func TestEinByeZaehltInDerRangliste(t *testing.T) {
	t.Parallel()

	turnier := turnierMit(3)

	runde(&turnier,
		spiel(1, 2, 2, 0),
		domain.Paarung{A: 3, B: 0}, // C bekommt den Bye
	)

	rangliste := domain.Rangliste(turnier)

	nachID := map[int64]domain.Rang{}
	for _, rang := range rangliste {
		nachID[rang.TeilnehmerID] = rang
	}

	if nachID[3].Matchpunkte != 3 {
		t.Errorf("C hat %d Matchpunkte, erwartet 3", nachID[3].Matchpunkte)
	}

	// **Und keine Gegnerquote**, denn C hatte keinen Gegner.
	if nachID[3].Gegnermatchquote != 0 {
		t.Errorf("C hat die Gegnermatchquote %v, erwartet 0 — ein Bye ist "+
			"kein Gegner", nachID[3].Gegnermatchquote)
	}
}

// **Eine Paarung ohne Ergebnis zählt nicht.**
//
// Eine laufende Runde darf die Rangliste nicht verfälschen: wer noch nicht
// gespielt hat, hat keine Niederlage.
func TestEineOffenePaarungZaehltNicht(t *testing.T) {
	t.Parallel()

	turnier := turnierMit(2)

	runde(&turnier, domain.Paarung{A: 1, B: 2})

	rangliste := domain.Rangliste(turnier)

	for _, rang := range rangliste {
		if rang.Matchpunkte != 0 || rang.Bilanz.Runden() != 0 {
			t.Errorf("%s hat %d Punkte aus %d Runden — gespielt wurde nicht",
				rang.Name, rang.Matchpunkte, rang.Bilanz.Runden())
		}
	}
}

// **Ein Ausgestiegener bleibt in der Rangliste.**
//
// Appendix C: seine Matchquote geht weiter in die Gegnerquote jedes Gegners
// ein, den er hatte. Wer ihn entfernte, veränderte die Rangliste aller
// anderen.
func TestEinAusgestiegenerBleibtInDerRangliste(t *testing.T) {
	t.Parallel()

	turnier := turnierMit(2)
	turnier.Teilnehmer[1].Ausgestiegen = true

	runde(&turnier, spiel(1, 2, 2, 0))

	rangliste := domain.Rangliste(turnier)

	if len(rangliste) != 2 {
		t.Fatalf("%d Ränge, erwartet 2 — der Ausgestiegene fehlt", len(rangliste))
	}

	if rangliste[0].Gegnermatchquote != domain.Quotenuntergrenze {
		t.Errorf("die Gegnerquote von A ist %v, erwartet die Untergrenze %v",
			rangliste[0].Gegnermatchquote, domain.Quotenuntergrenze)
	}
}

// **Bei völligem Gleichstand entscheidet die Anmeldung — nicht die
// Reihenfolge, in der der Bestand die Teilnehmer liefert.**
//
// **Das steht nicht in der MTR** — sie sagt nichts über den Fall, in dem alle
// vier Tiebreaker gleich sind. Ohne eine Regel hinge die Rangfolge daran, wie
// eine Abfrage zufällig sortiert.
//
// # Warum die Teilnehmer hier verkehrt herum stehen
//
// Weil `sort.SliceStable` die Eingabereihenfolge ohnehin hält. Ein Aufbau,
// der sie schon in Anmeldereihenfolge liefert, kann nicht zeigen, dass die
// Regel greift — die Gegenprobe blieb genau deshalb grün. Erst wenn die
// Liste **anders** ankommt als angemeldet, wird sichtbar, was `Stelle` tut.
func TestBeiVoelligemGleichstandEntscheidetDieAnmeldung(t *testing.T) {
	t.Parallel()

	turnier := domain.Turnier{Name: "Küchentisch", Spiel: "magic"}

	// **Verkehrt herum eingetragen**, mit ihren Anmeldestellen.
	turnier.Teilnehmer = []domain.Teilnehmer{
		{ID: 2, Name: "B", Stelle: 2},
		{ID: 1, Name: "A", Stelle: 1},
	}

	runde(&turnier, domain.Paarung{
		A: 1, B: 2,
		Ergebnis: &domain.Ergebnis{Unentschieden: 1},
	})

	rangliste := domain.Rangliste(turnier)

	if rangliste[0].TeilnehmerID != 1 {
		t.Errorf("Platz 1 ist %d — erwartet den zuerst Angemeldeten (1)",
			rangliste[0].TeilnehmerID)
	}
}

// **Ohne Anmeldestelle entscheidet die Reihenfolge der Liste.**
//
// Ein Turnier, das von Hand zusammengestellt wird — im Test, an der
// Kommandozeile —, hat keine Stellen. Dann ist die Liste selbst die
// Anmeldung, und das ist besser als eine Null, die alle gleichmacht.
func TestOhneStelleZaehltDieListenreihenfolge(t *testing.T) {
	t.Parallel()

	turnier := domain.Turnier{
		Name: "Küchentisch", Spiel: "magic",
		Teilnehmer: []domain.Teilnehmer{{ID: 9, Name: "I"}, {ID: 4, Name: "D"}},
	}

	runde(&turnier, domain.Paarung{
		A: 9, B: 4,
		Ergebnis: &domain.Ergebnis{Unentschieden: 1},
	})

	if erster := domain.Rangliste(turnier)[0].TeilnehmerID; erster != 9 {
		t.Errorf("Platz 1 ist %d — erwartet den zuerst Gelisteten (9)", erster)
	}
}

// **Und gemischt: eine fehlende Stelle heißt „hier in der Liste", nicht
// „ganz vorn".**
//
// # Warum dieser Fall nachgetragen wurde
//
// Weil die Gegenprobe zum Test darüber grün blieb. `sort.SliceStable` hält
// die Eingabereihenfolge, und solange **alle** ohne Stelle sind, kommt
// dasselbe heraus — der Zweig war unbetreten.
//
// **Wirksam wird er erst gemischt.** Steht ein Teilnehmer ohne Stelle
// **hinter** einem mit Stelle 1, so hieße eine Null „vor allen anderen", und
// er spränge nach vorn. Seine Listenstelle heißt: da, wo er steht.
func TestEineFehlendeStelleHeisstHierInDerListe(t *testing.T) {
	t.Parallel()

	turnier := domain.Turnier{
		Name: "Küchentisch", Spiel: "magic",
		Teilnehmer: []domain.Teilnehmer{
			{ID: 1, Name: "A", Stelle: 1},
			{ID: 2, Name: "B"}, // ohne Stelle, an zweiter Position
		},
	}

	runde(&turnier, domain.Paarung{
		A: 1, B: 2,
		Ergebnis: &domain.Ergebnis{Unentschieden: 1},
	})

	if erster := domain.Rangliste(turnier)[0].TeilnehmerID; erster != 1 {
		t.Errorf("Platz 1 ist %d — B hat keine Stelle und steht an zweiter "+
			"Position; das heißt nicht »ganz vorn«", erster)
	}
}

// **Und wenn die ersten drei gleich sind, entscheidet die Gegnerspielquote.**
//
// # Warum auch dieser Fall gebaut werden muss
//
// Die Gegenprobe zum Test darüber — Tiebreaker 4 weglassen — blieb zuerst
// grün: keiner der Fälle brauchte ihn. Ein Tiebreaker, den kein Test
// betritt, ist eine Zeile ohne Zusage.
//
// **Der Aufbau.** A und B spielen je **ein** Match und gewinnen es 2–0; damit
// sind Matchpunkte, Matchquote und Spielquote gleich. Ihre Gegner C und D
// haben dieselbe Matchbilanz (1–2) und damit dieselbe Matchquote — aber
// verschiedene **Spielquoten**, weil D eines seiner verlorenen Matches knapp
// verliert.
//
//	C  0–2 gegen A, 2–0 gegen E, 0–2 gegen F  → Spielquote 6/18  = 0,333
//	D  0–2 gegen B, 2–0 gegen G, 1–2 gegen H  → Spielquote 9/21  = 0,429
//
// Damit steht **B** vor A — obwohl A zuerst angemeldet wurde.
func TestZuletztEntscheidetDieGegnerspielquote(t *testing.T) {
	t.Parallel()

	turnier := turnierMit(8)

	runde(&turnier, spiel(1, 3, 2, 0), spiel(2, 4, 2, 0))
	runde(&turnier, spiel(3, 5, 2, 0), spiel(4, 7, 2, 0))

	// **Hier verliert der Erstgenannte** — 3 gegen 6 und 4 gegen 8. Ohne
	// eine solche Zeile bekäme `spiel` nie etwas anderes als einen Sieg für
	// A zu sehen, und die Umrechnung aus Bs Sicht bliebe unbetreten.
	runde(&turnier, spiel(3, 6, 0, 2), spiel(4, 8, 1, 2))

	nachID := map[int64]domain.Rang{}
	for _, rang := range domain.Rangliste(turnier) {
		nachID[rang.TeilnehmerID] = rang
	}

	a, b := nachID[1], nachID[2]

	// **Die ersten drei müssen gleich sein**, sonst entscheidet ein anderer.
	for _, gleich := range []struct {
		was           string
		links, rechts float64
	}{
		{"Matchpunkte", float64(a.Matchpunkte), float64(b.Matchpunkte)},
		{"Gegnermatchquote", a.Gegnermatchquote, b.Gegnermatchquote},
		{"Spielquote", a.Spielquote, b.Spielquote},
	} {
		if gegen(gleich.links, gleich.rechts) {
			t.Fatalf("%s: A hat %.6f, B hat %.6f — der Fall trägt nicht",
				gleich.was, gleich.links, gleich.rechts)
		}
	}

	if b.Gegnerspielquote <= a.Gegnerspielquote {
		t.Fatalf("A hat die Gegnerspielquote %.4f, B %.4f — der Fall trägt nicht",
			a.Gegnerspielquote, b.Gegnerspielquote)
	}

	if b.Platz >= a.Platz {
		t.Errorf("B steht auf Platz %d, A auf %d — bei sonst gleichen Werten "+
			"entscheidet die Gegnerspielquote (MTR §3.1)", b.Platz, a.Platz)
	}
}

// gegen sagt, ob zwei Quoten sich unterscheiden.
func gegen(links, rechts float64) bool {
	unterschied := links - rechts
	if unterschied < 0 {
		unterschied = -unterschied
	}

	return unterschied > 0.0005
}
