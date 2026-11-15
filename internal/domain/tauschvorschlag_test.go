package domain_test

// **Was beide hergeben und bekommen** — Werkstück `T4` von Phase 5.6
// (#1469).
//
// # Was hier gemessen wird
//
// **Dass der leere Tausch keiner ist.** Die Formel „möglichst gleicher Wert"
// hat ihr Optimum bei null zu null; ein Vorschlag, der sie allein befolgte,
// empföhle, nichts zu tauschen. Deshalb gilt eine Nichtleer-Bedingung, und
// sie steht vor dem Wert.
//
// **Dass nichts erfunden wird.** Was nicht im Angebot steht, steht auch nicht
// im Vorschlag — gleich wie gut es passte.
//
// **Und dass zwei Währungen nicht addiert werden.** Ein Vorschlag, der EUR und
// USD zusammenzählt, nennt eine Zahl, die es nicht gibt.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einTreffer baut einen Kandidaten mit Menge und Preis.
func einTreffer(bezeichnung string, menge int, preis int64, waehrung string) domain.Tauschkandidat {
	return domain.Tauschkandidat{
		Wunsch: domain.Wunschzeile{
			Druck: bezeichnung, Menge: menge, Rang: domain.RangNormal,
			Bezeichnung: bezeichnung,
		},
		Angebot: domain.Angebotszeile{
			Druck: bezeichnung, Menge: menge, Bezeichnung: bezeichnung,
			Preis: preis, Waehrung: waehrung,
		},
		Menge: menge,
	}
}

func TestEinLeererTauschIstKeinVorschlag(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		gebe, bekomme []domain.Tauschkandidat
	}{
		"nichts auf beiden Seiten": {nil, nil},
		"nur ich gebe":             {[]domain.Tauschkandidat{einTreffer("a", 1, 0, "")}, nil},
		"nur ich bekomme":          {nil, []domain.Tauschkandidat{einTreffer("b", 1, 0, "")}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, gibt := domain.Tauschvorschlagen(fall.gebe, fall.bekomme, domain.Tauschumfang{})
			if gibt {
				t.Error("es kam ein Vorschlag")
			}
		})
	}
}

func TestEinVorschlagNenntBeideSeiten(t *testing.T) {
	t.Parallel()

	vorschlag, gibt := domain.Tauschvorschlagen(
		[]domain.Tauschkandidat{einTreffer("shivan", 2, 0, "")},
		[]domain.Tauschkandidat{einTreffer("bolt", 1, 0, "")},
		domain.Tauschumfang{},
	)

	if !gibt {
		t.Fatal("es kam kein Vorschlag")
	}

	if vorschlag.Gebe.Stueck != 2 || vorschlag.Bekomme.Stueck != 1 {
		t.Errorf("es sind %d hin und %d her", vorschlag.Gebe.Stueck, vorschlag.Bekomme.Stueck)
	}
}

// **Der Deckel gilt je Seite, und er schneidet die unwichtigeren ab.**
func TestDerDeckelBegrenztJedeSeite(t *testing.T) {
	t.Parallel()

	dringend := einTreffer("dringend", 3, 0, "")
	dringend.Wunsch.Rang = domain.RangSehrHoch

	vorschlag, gibt := domain.Tauschvorschlagen(
		[]domain.Tauschkandidat{einTreffer("beilaeufig", 3, 0, ""), dringend},
		[]domain.Tauschkandidat{einTreffer("bolt", 1, 0, "")},
		domain.Tauschumfang{Hoechstzahl: 4},
	)

	if !gibt {
		t.Fatal("es kam kein Vorschlag")
	}

	if vorschlag.Gebe.Stueck != 4 {
		t.Fatalf("es sind %d Stück", vorschlag.Gebe.Stueck)
	}

	// **Der dringende zuerst**: der Deckel schneidet unten ab und nicht oben.
	if vorschlag.Gebe.Zeilen[0].Wunsch.Bezeichnung != "dringend" {
		t.Errorf("vorn steht %q", vorschlag.Gebe.Zeilen[0].Wunsch.Bezeichnung)
	}

	// **Und er erfindet nichts**: die zweite Zeile trägt den Rest.
	if vorschlag.Gebe.Zeilen[1].Menge != 1 {
		t.Errorf("die zweite Zeile hat %d Stück", vorschlag.Gebe.Zeilen[1].Menge)
	}
}

