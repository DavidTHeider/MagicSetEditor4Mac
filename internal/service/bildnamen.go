package service

import (
	"context"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/svg"
)

// Bildformat ist die Gestalt, in der ein Kartenbild herauskommt.
type Bildformat string

const (
	FormatPNG Bildformat = "png"
	FormatJPG Bildformat = "jpg"

	// **BMP und TIFF, weil MSE2 sie anbietet** (#826).
	//
	// `src/gui/set/window.cpp:660`, Fassung f9d9356 vom 2021-07-06, führt vier
	// Formate im Speichern-Dialog. Ohne diese beiden ist „mindestens MSE2s
	// Umfang" verfehlt — gemessen, nicht empfunden
	// (#824).
	FormatBMP  Bildformat = "bmp"
	FormatTIFF Bildformat = "tiff"

	// **SVG, und es ist das einzige, das nicht aus Bildpunkten entsteht**
	// (#852).
	//
	// Die vier darüber rastern und packen; SVG entsteht aus dem
	// **Zeichenplan** — ein Rasterbild in ein `<svg>` zu wickeln wäre der Name
	// ohne die Sache. Deshalb geht es in [Bildausfuhr.Schreibe] einen eigenen
	// Weg und nicht durch [PackeBild].
	//
	// **MSE2 kennt es nicht** (`grep -rin svg src/` über die Fassung f9d9356
	// vom 2021-07-06 → 0 Treffer). Es ist eine Zutat dieses Projekts und in
	// `docs/PORTIERUNGSKARTE.md` als **new** gekennzeichnet.
	FormatSVG Bildformat = "svg"
)

// Endung gibt die Dateiendung des Formats; leer, wenn es keines ist.
func (f Bildformat) Endung() string {
	switch f {
	case FormatPNG:
		return ".png"

	case FormatJPG:
		return ".jpg"

	case FormatBMP:
		return ".bmp"

	// **`.tif` und nicht `.tiff`** — so steht es in MSEs Filter, und wer die
	// Dateien nebeneinanderlegt, will denselben Namen.
	case FormatTIFF:
		return ".tif"

	case FormatSVG:
		return ".svg"
	}

	return ""
}

// VorlageStandard ist die Namensvorlage, mit der niemand etwas verliert.
//
// **Die Nummer steht vorn und dreistellig.** Damit ist die alphabetische
// Ordnung des Dateibrowsers die Ordnung des Sets, und zwei gleichnamige Karten
// stoßen gar nicht erst zusammen.
const VorlageStandard = "{nummer}-{name}"

// Bildname ist der Dateiname, unter dem eine Karte herauskommt.
type Bildname struct {
	KarteID int64
	Nummer  int
	Datei   string
}

// Bildausfuhr schreibt Karten als Bilder.
type Bildausfuhr struct {
	wache

	karten   Kartenquelle
	bauer    Planquelle
	rasterer Rasterquelle

	// vektor sind die Mittel für [FormatSVG]. Der Nullwert kann nichts, und
	// eine SVG entsteht dann mit Befunden statt mit Ebenen — sichtbar leer
	// statt still leer.
	vektor svg.Quellen
}

// NeueBildausfuhr baut die Ausfuhr über Bestand, Planbau und Rasterer.
func NeueBildausfuhr(
	karten Kartenquelle, bauer Planquelle, rasterer Rasterquelle,
) *Bildausfuhr {
	return &Bildausfuhr{karten: karten, bauer: bauer, rasterer: rasterer}
}

// MitPruefer setzt den Riegel, der vor dem Schreiben fragt, ob dieser Aufruf
// es darf (0078).
func (b *Bildausfuhr) MitPruefer(pruefer Pruefer) *Bildausfuhr {
	b.pruefer = pruefer

	return b
}

