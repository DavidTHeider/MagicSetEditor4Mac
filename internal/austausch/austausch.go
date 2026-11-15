// Package austausch schreibt und liest den portablen Umschlag — Werkstück N4
// von Phase 5.4 (#1457).
//
// **Ein Blatt** (§2) mit **einer** benannten Ausnahme: es kennt
// `internal/sicherung`. Der Grund ist derselbe wie bei `fernleitung` zu
// `leitung` und bei `svg` zu `render` — es ist der **zweite Verbraucher**
// derselben Sache: ein ZIP mit Manifest, Prüfsummen, Größengrenzen und
// Pfadschutz. Das noch einmal zu schreiben hieße, es noch einmal zu härten.
//
// # Zwei Manifeste, und das ist keine Doppelung
//
// **Sie beantworten zwei Fragen.** Das äußere (`sicherung.Beschreibung`) sagt,
// **ob der Behälter heil ist**: welche Einträge es gibt, wie groß sie sind,
// welche Prüfsumme jeder hat. Das innere ([Manifest]) sagt, **was darin
// fachlich steht**: welche Objekte, welcher Art, was sie brauchen, von wem sie
// kommen — und es ist unterschrieben.
//
// Ein einziges Manifest müsste beides tragen, und dann hinge die Härtung des
// Behälters an einer fachlichen Struktur.
//
// # Was eine Unterschrift sagt und was nicht
//
// **Sie sagt: das kommt von dieser Kennung, und niemand hat es danach
// geändert.** Sie sagt **nicht**, dass der Inhalt harmlos ist. #1457 schreibt
// es wörtlich: „Signiert bedeutet bekannte Herkunft, nicht sichere
// ausführbare Vorlage." Deshalb steht die Quarantäne **vor** dem Auspacken und
// nicht danach, und deshalb führt ein empfangenes MSE-Skript nichts aus,
// bevor jemand es freigibt.
package austausch

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// jetzt ist die Naht zur Uhr.
//
// **Als Veränderliche, damit ein Lauf den Fall herstellen kann, in dem das
// Manifest nicht zu schreiben ist** (§3): `time.Time.MarshalJSON` scheitert
// bei einem Jahr außerhalb von [0, 9999], und ohne diese Naht wäre der
// Fehlerzweig darunter eine Behauptung darüber, was im Ernstfall geschieht.
var jetzt = time.Now

// Endung ist die Dateiendung eines Umschlags.
const Endung = ".msepackage"

// Manifestname ist der Eintrag im Umschlag, der das Manifest trägt.
const Manifestname = "austausch.json"

// Unterschriftsname ist der Eintrag, der die Unterschrift trägt.
//
// **Getrennt vom Manifest und nicht darin**: eine Unterschrift über ein
// Dokument, das sie selbst enthält, ist ein Ringschluss — man müsste sie beim
// Prüfen erst wieder herausschneiden, und *wie* das geschieht, wäre eine
// zweite Verabredung, die niemand liest.
const Unterschriftsname = "austausch.sig"

// Fassung ist die Fassung dieses Umschlags.
//
// **Eine Abweichung wird abgewiesen und nicht heruntergehandelt** — dieselbe
// Linie wie beim Handschlag (N1) und bei der Einladung (N3). Wer zwei kann,
// bietet zwei an.
const Fassung = 1

// Objektart sagt, was ein Objekt ist.
type Objektart string

// Die Arten, die ein Umschlag tragen kann.
//
// **Eine Aufzählung und kein freier Text**: was hier steht, muss die
// empfangende Seite einordnen können, und ein unbekanntes Wort ist ein Befund
// und keine Einladung zum Raten (§16).
const (
	ArtKarte    Objektart = "karte"
	ArtSet      Objektart = "set"
	ArtVorlage  Objektart = "vorlage"
	ArtDeck     Objektart = "deck"
	ArtSammlung Objektart = "sammlungssicht"

	// ArtLeihe ist eine **digitale Deckleihe** samt ihrem eingefrorenen
	// Abzug (G6,
	// #1463).
	//
	// # Warum sie hier steht und nicht neben `ArtDeck`
	//
	// **Weil sie kein Deck ist.** Ein Deck geht in die Sammlung des
	// Empfängers; eine Leihe **ausdrücklich nicht** — das Issue sagt es:
	// „die Leihe importiert nicht automatisch eine dauerhafte Kopie in die
	// Sammlung." Wer sie als `deck` schickte, bekäme drüben genau das.
	//
	// # Zwei Unterschriften, und sie beantworten zwei Fragen
	//
	// **Die des Umschlags sagt, wer geschickt hat**; die der Leihe sagt, wer
	// **verliehen** hat. Sie fallen meist zusammen und müssen es nicht: ein
	// Dokument darf weitergereicht werden, und wer es weitergibt, erwirbt
	// dadurch kein Recht daran.
	ArtLeihe Objektart = "leihe"
)

