package service

// **Ein Texturfach im Arbeitsspeicher** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Warum es hier steht und nicht in einer Testdatei
//
// **Weil vier Läufe es brauchen und keiner von ihnen der erste ist.** Es
// erfüllt [Texturablage] ohne Budget und ohne Verdrängung — genau das, was ein
// Lauf über den **Vertrag** will: dass ein Bild wiedergefunden wird, wenn es
// abgelegt wurde. Ob es unter Druck wieder hinausfällt, misst
// `internal/vorschauweg` an seinem echten Fach, und dort gehört es hin.
//
// **Exportiert, weil `internal/app` es ebenso braucht** — die Bindung prüft
// denselben Vertrag eine Schicht höher, und zwei gleiche Doppel liefen beim
// ersten neuen Feld auseinander (§3).

import "sync"

// Texturfachprobe hält Bilder, bis jemand sie wegwirft.
type Texturfachprobe struct {
	schranke sync.Mutex
	bilder   map[string]probenbild

	// Hinterlegt zählt die Ablagen — für Läufe, die wissen wollen, **ob**
	// gezeichnet wurde.
	Hinterlegt int
}

// probenbild ist ein abgelegtes Bild samt seiner Maße.
//
// **Dieselbe Gestalt wie im echten Fach**, und aus demselben Grund: eine
// Probe, die die Maße nicht mitführte, ließe die Zusage „ein Treffer nennt
// dieselben Zahlen" ungeprüft — und zwar überall dort, wo geprüft wird (§3,
// „eine Vorrichtung, die großzügiger ist als der Betrieb").
type probenbild struct {
	bytes []byte

	breite int
	hoehe  int
}

// NeueTexturfachprobe baut eine.
func NeueTexturfachprobe() *Texturfachprobe {
	return &Texturfachprobe{bilder: map[string]probenbild{}}
}

// Adresse gibt die Adresse zu einem Schlüssel samt Maßen — und ob es ihn gibt.
func (t *Texturfachprobe) Adresse(schluessel string) (
	adresse string, breite, hoehe int, da bool,
) {
	t.schranke.Lock()
	defer t.schranke.Unlock()

	bild, da := t.bilder[schluessel]
	if !da {
		return "", 0, 0, false
	}

	return "/spielbild/" + schluessel + ".png", bild.breite, bild.hoehe, true
}

// Bild gibt die abgelegten Bytes — für Läufe, die **hinsehen** wollen.
//
// **Und einen davon gibt es**: dass die zweite Seite die rechte Hälfte zeigt,
// ist eine Aussage über Bildpunkte. Ein Lauf, der nur Schlüssel und Maße
// vergleicht, bleibt grün, wenn beide Seiten dieselbe Hälfte zeigen —
// gemessen am 2026-09-15 mit genau dieser Gegenprobe (§10).
func (t *Texturfachprobe) Bild(schluessel string) ([]byte, bool) {
	t.schranke.Lock()
	defer t.schranke.Unlock()

	bild, da := t.bilder[schluessel]

	return bild.bytes, da
}

// Vergiss wirft ein Bild weg.
func (t *Texturfachprobe) Vergiss(schluessel string) {
	t.schranke.Lock()
	defer t.schranke.Unlock()

	delete(t.bilder, schluessel)
}

// Hinterlege legt ein Bild samt seinen Maßen ab und gibt seine Adresse.
func (t *Texturfachprobe) Hinterlege(
	schluessel string, bild []byte, breite, hoehe int,
) string {
	t.schranke.Lock()
	defer t.schranke.Unlock()

	t.bilder[schluessel] = probenbild{
		bytes: bild, breite: breite, hoehe: hoehe,
	}
	t.Hinterlegt++

	return "/spielbild/" + schluessel + ".png"
}
