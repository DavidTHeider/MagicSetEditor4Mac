// **Die Leseseite des Archivs** — auspacken, prüfen, an den Ort legen.
//
// Aus `archiv.go` herausgelöst, als es über 400 Zeilen ging (§2). Der Schnitt
// folgt der Frage: dort steht, **was** eine Sicherung ist und wie sie
// entsteht, hier, wie man sie wieder auseinandernimmt — und was dabei nicht
// passieren darf.

package sicherung

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Entpacken legt alle Teile nach `verzeichnis` und gibt die Beschreibung.
//
// **Geprüft wird vor dem Schreiben eines jeden Teils**, nicht danach: eine
// Datei, die schon dasteht, ist schwerer wegzunehmen als eine, die nie
// entstand.
func Entpacken(pfad, verzeichnis string) (Beschreibung, error) {
	return EntpackenNur(pfad, verzeichnis)
}

// EntpackenNur legt die **genannten** Teile nach `verzeichnis`.
//
// Ohne Namen legt es alle — das ist [Entpacken]. Mit Namen ist es der Weg für
// einen Aufrufer, der nur einen Teil braucht: die Vorschau will wissen, was im
// Bestand steht, und ein Archiv mit einem Gigabyte Bildern dafür auszupacken
// wäre Arbeit vor einer Entscheidung, die noch niemand getroffen hat.
//
// **Das Manifest wird trotzdem ganz gelesen**, und die zurückgegebene
// Beschreibung führt **alle** Teile. Sie beschreibt das Archiv, nicht den
// Auszug — wer sie als Inhaltsverzeichnis des Verzeichnisses läse, läge falsch.
//
// **Ein genannter Teil, den es nicht gibt, ist ein Fehler** — wie bei
// [Entpacken]. Ein *nicht* genannter fehlender bleibt unbemerkt: geprüft wird,
// was geholt wird.
func EntpackenNur(pfad, verzeichnis string, nur ...string) (Beschreibung, error) {
	return EntpackenNurAls(pfad, verzeichnis, Manifestname, nur...)
}

// EntpackenNurAls tut dasselbe für ein Archiv mit eigenem Manifestnamen.
//
// **Die Gegenseite zu [PackenAls]** — und sie hat bis zum 2026-09-15 gefehlt.
// Wer mit einem eigenen Namen packte, konnte nicht mit demselben auspacken:
// der Leseweg suchte immer `sicherung.json`. Das fiel erst auf, als
// `internal/austausch` (N4,
// #1457)
// den ersten Umschlag zurücklas — eine halbe Zusage, und die andere Hälfte
// war nie geprüft worden (§10).
func EntpackenNurAls(
	pfad, verzeichnis, manifestname string, nur ...string,
) (Beschreibung, error) {
	archiv, err := zip.OpenReader(pfad)
	if err != nil {
		return Beschreibung{}, domain.MitGrund(domain.ErrValidation, err,
			"die Sicherung lässt sich nicht öffnen — wähle eine andere Datei")
	}

	defer func() { _ = archiv.Close() }() //nolint:errcheck // nur gelesen

	return entpackenAus(&archiv.Reader, verzeichnis, manifestname, nur...)
}

// entpackenAus legt die Teile eines geöffneten Archivs heraus — gleich, ob
// es offen auf der Platte liegt oder erst entschlüsselt wird
// ([EntpackenVerschluesselt]).
func entpackenAus(
	archiv *zip.Reader, verzeichnis, manifestname string, nur ...string,
) (Beschreibung, error) {
	// **`nur` ist ein Deckel ohne eigene Grenze**: jeder genannte Teil darf,
	// was jeder Teil darf ([HoechstensTeil], gehalten in [gesamtPruefen]).
	var deckel map[string]int64
	if len(nur) > 0 {
		deckel = map[string]int64{}
		for _, name := range nur {
			deckel[name] = HoechstensTeil
		}
	}

	return entpackenMit(archiv, verzeichnis, manifestname, HoechstensManifest, deckel)
}

// ErrUeberDemDeckel sagt, dass ein Teil über der Grenze **seines Formats**
// liegt — angekündigt oder tatsächlich (SEC-03,
// #2138).
var ErrUeberDemDeckel = fmt.Errorf("%w: über der Grenze seines Formats", domain.ErrValidation)

