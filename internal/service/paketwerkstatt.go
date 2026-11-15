// **Der Arbeitsbereich für Vorlagen** — Werkstück 2 von Phase 3.3
// (#1164).
//
// # Wofür
//
// Exportvorlagen ließen sich ausführen, nicht bearbeiten; ein `.mse-style`
// war nur über einen Texteditor außerhalb der Anwendung zu ändern. Das ist
// der Schritt dazwischen: eine Datei eines installierten Pakets aufschlagen,
// ändern und sicher zurückschreiben.
//
// # Was hier entschieden wird und nicht in `internal/msepaket`
//
// Das Blatt kann eine Datei zurückschreiben. **Ob es das darf**, ist fachlich
// und steht deshalb hier:
//
//   - **Welche Dateien bearbeitbar sind.** Ein `.png` ist keine Vorlage.
//   - **Dass ein Paket unter der Wurzel liegt.** Ein Pfad von außen ist keine
//     Paketwahl — der Name kommt aus dem Fenster, von der Kommandozeile, aus
//     einem MCP-Aufruf und über die Rückschleife (§16).
//   - **Woran ein Konflikt erkannt wird**: am Stand, den der Aufrufer beim
//     Lesen bekommen hat.

package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Textendungen sind die Endungen, die trotz Endung Text tragen.
//
// # Woher die Liste kommt
//
// Gezählt über die 526 Pakete von `full-magic-pack` am 2026-08-26: MSEs eigene
// Datendateien tragen **keine Endung** — `style`, `game`, `card_fields`,
// `include`, `locale`; 777 Dateien unter 133 verschiedenen Namen. Alles mit
// Endung ist Beiwerk: 30.245 `.png`, 3.242 `.jpg`, 225 `.bat`, 55 `.ttf`, 27
// `.dic`, 14 `.ini`. Die einzige Ausnahme sind 31 `.txt`.
//
// **Keine der 777 und keine der 31 trägt ein Nullbyte.** Das ist die
// Reichweite dieser Messung (§10): sie gilt für dieses Pack. Ein Paket, das
// eine `.png` mit Text füllt, gäbe es hier nicht zu bearbeiten — und das ist
// die richtige Seite, auf der man sich irrt.
var Textendungen = []string{".txt"}

// Paketwerkstatt liest und schreibt die Dateien installierter Pakete.
type Paketwerkstatt struct{ wurzel string }

// NeuePaketwerkstatt baut es über dem Paketverzeichnis.
//
// **Über einer Wurzel und nicht über freien Pfaden** — dieselbe Gestalt wie
// [NeuerPaketbezieher] und aus demselben Grund: was unter der Wurzel liegt,
// entscheidet dieses Programm, und alles andere ist nicht erreichbar.
func NeuePaketwerkstatt(wurzel string) *Paketwerkstatt { return &Paketwerkstatt{wurzel: wurzel} }

// Dateien nennt die bearbeitbaren Dateien eines Pakets.
func (v *Paketwerkstatt) Dateien(paket string) ([]string, error) {
	dateien, _, err := v.DateienUndBeiwerk(paket)

	return dateien, err
}

// DateienUndBeiwerk nennt die bearbeitbaren Dateien eines Pakets und, getrennt
// davon, sein **Beiwerk** — Bilder, Schriften, alles, was sich nicht als Text
// bearbeiten lässt (FMT-02, #1691). Ohne diese zweite Liste sähe niemand die
// Bilder, die er selbst in sein Paket gelegt hat.
func (v *Paketwerkstatt) DateienUndBeiwerk(paket string) (dateien, beiwerk []string, fehler error) {
	ort, err := v.ort(paket)
	if err != nil {
		return nil, nil, err
	}

	geoeffnet, oeffnefehler := msepaket.Oeffne(ort)
	if oeffnefehler != nil {
		return nil, nil, oeffnefehler
	}

	defer func() { _ = geoeffnet.Schliessen() }() //nolint:errcheck // nur gelesen

	alle, lesefehler := geoeffnet.Dateien()
	if lesefehler != nil {
		return nil, nil, lesefehler
	}

	dateien, beiwerk = make([]string, 0, len(alle)), []string{}

	for _, name := range alle {
		if Bearbeitbar(name) {
			dateien = append(dateien, name)
		} else {
			beiwerk = append(beiwerk, name)
		}
	}

	return dateien, beiwerk, nil
}

