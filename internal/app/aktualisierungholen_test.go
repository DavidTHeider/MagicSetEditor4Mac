// # Warum hier nichts nebenläufig läuft
//
// Der Riegel `laeuftGerade` ist ein **Paketwert**: „ein Ladevorgang je
// Prozess" (§7, Regel 4). Im Betrieb ist das richtig — es gibt eine Anwendung
// je Prozess. Im Test heißt es, dass zwei Läufe mit `t.Parallel()` einander
// den Riegel wegnehmen; einer bekommt dann „es läuft schon ein Ladevorgang"
// und ist rot, ohne dass an der Sache etwas falsch wäre.
//
// **Gefunden am 2026-08-26** als roter `make ci` in einem PR, der mit der
// Aktualisierung nichts zu tun hatte
// (#1162):
// rund jeder dritte volle Lauf. Ein flackernder Lauf ist schlimmer als ein
// langsamer — er lehrt, rote Läufe zu wiederholen statt zu lesen.
//
// **Deshalb steht in dieser Datei kein `t.Parallel()`**, und zwar in keinem
// Lauf: welcher den Riegel nimmt, hängt daran, wie weit er kommt, und diese
// Unterscheidung ist zu fein, um sie einem späteren Leser zuzumuten.

package app_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/aktualisierung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Laden legt eine Datei ab — oder gibt den vorgemerkten Fehler.
//
// **Der Fortschritt wird wirklich gemeldet**, denn genau darüber sagt der
// Test etwas: eine Naht, die ihn nur entgegennimmt, ließe den Weg ins Fenster
// ungeprüft.
func (f *freigabeprobe) Laden(
	_ context.Context, freigabe aktualisierung.Freigabe, verzeichnis string,
	melden aktualisierung.Fortschritt,
) (string, error) {
	f.geladen++

	if f.ladefehler != nil {
		return "", f.ladefehler
	}

	if f.waehrenddessen != nil {
		f.waehrenddessen()
	}

	if melden != nil {
		melden(1, 2)
		melden(2, 2)
	}

	pfad := filepath.Join(verzeichnis, freigabe.Abbild.Name)

	return pfad, os.WriteFile(pfad, []byte("abbild"), 0o600)
}

// **Mit „holen" wird geladen, und der Ort steht in der Antwort.**
func TestMitHolenWirdGeladen(t *testing.T) {
	ziel := t.TempDir()
	probe := &freigabeprobe{freigabe: aktualisierung.Freigabe{
		Fassung: "99.0.0",
		Abbild:  aktualisierung.Beigabe{Name: "neu.dmg", Adresse: "http://x/neu.dmg"},
	}}

	anwendung := mitDownloads(t, probe, domain.AktualisierungHolen, ziel)

	var stufen [][2]int64

	anwendung.MitMeldung(func(_ context.Context, ereignis string, daten ...any) {
		if ereignis != app.EreignisLadefortschritt || len(daten) != 2 {
			return
		}

		gelesen, einer := daten[0].(int64)
		ganz, zweiter := daten[1].(int64)

		if !einer || !zweiter {
			t.Errorf("der Fortschritt kam als %T/%T", daten[0], daten[1])

			return
		}

		stufen = append(stufen, [2]int64{gelesen, ganz})
	})

	antwort := anwendung.NeueFassungLaden(t.Context())

	if antwort.Fehler != nil {
		t.Fatalf("das Laden meldete %+v", antwort.Fehler)
	}

	if filepath.Base(antwort.Pfad) != "neu.dmg" {
		t.Errorf("die Datei liegt unter %q", antwort.Pfad)
	}

	if antwort.Fassung != "99.0.0" {
		t.Errorf("die Antwort nennt die Fassung %q", antwort.Fassung)
	}

	// **Der Fortschritt kommt im Fenster an** — 188 MB sind kein Augenblick.
	if len(stufen) != 2 || stufen[1] != [2]int64{2, 2} {
		t.Errorf("der Fortschritt kam als %+v an", stufen)
	}
}

