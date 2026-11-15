package mcp_test

// **Die Gruppen werden zu einer Tabelle** — abgetrennt am 2026-09-11, weil
// `zusagenliste_test.go` mit der siebzehnten Gruppe über die Zeilengrenze aus
// §2 ging (A5, #1444).
//
// # Der Schnitt folgt der Frage
//
// Dort stehen **die Zusagen** — je Gruppe, warum sie eine eigene ist und was
// sie nicht sagt. Hier steht, **wie aus vielen eine wird**: drei Zeilen
// Rechnung, die mit dem Inhalt nichts zu tun haben.

// zusammengefasst legt die Gruppen zu einer Tabelle zusammen.
//
// **Der Aufrufer fragt „ist dieses Werkzeug zugesagt"**, und die Antwort hängt
// nicht davon ab, unter welcher Gruppe — die sind für den **Leser** getrennt,
// nicht für die Rechnung.
func zusammengefasst(gruppen ...map[string]bool) map[string]bool {
	alle := map[string]bool{}

	for _, gruppe := range gruppen {
		for name := range gruppe {
			alle[name] = true
		}
	}

	return alle
}
