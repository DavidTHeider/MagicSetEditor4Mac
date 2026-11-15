package domain

// **Ein KI-Modell, so wie dieses Programm es kennt** — Werkstück A6 von
// Phase 5.3 (#1445).
//
// # Vier Dinge, und sie sind getrennt
//
// Das Issue verlangt es zuerst: „KIProvider, Modell-ID, Fähigkeit und Aufgabe
// getrennt." Der **Anbieter** ist, wen man fragt; der **Alias** ist, wonach
// man fragt; die **Fähigkeit** ist, was dabei herauskommen kann; die
// **Aufgabe** ist, wozu man es tut — und die steht nicht hier, sondern beim
// Auftrag.
//
// # Der Alias ist ein Name und kein Ding
//
// **Ein Anbieter darf tauschen, was hinter seinem Namen liegt.** `beispiel:3b`
// heißt heute dies und morgen das; das Issue verlangt deshalb
// „Modellaliasänderung **sichtbar**". Die [KIModell.Bindung] ist Alias **und**
// ausgeführtes Artefakt — bleibt sie gleich, ist es dasselbe.

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// KIModell ist ein Modell samt dem, was über es feststeht.
type KIModell struct {
	// Alias ist der Name, unter dem gerufen wird.
	Alias string

	// Anbieter ist, wen man fragt — `ollama`, `anthropic`, …
	Anbieter string

	// Ort ist, **wo** es rechnet — wie der Anbieter es erklärt.
	//
	// **Erklärt und nicht gemessen.** Das Issue sagt es scharf: „**Loopback
	// allein beweist keine lokale Inferenz**." Ein Dienst auf `127.0.0.1`
	// kann weiterleiten; die Adresse sagt, wohin gefragt wird, nicht, wo
	// gerechnet wird. Deshalb steht der Ort als Angabe hier — und wer ihn
	// falsch setzt, hat gelogen, nicht sich geirrt.
	Ort KIOrt

	// Laufzeit ist, womit es ausgeführt wird, samt Fassung.
	//
	// **Sie gehört zum Beleg**: dieselbe Datei tut unter zwei Laufzeiten
	// nicht dasselbe, und eine Fähigkeit ist immer die einer Kombination.
	Laufzeit string

	// Geprueft sind die Fähigkeiten **mit Beleg** — und nur die zählen.
	Geprueft []KIGeprueft

	// Herkunft ist die Kette bis zum ausgeführten Artefakt — leer heißt
	// **unbekannt** und ist erlaubt ([KIHerkunft]).
	Herkunft KIHerkunft

	// Einbettung ist der Vertrag, wenn dieses Modell einbettet — sonst nil.
	//
	// **Beides hängt zusammen**: wer [KIEinbettung] belegt, hat einen
	// Vertrag, und wer keinen Vektor liefert, verspricht auch keine Form.
	Einbettung *KIEinbettungsvertrag

	// Kontextfenster ist, wie viele Tokens das Modell auf einmal liest — 0
	// heißt: niemand sagt es (AI-01.AK-02, #1742). KontextBeleg sagt,
	// **woher** die Zahl stammt: angegeben, gelesen oder gemessen — eine Zahl
	// ohne Herkunft wäre die Behauptung, die A6 ausschließt.
	Kontextfenster int
	KontextBeleg   string
}

// Kann sagt, ob die Fähigkeit **belegt** ist.
//
// **Nicht »steht in einer Liste«** (Issue): ein Name in einem Katalog sagt,
// dass es das Modell gibt, nicht, dass es dieses kann.
func (m KIModell) Kann(faehigkeit KIFaehigkeit) bool {
	for _, geprueft := range m.Geprueft {
		if geprueft.Faehigkeit == faehigkeit && len(geprueft.Validate()) == 0 {
			return true
		}
	}

	return false
}

// Validate prüft Alias, Anbieter, Laufzeit, Belege, Herkunft und Vertrag.
func (m KIModell) Validate() []*FieldError {
	var befunde []*FieldError

	for feld, wert := range map[string]string{
		"alias": m.Alias, "anbieter": m.Anbieter, "laufzeit": m.Laufzeit,
	} {
		if strings.TrimSpace(wert) == "" {
			befunde = append(befunde, NewFieldError(feld, SchluesselPflichtfeld,
				"Ein Modell nennt seinen %s.", feld))
		}
	}

	if !IstKIOrt(m.Ort) {
		befunde = append(befunde, NewFieldError("ort", SchluesselValidierung,
			"%q ist kein bekannter Ort — es gibt »lokal« und »cloud«.",
			string(m.Ort)))
	}

	for _, geprueft := range m.Geprueft {
		befunde = append(befunde, geprueft.Validate()...)
	}

	if fehler := m.Herkunft.Pruefe(); fehler != nil {
		befunde = append(befunde, NewFieldError("herkunft", SchluesselValidierung,
			"%s", fehler.Error()))
	}

	return append(befunde, m.pruefeEinbettung()...)
}

// pruefeEinbettung hält Fähigkeit und Vertrag zusammen.
func (m KIModell) pruefeEinbettung() []*FieldError {
	bettetEin := m.Kann(KIEinbettung)

	switch {
	case bettetEin && m.Einbettung == nil:
		return []*FieldError{NewFieldError("einbettung", SchluesselPflichtfeld,
			"Ein Modell, das einbettet, nennt seinen Einbettungsvertrag — "+
				"ohne ihn ist an einem Vektor nichts zu prüfen.")}

	case !bettetEin && m.Einbettung != nil:
		return []*FieldError{NewFieldError("einbettung", SchluesselValidierung,
			"Dieses Modell hat keine belegte Einbettung und trägt trotzdem "+
				"einen Vertrag — ein Versprechen ohne Rumpf.")}

	case bettetEin:
		if fehler := m.Einbettung.Pruefe(); fehler != nil {
			return []*FieldError{NewFieldError("einbettung",
				SchluesselValidierung, "%s", fehler.Error())}
		}
	}

	return nil
}

// KIBindungUnbekannt steht in der Bindung, wo es keine Herkunft gibt.
const KIBindungUnbekannt = "unbekannt"

// Bindung ist Alias **und** ausgeführtes Artefakt.
//
// # Wofür es sie gibt
//
// **Damit ein Aliaswechsel auffällt.** Ein Anbieter tauscht hinter
// `beispiel:3b` ein Modell aus, und der Name bleibt derselbe; die Bindung
// ändert sich. Das Issue verlangt genau das.
//
// **Ohne Herkunft steht »unbekannt« darin** — nicht nichts und nichts
// Geratenes: eine leere Bindung sähe aus wie „nichts gebunden", und eine aus
// dem Alias abgeleitete wäre die erfundene Herkunftsangabe, die das Issue
// verbietet.
func (m KIModell) Bindung() string {
	artefakt := KIBindungUnbekannt

	for _, glied := range m.Herkunft.Glieder {
		if glied.Art == KIArtefakt {
			artefakt = glied.Quelle + "@" + glied.Revision

			break
		}
	}

	summe := sha256.Sum256([]byte(strings.Join(
		[]string{m.Anbieter, m.Alias, artefakt}, "\x1f")))

	return hex.EncodeToString(summe[:])
}
