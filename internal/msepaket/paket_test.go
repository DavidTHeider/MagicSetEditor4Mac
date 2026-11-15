package msepaket_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Verzeichnis und ZIP sind derselbe Gegenstand.
//
// Ein `.mse-set` ist im Betrieb meist ein ZIP-Archiv, ein Vorlagenpaket meist
// ein Verzeichnis — belegt in `docs/FORMATE.md`. Wer beides verschieden
// behandelt, schreibt jede Auswertung zweimal, und die seltener benutzte
// Fassung fällt zuerst um.
func TestVerzeichnisUndArchivGebenDasselbeHer(t *testing.T) {
	faelle := map[string]string{
		"als Verzeichnis": "wiederholung.mse-style",
		"als Archiv":      "gepackt.mse-style",
	}

	for name, paketname := range faelle {
		t.Run(name, func(t *testing.T) {
			paket := oeffneTestpaket(t, paketname)

			if paket.Art != msepaket.ArtStil {
				t.Errorf("Art %q, erwartet war %q", paket.Art, msepaket.ArtStil)
			}
			if paket.Spiel != "probe" {
				t.Errorf("Spiel %q, erwartet war %q", paket.Spiel, "probe")
			}
			if len(paket.Abhaengigkeiten) == 0 {
				t.Error("keine Abhängigkeit gelesen")
			}
		})
	}
}

func TestDieArtKommtAusDerNamensendung(t *testing.T) {
	faelle := map[string]msepaket.Art{
		"magic.mse-game":             msepaket.ArtSpiel,
		"magic-m15.mse-style":        msepaket.ArtStil,
		"meins.mse-set":              msepaket.ArtSet,
		"magic-mana.mse-symbol-font": msepaket.ArtSymbolschrift,
		"magic-blends.mse-include":   msepaket.ArtEinschub,
		"de.mse-locale":              msepaket.ArtSprache,
		"html.mse-export-template":   msepaket.ArtExportvorlage,
	}

	for name, erwartet := range faelle {
		t.Run(name, func(t *testing.T) {
			if art := msepaket.ArtAusName(name); art != erwartet {
				t.Errorf("Art %q, erwartet war %q", art, erwartet)
			}
		})
	}

	t.Run("etwas anderes", func(t *testing.T) {
		if art := msepaket.ArtAusName("liesmich.txt"); art != msepaket.ArtUnbekannt {
			t.Errorf("Art %q, erwartet war %q", art, msepaket.ArtUnbekannt)
		}
	})
}

// Versionen sind Datumsangaben, keine Zahlenfolgen.
//
// `2024-05-26` ist kein Semver, und ein Vergleich als Zeichenkette ginge nur
// deshalb gut, weil ISO-Daten zufällig lexikographisch sortieren. Auf diese
// Zufälligkeit soll sich nichts stützen.
func TestVersionenSindDaten(t *testing.T) {
	paket := oeffneTestpaket(t, "gepackt.mse-style")

	if paket.Version.String() != "2026-08-08" {
		t.Errorf("Version %q, erwartet war %q", paket.Version, "2026-08-08")
	}

	aelter := paket.Abhaengigkeiten[0].Version
	if !aelter.VorOder(paket.Version) {
		t.Errorf("%s gilt nicht als älter als %s", aelter, paket.Version)
	}
	if paket.Version.VorOder(aelter) {
		t.Errorf("%s gilt als älter als %s", paket.Version, aelter)
	}
}

func TestEineUnbrauchbareVersionIstEinBefund(t *testing.T) {
	if _, err := msepaket.DatumLesen("übermorgen"); err == nil {
		t.Error("eine unbrauchbare Version wurde angenommen")
	}
	if _, err := msepaket.DatumLesen("2026-13-01"); err == nil {
		t.Error("ein dreizehnter Monat wurde angenommen")
	}
}

