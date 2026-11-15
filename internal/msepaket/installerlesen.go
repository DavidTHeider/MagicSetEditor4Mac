package msepaket

// **Einen `.mse-installer` lesen und auspacken** — FMT-02.AK-04
// (#1691).
//
// # Was aus MSE2 übernommen ist (ported)
//
// Fassung f9d9356 vom 2021-07-06:
//
//   - **Die Hauptdatei** heißt `installer` und nennt je Paket einen Block
//     `package:` (`IMPLEMENT_REFLECTION(Installer)`, `src/data/installer.cpp`,
//     Zeilen 32–35; `PackageDescription`, Zeilen 234–244).
//   - **Ausgepackt wird, was unter dem Paketnamen liegt**, ohne das Präfix
//     (`PackageDirectory::actual_install`, `src/util/io/package_manager.cpp`,
//     Zeilen 376–400) — dort ohne Rücksicht auf Groß- und Kleinschreibung
//     (`is_substr_i`), hier ebenso.
//
// # Wo es enger ist als MSE2 (adapted)
//
// **Ein Installer kommt von fremder Hand**, und MSE2 prüft seine Einträge
// nicht: `install_dir + file.substr(name.length())` folgte einem `../`.
// Hier gilt:
//
//   - **Bis zum Schrägstrich**: `probe.mse-game-x/` gehört nicht zu
//     `probe.mse-game`.
//   - **Kein Weg hinaus**: ein Eintrag mit `..` oder absolutem Pfad wird
//     benannt und nicht ausgepackt; geschrieben wird durch eine [os.Root].
//   - **Kein Verweis**: ein symbolischer Link wird benannt und übergangen.
//   - **Ein Name ist ein Paketname** — ein Glied, mit einer der sechs
//     Vorlagenendungen; ein Set gehört nicht in einen Installer.
//   - **Ein Deckel** über die ausgepackten Bytes, den der Aufrufer setzt.

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
)

// ErrInstallerZuGross sagt, dass ein Installer mehr auspacken wollte, als
// der Aufrufer erlaubt hat.
var ErrInstallerZuGross = errors.New("der Installer ist größer als erlaubt")

// Installerpaket ist die Beschreibung eines Pakets in einem Installer.
type Installerpaket struct {
	Name            string
	Version         Datum
	KurzerName      string
	VollerName      string
	Gruppe          string
	Abhaengigkeiten []Abhaengigkeit
}

// Installer ist ein geöffneter `.mse-installer`.
type Installer struct {
	Pfad    string
	Pakete  []Installerpaket
	Befunde []Befund

	archiv  io.Closer
	dateien map[string][]*zip.File
}

// OeffneInstaller öffnet einen Installer und liest seine Beschreibung.
//
// **Die Einträge werden hier zugeordnet und nicht erst beim Auspacken** —
// wer fragt, was darin ist, soll auch erfahren, was davon nicht mitkommt.
func OeffneInstaller(pfad string) (*Installer, error) {
	// **Nach der Ressourcenregel** (#1886) — ein Installer ist ein Archiv
	// wie jedes Paket.
	archiv, griff, err := oeffneArchiv(pfad)
	// **Ein unsicherer Name ist hier ein Befund und kein Abbruch** — die
	// Prüfung unten benennt jeden einzeln; mit `zipinsecurepath=0` käme
	// dieser Fehler neben einem brauchbaren Leser.
	if err != nil && !errors.Is(err, zip.ErrInsecurePath) {
		return nil, fmt.Errorf("%q lässt sich nicht als Installer öffnen: %w", pfad, err)
	}

	installer := &Installer{Pfad: pfad, archiv: griff, dateien: map[string][]*zip.File{}}

	inhalt, err := leseBegrenzt(archiv, Installerdatei)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("%q trägt keine Hauptdatei %q — ist es ein Installer?", pfad, Installerdatei),
			griff.Close())
	}

	baum, err := Zerteile(Installerdatei, inhalt)
	if err != nil {
		return nil, errors.Join(err, griff.Close())
	}

	installer.beschreibungenLesen(baum)
	installer.eintraegeZuordnen(archiv.File)

	return installer, nil
}

