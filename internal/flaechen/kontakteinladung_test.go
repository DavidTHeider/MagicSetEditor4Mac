package flaechen_test

// **Einladen, vergleichen, freigeben über jede Fläche** — NET-01
// (#1705).
//
// # Was hier gemessen wird
//
// **Zwei Rechner mit je eigenem Bestand, Bund und Schlüssel** — der Austausch
// aus dem Artboard *Gegenüber*: Anna stellt an der Kommandozeile aus, Bert
// nimmt im Dienst an, und umgekehrt. **Beide sehen denselben Vergleichscode**,
// und eine Einladung wirkt einmal.
//
// **MCP fehlt mit Absicht**: annehmen und bestätigen sind
// Vertrauensentscheidungen (0044), und einladen gibt heraus, wie ein Rechner
// zu erreichen ist.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestEinladungUeberJedeFlaeche(t *testing.T) {
	anna, bert := stelleAuf(t), stelleAuf(t)
	bertsDienst := starteDienst(t, bert.umgebung)

	// **Anna stellt an der Kommandozeile aus**, Bert prüft im Dienst.
	code := zeilenwert(t, mussKommandozeile(t, anna.umgebung,
		"einladung-erstellen", "Anna", "handel"), "code")

	var vorschau struct {
		Name           string `json:"name"`
		Lage           string `json:"lage"`
		Vergleichscode string `json:"vergleichscode"`
	}

	rumpf, fehler := json.Marshal(map[string]string{"code": code})
	if fehler != nil {
		t.Fatal(fehler)
	}

	if gelesen := json.Unmarshal([]byte(mussREST(t, bertsDienst, http.MethodPost,
		"/peer/einladungen/pruefung", string(rumpf))), &vorschau); gelesen != nil {
		t.Fatal(gelesen)
	}

	if vorschau.Name != "Anna" || vorschau.Lage != "unbekannt" || vorschau.Vergleichscode == "" {
		t.Fatalf("Vorschau: %+v", vorschau)
	}

	// **Dieselbe Vorschau an der Kommandozeile** — und sie schreibt nichts.
	geprueft := mussKommandozeile(t, bert.umgebung, "einladung-pruefen", code)
	if zeilenwert(t, geprueft, "vergleich") != vorschau.Vergleichscode ||
		zeilenwert(t, geprueft, "lage") != "unbekannt" {
		t.Fatalf("die Kommandozeile prüft anders: %q", geprueft)
	}

	// **Bert nimmt im Dienst an, mit Vergleich und Freigabe.**
	annahme, fehler := json.Marshal(map[string]any{
		"code": code, "name": "Anna vom Stammtisch",
		"vergleichscode": vorschau.Vergleichscode, "freigeben": true,
	})
	if fehler != nil {
		t.Fatal(fehler)
	}

	if aus := mussREST(t, bertsDienst, http.MethodPost,
		"/peer/einladungen/annahme", string(annahme)); !strings.Contains(aus, `"rechte":["handel"]`) {
		t.Fatalf("Annahme: %s", aus)
	}

	// **Dieselbe Einladung ein zweites Mal** — an der Kommandozeile, und
	// sie scheitert dort genauso.
	if _, _, rc := kommandozeile(t, bert.umgebung,
		"einladung-annehmen", code, "Anna"); rc == 0 {
		t.Fatal("dieselbe Einladung zweimal angenommen")
	}

	// **Und umgekehrt**: Bert stellt im Dienst aus, Anna nimmt an der
	// Kommandozeile an — ohne Vergleich zuerst.
	var bertsEinladung struct {
		Code string `json:"code"`
	}

	if fehler := json.Unmarshal([]byte(mussREST(t, bertsDienst, http.MethodPost,
		"/peer/einladungen", `{"name":"Bert","zweck":"handel"}`)), &bertsEinladung); fehler != nil {
		t.Fatal(fehler)
	}

	liste := mussKommandozeile(t, anna.umgebung,
		"einladung-annehmen", bertsEinladung.Code, "Bert")
	annasCode := zeilenwert(t, strings.ReplaceAll(liste, "\tvergleich\t", "vergleich\t"), "vergleich")

	// **Derselbe Code auf beiden Seiten** — das ist der Vergleich.
	if annasCode != vorschau.Vergleichscode {
		t.Fatalf("Anna sieht %q, Bert sah %q", annasCode, vorschau.Vergleichscode)
	}

	// **Bestätigt mit dem Code**, an derselben Stelle wie mit dem Abdruck.
	if aus := mussKommandozeile(t, anna.umgebung,
		"kontakt-bestaetigen", bert.kennung, annasCode); !strings.Contains(aus, "bestaetigt") {
		t.Fatalf("Bestätigen: %q", aus)
	}
}

func TestEineFalscheEinladungHinterlaesstNichts(t *testing.T) {
	anna, bert := stelleAuf(t), stelleAuf(t)

	code := zeilenwert(t, mussKommandozeile(t, anna.umgebung,
		"einladung-erstellen", "Anna", "partie"), "code")

	// **Ein falscher Vergleichscode** — AK-04: kein halber Kontakt.
	if _, _, rc := kommandozeile(t, bert.umgebung, "einladung-annehmen",
		code, "Anna", "00000 00000 00000 00000 00000 00000"); rc == 0 {
		t.Fatal("ein falscher Code wurde angenommen")
	}

	if aus := mussKommandozeile(t, bert.umgebung, "kontakte"); strings.TrimSpace(aus) != "" {
		t.Fatalf("es blieb ein Kontakt: %q", aus)
	}

	// **Und `freigeben` ohne Vergleich** wird an der Bedienung abgewiesen,
	// weil das Wort an der Stelle des Codes steht.
	if _, _, rc := kommandozeile(t, bert.umgebung, "einladung-annehmen",
		code, "Anna", "", "freigeben"); rc == 0 {
		t.Fatal("freigegeben ohne Vergleich")
	}
}
