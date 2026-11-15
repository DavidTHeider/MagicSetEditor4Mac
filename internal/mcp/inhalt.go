package mcp

import "context"

// **Herausgelöst aus `server.go`**, als der Pfadvertrag (OPS-02.AK-04,
// #1719) die Datei über die Zeilengrenze aus §2 schob.

// inhalt baut das Ergebnis eines Werkzeugaufrufs.
func inhalt(text string, fehler bool) map[string]any {
	return map[string]any{
		"content": []map[string]any{{"type": "text", "text": text}},
		"isError": fehler,
	}
}

// nachgeladenerInhalt baut das Ergebnis eines Rufs, der geschrieben hat und
// nur nicht nachlesen konnte (#1970): **zwei Teile, und kein Fehler**.
func nachgeladenerInhalt(nachgeladen *Nachgeladen) map[string]any {
	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": nachgeladen.Ergebnis},
			{"type": "text", "text": nachgeladen.Error()},
		},
		"isError": false,
	}
}

// fehlerInhalt baut das Ergebnis eines gescheiterten Rufs: den Satz mit
// Vorgangskennung wie bisher **und**, wenn die Wurzel eine Fehlerform
// eingesetzt hat, dieselbe Angabe maschinenlesbar in `structuredContent`
// (F03, #2342). **Kein Werkzeug erklärt ein `outputSchema`**, an das sich
// die Form halten müsste — sie ist eine Zugabe, und der Satz bleibt.
func (s *Server) fehlerInhalt(ctx context.Context, fehler error, text string) map[string]any {
	ergebnis := inhalt(s.mitKennung(ctx, text), true)

	if s.fehlerform != nil {
		ergebnis["structuredContent"] = map[string]any{"fehler": s.fehlerform(ctx, fehler)}
	}

	return ergebnis
}

// MitFehlerform setzt, wie ein Fehler maschinenlesbar aussieht — die Wurzel
// setzt die Übersetzung aus `internal/app` ein, damit MCP dieselbe Klasse
// nennt wie Fenster, Kommandozeile und Dienst.
func (s *Server) MitFehlerform(form func(context.Context, error) any) *Server {
	s.fehlerform = form

	return s
}
