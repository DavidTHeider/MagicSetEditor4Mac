// **Die Fehlerwege des Rückspielens** — jeder betreten, kein halber Stand
// (#1175).
//
// # Warum sie hier neu stehen
//
// Sie waren gedeckt — durch `internal/app`, solange der Weg dort lag. Mit dem
// Umzug in den Dienst wäre die Deckung mitgewandert, wenn die Tests durch die
// Fassade gefahren wären; das taten sie, und genau deshalb blieben elf Zweige
// zurück, die niemand mehr betrat.
//
// **Was sie dabei gewinnen**: die Nähte des Dienstes sind Rückrufe, und ein
// Fehler darin ist eine Zeile statt einer zugesperrten Datei. Was vorher nur
// über das Dateisystem herzustellen war — „das Schließen scheitert" —, ist
// hier eine Angabe.

package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// scheiternderPruefer weist jede Sicherung ab.
func scheiternderPruefer(grund string) Sicherungspruefer {
	return func(context.Context, string) (Sicherungsbefund, error) {
		return Sicherungsbefund{}, errors.New(grund)
	}
}

// eineSicherung schreibt eine gültige Sicherung und gibt ihren Pfad.
func eineSicherung(t *testing.T, bau *wegbau) string {
	t.Helper()

	return gesichert(t, bau.weg, filepath.Join(bau.ort, "gut"+Sicherungsendung))
}

// **Eine Sicherung, die der Prüfer abweist, wird nicht übernommen.**
//
// Und zwar **vor** dem Schließen: die Anwendung hat davon nichts gemerkt.
func TestEinAbweisenderPrueferHaeltDasZurueckspielenAn(t *testing.T) {
	bau := neuerWegbau(t)
	ziel := eineSicherung(t, bau)

	bau.weg.pruefer = scheiternderPruefer("die Datei taugt nicht")

	_, fehler := bau.weg.Zurueckspielen(t.Context(), ziel)
	if fehler == nil {
		t.Fatal("die abgewiesene Sicherung wurde übernommen")
	}

	if bau.leben.geschlossen != 0 {
		t.Errorf("die Verbindung wurde %d-mal geschlossen — die Prüfung steht davor",
			bau.leben.geschlossen)
	}
}

// **Und die Vorschau ebenso.**
func TestEinAbweisenderPrueferMeldetSichAuchInDerVorschau(t *testing.T) {
	bau := neuerWegbau(t)
	ziel := eineSicherung(t, bau)

	bau.weg.pruefer = scheiternderPruefer("die Datei taugt nicht")

	if _, fehler := bau.weg.Beschauen(t.Context(), ziel); fehler == nil {
		t.Fatal("die Vorschau nahm die abgewiesene Sicherung an")
	}
}

// **Lässt sich der Bestand nicht schließen, wird nicht getauscht.**
//
// Eine Datei umzubenennen, die die eigene Verbindung noch offen hält, nennt
// SQLite undefiniert — und der Grund gehört in die Meldung.
func TestOhneSchliessenWirdNichtGetauscht(t *testing.T) {
	bau := neuerWegbau(t)
	ziel := eineSicherung(t, bau)

	bau.leben.schliessfehler = errors.New("die Verbindung hängt")

	_, fehler := bau.weg.Zurueckspielen(t.Context(), ziel)
	if fehler == nil {
		t.Fatal("es wurde getauscht, obwohl der Bestand offen blieb")
	}

	if !strings.Contains(fehler.Error(), "schließen") {
		t.Errorf("die Meldung nennt den Schritt nicht: %v", fehler)
	}
}

// **Hält ein anderer Prozess die Datei, wird nichts angefasst** — und die
// Anwendung bekommt ihre Datenbank zurück
// (#961).
func TestEinZweiterProzessHaeltDasZurueckspielenAn(t *testing.T) {
	bau := neuerWegbau(t)
	ziel := eineSicherung(t, bau)

	bau.weg.allein = func(context.Context, string) (func() error, error) {
		return nil, errors.New("ein anderer hält die Datei")
	}

	if _, fehler := bau.weg.Zurueckspielen(t.Context(), ziel); fehler == nil {
		t.Fatal("es wurde getauscht, obwohl ein zweiter Prozess dranhing")
	}

	if bau.leben.geoeffnet != 1 {
		t.Errorf("die Anwendung wurde %d-mal wieder geöffnet — erwartet war einmal",
			bau.leben.geoeffnet)
	}
}

