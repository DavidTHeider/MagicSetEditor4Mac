package domain

// **Was ein Index über sich selbst weiß** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Der Satz, um den herum das gebaut ist
//
// **„Gleich langer Vektor bedeutet nicht gleicher Vektorraum."** So steht es
// im Issue, und es ist die teuerste Verwechslung dieses Werkstücks: zwei
// Modelle mit 768 Zahlen liefern Listen derselben Länge und nichts, was sich
// vergleichen ließe. Wer einen Index aus Modell A mit Modell B befragt,
// bekommt Treffer — falsche, geordnet nach einem Abstand, der nichts bedeutet,
// und **ohne Meldung**.
//
// # Deshalb der Fingerabdruck
//
// Er nennt alles, was den Raum bestimmt: Modell, Revision, Artefakt,
// Laufzeit, Abstandsmaß und den **Einbettungsvertrag** aus A6 — Zerteiler,
// Anweisungen, Pooling, Normalisierung, Dimension.
//
// **Was er nicht nennt, ist der Korpus.** Ein Dokument mehr ändert nicht, wie
// verglichen wird; wer ihn mitzählte, baute bei jedem Nachladen neu.
//
// **Und der Reranker steht daneben.** Er ordnet um, was gefunden wurde; er
// bestimmt nicht, wie gesucht wird, und darf einen Index nicht ungültig
// machen. Das Issue sagt es als eigene Zeile: „Reranker hat eigenen
// Fingerprint."

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// Abstandsmass sagt, wie zwei Vektoren verglichen werden.
type Abstandsmass string

// Die zwei Maße.
//
// **Zwei, und sie sind nicht austauschbar.** Bei normalisierten Vektoren
// ordnen Cosinus und Skalarprodukt gleich; bei unnormalisierten nicht — und
// welche es sind, sagt der Vertrag.
const (
	AbstandCosinus Abstandsmass = "cosinus"
	AbstandPunkt   Abstandsmass = "punkt"
)

var abstandsmasse = []Abstandsmass{AbstandCosinus, AbstandPunkt}

// Abstandsmasse gibt die zwei in ihrer Reihenfolge.
func Abstandsmasse() []Abstandsmass { return slices.Clone(abstandsmasse) }

// Korpusstand sagt, in welchem Stand eine Quelle im Index liegt.
//
// **Je Quelle einer und nicht einer für alle.** Oracle-Texte, Rulings und das
// Regelwerk haben verschiedene Stände; ein gemeinsamer wäre die Mischung, die
// das Issue ausschließt: „keine Mischung verschiedener Regelstände als
// aktueller Stand."
type Korpusstand struct {
	Quelle Retrievalquelle

	// Revision ist der Stand, unter dem die Quelle geführt wird.
	Revision string

	// Datum ist das Gültigkeitsdatum — leer, wo die Quelle keines nennt.
	Datum string
}

// Vektorraum ist, was zwei Vektoren überhaupt vergleichbar macht.
//
// **Sechs Angaben, und keine ist entbehrlich.** Zwei Modelle mit derselben
// Dimension, derselben Zusammenfassung und demselben Zerteiler liefern
// trotzdem zwei verschiedene Räume — deshalb stehen Modell, Revision,
// Artefakt und Laufzeit **neben** dem Vertrag und nicht in ihm.
//
// **Und das Abstandsmaß gehört dazu**, weil es die Ordnung bestimmt: dieselben
// Vektoren, einmal nach Cosinus und einmal nach Skalarprodukt geordnet, sind
// zwei verschiedene Trefferlisten, sobald jemand die Normalisierung abschaltet.
type Vektorraum struct {
	Modell         string
	Modellrevision string

	// Artefakthash ist der Abdruck der Gewichte — leer, wo die Laufzeit
	// keinen herausgibt.
	Artefakthash string

	// Laufzeit nennt, was gerechnet hat (`ollama 0.33.3`).
	Laufzeit string

	Abstandsmass Abstandsmass

	// Vertrag ist der Einbettungsvertrag aus A6.
	Vertrag KIEinbettungsvertrag
}

// Abdruck ist die Kennung dieses Raums.
func (r Vektorraum) Abdruck() string { return abdruckVon(r.teile()) }

// PasstZu sagt, ob in diesem Raum gesucht werden darf, was jener gebaut hat.
//
// **Die Frage ist nicht „gleich lang?", sondern „gleicher Raum?".** Ein
// Encoder, der 384 Zahlen liefert, liefert damit noch lange nicht **diese**
// 384 Zahlen.
func (r Vektorraum) PasstZu(anderer Vektorraum) bool {
	return r.Abdruck() == anderer.Abdruck()
}

// teile sind die Angaben in fester Reihenfolge.
func (r Vektorraum) teile() []string {
	return []string{
		r.Modell, r.Modellrevision, r.Artefakthash, r.Laufzeit,
		string(r.Abstandsmass), r.Vertrag.Fingerabdruck(),
	}
}

// Indexmanifest beschreibt einen gebauten Index vollständig.
type Indexmanifest struct {
	// Schemafassung ist die Fassung des Dokumentvertrags.
	Schemafassung int

	// Zerteilfassung nennt Abschnittsbildung und Normalisierung.
	Zerteilfassung string

	// Indexfassung ist das Format des Index selbst.
	Indexfassung int

	Modell         string
	Modellrevision string

	// Artefakthash ist der Abdruck der Gewichte — leer, wo die Laufzeit
	// keinen herausgibt.
	Artefakthash string

	// Laufzeit nennt, was gerechnet hat (`ollama 0.33.3`).
	Laufzeit string

	Abstandsmass Abstandsmass

	// Vertrag ist der Einbettungsvertrag aus A6.
	Vertrag KIEinbettungsvertrag

	Korpusstaende []Korpusstand

	// Suchkonfiguration nennt das Verfahren der Zusammenführung (`rrf-60`).
	Suchkonfiguration string

	// Rerankerabdruck ist der Fingerabdruck des Rerankers — leer heißt:
	// keiner.
	Rerankerabdruck string
}