// EntpackenGedeckelt legt nur die Teile aus `deckel` heraus und hält jeden an
// **seiner eigenen** Grenze — an der Ankündigung im Manifest, **bevor**
// geschrieben wird; das Manifest selbst an `manifestDeckel`, bevor es ganz
// gelesen ist.
//
// **Für Teile, deren Format eine engere Grenze kennt als eine Sicherung**:
// das Manifest und die Unterschrift eines Umschlags wurden unter den 16 GiB
// je Teil herausgelegt und danach ganz eingelesen — bevor die Unterschrift
// geprüft war.
func EntpackenGedeckelt(
	pfad, verzeichnis, manifestname string, manifestDeckel int64, deckel map[string]int64,
) (Beschreibung, error) {
	archiv, err := zip.OpenReader(pfad)
	if err != nil {
		return Beschreibung{}, domain.MitGrund(domain.ErrValidation, err,
			"die Sicherung lässt sich nicht öffnen — wähle eine andere Datei")
	}

	defer func() { _ = archiv.Close() }() //nolint:errcheck // nur gelesen

	return entpackenMit(&archiv.Reader, verzeichnis, manifestname, manifestDeckel, deckel)
}

// entpackenMit ist der eine Weg hinter beiden: `deckel` nil heißt alle
// Teile, sonst nur die genannten, jeder an seiner Grenze.
func entpackenMit(
	archiv *zip.Reader, verzeichnis, manifestname string, manifestDeckel int64, deckel map[string]int64,
) (Beschreibung, error) {
	beschreibung, gewollt, nachName, fehler := gewollteTeile(archiv, manifestname, manifestDeckel, deckel)
	if fehler != nil {
		return Beschreibung{}, fehler
	}

	for _, angabe := range gewollt {
		if fehler := teilEntpacken(nachName[angabe.Name], angabe, verzeichnis); fehler != nil {
			return Beschreibung{}, fehler
		}
	}

	return beschreibung, nil
}

// gewollteTeile liest das Manifest und prüft alles, was es schon sagt —
// **bevor** ein Teil gelesen wird. Auspacken und Zurücklesen (OPS01, #2347)
// teilen es, damit eine Sicherung beim Schreiben an denselben Regeln
// gemessen wird wie beim Zurückspielen.
func gewollteTeile(
	archiv *zip.Reader, manifestname string, manifestDeckel int64, deckel map[string]int64,
) (Beschreibung, []Teilangabe, map[string]*zip.File, error) {
	// **Gezählt wird das ZIP-Verzeichnis, bevor irgendetwas gelesen wird.**
	//
	// `zip.OpenReader` hat es zu diesem Zeitpunkt schon vollständig im
	// Speicher; die Zahl hier hält nur den nächsten Schritt auf. Was **davor**
	// liegt, ist keine Grenze dieses Pakets, sondern eine von `archive/zip` —
	// und das steht bei [HoechstensEintraege] dabei.
	if len(archiv.File) > HoechstensEintraege {
		return Beschreibung{}, nil, nil, fmt.Errorf(
			"%w: dieses Archiv führt %d Einträge — über %d, und so viele hat "+
				"keine Sicherung dieser Anwendung",
			domain.ErrValidation, len(archiv.File), HoechstensEintraege)
	}

	beschreibung, err := manifestLesen(archiv, manifestname, manifestDeckel)
	if err != nil {
		return Beschreibung{}, nil, nil, err
	}

	if fehler := gesamtPruefen(beschreibung); fehler != nil {
		return Beschreibung{}, nil, nil, fehler
	}

	nachName := map[string]*zip.File{}
	for _, eintrag := range archiv.File {
		nachName[eintrag.Name] = eintrag
	}

	// **Erst alles prüfen, was das Manifest schon sagt, dann schreiben**
	// (Zweitblick über E, Befund 3): je Teil hintereinander stand ein
	// richtiges Teil schon im Verzeichnis, wenn das nächste über seiner
	// Grenze angekündigt war.
	gewollt := make([]Teilangabe, 0, len(beschreibung.Teile))

	for _, angabe := range beschreibung.Teile {
		if deckel != nil {
			grenze, istGewollt := deckel[angabe.Name]
			if !istGewollt {
				continue
			}

			if angabe.Groesse > grenze {
				return Beschreibung{}, nil, nil, fmt.Errorf(
					"%w: %s ist mit %d Byte angekündigt — sein Format trägt höchstens %d",
					ErrUeberDemDeckel, angabe.Name, angabe.Groesse, grenze)
			}
		}

		if _, da := nachName[angabe.Name]; !da {
			return Beschreibung{}, nil, nil, fmt.Errorf(
				"%w: der Sicherung fehlt %s — sie ist unvollständig",
				domain.ErrConflict, angabe.Name)
		}

		gewollt = append(gewollt, angabe)
	}

	return beschreibung, gewollt, nachName, nil
}

