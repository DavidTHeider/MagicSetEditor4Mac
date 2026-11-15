package schriften_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
)

// **Ein Fehlschlag wird gemerkt** — sonst läuft die Suche die Paketwurzel je
// Textebene neu ab.
//
// # Wie das aufgefallen ist
//
// Nicht beim Lesen. Ein CPU-Profil über den Zeichenweg von zwanzig Karten
// (#1029,
// Werkstück Ω.4) wies einem `filepath.WalkDir` unter `schriften.Suche` ein
// Drittel des Zeichenbaums zu — mehr als jede Punktrechnung.
//
// **Als Zahl über die Karte war das falsch gelesen**, und das gehört
// dazugesagt: bei vorhandenen Schriften sind es 13 Läufe über den **ganzen**
// Test, lauter erste Blicke. Was der Speicher wirklich einbringt, steht am
// Feld `fehlt` — 9,907 s gegen 398 ms, sobald eine Schrift **fehlt**.
//
// # Was der Test beobachtet
//
// Dass der zweite Blick die Wurzel **nicht** noch einmal abläuft, ist von
// außen nur an seiner Folge zu sehen: eine Datei, die danach in einem schon
// vorhandenen Unterverzeichnis auftaucht, ändert die Änderungszeit der Wurzel
// nicht — und wird deshalb nicht gefunden.
//
// **Das ist die Grenze, die `paketnamensindex` seit jeher führt**, und sie
// gilt ab jetzt für den Dateinamensweg genauso. Der andere Weg — ein Paket,
// das während der Sitzung **bezogen** wird — ändert die Wurzel und wird
// weiterhin gesehen; dafür steht
// [TestEinePaketschriftWirdAuchNachDemBezugGefunden], und ohne die Räumung in
// `fehlschlaege` wird der Test rot.
func TestEinFehlschlagWirdGemerkt(t *testing.T) {
	wurzel := t.TempDir()
	quelle, name := ersteSystemschrift(t)

	roh, err := os.ReadFile(quelle) //nolint:gosec // Pfad aus der Verzeichnisliste
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	// **Das Unterverzeichnis steht vor dem ersten Blick da.** Es später
	// anzulegen änderte die Wurzel, und dann prüfte dieser Test den Stand
	// statt das Merken.
	paket := filepath.Join(wurzel, "irgendein.mse-style")
	if err := os.Mkdir(paket, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	suche := schriften.Neue(nil).MitPaketen(wurzel)

	if _, err := suche.Suche(name, false); err == nil {
		t.Fatalf("die Schrift %q war schon da", name)
	}

	// **Unter dem eigenen Dateinamen** — der Weg, den `paketdatei` geht.
	if err := schreibe(t, paket, name+".ttf", roh); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, err := suche.Suche(name, false); err == nil {
		t.Errorf("der Fehlschlag zu %q wurde nicht gemerkt", name)
	}
}

// **Eine Schrift, die während der Sitzung installiert wird, wird gefunden.**
//
// # Was der erste Anlauf kaputt gemacht hat
//
// Der Fehlschlagspeicher hing allein an `paketstand()` — der Änderungszeit der
// **Paketwurzel**. Eine Datei, die jemand nach `~/Library/Fonts` legt, ändert
// die nicht.
//
// Bis dahin fand der Dateinamensweg sie: [Suche.datei] liest die Orte bei
// **jedem** Versuch neu. Der Speicher brach vorher ab, und dieselbe Schrift
// blieb bis zum Neustart unauffindbar — genau in dem Fall, für den man sie
// gerade installiert hat.
//
// Angestrichen vom Zweitblick über Phase Ω. Der Kommentar am Speicher berief
// sich darauf, dass [Suche.namensindex] ohnehin nur einmal je Sitzung gebaut
// werde; das deckt aber nur die Suche nach dem **eigenen** Namen einer
// Schrift, nicht den Weg über den Dateinamen.
func TestEineNeuInstallierteSchriftWirdGefunden(t *testing.T) {
	ort := t.TempDir()
	quelle, name := ersteSystemschrift(t)

	roh, err := os.ReadFile(quelle) //nolint:gosec // Pfad aus der Verzeichnisliste
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	suche := schriften.Neue([]string{ort})

	if _, err := suche.Suche(name, false); err == nil {
		t.Fatalf("die Schrift %q war schon da", name)
	}

	// **Jetzt wird sie installiert** — in einen Ort, den die Suche kennt.
	if err := schreibe(t, ort, name+".ttf", roh); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, err := suche.Suche(name, false); err != nil {
		t.Errorf("die neu installierte Schrift %q wurde nicht gefunden: %v", name, err)
	}
}
