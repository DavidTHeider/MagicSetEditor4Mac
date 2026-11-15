package domain

// **Nach außen eine Sichtkennung statt der Kennung** —
// #2480,
// Vermerk 0201.
//
// Die Kennung eines Objekts trägt Sitz, Deckplatz und Exemplar. Ein
// verborgenes Objekt verlor in der Sicht seinen Inhalt und behielt die
// Kennung — damit waren Geschwister einer gezeigten Karte zuzuordnen, und eine
// bekannte Karte blieb nach dem Mischen wiedererkennbar. **Nach außen geht
// jetzt ein Name, der aus der Kennung, dem Geheimnis der Partie und dem
// Wechsel des Objekts gerechnet ist** — ohne das Geheimnis nicht
// zurückzurechnen, und nach einem Wechsel ein anderer.

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
)

// sichtkennungslaenge ist die Länge des Namens ohne Vorsatz — dieselbe
// Rechnung wie bei [Sitzungsreferenz]: hundertacht Bit.
const sichtkennungslaenge = 18

// Sichtkennung gibt den Namen, unter dem ein Objekt nach außen erscheint.
func (p Partie) Sichtkennung(objekt Spielobjekt) string {
	unterschrift := hmac.New(sha256.New, []byte(p.Geheimnis))

	// **Der Fehler kann hier nicht auftreten** — derselbe Satz wie in
	// [Sitzungsreferenz].
	_, _ = unterschrift.Write([]byte(objekt.Kennung + "\x00" +
		strconv.FormatInt(objekt.Wechsel, 10)))

	return "s-" + base64.RawURLEncoding.EncodeToString(
		unterschrift.Sum(nil))[:sichtkennungslaenge]
}

// Sichtwechsel gibt `danach` mit neuen Wechseln für jedes Objekt, das seit
// `vorher` verborgen wurde oder dessen verborgene Lage sich änderte.
//
// **Eine Stelle nach jedem Befehl** und nicht ein Zähler an jedem Schreiber
// einer Zone (Vermerk 0201, „Verworfen"): ein neuer Schreiber kann es nicht
// vergessen.
func Sichtwechsel(vorher, danach Partie) Partie {
	alt := make(map[string]Spielobjekt, len(vorher.Objekte))
	for _, objekt := range vorher.Objekte {
		alt[objekt.Kennung] = objekt
	}

	naechste := danach.Sichtfolge + 1
	gewechselt := false
	objekte := make([]Spielobjekt, len(danach.Objekte))

	for stelle, objekt := range danach.Objekte {
		if frueher, da := alt[objekt.Kennung]; !da || wirdVerborgen(frueher, objekt) {
			objekt.Wechsel = naechste
			gewechselt = true
		}

		objekte[stelle] = objekt
	}

	if !gewechselt {
		return danach
	}

	danach.Objekte = objekte
	danach.Sichtfolge = naechste

	return danach
}

// wirdVerborgen sagt, ob ein Objekt mit diesem Schritt verborgen wurde oder
// seine verborgene Lage wechselte — CR 400.7 dort, wo es auf das Geheimnis
// ankommt.
//
// **Ein Wechsel zwischen offenen Zonen zählt nicht**: eine Kreatur, die
// stirbt, bleibt am Tisch dasselbe Ding, und das Fenster sieht sie vom
// Schlachtfeld auf den Friedhof gehen.
func wirdVerborgen(vorher, danach Spielobjekt) bool {
	switch {
	case danach.Verdeckt && !vorher.Verdeckt:
		return true
	case danach.Zone != vorher.Zone:
		return danach.Zone.Sicht() != SichtOffen
	default:
		return danach.Zone.Sicht() == SichtVerdeckt && danach.Ordnung != vorher.Ordnung
	}
}
