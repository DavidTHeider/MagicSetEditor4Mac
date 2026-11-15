package domain

// **Der Wunsch nach einem Bild und das, was über das fertige feststeht** —
// Werkstück A12 von Phase 5.3
// (#1451).
//
// # Warum Größe und Güte Aufzählungen sind
//
// **Weil beide in den Preis eingehen.** Ein freies Textfeld für die Größe
// hieße, dass jemand `4000x4000` tippt und einen 400 bekommt, für den er
// gewartet hat; ein freies Feld für die Güte hieße, dass er `max` tippt und
// dafür bezahlt.
//
// **`auto` steht nicht darunter, und das ist Absicht**
// ([0102](../../docs/entscheidungen/0102-wo-ein-erzeugtes-bild-liegt.md) §6):
// über einem Kostendeckel ist die eine Angabe, die man nicht machen will,
// „entscheide du, wie teuer".

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

// Die drei Größen, die dieses Programm anbietet.
//
// **Dokumentiert sind mehr** — freie Maße in Vielfachen von 16. Angeboten
// werden die drei benannten: eine Kartenillustration ist hoch, ein Hintergrund
// quer, und alles andere ist ein Sonderfall, den niemand verlangt hat.
const (
	KIBildgroesseQuadrat = "1024x1024"
	KIBildgroesseQuer    = "1536x1024"
	KIBildgroesseHoch    = "1024x1536"
)

// Die drei Güten, die dieses Programm anbietet.
const (
	KIBildgueteNiedrig = "low"
	KIBildgueteMittel  = "medium"
	KIBildgueteHoch    = "high"
)

// KIBildaufgabeHoechstens ist die Länge, ab der eine Aufgabe abgewiesen wird.
//
// **Zweitausend Zeichen.** Das ist mehr, als eine Bildbeschreibung je braucht,
// und weniger, als ein versehentlich eingefügtes Kapitel hat.
const KIBildaufgabeHoechstens = 2000

var (
	kiBildgroessen = []string{
		KIBildgroesseQuadrat, KIBildgroesseQuer, KIBildgroesseHoch,
	}
	kiBildgueten = []string{
		KIBildgueteNiedrig, KIBildgueteMittel, KIBildgueteHoch,
	}
)

// KIBildgroessen gibt die drei in ihrer Reihenfolge.
func KIBildgroessen() []string { return slices.Clone(kiBildgroessen) }

// KIBildgueten gibt die drei in ihrer Reihenfolge.
func KIBildgueten() []string { return slices.Clone(kiBildgueten) }

// KIBildauftrag ist der Wunsch nach einem Bild.
type KIBildauftrag struct {
	// Aufgabe ist, was ein Mensch getippt hat.
	Aufgabe string

	// Feld ist das Bildfeld der Karte, für das es gedacht ist.
	Feld string

	// Groesse und Guete sind leer erlaubt — dann entscheidet der Dienst.
	Groesse string
	Guete   string

	// Bearbeiten heißt: das Bild, das im Feld steht, wird verändert statt
	// eines neuen erzeugt.
	Bearbeiten bool
}

// Pruefe sagt, ob der Auftrag vollständig und angeboten ist.
func (a KIBildauftrag) Pruefe() error {
	if strings.TrimSpace(a.Aufgabe) == "" {
		return fmt.Errorf(
			"%w: ein Bildauftrag braucht eine Beschreibung", ErrValidation)
	}

	if len(a.Aufgabe) > KIBildaufgabeHoechstens {
		return fmt.Errorf(
			"%w: die Beschreibung hat %d Zeichen, erlaubt sind %d",
			ErrValidation, len(a.Aufgabe), KIBildaufgabeHoechstens)
	}

	if strings.TrimSpace(a.Feld) == "" {
		return fmt.Errorf(
			"%w: ein Bildauftrag braucht ein Feld", ErrValidation)
	}

	if a.Groesse != "" && !slices.Contains(kiBildgroessen, a.Groesse) {
		return fmt.Errorf("%w: die Größe %q wird nicht angeboten — es gibt %s",
			ErrValidation, a.Groesse, strings.Join(kiBildgroessen, ", "))
	}

	if a.Guete != "" && !slices.Contains(kiBildgueten, a.Guete) {
		return fmt.Errorf("%w: die Güte %q wird nicht angeboten — es gibt %s",
			ErrValidation, a.Guete, strings.Join(kiBildgueten, ", "))
	}

	return nil
}

// KIBildparameter ist eine Angabe, die über die Entstehung feststeht.
//
// **Als Paar und nicht als Verbund mit festen Feldern.** Was ein Anbieter
// nennt, ist bei jedem ein anderes — Seed hier, Schritte dort, nichts beim
// dritten —, und ein Verbund mit leeren Feldern sähe aus, als sei die Angabe
// „leer" statt „nicht genannt".
type KIBildparameter struct {
	Name string
	Wert string
}

// KIBildherkunft ist, was über ein erzeugtes Bild feststeht.
//
// # Die Ergänzung des Issues, Satz für Satz
//
// „Basisartefakt, optionale Adapter/LoRAs mit jeweiliger Revision/Lizenz/Hash,
// Konvertierung/Quantisierung, Runtimefassung" — das ist [KIHerkunft] aus A6,
// und sie steht hier als `Kette`. Für ein Cloudmodell ist sie leer, und leer
// heißt **unbekannt**, nicht „ohne Adapter".
//
// „sowie tatsächlich verfügbare Seed-/Generierungs-/Editparameter. Fehlende
// Angaben als unbekannt führen" — das sind die `Parameter`. Was nicht
// dasteht, wurde nicht genannt; erfunden wird nichts.
//
// „Seed allein garantiert keine pixelgleiche Reproduktion" — deshalb heißt
// dieses Feld `Parameter` und nicht `Rezept`.
type KIBildherkunft struct {
	Anbieter string
	Modell   string

	// Aufgabe ist, was ein Mensch getippt hat.
	Aufgabe string

	// Umschrift ist, was der Dienst daraus gemacht hat — leer, wenn er
	// nichts sagt.
	//
	// **Sie ist der Unterschied zwischen Wunsch und Auftrag.** Gezeichnet
	// wurde, was hier steht.
	Umschrift string

	Auftrag   string
	Zeitpunkt time.Time

	// Kette ist die Artefaktherkunft des Modells (A6) — leer heißt
	// unbekannt.
	Kette KIHerkunft

	Parameter []KIBildparameter
}

// Parameterwert gibt eine genannte Angabe — und sagt, ob sie genannt wurde.
func (h KIBildherkunft) Parameterwert(name string) (string, bool) {
	for _, parameter := range h.Parameter {
		if parameter.Name == name {
			return parameter.Wert, true
		}
	}

	return "", false
}

// AlsFeldherkunft gibt den Teil, den die Tafel aus A11 hält.
//
// **Ohne die Bildparameter.** Die Tafel `ki_feldherkunft` beantwortet für
// **jedes** Feld dieselbe Frage; Größe, Güte und Umschrift gelten nur für
// Bilder und stehen daneben.
func (h KIBildherkunft) AlsFeldherkunft(feld, bezug string) KIFeldherkunft {
	return KIFeldherkunft{
		Feld: feld, Anbieter: h.Anbieter, Modell: h.Modell,
		Aufgabe: h.Aufgabe, Auftrag: h.Auftrag, Zeitpunkt: h.Zeitpunkt,
		Bezug: bezug,
	}
}
