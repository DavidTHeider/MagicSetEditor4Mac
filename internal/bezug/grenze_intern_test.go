// Die Grenze und der Speicher — zwei Fragen an denselben Weg.
//
// # Warum ein Test **im** Paket
//
// Beide Fragen hängen an `hoechstens`, und die Grenze steht bei zwei Gigabyte.
// Ein Test von außen müsste ein Archiv dieser Größe bauen; das ist kein Test,
// sondern eine Wartezeit. Von innen lässt sich die Grenze für die Dauer eines
// Falls herunterdrehen — und dann fährt der Test **denselben** Weg, den ein
// zwei Gigabyte großes Archiv führe.
//
// Dieselbe Begründung wie bei `internal/render/kerningleser_intern_test.go`.
//
// # Was hier nicht geprüft wird
//
//   - **Zwei Gigabyte selbst.** Geprüft wird das Verhalten an der Grenze, nicht
//     die Zahl. Dass sie 2 GiB ist, sagt die Konstante.
//   - **Der Speicherbedarf des Auspackens im Ganzen.** Gemessen wird, dass der
//     **Inhalt** nicht ein zweites Mal entsteht — nicht, was gzip und tar
//     darüber hinaus brauchen.
package bezug

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
)

// mitGrenze dreht die Grenze für die Dauer eines Falls herunter.
func mitGrenze(t *testing.T, neu int64) {
	t.Helper()

	alt := hoechstens
	hoechstens = neu

	t.Cleanup(func() { hoechstens = alt })
}

// archiv baut ein `.tar.gz` mit einer Datei bekannter Größe.
//
// **Ohne Verdichtung**: der Test misst den Speicher für den **gelesenen**
// Inhalt, und der ist die verdichtete Größe. Ein gut verdichtetes Archiv wäre
// winzig und die Messung bedeutungslos.
func archivOhneVerdichtung(t *testing.T, name string, groesse int) []byte {
	t.Helper()

	var gepackt bytes.Buffer

	packer, err := gzip.NewWriterLevel(&gepackt, gzip.NoCompression)
	if err != nil {
		t.Fatalf("gzip: %v", err)
	}

	schreiber := tar.NewWriter(packer)

	kopf := &tar.Header{Name: name, Mode: 0o644, Size: int64(groesse), Typeflag: tar.TypeReg}
	if err := schreiber.WriteHeader(kopf); err != nil {
		t.Fatalf("Kopf: %v", err)
	}

	if _, err := schreiber.Write(make([]byte, groesse)); err != nil {
		t.Fatalf("Inhalt: %v", err)
	}

	if err := schreiber.Close(); err != nil {
		t.Fatalf("tar schließen: %v", err)
	}

	if err := packer.Close(); err != nil {
		t.Fatalf("gzip schließen: %v", err)
	}

	return gepackt.Bytes()
}

// **Auspacken macht keine zweite Kopie des Archivs.**
//
// Bis zum 2026-08-23 stand hier `gzip.NewReader(strings.NewReader(string(inhalt)))`.
// `string([]byte)` kopiert in Go **immer** — bei zwei Gigabyte sind das zwei
// Gigabyte mehr, und zwar bevor gzip anfängt
// (#898).
func TestAuspackenKopiertDasArchivNicht(t *testing.T) {
	const inhaltsgroesse = 16 << 20

	gepackt := archivOhneVerdichtung(t, "probe/datei.txt", inhaltsgroesse)

	// **Die verdichtete Größe ist der Maßstab**, denn sie ist es, die kopiert
	// würde. Mit `NoCompression` liegt sie knapp über der Inhaltsgröße.
	archivgroesse := uint64(len(gepackt))

	belegt := belegtWaehrend(t, func() {
		if _, err := auspacken(bytes.NewReader(gepackt), "probe", t.TempDir()); err != nil {
			t.Fatalf("auspacken: %v", err)
		}
	})

	// **Die Schwelle ist großzügig und trotzdem scharf.** gzip und tar
	// brauchen Puffer, und die geschriebene Datei geht durch `io.Copy`; das
	// alles liegt weit unter einer vollen Kopie des Archivs. Eine zweite Kopie
	// schlüge mit der ganzen Archivgröße zu Buche und risse die Schwelle.
	schwelle := archivgroesse / 2

	if belegt > schwelle {
		t.Errorf("das Auspacken belegte %d Bytes bei einem Archiv von %d — "+
			"mehr als die Schwelle %d. Eine zweite Kopie?",
			belegt, archivgroesse, schwelle)
	}

	t.Logf("Archiv %d Bytes, belegt %d Bytes (%.1f %% der Archivgröße)",
		archivgroesse, belegt, float64(belegt)*100/float64(archivgroesse))
}

// **Ein Archiv über der Grenze wird abgewiesen, nicht gekappt.**
//
// # Warum dieser Test jetzt über das Netz geht
//
// Er rief bis zum 2026-08-23 `lies` unmittelbar an — und `lies` gibt es nicht
// mehr: das Lesen bis zur Grenze steht seit
// #1005 in
// `internal/netz`. Der Test wäre grün geblieben und hätte **toten Code**
// geprüft; das ist die Sorte Grün, die nichts über den Betrieb sagt (§10).
//
// Geprüft wird deshalb der Weg, den ein Bezug wirklich nimmt: eine Gegenstelle
// schickt mehr, als die Grenze erlaubt.
func TestEinArchivUeberDerGrenzeWirdBenanntAbgewiesen(t *testing.T) {
	mitGrenze(t, 32)

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if _, err := schreiben.Write(make([]byte, 4096)); err != nil {
				t.Errorf("die Gegenstelle konnte nicht schreiben: %v", err)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	_, err := laden(t.Context(), gegenstelle.URL, "prüfung")
	if err == nil {
		t.Fatal("ein Archiv über der Grenze kam durch")
	}

	if !bytes.Contains([]byte(err.Error()), []byte("größer als")) {
		t.Errorf("die Meldung nennt die Grenze nicht: %v", err)
	}
}

// **Und eines genau auf der Grenze kommt durch.**
//
// Der Fall, an dem eine Prüfung `>` gegen `>=` entscheidet — und an dem eine
// Grenze, die um eins danebenliegt, ein gültiges Archiv abwiese.
func TestEinArchivGenauAufDerGrenzeKommtDurch(t *testing.T) {
	mitGrenze(t, 4096)

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if _, err := schreiben.Write(make([]byte, 4096)); err != nil {
				t.Errorf("die Gegenstelle konnte nicht schreiben: %v", err)
			}
		}))
	t.Cleanup(gegenstelle.Close)

	inhalt, err := laden(t.Context(), gegenstelle.URL, "prüfung")
	if err != nil {
		t.Fatalf("ein Archiv genau auf der Grenze wurde abgewiesen: %v", err)
	}

	if len(inhalt) != 4096 {
		t.Errorf("es kamen %d Bytes an", len(inhalt))
	}
}

// belegtWaehrend misst, wie viel ein Abschnitt insgesamt belegt hat.
//
// **`TotalAlloc` und nicht `HeapAlloc`**: gefragt ist, wie viel **entstanden**
// ist, nicht wie viel am Ende noch liegt. Eine Kopie, die der Sammler
// zwischendurch wegräumt, ist trotzdem entstanden.
func belegtWaehrend(t *testing.T, tun func()) uint64 {
	t.Helper()

	var vorher, nachher runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&vorher)

	tun()

	runtime.ReadMemStats(&nachher)

	return nachher.TotalAlloc - vorher.TotalAlloc
}
