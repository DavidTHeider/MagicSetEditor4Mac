// **Das Profilbild** — Werkstück 8 von Phase 4.12
// (#1351),
// Punkt 17 der Kommentierung vom 2026-09-05.
//
// # Warum es neben den Kartenbildern liegt und nicht darunter
//
// **Weil ein Kartenbild einem Set gehört und ein Profilbild einem Menschen.**
// Die Kartenbilder stehen in `beiwerk/<setID>/`; ein Profilbild dort abzulegen
// hieße, es zu verlieren, sobald jemand das Set löscht — und ein Mensch gehört
// keinem Set.
//
// **Deshalb `beiwerk/nutzer/`**, ein Name, den keine Setkennung annehmen kann:
// die sind Zahlen.
//
// # Und warum es als PNG landet, gleich was hereinkam
//
// **Dieselbe Antwort wie bei den Kartenbildern** — MSE selbst tut es so
// (`src/gui/value/image.cpp`, Zeile 51, Fassung f9d9356 vom 2021-07-06:
// „always use PNG images, see #69"), und hier kommt ein Grund dazu: das
// Umkodieren **beweist**, dass die gewählte Datei ein Bild ist. Wer eine `.exe` mit der Endung `.png` wählt, bekommt
// eine Meldung und keinen Eintrag im Bestand.

package service

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image"

	// Für `image.Decode`: dieselben drei Formate, die der Dateidialog
	// anbietet. Geschrieben wird immer PNG — gelesen wird großzügiger, damit
	// ein von Hand hineingelegtes Bild nicht als kaputt gemeldet wird.
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Nutzerbildordner ist das Verzeichnis unter der Wurzel.
//
// **Ein Wort und keine Zahl** — so kann es mit keinem Set kollidieren.
const Nutzerbildordner = "nutzer"

// Nutzerbildkante ist die Schranke, in die ein Profilbild passen muss.
//
// **Als Schranke und nicht als Maß**: das Seitenverhältnis der Quelle bleibt,
// und ein Bild passt **in** dieses Quadrat — dieselbe Bauform wie
// [Vorschaubreite]. Eine feste Größe stauchte ein hochkant aufgenommenes
// Gesicht.
//
// **256 und nicht 128**: ein Abzeichen von 128 Punkten steht auf einem
// Netzhautbildschirm doppelt so fein da, und die Datei bleibt unter 100 kB.
const Nutzerbildkante = 256

// Nutzerverzeichnis gibt den Ort der Profilbilder — leer ohne Wurzel.
func (b *Bildablage) Nutzerverzeichnis() string {
	if b.wurzel == "" {
		return ""
	}

	return filepath.Join(b.wurzel, Nutzerbildordner)
}

// NutzerbildAblegen schreibt das Bild eines Nutzers und gibt seinen Namen.
//
// **Der Name trägt die Kennung und nicht den Anzeigenamen**: der ändert sich,
// und ein Bild, das danach unter einem anderen Namen läge, wäre verloren.
//
// **Und jedes Ablegen bekommt einen eigenen Namen** — Befund 3 des
// Zweitblicks über Phase 6.1 und Befund 1 seines Nachlaufs. Hier stand „er
// wird überschrieben": ein Mensch hat **ein** Profilbild. Aber das
// Überschreiben geschah, **bevor** der Bestand die Fassung prüfte, und ein
// abgewiesenes Schreiben ließ das neue Bild unter dem alten Verweis stehen.
// Der erste Ersatz nahm den **Inhalt** als Namen — und zwei gleichzeitige
// Wechsel auf dasselbe Bild teilten sich damit eine Datei: der Verlierer
// räumte die des Gewinners weg. Jetzt ist der Name gewürfelt; der Bestand
// wechselt den Verweis, und erst danach geht das alte
// ([Bildablage.BildWegnehmen]).
func (b *Bildablage) NutzerbildAblegen(nutzerID int64, bild image.Image) (string, error) {
	return b.ablegenUnterNeuemNamen(fmt.Sprintf("nutzer-%d", nutzerID), bild, Nutzerbildkante)
}

// ablegenUnterNeuemNamen schreibt ein Bild unter einem **neuen** Namen aus Vorsatz
// und Zufall — für alles, worauf ein Verweis im Bestand zeigt.
//
// **Ein Rumpf für Profilbild und Abzeichen** — sie unterscheiden sich in
// nichts als dem Vorsatz, und zwei Fassungen liefen beim ersten Sonderfall
// auseinander.
func (b *Bildablage) ablegenUnterNeuemNamen(
	vorsatz string, bild image.Image, kante int,
) (string, error) {
	gepackt, fehler := gepacktAufKante(bild, kante)
	if fehler != nil {
		return "", fehler
	}

	// **Zwölf Stellen Zufall** — 48 Bit; zwei Ablagen desselben Menschen
	// treffen sich darin nicht. `crypto/rand.Read` gibt keinen Fehler zurück
	// (Go 1.27, `crypto/rand/rand.go`: er bricht das Programm ab).
	zufall := make([]byte, 6)
	_, _ = rand.Read(zufall) //nolint:errcheck // siehe darüber

	return b.schreibeBild(fmt.Sprintf("%s-%s.png", vorsatz, hex.EncodeToString(zufall)), gepackt)
}

