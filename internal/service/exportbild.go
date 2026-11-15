package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Bildschreiber ist die Naht zu `write_image_file` — der Funktion, die Dateien
// schreibt.
// (#216).
//
// # Erschlossen, nicht erfunden
//
// Fünf der neunzehn Exportvorlagen von `M15-Magic-Pack` rufen sie, in zwei
// Gestalten:
//
//	write_image_file(card, file: "{card.name}.jpg")
//	write_image_file(card, directory: "cardimages", file: "…")
//
// Sie gibt den **Namen** der geschriebenen Datei zurück; die Vorlagen hängen
// ihn in ein `<img src="…">` oder in eine Zeile ihrer Ausgabe.
//
// # Warum sie hier steht und nicht in `msescript`
//
// Weil sie eine Karte rastert und auf die Platte schreibt. `msescript` kennt
// nichts aus diesem Projekt (§2) und darf kein Dateisystem anfassen — der
// Wächter `internal/schichten/wirkung_test.go` hält das fest. Die Funktion
// wird deshalb von hier **in die Umgebung gelegt**, nicht in die Bibliothek
// eingetragen.
//
// # Die Ausbruchsgrenze
//
// Sie liegt nicht hier, sondern in [Bildausfuhr.Schreibe] — dort, wo der Pfad
// entsteht. Diese Datei reicht `verzeichnis` und `datei` weiter und prüft
// selbst nichts: eine zweite Prüfung daneben wäre eine zweite Wahrheit, und
// die läuft weg. Was hier steht, ist die Zusage, dass der Aufrufer **keinen**
// Pfad wählen kann, den die Vorlage bestimmt hat — das Zielverzeichnis kommt
// von außen und wird der Vorlage nur untergeordnet.
type Bildschreiber interface {
	Schreibe(ctx context.Context, karteID int64, verzeichnis, datei string,
		format Bildformat, zielbreite int) (*Bildschreibung, error)
}

// bildschreiber baut die Funktion, die in der Umgebung steht.
//
// **Zwei Gestalten, eine Funktion.**
//
//	ScriptObject<CardP>* card = dynamic_cast<ScriptObject<CardP>*>(input.get());
//	if (card) { image = conform_image(export_bitmap(…), options); }
//	else      { image = input->toImage()->generateConform(options); }
//
// `src/script/functions/export.cpp`, Zeilen 414–421, Fassung f9d9356 vom
// 2021-07-06. Bis zum 2026-08-21 kannte diese Umsetzung nur die erste; die
// zweite brauchen `magic-spoiler` und `magic-symbol-extractor` für das
// Setsymbol (#660).
func (l *ausfuhrlauf) bildschreiber() msescript.Eingebaut {
	return func(ctx context.Context, ruf msescript.Ruf) (msescript.Wert, error) {
		if len(ruf.Stellung) != 1 {
			return msescript.Nichts, fmt.Errorf(
				"%w: write_image_file braucht genau eine Karte oder ein Bild",
				domain.ErrValidation)
		}

		datei, benannt := ruf.Benannter("file")
		if !benannt {
			return msescript.Nichts, fmt.Errorf(
				"%w: write_image_file braucht ein Argument `file:`", domain.ErrValidation)
		}

		name := datei.AlsText()

		// **Ein Name, der schon geschrieben ist, wird nicht zweimal
		// geschrieben.**
		//
		//	if (ei.exported_images.find(file) != ei.exported_images.end()) {
		//	  SCRIPT_RETURN(file); // already written an image with this name
		//	}
		//
		// Zeilen 406–409. `magic-spoiler` ruft die Funktion **je Karte** mit
		// demselben Namen; über ein Set mit 300 Karten wäre dasselbe Symbol
		// sonst dreihundertmal gerastert.
		if _, schon := l.gelegt[name]; schon {
			return msescript.Text(name), nil
		}

		// **Erst die Karte, dann das Bild** — dieselbe Reihenfolge wie MSEs
		// `dynamic_cast<ScriptObject<CardP>*>`. Eine Karte trägt ihre Kennung
		// als Marke; alles andere geht durch `toImage()`, und das nimmt auch
		// eine Zeichenkette als Dateibild
		// (`ScriptString::toImage`, `src/script/value.cpp`, Zeilen 369–375,
		// Fassung f9d9356 vom 2021-07-06).
		if kennung, ist := kartenkennung(ruf.Stellung[0]); ist {
			return l.karteSchreiben(ctx, ruf, name, kennung)
		}

		geschrieben, fehler := l.bildwertSchreiben(ruf.Stellung[0], name,
			ganzeAngabe(ruf, "width"), ganzeAngabe(ruf, "height"))
		if fehler != nil {
			return msescript.Nichts, fehler
		}

		return msescript.Text(geschrieben), nil
	}
}

