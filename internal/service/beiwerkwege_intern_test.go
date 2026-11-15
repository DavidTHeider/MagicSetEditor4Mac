// **Die Rückwege des Beiwerks** (#1162).
//
// Sie stehen für sich, weil sie eine eigene Vorrichtung brauchen: ein
// Bilderverzeichnis mit **Abstand** zum Bestand. Warum, steht bei
// [beiwerkAnwendung].
package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// beiwerkAnwendung gibt eine Anwendung, deren Bilder **nicht** neben dem
// Bestand liegen.
//
// **Der Abstand ist die Vorrichtung**: nur so lässt sich das Bilderverzeichnis
// zusperren, ohne zugleich den Zwischenplatz des Auspackens zu treffen — und
// ein Lauf, der am falschen Zwischenschritt scheitert, prüft den falschen
// Zweig, obwohl er grün aussieht (§10).
func beiwerkAnwendung(t *testing.T) (bau *wegbau, unter, bilder string) {
	t.Helper()

	bau = neuerWegbau(t)
	unter = filepath.Join(bau.ort, "unter")
	bilder = filepath.Join(unter, "bilder")

	if err := os.MkdirAll(bilder, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	bau.weg.orte.Bilder = func() (string, error) { return bilder, nil }

	return bau, unter, bilder
}

// zusperren nimmt einem Verzeichnis das Schreibrecht — und gibt es zurück.
func zusperren(t *testing.T, ort string) {
	t.Helper()

	// #nosec G302 -- ein Verzeichnis braucht das Ausführrecht.
	if err := os.Chmod(ort, 0o500); err != nil {
		t.Fatalf("zusperren: %v", err)
	}

	// #nosec G302 -- ein Verzeichnis braucht das Ausführrecht.
	t.Cleanup(func() { _ = os.Chmod(ort, 0o700) }) //nolint:errcheck // Aufräumen
}

// **Lässt sich das Bilderverzeichnis nicht beiseitelegen, wird es gemeldet.**
func TestEinUnbeweglichesBilderverzeichnisMeldetSich(t *testing.T) {
	bau, unter, _ := beiwerkAnwendung(t)
	weg := bau.weg

	zusperren(t, unter)

	_, fehler := weg.beiwerkErsetzen(t.TempDir())
	if fehler == nil {
		t.Fatal("das unbewegliche Verzeichnis wurde übergangen")
	}

	if !strings.Contains(fehler.Error(), "beiseitelegen") {
		t.Errorf("die Meldung nennt den Schritt nicht: %v", fehler)
	}
}

// **Und ein alter Rückspielstand, der sich nicht wegräumen lässt, ebenso.**
//
// Er liegt vom vorigen Zurückspielen da. Bleibt er liegen, weil das
// Verzeichnis darüber zu ist, darf nicht der **nächste** Stand an seine
// Stelle rutschen.
func TestEinUnraeumbarerRueckspielstandMeldetSich(t *testing.T) {
	bau, unter, _ := beiwerkAnwendung(t)
	weg := bau.weg

	if err := os.MkdirAll(filepath.Join(unter, Beiwerkbeiseite, "7"), 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	zusperren(t, unter)

	_, fehler := weg.beiwerkErsetzen(t.TempDir())
	if fehler == nil {
		t.Fatal("der alte Rückspielstand wurde übergangen")
	}

	if !strings.Contains(fehler.Error(), "wegräumen") {
		t.Errorf("die Meldung nennt den Schritt nicht: %v", fehler)
	}
}

// **Kommen die Bilder nicht an ihren Ort, geht der alte Stand zurück.**
//
// Der Fall entsteht, wenn der Zwischenplatz zwischen Auspacken und Einsetzen
// verschwindet — ein Aufräumdienst, eine volle Platte. Geprüft wird nicht die
// Meldung, sondern **dass danach wieder der laufende Stand dasteht**.
func TestEinGescheitertesEinsetzenHoltDenAltenStandZurueck(t *testing.T) {
	bau, _, bilder := beiwerkAnwendung(t)
	weg := bau.weg

	if err := os.WriteFile(filepath.Join(bilder, "alt.png"),
		[]byte("der laufende Stand"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	_, fehler := weg.beiwerkErsetzen(filepath.Join(t.TempDir(), "gibt-es-nicht"))
	if fehler == nil {
		t.Fatal("das Einsetzen aus dem Nichts kam durch")
	}

	roh, lesefehler := os.ReadFile(filepath.Join(bilder, "alt.png")) // #nosec G304 -- aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("der laufende Stand ist fort: %v", lesefehler)
	}

	if string(roh) != "der laufende Stand" {
		t.Errorf("der laufende Stand kam anders zurück: %q", roh)
	}
}

// **Gab es vorher kein Bilderverzeichnis, räumt der Rückweg es wieder weg.**
//
// Sonst bliebe nach einem gescheiterten Zurückspielen ein Ordner stehen, den
// es vorher nicht gab — und beim nächsten Sichern wanderte er mit.
func TestDerRueckwegRaeumtEinVorherNichtVorhandenesVerzeichnis(t *testing.T) {
	bau, _, bilder := beiwerkAnwendung(t)
	weg := bau.weg

	if err := os.RemoveAll(bilder); err != nil {
		t.Fatalf("wegräumen: %v", err)
	}

	neu := filepath.Join(t.TempDir(), "beiwerk")
	if err := os.MkdirAll(neu, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	zurueck, fehler := weg.beiwerkErsetzen(neu)
	if fehler != nil {
		t.Fatalf("ersetzen: %v", fehler)
	}

	if _, err := os.Stat(bilder); err != nil {
		t.Fatalf("die Bilder kamen nicht an: %v", err)
	}

	if err := zurueck(); err != nil {
		t.Fatalf("zurückdrehen: %v", err)
	}

	if _, err := os.Stat(bilder); !os.IsNotExist(err) {
		t.Errorf("das Verzeichnis steht noch da: %v", err)
	}
}

// **Und ein Rückweg, der selbst nicht kann, sagt es.**
//
// Er wird nicht verschluckt: `SicherungZurueckspielen` fügt ihn dem
// ursprünglichen Fehler bei, damit im Protokoll steht, dass **zwei** Dinge
// schiefgingen.
func TestEinGescheiterterBeiwerkRueckwegMeldetSich(t *testing.T) {
	bau, unter, _ := beiwerkAnwendung(t)
	weg := bau.weg

	neu := filepath.Join(t.TempDir(), "beiwerk")
	if err := os.MkdirAll(neu, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	zurueck, fehler := weg.beiwerkErsetzen(neu)
	if fehler != nil {
		t.Fatalf("ersetzen: %v", fehler)
	}

	zusperren(t, unter)

	if err := zurueck(); err == nil {
		t.Fatal("der gescheiterte Rückweg meldete Erfolg")
	}
}
