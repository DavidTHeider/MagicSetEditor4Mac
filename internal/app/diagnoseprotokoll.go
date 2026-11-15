package app

// **Das Ende des Protokolls in der Diagnose** — abgetrennt von `diagnose.go`,
// seit das begrenzte Lesen die Datei über die Zeilengrenze aus §2 hob. Der
// Schnitt folgt der Frage: dort, was eine Diagnose trägt; hier, wie viel vom
// Protokoll dazugehört.

import (
	"bytes"
	"errors"
	"io"
	"os"
)

// protokollfenster ist, wie viel vom Ende des Protokolls die Diagnose liest.
//
// **Sechzehn Mebibyte, doppelt so viel, wie ein Diagnosepaket tragen darf**
// (`domain.Diagnosepaketgrenze`): zweihundert Zeilen passen auch dann hinein,
// wenn jede lang ist, und ein Protokoll, das das Paket sprengt, sprengt es
// weiterhin — die Grenze des Pakets bleibt die, die entscheidet
// (`TestEinZuGrossesPaketWirdAbgewiesen`). Was dieses Fenster hält, ist der
// Speicher, nicht der Inhalt.
const protokollfenster int64 = 16 << 20

// protokollschwanz liest das Ende des Protokolls — höchstens
// [protokollfenster] Byte.
//
// **Nur das Ende** — Befund 3 des Nachlaufs über Phase 6.1. Das Protokoll
// schneidet sich beim **Start** bei 4 MiB ab (`Protokoll`); während einer
// Sitzung wächst es ohne Grenze, und hier stand ein `os.ReadFile`, das die
// ganze Datei las, um zweihundert Zeilen zu behalten. Die erste Zeile des
// Fensters ist meist angeschnitten und fällt weg.
func protokollschwanz(pfad string) ([]byte, error) {
	// #nosec G304 -- der Pfad kommt aus [Protokollpfad], nicht von einem Aufrufer
	datei, err := os.Open(pfad)
	if err != nil {
		return nil, err
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // nur gelesen

	// **Wo das Fenster anfängt und wie groß es ist** — und wer das nicht
	// sagen kann, liest vom Anfang ein volles Fenster: begrenzt ist es so oder
	// so. `Seek` gibt bei einem Fehler die Stelle null zurück, und von dort
	// liest dann auch der Leser.
	anfang, groesse := int64(0), protokollfenster
	if angaben, statfehler := datei.Stat(); statfehler == nil {
		anfang = max(0, angaben.Size()-protokollfenster)
		groesse = min(protokollfenster, angaben.Size())
	}

	anfang, _ = datei.Seek(anfang, io.SeekStart) //nolint:errcheck // siehe darüber

	// **Ein Puffer, einmal angelegt** — nicht `io.ReadAll`, das verdoppelt
	// und dabei das Doppelte zuteilt. Gemessen unter `-race`: 69 MiB für ein
	// Fenster von 16 MiB, mit Umwandeln, Zerteilen und Zusammenfügen dahinter.
	// Wird die Datei zwischen Vermessen und Lesen kürzer, zählt, was da ist.
	roh := make([]byte, groesse)

	gelesen, err := io.ReadFull(datei, roh)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, err
	}

	roh = roh[:gelesen]

	if anfang > 0 {
		if schnitt := bytes.IndexByte(roh, '\n'); schnitt >= 0 {
			roh = roh[schnitt+1:]
		}
	}

	return roh, nil
}

// protokollende liest die letzten [Protokollzeilen] Zeilen.
//
// **Ein fehlendes Protokoll ist kein Fehler**, sondern eine Auskunft: eine
// frisch eingerichtete Anwendung hat keines, und wer dann eine Diagnose
// braucht, braucht sie besonders.
func protokollende(pfad string, vorher error, texte Dialogtexte) string {
	if vorher != nil {
		return texte.DiagnoseKeinProtokoll + ": " + vorher.Error()
	}

	roh, err := protokollschwanz(pfad)
	if err != nil {
		return texte.DiagnoseKeinProtokoll + ": " + err.Error()
	}

	// **Ohne den freien Wert einer Panik** (#2399, Vermerk 0193) — die
	// Diagnose und das Diagnosepaket können den Rechner verlassen.
	return paniktexteSchwaerzen(string(letzteZeilen(roh, Protokollzeilen))) + "\n"
}

// letzteZeilen gibt die letzten `anzahl` Zeilen — ohne abschließende
// Zeilenwechsel und **ohne Kopie**: gezählt wird vom Ende her, und erst das
// Ergebnis wird zur Zeichenkette.
func letzteZeilen(roh []byte, anzahl int) []byte {
	roh = bytes.TrimRight(roh, "\n")

	anfang, ende := 0, len(roh)

	for range anzahl {
		stelle := bytes.LastIndexByte(roh[:ende], '\n')
		if stelle < 0 {
			return roh
		}

		anfang, ende = stelle+1, stelle
	}

	return roh[anfang:]
}