// gesamtPruefen weist ein Archiv ab, dessen Teile zusammen zu viel wiegen.
//
// **Neben der Grenze je Teil und nicht statt ihr** (Befund 3 des Nachlaufs
// über #1162): hundert Teile zu je 15 GiB sind einzeln zulässig und zusammen
// anderthalb Terabyte. Geströmt wird ohnehin — was hier fehlte, ist nicht der
// Arbeitsspeicher, sondern der Platz auf dem Laufwerk, auf dem der laufende
// Bestand liegt.
func gesamtPruefen(beschreibung Beschreibung) error {
	var gesamt int64

	for _, angabe := range beschreibung.Teile {
		if angabe.Groesse < 0 || angabe.Groesse > HoechstensTeil {
			return fmt.Errorf(
				"%w: %s ist mit %d Byte zu groß angekündigt — über %d, und so "+
					"viel hat keine Sicherung dieser Anwendung",
				domain.ErrValidation, angabe.Name, angabe.Groesse, HoechstensTeil)
		}

		gesamt += angabe.Groesse

		if gesamt > HoechstensGesamt {
			return fmt.Errorf(
				"%w: die Teile dieser Sicherung wiegen zusammen über %d Byte "+
					"— so viel hat keine Sicherung dieser Anwendung",
				domain.ErrValidation, HoechstensGesamt)
		}
	}

	return nil
}

// manifestLesen holt das Manifest aus dem Archiv.
func manifestLesen(archiv *zip.Reader, manifestname string, hoechstens int64) (Beschreibung, error) {
	offen, err := archiv.Open(manifestname)
	if err != nil {
		return Beschreibung{}, fmt.Errorf(
			"%w: diese Datei trägt kein %s — sie ist keine Sicherung dieser "+
				"Anwendung", domain.ErrValidation, manifestname)
	}

	defer func() { _ = offen.Close() }() //nolint:errcheck // nur gelesen

	// **Mit Deckel gelesen.** Ein Manifest ist die Auskunft darüber, wie groß
	// etwas sein darf — und damit der eine Eintrag, dessen Größe niemand
	// vorher kennt. Ein Byte mehr als [HoechstensManifest] wird geholt, damit
	// „genau am Deckel" von „darüber" zu unterscheiden ist.
	roh, err := io.ReadAll(io.LimitReader(offen, hoechstens+1))
	if err != nil {
		return Beschreibung{}, domain.MitGrund(domain.ErrValidation, err,
			"das Manifest ist unlesbar — die Sicherung ist beschädigt")
	}

	if int64(len(roh)) > hoechstens {
		return Beschreibung{}, fmt.Errorf(
			"%w: das Manifest ist zu groß — über %d Byte, und so viel hat "+
				"keine Sicherung dieser Anwendung",
			ErrUeberDemDeckel, hoechstens)
	}

	var beschreibung Beschreibung
	if fehler := json.Unmarshal(roh, &beschreibung); fehler != nil {
		return Beschreibung{}, domain.MitGrund(domain.ErrValidation, fehler,
			"das Manifest ist unlesbar — die Sicherung ist beschädigt")
	}

	if len(beschreibung.Teile) > HoechstensEintraege {
		return Beschreibung{}, fmt.Errorf(
			"%w: das Manifest führt %d Teile — über %d, und so viele hat "+
				"keine Sicherung dieser Anwendung",
			domain.ErrValidation, len(beschreibung.Teile), HoechstensEintraege)
	}

	return beschreibung, nil
}

