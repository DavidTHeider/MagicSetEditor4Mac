//go:build darwin

// **Die Fehlerwege der Platzzählung** — Werkstück 4 von Phase 4.13
// (#1352).
//
// **Eigene Datei neben `platz_test.go`.** Der Schnitt folgt der Frage: dort
// steht, **was gezählt wird**, hier, **was geschieht, wenn es nicht geht**.

package app_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ohne Zuhause gibt es kein Bilderverzeichnis — und keine Zählung.**
//
// **Nicht „null Byte"**: wer nicht sagen kann, wo etwas liegt, hat es nicht
// gezählt (§8).
func TestOhneZuhauseGibtEsKeineZaehlung(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "bestand.db")

	// **Der Bestand steht, das Zuhause fällt weg** — genau die Lage, in der
	// `os.UserConfigDir` aufgibt.
	if err := os.WriteFile(ort, make([]byte, 16), 0o600); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	t.Setenv("HOME", "")

	if antwort := app.PlatzflaecheUeber(ort).Platz(t.Context()); antwort.Fehler == nil {
		t.Error("ohne Zuhause kam eine Zählung")
	}
}

// **Ein Verzeichnis, das es nicht gibt, lässt sich nicht auflisten.**
func TestEinFehlendesVerzeichnisReisstDieZaehlungAb(t *testing.T) {
	t.Parallel()

	// **Kein `t.TempDir()`**: das gibt es. Ein Ort darunter, den niemand
	// angelegt hat, ist der Fall.
	ort := filepath.Join(t.TempDir(), "gibtsnicht", "bestand.db")

	if antwort := app.PlatzflaecheUeber(ort).Platz(t.Context()); antwort.Fehler == nil {
		t.Error("ein fehlendes Verzeichnis gab eine Zählung")
	}
}

// **Und ein Bestand ohne seine Begleiter wird trotzdem gezählt.**
//
// **`-wal` und `-shm` gibt es nur, solange jemand schreibt.** Wer ihr Fehlen
// als Fehler nähme, könnte einen geschlossenen Bestand nie zählen.
func TestEinBestandOhneBegleiterWirdGezaehlt(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)

	if fehler := anwendung.Beenden(); fehler != nil {
		t.Fatalf("beenden: %v", fehler)
	}

	// **Und ein Bild, damit auch das Verzeichnis etwas hergibt.**
	bilder := filepath.Join(filepath.Dir(bestand), "beiwerk")
	if err := os.MkdirAll(bilder, 0o700); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	if err := os.WriteFile(filepath.Join(bilder, "karte.png"),
		make([]byte, 128), 0o600); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	antwort := app.PlatzflaecheUeber(bestand).Platz(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	nach := map[string]app.PlatzpostenDTO{}
	for _, posten := range antwort.Posten {
		nach[posten.Posten] = posten
	}

	if nach[string(domain.PostenBestand)].Dateien != 1 {
		t.Errorf("der Bestand zählt %d Dateien",
			nach[string(domain.PostenBestand)].Dateien)
	}

	if nach[string(domain.PostenBilder)].Bytes != 128 {
		t.Errorf("die Bilder wiegen %d Byte", nach[string(domain.PostenBilder)].Bytes)
	}
}

// **Ein Wegräumen, das an den Rechten scheitert, wird gemeldet** (§8).
func TestEinGescheitertesWegraeumenWirdGemeldet(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)

	neben := filepath.Dir(bestand)
	for _, name := range []string{"bestand-vor-0.1.0.db", "bestand-vor-0.2.0.db"} {
		if err := os.WriteFile(filepath.Join(neben, name),
			make([]byte, 16), 0o600); err != nil {
			t.Fatalf("vorbereiten: %v", err)
		}
	}

	// **Das Verzeichnis wird schreibgeschützt** — dann bleibt die Datei
	// liegen, und `os.Remove` sagt es.
	//nolint:gosec // G302: der Schreibschutz IST der Fall — ein Verzeichnis
	// mit 0600 ließe sich nicht betreten, und dann prüfte der Lauf etwas
	// anderes
	if err := os.Chmod(neben, 0o500); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	//nolint:gosec,errcheck // G302: dasselbe beim Aufräumen — ohne
	// Betretungsrecht räumt `t.TempDir` nicht ab
	t.Cleanup(func() { _ = os.Chmod(neben, 0o700) })

	geraeumt := anwendung.PlatzFreigeben(t.Context(),
		string(domain.PostenVorsicherungen))

	if geraeumt.Fehler == nil {
		t.Error("das Wegräumen glückte in einem schreibgeschützten Verzeichnis")
	}
}

// **Und beides steht auch an den Bindungen des Fensters** (§16).
func TestDieBindungenGebenDenPlatzWeiter(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)

	bindungen := app.NeueBindungen(anwendung)

	if antwort := bindungen.Platz(t.Context()); antwort.Fehler != nil {
		t.Errorf("Platz: %+v", antwort.Fehler)
	}

	if geraeumt := bindungen.PlatzFreigeben(t.Context(),
		string(domain.PostenVorsicherungen)); geraeumt.Fehler != nil {
		t.Errorf("PlatzFreigeben: %+v", geraeumt.Fehler)
	}
}