// **Lässt sich der Alleinanspruch nicht zurückgeben, wird es aufgeschrieben** —
// und der Weg läuft weiter.
//
// Er ist ein Griff ins Dateisystem, den niemand mehr braucht; daran das
// Zurückspielen scheitern zu lassen, machte aus einer Nebensache einen
// Abbruch.
func TestEinNichtZurueckgegebenerAnspruchHaeltNichtsAuf(t *testing.T) {
	bau := neuerWegbau(t)
	ziel := eineSicherung(t, bau)

	bau.weg.allein = func(context.Context, string) (func() error, error) {
		return func() error { return errors.New("der Griff klemmt") }, nil
	}

	if _, fehler := bau.weg.Zurueckspielen(t.Context(), ziel); fehler != nil {
		t.Fatalf("zurückspielen: %v", fehler)
	}
}

// **Lässt sich der zurückgespielte Stand nicht öffnen, geht es zurück.**
func TestEinGescheitertesOeffnenDrehtZurueck(t *testing.T) {
	bau := neuerWegbau(t)
	ziel := eineSicherung(t, bau)

	echt := bau.weg.oeffnen
	versuche := 0

	bau.weg.oeffnen = func(ctx context.Context) error {
		versuche++
		if versuche == 1 {
			return errors.New("die zurückgespielte Datei taugt nicht")
		}

		return echt(ctx)
	}

	_, fehler := bau.weg.Zurueckspielen(t.Context(), ziel)
	if fehler == nil {
		t.Fatal("das gescheiterte Öffnen wurde als Erfolg gemeldet")
	}

	// **Der alte Stand steht wieder da.**
	roh, lesefehler := os.ReadFile(bau.weg.orte.Datenbank) // #nosec G304 -- aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("der Bestand ist fort: %v", lesefehler)
	}

	if string(roh) != "laufender bestand" {
		t.Errorf("der Bestand ist nicht der alte: %q", roh)
	}
}

// **Scheitert auch das Zurückdrehen, bleibt der ursprüngliche Grund vorn.**
//
// Wer eine Sicherung einspielt, die sich nicht öffnen lässt, will das lesen —
// und nicht, was beim Aufräumen danach noch geschah.
func TestEinGescheitertesZurueckdrehenNenntBeideGruende(t *testing.T) {
	weg := wegAuf(t, filepath.Join(t.TempDir(), "bestand.db"))

	zurueck := func() error { return errors.New("der Rückweg ist versperrt") }
	grund := errors.New("die zurückgespielte Datei taugt nicht")

	fehler := weg.zurueckDrehen(t.Context(), zurueck, func() {}, grund)
	if !errors.Is(fehler, grund) {
		t.Errorf("der ursprüngliche Grund fehlt: %v", fehler)
	}

	if !strings.Contains(fehler.Error(), "versperrt") {
		t.Errorf("der Grund des Aufräumens fehlt: %v", fehler)
	}
}

