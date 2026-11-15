package render

import "math"

// **Die Vorsilbenbreite rastet auf ¼ Punkt** —
// #598.
//
// # Die Regel
//
//	double prev_width = 0;
//	size_t line_start = start;
//	for (size_t i = start ; i < end ; ++i) {
//	  …
//	  RealSize s = dc.GetTextExtent(content.substr(line_start - this->start, i - line_start + 1));
//	  out.push_back(CharInfo(RealSize(s.width - prev_width, s.height), …));
//	  prev_width = s.width;
//	}
//
// `FontTextElement::getCharInfo`, `src/render/text/font.cpp`, Zeilen 27–49,
// Fassung f9d9356 vom 2021-07-06. `GetTextExtent` gibt `w / (zoomX ·
// text_scaling)` mit ganzzahligem `w` und `text_scaling = 4`
// (`src/util/rotation.cpp`, Zeilen 344–351).
//
// **Jeder Vorschub ist damit `raster(vorsilbe_i) − raster(vorsilbe_{i−1})`, und
// die Summe über einen Lauf teleskopiert zu `raster(vorsilbe_n)`.** Eine
// einzige Rasterung, nicht eine je Zeichen.
//
// # Drei Ebenen, und nur eine ist richtig
//
// | Ebene | rastern? |
// | --- | --- |
// | je Zeichen | nein — die Differenzen heben sich auf |
// | je Wort (unser [Stueck]) | **nein** — das war der erste Anlauf: 20,679 → 24,728 |
// | je Abschnitt, kumulativ | **ja** |
// | je gesetzter Zeile | nein — `line_start` setzt an `\n` zurück, nicht am Umbruch |
//
// Der dritte Punkt ist der, den man leicht übersieht: der Wortumbruch geschieht
// bei MSE später, in `prepareLines`, und **partitioniert** die schon
// gerechneten Vorschübe. Der Vorschub eines Wortes mitten im Absatz hängt an
// allem, was seit dem letzten harten Umbruch davorsteht.
//
// **Hier ist der Absatz schon an den harten Umbrüchen geteilt** (`absatzband`),
// und ein Abschnitt ist ein Schriftlauf — MSEs Element. Der Rücksetzpunkt fällt
// damit auf den Abschnittsanfang.

// Textraster ist MSEs `text_scaling`.
//
// **Vier**, und die Zahl steht so im Original: `RotatedDC` teilt jede gemessene
// Breite durch `zoomX · text_scaling` (`src/util/rotation.cpp`, Zeile 349), und
// `text_scaling` ist 4, Fassung f9d9356 vom 2021-07-06. Dieselbe Zahl rastert
// die Zeilenhöhe und die Grundlinie (#546).
const Textraster = 4

// gerastet legt eine Breite auf das Viertelpunktraster.
//
// **Aufgerundet, nicht gerundet.** Neun Zeichenketten bei zwei Punktgrößen
// gegen wx 3.3.3: `ceil` trifft 18 von 18, `round` verfehlt drei — „Hallo" und
// „iiiii". Festgehalten als `TestDieBreiteTrifftWx` hinter `fremdpakete`.
//
// **Und in `mplantin.ttf` ist kein Kerning im Spiel**: die ungerundete Summe
// liegt bei allen achtzehn knapp **unter** wx' ganzer Zahl; gäbe es Kerning,
// läge sie mal darüber.
//
// **Das ist eine Aussage über eine Schrift, nicht über alle** —
// #752:
// `mplantin.ttf` trägt **gar keine** Kerningtabelle, und die Messung konnte
// über Kerning nichts sagen. `beleren-bold`
// dagegen führt ein `kern`-Merkmal in `GPOS`, und in ihr steht jeder
// Kartenname. Welche Schriften des Packs kernen, hält
// `kerningreichweite_intern_test.go` fest. **Wie weit es die Breite
// verschiebt, ist seit QA-03 gemessen**: mit beiden Formaten der Paartabelle
// trifft sie Core Text auf ein Tausendstel Punkt, ohne das zweite lag sie im
// Mittel 1,8 Punkt daneben (`docs/SCHRIFTUNTERSUCHUNG.md`).
func gerastet(breite float64) float64 {
	return math.Ceil(breite*Textraster) / Textraster
}

