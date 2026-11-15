// **Karten aus einem Skript — an jeder Fläche** — Werkstück B von Phase 4.11
// (#1337,
// #1349).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
//
// # Was hier geprüft wird
//
// Die Saat `probe.mse-game` bringt seit #1337 zwei `add cards script:`-Blöcke
// mit — dieselbe Gestalt wie `magic.mse-game/add_cards_scripts`, eigener
// Inhalt (§11). Der eine legt zwei Karten an, der andere nichts.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// skriptumgebung gibt ein eigenes Zuhause samt der Saat als Spielpaket.
//
// # Warum die Saat kopiert wird
//
// `app.Saatverzeichnis()` rechnet vom Programmpfad auf `…/Contents/Resources/saat`
// — im `.app`-Bündel. Unter `go test` gibt es das nicht, und die Fläche fände
// `probe.mse-game` nirgends: der erste Lauf gab **null Skripte an allen drei
// Flächen**, und `mse felder` gab eine leere Liste dazu.
//
// **Und ein eigenes Zuhause dazu** (§3): `app.Paketverzeichnis()` hängt an
// `HOME`. Ohne diese Zeile legte der Lauf das Paket im **echten**
// Paketverzeichnis dieses Rechners an — derselbe Fall, den
// `paketwerkstatt_test.go` beschreibt.
func skriptumgebung(t *testing.T) []string {
	t.Helper()

	umgebung := eigenesZuhause(t)

	zuhause := ""

	for _, eintrag := range umgebung {
		if wert, gefunden := strings.CutPrefix(eintrag, "HOME="); gefunden {
			zuhause = wert
		}
	}

	ziel := filepath.Join(zuhause, "Library", "Application Support",
		"MagicSetEditor4Mac", "pakete")

	for _, paket := range []string{"probe.mse-game", "probe-schlicht.mse-style"} {
		kopiereVerzeichnis(t, filepath.Join("..", "..", "saat", paket),
			filepath.Join(ziel, paket))
	}

	return umgebung
}

// kopiereVerzeichnis legt ein Paket flach an seinem Ziel ab.
//
// **Flach genügt**: die Saatpakete haben keine Unterverzeichnisse, und ein
// rekursiver Kopierer wäre Code, den kein Fall braucht.
func kopiereVerzeichnis(t *testing.T, quelle, ziel string) {
	t.Helper()

	if err := os.MkdirAll(ziel, 0o750); err != nil {
		t.Fatalf("%s anlegen: %v", ziel, err)
	}

	eintraege, err := os.ReadDir(quelle)
	if err != nil {
		t.Fatalf("%s lesen: %v", quelle, err)
	}

	for _, eintrag := range eintraege {
		if eintrag.IsDir() {
			continue
		}

		inhalt, lesefehler := os.ReadFile(filepath.Join(quelle, eintrag.Name())) //nolint:gosec // ein Pfad aus dem Repositorium
		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", eintrag.Name(), lesefehler)
		}

		// **Der Name kommt aus `os.ReadDir` über einem Pfad im
		// Repositorium** — er ist ein einzelner Eintrag ohne Trenner, und der
		// Quellordner ist die Saat.
		//nolint:gosec // G703: der Grund steht darüber
		if err := os.WriteFile(filepath.Join(ziel, eintrag.Name()),
			inhalt, 0o600); err != nil {
			t.Fatalf("%s schreiben: %v", eintrag.Name(), err)
		}
	}
}

// **Jede Fläche nennt dieselben Skripte.**
func TestDieKartenskripteStehenUeberallGleich(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Skriptset")

		pruefeSkripte(t, mussKommandozeile(t, umgebung,
			"kartenskripte", setID, "--json"))
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Skriptset")

		roh, istFehler := mcpRuf(t, umgebung, "lesen.kartenskripte",
			map[string]any{"setID": zahl(t, setID)})
		if istFehler {
			t.Fatalf("lesen.kartenskripte: %s", roh)
		}

		pruefeSkripte(t, roh)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Skriptset")

		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodGet, "/sets/"+setID+"/kartenskripte", "")
		if status != http.StatusOK {
			t.Fatalf("GET /sets/%s/kartenskripte: %d (%s)", setID, status, rumpf)
		}

		pruefeSkripte(t, rumpf)
	})
}

