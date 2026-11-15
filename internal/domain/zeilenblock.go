package domain

// **Ein markierter Block in einer fremden Datei** — OPS-09
// (#1907),
// Marke `new`: MSE2 richtet nichts außerhalb seines Programms ein.
//
// # Warum ein Block und keine Zeile
//
// **Weil widerrufen genau das nehmen muss, was eingesetzt wurde — und nichts
// sonst.** Eine Zeile `export PATH=…` kann ein Mensch selbst geschrieben
// haben; dieselbe Zeile zwischen zwei Marken hat dieses Programm geschrieben.
// So hält es auch `conda init` in der `.zshrc`, und so liest es jeder, der
// die Datei öffnet.
//
// # Was byte-gleich heißt
//
// **Widerrufen hinterlässt die Datei, wie sie vor dem Einsetzen war** (AK-02).
// Hing an ihr kein Zeilenende, musste eines davor — die Anfangsmarke sagt es,
// und widerrufen nimmt es mit.

import (
	"bytes"
	"strings"
)

// Die Marken eines Blocks — beide als Kommentar in Shell und TOML.
const (
	BlockAnfang = "# >>> MagicSetEditor4Mac — eingetragen auf Wunsch; „Widerrufen“ nimmt genau diesen Block"
	BlockEnde   = "# <<< MagicSetEditor4Mac"

	// blockErgaenzt hängt an der Anfangsmarke, wenn davor ein Zeilenende
	// ergänzt wurde.
	blockErgaenzt = " · Zeilenende davor ergänzt"
)

// Blockstand ist, was in einer Datei steht.
type Blockstand string

// Die vier Stände.
const (
	BlockFehlt       Blockstand = "fehlt"
	BlockEingetragen Blockstand = "eingetragen"
	BlockVeraltet    Blockstand = "veraltet"

	// BlockFremd ist ein Anfang ohne Ende — nicht mehr der Block, den dieses
	// Programm geschrieben hat, und deshalb nicht anzufassen.
	BlockFremd Blockstand = "fremd"
)

// blocklage ist, wo ein Block steht.
type blocklage struct {
	anfang, innen, ende int // Bytegrenzen: Anfangsmarke, Inhalt, nach der Endmarke
	ergaenzt            bool
}

// blockFinden sucht den ersten Block.
func blockFinden(inhalt []byte) (blocklage, Blockstand) {
	start := zeilenanfang(inhalt, BlockAnfang)
	if start < 0 {
		return blocklage{}, BlockFehlt
	}

	zeilenende := bytes.IndexByte(inhalt[start:], '\n')
	if zeilenende < 0 {
		return blocklage{}, BlockFremd
	}

	innen := start + zeilenende + 1
	schluss := zeilenanfang(inhalt[innen:], BlockEnde)

	if schluss < 0 {
		return blocklage{}, BlockFremd
	}

	ende := innen + schluss + len(BlockEnde)
	if ende < len(inhalt) && inhalt[ende] == '\n' {
		ende++
	}

	anfangszeile := string(inhalt[start : start+zeilenende])

	return blocklage{
		anfang: start, innen: innen, ende: ende,
		ergaenzt: strings.HasSuffix(anfangszeile, blockErgaenzt),
	}, BlockEingetragen
}

// zeilenanfang gibt die Stelle der ersten Zeile, die mit `marke` beginnt.
func zeilenanfang(inhalt []byte, marke string) int {
	// **Ohne Bedingung im Kopf**: `stelle` springt nur hinter ein
	// Zeilenende und bleibt damit höchstens bei `len(inhalt)`; dort findet
	// `IndexByte` keines mehr, und die Schleife endet.
	for stelle := 0; ; {
		if bytes.HasPrefix(inhalt[stelle:], []byte(marke)) {
			return stelle
		}

		naechste := bytes.IndexByte(inhalt[stelle:], '\n')
		if naechste < 0 {
			return -1
		}

		stelle += naechste + 1
	}
}

// BlockStand sagt, ob `block` eingetragen, veraltet, fremd oder nicht da ist.
func BlockStand(inhalt []byte, block string) Blockstand {
	lage, stand := blockFinden(inhalt)
	if stand != BlockEingetragen {
		return stand
	}

	endmarke := lage.innen + zeilenanfang(inhalt[lage.innen:], BlockEnde)
	if string(inhalt[lage.innen:endmarke]) != mitZeilenende(block) {
		return BlockVeraltet
	}

	return BlockEingetragen
}

// BlockEinsetzen setzt `block` ein — oder ersetzt einen veralteten an seiner
// Stelle. **Ein fremder bleibt, wie er ist.**
func BlockEinsetzen(inhalt []byte, block string) []byte {
	lage, stand := blockFinden(inhalt)

	switch stand {
	case BlockFremd:
		return inhalt
	case BlockEingetragen:
		endmarke := lage.innen + zeilenanfang(inhalt[lage.innen:], BlockEnde)

		neu := append([]byte{}, inhalt[:lage.innen]...)
		neu = append(neu, mitZeilenende(block)...)

		return append(neu, inhalt[endmarke:]...)
	}

	anfang := BlockAnfang
	neu := append([]byte{}, inhalt...)

	if len(neu) > 0 && neu[len(neu)-1] != '\n' {
		neu = append(neu, '\n')
		anfang += blockErgaenzt
	}

	return append(neu, anfang+"\n"+mitZeilenende(block)+BlockEnde+"\n"...)
}

// BlockEntfernen nimmt den Block — und das Zeilenende, das für ihn ergänzt
// wurde. **Ein fremder oder fehlender ändert nichts.**
func BlockEntfernen(inhalt []byte) []byte {
	lage, stand := blockFinden(inhalt)
	if stand != BlockEingetragen {
		return inhalt
	}

	anfang := lage.anfang
	if lage.ergaenzt && anfang > 0 {
		anfang--
	}

	return append(append([]byte{}, inhalt[:anfang]...), inhalt[lage.ende:]...)
}

func mitZeilenende(block string) string {
	if strings.HasSuffix(block, "\n") {
		return block
	}

	return block + "\n"
}
