// Der Wächter über die Formatliste der Bildausgabe.
//
// # Wofür
//
// **Ein Format, das niemand wählen kann, gibt es für Anwendende nicht.** Genau
// das war der schwerwiegendste Befund des Zweitblicks über Phase P: `internal/svg`
// stand mit Tests und 100 % Abdeckung da, `svg.AusPlan` wurde außerhalb dieser
// Tests **nirgends** gerufen, und die Auswahl in der Oberfläche bot vier
// Formate an (#852).
//
// Die erste Dimension aus §10 fragt: löst dieser Beitrag ein, wofür es die
// Anwendung gibt — und **woran ist das zu sehen**? Hier: daran, dass jedes
// Format, das `service.Bildformat` kennt, in der Auswahl steht.
//
// # Was er nicht prüft
//
// **Ob das Format wirklich schreibt.** Er hält zwei Listen gegeneinander, nicht
// eine Datei gegen ihren Inhalt — dafür stehen die Läufe in
// `internal/service`. Eine Auswahl mit einem Eintrag, dessen Weg abstürzt, käme
// hier durch.
//
// **Ob die Beschriftung stimmt.** Der Wert im `<option>` ist die Zusage; was
// daneben steht, kommt aus dem Sprachkatalog und wird dort geprüft.
package werkzeuge

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// gokonstante trifft `FormatXYZ Bildformat = "wert"`.
var gokonstante = regexp.MustCompile(`Format[A-Za-z]+\s+Bildformat\s*=\s*"([a-z]+)"`)

// jsxoption trifft `<option value="wert">`.
var jsxoption = regexp.MustCompile(`<option value="([a-z0-9]+)"`)

func TestJedesBildformatStehtInDerAuswahl(t *testing.T) {
	ausGo := treffer(t, "../service/bildnamen.go", gokonstante)
	if len(ausGo) < 4 {
		t.Fatalf("%d Formate in service.Bildformat gefunden — der Leser greift daneben: %v",
			len(ausGo), ausGo)
	}

	ausJSX := optionenAus(t, "../../frontend/src/seiten/Bildereinstellungen.tsx", "format")

	fehlend := ohne(ausGo, ausJSX)
	if len(fehlend) > 0 {
		t.Errorf("die Auswahl bietet diese Formate nicht an: %s",
			strings.Join(fehlend, ", "))
	}

	// **Und die Gegenrichtung.** Ein Wert in der Auswahl, den `Bildformat`
	// nicht kennt, wird vom Dienst abgewiesen — mit einer Fehlermeldung für
	// etwas, das die Oberfläche selbst angeboten hat.
	//
	// **Gelesen wird nur die Auswahl über `format`**, nicht jede Auswahl der
	// Datei. Hier stand der Satz, das gehe gut, „solange niemand eine zweite
	// ausschreibt" — und mit der Güte (Ω.4, #1029) hat es jemand getan: vier
	// `<option>`-Werte, die kein `Bildformat` sind, machten diesen Wächter
	// rot. Ein Wächter, der sich auf die Abwesenheit von etwas verlässt,
	// wartet auf den Tag, an dem es da ist.
	fremd := ohne(ausJSX, ausGo)
	if len(fremd) > 0 {
		t.Errorf("die Auswahl bietet Formate an, die service.Bildformat nicht kennt: %s",
			strings.Join(fremd, ", "))
	}
}

// treffer liest die erste Gruppe jedes Treffers aus einer Datei.
func treffer(t *testing.T, pfad string, muster *regexp.Regexp) []string {
	t.Helper()

	roh, err := os.ReadFile(pfad) //nolint:gosec // ein fester Pfad im Repositorium
	if err != nil {
		t.Fatalf("%s lesen: %v", pfad, err)
	}

	gruppenliste := muster.FindAllStringSubmatch(string(roh), -1)
	gefunden := make([]string, 0, len(gruppenliste))

	for _, gruppen := range gruppenliste {
		gefunden = append(gefunden, gruppen[1])
	}

	sort.Strings(gefunden)

	return gefunden
}

// ohne gibt die Einträge der ersten Liste, die in der zweiten fehlen.
func ohne(diese, jene []string) []string {
	var fehlend []string

	for _, eintrag := range diese {
		if !enthaelt(jene, eintrag) {
			fehlend = append(fehlend, eintrag)
		}
	}

	return fehlend
}

// optionenAus liest die `<option value="…">` **einer** Auswahl.
//
// **Welche, sagt der gebundene Zustand.** Die Auswahl über `format` beginnt
// mit `value={format}`, die über die Güte mit `value={guete}`; gelesen wird
// von dort bis zum nächsten `</select>`.
//
// # Was er nicht prüft
//
// **Ob die Auswahl überhaupt dasteht.** Ein leeres Ergebnis meldet er als
// Fehlschlag — sonst wäre eine entfernte Auswahl dasselbe wie eine
// vollständige.
func optionenAus(t *testing.T, pfad, feld string) []string {
	t.Helper()

	roh, err := os.ReadFile(pfad) //nolint:gosec // ein fester Pfad im Repositorium
	if err != nil {
		t.Fatalf("%s lesen: %v", pfad, err)
	}

	inhalt := string(roh)

	anfang := strings.Index(inhalt, "value={"+feld+"}")
	if anfang < 0 {
		t.Fatalf("%s: keine Auswahl über %q", pfad, feld)
	}

	ende := strings.Index(inhalt[anfang:], "</select>")
	if ende < 0 {
		t.Fatalf("%s: die Auswahl über %q endet nicht", pfad, feld)
	}

	gruppenliste := jsxoption.FindAllStringSubmatch(inhalt[anfang:anfang+ende], -1)
	if len(gruppenliste) == 0 {
		t.Fatalf("%s: die Auswahl über %q bietet nichts an", pfad, feld)
	}

	gefunden := make([]string, 0, len(gruppenliste))
	for _, gruppen := range gruppenliste {
		gefunden = append(gefunden, gruppen[1])
	}

	sort.Strings(gefunden)

	return gefunden
}
