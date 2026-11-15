// **Ein `leser` liest und schreibt nicht** — Werkstück 16 von Phase 4.12
// (#1351),
// nach [0073](../../docs/entscheidungen/0073-ein-leser-darf-wirklich-lesen.md).
//
// # Warum an den gebauten Programmen
//
// **Weil der Schutz an einer Verbindung hängt, die jeder Prozess selbst
// aufmacht.** Ein Lauf gegen ein Doppel bewiese die Entscheidung; hier wird
// bewiesen, dass der Bestand wirklich nichts annimmt.
package flaechen_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

// **Der ganze Weg an der Kommandozeile: hinein, lesen, nicht schreiben,
// hinaus.**
func TestEinLeserLiestUndSchreibtNichtAnDerZeile(t *testing.T) {
	t.Parallel()

	umgebung := bestandMitBund(t)

	eigenes := mussKommandozeile(t, umgebung, "profil", "--json")
	ich := eignerkennung(t, eigenes)
	meines := profilkennung(t, eigenes)

	angelegt := mussKommandozeile(t, umgebung, "nutzer-neu", "Zweiter", "--json")
	fremder, fremdes := profilEinesFremden(t, angelegt)

	// **Die Freigabe vergibt der Eigner** — dafür wird man er (Werkstück 17).
	mussKommandozeile(t, umgebung, "nutzer-wechseln", fremder)
	mussKommandozeile(t, umgebung, "freigabe-setzen", fremdes, ich, "leser")
	mussKommandozeile(t, umgebung, "nutzer-wechseln", ich)

	// **Und jetzt als Leser hinein** — das ging bis Werkstück 16 nicht.
	mussKommandozeile(t, umgebung, "profil", fremdes)

	// **Lesen geht.**
	mussKommandozeile(t, umgebung, "sets")

	// **Schreiben nicht — und die Meldung sagt, was los ist** (§8).
	aus, fehler, code := kommandozeile(t, umgebung,
		"einstellungen", "setzen", "sprache", "en")

	if code == 0 {
		t.Fatalf("der Leser schrieb eine Einstellung: %s", aus)
	}

	gesagt := aus + fehler
	if !strings.Contains(gesagt, "nur gelesen") {
		t.Errorf("die Meldung sagt nicht, was los ist: %s", gesagt)
	}

	// **Und der Weg hinaus bleibt offen** — sonst säße er fest.
	mussKommandozeile(t, umgebung, "profil", meines)

	// **Danach schreibt er wieder.** Ohne diese Zeile wäre der Lauf oben auch
	// dann grün, wenn gar nichts mehr schriebe (§10).
	mussKommandozeile(t, umgebung, "einstellungen", "setzen", "sprache", "de")
}

