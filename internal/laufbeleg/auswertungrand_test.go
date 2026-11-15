package laufbeleg_test

// **Die Wege, die kein echter Probelauf hergibt** — abgetrennt von
// `auswertung_test.go`, als die Datei über die Zeilengrenze aus §2 ging.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/laufbeleg"
)

// **Was kein echter Probelauf hergibt** — gebaut aus den Feldern, die die
// echten Ströme oben zeigen.
func TestDieSeltenenWegeHabenEinenGrund(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		ereignisse []laufbeleg.Ereignis
		rueckgabe  int
		status     laufbeleg.Status
		grund      string
	}{
		// **Scheitert eine Abhängigkeit**, nennt `FailedBuild` sie und nicht
		// das Paket, das getestet werden sollte.
		"eine Abhängigkeit übersetzt nicht": {
			[]laufbeleg.Ereignis{
				{Action: "build-output", ImportPath: "q/tief", Output: "# q/tief\n"},
				{Action: "build-output", ImportPath: "q/tief", Output: "tief.go:3:1: syntax error\n"},
				{Action: "build-fail", ImportPath: "q/tief"},
				{Action: "start", Package: "q/oben"},
				{Action: "fail", Package: "q/oben", FailedBuild: "q/tief"},
			},
			1,
			laufbeleg.Gescheitert, "Übersetzung gescheitert: tief.go:3:1: syntax error",
		},
		"ein Übersetzungsfehler ohne Meldung": {
			[]laufbeleg.Ereignis{
				{Action: "start", Package: "q/oben"},
				{Action: "fail", Package: "q/oben", FailedBuild: "q/oben [q/oben.test]"},
			},
			1,
			laufbeleg.Gescheitert, "Übersetzung gescheitert: ohne Meldung",
		},
		// **Von außen beendet**: kein Ende und keine Zeile, die es sagt.
		"ein Fall ohne Ende": {
			[]laufbeleg.Ereignis{
				{Action: "start", Package: "q/oben"},
				{Action: "run", Package: "q/oben", Test: "TestA"},
			},
			-1,
			laufbeleg.Gescheitert, "kein Ende gemeldet",
		},
		"ein Sprung ohne Meldung": {
			[]laufbeleg.Ereignis{
				{Action: "run", Package: "q/oben", Test: "TestA"},
				{Action: "output", Package: "q/oben", Test: "TestA", Output: "--- SKIP: TestA\n", OutputType: "frame"},
				{Action: "skip", Package: "q/oben", Test: "TestA"},
			},
			0,
			laufbeleg.Uebersprungen, "ohne Grund",
		},
		// **Der Grund eines Sprungs steht zuletzt** — nach allem, was davor
		// geloggt wurde.
		"ein Sprung nach einer Logzeile": {
			[]laufbeleg.Ereignis{
				{Action: "run", Package: "q/oben", Test: "TestA"},
				{Action: "output", Package: "q/oben", Test: "TestA", Output: "    a_test.go:4: zehn Sets gefunden\n"},
				{Action: "output", Package: "q/oben", Test: "TestA", Output: "    a_test.go:9: die Referenzen fehlen\n"},
				{Action: "skip", Package: "q/oben", Test: "TestA"},
			},
			0,
			laufbeleg.Uebersprungen, "die Referenzen fehlen",
		},
		// **Rot ohne Fehlerzeile**: `t.Fail()` sagt nichts, und die letzte
		// Ausgabe ist das Beste, was es gibt.
		"rot ohne Fehlerzeile": {
			[]laufbeleg.Ereignis{
				{Action: "run", Package: "q/oben", Test: "TestA"},
				{Action: "output", Package: "q/oben", Test: "TestA", Output: "    a_test.go:4: 679 Befunde\n"},
				{Action: "fail", Package: "q/oben", Test: "TestA"},
			},
			1,
			laufbeleg.Gescheitert, "679 Befunde",
		},
		"rot ganz ohne Ausgabe": {
			[]laufbeleg.Ereignis{
				{Action: "run", Package: "q/oben", Test: "TestA"},
				{Action: "fail", Package: "q/oben", Test: "TestA"},
			},
			1,
			laufbeleg.Gescheitert, "ohne Meldung",
		},
	}

	for name, einer := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			auswertung := laufbeleg.Auswerten(einer.ereignisse, laufbeleg.Erwartung{
				Faelle: []laufbeleg.Fall{{Paket: "q/oben", Test: "TestA"}},
			}, einer.rueckgabe)

			if auswertung.Status != einer.status || !strings.Contains(auswertung.Grund, einer.grund) {
				t.Errorf("Status %s, Grund %q — erwartet %s und %q",
					auswertung.Status, auswertung.Grund, einer.status, einer.grund)
			}
		})
	}
}

