// **Der lang laufende Agent** — Werkstück A14 von Phase 5.3
// (#1453).
//
// # Die Abnahme, Wort für Wort
//
// > Ein lang laufender echter MCP-Prozess liest Karte, erzeugt Vorschlag,
// > verweigert Apply mit Leserecht und erlaubt nur einen gültigen
// > autorisierten Apply.
//
// **Vier Aussagen, und drei davon hängen an der Reihenfolge.** Ein Lauf, der
// sie einzeln über je einen frischen Prozess führte, prüfte vier Fälle und
// **keine** Sitzung: was ein Ruf hinterlässt, sähe der nächste nie. Deshalb
// [mcpprozess] und nicht [mcpRuf].
//
// # Was hier gemessen wird, und woran
//
//   - **Ein Vorschlag ist keine Übernahme** — an der Karte, nicht an der
//     Antwort. Der Text sagt, was das Modell vorschlägt; ob etwas geschrieben
//     wurde, sagt nur ein zweites Lesen.
//   - **Ein Feldname, den der Vorschlag nicht trägt, schreibt nichts** — und
//     zwar auch nicht die Felder, die er trägt.
//   - **Genau eine Übernahme** — am Verlauf. Er zählt Schritte; die Karte
//     zeigte auch nach zwei gleichen Übernahmen denselben Wert.
//   - **Ein Leser bekommt keine** — und liest und schlägt trotzdem vor. Wäre
//     schon der Vorschlag abgewiesen, sagte der Lauf nichts über die
//     Übernahme.
//
// # Die vier Gegenproben, und was die letzte gezeigt hat
//
// Jede Zusage ist einmal entschärft worden, und jedes Mal wurde der Lauf rot
// (§10). Drei davon waren erwartbar:
//
//   - `bezug.Uebernehmen` fest auf `true` — der Vorschlag schreibt.
//   - `gewaehlteFelder` übergeht unbekannte Namen — die halbe Auswahl steht.
//   - Zweimal übernehmen — zwei Schritte im Verlauf.
//
// **Die vierte hat den ersten Anlauf widerlegt.** Ich hatte vermutet, dass
// den Leser der Schreibweg aufhält, und `app.Schreibnaht(db, …)` statt
// `arbeit` gesetzt: der Lauf blieb **grün**. Dann `Akteur.Darf` entschärft:
// wieder **grün**. Erst `preise{db: db}` in `internal/mcpnaehte/naehte.go`
// machte ihn rot.
//
// **Also hält genau eine Zeile diese Zusage** — die nur lesende Verbindung,
// die die KI-Naht bekommt. Weder die Akteursprüfung noch die Schreibnaht
// liegen auf diesem Weg; das ist die Bauweise aus
// [0073](../../docs/entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)
// und kein Mangel — aber gemessen und nicht angenommen.
//
// # Was er nicht deckt
//
// **Einen Anbieter, der Geld kostet.** Der braucht eine Freigabe, die nur das
// Fenster ausstellt ([0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md)),
// und ein Cloudanbieter mit Modell braucht eine Gegenstelle mit Schlüssel.
package flaechen_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// dasFeld ist das Feld, um das dieser Lauf geht.
//
// **Eines und nicht alle**: die Zusage lautet „nur ausgewählte gültige
// Felder", und ein Lauf über das ganze Paket verwischte, welche Übernahme
// welches Feld geschrieben hat.
const dasFeld = "name"

// derAuftrag ist, was der Agent das Modell fragt.
const derAuftrag = "Schreibe einen neuen Namen."

// derVorschlag ist, was das Modell antwortet — [dasFeld] und sonst nichts.
const derVorschlag = `{"name":"Aus dem Modell"}`

