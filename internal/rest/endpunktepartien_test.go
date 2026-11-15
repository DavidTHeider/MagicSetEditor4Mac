// **Die Endpunkte der Partien** — Werkstück G1 von Phase 5.5
// (#1458).
//
// **Aus `endpunktliste_test.go` herausgelöst**, als sie über 400 Zeilen ging
// (§2). Der Schnitt folgt der Frage: dort stehen die Wege am Bestand, hier
// die einer laufenden Partie.
package rest_test

import "net/http"

// endpunktePartien sind die zehn Wege einer Partie.
//
// **Die sechs ersten tragen eine Sitzkennung** — ohne sie gibt es die Frage
// „wer sieht das?" nicht, und die Projektion wüsste nicht, was sie kürzen
// soll.
//
// **Die vier der Verwaltung tragen keine** (PLAY-03.AK-05, #1711): sie
// betreffen die Partie als Ganzes und nicht ihre Sicht. Wer eine löscht,
// löscht sie für jeden; wer eine ausführt, bekommt den **vollen** Zustand.
func endpunktePartien() []endpunkt {
	return []endpunkt{
		{
			name: "eine Partie anlegen", methode: http.MethodPost,
			weg: "/partien",
			rumpf: `{"spiel":"magic","sitze":[` +
				`{"kennung":"sitz-1","name":"A","deckID":"1"},` +
				`{"kennung":"sitz-2","name":"B","deckID":"2"}]}`,
			erfolg: http.StatusCreated,
		},
		{
			name: "eine Partie ansehen", methode: http.MethodGet,
			weg: "/partien/p1?fuer=sitz-1", erfolg: http.StatusOK,
		},
		{
			// **Die eigenen Partien** (#1600)
			// — **ohne** Sitzkennung, und das ist der Unterschied zu allen
			// anderen hier: eine Liste ist keine Sicht. Sie nennt Sitze,
			// Stand und die Zahl der Objekte, und welche Partien es gibt,
			// hängt am Profil und nicht an einem Sitz.
			name: "die eigenen Partien nennen", methode: http.MethodGet,
			weg: "/partien", erfolg: http.StatusOK,
		},
		{
			// **Der Verlauf** (G5, #1462) — dieselbe Sitzkennung in der
			// Abfrage wie beim Ansehen, und aus demselben Grund: ohne sie
			// gäbe es die Frage „wer sieht das?" nicht.
			name: "den Verlauf einer Partie ansehen", methode: http.MethodGet,
			weg: "/partien/p1/verlauf?fuer=sitz-1", erfolg: http.StatusOK,
		},
		{
			name: "einen Spielbefehl stellen", methode: http.MethodPost,
			weg: "/partien/p1/befehle",
			rumpf: `{"absender":"sitz-1","kennung":"b1",` +
				`"art":"objekt.tappen"}`,
			erfolg: http.StatusOK,
		},
		{
			name: "den Stand einer Partie wechseln", methode: http.MethodPost,
			weg:    "/partien/p1/stand",
			rumpf:  `{"fuer":"sitz-1","stand":"pausieren"}`,
			erfolg: http.StatusOK,
		},
		{
			name: "eine Partie löschen", methode: http.MethodDelete,
			weg: "/partien/p1", erfolg: http.StatusOK,
		},
		{
			name: "eine Partie kopieren", methode: http.MethodPost,
			weg: "/partien/p1/kopie", erfolg: http.StatusCreated,
		},
		{
			name: "eine Partie ausführen", methode: http.MethodGet,
			weg: "/partien/p1/ausfuhr", erfolg: http.StatusOK,
		},
		{
			name: "eine Partie einlesen", methode: http.MethodPost,
			weg:    "/partien/einfuhr",
			rumpf:  `{"inhalt":"{\"fassung\":1}"}`,
			erfolg: http.StatusCreated,
		},
	}
}
