package fenster

// **Die Abnahme am nativen Fenster** — ENG-02.AK-04
// (#1727),
// die Menüleiste aus #255 und die FIX-Abnahmen #1746–#1748.
//
// # Warum es das gibt
//
// **Was nur am nativen Fenster zu sehen ist, blieb offen, bis ein Mensch es
// ansah** — die Menüleiste, die Kürzel, die Dateidialoge. Ein Werkzeug von
// außen kommt nicht hinein: `AXIsProcessTrusted()` ist für die Prozesse
// dieser Arbeitsumgebung falsch, und ohne Bedienungshilfen gibt es weder
// einen fremden Zugänglichkeitsbaum noch `CGEvent`
// (`docs/HANDPRUEFUNGEN.md`). **Von innen geht es**: die Anwendung darf ihre
// eigene Leiste lesen, ihre eigenen Einträge auslösen und Tastendrücke in
// ihre eigene Ereignisschlange legen — dieselben Wege, die ein Klick und ein
// Tastendruck nehmen, ab dem Punkt, an dem AppKit sie in der Hand hat.
//
// # Was hier steht und was nicht
//
// **Hier steht die Sprache der Leitung** — Befehl, Antwort, Tasten, Pfade in
// der Leiste — und sie ist geprüft wie alles in diesem Paket. **Die Griffe
// in AppKit stehen in `abnahme_*darwin.go`**, nur mit dem Bauschalter
// `abnahme`: ein Freigabebau trägt sie nie (`scripts/releaseschalter.sh`
// lässt nur `production` zu). **Ohne `MSE_ABNAHME` startet ein Abnahmebau
// gar nicht** ([abnahmeGrund]), und mit ihr nimmt Wails' MCP-Server nur
// Anfragen mit dem Schlüssel an, den allein die Leitung herausgibt
// ([mcpSchluessel]).
//
// **Was die Leitung nicht belegt**: dass ein Mensch den Eintrag findet, dass
// die Leiste gut aussieht, und alles, was der Fensterserver vor AppKit tut —
// eine Maus, die über die Leiste fährt, ein Ziehen aus dem Finder. Das bleibt
// eine Handprüfung und steht dort.

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"
	"unicode/utf8"
)

// Abnahmebefehl ist eine Zeile auf der Leitung.
type Abnahmebefehl struct {
	// Was ist einer von menue, klick, imfenster, taste, klicke, ziehe,
	// loslassen, bewege, rechtsklick, rufbuch, stoere, aktivieren, blatt,
	// blattordner, blattantwort, dialogantwort, ablegen, aufklappen,
	// zwischenablage, schluessel, fenster, beenden.
	Was string `json:"was"`

	// Pfad benennt einen Eintrag der Leiste über seine Beschriftungen.
	Pfad []string `json:"pfad,omitempty"`

	// Taste und Mit sind ein Tastendruck ([tasteAus]).
	Taste string   `json:"taste,omitempty"`
	Mit   []string `json:"mit,omitempty"`

	// X und Y sind die Stelle eines Klicks, in Punkten der Seite — so, wie
	// `getBoundingClientRect` sie nennt.
	X float64 `json:"x,omitempty"`
	Y float64 `json:"y,omitempty"`

	// Anzahl macht aus `klicke` einen Doppel- oder Dreifachklick; Mit gilt
	// dort als Umschalter am Klick (P4, #2372).
	Anzahl int `json:"anzahl,omitempty"`

	// Datei ist die Antwort an einen Dateidialog; Abbrechen schließt ihn. Bei
	// `speicher` ist sie der Ort des Heap-Profils (E06).
	Datei     string `json:"datei,omitempty"`
	Abbrechen bool   `json:"abbrechen,omitempty"`

	// Weg sind die Punkte eines Zugs — der erste drückt, der letzte lässt
	// los, es sei denn, Halten hält die Taste unten (A04, #2218).
	Weg    []Ort `json:"weg,omitempty"`
	Halten bool  `json:"halten,omitempty"`

	// Leeren leert das Rufbuch nach dem Abzug (B01, #2220).
	Leeren bool `json:"leeren,omitempty"`

	// Stoerungen verzögern Bindungen oder lassen sie fachlich scheitern
	// (B03, #2222) — leer hebt sie auf.
	Stoerungen []Stoerung `json:"stoerungen,omitempty"`
}

// Abnahmeantwort ist, was zurückgeht — eine Zeile je Befehl.
type Abnahmeantwort struct {
	OK     bool   `json:"ok"`
	Daten  any    `json:"daten,omitempty"`
	Fehler string `json:"fehler,omitempty"`
}

