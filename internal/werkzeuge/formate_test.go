// Der Wächter über `docs/FORMATE.md`.
//
// **Bis zum 2026-08-16 war sie der Beleg der Unabhängigkeit**: dieses Projekt
// stand unter MIT, der MSE2-Quelltext war tabu, und eine Feststellung ohne
// Quelle war von einer abgeschriebenen nicht zu unterscheiden.
//
// Seit dem Wechsel auf GPL-2.0
// ([0019](../../docs/entscheidungen/0019-gpl-2-0-statt-mit.md)) darf portiert
// werden, und die Quellenangabe hat einen anderen Zweck: sie **unterscheidet**,
// ob eine Feststellung aus einem Datenpaket erschlossen oder aus MSE2
// übernommen ist (`AGENTS.md` §11). Beides ist erlaubt; wer später etwas
// nachprüft, muss wissen, welches von beidem vor ihm liegt. Gefunden vom
// Zweitblick über die Portierung, der den überholten Zweck angestrichen hat.
//
// Eine Prosa-Regel darüber hält nichts. Also zählt hier jemand nach:
//
//  1. Jede Feststellung nennt Quelle, Datum, Sicherheitsgrad und Testfall.
//  2. Was nicht „belegt" ist, benennt einen Testfall — sonst wäre die Annahme
//     dahinter unsichtbar, und ein späterer Fund würde sie erben statt sie rot
//     zu machen.
//  3. Ein genannter Testfall existiert wirklich im Baum. Ein Verweis auf einen
//     Test, den es nicht gibt, ist schlimmer als keiner: er sieht aus wie ein
//     Beleg.
//
// # Was er nicht prüft
//
// **Zwei Dinge, und beide sind Lücken, keine Auslassungen.**
//
// Erstens: ob die **Quelle stimmt**. Dass `magic-m15.mse-style/style` in Zeile
// 12 das sagt, was danebensteht, kann er nicht nachsehen — das Paket liegt nicht
// im Repository und darf es nicht (`docs/entscheidungen/0005`).
//
// Zweitens, und das wiegt schwerer: ob der genannte Testfall die Feststellung
// **deckt**. Er prüft, dass eine Funktion dieses Namens existiert, nicht was sie
// tut. Der Abschnitt „Der Paketkopf" trifft sechs Aussagen — Kleinschreibung,
// Trennzeichen, Einrückung, Wiederholung, Versionsform, BOM — und nennt einen
// Test, der eine davon deckt. Der Zweitblick über Phase 2 hat das benannt, und
// es ist die halbe Zusage aus §10: was der Wächter nicht prüft, gehört
// danebengeschrieben, und hier steht es.
//
// Maschinell zu prüfen wäre es nur über eine Zuordnung Aussage → Test, also
// über eine feinere Zerlegung der Datei. Das ist möglich und nicht getan.
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// feststellung fängt einen Abschnitt samt seinen vier Angaben.
//
// `angabe` fängt **den ganzen Wert**, nicht nur seine erste Zeile. Der erste
// Anlauf endete auf `(.*)$`, und eine Begründung, die über zwei Zeilen läuft,
// war ab der zweiten unsichtbar — worauf der Wächter über die doppelten
// Sicherheitsgrade grün lief, obwohl drei Abschnitte ihn hätten rot machen
// müssen. Ein Wächter, der die Hälfte seines Eingabewerts nicht liest, prüft
// nicht die Datei, sondern ihren Anfang.
// Zeilenweise statt über einen Ausdruck: Gos Regex kennt keinen Lookahead, und
// ein Abschluss auf `\n\*\*` verschlänge den Anfang des nächsten Feldes — dann
// fände der Wächter nur noch jede zweite Angabe.
var (
	ueberschrift = regexp.MustCompile(`(?m)^## (.+)$`)
	angabe       = regexp.MustCompile(`^\*\*(Quelle|Erschlossen am|Sicherheit|Testfall):\*\* *(.*)$`)
)

// angabenAus liest die vier Angaben eines Abschnitts, jede über alle ihre Zeilen.
func angabenAus(abschnitt string) map[string]string {
	gefunden := map[string]string{}
	offen := ""

	for _, zeile := range strings.Split(abschnitt, "\n") {
		if treffer := angabe.FindStringSubmatch(zeile); treffer != nil {
			offen = treffer[1]
			gefunden[offen] = treffer[2]

			continue
		}

		// Eine Fortsetzungszeile gehört zur zuletzt begonnenen Angabe. Beendet
		// wird sie von einer Leerzeile, einem Trenner oder dem nächsten Feld.
		if offen == "" || strings.TrimSpace(zeile) == "" || strings.HasPrefix(zeile, "---") {
			offen = ""

			continue
		}

		gefunden[offen] += " " + strings.TrimSpace(zeile)
	}

	return gefunden
}

// keineFeststellung sind die Abschnitte, die keine sind.
var keineFeststellung = map[string]bool{
	"Wie eine Feststellung hier aussieht": true,
	"Noch offen":                          true,
}

