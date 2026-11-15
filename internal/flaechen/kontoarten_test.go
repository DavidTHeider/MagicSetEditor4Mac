// **Die vier Kontoarten — an jeder Fläche** — Werkstück 6 von Phase 4.12
// (#1351).
//
// # Die Abnahme
//
// **Eintragen, dann in einem zweiten Prozess anmelden.** Das ist der Punkt und
// nicht ein grüner Lauf: jeder Ruf hier startet ein **eigenes** Programm über
// demselben Bestand — die Anwendung schließen, öffnen, hereinkommen.
//
// **Ein Test, der beides in einem Prozess fährt, prüft einen Zwischenspeicher.**
package flaechen_test

import (
	"net/http"
	"slices"
	"strings"
	"testing"
)

// Das Kennwort der Läufe — lang genug für `kennwort.Mindestlaenge`.
const probekennwort = "einlangesgeheimnis"

// Der Bereich der Kontowerkzeuge und die beiden, die es darin gibt.
//
// **Zusammengesetzt und nicht ausgeschrieben**: ein Wächter über die
// Flächenläufe liest jede genannte Werkzeugkennung gegen den Katalog, und ein
// ausgeschriebener Name, den es nicht gibt, wäre dort ein Befund. Die Aussage
// hier ist ohnehin die über den **Bereich**.
//
// **Zwei und nicht eines**: `…loesen` steht seit Werkstück 4 da und nimmt
// eine fremde Kennung von einem Nutzer — es trägt kein Geheimnis. Wer ein
// drittes anmeldet, macht diesen Lauf rot und muss sagen, warum es kein
// Kennwort trägt.
const (
	kontobereich = "schreiben.konto."
	kontoanonym  = kontobereich + "anonym"
	kontoloesen  = kontobereich + "loesen"
)

// **Der ganze Weg über die Kommandozeile.**
//
// **Das Kennwort kommt von der Standardeingabe** und steht in keinem Argument
// — der Lauf fährt genau den Weg, den die Hilfe beschreibt.
func TestDerGanzeKontowegAnDerKommandozeile(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	eingetragen := nutzerkennung(t, mitEingabe(t, umgebung, probekennwort,
		"konto-neu", "wer@example.org", "--json"))

	herein := nutzerkennung(t, mitEingabe(t, umgebung, probekennwort,
		"konto-anmelden", "wer@example.org", "--json"))

	if herein != eingetragen {
		t.Errorf("herein kam %q, eingetragen war %q", herein, eingetragen)
	}

	// **Ein falsches Kennwort kommt nicht herein**, und die Meldung sagt
	// nicht, welche Hälfte falsch war.
	aus, fehler, code := mitEingabeUndCode(t, umgebung, "einanderesgeheimnis",
		"konto-anmelden", "wer@example.org")
	if code != 1 {
		t.Errorf("das falsche Kennwort kam durch: %d %s", code, aus)
	}

	if strings.Contains(fehler, "Kennwort stimmt") {
		t.Errorf("die Meldung verrät zu viel: %q", fehler)
	}

	// **Und ein anonymes Konto braucht gar nichts.**
	anonym := nutzerkennung(t,
		mussKommandozeile(t, umgebung, "konto-anonym", "--json"))
	if anonym == "" || anonym == eingetragen {
		t.Errorf("das anonyme Konto ist %q", anonym)
	}
}

// **Über MCP geht nur das anonyme Konto — und die anderen beiden gibt es
// dort nicht.**
//
// **Die zweite Hälfte ist die wichtigere** (§10): eine Liste dessen, was
// fehlt, liest sich wie eine Zusage, und niemand prüft nach. Hier wird
// nachgeprüft.
func TestUeberMCPGehtNurDasAnonymeKonto(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	if angelegt := nutzerkennung(t,
		mussMCP(t, umgebung, "schreiben.konto.anonym", map[string]any{})); angelegt == "" {
		t.Error("das anonyme Konto kam ohne Kennung")
	}

	// **Und sonst nichts an Konten** — gefragt wird der **gebaute** Server
	// über `tools/list` und nicht eine Liste im Quelltext.
	//
	// **Als Klasse und nicht als Aufzählung** (§10): eine Liste verbotener
	// Namen prüfte die Namen, die jemand sich ausgedacht hat, und ein
	// `schreiben.konto.eintragen` neben `…neu` käme durch. Verboten ist der
	// **Bereich**.
	angemeldet := angemeldeteWerkzeuge(t, umgebung)
	erlaubt := []string{kontoanonym, kontoloesen}

	for _, name := range angemeldet {
		if strings.HasPrefix(name, kontobereich) && !slices.Contains(erlaubt, name) {
			t.Errorf("%s ist angemeldet — im Bereich %s gehört über MCP nur %v, "+
				"weil alles andere ein Kennwort im Werkzeugaufruf trüge",
				name, kontobereich, erlaubt)
		}
	}

	// **Die Gegenprobe zur Aufzählung**: sie prüft die Liste und nicht die
	// Schreibweise. Ohne diese Zeile wäre sie auch grün, wenn `tools/list`
	// gar nichts zurückgäbe.
	if !slices.Contains(angemeldet, kontoanonym) {
		t.Fatalf("die Werkzeugliste kennt nicht einmal das anonyme Konto (%d Namen)",
			len(angemeldet))
	}
}

// **Und derselbe Weg über den Dienst.**
func TestDerGanzeKontowegUeberDenDienst(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)
	an := starteDienst(t, umgebung)

	eingetragen := nutzerkennung(t, mussREST(t, an, http.MethodPost, "/konten",
		`{"kennung":"+4915112345678","kennwort":"`+probekennwort+`"}`))

	herein := nutzerkennung(t, mussREST(t, an, http.MethodPost, "/konten/anmeldung",
		`{"kennung":"+4915112345678","kennwort":"`+probekennwort+`"}`))

	if herein != eingetragen {
		t.Errorf("herein kam %q, eingetragen war %q", herein, eingetragen)
	}

	// **Dieselbe Kennung ein zweites Mal ist ein Konflikt.**
	status, antwort := an.ruf(t, http.MethodPost, "/konten",
		`{"kennung":"+4915112345678","kennwort":"`+probekennwort+`"}`)
	if status != http.StatusConflict {
		t.Errorf("die belegte Kennung kam durch: %d (%s)", status, antwort)
	}

	// **Und das Kennwort steht in keiner Antwort** — auch nicht gehütet.
	alle := mussREST(t, an, http.MethodGet, "/nutzer", "")
	for _, verboten := range []string{"argon2", "kennwort", probekennwort} {
		if strings.Contains(alle, verboten) {
			t.Errorf("%q steht in der Nutzerliste: %s", verboten, alle)
		}
	}

	if angelegt := nutzerkennung(t,
		mussREST(t, an, http.MethodPost, "/konten/anonym", "")); angelegt == "" {
		t.Error("das anonyme Konto kam ohne Kennung")
	}
}