// Menuepunkt ist ein Eintrag der **echten** Leiste, so wie AppKit ihn hält —
// nicht, wie `internal/app` ihn gebaut hat.
type Menuepunkt struct {
	Titel   string `json:"titel"`
	Kuerzel string `json:"kuerzel,omitempty"`
	Aktiv   bool   `json:"aktiv"`
	Trenner bool   `json:"trenner,omitempty"`

	// Verborgen ist ein Eintrag, den die Leiste nicht zeigt — etwa die
	// Wahltaste-Fassung eines anderen. Gefunden wird er nicht.
	Verborgen bool         `json:"verborgen,omitempty"`
	Aktion    string       `json:"aktion,omitempty"`
	Dienste   bool         `json:"dienste,omitempty"`
	Unter     []Menuepunkt `json:"unter,omitempty"`
}

// Tastendruck ist, was AppKit für ein `NSEvent` braucht.
type Tastendruck struct {
	Code     uint16
	Zeichen  string
	Ohne     string
	Schalter uint64
}

// stapelwort legt `isARepeat` und `keyCode` in **ein** Wort, wie Apples
// arm64-Aufrufkonvention sie auf den Stapel legt: Argumente dort stehen nach
// ihrer eigenen Größe und Ausrichtung, nicht je acht Bytes — das `BOOL` an
// Stelle 0, der `unsigned short` an Stelle 2 (*Writing ARM64 code for Apple
// platforms*, Abschnitt über Stapelargumente). `keyEventWithType:…` hat zwölf
// Argumente samt Empfänger und Selektor; diese zwei sind die einzigen, für die
// kein Register bleibt.
//
// **Warum hier und nicht in der Bibliothek**: `objc.Send` reicht seine
// Argumente variadisch an purego weiter, und purego packt nur einen
// **typisierten** Aufruf nach dieser Konvention (`func.go`,
// `shouldBundleStackArgs`, purego v0.11.0-alpha.6 und unverändert in v0.11.1 —
// geändert hat sich dort nur, welche Felder einer Struktur zählen); variadisch bekommt jedes
// sein eigenes Wort. Gemessen am Fenster: jede Taste der Probe kam als
// `KeyA` an, und ⌘Q beendete nach einer Eingabe nicht mehr.
func stapelwort(wiederholung bool, code uint16) uint64 {
	wort := uint64(code) << 16
	if wiederholung {
		wort |= 1
	}

	return wort
}

// Die Umschalter aus `NSEvent.h` (`NSEventModifierFlags`), nachgelesen im
// macOS-SDK der Befehlszeilenwerkzeuge.
const (
	schalterUmschalt uint64 = 1 << 17
	schalterCtrl     uint64 = 1 << 18
	schalterWahl     uint64 = 1 << 19
	schalterBefehl   uint64 = 1 << 20
)

// umschalter ordnet die Namen der Leitung den Umschaltern zu.
var umschalter = map[string]uint64{
	"befehl":   schalterBefehl,
	"umschalt": schalterUmschalt,
	"wahl":     schalterWahl,
	"ctrl":     schalterCtrl,
}

// tastencodes sind die virtuellen Tastencodes aus `HIToolbox/Events.h`
// (`kVK_ANSI_*` und die benannten), nachgelesen im macOS-SDK. **Ein Code
// gehört zur Taste, nicht zum Zeichen**: AppKit liest beim Kürzel das
// Zeichen, WebKit beim Tippen beides.
var tastencodes = map[string]uint16{
	"a": 0x00, "s": 0x01, "d": 0x02, "f": 0x03, "h": 0x04, "g": 0x05,
	"z": 0x06, "x": 0x07, "c": 0x08, "v": 0x09, "b": 0x0B, "q": 0x0C,
	"w": 0x0D, "e": 0x0E, "r": 0x0F, "y": 0x10, "t": 0x11, "1": 0x12,
	"2": 0x13, "3": 0x14, "4": 0x15, "6": 0x16, "5": 0x17, "9": 0x19,
	"7": 0x1A, "8": 0x1C, "0": 0x1D, "o": 0x1F, "u": 0x20, "i": 0x22,
	"p": 0x23, "l": 0x25, "j": 0x26, "k": 0x28, ",": 0x2B, "n": 0x2D,
	"m": 0x2E, ".": 0x2F, "-": 0x1B, "/": 0x2C, ";": 0x29, "'": 0x27,
	"=": 0x18, "[": 0x21, "]": 0x1E, "\\": 0x2A, "`": 0x32,
	"Enter": 0x24, "Tab": 0x30, " ": 0x31, "Backspace": 0x33,
	"Escape": 0x35, "Delete": 0x75, "ArrowLeft": 0x7B, "ArrowRight": 0x7C,
	"ArrowDown": 0x7D, "ArrowUp": 0x7E, "F10": 0x6D,
	"PageUp": 0x74, "PageDown": 0x79, "Home": 0x73, "End": 0x77,
}

