// **Die Zwischenablage über jede Fläche ohne Fenster** — Werkstück 1 von
// Phase 3.7
// (#1168).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// **Eine Karte von Set A nach Set B — und nichts geht dabei verloren.**
//
// # Was hier die Zusage ist
//
// Nicht „der Weg läuft durch", sondern: **die Feldwerte der eingefügten Karte
// sind die der kopierten**. Das ist die Frage, an der eine Zwischenablage
// scheitert, ohne dass es auffällt — eine Karte entsteht, sie hat einen
// Namen, und drei Felder fehlen.
//
// Geprüft wird über **alle** Werte und nicht über einen genannten: ein Feld,
// das der Ablageschreiber vergisst, stünde sonst nicht in der Liste, gegen die
// verglichen wird.
func TestEineKarteWandertUeberJedeFlaecheOhneVerlust(t *testing.T) {
	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		vonID, nachID, karteID := zweiSets(t, umgebung)

		vorher := feldwerte(t, umgebung, vonID, karteID)

		inhalt := mussKommandozeile(t, umgebung, "ablage-kopieren", vonID, karteID)

		// **Über die Standardeingabe**, denn genau das ist die Ablage ohne
		// Fenster.
		eingefuegt := mitEingabe(t, umgebung, inhalt, "ablage-einfuegen", nachID, "--json")

		gleicheWerte(t, vorher, feldwerte(t, umgebung, nachID, ersteKarteVon(t, eingefuegt)))
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		vonID, nachID, karteID := zweiSets(t, umgebung)

		vorher := feldwerte(t, umgebung, vonID, karteID)

		text, istFehler := mcpRuf(t, umgebung, "lesen.ablage.kopieren",
			map[string]any{"setID": zahl(t, vonID), "karteIDs": []int64{zahl(t, karteID)}})
		if istFehler {
			t.Fatalf("lesen.ablage.kopieren: %s", text)
		}

		eingefuegt, istFehler := mcpRuf(t, umgebung, "schreiben.ablage.einfuegen",
			map[string]any{"setID": zahl(t, nachID), "inhalt": inhaltAus(t, text)})
		if istFehler {
			t.Fatalf("schreiben.ablage.einfuegen: %s", eingefuegt)
		}

		gleicheWerte(t, vorher,
			feldwerte(t, umgebung, nachID, ersteKarteVon(t, eingefuegt)))
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		vonID, nachID, karteID := zweiSets(t, umgebung)
		an := starteDienst(t, umgebung)

		vorher := feldwerte(t, umgebung, vonID, karteID)

		status, rumpf := an.ruf(t, http.MethodPost, "/sets/"+vonID+"/ablage/kopieren",
			`{"karten":[`+karteID+`]}`)
		if status != http.StatusOK {
			t.Fatalf("POST ablage/kopieren: %d (%s)", status, rumpf)
		}

		status, eingefuegt := an.ruf(t, http.MethodPost, "/sets/"+nachID+"/ablage/einfuegen",
			string(mussJSON(t, map[string]any{"inhalt": inhaltAus(t, rumpf)})))
		if status != http.StatusOK {
			t.Fatalf("POST ablage/einfuegen: %d (%s)", status, eingefuegt)
		}

		gleicheWerte(t, vorher,
			feldwerte(t, umgebung, nachID, ersteKarteVon(t, eingefuegt)))
	})
}

// **Ausschneiden nimmt die Karte mit** — und zwar erst, nachdem sie in der
// Ablage steht.
//
// # ported
//
//	bool CardListBase::canCut()   const { return canDelete(); }
//
// `src/gui/control/card_list.cpp`, Zeile 137, Fassung f9d9356 vom 2021-07-06.
// Ausschneiden ist Kopieren **und** Löschen; wer nur löscht, hat den Inhalt
// verloren, den er gerade wegnehmen wollte.
func TestAusschneidenNimmtDieKarteMit(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	vonID, nachID, karteID := zweiSets(t, umgebung)

	vorher := feldwerte(t, umgebung, vonID, karteID)
	vorZahl := len(kartenzeilenAus(t, mussKommandozeile(t, umgebung, "karten", vonID, "--json")))

	inhalt := mussKommandozeile(t, umgebung, "ablage-ausschneiden", vonID, karteID)

	nachZahl := len(kartenzeilenAus(t, mussKommandozeile(t, umgebung, "karten", vonID, "--json")))
	if nachZahl != vorZahl-1 {
		t.Fatalf("%d Karten im Quellset, erwartet %d", nachZahl, vorZahl-1)
	}

	eingefuegt := mitEingabe(t, umgebung, inhalt, "ablage-einfuegen", nachID, "--json")

	gleicheWerte(t, vorher, feldwerte(t, umgebung, nachID, ersteKarteVon(t, eingefuegt)))
}

