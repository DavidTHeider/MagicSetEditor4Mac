package domain

// **Wogegen ein KI-Vorschlag geprüft wird** — Werkstück A6 von Phase 5.3
// (#1445).
//
// # Aus den Feldern des Pakets und nicht aus einer Liste daneben
//
// Das Issue verlangt es: „Kanonisches Schema aus **tatsächlichen
// Paketfeldern** inklusive Typ/Enum/Mehrfachfeld und Schreibbarkeit." Eine
// zweite Aufzählung neben dem Paket liefe beim ersten neuen Stil auseinander
// — und zwar in der Richtung, die wehtut: die KI bekäme Felder angeboten, die
// es nicht gibt, oder dürfte welche setzen, die niemand setzen soll.
//
// # Ein leeres Schema ist kein Freibrief
//
// **Der wichtigste Satz des Issues zu diesem Punkt**: „Fehlendes/unlesbares
// Paket erzeugt für KI **keinen beliebigen Feldfreibrief**." Ein Schema, das
// nichts weiß, weiß nicht, dass alles erlaubt ist — es weiß nichts. Wer daraus
// „dann halt alles" machte, hätte die Prüfung genau dort abgeschaltet, wo sie
// am nötigsten ist.
//
// # Was es **nicht** prüft
//
// **Den Inhalt.** Das Issue sagt es scharf: „Schema validiert Form/Felder,
// **niemals Wahrheitsgehalt oder Spielbalance** einer KI-Aussage." Ob ein
// Kartentext gut ist, entscheidet ein Mensch.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
)

// KIFeldtyp ist die Art eines Feldes.
type KIFeldtyp string

// Die zwei, die für eine Prüfung zählen.
const (
	// KIText nimmt jeden Text.
	KIText KIFeldtyp = "text"

	// KIAuswahl nimmt nur, was in [KISchemafeld.Werte] steht.
	KIAuswahl KIFeldtyp = "auswahl"
)

// KISchemafeld ist ein Feld, so wie das Paket es führt.
type KISchemafeld struct {
	Name string
	Typ  KIFeldtyp

	// Werte sind die erlaubten bei [KIAuswahl] — sonst leer.
	Werte []string

	// Mehrfach sagt, ob das Feld mehrere Werte trägt.
	Mehrfach bool

	// Schreibbar sagt, ob eine KI es setzen darf.
	//
	// **Nicht dasselbe wie »es gibt das Feld«**: `casting cost` gibt es, und
	// trotzdem soll eine KI es nicht setzen. Wer beides zusammenlegte, ließe
	// den Unterschied zwischen „kenne ich nicht" und „darfst du nicht"
	// verschwinden — und damit die Meldung, die weiterhilft.
	Schreibbar bool
}

// KISchema ist die Menge der Felder eines Pakets samt seiner Fassung.
type KISchema struct {
	// Paket ist das Vorlagenpaket, aus dem die Felder stammen.
	Paket string

	// Fassung bindet das Schema an einen Stand.
	Fassung string

	Felder []KISchemafeld
}

// Unbekannt sagt, ob das Schema gar keine Felder kennt.
func (s KISchema) Unbekannt() bool { return len(s.Felder) == 0 }

// Kennung ist der Fingerabdruck dieses Schemas.
//
// **Paket, Fassung und Felder gehen ein**: ein Auftrag trägt sie mit, und
// ändert sich das Paket zwischen Auftrag und Übernahme, passt der Vorschlag
// vielleicht nicht mehr.
func (s KISchema) Kennung() string {
	teile := make([]string, 0, 2+len(s.Felder))
	teile = append(teile, s.Paket, s.Fassung)

	for _, feld := range s.Felder {
		teile = append(teile, strings.Join([]string{
			feld.Name, string(feld.Typ),
			strings.Join(feld.Werte, ","),
			fmt.Sprintf("%t/%t", feld.Mehrfach, feld.Schreibbar),
		}, ":"))
	}

	summe := sha256.Sum256([]byte(strings.Join(teile, "\x1f")))

	return hex.EncodeToString(summe[:])
}

// Pruefe hält einen Vorschlag gegen das Schema.
//
// **Ein leerer Vorschlag ist kein Befund**: ein Modell darf sich verweigern,
// und das Issue nennt „Refusal" neben den Fehlern. Eine Antwort ohne
// Vorschlag führt zu keiner Kartenänderung — genau wie gewollt.
func (s KISchema) Pruefe(vorschlag map[string][]string) []Befund {
	befunde := make([]Befund, 0, len(vorschlag))

	// **Die Namen werden sortiert durchlaufen**, damit zwei Läufe dieselbe
	// Reihenfolge geben — sonst wäre jede Meldung eine andere.
	namen := make([]string, 0, len(vorschlag))
	for name := range vorschlag {
		namen = append(namen, name)
	}

	slices.Sort(namen)

	for _, name := range namen {
		befunde = append(befunde, s.pruefeFeld(name, vorschlag[name])...)
	}

	return befunde
}

// pruefeFeld prüft einen einzelnen Feldvorschlag.
func (s KISchema) pruefeFeld(name string, werte []string) []Befund {
	feld, gibts := s.feld(name)
	if !gibts {
		return []Befund{{
			Datei: name, Art: ArtEingabe,
			Meldung: fmt.Sprintf(
				"Das Feld %q gibt es in diesem Paket nicht. Ein Vorschlag "+
					"dafür wird nicht übernommen.", name),
		}}
	}

	if !feld.Schreibbar {
		return []Befund{{
			Datei: name, Art: ArtEingabe,
			Meldung: fmt.Sprintf(
				"Das Feld %q gibt es, aber es wird nicht von einer KI "+
					"gesetzt.", name),
		}}
	}

	if !feld.Mehrfach && len(werte) > 1 {
		return []Befund{{
			Datei: name, Art: ArtEingabe,
			Meldung: fmt.Sprintf(
				"Das Feld %q trägt einen Wert, der Vorschlag nennt %d.",
				name, len(werte)),
		}}
	}

	var befunde []Befund

	for _, wert := range werte {
		if feld.Typ == KIAuswahl && !slices.Contains(feld.Werte, wert) {
			befunde = append(befunde, Befund{
				Datei: name, Ort: wert, Art: ArtEingabe,
				Meldung: fmt.Sprintf(
					"%q ist für das Feld %q nicht vorgesehen — erlaubt sind %s.",
					wert, name, strings.Join(feld.Werte, ", ")),
			})
		}
	}

	return befunde
}

// feld sucht ein Feld nach Namen.
func (s KISchema) feld(name string) (KISchemafeld, bool) {
	for _, feld := range s.Felder {
		if feld.Name == name {
			return feld, true
		}
	}

	return KISchemafeld{}, false
}
