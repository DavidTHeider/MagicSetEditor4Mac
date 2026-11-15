package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/aktualisierung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// freigabeprobe gibt eine feste Freigabe — oder einen Fehler.
type freigabeprobe struct {
	freigabe   aktualisierung.Freigabe
	fehler     error
	ladefehler error
	gefragt    int
	geladen    int

	// **Für die Verträglichkeitsprüfung** (O6, #1430): das Manifest, das
	// diese Freigabe über sich sagt, und was beim Holen schiefgeht.
	manifest       domain.Fassungsmanifest
	manifestfehler error
	summen         []byte
	summenfehler   error

	// waehrenddessen läuft **im** Ladevorgang — für den Riegel aus §7,
	// Regel 4: ein zweiter Aufruf, der erst danach käme, prüfte ihn nicht.
	waehrenddessen func()
}

func (f *freigabeprobe) NeuesteFreigabe(context.Context) (aktualisierung.Freigabe, error) {
	f.gefragt++

	return f.freigabe, f.fehler
}

// **Ohne Einstellung geht nichts hinaus** — §16, Regel 6.
//
// Das ist die wichtigste Zusage dieses Wegs, und sie ist von außen nur an
// einer Sache zu sehen: **der Klient wurde nicht gefragt**. Ein Test, der nur
// die Antwort prüft, sähe „nichts Neues" und wäre zufrieden.
func TestOhneEinstellungWirdNichtGefragt(t *testing.T) {
	t.Parallel()

	probe := &freigabeprobe{freigabe: aktualisierung.Freigabe{Fassung: "99.0.0"}}
	anwendung := mitFassungen(t, probe, "")

	antwort := anwendung.NachNeuerFassungSehen(t.Context())

	if probe.gefragt != 0 {
		t.Errorf("der Klient wurde %d-mal gefragt", probe.gefragt)
	}

	if antwort.Gefragt {
		t.Error("die Antwort sagt, es sei gefragt worden")
	}

	// **„Nicht gefragt" ist nicht „nichts Neues".** Ohne diese Unterscheidung
	// sähe eine abgeschaltete Prüfung aus wie eine, die nichts gefunden hat.
	if antwort.Neuer {
		t.Error("die Antwort meldet eine neuere Fassung, ohne gefragt zu haben")
	}
}

// **Und mit „aus" ebenso wenig** — der ausdrücklich gewählte Fall.
func TestMitAusWirdNichtGefragt(t *testing.T) {
	t.Parallel()

	probe := &freigabeprobe{freigabe: aktualisierung.Freigabe{Fassung: "99.0.0"}}
	anwendung := mitFassungen(t, probe, domain.AktualisierungAus)

	if anwendung.NachNeuerFassungSehen(t.Context()).Gefragt || probe.gefragt != 0 {
		t.Error("mit der Stufe aus wurde gefragt")
	}
}

