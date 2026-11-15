package flaechen_test

// **Ein Feld mit Wertskript speichert über jede Fläche das Gerechnete** —
// #2396 (Issue (#2396)).
//
// MSE rechnet nach einer Eingabe das Wertskript und zieht nach, wer das Feld
// liest (`SetScriptManager::updateValue`, `src/script/script_manager.cpp`,
// Zeilen 249–262, Fassung f9d9356 vom 2021-07-06). **Dieselbe Fähigkeit,
// überall dieselbe Antwort** (§16): die Kommandozeile und der Dienst setzen
// dasselbe Feld an zwei Karten, und in beiden steht danach dasselbe.
//
// **Brechend**, entschieden mit 0.9.0 (0033, Regel 7): wer ein Feld setzt,
// findet den gerechneten Wert und nicht den übergebenen.

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// mitWertskripten hängt an die Kopie des Probespiels zwei Felder: `zeile`
// klammert, `echo` liest `zeile`.
func mitWertskripten(t *testing.T, umgebung []string) {
	t.Helper()

	spiel := filepath.Join(zuhauseAus(t, umgebung), "Library", "Application Support",
		"MagicSetEditor4Mac", "pakete", "probe.mse-game", "game")

	roh, fehler := os.ReadFile(spiel) //nolint:gosec // die Kopie dieses Laufs
	if fehler != nil {
		t.Fatalf("Spiel lesen: %v", fehler)
	}

	const felder = "card field:\n\ttype: text\n\tname: zeile\n\tscript: \"[\" + value + \"]\"\n" +
		"card field:\n\ttype: text\n\tname: echo\n\tscript: card.zeile + \"!\"\n\n"

	neu := strings.Replace(string(roh), "set field:", felder+"set field:", 1)
	//nolint:gosec // G703: die Kopie dieses Laufs
	if fehler := os.WriteFile(spiel, []byte(neu), 0o600); fehler != nil {
		t.Fatalf("Spiel schreiben: %v", fehler)
	}
}

// gespeichert liest die Werte einer Karte aus dem Bestand.
func gespeichert(t *testing.T, umgebung []string, karteID string) map[string]string {
	t.Helper()

	db, fehler := storage.Oeffnen(t.Context(), bestandsortAus(t, umgebung))
	if fehler != nil {
		t.Fatalf("den Bestand öffnen: %v", fehler)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur zum Aufräumen

	ctx, profilfehler := storage.MitGezeigtemProfil(t.Context(), db)
	if profilfehler != nil {
		t.Fatalf("das Profil lesen: %v", profilfehler)
	}

	kennung, _ := strconv.ParseInt(karteID, 10, 64) //nolint:errcheck // die Kennung kam aus der Liste

	karte, fehler := storage.NeuerSetLeser(db).KarteLesen(ctx, kennung)
	if fehler != nil || karte == nil {
		t.Fatalf("die Karte lesen: %v", fehler)
	}

	werte := make(map[string]string, len(karte.Values))
	for _, wert := range karte.Values {
		werte[wert.Field] = wert.Text
	}

	return werte
}

func TestEinFeldMitWertskriptSpeichertUeberJedeFlaecheDasGerechnete(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	mitWertskripten(t, umgebung)
	an := starteDienst(t, umgebung)

	setID := neuesSet(t, umgebung, "Nachzug")
	mussKommandozeile(t, umgebung, "karte", setID, "anlegen")
	mussKommandozeile(t, umgebung, "karte", setID, "anlegen")

	karten := kartenkennungenAus(t, mussKommandozeile(t, umgebung, "karten", setID, "--json"))
	if len(karten) != 2 {
		t.Fatalf("%d Karten, erwartet zwei", len(karten))
	}

	mussKommandozeile(t, umgebung, "feld", setID, karten[0], "zeile", "Funke")

	if status, rumpf := an.ruf(t, http.MethodPut,
		"/sets/"+setID+"/karten/"+karten[1]+"/felder/zeile", `{"wert":"Funke"}`); status != http.StatusOK {
		t.Fatalf("über den Dienst: %d (%s)", status, rumpf)
	}

	for flaeche, karteID := range map[string]string{"Kommandozeile": karten[0], "Dienst": karten[1]} {
		werte := gespeichert(t, umgebung, karteID)
		if werte["zeile"] != "[Funke]" || werte["echo"] != "[Funke]!" {
			t.Errorf("über %s steht zeile=%q, echo=%q — erwartet [Funke] und [Funke]!",
				flaeche, werte["zeile"], werte["echo"])
		}
	}
}

// kartenkennungenAus liest alle Kartenkennungen aus der Antwort von
// `mse karten`.
func kartenkennungenAus(t *testing.T, roh string) []string {
	t.Helper()

	var antwort struct {
		Karten []struct {
			ID string `json:"id"`
		} `json:"karten"`
	}

	if fehler := json.Unmarshal([]byte(roh), &antwort); fehler != nil {
		t.Fatalf("kein JSON: %v\n%s", fehler, roh)
	}

	kennungen := make([]string, 0, len(antwort.Karten))
	for _, karte := range antwort.Karten {
		kennungen = append(kennungen, karte.ID)
	}

	return kennungen
}
