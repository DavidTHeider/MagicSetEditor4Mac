package domain

// **Das Kontextpaket** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Warum ein Paket und kein freier Promptbestand
//
// **Weil ein Modell nur belegen kann, was ihm jemand gegeben hat.** Das Issue
// nennt es als eigene Überschrift: „Kontextpaket statt freiem Promptbestand".
// Der Dienst wählt aus, entdoppelt, ordnet und kürzt — und was dabei
// herauskommt, trägt **Beleg-IDs**, die der Dienst vergeben hat.
//
// # Der Satz, um den herum das gebaut ist
//
// **„Belege werden vom Service vergeben. Modell darf nur diese IDs
// referenzieren; unbekannte IDs/erfundene Regelnummern/URLs ergeben
// Validierungsbefund, keine vertrauenswürdige Quellenanzeige."**
//
// Damit ist eine erfundene Regelnummer kein Darstellungsproblem, sondern ein
// **Befund**: sie fällt auf, bevor jemand sie liest.
//
// # Und warum das Manifest unveränderlich ist
//
// **Weil es der Beleg über den Beleg ist.** Wer es nachträglich ändern
// könnte, könnte behaupten, ein Modell habe etwas anderes gesehen. Das
// Manifest entsteht mit dem Paket und wird nie fortgeschrieben; ein anderes
// Zielmodell bekommt ein **neues** Paket ([Kontextmanifest.PasstZu]).

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Belegkennung ist die Kennung, unter der ein Modell einen Beleg nennen darf.
//
// **Kurz und sprechend** (`B1`, `B2`): ein Modell soll sie in einen Satz
// setzen können, ohne dass der Satz zur Hälfte aus Kennung besteht.
type Belegkennung string

// Beleg ist ein Ausschnitt, auf den sich eine Aussage stützen darf.
type Beleg struct {
	Kennung Belegkennung

	// Quellkennung ist die Kennung des Abschnitts im Index.
	Quellkennung string

	Quelle Retrievalquelle

	// Kennungssystem, Fachkennung und Semantik sind die Identität nach A2.
	Kennungssystem Kennungssystem
	Fachkennung    string
	Semantik       Semantik

	Titel string

	// Ausschnitt ist **der Text, der wirklich mitgegeben wurde** — nicht der
	// ganze Abschnitt.
	//
	// **Das Issue verlangt es**: „tatsächlich verwendete Ausschnitte". Wer
	// den ganzen Abschnitt festhielte, behauptete, das Modell habe ihn
	// gesehen.
	Ausschnitt string

	// Von und Bis sind der Zeichenbereich des Ausschnitts im Abschnitt.
	Von int
	Bis int

	// Inhaltshash ist der Abdruck des **Ausschnitts**.
	Inhaltshash string

	Sprache     string
	Regelnummer string
	Urheber     Rulingurheber

	Bezugsadresse string

	Quellrevision string
	Quelldatum    string
	Abgerufen     string

	// Veraltet sagt, dass die Quelle nicht mehr dem aktiven Stand entspricht.
	//
	// **Das Issue: „Alte Quellen als alt anzeigen."** Ein Regelstand von
	// gestern ist eine Auskunft und kein Fehler — solange jemand sieht, dass
	// er von gestern ist.
	Veraltet bool
}

// Kontextmanifest beschreibt ein Paket vollständig und unveränderlich.
type Kontextmanifest struct {
	// PaketID ist die Kennung des Pakets.
	PaketID string

	Aufgabe string

	// Zielrevision bindet das Paket an den Stand der Zielkarte.
	Zielrevision string

	// Zielmodell und Laufzeit sind die Bindung, für die budgetiert wurde.
	//
	// **Ein anderes Zielmodell ist ein anderes Budget** — die Zählweise und
	// die Grenzen hängen am Modell, und dasselbe Paket dort wiederzuverwenden
	// hieße, eine Zahl zu übernehmen, die für ein anderes galt.
	Zielanbieter string
	Zielmodell   string
	Laufzeit     string

	// Zaehlweise nennt, **wie** Tokens gezählt wurden.
	//
	// **Weil es kein gemeinsames Maß gibt**: was ein Modell als Token zählt,
	// zählt das nächste anders. Ohne die Angabe ist die Zahl bedeutungslos.
	Zaehlweise string

	// Kontextgrenze ist das Fenster, Ausgabereserve der Teil davon, der für
	// die Antwort freigehalten wird.
	Kontextgrenze  int
	Ausgabereserve int

	// Verbraucht ist, was die Belege wirklich kosten.
	Verbraucht int

	// Indexfingerabdruck, Modellfingerabdruck und Suchkonfiguration sagen,
	// **wogegen** gesucht wurde.
	Indexfingerabdruck string
	Raumabdruck        string
	Suchkonfiguration  string
	Rerankerabdruck    string

	Korpusstaende []Korpusstand

	// Kontextpolicy nennt die Fassung der Regel, nach der ausgewählt wurde.
	Kontextpolicy string

	// Belege sind die vergebenen Kennungen samt ihrer Herkunft.
	Belege []Beleg

	// ProfilID bindet das Paket an den, für den es gebaut wurde.
	ProfilID int64

	// Erstellt und Laeuftab sind Entstehung und Ablauf.
	Erstellt time.Time
	Laeuftab time.Time
}

