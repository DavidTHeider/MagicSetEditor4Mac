// **Wie ein Name auf ein Glied trifft** — die beiden Regeln, nach denen
// [Wert.Eintrag] sucht.
//
// # Warum eine eigene Datei
//
// `wert.go` stand an der Zeilengrenze aus §2, als der Zugriff über eine
// **Stelle** dazukam
// (#1356).
// Der Schnitt folgt der Frage: *was* ein Wert ist, steht dort; *wie* ein Name
// darin gefunden wird, hier.
//
// **Und die beiden gehören zusammen.** MSE sucht in derselben Reihenfolge:
// erst über den Namen (`ScriptCustomCollection::getMember`,
// `src/script/value.cpp` Zeilen 512–519), dann über die Stelle
// (`ScriptValue::getMember` → `getIndex`, ebd. Zeilen 64–73). Fassung
// `f9d9356` vom 2021-07-06.

package msescript

import (
	"strconv"
	"strings"
)

// gliedtrifft bildet MSEs [canonical_name_compare] nach.
//
// **Der gesuchte Name ist der Präfix, nicht umgekehrt.** `separator` findet
// `separators`; `separators` findet **kein** `separator`, weil die Schleife
// dann über das Ende des Gliednamens liefe.
//
// **Und ein `_` im gesuchten Namen trifft ein Leerzeichen im Glied** — MSEs
// Ausnahme in derselben Zeile.
func gliedtrifft(gesucht, glied string) bool {
	if len(gesucht) > len(glied) {
		return false
	}

	for stelle := range len(gesucht) {
		if gesucht[stelle] == glied[stelle] {
			continue
		}

		if gesucht[stelle] == '_' && glied[stelle] == ' ' {
			continue
		}

		return false
	}

	return true
}

func (w Wert) String() string {
	switch w.Art {
	case ArtAbbildung:
		teile := make([]string, 0, len(w.Eintraege))
		for i := range w.Eintraege {
			teile = append(teile, w.Eintraege[i].Name+": "+w.Eintraege[i].Wert.String())
		}

		return "[" + strings.Join(teile, ", ") + "]"

	case ArtFunktion:
		return "Funktion"

	case ArtTextwert:
		return strconv.Quote(w.Text)

	case ArtNichts:
		return "nil"

	case ArtZahlwert, ArtWahrheitswert, ArtFarbe:
	}

	return w.AlsText()
}

// anStelle gibt das Glied an der genannten Stelle — MSEs `getIndex`.
//
// **Ein Name, der sich als Zahl lesen lässt, ist eine Stelle.** So steht es in
// der Basisklasse und gilt damit für jeden Wert:
//
//	long index;
//	if (name.ToLong(&index)) return getIndex(index);
//
// **Gezählt werden allein die unbenannten Glieder.** `getIndex` greift in
// `value`, und dort stehen nur die ohne Schlüssel
// (`ScriptCustomCollection::getIndex`, Zeilen 520–526). Ein benanntes Glied
// dazwischen verschiebt die Stellen also **nicht**.
//
// **Und eine Abbildung aus lauter benannten Gliedern hat keine Stellen.** Das
// ist die Zusage aus #653:
// `set.0` gibt keine Karte, weil `set` seine Felder als Namen führt. In MSE
// fällt `ScriptObject<Set*>::getMember` auf `ScriptValue::getIndex` zurück,
// und das ist ein Fehlerwert.
func anStelle(unter []Eintrag, name string) (Wert, bool) {
	stelle, err := strconv.Atoi(name)
	if err != nil || stelle < 0 {
		return Nichts, false
	}

	gezaehlt := 0

	for i := range unter {
		if unter[i].Name != "" {
			continue
		}

		if gezaehlt == stelle {
			return unter[i].Wert, true
		}

		gezaehlt++
	}

	return Nichts, false
}