// Lesen gibt Inhalt und Stand einer Datei.
//
// **Der Stand ist die SHA-256 des Inhalts** und keine Änderungszeit: eine
// Uhr geht auf zwei Rechnern verschieden, und ein Dateisystem rundet sie.
func (v *Paketwerkstatt) Lesen(paket, datei string) (inhalt, stand string, fehler error) {
	ort, err := v.ort(paket)
	if err != nil {
		return "", "", err
	}

	if !Bearbeitbar(datei) {
		return "", "", nichtBearbeitbar(datei)
	}

	geoeffnet, oeffnefehler := msepaket.Oeffne(ort)
	if oeffnefehler != nil {
		return "", "", oeffnefehler
	}

	defer func() { _ = geoeffnet.Schliessen() }() //nolint:errcheck // nur gelesen

	roh, lesefehler := geoeffnet.Datei(datei)
	if lesefehler != nil {
		return "", "", fmt.Errorf("%w: %w", domain.ErrNotFound, lesefehler)
	}

	return string(roh), Stand(roh), nil
}

// Schreiben setzt eine Datei zurück und gibt den neuen Stand.
//
// **Der neue Stand kommt mit der Antwort.** Ohne ihn müsste der Aufrufer nach
// jedem Speichern neu lesen, und zwischen Antwort und Lesen läge wieder ein
// Fenster, in dem jemand anderes schreibt (§7, Regel 2).
func (v *Paketwerkstatt) Schreiben(ctx context.Context, paket, datei, inhalt, erwartet string) (string, error) {
	ort, err := v.ort(paket)
	if err != nil {
		return "", err
	}

	if !Bearbeitbar(datei) {
		return "", nichtBearbeitbar(datei)
	}

	// **Die geteilte Sperre am umgebenden Bezugsbaum** (#1190). Die
	// Schreibsperre von `msepaket` liegt an der **Paketdatei**; ein Bezug
	// tauscht den **Baum** aus, und die beiden Ebenen sahen einander nicht.
	//
	// **Geteilt und nicht ausschließlich**: mehrere dürfen gleichzeitig an
	// Dateien darunter arbeiten — was zwei Schreiber derselben Datei angeht,
	// entscheidet die Sperre eine Ebene tiefer.
	fertig, sperrfehler := GeteiltAmBaum(ctx, bezugsbaum(v.wurzel, ort))
	if sperrfehler != nil {
		return "", sperrfehler
	}

	defer fertig()

	if fehler := msepaket.DateiSchreiben(ctx, ort, datei, []byte(inhalt), erwartet); fehler != nil {
		return "", fehler
	}

	return Stand([]byte(inhalt)), nil
}

// Texte gibt jede bearbeitbare Datei des Pakets samt Inhalt — der Stoff des
// Symbolindex ([Paketsymbole], U7, #1418).
//
// **Das Paket wird einmal geöffnet**, nicht je Datei: `Dateien` und `Lesen`
// nacheinander öffneten es für jede Datei neu, und ein Index über zwanzig
// Dateien wäre zwanzig Öffnungen.
//
// **Eine Datei, die zwischen Aufzählen und Lesen verschwindet, fehlt im
// Index** — und nimmt den anderen nichts (§8). Der Fall ist ein Wettlauf mit
// einem Editor daneben, und dafür einen Lauf abzubrechen wäre die härtere
// Antwort, als er verdient.
func (v *Paketwerkstatt) Texte(paket string) ([]Paketdateitext, error) {
	ort, err := v.ort(paket)
	if err != nil {
		return nil, err
	}

	geoeffnet, oeffnefehler := msepaket.Oeffne(ort)
	if oeffnefehler != nil {
		return nil, oeffnefehler
	}

	defer func() { _ = geoeffnet.Schliessen() }() //nolint:errcheck // nur gelesen

	alle, lesefehler := geoeffnet.Dateien()
	if lesefehler != nil {
		return nil, lesefehler
	}

	texte := make([]Paketdateitext, 0, len(alle))

	for _, name := range alle {
		if !Bearbeitbar(name) {
			continue
		}

		roh, dateifehler := geoeffnet.Datei(name)
		if dateifehler != nil {
			continue
		}

		texte = append(texte, Paketdateitext{Name: name, Inhalt: string(roh)})
	}

	return texte, nil
}

