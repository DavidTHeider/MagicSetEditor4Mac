// **Die Rolle über die übrigen Schreibklassen** — C04
// (#2229).
//
// Der Lauf daneben fährt die Klassen, an denen fremde Arbeit am meisten
// verliert. Hier gehen die übrigen mit Profilträger in dieselben vier Lagen:
// Symbol, Stiloption, Marke, Kartenbild, Umschlag, Gegenüber, Freigabe an
// ein Gegenüber, Handel, Partie, Paket, Zurücksetzen, KI-Messung,
// Kostengrenze, Preise und Suchindex — je über jede Fläche, die die
// Fähigkeit hat (`FAEHIGKEITEN.md`), und wo es billig ist, **an der
// Wirkung**.
//
// **Was die Lage braucht, legt der Eigentümer an** (`richteLageEin` mit
// Vorbereitung): ein zweites Set mit Bildfeld, drei Gegenüber, die eigene
// Identität für den Handel, zwei Decks für die Partie und die Probelaufzeit
// für den Index.
//
// **Was er nicht fährt**: das Fenster — es fährt die Lagen in
// `internal/app`.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

// gegenueber ist ein fremder Peer, den eine Lage als Kontakt kennt.
type gegenueber struct{ kennung, abdruck string }

func neuesGegenueber(t *testing.T) gegenueber {
	t.Helper()

	fremd, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	return gegenueber{string(fremd.Kennung()), fremd.Fingerabdruck()}
}

func TestDieUebrigenSchreibklassenFolgenDerRolle(t *testing.T) {
	t.Parallel()

	for _, lage := range []rollenlage{
		{"eigenes Profil", "", false, true},
		{"fremd als Mitarbeiter", "mitarbeiter", true, true},
		{"fremd als Leser", "leser", true, false},
		{"fremd ohne Rolle", "", true, false},
	} {
		t.Run(lage.name, func(t *testing.T) {
			t.Parallel()

			fahreUebrigeKlassen(t, lage)
		})
	}
}

