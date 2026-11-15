package rest_test

// **Die Betriebswege des Dienstes** — Gesundheit, Metriken, Meta und
// Verträglichkeit.
//
// **Eine eigene Datei**, weil `endpunkteneben_test.go` mit ihnen über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort stehen die
// Wege **neben** dem Bestand, hier die über den **Betrieb** — was dieser
// Prozess ist, kann und getan hat.

import "net/http"

// endpunkteBetrieb sind die Wege, die über den Prozess Auskunft geben.
func endpunkteBetrieb() []endpunkt {
	return []endpunkt{
		// **Der Betriebszustand** (O9, #1433) — zwei Wege, und `/bereit`
		// antwortet mit dem Statuscode.
		{
			name: "Gesundheit", methode: http.MethodGet,
			weg: "/gesundheit", erfolg: http.StatusOK,
		},
		{
			name: "Bereit", methode: http.MethodGet,
			weg: "/bereit", erfolg: http.StatusOK,
		},

		// **Und die Metriken** (O9, #1433) — Zahlen mit endlich vielen Reihen.
		{
			name: "Metriken", methode: http.MethodGet,
			weg: "/metriken", erfolg: http.StatusOK,
		},
		{
			name: "Meta", methode: http.MethodGet,
			weg: "/meta", erfolg: http.StatusOK,
		},
		{
			name: "Verträglichkeit", methode: http.MethodGet,
			weg: "/vertraeglichkeit", erfolg: http.StatusOK,
		},

		// **Und die Rückkehren** (O7, #1431) — was zurückgehen kann, was eine
		// Sicherung kostet, und worüber nach einer Rückkehr nichts feststeht.
		{
			name: "Rücksprungarten", methode: http.MethodGet,
			weg: "/ruecksprung/arten", erfolg: http.StatusOK,
		},
		{
			name: "Rücksprunglage", methode: http.MethodGet,
			weg:    "/ruecksprung/lage?datei=/tmp/probe.msesicherung",
			erfolg: http.StatusOK,
		},
		{
			name: "Rückkehren", methode: http.MethodGet,
			weg: "/rueckkehren", erfolg: http.StatusOK,
		},
		{
			name: "Rückkehr quittieren", methode: http.MethodPost,
			weg: "/rueckkehren/1/quittung", erfolg: http.StatusOK,
		},

		// **Und die Datenpflege** (O8, #1432) — nachsehen und wegräumen.
		{
			name: "Pflegevorschau", methode: http.MethodGet,
			weg: "/pflege/vorschau", erfolg: http.StatusOK,
		},
		{
			name: "Pflegen", methode: http.MethodPost, weg: "/pflege",
			rumpf:  `{"erstellt":"2026-09-10T12:00:00Z","wahl":[]}`,
			erfolg: http.StatusOK,
		},

		// **Und die abgestufte Rücksetzung** (O4, #1428) — die Reihe,
		// nachsehen und ausführen.
		{
			name: "Rücksetzungsarten", methode: http.MethodGet,
			weg: "/ruecksetzung/arten", erfolg: http.StatusOK,
		},
		{
			name: "Rücksetzungsvorschau", methode: http.MethodGet,
			weg: "/ruecksetzung/layout/vorschau", erfolg: http.StatusOK,
		},
		{
			name: "Zurücksetzen", methode: http.MethodPost,
			weg: "/ruecksetzung/layout", rumpf: `{"name":""}`, erfolg: http.StatusOK,
		},

		// **Und der lokale Suchindex** (A15, #1514) — drei lesende Wege und
		// ein schreibender; `DELETE` auf denselben Weg bricht den Aufbau ab.
		{
			name: "Indexstand", methode: http.MethodGet,
			weg: "/index", erfolg: http.StatusOK,
		},
		{
			name: "Indexsuche", methode: http.MethodGet,
			weg: "/index/suche?frage=lightning", erfolg: http.StatusOK,
		},
		{
			name: "Indexaufbau", methode: http.MethodPost, weg: "/index/aufbau",
			rumpf:  `{"modell":"all-minilm","quelle":"/tmp/o.jsonl.gz","revision":"1"}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Indexabbruch", methode: http.MethodDelete,
			weg: "/index/aufbau", erfolg: http.StatusOK,
		},

		// **Und die Recherche darüber** (A16, #1515) — zwei lesende Wege
		// und einer, der wie ein Schreiben aussieht und keines ist.
		{
			name: "Recherche", methode: http.MethodGet,
			weg: "/recherche?fall=regeln&frage=trampeln", erfolg: http.StatusOK,
		},
		{
			name: "Ähnliche Karten", methode: http.MethodGet,
			weg: "/recherche/aehnlich/2", erfolg: http.StatusOK,
		},
		{
			name: "Kontextpaket", methode: http.MethodPost, weg: "/kontextpaket",
			rumpf: `{"aufgabe":"a","fall":"regeln","frage":"b",` +
				`"anbieter":"ollama","modell":"m","kontextgrenze":8192,` +
				`"ausgabereserve":1024}`,
			erfolg: http.StatusOK,
		},

		// **Und das Kostenjournal** (O12, #1439) — Stand, Posten, Grenze.
		{
			name: "Kostenstand", methode: http.MethodGet,
			weg: "/kosten/anbieter", erfolg: http.StatusOK,
		},
		{
			name: "Kostenposten", methode: http.MethodGet,
			weg: "/kosten/anbieter/posten", erfolg: http.StatusOK,
		},
		{
			name: "Kostenbudget", methode: http.MethodPut,
			weg:    "/kosten/anbieter/budget",
			rumpf:  `{"zeitraum":"tag","wert":5000,"waehrung":"EUR"}`,
			erfolg: http.StatusOK,
		},
	}
}
