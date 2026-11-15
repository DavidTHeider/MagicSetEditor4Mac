package msepaket_test

// **Ein Installer von fremder Hand** — QA-06b
// (#1967),
// Fläche 3 aus `docs/FUZZZIELE.md`.
//
// # Warum er ein eigenes Ziel braucht
//
// **Ein Installer ist das Paket, das jemand weiterreicht** — per Nachricht
// oder Download, und [msepaket.Installer.Auspacken] schreibt seine Einträge
// auf die Platte. `FuzzBeschaedigtesPaket` erreicht den Archivleser, aber
// weder die Zuordnung der Einträge zu den Paketen noch das Schreiben.
//
// **Das Archiv wird gebaut und nicht geraten**: der Fuzzer wählt Hauptdatei,
// Eintragsnamen, Inhalte und Verweise; ein ZIP aus Zufallsbytes käme fast nie
// bis zur Zuordnung.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger
//   - ein Installer **und** ein Fehler, oder keins von beidem
//   - zweimal dieselbe Datei, zwei Beschreibungen
//   - ein Befund ohne Meldung
//   - **ein angenommener Paketname, der kein einzelnes Glied ist**, keine
//     Vorlagenendung trägt oder ein Set ist
//   - **zwei angenommene Namen, die die Zuordnung für denselben hält** —
//     zugeordnet wird über [strings.EqualFold], also gilt dieselbe Gleichheit
//   - nach dem Auspacken: **etwas außerhalb des Ziels**, ein Verweis oder
//     etwas anderes als Datei und Ordner im Ziel, **eine Datei, deren Bytes
//     kein Eintrag unter diesem Paket trägt**, andere Bytes auf der Platte als
//     gemeldet, mehr als der Deckel bei gemeldetem Erfolg
//
// # Was nicht zugesichert wird
//
// **Dass ein abgebrochenes Auspacken nichts hinterlässt** — bei
// [msepaket.ErrInstallerZuGross] liegt die Datei, an der es abbrach, mit
// einem Byte über dem Deckel da. Den Bauort räumt der Aufrufer
// (`service.Paketverwalter.einlesenEines`), und das ist seine Zusage.
//
//	go test ./internal/msepaket -run Fuzz -fuzz FuzzInstaller -fuzztime 2m