// kartenkennung liest die Marke eines Kartenwerts.
//
// **Die Marke sagt, welche Karte gemeint ist.** Der Grund, warum es sie gibt,
// steht am Feld in `msescript/wert.go`: eine Karte im Skript ist eine Abbildung
// aus Feldname und Text und trägt sonst keine Identität.
func kartenkennung(wert msescript.Wert) (int64, bool) {
	kennung, fehler := strconv.ParseInt(wert.Marke, 10, 64)

	return kennung, fehler == nil
}

// karteSchreiben ist die erste Gestalt — eine Karte wird gerastert.
func (l *ausfuhrlauf) karteSchreiben(
	ctx context.Context, ruf msescript.Ruf, name string, kennung int64,
) (msescript.Wert, error) {
	mittel := l.mittel

	if mittel.Bilder == nil {
		return msescript.Nichts, fmt.Errorf(
			"%w: write_image_file kann hier keine Karten schreiben", domain.ErrValidation)
	}

	ziel := mittel.Bildziel

	// **`directory:` liegt *unter* dem Zielverzeichnis, nicht daneben.**
	// Die Vorlage bestimmt einen Namen, nicht einen Ort.
	if unter, gesetzt := ruf.Benannter("directory"); gesetzt {
		ziel = unterverzeichnis(ziel, unter.AlsText())
	}

	// **Vor jedem Bild die Frage, ob es weitergeht** (FMT-03). Wie viele es
	// werden, weiß nur das Skript; gezählt wird deshalb ohne Gesamtzahl.
	if fehler := Schritt(ctx, domain.Fortschritt{
		Erledigt: l.abgelegt, Einheit: domain.EinheitBild,
	}); fehler != nil {
		return msescript.Nichts, fehler
	}

	geschrieben, fehler := mittel.Bilder.Schreibe(ctx, kennung, ziel,
		name, formatAusName(name), mittel.Bildbreite)
	if fehler != nil {
		return msescript.Nichts, fehler
	}

	// **Auch eine geschriebene Karte kommt in die Merkliste** — MSE trägt sie
	// dort ein, gleich welcher Gestalt (Zeile 425).
	l.gelegt[name] = [2]int{geschrieben.Breite, geschrieben.Hoehe}
	l.abgelegt++

	return msescript.Text(basisname(geschrieben.Pfad)), nil
}

// unterverzeichnis hängt einen von der Vorlage genannten Ordner an das Ziel.
//
// **Angenommen wird nur ein einzelner, harmloser Name; alles andere wird
// verworfen.** Nicht bereinigt — verworfen. Der erste Anlauf strich die
// Trennzeichen heraus und ließ aus `../..` ein Verzeichnis `....` übrig; das
// bricht zwar nicht aus, aber es ist ein Name, den niemand gemeint hat, und
// eine Regel, die aus einer Eingabe etwas anderes macht, muss man jedes Mal
// nachrechnen. Gefunden hat es die Gegenprobe in
// `TestEineVorlageKannNichtAusDemZielverzeichnisAusbrechen`.
//
// Die Vorlage darf einen Ordner **benennen**, nicht einen Ort **wählen**.
func unterverzeichnis(ziel, genannt string) string {
	name := strings.TrimSpace(genannt)

	// Ein Name und kein Pfad: kein Trenner, kein Null-Byte, kein `..`, und
	// nicht `.` — was übrig bleibt, kann das Ziel nicht verlassen.
	if name == "" || name == "." || strings.Contains(name, "..") ||
		strings.ContainsAny(name, "/\\\x00") || name != filepath.Base(name) {
		return ziel
	}

	return filepath.Join(ziel, name)
}

// formatAusName liest das Format an der Endung — `.jpg` heißt JPG, sonst PNG.
//
// **Die Vorlagen schreiben beides**: `magic-mws` legt `.jpg` an,
// `magic-planesculptors` `.png`. Die Endung ist die einzige Angabe, die sie
// dazu machen.
func formatAusName(datei string) Bildformat {
	// **Beide Schreibweisen je Format**, weil beide vorkommen: MSE schlägt
	// `.jpg` und `.tif` vor, andere Programme schreiben `.jpeg` und `.tiff`
	// (#826).
	nach := map[string]Bildformat{
		".jpg":  FormatJPG,
		".jpeg": FormatJPG,
		".bmp":  FormatBMP,
		".tif":  FormatTIFF,
		".tiff": FormatTIFF,
	}

	if format, bekannt := nach[strings.ToLower(filepath.Ext(datei))]; bekannt {
		return format
	}

	return FormatPNG
}

// basisname gibt den Dateinamen ohne Verzeichnis.
//
// **Das ist es, was die Vorlage zurückbekommt.** Sie hängt ihn in ein
// `<img src="…">`; ein absoluter Pfad des Rechners, auf dem exportiert wurde,
// gehörte dort nicht hinein.
func basisname(pfad string) string { return filepath.Base(pfad) }