// Pruefe sagt, ob das Manifest vollständig ist.
func (m Indexmanifest) Pruefe() error {
	for name, wert := range map[string]string{
		"eine Zerteilfassung":    m.Zerteilfassung,
		"ein Modell":             m.Modell,
		"eine Modellrevision":    m.Modellrevision,
		"eine Laufzeit":          m.Laufzeit,
		"eine Suchkonfiguration": m.Suchkonfiguration,
	} {
		if strings.TrimSpace(wert) == "" {
			return fmt.Errorf("%w: ein Indexmanifest braucht %s", ErrValidation, name)
		}
	}

	if m.Schemafassung < 1 || m.Indexfassung < 1 {
		return fmt.Errorf(
			"%w: Schema- und Indexfassung zählen ab eins — hier stehen %d und %d",
			ErrValidation, m.Schemafassung, m.Indexfassung)
	}

	if !slices.Contains(abstandsmasse, m.Abstandsmass) {
		return fmt.Errorf("%w: %q ist kein Abstandsmaß", ErrValidation, m.Abstandsmass)
	}

	if fehler := m.Vertrag.Pruefe(); fehler != nil {
		return fehler
	}

	return m.pruefeKorpus()
}

// pruefeKorpus hält, dass jede Quelle ihren Stand nennt.
func (m Indexmanifest) pruefeKorpus() error {
	if len(m.Korpusstaende) == 0 {
		return fmt.Errorf(
			"%w: ein Indexmanifest ohne Korpusstand beschreibt einen leeren Index",
			ErrValidation)
	}

	for _, stand := range m.Korpusstaende {
		if !slices.Contains(retrievalquellen, stand.Quelle) {
			return fmt.Errorf("%w: %q ist keine Retrievalquelle",
				ErrValidation, stand.Quelle)
		}

		if strings.TrimSpace(stand.Revision) == "" {
			return fmt.Errorf("%w: der Stand von %q nennt keine Revision",
				ErrValidation, stand.Quelle)
		}
	}

	return nil
}

// Fingerabdruck gibt den Abdruck **dieses Index** — Raum und Verfahren.
//
// **Er zählt den Reranker mit und den Korpus nicht.** Zwei Indizes mit
// demselben Abdruck sind austauschbar; einer mit anderem Reranker ordnet
// anders und ist deshalb nicht derselbe Index — er sucht aber im selben Raum
// ([Indexmanifest.PasstZu]).
func (m Indexmanifest) Fingerabdruck() string {
	return abdruckVon(append(m.raumteile(),
		m.Suchkonfiguration, m.Rerankerabdruck,
		strconv.Itoa(m.Schemafassung), strconv.Itoa(m.Indexfassung),
		m.Zerteilfassung))
}

// Raum gibt den Vektorraum, in dem dieser Index liegt.
//
// **Er steht neben dem Manifest und nicht darin**, weil ein Encoder ihn
// nennen kann, ohne einen Index gebaut zu haben: wer sucht, vergleicht seinen
// Raum mit diesem — Korpus, Reranker und Suchkonfiguration sind dabei ohne
// Belang.
func (m Indexmanifest) Raum() Vektorraum {
	return Vektorraum{
		Modell: m.Modell, Modellrevision: m.Modellrevision,
		Artefakthash: m.Artefakthash, Laufzeit: m.Laufzeit,
		Abstandsmass: m.Abstandsmass, Vertrag: m.Vertrag,
	}
}

// PasstZu sagt, ob ein Encoder dieses Manifests im Raum des anderen suchen
// darf.
//
// **Modell, Revision, Artefakt, Laufzeit, Abstandsmaß und der ganze
// Einbettungsvertrag** müssen übereinstimmen; Korpus, Reranker und
// Suchkonfiguration nicht.
func (m Indexmanifest) PasstZu(anderer Indexmanifest) bool {
	return m.Raum().PasstZu(anderer.Raum())
}

// raumteile sind die Angaben, die den Vektorraum bestimmen.
func (m Indexmanifest) raumteile() []string { return m.Raum().teile() }

// abdruckVon fasst Teile zu einem Abdruck zusammen.
//
// **Mit einem Trennzeichen, das in keinem Teil vorkommt.** Ohne es wären
// `("ab", "c")` und `("a", "bc")` derselbe Abdruck — und zwei verschiedene
// Räume sähen gleich aus.
func abdruckVon(teile []string) string {
	summe := sha256.Sum256([]byte(strings.Join(teile, "\x1f")))

	return hex.EncodeToString(summe[:])
}

// RetrievalschemaFassung und RetrievalindexFassung sind die Fassungen, unter
// denen dieses Programm baut.
//
// **Sie stehen in `domain` und nicht im Aufrufer**, weil sie eine Eigenschaft
// des Vertrags sind und keine Wahl: ein Index, der unter Schemafassung 1
// gebaut wurde, enthält die Felder von [Retrievaldokument] in dieser Fassung.
// Wer das Dokument ändert, zählt hier hoch — und ein alter Index wird
// dadurch **sichtbar** alt, statt still falsch gelesen zu werden.
const (
	RetrievalschemaFassung = 1
	RetrievalindexFassung  = 1
)
