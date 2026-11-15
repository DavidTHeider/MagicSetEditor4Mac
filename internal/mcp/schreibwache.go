package mcp

import (
	"context"
	"encoding/json"
	"net/http"
)

// MitSchreibwache setzt die Frage, die jedes **schreibende** Werkzeug vor
// seinem Lauf stellt — ID-03.AK-02
// (#1702).
//
// # Warum es sie gibt
//
// **Der Server hält seine Verbindung über seine ganze Laufzeit**, aber jeder
// Ruf arbeitet im Profil, das der Bestand **gerade** zeigt. Wer nach dem
// Start als Leser in ein fremdes Profil wechselte oder dem eine Freigabe
// entzogen wurde, schrieb über die schreibende Verbindung vom Start weiter —
// mit Rechten, die er nicht mehr hatte.
//
// **Die Frage stellt die Wurzel, nicht dieses Blatt**: `internal/mcp` weiß
// nicht, was ein Akteur ist. `cmd/mse-mcp` baut ihn je Ruf frisch und fragt,
// ob er schreiben darf.
//
// # Was sie nicht fragt
//
// **Lesende Werkzeuge** — sie schreiben nicht. **Und Werkzeuge, die den
// Zeiger bewegen** ([Werkzeug.Zeigerwechsel]): wer nur noch lesen darf, muss
// hinauskommen.
func (s *Server) MitSchreibwache(wache func(ctx context.Context) error) *Server {
	s.schreibwache = wache

	return s
}

// schreibenVerwehrt sagt, ob dieses Werkzeug hier nicht laufen darf.
func (s *Server) schreibenVerwehrt(ctx context.Context, werkzeug Werkzeug) error {
	if s.schreibwache == nil || werkzeug.Art != Schreibend || werkzeug.Zeigerwechsel {
		return nil
	}

	return s.schreibwache(ctx)
}

// MitSchemafreigabe setzt die Freigabe, die **jeder** Ruf vor allem anderen
// holt und bis zu seinem Ende hält — Zweitblick über Phase 2 von 0.9.0,
// Befunde 1 und 2 (#2323). Der Grund steht am Dienst
// (`rest.Dienst.MitSchemafreigabe`): dieselbe Lage, dieselbe Antwort.
func (s *Server) MitSchemafreigabe(freigabe func(ctx context.Context) (func(), error)) *Server {
	s.schemafreigabe = freigabe

	return s
}

// freigegeben markiert einen Ruf, der seine Freigabe schon hält — über HTTP
// holt sie die Anfrage, bevor sie den Bestand liest (`ueberhttp.go`).
type freigegeben struct{}

// freigabeHolen holt die Freigabe — ohne eingesetzte gibt es nichts zu
// halten, und ein Ruf unter einer gehaltenen holt sie nicht ein zweites Mal.
func (s *Server) freigabeHolen(ctx context.Context) (frei func(), err error) {
	if s.schemafreigabe == nil || ctx.Value(freigegeben{}) != nil {
		return func() {}, nil
	}

	return s.schemafreigabe(ctx)
}

// unterFreigabe holt die Freigabe für eine Anfrage über HTTP und gibt den
// Kontext, der sie trägt — oder schreibt das Nein (Nachlauf über Phase 2 von
// 0.9.0, Befund 3: `prozessTaugt` las Profil und Schalter vor ihr).
func (s *Server) unterFreigabe(schreiben http.ResponseWriter, anfrage *http.Request) (*http.Request, func(), bool) {
	frei, verwehrt := s.freigabeHolen(anfrage.Context())
	if verwehrt != nil {
		schreiben.Header().Set("Content-Type", "application/json")
		schreiben.Header().Set("Retry-After", "1")
		schreiben.WriteHeader(http.StatusServiceUnavailable)

		gesagt, _ := json.Marshal(map[string]any{ //nolint:errcheck // Zeichenketten
			"error": map[string]any{"code": "belegt", "message": fehlertext(verwehrt)},
		})
		_, _ = schreiben.Write(gesagt) //nolint:errcheck // die Verbindung ist ohnehin am Ende

		return nil, nil, false
	}

	return anfrage.WithContext(context.WithValue(anfrage.Context(), freigegeben{}, true)), frei, true
}
