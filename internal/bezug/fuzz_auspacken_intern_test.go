package bezug

// **Der Entpacker gegen ein feindliches Archiv** —
// #1289.
//
// # Warum hier am schärfsten
//
// Alles andere in diesem Projekt liest Dateien, die jemand **hat**. `bezug`
// lädt sie aus dem Netz: das Archiv ist von einem Fremden gebaut, und wer es
// baut, sucht sich seine Bytes aus. Das ist die einzige Stelle, an der der
// Angreifer die Eingabe **wählt**.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
// Aus dem Issue, wörtlich:
//
//   - Panic · Hänger
//   - **Schreiben außerhalb des Ziels**
//   - **Annahme eines Link- oder Sondertyps**
//   - Erfolg gemeldet, Zustand strukturell ungültig
//
// **Zugesichert ist hier die Wirkung auf der Platte**, nicht die Antwort des
// Entpackers. Ein Entpacker, der „übergangen" sagt und trotzdem schreibt, wäre
// der schlimmste Fall, und genau den fängt der Nachbarordner.
//
// **Und damit ist die zweite Zeile nur halb gedeckt, was hier stehen muss.**
// „Annahme eines Link- oder Sondertyps" hat zwei Hälften: *es entstand ein
// Sondertyp auf der Platte* — die fängt [nurGewoehnlichesImZiel] — und *ein
// feindlicher Eintrag wurde angenommen*. Die zweite ist an der Platte gar
// nicht zu sehen: fiele die Typprüfung weg, legte `schreiben` den Linkkopf als
// **gewöhnliche** Datei mit Modus 0600 an, und dieser Lauf bliebe grün.
//
// Sie wird deshalb dort gehalten, wo sie beobachtbar ist — an der **Ablehnung**,
// mit benannten Fällen: `TestEinArchivKannNichtAusbrechen` (`bezug_test.go`,
// ein Verweis nach oben) und `TestEineWarteschlangeWirdUebergangen`
// (`raender_test.go`, ein Sondertyp). Angestrichen vom Zweitblick über
// Phase 4.2; vorher stand hier „zugesichert" für beide Hälften.
//
// # Warum die Eingabe das **tar** ist und nicht das gzip
//
// Weil gzip eine Prüfsumme trägt. Ein Fuzzer, der ein gültiges gzip mutiert,
// erzeugt in aller Regel ein **ungültiges** — und der Lauf endet in der ersten
// Zeile von [auspacken], ohne den tar-Leser je zu erreichen. Die Eingabe ist
// deshalb der rohe tar-Strom; verdichtet wird hier.
//
// **Was dadurch ungeprüft bleibt**, steht damit auch fest: der gzip-Leser
// selbst. Ihn deckt `verdichtungsbombe_intern_test.go` an der Stelle, an der
// es zählt — beim Größenbudget.
//
//	go test ./internal/bezug -run Fuzz -fuzz FuzzAuspacken -fuzztime 2m -fuzzminimizetime 5s
//
// **Mit kurzer Minimierzeit, und das ist gemessen** (FMT-06,
// #1695).
// Ein Durchlauf kostet hier rund 3,5 ms, und die Vorgabe minimiert jede neue
// Eingabe bis zu 60 Sekunden, bevor sie zählt: eine Minute mit der Vorgabe
// fand **keine einzige** neue Eingabe (42 Ausführungen), zwei Minuten mit 5 s
// fanden 165. Die Zahlen stehen in `docs/RELEASE.md`.

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// FuzzAuspacken fährt den Entpacker gegen erfundene tar-Ströme.
func FuzzAuspacken(f *testing.F) {
	kleineGrenzen(f)

	for _, saat := range archivsaat(f) {
		f.Add(saat)
	}

	f.Fuzz(func(t *testing.T, roh []byte) {
		// **Beobachtet wird eine Ebene über dem Zuhause.**
		//
		// Der erste Anlauf ging nur `heim` durch. Ein Archiv mit genügend `..`
		// steigt darüber hinaus, und dort sah niemand hin — der Wächter deckte
		// die Klasse „Schreiben außerhalb des Ziels" schmaler ab, als er
		// behauptete (Zweitblick über Phase 4.2).
		draussen := t.TempDir()

		heim := filepath.Join(draussen, "heim")
		if err := os.MkdirAll(heim, 0o750); err != nil {
			t.Fatalf("das Zuhause anlegen: %v", err)
		}

		// **Das Ziel liegt tief, und das ist die Zusage.** Ein `..` im Archiv
		// steigt um eine Ebene; acht davon bleiben damit noch **innerhalb**
		// von `draussen`, und der Lauf sieht sie.
		//
		// Der erste Anlauf legte einen Nachbarordner `heim/verboten` an und
		// prüfte **den** — ein Ausbruch aus `heim/ziel` landet aber in `heim`
		// und nicht in `heim/verboten`. Die Gegenprobe blieb deshalb grün,
		// obwohl die Wegprüfung ausgeschaltet war: nicht der Wächter war
		// falsch, sondern die Stelle, an der er hinsah (§10).
		//
		// **Und die Beobachtung hat eine Grenze, die hier stehen muss.** Wer
		// mehr als acht Ebenen steigt, landet im gemeinsamen Temp-Bereich
		// dieses Rechners — das sieht kein Lauf, der ihn nicht besitzt. Der
		// zweite Anlauf hob die Grenze von drei auf acht und schrieb sie auf;
		// vorher stand hier nichts darüber (Zweitblick über Phase 4.2).
		ziel := filepath.Join(heim, "eins", "zwei", "drei", "vier",
			"fuenf", "sechs", "sieben", "ziel")
		if err := os.MkdirAll(ziel, 0o750); err != nil {
			t.Fatalf("das Ziel anlegen: %v", err)
		}

		var verdichtet bytes.Buffer

		schreiber := gzip.NewWriter(&verdichtet)
		if _, err := schreiber.Write(roh); err != nil {
			t.Fatalf("verdichten: %v", err)
		}

		if err := schreiber.Close(); err != nil {
			t.Fatalf("den Verdichter schließen: %v", err)
		}

		// **Der Rückgabewert ist hier nicht die Zusage.** Ein Archiv darf
		// abgelehnt werden, und ein halb ausgepacktes darf Reste hinterlassen
		// — im **Ziel**. Was zählt, steht darunter.
		//
		_, auspackfehler := auspacken(&verdichtet, "", ziel)
		_ = auspackfehler

		nurImZiel(t, draussen, ziel)
		nurGewoehnlichesImZiel(t, ziel)
	})
}

