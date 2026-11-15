package msepaket

// **Pakete als `.mse-installer` weitergeben** — FMT-02.AK-04
// (#1691),
// Marke **ported**.
//
// # Was aus MSE2 übernommen ist
//
// Alles aus Fassung f9d9356 vom 2021-07-06:
//
//   - **Der Aufbau** aus `Installer::addPackage` (`src/data/installer.cpp`,
//     Zeilen 172–194): jede Datei eines Pakets unter `<paket>/<datei>`, eine
//     Beschreibung je Paket, doppelte Pakete nur einmal.
//   - **Die Beschreibung** aus `PackageDescription` (dieselbe Datei, Zeilen
//     198–224, und ihre Reflexion, Zeilen 234–244): ein leerer voller Name
//     wird der kurze, eine leere Gruppe wird aus dem Namen gebildet.
//   - **Der Kurzname** aus `Packaged::validate` (`src/util/io/package.cpp`,
//     Zeilen 621–626): ist er leer, wird er der Dateiname ohne Endung —
//     **immer**, auch wenn ein voller Name dasteht; die Zeile, die ihn
//     setzte, überschreibt MSE2 in der nächsten.
//   - **Welche Dateien reisen** aus `Package::openSubdir` (`package.cpp`,
//     Zeilen 372–392) und `ignore_file` (`src/util/file_utils.cpp`, Zeilen
//     33–37): Unterordner mit Punkt vorn bleiben draußen, ebenso `Thumbs.db`.
//     **Adapted**: auch `.DS_Store` — der Zweck von `ignore_file` ist, was
//     der Dateimanager ungefragt hinlegt, und auf dem Mac ist das der Finder
//     (§15 Satz 9).
//   - **Die Hauptdatei** so, wie MSE2 sie am 2026-09-29 mit
//     `--create-installer` geschrieben hat: BOM, Schlüssel mit Unterstrich,
//     `mse_version: 0.3.7` (`file_version_installer`,
//     `src/util/version.cpp`, Zeile 94), `depends_on:` in einer Zeile.

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// installerfassung ist `file_version_installer` von MSE2 — 0.3.7.
const installerfassung = "0.3.7"

// Installerdatei ist der Name der Hauptdatei eines Installers.
const Installerdatei = "installer"

// rangVorgabe ist `position_hint` eines Pakets ohne Angabe —
// `Packaged::Packaged` (`package.cpp`, Zeilen 548–551).
const rangVorgabe = 100000

// SchreibeInstaller packt die Pakete in einen Installer.
//
// **Die Reihenfolge ist die gegebene**, und ein Paket, das zweimal genannt
// ist, kommt einmal hinein.
func SchreibeInstaller(ziel io.Writer, pakete []*Paket) error {
	if len(pakete) == 0 {
		return errors.New("ein Installer braucht wenigstens ein Paket")
	}

	archiv := zip.NewWriter(ziel)

	var haupt strings.Builder
	haupt.WriteString("\ufeffmse_version: " + installerfassung + "\n" +
		"short_name: \nfull_name: \nicon: \nposition_hint: " + strconv.Itoa(rangVorgabe) + "\n" +
		"installer_group: \nversion: 0.0.0\ncompatible_version: 0.0.0\n")

	gesehen := map[string]bool{}

	for _, paket := range pakete {
		name := filepath.Base(paket.Pfad)
		if gesehen[name] {
			continue
		}

		gesehen[name] = true

		if err := dateienEinpacken(archiv, paket, name); err != nil {
			return err
		}

		beschreibungSchreiben(&haupt, paket, name)
	}

	kopf, err := archiv.Create(Installerdatei)
	if err != nil {
		return fmt.Errorf("die Hauptdatei des Installers anlegen: %w", err)
	}

	// **Ein Fehler in einem Weg**: die Hauptdatei ist klein und landet im
	// Puffer des Archivs — ob die Platte sie nimmt, sagt erst der Abschluss.
	_, schreibfehler := io.WriteString(kopf, haupt.String())
	if err := errors.Join(schreibfehler, archiv.Close()); err != nil {
		return fmt.Errorf("den Installer abschließen: %w", err)
	}

	return nil
}

