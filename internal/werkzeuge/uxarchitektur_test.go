// **Jede Seite hat einen Arbeitsweg, und jeder Kernweg hat seine Zahl** —
// Werkstück U1 von Phase 5.1
// (#1412).
//
// # Warum es diesen Wächter gibt
//
// `docs/UX-ARCHITEKTUR.md` ordnet jede Seite und jede Lasche einem der
// sieben Arbeitswege zu und nennt für jeden Kernweg den gemessenen
// Ausgangswert und das Ziel. Die Abnahme des Issues verlangt beides —
// *„keine Funktion verschwindet durch Neugruppierung"* und *„gemessener
// Ausgangswert und numerisches Ziel, bevor ein Arbeitsplatz umgebaut wird"*.
//
// Eine Tabelle von Hand ist grün, sobald sie zu kurz ist (`AGENTS.md` §3).
// Die Seitenliste kommt deshalb aus `navigation.ts`, und was dort steht,
// muss hier eine Zeile haben — **und umgekehrt**: ein Schlüssel im Dokument,
// den es nicht gibt, ist eine Seite, die jemand sucht und nicht findet.
//
// # Was er hält und was nicht
//
//   - **Er hält**: jede Seite aus `navigation.ts` und die zwei reservierten
//     Bereiche stehen in der Seitentabelle, jede Zeile nennt mindestens
//     einen der sieben Kernwege, jeder Kernweg hat eine Zeile mit
//     Ausgangswert und Ziel, und die Reservierung von ⌘9 und ⌘0 steht in
//     beiden Dateien gleich.
//   - **Er hält nicht**, ob die Zuordnung trifft, ob die Zahl stimmt und ob
//     das Ziel erreichbar ist. Das sieht nur ein Mensch — und der Messlauf
//     daneben (`scripts/arbeitswegmass.mjs`), der die Zahl erhebt.
package werkzeuge_test

import (
	"regexp"
	"strings"
	"testing"
)

// uxKernwege sind die sieben Arbeitswege — die Schlüssel, die die Tabelle
// trägt. Sechs gibt es, einer ist reserviert.
var uxKernwege = []string{
	"karte", "paket", "sammlung", "deck", "profil", "spielen", "handeln",
}

// uxReservierteBereiche sind die Bereiche, die es noch nicht gibt; sie stehen
// in der Seitentabelle mit ihrer Taste.
//
// **Keiner mehr, und das ist das Ende der Reservierung.** `spielen` stand
// hier bis zum 2026-09-15 und `handeln` bis zum 2026-09-19; beide sind seit
// Werkstück G1 von Phase 5.5
// (#1458)
// und `T9` von Phase 5.6
// (#1474)
// **Seiten** — sie kommen damit über `anwendungsschluessel` herein wie jede
// andere. Eine Reservierung endet damit, dass ihr Bereich da ist; stünde sie
// weiter hier, verlangte dieser Wächter eine Zeile, die es zweimal gäbe.
//
// **Die Aufzählung bleibt, weil die nächste Reservierung eine Stelle
// braucht** — dieselbe Überlegung wie bei `reservierteKuerzel` drüben. Was
// sie kostet, steht am Lauf darunter: über einer leeren Liste prüft
// [TestDieReservierungStehtInBeidenDateien] nichts, und das steht dort.
var uxReservierteBereiche = map[string]string{}

// uxZeile findet eine Zeile, deren erste Zelle einen Schlüssel in
// Backticks trägt: „| `karten` | … |".
var uxZeile = regexp.MustCompile("(?m)^\\| `([a-z]+)` \\|(.*)$")

// uxReservierungImQuelltext findet `'9': 'spielen'` in `navigation.ts`.
var uxReservierungImQuelltext = regexp.MustCompile(`'([09])': '([a-z]+)'`)