// **Mit „melden" wird nicht geladen** — die Stufe ist die Grenze.
//
// Das ist die Zusage, die von außen nur an einer Sache zu sehen ist: der
// Klient wurde **nicht** gefragt. Ein Test, der nur den Fehler prüft, sähe
// eine Meldung und wäre zufrieden.
func TestMitMeldenWirdNichtGeladen(t *testing.T) {
	probe := &freigabeprobe{freigabe: aktualisierung.Freigabe{Fassung: "99.0.0"}}
	anwendung := mitDownloads(t, probe, domain.AktualisierungMelden, t.TempDir())

	antwort := anwendung.NeueFassungLaden(t.Context())

	if probe.geladen != 0 || probe.gefragt != 0 {
		t.Errorf("es wurde %d-mal gefragt und %d-mal geladen", probe.gefragt, probe.geladen)
	}

	if antwort.Fehler == nil {
		t.Error("die Stufe melden lud, ohne es zu sagen")
	}
}

// **Und mit „aus" erst recht nicht.**
func TestMitAusWirdNichtGeladen(t *testing.T) {
	probe := &freigabeprobe{freigabe: aktualisierung.Freigabe{Fassung: "99.0.0"}}
	anwendung := mitDownloads(t, probe, domain.AktualisierungAus, t.TempDir())

	anwendung.NeueFassungLaden(t.Context())

	if probe.gefragt != 0 || probe.geladen != 0 {
		t.Error("mit der Stufe aus wurde gefragt oder geladen")
	}
}

// **Die laufende Fassung wird nicht geholt.**
//
// 188 MB für das, was schon läuft, wäre die teuerste Art, nichts zu tun.
func TestDieLaufendeFassungWirdNichtGeholt(t *testing.T) {
	probe := &freigabeprobe{freigabe: aktualisierung.Freigabe{Fassung: app.Fassung}}
	anwendung := mitDownloads(t, probe, domain.AktualisierungHolen, t.TempDir())

	antwort := anwendung.NeueFassungLaden(t.Context())

	if probe.geladen != 0 {
		t.Error("die laufende Fassung wurde geladen")
	}

	if antwort.Neuer || antwort.Pfad != "" {
		t.Errorf("die Antwort meldet %+v", antwort)
	}
}

// **Ein Fehler des Dienstes steht in der Antwort** — §16, Satz 5.
func TestEinLadefehlerStehtInDerAntwort(t *testing.T) {
	probe := &freigabeprobe{
		freigabe:   aktualisierung.Freigabe{Fassung: "99.0.0"},
		ladefehler: errors.New("die Leitung ist weg"),
	}
	anwendung := mitDownloads(t, probe, domain.AktualisierungHolen, t.TempDir())

	if antwort := anwendung.NeueFassungLaden(t.Context()); antwort.Fehler == nil {
		t.Error("der Ladefehler ging als Erfolg durch")
	}
}

// **Ein Fehler beim Fragen ebenso.**
func TestEinFragefehlerHaeltDasLadenAn(t *testing.T) {
	probe := &freigabeprobe{fehler: errors.New("kein Zugriff")}
	anwendung := mitDownloads(t, probe, domain.AktualisierungHolen, t.TempDir())

	antwort := anwendung.NeueFassungLaden(t.Context())

	if antwort.Fehler == nil {
		t.Error("der Fragefehler ging als Erfolg durch")
	}

	if probe.geladen != 0 {
		t.Error("nach einem Fragefehler wurde geladen")
	}
}

// **Eine unlesbare Fassung führt zu keinem Ladevorgang.**
func TestEineUnlesbareFassungFuehrtZuKeinemLaden(t *testing.T) {
	probe := &freigabeprobe{freigabe: aktualisierung.Freigabe{Fassung: "keine Zahl"}}
	anwendung := mitDownloads(t, probe, domain.AktualisierungHolen, t.TempDir())

	if antwort := anwendung.NeueFassungLaden(t.Context()); antwort.Fehler == nil {
		t.Error("eine unlesbare Fassung ging durch")
	}

	if probe.geladen != 0 {
		t.Error("eine unlesbare Fassung wurde geladen")
	}
}

// **Ohne Downloadverzeichnis wird das gesagt, nicht geraten.**
func TestOhneDownloadverzeichnisKommtEinFehler(t *testing.T) {
	probe := &freigabeprobe{freigabe: aktualisierung.Freigabe{Fassung: "99.0.0"}}
	anwendung := mitDownloads(t, probe, domain.AktualisierungHolen, "")

	anwendung.MitDownloadort(func() (string, error) {
		return "", errors.New("kein Zuhause")
	})

	if antwort := anwendung.NeueFassungLaden(t.Context()); antwort.Fehler == nil {
		t.Error("ohne Downloadverzeichnis kam keine Meldung")
	}

	if probe.geladen != 0 {
		t.Error("ohne Downloadverzeichnis wurde geladen")
	}
}

