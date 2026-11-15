package schichten

// **Die Prüfung der Composition Roots — das Werk dahinter.**
//
// **Eine eigene Datei**, weil `composition_test.go` mit ihr über die
// Zeilengrenze aus §2 ging. Der Schnitt folgt der Frage: dort steht, **was**
// erlaubt ist — die Listen —, hier, **wie** es geprüft wird.

import (
	"slices"
	"strings"
)

func (b Befund) String() string {
	ort := b.Paket
	if b.Datei != "" {
		ort += "/" + b.Datei
	}

	return ort + " importiert " + b.Import + " — " + b.Grund
}

// Einfuhr ist ein Import an seiner Fundstelle.
type Einfuhr struct {
	Datei string
	Pfad  string
}

// pruefe wendet die Regel auf eine Abbildung Paket → Importe an.
//
// Rein und ohne Dateisystem: nur so lässt sich jede Zeile der Tabelle an einem
// erfundenen Verstoß vorführen, statt ihn in den Baum einzubauen.
func pruefe(importe map[string][]Einfuhr) []Befund {
	var befunde []Befund

	for paket, liste := range importe {
		if _, ausgenommen := keineSchicht[paket]; ausgenommen {
			continue
		}

		erlaubteZiele, gefuehrt := erlaubt[paket]
		if !gefuehrt {
			befunde = append(befunde, Befund{
				Paket: paket,
				Grund: "das Paket steht weder in erlaubt noch in keineSchicht (AGENTS.md §2)",
			})

			continue
		}

		for _, einfuhr := range liste {
			if befund, verstoss := pruefeImport(paket, einfuhr, erlaubteZiele); verstoss {
				befunde = append(befunde, befund)
			}
		}
	}

	slices.SortFunc(befunde, func(a, b Befund) int { return strings.Compare(a.String(), b.String()) })

	return befunde
}

func pruefeImport(paket string, einfuhr Einfuhr, erlaubteZiele []string) (Befund, bool) {
	ziel := einfuhr.Pfad

	for _, verboten := range verboteneFremdpakete[paket] {
		if ziel == verboten {
			return Befund{
				Paket: paket, Datei: einfuhr.Datei, Import: ziel,
				Grund: "in dieser Schicht verboten (AGENTS.md §2)",
			}, true
		}
	}

	eigenes, istEigenes := strings.CutPrefix(ziel, modulpfad+"/internal/")
	if !istEigenes {
		return Befund{}, false
	}

	eigenes, _, _ = strings.Cut(eigenes, "/")

	if eigenes == paket {
		return Befund{}, false
	}

	if !slices.Contains(erlaubteZiele, eigenes) {
		// **Eine Kante allein für Tests** — die Aufzählung und ihr Preis
		// stehen bei [nurInTests].
		if strings.HasSuffix(einfuhr.Datei, "_test.go") &&
			slices.Contains(nurInTests[paket], eigenes) {
			return Befund{}, false
		}

		return Befund{
			Paket: paket, Datei: einfuhr.Datei, Import: ziel,
			Grund: "nicht erlaubt; zulässig sind: " + liste(erlaubteZiele),
		}, true
	}

	// Die Kante ist erlaubt — aber vielleicht nur in einer bestimmten Datei.
	//
	// **Nicht für Testdateien.** Die Verengung gilt der Verdrahtung: sie soll
	// verhindern, dass die Composition-Root-Ausnahme sich still über das ganze
	// Paket ausbreitet. Ein Test, der eine Datenbank vorbereitet, um genau
	// diese Verdrahtung zu prüfen, braucht dasselbe Paket — ihn abzuweisen
	// hieße, die Prüfung der Ausnahme zur Verletzung der Ausnahme zu erklären.
	//
	// Was das kostet, gehört danebengeschrieben: eine fachliche Entscheidung,
	// die jemand in eine Testdatei in `internal/app` legt, sieht dieser
	// Wächter nicht. Er sah sie vorher auch nicht.
	if strings.HasSuffix(einfuhr.Datei, "_test.go") {
		return Befund{}, false
	}

	if dateien, verengt := nurInDatei[paket][eigenes]; verengt &&
		!slices.Contains(dateien, einfuhr.Datei) {
		return Befund{
			Paket: paket, Datei: einfuhr.Datei, Import: ziel,
			Grund: "nur in " + strings.Join(dateien, " und ") +
				" erlaubt (Composition Root, AGENTS.md §2)",
		}, true
	}

	return Befund{}, false
}

func liste(erlaubteZiele []string) string {
	if len(erlaubteZiele) == 0 {
		return "nichts aus dem Projekt"
	}

	return strings.Join(erlaubteZiele, ", ")
}

// ---------------------------------------------------------------- Gegenproben
