package msepaket

// **Das Ende eines Archivs, ohne sein Verzeichnis zu lesen** —
// #1886.
//
// **Hierher umgezogen aus `internal/service/zipverzeichnis.go`** (#1885),
// damit die Ressourcenregel für Pakete und die Dateiansicht **dasselbe** Ende
// lesen: übernommen ist das Verhalten von `archive/zip` (Go 1.27.1,
// `src/archive/zip/reader.go`) — die Suche der Endsignatur von hinten samt
// Prüfung der Kommentarlänge (`readDirectoryEnd`, `findSignatureInBlock`) und
// zip64 über den Verweis davor (`findDirectory64End`, `readDirectory64End`).

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	endsignatur     = 0x06054b50
	endlaenge       = 22
	zip64verweis    = 0x07064b50
	zip64verweislen = 20
	zip64ende       = 0x06064b50
	zip64endelen    = 56
)

// errKeinArchivende heißt: am Ende steht nichts, was ein Archiv beschlösse.
var errKeinArchivende = errors.New("kein lesbares Verzeichnisende")

// LiesArchivende sucht das Verzeichnisende in den letzten 65 KiB und folgt, wo es
// voll ist, dem zip64-Ende.
func LiesArchivende(quelle io.ReaderAt, groesse int64) (Archivende, error) {
	block := min(groesse, 65*1024+endlaenge)
	roh := make([]byte, block)

	if _, fehler := quelle.ReadAt(roh, groesse-block); fehler != nil && !errors.Is(fehler, io.EOF) {
		return Archivende{}, fehler
	}

	stelle := endsignaturSuchen(roh)
	if stelle < 0 {
		return Archivende{}, errKeinArchivende
	}

	feld := roh[stelle:]
	ende := Archivende{
		Zahl:       uint64(binary.LittleEndian.Uint16(feld[10:])),
		Groesse:    uint64(binary.LittleEndian.Uint32(feld[12:])),
		Versatz:    uint64(binary.LittleEndian.Uint32(feld[16:])),
		Endversatz: groesse - block + int64(stelle),
	}

	if ende.Zahl == 0xffff || ende.Groesse == 0xffffffff || ende.Versatz == 0xffffffff {
		if fehler := zip64Lesen(quelle, &ende); fehler != nil {
			return Archivende{}, fehler
		}
	}

	if ende.Groesse > 1<<63-1 || ende.Versatz > 1<<63-1 {
		return Archivende{}, errKeinArchivende
	}

	return ende, nil
}

// endsignaturSuchen sucht von hinten — und nimmt nur ein Ende, dessen
// Kommentar in den Block passt.
func endsignaturSuchen(roh []byte) int {
	for stelle := len(roh) - endlaenge; stelle >= 0; stelle-- {
		if binary.LittleEndian.Uint32(roh[stelle:]) != endsignatur {
			continue
		}

		if stelle+endlaenge+int(binary.LittleEndian.Uint16(roh[stelle+20:])) > len(roh) {
			return -1
		}

		return stelle
	}

	return -1
}

// zip64Lesen liest das zip64-Ende über den Verweis unmittelbar davor.
func zip64Lesen(quelle io.ReaderAt, ende *Archivende) error {
	stelle := ende.Endversatz - zip64verweislen
	if stelle < 0 {
		return nil
	}

	verweis := make([]byte, zip64verweislen)
	if _, fehler := quelle.ReadAt(verweis, stelle); fehler != nil {
		return fehler
	}

	if binary.LittleEndian.Uint32(verweis) != zip64verweis ||
		binary.LittleEndian.Uint32(verweis[4:]) != 0 || binary.LittleEndian.Uint32(verweis[16:]) != 1 {
		return nil
	}

	// #nosec G115 -- ein Versatz; ReadAt weist einen negativen ab
	anfang := int64(binary.LittleEndian.Uint64(verweis[8:]))
	roh := make([]byte, zip64endelen)

	if _, fehler := quelle.ReadAt(roh, anfang); fehler != nil {
		return fehler
	}

	if binary.LittleEndian.Uint32(roh) != zip64ende {
		return errKeinArchivende
	}

	ende.Zahl = binary.LittleEndian.Uint64(roh[32:])
	ende.Groesse = binary.LittleEndian.Uint64(roh[40:])
	ende.Versatz = binary.LittleEndian.Uint64(roh[48:])
	ende.Endversatz = anfang

	return nil
}