// **Mit „melden" wird gefragt, und eine neuere Fassung kommt an.**
func TestMitMeldenKommtDieNeuereFassungAn(t *testing.T) {
	t.Parallel()

	probe := &freigabeprobe{freigabe: aktualisierung.Freigabe{
		Fassung: "99.0.0", Seite: "https://beispiel.invalid/v99",
	}}
	anwendung := mitFassungen(t, probe, domain.AktualisierungMelden)

	antwort := anwendung.NachNeuerFassungSehen(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if !antwort.Gefragt || !antwort.Neuer {
		t.Errorf("gefragt=%v neuer=%v", antwort.Gefragt, antwort.Neuer)
	}

	if antwort.Fassung != "99.0.0" || antwort.Seite == "" {
		t.Errorf("Fassung=%q Seite=%q", antwort.Fassung, antwort.Seite)
	}
}

// **Die laufende Fassung ist nicht neuer als sie selbst.**
func TestDieLaufendeFassungIstNichtNeuer(t *testing.T) {
	t.Parallel()

	probe := &freigabeprobe{freigabe: aktualisierung.Freigabe{Fassung: app.Fassung}}
	anwendung := mitFassungen(t, probe, domain.AktualisierungMelden)

	antwort := anwendung.NachNeuerFassungSehen(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if !antwort.Gefragt {
		t.Error("es wurde nicht gefragt")
	}

	if antwort.Neuer {
		t.Errorf("%q gilt als neuer als sich selbst", app.Fassung)
	}
}

// **Ein Fehler des fremden Dienstes hält nichts an** (§16, Satz 5) — er steht
// in der Antwort, und die Anwendung arbeitet weiter.
func TestEinFehlerDesDienstesStehtInDerAntwort(t *testing.T) {
	t.Parallel()

	probe := &freigabeprobe{fehler: errors.New("kein Netz")}
	anwendung := mitFassungen(t, probe, domain.AktualisierungMelden)

	antwort := anwendung.NachNeuerFassungSehen(t.Context())
	if antwort.Fehler == nil {
		t.Fatal("der Fehler kam nicht an")
	}

	if !antwort.Gefragt {
		t.Error("gefragt wurde er trotzdem — das gehört in die Antwort")
	}
}

// **Eine unlesbare Fassung wird abgewiesen und nicht als „neuer" gedeutet.**
func TestEineUnlesbareFassungMeldetNichtNeuer(t *testing.T) {
	t.Parallel()

	probe := &freigabeprobe{freigabe: aktualisierung.Freigabe{Fassung: "keine"}}
	anwendung := mitFassungen(t, probe, domain.AktualisierungMelden)

	antwort := anwendung.NachNeuerFassungSehen(t.Context())
	if antwort.Fehler == nil {
		t.Fatal("eine unlesbare Fassung ging durch")
	}

	if antwort.Neuer {
		t.Error("sie gilt als neuer")
	}
}

// mitFassungen baut eine Anwendung mit gesetzter Wahl und gesetztem Zugang.
func mitFassungen(t *testing.T, anleger app.Fassungsanleger, wahl string) *app.App {
	t.Helper()

	anwendung := app.NeueApp("")
	anwendung.MitEinstellungen(&einstellungsprobe{
		werte: map[string]string{domain.EinstellungAktualisierung: wahl},
	})
	anwendung.MitFassungen(anleger)

	return anwendung
}

// **Ein unlesbarer Bestand meldet sich, statt zu schweigen.**
//
// Ohne Einstellungen ist nicht zu sagen, ob gefragt werden darf — und „nicht
// zu sagen" ist etwas anderes als „aus" (§8).
func TestOhneEinstellungenKommtEinFehler(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitEinstellungen(&einstellungsprobe{fehler: errors.New("kein Bestand")})
	anwendung.MitFassungen(&freigabeprobe{})

	if antwort := anwendung.NachNeuerFassungSehen(t.Context()); antwort.Fehler == nil {
		t.Error("ein unlesbarer Bestand ging als nicht gefragt durch")
	}
}

// **Ohne eingerichteten Zugang wird das gesagt und nicht geschwiegen.**
//
// Der Fall entsteht, wenn jemand eine Anwendung ohne Composition Root baut —
// im Fenster gibt es ihn nicht. Ein stilles „nichts Neues" wäre trotzdem die
// falsche Antwort: die Einstellung sagt „melden", und gemeldet wurde nichts.
func TestOhneZugangKommtEinFehler(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitEinstellungen(&einstellungsprobe{
		werte: map[string]string{domain.EinstellungAktualisierung: domain.AktualisierungMelden},
	})

	if antwort := anwendung.NachNeuerFassungSehen(t.Context()); antwort.Fehler == nil {
		t.Error("ohne Zugang kam keine Meldung")
	}
}

// **Das Manifest und die signierte Liste** — beides am selben Anleger, wie im
// Betrieb (O6, #1430).
func (f *freigabeprobe) Manifest(
	context.Context, aktualisierung.Freigabe, []byte,
) (domain.Fassungsmanifest, error) {
	return f.manifest, f.manifestfehler
}

func (f *freigabeprobe) Pruefsummen(
	context.Context, aktualisierung.Freigabe,
) ([]byte, error) {
	return f.summen, f.summenfehler
}

// **Die Antwort sagt, ob geladen werden darf** — gefunden am 2026-10-17 beim
// Handbuchbild (AP-12, #2369): unter „melden“ bot das Band „Herunterladen“
// an, und [App.NeueFassungLaden] wies es ab. Die Grenze bleibt hier; das
// Fenster zeigt den Knopf nur, wenn sie es erlaubt.
func TestDieAntwortSagtObGeladenWerdenDarf(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		stufe string
		holen bool
	}{
		{domain.AktualisierungMelden, false},
		{domain.AktualisierungHolen, true},
	} {
		probe := &freigabeprobe{freigabe: aktualisierung.Freigabe{
			Fassung: "99.0.0", Seite: "https://beispiel.invalid/v99",
		}}
		antwort := mitFassungen(t, probe, fall.stufe).NachNeuerFassungSehen(t.Context())

		if antwort.Fehler != nil {
			t.Fatalf("%s: %+v", fall.stufe, antwort.Fehler)
		}

		if antwort.Holen != fall.holen {
			t.Errorf("%s: holen=%v, erwartet %v", fall.stufe, antwort.Holen, fall.holen)
		}
	}
}
