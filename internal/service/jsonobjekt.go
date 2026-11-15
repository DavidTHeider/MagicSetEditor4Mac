package service

// **Ein Eintrag in einem JSON-Objekt, ohne die Ordnung daneben zu
// verlieren** — OPS-09b
// (#1961).
//
// # Warum nicht `encoding/json` mit einer Karte
//
// **Weil eine Karte die Reihenfolge der Schlüssel vergisst.** Wer
// `claude_desktop_config.json` über `map[string]any` liest und schreibt,
// sortiert jedem Menschen seine Datei um, und ein Widerruf fände den Stand
// davor nicht wieder. Hier wird je Ebene eine Liste von Paaren gelesen; was
// nicht angefasst wird, reist als rohes JSON mit.
//
// # Was es nicht verspricht
//
// **Die Einrückung.** Geschrieben wird mit zwei Leerzeichen; byte-gleich
// zurück geht es nur über den gemerkten Stand davor (`schnittstellenjson.go`).

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// errKeinObjekt ist ein Dokument, das kein Objekt der erwarteten Gestalt ist.
var errKeinObjekt = errors.New("die Datei ist kein JSON-Objekt dieser Gestalt")

// paar ist ein Schlüssel mit seinem rohen Wert.
type paar struct {
	schluessel string
	wert       json.RawMessage
}

// objektLesen liest ein Objekt als Paare — ein leeres Dokument als leeres.
func objektLesen(roh []byte) ([]paar, error) {
	if len(bytes.TrimSpace(roh)) == 0 {
		return nil, nil
	}

	leser := json.NewDecoder(bytes.NewReader(roh))

	if anfang, fehler := leser.Token(); fehler != nil || anfang != json.Delim('{') {
		return nil, errKeinObjekt
	}

	var paare []paar

	for leser.More() {
		marke, fehler := leser.Token()

		schluessel, istText := marke.(string)
		if fehler != nil || !istText || doppelt(paare, schluessel) {
			return nil, errKeinObjekt
		}

		var wert json.RawMessage
		if fehler := leser.Decode(&wert); fehler != nil {
			return nil, errKeinObjekt
		}

		paare = append(paare, paar{schluessel: schluessel, wert: wert})
	}

	if ende, fehler := leser.Token(); fehler != nil || ende != json.Delim('}') {
		return nil, errKeinObjekt
	}

	// **Nach dem Objekt kommt nur das Ende** — ein kaputter Nachlauf gibt
	// einen Syntaxfehler und nicht `io.EOF`, und beides als Ende zu nehmen
	// hieß, ihn beim Umschreiben verschwinden zu lassen (Zweitblick über
	// Phase 6.12, Z8).
	if _, fehler := leser.Token(); !errors.Is(fehler, io.EOF) {
		return nil, errKeinObjekt
	}

	return paare, nil
}

// doppelt sagt, ob ein Schlüssel schon dasteht — zwei gleiche hießen, dass
// niemand weiß, welcher gilt.
func doppelt(paare []paar, schluessel string) bool {
	for _, einer := range paare {
		if einer.schluessel == schluessel {
			return true
		}
	}

	return false
}

// objektSchreiben schreibt Paare als eingerücktes Objekt.
func objektSchreiben(paare []paar) []byte {
	var roh bytes.Buffer

	roh.WriteByte('{')

	for stelle, einer := range paare {
		if stelle > 0 {
			roh.WriteByte(',')
		}

		name, _ := json.Marshal(einer.schluessel) //nolint:errcheck // eine Zeichenkette lässt sich immer schreiben
		roh.Write(name)
		roh.WriteByte(':')
		roh.Write(einer.wert)
	}

	roh.WriteByte('}')

	var eingerueckt bytes.Buffer
	_ = json.Indent(&eingerueckt, roh.Bytes(), "", "  ") //nolint:errcheck // aus gelesenen Werten gebaut, also gültig

	eingerueckt.WriteByte('\n')

	return eingerueckt.Bytes()
}

// innenLesen liest das innere Objekt unter `aussen` — fehlt es, ist es leer.
func innenLesen(paare []paar, aussen string) ([]paar, int, error) {
	for stelle, einer := range paare {
		if einer.schluessel == aussen {
			innen, fehler := objektLesen(einer.wert)
			if fehler != nil || len(bytes.TrimSpace(einer.wert)) == 0 {
				return nil, 0, fmt.Errorf("%w: %q ist kein Objekt", errKeinObjekt, aussen)
			}

			return innen, stelle, nil
		}
	}

	return nil, -1, nil
}

// jsonEintrag liest `mcpServers.name` — der zweite Wert sagt, ob er dasteht.
func jsonEintrag(roh []byte, name string) (wert json.RawMessage, da bool, fehler error) {
	paare, fehler := objektLesen(roh)
	if fehler != nil {
		return nil, false, fehler
	}

	innen, _, fehler := innenLesen(paare, mcpObjekt)
	if fehler != nil {
		return nil, false, fehler
	}

	for _, einer := range innen {
		if einer.schluessel == name {
			return einer.wert, true, nil
		}
	}

	return nil, false, nil
}

// jsonEintragSetzen setzt `mcpServers.name` auf `wert` — an seiner Stelle,
// wenn es ihn gibt, sonst am Ende.
func jsonEintragSetzen(roh []byte, name string, wert json.RawMessage) ([]byte, error) {
	return jsonInnenAendern(roh, func(innen []paar) []paar {
		for stelle, einer := range innen {
			if einer.schluessel == name {
				innen[stelle].wert = wert

				return innen
			}
		}

		return append(innen, paar{schluessel: name, wert: wert})
	})
}

// jsonEintragEntfernen nimmt `mcpServers.name` heraus — und sonst nichts.
func jsonEintragEntfernen(roh []byte, name string) ([]byte, error) {
	paare, fehler := objektLesen(roh)
	if fehler != nil {
		return nil, fehler
	}

	if _, stelle, fehler := innenLesen(paare, mcpObjekt); fehler != nil {
		return nil, fehler
	} else if stelle < 0 {
		return objektSchreiben(paare), nil
	}

	return jsonInnenAendern(roh, func(innen []paar) []paar {
		uebrig := innen[:0]

		for _, einer := range innen {
			if einer.schluessel != name {
				uebrig = append(uebrig, einer)
			}
		}

		return uebrig
	})
}

// jsonInnenAendern ändert `mcpServers` und schreibt das ganze Objekt.
func jsonInnenAendern(roh []byte, aendern func([]paar) []paar) ([]byte, error) {
	paare, fehler := objektLesen(roh)
	if fehler != nil {
		return nil, fehler
	}

	innen, stelle, fehler := innenLesen(paare, mcpObjekt)
	if fehler != nil {
		return nil, fehler
	}

	neu := json.RawMessage(objektSchreiben(aendern(innen)))

	if stelle < 0 {
		return objektSchreiben(append(paare, paar{schluessel: mcpObjekt, wert: neu})), nil
	}

	paare[stelle].wert = neu

	return objektSchreiben(paare), nil
}