// nurImZiel sichert zu, dass unterhalb von `heim` nichts außerhalb von `ziel`
// entstanden ist.
//
// **Der Trenner gehört an den Vergleich.** `strings.HasPrefix(pfad, ziel)`
// nähme auch `…/zielX` als „im Ziel" an; ein Archiv, das genau dorthin
// schriebe, käme durch. Angestrichen vom Zweitblick über Phase 4.2.
//
// **Geprüft wird die Platte und nicht der Rückgabewert.** Ein Entpacker, der
// „übergangen" meldet und trotzdem schreibt, wäre der schlimmste Fall — und
// den sähe niemand, der nur die Antwort liest.
func nurImZiel(t *testing.T, heim, ziel string) {
	t.Helper()

	fehler := filepath.WalkDir(heim, func(pfad string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if pfad == heim || pfad == ziel ||
			strings.HasPrefix(pfad, ziel+string(filepath.Separator)) {
			return nil
		}

		// **Die Ordner auf dem Weg zum Ziel gehören dazu.**
		if strings.HasPrefix(ziel, pfad+string(filepath.Separator)) {
			return nil
		}

		t.Errorf("es wurde außerhalb des Ziels geschrieben: %s", pfad)

		return nil
	})
	if fehler != nil {
		t.Fatalf("das Zuhause durchgehen: %v", fehler)
	}
}

