package austausch_test

// **Die Quarantäne** — Werkstück N4 von Phase 5.4
// (#1457).
//
// # Was hier gemessen wird
//
// **Dass ein Umschlag abgewiesen wird, bevor etwas auf der Platte steht.**
// #1457 verlangt es namentlich: „Archivbomben, Pfadtraversal, Symlink,
// übergroße Bilder und Hashanfragen außerhalb des Manifests werden mit selbst
// erstellten Fixtures verworfen."
//
// **Selbst erstellt, und das ist keine Formsache** (§11): fremde Vorlagen
// liegen hier nicht, und ein Umschlag mit einem Pfadtraversal ist ohnehin
// nichts, was man irgendwo herunterlädt.

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/austausch"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// **Ein Name, der aus dem Umschlag herausführt, wird beim Schreiben
// abgewiesen.**
//
// **Beim Schreiben schon** — wer einen solchen Umschlag baute, hätte ihn
// jemandem geschickt.
func TestEinNameDerHerausfuehrtWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, weg := range map[string]string{
		"Rückwärtsschritt":      "../draussen.json",
		"verstecktes Rückwärts": "a/../../draussen.json",
		"absolut":               "/etc/passwd",
		"Windowstrenner":        `a\b.json`,
		"Laufwerk":              "C:datei.json",
		"leer":                  "",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := austausch.Schreibe(
				filepath.Join(t.TempDir(), "x"+austausch.Endung),
				[]austausch.Inhalt{{
					Art: austausch.ArtKarte, Name: weg, Bytes: []byte("x"),
				}}, eineIdentitaet(t))

			if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("%q kam durch: %v", weg, fehler)
			}
		})
	}
}

// **Ein Name steht kanonisch da** — QA-06b
// (#1967).
//
// Der Saatlauf von `FuzzUmschlag` nahm ein Objekt namens `.` an: ausgepackt
// wurde es **an die Stelle des Verzeichnisses**, und die Sendung nannte danach
// eine Datei als ihr Verzeichnis. `a/../b.json` führt nicht hinaus und nennt
// trotzdem eine andere Datei, als dasteht — dieselbe Klasse, die der
// Zweitblick über Phase 3 an `msepaket.ZerlegePaketpfad` geschlossen hat.
func TestEinNameStehtKanonischDa(t *testing.T) {
	t.Parallel()

	for name, weg := range map[string]string{
		"der Umschlag selbst":        ".",
		"ein Umweg":                  "a/../b.json",
		"ein Punktglied":             "a/./b.json",
		"ein doppelter Schrägstrich": "a//b.json",
		"ein Schrägstrich am Ende":   "a/",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := austausch.Schreibe(
				filepath.Join(t.TempDir(), "x"+austausch.Endung),
				[]austausch.Inhalt{{
					Art: austausch.ArtKarte, Name: weg, Bytes: []byte("x"),
				}}, eineIdentitaet(t))

			if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("%q kam durch: %v", weg, fehler)
			}
		})
	}
}

// **Ein zu tiefer und ein zu langer Name ebenso.**
func TestZuTiefUndZuLangWerdenAbgewiesen(t *testing.T) {
	t.Parallel()

	tief := ""
	for range austausch.HoechstensTiefe + 1 {
		tief += "a/"
	}

	lang := ""
	for range austausch.HoechstensName + 1 {
		lang += "a"
	}

	for name, weg := range map[string]string{"zu tief": tief + "x", "zu lang": lang} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := austausch.Schreibe(
				filepath.Join(t.TempDir(), "x"+austausch.Endung),
				[]austausch.Inhalt{{
					Art: austausch.ArtKarte, Name: weg, Bytes: []byte("x"),
				}}, eineIdentitaet(t))

			if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("kam durch: %v", fehler)
			}
		})
	}
}

// **Eine gefälschte Unterschrift trägt nicht.**
//
// # Wie der Umschlag gefälscht wird
//
// **Indem das Manifest geändert wird und die Unterschrift stehen bleibt** —
// genau das, was ein Mitleser täte. Gebaut wird der gefälschte Umschlag hier,
// Eintrag für Eintrag; ein echter, den jemand nachträglich ändert, sähe
// genauso aus.
func TestEineGefaelschteUnterschriftTraegtNicht(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	echt := austausch.Manifest{
		Fassung: 1, Urheber: string(ich.Kennung()), Erstellt: time.Now().UTC(),
		Objekte: []austausch.Objekt{{
			Art: austausch.ArtKarte, Name: "a.json", Bezeichnung: "A",
			Summe:   "0000000000000000000000000000000000000000000000000000000000000000",
			Groesse: 1,
		}},
	}

	roh, fehler := json.Marshal(echt)
	if fehler != nil {
		t.Fatalf("Manifest: %v", fehler)
	}

	// **Die Unterschrift gilt dem echten Manifest** — und der Umschlag trägt
	// ein anderes.
	unterschrift := ich.Signiere(roh)

	gefaelscht := echt
	gefaelscht.Objekte[0].Bezeichnung = "etwas ganz anderes"

	falschRoh, fehler := json.Marshal(gefaelscht)
	if fehler != nil {
		t.Fatalf("Manifest: %v", fehler)
	}

	pfad := baueUmschlag(t, map[string][]byte{
		"austausch.json": falschRoh,
		"austausch.sig":  unterschrift,
		"a.json":         []byte("x"),
	})

	if _, fehler := austausch.Lies(pfad, t.TempDir()); !errors.Is(
		fehler, domain.ErrUmschlagUnterschrift) {
		t.Errorf("erwartet war ErrUmschlagUnterschrift, kam: %v", fehler)
	}
}

