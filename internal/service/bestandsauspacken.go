// **Eine Sicherung mit Bildern zurückspielen** —
// #1162.
//
// # Zwei Formate, ein Weg
//
// Seit 0.3.0 ist eine Sicherung ein **Archiv** aus Bestand und Beiwerk. Was
// 0.1.0 und 0.2.0 geschrieben haben, ist eine **nackte Datenbank** — und die
// muss weiterhin zurückzuspielen sein. Eine Behebung, die jede vorhandene
// Sicherung wertlos macht, wäre schlimmer als der Fehler.
//
// Unterschieden wird an den ersten vier Bytes (`sicherung.IstArchiv`), nicht
// an der Endung: wer eine Datei umbenennt, ändert nicht, was drin steht.

package service

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// Beiwerkbeiseite ist der Name, unter dem das alte Beiwerk landet.
//
// **Neben dem Beiwerk und nicht darin**: ein Ordner, der sich selbst enthält,
// wäre beim nächsten Sichern ein Ordner voller alter Ordner.
const Beiwerkbeiseite = "beiwerk-vorher"

// ausgepackteSicherung ist eine Sicherung, die auf der Platte liegt.
type ausgepackteSicherung struct {
	// Bestand ist die Datenbank, die zurückgespielt werden soll.
	Bestand string

	// Beiwerk ist das Verzeichnis der Bilder — leer, wenn keines dabei war.
	Beiwerk string

	// Wegraeumen gibt den Zwischenplatz wieder frei.
	Wegraeumen func()

	// Archiv sagt, ob die gewählte Datei eine `.msesicherung` war.
	//
	// **Getrennt von allem anderen**, weil daran hängt, ob die Sicherung über
	// Bilder überhaupt **etwas sagt**. Eine nackte `.db` tut es nicht.
	Archiv bool

	// Verschluesselt sagt, ob sie ein Kennwort trug (OPS-06).
	Verschluesselt bool

	// Pakete sind die Verweise aus dem Manifest, Paketplatz der Ordner der
	// eigenen Pakete im Zwischenplatz — leer, wenn sie keine trug (OPS-06).
	Pakete     []sicherung.Paketverweis
	Paketplatz string

	// TraegtBilder sagt, ob das **Archiv** einen Bilderstand beschreibt.
	//
	// **Aus dem Manifest und nicht von der Platte**: die Vorschau packt die
	// Bilder gar nicht aus, und ein `os.Stat` sagte dort „keine", wo das
	// Archiv sehr wohl welche führt.
	TraegtBilder bool

	// BilderErsetzen sagt, ob das Bilderverzeichnis ausgetauscht wird.
	//
	// **Eine Aussage über diesen Auszug, nicht über das Archiv** — und
	// deshalb heißt sie so und nicht `MitBildern`. Sie ist wahr, wenn hier
	// **vollständig** ausgepackt wurde und die Sicherung damit einen
	// Bilderstand beschreibt; dann gilt [Beiwerk], leer oder nicht.
	//
	// **Getrennt von `Beiwerk != ""`**, weil eine neue Sicherung **keine**
	// Bilder tragen kann und das etwas anderes heißt als eine alte, die von
	// ihnen **nichts weiß**: im ersten Fall wird geleert, im zweiten bleibt
	// stehen.
	//
	// **Und falsch bei einem Auszug**, gleich was im Archiv steht: die
	// Vorschau holt allein den Bestand, und wer auf ihrem Ergebnis das
	// Bilderverzeichnis ersetzte, ersetzte es durch nichts. Ein Kommentar
	// „wird dort nicht gelesen" hielte nur die halbe Zusage (§10).
	BilderErsetzen bool
}

// Bilderlage sagt, was beim Zurückspielen mit den Bildern geschieht.
//
// **Sie liest [ausgepackteSicherung.Archiv] und nicht `BilderErsetzen`.** Das
// andere Feld ist eine Aussage über **diesen Auszug** — die Vorschau holt nur
// den Bestand, und für sie ist es immer falsch. Was die Rückfrage zeigen
// muss, ist aber die Aussage über das **Archiv**: was geschähe, wenn jemand
// bestätigt.
func (s ausgepackteSicherung) Bilderlage() string {
	switch {
	case !s.Archiv:
		return BilderUnberuehrt
	case s.TraegtBilder:
		return BilderErsetzt
	default:
		return BilderGeleert
	}
}

