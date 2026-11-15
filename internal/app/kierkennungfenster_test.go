package app_test

// **Die Kartenerkennung am Fenster** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// # Warum eine eigene Datei
//
// **Weil sie eine andere Frage beantwortet.** Nebenan steht, was aus einem
// Bild wird — und das gilt an jeder Fläche. Hier steht, was **nur** das
// Fenster tut: nach der Datei fragen. §16 nennt den Dateiwahldialog als die
// benannte Ausnahme von „dieselbe Fähigkeit überall", und eine Ausnahme
// gehört sichtbar an eine Stelle.
//
// Der Schnitt folgt der Frage und nicht der Zeilenzahl (§2) — die 400 waren
// der Anlass.

import (
	"context"
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Am Fenster fragt ein leerer Pfad nach der Datei.**
//
// # Warum der Dialog an der Bindung hängt und nicht in `App`
//
// **Weil es ihn ohne Fenster nicht gibt.** §16 verlangt dieselbe Fähigkeit
// an jeder Fläche und nennt den Dateiwahldialog als benannte Ausnahme; stünde
// er in `App`, öffnete ein MCP-Werkzeug mit leerem Pfad ein Fenster auf einem
// Rechner, an dem niemand sitzt.
//
// **Die Gegenprobe steht daneben**: [TestOhnePfadWirdNichtsGelesen] hält,
// dass derselbe leere Pfad **ohne** Bindung ein Fehler bleibt.
func TestDieFensterbindungFragtNachDerDatei(t *testing.T) {
	t.Parallel()

	welt := einesichtwelt(t)
	deuter := &bilddeuterprobe{deutung: service.Bilddeutung{
		Kandidaten: []service.Bildkandidat{},
	}}

	welt.anwendung.MitBilderkenner(deuter)
	welt.laufzeit.antwort = app.KIRohantwort{
		Inhalt: `{"beobachtungen":[{"art":"name","wert":"Bolt","lesbar":true}]}`,
	}

	pfad := einProbebildAufPlatte(t)
	gefragt := 0

	welt.anwendung.MitDateiwahl(func(_ context.Context, auswahl app.Dateiauswahl) (
		string, error,
	) {
		gefragt++

		// **Und er geht dort auf, wo Bilder liegen** — der Filter ist der
		// des Bildfelds, nicht „alle Dateien".
		if auswahl.Muster == "" {
			t.Errorf("der Dialog hat kein Muster: %+v", auswahl)
		}

		return pfad, nil
	})

	welt.anwendung.Bereit(probekontext(t))

	antwort := app.NeueBindungen(welt.anwendung).KIKarteErkennen(t.Context(),
		app.KIErkennungsbezug{Anbieter: "ollama", Alias: "sieht:8b"})
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if gefragt != 1 {
		t.Errorf("der Dialog ging %dmal auf", gefragt)
	}

	if len(deuter.gesehen.Beobachtungen) != 1 {
		t.Errorf("gelesen wurde %+v", deuter.gesehen.Beobachtungen)
	}
}

// **Ein geschlossener Dialog ist kein Fehler — und keine Auskunft über das
// Bild.**
//
// **Zwei Zustände und nicht einer**: „ich habe nichts gelesen" und „es wurde
// gar nicht gefragt" sehen ohne [KIErkennungAntwort.Abgebrochen] gleich aus,
// und das Fenster zeigte dann „kein Druck passt" für ein Bild, das niemand
// gewählt hat.
func TestEinAbgebrochenerDialogIstKeinFehler(t *testing.T) {
	t.Parallel()

	welt := einesichtwelt(t)
	welt.anwendung.MitBilderkenner(&bilddeuterprobe{})
	welt.anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (
		string, error,
	) {
		return "", nil
	})

	welt.anwendung.Bereit(probekontext(t))

	antwort := app.NeueBindungen(welt.anwendung).KIKarteErkennen(t.Context(),
		app.KIErkennungsbezug{Anbieter: "ollama", Alias: "sieht:8b"})

	if antwort.Fehler != nil {
		t.Fatalf("ein Abbruch ist kein Fehler: %+v", antwort.Fehler)
	}

	if !antwort.Abgebrochen {
		t.Error("der Abbruch steht nicht in der Antwort")
	}

	if len(welt.laufzeit.gesehene) != 0 {
		t.Errorf("es ging trotzdem etwas hinaus: %+v", welt.laufzeit.gesehene)
	}
}

