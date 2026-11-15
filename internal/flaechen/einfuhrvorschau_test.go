// **Eine Einfuhr vorab ansehen — an jeder Fläche** — DOM-03.AK-01
// (#1698).
//
// # Der Ablauf
//
// Die Sammlung mit einem Exemplar über vier Stück ausgeben → dasselbe CSV
// vorab ansehen → mit `auslassen` zurücklesen → die Menge ist noch vier.
//
// **Dieselbe Antwort an jeder Fläche**: die Vorschau trifft die eine Zeile
// (vier → acht beim Addieren), und das Zurücklesen mit Auslassen lässt sie
// stehen. Vor der Vorschau verdoppelte genau dieser Weg jede Menge still.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestDieEigeneAusfuhrLaesstSichUeberallVorabAnsehen(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitExemplar(t)
		text := mussKommandozeile(t, umgebung, "sammlung-ausgeben", "csv")

		pruefeVorschau(t, mitEingabe(t, umgebung, text, "sammlung-vorschau", "--json"))

		mitEingabe(t, umgebung, text, "sammlung-einlesen", "--auslassen")
		pruefeMengeVier(t, mussKommandozeile(t, umgebung, "sammlung", "--json"))
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitExemplar(t)
		text := mussKommandozeile(t, umgebung, "sammlung-ausgeben", "csv")

		pruefeVorschau(t, mussMCP(t, umgebung, "lesen.sammlung.vorschau", map[string]any{"text": text}))

		mussMCP(t, umgebung, "schreiben.sammlung.einlesen",
			map[string]any{"text": text, "zusammenfuehrung": "auslassen"})
		pruefeMengeVier(t, mussKommandozeile(t, umgebung, "sammlung", "--json"))
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitExemplar(t)
		text := mussKommandozeile(t, umgebung, "sammlung-ausgeben", "csv")
		an := starteDienst(t, umgebung)

		pruefeVorschau(t, mussREST(t, an, http.MethodPost, "/sammlung/vorschau",
			string(mussJSON(t, map[string]any{"text": text}))))

		mussREST(t, an, http.MethodPost, "/sammlung/einlesen",
			string(mussJSON(t, map[string]any{"text": text, "zusammenfuehrung": "auslassen"})))
		pruefeMengeVier(t, mussKommandozeile(t, umgebung, "sammlung", "--json"))
	})
}

// **Die Liste vorab — über dieselbe Gegenstelle wie ihr Einlesen.** Leer
// ist die erste eindeutige Zeile neu; nach dem Einlesen trifft sie sich
// selbst.
func TestEineListeLaesstSichUeberallVorabAnsehen(t *testing.T) {
	t.Parallel()

	zeile := map[string]any{
		"printing": "00000000-0000-4000-8000-000000000001",
		"menge":    2, "zustand": "nearMint", "sprache": "en",
		"name": "Erste Probe", "setcode": "probe", "nummer": "1",
	}

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := scryfallumgebung(t, scryfallgegenstelle(t))

		pruefeArt(t, mitEingabe(t, umgebung, probeliste, "liste-vorschau", "--json"), "neu")
		mitEingabe(t, umgebung, probeliste, "liste-einlesen")
		pruefeArt(t, mitEingabe(t, umgebung, probeliste, "liste-vorschau", "--json"), "zusammen")
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := scryfallumgebung(t, scryfallgegenstelle(t))
		argumente := map[string]any{"zeilen": []any{zeile}}

		pruefeArt(t, mussMCP(t, umgebung, "lesen.liste.vorschau", argumente), "neu")
		mussMCP(t, umgebung, "schreiben.liste.einlesen", argumente)
		pruefeArt(t, mussMCP(t, umgebung, "lesen.liste.vorschau", argumente), "zusammen")
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := scryfallumgebung(t, scryfallgegenstelle(t))
		an := starteDienst(t, umgebung)
		rumpf := string(mussJSON(t, map[string]any{"zeilen": []any{zeile}}))

		pruefeArt(t, mussREST(t, an, http.MethodPost, "/sammlung/liste/vorschau", rumpf), "neu")
		mussREST(t, an, http.MethodPost, "/sammlung/liste/einlesen", rumpf)
		pruefeArt(t, mussREST(t, an, http.MethodPost, "/sammlung/liste/vorschau", rumpf), "zusammen")
	})
}

// vorschauposten liest die Posten einer Vorschau.
func vorschauposten(t *testing.T, roh string) []struct {
	Art     string `json:"art"`
	Vorher  int    `json:"vorher"`
	Nachher int    `json:"nachher"`
} {
	t.Helper()

	var vorschau struct {
		Posten []struct {
			Art     string `json:"art"`
			Vorher  int    `json:"vorher"`
			Nachher int    `json:"nachher"`
		} `json:"posten"`
	}

	if err := json.Unmarshal([]byte(roh), &vorschau); err != nil {
		t.Fatalf("die Vorschau ist kein gültiges JSON: %v (%s)", err, roh)
	}

	return vorschau.Posten
}

// pruefeVorschau hält, dass die eigene Ausfuhr ihre eine Zeile trifft.
func pruefeVorschau(t *testing.T, roh string) {
	t.Helper()

	posten := vorschauposten(t, roh)
	if len(posten) != 1 || posten[0].Art != "zusammen" || posten[0].Vorher != 4 || posten[0].Nachher != 8 {
		t.Errorf("die Vorschau ist %s", roh)
	}
}

// pruefeArt hält die Art des ersten Postens.
func pruefeArt(t *testing.T, roh, art string) {
	t.Helper()

	posten := vorschauposten(t, roh)
	if len(posten) == 0 || posten[0].Art != art {
		t.Errorf("erwartet %q, die Vorschau ist %s", art, roh)
	}
}

// pruefeMengeVier hält, dass beim Auslassen nichts dazukam.
func pruefeMengeVier(t *testing.T, roh string) {
	t.Helper()

	var sammlung struct {
		Exemplare []struct {
			Menge int `json:"menge"`
		} `json:"exemplare"`
	}

	if err := json.Unmarshal([]byte(roh), &sammlung); err != nil {
		t.Fatalf("die Sammlung lesen: %v (%s)", err, roh)
	}

	if len(sammlung.Exemplare) != 1 || sammlung.Exemplare[0].Menge != 4 {
		t.Errorf("die Sammlung ist %s — beim Auslassen kam etwas dazu", roh)
	}
}