// **Der Wechsel in ein Leserprofil geht über MCP und den Dienst nicht — und
// die Absage sagt, warum.**
//
// # Warum das eine Zusage ist und kein Mangel
//
// **Beide halten eine Verbindung über ihre ganze Laufzeit**, und jede Naht hat
// sie sich beim Aufbau abgeholt; sie mitten im Betrieb auf eine nur lesende
// umzuhängen hieße, sie alle neu zu bauen. Derselbe Grund, aus dem dort auch
// das Zurückspielen fehlt und die Verschlüsselung nicht geschaltet wird.
//
// **Ein Lauf über eine Abwesenheit ist selten richtig.** Hier ist er es: die
// Begründung im Katalog ist eine Zusage, und ohne diesen Lauf stünde sie da,
// während ein späterer Handgriff sie stillschweigend aufhebt.
func TestDerLeserwechselGehtUeberMCPUndDienstNicht(t *testing.T) {
	t.Parallel()

	umgebung := bestandMitBund(t)

	eigenes := mussKommandozeile(t, umgebung, "profil", "--json")
	ich := eignerkennung(t, eigenes)

	angelegt := mussKommandozeile(t, umgebung, "nutzer-neu", "Zweiter", "--json")
	fremder, fremdes := profilEinesFremden(t, angelegt)

	mussKommandozeile(t, umgebung, "nutzer-wechseln", fremder)
	mussKommandozeile(t, umgebung, "freigabe-setzen", fremdes, ich, "leser")
	mussKommandozeile(t, umgebung, "nutzer-wechseln", ich)

	// **Über MCP: abgelehnt.**
	roh, istFehler := mcpRuf(t, umgebung, "schreiben.profil.wechseln",
		map[string]any{"profilID": fremdes})

	if !istFehler {
		t.Errorf("MCP wechselte in ein Leserprofil: %s", roh)
	}

	if !strings.Contains(roh, "nur lesen") {
		t.Errorf("die Absage sagt nicht, was los ist: %s", roh)
	}

	// **Über den Dienst ebenso.**
	an := starteDienst(t, umgebung)

	status, antwort := an.ruf(t, http.MethodPut, "/profil",
		`{"profilID":"`+fremdes+`"}`)

	if status < 400 {
		t.Errorf("der Dienst wechselte in ein Leserprofil: %d (%s)", status, antwort)
	}

	// **Und beide arbeiten danach weiter** — die Absage hat nichts bewegt.
	if steht := profilkennung(t,
		mussKommandozeile(t, umgebung, "profil", "--json")); steht == fremdes {
		t.Error("der Zeiger steht doch im fremden Profil")
	}
}

