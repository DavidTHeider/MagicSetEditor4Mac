// Der Wächter über die Güteliste der Bildausgabe.
//
// # Wofür
//
// Dieselbe Frage wie beim Format, und aus demselben Anlass: **eine Güte, die
// niemand wählen kann, gibt es für Anwendende nicht.** Werkstück Ω.4 verlangt
// ausdrücklich, dass alle vier über **jede** Fläche erreichbar sind
// (#1029).
//
// Die drei Flächen ohne Fenster fährt `internal/flaechen/guete_test.go`
// wirklich; das Fenster läuft in keinem Go-Prozess, und für es hält dieser
// Wächter die Liste.
//
// # Was er nicht prüft
//
// **Ob die Wahl ankommt.** Dass der gewählte Wert bis zur Ausfuhr reist, steht
// in `internal/app/bildausfuhr_test.go`; dass er dort etwas ändert, in
// `internal/render`. Dieser Wächter hält zwei Listen gegeneinander.
//
// **Ob die Beschriftung stimmt.** Wie beim Format: der Wert im `<option>` ist
// die Zusage, der Text daneben kommt aus dem Sprachkatalog.
package werkzeuge

import (
	"regexp"
	"strings"
	"testing"
)

// gonamensfall trifft `case GueteXYZ:` samt dem `return "wert"` darunter.
//
// **Aus `Guete.Name` und nicht aus `Gueten`**: die Aufzählung nennt die Werte,
// die Namensfunktion die Zeichenketten — und über die Flächen reist die
// Zeichenkette.
var gonamensfall = regexp.MustCompile(`case Guete[A-Za-z]+:\s*\n\s*return "([a-z]+)"`)

func TestJedeGueteStehtInDerAuswahl(t *testing.T) {
	ausGo := treffer(t, "../render/guete.go", gonamensfall)
	if len(ausGo) != 4 {
		t.Fatalf("%d Güten in render.Guete gefunden, erwartet sind vier: %v",
			len(ausGo), ausGo)
	}

	ausJSX := optionenAus(t, "../../frontend/src/seiten/Bildereinstellungen.tsx", "guete")

	if fehlend := ohne(ausGo, ausJSX); len(fehlend) > 0 {
		t.Errorf("die Auswahl bietet diese Güten nicht an: %s",
			strings.Join(fehlend, ", "))
	}

	if fremd := ohne(ausJSX, ausGo); len(fremd) > 0 {
		t.Errorf("die Auswahl bietet Güten an, die render.Guete nicht kennt: %s",
			strings.Join(fremd, ", "))
	}
}
