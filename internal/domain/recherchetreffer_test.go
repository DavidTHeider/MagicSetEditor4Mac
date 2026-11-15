package domain_test

// **Bündeln und Ausschließen** — Werkstück A16 von Phase 5.3
// (#1515).
//
// **Zwei Regeln, die beide aus dem Issue kommen**: „Oracle-Duplikate
// nachvollziehbar bündeln" und „Eigene Ausgangskarte aus
// Ähnlichkeitstreffern ausschließen".

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einRechercheteil baut einen Treffer.
func einRechercheteil(kennung, fachkennung string, rang int) domain.Recherchetreffer {
	return domain.Recherchetreffer{
		Rang: rang, Kennung: kennung,
		Kennungssystem: domain.SystemScryfall, Fachkennung: fachkennung,
		Semantik: domain.SemantikKonzept, Quelle: domain.QuelleOracle,
	}
}

// **Zwei Seiten derselben Karte sind ein Treffer.**
func TestZweiSeitenDerselbenKarteSindEinTreffer(t *testing.T) {
	t.Parallel()

	gebuendelt := domain.GebuendeltNach([]domain.Recherchetreffer{
		einRechercheteil("o1:0", "o1", 1),
		einRechercheteil("o2:0", "o2", 2),
		einRechercheteil("o1:1", "o1", 3),
	})

	if len(gebuendelt) != 2 {
		t.Fatalf("gebündelt wurde zu %+v", gebuendelt)
	}

	// **Der erste bleibt vorn** — er stand höher im Rang.
	if gebuendelt[0].Kennung != "o1:0" || len(gebuendelt[0].Weitere) != 1 ||
		gebuendelt[0].Weitere[0] != "o1:1" {
		t.Fatalf("der erste Treffer ist %+v", gebuendelt[0])
	}

	// **Und die Ränge zählen neu durch.**
	for stelle, treffer := range gebuendelt {
		if treffer.Rang != stelle+1 {
			t.Errorf("der %d. trägt Rang %d", stelle+1, treffer.Rang)
		}
	}
}

// **Was keine Fachkennung hat, wird nicht gebündelt.**
//
// **Zwei Regelstellen sind zwei Treffer** — sie tragen keine Kartenidentität,
// und über einen leeren Schlüssel fielen sie alle zusammen.
func TestOhneFachkennungWirdNichtGebuendelt(t *testing.T) {
	t.Parallel()

	gebuendelt := domain.GebuendeltNach([]domain.Recherchetreffer{
		{Rang: 1, Kennung: "regel:205.4a", Quelle: domain.QuelleRegelwerk},
		{Rang: 2, Kennung: "regel:509.1a", Quelle: domain.QuelleRegelwerk},
	})

	if len(gebuendelt) != 2 {
		t.Fatalf("gebündelt wurde zu %+v", gebuendelt)
	}
}

// **Dieselbe Fachkennung in zwei Systemen ist zweierlei.**
//
// **Der Schlüssel ist System **und** Kennung** — `scryfall:abc` und
// `mtgjson:abc` sind zwei verschiedene Karten, und ein Bündel darüber wäre
// eine erfundene Gleichsetzung.
func TestDieselbeKennungInZweiSystemenIstZweierlei(t *testing.T) {
	t.Parallel()

	zweiter := einRechercheteil("m:abc:0", "abc", 2)
	zweiter.Kennungssystem = domain.SystemMTGJSON

	gebuendelt := domain.GebuendeltNach([]domain.Recherchetreffer{
		einRechercheteil("s:abc:0", "abc", 1), zweiter,
	})

	if len(gebuendelt) != 2 {
		t.Fatalf("gebündelt wurde zu %+v", gebuendelt)
	}
}

// **Die Ausgangskarte kommt in ihren eigenen Ähnlichkeitstreffern nicht vor.**
//
// **Auch nicht über ihre zweite Seite**: `o1:0` und `o1:1` sind zwei
// Abschnitte derselben Karte, und wer nur den einen ausschlösse, ließe sie
// über den anderen wieder herein.
func TestDieAusgangskarteKommtNichtInIhrenTreffernVor(t *testing.T) {
	t.Parallel()

	geblieben := domain.OhneAusgangskarte([]domain.Recherchetreffer{
		einRechercheteil("o1:0", "o1", 1),
		einRechercheteil("o1:1", "o1", 2),
		einRechercheteil("o2:0", "o2", 3),
	}, []string{"o1:0"})

	if len(geblieben) != 1 || geblieben[0].Kennung != "o2:0" {
		t.Fatalf("geblieben ist %+v", geblieben)
	}

	if geblieben[0].Rang != 1 {
		t.Errorf("der Rang ist %d", geblieben[0].Rang)
	}
}

// **Ohne Ausgangskarte wird nichts ausgeschlossen.**
func TestOhneAusgangskarteWirdNichtsAusgeschlossen(t *testing.T) {
	t.Parallel()

	treffer := []domain.Recherchetreffer{einRechercheteil("o1:0", "o1", 1)}

	if geblieben := domain.OhneAusgangskarte(treffer, nil); len(geblieben) != 1 {
		t.Fatalf("geblieben ist %+v", geblieben)
	}

	// **Und eine Kennung ohne Doppelpunkt schließt nur sich selbst aus.**
	geblieben := domain.OhneAusgangskarte([]domain.Recherchetreffer{
		einRechercheteil("ohnetrenner", "o1", 1),
		einRechercheteil("o2:0", "o2", 2),
	}, []string{"ohnetrenner"})

	if len(geblieben) != 1 || geblieben[0].Kennung != "o2:0" {
		t.Fatalf("geblieben ist %+v", geblieben)
	}
}