// MitVektor hängt die Mittel für die SVG-Ausgabe an.
//
// **Getrennt vom Bau**, weil ein Aufrufer, der nur PNG schreibt, sie nicht
// braucht — dieselbe Bauform wie `render.Rasterer.MitSchriften`.
//
// **[svg.Quellen.Vergroesserung] wird hier nicht gesetzt.** Sie gehört zum
// einzelnen Lauf und nicht zur Verdrahtung; [Bildausfuhr.Schreibe] rechnet sie
// aus der Zielbreite, so wie für jedes andere Format.
func (b *Bildausfuhr) MitVektor(quellen svg.Quellen) *Bildausfuhr {
	b.vektor = quellen

	return b
}

// Namen gibt für jede Karte eines Sets den Dateinamen nach Vorlage.
//
// # Warum die Namen vorab entstehen und nicht beim Schreiben
//
// **Eindeutigkeit ist eine Aussage über den ganzen Lauf.** Wer den Namen erst
// im Augenblick des Schreibens bildet, kann die Kollision nur gegen das
// Dateisystem prüfen — und dann heißt eine zweite Ausgabe desselben Sets
// `Feuerball-2.png`, obwohl niemand etwas doppelt hat. Hier entstehen alle
// Namen aus einer Liste, und der Lauf kann sie danach der Reihe nach schreiben.
func (b *Bildausfuhr) Namen(
	ctx context.Context, setID int64, vorlage string, format Bildformat,
	verhalten string, auswahl domain.Kartenauswahl,
) ([]Bildname, error) {
	endung := format.Endung()
	if endung == "" {
		return nil, fmt.Errorf("%w: %q ist kein Bildformat", domain.ErrValidation, format)
	}

	if strings.TrimSpace(vorlage) == "" {
		vorlage = VorlageStandard
	}

	alle, err := b.karten.Karten(ctx, setID)
	if err != nil {
		return nil, fmt.Errorf("die Karten lesen: %w", err)
	}

	// **Und eine Teilmenge geht auch** — Befund 2 des Zweitblicks über
	// Phase 3.6. MSE2 stellt die Wahl vor der Bildausgabe genauso
	// (`SetWindow::onFileExportImages`, `src/gui/set/window.cpp`, Zeilen
	// 668–673, Fassung f9d9356 vom 2021-07-06); hier fehlte sie, während
	// Druckbogen und Exportvorlage sie schon hatten.
	//
	// **Vor der Namensvergabe.** MSEs `used` zählt die schon vergebenen Namen
	// mit; wer erst alle Namen bildete und dann filterte, bekäme Lücken in der
	// Nummerierung — `Karte-1`, `Karte-3`.
	if fehler := auswahl.Pruefe(alle); fehler != nil {
		return nil, fehler
	}

	karten := auswahl.Anwenden(alle)

	namen := make([]Bildname, 0, len(karten))

	// **Was dieser Durchgang schon vergeben hat** — MSEs `used`
	// (`src/data/format/image.cpp`, Zeile 78, Fassung f9d9356 vom
	// 2021-07-06). Gelesen wird es nur von `nummerierenneu`; gefüllt wird es
	// immer, wie bei MSE.
	//
	// **Kleingeschrieben als Schlüssel**, und das ist eine benannte Abweichung
	// (`adapted`): MSEs `std::set<String>` vergleicht genau. macOS
	// unterscheidet in der Vorgabe **nicht** zwischen `Feuerball` und
	// `feuerball` — ein Lauf, der beide schriebe, verlöre die erste Karte auf
	// genau der Hälfte aller Rechner.
	gelaufen := map[string]bool{}

	for _, karte := range karten {
		stamm := einsetzen(vorlage, karte)
		if stamm == "" {
			// **Nicht `.png`.** Ein leerer Name ist unter macOS erlaubt, aber
			// niemand findet die Datei wieder — und die zweite namenlose Karte
			// überschriebe die erste.
			stamm = fmt.Sprintf("karte-%03d", karte.Number)
		}

		// **Nur der laufende Durchgang wird hier aufgelöst** (#491).
		//
		// `nummerieren` und `altbehalten` fragen die **Platte**, und die steht
		// erst beim Schreiben fest: das Zielverzeichnis wählt jemand danach.
		// Sie werden deshalb in [Bildausfuhr.Schreibe] entschieden — dieselbe
		// Funktion, dieselbe Regel, zwei Aufrufstellen.
		datei := kuerze(stamm) + endung
		datei, _ = Namenskonflikt("", datei, verhalten, gelaufen)
		gelaufen[strings.ToLower(datei)] = true

		namen = append(namen, Bildname{
			KarteID: karte.ID,
			Nummer:  karte.Number,
			Datei:   datei,
		})
	}

	return namen, nil
}

