package austausch

// **Manifest und Unterschrift haben eigene Grenzen, und sie greifen vor dem
// Schreiben** — SEC-03
// (#2138).
//
// Beide wurden unter den allgemeinen Grenzen einer Sicherung herausgelegt —
// 16 GiB je Teil — und danach ganz eingelesen, **bevor** die Unterschrift
// geprüft war. Ein Umschlag von wenigen Bytes, dessen äußeres Manifest für
// `austausch.json` 16 GiB ankündigte, kam bis zum Schreiben.

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// umschlagMit packt einen Umschlag mit einem Objekt, dessen Manifest auf
// `manifestgroesse` Bytes aufgefüllt ist und dessen Unterschrift
// `unterschrift` ist — `nil` heißt: die echte.
func umschlagMit(t *testing.T, manifestgroesse int, unterschrift []byte) string {
	t.Helper()

	ich, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatal(fehler)
	}

	inhalt := []byte("{}")
	summe := sha256.Sum256(inhalt)

	roh := []byte(`{"fassung":1,"urheber":"` + string(ich.Kennung()) + `","objekte":[` +
		`{"art":"karte","name":"a.json","sha256":"` + hex.EncodeToString(summe[:]) +
		`","groesse":` + strconv.Itoa(len(inhalt)) + `}]}`)

	// **Leerraum ist gültiges JSON** — so wiegt das Manifest genau so viel,
	// wie der Fall verlangt, und bleibt deutbar.
	if manifestgroesse > len(roh) {
		roh = append(roh, bytes.Repeat([]byte(" "), manifestgroesse-len(roh))...)
	}

	if unterschrift == nil {
		unterschrift = ich.Signiere(roh)
	}

	ort := filepath.Join(t.TempDir(), "probe"+Endung)
	if packfehler := sicherung.PackenAls(ort, aeusseresManifest, []sicherung.Teil{
		{Name: "a.json", Inhalt: inhalt},
		{Name: Manifestname, Inhalt: roh},
		{Name: Unterschriftsname, Inhalt: unterschrift},
	}, sicherung.Angaben{}); packfehler != nil {
		t.Fatal(packfehler)
	}

	return ort
}

func TestDasManifestHatEineEigeneGrenze(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		groesse int
		ueber   bool
	}{
		"eins unter der Grenze": {groesse: int(HoechstensManifest) - 1},
		"genau auf der Grenze":  {groesse: int(HoechstensManifest)},
		"eins darüber":          {groesse: int(HoechstensManifest) + 1, ueber: true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ziel := filepath.Join(t.TempDir(), "ziel")
			_, fehler := Lies(umschlagMit(t, fall.groesse, nil), ziel)

			if ueber := errors.Is(fehler, sicherung.ErrUeberDemDeckel); ueber != fall.ueber {
				t.Fatalf("über dem Deckel: %v, erwartet %v — %v", ueber, fall.ueber, fehler)
			}

			if !fall.ueber && fehler != nil {
				t.Errorf("ein Manifest innerhalb der Grenze: %v", fehler)
			}
		})
	}
}

func TestDieUnterschriftHatEineEigeneGrenze(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "ziel")
	_, fehler := Lies(umschlagMit(t, 0, make([]byte, HoechstensUnterschrift+1)), ziel)

	if !errors.Is(fehler, sicherung.ErrUeberDemDeckel) {
		t.Fatalf("eine Unterschrift über der Grenze: %v", fehler)
	}

	// **Genau auf der Grenze** ist sie keine gültige Unterschrift — aber das
	// sagt die Prüfung der Unterschrift, nicht der Deckel.
	_, fehler = Lies(umschlagMit(t, 0, make([]byte, HoechstensUnterschrift)), filepath.Join(t.TempDir(), "ziel"))
	if errors.Is(fehler, sicherung.ErrUeberDemDeckel) || fehler == nil {
		t.Errorf("auf der Grenze: %v", fehler)
	}
}

// **Eine riesige Ankündigung bei winzigem Inhalt** wird abgewiesen, bevor
// irgendetwas geschrieben ist — **in jeder Reihenfolge** (Zweitblick über E,
// Befund 3): stand das zu große Teil hinter einem richtigen, war das
// richtige schon herausgelegt, als die Grenze griff.
func TestEineRiesigeAnkuendigungWirdVorDemSchreibenAbgewiesen(t *testing.T) {
	t.Parallel()

	leer := sha256.Sum256([]byte("{}"))
	manifest := `{"name":"` + Manifestname + `","groesse":2,"sha256":"` + hex.EncodeToString(leer[:]) + `"}`
	riesig := `{"name":"` + Manifestname + `","groesse":17179869184,"sha256":"00"}`
	unterschrift := `{"name":"` + Unterschriftsname + `","groesse":64,"sha256":"00"}`
	zuLang := `{"name":"` + Unterschriftsname + `","groesse":1025,"sha256":"00"}`

	for name, teile := range map[string]string{
		"das erste Teil zu groß":  riesig + "," + unterschrift,
		"das zweite Teil zu groß": manifest + "," + zuLang,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ort := filepath.Join(t.TempDir(), "riese"+Endung)

			var puffer bytes.Buffer
			packer := zip.NewWriter(&puffer)

			for name, inhalt := range map[string]string{
				aeusseresManifest: `{"teile":[` + teile + `]}`,
				Manifestname:      "{}",
				Unterschriftsname: "x",
			} {
				schreiber, fehler := packer.Create(name)
				if fehler != nil {
					t.Fatal(fehler)
				}

				if _, fehler := schreiber.Write([]byte(inhalt)); fehler != nil {
					t.Fatal(fehler)
				}
			}

			if fehler := packer.Close(); fehler != nil {
				t.Fatal(fehler)
			}

			if fehler := os.WriteFile(ort, puffer.Bytes(), 0o600); fehler != nil {
				t.Fatal(fehler)
			}

			ziel := filepath.Join(t.TempDir(), "ziel")
			if _, fehler := Lies(ort, ziel); !errors.Is(fehler, sicherung.ErrUeberDemDeckel) {
				t.Fatalf("eine riesige Ankündigung: %v", fehler)
			}

			eintraege, lesefehler := os.ReadDir(filepath.Join(ziel, ".pruefung"))
			if lesefehler != nil && !errors.Is(lesefehler, fs.ErrNotExist) {
				t.Fatal(lesefehler)
			}

			if len(eintraege) != 0 {
				t.Errorf("vor der Abweisung wurde geschrieben: %v", eintraege)
			}
		})
	}
}
