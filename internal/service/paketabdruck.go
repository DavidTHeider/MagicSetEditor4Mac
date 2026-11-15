package service

// **Der Abdruck eines Pakets** — DOM-01.AK-01
// (#1696):
// „lokale Änderungen sind am Objekt sichtbar und filterbar."
//
// # Warum Größe und Zeit und nicht der Inhalt
//
// **Weil die Liste bei jedem Blick gerechnet wird.** Ein bezogenes
// Repositorium wiegt über ein Gigabyte; seinen Inhalt bei jedem Öffnen der
// Seite zu lesen, kostete Sekunden. Der Abdruck liest nur, was das
// Dateisystem ohnehin weiß: Name, Größe und Änderungszeit jeder Datei.
//
// **Was das kostet, gehört ausgesprochen**: eine Datei, die jemand nur
// berührt, gilt als geändert — und eine, die jemand auf Byte und Sekunde
// genau zurückschreibt, als ungeändert. Das erste ist eine Vorsicht, das
// zweite kommt ohne Absicht nicht vor.
//
// Portierungsmarke **adapted**, und das Maß ist dasselbe: MSE2 vergleicht in
// `PackageVersion::check_status` die **Änderungszeit jeder Datei** mit der
// Liste, die beim Segnen (`bless`) festgehalten wurde, und zählt eine
// hinzugekommene oder verschwundene Datei als Änderung
// (`src/util/io/package_manager.cpp`, Zeilen 446–484, Fassung f9d9356 vom
// 2021-07-06). Hier kommt die Größe dazu, und gesegnet wird beim Einbinden.
// **MSE2 zeigt das Ergebnis nicht**: `InstallablePackage::determineStatus`
// liest es nur hinter `USE_MODIFIED_CHECK`, und der Schalter ist aus, „weil
// wir Pakete nicht segnen können" (`src/data/installer.cpp`, Zeilen 25 und
// 306). Hier segnet das Einbinden, und damit gibt es den Grund nicht.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// abdruckVon rechnet den Abdruck eines Pakets — eines Verzeichnisses oder
// einer einzelnen Datei.
func abdruckVon(pfad string) (string, error) {
	summe := sha256.New()

	// **`Walk` und nicht `WalkDir`**: es liefert die Angaben jeder Datei mit,
	// und ein zweiter Ruf danach wäre ein Fehlerweg, den nur ein Rennen
	// beträte.
	fehler := filepath.Walk(pfad, func(ort string, angaben os.FileInfo, fehler error) error {
		if fehler != nil || angaben.IsDir() {
			return fehler
		}

		name, _ := filepath.Rel(pfad, ort) //nolint:errcheck // ort liegt unter pfad, Walk sagt es zu
		// **`hash.Hash.Write` scheitert nie** — die Schnittstelle sagt es zu,
		// und errcheck kennt die Ausnahme. Ein `Fprintf` davor gäbe einen
		// Fehlerweg, den es nicht gibt.
		summe.Write([]byte(filepath.ToSlash(name) + "\x00" + strconv.FormatInt(angaben.Size(), 10) +
			"\x00" + strconv.FormatInt(angaben.ModTime().UnixNano(), 10) + "\n"))

		return nil
	})
	if fehler != nil {
		return "", fmt.Errorf("den Abdruck von %q rechnen: %w", pfad, fehler)
	}

	return "abdruck:" + hex.EncodeToString(summe.Sum(nil)), nil
}

// mitAenderung setzt `Geaendert` an jedem Paket, das einen Abdruck trägt und
// heute anders dasteht. **Ein Paket, dessen Abdruck sich nicht rechnen
// lässt, gilt als geändert** — es steht jedenfalls nicht mehr so da wie beim
// Einbinden.
func mitAenderung(pakete []domain.Paket) []domain.Paket {
	for i, paket := range pakete {
		if paket.Abdruck == "" {
			continue
		}

		jetzt, fehler := abdruckVon(paket.Pfad)
		pakete[i].Geaendert = fehler != nil || jetzt != paket.Abdruck
	}

	return pakete
}
