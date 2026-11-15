package domain_test

// **Aus Werten werden Gruppen** — Werkstück C von Phase 4.11
// (#1338).
//
// # Was am 2026-08-31 nachgelesen wurde (§15)
//
// `GraphData::GraphData` (`src/gui/control/graph.cpp`, Zeilen 85–135),
// `GraphAxis::addGroup` (Zeilen 24–32), `GraphDataPre::splitList`
// (Zeilen 42–60) und `to_bin` (Zeilen 67–73), Fassung f9d9356 vom 2021-07-06.
//
// Drei Ordnungen, und MSE wählt in dieser Reihenfolge:
//
//  1. **numerisch** — die Werte werden als Zahl gelesen, fehlende ganze
//     Zahlen dazwischen bekommen einen leeren Balken;
//  2. **erklärt** — `group:` gibt die Reihenfolge, auch für Gruppen mit null;
//  3. **sonst** — `smart_less`, also [domain.Vergleiche].

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die erklärte Reihenfolge gilt — samt der Gruppe, die niemand trifft.**
//
// Das ist MSEs zweiter Zweig: `FOR_EACH_CONST(gn, *a->order) a->addGroup(gn,
// counts[gn])`. `counts[gn]` legt für einen unbekannten Namen eine Null an —
// die leere Gruppe steht also da, und das ist die Absicht: eine Farbe ohne
// Karten ist eine Aussage.
func TestDieErklaerteReihenfolgeGiltSamtLeererGruppe(t *testing.T) {
	t.Parallel()

	gruppen := domain.Auszaehlen(
		[]string{"Weiss", "Rot", "Weiss"},
		domain.Auszaehlart{Gruppen: []string{"Land", "Weiss", "Rot"}},
	)

	if len(gruppen) != 3 {
		t.Fatalf("%d Gruppen: %+v", len(gruppen), gruppen)
	}

	if gruppen[0].Name != "Land" || gruppen[0].Karten != 0 {
		t.Errorf("die erste Gruppe ist %+v", gruppen[0])
	}

	if gruppen[1].Name != "Weiss" || gruppen[1].Karten != 2 {
		t.Errorf("die zweite Gruppe ist %+v", gruppen[1])
	}
}

// **Ohne erklärte Reihenfolge wird klug geordnet** — `smart_less`, also
// `Karte 2` vor `Karte 10`.
func TestOhneErklaerteReihenfolgeWirdKlugGeordnet(t *testing.T) {
	t.Parallel()

	gruppen := domain.Auszaehlen(
		[]string{"Karte 10", "Karte 2", "Karte 10"},
		domain.Auszaehlart{},
	)

	if len(gruppen) != 2 {
		t.Fatalf("%d Gruppen: %+v", len(gruppen), gruppen)
	}

	if gruppen[0].Name != "Karte 2" || gruppen[1].Karten != 2 {
		t.Errorf("die Ordnung lautet %+v", gruppen)
	}
}

// **Numerisch heißt: als Zahl, und die Lücke dazwischen steht da.**
//
// MSE füllt sie ausdrücklich (`add 0 bars before this value`). Ein Balken, der
// fehlt, sähe aus wie einer, der nie erklärt wurde.
func TestEineNumerischeAchseFuelltIhreLuecken(t *testing.T) {
	t.Parallel()

	gruppen := domain.Auszaehlen(
		[]string{"1", "3", "3"},
		domain.Auszaehlart{Numerisch: true},
	)

	// **Und die Achse beginnt bei null**, auch wenn keine Karte dort steht.
	// MSEs Zähler `prev` startet auf 0, und die Schleife füllt von dort bis
	// zum ersten Wert. Der erste Anlauf dieses Tests erwartete drei Gruppen
	// und bekam vier — nachgelesen statt geraten hätte es die 0 gleich
	// genannt (§15).
	if len(gruppen) != 4 {
		t.Fatalf("%d Gruppen: %+v", len(gruppen), gruppen)
	}

	erwartet := []domain.Auszaehlgruppe{
		{Name: "0", Karten: 0},
		{Name: "1", Karten: 1},
		{Name: "2", Karten: 0},
		{Name: "3", Karten: 2},
	}

	for i, soll := range erwartet {
		if gruppen[i].Name != soll.Name || gruppen[i].Karten != soll.Karten {
			t.Errorf("Gruppe %d ist %+v, erwartet %+v", i, gruppen[i], soll)
		}
	}
}