// Kontextpolicyfassung ist die Fassung der Auswahlregel.
//
// **Sie steht im Manifest und wird hochgezählt, wenn sich die Auswahl
// ändert.** Zwei Pakete mit verschiedener Fassung sind nicht vergleichbar —
// und ein Vergleich ohne die Angabe verglich zwei Regeln statt zweier
// Anfragen.
const Kontextpolicyfassung = "kontext-1"

// Kontextfrist ist, wie lange ein Paket gilt.
//
// **Eine Stunde**, dieselbe wie beim Bildvorrat aus A12: lang genug, um es
// anzusehen und abzuschicken, kurz genug, dass ein Widerruf greift, bevor
// jemand es benutzt.
const Kontextfrist = time.Hour

// Pruefe sagt, ob das Manifest vollständig ist.
func (m Kontextmanifest) Pruefe() error {
	for name, wert := range map[string]string{
		"eine Paketkennung":  m.PaketID,
		"eine Aufgabe":       m.Aufgabe,
		"ein Zielmodell":     m.Zielmodell,
		"einen Anbieter":     m.Zielanbieter,
		"eine Zählweise":     m.Zaehlweise,
		"eine Kontextpolicy": m.Kontextpolicy,
	} {
		if strings.TrimSpace(wert) == "" {
			return fmt.Errorf("%w: ein Kontextmanifest braucht %s",
				ErrValidation, name)
		}
	}

	if m.ProfilID == 0 {
		return fmt.Errorf(
			"%w: ein Kontextpaket gehört einem Profil", ErrValidation)
	}

	return m.pruefeBudget()
}

// pruefeBudget hält die drei Zahlen zusammen.
func (m Kontextmanifest) pruefeBudget() error {
	if m.Kontextgrenze <= 0 || m.Ausgabereserve < 0 {
		return fmt.Errorf(
			"%w: das Kontextfenster ist %d und die Ausgabereserve %d — "+
				"beides zählt ab null, das Fenster ab eins",
			ErrValidation, m.Kontextgrenze, m.Ausgabereserve)
	}

	if m.Ausgabereserve >= m.Kontextgrenze {
		return fmt.Errorf(
			"%w: die Ausgabereserve %d füllt das Fenster von %d ganz — "+
				"dann bleibt für den Kontext nichts",
			ErrValidation, m.Ausgabereserve, m.Kontextgrenze)
	}

	if uebrig := m.Kontextgrenze - m.Ausgabereserve; m.Verbraucht > uebrig {
		return fmt.Errorf(
			"%w: die Belege brauchen %d von %d verfügbaren Stellen",
			ErrValidation, m.Verbraucht, uebrig)
	}

	return nil
}

// Verfuegbar ist, was für den Kontext übrig bleibt.
func (m Kontextmanifest) Verfuegbar() int {
	return m.Kontextgrenze - m.Ausgabereserve
}

// PasstZu sagt, ob dieses Paket für ein Zielmodell gilt.
//
// **Anbieter, Modell, Laufzeit und Zählweise müssen übereinstimmen.** Das
// Issue: „Bei Zielmodellwechsel das Paket erneut budgetieren und erforderliche
// Freigabe erneuern; gleiche Aufgabe/Zielrevision genügt nicht zur
// Wiederverwendung."
func (m Kontextmanifest) PasstZu(anderer Kontextmanifest) bool {
	return m.Zielanbieter == anderer.Zielanbieter &&
		m.Zielmodell == anderer.Zielmodell &&
		m.Laufzeit == anderer.Laufzeit &&
		m.Zaehlweise == anderer.Zaehlweise
}

// Abgelaufen sagt, ob die Frist um ist.
func (m Kontextmanifest) Abgelaufen(jetzt time.Time) bool {
	return !jetzt.Before(m.Laeuftab)
}

// Kennungen gibt die Belegkennungen, die ein Modell nennen darf.
func (m Kontextmanifest) Kennungen() []Belegkennung {
	kennungen := make([]Belegkennung, 0, len(m.Belege))
	for _, beleg := range m.Belege {
		kennungen = append(kennungen, beleg.Kennung)
	}

	return kennungen
}

// BelegkennungFuer baut die Kennung des n-ten Belegs.
func BelegkennungFuer(stelle int) Belegkennung {
	return Belegkennung("B" + strconv.Itoa(stelle+1))
}

// Belegabdruck ist der Abdruck eines Ausschnitts.
//
// **Über den Ausschnitt und nicht über den Abschnitt**: er sagt, ob **das**
// noch dasteht, was das Modell gesehen hat.
func Belegabdruck(ausschnitt string) string {
	summe := sha256.Sum256([]byte(ausschnitt))

	return "sha256:" + hex.EncodeToString(summe[:])
}
