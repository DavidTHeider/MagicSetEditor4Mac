package werkzeuge_test

// **Jede Abfrage auf eine Wurzeltabelle klemmt auf ein Profil** — Werkstück 2
// von Phase 4.12
// (#1351),
// nach Vermerk 0066, Entscheidung 2.
//
// # Warum es diesen Wächter gibt
//
// **SQLite hat kein `ROW LEVEL SECURITY`.** Der Ersatz hat drei Teile, und
// der dritte ist der, der trägt:
//
//  1. eine `profil_id` in jeder Wurzeltabelle,
//  2. jede Abfrage klemmt darauf,
//  3. **ein Wächter, der eine Abfrage ohne die Klausel rot macht.**
//
// Ohne den dritten ist es eine Konvention, und Konventionen halten nicht.
//
// # Was er **nicht** hält
//
// **Die Abkömmlinge.** `card_values` wird über `card_id` gelesen, nicht über
// ein Profil — die Zeile ist nur so sicher wie die Abfrage, die die Kennung
// geliefert hat. Das ist der Unterschied zu echtem RLS, und er steht in
// Vermerk 0066 beim Namen.
//
// **Und den Sinn der Klausel.** Er liest, dass `profil_id` vorkommt; ob sie
// gegen den richtigen Wert steht, sagt kein Textmuster. Dafür stehen die
// Läufe in `internal/storage` und der Flächenlauf.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// wurzeltabellen sind die, die einem Profil gehören.
//
// **Zehn und nicht dreiunddreißig**: die Abkömmlinge erben über ihren
// Fremdschlüssel. Was **kein** Profil hat, steht in der Wanderung 0035 mit
// Begründung — der Paketkatalog gehört dem Rechner, Preise sind fremde Daten,
// `bestand` ist die Buchführung über das Schema selbst.
var wurzeltabellen = []string{
	"sets", "decks", "sammlung", "turniere", "formate",
	"settings", "export_options",
	"spiel_autoersetzungen", "spiel_listenspalten", "spiel_sortierung",
}

// ohneKlausel sind die Stellen, die bewusst nicht klemmen — mit Grund.
//
// **Jede Zeile ist eine Entscheidung.** Eine Ausnahme ohne Begründung wäre
// eine stille Absenkung der Zusage, und genau das verhindert dieser Wächter.
var ohneKlausel = map[string]string{
	"migrationen": "die Wanderungen legen die Spalte erst an und füllen sie",
	"waisenzaehler.go": "die Diagnose zählt **alle** verwaisten Zeilen, über " +
		"jedes Profil hinweg — sie ist die Frage nach der Gesundheit der Datei " +
		"und nicht nach dem Besitz",
	"mindestform.go": "sie prüft die Gestalt des Schemas, nicht seinen Inhalt",
	"herkunftverweis.go": "der Eingang des Dienstes gehört dem Rechner (#1944): eine Quelle, " +
		"auf die ein **fremdes** Profil noch verweist, darf nicht fort. Gefragt wird nur ja " +
		"oder nein, und kein Wert eines Sets verlässt die Abfrage",
	"pflegestellen.go": "die Analyse zählt **verwaiste** Zeilen (O8, #1432) — " +
		"also gerade die, die zu **keinem** Profil mehr gehören, weil ihr " +
		"Elternteil fehlt. Ein `profil_id` in der Klausel schlösse sie aus, " +
		"und der Wächter über die Gesundheit der Datei sähe nichts mehr; " +
		"dieselbe Lage wie bei `waisenzaehler.go` darüber. **Was ein Profil " +
		"hat, klemmt trotzdem**: `moeglicheDubletten` gruppiert nach " +
		"`profil_id`, und die Sammlung wird nicht angefasst. **Und was ein " +
		"eigenes hat, klemmt darauf**: `alteProtokolle` trägt " +
		"`profilgebunden` (G5, #1462) — eine beendete Partie gehört einem " +
		"Profil, und ein Pflegelauf, der das überginge, böte ein fremdes " +
		"Protokoll zum Löschen an",
	"bestandsstand.go": "der Vergleich vor einer Rückkehr geht über **jedes** " +
		"Profil (O7, #1431) — ein Zurückspielen tauscht den ganzen Bestand, " +
		"und wer nur das eigene Profil zählte, verschwiege die Unsicherheit " +
		"der anderen. **Die Trennung bleibt trotzdem**: jede gelesene Menge " +
		"trägt ihr `profil_id` mit (`sammlungsmengen` gruppiert danach), und " +
		"das **Lesen** der Epoche klemmt wieder — `wiederherstellungen.go` " +
		"gibt einem Profil nur seine eigenen Mengen",
}

