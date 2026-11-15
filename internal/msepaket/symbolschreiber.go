package msepaket

import "fmt"

// SymbolFassung ist die Fassung, die MSE in eine `.mse-symbol` schreibt.
//
//	const Version file_version_symbol          =   305; // 0.3.5
//
// `src/util/version.cpp`, Zeile 95, Fassung f9d9356 vom 2021-07-06. Die
// Zeichenkette entsteht dort aus der Zahl (`Version::toString`, Zeilen 17–30):
// 305 → 0.3.5.
const SymbolFassung = "0.3.5"

// SchreibeSymbolzeichnung gibt eine Zeichnung als Dateiinhalt zurück.
//
// # Warum es das gibt
//
// Bis Phase X wurde `.mse-symbol` **gelesen und nie geschrieben**. Ein eigenes
// Setsymbol ließ sich damit nicht anlegen, nur eines übernehmen, das schon
// existiert — und auf diesem Rechner existiert keines: unter den zehn
// kanonischen Sets trägt **keines** ein Symbol (`symbol:` ist überall leer),
// und `M15-Magic-Pack` enthält **null** `.mse-symbol`-Dateien; seine
// `.mse-symbol-font`-Verzeichnisse sind PNG-Sammlungen. Gemessen am
// 2026-08-24; die Aussage reicht so weit wie diese beiden Quellen.
//
// # Was hier über MSE nachgelesen ist
//
// Die Reihenfolge der Felder ist die der Reflexion, denn genau in ihr schreibt
// MSE:
//
// Alles aus `src/data/symbol.cpp`, Fassung f9d9356 vom 2021-07-06:
//
//   - `SymbolShape`: Basis (`type`, `name`), dann `combine`, dann die Punkte
//     (Zeilen 162–166)
//   - `ControlPoint`: `position`, `lock`, `line_after`, und `handle_before`
//     bzw. `handle_after` **nur bei einer Kurve** (Zeilen 26–36)
//   - `SymbolSymmetry`: Basis, `kind`, `copies`, `center`, `handle`, `part`
//     (Zeilen 278–285)
//   - `SymbolGroup`: Basis, `part` (Zeilen 323–326)
//
// Die Mehrzahl wird zur Einzahl — `points` steht als `point:` in der Datei
// (`singular_form`, `src/util/string.cpp`, Zeilen 153–160).
//
// # Zwei Abweichungen, und beide sind benannt
//
// **Unterstriche statt Leerzeichen.** MSEs Schreiber gibt den Feldnamen aus,
// wie er im Code steht, ohne Rückumwandlung (`src/util/io/writer.cpp`, Zeile
// 56) — also `line_after`, nicht `line after`. Hier wird es genauso gemacht,
// damit unsere Datei aussieht wie seine.
//
// **Ein leeres `name:` bekommt kein Leerzeichen dahinter.** MSE schreibt
// `": "` und dann den leeren Wert, also eine Zeile mit Abschlussleerzeichen
// (ebd., Zeile 99). Der Schreiber dieses Projekts lässt es weg; MSEs Leser
// schneidet ohnehin ab (`trim(key)`, `reader.cpp`, Zeile 217), und ein
// Abschlussleerzeichen in jeder erzeugten Zeile wäre in jedem Diff zu sehen.
func SchreibeSymbolzeichnung(zeichnung Symbolzeichnung) []byte {
	baum := &Baum{
		Datei:  "symbol",
		MitBOM: true,
		Wurzel: []*Knoten{eintrag("mse_version", SymbolFassung)},
	}

	for _, teil := range zeichnung.Teile {
		baum.Wurzel = append(baum.Wurzel, teilKnoten(teil))
	}

	return Schreibe(baum)
}

// teilKnoten baut einen `part:`-Block.
func teilKnoten(teil Symbolteil) *Knoten {
	knoten := &Knoten{Zeilenart: ArtEintrag, Schluessel: "part"}

	knoten.Kinder = append(knoten.Kinder,
		eintrag("type", typname(teil.Art)),
		eintrag("name", teil.Name))

	switch teil.Art {
	case TeilForm:
		kombination := teil.Kombination
		if kombination == "" {
			kombination = KombinationUeberdecken
		}

		knoten.Kinder = append(knoten.Kinder, eintrag("combine", string(kombination)))

		for _, punkt := range teil.Punkte {
			knoten.Kinder = append(knoten.Kinder, punktKnoten(punkt))
		}

	case TeilSymmetrie:
		spiegelart := teil.Spiegelart
		if spiegelart == "" {
			spiegelart = SpiegelungDrehung
		}

		knoten.Kinder = append(knoten.Kinder,
			eintrag("kind", string(spiegelart)),
			eintrag("copies", fmt.Sprintf("%d", teil.Kopien)),
			eintrag("center", vektorText(teil.Mitte)),
			eintrag("handle", vektorText(teil.Griff)))

		fallthrough

	case TeilGruppe:
		for _, kind := range teil.Teile {
			knoten.Kinder = append(knoten.Kinder, teilKnoten(kind))
		}
	}

	return knoten
}

// punktKnoten baut einen `point:`-Block.
//
// **Die Griffe stehen nur bei einer Kurve da.** MSE schreibt sie unter
// derselben Bedingung (`src/data/symbol.cpp`, Zeilen 30–35, Fassung f9d9356
// vom 2021-07-06); ein Griff an einer
// Geraden wäre eine Zahl, die niemand benutzt.
func punktKnoten(punkt Kontrollpunkt) *Knoten {
	sperre := punkt.Sperre
	if sperre == "" {
		sperre = "free"
	}

	art := "line"
	if punkt.KurveDanach {
		art = "curve"
	}

	knoten := &Knoten{Zeilenart: ArtEintrag, Schluessel: "point", Kinder: []*Knoten{
		eintrag("position", vektorText(punkt.Stelle)),
		eintrag("lock", sperre),
		eintrag("line_after", art),
	}}

	if punkt.KurveDavor {
		knoten.Kinder = append(knoten.Kinder, eintrag("handle_before", vektorText(punkt.GriffDavor)))
	}

	if punkt.KurveDanach {
		knoten.Kinder = append(knoten.Kinder, eintrag("handle_after", vektorText(punkt.GriffDanach)))
	}

	return knoten
}

// vektorText schreibt ein `(x,y)`.
//
//	handle(String::Format(_("(%.10lf,%.10lf)"), vec.x, vec.y));
//
// `src/util/io/writer.cpp`, Zeile 131, Fassung f9d9356 vom 2021-07-06. **Zehn
// Nachkommastellen, immer** — auch bei einer glatten Null.
func vektorText(vektor Vektor) string {
	return fmt.Sprintf("(%.10f,%.10f)", vektor.X, vektor.Y)
}

// typname gibt den `type:`-Wert einer Teilart — die Gegenrichtung zu
// [teilarten].
//
// **Ausgeschrieben, auch bei der Form.** MSE liest einen leeren Typ als
// `shape`, schreibt ihn aber aus (`SymbolShape::typeName`, `src/data/symbol.cpp`,
// Zeile 218, Fassung f9d9356 vom 2021-07-06). Wer sich auf den Standardwert
// verlässt, schreibt eine Datei, die von einer anderen Lesart abhängt.
func typname(art Symbolteilart) string {
	for name, kandidat := range teilarten {
		if kandidat == art && name != "" {
			return name
		}
	}

	return "shape"
}

// eintrag baut einen einfachen `schlüssel: wert`-Knoten.
func eintrag(schluessel, wert string) *Knoten {
	return &Knoten{Zeilenart: ArtEintrag, Schluessel: schluessel, Wert: wert}
}
