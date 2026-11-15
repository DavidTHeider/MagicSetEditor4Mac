package mtgjson_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// **Ein unbekanntes Feld gibt einen Befund je Set, nicht einen je Karte.**
//
// # Woher diese Zusage kommt
//
// Aus dem Lauf gegen bezogene Daten (Build-Tag `fremdpakete`). Eine echte
// MTGJSON-Datei — *Limited Edition Alpha*, 295 Karten — führt **26** Felder, die
// die Zuordnung nicht kennt, und fast jede Karte führt fast alle: `uuid`,
// `identifiers`, `legalities`, `purchaseUrls`, `foreignData` und so fort. Je
// Karte gemeldet ergab das **7.669 Befunde** für eine Datei, die vollständig
// gelesen wurde.
//
// Das ist keine Auskunft mehr, sondern ihr Gegenteil: wer 7.669 Zeilen sieht,
// liest die eine nicht, die von einem wirklichen Verlust handelt. Gebündelt sind
// es 26 — und jede sagt, wie viele Karten sie betrifft.
//
// **Der Wert bleibt trotzdem an jeder Karte.** Gebündelt wird die *Meldung*,
// nicht das Mitführen.
func TestEinUnbekanntesFeldGibtEinenBefundJeSet(t *testing.T) {
	quelle, err := mtgjson.Lies(pfad("probe.json"), mtgjson.ZuordnungMagic)
	if err != nil {
		t.Fatalf("Lies: %v", err)
	}

	if len(quelle.Befunde) != 1 {
		t.Fatalf("Befunde: %v, erwartet genau einen für »uuid«", quelle.Befunde)
	}

	befund := quelle.Befunde[0]

	if befund.Ort != "PRB → uuid" {
		t.Errorf("Ort: %q, erwartet %q", befund.Ort, "PRB → uuid")
	}

	// Die Zahl gehört dazu: „ein Feld fehlt" und „ein Feld fehlt an allen 295
	// Karten" sind zwei verschiedene Auskünfte.
	if !strings.Contains(befund.Meldung, "2 Karten") {
		t.Errorf("Meldung: %q — sie sagt nicht, wie viele Karten es betrifft", befund.Meldung)
	}

	// Und der Wert steht weiterhin an **jeder** Karte.
	for nummer, karte := range quelle.Sets[0].Karten {
		if _, da := wertVon(karte.Rest, "uuid"); !da {
			t.Errorf("Karte %d führt »uuid« nicht mehr mit", nummer+1)
		}
	}
}

// Zwei Sets in einer Datei werden getrennt gemeldet.
//
// Sonst stünde eine Zahl da, die über Sets summiert — und niemand wüsste, ob
// das Feld in einem Set fehlt oder in allen.
func TestJedesSetBekommtSeinenEigenenBefund(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "zwei.json")

	inhalt := `{"data": {
	  "AAA": {"code":"AAA","cards":[{"name":"Eine","uuid":"1"}]},
	  "BBB": {"code":"BBB","cards":[{"name":"Zwei","uuid":"2"}]}
	}}`

	if err := os.WriteFile(ort, []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	quelle, err := mtgjson.Lies(ort, mtgjson.ZuordnungMagic)
	if err != nil {
		t.Fatalf("Lies: %v", err)
	}

	if len(quelle.Befunde) != 2 {
		t.Fatalf("Befunde: %v, erwartet einen je Set", quelle.Befunde)
	}

	for i, erwartet := range []string{"AAA → uuid", "BBB → uuid"} {
		if quelle.Befunde[i].Ort != erwartet {
			t.Errorf("Befund %d: Ort %q, erwartet %q", i, quelle.Befunde[i].Ort, erwartet)
		}
	}
}

// **Eine Namenskollision bleibt je Karte.**
//
// Sie trägt den Wert in ihrer Meldung — das ist der Grund, warum sie überhaupt
// eine ist. Sie zu bündeln hieße, alle Werte bis auf einen wegzuwerfen, und
// dann wäre die Meldung wieder eine Nachricht über etwas, das fort ist.
func TestEineNamenskollisionBleibtJeKarte(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "kollisionen.json")

	inhalt := `{"code":"KOL","cards":[
	  {"name":"Eine","text":"gedeutet","rule text":"erste"},
	  {"name":"Zwei","text":"gedeutet","rule text":"zweite"}
	]}`

	if err := os.WriteFile(ort, []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	quelle, err := mtgjson.Lies(ort, mtgjson.ZuordnungMagic)
	if err != nil {
		t.Fatalf("Lies: %v", err)
	}

	if len(quelle.Befunde) != 2 {
		t.Fatalf("Befunde: %v, erwartet einen je Karte", quelle.Befunde)
	}

	for _, wert := range []string{"erste", "zweite"} {
		gefunden := false

		for _, befund := range quelle.Befunde {
			if strings.Contains(befund.Meldung, wert) {
				gefunden = true
			}
		}

		if !gefunden {
			t.Errorf("der Wert %q steht in keiner Meldung und ist damit verloren", wert)
		}
	}
}
