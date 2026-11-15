// Der Wächter über die gebauten Bildverknüpfungen.
//
// # Wofür
//
// **Eine Funktion, die die Sprache kennt und der Rasterer nicht, zeichnet
// stillschweigend das falsche Bild.** `internal/msescript` trägt jede
// Bildfunktion in seine Standardbibliothek ein; `internal/render` führt aus,
// was in `gebauteVerknuepfungen` steht. Läuft das auseinander, bekommt eine
// fremde Vorlage ihr **ungedrehtes, ungefiltertes Ausgangsbild** und einen
// Befund, den niemand liest.
//
// Genau das war der erste Befund des Zweitblicks über die Phasen 4.0 bis 4.9:
// `rotate`, `rotate_image` und `built_in_image` waren registriert und nicht
// gebaut.
//
// # Was er nicht prüft
//
// **Ob die Ausführung stimmt.** Er hält zwei Listen gegeneinander. Eine
// Verknüpfung, die eingetragen ist und falsch rechnet, käme hier durch — dafür
// stehen die Tests in `internal/render`.
package werkzeuge

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// eintragungen liest die Namen aus `eintragen("…", …)`.
var eintragungen = regexp.MustCompile(`eintragen\("([a-z_0-9]+)"`)

// knotenarten liest die Namen aus `verknuepfeBild("…", …)` — die **Knotenart**
// kann von der Funktion abweichen: `rotate_image` erzeugt einen `rotate`.
var knotenarten = regexp.MustCompile(`verknuepfeBild\("([a-z_0-9]+)"`)

// eintraegeDerListe liest die Schlüssel einer Go-Map aus ihrem Literal.
//
// **Seit dem 2026-09-03 trägt jeder Eintrag seine Ausführung** (Befund 3 des
// Nachlaufs über die Phasen 4.0 bis 4.9): `"rotate": {fuehreAus: …}` statt
// `"rotate": true`. Das war die Korrektur, gegen die dieser Wächter dann
// selbst rot lief — und das ist richtig so: er las eine Form, die es nicht
// mehr gibt.
var eintraegeDerListe = regexp.MustCompile(`"([a-z_0-9]+)":\s+\{`)

func TestJedeRegistrierteBildfunktionIstGebaut(t *testing.T) {
	t.Parallel()

	gebaut := gebauteVerknuepfungenLesen(t)

	for name, knoten := range registrierteBildknoten(t) {
		if !gebaut[knoten] {
			t.Errorf("`%s` erzeugt die Knotenart `%s`, und der Rasterer führt "+
				"sie nicht aus — eintragen in `gebauteVerknuepfungen` "+
				"(internal/render/bildbaum.go) oder hier begründen", name, knoten)
		}
	}
}

// registrierteBildknoten gibt je registrierter Bildfunktion ihre Knotenart.
//
// **Nicht jede Eintragung ist eine Verknüpfung.** `to_image` wandelt einen
// Wert und erzeugt keinen eigenen Knoten; `built_in_image` erzeugt einen ohne
// `verknuepfeBild`. Beide Sonderfälle stehen hier namentlich, damit ein
// dritter auffällt statt durchzulaufen.
func registrierteBildknoten(t *testing.T) map[string]string {
	t.Helper()

	const ort = "../msescript/bibliothek_bild.go"

	roh, err := os.ReadFile(ort) //nolint:gosec // ein fester Pfad im Repositorium
	if err != nil {
		t.Fatalf("%s lesen: %v", ort, err)
	}

	knoten := map[string]string{}

	for _, zeile := range strings.Split(string(roh), "\n") {
		nackt := strings.TrimSpace(zeile)
		if strings.HasPrefix(nackt, "//") {
			continue
		}

		name := eintragungen.FindStringSubmatch(nackt)
		if name == nil {
			continue
		}

		switch name[1] {
		case "to_image":
			// Eine Wandlung, kein Knoten.
			continue
		case "built_in_image":
			knoten[name[1]] = "built_in_image"

			continue
		}

		art := knotenarten.FindStringSubmatch(nackt)
		if art == nil {
			t.Fatalf("`%s` ist eingetragen, und die Zeile nennt keine "+
				"Knotenart — der Wächter kann sie nicht zuordnen: %s",
				name[1], nackt)
		}

		knoten[name[1]] = art[1]
	}

	if len(knoten) < 10 {
		t.Fatalf("nur %d Bildfunktionen gefunden — der Leser greift ins Leere",
			len(knoten))
	}

	return knoten
}

// gebauteVerknuepfungenLesen liest die Liste des Rasterers.
func gebauteVerknuepfungenLesen(t *testing.T) map[string]bool {
	t.Helper()

	const ort = "../render/bildbaum.go"

	roh, err := os.ReadFile(ort) //nolint:gosec // ein fester Pfad im Repositorium
	if err != nil {
		t.Fatalf("%s lesen: %v", ort, err)
	}

	nach := strings.SplitN(string(roh),
		"\tgebauteVerknuepfungen = map[string]verknuepfung{", 2)
	if len(nach) != 2 {
		t.Fatal("`gebauteVerknuepfungen` steht nicht mehr in internal/render/bildbaum.go")
	}

	block, _, gefunden := strings.Cut(nach[1], "\n\t}")
	if !gefunden {
		t.Fatal("das Literal von `gebauteVerknuepfungen` endet nicht")
	}

	gebaut := map[string]bool{}
	for _, treffer := range eintraegeDerListe.FindAllStringSubmatch(block, -1) {
		gebaut[treffer[1]] = true
	}

	if len(gebaut) < 10 {
		t.Fatalf("nur %d gebaute Verknüpfungen gefunden — der Leser greift ins Leere",
			len(gebaut))
	}

	namen := make([]string, 0, len(gebaut))
	for name := range gebaut {
		namen = append(namen, name)
	}

	sort.Strings(namen)
	t.Logf("gebaut: %s", strings.Join(namen, ", "))

	return gebaut
}
