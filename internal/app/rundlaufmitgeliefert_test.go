//go:build fremdpakete

package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// **Der Rundlauf über das mitgelieferte Set** —
// #702.
//
// # Was er prüft, das kein anderer prüft
//
// `TestWegEinSetOeffnenAendernUndSichern` geht denselben Weg über ein Set mit
// **einer** Karte und drei Feldern — so sieht die Saat aus. Hier sind es 107
// Karten und 108 Dateien, jede davon ein Kartenbild: der Fall, den jemand
// wirklich hat.
//
// Geprüft wird die Zusage, die einen Rundlauf von einem Neuschreiben
// unterscheidet: die Änderung steht drin, **und** was niemand angefasst hat,
// steht unverändert da — die Bilder eingeschlossen.
//
//	MSE_BUENDEL=… go test -tags fremdpakete ./internal/app/ -run TestRundlaufUeberDasMitgelieferteSet -v
func TestRundlaufUeberDasMitgelieferteSet(t *testing.T) {
	buendel := os.Getenv("MSE_BUENDEL")
	if buendel == "" {
		t.Skip("MSE_BUENDEL ist nicht gesetzt")
	}

	sets := filepath.Join(buendel, "Contents", "Resources", "beispielsets")
	vorlagen := filepath.Join(buendel, "Contents", "Resources", "vorlagen")

	eintraege, err := os.ReadDir(sets)
	if err != nil || len(eintraege) == 0 {
		t.Skip("das Bündel trägt kein Beispielset")
	}

	quelle := filepath.Join(sets, eintraege[0].Name())

	// **Eine Kopie und nicht das Original.** Ein Lauf, der ins Bündel
	// zurückschriebe, änderte die Anwendung, die er prüft.
	ziel := filepath.Join(t.TempDir(), eintraege[0].Name())

	if err := kopiereBaum(quelle, ziel); err != nil {
		t.Fatalf("kopieren: %v", err)
	}

	vorher := baumgroessen(t, ziel)

	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitVorlagenort(func() string { return vorlagen })

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	schritt(t, "das mitgelieferte Set wird geöffnet")

	geoeffnet := anwendung.SetOeffnen(t.Context(), ziel)
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	karten := anwendung.Karten(t.Context(), geoeffnet.SetID)
	if karten.Fehler != nil || len(karten.Karten) == 0 {
		t.Fatalf("Karten: %+v", karten)
	}

	schritt(t, "ein Feld wird geändert")

	if antwort := anwendung.FeldSetzen(
		probekontext(t), "1", karten.Karten[0].ID, "name", "Eisball",
	); antwort.Fehler != nil {
		t.Fatalf("Feld setzen: %+v", antwort.Fehler)
	}

	schritt(t, "das Set wird gesichert")

	gesichert := anwendung.SetSpeichern(t.Context(), geoeffnet.SetID, ziel)
	if gesichert.Fehler != nil {
		t.Fatalf("sichern: %+v", gesichert.Fehler)
	}

	schritt(t, "und die Bilder sind noch da")

	nachher := baumgroessen(t, ziel)

	for name, groesse := range vorher {
		if name == "/set" {
			continue
		}

		if nachher[name] != groesse {
			t.Errorf("%s: %d Bytes vorher, %d nachher", name, groesse, nachher[name])
		}
	}

	if len(nachher) != len(vorher) {
		t.Errorf("%d Einträge vorher, %d nachher", len(vorher), len(nachher))
	}

	schritt(t, "das Set wird wieder geöffnet")

	zweite := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	zweite.MitVorlagenort(func() string { return vorlagen })

	if err := zweite.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if fehler := zweite.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	wieder := zweite.SetOeffnen(t.Context(), ziel)
	if wieder.Fehler != nil {
		t.Fatalf("wieder öffnen: %+v", wieder.Fehler)
	}

	nochmal := zweite.Karten(t.Context(), wieder.SetID)
	if nochmal.Fehler != nil {
		t.Fatalf("Karten: %+v", nochmal.Fehler)
	}

	if len(nochmal.Karten) != len(karten.Karten) {
		t.Errorf("%d Karten vorher, %d nachher",
			len(karten.Karten), len(nochmal.Karten))
	}

	// **Alle Feldwerte kommen zurück, nicht nur der geänderte** —
	// #704.
	// Vor dem Fund waren es 2 von 18: ein `notes` über drei Zeilen wurde
	// einzeilig geschrieben, und die zweite Zeile beendete den `card:`-Block.
	if len(nochmal.Karten[0].Werte) != len(karten.Karten[0].Werte) {
		t.Errorf("%d Feldwerte vorher, %d nachher",
			len(karten.Karten[0].Werte), len(nochmal.Karten[0].Werte))
	}

	if name := feldwert(nochmal.Karten[0], "name"); name != "Eisball" {
		t.Errorf("der Name lautet %q, erwartet Eisball", name)
	}

	// **Und ein Feld, das niemand angefasst hat** — die zweite Hälfte der
	// Zusage. Ein Rundlauf, der nur die Änderung wiederfindet, wäre ein
	// Neuschreiben mit einem Feld darin.
	if vorher, nachher := feldwert(karten.Karten[0], "notes"), feldwert(nochmal.Karten[0], "notes"); vorher != nachher {
		t.Errorf("notes vorher %q, nachher %q", vorher, nachher)
	}

	if text := feldwert(nochmal.Karten[0], "rule text"); text == "" {
		t.Error("ein unberührtes Feld ist beim Rundlauf verlorengegangen")
	}

	t.Logf("%d Karten, %d Dateien unverändert",
		len(nochmal.Karten), len(nachher)-1)
}

// kopiereBaum legt ein Verzeichnis an einem zweiten Ort ab.
//
//nolint:gosec // beide Orte kommen aus MSE_BUENDEL und t.TempDir()
func kopiereBaum(von, nach string) error {
	return filepath.Walk(von, func(pfad string, angaben os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		ziel := filepath.Join(nach, strings.TrimPrefix(pfad, von))
		if angaben.IsDir() {
			return os.MkdirAll(ziel, 0o750)
		}

		inhalt, fehler := os.ReadFile(pfad) //nolint:gosec // eigenes Bündel
		if fehler != nil {
			return fehler
		}

		return os.WriteFile(ziel, inhalt, 0o600)
	})
}

// baumgroessen gibt je Datei ihre Größe.
//
//nolint:gosec // die Wurzel kommt aus t.TempDir()
func baumgroessen(t *testing.T, wurzel string) map[string]int64 {
	t.Helper()

	groessen := map[string]int64{}

	fehler := filepath.Walk(wurzel, func(pfad string, angaben os.FileInfo, err error) error {
		if err != nil || angaben.IsDir() {
			return err
		}

		groessen[strings.TrimPrefix(pfad, wurzel)] = angaben.Size()

		return nil
	})
	if fehler != nil {
		t.Fatalf("%s durchgehen: %v", wurzel, fehler)
	}

	return groessen
}

// feldwert liest einen Feldwert aus einer Karte.
func feldwert(karte KarteDTO, feld string) string {
	for _, wert := range karte.Werte {
		if strings.EqualFold(wert.Feld, feld) {
			return wert.Wert
		}
	}

	return ""
}
