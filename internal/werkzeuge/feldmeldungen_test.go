// Der Wächter über den Fremdtext in einem Feldbefund —
// Befund 3 des Zweitblicks über Phase 4.13
// (#1352).
//
// # Die Lücke
//
// `meldungswege_test.go` hält die Zusage aus §8 an **einer** Stelle: jedem
// `fmt.Errorf`, das eines der neun Wahrzeichen umschließt. `FieldError` geht
// daran vorbei — er entsteht über [domain.NewFieldError] mit einer Vorlage und
// Werten, und wer dort `%v` mit einem fremden Fehler füllt, schreibt dessen
// Text in `Message`. Die Fassade gibt ihn unverändert weiter.
//
// **Sechs Stellen taten das am 2026-09-07**, und der Zweitblick hat sie
// angestrichen.
//
// # Warum nicht alle sechs falsch sind
//
// **Bei fremden Daten ist die Stelle die Auskunft** (§8): „ein Set, von dem
// 300 Karten lesbar sind und zwei nicht, wird mit 300 Karten und zwei
// benannten Befunden geöffnet". Ein JSON-Fehler, der Zeile und Zeichen nennt,
// ist genau das — ihn zu verschlucken machte aus „hier ist der Fehler" ein
// „irgendetwas stimmt nicht".
//
// **Ein `strconv`-Fehler ist es nicht.** „invalid syntax" sagt niemandem
// etwas, den es angeht.
//
// # Was der Wächter deshalb verlangt
//
// Nicht das Verbot, sondern die **Entscheidung**: jede Stelle, die einen
// Fremdfehler in einen Feldbefund setzt, steht unten mit ihrem Grund. Wer eine
// neue schreibt, wird gefragt.
//
// # Was er nicht prüft
//
// **Ob der Grund stimmt.** Er sieht, dass einer dasteht — nicht, dass er
// trägt. Das ist eine Frage ans Review (§10).
package werkzeuge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// feldmeldungsausnahme ist eine Stelle, an der ein fremder Text in einen
// Feldbefund gehört — **mit ihrem Grund**.
type feldmeldungsausnahme struct {
	datei    string
	satzteil string
	grund    string
}

var feldmeldungsausnahmen = []feldmeldungsausnahme{
	{
		"gueltigkeit.go", "Das ist kein JSON",
		"fremde Daten: der Fehler nennt Zeile und Zeichen, und das ist die Auskunft (§8)",
	},
	{
		"gueltigkeit.go", "weder ein Set noch eine Abbildung",
		"dasselbe — die Gestalt der fremden Datei ist der Befund",
	},
	{
		"gueltigkeit.go", "ist kein Set",
		"dasselbe, je Set",
	},
	{
		"paketanlage.go", "ist nicht lesbar",
		"der Fehler kommt aus `msepaket` und nennt Datei, Zeile und Feld — eigener Text, keine Go-Interna",
	},

	// **Der Austausch der Einstellungen** — Werkstück O3 von Phase 5.2
	// (#1427).
	// Vier Stellen, ein Grund: die Datei kommt von irgendwoher, und wo sie
	// bricht, ist die Auskunft (§8). „Diese Datei ist kein JSON" ohne Zeile
	// und Zeichen schickt jemanden auf die Suche durch eine Datei, die er
	// nicht geschrieben hat.
	{
		"einstellungsaustausch.go", "ist kein JSON",
		"fremde Daten: der Fehler nennt Zeile und Zeichen, und das ist die Auskunft (§8)",
	},
	{
		"einstellungsaustausch.go", "trägt kein Objekt",
		"dasselbe — die Gestalt der fremden Datei ist der Befund",
	},
	{
		"einstellungsaustausch.go", "ist keine Zahl",
		"dasselbe, für das Feld »fassung«",
	},
	{
		"einstellungsaustausch.go", "nicht die erwartete Gestalt",
		"dasselbe, für den Block »bereiche« — welcher Abschnitt bricht, steht im Fehler",
	},

	// **Die Partiedatei** — PLAY-03.AK-05 (#1711). Dieselben vier Fälle wie
	// eine Zeile darüber und aus demselben Grund: was hier hereinkommt, ist
	// eine Datei von irgendwoher, und **wo** sie bricht, ist die Auskunft.
	{
		"partieaustausch.go", "ist kein JSON",
		"fremde Daten: der Fehler nennt Zeile und Zeichen, und das ist die Auskunft (§8)",
	},
	{
		"partieaustausch.go", "trägt kein Objekt",
		"dasselbe — die Gestalt der fremden Datei ist der Befund",
	},
	{
		"partieaustausch.go", "ist keine Zahl",
		"dasselbe, für das Feld »fassung«",
	},
	{
		"partieaustausch.go", "nicht die erwartete Gestalt",
		"dasselbe, für den Block »partie« — welcher Teil bricht, steht im Fehler",
	},

	// **Die Turnierdatei** — DOM-04.AK-03 (#1699). Dieselben vier Fälle wie
	// bei der Partie und aus demselben Grund.
	{
		"turnieraustausch.go", "ist kein JSON",
		"fremde Daten: der Fehler nennt Zeile und Zeichen, und das ist die Auskunft (§8)",
	},
	{
		"turnieraustausch.go", "trägt kein Objekt",
		"dasselbe — die Gestalt der fremden Datei ist der Befund",
	},
	{
		"turnieraustausch.go", "ist keine Zahl",
		"dasselbe, für das Feld »fassung«",
	},
	{
		"turnieraustausch.go", "nicht die erwartete Gestalt",
		"dasselbe, für den Block »turnier« — welcher Teil bricht, steht im Fehler",
	},
}