// **Was bepreist ist, wiegt** — und was nicht, wiegt nicht.
func TestEinVorschlagWiegtNurMitVollstaendigenPreisen(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		gebe     []domain.Tauschkandidat
		bepreist bool
		wert     int64
		waehrung string
	}{
		"alles bepreist": {
			[]domain.Tauschkandidat{einTreffer("a", 2, 417, "EUR")}, true, 834, "EUR",
		},
		"eine Zeile ohne Preis": {
			[]domain.Tauschkandidat{einTreffer("a", 1, 417, "EUR"), einTreffer("b", 1, 0, "")},
			false, 0, "",
		},
		// **Ein Preis ohne Währung ist eine Zahl ohne Bedeutung** — derselbe
		// Satz wie an den Erwerbsdaten eines Exemplars.
		"ein Preis ohne Währung": {
			[]domain.Tauschkandidat{einTreffer("a", 1, 417, "")}, false, 0, "",
		},
		// **Zwei Währungen werden nicht addiert.** Die Summe wäre eine Zahl,
		// die es nicht gibt.
		"zwei Währungen": {
			[]domain.Tauschkandidat{einTreffer("a", 1, 417, "EUR"), einTreffer("b", 1, 500, "USD")},
			false, 0, "",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			vorschlag, gibt := domain.Tauschvorschlagen(
				fall.gebe,
				[]domain.Tauschkandidat{einTreffer("gegen", 1, 100, "EUR")},
				domain.Tauschumfang{},
			)

			if !gibt {
				t.Fatal("es kam kein Vorschlag")
			}

			if vorschlag.Gebe.Bepreist != fall.bepreist {
				t.Fatalf("bepreist ist %v", vorschlag.Gebe.Bepreist)
			}

			if vorschlag.Gebe.Wert != fall.wert || vorschlag.Gebe.Waehrung != fall.waehrung {
				t.Errorf("es sind %d %q", vorschlag.Gebe.Wert, vorschlag.Gebe.Waehrung)
			}
		})
	}
}

// **Der Vorschlag erfindet nichts** — er nennt nur, was im Angebot stand.
func TestEinVorschlagErfindetNichts(t *testing.T) {
	t.Parallel()

	vorschlag, _ := domain.Tauschvorschlagen(
		[]domain.Tauschkandidat{einTreffer("shivan", 1, 0, "")},
		[]domain.Tauschkandidat{einTreffer("bolt", 1, 0, "")},
		domain.Tauschumfang{Hoechstzahl: 99},
	)

	for _, zeile := range vorschlag.Gebe.Zeilen {
		if zeile.Angebot.Bezeichnung != "shivan" {
			t.Errorf("im Vorschlag steht %q", zeile.Angebot.Bezeichnung)
		}
	}

	if vorschlag.Gebe.Stueck != 1 {
		t.Errorf("es sind %d Stück, obwohl nur eines dalag", vorschlag.Gebe.Stueck)
	}
}

// **Zweimal dieselbe Eingabe, zweimal derselbe Vorschlag** (§10).
func TestEinVorschlagIstBeimZweitenMalDerselbe(t *testing.T) {
	t.Parallel()

	gebe := []domain.Tauschkandidat{
		einTreffer("b", 1, 0, ""), einTreffer("a", 1, 0, ""), einTreffer("c", 1, 0, ""),
	}

	bekomme := []domain.Tauschkandidat{einTreffer("x", 1, 0, "")}

	erster, _ := domain.Tauschvorschlagen(gebe, bekomme, domain.Tauschumfang{Hoechstzahl: 2})
	zweiter, _ := domain.Tauschvorschlagen(gebe, bekomme, domain.Tauschumfang{Hoechstzahl: 2})

	if len(erster.Gebe.Zeilen) != len(zweiter.Gebe.Zeilen) {
		t.Fatalf("es sind %d und %d Zeilen", len(erster.Gebe.Zeilen), len(zweiter.Gebe.Zeilen))
	}

	for stelle := range erster.Gebe.Zeilen {
		if erster.Gebe.Zeilen[stelle].Angebot.Druck != zweiter.Gebe.Zeilen[stelle].Angebot.Druck {
			t.Errorf("an Stelle %d stehen %q und %q", stelle,
				erster.Gebe.Zeilen[stelle].Angebot.Druck, zweiter.Gebe.Zeilen[stelle].Angebot.Druck)
		}
	}
}

// **Die offenen Stücke stehen dran** — ein Vorschlag sagt auch, was er nicht
// deckt.
func TestEinVorschlagNenntSeineLuecke(t *testing.T) {
	t.Parallel()

	halb := einTreffer("shivan", 1, 0, "")
	halb.Wunsch.Menge, halb.Fehlend = 4, 3

	vorschlag, _ := domain.Tauschvorschlagen(
		[]domain.Tauschkandidat{halb},
		[]domain.Tauschkandidat{einTreffer("bolt", 1, 0, "")},
		domain.Tauschumfang{},
	)

	if vorschlag.Gebe.Gedeckt != 1 || vorschlag.Gebe.Offen != 3 {
		t.Errorf("gedeckt sind %d, offen %d", vorschlag.Gebe.Gedeckt, vorschlag.Gebe.Offen)
	}
}