// Dateien im Paket werden über `path` angesprochen, nie über `filepath`.
//
// Paketverweise tragen `/` — `/magic-modules.mse-include/cards/…` —, und unter
// Windows trennt `filepath` mit `\`. Unter macOS fällt der Unterschied nicht
// auf; unter Windows sofort (`docs/PORTIERUNG.md`).
func TestDateienWerdenUeberSchraegstricheAngesprochen(t *testing.T) {
	paket := oeffneTestpaket(t, "gepackt.mse-style")

	inhalt, err := paket.Datei("bilder/rahmen.png")
	if err != nil {
		t.Fatalf("bilder/rahmen.png nicht lesbar: %v", err)
	}
	if len(inhalt) == 0 {
		t.Error("die Datei ist leer")
	}
}

// Ein Verweis, der aus dem Paket herausführt, wird abgewiesen.
//
// Ein Paket kommt von außen — heruntergeladen, weitergereicht, aus einer
// fremden Sammlung. Ein Eintrag `../../../.ssh/id_ed25519` darin ist kein
// theoretischer Fall, sondern der Grund, warum diese Prüfung existiert.
func TestEinVerweisAusDemPaketHerausWirdAbgewiesen(t *testing.T) {
	paket := oeffneTestpaket(t, "gepackt.mse-style")

	faelle := []string{
		"../geheim",
		"bilder/../../geheim",
		"/etc/passwd",
		"./../geheim",
	}

	for _, verweis := range faelle {
		t.Run(verweis, func(t *testing.T) {
			_, err := paket.Datei(verweis)

			if !errors.Is(err, msepaket.ErrAusbruch) {
				t.Errorf("Fehler %v, erwartet war ErrAusbruch", err)
			}
		})
	}
}

func TestEineFehlendeDateiIstEinNormalerFehler(t *testing.T) {
	paket := oeffneTestpaket(t, "gepackt.mse-style")

	if _, err := paket.Datei("gibtsnicht.png"); err == nil {
		t.Error("eine fehlende Datei wurde angenommen")
	} else if errors.Is(err, msepaket.ErrAusbruch) {
		t.Error("eine fehlende Datei gilt als Ausbruch — das verwischt zwei sehr verschiedene Fälle")
	}
}

// Eine fehlende Abhängigkeit ist ein Befund, kein Abbruch.
//
// Fremde Daten sind der Normalfall (AGENTS.md §8). Ein Paket, dessen zweite
// Abhängigkeit nicht aufzulösen ist, wird mit einem benannten Vorbehalt
// geöffnet — nicht abgewiesen.
func TestEineUnvollstaendigeAbhaengigkeitIstEinBefund(t *testing.T) {
	paket := oeffneTestpaket(t, "unvollstaendig.mse-style")

	if len(paket.Befunde) == 0 {
		t.Fatal("keine Befunde — die unvollständige Abhängigkeit ging still durch")
	}

	befund := paket.Befunde[0]
	if befund.Stelle.Zeile == 0 {
		t.Error("der Befund nennt keine Zeile")
	}
	if befund.Meldung == "" {
		t.Error("der Befund trägt keine Meldung")
	}

	// Und das Paket ist trotzdem benutzbar.
	if paket.Spiel != "probe" {
		t.Errorf("Spiel %q — das Paket wurde trotz Befund nicht gelesen", paket.Spiel)
	}
}

func TestEinPaketOhneHauptdateiWirdAbgewiesen(t *testing.T) {
	if _, err := msepaket.Oeffne(t.TempDir() + "/leer.mse-style"); err == nil {
		t.Error("ein Pfad ohne Paket wurde angenommen")
	}
}

// ---------------------------------------------------------------- Vorrichtung

func oeffneTestpaket(t *testing.T, name string) *msepaket.Paket {
	t.Helper()

	paket, err := msepaket.Oeffne(filepath.Join(testdaten, name))
	if err != nil {
		t.Fatalf("%s nicht öffenbar: %v", name, err)
	}
	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("%s nicht schließbar: %v", name, fehler)
		}
	})

	return paket
}