func fahreUebrigeKlassen(t *testing.T, lage rollenlage) {
	t.Helper()

	umgebung := append(mitEigenerSaat(t),
		app.Scryfallumgebung+"="+scryfallgegenstelle(t).URL,
		schluesselbund.Bundort+"="+probebund(t))
	notiert, gesperrt, freigegeben := neuesGegenueber(t), neuesGegenueber(t), neuesGegenueber(t)
	laufzeit, sammeldatei := starteLaufzeit(t, derVorschlag), eineSammeldatei(t)

	var kunstSetID, kunstKarteID, ersterDeck, zweiterDeck string

	// messungsprofil ist das Profil der Lage, in dem eine Messung liegt — das
	// Vergessen hat sonst nichts, woran es sich zeigt (C04-Rest, #2284).
	var messungsprofil int64

	setID := richteLageEin(t, umgebung, lage, func(string) {
		kunstSetID, kunstKarteID = kunstset(t, umgebung)
		ersterDeck, zweiterDeck = zweiDecks(t, umgebung)

		mussKommandozeile(t, umgebung, "peer-anlegen")

		for _, wer := range []gegenueber{notiert, gesperrt, freigegeben} {
			mussKommandozeile(t, umgebung, "kontakt-aufnehmen", wer.kennung, "Gegenüber")
		}

		mussKommandozeile(t, umgebung, "kontakt-bestaetigen", freigegeben.kennung, freigegeben.abdruck)
		kiEinschalten(t, umgebung, laufzeit.adresse)
		messungsprofil = merkeEineMessung(t, umgebung)
	})
	karteID := ersteKarte(t, mussKommandozeile(t, umgebung, "karten", setID, "--json"))
	mcp := starteMCPProzess(t, umgebung)
	an := starteDienst(t, umgebung)
	ablage, quelle, kreis := t.TempDir(), probebild(t), kreisbild(t)

	kommandozeileGelang := func(argumente ...string) bool {
		_, _, code := kommandozeile(t, umgebung, argumente...)

		return code == 0
	}
	eingabeGelang := func(eingabe string, argumente ...string) bool {
		_, _, code := mitEingabeUndCode(t, umgebung, eingabe, argumente...)

		return code == 0
	}
	mcpGelang := func(werkzeug string, argumente map[string]any) bool {
		_, istFehler := mcp.ruf(werkzeug, argumente)

		return !istFehler
	}
	dienstGelang := func(methode, weg, rumpf string) bool {
		status, _ := an.ruf(t, methode, weg, rumpf)

		return status >= http.StatusOK && status < http.StatusMultipleChoices
	}

	set, karte := json.Number(setID), json.Number(karteID)
	sitze := sitzwuensche(ersterDeck, zweiterDeck)
	handel := `{"art":"tausch","gegenueber":["` + dasGegenueberImHandel + `"],"positionen":` + diePositionen + `}`

	for wirkung, gelang := range map[string]map[string]bool{
		"ein Symbol": {
			"Kommandozeile": eingabeGelang(symbolteile, "symbol-setzen", setID),
			"MCP":           mcpGelang("schreiben.symbol", map[string]any{"setID": set, "teile": rohTeile(t)}),
			"Dienst":        dienstGelang(http.MethodPut, "/sets/"+setID+"/symbol", `{"teile":`+symbolteile+`,"unter":""}`),
		},
		// **Und nachgezeichnet**: derselbe Weg in die Ablage, über ein Bild.
		"ein nachgezeichnetes Symbol": {
			"Kommandozeile": kommandozeileGelang("symbol-nachzeichnen", setID, kreis),
			"MCP":           mcpGelang("schreiben.symbolnachzeichnen", map[string]any{"setID": set, "bildpfad": kreis}),
			"Dienst": dienstGelang(http.MethodPost, "/sets/"+setID+"/symbol-nachzeichnen",
				string(mussJSON(t, map[string]string{"bildpfad": kreis}))),
		},
		"eine Stiloption": {
			"Kommandozeile": kommandozeileGelang("stiloption", setID, "setzen", "hintergrund", "dunkel"),
			"MCP": mcpGelang("schreiben.stiloption", map[string]any{
				"setID": set, "feld": "hintergrund", "wert": "dunkel",
			}),
			"Dienst": dienstGelang(http.MethodPut, "/sets/"+setID+"/stiloptionen/hintergrund", `{"wert":"dunkel"}`),
		},
		"eine Marke": {
			"Kommandozeile": kommandozeileGelang("marke", setID, karteID, "Zeile"),
			"MCP":           mcpGelang("schreiben.marke", map[string]any{"setID": set, "karteID": karte, "marke": "Modell"}),
			"Dienst":        dienstGelang(http.MethodPut, "/sets/"+setID+"/karten/"+karteID+"/marken/Dienst", ""),
		},
		"ein Kartenbild": {
			"Kommandozeile": kommandozeileGelang("bild-setzen", kunstSetID, kunstKarteID, "kunst", quelle),
			"MCP": mcpGelang("schreiben.karte.bild", map[string]any{
				"setID": json.Number(kunstSetID), "karteID": json.Number(kunstKarteID), "feld": "kunst", "pfad": quelle,
			}),
			"Dienst": dienstGelang(http.MethodPut, "/sets/"+kunstSetID+"/karten/"+kunstKarteID+"/felder/kunst/bild",
				`{"pfad":`+alsJSONText(t, quelle)+`}`),
		},
		"ein Umschlag": {
			"Kommandozeile": kommandozeileGelang("umschlag-schnueren", setID, filepath.Join(ablage, "zeile.msepackage")),
			"MCP": mcpGelang("schreiben.umschlag.schnueren", map[string]any{
				"setID": setID, "ziel": filepath.Join(ablage, "modell.msepackage"),
			}),
			"Dienst": dienstGelang(http.MethodPost, "/umschlag",
				`{"setID":"`+setID+`","ziel":`+alsJSONText(t, filepath.Join(ablage, "dienst.msepackage"))+`}`),
		},
		// **Drei Wege derselben Klasse**: notieren an Zeile und Dienst,
		// sperren über MCP — das einzige Vertrauenswerkzeug dort.
		"ein Gegenüber": {
			"Kommandozeile": kommandozeileGelang("kontakt-notiz", notiert.kennung, "Aus der Zeile"),
			"MCP":           mcpGelang("schreiben.peerkontakt.sperren", map[string]any{"kennung": gesperrt.kennung}),
			"Dienst":        dienstGelang(http.MethodPut, "/peer/kontakte/"+notiert.kennung+"/notiz", `{"notiz":"Aus dem Dienst"}`),
		},
		"eine Freigabe an ein Gegenüber": {
			"Kommandozeile": kommandozeileGelang("kontakt-freigeben", freigegeben.kennung, "handel", "sammlung"),
			"Dienst": dienstGelang(http.MethodPost, "/peer/kontakte/"+freigegeben.kennung+"/freigaben",
				`{"zweck":"handel","umfang":"sammlung"}`),
		},
		"ein Handel": {
			"Kommandozeile": eingabeGelang(diePositionen, "handel", "eroeffnen", "", "tausch", dasGegenueberImHandel),
			"Dienst":        dienstGelang(http.MethodPost, "/handel", handel),
		},
		"eine Partie": {
			"Kommandozeile": kommandozeileGelang("partie-anlegen", "magic", "anna:Anna:"+ersterDeck, "bert:Bert:"+zweiterDeck),
			"MCP":           mcpGelang("schreiben.partie.anlegen", map[string]any{"spiel": "magic", "sitze": sitze}),
			"Dienst": dienstGelang(http.MethodPost, "/partien",
				string(mussJSON(t, map[string]any{"spiel": "magic", "sitze": sitze}))),
		},
		"ein Paket": {
			"Kommandozeile": kommandozeileGelang("paket-anlegen", "mse-game", "zeilenspiel", "Spiel der Zeile"),
			"MCP":           mcpGelang("schreiben.paket.anlegen", map[string]any{"art": "mse-game", "name": "modellspiel"}),
			"Dienst":        dienstGelang(http.MethodPost, "/pakete/neu", `{"art":"mse-game","name":"dienstspiel"}`),
		},
		"ein Zurücksetzen": {
			"Kommandozeile": kommandozeileGelang("zuruecksetzen", "layout", ""),
			"Dienst":        dienstGelang(http.MethodPost, "/ruecksetzung/layout", `{"name":""}`),
		},
		"eine vergessene Messung": {
			"Kommandozeile": kommandozeileGelang("ki-modell-vergessen", "openai", "gpt-probe"),
			"MCP":           mcpGelang("schreiben.ki.modell.vergessen", map[string]any{"anbieter": "openai", "alias": "gpt-probe"}),
			"Dienst":        dienstGelang(http.MethodPost, "/ki/modelle/vergessen", `{"anbieter":"openai","alias":"gpt-probe"}`),
		},
		"ein Preis": {
			"Kommandozeile": kommandozeileGelang("preise-holen", "bolt-m10"),
			"MCP":           mcpGelang("schreiben.preiseholen", map[string]any{"printing": "bolt-m10"}),
			"Dienst":        dienstGelang(http.MethodPost, "/drucke/bolt-m10/preise", ""),
		},
		"ein Suchindex": {
			"Kommandozeile": kommandozeileGelang("indexaufbau", probeeinbetter, sammeldatei, "prf-zeile", "2026-01-01"),
			"Dienst": dienstGelang(http.MethodPost, "/index/aufbau", string(mussJSON(t, map[string]string{
				"modell": probeeinbetter, "quelle": sammeldatei, "revision": "prf-dienst", "quelldatum": "2026-01-01",
			}))),
		},
		"eine Kostengrenze": {
			"Kommandozeile": kommandozeileGelang("kostenbudget", "anbieter", "tag", "50000", "EUR"),
			"Dienst":        dienstGelang(http.MethodPut, "/kosten/anbieter/budget", `{"zeitraum":"tag","wert":50000,"waehrung":"EUR"}`),
		},
	} {
		for flaeche, ja := range gelang {
			if ja != lage.darf {
				t.Errorf("%s über %s ging %v, erwartet %v", wirkung, flaeche, ja, lage.darf)
			}
		}
	}

	pruefeUebrigeWirkung(t, umgebung, lage, setID, ablage)
	pruefeMessung(t, umgebung, lage, messungsprofil)
}

