// **Die Rückwege des Zurückspielens** (#940).
//
// Sie brauchen Nähte, die von außen nicht zu setzen sind — eine Datenbank, die
// sich nicht schließen lässt, und eine, die sich nicht wieder öffnen lässt.
// Deshalb steht diese Datei im Paket.
package app

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// gepruefteSicherung gibt einen Prüfer, der alles durchwinkt.
//
// **Die Prüfung selbst steht in `internal/storage` und ist dort belegt.** Hier
// geht es um das, was **danach** kommt — und dafür muss davor etwas
// durchkommen.
func gepruefteSicherung() Sicherungspruefer {
	return func(context.Context, string) (Sicherungsbefund, error) {
		return Sicherungsbefund{AngelegtAm: "2026-08-23T00:00:00.000Z"}, nil
	}
}

// vorbereitet baut eine Anwendung mit echter Datei und durchwinkender Prüfung.
func vorbereitet(t *testing.T) (anwendung *App, bestand, sicherung string) {
	t.Helper()

	ort := t.TempDir()
	bestand = filepath.Join(ort, "bestand.db")
	sicherung = filepath.Join(ort, "sicherung.db")

	anwendung = NeueApp(bestand)
	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	anwendung.MitSicherungspruefer(gepruefteSicherung())

	return anwendung, bestand, sicherung
}