func TestDieHauptdateiFolgtDerEndung(t *testing.T) {
	faelle := map[msepaket.Art]string{
		msepaket.ArtSpiel:         "game",
		msepaket.ArtStil:          "style",
		msepaket.ArtSet:           "set",
		msepaket.ArtSymbolschrift: "symbol-font",
		msepaket.ArtUnbekannt:     "",
	}

	for art, erwartet := range faelle {
		t.Run(string(art), func(t *testing.T) {
			if haupt := art.Hauptdatei(); haupt != erwartet {
				t.Errorf("Hauptdatei %q, erwartet war %q", haupt, erwartet)
			}
		})
	}
}

func TestDatumVergleichtJahrMonatUndTag(t *testing.T) {
	faelle := []struct {
		name    string
		links   string
		rechts  string
		vorOder bool
	}{
		{"früheres Jahr", "2025-12-31", "2026-01-01", true},
		{"späteres Jahr", "2026-01-01", "2025-12-31", false},
		{"früherer Monat", "2026-01-31", "2026-02-01", true},
		{"späterer Monat", "2026-02-01", "2026-01-31", false},
		{"früherer Tag", "2026-08-07", "2026-08-08", true},
		{"späterer Tag", "2026-08-09", "2026-08-08", false},
		{"derselbe Tag", "2026-08-08", "2026-08-08", true},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			links, err := msepaket.DatumLesen(fall.links)
			if err != nil {
				t.Fatalf("%s: %v", fall.links, err)
			}
			rechts, fehler := msepaket.DatumLesen(fall.rechts)
			if fehler != nil {
				t.Fatalf("%s: %v", fall.rechts, fehler)
			}

			if links.VorOder(rechts) != fall.vorOder {
				t.Errorf("%s.VorOder(%s) = %v, erwartet war %v",
					links, rechts, links.VorOder(rechts), fall.vorOder)
			}
		})
	}
}

// Ein fehlendes Datum ist leer, nicht der 1. Januar des Jahres null.
//
// Der Unterschied zählt: „keine Version angegeben" und „Version 0000-01-01"
// sind zwei verschiedene Aussagen, und ein Vergleich behandelt sie gleich,
// wenn niemand sie trennt.
func TestEinFehlendesDatumIstLeer(t *testing.T) {
	paket := oeffneTestpaket(t, "wiederholung.mse-style")

	if !paket.Version.Leer() {
		t.Errorf("Version %s gilt nicht als leer", paket.Version)
	}

	gesetzt := oeffneTestpaket(t, "gepackt.mse-style")
	if gesetzt.Version.Leer() {
		t.Error("eine gesetzte Version gilt als leer")
	}
}

func TestEinBefundSchreibtSichWieEineFundstelle(t *testing.T) {
	paket := oeffneTestpaket(t, "unvollstaendig.mse-style")

	text := paket.Befunde[0].String()
	if !strings.Contains(text, "unvollstaendig.mse-style/style:") {
		t.Errorf("der Befund %q nennt seine Datei nicht", text)
	}
}

// Ein kaputtes Archiv wird abgewiesen, nicht halb gelesen.
func TestEinKaputtesArchivWirdAbgewiesen(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "kaputt.mse-set")
	if err := os.WriteFile(pfad, []byte("das ist kein ZIP"), 0o600); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	if _, err := msepaket.Oeffne(pfad); err == nil {
		t.Error("ein kaputtes Archiv wurde angenommen")
	}
}

// Ein Paket ohne erkennbare Endung wird abgewiesen.
//
// Nicht stillschweigend als Verzeichnis gelesen: welche Hauptdatei gesucht
// werden müsste, folgt aus der Art — ohne sie gibt es nichts zu lesen.
func TestEinPaketOhneBekannteEndungWirdAbgewiesen(t *testing.T) {
	verzeichnis := filepath.Join(t.TempDir(), "irgendwas")
	if err := os.MkdirAll(verzeichnis, 0o700); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	_, err := msepaket.Oeffne(verzeichnis)
	if err == nil {
		t.Fatal("ein Verzeichnis ohne Paketendung wurde angenommen")
	}
	if !strings.Contains(err.Error(), "Paketendung") {
		t.Errorf("die Meldung %q nennt den Grund nicht", err.Error())
	}
}

