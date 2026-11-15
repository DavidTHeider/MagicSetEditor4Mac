package render

import (
	"image"
	"strings"
)

// Bildbaum ist ein Bild, wie eine Vorlage es beschreibt.
//
// # Warum ein Baum und kein Pfad
//
// **In MSE gibt ein Bildskript keinen Dateinamen zurück, sondern ein Bild** —
// **fünf Erzeuger und sechzehn Umformungen** — gezählt in
// `src/gfx/generated_image.hpp`, Fassung f9d9356 vom 2021-07-06: 23 Klassen,
// davon `GeneratedImage` die Oberklasse und `SimpleFilterImage` eine
// Zwischenklasse.
//
// **Erzeuger** (aus nichts ein Bild): `BlankImage`, `PackagedImage`,
// `BuiltInImage`, `SymbolToImage`, `ImageValueToImage`.
//
// **Umformungen** (aus Bildern ein Bild): `LinearBlendImage`,
// `MaskedBlendImage`, `CombineBlendImage` und die dreizehn
// `SimpleFilterImage` — `SetMask`, `SetAlpha`, `SetCombine`, `Saturate`,
// `Invert`, `Recolor`, `Recolor2`, `FlipHorizontal`, `FlipVertical`,
// `Rotate`, `Enlarge`, `Crop`, `DropShadow`.
//
// Solange der Plan einen **Pfad** trug, hatte `recolor_image` nichts
// zurückzugeben — und 965 der 2.729 Befunde über `M15-Magic-Pack` lauteten
// genau darauf
// (#270).
//
// # Der leere Baum ist „kein Bild"
//
// Das ist die Stelle, an der vorher die leere Zeichenkette stand. [Leer] sagt
// es, damit niemand die Knotenart gegen `""` prüfen muss.
//
// Fassung f9d9356 vom 2021-07-06.
type Bildbaum struct {
	// Knotenart ist `datei` für ein Blatt, sonst der Name der Verknüpfung.
	Knotenart string

	// Pfad ist der Paketpfad eines Blattes.
	Pfad string

	// Kinder sind die Bilder, aus denen dieser Knoten entsteht — **benannt**.
	//
	// Bei `recolor_image` heißt das eine `image`; bei `masked_blend` heißen die
	// drei `light`, `dark` und `mask`, und welches welches ist, steht im Namen.
	Kinder []Bildkind

	// Farben sind benannte Farbargumente einer Verknüpfung.
	Farben map[string]Farbe

	// Zahlen sind benannte Zahlargumente einer Verknüpfung.
	Zahlen map[string]float64

	// Texte sind die übrigen benannten Argumente.
	Texte map[string]string

	// Symbol ist die **Vektorzeichnung** eines `symbol_variation`-Knotens.
	//
	// **Die Gestalt steht im Plan, nicht im Rasterer.** Wer die `.mse-symbol`
	// liest, ist `internal/service`; hier kommt an, was [RastereSymbol]
	// braucht (Vermerk 0022,
	// #323).
	Symbol []Symbolteil
}

// Bildkind ist ein benanntes Kind eines Bildbaums.
type Bildkind struct {
	Name string
	Baum Bildbaum
}

// Kind gibt das Kind mit diesem Namen.
func (b Bildbaum) Kind(name string) (Bildbaum, bool) {
	for _, kind := range b.Kinder {
		if kind.Name == name {
			return kind.Baum, true
		}
	}

	return Bildbaum{}, false
}

// Blatt ist ein Bild, das nur aus einer Datei besteht.
//
// **Die häufigste Form**, und die einzige, die es vor dem Bildbaum gab: eine
// Ebene, deren `image:` einen Pfad ergibt.
func Blatt(pfad string) Bildbaum {
	// **Auch ein Pfad aus Leerzeichen ist keiner.** Vor dem Bildbaum prüfte
	// `planpruefung.go` mit `strings.TrimSpace`; wer das hier wegließe, machte
	// aus einem gefangenen Mangel eine Ebene, die beim Laden scheitert.
	pfad = strings.TrimSpace(pfad)
	if pfad == "" {
		return Bildbaum{}
	}

	return Bildbaum{Knotenart: "datei", Pfad: pfad}
}

// Leer sagt, ob dieser Baum gar kein Bild beschreibt.
func (b Bildbaum) Leer() bool { return b.Knotenart == "" }

// IstBlatt sagt, ob dieser Baum unmittelbar eine Datei nennt.
func (b Bildbaum) IstBlatt() bool { return b.Knotenart == "datei" }

// Beschreibung nennt den Baum in einer Zeile — für Befunde.
//
// **Der Pfad des tiefsten Blattes**, denn der ist es, den ein Leser wiedererkennt:
// „`/paket/rahmen.png` ist nicht lesbar" hilft, „ein Bildbaum ist nicht lesbar"
// nicht.
// **Und sie hält den Tiefenvertrag** — siehe [Baumtiefe].
func (b Bildbaum) Beschreibung() string { return b.beschreibung(0) }

