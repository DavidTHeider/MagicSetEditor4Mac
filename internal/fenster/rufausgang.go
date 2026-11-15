package fenster

// **Wie das Rufbuch eine Antwort einordnet** — B01
// (#2220).
//
// **Abgewiesen** (die Laufzeit meldet einen Fehler — eine Panik des Tors,
// ein unbekannter Name), **abgebrochen** (eine überholte Vorschau, ein Lauf,
// den jemand anhielt), **fachfehler** (die Bindung antwortet mit einem anderen
// gesetzten `fehler`), **gelungen** — oder **unbekannt**.
//
// # Der ganze Strom, nicht Anfang und Ende
//
// Bis zum Zweitblick über Phase B (Befund 4) galt als gelungen, was sich
// nicht lesen ließ: eine Antwort über 64 KiB war abgeschnitten, und ihr
// `fehler` — bei den Antworten der Bindungen das letzte Feld — stand hinter
// der Schnittstelle. Der Nachlauf fand am Behelf mit Anfang und Ende zwei
// weitere Wege vorbei: ein verschachteltes `"fehler":` hinter dem echten und
// einen Anfang, der genau nach einem Feld endete. **Jetzt liest der
// [fehlerspaeher] jede Antwort Byte für Byte mit**: er zählt die Tiefe, kennt
// Zeichenketten samt Escapes und greift nur das Feld `fehler` der obersten
// Ebene. Eine Antwort, deren Wurzel nie schließt, ist unbekannt.
//
// # Was es nicht sieht
//
// **Ob das JSON gültig ist** — gezählt werden Klammern und Zeichenketten,
// geprüft wird die Syntax nur bei Antworten bis 64 KiB (`ausgangGanz`) und
// beim Wert des Fehlerfelds.

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// fehlerform ist, was ein Fehler der Fassade trägt und das Rufbuch braucht.
type fehlerform struct {
	Code       string `json:"code"`
	Schluessel string `json:"schluessel"`
}

// mitschreiber reicht die Antwort durch und behält Status, Anfang und Ende.
type mitschreiber struct {
	http.ResponseWriter

	status  int
	anfang  bytes.Buffer
	laenge  int
	spaeher fehlerspaeher
}

func (m *mitschreiber) WriteHeader(status int) {
	m.status = status
	m.ResponseWriter.WriteHeader(status)
}

func (m *mitschreiber) Write(daten []byte) (int, error) {
	if frei := rufgrenze - m.anfang.Len(); frei > 0 {
		m.anfang.Write(daten[:min(frei, len(daten))])
	}

	m.laenge += len(daten)
	m.spaeher.lies(daten)

	return m.ResponseWriter.Write(daten) //nolint:wrapcheck // durchgereicht, wie es kam
}

// ausgang ordnet die mitgeschriebene Antwort ein.
func (m *mitschreiber) ausgang() string {
	if m.status >= http.StatusBadRequest {
		return "abgewiesen"
	}

	if m.laenge <= rufgrenze {
		return ausgangGanz(m.anfang.Bytes())
	}

	return m.spaeher.ausgang()
}

// ausgangGanz ordnet eine vollständig gelesene Antwort ein.
func ausgangGanz(rumpf []byte) string {
	// **Eine Bindung, die nur einen Fehler zurückgibt** (`*Fehler`, etwa
	// `ExemplarAendern`), gibt ihn ohne Umschlag — und `null`, wenn es gelang.
	var antwort struct {
		Fehler *fehlerform `json:"fehler"`
		fehlerform
	}

	if json.Unmarshal(rumpf, &antwort) != nil {
		// **Eine Zeichenkette, eine Zahl, eine Liste** sind gelesen; nur was
		// gar kein JSON ist, bleibt unbekannt.
		if json.Valid(rumpf) {
			return "gelungen"
		}

		return "unbekannt"
	}

	fehler := antwort.Fehler
	if fehler == nil && antwort.Code != "" && antwort.Schluessel != "" {
		fehler = &antwort.fehlerform
	}

	return einordnen(fehler)
}