// **Und ein MCP-Server, der schon im Leserprofil startet, schreibt nicht.**
//
// # Warum dieser Lauf nachgetragen wurde
//
// **Weil der vorige die andere Hälfte prüfte.** Er prüft den **Wechsel** — und
// den weisen MCP und Dienst ab. Was er nicht prüft, ist der Prozess, der
// startet, während der Zeiger **schon** dort steht: dann öffnet er gleich nur
// lesend, und alle Werkzeuge müssen an dieser Verbindung hängen.
//
// **Genau dort war ein Loch.** `cmd/mse-mcp` baute seine Schreibnaht — die für
// **Kartenfelder** — weiter über die schreibende Verbindung, während der Leser
// über die zweite las. Angestrichen vom Zweitblick über Phase 4.12; der Dienst
// machte dieselbe Naht von Anfang an richtig.
//
// **Über ein Kartenfeld und nicht über eine Einstellung.** Der erste Anlauf
// nahm `schreiben.einstellung` — und blieb grün, als der Fehler wieder
// eingebaut wurde: die Einstellungsnaht hing nie an `Schreibnaht`. Eine
// Gegenprobe, die nichts ändert, ist keine (§10).
func TestEinLeserSchreibtAuchUeberMCPUndDienstNicht(t *testing.T) {
	t.Parallel()

	umgebung := saatMitBund(t)

	eigenes := mussKommandozeile(t, umgebung, "profil", "--json")
	ich := eignerkennung(t, eigenes)
	meines := profilkennung(t, eigenes)

	angelegt := mussKommandozeile(t, umgebung, "nutzer-neu", "Zweiter", "--json")
	fremder, fremdes := profilEinesFremden(t, angelegt)

	// **Das Set gehört dem Fremden** — sonst sähe der Leser es gar nicht, und
	// die Absage käme von „gibt es nicht" statt von der Verbindung.
	mussKommandozeile(t, umgebung, "nutzer-wechseln", fremder)

	setID := leeresSet(t, umgebung, "Leseprobe")
	mussKommandozeile(t, umgebung, "karte", setID, "anlegen")

	karteID := ersteKarte(t, mussKommandozeile(t, umgebung, "karten", setID, "--json"))

	mussKommandozeile(t, umgebung, "freigabe-setzen", fremdes, ich, "leser")
	mussKommandozeile(t, umgebung, "nutzer-wechseln", ich)

	// **Der Zeiger steht dort, bevor die beiden Prozesse aufmachen** — an der
	// Kommandozeile, denn nur sie kann wechseln.
	mussKommandozeile(t, umgebung, "profil", fremdes)

	// **Über MCP: lesen ja, schreiben nein.**
	gelesen, lesefehler := mcpRuf(t, umgebung, "lesen.karten",
		map[string]any{"setID": zahl(t, setID)})
	if lesefehler {
		t.Fatalf("der Leser konnte über MCP nicht lesen: %s", gelesen)
	}

	roh, istFehler := mcpRuf(t, umgebung, "schreiben.feld", map[string]any{
		"setID": zahl(t, setID), "karteID": zahl(t, karteID),
		"feld": "name", "wert": "Fremd",
	})

	if !istFehler {
		t.Errorf("der Leser schrieb über MCP ein Kartenfeld: %s", roh)
	}

	if !strings.Contains(roh, "nur gelesen") {
		t.Errorf("die Absage sagt nicht, was los ist: %s", roh)
	}

	// **Und ein Zusatzwert ebenso** (#2298): er geht durch denselben
	// Schreibweg wie ein Kartenfeld, also muss er auch an derselben Rolle
	// scheitern.
	if zusatz, abgewiesen := mcpRuf(t, umgebung, "schreiben.zusatzwert", map[string]any{
		"setID": zahl(t, setID), "karteID": zahl(t, karteID),
		"feld": "vermerk", "wert": "Fremd",
	}); !abgewiesen || !strings.Contains(zusatz, "nur gelesen") {
		t.Errorf("der Leser schrieb über MCP einen Zusatzwert: %s", zusatz)
	}

	// **Über den Dienst ebenso.**
	an := starteDienst(t, umgebung)

	status, antwort := an.ruf(t, http.MethodPut,
		"/sets/"+setID+"/karten/"+karteID+"/felder/name", `{"wert":"Fremd"}`)

	if status < 400 {
		t.Errorf("der Leser schrieb über den Dienst: %d (%s)", status, antwort)
	}

	if status, antwort := an.ruf(t, http.MethodPut,
		"/sets/"+setID+"/karten/"+karteID+"/zusatzfelder/vermerk", `{"wert":"Fremd"}`); status < 400 {
		t.Errorf("der Leser schrieb über den Dienst einen Zusatzwert: %d (%s)", status, antwort)
	}

	// **Und auch nicht über die Nutzerverwaltung** — der Nachlauf zum
	// Zweitblick über Phase 4.12.
	//
	// **Die benannte Ausnahme war zu breit**: Token, Vorgaben und die
	// Sicherung hingen an der schreibenden Verbindung, weil sie mit den
	// Nutzern zusammen verdrahtet werden. Der Weg hinaus schreibt aber den
	// **Zeiger** — ein Token schreibt in ein **Profil**, und es trägt Rechte.
	vorgabe, vorgabefehler := mcpRuf(t, umgebung, "schreiben.profilvorgabe",
		map[string]any{
			"profilID": fremdes, "art": "format", "werte": []string{},
		})

	if !vorgabefehler {
		t.Errorf("der Leser setzte über MCP eine Vorgabe: %s", vorgabe)
	}

	marke, markenfehler := mcpRuf(t, umgebung, "schreiben.token.anlegen",
		map[string]any{
			"profilID": fremdes, "bezeichnung": "Probe", "principal": "Rechner", "scope": "lesen",
		})

	if !markenfehler {
		t.Errorf("der Leser legte über MCP ein Token an: %s", marke)
	}

	// **Und der Weg hinaus bleibt offen** — sonst säße er fest.
	mussKommandozeile(t, umgebung, "profil", meines)
}

// saatMitBund gibt eine Umgebung mit der Saat **und** einem eigenen
// Schlüsselbund.
//
// **Beides**, weil dieser Lauf ein Set anlegt (das braucht einen Stil aus der
// Saat) und Nutzer wechselt (das fasst den Bund an). Ohne den Bund griffe er
// in den echten Schlüsselbund dieses Rechners (§3).
func saatMitBund(t *testing.T) []string {
	t.Helper()

	return append(mitEigenerSaat(t), schluesselbund.Bundort+"="+probebund(t))
}
