package domain

// **Was hereinkommt, nennt Sichtkennungen** — #2480, Vermerk 0201, Punkte
// 6 und 7.

import (
	"cmp"
	"slices"
	"strconv"
	"strings"
)

// Echtnamen gibt eine Nutzlast, in der jede Sichtkennung wieder die
// Kennung des Objekts trägt, das sie **jetzt** meint.
//
// **Was keine Sichtkennung ist, bleibt** — ein Sitz, eine Zone, eine neue
// Kennung, die ein Befehl erst vergibt, und eine echte Kennung von einem
// Werkzeug des Wirts.
func (p Partie) Echtnamen(nutzlast map[string]any) map[string]any {
	kennungen := make(map[string]string, len(p.Objekte))
	for _, objekt := range p.Objekte {
		kennungen[p.Sichtkennung(objekt)] = objekt.Kennung
	}

	gelesen, _ := umbenannt(nutzlast, kennungen).(map[string]any) //nolint:errcheck // eine Tabelle bleibt eine

	return gelesen
}

// NenntKennung sagt, ob eine Nutzlast die echte Kennung eines Objekts nennt —
// und welche.
//
// **Ein Gast über die Leitung kennt keine** (Vermerk 0201, Punkt 6): er sieht
// nur Sichtkennungen. Nennt er trotzdem eine, hat er sie erraten — und ein
// erratener Zugriff auf ein verborgenes Objekt ist genau das Leck, das hier
// geschlossen wird.
func (p Partie) NenntKennung(nutzlast map[string]any) (string, bool) {
	echte := make(map[string]bool, len(p.Objekte))
	for _, objekt := range p.Objekte {
		echte[objekt.Kennung] = true
	}

	gefunden := ""

	besuche(nutzlast, func(wert string) {
		if gefunden == "" && echte[wert] {
			gefunden = wert
		}
	})

	return gefunden, gefunden != ""
}

// OhneKennungen gibt einen Satz, in dem jede Kennung ihre Sichtkennung trägt.
//
// **Die längste zuerst** — sonst ersetzte „bo/p0-12/1" den Anfang von
// „bo/p0-12/10".
func (p Partie) OhneKennungen(satz string) string {
	objekte := slices.Clone(p.Objekte)
	slices.SortFunc(objekte, func(a, b Spielobjekt) int {
		return cmp.Compare(len(b.Kennung), len(a.Kennung))
	})

	paare := make([]string, 0, 2*len(objekte))
	for _, objekt := range objekte {
		paare = append(paare, objekt.Kennung, p.Sichtkennung(objekt))
	}

	return strings.NewReplacer(paare...).Replace(satz)
}

// FreieKennung gibt die erste Kennung `wer/wofuer/n`, die es in dieser Partie
// noch nicht gibt — für ein Objekt, dessen Befehl keine nennt (#2480).
func (p Partie) FreieKennung(wer, wofuer string) string {
	vergeben := make(map[string]bool, len(p.Objekte))
	for _, objekt := range p.Objekte {
		vergeben[objekt.Kennung] = true
	}

	for lauf := 0; ; lauf++ {
		if kennung := wer + "/" + wofuer + "/" + strconv.Itoa(lauf); !vergeben[kennung] {
			return kennung
		}
	}
}
