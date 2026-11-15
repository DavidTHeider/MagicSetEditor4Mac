// **Die Wege, an denen das Sichern und Zurückspielen scheitert** (#1162).
//
// Sie brauchen Nähte, die von außen nicht zu setzen sind — ein Öffnen, das
// fehlschlägt, und ein Bilderverzeichnis, das sich nicht anfassen lässt.
// Deshalb steht diese Datei im Paket.
package service

import (
	"archive/zip"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// mitBilderort baut einen Sicherungsweg samt Bilderverzeichnis.
func mitBilderort(t *testing.T) (weg *Bestandsweg, ort, bilder string) {
	t.Helper()

	bau := neuerWegbau(t)

	return bau.weg, bau.ort, bau.bilder
}

// mitZaehlern gibt zusätzlich die Zähler des Lebenszyklus.
//
// **Getrennt, weil die meisten Fälle sie nicht brauchen** — und ein Rückgabe
// mehr in jedem Aufruf wäre Rauschen um der Einheitlichkeit willen.
func mitZaehlern(t *testing.T) (bau *wegbau, ort, bilder string) {
	t.Helper()

	bau = neuerWegbau(t)

	return bau, bau.ort, bau.bilder
}

// gesichert schreibt eine echte Sicherung und gibt ihren Pfad.
func gesichert(t *testing.T, weg *Bestandsweg, ziel string) string {
	t.Helper()

	if fehler := weg.Sichern(t.Context(), ziel); fehler != nil {
		t.Fatalf("sichern: %v", fehler)
	}

	return ziel
}

// **Ohne Bilderverzeichnis wird gar nicht erst gesichert.**
//
// Eine Sicherung, die den Ort der Bilder nicht kennt, wüsste nicht, ob sie
// vollständig ist — und eine unvollständige, die sich vollständig nennt, ist
// der Fehler, um den es hier geht.
func TestOhneBilderortWirdNichtGesichert(t *testing.T) {
	weg, ort, _ := mitBilderort(t)
	weg.orte.Bilder = func() (string, error) { return "", errors.New("kein Zuhause") }

	fehler := weg.Sichern(t.Context(), filepath.Join(ort, "x"+sicherung.Endung))
	if fehler == nil {
		t.Fatal("das Sichern kam durch")
	}

	if !strings.Contains(fehler.Error(), "Bilderverzeichnis") {
		t.Errorf("die Meldung sagt nichts über den Grund: %q", fehler.Error())
	}

	if _, err := os.Stat(filepath.Join(ort, "x"+sicherung.Endung)); !os.IsNotExist(err) {
		t.Error("es liegt eine Sicherung da, die nur den halben Stand trägt")
	}
}

// gesichertMitFehler erwartet einen Fehlschlag beim Sichern.
func gesichertMitFehler(t *testing.T, weg *Bestandsweg, ziel string) {
	t.Helper()

	if fehler := weg.Sichern(t.Context(), ziel); fehler == nil {
		t.Fatal("das Sichern kam durch")
	}
}

// **Ein unlesbarer Unterordner bricht das Sichern ab.**
//
// Er wird **nicht übersprungen**: eine Sicherung, die stillschweigend ohne die
// Bilder eines Sets entsteht, ist genau die Zusage mit Loch, gegen die dieser
// ganze Weg gebaut ist.
func TestEinUnlesbarerBilderordnerBrichtDasSichernAb(t *testing.T) {
	weg, ort, bilder := mitBilderort(t)

	zu := filepath.Join(bilder, "7")
	if err := os.MkdirAll(zu, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.Chmod(zu, 0o000); err != nil {
		t.Fatalf("zusperren: %v", err)
	}

	// #nosec G302 -- ein Verzeichnis braucht das Ausführrecht.
	t.Cleanup(func() { _ = os.Chmod(zu, 0o700) }) //nolint:errcheck // Aufräumen

	gesichertMitFehler(t, weg, filepath.Join(ort, "y"+sicherung.Endung))
}

// **Ein Zwischenplatz, der sich nicht anlegen lässt, meldet sich.**
func TestOhneZwischenplatzWirdNichtGesichert(t *testing.T) {
	weg, ort, _ := mitBilderort(t)

	eng := filepath.Join(ort, "eng")
	if err := os.Mkdir(eng, 0o500); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	// #nosec G302 -- ein Verzeichnis braucht das Ausführrecht.
	t.Cleanup(func() { _ = os.Chmod(eng, 0o700) }) //nolint:errcheck // Aufräumen

	gesichertMitFehler(t, weg, filepath.Join(eng, "z"+sicherung.Endung))
}

// **Scheitert das Schreiben des Bestands, entsteht kein Archiv.**
func TestEinGescheiterterBestandGibtKeinArchiv(t *testing.T) {
	weg, ort, _ := mitBilderort(t)
	weg.MitBestandssicherer(bockigerSicherer{})

	ziel := filepath.Join(ort, "halb"+sicherung.Endung)
	gesichertMitFehler(t, weg, ziel)

	if _, err := os.Stat(ziel); !os.IsNotExist(err) {
		t.Error("es liegt ein Archiv ohne Bestand da")
	}
}

// bockigerSicherer schreibt nichts und sagt es.
type bockigerSicherer struct{}

func (bockigerSicherer) Sichern(context.Context, string) error {
	return errors.New("die Datenbank gibt nichts her")
}

// **Ein Archiv ohne Bestand wird benannt abgewiesen.**
//
// Es kann entstehen, wenn jemand eine fremde ZIP-Datei wählt — und dann soll
// dastehen, was ihr fehlt, statt „unbrauchbar".
func TestEinArchivOhneBestandWirdBenanntAbgewiesen(t *testing.T) {
	weg, ort, _ := mitBilderort(t)

	leer := filepath.Join(ort, "leer.txt")
	if err := os.WriteFile(leer, []byte("nur ein Bild"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	ohne := filepath.Join(ort, "ohne"+sicherung.Endung)
	if err := sicherung.Packen(ohne,
		[]sicherung.Teil{{Name: Beiwerkordner + "/7/kunst.png", Quelle: leer}},
		sicherung.Angaben{Fassung: Pruefassung, Erstellt: time.Now().UTC()}); err != nil {
		t.Fatalf("packen: %v", err)
	}

	_, rueckfehler := weg.Zurueckspielen(t.Context(), ohne)
	if rueckfehler == nil {
		t.Fatal("das Archiv ohne Bestand wurde angenommen")
	}

	if !strings.Contains(rueckfehler.Error(), Bestandsname) {
		t.Errorf("die Meldung nennt nicht, was fehlt: %q", rueckfehler.Error())
	}
}

// **Ein Archiv, das keines ist, wird abgewiesen — auch in der Vorschau.**
//
// Die ersten vier Bytes sagen „ZIP", der Rest sagt nichts. Der Weg muss das
// aushalten, ohne die laufende Anwendung anzufassen.
func TestEinKaputtesArchivWirdAbgewiesen(t *testing.T) {
	weg, ort, _ := mitBilderort(t)

	kaputt := filepath.Join(ort, "kaputt"+sicherung.Endung)
	if err := os.WriteFile(kaputt,
		append([]byte("PK\x03\x04"), []byte("und dann nichts Gültiges")...), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, rueckfehler := weg.Zurueckspielen(t.Context(), kaputt); rueckfehler == nil {
		t.Fatal("das kaputte Archiv wurde angenommen")
	}

	_, schaufehler := weg.Beschauen(t.Context(), kaputt)
	if schaufehler == nil {
		t.Fatal("die Vorschau nahm das kaputte Archiv an")
	}
}

// **Eine Sicherung ohne Bilder räumt das Bilderverzeichnis.**
//
// Sie ist der Gegenfall zur alten `.db`: sie **sagt** etwas über die Bilder,
// nämlich dass es keine gab. Bliebe der Ordner stehen, entstünde ein Stand,
// den es nie gab.
func TestEineSicherungOhneBilderRaeumtDasVerzeichnis(t *testing.T) {
	weg, ort, bilder := mitBilderort(t)

	ziel := gesichert(t, weg, filepath.Join(ort, "ohnebilder"+sicherung.Endung))

	if err := os.MkdirAll(filepath.Join(bilder, "9"), 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(bilder, "9", "spaet.png"),
		[]byte("kam danach"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, rueckfehler := weg.Zurueckspielen(t.Context(), ziel); rueckfehler != nil {
		t.Fatalf("zurückspielen: %v", rueckfehler)
	}

	if _, err := os.Stat(filepath.Join(bilder, "9", "spaet.png")); !os.IsNotExist(err) {
		t.Errorf("das spätere Bild steht noch da: %v", err)
	}
}

// **Ohne Bilderverzeichnis wird nicht zurückgespielt** — und der Bestand
// bleibt der alte.
func TestOhneBilderortWirdNichtZurueckgespielt(t *testing.T) {
	bau, ort, _ := mitZaehlern(t)
	weg := bau.weg

	ziel := gesichert(t, weg, filepath.Join(ort, "gut"+sicherung.Endung))

	weg.orte.Bilder = func() (string, error) { return "", errors.New("kein Zuhause") }

	if _, rueckfehler := weg.Zurueckspielen(t.Context(), ziel); rueckfehler == nil {
		t.Fatal("es wurde zurückgespielt, ohne die Bilder zu kennen")
	}

	// **Die Anwendung läuft weiter** — das ist die Zusage des Rückwegs. Der
	// Weg hat sie geschlossen und wieder geöffnet; die Zähler stehen gleich.
	if bau.leben.geschlossen != bau.leben.geoeffnet {
		t.Errorf("geschlossen %d-mal, geöffnet %d-mal — die Anwendung steht ohne Datenbank da",
			bau.leben.geschlossen, bau.leben.geoeffnet)
	}
}

// **Lässt sich der zurückgespielte Stand nicht öffnen, gehen beide Teile
// zurück** — der Bestand und die Bilder.
//
// Das ist der Fall, an dem sich entscheidet, ob der Rückweg vollständig ist:
// ein Zurückdrehen, das die Datenbank zurückholt und die Bilder des neuen
// Stands stehen lässt, hinterließe genau die Vermischung, die niemand
// auflösen kann.
func TestEinGescheitertesOeffnenHoltAuchDieBilderZurueck(t *testing.T) {
	weg, ort, bilder := mitBilderort(t)

	if err := os.MkdirAll(bilder, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(bilder, "alt.png"),
		[]byte("der laufende Stand"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	ziel := gesichert(t, weg, filepath.Join(ort, "gut"+sicherung.Endung))

	if err := os.WriteFile(filepath.Join(bilder, "alt.png"),
		[]byte("danach geändert"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	// **Der erste Öffnungsversuch scheitert** — der auf den zurückgespielten
	// Stand. Der zweite, der zum alten gehört, gelingt.
	echt := weg.oeffnen
	versuche := 0

	beimZweiten := ""

	weg.oeffnen = func(ctx context.Context) error {
		versuche++
		if versuche == 1 {
			return errors.New("die zurückgespielte Datei taugt nicht")
		}

		// **Was beim Öffnen des alten Stands an Bildern daliegt** — A05,
		// Befund 3 des Zweitblicks über Phase A: der alte Bestand wurde
		// geöffnet, bevor seine Bilder zurück waren.
		roh, _ := os.ReadFile(filepath.Join(bilder, "alt.png")) //nolint:errcheck,gosec // aus t.TempDir(); fehlt sie, vergleicht der Lauf ""
		beimZweiten = string(roh)

		return echt(ctx)
	}

	if _, rueckfehler := weg.Zurueckspielen(t.Context(), ziel); rueckfehler == nil {
		t.Fatal("das gescheiterte Öffnen wurde als Erfolg gemeldet")
	}

	if beimZweiten != "danach geändert" {
		t.Errorf("der alte Bestand ging auf, während die Bilder noch %q waren", beimZweiten)
	}

	roh, err := os.ReadFile(filepath.Join(bilder, "alt.png")) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("das Bild des laufenden Stands ist fort: %v", err)
	}

	if string(roh) != "danach geändert" {
		t.Errorf("die Bilder wurden nicht zurückgedreht: %q", roh)
	}
}

// **Und ein Archiv ohne Kopf ist keine Sicherung** — die Zusicherung, dass
// [sicherung.IstArchiv] nicht rät.
func TestEineDateiOhneZipkopfGehtDenAltenWeg(t *testing.T) {
	ort := t.TempDir()
	fremd := filepath.Join(ort, "notiz.db")

	if err := os.WriteFile(fremd, []byte("kein ZIP"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if sicherung.IstArchiv(fremd) {
		t.Error("eine Textdatei gilt als Archiv")
	}

	// **Und die Gegenprobe**: ein echtes ZIP wird erkannt.
	echt := filepath.Join(ort, "echt.zip")

	datei, err := os.Create(echt) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	packer := zip.NewWriter(datei)
	if _, err := packer.Create("leer"); err != nil {
		t.Fatalf("eintragen: %v", err)
	}

	if err := errors.Join(packer.Close(), datei.Close()); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	if !sicherung.IstArchiv(echt) {
		t.Error("ein echtes ZIP gilt nicht als Archiv")
	}
}

// **Ist neben dem Bestand kein Platz, wird nichts angefasst.**
//
// Der Zwischenplatz entsteht **dort** und nicht in `/tmp`: das Umbenennen am
// Ende muss im selben Dateisystem bleiben. Lässt er sich nicht anlegen, ist
// das das Ende des Wegs — und der laufende Stand hat nichts davon gemerkt.
func TestOhneZwischenplatzWirdNichtZurueckgespielt(t *testing.T) {
	bau, _, _ := beiwerkAnwendung(t)
	weg := bau.weg

	ziel := gesichert(t, weg, filepath.Join(t.TempDir(), "gut"+sicherung.Endung))

	zusperren(t, filepath.Dir(weg.orte.Datenbank))

	_, rueckfehler := weg.Zurueckspielen(t.Context(), ziel)
	if rueckfehler == nil {
		t.Fatal("es wurde zurückgespielt, obwohl kein Platz war")
	}

	// **Die Anwendung läuft weiter** — nichts wurde geschlossen, nichts
	// getauscht.
	if bau.leben.geschlossen != bau.leben.geoeffnet {
		t.Errorf("geschlossen %d-mal, geöffnet %d-mal — die Anwendung steht ohne Datenbank da",
			bau.leben.geschlossen, bau.leben.geoeffnet)
	}
}