// teilEntpacken schreibt einen Teil heraus und prüft dabei seine Summe.
func teilEntpacken(eintrag *zip.File, angabe Teilangabe, verzeichnis string) error {
	// **Die Größe ist hier schon geprüft** — [gesamtPruefen] geht das ganze
	// Manifest durch, bevor ein Teil angefasst wird. Eine zweite Prüfung an
	// dieser Stelle wäre ein Zweig, den kein Aufruf betreten kann (§3).
	ziel, err := sichererWeg(verzeichnis, angabe.Name)
	if err != nil {
		return err
	}

	if fehler := os.MkdirAll(filepath.Dir(ziel), 0o750); fehler != nil {
		return fmt.Errorf("%s: das Verzeichnis anlegen: %w", angabe.Name, fehler)
	}

	// **In eine Nachbardatei geströmt, geprüft, dann umbenannt.**
	//
	// Bis zum Zweitblick über
	// #1162
	// ging der ganze Teil erst in den Arbeitsspeicher — das hielt zwar die
	// Zusage „nichts landet ungeprüft", legte aber den Bestand vollständig
	// hinein. Bei einer großen Sammlung sind das hunderte Megabyte, und bei
	// einer präparierten Sicherung so viel, wie ihr Manifest ankündigt.
	//
	// Der Strom hält dieselbe Zusage: unter dem endgültigen Namen steht nichts,
	// bevor die Summe stimmt.
	unterwegs := ziel + ".unfertig"

	geschrieben, summe, err := stroemen(eintrag, unterwegs, angabe.Groesse)
	if err != nil {
		return errors.Join(fmt.Errorf("%s aus der Sicherung lesen: %w",
			angabe.Name, err), os.Remove(unterwegs))
	}

	if geschrieben != angabe.Groesse || summe != angabe.Summe {
		return errors.Join(fmt.Errorf(
			"%w: %s stimmt nicht mit dem Manifest überein — die Sicherung ist "+
				"beschädigt oder verändert", domain.ErrConflict, angabe.Name),
			os.Remove(unterwegs))
	}

	if fehler := os.Rename(unterwegs, ziel); fehler != nil {
		return errors.Join(fmt.Errorf("%s an seinen Ort legen: %w",
			angabe.Name, fehler), os.Remove(unterwegs))
	}

	return nil
}

// stroemen schreibt einen Eintrag nach `unterwegs` und gibt Länge und Summe.
//
// **Höchstens ein Byte mehr als angekündigt** wird gelesen: so unterscheidet
// der Aufrufer „genau so lang" von „länger", ohne den Rest zu holen.
func stroemen(eintrag *zip.File, unterwegs string, deckel int64) (geschrieben int64, summe string, fehler error) {
	offen, err := eintrag.Open()
	if err != nil {
		return 0, "", err
	}

	defer func() { _ = offen.Close() }() //nolint:errcheck // nur gelesen

	// #nosec G304 -- der Weg kommt aus `sichererWeg`, nicht aus dem Archiv
	datei, err := os.OpenFile(unterwegs, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return 0, "", err
	}

	// **Schreiben und Schließen auf einem Weg**, wie in [Packen] und aus
	// demselben Grund: eine eben angelegte Datei scheitert beim Schließen
	// nicht auf eine Art, die ein Test herbeiführen könnte.
	rechner := sha256.New()

	geschrieben, err = io.Copy(io.MultiWriter(datei, rechner),
		io.LimitReader(offen, deckel+1))
	if schliessfehler := datei.Close(); err == nil {
		err = schliessfehler
	}

	if err != nil {
		return 0, "", err
	}

	return geschrieben, hex.EncodeToString(rechner.Sum(nil)), nil
}

// sichererWeg gibt den Zielpfad — oder einen Fehler, wenn er hinausführt.
//
// **Dieselbe Wache wie beim Auspacken eines Pakets** (`internal/bezug`): ein
// Eintrag `../../…` schriebe außerhalb des Ziels. Eine Sicherung kommt von
// derselben Person, die sie zurückspielt — aber eine Datei aus dem Netz sieht
// genauso aus.
func sichererWeg(verzeichnis, name string) (string, error) {
	if path.IsAbs(name) || strings.HasPrefix(name, "..") {
		return "", fmt.Errorf("%w: der Eintrag %q führt aus der Sicherung hinaus",
			domain.ErrValidation, name)
	}

	sauber := path.Clean(name)
	if sauber == ".." || strings.HasPrefix(sauber, "../") {
		return "", fmt.Errorf("%w: der Eintrag %q führt aus der Sicherung hinaus",
			domain.ErrValidation, name)
	}

	return filepath.Join(verzeichnis, filepath.FromSlash(sauber)), nil
}