// beschreibungenLesen liest die `package:`-Blöcke der Hauptdatei.
func (i *Installer) beschreibungenLesen(baum *Baum) {
	gesehen := map[string]bool{}

	for _, block := range baum.Alle("package") {
		name := wert(block.Ersten("name"))

		art := ArtAusName(name)
		switch {
		case name == "" || strings.ContainsAny(name, `/\`) || name != path.Clean(name) || strings.HasPrefix(name, "."):
			i.melde(block.Stelle, fmt.Sprintf("%q ist kein Paketname — übergangen", name))

			continue
		case art == ArtUnbekannt || art == ArtSet:
			i.melde(block.Stelle, fmt.Sprintf("%q trägt keine Vorlagenendung — übergangen", name))

			continue
		case gesehen[strings.ToLower(name)]:
			i.melde(block.Stelle, fmt.Sprintf("%q steht zweimal im Installer — der zweite übergangen", name))

			continue
		}

		gesehen[strings.ToLower(name)] = true
		i.Pakete = append(i.Pakete, i.beschreibung(block, name))
	}
}

// beschreibung liest einen Block; eine unlesbare Fassung ist ein Befund.
func (i *Installer) beschreibung(block *Knoten, name string) Installerpaket {
	paket := Installerpaket{
		Name:       name,
		KurzerName: wert(block.Ersten("short name")),
		VollerName: wert(block.Ersten("full name")),
		Gruppe:     strings.TrimSpace(wert(block.Ersten("installer group"))),
	}

	// **`0.0.0` heißt „keine"** — so schreibt MSE2 eine leere Fassung.
	if fassung := block.Ersten("version"); fassung != nil && fassung.Wert != "" && fassung.Wert != "0.0.0" {
		datum, err := DatumLesen(fassung.Wert)
		if err != nil {
			i.melde(fassung.Stelle, err.Error())
		}

		paket.Version = datum
	}

	for _, verweis := range block.Alle("depends on") {
		abhaengigkeit, stelle, err := abhaengigkeitAus(verweis)
		if err != nil {
			i.melde(stelle, err.Error())
		}

		if abhaengigkeit.Paket != "" {
			paket.Abhaengigkeiten = append(paket.Abhaengigkeiten, abhaengigkeit)
		}
	}

	return paket
}

// eintraegeZuordnen hängt jede Datei an das Paket, unter dessen Namen sie
// liegt — was keinem genannten gehört, bleibt still liegen, wie bei MSE2.
func (i *Installer) eintraegeZuordnen(dateien []*zip.File) {
	for _, datei := range dateien {
		if strings.HasSuffix(datei.Name, "/") {
			continue
		}

		for _, paket := range i.Pakete {
			vorn := len(paket.Name) + 1
			if len(datei.Name) <= vorn || !strings.EqualFold(datei.Name[:vorn], paket.Name+"/") {
				continue
			}

			stelle := Stelle{Datei: datei.Name}

			switch rest := datei.Name[vorn:]; {
			case !fs.ValidPath(rest) || strings.Contains(rest, `\`):
				i.melde(stelle, "der Eintrag führt aus dem Paket heraus — nicht ausgepackt")
			case datei.Mode()&fs.ModeSymlink != 0:
				i.melde(stelle, "der Eintrag ist ein Verweis — nicht ausgepackt")
			default:
				i.dateien[paket.Name] = append(i.dateien[paket.Name], datei)
			}

			break
		}
	}
}

// Auspacken legt die Dateien eines Pakets in `ziel` und gibt die Bytes.
//
// **Mehr als `deckel` Bytes gibt es nicht** — der Aufrufer zählt über alle
// Pakete eines Installers.
func (i *Installer) Auspacken(name, ziel string, deckel int64) (int64, error) {
	dateien := i.dateien[name]
	if len(dateien) == 0 {
		return 0, fmt.Errorf("%q trägt im Installer keine Datei", name)
	}

	wurzel, err := os.OpenRoot(ziel)
	if err != nil {
		return 0, fmt.Errorf("das Ziel %q öffnen: %w", ziel, err)
	}

	defer func() { _ = wurzel.Close() }() //nolint:errcheck // nur geschrieben, jede Datei ist geschlossen

	var benutzt int64

	for _, datei := range dateien {
		geschrieben, err := dateiAuspacken(wurzel, datei, datei.Name[len(name)+1:], deckel-benutzt)
		benutzt += geschrieben

		if err != nil {
			return benutzt, err
		}
	}

	return benutzt, nil
}

// dateiAuspacken schreibt einen Eintrag — höchstens `rest` Bytes.
func dateiAuspacken(wurzel *os.Root, datei *zip.File, name string, rest int64) (int64, error) {
	if ordner := path.Dir(name); ordner != "." {
		if err := wurzel.MkdirAll(ordner, 0o750); err != nil {
			return 0, fmt.Errorf("den Ordner %q anlegen: %w", ordner, err)
		}
	}

	leser, err := datei.Open()
	if err != nil {
		return 0, fmt.Errorf("%q im Installer lesen: %w", datei.Name, err)
	}

	defer func() { _ = leser.Close() }() //nolint:errcheck // nur gelesen

	ziel, err := wurzel.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return 0, fmt.Errorf("%q anlegen: %w", name, err)
	}

	geschrieben, kopierfehler := io.Copy(ziel, io.LimitReader(leser, rest+1))

	if err := errors.Join(kopierfehler, ziel.Close()); err != nil {
		return geschrieben, fmt.Errorf("%q auspacken: %w", name, err)
	}

	if geschrieben > rest {
		return geschrieben, fmt.Errorf("%w — bei %q abgebrochen", ErrInstallerZuGross, name)
	}

	return geschrieben, nil
}

// Schliessen gibt das Archiv frei.
func (i *Installer) Schliessen() error {
	return i.archiv.Close()
}

func (i *Installer) melde(stelle Stelle, meldung string) {
	i.Befunde = append(i.Befunde, Befund{Stelle: stelle, Meldung: meldung})
}