// **Ein Dialog, der stolpert, meldet sich.**
func TestEinStolperndersDialogMeldetSich(t *testing.T) {
	t.Parallel()

	welt := einesichtwelt(t)
	welt.anwendung.MitBilderkenner(&bilddeuterprobe{})
	welt.anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (
		string, error,
	) {
		return "", fmt.Errorf("%w: der Dialog ging nicht auf", domain.ErrValidation)
	})

	welt.anwendung.Bereit(probekontext(t))

	antwort := app.NeueBindungen(welt.anwendung).KIKarteErkennen(t.Context(),
		app.KIErkennungsbezug{Anbieter: "ollama", Alias: "sieht:8b"})

	if antwort.Fehler == nil {
		t.Fatal("der Fehler des Dialogs kam nicht durch")
	}

	if antwort.Abgebrochen {
		t.Error("ein Fehler ist kein Abbruch — das sind zwei Auskünfte")
	}
}

// **Und ein Pfad im Aufruf lässt den Dialog zu.**
//
// **Das ist die Gegenprobe zum Dialog**: ohne sie sagte der Lauf darüber nur,
// dass er aufgeht — nicht, dass er es **nur dann** tut. Am Fenster zieht
// jemand eine Datei ins Fenster, und eine Rückfrage darauf wäre eine Frage
// auf eine Antwort (vgl. [App.BildAusDatei]).
func TestEinGenannterPfadLaesstDenDialogZu(t *testing.T) {
	t.Parallel()

	welt := einesichtwelt(t)
	welt.anwendung.MitBilderkenner(&bilddeuterprobe{deutung: service.Bilddeutung{
		Kandidaten: []service.Bildkandidat{},
	}})
	welt.laufzeit.antwort = app.KIRohantwort{
		Inhalt: `{"beobachtungen":[{"art":"name","wert":"Bolt","lesbar":true}]}`,
	}

	welt.anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (
		string, error,
	) {
		t.Error("der Dialog ging auf, obwohl der Pfad dastand")

		return "", nil
	})

	welt.anwendung.Bereit(probekontext(t))

	antwort := app.NeueBindungen(welt.anwendung).KIKarteErkennen(t.Context(),
		app.KIErkennungsbezug{
			Pfad: einProbebildAufPlatte(t), Anbieter: "ollama",
			Alias: "sieht:8b",
		})
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}
}

// **Eine volle Deutung kommt vollständig an der Fläche an.**
//
// # Was dieser Lauf hält
//
// **Die Übersetzung selbst.** Die Läufe nebenan geben dem Deuter eine leere
// Deutung zurück und prüfen den **Weg**; hier steht, dass aus jedem Teil auch
// ein Feld wird — Beobachtungen, nicht gelesene Stellen, Befunde,
// **Kandidaten samt ihren Stützen**, die Zahl der Drucke und das
// „mindestens" daneben.
//
// **Ohne ihn wäre die Zusage der Antwort unbelegt**: eine Übersetzung, die
// die Kandidatenschleife ausließe, käme an keinem anderen Lauf vorbei.
func TestEineVolleDeutungKommtAnDerFlaecheAn(t *testing.T) {
	t.Parallel()

	welt := einesichtwelt(t)
	welt.anwendung.MitBilderkenner(&bilddeuterprobe{deutung: service.Bilddeutung{
		Erkennung: domain.Bilderkennung{
			Beobachtungen: []domain.Beobachtung{
				{Art: domain.EvidenzName, Wert: "Shivan Dragon", Lesbar: true},
			},
			NichtGelesen: []domain.Evidenzart{domain.EvidenzNummer},
		},
		Befunde: []domain.Belegbefund{{
			Art: domain.BefundWertOhneLesbarkeit, Stelle: "nummer",
			Meldung: "geraten",
		}},
		Kandidaten: []service.Bildkandidat{{
			Treffer: service.Suchtreffer{
				Kennung: "lea-174", Name: "Shivan Dragon", Setcode: "lea",
				Setname: "Limited Edition Alpha", Nummer: "174",
				Seltenheit: "rare", Bild: "https://example.invalid/174.png",
			},
			Gestuetzt: []domain.Evidenzart{domain.EvidenzName},
		}},
		Drucke: 72, Weitere: true, Alter: "1h0m0s",
	}})

	welt.laufzeit.antwort = app.KIRohantwort{
		Inhalt: `{"beobachtungen":[{"art":"name","wert":"x","lesbar":true}]}`,
	}

	antwort := welt.anwendung.KIKarteErkennen(probekontext(t),
		app.KIErkennungsbezug{
			Pfad: einProbebildAufPlatte(t), Anbieter: "ollama",
			Alias: "sieht:8b",
		})
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Kandidaten) != 1 ||
		antwort.Kandidaten[0].Kennung != "lea-174" ||
		len(antwort.Kandidaten[0].Gestuetzt) != 1 ||
		antwort.Kandidaten[0].Gestuetzt[0] != string(domain.EvidenzName) {
		t.Errorf("die Kandidaten sind %+v", antwort.Kandidaten)
	}

	if len(antwort.NichtGelesen) != 1 || len(antwort.Befunde) != 1 {
		t.Errorf("nicht gelesen %v, Befunde %+v",
			antwort.NichtGelesen, antwort.Befunde)
	}

	// **72 passende Drucke sind nicht eindeutig** — und `weitere` sagt, dass
	// es dahinter weitergeht.
	if antwort.Drucke != 72 || !antwort.Weitere || antwort.Eindeutig {
		t.Errorf("drucke %d, weitere %v, eindeutig %v",
			antwort.Drucke, antwort.Weitere, antwort.Eindeutig)
	}

	if antwort.Alter != "1h0m0s" {
		t.Errorf("das Alter ist %q", antwort.Alter)
	}
}

