package service

// **Das Verzeichnis eines ZIP, gelesen bis zur Grenze und nicht weiter** —
// Befund 1 des Zweitblicks über UX-13
// (#1685).
//
// # Warum nicht `archive/zip`
//
// **Weil es das ganze Verzeichnis liest, bevor es etwas zurückgibt.**
// Nachgelesen in Go 1.27.1, `src/archive/zip/reader.go`, `(*Reader).init`:
// die Schleife liest Kopf um Kopf, bis einer nicht passt, und legt je Kopf
// eine Struktur an; die Zahl aus dem Verzeichnisende bestimmt vorher sogar die
// Kapazität der Liste. Eine Grenze dahinter begrenzte die **Anzeige**, nicht
// den Speicher — ein Archiv aus Millionen leerer Einträge hielte das Fenster
// an.
//
// # Was von dort übernommen ist
//
// **Das Lesen des Verzeichnisendes und eines Kopfes**, in Verhalten und nicht
// in Zeilen: die Suche der Endsignatur von hinten samt Prüfung der
// Kommentarlänge (`readDirectoryEnd`, `findSignatureInBlock`), zip64 über den
// Verweis davor (`findDirectory64End`, `readDirectory64End`), der Vorsatz vor
// dem Archiv über `baseOffset` samt dem Rückfall auf null, die Größe aus dem
// zip64-Feld, wenn die alte voll ist (`readDirectoryHeader`), und die
// Unix-Art aus den äußeren Merkmalen (`(*FileHeader).Mode`). **Und der
// Vergleich am Ende**: ein kaputter Kopf, bevor die genannte Zahl erreicht
// ist, macht das Archiv unlesbar — dort über 16 Bit, hier auch.
//
// # Wo er bewusst abweicht
//
// **Er hört an der Grenze auf** — das ist sein Zweck; was dahinter kaputt
// ist, sieht er nicht. **Er liest nur die ausgepackte Größe** aus dem
// zip64-Feld und prüft die gepackte und den Versatz nicht, weil die Ansicht
// sie nicht zeigt. **Er dekodiert Namen nicht** (Go unterscheidet UTF-8 und
// CP437 für `NonUTF8`); ein Name erscheint, wie er in den Bytes steht, und
// wird an ungültigen Stellen ersetzt. **Und er fragt `zipinsecurepath` nicht**:
// ein Name mit `..` oder `/` am Anfang wird gezählt und markiert, nicht
// abgewiesen — die Ansicht soll ihn zeigen.

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Das Verzeichnisende liest `msepaket`** ([msepaket.LiesArchivende]) —
// dieselbe Stelle wie für die Ressourcenregel der Pakete (#1886).
const (
	kopfsignatur = 0x02014b50
	kopflaenge   = 46

	// namensgrenze ist, wie viel eines Namens behalten wird. **Ein Kopf darf
	// einen Namen von 64 KiB tragen**; zweitausend davon wären 128 MiB für
	// eine Liste, die niemand so breit liest.
	namensgrenze = 1024
)

var errKeinZip = errors.New("kein lesbares ZIP-Verzeichnis")