func TestJederFremdtextImFeldbefundIstEntschieden(t *testing.T) {
	t.Parallel()

	dateien := gowurzeln(t, "..")
	ohne := make([]string, 0, len(dateien))

	// **Über `internal` ganz und nicht über sechs ausgewählte Pakete** —
	// Nachlauf zum Zweitblick (2026-09-07). `bezug` und `netz` bauen
	// ebenfalls Feldbefunde, und eine Aufzählung, die sie nicht kennt, ist
	// grün, ohne sie gesehen zu haben.
	for _, datei := range dateien {
		roh, err := os.ReadFile(datei) //nolint:gosec // Pfad aus dem Baum selbst
		if err != nil {
			t.Fatalf("%s: %v", datei, err)
		}

		ohne = append(ohne, fremdtexte(filepath.Base(datei), string(roh))...)
	}

	if len(ohne) > 0 {
		t.Errorf("diese Stellen setzen einen fremden Fehlertext in einen "+
			"Feldbefund, und er erreicht die Oberfläche. Entweder er gehört "+
			"dorthin — dann trage die Stelle mit ihrem Grund in "+
			"`feldmeldungsausnahmen` ein — oder er gehört über "+
			"`domain.MitGrund` ins Protokoll:\n%s", strings.Join(ohne, "\n"))
	}
}

// fremdtexte findet die Feldbefunde einer Datei, die einen Platzhalter für
// einen Fehler tragen.
//
// **Über den Text und nicht über den AST**: ein `NewFieldError` steht über
// mehrere Zeilen, und die Vorlage ist ein Literal. Was zählt, ist das `%v` oder
// `%w` darin — und das steht immer in derselben Zeichenkette.
func fremdtexte(name, inhalt string) []string {
	var gefunden []string

	zeilen := strings.Split(inhalt, "\n")

	for nummer, zeile := range zeilen {
		// **Beide Wege** — Nachlauf zum Zweitblick: `FieldError` ist ein
		// **exportierter** Verbund, und wer ihn unmittelbar baut, geht am
		// Konstruktor vorbei. Der Wächter kannte nur den Konstruktor.
		if !strings.Contains(zeile, "NewFieldError(") &&
			!strings.Contains(zeile, "FieldError{") {
			continue
		}

		// **Die Vorlage steht in den nächsten Zeilen** — bei einem Verbund
		// mit benannten Feldern können es mehr sein als bei einem Ruf.
		bis := min(nummer+6, len(zeilen))
		block := strings.Join(zeilen[nummer:bis], "\n")

		if !strings.Contains(block, "%v") && !strings.Contains(block, "%w") {
			continue
		}

		if entschieden(name, block) {
			continue
		}

		gefunden = append(gefunden,
			name+":"+itoa(nummer+1)+": "+strings.TrimSpace(zeile))
	}

	return gefunden
}

// entschieden sagt, ob eine Stelle mit ihrem Grund geführt wird.
func entschieden(name, block string) bool {
	for _, ausnahme := range feldmeldungsausnahmen {
		if ausnahme.datei == name && strings.Contains(block, ausnahme.satzteil) &&
			ausnahme.grund != "" {
			return true
		}
	}

	return false
}