// grade sind die zulässigen Sicherheitsgrade.
var grade = []string{"belegt", "plausibel", "Vermutung"}

// TestKeineFeststellungBleibtBeiEinemPaket verlangt, was F05
// (#2247)
// verlangt: **jede Feststellung an mindestens zwei Paketen** — oder am
// Quelltext von MSE2 — **oder ausdrücklich als Vermutung**.
//
// „plausibel" hieß „an einem Paket gesehen, nicht gegengeprüft", und es war
// der bequeme Grad dazwischen: er sah aus wie ein Beleg und verlangte keinen
// Testfall, der die Annahme benennt. Am 2026-10-10 standen zwölf
// Feststellungen so da; jede ist seither nachgeprüft — acht belegt, vier als
// Vermutung. **Der Grad bleibt in [grade]**, damit dieser Wächter ihn beim
// Namen nennen kann, statt „kein Grad" zu sagen.
func TestKeineFeststellungBleibtBeiEinemPaket(t *testing.T) {
	for _, fund := range feststellungen(t) {
		if ersterGrad(fund.angabe["Sicherheit"]) == "plausibel" {
			t.Errorf("%q steht als „plausibel“ da — an einem Paket gesehen. Ein zweites "+
				"Paket oder der Quelltext von MSE2 macht sie belegt; sonst ist sie eine Vermutung (F05)",
				fund.titel)
		}
	}
}

type feststellung struct {
	titel  string
	angabe map[string]string
}

func TestJedeFeststellungTraegtIhreVierAngaben(t *testing.T) {
	for _, fund := range feststellungen(t) {
		t.Run(fund.titel, func(t *testing.T) {
			for _, pflicht := range []string{"Quelle", "Erschlossen am", "Sicherheit", "Testfall"} {
				if strings.TrimSpace(fund.angabe[pflicht]) == "" {
					t.Errorf("die Feststellung %q nennt kein %q", fund.titel, pflicht)
				}
			}

			grad := ersterGrad(fund.angabe["Sicherheit"])
			if grad == "" {
				t.Errorf("der Sicherheitsgrad %q ist keiner von %v",
					fund.angabe["Sicherheit"], grade)
			}
		})
	}
}

// Was nicht „belegt" ist und in Code steht, benennt einen Testfall.
//
// Ein Gedankenstrich ist nur dort erlaubt, wo die Feststellung belegt ist — oder
// wo sie ausdrücklich `— (noch nicht in Code)` sagt.
//
// **Die Ausnahme steht in der Angabe, nicht in der Prosa daneben.** Der erste
// Anlauf suchte sie im Fließtext des Sicherheitsgrades und fand sie nicht, weil
// sie auf einer Folgezeile stand. Eine Ausnahme, die der Wächter nicht lesen
// kann, ist keine — sie ist ein roter Lauf mit Erklärung.
func TestJedeUnbelegteFeststellungBenenntEinenTestfall(t *testing.T) {
	for _, fund := range feststellungen(t) {
		grad := ersterGrad(fund.angabe["Sicherheit"])
		testfall := strings.TrimSpace(fund.angabe["Testfall"])

		t.Run(fund.titel, func(t *testing.T) {
			if grad == "belegt" || strings.Contains(testfall, "noch nicht in Code") {
				return
			}

			if testfall == "—" || testfall == "-" {
				t.Errorf("%q ist %q und benennt trotzdem keinen Testfall — "+
					"die Annahme dahinter ist damit unsichtbar", fund.titel, grad)
			}
		})
	}
}

// Ein genannter Testfall existiert wirklich.
//
// Ein Verweis auf einen Test, den es nicht gibt, ist schlimmer als keiner: er
// sieht aus wie ein Beleg. Genau das ist die Klasse, die in Phase 1 dreimal
// aufgetreten ist — eine Behauptung, die niemand nachgelesen hat.
func TestJederGenannteTestfallExistiert(t *testing.T) {
	vorhanden := testnamen(t)

	for _, fund := range feststellungen(t) {
		testfall := strings.TrimSpace(fund.angabe["Testfall"])
		if strings.HasPrefix(testfall, "—") || strings.HasPrefix(testfall, "-") || testfall == "" {
			continue
		}

		// **Mehrere Namen sind erlaubt, durch Komma getrennt.**
		//
		// Eine Feststellung, die zwei Tests decken, soll beide nennen können —
		// vorher galt die ganze Zeile als **ein** Name, und wer einen zweiten
		// eintrug, bekam einen Wächter, der über den Doppelnamen stolperte
		// statt über einen fehlenden Test.
		for _, name := range strings.Split(testfall, ",") {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}

			t.Run(name, func(t *testing.T) {
				if !slices.Contains(vorhanden, name) {
					t.Errorf("docs/FORMATE.md nennt für %q den Testfall %q — den gibt es nicht",
						fund.titel, name)
				}
			})
		}
	}
}

