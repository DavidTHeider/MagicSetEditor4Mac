package msepaket

import (
	"path"
	"path/filepath"
	"strings"
)

// **Ein Setsymbol hat benannte Fassungen — `variation:`.**
//
//	IMPLEMENT_REFLECTION_NO_SCRIPT(SymbolVariation) {
//	  REFLECT(name);
//	  REFLECT(border_radius);
//	  REFLECT_NAMELESS(filter);
//	}
//	SymbolVariation::SymbolVariation() : border_radius(0.05) {}
//
// `src/data/field/symbol.cpp`, Zeilen 31–46, Fassung f9d9356 vom 2021-07-06.
//
// **Sie stehen an zwei Orten**: das Spiel erklärt sie am `set field:`, die
// Vorlage ergänzt sie unter `set info style:`. `magic.mse-game` führt `common`,
// `uncommon`, `rare` und `mythic rare`; `magic-m15.mse-style` legt
// `invertedcommon` dazu.
//
// **Daran hängt das Seltenheitssymbol jeder Karte** —
// `symbol_variation(symbol: set.symbol, variation: "common")`
// (#323).

// Symbolfassung ist ein `variation:`-Block.
//
// **Die Farben stehen roh.** Ob `rgb(0,0,0)` oder `black` — was sie bedeuten,
// weiß `internal/service`; `msepaket` liest, was dasteht (§2).
type Symbolfassung struct {
	Name string

	// Feld ist das Setfeld, unter dem diese Fassung erklärt ist — `symbol`
	// oder `masterpiece symbol`.
	//
	// **MSE sucht die Fassung am Stil des konkreten Feldes**
	// (`value->fieldP`, `src/script/functions/image.cpp`, Zeile 161, Fassung
	// f9d9356 vom 2021-07-06); wir sammelten sie über alle Felder in eine
	// Liste. `magic.mse-game/set_fields` erklärt beide, und sie unterscheiden
	// sich an genau einer Stelle: `watermark` ist unter `symbol:`
	// `fill color: rgb(0,0,0)`, unter `masterpiece symbol:` `rgb(255,255,255)`
	// (#548).
	Feld string

	// Randbreite ist `border radius:` — ein **Anteil** der Symbolgröße.
	// Ohne Angabe gilt 0,05.
	Randbreite float64

	// Fuellart ist `fill type:` — `solid`, `linear gradient` oder
	// `radial gradient`.
	Fuellart string

	// Die einfachen Farben gelten für `solid`, die nummerierten für die beiden
	// Verläufe.
	Fuellfarbe  string
	Randfarbe   string
	Fuellfarbe1 string
	Fuellfarbe2 string
	Randfarbe1  string
	Randfarbe2  string

	// Mitte und Ende gelten für den linearen Verlauf — `center x:` und
	// `end x:`, Anteile der Fläche.
	MitteX, MitteY float64
	EndeX, EndeY   float64
}

// symbolfassungenAus liest die `variation:`-Blöcke unter `set info style:`.
//
// **Nur die des Symbols.** Ein `set info style:` beschreibt jedes Setfeld;
// Fassungen hat allein das Symbolfeld, und ein Leser über alle Felder
// sammelte Blöcke ein, die niemand so gemeint hat.
func symbolfassungenAus(kopf *Baum) []Symbolfassung {
	return fassungenUnter(kopf, "set info style")
}

// fassungenUnter liest die `variation:`-Blöcke unter einem Stilblock.
func fassungenUnter(kopf *Baum, blockname string) []Symbolfassung {
	var fassungen []Symbolfassung

	for _, block := range kopf.Alle(blockname) {
		for _, feld := range block.Kinder {
			if feld.Zeilenart != ArtEintrag {
				continue
			}

			for _, knoten := range feld.Alle("variation") {
				fassung := symbolfassungAus(knoten)
				fassung.Feld = feld.Schluessel
				fassungen = append(fassungen, fassung)
			}
		}
	}

	return fassungen
}

// symbolfassungAus deutet einen `variation:`-Block.
func symbolfassungAus(knoten *Knoten) Symbolfassung {
	fassung := Symbolfassung{
		Name:     strings.TrimSpace(wertOderLeer(knoten.Ersten("name"))),
		Fuellart: strings.TrimSpace(wertOderLeer(knoten.Ersten("fill type"))),

		Fuellfarbe:  strings.TrimSpace(wertOderLeer(knoten.Ersten("fill color"))),
		Randfarbe:   strings.TrimSpace(wertOderLeer(knoten.Ersten("border color"))),
		Fuellfarbe1: strings.TrimSpace(wertOderLeer(knoten.Ersten("fill color 1"))),
		Fuellfarbe2: strings.TrimSpace(wertOderLeer(knoten.Ersten("fill color 2"))),
		Randfarbe1:  strings.TrimSpace(wertOderLeer(knoten.Ersten("border color 1"))),
		Randfarbe2:  strings.TrimSpace(wertOderLeer(knoten.Ersten("border color 2"))),

		// **Die Vorgaben sind 0,5/0,5 und 1/1**, nicht null:
		//
		//	GradientSymbolFilter::GradientSymbolFilter()
		//	  : center_x(0.5), center_y(0.5)
		//	  , end_x(1), end_y(1)
		//
		// `src/render/symbol/filter.cpp`, Zeilen 138–142, Fassung f9d9356 vom
		// 2021-07-06. Keine Fassung des Packs nennt diese Angaben — mit null
		// wurde MSEs `len` eins und `t` überall null, und **jeder Verlauf war
		// eine einfarbige Fläche**
		// (#270).
		MitteX: zahlOder(knoten.Ersten("center x"), 0.5),
		MitteY: zahlOder(knoten.Ersten("center y"), 0.5),
		EndeX:  zahlOder(knoten.Ersten("end x"), 1),
		EndeY:  zahlOder(knoten.Ersten("end y"), 1),
	}

	// **Ohne Angabe gilt 0,05** — MSEs `border_radius(0.05)`.
	fassung.Randbreite = 0.05
	if roh := knoten.Ersten("border radius"); roh != nil {
		fassung.Randbreite = zahlOderNull(roh)
	}

	return fassung
}

