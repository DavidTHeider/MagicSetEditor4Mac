package app

// **Welche Dateien zu welchem Posten gehören** — Werkstück 4 von Phase 4.13
// (#1352).
//
// Eigene Datei, weil `platz.go` mit ihr über 400 Zeilen ginge (§2). Der
// Schnitt folgt der Frage: dort steht, **was die Bindung antwortet**, hier,
// **wo nachgesehen wird**.

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// platzdatei ist eine gefundene Datei samt ihrem Gewicht und Alter.
type platzdatei struct {
	pfad  string
	wiegt int64
	alter time.Time
}

// dateienDesPostens sammelt, was zu einem Posten gehört.
//
// # Warum ein fehlendes Verzeichnis kein Fehler ist
//
// **Weil es der Normalfall beim ersten Start ist.** `beiwerk/` entsteht mit
// dem ersten Bild, `pakete/` mit dem ersten Bezug — und „null Byte" ist die
// richtige Antwort, nicht „geht nicht" (§8).
func dateienDesPostens(posten domain.Platzposten, bestandspfad string) (
	[]platzdatei, error,
) {
	switch posten {
	case domain.PostenBestand:
		// **Die Begleiter zählen mit.** `-wal` kann größer sein als der
		// Bestand selbst; wer ihn wegließe, nennte eine Zahl, die niemand auf
		// der Platte wiederfindet.
		return einzelne(bestandspfad, bestandspfad+"-wal", bestandspfad+"-shm"), nil

	case domain.PostenBilder:
		ort, fehler := Bilderverzeichnis()
		if fehler != nil {
			return nil, fehler
		}

		return unterhalb(ort, nil)

	case domain.PostenPakete:
		ort, fehler := Paketverzeichnis()
		if fehler != nil {
			return nil, fehler
		}

		return unterhalb(ort, nil)

	case domain.PostenVorsicherungen:
		return vorsicherungen(bestandspfad)

	}

	// **Das Protokoll als Ausgang und nicht als fünfter Fall.** Gerufen wird
	// nur mit den fünf Posten, die sich für sich allein bestimmen lassen; ein
	// `default`, der „unbekannt" behandelte, wäre ein Zweig, den kein Aufruf
	// betritt — also eine Behauptung darüber, was im Ernstfall geschieht (§3).
	//
	// **`sonstiges` steht nicht darunter**, und der Grund ist kein Geschmack:
	// es ist der **Rest**, und ein Rest lässt sich nur dort bestimmen, wo die
	// anderen schon gezählt sind — also in [App.Platz]. Stünde er hier, müsste
	// er die fünf noch einmal rufen, und deren Fehlerzweige wären dann ein
	// zweites Mal da: unerreichbar, weil der erste Ruf längst abgebrochen
	// hätte. Genau das hat die Abdeckung am 2026-09-07 angestrichen.
	return einzelne(protokolldateien(bestandspfad)...), nil
}

// laufendesProtokoll ist die Datei, in die gerade geschrieben wird.
//
// **Sie hat einen eigenen Namen, weil `PlatzFreigeben` sie stehen lässt.** Ein
// Index in eine Liste sagte dasselbe und wäre beim nächsten Umbau still
// falsch.
func laufendesProtokoll(bestandspfad string) string {
	return filepath.Join(filepath.Dir(bestandspfad), Protokollname)
}

// protokolldateien nennt das laufende Protokoll und seine Vorfassungen.
//
// **Alle, die eine Nummer tragen**, und nicht nur `.1`: seit OPS-04 bewahrt
// die Drehung bis zu zehn auf, und die übrigen fielen bis zum Zweitblick über
// Phase 6.4 in den Rest, den niemand räumen kann. Ein anderer eingestellter
// Ort liegt außerhalb des Anwendungsordners und damit außerhalb dieser
// Übersicht.
func protokolldateien(bestandspfad string) []string {
	ordner := filepath.Dir(bestandspfad)
	dateien := []string{laufendesProtokoll(bestandspfad)}

	// Ein unlesbarer Ordner hat keine Vorfassungen, die sich zählen ließen —
	// dieselbe Entscheidung wie in [einzelne].
	eintraege, _ := os.ReadDir(ordner) //nolint:errcheck // siehe oben
	for _, eintrag := range eintraege {
		nummer, ist := strings.CutPrefix(eintrag.Name(), Protokollname+".")
		if ist && nummer != "" && strings.Trim(nummer, "0123456789") == "" {
			dateien = append(dateien, filepath.Join(ordner, eintrag.Name()))
		}
	}

	return dateien
}

// einzelne sammelt benannte Dateien — was fehlt, fehlt.
//
// **Jeder Fehlschlag heißt „nicht da", und das ist eine Entscheidung.** Die
// Begleiter `-wal` und `-shm` gibt es nur, solange jemand schreibt; sie fehlen
// im Normalfall. Ein Zweig, der „fehlt" von „darf ich nicht lesen"
// unterschiede, wäre für den Bestand selbst nicht zu betreten: wer ihn nicht
// lesen kann, hat die Anwendung gar nicht erst gestartet.
func einzelne(pfade ...string) []platzdatei {
	gefunden := make([]platzdatei, 0, len(pfade))

	for _, pfad := range pfade {
		angabe, fehler := os.Stat(pfad)
		if fehler != nil {
			continue
		}

		gefunden = append(gefunden, platzdatei{
			pfad: pfad, wiegt: angabe.Size(), alter: angabe.ModTime(),
		})
	}

	return gefunden
}

