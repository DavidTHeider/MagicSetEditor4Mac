package flaechen_test

// **Die Liste der Handelsvorgänge und ihre Entwürfe — über drei Flächen** —
// TRADE-01 (#1715).
//
// Eröffnet wird an der Kommandozeile; Kommandozeile, Dienst und MCP nennen
// denselben Vorgang mit derselben Lage. Kopiert wird an der Kommandozeile,
// geändert und gelöscht im Dienst — und **dasselbe Nein** an beiden: was
// läuft, geht nicht ins Archiv. §16: dieselbe Fähigkeit, dieselbe Antwort.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

// Die Wege, wie sie im Katalog stehen.
const (
	derEntwurfsweg = "PUT /handel/{vorgang}/entwurf"
	derVerwerfweg  = "DELETE /handel/{vorgang}/entwurf"
	derKopieweg    = "POST /handel/{vorgang}/kopie"
	derArchivweg   = "PUT /handel/{vorgang}/archiv"
)

// amVorgang setzt den Vorgang in einen Weg ein.
func amVorgang(weg, vorgang string) string {
	return strings.ReplaceAll(pfadAus(weg), "{vorgang}", vorgang)
}

func TestDieHandelslisteIstUeberDreiFlaechenDieselbe(t *testing.T) {
	umgebung := append(mitEigenerSaat(t), schluesselbund.Bundort+"="+probebund(t))
	an := starteDienst(t, umgebung)

	mussKommandozeile(t, umgebung, "peer-anlegen")

	eroeffnet := mitEingabe(t, umgebung, diePositionen, "handel", "eroeffnen", "", "tausch", dasGegenueberImHandel)
	kennung, _, _ := strings.Cut(eroeffnet, "\t")

	// **Dieselbe Zeile an drei Flächen** — Kennung und Lage.
	for flaeche, aus := range map[string]string{
		"Kommandozeile": mussKommandozeile(t, umgebung, "handel", "liste", "--lage=dran", "--json"),
		"Dienst":        mussREST(t, an, http.MethodGet, "/handel?lage=dran", ""),
		"MCP":           mussMCP(t, umgebung, "lesen.handelsuebersicht", map[string]any{"lagen": []string{"dran"}}),
	} {
		var liste struct {
			Zeilen []struct {
				Vorgang struct {
					Vorgang string `json:"vorgang"`
				} `json:"vorgang"`
				Lage      string `json:"lage"`
				Naechster string `json:"naechster"`
			} `json:"zeilen"`
		}

		if fehler := json.Unmarshal([]byte(aus), &liste); fehler != nil || len(liste.Zeilen) != 1 ||
			liste.Zeilen[0].Vorgang.Vorgang != kennung || liste.Zeilen[0].Lage != "dran" ||
			liste.Zeilen[0].Naechster != "anbieten" {
			t.Errorf("%s: %s (%v)", flaeche, aus, fehler)
		}
	}

	// **Kopieren an beiden Flächen, ändern und löschen im Dienst.**
	kopiert := mussKommandozeile(t, umgebung, "handel", "kopieren", kennung)
	kopie, _, _ := strings.Cut(kopiert, "\t")

	if kopie == kennung || !strings.HasPrefix(kopie, "h-") {
		t.Fatalf("kopiert wurde %q", kopiert)
	}

	if aus := mussREST(t, an, http.MethodPost, amVorgang(derKopieweg, kennung), ""); !strings.Contains(aus, `"zustand":"entwurf"`) {
		t.Errorf("die Kopie im Dienst: %s", aus)
	}

	mussREST(t, an, http.MethodPut, amVorgang(derEntwurfsweg, kopie),
		`{"art":"tausch","gegenueber":["`+dasGegenueberImHandel+`"],"positionen":`+diePositionen+`}`)
	mussREST(t, an, http.MethodDelete, amVorgang(derVerwerfweg, kopie), "")

	// **Dasselbe Nein**: was läuft, geht nicht ins Archiv.
	if _, _, code := kommandozeile(t, umgebung, "handel", "archivieren", kennung); code != 1 {
		t.Errorf("die Kommandozeile gab %d, erwartet 1", code)
	}

	if status, antwort := an.ruf(t, http.MethodPut, amVorgang(derArchivweg, kennung), ""); status != http.StatusConflict {
		t.Errorf("der Dienst gab %d: %s", status, antwort)
	}

	if aus := mussKommandozeile(t, umgebung, "handel", "verwerfen", kennung); strings.TrimSpace(aus) != "verworfen" {
		t.Errorf("verwerfen: %q", aus)
	}

	// **Übrig ist die zweite Kopie** — ein Entwurf, den der Dienst anlegte.
	//
	// **Gezählt werden Zeilen und nicht Wörter**: seit TRADE-02 (#1716)
	// trägt auch der Vorgang in der Zeile seine Lage.
	aus := mussREST(t, an, http.MethodGet, "/handel", "")

	var uebrig struct {
		Zeilen []struct {
			Lage string `json:"lage"`
		} `json:"zeilen"`
	}

	if fehler := json.Unmarshal([]byte(aus), &uebrig); fehler != nil ||
		len(uebrig.Zeilen) != 1 || uebrig.Zeilen[0].Lage != "dran" {
		t.Errorf("nach dem Verwerfen: %s", aus)
	}
}