// **Eine unlesbare Marke, die es gibt, wird aufgeschrieben.**
//
// Der Zweig daneben — sie fehlt — ist der Normalfall und hält nichts auf. Was
// hier zählt, ist der Unterschied: ein Lesefehler ist ein Befund, kein
// Nichtvorhandensein.
func TestEineUnlesbareMarkeWirdAufgeschrieben(t *testing.T) {
	bau := neuerWegbau(t)

	// **Ein Verzeichnis am Ort der Marke** — `os.ReadFile` gibt darauf einen
	// Fehler, der nicht `os.ErrNotExist` ist.
	if err := os.Mkdir(bau.weg.markenpfad(), 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	// **Der Start läuft trotzdem** — das ist die Zusage.
	bau.weg.RueckspielVollenden()
}

// **Eine nackte `.db`, die sich nicht kopieren lässt, meldet sich.**
func TestEineUnkopierbareDatenbankMeldetSich(t *testing.T) {
	bau := neuerWegbau(t)

	if _, fehler := bau.weg.Zurueckspielen(
		t.Context(), filepath.Join(bau.ort, "gibtesnicht.db")); fehler == nil {
		t.Fatal("die fehlende Datei wurde angenommen")
	}
}

// **Lässt sich der Bestand nicht tauschen, bekommt die Anwendung ihre
// Datenbank zurück.**
//
// Sonst stünde sie ohne da, und der nächste Aufruf meldete etwas ganz anderes
// als den eigentlichen Grund.
func TestEinGescheiterterTauschGibtDieDatenbankZurueck(t *testing.T) {
	bau := neuerWegbau(t)
	ziel := eineSicherung(t, bau)

	// **Ein volles Verzeichnis am Ort der Rückspielsicherung.** Dorthin geht
	// der alte Bestand; `os.Rename` auf ein nicht leeres Verzeichnis
	// scheitert, und damit scheitert der Tausch an seinem ersten Schritt.
	voll := filepath.Join(filepath.Dir(bau.weg.orte.Datenbank), Rueckspielsicherung)
	if err := os.MkdirAll(filepath.Join(voll, "drin"), 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if _, fehler := bau.weg.Zurueckspielen(t.Context(), ziel); fehler == nil {
		t.Fatal("der Tausch wurde als Erfolg gemeldet")
	}

	if bau.leben.geoeffnet != 1 {
		t.Errorf("die Anwendung wurde %d-mal wieder geöffnet — erwartet war einmal",
			bau.leben.geoeffnet)
	}
}

// **Und dasselbe für eine Sicherung ohne Bilder.**
//
// # Warum der Fall doppelt vorkommt
//
// [Bestandsweg.Zurueckspielen] hat **zwei** Ausgänge — einen mit Bildertausch
// und einen ohne —, und beide öffnen wieder. Ein Lauf über nur einen ließe den
// anderen unbetreten; sichtbar wäre das allein an der Abdeckung.
func TestEinGescheitertesOeffnenDrehtAuchOhneBilderZurueck(t *testing.T) {
	bau := neuerWegbau(t)

	// **Eine nackte `.db` aus 0.2.0** — sie sagt über Bilder nichts, und der
	// Weg nimmt deshalb den Ausgang ohne Bildertausch.
	alt := filepath.Join(bau.ort, "alt.db")
	if err := os.WriteFile(alt, []byte("aus 0.2.0"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	echt := bau.weg.oeffnen
	versuche := 0

	bau.weg.oeffnen = func(ctx context.Context) error {
		versuche++
		if versuche == 1 {
			return errors.New("die zurückgespielte Datei taugt nicht")
		}

		return echt(ctx)
	}

	if _, fehler := bau.weg.Zurueckspielen(t.Context(), alt); fehler == nil {
		t.Fatal("das gescheiterte Öffnen wurde als Erfolg gemeldet")
	}

	roh, lesefehler := os.ReadFile(bau.weg.orte.Datenbank) // #nosec G304 -- aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("der Bestand ist fort: %v", lesefehler)
	}

	if string(roh) != "laufender bestand" {
		t.Errorf("der Bestand ist nicht der alte: %q", roh)
	}
}

// **Ein abgebrochener Kontext hält den Rückweg nicht auf.**
//
// # Der Befund
//
// Ab dem Schließen hat die Anwendung keine Datenbank. Wer den Vorgang in
// diesem Augenblick abbricht, will nicht, dass sie **ohne** eine dasteht —
// ein abgebrochener Kontext ließe `storage.Oeffnen` scheitern, und der
// Rückweg scheiterte gleich mit.
//
// Angestrichen als Befund 1 des Zweitblicks über Phase 3.8.
//
// # Wie er hergestellt wird
//
// **Der Abbruch geschieht im Schließen selbst** — das ist der Augenblick, in
// dem er wehtut. Danach läuft alles Weitere mit einem Kontext, den der
// Aufrufer schon aufgegeben hat.
func TestEinAbbruchWaehrendDesTauschsHaeltDenRueckwegNichtAuf(t *testing.T) {
	bau := neuerWegbau(t)
	ziel := eineSicherung(t, bau)

	ctx, abbrechen := context.WithCancel(t.Context())
	t.Cleanup(abbrechen)

	// **Der Weg merkt sich, mit welchem Kontext geöffnet wurde.**
	var geoeffnetMit []error

	echt := bau.weg.oeffnen
	bau.weg.oeffnen = func(innerer context.Context) error {
		geoeffnetMit = append(geoeffnetMit, innerer.Err())

		return echt(innerer)
	}

	schliessenEcht := bau.weg.schliessen
	bau.weg.schliessen = func() error {
		abbrechen()

		return schliessenEcht()
	}

	if _, fehler := bau.weg.Zurueckspielen(ctx, ziel); fehler != nil {
		t.Fatalf("zurückspielen: %v", fehler)
	}

	if len(geoeffnetMit) == 0 {
		t.Fatal("es wurde gar nicht wieder geöffnet")
	}

	for stelle, fehler := range geoeffnetMit {
		if fehler != nil {
			t.Errorf("das %d. Wiederöffnen lief mit einem abgebrochenen Kontext: %v",
				stelle+1, fehler)
		}
	}
}
