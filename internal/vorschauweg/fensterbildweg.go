package vorschauweg

// **Der Weg zum eigenen Fensterbild** — UX-09.AK-01
// (#1681).
//
// **Eine Adresse und keine Bytes**, aus demselben Grund wie beim Tischbild:
// ein Hintergrund von 2.560 Punkten käme als Base64 über die Brücke, bei
// jedem Start.

import "net/http"

// Fensterpfad ist der Anfang jeder Fensterbildadresse.
const Fensterpfad = "/fensterbild/"

// Fensterbildquelle gibt das abgelegte Fensterbild eines Menschen.
type Fensterbildquelle interface {
	Fensterbild(nutzerID string) ([]byte, bool)
}

// MitFensterbildern hängt die Quelle an den Weg.
func (w *Weg) MitFensterbildern(quelle Fensterbildquelle) *Weg {
	w.fensterbilder = quelle

	return w
}

// fensterbild beantwortet den Abruf eines Fensterbildes.
func (w *Weg) fensterbild(schreiben http.ResponseWriter, anfrage *http.Request) {
	if w.fensterbilder == nil {
		http.NotFound(schreiben, anfrage)

		return
	}

	eigenesBild(schreiben, anfrage, Fensterpfad, w.fensterbilder.Fensterbild)
}