// nurGewoehnlichesImZiel sichert zu, dass im Ziel kein Sondertyp liegt.
//
// **Ein Symlink ist der Fund, nicht die Datei, auf die er zeigt.** Wer einen
// anlegt, hat den Ausbruch nur auf später verschoben: der nächste Schreiber
// folgt ihm.
//
// **Heute kann keiner der beiden Teile auslösen, und das gehört gesagt.**
// `schreiben` legt jede Datei mit `os.OpenFile(name, …, 0o600)` an: eine
// gewöhnliche Datei, ein fester Modus. Ein Verweis im Archiv landete als
// gewöhnliche Datei, und ein `mode: 0755` im Kopf erreicht die Platte nicht —
// beides nachgesehen am 2026-08-29.
//
// **Gegenprobe, und sie blieb grün**: `ausfuehrbar` auf `false` gesetzt, und
// nichts wurde ausführbar. Das ist kein Fund, sondern die Auskunft, dass diese
// Zusage zweimal gehalten wird — von `ausfuehrbar` und vom festen Modus.
//
// Was die Zeilen hier halten, ist deshalb eine **Regression**: der Tag, an dem
// jemand `os.Symlink` schreibt oder `kopf.FileInfo().Mode()` übernimmt. Ein
// Wächter, der heute nicht auslösen **kann**, ist kein Beleg für heute — und
// das steht hier, damit ihn niemand dafür hält (§10).
//
// **Warum `ausfuehrbar` trotzdem nicht überflüssig ist**: es weist auch über
// die **Endung** ab (`nichtAuspacken`), und eine `.dylib` mit Modus 0600 ist
// weiterhin ladbar.
func nurGewoehnlichesImZiel(t *testing.T, ziel string) {
	t.Helper()

	fehler := filepath.WalkDir(ziel, func(pfad string, eintrag fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if eintrag.IsDir() {
			return nil
		}

		auskunft, lesefehler := eintrag.Info()
		if lesefehler != nil {
			return lesefehler
		}

		if auskunft.Mode()&fs.ModeType != 0 {
			t.Errorf("%s ist keine gewöhnliche Datei, sondern %s",
				pfad, auskunft.Mode())
		}

		if auskunft.Mode().Perm()&0o111 != 0 {
			t.Errorf("%s ist ausführbar: %s", pfad, auskunft.Mode())
		}

		return nil
	})
	if fehler != nil {
		t.Fatalf("das Ziel durchgehen: %v", fehler)
	}
}

// archivsaat baut die tar-Ströme, mit denen der Fuzzer beginnt.
//
// **Jeder ist ein Angriff, den `auspacken` schon kennt** — der Punkt ist
// nicht, dass sie durchkommen, sondern dass der Fuzzer von ihnen aus mutiert.
func archivsaat(f *testing.F) [][]byte {
	f.Helper()

	faelle := []struct {
		kopf   tar.Header
		inhalt string
	}{
		{tar.Header{Name: "paket/datei.txt", Mode: 0o644, Typeflag: tar.TypeReg}, "inhalt"},
		{tar.Header{Name: "../ausbruch.txt", Mode: 0o644, Typeflag: tar.TypeReg}, "raus"},
		{tar.Header{Name: "paket/../../weg.txt", Mode: 0o644, Typeflag: tar.TypeReg}, "raus"},
		{
			tar.Header{
				Name: "paket/" + strings.Repeat("../", 8) + "weit-raus.txt",
				Mode: 0o644, Typeflag: tar.TypeReg,
			},
			"sehr weit raus",
		},
		{tar.Header{Name: "/absolut.txt", Mode: 0o644, Typeflag: tar.TypeReg}, "raus"},
		{tar.Header{Name: "paket/link", Typeflag: tar.TypeSymlink, Linkname: "/etc/passwd"}, ""},
		{tar.Header{Name: "paket/hart", Typeflag: tar.TypeLink, Linkname: "paket/datei.txt"}, ""},
		{tar.Header{Name: "paket/prog", Mode: 0o755, Typeflag: tar.TypeReg}, "#!/bin/sh"},
		{tar.Header{Name: "paket/ordner/", Mode: 0o755, Typeflag: tar.TypeDir}, ""},
		{tar.Header{Name: "paket/gerät", Typeflag: tar.TypeChar, Devmajor: 1, Devminor: 3}, ""},
		{tar.Header{Name: "paket/Ω-datei.txt", Mode: 0o644, Typeflag: tar.TypeReg}, "unicode"},
		{tar.Header{Name: strings.Repeat("a/", 60) + "tief.txt", Mode: 0o644, Typeflag: tar.TypeReg}, "tief"},
	}

	saat := make([][]byte, 0, len(faelle)+1)

	for _, fall := range faelle {
		saat = append(saat, einArchiv(f, fall.kopf, fall.inhalt))
	}

	// **Und eines mit zwei gleichen Namen** — die Frage, wer gewinnt.
	saat = append(saat, zweiEintraege(f))

	return saat
}