// vorsilbe sammelt die rohe Breite eines Laufs, Stück für Stück.
//
// **Roh und nicht gerastert** — gerastert wird die Summe, und die Differenz
// zweier gerasterter Summen ist die Breite eines Stücks.
type vorsilbe struct {
	summe       float64
	vorherRastr float64

	// letztes ist das zuletzt gezählte Zeichen — für das Paar über die
	// Stücknaht hinweg
	// (#752).
	//
	// **Weil ein Zug in Stücken ankommt**: Wort für Wort beim Umbruch. Wer nur
	// innerhalb eines Stücks kernte, verlöre die Paare an den Wortgrenzen —
	// ` A` und ` W` sind zwei der häufigsten in `beleren-bold`.
	letztes rune

	// ersatz gibt die Schrift für ein Zeichen, das die des Stücks nicht führt
	// — dieselbe, aus der gezeichnet wird ([Schriftsatz.Ersatz]).
	//
	// **Gemessen wie gezeichnet**, seit dem 2026-09-24: MSE misst über
	// `GetTextExtent`, und die Plattform weicht dabei selbst aus. Hier maß das
	// Stück mit seiner eigenen Schrift, die für ein fehlendes Zeichen null
	// meldet; der Puffer war dann so schmal wie das Gemessene, und von der
	// ausgewichenen Glyphe blieb ein Streifen (`zeichenrand_test.go`).
	ersatz Ersatzschrift
}

// nimm hängt ein Stück an und gibt dessen gerasterte Breite.
//
// **Ohne die Laufweite**, und das ist kein Weglassen: MSE rastert, was
// `GetTextExtent` misst — die **Glyphenvorschübe**. Unsere Laufweite ist ein
// Zusatz **darauf**, gesetzt aus `separation:`; sie steht zwischen den
// Buchstaben und wird dort gezählt, wo sie schon gezählt wurde
// ([breiteDerStuecke] und [verteile]).
//
// **Wie weit diese Aussage reicht** (§10): über die 190 Stilpakete von
// `M15-Magic-Pack` steht `separation:` **nullmal** in einem `font:`-Block, und
// kein Weg in `internal/service` setzt [Satz.Laufweite]. Im Betrieb ist sie
// null; geprüft wird sie in `internal/render`.
func (v *vorsilbe) nimm(text string, satz Satz, schrift Schrift) float64 {
	if schrift != nil {
		for _, zeichen := range gesetzt(text, satz, schrift) {
			verwendet := ausgewichen(schrift, v.ersatz, zeichen)
			v.summe += kernungZwischen(v.letztes, zeichen, satz, verwendet)
			v.summe += verwendet.Breite(zeichen) * satz.Groesse
			v.letztes = zeichen
		}
	}

	jetzt := gerastet(v.summe)

	glyphen := jetzt - v.vorherRastr
	v.vorherRastr = jetzt

	// **Die Laufweite kommt auf die gerasterte Breite, nicht hinein.** Sie
	// wird je Stück gezählt und an den Stücknähten von [breiteDerStuecke]
	// nachgetragen — dieselbe Rechnung wie vor #598, unverändert.
	return glyphen + float64(max(buchstabenIn(text)-1, 0))*satz.Laufweite
}

// Zeichenfuehrer ist eine Schrift, die sagt, ob sie ein Zeichen führt.
//
// **Die Frage und nicht die Breite**: ein Zeichen ohne Vorschub — eine
// kombinierende Marke — führt die Schrift trotzdem, und ausgewichen wird nur,
// wo sie es nicht tut. Eine Schrift ohne diese Auskunft gilt als vollständig.
type Zeichenfuehrer interface {
	Fuehrt(zeichen rune) bool
}

// ausgewichen gibt die Schrift, aus der ein Zeichen gemessen wird — die des
// Stücks, oder die Ersatzschrift, wo jene es nicht führt. Dieselbe Wahl
// trifft das Zeichnen ([Rasterer.zeichneStueck]): eine Glyphe aus der einen
// und eine Breite aus der anderen ergäbe einen Satz, den keine meint.
func ausgewichen(schrift Schrift, ersatz Ersatzschrift, zeichen rune) Schrift {
	fuehrer, sagt := schrift.(Zeichenfuehrer)
	if ersatz == nil || !sagt || fuehrer.Fuehrt(zeichen) {
		return schrift
	}

	if andere, gefunden := ersatz(zeichen); gefunden {
		return andere
	}

	return schrift
}
