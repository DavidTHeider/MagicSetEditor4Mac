package service

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"

// Konsoleneintrag ist ein Eintrag einer Abbildung im Ergebnis der Konsole —
// mit seinen eigenen Einträgen, wenn er selbst eine ist
// (#1424).
//
// **Rekursiv und nicht flach**, weil `card` und `set` Abbildungen über
// Abbildungen sind, und ein Baum die einzige Form ist, in der man sie liest.
type Konsoleneintrag struct {
	Name      string
	Quelltext string
	Art       string
	Eintraege []Konsoleneintrag
}

// eintraegeVon baut die Einträge eines Werts — leer für alles, was keine
// Abbildung ist, und **nie nil**, damit die Wails-Bindung `[]` schreibt.
func eintraegeVon(wert msescript.Wert) []Konsoleneintrag {
	eintraege := make([]Konsoleneintrag, 0, len(wert.Eintraege))
	if wert.Art != msescript.ArtAbbildung {
		return eintraege
	}

	for _, eintrag := range wert.Eintraege {
		eintraege = append(eintraege, Konsoleneintrag{
			Name:      eintrag.Name,
			Quelltext: eintrag.Wert.AlsQuelltext(),
			Art:       string(eintrag.Wert.Art),
			Eintraege: eintraegeVon(eintrag.Wert),
		})
	}

	return eintraege
}