// **Ein Prozess: lesen, vorschlagen, genau einmal übernehmen.**
func TestEinLangLaufenderAgentUebernimmtGenauEinmal(t *testing.T) {
	t.Parallel()

	umgebung, laufzeit := einAgentenbestand(t)

	setID, karteID := eineAbnahmekarte(t, umgebung)
	agent := starteMCPProzess(t, umgebung)

	// **1. Lesen.**
	vorher := feldwertMCP(t, agent.muss("lesen.karten",
		map[string]any{"setID": zahl(t, setID)}), karteID)

	if vorher == "Aus dem Modell" {
		t.Fatalf("die Karte trägt den Wert schon vor dem Lauf: %q", vorher)
	}

	// **2. Vorschlagen — und die Karte bleibt, wie sie war.**
	if text := agent.muss("schreiben.ki.vorschlag", vorschlagsruf(karteID, nil)); !strings.Contains(text, "Aus dem Modell") {
		t.Errorf("der Vorschlag nennt das Feld nicht: %s", text)
	}

	if laufzeit.auftraege() != 1 {
		t.Errorf("die Laufzeit sah %d Aufträge statt einem", laufzeit.auftraege())
	}

	istNoch(t, agent, setID, karteID, vorher, "ein Vorschlag hat geschrieben")

	// **3. Ein Feld, das der Vorschlag nicht trägt — und nichts wird
	// geschrieben.**
	//
	// **Die Auswahl nennt beide**, das gute Feld und das erfundene. Nur so
	// misst der Fall die Zusage aus `service.gewaehlteFelder`: „eine Auswahl
	// mit einem Tippfehler darin darf nicht dazu führen, dass die Felder
	// davor schon stehen." Mit `farbe` allein bliebe nichts zu schreiben
	// übrig, und der Lauf wäre auch dann grün, wenn die Prüfung fehlte.
	agent.scheitert("schreiben.ki.vorschlag",
		vorschlagsruf(karteID, []string{"name", "farbe"}), "farbe")
	istNoch(t, agent, setID, karteID, vorher, "eine abgewiesene Auswahl hat geschrieben")

	// **4. Und genau einer geht durch.**
	agent.muss("schreiben.ki.vorschlag", vorschlagsruf(karteID, []string{"name"}))

	nachher := feldwertMCP(t, agent.muss("lesen.karten",
		map[string]any{"setID": zahl(t, setID)}), karteID)
	if nachher != "Aus dem Modell" {
		t.Errorf("die Übernahme kam nicht an: %q", nachher)
	}

	// **5. Einmal, und nicht dreimal.**
	//
	// **Am Verlauf und nicht an der Karte**: drei Übernahmen desselben Werts
	// gäben denselben Wert. Der Verlauf zählt Schritte.
	if schritte := schritteAufFeld(t, agent.muss("lesen.verlauf",
		map[string]any{"setID": zahl(t, setID)})); schritte != 1 {
		t.Errorf("%d Schritte auf »name« statt einem", schritte)
	}
}

// **Ein Leser liest — und lässt kein Modell arbeiten.**
//
// # Warum der Zeiger vor dem Start umgestellt wird
//
// **Weil dieser Prozess ihn selbst nicht umstellen dürfte.** `ProfilWechseln`
// geht über MCP `OhneLeserwechsel` (`internal/mcpnaehte/nutzer.go`): eine
// Sitzung, die mitten im Betrieb in ein Leserprofil ginge, behielte ihre
// **schreibende** Verbindung, weil jede Naht sie beim Aufbau geholt hat.
// Steht der Zeiger schon dort, öffnet der Prozess gleich nur lesend — und
// **das** ist der Fall, den ein Agent trifft, der als Leser gestartet wird.
func TestEinLangLaufenderLeserBekommtKeineUebernahme(t *testing.T) {
	t.Parallel()

	umgebung, laufzeit := einAgentenbestand(t)
	setID, karteID := imFremdenProfilAlsLeser(t, umgebung, laufzeit.adresse)

	agent := starteMCPProzess(t, umgebung)

	// **Lesen geht.**
	vorher := feldwertMCP(t, agent.muss("lesen.karten",
		map[string]any{"setID": zahl(t, setID)}), karteID)

	// **Vorschlagen nicht** — ID-03 (#1702). Hier stand „Vorschlagen auch:
	// es schreibt nichts, und deshalb hält die nur lesende Verbindung es
	// nicht auf". Das war der Mechanismus und nicht die Regel: die Matrix
	// führt die Fähigkeit seit A7 unter `lokalSchreiben`, Mindestrolle
	// Mitarbeiter, weil ein Modellruf Zeit und Strom kostet. Jetzt fragt
	// jeder schreibende Ruf den Akteur — und das Modell wird gar nicht erst
	// gefragt.
	auftraege := laufzeit.gefragt.Load()

	agent.scheitert("schreiben.ki.vorschlag", vorschlagsruf(karteID, nil), "nur gelesen")

	// **Übernehmen erst recht nicht.**
	agent.scheitert("schreiben.ki.vorschlag",
		vorschlagsruf(karteID, []string{"name"}), "nur gelesen")

	if gerechnet := laufzeit.gefragt.Load() - auftraege; gerechnet != 0 {
		t.Errorf("für einen Leser rechnete das Modell %d-mal", gerechnet)
	}

	istNoch(t, agent, setID, karteID, vorher, "ein Leser hat geschrieben")
}

// vorschlagsruf baut die Argumente eines Vorschlags — mit Auswahl übernimmt er.
func vorschlagsruf(karteID int64, felder []string) map[string]any {
	ruf := map[string]any{
		"karteId": karteID, "anbieter": "ollama",
		"alias": probemodell, "auftrag": derAuftrag,
	}

	if felder != nil {
		ruf["uebernehmen"] = true
		ruf["felder"] = felder
	}

	return ruf
}

// istNoch prüft, dass ein Feld unverändert dasteht.
func istNoch(
	t *testing.T, agent *mcpprozess, setID string, karteID int64, erwartet, klage string,
) {
	t.Helper()

	jetzt := feldwertMCP(t, agent.muss("lesen.karten",
		map[string]any{"setID": zahl(t, setID)}), karteID)
	if jetzt != erwartet {
		t.Errorf("%s: %q statt %q", klage, jetzt, erwartet)
	}
}