// einArchiv baut einen tar-Strom mit einem Eintrag.
//
// **Fehler werden hier gemeldet und nicht verschluckt.** Eine Saat, die nicht
// entsteht, wäre ein Fuzzlauf ohne Ausgangspunkt — und das sähe aus wie ein
// Lauf, der nichts gefunden hat.
func einArchiv(f *testing.F, kopf tar.Header, inhalt string) []byte {
	f.Helper()

	var gebaut bytes.Buffer

	schreiber := tar.NewWriter(&gebaut)
	kopf.Size = int64(len(inhalt))

	if err := schreiber.WriteHeader(&kopf); err != nil {
		f.Fatalf("den Kopf %q schreiben: %v", kopf.Name, err)
	}

	if _, err := schreiber.Write([]byte(inhalt)); err != nil {
		f.Fatalf("den Inhalt zu %q schreiben: %v", kopf.Name, err)
	}

	if err := schreiber.Close(); err != nil {
		f.Fatalf("das Archiv zu %q schließen: %v", kopf.Name, err)
	}

	return gebaut.Bytes()
}

// zweiEintraege baut einen Strom mit demselben Namen zweimal.
func zweiEintraege(f *testing.F) []byte {
	f.Helper()

	var gebaut bytes.Buffer

	schreiber := tar.NewWriter(&gebaut)

	for _, inhalt := range []string{"erst", "dann"} {
		kopf := &tar.Header{
			Name: "paket/gleich.txt", Mode: 0o644,
			Typeflag: tar.TypeReg, Size: int64(len(inhalt)),
		}

		if err := schreiber.WriteHeader(kopf); err != nil {
			f.Fatalf("den zweiten Kopf schreiben: %v", err)
		}

		if _, err := schreiber.Write([]byte(inhalt)); err != nil {
			f.Fatalf("den zweiten Inhalt schreiben: %v", err)
		}
	}

	if err := schreiber.Close(); err != nil {
		f.Fatalf("das zweite Archiv schließen: %v", err)
	}

	return gebaut.Bytes()
}

// kleineGrenzen setzt die Budgets für die Dauer des Laufs herunter.
//
// # Warum das sein muss
//
// **Gemessen am 2026-08-29**: mit den echten Budgets (2 GiB je Eintrag, 4 GiB
// gesamt) schaffte der Fuzzer in 90 Sekunden **45 Ausführungen** — null pro
// Sekunde. Ein verdichteter Strom aus Nullbytes ist ein paar hundert Byte
// groß und packt sich zu Gigabyte aus; jede einzelne Ausführung kopierte
// minutenlang.
//
// **Das ist kein Fund, sondern das Budget bei der Arbeit.** Es begrenzt
// **Bytes** und nicht Zeit, und das ist richtig: ein echtes Vorlagenpaket
// darf groß sein. Was es für einen Fuzzer bedeutet, ist trotzdem: keine
// Abdeckung.
//
// Mit einem Megabyte je Eintrag und vier gesamt fährt derselbe Lauf **über
// zwanzigtausend** Ausführungen — und die Zusagen, um die es geht (kein
// Ausbruch, kein Sondertyp), hängen an keiner Größe.
//
// **Was dadurch ungeprüft bleibt**: das Verhalten **an** der echten Grenze.
// Das hält `verdichtungsbombe_intern_test.go` mit einem gebauten Fall.
func kleineGrenzen(f *testing.F) {
	f.Helper()

	altEintrag, altGesamt, altZahl := hoechstensEintrag, hoechstensGesamt, hoechstensEintraege

	hoechstensEintrag = 1 << 20
	hoechstensGesamt = 4 << 20
	hoechstensEintraege = 64

	f.Cleanup(func() {
		hoechstensEintrag, hoechstensGesamt, hoechstensEintraege = altEintrag, altGesamt, altZahl
	})
}