// **Ein geänderter Inhalt fällt an seiner Prüfsumme auf.**
func TestEinGeaenderterInhaltFaelltAuf(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	pfad := einUmschlag(t, ich, austausch.Inhalt{
		Art: austausch.ArtKarte, Name: "a.json", Bezeichnung: "A",
		Bytes: []byte("echt"),
	})

	sendung, fehler := austausch.Lies(pfad, t.TempDir())
	if fehler != nil {
		t.Fatalf("Lies: %v", fehler)
	}

	// **Denselben Umschlag noch einmal bauen, mit anderem Inhalt und
	// demselben Manifest** — das ist der Fall „unterwegs geändert".
	manifest, fehler := json.Marshal(sendung.Manifest)
	if fehler != nil {
		t.Fatalf("Manifest: %v", fehler)
	}

	verfaelscht := baueUmschlag(t, map[string][]byte{
		"austausch.json": manifest,
		"austausch.sig":  ich.Signiere(manifest),
		"a.json":         []byte("gefälscht"),
	})

	if _, fehler := austausch.Lies(verfaelscht, t.TempDir()); !errors.Is(
		fehler, domain.ErrUmschlagInhalt) {
		t.Errorf("erwartet war ErrUmschlagInhalt, kam: %v", fehler)
	}
}

// **Eine fremde Fassung wird abgewiesen, nicht heruntergehandelt.**
func TestEineFremdeFassungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	ich := eineIdentitaet(t)

	manifest := austausch.Manifest{
		Fassung: 99, Urheber: string(ich.Kennung()),
		Objekte: []austausch.Objekt{{
			Art: austausch.ArtKarte, Name: "a.json",
			Summe:   "0000000000000000000000000000000000000000000000000000000000000000",
			Groesse: 1,
		}},
	}

	roh, fehler := json.Marshal(manifest)
	if fehler != nil {
		t.Fatalf("Manifest: %v", fehler)
	}

	pfad := baueUmschlag(t, map[string][]byte{
		"austausch.json": roh,
		"austausch.sig":  ich.Signiere(roh),
		"a.json":         []byte("x"),
	})

	if _, fehler := austausch.Lies(pfad, t.TempDir()); !errors.Is(
		fehler, domain.ErrUmschlagFassung) {
		t.Errorf("erwartet war ErrUmschlagFassung, kam: %v", fehler)
	}
}

// baueUmschlag schreibt ein ZIP von Hand — für die Fälle, die `Schreibe`
// gar nicht erzeugen kann.
func baueUmschlag(t *testing.T, eintraege map[string][]byte) string {
	t.Helper()

	return baueUmschlagMitLuecke(t, eintraege)
}

// baueUmschlagMitLuecke nennt zusätzlich Namen im **äußeren** Manifest, die
// im ZIP nicht liegen.
//
// **Für den Fall, den `baueUmschlag` nicht herstellen kann**: dort leitet
// sich das äußere Manifest aus den Einträgen ab, und damit kennt es nie
// einen, den es nicht gibt.
func baueUmschlagMitLuecke(
	t *testing.T, eintraege map[string][]byte, luecken ...string,
) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "gebaut"+austausch.Endung)

	datei, fehler := os.Create(pfad) //#nosec G304 -- t.TempDir()
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // im Lauf

	schreiber := zip.NewWriter(datei)

	// **Das äußere Manifest gehört dazu** — ohne es hält `sicherung` die
	// Datei gar nicht erst für einen Umschlag, und der Lauf prüfte dann
	// etwas anderes als gemeint.
	angaben := teileFuer(eintraege)
	for _, luecke := range luecken {
		angaben = append(angaben, sicherung.Teilangabe{
			Name: luecke, Groesse: 1, Summe: eineSumme,
		})
	}

	aussen := sicherung.Beschreibung{
		Angaben: sicherung.Angaben{
			Fassung: "probe", Erstellt: time.Now().UTC(),
		},
		Teile: angaben,
	}

	aussenRoh, fehler := json.Marshal(aussen)
	if fehler != nil {
		t.Fatalf("äußeres Manifest: %v", fehler)
	}

	alles := map[string][]byte{"umschlag.json": aussenRoh}
	for name, inhalt := range eintraege {
		alles[name] = inhalt
	}

	for name, inhalt := range alles {
		eintrag, fehler := schreiber.Create(name)
		if fehler != nil {
			t.Fatalf("Eintrag %q: %v", name, fehler)
		}

		if _, fehler := eintrag.Write(inhalt); fehler != nil {
			t.Fatalf("schreiben: %v", fehler)
		}
	}

	if fehler := schreiber.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	return pfad
}

// teileFuer baut die Angaben des äußeren Manifests.
//
// **Mit echten Prüfsummen** — das äußere Manifest prüft `sicherung`, und
// falsche Summen dort ließen den Lauf an der falschen Stelle scheitern.
func teileFuer(eintraege map[string][]byte) []sicherung.Teilangabe {
	angaben := make([]sicherung.Teilangabe, 0, len(eintraege))

	for name, inhalt := range eintraege {
		summe := sha256.Sum256(inhalt)

		angaben = append(angaben, sicherung.Teilangabe{
			Name: name, Groesse: int64(len(inhalt)),
			Summe: hex.EncodeToString(summe[:]),
		})
	}

	return angaben
}
