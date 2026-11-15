package vorschauweg

import "strings"

// vorschaumarke liest die Marke aus der Adresse eines Vorschaubildes.
//
// **Ein Name hinter der Marke sucht nichts** — AP-12
// (#2369),
// B-24: zieht man das Bild in den Finder, nennt WebKit die Datei nach dem
// letzten Teil der Adresse. Unter `<Pfad><marke>/<name>.png` heißt sie wie
// die Karte; gesucht wird weiter nur die Marke.
func vorschaumarke(pfad string) string {
	rest := strings.TrimPrefix(pfad, Pfad)
	if vorn, _, mitNamen := strings.Cut(rest, "/"); mitNamen {
		rest = vorn
	}

	return strings.TrimSuffix(rest, ".png")
}