// einsetzen ersetzt die Platzhalter der Vorlage durch die Werte der Karte.
//
// **Der Wert wird vor dem Einsetzen gesäubert, nicht danach.** Ein Kartenname
// `../../oben` in der Vorlage `{nummer}-{name}` ergäbe sonst `001-..-..-oben`:
// das Ergebnis führt zwar nicht mehr aus dem Verzeichnis heraus, trägt den
// Versuch aber im Namen. Und ein Name mit führendem Punkt bliebe unsichtbar,
// weil der Punkt dann in der Mitte steht.
func einsetzen(vorlage string, karte *domain.CardRow) string {
	name, _ := karte.Value("name")

	gesetzt := strings.NewReplacer(
		"{nummer}", fmt.Sprintf("%03d", karte.Number),
		"{name}", saeubere(name),
	).Replace(vorlage)

	// **Und die Vorlage selbst ist auch eine Eingabe.** Sie kommt aus einem
	// Feld der Oberfläche; wer dort `../` tippt, hat denselben Weg gefunden.
	return saeubere(gesetzt)
}

// grenzeName ist die Byte-Länge, auf die ein Namensstamm gekürzt wird.
//
// **macOS lässt 255 Bytes je Dateiname zu.** Der Rest bis dahin gehört der
// Endung und dem Zusatz, den `eindeutig` anhängt — beides entsteht erst nach
// dem Kürzen, und beides muss noch hineinpassen.
const grenzeName = 200

// kuerze schneidet einen Namensstamm auf `grenzeName` Bytes.
//
// **Nach Bytes gemessen, an Zeichen geschnitten.** Ein Schnitt mitten in einer
// UTF-8-Folge ergäbe ein Ersatzzeichen, und das ist im Dateinamen nicht mehr zu
// tilgen. Ein zu langer Name ließe das Schreiben scheitern — und zwar erst nach
// dem Rastern, also nach der teuren Hälfte.
func kuerze(stamm string) string {
	if len(stamm) <= grenzeName {
		return stamm
	}

	gekuerzt := stamm[:grenzeName]
	for !utf8.ValidString(gekuerzt) {
		gekuerzt = gekuerzt[:len(gekuerzt)-1]
	}

	return strings.TrimRight(gekuerzt, " .-")
}

// saeubere macht aus einer Eingabe einen Dateinamen.
//
// **Ein Kartenname ist eine Eingabe, kein Dateiname** (§8): er kommt aus einem
// fremden Paket und darf alles enthalten — Schrägstriche, Doppelpunkte,
// Steuerzeichen, `..`. Wer ihn ungeprüft zusammensetzt, hat einen Schreibweg,
// der aus dem gewählten Verzeichnis herausführt.
func saeubere(roh string) string {
	var gebaut strings.Builder

	for _, zeichen := range roh {
		switch {
		case zeichen == '/' || zeichen == '\\' || zeichen == ':':
			gebaut.WriteRune('-')

		case unicode.IsControl(zeichen):
			gebaut.WriteRune('-')

		default:
			gebaut.WriteRune(zeichen)
		}
	}

	// Führende Punkte fallen weg: `..` führte ein Verzeichnis höher, und eine
	// Datei, die mit einem Punkt beginnt, ist im Finder unsichtbar.
	sauber := strings.Trim(gebaut.String(), " .-")

	// Mehrere Trennstriche hintereinander sind das Ergebnis des Ersetzens, nicht
	// der Absicht.
	for strings.Contains(sauber, "--") {
		sauber = strings.ReplaceAll(sauber, "--", "-")
	}

	return sauber
}
