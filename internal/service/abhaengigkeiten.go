package service

import (
	"context"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Abhaengigkeitsart sagt, was mit einer Abhängigkeit nicht stimmt.
//
// **Ein Schlüssel und kein Satz** (§9): der Satz steht in den Sprachkatalogen
// des Frontends. Ein deutscher Fließtext von hier wäre in der englischen
// Oberfläche deutsch.
type Abhaengigkeitsart string

const (
	// AbhaengigkeitFehlt: das gebrauchte Paket ist nicht installiert.
	AbhaengigkeitFehlt Abhaengigkeitsart = "fehlt"

	// AbhaengigkeitZuAlt: es ist installiert, aber älter als verlangt.
	AbhaengigkeitZuAlt Abhaengigkeitsart = "zualt"

	// AbhaengigkeitUnlesbar: eine der beiden Versionen ist kein Datum.
	//
	// **Ein eigener Befund und kein Urteil.** Eine krumme Version ist ein
	// Befund, kein Abbruch (Phase 2, Punkt 3) — und ein Vergleich gegen etwas,
	// das kein Datum ist, wäre eine Behauptung.
	AbhaengigkeitUnlesbar Abhaengigkeitsart = "unlesbar"

	// AbhaengigkeitZyklus: die Pakete brauchen einander im Kreis.
	AbhaengigkeitZyklus Abhaengigkeitsart = "zyklus"
)

// Abhaengigkeitsbefund ist ein Vorbehalt gegen eine Abhängigkeit.
//
// **Angaben, kein Satz.** Wer den Satz baut, ist das Frontend; hier stehen die
// Werte, die er nennt.
type Abhaengigkeitsbefund struct {
	// Paket ist der, der etwas braucht — `magic-m15.mse-style`.
	Paket string

	// Braucht ist das gebrauchte Paket, wie `depends on:` es nennt.
	Braucht string

	Art Abhaengigkeitsart

	// Verlangt ist die Version aus `depends on:`; Da die installierte.
	Verlangt string
	Da       string
}

// Abhaengigkeiten prüft die Abhängigkeiten aller installierten Pakete.
//
// # Die drei Zusagen aus #57
//
//   - **Eine zu alte Abhängigkeit wird benannt, bevor eine Vorlage
//     stillschweigend falsch rendert.** Ein Stil rechnet mit den Funktionen
//     seines Spiels; ist das Spiel älter als verlangt, fehlen sie — und was
//     dabei herauskommt, ist eine Karte, der man es nicht ansieht.
//   - **Ein Zyklus wird erkannt.**
//   - **Fehlende Abhängigkeiten sind ein Befund mit Vorschlag, kein Abbruch.**
//     Der Vorschlag steht in der Oberfläche: was fehlt, lässt sich beziehen
//     oder übernehmen. Hier steht, **was** fehlt.
//
// # Was sie nicht prüft
//
// **Ob eine Vorlage ohne ihre Abhängigkeit trotzdem läuft.** Das entscheidet
// der Planbauer beim Deuten, und er benennt es dort (§8). Diese Prüfung liest
// den Katalog und sonst nichts — sie öffnet kein einziges Paket.
func (v *Paketverwalter) Abhaengigkeiten(ctx context.Context) ([]Abhaengigkeitsbefund, error) {
	pakete, err := v.katalog.Pakete(ctx)
	if err != nil {
		return nil, err
	}

	da := make(map[string]domain.Paket, len(pakete))
	for _, paket := range pakete {
		da[paket.Name+"."+paket.Art] = paket
	}

	befunde := make([]Abhaengigkeitsbefund, 0)

	for _, paket := range pakete {
		voll := paket.Name + "." + paket.Art

		for _, bedarf := range paket.Braucht {
			if befund := pruefeBedarf(voll, bedarf, da); befund != nil {
				befunde = append(befunde, *befund)
			}
		}
	}

	return append(befunde, zyklen(pakete)...), nil
}

// pruefeBedarf urteilt über eine einzelne Abhängigkeit.
func pruefeBedarf(
	paket string, bedarf domain.Bedarf, da map[string]domain.Paket,
) *Abhaengigkeitsbefund {
	gebraucht, installiert := da[bedarf.Paket]
	if !installiert {
		return &Abhaengigkeitsbefund{
			Paket: paket, Braucht: bedarf.Paket,
			Art: AbhaengigkeitFehlt, Verlangt: bedarf.Version,
		}
	}

	// **Ohne verlangte Version gibt es nichts zu vergleichen.** `depends on:`
	// ohne `version:` kommt vor; daraus ein Urteil zu machen hieße, eine Regel
	// zu erfinden.
	if strings.TrimSpace(bedarf.Version) == "" {
		return nil
	}

	verlangt, fehler := msepaket.DatumLesen(bedarf.Version)
	vorhanden, zweiter := msepaket.DatumLesen(gebraucht.Version)

	if fehler != nil || zweiter != nil {
		return &Abhaengigkeitsbefund{
			Paket: paket, Braucht: bedarf.Paket, Art: AbhaengigkeitUnlesbar,
			Verlangt: bedarf.Version, Da: gebraucht.Version,
		}
	}

	// **`VorOder` und kein Zeichenkettenvergleich.** Zwei ISO-Daten sortieren
	// zufällig auch als Text; auf diese Zufälligkeit soll sich nichts stützen
	// (der Grund steht bei [msepaket.Datum]).
	if vorhanden.VorOder(verlangt) && vorhanden != verlangt {
		return &Abhaengigkeitsbefund{
			Paket: paket, Braucht: bedarf.Paket, Art: AbhaengigkeitZuAlt,
			Verlangt: bedarf.Version, Da: gebraucht.Version,
		}
	}

	return nil
}

// zyklen findet Pakete, die einander im Kreis brauchen.
//
// **Tiefensuche mit drei Farben**, und gemeldet wird die **Rückkante**: das
// Paar, an dem der Kreis sich schließt.
//
// **Ein Befund je Kreis, nicht je beteiligtem Paket.**
//
// **Und ein Befund ist die bessere Antwort**: ein Kreis über drei Pakete ist
// **ein** Sachverhalt. Drei Zeilen darüber lesen sich wie drei Probleme, und
// wer eines davon behebt, hat alle drei behoben.
func zyklen(pakete []domain.Paket) []Abhaengigkeitsbefund {
	kanten := make(map[string][]string, len(pakete))
	for _, paket := range pakete {
		voll := paket.Name + "." + paket.Art
		for _, bedarf := range paket.Braucht {
			kanten[voll] = append(kanten[voll], bedarf.Paket)
		}
	}

	// 0 = ungesehen, 1 = im Lauf, 2 = fertig
	farbe := make(map[string]int, len(kanten))
	gefunden := make([]Abhaengigkeitsbefund, 0)

	var gehe func(knoten string)

	gehe = func(knoten string) {
		farbe[knoten] = 1

		for _, ziel := range kanten[knoten] {
			switch farbe[ziel] {
			case 0:
				gehe(ziel)

			case 1:
				gefunden = append(gefunden, Abhaengigkeitsbefund{
					Paket: knoten, Braucht: ziel, Art: AbhaengigkeitZyklus,
				})
			}
		}

		farbe[knoten] = 2
	}

	// **In fester Reihenfolge**, damit derselbe Bestand dieselbe Liste gibt.
	// Eine Abbildung in Go hat keine, und eine Anzeige, die bei jedem Aufruf
	// eine andere Reihenfolge zeigt, sieht aus wie ein wechselnder Befund.
	namen := make([]string, 0, len(kanten))
	for knoten := range kanten {
		namen = append(namen, knoten)
	}

	slices.Sort(namen)

	for _, knoten := range namen {
		if farbe[knoten] == 0 {
			gehe(knoten)
		}
	}

	return gefunden
}