// **Lässt sich der Bestand nicht schließen, wird nicht getauscht.**
func TestSicherungZurueckspielenBrichtAbWennDasSchliessenScheitert(t *testing.T) {
	anwendung, bestand, sicherung := vorbereitet(t)

	// **Eine Datenbank, deren Verbindung sich nicht schließen lässt** — der
	// Treiber aus `verdrahtung_intern_test.go`. Bis #1485 stand hier ein
	// Standleser, dessen Leihgabe klemmte; der hält nichts mehr fest.
	if fehler := anwendung.db.Close(); fehler != nil {
		t.Fatalf("den echten Bestand schließen: %v", fehler)
	}

	klemmend, err := sql.Open("app-nicht-schliessbar", "")
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	// Erst eine Verbindung im Vorrat lässt `Close` scheitern — ohne sie gäbe
	// es nichts zu schließen.
	if fehler := klemmend.PingContext(t.Context()); fehler != nil {
		t.Fatalf("anpingen: %v", fehler)
	}

	anwendung.db = klemmend

	if err := os.WriteFile(sicherung, []byte("egal"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	antwort := anwendung.SicherungZurueckspielen(t.Context(), sicherung, "")
	if antwort.Fehler == nil {
		t.Fatal("der Fehler beim Schließen wurde übergangen")
	}

	// **Der Bestand liegt unangetastet da.**
	if _, err := os.Stat(bestand); err != nil {
		t.Errorf("der Bestand ist fort: %v", err)
	}

	if _, err := os.Stat(filepath.Join(filepath.Dir(bestand), Rueckspielsicherung)); !os.IsNotExist(err) {
		t.Error("es wurde beiseitegelegt, obwohl gar nicht getauscht wurde")
	}
}

// **Scheitert der Tausch, bekommt die Anwendung ihre Datenbank zurück.**
func TestSicherungZurueckspielenOeffnetNachEinemGescheitertenTauschWieder(t *testing.T) {
	anwendung, _, sicherung := vorbereitet(t)

	// Eine „Sicherung", die ein Verzeichnis ist: lesbar geöffnet, nicht
	// übertragbar.
	if err := os.MkdirAll(filepath.Join(sicherung, "inhalt"), 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if antwort := anwendung.SicherungZurueckspielen(t.Context(), sicherung, ""); antwort.Fehler == nil {
		t.Fatal("der gescheiterte Tausch wurde als Erfolg gemeldet")
	}

	// **Und sie läuft weiter** — das ist die eigentliche Zusage.
	if antwort := anwendung.Einstellungen(t.Context()); antwort.Fehler != nil {
		t.Fatalf("die Anwendung hat keine Datenbank mehr: %+v", antwort.Fehler)
	}

	t.Cleanup(func() { _ = anwendung.Beenden() }) //nolint:errcheck // Aufräumen
}

// **Lässt sich der zurückgespielte Stand nicht öffnen, geht der Rückweg zurück.**
func TestSicherungZurueckspielenDrehtZurueckWennDasOeffnenScheitert(t *testing.T) {
	anwendung, bestand, sicherung := vorbereitet(t)

	if err := os.WriteFile(sicherung, []byte("keine Datenbank"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	echt := anwendung.oeffnen
	versuche := 0

	anwendung.oeffnen = func(ctx context.Context, quelle string) (*sql.DB, error) {
		versuche++
		if versuche == 1 {
			return nil, errors.New("die zurückgespielte Datei taugt nicht")
		}

		return echt(ctx, quelle)
	}

	antwort := anwendung.SicherungZurueckspielen(t.Context(), sicherung, "")
	if antwort.Fehler == nil {
		t.Fatal("die unbrauchbare Datei wurde übernommen")
	}

	// **Der alte Stand steht wieder am Ort des Bestands** — und die Anwendung
	// läuft mit ihm.
	roh, err := os.ReadFile(bestand) // #nosec G304 -- Testpfad aus t.TempDir()
	if err != nil {
		t.Fatalf("der Bestand ist fort: %v", err)
	}

	if strings.HasPrefix(string(roh), "keine Datenbank") {
		t.Error("am Ort des Bestands steht noch die unbrauchbare Datei")
	}

	if antwort := anwendung.Einstellungen(t.Context()); antwort.Fehler != nil {
		t.Fatalf("die Anwendung hat keine Datenbank mehr: %+v", antwort.Fehler)
	}

	t.Cleanup(func() { _ = anwendung.Beenden() }) //nolint:errcheck // Aufräumen
}

// **Und wenn auch der Rückweg scheitert, steht es im Protokoll.**
//
// Der Fall ist keine Erfindung: er entsteht, wenn zwischen Tausch und
// Rücknahme jemand die beiseitegelegte Datei entfernt. Hergestellt wird er
// hier, indem genau das im Augenblick des gescheiterten Öffnens geschieht.
func TestSicherungZurueckspielenMeldetEinenGescheitertenRueckweg(t *testing.T) {
	anwendung, bestand, sicherung := vorbereitet(t)

	if err := os.WriteFile(sicherung, []byte("keine Datenbank"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	beiseite := filepath.Join(filepath.Dir(bestand), Rueckspielsicherung)

	anwendung.oeffnen = func(context.Context, string) (*sql.DB, error) {
		if fehler := os.RemoveAll(beiseite); fehler != nil {
			t.Errorf("beiseitegelegte Datei entfernen: %v", fehler)
		}

		return nil, errors.New("die zurückgespielte Datei taugt nicht")
	}

	antwort := anwendung.SicherungZurueckspielen(t.Context(), sicherung, "")
	if antwort.Fehler == nil {
		t.Fatal("der doppelte Fehlschlag wurde als Erfolg gemeldet")
	}
}

// **Hält ein zweiter Prozess die Datei, wird nicht getauscht**
// (#961).
//
// Die zweite Verbindung steht hier für den zweiten **Prozess** — für SQLite
// ist beides dasselbe, es sperrt über die Datei.
func TestSicherungZurueckspielenWeistEinenZweitenVerbraucherAb(t *testing.T) {
	anwendung, bestand, sicherung := vorbereitet(t)

	if err := os.WriteFile(sicherung, []byte("egal"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	// Ein zweiter Verbraucher, der die Datei offen hält — wie `mse-mcp`.
	zweiter, err := storage.FuerLaeufe("file:" + bestand)
	if err != nil {
		t.Fatalf("die zweite Verbindung öffnen: %v", err)
	}

	//nolint:rowserrcheck // die Zeilen sind die Vorrichtung, nicht das Ergebnis:
	// sie bleiben offen, damit die Datei belegt ist.
	zeilen, err := zweiter.QueryContext(probekontext(t), `SELECT name FROM migrationen`)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	t.Cleanup(func() {
		_ = zeilen.Close()  //nolint:errcheck // nur zum Aufräumen
		_ = zweiter.Close() //nolint:errcheck // nur zum Aufräumen
	})

	antwort := anwendung.SicherungZurueckspielen(t.Context(), sicherung, "")
	if antwort.Fehler == nil {
		t.Fatal("getauscht wurde, obwohl ein zweiter Verbraucher die Datei hält")
	}

	// **Der Bestand ist unangetastet**, und die Anwendung läuft weiter.
	if _, err := os.Stat(filepath.Join(filepath.Dir(bestand), Rueckspielsicherung)); !os.IsNotExist(err) {
		t.Error("es wurde beiseitegelegt, obwohl gar nicht getauscht wurde")
	}

	if antwort := anwendung.Einstellungen(t.Context()); antwort.Fehler != nil {
		t.Fatalf("die Anwendung hat keine Datenbank mehr: %+v", antwort.Fehler)
	}
}

// **Die Naht lässt sich setzen** — für einen Test, der den Fall nachstellt,
// ohne eine zweite Verbindung zu brauchen.
func TestMitAlleinpruefungSetztDieNaht(t *testing.T) {
	anwendung, _, sicherung := vorbereitet(t)
	anwendung.MitAlleinpruefung(func(context.Context, string) (func() error, error) {
		return nil, errors.New("jemand hält die Datei")
	})

	if err := os.WriteFile(sicherung, []byte("egal"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if antwort := anwendung.SicherungZurueckspielen(t.Context(), sicherung, ""); antwort.Fehler == nil {
		t.Fatal("die gesetzte Prüfung wurde übergangen")
	}
}

// **Und der Anspruch wird zurückgegeben, auch wenn der Tausch scheitert.**
//
// Er gilt über den ganzen Tausch (#1163, Befund Z6-02). Bliebe er auf einem
// Fehlerweg liegen, stünde er dem **eigenen** nächsten Öffnen im Weg — `flock`
// gilt für die offene Datei und nicht für den Prozess.
func TestDerAlleinanspruchWirdAufJedemWegZurueckgegeben(t *testing.T) {
	anwendung, _, sicherung := vorbereitet(t)

	if err := os.WriteFile(sicherung, []byte("keine Datenbank"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	var zurueckgegeben bool

	anwendung.MitAlleinpruefung(func(context.Context, string) (func() error, error) {
		return func() error {
			zurueckgegeben = true

			return nil
		}, nil
	})

	// **Ein Tausch, der scheitert**: die Datei ist keine Datenbank.
	anwendung.oeffnen = func(context.Context, string) (*sql.DB, error) {
		return nil, errors.New("die zurückgespielte Datei taugt nicht")
	}

	if antwort := anwendung.SicherungZurueckspielen(t.Context(), sicherung, ""); antwort.Fehler == nil {
		t.Fatal("die unbrauchbare Datei wurde übernommen")
	}

	if !zurueckgegeben {
		t.Error("der Alleinanspruch blieb liegen — er stünde dem nächsten Öffnen im Weg")
	}
}

// **Bleibt er liegen, steht es im Protokoll.**
//
// Es ist ein Befund über eine Nachbardatei und kein Grund, den Vorgang zu
// verwerfen: der Tausch ist an dieser Stelle schon geschehen.
func TestEinGescheitertesZurueckgebenStehtImProtokoll(t *testing.T) {
	anwendung, _, sicherung := vorbereitet(t)

	// **Eine echte Sicherung**: der Tausch muss gelingen, damit der Lauf vom
	// Zurückgeben handelt und nicht von einer unbrauchbaren Datei.
	if err := anwendung.bestandssicherer.Sichern(t.Context(), sicherung); err != nil {
		t.Fatalf("sichern: %v", err)
	}

	anwendung.MitAlleinpruefung(func(context.Context, string) (func() error, error) {
		return func() error { return errors.New("die Sperre klemmt") }, nil
	})

	if antwort := anwendung.SicherungZurueckspielen(t.Context(), sicherung, ""); antwort.Fehler != nil {
		t.Fatalf("das Zurückspielen scheiterte an der Sperre: %+v", antwort.Fehler)
	}

	t.Cleanup(func() { _ = anwendung.Beenden() }) //nolint:errcheck // Aufräumen
}
