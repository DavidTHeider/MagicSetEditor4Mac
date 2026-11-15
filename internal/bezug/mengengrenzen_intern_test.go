// **Was ein Archiv insgesamt kosten darf** — Befund Z7-01 des
// Audits für 0.3.0 (Entwicklungsarchiv), Phase 3.2
// (#1163).
//
// # Der Befund
//
// `io.LimitReader` deckelte den **einzelnen** Eintrag. Ein Archiv mit tausend
// Einträgen zu je zwei Gigabyte war damit unbegrenzt — und ein Eintrag
// **über** der Grenze wurde stillschweigend **abgeschnitten**: `io.Copy` gab
// keinen Fehler, und auf der Platte lag eine beschädigte Datei, die aussah wie
// eine ganze.
//
// # Warum ein Test **im** Paket
//
// Dieselbe Begründung wie bei `grenze_intern_test.go`: die Grenzen stehen bei
// Gigabyte, und ein Test von außen müsste solche Archive bauen. Von innen
// lassen sie sich herunterdrehen, und dann fährt der Lauf **denselben** Weg.
//
// **Was hier nicht geprüft wird**: ob die Zahlen richtig gewählt sind. Geprüft
// wird, dass sie greifen.
package bezug

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// mitMengengrenzen dreht alle drei Grenzen für die Dauer eines Falls herunter.
func mitMengengrenzen(t *testing.T, eintrag, gesamt int64, anzahl int) {
	t.Helper()

	alteEintrag, alteGesamt, alteAnzahl := hoechstensEintrag, hoechstensGesamt, hoechstensEintraege
	hoechstensEintrag, hoechstensGesamt, hoechstensEintraege = eintrag, gesamt, anzahl

	t.Cleanup(func() {
		hoechstensEintrag, hoechstensGesamt, hoechstensEintraege = alteEintrag, alteGesamt, alteAnzahl
	})
}

// vieleEintraege baut ein `tar.gz` mit `anzahl` Dateien zu je `groesse` Bytes.
func vieleEintraege(t *testing.T, anzahl, groesse int) []byte {
	t.Helper()

	var gepackt bytes.Buffer

	packer := gzip.NewWriter(&gepackt)
	schreiber := tar.NewWriter(packer)

	// **Nullbytes**: sie verdichten sich fast auf nichts, und genau das ist
	// die Gestalt einer Kompressionsbombe.
	inhalt := make([]byte, groesse)

	for i := range anzahl {
		if err := schreiber.WriteHeader(&tar.Header{
			Name:     "wurzel/data/datei" + string(rune('a'+i%26)) + string(rune('a'+i/26)),
			Mode:     0o644,
			Size:     int64(groesse),
			Typeflag: tar.TypeReg,
		}); err != nil {
			t.Fatalf("Kopf schreiben: %v", err)
		}

		if _, err := schreiber.Write(inhalt); err != nil {
			t.Fatalf("Inhalt schreiben: %v", err)
		}
	}

	if err := schreiber.Close(); err != nil {
		t.Fatalf("tar schließen: %v", err)
	}

	if err := packer.Close(); err != nil {
		t.Fatalf("gzip schließen: %v", err)
	}

	return gepackt.Bytes()
}

// **Ein Archiv, dessen Einträge zusammen zu viel wiegen, wird abgewiesen.**
//
// Jeder für sich ist unter der Grenze. Ohne ein Budget über alle wäre ein
// Archiv mit tausend zulässigen Einträgen unbegrenzt.
func TestEinZuSchweresArchivWirdAbgewiesen(t *testing.T) {
	mitMengengrenzen(t, 1<<20, 4<<10, 100)

	_, fehler := auspacken(bytes.NewReader(vieleEintraege(t, 5, 2<<10)), "", t.TempDir())
	if fehler == nil {
		t.Fatal("das zu schwere Archiv kam durch")
	}

	if !strings.Contains(fehler.Error(), "zusammen") {
		t.Errorf("die Meldung nennt den Grund nicht: %v", fehler)
	}
}

// **Ein Archiv mit zu vielen Einträgen wird abgewiesen.**
//
// Der Fall, den ein Bytebudget durchlässt: sehr viele sehr kleine Einträge.
// Jeder kostet einen Systemaufruf und einen Verzeichniseintrag.
func TestEinArchivMitZuVielenEintraegenWirdAbgewiesen(t *testing.T) {
	mitMengengrenzen(t, 1<<20, 1<<30, 3)

	_, fehler := auspacken(bytes.NewReader(vieleEintraege(t, 5, 8)), "", t.TempDir())
	if fehler == nil {
		t.Fatal("das Archiv mit zu vielen Einträgen kam durch")
	}

	if !strings.Contains(fehler.Error(), "Einträge") {
		t.Errorf("die Meldung nennt den Grund nicht: %v", fehler)
	}
}

// **Ein zu großer Eintrag ist ein Fehler und kein halber Erfolg.**
//
// # Der teuerste der drei
//
// Bis hierher schnitt `io.LimitReader` ihn ab, und `io.Copy` meldete nichts.
// Auf der Platte lag eine Datei, die aussah wie eine ganze — ein Paket, das
// sich öffnen lässt und dessen Vorlage in der Mitte aufhört.
//
// **Ein stiller Fehlschlag ist schlimmer als ein lauter**: er wird erst dann
// sichtbar, wenn jemand die Karte zeichnet, und dann sucht er ihn im Paket.
func TestEinZuGrosserEintragIstEinFehler(t *testing.T) {
	mitMengengrenzen(t, 4<<10, 1<<30, 100)

	ziel := t.TempDir()

	_, fehler := auspacken(bytes.NewReader(vieleEintraege(t, 1, 16<<10)), "", ziel)
	if fehler == nil {
		t.Fatal("der zu große Eintrag kam als Erfolg durch")
	}

	if !strings.Contains(fehler.Error(), "zu groß") {
		t.Errorf("die Meldung nennt den Grund nicht: %v", fehler)
	}
}