func (b Bildbaum) beschreibung(tiefe int) string {
	// **Der Abbruch sagt sich.** Eine Beschreibung, die still auf vierzig
	// Ebenen endet, sähe aus wie ein echter Baum.
	if tiefe > Baumtiefe {
		return "…"
	}

	if b.Leer() {
		return "kein Bild"
	}

	if b.IstBlatt() {
		return b.Pfad
	}

	if len(b.Kinder) == 0 {
		return b.Knotenart
	}

	return b.Knotenart + "(" + b.Kinder[0].Baum.beschreibung(tiefe+1) + ")"
}

// Mischart sagt, wie diese Ebene auf das Bisherige gelegt wird.
//
// **Leer heißt „normal"** — das Bild ersetzt, was darunter liegt, mit seinem
// Alphakanal.
//
// # Ein reiner Baumdurchlauf
//
// MSE fragt die Mischart am Baum, **ohne ein Bild zu erzeugen**
// (Fassung f9d9356 vom 2021-07-06):
//
//	GeneratedImage::combine()      → COMBINE_DEFAULT   (generated_image.hpp:51)
//	SimpleFilterImage::combine()   → image->combine()  (ebd. :79)
//	LinearBlendImage::combine()    → image1->combine() (generated_image.cpp:104)
//	MaskedBlendImage::combine()    → light->combine()  (ebd. :122)
//	CombineBlendImage::combine()   → image1->combine() (ebd. :139)
//	SetCombineImage::combine()     → seine eigene      (ebd. :178)
//
// **`masked_blend` nimmt die helle Seite**, nicht die dunkle; die beiden
// Mischungen nehmen ihr erstes Bild. Wer das umdreht, bekommt eine Mischart,
// die im Bild nicht vorkommt.
//
// **Und sie hält den Tiefenvertrag** — siehe [Baumtiefe].
func (b Bildbaum) Mischart() string { return b.mischart(0) }

func (b Bildbaum) mischart(tiefe int) string {
	if tiefe > Baumtiefe || b.Leer() || b.IstBlatt() {
		return ""
	}

	if b.Knotenart == "set_combine" {
		return b.Texte["combine"]
	}

	// **Welches Kind gefragt wird, hängt am Knoten.**
	name := "input"

	switch {
	// **`set_mask` nennt sein Ausgangsbild `image`** und nicht `input` —
	// als einzige Verknüpfung (#483).
	case b.Knotenart == "set_mask":
		name = "image"
	case einfacheFilter[b.Knotenart]:
	case b.Knotenart == "masked_blend":
		name = "light"
	case b.Knotenart == "linear_blend", b.Knotenart == "combine_blend":
		name = "image1"
	default:
		return ""
	}

	kind, da := b.Kind(name)
	if !da {
		return ""
	}

	return kind.mischart(tiefe + 1)
}

// einfacheFilter sind die Unterklassen von `SimpleFilterImage` — sie reichen
// die Mischart ihres Kindes durch (`generated_image.hpp`, Zeilen 153–308).
//
// **`set_combine` steht nicht darin**, obwohl es eine ist: es überschreibt
// `combine()` und wird davor abgefangen.
var einfacheFilter = map[string]bool{
	"set_mask":        true,
	"set_alpha":       true,
	"saturate":        true,
	"invert_image":    true,
	"recolor_image":   true,
	"flip_horizontal": true,
	"flip_vertical":   true,
	"rotate":          true,
	"enlarge":         true,
	"crop":            true,
	"drop_shadow":     true,
}

// gebauteVerknuepfungen sind die Knotenarten, die [Rasterer.ausBaum] ausführt
// — **und zugleich, wie sie ausgeführt werden.**
//
// # Warum eine Liste und kein `switch` daneben
//
// Bis zum 2026-09-03 standen hier Namen und dort ein `switch`: **zwei
// Wahrheiten über dieselbe Frage.** Der Nachlauf über die Phasen 4.0 bis 4.9
// hat es angestrichen und die Gegenprobe genannt — wer `case "rotate"` aus
// dem Verteiler entfernt, lässt `rotate` in der Liste stehen, und der Wächter
// darüber bleibt **grün**, obwohl der Fehler wiederhergestellt ist, gegen den
// er gebaut wurde.
//
// Nachgestellt und gesehen: mit entferntem `case` war
// `TestJedeRegistrierteBildfunktionIstGebaut` grün; rot wurden allein die
// Rendertests. Der Wächter hielt eine Behauptungsliste, nicht die Naht.
//
// **Jetzt trägt jeder Eintrag seine Ausführung.** Eine Knotenart, die
// dasteht, wird ausgeführt — anders geht es nicht mehr.
// **Sie wird in `init` gefüllt und nicht als Literal geschrieben.** Go weist
// einen Initialisierungszyklus ab: `drehe` ruft `ausBaumTief`, das über
// `ausBaumRoh` wieder auf diese Liste sieht. Ein `init` bricht ihn, ohne die
// Aussage zu ändern — die Liste ist zur Laufzeit dieselbe.
var gebauteVerknuepfungen = map[string]verknuepfung{}