// Eine unbrauchbare Version im Kopf ist ein Befund, kein Abbruch.
func TestEineUnbrauchbareVersionImKopfIstEinBefund(t *testing.T) {
	verzeichnis := filepath.Join(t.TempDir(), "krumm.mse-style")
	if err := os.MkdirAll(verzeichnis, 0o700); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}
	if err := os.WriteFile(filepath.Join(verzeichnis, "style"),
		[]byte("game: probe\nversion: übermorgen\n"), 0o600); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	paket, err := msepaket.Oeffne(verzeichnis)
	if err != nil {
		t.Fatalf("das Paket wurde abgewiesen statt mit Befund geöffnet: %v", err)
	}
	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	if len(paket.Befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet war einer", len(paket.Befunde))
	}
	if !paket.Version.Leer() {
		t.Errorf("Version %s — eine unlesbare Version sollte leer bleiben", paket.Version)
	}
}

// Eine Hauptdatei, die sich nicht zerteilen lässt, beendet das Öffnen.
//
// Anders als eine krumme Version: ohne Kopf gibt es kein Paket, über das sich
// etwas aussagen ließe. Ein Befund wäre hier eine Behauptung, das Gelesene sei
// brauchbar.
//
// **Der Anlass ist seit #172 ein anderer**: `kein doppelpunkt` ist seit
// [0014](../../docs/entscheidungen/0014-eine-mse-datei-ist-nicht-notwendig-ein-baum.md)
// Text. Übrig bleibt der leere Schlüssel — der letzte Weg, auf dem eine Datei
// noch abgewiesen wird, und deshalb bleibt dieser Test.
func TestEineUnzerteilbareHauptdateiBeendetDasOeffnen(t *testing.T) {
	verzeichnis := filepath.Join(t.TempDir(), "wirr.mse-style")
	if err := os.MkdirAll(verzeichnis, 0o700); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}
	if err := os.WriteFile(filepath.Join(verzeichnis, "style"),
		[]byte(": ohne schlüssel\n"), 0o600); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	if _, err := msepaket.Oeffne(verzeichnis); err == nil {
		t.Error("eine unzerteilbare Hauptdatei wurde angenommen")
	}
}

func TestSchliessenZweimalIstStill(t *testing.T) {
	paket, err := msepaket.Oeffne(filepath.Join(testdaten, "gepackt.mse-style"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := paket.Schliessen(); fehler != nil {
		t.Fatalf("erstes Schließen: %v", fehler)
	}
	if fehler := paket.Schliessen(); fehler != nil {
		t.Errorf("zweites Schließen meldete %v", fehler)
	}
}

// Ein Paket mit richtiger Endung, aber ohne Hauptdatei.
//
// Die Meldung nennt, **welche** Datei gesucht wurde. Ohne sie wäre der
// häufigste Fall — ein Paket, dessen Hauptdatei anders heißt als vermutet —
// ein Rätsel statt eines Befundes; und die Regel „Endung ohne mse-" ist für
// fünf der sieben Arten nur Vermutung (`docs/FORMATE.md`).
func TestEinPaketMitEndungAberOhneHauptdatei(t *testing.T) {
	verzeichnis := filepath.Join(t.TempDir(), "leer.mse-symbol-font")
	if err := os.MkdirAll(verzeichnis, 0o700); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	_, err := msepaket.Oeffne(verzeichnis)
	if err == nil {
		t.Fatal("ein Paket ohne Hauptdatei wurde angenommen")
	}
	if !strings.Contains(err.Error(), "symbol-font") {
		t.Errorf("die Meldung %q nennt nicht, welche Datei gesucht wurde", err.Error())
	}
}