// **Ein Ruf ohne Profil wird abgewiesen, bevor etwas hinausgeht.**
func TestEineErkennungOhneProfilGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	welt := einesichtwelt(t)
	welt.anwendung.MitBilderkenner(&bilddeuterprobe{})

	antwort := welt.anwendung.KIKarteErkennen(t.Context(),
		app.KIErkennungsbezug{
			Pfad: einProbebildAufPlatte(t), Anbieter: "ollama",
			Alias: "sieht:8b",
		})

	if antwort.Fehler == nil {
		t.Fatal("der Ruf ohne Profil kam durch")
	}

	if len(welt.laufzeit.gesehene) != 0 {
		t.Errorf("es ging trotzdem etwas hinaus: %+v", welt.laufzeit.gesehene)
	}
}

// **Und ein Anbieter, den es nicht gibt, ebenso.**
func TestEinUnbekannterAnbieterErkenntNichts(t *testing.T) {
	t.Parallel()

	welt := einesichtwelt(t)
	welt.anwendung.MitBilderkenner(&bilddeuterprobe{})

	antwort := welt.anwendung.KIKarteErkennen(probekontext(t),
		app.KIErkennungsbezug{
			Pfad: einProbebildAufPlatte(t), Anbieter: "gibtesnicht",
			Alias: "sieht:8b",
		})

	if antwort.Fehler == nil {
		t.Fatal("der unbekannte Anbieter kam durch")
	}
}

// **Eine abgeschaltete Laufzeit erkennt nichts — und der Posten wird
// freigegeben.**
//
// # Was dieser Lauf hält
//
// **Den Weg zwischen Buchen und Werkzeugsatz.** Gebucht wird **vor** dem
// Werk (dieselbe Reihenfolge wie beim Batch); steht das Werk dann nicht, ist
// Geld reserviert, das niemand ausgegeben hat. Es wird **freigegeben** und
// nicht als unklar gemerkt: hinausgegangen ist nichts
// (#1542).
//
// **Ein abgeschalteter Anbieter ist dafür der ehrliche Fall** — ein
// unbekannter gilt als Cloud und scheitert schon beim Buchen.
func TestEineAbgeschalteteLaufzeitErkenntNichts(t *testing.T) {
	t.Parallel()

	welt := einesichtwelt(t)
	welt.anwendung.MitBilderkenner(&bilddeuterprobe{})

	if fehler := storage.NeueEinstellungen(welt.db).Setzen(probekontext(t),
		domain.KILaufzeit, domain.KILaufzeitAus); fehler != nil {
		t.Fatalf("abschalten: %v", fehler)
	}

	antwort := welt.anwendung.KIKarteErkennen(probekontext(t),
		app.KIErkennungsbezug{
			Pfad: einProbebildAufPlatte(t), Anbieter: "ollama",
			Alias: "sieht:8b",
		})

	if antwort.Fehler == nil {
		t.Fatal("die abgeschaltete Laufzeit kam durch")
	}

	if len(welt.laufzeit.gesehene) != 0 {
		t.Errorf("es ging trotzdem etwas hinaus: %+v", welt.laufzeit.gesehene)
	}

	// **Und im Journal steht nichts** — ein Modell auf diesem Rechner bucht
	// nicht ([kibudget.Reservieren] gibt für `lokal` den Posten ohne Kosten).
	//
	// **Das ist die Zusage, die hier zu prüfen ist**, und nicht „der Posten
	// wurde freigegeben": ein Journaleintrag für einen Lauf, der nichts
	// kostet, wäre eine Zeile, die eine Rechnung vortäuscht.
	stand := welt.anwendung.Kostenposten(probekontext(t), "ollama",
		string(domain.ZeitraumTag))
	if stand.Fehler != nil {
		t.Fatalf("%+v", stand.Fehler)
	}

	if len(stand.Posten) != 0 {
		t.Errorf("ein lokaler Lauf hat gebucht: %+v", stand.Posten)
	}
}