func TestJedeSeiteHatEinenArbeitsweg(t *testing.T) {
	t.Parallel()

	dokument := lies(t, "docs/UX-ARCHITEKTUR.md")
	navigation := lies(t, "frontend/src/navigation.ts")

	seitentabelle := uxAbschnitt(t, dokument, "## Jede Seite und jede Lasche")
	zeilen := uxZeilenAus(seitentabelle)

	// **Vierundzwanzig Seiten und ein Bereich** — die Zahl ist die Vorgabe
	// der Länge, nicht ihre Prüfung; die steht unten.
	gefuehrt := make([]string, 0, 25)
	for _, liste := range []string{
		"export const anwendungsschluessel = [",
		"export const setschluessel = [",
	} {
		gefuehrt = append(gefuehrt, schluesselAus(navigation, liste)...)
	}

	// **Eine leere Liste macht diesen Wächter hohl** — dieselbe Untergrenze
	// wie in `serienmassseiten_test.go`.
	if len(gefuehrt) < 10 {
		t.Fatalf("navigation.ts führt %d Seiten — der Leser greift ins Leere", len(gefuehrt))
	}

	for bereich := range uxReservierteBereiche {
		gefuehrt = append(gefuehrt, bereich)
	}

	bekannt := map[string]bool{}
	for _, schluessel := range gefuehrt {
		bekannt[schluessel] = true

		zeile, da := zeilen[schluessel]
		if !da {
			t.Errorf("die Seite %q hat keine Zeile in docs/UX-ARCHITEKTUR.md — sie "+
				"ist keinem Arbeitsweg zugeordnet (#1412)", schluessel)

			continue
		}

		if !uxNenntKernweg(zeile) {
			t.Errorf("die Zeile zu %q nennt keinen der sieben Kernwege %v", schluessel, uxKernwege)
		}
	}

	// **Und umgekehrt.**
	for schluessel := range zeilen {
		if !bekannt[schluessel] {
			t.Errorf("docs/UX-ARCHITEKTUR.md führt die Seite %q — navigation.ts kennt "+
				"sie nicht", schluessel)
		}
	}
}

func TestJederKernwegHatAusgangswertUndZiel(t *testing.T) {
	t.Parallel()

	dokument := lies(t, "docs/UX-ARCHITEKTUR.md")
	zeilen := uxZeilenAus(uxAbschnitt(t, dokument, "## Die Kernwege, gemessen"))

	for _, kernweg := range uxKernwege {
		zeile, da := zeilen[kernweg]
		if !da {
			t.Errorf("der Kernweg %q hat keine Zeile in der Messtabelle", kernweg)

			continue
		}

		zellen := strings.Split(zeile, "|")
		if len(zellen) < 5 {
			t.Errorf("die Zeile zu %q hat %d Zellen — erwartet sind Ausgangsbestand, "+
				"Ausgangswert und Ziel", kernweg, len(zellen))

			continue
		}

		for _, zelle := range zellen[1 : len(zellen)-1] {
			zelle = strings.TrimSpace(zelle)
			if zelle == "" || zelle == "?" || strings.EqualFold(zelle, "offen") {
				t.Errorf("die Zeile zu %q hat eine leere Zelle — eine Messung ohne "+
					"Zahl ist keine (AGENTS.md §10)", kernweg)
			}
		}

		// **Das Ziel ist eine Zahl.** Die letzte Zelle muss eine Ziffer
		// tragen; „später" oder „besser" wäre kein numerisches Ziel.
		ziel := strings.TrimSpace(zellen[len(zellen)-2])
		if !strings.ContainsAny(ziel, "0123456789") {
			t.Errorf("das Ziel zu %q lautet %q — keine Zahl darin", kernweg, ziel)
		}
	}
}

// **Was er hält, solange es keine Reservierung gibt: nichts.**
//
// Er läuft über [uxReservierteBereiche], und die ist seit `T9` leer — eine
// leere Schleife ist grün, ohne etwas gesehen zu haben (§10). Damit das
// dasteht und nicht nur gilt, prüft er zuerst die **Gegenrichtung**: was
// `navigation.ts` reserviert, muss hier stehen. Ein Eintrag dort, den
// niemand hier nachträgt, macht ihn rot statt still durchzulassen.
func TestDieReservierungStehtInBeidenDateien(t *testing.T) {
	t.Parallel()

	navigation := lies(t, "frontend/src/navigation.ts")
	dokument := lies(t, "docs/UX-ARCHITEKTUR.md")

	imQuelltext := map[string]string{}
	for _, treffer := range uxReservierungImQuelltext.FindAllStringSubmatch(navigation, -1) {
		imQuelltext[treffer[2]] = "⌘" + treffer[1]
	}

	zeilen := uxZeilenAus(uxAbschnitt(t, dokument, "## Jede Seite und jede Lasche"))

	// **Die Gegenrichtung zuerst** — siehe den Kopf dieses Laufs.
	for bereich, taste := range imQuelltext {
		if uxReservierteBereiche[bereich] != taste {
			t.Errorf("navigation.ts reserviert %s für %q — dieser Wächter "+
				"kennt die Reservierung nicht", taste, bereich)
		}
	}

	for bereich, taste := range uxReservierteBereiche {
		if imQuelltext[bereich] != taste {
			t.Errorf("navigation.ts reserviert %s nicht für %q (gefunden: %q)",
				taste, bereich, imQuelltext[bereich])
		}

		if !strings.Contains(zeilen[bereich], taste) {
			t.Errorf("die Zeile zu %q in docs/UX-ARCHITEKTUR.md nennt %s nicht", bereich, taste)
		}
	}
}