// Eine Feststellung trägt **einen** Sicherheitsgrad, nicht zwei.
//
// `ersterGrad` liest das erste Wort. Eine Angabe wie „belegt für X. Plausibel
// für Y" gilt damit als belegt — und Y darf ohne Testfall bleiben, obwohl es
// eine Vermutung ist. Der Wächter hielte die halbe Zusage und machte die
// andere unsichtbar (AGENTS.md §10).
//
// Beim Aufschreiben der MSE-Script-Grammatik sind gleich drei solcher
// Abschnitte entstanden — der Griff ist naheliegend, weil eine Beobachtung
// selten ganz belegt oder ganz vermutet ist. Der Ausweg ist nicht, das
// zuzulassen, sondern zu **trennen**: was verschieden sicher ist, ist eine
// verschiedene Feststellung.
func TestKeineFeststellungNenntZweiSicherheitsgrade(t *testing.T) {
	for _, fund := range feststellungen(t) {
		t.Run(fund.titel, func(t *testing.T) {
			genannt := genannteGrade(fund.angabe["Sicherheit"])

			if len(genannt) > 1 {
				t.Errorf("%q nennt %v — was verschieden sicher ist, gehört in "+
					"verschiedene Feststellungen", fund.titel, genannt)
			}
		})
	}
}

// genannteGrade sammelt die Grade, die als **Angabe** auftreten.
//
// Am Satzanfang, nicht irgendwo: „Vermutung — dass ein Set als Archiv vorliegt,
// ist nicht belegt" nennt einen Grad und erwähnt einen zweiten. Der erste
// Anlauf suchte die bloße Zeichenkette und machte daraus einen Befund — ein
// Wächter, der Prosa für eine Angabe hält, wird abgeschaltet statt befolgt.
func genannteGrade(sicherheit string) []string {
	var genannt []string

	for satz := range strings.SplitSeq(sicherheit, ". ") {
		grad := ersterGrad(satz)
		if grad != "" && !slices.Contains(genannt, grad) {
			genannt = append(genannt, grad)
		}
	}

	return genannt
}

func TestEsGibtUeberhauptFeststellungen(t *testing.T) {
	if len(feststellungen(t)) < 5 {
		t.Errorf("nur %d Feststellungen — hat sich der Aufbau der Datei geändert?",
			len(feststellungen(t)))
	}
}

// ------------------------------------------------------------------ Hilfsmittel

// ersterGrad nennt den Grad, mit dem eine Angabe beginnt.
//
// Hervorhebung und Großschreibung werden abgeräumt: `**Plausibel**` ist
// derselbe Grad wie `plausibel`. Ohne das prüfte der Wächter die Schreibweise
// statt der Aussage — und `**Plausibel**` rutschte durch, ausgerechnet in der
// Form, die jemand wählt, wenn ihm der Vorbehalt wichtig ist.
func ersterGrad(sicherheit string) string {
	schlicht := strings.ToLower(strings.TrimLeft(strings.TrimSpace(sicherheit), "*_"))

	for _, grad := range grade {
		if strings.HasPrefix(schlicht, strings.ToLower(grad)) {
			return grad
		}
	}

	return ""
}

// feststellungen zerlegt FORMATE.md in seine Abschnitte.
func feststellungen(t *testing.T) []feststellung {
	t.Helper()

	inhalt := string(projektdatei(t, "docs/FORMATE.md"))

	stellen := ueberschrift.FindAllStringSubmatchIndex(inhalt, -1)
	if len(stellen) == 0 {
		t.Fatal("docs/FORMATE.md enthält keine Überschriften — hat sich der Aufbau geändert?")
	}

	var gefunden []feststellung

	for i, stelle := range stellen {
		titel := strings.TrimSpace(inhalt[stelle[2]:stelle[3]])
		if keineFeststellung[titel] {
			continue
		}

		ende := len(inhalt)
		if i+1 < len(stellen) {
			ende = stellen[i+1][0]
		}

		gefunden = append(gefunden, feststellung{
			titel:  titel,
			angabe: angabenAus(inhalt[stelle[1]:ende]),
		})
	}

	return gefunden
}

// testnamen sammelt jede Testfunktion im Baum.
var testfunktion = regexp.MustCompile(`(?m)^func (Test\w+)\(`)

func testnamen(t *testing.T) []string {
	t.Helper()

	var namen []string

	fehler := filepath.WalkDir(filepath.Join(wurzel, "internal"),
		func(pfad string, eintrag os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if eintrag.IsDir() || !strings.HasSuffix(pfad, "_test.go") {
				return nil
			}

			inhalt, lesefehler := os.ReadFile(pfad) //nolint:gosec // Testdatei im Repository
			if lesefehler != nil {
				return lesefehler
			}

			for _, treffer := range testfunktion.FindAllStringSubmatch(string(inhalt), -1) {
				namen = append(namen, treffer[1])
			}

			return nil
		})
	if fehler != nil {
		t.Fatalf("Tests nicht auffindbar: %v", fehler)
	}

	if len(namen) == 0 {
		t.Fatal("keine Testfunktion gefunden")
	}

	return namen
}