func init() {
	gebauteVerknuepfungen = map[string]verknuepfung{
		"datei":            {},
		"blank_image":      {},
		"symbol_variation": {},
		"built_in_image":   {},
		Platzhalterknoten:  {},

		"masked_blend":    {fuehreAus: (*Rasterer).mischeUeberMaske},
		"linear_blend":    {fuehreAus: (*Rasterer).blendeLinear},
		"recolor_image":   {fuehreAus: (*Rasterer).faerbeUm},
		"set_alpha":       {fuehreAus: (*Rasterer).setzeDeckung},
		"set_combine":     {fuehreAus: (*Rasterer).reicheDurch},
		"combine_blend":   {fuehreAus: (*Rasterer).mischeArten},
		"set_mask":        {fuehreAus: (*Rasterer).setzeMaske},
		"enlarge":         {fuehreAus: (*Rasterer).vergroessere},
		"drop_shadow":     {fuehreAus: (*Rasterer).schlagschatten},
		"saturate":        {fuehreAus: (*Rasterer).saettige},
		"invert_image":    {fuehreAus: (*Rasterer).kehreUm},
		"crop":            {fuehreAus: (*Rasterer).schneide},
		"flip_horizontal": {fuehreAus: (*Rasterer).spiegeleWaagerecht},
		"flip_vertical":   {fuehreAus: (*Rasterer).spiegeleSenkrecht},
		"rotate":          {fuehreAus: (*Rasterer).drehe},
	}
}

// verknuepfung ist eine Knotenart samt ihrer Ausführung.
//
// **`fuehreAus` fehlt bei den vier Erzeugern** — `datei`, `blank_image`,
// `symbol_variation` und `built_in_image` haben kein Ausgangsbild, das man
// umformen könnte; sie *sind* die Quelle
// (#509) und
// werden vor dem Verteiler behandelt. Sie stehen trotzdem in der Liste, weil
// [Bildbaum.Gebaut] sie kennen muss.
type verknuepfung struct {
	fuehreAus func(*Rasterer, Bildbaum, int) (image.Image, string)
}

// Gebaut sagt, ob dieser Knoten **und alle darunter** ausgeführt werden können.
//
// # Wozu eine Messung das braucht
//
// Seit dem Bildbaum zeichnet eine nicht gebaute Verknüpfung ihr Ausgangsbild.
// Das ist richtig — sichtbar falsch schlägt unsichtbar fehlend —, aber es macht
// eine Zählung über **Planbefunde** blind: der Plan trägt eine Bildebene, und
// dass sie nur zur Hälfte stimmt, steht erst im Rasterbefund.
//
// Der Zweitblick über die Phasen B bis D hat genau das angestrichen: „die 965
// früheren `recolor_image`-Fehler können aus dem Zähler verschwinden, obwohl
// die Verknüpfung weiterhin nur als Ausgangsbild gezeichnet wird".
//
// **Eine Messgröße darf ihre Bedeutung nicht stillschweigend wechseln.**
// **Und sie hält den Tiefenvertrag** — siehe [Baumtiefe].
func (b Bildbaum) Gebaut() bool { return b.gebaut(0) }

func (b Bildbaum) gebaut(tiefe int) bool {
	// **Jenseits der Grenze gilt der Baum als gebaut.** Er wird dort nicht
	// mehr gezeichnet, sondern abgewiesen ([Rasterer.ausBaumTief]); ihn hier
	// als „nicht gebaut" zu zählen, verschöbe einen Zähler über eine
	// Verknüpfung, die es gar nicht bis zur Ausführung schafft.
	if tiefe > Baumtiefe {
		return true
	}

	if _, gebaut := gebauteVerknuepfungen[b.Knotenart]; !gebaut {
		return false
	}

	for _, kind := range b.Kinder {
		if !kind.Baum.gebaut(tiefe + 1) {
			return false
		}
	}

	return true
}

// Blaetter gibt die Paketpfade aller Blätter dieses Baums.
//
// **Alle und nicht das erste.** Der Messwächter prüfte bis zum 2026-08-16 nur
// den Pfad am Wurzelknoten — bei einem Baum ist der **leer**, und damit prüfte
// er nichts.
// **Und sie hält den Tiefenvertrag** — siehe [Baumtiefe].
func (b Bildbaum) Blaetter() []string { return b.blaetter(0) }

func (b Bildbaum) blaetter(tiefe int) []string {
	if tiefe > Baumtiefe {
		return nil
	}

	if b.IstBlatt() {
		return []string{b.Pfad}
	}

	var pfade []string
	for _, kind := range b.Kinder {
		pfade = append(pfade, kind.Baum.blaetter(tiefe+1)...)
	}

	return pfade
}