// benannteZeichen sind die Zeichen der benannten Tasten — die Pfeile aus
// `NSEvent.h` (`NSUpArrowFunctionKey` …), die übrigen als Steuerzeichen, wie
// AppKit sie in `characters` liefert.
var benannteZeichen = map[string]string{
	"Enter": "\r", "Tab": "\t", "Backspace": "\x7f", "Escape": "\x1b",
	"Delete": "\uf728", "ArrowUp": "\uf700", "ArrowDown": "\uf701",
	"ArrowLeft": "\uf702", "ArrowRight": "\uf703", "F10": "\uf70d",
	"PageUp": "\uf72c", "PageDown": "\uf72d", "Home": "\uf729", "End": "\uf72b",
}

var errUnbekannteTaste = errors.New("unbekannte Taste")

// tasteAus übersetzt einen Tastendruck der Leitung.
//
// **Ein Zeichen ohne eigene Taste geht als Zeichen**: ein Umlaut oder ein
// Doppelpunkt hat auf einer deutschen Tastatur eine andere Stelle als auf
// einer amerikanischen, und WebKit fügt ein, was in `characters` steht. Der
// Code bleibt dann null — benannte Tasten ohne Code sind ein Fehler.
//
// **Mit Umschalt wird ein Buchstabe groß** — in `characters`, nicht in
// `charactersIgnoringModifiers`: so liefert AppKit ihn, und ⇧⌘S ist nur so
// von ⌘S zu unterscheiden.
func tasteAus(taste string, mit []string) (Tastendruck, error) {
	code, bekannt := tastencodes[taste]
	if !bekannt && utf8.RuneCountInString(taste) != 1 {
		return Tastendruck{}, fmt.Errorf("%w: %q", errUnbekannteTaste, taste)
	}

	var schalter uint64

	for _, name := range mit {
		wert, da := umschalter[name]
		if !da {
			return Tastendruck{}, fmt.Errorf("%w: Umschalter %q", errUnbekannteTaste, name)
		}

		schalter |= wert
	}

	zeichen, benannt := benannteZeichen[taste]
	if !benannt {
		zeichen = taste
	}

	ohne := zeichen
	if schalter&schalterUmschalt != 0 {
		zeichen = strings.ToUpper(zeichen)
	}

	return Tastendruck{Code: code, Zeichen: zeichen, Ohne: ohne, Schalter: schalter}, nil
}

// kuerzelText schreibt ein Kürzel so, wie macOS es in der Leiste zeigt — in
// der Reihenfolge ⌃⌥⇧⌘.
func kuerzelText(aequivalent string, schalter uint64) string {
	if aequivalent == "" {
		return ""
	}

	var text strings.Builder

	for _, stelle := range []struct {
		wert    uint64
		zeichen string
	}{
		{schalterCtrl, "⌃"}, {schalterWahl, "⌥"}, {schalterUmschalt, "⇧"}, {schalterBefehl, "⌘"},
	} {
		if schalter&stelle.wert != 0 {
			text.WriteString(stelle.zeichen)
		}
	}

	text.WriteString(strings.ToUpper(aequivalent))

	return text.String()
}

var errKeinEintrag = errors.New("kein solcher Eintrag")

// pfadIn sucht einen Eintrag über seine Beschriftungen und gibt die Stellen
// je Ebene. **Die Fehlermeldung nennt, was dort steht** — wer einen Titel
// falsch schreibt, sieht die richtigen, statt zu raten.
func pfadIn(punkte []Menuepunkt, pfad []string) ([]int, error) {
	if len(pfad) == 0 {
		return nil, fmt.Errorf("%w: leerer Pfad", errKeinEintrag)
	}

	stellen := make([]int, 0, len(pfad))
	ebene := punkte

	for tiefe, titel := range pfad {
		stelle := slices.IndexFunc(ebene, func(punkt Menuepunkt) bool {
			return sichtbar(punkt) && punkt.Titel == titel
		})
		if stelle < 0 {
			return nil, fmt.Errorf("%w: %q in Ebene %d, dort stehen %s",
				errKeinEintrag, titel, tiefe+1, titelListe(ebene))
		}

		stellen = append(stellen, stelle)
		ebene = ebene[stelle].Unter
	}

	return stellen, nil
}