// mindestfenster findet `mindestbreite = 1024` und `mindesthoehe = 700` in
// der Hülle.
var mindestfenster = regexp.MustCompile(`mindest(breite|hoehe)\s+=\s+(\d+)`)

// fensterquelle ist, wo das Mindestmaß steht — seit der Klemmung
// (UX-14.AK-04) in `bildschirm.go`, denn sie braucht es ebenfalls; die
// Fensteroptionen in `ereignisse.go` nehmen es von dort
// ([fensteroptionenNehmenEs]).
const fensterquelle = "internal/fenster/bildschirm.go"

// fensteroptionenNehmenEs findet `MinWidth: mindestbreite` und
// `MinHeight: mindesthoehe` in den Fensteroptionen.
var fensteroptionenNehmenEs = regexp.MustCompile(`Min(Width|Height):\s+mindest(breite|hoehe)`)

// **Das Mindestfenster steht in der Hülle und im Dokument gleich** — Vermerk
// 0076 legt es fest, die Hülle setzt es, und `docs/UX-ARCHITEKTUR.md` nennt
// es. Drei Stellen, eine Zahl; ohne Wächter liefe die im Dokument als erste
// davon.
func TestDasMindestfensterStehtInBeidenDateien(t *testing.T) {
	t.Parallel()

	quelle := lies(t, fensterquelle)
	dokument := lies(t, "docs/UX-ARCHITEKTUR.md")

	masse := map[string]string{}
	for _, treffer := range mindestfenster.FindAllStringSubmatch(quelle, -1) {
		masse[treffer[1]] = treffer[2]
	}

	if masse["breite"] == "" || masse["hoehe"] == "" {
		t.Fatalf("%s nennt kein Mindestfenster — gefunden: %v", fensterquelle, masse)
	}

	if genommen := fensteroptionenNehmenEs.FindAllString(lies(t, "internal/fenster/ereignisse.go"), -1); len(genommen) != 2 {
		t.Errorf("die Fensteroptionen nehmen das Mindestmaß nicht aus %s: %v", fensterquelle, genommen)
	}

	// **In der Schreibweise des Dokuments**: „1.024 × 700", mit Tausenderpunkt.
	breite := masse["breite"]
	if len(breite) == 4 {
		breite = breite[:1] + "." + breite[1:]
	}

	erwartet := breite + " × " + masse["hoehe"]
	if !strings.Contains(dokument, erwartet) {
		t.Errorf("docs/UX-ARCHITEKTUR.md nennt das Mindestfenster %q nicht, das die Hülle setzt", erwartet)
	}
}

// uxAbschnitt schneidet den Text zwischen einer Überschrift und der nächsten
// gleichrangigen heraus — und scheitert, wenn es sie nicht gibt.
func uxAbschnitt(t *testing.T, dokument, ueberschrift string) string {
	t.Helper()

	_, ab, gefunden := strings.Cut(dokument, "\n"+ueberschrift+"\n")
	if !gefunden {
		t.Fatalf("docs/UX-ARCHITEKTUR.md hat keinen Abschnitt %q", ueberschrift)
	}

	bis, _, _ := strings.Cut(ab, "\n## ")

	return bis
}

// uxZeilenAus liest die Tabellenzeilen eines Abschnitts, je Schlüssel eine.
func uxZeilenAus(abschnitt string) map[string]string {
	zeilen := map[string]string{}

	for _, treffer := range uxZeile.FindAllStringSubmatch(abschnitt, -1) {
		zeilen[treffer[1]] = treffer[2]
	}

	return zeilen
}

// uxNenntKernweg sagt, ob eine Zeile einen der sieben Kernwege trägt —
// in Backticks, damit „Karte" im Fließtext nicht zählt.
func uxNenntKernweg(zeile string) bool {
	for _, kernweg := range uxKernwege {
		if strings.Contains(zeile, "`"+kernweg+"`") {
			return true
		}
	}

	return false
}
