package service

import (
	"archive/zip"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// unlesbar ordnet ein, was an einer Datei **selbst** scheitert — FMT-06.AK-04
// (#1695).
//
// # Warum eine Ablehnung und keine Panne
//
// **Eine Datei, die kein gültiges UTF-8 ist, ein Archiv, das mittendrin
// abbricht, eine Datei über der Grenze** — das ist eine Eigenschaft der
// Eingabe, und wer sie gewählt hat, kann etwas tun: eine andere Datei nehmen,
// die Stelle berichtigen. Bis FMT-06 kamen alle drei als „Es ist ein
// unerwarteter Fehler aufgetreten" an, und die Stelle, die `msepaket` seit
// AK-01 nennt, erreichte niemanden.
//
// # Wie sie ankommt
//
// **Als Validierung mit eigenem Satz** (`domain.ErrValidation`): die
// Oberfläche setzt ihn in „Das ging nicht: … Berichtige die Angabe …" ein
// (`app.validierungsgrund`). Der Satz nennt Datei, Zeile und Spalte in Worten;
// die technische Ursache reist als Grund ins Protokoll ([0075](../../docs/entscheidungen/0075-der-satz-und-sein-grund.md)).
//
// **Der Satz beginnt mit der Datei**, nicht mit dem Vorgang: im Fenster steht
// er hinter „Das ging nicht:", und „Set öffnen:" davor sagte zweimal, was
// scheiterte. Die Stelle nennt die Datei **im** Paket — `msepaket` setzt den
// Paketnamen davor, und der steht schon am Anfang.
//
// **Was keines davon ist, ordnet es nicht ein** und gibt `nil` — eine Datei,
// die es nicht gibt, oder ein Verzeichnis ohne Leserecht sagt der Aufrufer
// selbst, über [unzugaenglich].
func unlesbar(pfad string, fehler error) error {
	name := filepath.Base(pfad)

	var stelle *msepaket.Zerteilfehler
	if errors.As(fehler, &stelle) {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"%s lässt sich nicht lesen — in »%s«, Zeile %d, Spalte %d: %s",
			name, strings.TrimPrefix(stelle.Stelle.Datei, name+"/"),
			stelle.Stelle.Zeile, stelle.Stelle.Spalte, stelle.Meldung)
	}

	if errors.Is(fehler, zip.ErrFormat) || errors.Is(fehler, zip.ErrChecksum) {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"%s ist kein lesbares Archiv — beschädigt oder unvollständig", name)
	}

	// **Nur die Grenze der Paketdatei.** Die des Bildes (`ErrBildZuGross`)
	// greift erst beim Auspacken eines Bildes, und das geschieht beim Öffnen
	// nicht — ein Zweig dafür hätte keinen Fall.
	// **Die Ressourcenregel für Archive** (#1886): ein Archiv, das mehr
	// verspricht, als ein Paket haben darf, wird nicht gelesen.
	if errors.Is(fehler, msepaket.ErrArchivZuGross) {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"%s lässt sich nicht öffnen — das Archiv hat mehr als %d Einträge "+
				"oder ein Verzeichnis über %d MiB",
			name, msepaket.GrenzeArchiveintraege, msepaket.GrenzeArchivverzeichnis>>20)
	}

	if errors.Is(fehler, msepaket.ErrZuGross) {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"%s lässt sich nicht öffnen — eine Datei darin ist größer als %d MiB",
			name, msepaket.GrenzePaketdatei>>20)
	}

	return nil
}

// unzugaenglich ordnet ein, was **vor** dem Inhalt scheitert — die Datei
// fehlt, oder sie lässt sich nicht öffnen (FMT-03.AK-01,
// #1692).
//
// **Bis dahin sagte es kein Aufrufer**: eine fehlende MTG-JSON-Datei und
// jede Datei ohne Leserecht gingen als unbekannter Fehler hinaus, und der
// Dienst machte daraus `500` — ein Tippfehler im Pfad sah aus wie ein
// kaputtes Programm. Ein Setpaket, das fehlt, meldet `msepaket` schon selbst
// als nicht gefunden; dieser Zweig trifft dort nur das Leserecht.
func unzugaenglich(pfad string, fehler error) error {
	name := filepath.Base(pfad)

	switch {
	case errors.Is(fehler, fs.ErrNotExist):
		return domain.MitGrund(domain.ErrNotFound, fehler, "%s gibt es nicht", name)
	case errors.Is(fehler, fs.ErrPermission):
		return domain.MitGrund(domain.ErrValidation, fehler,
			"%s lässt sich nicht lesen — keine Berechtigung", name)
	}

	return nil
}
