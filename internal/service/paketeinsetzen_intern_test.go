// **Ein gescheitertes Einsetzen lässt das bisherige Paket stehen** — der
// zweite Zweitblick über die Freigabe 0.3.0, Befund 3
// (#1170).
//
// # Der Befund
//
// Der Bezug räumte den bisherigen Ort mit `os.RemoveAll` und benannte das
// bezogene Paket erst danach um. Zwischen beiden Zügen ist das Paket, das
// jemand hatte, **weg**. Die Begründung daneben lautete, nach einem
// geglückten Räumen könne das Umbenennen kaum noch scheitern — „kaum" ist
// keine Zusage, und ein Abbruch dazwischen fragt nicht danach.
//
// # Wie der rote Lauf herbeigeführt wird
//
// Über den eingesetzten `umbenennen`, aus demselben Grund wie bei
// `austauschen` in `internal/msepaket`: einen Fehlschlag genau zwischen zwei
// Umbenennungen liefert kein Dateisystem auf Bestellung (§3).
package service

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// paketOrt legt ein „bisheriges Paket" mit einer erkennbaren Datei darin an.
func paketOrt(t *testing.T) (wurzel, ort, vorlaeufig string) {
	t.Helper()

	wurzel = t.TempDir()
	ort = filepath.Join(wurzel, "M15-Magic-Pack")
	vorlaeufig = filepath.Join(wurzel, ".M15-Magic-Pack-neu")

	for verzeichnis, inhalt := range map[string]string{ort: "alt", vorlaeufig: "neu"} {
		if err := os.MkdirAll(verzeichnis, 0o750); err != nil {
			t.Fatalf("anlegen: %v", err)
		}

		if err := os.WriteFile(filepath.Join(verzeichnis, "kennung.txt"),
			[]byte(inhalt), 0o600); err != nil {
			t.Fatalf("schreiben: %v", err)
		}
	}

	return wurzel, ort, vorlaeufig
}

// kennung gibt den Inhalt der Kennungsdatei am Ort — oder "" bei Fehlen.
func kennung(t *testing.T, ort string) string {
	t.Helper()

	roh, err := os.ReadFile(filepath.Join(ort, "kennung.txt")) // #nosec G304 -- t.TempDir()
	if err != nil {
		return ""
	}

	return string(roh)
}

func TestEinGescheitertesEinsetzenLaesstDasBisherigePaketStehen(t *testing.T) {
	t.Parallel()

	_, ort, vorlaeufig := paketOrt(t)
	kaputt := errors.New("das Dateisystem lehnt ab")

	// **Nur der Zug ins Ziel scheitert.** Das Beiseiteräumen gelingt; genau
	// dieser Augenblick ist der, in dem der bisherige Stand verschwand.
	fehler := paketEinsetzen(ort, vorlaeufig, func(alt, neu string) error {
		if alt == vorlaeufig {
			return kaputt
		}

		return os.Rename(alt, neu)
	})

	if !errors.Is(fehler, kaputt) {
		t.Fatalf("Fehler: %v, erwartet %v", fehler, kaputt)
	}

	if da := kennung(t, ort); da != "alt" {
		t.Errorf("am Ort steht %q, erwartet \"alt\" — der bisherige Stand ist "+
			"weg, und das neue Paket kam nie an", da)
	}
}

// **Der geglückte Weg räumt den beiseitegelegten Stand weg.** Sonst bliebe
// bei jedem Bezug eine vollständige Kopie liegen, und die wiegt bei
// `M15-Magic-Pack` über ein Gigabyte.
func TestEinGeglueckterBezugLaesstNichtsLiegen(t *testing.T) {
	t.Parallel()

	wurzel, ort, vorlaeufig := paketOrt(t)

	if err := paketEinsetzen(ort, vorlaeufig, os.Rename); err != nil {
		t.Fatalf("einsetzen: %v", err)
	}

	if da := kennung(t, ort); da != "neu" {
		t.Errorf("am Ort steht %q, erwartet \"neu\"", da)
	}

	eintraege, err := os.ReadDir(wurzel)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(eintraege) != 1 {
		for _, eintrag := range eintraege {
			t.Logf("liegengeblieben: %s", eintrag.Name())
		}

		t.Errorf("%d Einträge in der Wurzel, erwartet 1", len(eintraege))
	}
}