// **Ein fremdes Spiel wird abgewiesen, und zwar überall.**
//
// # Warum gerade dieser Fall
//
// Er ist die **Grenze** der Ablage: die Feldnamen eines Spiels sagen im
// anderen nichts. Ginge er durch, entstünden Karten mit Werten, die keine
// Vorlage kennt — und niemand sähe es, bis jemand sie zu zeichnen versucht.
//
// **Und es darf keine halbe Karte zurückbleiben** (§8): geprüft wird vor dem
// Anlegen, nicht dazwischen.
func TestEinFremdesSpielWirdUeberallAbgewiesen(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	vonID, _, karteID := zweiSets(t, umgebung)

	inhalt := mussKommandozeile(t, umgebung, "ablage-kopieren", vonID, karteID)

	// **Das Spiel wird im Text ausgetauscht** — dieselbe Karte, ein anderes
	// Spiel. Ein zweites Set einer anderen Vorlage anzulegen ginge auch,
	// verlangte aber ein zweites Paket in der Saat.
	fremd := strings.Replace(inhalt, "game_name: probe", "game_name: magic", 1)
	if fremd == inhalt {
		t.Fatalf("der Kopf nennt das Spiel nicht:\n%s", inhalt)
	}

	nachID := leeresSet(t, umgebung, "Ziel")

	t.Run("Kommandozeile", func(t *testing.T) {
		_, _, code := mitEingabeUndCode(t, umgebung, fremd, "ablage-einfuegen", nachID)
		if code == 0 {
			t.Fatal("das fremde Spiel kam durch")
		}
	})

	t.Run("MCP", func(t *testing.T) {
		text, istFehler := mcpRuf(t, umgebung, "schreiben.ablage.einfuegen",
			map[string]any{"setID": zahl(t, nachID), "inhalt": fremd})
		if !istFehler {
			t.Fatalf("das fremde Spiel kam durch: %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost, "/sets/"+nachID+"/ablage/einfuegen",
			string(mussJSON(t, map[string]any{"inhalt": fremd})))
		if status == http.StatusOK {
			t.Fatalf("das fremde Spiel kam durch: %s", rumpf)
		}
	})

	// **Und danach ist das Zielset leer geblieben.**
	if zeilen := kartenzeilenAus(t, mussKommandozeile(t, umgebung, "karten", nachID, "--json")); len(zeilen) != 0 {
		t.Errorf("%d Karten im Zielset, erwartet keine", len(zeilen))
	}
}

// zweiSets gibt ein Quellset mit einer Karte und ein leeres Zielset.
func zweiSets(t *testing.T, umgebung []string) (vonID, nachID, karteID string) {
	t.Helper()

	vonID = leeresSet(t, umgebung, "Quelle")

	mussKommandozeile(t, umgebung, "karte", vonID, "anlegen")

	karten := kartenzeilenAus(t, mussKommandozeile(t, umgebung, "karten", vonID, "--json"))
	if len(karten) == 0 {
		t.Fatal("das Quellset hat keine Karte")
	}

	karteID = karten[0].ID

	// **Werte in mehreren Feldern**, denn ein Verlust zeigt sich nicht am
	// Namen allein.
	for feld, wert := range map[string]string{
		"name": "Wanderfalke", "rule text": "Fliegend.", "type": "Kreatur — Vogel",
	} {
		mussKommandozeile(t, umgebung, "feld", vonID, karteID, feld, wert)
	}

	return vonID, leeresSet(t, umgebung, "Ziel"), karteID
}

// leeresSet legt ein leeres Set der Saatvorlage an.
//
// **Nicht `probeset`** — das ist der Pfad der Datei in `testdaten/`, und zwei
// Namen für zwei Dinge sind billiger als ein Name für beide.
func leeresSet(t *testing.T, umgebung []string, name string) string {
	t.Helper()

	var antwort struct {
		ID string `json:"id"`
	}

	roh := mussKommandozeile(t, umgebung, "set-neu", name, "probe", "probe-schlicht", "--json")
	if err := json.Unmarshal([]byte(roh), &antwort); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, roh)
	}

	return antwort.ID
}

// feldwerte gibt die Werte einer Karte als Abbildung Feld → Wert.
func feldwerte(t *testing.T, umgebung []string, setID, karteID string) map[string]string {
	t.Helper()

	var antwort struct {
		Karten []struct {
			ID    string            `json:"id"`
			Werte map[string]string `json:"werte"`
		} `json:"karten"`
	}

	roh := mussKommandozeile(t, umgebung, "karten", setID, "--json")
	if err := json.Unmarshal([]byte(roh), &antwort); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, roh)
	}

	for _, karte := range antwort.Karten {
		if karte.ID != karteID {
			continue
		}

		return karte.Werte
	}

	t.Fatalf("die Karte %s steht nicht in %s", karteID, setID)

	return nil
}

// gleicheWerte vergleicht zwei Wertabbildungen und nennt jede Abweichung.
func gleicheWerte(t *testing.T, vorher, nachher map[string]string) {
	t.Helper()

	if len(vorher) == 0 {
		t.Fatal("die kopierte Karte hatte keine Werte — der Vergleich sagte nichts")
	}

	for feld, wert := range vorher {
		if nachher[feld] != wert {
			t.Errorf("%q ist nach dem Einfügen %q, vorher war es %q",
				feld, nachher[feld], wert)
		}
	}

	for feld, wert := range nachher {
		if _, stand := vorher[feld]; !stand {
			t.Errorf("%q steht nach dem Einfügen mit %q da und stand vorher nicht", feld, wert)
		}
	}
}

// ersteKarteVon liest die erste Kennung aus einer Einfügeantwort.
func ersteKarteVon(t *testing.T, roh string) string {
	t.Helper()

	var antwort struct {
		Karten []string `json:"karten"`
	}

	if err := json.Unmarshal([]byte(roh), &antwort); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, roh)
	}

	if len(antwort.Karten) != 1 {
		t.Fatalf("%d Karten eingefügt, erwartet eine:\n%s", len(antwort.Karten), roh)
	}

	return antwort.Karten[0]
}

// inhaltAus liest das Feld `inhalt` aus einer Kopierantwort.
func inhaltAus(t *testing.T, roh string) string {
	t.Helper()

	var antwort struct {
		Inhalt string `json:"inhalt"`
	}

	if err := json.Unmarshal([]byte(roh), &antwort); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, roh)
	}

	if strings.TrimSpace(antwort.Inhalt) == "" {
		t.Fatalf("die Antwort trägt keinen Inhalt:\n%s", roh)
	}

	return antwort.Inhalt
}