// fehlerAus liest den Wert eines Fehlerfelds — `null` heißt: kein Fehler.
func fehlerAus(wert json.RawMessage) (*fehlerform, bool) {
	var fehler *fehlerform
	if json.Unmarshal(wert, &fehler) != nil {
		return nil, false
	}

	return fehler, true
}

// einordnen sagt, was ein gelesener Fehler bedeutet.
func einordnen(fehler *fehlerform) string {
	switch {
	case fehler == nil:
		return "gelungen"
	case fehler.Code == "abgebrochen":
		return "abgebrochen"
	default:
		return "fachfehler"
	}
}

// fehlerspaeher liest eine Antwort im Vorbeigehen und behält von ihr nur, ob
// ihre Wurzel geschlossen ist und was im Feld `fehler` der obersten Ebene
// steht — höchstens [endgrenze] Bytes davon.
type fehlerspaeher struct {
	tiefe       int
	wurzel      byte // `{`, `[` oder 0 vor dem ersten Zeichen
	geschlossen bool

	inText, escape bool
	text           []byte // die letzte Zeichenkette der obersten Ebene
	schluessel     string // der Schlüssel, dessen Wert gerade folgt

	imFehler bool   // der Wert des Fehlerfelds wird gerade gelesen
	fehler   []byte // dieser Wert
	gefunden bool
	zuLang   bool
}

// lies nimmt das nächste Stück der Antwort.
func (f *fehlerspaeher) lies(daten []byte) {
	for _, zeichen := range daten {
		f.zeichen(zeichen)
	}
}

func (f *fehlerspaeher) zeichen(zeichen byte) {
	if f.imFehler {
		f.mitschreiben(zeichen)
	}

	if f.inText {
		f.imText(zeichen)

		return
	}

	switch zeichen {
	case '"':
		f.inText = true
		f.text = f.text[:0]
	case '{', '[':
		if f.wurzel == 0 {
			f.wurzel = zeichen
		}

		f.tiefe++
	case '}', ']':
		f.tiefe--
		// **Nur das Ende der Wurzel beendet den Wert** — das Schließen des
		// Fehlerobjekts selbst gehört noch zu ihm.
		if f.tiefe == 0 {
			f.geschlossen = true
			f.wertEnde()
		}
	case ':':
		if f.tiefe == 1 {
			f.schluessel = string(f.text)
			f.imFehler = f.schluessel == "fehler" && f.wurzel == '{'
		}
	case ',':
		if f.tiefe == 1 {
			f.wertEnde()
		}
	}
}

// imText liest ein Zeichen einer Zeichenkette — samt Escape.
func (f *fehlerspaeher) imText(zeichen byte) {
	switch {
	case f.escape:
		f.escape = false
	case zeichen == '\\':
		f.escape = true
	case zeichen == '"':
		f.inText = false

		return
	}

	if f.tiefe == 1 && len(f.text) < 64 {
		f.text = append(f.text, zeichen)
	}
}

// mitschreiben behält das Zeichen als Teil des Fehlerwerts.
func (f *fehlerspaeher) mitschreiben(zeichen byte) {
	if len(f.fehler) >= endgrenze {
		f.zuLang = true

		return
	}

	f.fehler = append(f.fehler, zeichen)
}

// wertEnde schließt den Fehlerwert — gerufen, wenn die oberste Ebene einen
// Wert beendet: mit einem Komma oder mit dem Ende der Wurzel.
func (f *fehlerspaeher) wertEnde() {
	if !f.imFehler {
		return
	}

	f.imFehler = false
	f.gefunden = true
	// **Das Komma oder die Klammer gehört nicht zum Wert** — und das Zeichen
	// wurde schon mitgeschrieben.
	f.fehler = bytes.TrimSpace(f.fehler[:len(f.fehler)-1])
}

// ausgang ordnet ein, was der Späher gesehen hat.
func (f *fehlerspaeher) ausgang() string {
	switch {
	case !f.geschlossen || f.zuLang:
		return "unbekannt"
	case !f.gefunden:
		return "gelungen"
	}

	fehler, gelesen := fehlerAus(f.fehler)
	if !gelesen {
		return "unbekannt"
	}

	return einordnen(fehler)
}