// Spielsymbolfassungen liest die `variation:`-Blöcke der `set field:`-Blöcke.
//
// **Das Spiel erklärt die Grundfassungen, die Vorlage ergänzt sie.**
// `magic.mse-game` führt `common`, `uncommon`, `rare` und `mythic rare`;
// `magic-m15.mse-style` legt `invertedcommon` dazu. Beide Listen gelten
// zusammen, und die der Vorlage steht **hinten** — sie gewinnt bei gleichem
// Namen, so wie eine Vorlage überall das Spiel überschreiben darf.
// **Sie stehen unter `default set style:`**, nicht am `set field:`. Das Feld
// erklärt, *was* ein Setsymbol ist; wie es aussieht, steht im Stil — und der
// des Spiels ist die Vorgabe für jede Vorlage
// (`magic.mse-game/set_fields`, Zeilen 535–571).
func (p *Paket) Spielsymbolfassungen() []Symbolfassung {
	return fassungenUnter(p.Kopf, "default set style")
}

// Symbolgestalt liest die Gestalt einer `.mse-symbol` aus einem Paket.
//
// **Sie steht an der Bildquelle, weil dort die Pakete liegen.** Der Planbauer
// öffnet keine; er fragt die Quelle, die er ohnehin hat
// (#323).
//
// **Was nicht lesbar ist, ist leer** — und der Rasterer benennt es. Ein Fehler
// hier bräche den ganzen Plan für eine Datei, die auf den meisten Karten gar
// nicht gebraucht wird (§8).
func (b *Bildquelle) Symbolgestalt(verweis string) Symbolzeichnung {
	// **Das Set-Paket zuerst**, genau wie bei [Bildquelle.Bild]: sein Ort steht
	// im Verweis und wird nicht gesucht. `set.symbol` nennt eine `.mse-symbol`
	// **im Set**, und MSE liest sie aus dem lokalen Paket
	// (`SymbolToImage::generate`, `src/gfx/generated_image.cpp`, Zeilen
	// 453–462, Fassung f9d9356 vom 2021-07-06).
	if ort, datei, ist := ZerlegeSetpaketverweis(verweis); ist {
		return b.gestaltAusSetpaket(verweis, ort, datei)
	}

	pfad, err := ZerlegePaketpfad(verweis)
	if err != nil {
		return Symbolzeichnung{}
	}

	for _, wurzel := range b.wurzeln {
		rohdaten, fehler := paketdatei(filepath.Join(wurzel, pfad.Paket), pfad.Datei)
		if fehler != nil {
			continue
		}

		// **Eine unlesbare Datei ist eine leere Gestalt**, kein Fehler: der
		// Rasterer benennt sie, und ein Abbruch hier nähme der Karte alles
		// andere mit (§8).
		baum, fehler := Zerteile(verweis, rohdaten)
		if fehler != nil {
			return Symbolzeichnung{}
		}

		zeichnung, _ := Symbolzeichnungsmodell(baum)

		return zeichnung
	}

	return Symbolzeichnung{}
}

// gestaltAusSetpaket liest eine Zeichnung aus dem Paket an seinem Ort.
//
// **Dieselbe Prüfung wie bei einem Bild** ([Bildquelle.ausSetpaket]): der
// Verweis nennt einen Ort, und der wird geprüft, bevor er auf die Platte geht.
//
// **Was nicht lesbar ist, ist leer** — und der Rasterer benennt es, so wie im
// gewöhnlichen Weg darüber.
func (b *Bildquelle) gestaltAusSetpaket(verweis, ort, datei string) Symbolzeichnung {
	if err := pruefeSetpaketverweis(verweis, ort, datei); err != nil {
		return Symbolzeichnung{}
	}

	rohdaten, err := paketdatei(ort, path.Clean(datei))
	if err != nil {
		return Symbolzeichnung{}
	}

	baum, err := Zerteile(verweis, rohdaten)
	if err != nil {
		return Symbolzeichnung{}
	}

	zeichnung, _ := Symbolzeichnungsmodell(baum)

	return zeichnung
}

// zahlOder gibt die Zahl eines Knotens oder die Vorgabe, wenn er fehlt.
//
// **Ein Knoten mit `0` ist keine fehlende Angabe.** Deshalb steht hier nicht
// „null heißt Vorgabe": `center x: 0` ist eine Aussage.
func zahlOder(knoten *Knoten, vorgabe float64) float64 {
	if knoten == nil {
		return vorgabe
	}

	return zahlOderNull(knoten)
}