// **Und jede Fläche legt dieselben Karten an — in einem Zug.**
func TestEinKartenskriptLaeuftUeberallGleich(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Skriptset")

		mussKommandozeile(t, umgebung, "kartenskript", setID, "Zwei Farben")

		pruefeZweiKarten(t, umgebung, setID)
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Skriptset")

		roh, istFehler := mcpRuf(t, umgebung, "schreiben.kartenskript",
			map[string]any{"setID": zahl(t, setID), "name": "Zwei Farben"})
		if istFehler {
			t.Fatalf("schreiben.kartenskript: %s", roh)
		}

		pruefeZweiKarten(t, umgebung, setID)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := skriptumgebung(t)
		setID := neuesSet(t, umgebung, "Skriptset")

		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost, "/sets/"+setID+"/kartenskripte",
			`{"name":"Zwei Farben"}`)
		if status != http.StatusCreated {
			t.Fatalf("POST /sets/%s/kartenskripte: %d (%s)", setID, status, rumpf)
		}

		pruefeZweiKarten(t, umgebung, setID)
	})
}

// pruefeSkripte hält die fachliche Zusage — an jeder Fläche dieselbe.
func pruefeSkripte(t *testing.T, roh string) {
	t.Helper()

	var stand struct {
		Skripte []struct {
			Name         string `json:"name"`
			Kuerzel      string `json:"kuerzel"`
			Beschreibung string `json:"beschreibung"`
		} `json:"skripte"`
	}

	if err := json.Unmarshal([]byte(roh), &stand); err != nil {
		t.Fatalf("die Antwort ist kein JSON: %v — %s", err, roh)
	}

	if len(stand.Skripte) != 2 {
		t.Fatalf("%d Skripte, erwartet waren zwei: %s", len(stand.Skripte), roh)
	}

	// **Das `&` ist abgetrennt.** In der Datei steht `&Zwei Farben`; was eine
	// Fläche zeigt, ist der Name, und das Kürzel steht daneben.
	if stand.Skripte[0].Name != "Zwei Farben" || stand.Skripte[0].Kuerzel != "Z" {
		t.Errorf("das erste Skript lautet %+v", stand.Skripte[0])
	}

	if stand.Skripte[0].Beschreibung == "" {
		t.Error("die Beschreibung fehlt")
	}
}

// pruefeZweiKarten sieht nach, was im Set steht.
//
// **Über die Kartenliste und nicht über die Antwort des Laufs.** Der Lauf sagt,
// was er getan haben will; die Liste sagt, was dasteht — und das ist die
// Abnahme (Satz 4 des Audits).
func pruefeZweiKarten(t *testing.T, umgebung []string, setID string) {
	t.Helper()

	roh := mussKommandozeile(t, umgebung, "karten", setID, "--json")

	var stand struct {
		Karten []struct {
			Werte map[string]string `json:"werte"`
		} `json:"karten"`
	}

	if err := json.Unmarshal([]byte(roh), &stand); err != nil {
		t.Fatalf("die Kartenliste ist kein JSON: %v — %s", err, roh)
	}

	if len(stand.Karten) != 2 {
		t.Fatalf("%d Karten im Set, erwartet waren zwei: %s", len(stand.Karten), roh)
	}

	namen := map[string]string{}

	for _, karte := range stand.Karten {
		namen[karte.Werte["name"]] = karte.Werte["farbe"]
	}

	if namen["Rot"] != "rot" || namen["Blau"] != "blau" {
		t.Errorf("die Karten lauten %v", namen)
	}
}
