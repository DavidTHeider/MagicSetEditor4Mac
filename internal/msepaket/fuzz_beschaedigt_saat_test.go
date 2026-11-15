package msepaket_test

// Die Saatkörner für [FuzzBeschaedigtesPaket] und ihre Bauer.
//
// **Eine eigene Datei, weil sie eine eigene Frage beantwortet** (§2): nicht
// *was zugesichert ist* — das steht in `fuzz_beschaedigt_test.go` —, sondern
// *womit der Fuzzer anfängt*. Beides in einer Datei ging über die Zeilengrenze,
// als der Nachlauf über Phase 4.2 den Vergleich der ganzen Paketantwort
// verlangte
// (#1289).

import (
	"archive/zip"
	"bytes"
	"testing"
)

// heilesArchiv baut ein wohlgeformtes `.mse-set` als Bytefolge.
func heilesArchiv(f *testing.F) []byte {
	f.Helper()

	var puffer bytes.Buffer

	packer := zip.NewWriter(&puffer)

	// **Feste Reihenfolge und ein Eintrag ungepackt.** Über eine `map` zu
	// gehen gäbe bei jedem Lauf eine andere Bytefolge — ein Saatkorn, das sich
	// von Lauf zu Lauf ändert, ist keines. Und `Store` ist die Form, in der
	// MSE2 seine Bilder ablegt: dort geht ein verdrehtes Byte unverändert durch
	// den Entpacker, und **nur** die Prüfsumme merkt es.
	for _, eintrag := range []struct {
		name      string
		inhalt    string
		verfahren uint16
	}{
		{"set", kopfinhalt, zip.Deflate},
		{beigabe, beigabeninhalt, zip.Store},
	} {
		strom, err := packer.CreateHeader(&zip.FileHeader{
			Name: eintrag.name, Method: eintrag.verfahren,
		})
		if err != nil {
			f.Fatalf("Eintrag: %v", err)
		}

		if _, err := strom.Write([]byte(eintrag.inhalt)); err != nil {
			f.Fatalf("schreiben: %v", err)
		}
	}

	if err := packer.Close(); err != nil {
		f.Fatalf("schließen: %v", err)
	}

	return puffer.Bytes()
}

// mitVerdrehtemByte kippt ein Bit mitten im Archiv.
//
// **Mitten hinein und nicht ans Ende**: das Verzeichnis eines Archivs steht
// hinten, und wer es trifft, bekommt eine Ablehnung beim Öffnen. Interessant
// ist der andere Fall — ein Archiv, das aufgeht, und ein Eintrag, der nicht
// mehr das ist, was seine Prüfsumme sagt.
func mitVerdrehtemByte(roh []byte) []byte {
	verdreht := bytes.Clone(roh)
	if len(verdreht) > 40 {
		verdreht[len(verdreht)/3] ^= 0xff
	}

	return verdreht
}

// Der Inhalt des heilen Archivs, an einer Stelle und mit Namen.
const (
	beigabe        = "bild1"
	kopfinhalt     = "mse version: 2.0.0\ngame: probe\ncard:\n\tname: A\ncard:\n\tname: B\n"
	beigabeninhalt = "nicht wirklich ein Bild"
)

// mitVerdrehterBeigabe verdreht ein Byte **in der Beigabe**, nicht im Kopf.
//
// **Und darauf kommt es an.** Wer die Hauptdatei trifft, bekommt eine Ablehnung
// beim Öffnen — der Weg, den [dateienPruefen] gar nicht sieht. Der Fall, für den
// die Prüfsummentafel dasteht, ist der andere: ein Paket, das **aufgeht**, und
// eine Datei darin, die nicht mehr das ist, was das Archiv über sie sagt.
//
// Gefunden beim Gegenprobieren: mit dem Byte an einer beliebigen Stelle blieb
// der eingebaute Fehler unbemerkt, weil kein Saatkorn je bis zum Lesen kam.
func mitVerdrehterBeigabe(roh []byte) []byte {
	stelle := bytes.Index(roh, []byte(beigabeninhalt))
	if stelle < 0 {
		return roh
	}

	verdreht := bytes.Clone(roh)
	verdreht[stelle+1] ^= 0xff

	return verdreht
}

// mitDoppeltemNamen baut ein Archiv, das denselben Eintrag zweimal führt.
//
// **Ein ZIP darf das**, und welchen der beiden ein Leser nimmt, steht nirgends.
// Das Saatkorn steht hier, damit die Ausnahme in [crcTafel] einen Fall hat: ohne
// es blieb der falsche Sentinel unbemerkt.
func mitDoppeltemNamen(f *testing.F) []byte {
	f.Helper()

	var puffer bytes.Buffer

	packer := zip.NewWriter(&puffer)

	for _, inhalt := range []string{beigabeninhalt, "etwas ganz anderes"} {
		strom, err := packer.CreateHeader(&zip.FileHeader{
			Name: beigabe, Method: zip.Store,
		})
		if err != nil {
			f.Fatalf("Eintrag: %v", err)
		}

		if _, err := strom.Write([]byte(inhalt)); err != nil {
			f.Fatalf("schreiben: %v", err)
		}
	}

	kopf, err := packer.Create("set")
	if err != nil {
		f.Fatalf("Eintrag: %v", err)
	}

	if _, err := kopf.Write([]byte(kopfinhalt)); err != nil {
		f.Fatalf("schreiben: %v", err)
	}

	if err := packer.Close(); err != nil {
		f.Fatalf("schließen: %v", err)
	}

	return puffer.Bytes()
}