// einAgentenbestand gibt eine Umgebung mit eingeschalteter Probelaufzeit.
func einAgentenbestand(t *testing.T) ([]string, *laufzeitstelle) {
	t.Helper()

	// **Mit eigener Saat**, denn ein Vorschlag baut sein Schema aus den
	// **tatsächlichen** Feldern des Pakets — ohne `probe.mse-game` scheitert
	// er, bevor irgendein Modell gefragt wird.
	umgebung := mitEigenerSaat(t)
	laufzeit := starteLaufzeit(t, derVorschlag)

	kiEinschalten(t, umgebung, laufzeit.adresse)

	return umgebung, laufzeit
}

// kiEinschalten stellt die Probelaufzeit im **gezeigten** Profil ein.
//
// **Im gezeigten und nicht global**: Einstellungen hängen am Profil, und
// genau das hat den ersten Anlauf des Leserfalls rot gemacht.
func kiEinschalten(t *testing.T, umgebung []string, adresse string) {
	t.Helper()

	mussKommandozeile(t, umgebung, "einstellungen", "setzen",
		domain.KIAdresse, adresse)
	mussKommandozeile(t, umgebung, "einstellungen", "setzen",
		domain.KILaufzeit, domain.KILaufzeitAn)
}

// eineAbnahmekarte fährt das Probeset ein und gibt Set und erste Karte.
func eineAbnahmekarte(t *testing.T, umgebung []string) (setID string, karteID int64) {
	t.Helper()

	setID = setkennungAus(t, mussKommandozeile(t, umgebung,
		"oeffnen", probeset(t), "--json"))
	karten := kartenzeilenAus(t, mussKommandozeile(t, umgebung,
		"karten", setID, "--json"))

	if len(karten) == 0 {
		t.Fatalf("das Probeset hat keine Karte")
	}

	return setID, zahl(t, karten[0].ID)
}

// imFremdenProfilAlsLeser baut den Fall im Profil eines anderen Nutzers auf
// und stellt den Zeiger mit Leserecht dorthin.
//
// # Warum alles dort drüben entsteht
//
// **Weil Einstellungen und Sets am Profil hängen.** Der erste Anlauf legte
// beides im eigenen an und wechselte danach: der Leser sah kein Set, und der
// Vorschlag scheiterte an „»ollama« ist nicht eingeschaltet" — an einer
// Einstellung des **anderen** Profils. Das wäre eine grüne Abweisung ohne
// Aussage über das Leserecht gewesen (§10).
func imFremdenProfilAlsLeser(t *testing.T, umgebung []string, adresse string) (
	setID string, karteID int64,
) {
	t.Helper()

	ich := eignerkennung(t, mussKommandozeile(t, umgebung, "profil", "--json"))
	fremder, fremdes := profilEinesFremden(t,
		mussKommandozeile(t, umgebung, "nutzer-neu", "Zweiter", "--json"))

	mussKommandozeile(t, umgebung, "nutzer-wechseln", fremder)

	// **Der Eigner richtet ein, was der Leser danach nur noch benutzt.**
	kiEinschalten(t, umgebung, adresse)
	setID, karteID = eineAbnahmekarte(t, umgebung)

	// **Die Freigabe vergibt der Eigner** — dafür ist man er.
	mussKommandozeile(t, umgebung, "freigabe-setzen", fremdes, ich, "leser")
	mussKommandozeile(t, umgebung, "nutzer-wechseln", ich)
	mussKommandozeile(t, umgebung, "profil", fremdes)

	return setID, karteID
}

// feldwertMCP liest den Wert eines Feldes einer Karte aus `lesen.karten`.
func feldwertMCP(t *testing.T, roh string, karteID int64) string {
	t.Helper()

	var karten []struct {
		ID     int64 `json:"ID"`
		Values []struct {
			Field string `json:"Field"`
			Text  string `json:"Text"`
		} `json:"Values"`
	}

	if fehler := json.Unmarshal([]byte(roh), &karten); fehler != nil {
		t.Fatalf("die Kartenliste lesen: %v — %s", fehler, roh)
	}

	for _, karte := range karten {
		if karte.ID != karteID {
			continue
		}

		for _, wert := range karte.Values {
			if wert.Field == dasFeld {
				return wert.Text
			}
		}

		t.Fatalf("die Karte %d hat kein Feld %q", karteID, dasFeld)
	}

	t.Fatalf("die Karte %d steht nicht in der Liste", karteID)

	return ""
}

// schritteAufFeld zählt die Verlaufsschritte, die ein Feld geändert haben.
func schritteAufFeld(t *testing.T, roh string) int {
	t.Helper()

	var antwort struct {
		Schritte []struct {
			Feld string `json:"feld"`
		} `json:"schritte"`
	}

	if fehler := json.Unmarshal([]byte(roh), &antwort); fehler != nil {
		t.Fatalf("den Verlauf lesen: %v — %s", fehler, roh)
	}

	gezaehlt := 0

	for _, schritt := range antwort.Schritte {
		if schritt.Feld == dasFeld {
			gezaehlt++
		}
	}

	return gezaehlt
}
