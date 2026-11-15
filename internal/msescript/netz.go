package msescript

import (
	"context"
	"slices"
	"strings"
)

// Netz rechnet benannte Skripte, die einander lesen dürfen.
//
// # Wofür
//
// Ein Kartenfeld ist ein Skript, und es liest andere Felder: die Breite des
// Typfelds hängt an `card.type_1`, die Rahmenfarbe an `card.rarity`. Ändert
// jemand die Seltenheit, muss die Rahmenfarbe neu — und **nur** sie.
//
// Das Netz weiß nicht, was eine Karte ist. Es kennt Namen, Skripte und wer
// wen liest; `internal/service` legt darüber, was davon ein Kartenfeld ist.
// Anders ginge es nicht: `msescript` kennt `domain` nicht (AGENTS.md §2), und
// ein Interpreter mit einer Kartenentität darin wäre nicht mehr gegen eine
// Tabelle aus Ausdruck und Ergebnis zu prüfen.
//
// # Zwei Zusagen, die anders lauten als beim einzelnen Auswerter
//
// **Ein Netzskript rechnet in der Umgebung des Netzes, nicht in der des
// Aufrufers.** Eine geschweifte Klammer bekommt ihre Umgebung gereicht
// (`umgebung.go`) — ein Netzskript nicht. Das ist kein Versehen, sondern die
// Bedingung dafür, dass ein gerechneter Wert liegen bleiben darf: der Behälter
// ist nach dem **Namen** geführt, und ein Wert, der von den Namen des Aufrufers
// abhinge, wäre für den nächsten Aufrufer falsch.
//
// Der Zweitblick über Phase 3 hat den Unterschied angestrichen. Wer ihn
// aufheben will, muss zuerst den Behälter aufgeben oder ihn nach Umgebung
// **und** Namen führen.
//
// **Ein Netz gehört einem Aufrufer zur Zeit.** `Wert` und `Geaendert`
// schreiben beide in dieselben Abbildungen, und `Wert` führt einen Stapel für
// die Kreiserkennung. Zwei überlappende Läufe ergäben ein Datenrennen und
// falsche Kreise. Wer nebenläufig rechnen will, nimmt je Lauf ein eigenes Netz
// oder reiht die Läufe auf.
type Netz struct {
	skripte  map[string]*Knoten
	umgebung *Umgebung

	werte  map[string]Wert
	liest  map[string][]string
	imLauf []string
}

// NeuesNetz baut ein Netz über bereits zerteilten Skripten.
func NeuesNetz(skripte map[string]*Knoten, umgebung *Umgebung) *Netz {
	if umgebung == nil {
		umgebung = Bibliothek()
	}

	return &Netz{
		skripte:  skripte,
		umgebung: umgebung,
		werte:    map[string]Wert{},
		liest:    map[string][]string{},
	}
}

// Wert rechnet ein Skript und merkt sich, was es gelesen hat.
//
// Einmal gerechnete Werte bleiben liegen, bis `Geaendert` sie verwirft. Ohne das
// rechnete eine Vorschau bei jedem Tastendruck jedes Feld neu.
func (n *Netz) Wert(ctx context.Context, name string) (Wert, error) {
	if wert, da := n.werte[name]; da {
		return wert, nil
	}

	skript, da := n.skripte[name]
	if !da {
		return Nichts, NeuerFehler(0, 0, SchluesselKeinSkript, name)
	}

	// **Der Zyklus wird benannt, nicht nur gemeldet.** „a hängt von b ab, b von
	// a" ist zu finden; „irgendwo im Set ist ein Kreis" nicht.
	if kreis := slices.Index(n.imLauf, name); kreis >= 0 {
		return Nichts, NeuerFehler(skript.Zeile, skript.Spalte, SchluesselSkriptkreis,
			strings.Join(append(n.imLauf[kreis:], name), " → "))
	}

	n.imLauf = append(n.imLauf, name)
	defer func() { n.imLauf = n.imLauf[:len(n.imLauf)-1] }()

	wert, gelesen, err := n.rechne(ctx, skript)
	if err != nil {
		return Nichts, err
	}

	n.werte[name] = wert
	n.liest[name] = gelesen

	return wert, nil
}

// rechne wertet ein Skript aus und sammelt dabei die gelesenen Namen.
func (n *Netz) rechne(ctx context.Context, skript *Knoten) (Wert, []string, error) {
	gelesen := map[string]bool{}

	umgebung := n.umgebung.Darueber()
	for name := range n.skripte {
		umgebung.Setze(name, Wert{Art: ArtFunktion, Funktion: &Funktionsteil{Netzname: name}})
	}

	a := &auswerter{ctx: ctx, netz: n, gelesen: gelesen}

	wert, err := a.werte(skript, umgebung.Darueber())
	if err != nil {
		return Nichts, nil, err
	}

	namen := make([]string, 0, len(gelesen))
	for name := range gelesen {
		namen = append(namen, name)
	}

	slices.Sort(namen)

	return wert, namen, nil
}

// Geaendert verwirft einen Wert und alles, was ihn liest — mittelbar auch.
//
// Zurückgegeben wird, was neu zu rechnen ist. Der Aufrufer entscheidet, ob er
// es sofort tut oder auf die nächste Anzeige wartet; das Netz drängt sich nicht
// auf.
func (n *Netz) Geaendert(name string) []string {
	betroffen := map[string]bool{name: true}

	// Bis sich nichts mehr ändert: ein Feld kann über drei Ecken hängen.
	for gewachsen := true; gewachsen; {
		gewachsen = false

		for leser, gelesen := range n.liest {
			if betroffen[leser] {
				continue
			}

			for _, quelle := range gelesen {
				if betroffen[quelle] {
					betroffen[leser] = true
					gewachsen = true

					break
				}
			}
		}
	}

	namen := make([]string, 0, len(betroffen))

	for betroffener := range betroffen {
		delete(n.werte, betroffener)
		delete(n.liest, betroffener)

		if betroffener != name {
			namen = append(namen, betroffener)
		}
	}

	slices.Sort(namen)

	return namen
}

// Liest sagt, welche Namen ein gerechnetes Skript gelesen hat.
func (n *Netz) Liest(name string) []string { return n.liest[name] }