// **Und ein Archiv unter allen drei Grenzen kommt durch.**
//
// Die Gegenprobe: ohne sie sagten die drei Läufe darüber auch dann ja, wenn
// **jedes** Archiv abgewiesen würde.
func TestEinArchivUnterDenGrenzenKommtDurch(t *testing.T) {
	mitMengengrenzen(t, 1<<20, 1<<30, 100)

	uebergangen, fehler := auspacken(bytes.NewReader(vieleEintraege(t, 5, 2<<10)), "", t.TempDir())
	if fehler != nil {
		t.Fatalf("das zulässige Archiv wurde abgewiesen: %v", fehler)
	}

	if len(uebergangen) != 0 {
		t.Errorf("es wurde etwas übergangen: %v", uebergangen)
	}
}

// **Das Gesamtbudget greift, bevor ein weiterer Eintrag es überschreitet.**
//
// # Der Befund des Zweitblicks
//
// Der erste Anlauf gab jedem Eintrag die volle **Einzelgrenze** und zählte
// **danach**: bei knapp am Budget durfte der nächste noch einmal so viel
// schreiben, bevor die Abweisung kam. Die zugesagten vier Gigabyte wurden
// damit fast sechs.
//
// **Gemessen wird, was auf der Platte liegt** — nicht, ob eine Meldung kommt.
// Die kam vorher auch.
func TestDasGesamtbudgetGreiftVorDemUeberschreiten(t *testing.T) {
	// **Einzelgrenze 8 KiB, Budget 10 KiB.** Drei Einträge zu je 4 KiB: nach
	// zweien sind 8 KiB da, und der dritte darf höchstens noch 2 KiB kosten.
	mitMengengrenzen(t, 8<<10, 10<<10, 100)

	ziel := t.TempDir()

	_, fehler := auspacken(bytes.NewReader(vieleEintraege(t, 3, 4<<10)), "", ziel)
	if fehler == nil {
		t.Fatal("das zu schwere Archiv kam durch")
	}

	if !strings.Contains(fehler.Error(), "zusammen") {
		t.Errorf("die Meldung nennt den Grund nicht: %v", fehler)
	}

	// **Höchstens das Budget plus ein Byte.** Der erste Anlauf ließ hier
	// 8 + 8 = 16 KiB zu.
	gewicht := gewichtUnter(t, ziel)
	if gewicht > 10<<10+1 {
		t.Errorf("auf der Platte liegen %d Byte — das Budget ist %d", gewicht, 10<<10)
	}
}

// **Ein Archiv mit sehr vielen Verzeichnissen wird abgewiesen.**
//
// # Der Befund des Zweitblicks
//
// Gezählt wurde **hinter** der Auswahl. Verzeichnisse, Sondertypen, Einträge
// außerhalb des gewählten Unterverzeichnisses und übergangene Programme kamen
// gar nicht bis zum Zähler — ein Tar mit Millionen davon ging vollständig an
// der Grenze vorbei und kostete trotzdem Laufzeit für jeden.
//
// **Die Klasse ist „nicht ausgepackte Tar-Einträge"**, nicht nur das Beispiel
// Verzeichnis.
func TestEinArchivMitVielenVerzeichnissenWirdAbgewiesen(t *testing.T) {
	mitMengengrenzen(t, 1<<20, 1<<30, 3)

	_, fehler := auspacken(bytes.NewReader(nurVerzeichnisse(t, 5)), "", t.TempDir())
	if fehler == nil {
		t.Fatal("das Archiv mit zu vielen Verzeichnissen kam durch")
	}

	if !strings.Contains(fehler.Error(), "Einträge") {
		t.Errorf("die Meldung nennt den Grund nicht: %v", fehler)
	}
}

// nurVerzeichnisse baut ein `tar.gz` aus lauter Verzeichniseinträgen.
func nurVerzeichnisse(t *testing.T, anzahl int) []byte {
	t.Helper()

	var gepackt bytes.Buffer

	packer := gzip.NewWriter(&gepackt)
	schreiber := tar.NewWriter(packer)

	for i := range anzahl {
		if err := schreiber.WriteHeader(&tar.Header{
			Name:     "wurzel/data/ordner" + strconv.Itoa(i) + "/",
			Mode:     0o755,
			Typeflag: tar.TypeDir,
		}); err != nil {
			t.Fatalf("Kopf schreiben: %v", err)
		}
	}

	if err := schreiber.Close(); err != nil {
		t.Fatalf("tar schließen: %v", err)
	}

	if err := packer.Close(); err != nil {
		t.Fatalf("gzip schließen: %v", err)
	}

	return gepackt.Bytes()
}

// gewichtUnter summiert die Dateigrößen unter einem Verzeichnis.
func gewichtUnter(t *testing.T, wurzel string) int64 {
	t.Helper()

	var gesamt int64

	if err := filepath.WalkDir(wurzel, func(_ string, eintrag os.DirEntry, unterwegs error) error {
		if unterwegs != nil || eintrag.IsDir() {
			return unterwegs
		}

		angaben, fehler := eintrag.Info()
		if fehler != nil {
			return fehler
		}

		gesamt += angaben.Size()

		return nil
	}); err != nil {
		t.Fatalf("zählen: %v", err)
	}

	return gesamt
}