// Bearbeitbar sagt, ob eine Datei eines Pakets eine Vorlage ist.
func Bearbeitbar(name string) bool {
	endung := path.Ext(path.Base(name))
	if endung == "" {
		return true
	}

	for _, erlaubt := range Textendungen {
		if strings.EqualFold(endung, erlaubt) {
			return true
		}
	}

	return false
}

// Stand gibt den Stand eines Inhalts.
func Stand(inhalt []byte) string {
	roh := sha256.Sum256(inhalt)

	return hex.EncodeToString(roh[:])
}

// ort löst die Angabe eines Pakets auf und hält sie unter der Wurzel.
//
// # Warum ein Pfad und nicht ein Name
//
// **Weil Pakete verschachtelt liegen.** Bezogen wird eine ganze Sammlung, und
// darin steckt der Baum, den ihr Urheber gewählt hat:
//
//	pakete/M15-Magic-Pack/data/magic-m15.mse-style
//
// Der erste Anlauf verlangte einen reinen Namen — und erreichte damit **kein
// einziges** installiertes Paket. Gefunden bei der Sichtprüfung im laufenden
// Fenster: die Seite bot 327 Pakete an und sagte bei jedem „hat keine Datei
// zum Bearbeiten".
//
// # Was stattdessen geprüft wird
//
// **Dass das Ergebnis unter der Wurzel liegt** — und das ist die eigentliche
// Zusage. Ein `..` mitten im Pfad, ein absoluter Pfad woanders hin, ein
// `a/../../b`: alle drei enden außerhalb, und alle drei fängt derselbe
// Vergleich. Eine Aufzählung verbotener Zeichen fängt nur die, an die jemand
// gedacht hat.
//
// **Absolut ist erlaubt**, denn genau das gibt der Paketkatalog: `PaketDTO`
// trägt den vollen Pfad, und das Fenster reicht ihn weiter. Relativ ist
// erlaubt, weil es sich an der Kommandozeile tippen lässt.
func (v *Paketwerkstatt) ort(paket string) (string, error) {
	if paket == "" {
		return "", fmt.Errorf("%w: ohne Paket gibt es nichts zu lesen", domain.ErrValidation)
	}

	ort := filepath.FromSlash(paket)
	if !filepath.IsAbs(ort) {
		ort = filepath.Join(v.wurzel, ort)
	}

	ort = filepath.Clean(ort)

	wurzel := filepath.Clean(v.wurzel)
	if !strings.HasPrefix(ort, wurzel+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: %q liegt nicht im Paketverzeichnis",
			domain.ErrValidation, paket)
	}

	return ort, nil
}

// nichtBearbeitbar meldet eine Datei, die keine Vorlage ist.
func nichtBearbeitbar(datei string) error {
	return fmt.Errorf("%w: %s ist keine Vorlage, sondern Beiwerk", domain.ErrValidation, datei)
}

// bezugsbaum gibt den Bezugsbaum, in dem ein Paket liegt.
//
// **Das erste Glied unter der Wurzel.** `internal/bezug` holt ein ganzes
// Repositorium und legt es als `<wurzel>/<name>/` ab; die Pakete darin stehen
// tiefer — `<wurzel>/M15-Magic-Pack/data/magic-m15.mse-style`. Gesperrt wird
// gegen den **Tausch des Baums**, und der betrifft das erste Glied
// (#1190).
//
// **Ein Paket unmittelbar unter der Wurzel ist sein eigener Baum.** Es kam
// nicht über einen Bezug, und die Sperre trifft dann genau es — das ist
// richtig: `Beziehen` könnte es unter demselben Namen ersetzen.
func bezugsbaum(wurzel, ort string) string {
	rest, err := filepath.Rel(filepath.Clean(wurzel), ort)
	if err != nil {
		return ort
	}

	glieder := strings.Split(rest, string(filepath.Separator))
	if len(glieder) == 0 || glieder[0] == "" || glieder[0] == ".." {
		return ort
	}

	return filepath.Join(filepath.Clean(wurzel), glieder[0])
}
