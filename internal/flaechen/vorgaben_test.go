// **Vorgaben, die auf Zeilen zeigen — an jeder Fläche** — Werkstück 15 von
// Phase 4.12
// (#1351),
// Punkte 4 bis 7.
//
// # Der Ablauf
//
// Alle fünf Arten lesen → eine setzen → wiederlesen → abräumen. **Und zweimal
// dieselbe Ablehnung**: ein Wert, den es nicht gibt, und eine Art, die es
// nicht gibt.
//
// # Was diese Läufe belegen und die der Schichten nicht
//
// **Dass das Angebot aus dem echten Bestand kommt.** In `internal/service`
// ist es eine Schnittstelle, und der Lauf erfindet, was sie sagt; hier
// entscheidet, was wirklich installiert ist — und `mse paket-anlegen` legt es
// an.
package flaechen_test

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// **Der ganze Weg über die Kommandozeile.**
func TestDerGanzeVorgabenwegAnDerKommandozeile(t *testing.T) {
	t.Parallel()

	umgebung := probespiel(t)

	alle := mussKommandozeile(t, umgebung, "vorgaben", "1")
	for _, art := range []string{"format", "spiel", "vorlage", "deck", "lagerort"} {
		if !strings.Contains(alle, art) {
			t.Errorf("die Art %q fehlt in %q", art, alle)
		}
	}

	// **Das Spiel gibt es**, weil `eigenerBestand` die Saat einbindet.
	mitEingabe(t, umgebung, `["probe"]`, "vorgabe-setzen", "1", "spiel")

	if nachher := mussKommandozeile(t, umgebung, "vorgaben", "1"); !strings.Contains(
		nachher, "probe") {
		t.Errorf("die gesetzte Vorgabe fehlt in %q", nachher)
	}

	mitEingabe(t, umgebung, `[]`, "vorgabe-setzen", "1", "spiel")

	if geraeumt := mussKommandozeile(t, umgebung, "vorgaben", "1"); strings.Contains(
		geraeumt, "probe") {
		t.Errorf("die abgeräumte Vorgabe steht noch da: %q", geraeumt)
	}

	// **Und was es nicht gibt, wird abgewiesen** — gegen dasselbe Angebot,
	// aus dem die Auswahl gebaut wird.
	if _, _, code := mitEingabeUndCode(t, umgebung, `["kartenrücken"]`,
		"vorgabe-setzen", "1", "spiel"); code == 0 {
		t.Error("ein Spiel, das es nicht gibt, kam durch")
	}
}

// **Derselbe Weg über MCP.**
func TestDerGanzeVorgabenwegUeberMCP(t *testing.T) {
	t.Parallel()

	umgebung := probespiel(t)

	alle := mussMCP(t, umgebung, "lesen.profilvorgaben",
		map[string]any{"profilID": "1"})
	if !strings.Contains(alle, `"moeglich"`) {
		t.Errorf("das Angebot fehlt in %s", alle)
	}

	mussMCP(t, umgebung, "schreiben.profilvorgabe", map[string]any{
		"profilID": "1", "art": "spiel", "werte": []string{"probe"},
	})

	nachher := mussMCP(t, umgebung, "lesen.profilvorgaben",
		map[string]any{"profilID": "1"})
	if !strings.Contains(nachher, "probe") {
		t.Errorf("die gesetzte Vorgabe fehlt in %s", nachher)
	}

	// **Und eine Art, die es nicht gibt, ist ein Nein.**
	roh, istFehler := mcpRuf(t, umgebung, "schreiben.profilvorgabe",
		map[string]any{
			"profilID": "1", "art": "kartenrücken", "werte": []string{"1"},
		})
	if !istFehler {
		t.Errorf("die erfundene Art kam durch: %s", roh)
	}
}