// **Ein Paketort, der als Datei dasteht, reißt die Zählung ab.**
//
// # Warum genau dieser Fall
//
// **`Paketverzeichnis` legt an, `Bilderverzeichnis` nicht** — der Grund steht
// dort: ein Ordner, den die allermeisten nie füllen, soll nicht auf jedem
// Rechner entstehen. Damit ist der Paketort der einzige der beiden, dessen
// **Anlegen** scheitern kann, und eine Datei an seiner Stelle ist die
// einfachste Art, das zu erreichen.
func TestEinPaketortAlsDateiReisstDieZaehlungAb(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)

	if fehler := anwendung.Beenden(); fehler != nil {
		t.Fatalf("beenden: %v", fehler)
	}

	// **Der Ort steht schon als Datei da** — `MkdirAll` gibt dann auf.
	if err := os.RemoveAll(filepath.Join(filepath.Dir(bestand), "pakete")); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	if err := os.WriteFile(filepath.Join(filepath.Dir(bestand), "pakete"),
		make([]byte, 4), 0o600); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	if antwort := app.PlatzflaecheUeber(bestand).Platz(t.Context()); antwort.Fehler == nil {
		t.Error("ein Paketort als Datei gab eine Zählung")
	}
}

// **Und ein Wegräumen, das schon beim Nachsehen scheitert, ebenso.**
//
// **Ein anderer Weg als der über die Rechte am Löschen**: dort steht die Liste
// schon da, hier lässt sie sich gar nicht erst holen.
func TestEinWegraeumenOhneListeWirdGemeldet(t *testing.T) {
	t.Parallel()

	neben := filepath.Join(t.TempDir(), "verschlossen")
	if err := os.Mkdir(neben, 0o700); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	bestand := filepath.Join(neben, "bestand.db")
	if err := os.WriteFile(bestand, make([]byte, 16), 0o600); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	if err := os.Chmod(neben, 0o000); err != nil {
		t.Fatalf("vorbereiten: %v", err)
	}

	//nolint:gosec,errcheck // G302: dasselbe beim Aufräumen — ohne
	// Betretungsrecht räumt `t.TempDir` nicht ab
	t.Cleanup(func() { _ = os.Chmod(neben, 0o700) })

	geraeumt := app.PlatzflaecheUeber(bestand).PlatzFreigeben(
		t.Context(), string(domain.PostenVorsicherungen))

	if geraeumt.Fehler == nil {
		t.Error("ein unlesbares Verzeichnis ließ sich wegräumen")
	}
}

// **Ein halber Lauf meldet den halben Stand — nicht den alten.**
//
// # Der Befund
//
// **Befund 4 des Zweitblicks über Phase 4.13** (2026-09-07): `PlatzFreigeben`
// löscht nacheinander. Scheitert ein späteres `Remove`, sind die früheren
// **schon fort** — die Antwort trug aber eine leere Liste und einen Fehler.
// Das Fenster zeigte den Fehler und behielt seine alte Zählung: eine Zahl, die
// nicht mehr stimmt, und niemand sieht es.
//
// **Das ist §7 Regel 3**: geschrieben ist geschrieben. Was gelöscht wurde,
// wird durch einen späteren Fehler nicht ungelöscht.
//
// # Warum es eine Naht braucht
//
// Der Fall „das **dritte** Löschen scheitert" lässt sich mit Dateirechten
// nicht herstellen: alle Vorsicherungen liegen im selben Verzeichnis, und ein
// unschreibbares Verzeichnis lässt schon das erste scheitern. Der vorhandene
// Lauf entzog dem ganzen Ordner die Rechte und betrat diesen Fall nie.
func TestEinHalberLaufMeldetDenHalbenStand(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)

	if fehler := anwendung.Beenden(); fehler != nil {
		t.Fatalf("beenden: %v", fehler)
	}

	ort := filepath.Dir(bestand)

	// **Vier Sicherungen**: zwei bleiben (jüngste und jüngste lesbare), zwei
	// sollen weg — und beim zweiten davon geht es schief.
	for _, name := range []string{
		"bestand-vor-0.1.0.db", "bestand-vor-0.2.0.db",
		"bestand-vor-0.3.0.db", "bestand-vor-0.4.0.db",
	} {
		if fehler := os.WriteFile(
			filepath.Join(ort, name), make([]byte, 256), 0o600); fehler != nil {
			t.Fatalf("vorbereiten: %v", fehler)
		}
	}

	flaeche := app.PlatzflaecheUeber(bestand)

	geloescht := 0
	flaeche.MitDateientferner(func(pfad string) error {
		geloescht++

		if geloescht == 2 {
			return errors.New("planmäßig gescheitert")
		}

		return os.Remove(pfad)
	})

	antwort := flaeche.PlatzFreigeben(t.Context(), string(domain.PostenVorsicherungen))

	if antwort.Fehler == nil {
		t.Fatal("der Lauf glückte, obwohl das zweite Löschen scheiterte")
	}

	// **Und trotzdem steht die Zählung da** — mit dem, was wirklich noch liegt.
	if len(antwort.Posten) == 0 {
		t.Fatal("die Antwort trägt keine Zählung — das Fenster behielte die alte")
	}

	for _, posten := range antwort.Posten {
		if posten.Posten != string(domain.PostenVorsicherungen) {
			continue
		}

		// **Eine ist fort**, also drei von vier.
		if posten.Dateien != 3 {
			t.Errorf("gezählt sind %d Sicherungen, auf der Platte liegen 3",
				posten.Dateien)
		}
	}
}
