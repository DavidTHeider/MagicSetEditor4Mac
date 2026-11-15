package domain

// **Woher ein KI-Modell kommt** — Werkstück A6 von Phase 5.3
// (#1445).
//
// # Warum eine Kette und kein Feld
//
// **Weil zwischen dem, was jemand veröffentlicht hat, und dem, was hier
// läuft, mehrere Schritte liegen.** Ein Grundmodell wird mit einem Adapter
// verfeinert, das Ergebnis in ein anderes Format gegossen und quantisiert,
// und was die Laufzeit dann lädt, ist eine Datei mit einem eigenen Hash. Ein
// einzelnes Feld „Modell" beantwortete die Frage „was ist da gelaufen" nicht.
//
// # Unbekannt ist eine Auskunft
//
// **Eine leere Kette ist erlaubt und heißt: wir wissen es nicht.** Das Issue
// sagt es scharf: „Fehlende Herkunft explizit unbekannt; **kein erratener
// HF-Commit aus einem Ollama-Alias**." Ein Alias nennt einen Namen, keine
// Revision — wer daraus eine ableitet, hat eine Herkunftsangabe erfunden, und
// die ist schlimmer als keine.
//
// **Und eine halbe Kette gibt es nicht.** Wer Glieder nennt, nennt sie
// vollständig: Quelle, unveränderliche Revision, Dateihashes, Lizenz. Sonst
// stünde neben jedem Glied die Frage, ob das Fehlende „unbekannt" oder
// „vergessen" heißt.
//
// # Was eine Kette **nicht** belegt
//
// **Weder sichere Ausführung noch lokale Inferenz** — das Issue sagt auch
// das. Ein Hash sagt, welche Bytes geladen wurden; er sagt nicht, was sie tun,
// und schon gar nicht, wo sie es tun. Die Netzfrage entscheidet der Transport
// und nicht diese Kette.

import "fmt"

// KIHerkunftsart ist ein Glied der Kette.
type KIHerkunftsart string

// Die vier Arten.
const (
	// KIBasis ist das veröffentlichte Grundmodell.
	KIBasis KIHerkunftsart = "basis"

	// KIAdapter ist eine Verfeinerung darauf — LoRA und Verwandtes.
	KIAdapter KIHerkunftsart = "adapter"

	// KIUmwandlung ist ein Formatwechsel oder eine Quantisierung.
	KIUmwandlung KIHerkunftsart = "umwandlung"

	// KIArtefakt ist, was die Laufzeit **wirklich geladen** hat.
	//
	// **Ohne dieses Glied beschreibt eine Kette, was es gibt, und nicht, was
	// gelaufen ist.**
	KIArtefakt KIHerkunftsart = "artefakt"
)

var kiHerkunftsarten = []KIHerkunftsart{
	KIBasis, KIAdapter, KIUmwandlung, KIArtefakt,
}

// KIHoechstensGlieder begrenzt die Kette.
//
// **Acht ist großzügig und trotzdem eine Grenze**: Basis, zwei Adapter, zwei
// Umwandlungen und ein Artefakt sind sechs. Ohne Grenze wäre eine Kette mit
// tausend Gliedern zulässig, und was sie belegte, könnte niemand lesen.
const KIHoechstensGlieder = 8

// KIDateihash ist eine Datei mit ihrer Prüfsumme.
type KIDateihash struct {
	Name string
	Hash string
}

// KIGlied ist ein Schritt auf dem Weg zum ausgeführten Modell.
type KIGlied struct {
	Art KIHerkunftsart

	// Kennung ist der Name dieses Glieds **innerhalb der Kette** — sie
	// verbindet es mit dem nächsten.
	Kennung string

	// Aus nennt das Glied, aus dem dieses hervorging — leer bei der Wurzel.
	Aus string

	// Quelle ist das Repositorium oder der Ort, an dem es liegt.
	Quelle string

	// Revision ist die **unveränderliche** Kennzeichnung — ein Commit, ein
	// Digest, ein Datum bei einer eigenen Umwandlung.
	//
	// **Ein Zweigname ist keine**: er zeigt morgen woandershin.
	Revision string

	Dateien []KIDateihash

	// Lizenz ist der Lizenzbeleg dieses Glieds.
	//
	// **Je Glied und nicht je Modell**: ein Adapter kann anders lizenziert
	// sein als seine Basis, und wer nur eine Lizenz führte, nennte die
	// falsche.
	Lizenz string
}

