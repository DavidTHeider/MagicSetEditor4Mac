package rest_test

// **Die KI-Endpunkte** — Werkstück A7 von Phase 5.3
// (#1446).
//
// **Eine eigene Datei**, weil `endpunkteneben_test.go` mit ihnen über die
// Zeilengrenze aus §2 ginge.

import "net/http"

// kiendpunkte gibt die Wege der KI.
func kiendpunkte() []endpunkt {
	return []endpunkt{
		{
			name: "KI-Modelle lesen", methode: http.MethodGet,
			weg: "/ki/modelle", erfolg: http.StatusOK,
		},
		{
			name: "KI-Vorschlag holen", methode: http.MethodPost,
			weg: "/karten/1/ki-vorschlag", erfolg: http.StatusOK,
			rumpf: `{"alias": "beispiel:3b", "auftrag": "Schreibe."}`,
		},

		// **Der Batch steht nicht unter einer Karte** (A13.1, #1452): er
		// meint mehrere, und ein Weg unter **einer** behauptete eine
		// Zugehörigkeit, die es nicht gibt.
		{
			name: "einen Batch vorschlagen", methode: http.MethodPost,
			weg: "/ki/batch", erfolg: http.StatusOK,
			rumpf: `{"karteIds": [1, 2], "alias": "beispiel:3b", ` +
				`"auftrag": "Schreibe."}`,
		},

		// **Die Gruppe steht ebenfalls nicht unter einer Karte** (A13.1b,
		// #1452): welche Karten sie trifft, entscheidet erst der Server aus
		// den Aufträgen — und ein Weg je Karte wäre wieder ein Aufruf je
		// Karte, also N Transaktionen.
		{
			name: "eine Gruppe übernehmen", methode: http.MethodPost,
			weg: "/ki/gruppe", erfolg: http.StatusOK,
			rumpf: `{"wahl": [{"kennung": "a1", "felder": ["name"]}]}`,
		},

		// **Und die Erkennung steht unter gar keiner** (A13.3, #1452): ein
		// Foto wird gelesen, **bevor** eine Karte feststeht — und welche es
		// wird, entscheidet ein Mensch.
		{
			name: "ein Kartenfoto lesen", methode: http.MethodPost,
			weg: "/ki/erkennung", erfolg: http.StatusOK,
			rumpf: `{"pfad":"/tmp/karte.png","anbieter":"anthropic",` +
				`"alias":"claude-probe"}`,
		},

		// **Die Einschätzung steht unter ihrer Karte** (A13.2a, #1452) — sie
		// meint genau eine, anders als der Batch darüber.
		{
			name: "eine Einschätzung holen", methode: http.MethodPost,
			weg: "/karten/1/ki-einschaetzung", erfolg: http.StatusOK,
			rumpf: `{"art": "regel", "alias": "beispiel:3b", ` +
				`"frage": "Wie wirkt sie?"}`,
		},

		// **Die zwei Wege der Messung** — A9, #1448. `POST` und nicht `GET`:
		// die Messung kostet Geld und legt eine Aussage im Bestand ab; ein
		// `GET`, der beides tut, wäre ein Weg, den ein Zwischenspeicher
		// wiederholen dürfte.
		{
			name: "ein Modell messen", methode: http.MethodPost,
			weg: "/ki/modelle/pruefen", erfolg: http.StatusOK,
			rumpf: `{"anbieter": "openai", "alias": "gpt-probe"}`,
		},
		{
			name: "eine Messung vergessen", methode: http.MethodPost,
			weg: "/ki/modelle/vergessen", erfolg: http.StatusOK,
			rumpf: `{"anbieter": "openai", "alias": "gpt-probe"}`,
		},

		// **Der Sitzungsschalter des Gerätemodells** — AI-05, #2042. `POST`,
		// weil er einen Stand des Dienstes ändert.
		{
			name: "das Gerätemodell für die Sitzung abschalten", methode: http.MethodPost,
			weg: "/ki/geraetemodell/sitzung", erfolg: http.StatusOK,
			rumpf: `{"an": false}`,
		},
	}
}