// ablegenUnter schreibt ein Bild unter einem **festen** Namen — für das
// Tischbild, auf das kein Verweis zeigt: die Datei **ist** dort der Stand,
// und es gibt keine Fassung, die ein Schreiben abweisen könnte.
func (b *Bildablage) ablegenUnter(name string, bild image.Image, kante int) (string, error) {
	gepackt, fehler := gepacktAufKante(bild, kante)
	if fehler != nil {
		return "", fehler
	}

	return b.schreibeBild(name, gepackt)
}

// gepacktAufKante bringt ein Bild in die Schranke und kodiert es als PNG.
//
// **Erst in den Speicher, dann auf die Platte** — derselbe Grund wie bei
// [Bildablage.Ablegen]: ein gescheitertes `Close` nach einem geglückten
// Schreiben ist auf einem Netzlaufwerk möglich und in keinem Test.
func gepacktAufKante(bild image.Image, kante int) ([]byte, error) {
	var gepackt bytes.Buffer
	if err := png.Encode(&gepackt, aufKante(bild, kante)); err != nil {
		return nil, fmt.Errorf("das Bild ließ sich nicht kodieren: %w", err)
	}

	return gepackt.Bytes(), nil
}

// schreibeBild legt die kodierten Bytes unter einem Namen ab.
func (b *Bildablage) schreibeBild(name string, gepackt []byte) (string, error) {
	ort := b.Nutzerverzeichnis()
	if ort == "" {
		return "", fmt.Errorf("%w: für Bilder ist kein Verzeichnis eingerichtet",
			domain.ErrValidation)
	}

	if err := os.MkdirAll(ort, 0o750); err != nil {
		return "", fmt.Errorf("Bildverzeichnis anlegen: %w", err)
	}

	// **Neu mit 0600, sonst mit dem Modus der Datei davor** — der gewöhnliche
	// Vertrag von [schreibeAtomar] (SEC-20, #2133).
	if err := schreibeAtomar(filepath.Join(ort, name), gepackt); err != nil {
		return "", fmt.Errorf("das Bild ablegen: %w", err)
	}

	return name, nil
}

// NutzerbildWegnehmen entfernt **jedes** Bild eines Nutzers — für das
// Löschen des Nutzers.
//
// **Jedes, nicht eines**: seit ein Bild seinen Inhalt im Namen trägt, kann
// neben dem gültigen ein altes liegen, dessen Wegräumen scheiterte — und aus
// Fassungen vor dem 2026-09-23 eines unter dem festen Namen `nutzer-7.png`.
//
// **Ein fehlendes Bild ist kein Fehler**: dass dort schon jemand aufgeräumt
// hat, ist keine schlechte Nachricht — dieselbe Regel wie bei
// [Bildablage.Wegnehmen].
func (b *Bildablage) NutzerbildWegnehmen(nutzerID int64) error {
	return b.alleWegnehmen(fmt.Sprintf("nutzer-%d", nutzerID))
}

// BildWegnehmen entfernt **ein** Bild unter seinem Namen — das alte nach
// einem Wechsel, das neue nach einem abgewiesenen.
//
// **Nur ein Name, kein Pfad**: er kommt aus dem Bestand, und ein Name mit
// Trenner oder `..` wird abgewiesen, statt außerhalb des Ordners zu löschen.
func (b *Bildablage) BildWegnehmen(name string) error {
	if name == "" || name != filepath.Base(name) || strings.HasPrefix(name, ".") {
		return fmt.Errorf("%w: %q ist kein Bildname", domain.ErrValidation, name)
	}

	return b.wegnehmen(name)
}

// alleWegnehmen entfernt das Bild unter dem festen Namen und jedes mit
// Inhalt im Namen.
func (b *Bildablage) alleWegnehmen(vorsatz string) error {
	ort := b.Nutzerverzeichnis()
	if ort == "" {
		return nil
	}

	// **Das Muster kann nicht scheitern**: der Vorsatz besteht aus einem Wort
	// und einer Zahl, und `filepath.Glob` klagt nur über ein kaputtes Muster.
	namen, _ := filepath.Glob(filepath.Join(ort, vorsatz+"-*.png")) //nolint:errcheck // siehe darüber

	var fehler error
	for _, name := range append([]string{vorsatz + ".png"}, namen...) {
		fehler = errors.Join(fehler, b.wegnehmen(filepath.Base(name)))
	}

	return fehler
}