import (
	"archive/zip"
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// installerkopf ist eine Hauptdatei mit zwei Paketen, wie MSE2 sie schreibt.
const installerkopf = "mse_version: 0.3.7\nversion: 0.0.0\n" +
	"package:\n\tname: probe.mse-game\n\tversion: 2026-09-01\n\tshort_name: Probe\n" +
	"package:\n\tname: probe-m.mse-style\n\tdepends_on: probe.mse-game 2026-09-01\n"

// **Was ein Installer auspackt, liegt im Ziel und trägt die Bytes seines
// Eintrags.**
func FuzzInstaller(f *testing.F) {
	for _, saat := range []struct {
		kopf, namen string
		inhalt      []byte
		deckel      uint16
	}{
		{installerkopf, "probe.mse-game/game\nprobe.mse-game/bilder/a.png\nprobe-m.mse-style/style", []byte("mse version: 2.0.0"), 4096},
		{installerkopf, "probe.mse-game/../draussen\nPROBE.mse-game/x\n~probe.mse-game/verweis", []byte("../../etc"), 64},
		{installerkopf, "probe.mse-game/a\nprobe.mse-game/a\nprobe.mse-game/a/b", []byte("doppelt"), 4096},
		{installerkopf, "probe.mse-game/gross", bytes.Repeat([]byte("x"), 300), 10},
		{"package:\n\tname: ../x.mse-game\npackage:\n\tname: a.mse-set\npackage:\n\tname: .mse-game\n", "x.mse-game/y", nil, 0},
		{"package:\n\tname: a.mse-game\npackage:\n\tname: A.mse-game\n", "a.mse-game/b\nA.mse-game/c", []byte("z"), 100},
		{"package:\n\tname: probe.mse-game\n\tversion: kaputt\n", "probe.mse-game/\\..\\x", []byte("w"), 100},
		{"", "", nil, 0},
	} {
		f.Add(saat.kopf, saat.namen, saat.inhalt, saat.deckel)
	}

	f.Fuzz(func(t *testing.T, kopf, namen string, inhalt []byte, deckel uint16) {
		eintraege := installereintraege(kopf, namen, inhalt)

		ort := filepath.Join(t.TempDir(), "probe.mse-installer")
		if !installerHinlegen(t, ort, eintraege) {
			return
		}

		installer := installerGeoeffnet(t, ort)
		if installer == nil {
			return
		}

		defer func() { _ = installer.Schliessen() }() //nolint:errcheck // nur gelesen

		for _, paket := range installer.Pakete {
			pruefeAuspacken(t, installer, paket.Name, eintraege, int64(deckel))
		}
	})
}

// installereintrag ist ein Eintrag des gebauten Archivs.
type installereintrag struct {
	name    string
	inhalt  []byte
	verweis bool
}

// installereintraege macht aus den Fuzzwerten die Einträge: die Hauptdatei,
// dann je Zeile von `namen` einen — `~` vorn heißt Verweis.
func installereintraege(kopf, namen string, inhalt []byte) []installereintrag {
	eintraege := []installereintrag{{name: msepaket.Installerdatei, inhalt: []byte(kopf)}}

	if namen == "" {
		return eintraege
	}

	for stelle, zeile := range strings.SplitN(namen, "\n", 24) {
		name, verweis := strings.CutPrefix(zeile, "~")

		// **Jeder Eintrag bekommt andere Bytes** — sonst sagte der Vergleich
		// unten nicht, **welcher** Eintrag auf der Platte liegt.
		bytesDesEintrags := append([]byte{byte(stelle)}, inhalt...)

		// **Ein Name auf `/` ist ein Ordner und trägt keine** — `archive/zip`
		// weist sie ab („write to directory"); gefunden von der vollen
		// Kampagne (C02, #2227, `FuzzInstaller/9c6f04f43a8d5ff1`), ein Fund
		// an dieser Vorrichtung und nicht am Leser. Der Ordnereintrag bleibt
		// eine Eingabe wie jede andere.
		if strings.HasSuffix(name, "/") {
			bytesDesEintrags = nil
		}

		eintraege = append(eintraege, installereintrag{
			name: name, verweis: verweis, inhalt: bytesDesEintrags,
		})
	}

	return eintraege
}

// installerHinlegen schreibt das Archiv; `false` heißt: `archive/zip` nahm
// einen Namen nicht an, und es gibt nichts zu prüfen.
func installerHinlegen(t *testing.T, ort string, eintraege []installereintrag) bool {
	t.Helper()

	var puffer bytes.Buffer

	archiv := zip.NewWriter(&puffer)

	for _, eintrag := range eintraege {
		kopf := &zip.FileHeader{Name: eintrag.name, Method: zip.Deflate}
		if eintrag.verweis {
			kopf.SetMode(fs.ModeSymlink | 0o777)
		}

		schreiber, fehler := archiv.CreateHeader(kopf)
		if fehler != nil {
			return false
		}

		if _, fehler := schreiber.Write(eintrag.inhalt); fehler != nil {
			t.Fatalf("in den Puffer schreiben: %v", fehler)
		}
	}

	if fehler := archiv.Close(); fehler != nil {
		return false
	}

	if fehler := os.WriteFile(ort, puffer.Bytes(), 0o600); fehler != nil {
		t.Fatalf("hinlegen: %v", fehler)
	}

	return true
}

// installerGeoeffnet öffnet zweimal und hält die Zusagen über die Antwort.
func installerGeoeffnet(t *testing.T, ort string) *msepaket.Installer {
	t.Helper()

	installer, fehler := msepaket.OeffneInstaller(ort)
	nochmal, nochmalFehler := msepaket.OeffneInstaller(ort)

	if (fehler == nil) != (nochmalFehler == nil) {
		t.Fatalf("dieselbe Datei einmal geöffnet und einmal nicht: %v / %v", fehler, nochmalFehler)
	}

	if fehler != nil {
		if installer != nil {
			t.Errorf("abgelehnt **und** ein Installer zurück: %v", fehler)
		}

		if fehler.Error() != nochmalFehler.Error() {
			t.Errorf("zwei Meldungen: %q gegen %q", fehler, nochmalFehler)
		}

		return nil
	}

	defer func() { _ = nochmal.Schliessen() }() //nolint:errcheck // nur gelesen

	if installer == nil {
		t.Fatal("kein Fehler und kein Installer — das ist keine Antwort")
	}

	if !reflect.DeepEqual(installer.Pakete, nochmal.Pakete) || !reflect.DeepEqual(installer.Befunde, nochmal.Befunde) {
		t.Errorf("dieselbe Datei, zwei Beschreibungen: %+v gegen %+v", installer.Pakete, nochmal.Pakete)
	}

	for _, befund := range installer.Befunde {
		if befund.Meldung == "" {
			t.Error("ein Befund ohne Meldung")
		}
	}

	for stelle, paket := range installer.Pakete {
		pruefePaketname(t, paket.Name)

		for _, frueher := range installer.Pakete[:stelle] {
			if strings.EqualFold(frueher.Name, paket.Name) {
				t.Errorf("%q und %q sind für die Zuordnung derselbe Name", frueher.Name, paket.Name)
			}
		}
	}

	return installer
}

// pruefePaketname hält einen angenommenen Namen gegen das, was einer ist.
func pruefePaketname(t *testing.T, name string) {
	t.Helper()

	art := msepaket.ArtAusName(name)

	if name == "" || strings.ContainsAny(name, `/\`) || name == "." || name == ".." ||
		path.Clean(name) != name || art == msepaket.ArtUnbekannt || art == msepaket.ArtSet {
		t.Errorf("%q ist als Paketname angenommen", name)
	}
}

// pruefeAuspacken packt ein Paket in ein frisches Ziel und sieht nach, was dort
// liegt — mit `os.Lstat` und dem eigenen Leser, nicht mit dem geprüften.
func pruefeAuspacken(t *testing.T, installer *msepaket.Installer, name string,
	eintraege []installereintrag, deckel int64,
) {
	t.Helper()

	rahmen := t.TempDir()
	ziel := filepath.Join(rahmen, "ziel")

	if fehler := os.Mkdir(ziel, 0o750); fehler != nil {
		t.Fatalf("das Ziel anlegen: %v", fehler)
	}

	benutzt, fehler := installer.Auspacken(name, ziel, deckel)

	if fehler == nil && benutzt > deckel {
		t.Errorf("%q: %d Byte bei einem Deckel von %d, und kein Fehler", name, benutzt, deckel)
	}

	if fehler != nil && fehler.Error() == "" {
		t.Errorf("%q: stumm abgelehnt", name)
	}

	if daneben, _ := os.ReadDir(rahmen); len(daneben) != 1 { //nolint:errcheck // leer heißt auch daneben
		t.Errorf("%q: neben dem Ziel liegt etwas — %d Einträge", name, len(daneben))
	}

	var aufDerPlatte int64

	gang := filepath.WalkDir(ziel, func(weg string, _ fs.DirEntry, err error) error {
		if err != nil || weg == ziel {
			return err
		}

		angaben, err := os.Lstat(weg)
		if err != nil {
			return err
		}

		if angaben.IsDir() {
			return nil
		}

		if !angaben.Mode().IsRegular() {
			t.Errorf("%q: %s ist weder Datei noch Ordner (%v)", name, weg, angaben.Mode())

			return nil
		}

		aufDerPlatte += angaben.Size()

		relativ, _ := filepath.Rel(ziel, weg) //nolint:errcheck // unter ziel gefunden
		pruefeDateiinhalt(t, name, filepath.ToSlash(relativ), weg, eintraege, fehler == nil)

		return nil
	})
	if gang != nil {
		t.Fatalf("%q: das Ziel durchgehen: %v", name, gang)
	}

	if aufDerPlatte != benutzt {
		t.Errorf("%q: %d Byte gemeldet, %d liegen da (%v)", name, benutzt, aufDerPlatte, fehler)
	}

	if errors.Is(fehler, msepaket.ErrInstallerZuGross) && benutzt <= deckel {
		t.Errorf("%q: „zu groß“ bei %d Byte und einem Deckel von %d", name, benutzt, deckel)
	}
}

// pruefeDateiinhalt sucht den Eintrag, aus dem eine Datei stammt.
//
// **Bei Erfolg genau seine Bytes, bei Abbruch ihr Anfang** — siehe Kopf.
func pruefeDateiinhalt(t *testing.T, paket, relativ, weg string,
	eintraege []installereintrag, ganz bool,
) {
	t.Helper()

	liegt, fehler := os.ReadFile(weg) //nolint:gosec // unter t.TempDir()
	if fehler != nil {
		t.Fatalf("%s lesen: %v", weg, fehler)
	}

	vorn := len(paket) + 1

	for _, eintrag := range eintraege {
		if eintrag.verweis || len(eintrag.name) <= vorn ||
			!strings.EqualFold(eintrag.name[:vorn], paket+"/") || eintrag.name[vorn:] != relativ {
			continue
		}

		if bytes.Equal(liegt, eintrag.inhalt) || (!ganz && bytes.HasPrefix(eintrag.inhalt, liegt)) {
			return
		}
	}

	t.Errorf("%s/%s trägt %d Byte, die kein Eintrag unter diesem Paket hat", paket, relativ, len(liegt))
}