// auspacken macht aus einer gewählten Datei eine Sicherung auf der Platte.
//
// **`nur` sagt, welche Teile geholt werden.** Ohne Angabe alle; die Vorschau
// nennt allein [Bestandsname], weil sie nur ihn ansieht — und ein Archiv
// zweimal auszupacken (einmal für die Rückfrage, einmal für die Antwort
// darauf) wäre Wartezeit vor einer Entscheidung, die noch aussteht.
func (b *Bestandsweg) auspacken(ctx context.Context, pfad, bestandspfad string, nur ...string) (ausgepackteSicherung, error) {
	// **Eine Datei, die nicht daliegt, ist nicht gefunden** — und keine
	// nackte Datenbank, an deren Kopie das System scheitert. Bis C01 (#2226)
	// kam der rohe Fehler als „unerwarteter Fehler" im Fenster an; zwischen
	// Dialog und Zurückspielen kann eine Datei verschwinden, und die Flächen
	// ohne Fenster nehmen einen Pfad ohne Dialog.
	if _, statfehler := os.Stat(pfad); errors.Is(statfehler, fs.ErrNotExist) {
		return ausgepackteSicherung{}, fmt.Errorf("%w: die Sicherung %s liegt nicht da",
			domain.ErrNotFound, filepath.Base(pfad))
	}

	verschluesselt := sicherung.IstVerschluesselt(pfad)

	if !verschluesselt && !sicherung.IstArchiv(pfad) {
		// **Eine nackte Datenbank aus 0.1.0 oder 0.2.0.** Sie wird kopiert
		// wie eh und je; Bilder trägt sie keine, und das vorhandene Beiwerk
		// bleibt deshalb stehen.
		eingehend, fehler := eingehendeKopie(pfad, bestandspfad)
		if fehler != nil {
			return ausgepackteSicherung{}, fehler
		}

		return ausgepackteSicherung{
			Bestand:    eingehend,
			Wegraeumen: func() { eingehendeWegraeumen(eingehend) },
		}, nil
	}

	platz, err := os.MkdirTemp(filepath.Dir(bestandspfad), Rueckspielvorsatz)
	if err != nil {
		return ausgepackteSicherung{}, fmt.Errorf("den Zwischenplatz anlegen: %w", err)
	}

	wegraeumen := func() { _ = os.RemoveAll(platz) } //nolint:errcheck // kein Datenverlust

	// **Eine verschlüsselte Sicherung wird im Strom geöffnet** (OPS-06): das
	// Archiv darin liegt nie entschlüsselt auf der Platte, nur die Teile, die
	// ohnehin ausgepackt werden.
	entpacken := sicherung.EntpackenNur
	if verschluesselt {
		kennwort := sicherungswunschAus(ctx).Kennwort
		entpacken = func(pfad, platz string, nur ...string) (sicherung.Beschreibung, error) {
			return sicherung.EntpackenVerschluesselt(pfad, platz, kennwort, nur...)
		}
	}

	beschreibung, entpackfehler := entpacken(pfad, platz, nur...)
	if entpackfehler != nil {
		wegraeumen()

		return ausgepackteSicherung{}, entpackfehler
	}

	// **Die Fassung im Manifest wird gelesen, nicht nur geschrieben.**
	//
	// Bis zum Zweitblick über #1162 verwarf dieser Aufrufer die Beschreibung
	// mit `_` — und der Kommentar an [sicherung.Angaben.Fassung] sagte
	// derweil, sie stehe da, „damit ein Rückspielweg entscheiden kann".
	// Geprüft wurden allein die Migrationen im Bestand; eine Sicherung einer
	// künftigen Fassung, deren **Schema** gleich blieb und deren
	// **Archivgestalt** sich änderte, kam durch.
	if fassungsfehler := b.fassungPruefen(beschreibung.Fassung); fassungsfehler != nil {
		wegraeumen()

		return ausgepackteSicherung{}, fassungsfehler
	}

	bestand := filepath.Join(platz, Bestandsname)
	if _, statfehler := os.Stat(bestand); statfehler != nil {
		wegraeumen()

		return ausgepackteSicherung{}, fmt.Errorf(
			"%w: der Sicherung fehlt %s — sie trägt keinen Bestand",
			domain.ErrValidation, Bestandsname)
	}

	beiwerk := filepath.Join(platz, Beiwerkordner)

	angaben, statfehler := os.Stat(beiwerk)

	switch {
	case statfehler != nil:
		// **Kein Beiwerk ist ein gültiger Fall**: wer nie ein Bild eingesetzt
		// hat, hat keines. Die Sicherung sagt trotzdem „hier sind alle" —
		// und das heißt: keine.
		beiwerk = ""

	case !angaben.IsDir():
		// **Und eine Datei dieses Namens ist keiner.** Geprüft wurde bis zum
		// Zweitblick über #1162 die bloße Existenz; ein Archiv mit einem
		// gewöhnlichen Eintrag `beiwerk` kam damit an den Ort des
		// Bilderverzeichnisses — das Zurückspielen meldete Erfolg, die Bilder
		// fehlten, und jedes spätere Einsetzen scheiterte daran, dass unter
		// einer Datei kein Ordner entsteht.
		wegraeumen()

		return ausgepackteSicherung{}, fmt.Errorf(
			"%w: in dieser Sicherung ist %s eine Datei und kein Verzeichnis — "+
				"sie stammt nicht von dieser Anwendung",
			domain.ErrValidation, Beiwerkordner)
	}

	return ausgepackteSicherung{
		Bestand:        bestand,
		Beiwerk:        beiwerk,
		Wegraeumen:     wegraeumen,
		Archiv:         true,
		Verschluesselt: verschluesselt,
		Pakete:         beschreibung.Pakete,
		Paketplatz:     paketplatz(platz),
		TraegtBilder:   traegtBilder(beschreibung),
		BilderErsetzen: len(nur) == 0,
	}, nil
}

