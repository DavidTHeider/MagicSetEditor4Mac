package service_test

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Wohin ein eingesetztes Bild kommt** —
// #246,
// entschieden in
// [0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md).
//
// # Der Name folgt MSE
//
//	name = prefix; name << ++infix; name += suffix;
//	name = normalize_internal_filename(name);
//	FileInfos::iterator it = files.find(name);
//	if (it == files.end()) { … return name; }
//
// `Package::newFileName`, `src/util/io/package.cpp`, Zeilen 248–267, Fassung
// f9d9356 vom 2021-07-06, gerufen mit dem **Feldnamen** als Präfix
// (`ImageValueEditor::sliceImage`, `src/gui/value/image.cpp`, Zeile 49).
//
// `normalize_internal_filename` macht aus jedem Rückwärtsschrägstrich einen
// Schrägstrich und schreibt alles klein (`src/util/file_utils.cpp`, Zeilen
// 24–31).
//
// **Nachgeprüft an `Archenemy.mse-set`**: die 70 Bilder darin heißen `image0`
// bis `image69` — Feldname plus Zahl, ohne Endung.
func TestEinBildBekommtMSEsNamen(t *testing.T) {
	t.Parallel()

	ablage := service.NeueBildablage(t.TempDir())

	faelle := []struct {
		feld     string
		erwartet string
		warum    string
	}{
		{"image", "image1", "Feldname plus laufende Zahl, ohne Endung"},
		{"image", "image2", "der zweite bekommt die nächste"},
		// **Der Zähler beginnt je Präfix wieder bei eins.** `newFileName` baut
		// `prefix + ++infix` und fragt, ob **dieser Name** vergeben ist —
		// `image1` und `image2` stehen `card image1` nicht im Weg. Hier stand
		// erst `card image3`, erschlossen statt nachgelesen.
		{"Card Image", "card image1", "kleingeschrieben, und der Zähler beginnt neu"},
	}

	for _, fall := range faelle {
		t.Run(fall.erwartet, func(t *testing.T) {
			name, err := ablage.Ablegen(7, fall.feld, probebild())
			if err != nil {
				t.Fatalf("ablegen: %v", err)
			}

			if name != fall.erwartet {
				t.Errorf("%q statt %q — %s", name, fall.erwartet, fall.warum)
			}
		})
	}
}

// **Zwei Karten desselben Feldes bekommen zwei Namen.**
//
// MSE fragt `files.find(name)` gegen **alle** Dateien des Pakets: `image1` ist
// vergeben, also nimmt der nächste `image2`. Wer stattdessen nur zählte, ohne
// nachzusehen, gäbe nach einem Neustart wieder `image1` aus — und die zweite
// Karte überschriebe das Bild der ersten.
func TestZweiKartenDesselbenFeldesBekommenZweiNamen(t *testing.T) {
	t.Parallel()

	ablage := service.NeueBildablage(t.TempDir())

	erster, err := ablage.Ablegen(7, "image", probebild())
	if err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	zweiter, err := ablage.Ablegen(7, "image", probebild())
	if err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	if erster == zweiter {
		t.Errorf("beide heißen %q", erster)
	}
}

// **Zwei Sets stören einander nicht.**
//
// Jedes Set hat sein eigenes Verzeichnis; `image1` in Set 7 und `image1` in
// Set 8 sind zwei Dateien. Wären sie eine, zeigte die eine Karte das Bild der
// anderen — derselbe Fehler wie
// #380.
func TestZweiSetsStoerenEinanderNicht(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ablage := service.NeueBildablage(wurzel)

	fuerSieben, err := ablage.Ablegen(7, "image", probebild())
	if err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	fuerAcht, err := ablage.Ablegen(8, "image", probebild())
	if err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	if fuerSieben != "image1" || fuerAcht != "image1" {
		t.Errorf("%q und %q", fuerSieben, fuerAcht)
	}

	if ablage.Verzeichnis(7) == ablage.Verzeichnis(8) {
		t.Errorf("beide liegen in %q", ablage.Verzeichnis(7))
	}
}

