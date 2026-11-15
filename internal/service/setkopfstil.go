package service

// **Spiel und Stil im Kopf eines geschriebenen Sets** — F05
// (#2247),
// Befund 2 des Zweitblicks über Phase G, zweiter Teil.
//
// # Was MSE2 schreibt und liest
//
// **Die Kurzform.** `StyleSheet::stylesheetName` nimmt dem Paketnamen das
// Spiel samt Bindestrich, und beim Lesen sucht `byGameAndName`
// `<spiel>-<kurzform>.mse-style` (`src/data/stylesheet.cpp`, Zeilen 26–60,
// Fassung f9d9356 vom 2021-07-06). Den Block `styling:` ordnet MSE2 nach
// dem **vollen** Namen des Pakets (`Set::stylingDataFor`, `src/data/set.cpp`,
// Zeile 99, dieselbe Fassung).
//
// # Warum eine Auflösung
//
// **Der Bestand kennt beide Formen**: ein eingelesenes Set trägt die
// Kurzform aus seiner Datei, ein hier angelegtes trug bis F05 den vollen
// Paketnamen. Welche dasteht, sagt nur, welches Paket es gibt — dieselbe
// Reihenfolge wie beim Lesen ([stilpaketnamen]). Ohne Auflösung bleibt der
// Wert, wie er im Bestand steht.

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Stilaufloesung sagt, unter welchem Namen das Stilpaket eines Sets liegt.
type Stilaufloesung func(stil, spiel string) (paket string, gefunden bool)

// StilaufloesungUeber löst über die Paketwurzeln auf, die auch Stil und Bild
// lesen — ein Paket gibt es, wenn seine Datei `style` sich lesen lässt.
func StilaufloesungUeber(wurzeln ...string) Stilaufloesung {
	leser := msepaket.NeuerPaketleser(wurzeln...)

	return func(stil, spiel string) (string, bool) {
		for _, kandidat := range stilpaketnamen(stil, spiel) {
			if _, err := leser.Lies(msepaket.Paketpfad{Paket: kandidat + "." + string(msepaket.ArtStil), Datei: "style"}); err == nil {
				return kandidat, true
			}
		}

		return "", false
	}
}

// stilnamen gibt den vollen Paketnamen und die Kurzform des Stils eines Sets
// — `paket` leer, wo sich nichts auflösen ließ.
func stilnamen(kopf domain.SetHead, aufloesung Stilaufloesung) (paket, kurz string) {
	if aufloesung == nil || kopf.Style == "" {
		return "", kopf.Style
	}

	paket, gefunden := aufloesung(kopf.Style, kopf.Game)
	if !gefunden {
		return "", kopf.Style
	}

	if kurz, mitSpiel := strings.CutPrefix(paket, kopf.Game+"-"); mitSpiel {
		return paket, kurz
	}

	return paket, paket
}

// kopfAngleichen setzt `game:` und `stylesheet:` auf den Stand des Bestands.
//
// **Nur, was der Bestand nennt, und nur, wo es anders ist**: eine gleiche
// Zeile bleibt Byte für Byte, und ein Set ohne Stil im Bestand verliert seine
// Zeile nicht. **`stylesheet:` kommt dazu, wo sie fehlt** — sonst ginge ein
// gewählter Stil verloren; `game:` nicht: eine Datei ohne Spiel ist keine,
// die dieser Schreiber reparieren soll.
func kopfAngleichen(baum *msepaket.Baum, kopf domain.SetHead, kurz string) {
	ende := zeilenendeVon(baum.Wurzel)

	if zeile := baum.Ersten("game"); zeile != nil {
		kopfzeileSetzen(baum, "game", kopf.Game, "", ende)
	}

	kopfzeileSetzen(baum, "stylesheet", kurz, "game", ende)
}

// kopfzeileSetzen setzt eine Zeile der obersten Ebene — fehlt sie, kommt sie
// hinter `nach`, ohne `nach` an den Anfang.
func kopfzeileSetzen(baum *msepaket.Baum, schluessel, wert, nach, ende string) {
	if wert == "" {
		return
	}

	if zeile := baum.Ersten(schluessel); zeile != nil {
		if zeile.Wert != wert {
			zeile.Roh, zeile.Wert = "", wert
		}

		return
	}

	neu := &msepaket.Knoten{Zeilenart: msepaket.ArtEintrag, Schluessel: schluessel, Wert: wert, Zeilenende: ende}
	stelle := 0

	for i, zeile := range baum.Wurzel {
		if zeile.Schluessel == nach {
			stelle = i + 1
		}
	}

	// **Die Zeile davor bekommt ihr Ende** (Nachlauf über Phase G, zweiter
	// Teil, Befund 2) — dieselbe Grenze wie in [angehaengt]: endete die Datei
	// mit `game: probe` ohne Umbruch, klebte die neue Zeile daran.
	if letzte := letzteZeile(baum.Wurzel[:stelle]); letzte != nil && letzte.Zeilenende == "" {
		letzte.Zeilenende = ende
	}

	baum.Wurzel = append(baum.Wurzel[:stelle], append([]*msepaket.Knoten{neu}, baum.Wurzel[stelle:]...)...)
}
