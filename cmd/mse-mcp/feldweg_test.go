package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Der Server schreibt über dieselbe fachliche Naht wie das Fenster** —
// #710.
//
// # Der Befund
//
// Aus dem Zweitblick durch Codex vom 2026-08-21: das Fenster setzt seit
// #696 den
// `Feldweg` ein — ein berechnetes Feld geht in die Felder, aus denen es
// entsteht. Der **zweite** Composition Root verdrahtete weiterhin
// `storage.NeuerKartenschreiber` unmittelbar.
//
// Folge: über MCP entstand ein roher `text`-Wert, den das Paket nie liest.
// Derselbe Aufruf, zwei Ergebnisse — und der Kommentar im `Feldweg` nannte
// ausgerechnet den MCP-Weg als Grund für seine Meldung.
func TestDerServerSchreibtUeberDenFeldweg(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	pfad, err := app.Datenbankpfad()
	if err != nil {
		t.Fatalf("Pfad: %v", err)
	}

	if fehler := os.MkdirAll(filepath.Dir(pfad), 0o700); fehler != nil {
		t.Fatalf("Verzeichnis: %v", fehler)
	}

	// **Ein Spielpaket, das `text` berechnet** — wie `magic.mse-game`.
	pakete, err := app.Paketverzeichnis()
	if err != nil {
		t.Fatalf("Pakete: %v", err)
	}

	spielpaket(t, pakete)

	db, err := storage.Oeffnen(probekontext(t), "file:"+pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "probe",
		Cards: []*domain.Card{{Values: []domain.Value{{Field: "name", Text: "Feuerball"}}}},
	}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	karten, err := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	aendern := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":` +
		`{"name":"schreiben.feld","arguments":{"setID":` +
		strconv.FormatInt(setID, 10) +
		`,"karteID":` +
		strconv.FormatInt(karten[0].ID, 10) +
		`,"feld":"text","wert":"Regel<sep><line>\n</line></sep>Flavor"}}}`

	sitzung := `{"jsonrpc":"2.0","id":1,"method":"initialize"}` + "\n" + aendern + "\n"

	if fehler := lauf(probekontext(t), nil, strings.NewReader(sitzung), &strings.Builder{}); fehler != nil {
		t.Fatalf("lauf: %v", fehler)
	}

	nachher, fehler := storage.Oeffnen(probekontext(t), "file:"+pfad)
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = nachher.Close() })

	geaendert, fehler := storage.NeuerSetLeser(nachher).Karten(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	werte := map[string]string{}
	for _, wert := range geaendert[0].Values {
		werte[wert.Field] = wert.Text
	}

	if werte["rule text"] != "Regel" || werte["flavor text"] != "Flavor" {
		t.Errorf("Werte = %v", werte)
	}

	// **Und kein roher `text`-Wert daneben.** Er wäre einer, den das Paket
	// nie liest — und der beim nächsten Öffnen still verschwindet.
	if _, da := werte["text"]; da {
		t.Errorf("`text` steht im Bestand: %v", werte)
	}
}

// spielpaket legt ein Spiel an, dessen `text` aus zwei Feldern entsteht.
func spielpaket(t *testing.T, wurzel string) {
	t.Helper()

	ort := filepath.Join(wurzel, "probe.mse-game")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := `mse version: 2.0.0
short name: probe
card field:
	type: text
	name: text
	save value: false
	script:
		combined_editor(field1: card.rule_text, separator: "<line>\n</line>", field2: card.flavor_text)
card field:
	type: text
	name: rule text
card field:
	type: text
	name: flavor text
`

	if err := os.WriteFile(filepath.Join(ort, "game"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}
}