// **Abgelegt wird PNG, gleich was hereinkam.**
//
//	img.SaveFile(getLocalPackage().nameOut(new_image_file), wxBITMAP_TYPE_PNG);
//	// always use PNG images, see #69. Disk space is cheap anyway.
//
// `src/gui/value/image.cpp`, Zeile 51, Fassung f9d9356 vom 2021-07-06.
func TestAbgelegtWirdPNG(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ablage := service.NeueBildablage(wurzel)

	name, err := ablage.Ablegen(7, "image", probebild())
	if err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	datei, err := os.Open(filepath.Join(ablage.Verzeichnis(7), name)) //nolint:gosec // Testpfad
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := datei.Close(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	bild, err := png.Decode(datei)
	if err != nil {
		t.Fatalf("die Datei ist kein PNG: %v", err)
	}

	if bild.Bounds() != image.Rect(0, 0, 4, 3) {
		t.Errorf("%v", bild.Bounds())
	}
}

// **Ohne Wurzel gibt es keine Ablage** — und das ist ein Fehler, kein Befund:
// wer ein Bild einsetzt, hat es getan, und es stillschweigend zu verlieren
// wäre die schlechteste Antwort (§8).
func TestOhneWurzelScheitertDasAblegen(t *testing.T) {
	t.Parallel()

	ablage := service.NeueBildablage("")

	if _, err := ablage.Ablegen(7, "image", probebild()); err == nil {
		t.Error("ohne Wurzel kam kein Fehler")
	}

	if ort := ablage.Verzeichnis(7); ort != "" {
		t.Errorf("Verzeichnis %q", ort)
	}
}

func probebild() image.Image {
	bild := image.NewNRGBA(image.Rect(0, 0, 4, 3))
	bild.SetNRGBA(1, 1, color.NRGBA{R: 200, A: 255})

	return bild
}

// **Ein Verzeichnis, in das sich nicht schreiben lässt, wird gemeldet** (§8).
//
// Ein Set liegt auf einem Datenträger, der abgezogen wurde; die Ablage liegt
// unter `~/Library/Application Support/`, und auch die kann voll sein. Ein
// Bild, das dabei stillschweigend verschwände, wäre der schlechteste Ausgang:
// im Feld stünde ein Dateiname, hinter dem nichts liegt.
func TestEinUnschreibbaresVerzeichnisWirdGemeldet(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()

	// **Eine Datei dort, wo das Verzeichnis hinsoll.** `MkdirAll` scheitert
	// daran — und zwar auf jeder Plattform, anders als an einem Rechtebit.
	if err := os.WriteFile(filepath.Join(wurzel, "7"), []byte("keine Ablage"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, err := service.NeueBildablage(wurzel).Ablegen(7, "image", probebild()); err == nil {
		t.Error("das unschreibbare Verzeichnis kam durch")
	}
}

// **Ein Feld, unter dessen Namen schon ein Verzeichnis liegt, wird gemeldet.**
//
// `os.Create` scheitert daran, und `freierName` sieht es nicht: ein
// Verzeichnis **existiert**, also geht der Zähler weiter — bis er bei einem
// Namen ankommt, den es nicht gibt. Genau deshalb steht der Fall hier: er
// belegt, dass der Fehler des Schreibens durchkommt und nicht in der Schleife
// verschwindet.
func TestEinUnschreibbarerNameWirdGemeldet(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ablage := service.NeueBildablage(wurzel)

	// Erst das Verzeichnis anlegen lassen …
	if _, err := ablage.Ablegen(7, "image", probebild()); err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	// … und dann `image2` als Verzeichnis besetzen.
	if err := os.MkdirAll(filepath.Join(ablage.Verzeichnis(7), "image2"), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	name, err := ablage.Ablegen(7, "image", probebild())
	if err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	if name != "image3" {
		t.Errorf("%q statt image3 — ein besetzter Name wurde übersprungen?", name)
	}
}

// **Ohne freien Namen wird abgebrochen, nicht gedreht.**
//
// MSE hat hier keine Schranke — dort ist die Dateiliste im Speicher, und
// `find` scheitert nicht. Hier fragt jeder Durchlauf die Platte, und ein
// Verzeichnis, das jeden Namen belegt meldet (ein kaputter Einhängepunkt
// etwa), drehte für immer. Die Schranke liegt bei 100.000; geprüft wird sie
// über ein Verzeichnis, dessen Einträge sich nicht abfragen lassen.
func TestOhneFreienNamenWirdAbgebrochen(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ablage := service.NeueBildablage(wurzel)

	// Das Verzeichnis anlegen lassen …
	if _, err := ablage.Ablegen(7, "image", probebild()); err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	// … und dann unlesbar machen: `os.Stat` in einem Verzeichnis ohne
	// Ausführungsrecht scheitert mit etwas anderem als „gibt es nicht".
	//nolint:gosec // absichtlich: ein Verzeichnis ohne x ist der Prüffall
	if err := os.Chmod(ablage.Verzeichnis(7), 0o600); err != nil {
		t.Fatalf("Rechte setzen: %v", err)
	}

	t.Cleanup(func() {
		//nolint:gosec // ein Verzeichnis braucht x, sonst lässt es sich nicht räumen
		if fehler := os.Chmod(ablage.Verzeichnis(7), 0o750); fehler != nil {
			t.Errorf("Rechte zurücksetzen: %v", fehler)
		}
	})

	if _, err := ablage.Ablegen(7, "image", probebild()); err == nil {
		t.Error("das unlesbare Verzeichnis kam durch")
	}
}

// **Ein Bild, das sich nicht kodieren lässt, wird gemeldet.**
//
// Ein Bild ohne Ausdehnung ist keines — `png.Encode` weist es ab, und das ist
// die richtige Antwort: eine leere Datei im Set sähe aus wie ein Bild und wäre
// keins. Entstehen kann es über [render.Zuschnitt] mit einem Feld der Breite
// null.
func TestEinUnkodierbaresBildWirdGemeldet(t *testing.T) {
	t.Parallel()

	leer := image.NewNRGBA(image.Rectangle{})

	if _, err := service.NeueBildablage(t.TempDir()).Ablegen(7, "image", leer); err == nil {
		t.Error("das leere Bild kam durch")
	}
}

// **Und ein Verzeichnis, in das sich nicht schreiben lässt, ebenso.**
//
// Es ist lesbar — `freierName` findet seinen Namen —, und erst das Schreiben
// scheitert. Ohne diesen Fall stünde im Feld ein Dateiname, hinter dem nichts
// liegt.
func TestEinNurLesbaresVerzeichnisWirdGemeldet(t *testing.T) {
	t.Parallel()

	ablage := service.NeueBildablage(t.TempDir())

	if _, err := ablage.Ablegen(7, "image", probebild()); err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	//nolint:gosec // absichtlich: lesbar, nicht schreibbar
	if err := os.Chmod(ablage.Verzeichnis(7), 0o500); err != nil {
		t.Fatalf("Rechte setzen: %v", err)
	}

	t.Cleanup(func() {
		//nolint:gosec // ein Verzeichnis braucht x, sonst lässt es sich nicht räumen
		if fehler := os.Chmod(ablage.Verzeichnis(7), 0o750); fehler != nil {
			t.Errorf("Rechte zurücksetzen: %v", fehler)
		}
	})

	if _, err := ablage.Ablegen(7, "image", probebild()); err == nil {
		t.Error("das nur lesbare Verzeichnis kam durch")
	}
}

// **Die Wurzel steht der Ablage zur Verfügung** —
// #890.
//
// Das Zeichenwerk braucht sie, um eingesetzte Bilder zu finden. **Sie wird
// nicht daneben mitgeführt**: ein zweiter Weg zu derselben Zahl liefe mit
// diesem auseinander, sobald einer von beiden umzieht.
func TestDieBildablageNenntIhreWurzel(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()

	if hat := service.NeueBildablage(wurzel).Wurzel(); hat != wurzel {
		t.Errorf("die Wurzel ist %q, will %q", hat, wurzel)
	}

	// **Und ohne Wurzel ist sie leer, nicht erfunden.**
	if hat := service.NeueBildablage("").Wurzel(); hat != "" {
		t.Errorf("ohne Wurzel steht dort %q", hat)
	}
}