// sichtbar ist ein Eintrag, den ein Mensch in der Leiste sieht und klicken
// kann — kein Trenner und nicht verborgen.
func sichtbar(punkt Menuepunkt) bool {
	return !punkt.Trenner && !punkt.Verborgen
}

// titelListe nennt die sichtbaren Beschriftungen einer Ebene.
func titelListe(punkte []Menuepunkt) string {
	titel := make([]string, 0, len(punkte))

	for _, punkt := range punkte {
		if sichtbar(punkt) {
			titel = append(titel, "»"+punkt.Titel+"«")
		}
	}

	return strings.Join(titel, ", ")
}

// befehlLesen liest eine Zeile der Leitung. **Unbekannte Felder sind ein
// Fehler**: ein vertippter Name wäre sonst ein Befehl ohne Wirkung, der wie
// einer mit Wirkung aussieht.
func befehlLesen(zeile []byte) (Abnahmebefehl, error) {
	var befehl Abnahmebefehl

	leser := json.NewDecoder(strings.NewReader(string(zeile)))
	leser.DisallowUnknownFields()

	if fehler := leser.Decode(&befehl); fehler != nil {
		return Abnahmebefehl{}, fmt.Errorf("die Zeile ist kein Befehl: %w", fehler)
	}

	return befehl, nil
}

// antwortAus macht aus Ergebnis und Fehler eine Zeile.
func antwortAus(daten any, fehler error) Abnahmeantwort {
	if fehler != nil {
		return Abnahmeantwort{Fehler: fehler.Error()}
	}

	return Abnahmeantwort{OK: true, Daten: daten}
}

// ersatzablage ist die Zwischenablage des Abnahmebaus — **nicht die des
// Rechners**.
//
// # Warum es sie gibt
//
// Kernweg 6 aus ENG-02.AK-04 kopiert Karten und setzt sie in ein anderes Set
// ein. Über die echte Zwischenablage griffe die Probe in die des Menschen an
// diesem Rechner: sie überschriebe, was er kopiert hat, und läse, was er
// nicht teilen wollte. **Deshalb geht die Naht im Abnahmebau hierher**
// (`wailshuelle.ablageLesen`, `ablageSchreiben`), so wie ein Dateidialog eine
// vorbereitete Antwort bekommt. **Belegt ist damit der Weg bis zur Naht und
// von ihr zurück — Knopf, Bindung, Anwendung, Bestand —, nicht
// `NSPasteboard` selbst**; das ist Wails' `Clipboard` und eine Zeile.
type ersatzablage struct {
	sync.Mutex

	text string
}

func (e *ersatzablage) lesen() string {
	e.Lock()
	defer e.Unlock()

	return e.text
}

func (e *ersatzablage) schreiben(text string) {
	e.Lock()
	defer e.Unlock()

	e.text = text
}

// abnahmeGrund sagt, warum ein Abnahmebau nicht starten darf. **Ohne Leitung
// ist er kein Werkzeug der Probe**, sondern ein Programm mit einem offenen
// MCP-Server — Wails startet ihn in `App.Run` von selbst, sobald der Schalter
// `mcp` gesetzt ist — und mit dem echten Bestand (N1 des Nachlaufs über
// Phase 6.3).
func abnahmeGrund(ort string) error {
	if ort == "" {
		return errors.New("ein Abnahmebau startet nur aus der Hüllenprobe — MSE_ABNAHME fehlt")
	}

	return nil
}

// mcpSchluessel ist der Schlüssel für Wails' MCP-Server: 32 Bytes aus der
// Quelle, hexadezimal. **Ohne ihn nimmt der Server jede Anfrage von diesem
// Rechner an** (`authorizedMCPRequest`, Wails v3.0.0-beta.23); mit ihm nur
// die, die ihn kennt — und erfährt ihn nur, wer die Leitung öffnen darf.
func mcpSchluessel(quelle io.Reader) (string, error) {
	roh := make([]byte, 32)
	if _, fehler := io.ReadFull(quelle, roh); fehler != nil {
		return "", fmt.Errorf("der Schlüssel des MCP-Servers: %w", fehler)
	}

	return hex.EncodeToString(roh), nil
}
