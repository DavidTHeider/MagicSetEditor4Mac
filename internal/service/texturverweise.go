package service

import "sync"

// **Wie viele Karten auf ein Spielbild zeigen** — E06
// (#2278).
//
// Der Schlüssel eines Spielbilds ist sein **Inhalt**: sieben gleiche Karten
// teilen sich eines. Ändert sich eine Karte, bekommt sie ein neues, und das
// alte gehört danach niemandem mehr — gemessen im Zwei-Stunden-Lauf vom
// 2026-10-11 wuchs das Fach dadurch um rund 1,7 MB je Minute, bis sein Budget
// es verdrängt hätte. Hier steht, wem ein Bild gehört; das Fach zählt nur
// Bytes.
//
// **Was es nicht hält**: Bilder, die eine gelöschte Karte zuletzt hielt. Sie
// bleiben, bis das Budget sie verdrängt — eine Löschung erreicht diesen
// Dienst nicht, und ein Bild, das noch auf dem Tisch liegt, ist ohnehin
// geladen.

// texturhalter ist eine Seite einer Karte in einer Güte.
type texturhalter struct {
	karte int64
	seite Kartenseite
	guete Texturguete
}

// texturverweise führt je Halter seinen Schlüssel und je Schlüssel die Zahl
// seiner Halter.
type texturverweise struct {
	schranke sync.Mutex
	halter   map[texturhalter]string
	zahl     map[string]int
}

// binde setzt den Schlüssel eines Halters und sagt, ob dabei ein Bild
// verwaist ist — und welches.
func (v *texturverweise) binde(halter texturhalter, schluessel string) (verwaist string, da bool) {
	v.schranke.Lock()
	defer v.schranke.Unlock()

	if v.halter == nil {
		v.halter, v.zahl = map[texturhalter]string{}, map[string]int{}
	}

	alt, hatte := v.halter[halter]
	if hatte && alt == schluessel {
		return "", false
	}

	v.halter[halter] = schluessel
	v.zahl[schluessel]++

	if !hatte {
		return "", false
	}

	v.zahl[alt]--
	if v.zahl[alt] > 0 {
		return "", false
	}

	delete(v.zahl, alt)

	return alt, true
}