// **Jede Abfrage auf eine Wurzeltabelle nennt `profil_id`.**
func TestJedeWurzelabfrageKlemmtAufEinProfil(t *testing.T) {
	t.Parallel()

	muster := regexp.MustCompile(
		`(?i)\b(FROM|INTO|UPDATE|JOIN)\s+(` + strings.Join(wurzeltabellen, "|") + `)\b`)

	dateien, err := filepath.Glob(filepath.Join("..", "storage", "*.go"))
	if err != nil {
		t.Fatalf("die Dateien finden: %v", err)
	}

	geprueft := 0

	for _, pfad := range dateien {
		name := filepath.Base(pfad)

		if strings.HasSuffix(name, "_test.go") || begruendet(name) {
			continue
		}

		inhalt, lesefehler := os.ReadFile(pfad) //nolint:gosec // aus Glob
		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", pfad, lesefehler)
		}

		for _, abfrage := range abfragen(string(inhalt)) {
			if !muster.MatchString(abfrage) {
				continue
			}

			geprueft++

			if !strings.Contains(abfrage, "profil_id") {
				t.Errorf("%s: eine Abfrage auf eine Wurzeltabelle klemmt nicht "+
					"auf ein Profil:\n%s", name, gekuerzt(abfrage))
			}
		}
	}

	// **Und der Wächter selbst muss etwas gesehen haben.** Ein Muster, das
	// danebengreift, wäre grün und prüfte nichts — dieselbe Falle wie beim
	// Kontrastwächter, der 668 Klassen nicht sah.
	if geprueft < 50 {
		t.Errorf("nur %d Abfragen geprüft — das Muster greift daneben", geprueft)
	}
}

// **Jede Ausnahme gibt es, und sie ist begründet.**
func TestJedeAusnahmeVonDerProfilklauselIstBegruendet(t *testing.T) {
	t.Parallel()

	for name, grund := range ohneKlausel {
		if len(grund) < 40 {
			t.Errorf("die Ausnahme %q ist nicht begründet", name)
		}
	}
}

// begruendet sagt, ob eine Datei ausgenommen ist.
func begruendet(name string) bool {
	for ausgenommen := range ohneKlausel {
		if strings.Contains(name, ausgenommen) {
			return true
		}
	}

	return false
}

// abfragen schneidet die SQL-Zeichenketten aus einer Go-Datei.
//
// **Rohzeichenketten und nichts sonst.** Jede Abfrage dieses Speichers steht
// zwischen Backticks; eine zusammengesetzte gäbe es nur dort, wo `gosec`
// ohnehin nachfragt.
func abfragen(inhalt string) []string {
	var gefunden []string

	teile := strings.Split(inhalt, "`")
	for i := 1; i < len(teile); i += 2 {
		gefunden = append(gefunden, teile[i])
	}

	return gefunden
}

// gekuerzt gibt die ersten Zeilen einer Abfrage.
func gekuerzt(abfrage string) string {
	zeilen := strings.Split(strings.TrimSpace(abfrage), "\n")
	if len(zeilen) > 3 {
		zeilen = append(zeilen[:3], "…")
	}

	return strings.Join(zeilen, "\n")
}
