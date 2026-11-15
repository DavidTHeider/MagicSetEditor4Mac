package laufbeleg_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/laufbeleg"
)

// fall baut einen Fall im Probemodul.
func fall(paket, test string) laufbeleg.Fall {
	return laufbeleg.Fall{Paket: "probe.test/" + paket, Test: test}
}

// nurPakete behält die Ereignisse der genannten Pakete.
func nurPakete(ereignisse []laufbeleg.Ereignis, pakete ...string) []laufbeleg.Ereignis {
	behalten := []laufbeleg.Ereignis{}

	for _, ereignis := range ereignisse {
		for _, paket := range pakete {
			if ereignis.Package == "probe.test/"+paket ||
				strings.HasPrefix(ereignis.ImportPath, "probe.test/"+paket+" ") {
				behalten = append(behalten, ereignis)
			}
		}
	}

	return behalten
}

// ohneTest nimmt die Ereignisse eines Tests heraus.
func ohneTest(ereignisse []laufbeleg.Ereignis, test string) []laufbeleg.Ereignis {
	behalten := []laufbeleg.Ereignis{}

	for _, ereignis := range ereignisse {
		if ereignis.Test != test {
			behalten = append(behalten, ereignis)
		}
	}

	return behalten
}

// ergebnisVon sucht einen Fall in der Auswertung.
func ergebnisVon(t *testing.T, auswertung laufbeleg.Auswertung,
	gesucht laufbeleg.Fall,
) laufbeleg.Fallergebnis {
	t.Helper()

	for _, ergebnis := range auswertung.Faelle {
		if ergebnis.Fall == gesucht {
			return ergebnis
		}
	}

	t.Fatalf("%v steht nicht in der Auswertung: %+v", gesucht, auswertung.Faelle)

	return laufbeleg.Fallergebnis{}
}

// **Jeder Fall trägt Ausführung und Ergebnis getrennt** (AK-03) — und einen
// Grund, sobald etwas nicht bestanden hat.
func TestJederFallSagtAusfuehrungUndErgebnis(t *testing.T) {
	t.Parallel()

	gesamt := append(strom(t, "strom.jsonl"), strom(t, "zeit.jsonl")...)

	faelle := map[string]struct {
		fall        laufbeleg.Fall
		ausfuehrung laufbeleg.Status
		ergebnis    laufbeleg.Status
		grund       string
	}{
		"bestanden":                   {fall("gut", "TestGut"), laufbeleg.Bestanden, laufbeleg.Bestanden, ""},
		"bestanden mit Unterfall":     {fall("gut", "TestMitUnterfall"), laufbeleg.Bestanden, laufbeleg.Bestanden, ""},
		"rot, aber gemessen":          {fall("rot", "TestRot"), laufbeleg.Bestanden, laufbeleg.Gescheitert, "erwartet 0, bekam 679"},
		"übersprungen":                {fall("sprung", "TestSprung"), laufbeleg.Uebersprungen, laufbeleg.Uebersprungen, "die kanonischen Sets liegen nicht vor"},
		"abgestürzt":                  {fall("panik", "TestPanik"), laufbeleg.Gescheitert, laufbeleg.Gescheitert, "abgebrochen: panic: assignment to entry in nil map"},
		"nach dem Absturz":            {fall("panik", "TestDanach"), laufbeleg.NichtGelaufen, laufbeleg.NichtGelaufen, "brach vorher ab"},
		"nicht übersetzt":             {fall("kaputt", "TestKaputt"), laufbeleg.Gescheitert, laufbeleg.Gescheitert, "Übersetzung gescheitert: kaputt/kaputt_test.go:5:33: undefined: gibtsNicht"},
		"über der Zeit":               {fall("zeit", "TestZuLang"), laufbeleg.Gescheitert, laufbeleg.Gescheitert, "abgebrochen: panic: test timed out after 1s"},
		"ein Paket, das nicht lief":   {fall("fehlt", "TestNie"), laufbeleg.NichtGelaufen, laufbeleg.NichtGelaufen, "das Paket lief nicht"},
		"ein Test, den es nicht gibt": {fall("gut", "TestUmbenannt"), laufbeleg.NichtGelaufen, laufbeleg.NichtGelaufen, "nicht gelaufen"},
	}

	erwartung := laufbeleg.Erwartung{}
	for _, einer := range faelle {
		erwartung.Faelle = append(erwartung.Faelle, einer.fall)
	}

	auswertung := laufbeleg.Auswerten(gesamt, erwartung, 1)

	for name, einer := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ergebnis := ergebnisVon(t, auswertung, einer.fall)

			if ergebnis.Ausfuehrung != einer.ausfuehrung || ergebnis.Ergebnis != einer.ergebnis {
				t.Errorf("Ausführung %s, Ergebnis %s — erwartet %s und %s",
					ergebnis.Ausfuehrung, ergebnis.Ergebnis, einer.ausfuehrung, einer.ergebnis)
			}

			if !strings.Contains(ergebnis.Grund, einer.grund) ||
				(einer.grund == "") != (ergebnis.Grund == "") {
				t.Errorf("Grund %q, erwartet %q", ergebnis.Grund, einer.grund)
			}
		})
	}
}