// KIHerkunft ist die ganze Kette.
type KIHerkunft struct{ Glieder []KIGlied }

// Unbekannt sagt, ob es gar keine Angabe gibt.
func (h KIHerkunft) Unbekannt() bool { return len(h.Glieder) == 0 }

// Pruefe sagt, ob die Kette vollständig, begrenzt und azyklisch ist.
//
// **Eine leere Kette geht durch** — sie ist die Auskunft „unbekannt".
func (h KIHerkunft) Pruefe() error {
	if h.Unbekannt() {
		return nil
	}

	if len(h.Glieder) > KIHoechstensGlieder {
		return fmt.Errorf("%w: eine Herkunft hat höchstens %d Glieder, diese hat %d",
			ErrValidation, KIHoechstensGlieder, len(h.Glieder))
	}

	nachKennung := map[string]KIGlied{}

	for _, glied := range h.Glieder {
		if fehler := glied.pruefe(); fehler != nil {
			return fehler
		}

		if _, doppelt := nachKennung[glied.Kennung]; doppelt {
			return fmt.Errorf("%w: die Kennung %q steht zweimal in der Kette",
				ErrValidation, glied.Kennung)
		}

		nachKennung[glied.Kennung] = glied
	}

	if fehler := h.pruefeKanten(nachKennung); fehler != nil {
		return fehler
	}

	return h.pruefeArtefakt()
}

// pruefeKanten prüft Vorgänger und Kreisfreiheit.
func (h KIHerkunft) pruefeKanten(nachKennung map[string]KIGlied) error {
	for _, glied := range h.Glieder {
		if glied.Aus == "" {
			continue
		}

		if _, gibts := nachKennung[glied.Aus]; !gibts {
			return fmt.Errorf("%w: %q stammt aus %q — das Glied gibt es in "+
				"dieser Kette nicht", ErrValidation, glied.Kennung, glied.Aus)
		}

		// **Der Weg nach oben wird abgelaufen**, und zwar höchstens so weit,
		// wie die Kette lang ist. Wer danach noch läuft, läuft im Kreis —
		// nichts stammt von sich selbst.
		bei, schritte := glied, 0

		for bei.Aus != "" {
			schritte++

			if schritte > len(h.Glieder) {
				return fmt.Errorf("%w: die Herkunft von %q läuft im Kreis",
					ErrValidation, glied.Kennung)
			}

			bei = nachKennung[bei.Aus]
		}
	}

	return nil
}

// pruefeArtefakt verlangt genau ein ausgeführtes Artefakt.
func (h KIHerkunft) pruefeArtefakt() error {
	artefakte := 0

	for _, glied := range h.Glieder {
		if glied.Art == KIArtefakt {
			artefakte++
		}
	}

	if artefakte != 1 {
		return fmt.Errorf("%w: eine Herkunft nennt genau **ein** ausgeführtes "+
			"Artefakt, diese nennt %d — ohne es beschreibt sie, was es gibt, "+
			"und nicht, was gelaufen ist", ErrValidation, artefakte)
	}

	return nil
}

// pruefe sagt, ob ein Glied vollständig ist.
func (g KIGlied) pruefe() error {
	if !kennt(kiHerkunftsarten, g.Art) {
		return fmt.Errorf("%w: %q ist keine Art eines Herkunftsglieds",
			ErrValidation, string(g.Art))
	}

	for feld, wert := range map[string]string{
		"Kennung": g.Kennung, "Quelle": g.Quelle,
		"Revision": g.Revision, "Lizenz": g.Lizenz,
	} {
		if wert == "" {
			return fmt.Errorf("%w: dem Glied %q fehlt %s — eine halbe Herkunft "+
				"ist schlimmer als keine", ErrValidation, g.Kennung, feld)
		}
	}

	if len(g.Dateien) == 0 {
		return fmt.Errorf("%w: das Glied %q nennt keine Datei mit Prüfsumme",
			ErrValidation, g.Kennung)
	}

	for _, datei := range g.Dateien {
		if datei.Name == "" || datei.Hash == "" {
			return fmt.Errorf("%w: das Glied %q nennt eine Datei ohne Namen "+
				"oder ohne Prüfsumme", ErrValidation, g.Kennung)
		}
	}

	return nil
}