// wegnehmen entfernt eine Datei unter diesem Namen.
func (b *Bildablage) wegnehmen(name string) error {
	ort := b.Nutzerverzeichnis()
	if ort == "" {
		return nil
	}

	if err := os.Remove(filepath.Join(ort, name)); err != nil &&
		!os.IsNotExist(err) {
		return fmt.Errorf("das Bild entfernen: %w", err)
	}

	return nil
}

// ProfilbadgeAblegen schreibt das Abzeichen eines Profils und gibt seinen Namen.
//
// **Neben den Profilbildern und nicht darunter** — Werkstück 9 (#1351),
// Punkt 13. Ein Abzeichen liegt **über** dem Nutzerfoto, rechts unten;
// es gehört einem **Profil** und nicht dem Menschen, und der Name sagt es:
// `profil-7-….png` neben `nutzer-3-….png`. Ein eigenes Verzeichnis wäre eine
// zweite Stelle für dieselbe Frage.
func (b *Bildablage) ProfilbadgeAblegen(profilID int64, bild image.Image) (string, error) {
	return b.ablegenUnterNeuemNamen(fmt.Sprintf("profil-%d", profilID), bild, Nutzerbildkante)
}

// ProfilbadgeWegnehmen entfernt **jedes** Abzeichen eines Profils — für das
// Löschen des Profils; derselbe Grund wie bei [Bildablage.NutzerbildWegnehmen].
func (b *Bildablage) ProfilbadgeWegnehmen(profilID int64) error {
	return b.alleWegnehmen(fmt.Sprintf("profil-%d", profilID))
}

// nutzerbildmass bringt ein Bild in die Schranke.
//
// **Beim Ablegen und nicht beim Lesen.** Ein Profilbild ist ein Abzeichen
// neben einem Namen; wer die Aufnahme in Originalgröße behielte,
// entschlüsselte für jede Anzeige ein Megabyte und schickte es als
// Daten-URI über die Naht.
//
// **Und es wird nie vergrößert.** Ein Bild, das schon hineinpasst, kommt
// unverändert zurück — hochgerechnet wäre es nur unschärfer. Das schließt
// das Bild ohne Ausdehnung ein: es passt hinein, kommt unverändert zurück,
// und `png.Encode` weist es ab, wo die Meldung hingehört.
//
// **Die Kante kommt vom Aufrufer** seit Werkstück G3 von Phase 5.5
// (#1460):
// ein Profilbild ist 256 Punkte, ein Tischbild 1920.
func aufKante(bild image.Image, kante int) image.Image {
	masse := bild.Bounds()
	if masse.Dx() <= kante && masse.Dy() <= kante {
		return bild
	}

	// **Mindestens eine Zeile und eine Spalte.** Ganzzahlig gerechnet ergäbe
	// ein 1000×1 sonst die Höhe null, und ein Bild ohne Fläche ist keines.
	if masse.Dx() >= masse.Dy() {
		return render.ZuschnittAus(bild, masse, kante,
			max(1, masse.Dy()*kante/masse.Dx()))
	}

	return render.ZuschnittAus(bild, masse,
		max(1, masse.Dx()*kante/masse.Dy()), kante)
}

// NutzerbildLesen gibt das abgelegte Bild — `nil` ohne Fehler heißt: keines.
//
// **Dieselbe Zusage wie [Vorschaubildwerk.Kartenbild]**, und aus demselben
// Grund: die meisten Menschen haben kein Profilbild. Ein Fehler wäre eine
// Meldung, die niemand braucht und die jede Liste rot färbte.
//
// **Eine kaputte Datei ist aber etwas anderes als keine.** „Es gibt keines"
// ist eine Auskunft, „da liegt etwas Unlesbares" ein Befund (§8); wer beides
// gleich behandelte, ließe eine beschädigte Ablage unbemerkt.
func (b *Bildablage) NutzerbildLesen(name string) (image.Image, error) {
	ort := b.Nutzerverzeichnis()
	if ort == "" || strings.TrimSpace(name) == "" {
		return nil, nil
	}

	// #nosec G304 -- ein Dateiname aus dem eigenen Bestand; `domain.Nutzer`
	// weist Pfadtrenner und `..` beim Schreiben ab
	datei, err := os.Open(filepath.Join(ort, name))
	if err != nil {
		return nil, nil //nolint:nilerr // nicht da heißt: keines
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // gelesen ist gelesen

	bild, _, decodefehler := msepaket.DekodiereBildAus(datei)
	if decodefehler != nil {
		return nil, msepaket.ZuGrossOder(decodefehler, fmt.Errorf("%w: %s ist kein lesbares Bild: %s",
			domain.ErrValidation, name, decodefehler.Error()))
	}

	return bild, nil
}