// **Was der ganze Lauf ist** — der schwerste Stand bestimmt ihn (AK-02).
func TestDerLaufHatDenSchwerstenStand(t *testing.T) {
	t.Parallel()

	gesamt := strom(t, "strom.jsonl")
	gut := []laufbeleg.Fall{fall("gut", "TestGut"), fall("gut", "TestMitUnterfall")}
	trend := fall("rot", "TestTrend")
	sprung := fall("sprung", "TestSprung")

	faelle := map[string]struct {
		ereignisse []laufbeleg.Ereignis
		erwartung  laufbeleg.Erwartung
		rueckgabe  int
		status     laufbeleg.Status
		grund      string
	}{
		"alles bestanden": {
			nurPakete(gesamt, "gut"),
			laufbeleg.Erwartung{Faelle: gut},
			0,
			laufbeleg.Bestanden, "",
		},
		"ein Trend darf rot sein, wenn er gemessen hat": {
			ohneTest(nurPakete(gesamt, "gut", "rot"), "TestRot"),
			laufbeleg.Erwartung{
				Faelle: append([]laufbeleg.Fall{trend}, gut...),
				Trend:  map[laufbeleg.Fall]string{trend: "als Verlauf zu lesen"},
			},
			1,
			laufbeleg.Bestanden, "",
		},
		"derselbe Fall als Pflicht ist rot": {
			nurPakete(gesamt, "rot"),
			laufbeleg.Erwartung{Faelle: []laufbeleg.Fall{trend}},
			1,
			laufbeleg.Gescheitert, "probe.test/rot.TestTrend: FAIL — Abstand über der Toleranz",
		},
		"ein Sprung ist kein Bestehen": {
			nurPakete(gesamt, "gut", "sprung"),
			laufbeleg.Erwartung{Faelle: append([]laufbeleg.Fall{sprung}, gut...)},
			0,
			laufbeleg.Uebersprungen, "die kanonischen Sets liegen nicht vor",
		},
		"ein erlaubter Sprung schon": {
			nurPakete(gesamt, "sprung"),
			laufbeleg.Erwartung{
				Faelle:   []laufbeleg.Fall{sprung},
				Optional: map[laufbeleg.Fall]string{sprung: "Vorrichtung"},
			},
			0,
			laufbeleg.Bestanden, "",
		},
		"ein Trend, der übersprungen wird, hat nicht gemessen": {
			nurPakete(gesamt, "sprung"),
			laufbeleg.Erwartung{
				Faelle: []laufbeleg.Fall{sprung},
				Trend:  map[laufbeleg.Fall]string{sprung: "Verlauf"},
			},
			0,
			laufbeleg.Uebersprungen, "",
		},
		"ein erlaubter Sprung, der rot wird, bleibt rot": {
			nurPakete(gesamt, "rot"),
			laufbeleg.Erwartung{
				Faelle:   []laufbeleg.Fall{fall("rot", "TestRot")},
				Optional: map[laufbeleg.Fall]string{fall("rot", "TestRot"): "Vorrichtung"},
			},
			1,
			laufbeleg.Gescheitert, "",
		},
		"null Messfälle": {
			nurPakete(gesamt, "gut"),
			laufbeleg.Erwartung{},
			0,
			laufbeleg.NichtGelaufen, "keine Messfälle erwartet",
		},
		"ein Rückgabewert, den kein Fall erklärt": {
			nurPakete(gesamt, "gut"),
			laufbeleg.Erwartung{Faelle: gut},
			1,
			laufbeleg.Gescheitert, "der Prozess endete mit 1, und kein Fall erklärt es",
		},
		"ein abgebrochener Prozess": {
			nurPakete(gesamt, "gut"),
			laufbeleg.Erwartung{Faelle: gut},
			-1,
			laufbeleg.Gescheitert, "der Prozess wurde abgebrochen",
		},
		"ein weiterer Fall, der rot ist": {
			nurPakete(gesamt, "gut", "rot"),
			laufbeleg.Erwartung{Faelle: gut},
			1,
			laufbeleg.Gescheitert, "probe.test/rot.TestRot",
		},
		"ein Übersetzungsfehler in einem Paket, das keiner erwartet": {
			nurPakete(gesamt, "gut", "kaputt"),
			laufbeleg.Erwartung{Faelle: gut},
			1,
			laufbeleg.Gescheitert, "probe.test/kaputt: Übersetzung gescheitert",
		},
		"aus dem Zwischenspeicher": {
			strom(t, "zwischenspeicher.jsonl"),
			laufbeleg.Erwartung{Faelle: gut},
			0,
			laufbeleg.NichtGelaufen, "aus dem Zwischenspeicher",
		},
	}

	for name, einer := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			auswertung := laufbeleg.Auswerten(einer.ereignisse, einer.erwartung, einer.rueckgabe)

			if auswertung.Status != einer.status {
				t.Errorf("Status %s, erwartet %s — Grund %q", auswertung.Status,
					einer.status, auswertung.Grund)
			}

			if !strings.Contains(auswertung.Grund, einer.grund) ||
				(einer.status == laufbeleg.Bestanden) != (auswertung.Grund == "") {
				t.Errorf("Grund %q, erwartet %q", auswertung.Grund, einer.grund)
			}
		})
	}
}

