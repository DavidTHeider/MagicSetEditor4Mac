package msepaket

// **Ein Paket als Archiv neu anlegen** — Werkstück C von Phase 4.9
// (#1324).
//
// # Die Lücke, die das schließt
//
// `docs/MSE2BESTAND.md` führt sie beim Befehl `save set as directory`: unsere
// **einzige** Form war das Verzeichnis. MSE2 hat beide und macht das Archiv
// zur **Vorgabe**:
//
//	set->saveAs(name, true, true);
//
// `src/gui/set_window.cpp`, Zeilen 605–614, Fassung f9d9356 vom 2021-07-06.
//
// **`insArchiv` konnte ein vorhandenes umschreiben, aber keines anlegen.**
// Damit war MSEs Vorgabeform hier gar nicht erreichbar — und wer ein Set
// weitergibt, gab einen Ordner weiter statt einer Datei.

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"sort"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// ArchivAnlegen schreibt ein neues Paketarchiv unter `ziel`.
//
// **Über austauschbare Griffe** — dieselbe Gestalt wie bei [nachbardatei] und
// aus demselben Grund: ein Fehlschlag beim Schreiben oder beim Schließen ist
// auf einer gesunden Platte nicht zu bestellen, und ohne diese Naht bliebe
// „das abgebrochene Anlegen" ungeprüft (§3).
//
// # Die Hauptdatei muss dabei sein
//
// **Sonst entstünde ein Archiv, das `Oeffne` nicht liest** — und das fiele
// erst auf, wenn jemand es aufmacht. Welche das ist, sagt [Art.Hauptdatei];
// die Art kommt aus der Endung des Ziels.
//
// # Die Namen werden sortiert geschrieben
//
// **Damit zweimal dasselbe herauskommt.** Eine Go-Abbildung hat keine
// Reihenfolge; ohne diese Zeile unterschieden sich zwei Archive mit gleichem
// Inhalt Byte für Byte, und ein Rundlauftest verglich Zufall.
func ArchivAnlegen(ziel string, dateien map[string][]byte) error {
	return archivAnlegen(ziel, dateien, echteArchivhilfen())
}

// archivziel ist der Griff auf die Datei, die entsteht.
type archivziel interface {
	io.Writer
	Close() error
}

// archivhilfen sind die Griffe, die ein Lauf austauschen kann.
type archivhilfen struct {
	anlegen    func(ziel string) (archivziel, error)
	wegraeumen func(ziel string) error

	// tauschen ersetzt das Ziel durch die fertige Datei daneben.
	tauschen func(von, nach string) error
}

// echteArchivhilfen greifen auf die Platte.
func echteArchivhilfen() archivhilfen {
	return archivhilfen{
		anlegen: func(ziel string) (archivziel, error) {
			// #nosec G304 -- der Aufrufer wählt den Ort, und er prüft ihn:
			// `app.harmloserName` weist jeden Pfad ab und lässt nur einen
			// Dateinamen durch.
			return os.Create(ziel)
		},
		wegraeumen: os.Remove,
		tauschen:   os.Rename,
	}
}

func archivAnlegen(
	ziel string, dateien map[string][]byte, hilfen archivhilfen,
) error {
	art := ArtAusName(ziel)

	haupt := art.Hauptdatei()
	if haupt == "" {
		return fmt.Errorf(
			"%w: %q trägt keine der sieben Paketendungen", domain.ErrValidation, ziel)
	}

	if _, da := dateien[haupt]; !da {
		return fmt.Errorf(
			"%w: dem Archiv %q fehlt seine Hauptdatei %q — ohne sie ließe es "+
				"sich nicht wieder öffnen",
			domain.ErrValidation, ziel, haupt)
	}

	// **Geschrieben wird daneben, getauscht wird am Ende** — Befund 2 des
	// Nachlaufs über die Phasen 4.0 bis 4.9.
	//
	// # Was der erste Anlauf getroffen hat, und was nicht
	//
	// Er räumte nach einem gescheiterten `Close` das Ziel weg — und das ist
	// zu wenig: geöffnet wurde am **endgültigen** Namen, und `os.Create`
	// schneidet eine vorhandene Datei zuvor ab. Wer über ein fertiges Archiv
	// schreibt und dabei scheitert, hat es verloren; der Fehlerweg löschte
	// danach auch noch den Rest.
	//
	// # Wie MSE2 es macht
	//
	//	String tempFile = saveAs + _(".tmp");
	//	…
	//	wxRenameFile(tempFile, saveAs);
	//
	// `src/util/io/package.cpp`, Zeilen 446–489, Fassung f9d9356 vom
	// 2021-07-06. **Ohne sein `.bak`**: MSE2 hält damit die vorige Fassung
	// eines Sets, an dem jemand arbeitet; dies hier ist eine **Ausfuhr**, und
	// eine stille zweite Datei daneben wäre eine Zugabe, um die niemand
	// gebeten hat.
	unterwegs := ziel + ".tmp"

	datei, err := hilfen.anlegen(unterwegs)
	if err != nil {
		return fmt.Errorf("das Archiv %q anlegen: %w", ziel, err)
	}

	if fehler := hineinpacken(datei, dateien); fehler != nil {
		_ = datei.Close()                //nolint:errcheck // der Fehler darunter zählt
		_ = hilfen.wegraeumen(unterwegs) //nolint:errcheck // ein halbes Archiv bleibt nicht liegen

		return fehler
	}

	if fehler := datei.Close(); fehler != nil {
		// **Auch hier wird weggeräumt** — Befund 5 des Zweitblicks über die
		// Phasen 4.0 bis 4.9. Ein Packer schreibt gepuffert: was beim
		// Abschluss schiefgeht, hinterlässt eine Datei, die aussieht wie ein
		// Archiv und keines ist.
		_ = hilfen.wegraeumen(unterwegs) //nolint:errcheck // der Fehler darunter zählt

		return fmt.Errorf("das Archiv %q schließen: %w", ziel, fehler)
	}

	if fehler := hilfen.tauschen(unterwegs, ziel); fehler != nil {
		_ = hilfen.wegraeumen(unterwegs) //nolint:errcheck // der Fehler darunter zählt

		return fmt.Errorf("das Archiv %q an seinen Platz legen: %w", ziel, fehler)
	}

	return nil
}

// hineinpacken schreibt die Dateien in den Packer.
//
// **Ein halbes Archiv bleibt nicht liegen** — der Aufrufer räumt es weg. Das
// ist dieselbe Regel wie beim Schreibweg nebenan: was scheitert, hinterlässt
// keinen Stand, den jemand für einen ganzen hält (§8).
func hineinpacken(nach io.Writer, dateien map[string][]byte) error {
	namen := make([]string, 0, len(dateien))
	for name := range dateien {
		namen = append(namen, name)
	}

	sort.Strings(namen)

	packer := zip.NewWriter(nach)

	for _, name := range namen {
		sauber, err := paketpfad(name)
		if err != nil {
			return err
		}

		// **Immer mit Schrägstrichen.** Ein Paketname trennt so, auch unter
		// Windows (`docs/PORTIERUNG.md`) — und `zip` schreibt, was dasteht.
		ziel, fehler := packer.Create(path.Clean(sauber))
		if fehler == nil {
			_, fehler = ziel.Write(dateien[name])
		}

		if fehler != nil {
			return fmt.Errorf("%s ins Archiv schreiben: %w", name, fehler)
		}
	}

	if err := packer.Close(); err != nil {
		return fmt.Errorf("das Archiv abschließen: %w", err)
	}

	return nil
}