// dateienEinpacken legt jede Datei des Pakets unter seinem Namen ab.
func dateienEinpacken(archiv *zip.Writer, paket *Paket, name string) error {
	return fs.WalkDir(paket.dateien, ".", func(pfad string, eintrag fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("die Dateien von %s aufzählen: %w", name, err)
		}

		if eintrag.IsDir() {
			if pfad != "." && strings.HasPrefix(eintrag.Name(), ".") {
				return fs.SkipDir
			}

			return nil
		}

		if eintrag.Name() == "Thumbs.db" || eintrag.Name() == ".DS_Store" {
			return nil
		}

		inhalt, err := leseBegrenzt(paket.dateien, pfad)
		if err != nil {
			return fmt.Errorf("%s in %s lesen: %w", pfad, name, err)
		}

		schreiber, err := archiv.CreateHeader(&zip.FileHeader{Name: path.Join(name, pfad), Method: zip.Deflate})
		if err != nil {
			return fmt.Errorf("%s in den Installer legen: %w", pfad, err)
		}

		if _, err := io.Copy(schreiber, bytes.NewReader(inhalt)); err != nil {
			return fmt.Errorf("%s in den Installer schreiben: %w", pfad, err)
		}

		return nil
	})
}

// beschreibungSchreiben hängt den `package:`-Block eines Pakets an.
func beschreibungSchreiben(haupt *strings.Builder, paket *Paket, name string) {
	kurz := paket.KurzerName
	if kurz == "" {
		kurz = strings.TrimSuffix(name, path.Ext(name))
	}

	voll := paket.VollerName
	if voll == "" {
		voll = kurz
	}

	rang := paket.Rang
	if rang == RangOhneHinweis {
		rang = rangVorgabe
	}

	fmt.Fprintf(haupt, "package:\n\tname: %s\n\tversion: %s\n\tshort_name: %s\n\tfull_name: %s\n"+
		"\ticon_url: %s\n\tinstaller_group: %s\n\tposition_hint: %d\n\tdescription: \n",
		name, fassungstext(paket.Version), kurz, voll, wert(paket.Kopf.Ersten("icon")),
		installergruppe(paket, name, kurz), rang)

	for _, abhaengigkeit := range paket.Abhaengigkeiten {
		if abhaengigkeit.Version.Leer() {
			fmt.Fprintf(haupt, "\tdepends_on: %s\n", abhaengigkeit.Paket)

			continue
		}

		fmt.Fprintf(haupt, "\tdepends_on: %s %s\n", abhaengigkeit.Paket, abhaengigkeit.Version)
	}
}

// installergruppe gibt `installer group` — oder bildet sie wie MSE2 aus dem
// Namen: bis zum letzten Bindestrich, Bindestriche als Schrägstriche, darunter
// „Game files" für ein Spiel und sonst der Kurzname.
func installergruppe(paket *Paket, name, kurz string) string {
	if paket.Gruppe != "" {
		return paket.Gruppe
	}

	gruppe := strings.TrimSuffix(name, path.Ext(name))
	if strich := strings.LastIndex(gruppe, "-"); strich >= 0 {
		gruppe = gruppe[:strich]
	}

	gruppe = "unclassified/" + strings.ReplaceAll(gruppe, "-", "/") + "/"

	if paket.Art == ArtSpiel {
		return gruppe + "Game files"
	}

	return gruppe + kurz
}

// fassungstext schreibt eine Fassung wie MSE2 — ohne Angabe `0.0.0`.
func fassungstext(fassung Datum) string {
	if fassung.Leer() {
		return "0.0.0"
	}

	return fassung.String()
}