// **Ein roter Trendlauf liefert trotzdem seine Kennzahlen** (AK-03).
func TestEinRoterTrendLiefertKennzahlen(t *testing.T) {
	t.Parallel()

	trend := fall("rot", "TestTrend")
	auswertung := laufbeleg.Auswerten(nurPakete(strom(t, "strom.jsonl"), "gut", "rot"),
		laufbeleg.Erwartung{
			Faelle: []laufbeleg.Fall{trend, fall("gut", "TestGut")},
			Trend:  map[laufbeleg.Fall]string{trend: "Verlauf"},
		}, 1)

	if auswertung.Kennzahlen["trend.mittel"] != "20.0560" ||
		auswertung.Kennzahlen["gut.zahl"] != "7" {
		t.Errorf("Kennzahlen %v", auswertung.Kennzahlen)
	}

	if ergebnis := ergebnisVon(t, auswertung, trend); ergebnis.Art != laufbeleg.Trend ||
		ergebnis.Ergebnis != laufbeleg.Gescheitert {
		t.Errorf("der Trend steht als %+v", ergebnis)
	}

	// **Die weiteren Fälle werden gezählt**, auch wenn keiner rot war.
	if auswertung.Weitere.Gelaufen != 2 || auswertung.Weitere.Gescheitert != 1 {
		t.Errorf("weitere %+v, erwartet zwei gelaufene, einer rot", auswertung.Weitere)
	}
}

// **Zwei Werte für dieselbe Kennzahl sind ein Befund**, kein stilles Überschreiben.
func TestEineKennzahlMitZweiWertenIstEinBefund(t *testing.T) {
	t.Parallel()

	ereignisse := []laufbeleg.Ereignis{
		{Action: "run", Package: "p", Test: "TestA"},
		{Action: "output", Package: "p", Test: "TestA", Output: "    a_test.go:3: kennzahl x=1\n"},
		{Action: "output", Package: "p", Test: "TestA/unter", Output: "    a_test.go:4: kennzahl x=1\n"},
		{Action: "output", Package: "p", Test: "TestA", Output: "    a_test.go:5: kennzahl x=2\n"},
		{Action: "pass", Package: "p", Test: "TestA"},
		{Action: "pass", Package: "p"},
	}

	auswertung := laufbeleg.Auswerten(ereignisse,
		laufbeleg.Erwartung{Faelle: []laufbeleg.Fall{{Paket: "p", Test: "TestA"}}}, 0)

	if len(auswertung.Fehler) != 1 || !strings.Contains(auswertung.Fehler[0], "kennzahl x") {
		t.Errorf("Fehler %q, erwartet einen über x", auswertung.Fehler)
	}

	if auswertung.Status != laufbeleg.Gescheitert {
		t.Errorf("Status %s — eine widersprüchliche Kennzahl macht den Beleg ungültig",
			auswertung.Status)
	}
}