// paketplatz ist der Ordner der eigenen Pakete im Zwischenplatz — leer,
// wenn keine ausgepackt wurden.
func paketplatz(platz string) string {
	ordner := filepath.Join(platz, Paketordner)
	if _, fehler := os.Stat(ordner); fehler != nil {
		return ""
	}

	return ordner
}

// traegtBilder sagt, ob ein Manifest einen Bildereintrag führt.
func traegtBilder(beschreibung sicherung.Beschreibung) bool {
	for _, teil := range beschreibung.Teile {
		if strings.HasPrefix(teil.Name, Beiwerkordner+"/") {
			return true
		}
	}

	return false
}

// beiwerkErsetzen tauscht das Bilderverzeichnis und gibt den Rückweg.
//
// **Nach dem Bestand und nicht davor.** Scheitert der Tausch der Datenbank,
// sind die Bilder ohne Bedeutung; scheitert er hier, steht die Datenbank schon
// — und dann ist der Rückweg beider Teile das, was zählt.
func (b *Bestandsweg) beiwerkErsetzen(neu string) (func() error, error) {
	ziel, err := b.orte.Bilder()
	if err != nil {
		return nil, fmt.Errorf("das Bilderverzeichnis finden: %w", err)
	}

	beiseite := filepath.Join(filepath.Dir(ziel), Beiwerkbeiseite)

	// **Ein früherer Rückspielstand wird weggeräumt.** Sonst scheiterte der
	// zweite Rückspielvorgang am Namen — und der erste hat seinen Zweck
	// erfüllt, sobald jemand weiterarbeitet.
	if fehler := os.RemoveAll(beiseite); fehler != nil {
		return nil, fmt.Errorf("den alten Rückspielstand wegräumen: %w", fehler)
	}

	gabEsEins, err := verschieben(ziel, beiseite)
	if err != nil {
		return nil, fmt.Errorf("das Bilderverzeichnis beiseitelegen: %w", err)
	}

	zurueck := func() error {
		if fehler := os.RemoveAll(ziel); fehler != nil {
			return fehler
		}

		if !gabEsEins {
			return nil
		}

		_, fehler := verschieben(beiseite, ziel)

		return fehler
	}

	if neu == "" {
		// **Eine Sicherung ohne Bilder ersetzt das Verzeichnis durch nichts.**
		// Das ist der Stand, den sie beschreibt — und ein stehengebliebener
		// Ordner wäre eine Vermischung zweier Stände.
		return zurueck, nil
	}

	if fehler := os.Rename(neu, ziel); fehler != nil {
		return zurueck, errors.Join(
			fmt.Errorf("die Bilder an ihren Ort legen: %w", fehler), rueckweg(zurueck))
	}

	return zurueck, nil
}

// fassungPruefen weist eine Sicherung ab, deren Archivfassung nicht passt.
//
// # Warum hier nicht gilt, was für die nackte `.db` gilt
//
// Dort ist „keine Angabe" eine echte Auskunft: eine Sicherung aus 0.1.0 oder
// 0.2.0 **kann** keine tragen, und aus „weiß nichts darüber" ein „stammt aus
// der Zukunft" zu machen wäre der Fehlschluss, gegen den §8 steht.
//
// **Für ein Archiv gilt das Gegenteil**, und der erste Anlauf hat es
// verwechselt: es gibt kein altes `.msesicherung`-Format ohne dieses Feld —
// dieses Format **beginnt** mit 0.3.0. Eine leere oder unlesbare Angabe kommt
// deshalb nicht von einer älteren Fassung, sondern von etwas anderem: einem
// fremden Werkzeug, einer veränderten Datei, einer künftigen Gestalt. Sie
// durchzulassen hieße, die Wache mit `fassung: "morgen"` zu umgehen, die
// `99.0.0` abweist (Befund 5 des Nachlaufs über #1162).
//
// **Älter ist in Ordnung.** Eine Sicherung aus 0.3.0 in 0.4.0 zurückzuspielen
// ist der Normalfall; dass ihr Bestand migriert werden muss, sagt die
// Schemaprüfung gleich danach.
func (b *Bestandsweg) fassungPruefen(fassung string) error {
	if fassung == "" {
		return &domain.FremdeArchivfassung{Hier: b.fassung}
	}

	neuer, lesefehler := domain.FassungIstNeuer(fassung, b.fassung)
	if lesefehler != nil || neuer {
		return &domain.FremdeArchivfassung{Dort: fassung, Hier: b.fassung}
	}

	return nil
}