// **Und ein Bezug ohne bisherigen Stand kommt ohne Beiseiteräumen aus.**
func TestEinErstbezugBrauchtKeinBeiseiteraeumen(t *testing.T) {
	t.Parallel()

	wurzel, ort, vorlaeufig := paketOrt(t)

	if err := os.RemoveAll(ort); err != nil {
		t.Fatalf("räumen: %v", err)
	}

	if err := paketEinsetzen(ort, vorlaeufig, os.Rename); err != nil {
		t.Fatalf("einsetzen: %v", err)
	}

	if da := kennung(t, ort); da != "neu" {
		t.Errorf("am Ort steht %q, erwartet \"neu\"", da)
	}

	if _, err := os.Stat(filepath.Join(wurzel, "M15-Magic-Pack.vorher")); err == nil {
		t.Error("es liegt ein .vorher da, obwohl es keinen vorigen Stand gab")
	}
}

// **Scheitert auch das Zurückholen**, sagt die Meldung, wo der bisherige Stand
// liegt. Das ist dann das Einzige, was noch hilft.
func TestScheitertAuchDasZurueckholenNenntDieMeldungDenOrt(t *testing.T) {
	t.Parallel()

	_, ort, vorlaeufig := paketOrt(t)

	fehler := paketEinsetzen(ort, vorlaeufig, func(alt, neu string) error {
		if alt == ort {
			return os.Rename(alt, neu)
		}

		return errors.New("das Dateisystem lehnt ab")
	})

	if fehler == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(fehler.Error(), ort+".vorher") {
		t.Errorf("die Meldung %q nennt nicht, wo der bisherige Stand liegt",
			fehler.Error())
	}
}

// **Ein Erstbezug, dessen Zug ins Ziel scheitert, meldet ihn schlicht.**
//
// Es gibt keinen bisherigen Stand, also auch nichts zurückzuholen — und der
// Zweig, der das feststellt, ist einer, den kein anderer Lauf betritt.
func TestEinGescheiterterErstbezugMeldetDenZugSchlicht(t *testing.T) {
	t.Parallel()

	_, ort, vorlaeufig := paketOrt(t)
	kaputt := errors.New("das Dateisystem lehnt ab")

	if err := os.RemoveAll(ort); err != nil {
		t.Fatalf("räumen: %v", err)
	}

	fehler := paketEinsetzen(ort, vorlaeufig, func(string, string) error { return kaputt })
	if !errors.Is(fehler, kaputt) {
		t.Fatalf("Fehler: %v, erwartet %v", fehler, kaputt)
	}

	if strings.Contains(fehler.Error(), ".vorher") {
		t.Errorf("die Meldung %q nennt einen bisherigen Stand, den es nie gab",
			fehler.Error())
	}
}

// **Scheitert schon das Beiseiteräumen, bleibt alles, wie es war.**
func TestScheitertDasBeiseiteraeumenBleibtAllesStehen(t *testing.T) {
	t.Parallel()

	_, ort, vorlaeufig := paketOrt(t)
	kaputt := errors.New("das Dateisystem lehnt ab")

	fehler := paketEinsetzen(ort, vorlaeufig, func(alt, _ string) error {
		if alt == ort {
			return kaputt
		}

		t.Error("es wurde weitergemacht, obwohl das Beiseiteräumen scheiterte")

		return nil
	})

	if !errors.Is(fehler, kaputt) {
		t.Fatalf("Fehler: %v, erwartet %v", fehler, kaputt)
	}

	if da := kennung(t, ort); da != "alt" {
		t.Errorf("am Ort steht %q, erwartet \"alt\"", da)
	}
}