// **Ein unerwarteter Sprung wird gezählt** — und macht den Lauf nicht rot.
func TestEinUnerwarteterSprungWirdGezaehlt(t *testing.T) {
	t.Parallel()

	auswertung := laufbeleg.Auswerten(nurPakete(strom(t, "strom.jsonl"), "gut", "sprung"),
		laufbeleg.Erwartung{Faelle: []laufbeleg.Fall{fall("gut", "TestGut")}}, 0)

	if auswertung.Weitere.Uebersprungen != 1 || auswertung.Status != laufbeleg.Bestanden {
		t.Errorf("weitere %+v, Status %s", auswertung.Weitere, auswertung.Status)
	}
}

// **Ein Paket, das rot endet, ohne dass ein Fall darin rot war**, ist ein
// geschluckter Fehler (AK-03) — auch dann, wenn ein roter Trend in einem
// anderen Paket den Rückgabewert schon erklärt. Etwa ein Fehler nach den
// Läufen, den `go test` mit `FAIL` und ohne roten Fall meldet.
func TestEinPaketDasOhneGrundRotEndetIstGescheitert(t *testing.T) {
	t.Parallel()

	trend := laufbeleg.Fall{Paket: "q/trend", Test: "TestAbstand"}
	ereignisse := []laufbeleg.Ereignis{
		{Action: "run", Package: "q/oben", Test: "TestA"},
		{Action: "pass", Package: "q/oben", Test: "TestA"},
		{Action: "output", Package: "q/oben", Output: "FAIL\tq/oben\t0.1s\n"},
		{Action: "fail", Package: "q/oben"},
		{Action: "run", Package: "q/trend", Test: "TestAbstand"},
		{Action: "fail", Package: "q/trend", Test: "TestAbstand"},
		{Action: "fail", Package: "q/trend"},
	}

	auswertung := laufbeleg.Auswerten(ereignisse, laufbeleg.Erwartung{
		Faelle: []laufbeleg.Fall{{Paket: "q/oben", Test: "TestA"}, trend},
		Trend:  map[laufbeleg.Fall]string{trend: "als Verlauf zu lesen"},
	}, 1)

	if auswertung.Status != laufbeleg.Gescheitert ||
		!strings.Contains(auswertung.Grund, "q/oben: das Paket endete rot") {
		t.Errorf("Status %s, Grund %q — erwartet FAIL über q/oben", auswertung.Status, auswertung.Grund)
	}

	// **Die Gegenprobe**: dasselbe Paket, grün beendet, ist kein Befund —
	// sonst hielte die Zeile oben jedes rote Ende für geschluckt.
	ereignisse[3].Action = "pass"
	if gruen := laufbeleg.Auswerten(ereignisse, laufbeleg.Erwartung{
		Faelle: []laufbeleg.Fall{{Paket: "q/oben", Test: "TestA"}, trend},
		Trend:  map[laufbeleg.Fall]string{trend: "als Verlauf zu lesen"},
	}, 1); gruen.Status != laufbeleg.Bestanden {
		t.Errorf("Status %s, Grund %q — erwartet PASS", gruen.Status, gruen.Grund)
	}
}