// archivEintraege listet ein ZIP, ohne etwas auszupacken — und ohne mehr als
// `hoechstens` Köpfe zu lesen.
//
// **Die Größe kommt vom Rufer**, der die Datei eben angesehen hat; ist sie
// seither eine andere, scheitert das Lesen an ihr und nicht an einer Annahme.
func archivEintraege(pfad string, groesse int64, hoechstens int) (eintraege []Dateieintrag, unsicher int, unlesbar bool) {
	datei, fehler := os.Open(pfad) // #nosec G304 -- die Datei, die jemand ansehen will
	if fehler != nil {
		return nil, 0, true
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // nur gelesen

	eintraege, unsicher, fehler = verzeichnisLesen(datei, groesse, hoechstens)
	if fehler != nil {
		return nil, 0, true
	}

	return eintraege, unsicher, false
}

// verzeichnisLesen liest Kopf um Kopf, bis die Grenze erreicht ist.
func verzeichnisLesen(quelle io.ReaderAt, groesse int64, hoechstens int) ([]Dateieintrag, int, error) {
	ende, fehler := msepaket.LiesArchivende(quelle, groesse)
	if fehler != nil {
		return nil, 0, fehler
	}

	// #nosec G115 -- LiesArchivende hat beide auf 63 Bit begrenzt
	verzeichnisgroesse, versatz := int64(ende.Groesse), int64(ende.Versatz)

	basis := ende.Endversatz - verzeichnisgroesse - versatz
	if anfang := basis + versatz; anfang < 0 || anfang >= groesse {
		return nil, 0, errKeinZip
	}

	// **Ein Vorsatz, den das Ende falsch nennt**: steht an Versatz null
	// schon ein Kopf, gilt null — dieselbe Nachsicht wie `archive/zip`.
	if basis > 0 {
		if _, fehler := kopfLesen(bufio.NewReader(io.NewSectionReader(quelle, versatz, groesse))); fehler == nil {
			basis = 0
		}
	}

	leser := bufio.NewReader(io.NewSectionReader(quelle, basis+versatz, groesse))

	var (
		eintraege []Dateieintrag
		unsicher  int
	)

	for len(eintraege) < hoechstens {
		einer, fehler := kopfLesen(leser)
		if fehler != nil {
			// #nosec G115 -- bewusst über 16 Bit, siehe Kopf
			if uint16(len(eintraege)) != uint16(ende.Zahl) {
				return nil, 0, errKeinZip
			}

			break
		}

		eintraege = append(eintraege, einer)
		if einer.Unsicher {
			unsicher++
		}
	}

	return eintraege, unsicher, nil
}

// kopfLesen liest einen Kopf des Verzeichnisses und macht daraus einen
// Eintrag der Ansicht.
func kopfLesen(leser *bufio.Reader) (Dateieintrag, error) {
	var kopf [kopflaenge]byte
	if _, fehler := io.ReadFull(leser, kopf[:]); fehler != nil {
		return Dateieintrag{}, fehler
	}

	if binary.LittleEndian.Uint32(kopf[:]) != kopfsignatur {
		return Dateieintrag{}, errKeinZip
	}

	ersteller := binary.LittleEndian.Uint16(kopf[4:]) >> 8
	ausgepackt := uint64(binary.LittleEndian.Uint32(kopf[24:]))
	namenslaenge := int(binary.LittleEndian.Uint16(kopf[28:]))
	extralaenge := int(binary.LittleEndian.Uint16(kopf[30:]))
	kommentarlaenge := int(binary.LittleEndian.Uint16(kopf[32:]))
	merkmale := binary.LittleEndian.Uint32(kopf[38:])

	rest := make([]byte, namenslaenge+extralaenge)
	if _, fehler := io.ReadFull(leser, rest); fehler != nil {
		return Dateieintrag{}, fehler
	}

	if _, fehler := leser.Discard(kommentarlaenge); fehler != nil {
		return Dateieintrag{}, fehler
	}

	name := string(rest[:namenslaenge])
	if ausgepackt == 0xffffffff {
		groesse, lesbar := zip64Groesse(rest[namenslaenge:])
		if !lesbar {
			return Dateieintrag{}, errKeinZip
		}

		ausgepackt = groesse
	}

	// **Unix und macOS tragen ihre Art in den oberen 16 Bit** (Ersteller 3
	// und 19); ein Verzeichnis erkennt jeder Ersteller am Schrägstrich.
	art := uint32(0)
	if ersteller == 3 || ersteller == 19 {
		art = (merkmale >> 16) & 0xf000
	}

	gefaehrlich := strings.HasPrefix(name, "/") || strings.HasPrefix(path.Clean(name), "..") || art == 0xa000

	return Dateieintrag{
		Name:        gekuerzterName(name),
		Groesse:     int64(min(ausgepackt, 1<<63-1)), //nolint:gosec // oben begrenzt
		Verzeichnis: strings.HasSuffix(name, "/") || art == 0x4000,
		Unsicher:    gefaehrlich,
	}, nil
}

// zip64Groesse sucht die ausgepackte Größe im zip64-Feld — sie steht dort
// zuerst, wenn die alte voll ist.
//
// **Ein zip64-Feld mit weniger als acht Bytes ist kaputt**, wie in Go
// (`readDirectoryHeader`: `ErrFormat`). **Fehlt das Feld ganz** oder ist es
// abgeschnitten, bleibt die volle alte Größe — auch das wie in Go, das die
// Suche dann abbricht, statt abzuweisen.
func zip64Groesse(extra []byte) (groesse uint64, lesbar bool) {
	for len(extra) >= 4 {
		marke, laenge := binary.LittleEndian.Uint16(extra), int(binary.LittleEndian.Uint16(extra[2:]))
		if len(extra) < 4+laenge {
			break
		}

		if marke == 0x0001 {
			if laenge < 8 {
				return 0, false
			}

			return binary.LittleEndian.Uint64(extra[4:]), true
		}

		extra = extra[4+laenge:]
	}

	return 0xffffffff, true
}

// gekuerzterName behält höchstens [namensgrenze] Bytes, ohne ein Zeichen zu
// zerschneiden.
func gekuerzterName(name string) string {
	if len(name) <= namensgrenze {
		return name
	}

	return string(bytes.ToValidUTF8([]byte(name[:namensgrenze]), nil)) + "…"
}
