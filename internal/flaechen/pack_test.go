// Das Zufallspack über jede Fläche
// (#889,
// Befund 7 des Zweitblicks über Phase W).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
package flaechen_test

import (
	"net/http"
	"strings"
	"testing"
)

// **Derselbe Same gibt an jeder Fläche ohne Fenster dieselbe Ziehung.**
//
// Das ist der schärfere Teil von §16: der Wächter über `FAEHIGKEITEN.md`
// sieht, **dass** eine Fähigkeit an vier Flächen steht — nicht, dass sie
// dasselbe tut. Hier wird es gefahren.
//
// **Und der Same ist es, der die Prüfung überhaupt möglich macht.** Ohne ihn
// wäre jede Ziehung einmalig, und drei Antworten zu vergleichen hieße, drei
// Zufälle zu vergleichen.
//
// # Was er nicht hält
//
// **Das Fenster.** Er fährt Kommandozeile, MCP und REST, das Fenster nicht
// (gemessen in `FLAECHENPARITAET.md`); die Fensterseite prüft sie gegen ihre eigene
// Vorrichtung (`zufallspack.test.tsx`). Der Name hieß bis zum 2026-08-25
// „über jede Fläche" und behauptete damit vier — angestrichen vom Nachlauf
// des Zweitblicks über Phase W.
//
// **Was ihn trotzdem trägt**: alle drei gehen über `app.PackUeber`, und die
// Bindung des Fensters über denselben `service.Packbereich`. Die Naht ist
// dieselbe; ungeprüft bleibt allein die Wails-Übersetzung darüber.
func TestDasZufallspackUeberJedeFlaecheOhneFenster(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := setImProbebestand(t, umgebung)

	const art = "alles"

	// **Die Packarten zuerst** — ohne sie wüsste niemand, was `art` sein darf.
	t.Run("die Packarten stehen an jeder Fläche", func(t *testing.T) {
		aus := mussKommandozeile(t, umgebung, "packarten", setID)

		text, istFehler := mcpRuf(t, umgebung, "lesen.packarten",
			map[string]any{"setID": zahl(t, setID)})
		if istFehler {
			t.Fatalf("lesen.packarten: %s", text)
		}

		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodGet, "/sets/"+setID+"/packarten", "")
		if status != http.StatusOK {
			t.Fatalf("GET packarten: %d (%s)", status, rumpf)
		}

		// **Die Saat trägt die Packart, also muss sie überall stehen.**
		//
		// Der erste Anlauf ließ eine leere Antwort an allen dreien als „einig"
		// durchgehen — und wäre damit auch grün gewesen, wenn keine Fläche je
		// eine Packart gefunden hätte. Angestrichen vom Nachlauf des
		// Zweitblicks über Phase W: „alle sagen dasselbe" ist keine Zusage,
		// wenn das Gemeinsame das Nichts sein darf.
		for name, antwort := range map[string]string{
			"CLI": aus, "MCP": text, "REST": rumpf,
		} {
			if !strings.Contains(antwort, art) {
				t.Errorf("%s kennt die Packart %q nicht:\n%s", name, art, antwort)
			}
		}
	})

	t.Run("und dieselbe Ziehung", func(t *testing.T) {
		const same = "4242"

		ausCLI := mussKommandozeile(t, umgebung, "pack", setID, art, same)

		ausMCP, istFehler := mcpRuf(t, umgebung, "lesen.pack.ziehen",
			map[string]any{"setID": zahl(t, setID), "art": art, "same": same})
		if istFehler {
			t.Fatalf("lesen.pack.ziehen: %s", ausMCP)
		}

		an := starteDienst(t, umgebung)

		status, ausREST := an.ruf(t, http.MethodPost,
			"/sets/"+setID+"/packarten/"+art+"/ziehung?same="+same, "")
		if status != http.StatusOK {
			t.Fatalf("POST ziehung: %d (%s)", status, ausREST)
		}

		// **Verglichen werden die Kennungen in ihrer Reihenfolge** — nicht die
		// Gestalt der Antwort. Jede Fläche hat ihre eigene: die Kommandozeile
		// schreibt Spalten, MCP und REST je ein eigenes JSON.
		cli := kennungenAus(ausCLI)
		mcp := kennungenAusJSON(ausMCP, `"ID"`)
		rest := kennungenAusJSON(ausREST, `"id"`)

		if len(cli) == 0 {
			t.Fatalf("die Kommandozeile zog nichts:\n%s", ausCLI)
		}

		if strings.Join(cli, ",") != strings.Join(mcp, ",") {
			t.Errorf("CLI zog %v, MCP %v — roh:\n%s", cli, mcp, ausMCP)
		}

		if strings.Join(cli, ",") != strings.Join(rest, ",") {
			t.Errorf("CLI zog %v, REST %v", cli, rest)
		}
	})
}

// kennungenAus liest die erste Spalte jeder Zeile.
func kennungenAus(aus string) []string {
	var gefunden []string

	for _, zeile := range strings.Split(strings.TrimSpace(aus), "\n") {
		if vorn, _, gut := strings.Cut(zeile, "\t"); gut && vorn != "" {
			gefunden = append(gefunden, vorn)
		}
	}

	return gefunden
}

// kennungenAusJSON liest jede Kennung hinter einer Marke.
//
// **Von Hand und nicht über einen Verbund**: die beiden Flächen schreiben
// verschiedene Marken, und ein Verbund je Fläche prüfte am Ende die Verbünde
// dieses Tests statt die Antworten.
//
// **Der Doppelpunkt und sein Abstand bleiben draußen**: MCP schreibt eingerückt
// (`"ID": "2"`), der Dienst gedrängt (`"id":"2"`). Wer die Marke mit
// Anführungszeichen schriebe, prüfte die Formatierung statt der Kennung.
func kennungenAusJSON(roh, marke string) []string {
	var gefunden []string

	rest := roh
	for {
		_, nach, gut := strings.Cut(rest, marke)
		if !gut {
			return gefunden
		}

		_, wert, offen := strings.Cut(nach, `"`)
		if !offen {
			return gefunden
		}

		kennung, weiter, zu := strings.Cut(wert, `"`)
		if !zu {
			return gefunden
		}

		gefunden = append(gefunden, kennung)
		rest = weiter
	}
}
