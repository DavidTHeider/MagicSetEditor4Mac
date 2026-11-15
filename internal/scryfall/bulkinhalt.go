package scryfall

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// bulkobjekt sagt, was in der ersten Zeile einer Sammeldatei dieser Art steht.
//
// **Gemessen am 2026-08-24** an der abgelegten `oracle_cards.jsonl.gz`
// (24.531.600 Bytes): die erste Zeile ist ein JSON-Objekt mit
// `"object":"card"` und 69 Feldern.
//
// **Nur die eine Art steht hier, und zwar weil nur sie nachgemessen ist.**
// Scryfall bietet daneben `default_cards`, `unique_artwork`, `all_cards` und
// `rulings` an; welchen Wert deren erste Zeile trägt, ist hier **nicht**
// nachgelesen worden, und ein geratener wäre eine Behauptung statt einer
// Prüfung (§10). Dieses Programm holt `oracle_cards` und sonst nichts
// (`app.bulkart`).
//
// Eine Art ohne Eintrag wird deshalb **schwächer** geprüft: die erste Zeile
// muss ein JSON-Objekt mit irgendeinem `object` sein. Wer eine Art aufnimmt,
// misst ihre erste Zeile und trägt sie hier ein.
var bulkobjekt = map[string]string{"oracle_cards": "card"}

// bulkmindestanteil ist der Anteil der angekündigten Größe, unter dem eine
// Sammeldatei als abgeschnitten gilt.
//
// **Die Hälfte, und nur nach unten.** `compressed_size` stammt aus dem
// Verzeichnis, und das darf bis zu einem Tag alt sein; Scryfall veröffentlicht
// täglich neu, die Datei ist zwischen Verzeichnis und Abruf also legitim
// gewachsen. Eine Prüfung nach oben wiese den Normalfall ab.
const bulkmindestanteil = 2

// pruefeSammeldatei sieht nach, ob die geholten Bytes eine Sammeldatei sind.
//
// # Warum überhaupt
//
// Der Zweischritt beim Ablegen schließt das Fenster einer halb
// **geschriebenen** Datei. Eine vollständig übertragene, inhaltlich falsche
// geht durch ihn hindurch — und ersetzt damit den letzten guten Stand, aus
// dem der Betrieb ohne Netz lebt
// (#1027).
// Das ist die Umkehrung von §16, Regel 5.
//
// # Was sie hält
//
//   - Eine Antwort, die kein gzip ist — jede HTML-Fehlerseite.
//   - Ein gzip, dessen erste Zeile kein JSON-Objekt ist.
//   - Ein JSON-Objekt, das für diese Art das falsche ist — etwa ein
//     Fehlerobjekt von Scryfall selbst.
//   - Eine Datei ohne eine einzige Zeile.
//   - Eine Datei, die weniger als die Hälfte der angekündigten Größe wiegt.
//
// # Was sie nicht hält
//
// **Eine Datei mit gültiger erster Zeile und Unsinn ab der zweiten.** Sie
// ganz zu lesen kostete bei 24 MB gepackt Sekunden bei jedem Abruf, und der
// Fall, gegen den es geht — eine Fehlerseite, ein Abbruch —, trifft die erste
// Zeile immer. Der schlechtere Tausch wäre, jeden Abruf zu verlangsamen.
//
// **Und sie sagt nichts über die Vollständigkeit.** Eine Sammeldatei mit drei
// Karten statt 30.000 kommt durch, solange sie über der halben angekündigten
// Größe liegt.
func pruefeSammeldatei(rumpf []byte, eintrag Bulkeintrag) error {
	if eintrag.Groesse > 0 && int64(len(rumpf))*bulkmindestanteil < eintrag.Groesse {
		return fmt.Errorf(
			"%w: die Sammeldatei wiegt %d Bytes, angekündigt waren %d",
			domain.ErrValidation, len(rumpf), eintrag.Groesse)
	}

	entpackt, gzfehler := gzip.NewReader(bytes.NewReader(rumpf))
	if gzfehler != nil {
		return fmt.Errorf("%w: die Antwort ist kein gzip: %s",
			domain.ErrValidation, gzfehler.Error())
	}

	defer func() { _ = entpackt.Close() }() //nolint:errcheck // gelesen ist gelesen

	leser := bufio.NewScanner(entpackt)
	leser.Buffer(make([]byte, 0, 64<<10), zeilengrenze)

	for leser.Scan() {
		zeile := bytes.TrimSpace(leser.Bytes())
		if len(zeile) == 0 {
			continue
		}

		return pruefeErsteZeile(zeile, eintrag.Art)
	}

	if fehler := leser.Err(); fehler != nil {
		return fmt.Errorf("%w: die Sammeldatei ließ sich nicht lesen: %s",
			domain.ErrValidation, fehler.Error())
	}

	return fmt.Errorf("%w: die Sammeldatei enthält keine Zeile", domain.ErrValidation)
}

// pruefeErsteZeile sieht nach, ob die erste Zeile das erwartete Objekt ist.
func pruefeErsteZeile(zeile []byte, art string) error {
	var kopf struct {
		Objekt string `json:"object"`
	}

	if fehler := json.Unmarshal(zeile, &kopf); fehler != nil {
		return fmt.Errorf("%w: die erste Zeile der Sammeldatei ist kein JSON: %s",
			domain.ErrValidation, fehler.Error())
	}

	erwartet, gemessen := bulkobjekt[art]

	switch {
	case kopf.Objekt == "":
		return fmt.Errorf(
			"%w: die erste Zeile der Sammeldatei nennt kein `object`",
			domain.ErrValidation)

	case gemessen && kopf.Objekt != erwartet:
		return fmt.Errorf(
			"%w: die erste Zeile der Sammeldatei %q ist ein %q und kein %q",
			domain.ErrValidation, art, kopf.Objekt, erwartet)
	}

	return nil
}
