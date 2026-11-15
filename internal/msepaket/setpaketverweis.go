package msepaket

import (
	"fmt"
	"image"
	"path"
	"strings"
	"unicode"
)

// setpaketschema steht vor jedem Verweis auf ein Paket an einem Ort.
//
// **Ausgeschrieben und nicht ein Zeichen**, weil dieser Verweis in Fehlermeldungen
// landet: „setpaket:/Users/…/Mein Set.mse-set!image1 entschlüsseln: …" sagt
// jemandem, wo er nachsehen muss.
const setpaketschema = "setpaket:"

// setpakettrenner trennt den Ort von der Datei.
//
// **`!` wie in einem Zip-Verweis**, und getrennt wird am **letzten**: ein
// Verzeichnisname darf ein Ausrufezeichen tragen, ein Dateiname in einem
// MSE-Set trägt keines (`image1`, `image2`, …).
const setpakettrenner = "!"

// SetpaketVerweis benennt eine Datei in einem Paket, dessen Ort feststeht.
//
// # Warum es diese zweite Form gibt
//
// Ein Paketpfad (`/magic-m15.mse-style/card.jpg`) nennt ein Paket **beim
// Namen**; wo es liegt, entscheiden die Paketwurzeln. Für ein Set geht das
// nicht: es liegt dort, wo jemand es abgelegt hat, und sein Ort steht in
// [domain.SetHead.Origin].
//
// MSE löst dieselbe Frage über das **lokale** Paket:
//
//	auto image_file = getLocalPackage().openIn(value().filename);
//
// `ImageValueViewer::draw`, `src/render/value/image.cpp`, Zeilen 30–33,
// Fassung f9d9356 vom 2021-07-06.
//
// # Warum der ganze Ort mitreist und nicht der Paketname
//
// Der Bildspeicher merkt sich nach Verweis. Zwei Sets nennen ihre Illustration
// beide `image1` — ein Verweis nach Namen zeigte auf der einen Karte das Bild
// der anderen (#380).
//
// # Warum nicht das Elternverzeichnis eine Wurzel wird
//
// Weil damit **jedes** Nachbarpaket auflösbar würde: ein Set neben einem
// fremden Verzeichnis machte dessen Dateien für jede Vorlage erreichbar.
func SetpaketVerweis(ort, datei string) string {
	return setpaketschema + ort + setpakettrenner + datei
}

// ZerlegeSetpaketverweis liest einen Verweis auf ein Paket an einem Ort zurück.
//
// `ist` sagt, ob es überhaupt einer ist — ein gewöhnlicher Paketpfad geht
// unberührt weiter.
//
// **Geschnitten wird nur links.** Leerraum am Ende gehört zum Dateinamen: der
// Zerteiler schneidet einen Wert nur links, wie MSE2 (`trim_left`,
// `src/util/io/reader.cpp` Zeile 224, Fassung f9d9356 vom 2021-07-06), und
// `image: image1 ` nennt die Datei `image1 `. Bis QA-06b (#1967) öffnete der
// Verweis `image1` — gefunden von `FuzzPaketpfad`.
func ZerlegeSetpaketverweis(verweis string) (ort, datei string, ist bool) {
	rumpf, hatSchema := strings.CutPrefix(strings.TrimLeftFunc(verweis, unicode.IsSpace), setpaketschema)
	if !hatSchema {
		return "", "", false
	}

	trenner := strings.LastIndex(rumpf, setpakettrenner)
	if trenner < 0 {
		return "", "", true
	}

	return rumpf[:trenner], rumpf[trenner+len(setpakettrenner):], true
}

// pruefeSetpaketverweis weist zurück, was nicht in ein Paket führt.
//
// **`.` und `..` am rohen Namen**, aus demselben Grund wie bei
// [ZerlegePaketpfad]: wegkanonisieren machte aus einem Ausbruchsversuch einen
// gültigen Verweis auf etwas anderes.
func pruefeSetpaketverweis(verweis, ort, datei string) error {
	if strings.TrimSpace(ort) == "" {
		return fmt.Errorf("%w: %q nennt keinen Ort", ErrPaketpfad, verweis)
	}

	if strings.TrimSpace(datei) == "" {
		return fmt.Errorf("%w: %q nennt keine Datei", ErrPaketpfad, verweis)
	}

	for teil := range strings.SplitSeq(datei, "/") {
		if teil == "." || teil == ".." {
			return fmt.Errorf("%w: %q enthält %q", ErrPaketpfad, verweis, teil)
		}
	}

	return nil
}

// ausSetpaket liest ein Bild aus einem Paket an seinem Ort.
//
// **Ohne die Wurzeln.** Der Ort steht im Verweis; dort zu suchen, wo das Bild
// nicht liegt, brächte keinen Treffer und könnte einen falschen bringen.
func (b *Bildquelle) ausSetpaket(verweis, ort, datei string) (image.Image, error) {
	if err := pruefeSetpaketverweis(verweis, ort, datei); err != nil {
		return nil, err
	}

	if gemerkt, da := b.nachschlagen(verweis); da {
		return gemerkt, nil
	}

	rohdaten, err := paketdatei(ort, path.Clean(datei))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", verweis, err)
	}

	bild, _, err := DekodiereBild(rohdaten)
	if err != nil {
		return nil, fmt.Errorf("%s entschlüsseln: %w", verweis, err)
	}

	b.merken(verweis, bild)

	return bild, nil
}