// **Ohne eingerichteten Zugang wird das gesagt und nicht geschwiegen.**
func TestOhneZugangKommtBeimLadenEinFehler(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitEinstellungen(&einstellungsprobe{
		werte: map[string]string{domain.EinstellungAktualisierung: domain.AktualisierungHolen},
	})

	if antwort := anwendung.NeueFassungLaden(t.Context()); antwort.Fehler == nil {
		t.Error("ohne Zugang kam keine Meldung")
	}
}

// **Ein unlesbarer Bestand meldet sich, statt zu schweigen.**
func TestOhneEinstellungenKommtBeimLadenEinFehler(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitEinstellungen(&einstellungsprobe{fehler: errors.New("kein Bestand")})
	anwendung.MitFassungen(&freigabeprobe{})

	if antwort := anwendung.NeueFassungLaden(t.Context()); antwort.Fehler == nil {
		t.Error("ein unlesbarer Bestand ging als nicht geladen durch")
	}
}

// mitDownloads baut eine Anwendung mit Wahl, Zugang und Ablageort.
func mitDownloads(
	t *testing.T, anleger app.Fassungsanleger, wahl, ziel string,
) *app.App {
	t.Helper()

	anwendung := mitFassungen(t, anleger, wahl)
	anwendung.MitDownloadort(func() (string, error) { return ziel, nil })
	anwendung.MitMeldung(func(context.Context, string, ...any) {})

	return anwendung
}

// **Die beiden Bindungen reichen durch** —
// #1100.
//
// Der Weg über `*Bindungen` ist der **einzige**, den das Fenster hat: gebunden
// wird `*Bindungen` (`main.go`), und ein Verfahren auf `*App` sieht es nicht.
// Genau davor stand Stufe 1 (§10, erste Dimension).
func TestDieAktualisierungsbindungenReichenDurch(t *testing.T) {
	ziel := t.TempDir()
	probe := &freigabeprobe{freigabe: aktualisierung.Freigabe{
		Fassung: "99.0.0",
		Abbild:  aktualisierung.Beigabe{Name: "neu.dmg"},
	}}

	bindungen := app.NeueBindungen(mitDownloads(t, probe, domain.AktualisierungHolen, ziel))

	if antwort := bindungen.NachNeuerFassungSehen(t.Context()); !antwort.Neuer {
		t.Errorf("das Nachsehen gab %+v", antwort)
	}

	if antwort := bindungen.NeueFassungLaden(t.Context()); filepath.Base(antwort.Pfad) != "neu.dmg" {
		t.Errorf("das Laden gab %+v", antwort)
	}
}

// **Auch das Fenster lädt nicht zweimal zugleich.**
//
// # Warum das eigens dasteht
//
// Der Riegel stand zuerst nur in [app.FassungLaden] — also an der Naht für
// Kommandozeile und Dienst. Die Bindung des Fensters ruft `NeueFassungLaden`
// und ging daran vorbei; „ein Vorgang je Prozess" galt damit für zwei von drei
// Flächen. Der Wächter im Frontendhaken hilft dort nicht: er hält seinen Satz
// **je Hakeninstanz**.
//
// Angestrichen vom Nachlauf des Gesamtreviews vor 0.2.0.
func TestAuchDasFensterLaedtNichtZweimalZugleich(t *testing.T) {
	probe := &freigabeprobe{freigabe: aktualisierung.Freigabe{
		Fassung: "99.0.0",
		Abbild:  aktualisierung.Beigabe{Name: "neu.dmg"},
	}}

	anwendung := mitDownloads(t, probe, domain.AktualisierungHolen, t.TempDir())

	var zweite app.LadeAntwort

	fertig := make(chan struct{})

	probe.waehrenddessen = func() {
		zweite = anwendung.NeueFassungLaden(t.Context())

		close(fertig)
	}

	if erste := anwendung.NeueFassungLaden(t.Context()); erste.Fehler != nil {
		t.Fatalf("der erste Ladevorgang: %+v", erste.Fehler)
	}

	<-fertig

	if zweite.Fehler == nil {
		t.Error("der zweite Ladevorgang kam durch")
	}
}