// unterhalb sammelt alles unter einem Verzeichnis — ohne das, was schon
// jemandem gehört.
//
// **Ein Walk und nicht zwei.** `beansprucht` darf `nil` sein und heißt dann
// „nimm alles"; die Posten *Bilder* und *Pakete* rufen so. Der Posten
// *Sonstiges* reicht die Pfade der fünf anderen herein — und bekommt damit
// **dieselben** Fehlerzweige wie sie, statt eigener, die niemand betreten
// kann (Werkstück 6 von Phase 4.13).
func unterhalb(ort string, beansprucht map[string]bool) ([]platzdatei, error) {
	var gefunden []platzdatei

	fehler := filepath.WalkDir(ort, func(
		pfad string, eintrag fs.DirEntry, unterwegs error,
	) error {
		if unterwegs != nil {
			// **Ein fehlendes Verzeichnis ist der erste Start.**
			if os.IsNotExist(unterwegs) {
				return filepath.SkipAll
			}

			return unterwegs
		}

		if eintrag.IsDir() || beansprucht[pfad] {
			return nil
		}

		// **Eine Datei, die zwischen Auflisten und Nachsehen verschwindet,
		// belegt keinen Platz.** Der Fehler ist hier kein Mangel, sondern die
		// Antwort — und ein Zweig, der ihn meldete, machte aus einem Rennen
		// eine Fehlermeldung.
		angabe, angabefehler := eintrag.Info()
		if angabefehler == nil {
			gefunden = append(gefunden, platzdatei{
				pfad: pfad, wiegt: angabe.Size(), alter: angabe.ModTime(),
			})
		}

		return nil
	})
	if fehler != nil {
		return nil, fehler
	}

	return gefunden, nil
}

// sicherungsmuster ist die Form, die eine Vorsicherung wirklich hat.
//
// # Warum ein Muster und nicht der bloße Namensanfang
//
// **Befund 1 des Zweitblicks über Phase 4.13** (2026-09-07): hier stand
// `strings.HasPrefix(name, "bestand-vor-")`, und damit galt **jede** Datei mit
// diesem Anfang als räumbar — auch ein `bestand-vor-notizen.txt`, das jemand
// dort abgelegt hat, oder ein `bestand-vor-0.1.0.db.bak`, das er von Hand
// gezogen hat. Beide waren nach einem Klick auf *Ältere wegräumen* fort.
//
// **Eine löschende Funktion darf keinem Namensanfang vertrauen.** Was die
// Anwendung nicht als ihr eigenes Erzeugnis wiedererkennt, ist *Sonstiges*
// und wird nicht angefasst — das sagt [domain.PostenSonstiges], und hier stand
// das Gegenteil.
//
// # Was das Muster fängt
//
//	bestand-vor-0.4.0.db                 die Sicherung selbst
//	bestand-vor-0.4.0-a1b2c3.db          die zweite unter derselben Fassung
//	bestand-vor-0.4.0.db-wal, …-shm      ihre Begleiter
//	bestand-vor-0.4.0.db.beiseite-…      beim Zurückspielen beiseitegelegt
//
// **Die Fassung ist die Bedingung**, nicht das `.db`.
//
// # Warum das erste Muster nicht reichte
//
// **Der Nachlauf zum Zweitblick hat es angestrichen** (2026-09-07): hier stand
// `bestand-vor-.+\.db`, und damit war ein `bestand-vor-notizen.db` wieder eine
// räumbare Sicherung. Die Gegenprobe deckte nur fremde **Endungen** ab —
// `.txt`, `.bak`, `.md` —, nicht fremde **Namen** mit der richtigen Endung.
//
// **Das ist dieselbe Klasse eine Ebene tiefer**: erst wurde einem
// Namensanfang vertraut, dann einer Endung. Was den Erzeuger wiedererkennt,
// ist die **Fassung** — `Sicherungsname` schreibt sie, und bei einem zweiten
// Stand hängt `sicherungVorMigration` zwölf Hexzeichen an
// ([storage.Bestandsmarke]).
var sicherungsmuster = regexp.MustCompile(
	`^` + service.Vorsicherungsform + `(-wal|-shm|\.beiseite-.+)?$`)

// vorsicherungen findet die Sicherungen vor Migrationen neben dem Bestand.
func vorsicherungen(bestandspfad string) ([]platzdatei, error) {
	neben := filepath.Dir(bestandspfad)

	eintraege, fehler := os.ReadDir(neben)
	if fehler != nil {
		return nil, fehler
	}

	var gefunden []platzdatei

	for _, eintrag := range eintraege {
		if eintrag.IsDir() || !sicherungsmuster.MatchString(eintrag.Name()) {
			continue
		}

		// **Wie bei [unterhalb]**: eine Datei, die inzwischen fort ist, belegt
		// keinen Platz.
		if angabe, angabefehler := eintrag.Info(); angabefehler == nil {
			gefunden = append(gefunden, platzdatei{
				pfad:  filepath.Join(neben, eintrag.Name()),
				wiegt: angabe.Size(), alter: angabe.ModTime(),
			})
		}
	}

	return gefunden, nil
}