// **Eine Klassenbreite fasst zusammen — mit MSEs Beschriftung.**
//
//	int bin = ceil(value / bin_size);
//	String::Format("%.0f%c%.0f", (bin-1) * bin_size + 1, EN_DASH, bin * bin_size)
//
// Also `1–5` für alles von 1 bis 5, und der Strich ist ein Halbgeviertstrich.
func TestEineKlassenbreiteFasstZusammen(t *testing.T) {
	t.Parallel()

	gruppen := domain.Auszaehlen(
		[]string{"1", "4", "7"},
		domain.Auszaehlart{Numerisch: true, Klassenbreite: 5},
	)

	// **Drei und nicht zwei**: die 0 steht vorn, siehe den Fall darüber.
	if len(gruppen) != 3 {
		t.Fatalf("%d Gruppen: %+v", len(gruppen), gruppen)
	}

	if gruppen[0].Name != "0" || gruppen[0].Karten != 0 {
		t.Errorf("die Achse beginnt mit %+v", gruppen[0])
	}

	// **`1` und `4` liegen in derselben Klasse** — `ceil(1/5) == ceil(4/5)`.
	if gruppen[1].Name != "1–5" || gruppen[1].Karten != 2 {
		t.Errorf("die erste Klasse ist %+v", gruppen[1])
	}

	if gruppen[2].Name != "6–10" || gruppen[2].Karten != 1 {
		t.Errorf("die zweite Klasse ist %+v", gruppen[2])
	}
}

// **Die Null bleibt die Null.** `to_bin` gibt für `value == 0` die Zahl selbst
// zurück, auch mit Klassenbreite — sonst hieße „keine Manakosten" `1–5`.
func TestDieNullBleibtDieNull(t *testing.T) {
	t.Parallel()

	gruppen := domain.Auszaehlen(
		[]string{"0", "0"},
		domain.Auszaehlart{Numerisch: true, Klassenbreite: 5},
	)

	if len(gruppen) != 1 || gruppen[0].Name != "0" || gruppen[0].Karten != 2 {
		t.Errorf("die Gruppen lauten %+v", gruppen)
	}
}

// **Ein Wert, der keine Zahl ist, zählt trotzdem** — MSEs `non-numeric, add
// anyway`. Eine numerische Achse mit einem Text darin verliert ihn nicht.
func TestEineNumerischeAchseVerliertKeinenText(t *testing.T) {
	t.Parallel()

	gruppen := domain.Auszaehlen(
		[]string{"2", "X"},
		domain.Auszaehlart{Numerisch: true},
	)

	gefunden := false

	for _, gruppe := range gruppen {
		if gruppe.Name == "X" && gruppe.Karten == 1 {
			gefunden = true
		}
	}

	if !gefunden {
		t.Errorf("der Textwert X fehlt: %+v", gruppen)
	}
}

// **Die Farbe kommt aus der Tafel, wo eine dasteht.**
func TestEineGruppeTraegtIhreErklaerteFarbe(t *testing.T) {
	t.Parallel()

	gruppen := domain.Auszaehlen([]string{"Weiss"}, domain.Auszaehlart{
		Gruppen: []string{"Weiss"},
		Farben:  map[string]string{"Weiss": "rgb(255,231,185)"},
	})

	if len(gruppen) != 1 || gruppen[0].Farbe != "rgb(255,231,185)" {
		t.Errorf("die Gruppen lauten %+v", gruppen)
	}
}

// **Ohne Werte kommt eine leere Liste und kein `nil`.**
func TestOhneWerteKommtEineLeereListe(t *testing.T) {
	t.Parallel()

	if gruppen := domain.Auszaehlen(nil, domain.Auszaehlart{}); gruppen == nil {
		t.Error("die Liste ist nil")
	}
}