// Objekt ist ein Stück Inhalt im Umschlag.
type Objekt struct {
	// Art sagt, was es ist.
	Art Objektart `json:"art"`

	// Name ist der Weg im Umschlag, mit Schrägstrichen.
	Name string `json:"name"`

	// Bezeichnung ist, was ein Mensch liest — **kein Nachweis**.
	Bezeichnung string `json:"bezeichnung"`

	// Summe ist die SHA-256 des Inhalts, hexadezimal.
	Summe string `json:"sha256"`

	// Groesse ist die Zahl der Bytes.
	Groesse int64 `json:"groesse"`

	// Braucht nennt die Namen, ohne die dieses Objekt unvollständig ist.
	//
	// **Namen und keine Kennungen**: was gebraucht wird, liegt im selben
	// Umschlag oder gar nicht. Ein Verweis nach draußen wäre eine Abfrage
	// über den Bestand des Empfängers, und #1457 verbietet sie ausdrücklich.
	Braucht []string `json:"braucht"`
}

// Manifest sagt, was in einem Umschlag steht.
type Manifest struct {
	Fassung int `json:"fassung"`

	// Urheber ist die Peerkennung dessen, der unterschrieben hat.
	Urheber string `json:"urheber"`

	// Erstellt ist der Zeitpunkt — **kein Nachweis**, sondern eine Auskunft.
	Erstellt time.Time `json:"erstellt"`

	Objekte []Objekt `json:"objekte"`
}

// Schreibe legt einen Umschlag an und unterschreibt sein Manifest.
//
// **Die Objekte kommen als fertige Bytes herein.** Was eine Karte zu einem
// `.mse-set` macht, weiß `msepaket`; dieses Blatt packt und unterschreibt. Wer
// beides vermischte, könnte den Umschlag nicht prüfen, ohne ein Spiel zu
// kennen.
func Schreibe(
	ziel string, inhalte []Inhalt, ich peer.Identitaet,
) (Manifest, error) {
	if len(inhalte) == 0 {
		return Manifest{}, fmt.Errorf(
			"%w: ein Umschlag ohne Objekte ist keiner", domain.ErrValidation)
	}

	manifest := Manifest{
		Fassung:  Fassung,
		Urheber:  string(ich.Kennung()),
		Erstellt: jetzt().UTC(),
		Objekte:  make([]Objekt, 0, len(inhalte)),
	}

	teile := make([]sicherung.Teil, 0, len(inhalte)+2)

	for _, inhalt := range inhalte {
		if fehler := inhalt.pruefe(); fehler != nil {
			return Manifest{}, fehler
		}

		summe := sha256.Sum256(inhalt.Bytes)

		manifest.Objekte = append(manifest.Objekte, Objekt{
			Art:         inhalt.Art,
			Name:        inhalt.Name,
			Bezeichnung: inhalt.Bezeichnung,
			Summe:       hex.EncodeToString(summe[:]),
			Groesse:     int64(len(inhalt.Bytes)),
			Braucht:     inhalt.Braucht,
		})

		teile = append(teile, sicherung.Teil{
			Name: inhalt.Name, Inhalt: inhalt.Bytes,
		})
	}

	roh, fehler := json.Marshal(manifest)
	if fehler != nil {
		return Manifest{}, domain.MitGrund(domain.ErrValidation, fehler,
			"das Manifest schreiben")
	}

	teile = append(teile,
		sicherung.Teil{Name: Manifestname, Inhalt: roh},
		sicherung.Teil{Name: Unterschriftsname, Inhalt: ich.Signiere(roh)})

	if fehler := sicherung.PackenAls(
		ziel, aeusseresManifest, teile, sicherung.Angaben{},
	); fehler != nil {
		return Manifest{}, fehler
	}

	return manifest, nil
}

// aeusseresManifest ist der Name des Behältermanifests.
//
// **Ein eigener Name und nicht der von `sicherung`**: ein Umschlag ist keine
// Sicherung, und `sicherung.IstArchiv` soll ihn nicht dafür halten.
const aeusseresManifest = "umschlag.json"

// Inhalt ist ein Objekt, wie es hereingereicht wird.
type Inhalt struct {
	Art         Objektart
	Name        string
	Bezeichnung string
	Bytes       []byte
	Braucht     []string
}

// pruefe hält einen Inhalt gegen das, was einer sein kann.
func (i Inhalt) pruefe() error {
	if !i.Art.gilt() {
		return fmt.Errorf("%w: %q ist keine Objektart",
			domain.ErrValidation, i.Art)
	}

	return pruefeName(i.Name)
}

// gilt sagt, ob eine Art bekannt ist.
func (a Objektart) gilt() bool {
	switch a {
	case ArtKarte, ArtSet, ArtVorlage, ArtDeck, ArtSammlung, ArtLeihe:
		return true
	default:
		return false
	}
}