// pruefeUebrigeWirkung sieht nach, ob im Profil der Lage genau steht, was
// entstehen durfte — je Klasse, an der es eine Liste gibt.
func pruefeUebrigeWirkung(t *testing.T, umgebung []string, lage rollenlage, setID, ablage string) {
	t.Helper()

	abgelegt, fehler := filepath.Glob(filepath.Join(ablage, "*.msepackage"))
	if fehler != nil {
		t.Fatalf("die Ablage lesen: %v", fehler)
	}

	if will := map[bool]int{true: 3, false: 0}[lage.darf]; len(abgelegt) != will {
		t.Errorf("%d Umschläge abgelegt, erwartet %d", len(abgelegt), will)
	}

	for _, liste := range []struct {
		aufruf []string
		marke  string
	}{
		{[]string{"marken", setID, "--json"}, `"zeile"`},
		// **„Aus de"**: Zeile und Dienst schreiben dieselbe Notiz, der
		// zweite überschreibt den ersten.
		{[]string{"kontakte"}, "notiz\tAus de"},
		{[]string{"kosten", "anbieter"}, "verfuegbar\t50000"},
	} {
		if inhalt := mussKommandozeile(t, umgebung, liste.aufruf...); strings.Contains(inhalt, liste.marke) != lage.darf {
			t.Errorf("%v der Lage nennt %q: %v, erwartet %v:\n%s",
				liste.aufruf, liste.marke, !lage.darf, lage.darf, inhalt)
		}
	}
}
