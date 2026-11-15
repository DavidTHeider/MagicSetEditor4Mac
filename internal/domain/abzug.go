// **Der Abzug eines Tisches** — PLAY-05.AK-03 und AK-04
// (#1713).
//
// # Warum das hier steht und nicht an der Naht
//
// **Weil es eine Prüfung ist und keine Übersetzung** (§2). Was aus dem
// Fenster kommt, ist eine Zeichenkette, die ein Bild zu sein **behauptet**;
// ob sie eines ist, entscheidet die Signatur und nicht der Vorsatz. Das ist
// dieselbe Haltung wie gegenüber einem MSE-Paket von 2007: fremde Daten sind
// der Normalfall (§8).
//
// # Und warum die Signatur geprüft wird
//
// **Weil dieser Weg eine Datei schreibt.** Wer eine `.png` anlegt, in der
// etwas anderes steht, hat kein Bild ausgegeben, sondern eine Falle gestellt
// — und der Vorsatz eines Data-URLs ist eine Angabe des Absenders.

package domain

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
)

// Abzugsvorsatz ist, womit ein PNG-Data-URL anfängt.
const Abzugsvorsatz = "data:image/png;base64,"

// Abzugsgrenze ist, wie viele Bytes ein Abzug höchstens trägt.
//
// **64 MiB** — ein Abzug von 50 Millionen Bildpunkten wiegt als PNG deutlich
// weniger, und was darüber liegt, ist kein Tischbild mehr. Die Zahl ist
// gewählt und nicht gemessen; sie steht hier, damit eine fremde Zeichenkette
// nicht den Arbeitsspeicher bestimmt (derselbe Satz wie bei der Längenangabe
// in `internal/leitung`).
const Abzugsgrenze = 64 << 20

// pngsignatur sind die acht Bytes, mit denen jede PNG-Datei anfängt.
var pngsignatur = []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}

// DeuteAbzug liest die Bytes eines PNG-Abzugs aus einem Data-URL.
func DeuteAbzug(datenURL string) ([]byte, error) {
	if datenURL == "" {
		return nil, fmt.Errorf("%w: der Abzug ist leer", ErrValidation)
	}

	if !strings.HasPrefix(datenURL, Abzugsvorsatz) {
		return nil, fmt.Errorf("%w: der Abzug ist kein PNG", ErrValidation)
	}

	kodiert := strings.TrimPrefix(datenURL, Abzugsvorsatz)

	// **Die Länge wird geprüft, bevor zugeteilt wird.** Base64 trägt drei
	// Bytes in vier Zeichen; wer erst dekodierte und dann maß, ließe eine
	// fremde Zeichenkette den Arbeitsspeicher bestimmen — derselbe Satz wie
	// bei der Längenangabe in `internal/leitung`.
	//
	// **Und danach wird nicht noch einmal gemessen.** `len(kodiert)/4*3` ist
	// die **Obergrenze** der dekodierten Länge: vier Zeichen werden zu
	// höchstens drei Bytes. Wer das Ergebnis noch einmal gegen dieselbe
	// Grenze hielte, schriebe einen Zweig hin, den kein Aufruf erreicht (§3).
	if len(kodiert)/4*3 > Abzugsgrenze {
		return nil, fmt.Errorf("%w: der Abzug ist zu groß", ErrValidation)
	}

	rohes, fehler := base64.StdEncoding.DecodeString(kodiert)
	if fehler != nil {
		return nil, MitGrund(ErrValidation, fehler,
			"der Abzug ist nicht lesbar")
	}

	if !bytes.HasPrefix(rohes, pngsignatur) {
		return nil, fmt.Errorf("%w: der Abzug ist kein PNG", ErrValidation)
	}

	return rohes, nil
}