// **Und derselbe Weg über den Dienst.**
func TestDerGanzeVorgabenwegUeberDenDienst(t *testing.T) {
	t.Parallel()

	umgebung := probespiel(t)
	an := starteDienst(t, umgebung)

	alle := mussREST(t, an, http.MethodGet, "/profile/1/vorgaben", "")
	if !strings.Contains(alle, `"lagerort"`) {
		t.Errorf("die Art `lagerort` fehlt in %s", alle)
	}

	mussREST(t, an, http.MethodPut, "/profile/1/vorgaben/spiel",
		`{"werte":["probe"]}`)

	nachher := mussREST(t, an, http.MethodGet, "/profile/1/vorgaben", "")
	if !strings.Contains(nachher, "probe") {
		t.Errorf("die gesetzte Vorgabe fehlt in %s", nachher)
	}

	// **Ein `PUT` ohne `werte` räumt ab** — es als Auslassung zu lesen ließe
	// einen Ruf ohne Wirkung zu, und der ist bei einem Ort, der die Werte
	// **ist**, keine sinnvolle Anfrage.
	mussREST(t, an, http.MethodPut, "/profile/1/vorgaben/spiel", `{}`)

	geraeumt := mussREST(t, an, http.MethodGet, "/profile/1/vorgaben", "")
	if strings.Contains(geraeumt, `"werte":["probe"]`) {
		t.Errorf("die abgeräumte Vorgabe steht noch da: %s", geraeumt)
	}

	// **Und ein Profil, das es nicht gibt, ist kein leeres** (§8).
	if status, _ := an.ruf(
		t, http.MethodGet, "/profile/4711/vorgaben", ""); status < 400 {
		t.Errorf("ein Profil, das es nicht gibt, gab %d", status)
	}
}

// **Und die drei Flächen sagen dasselbe.**
//
// **Das ist die Zusage aus §16**, und sie ist an keiner einzelnen Fläche zu
// prüfen: dass alle drei dieselben fünf Arten mit demselben Angebot nennen,
// sieht nur ein Lauf, der alle drei fragt.
func TestDieDreiFlaechenNennenDieselbenVorgaben(t *testing.T) {
	t.Parallel()

	umgebung := probespiel(t)

	ausMCP := mussMCP(t, umgebung, "lesen.profilvorgaben",
		map[string]any{"profilID": "1"})

	an := starteDienst(t, umgebung)
	ausREST := mussREST(t, an, http.MethodGet, "/profile/1/vorgaben", "")

	ausCLI := mussKommandozeile(t, umgebung, "vorgaben", "1", "--json")

	for _, art := range []string{"format", "spiel", "vorlage", "deck", "lagerort"} {
		for name, ausgabe := range map[string]string{
			"MCP": ausMCP, "REST": ausREST, "CLI": ausCLI,
		} {
			if !strings.Contains(ausgabe, `"`+art+`"`) {
				t.Errorf("%s nennt die Art %q nicht: %s", name, art, ausgabe)
			}
		}
	}
}

// probespiel gibt eine Umgebung mit **eigenem Zuhause** und einem Spielpaket
// darin.
//
// **Eigenes Zuhause und nicht das dieses Rechners** (§3): das Angebot kommt
// aus dem Paketkatalog, und der liest die Platte. Ein Lauf gegen das echte
// Zuhause prüfte, was dort zufällig installiert ist — und wäre grün, **weil**
// er funktioniert hat.
func probespiel(t *testing.T) []string {
	t.Helper()

	umgebung := eigenesZuhause(t)

	zuhause := ""

	for _, eintrag := range umgebung {
		if wert, gefunden := strings.CutPrefix(eintrag, "HOME="); gefunden {
			zuhause = wert
		}
	}

	// **Derselbe Weg wie `app.Paketverzeichnis()`**, von Hand nachgebaut — wie
	// bei `probepaket`: die Flächen laufen als eigene Prozesse, und dieser
	// Lauf kann ihnen den Ort nicht abfragen.
	ort := filepath.Join(zuhause, "Library", "Application Support",
		"MagicSetEditor4Mac", "pakete", "probe.mse-game")

	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(ort, "game"),
		[]byte("mse version: 2.0.0\nshort name: Probe\n"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return umgebung
}
