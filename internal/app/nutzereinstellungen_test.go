package app_test

// **Die Einstellungen, die dem Menschen gehören, am Fenster** — Werkstück 8
// von Phase 4.12
// (#1351),
// Punkte 9 und 11.

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Antwort ist vollständig — auch was niemand gesetzt hat.**
//
// **Sonst müsste jede Fläche die Vorgabe noch einmal kennen**, und die zweite
// Liste liefe zurück.
func TestDieNutzereinstellungenSindVollstaendig(t *testing.T) {
	antwort := angelaufeneAnwendung(t).Nutzereinstellungen(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Einstellungen) != len(domain.Nutzereinstellungsreihe()) {
		t.Fatalf("es kamen %d von %d", len(antwort.Einstellungen),
			len(domain.Nutzereinstellungsreihe()))
	}

	// **Und jede trägt ihre Werte mit** — ohne sie baute das Fenster die
	// Auswahlliste selbst, und die erlaubten Werte stünden an zwei Stellen.
	//
	// **Gegen die Domäne und nicht gegen »nicht leer«** — Werkstück O3 von
	// Phase 5.2 (#1427):
	// seit den Spaltenbreiten gibt es Einstellungen **ohne** Werteliste, und
	// „nicht leer" wäre für sie falsch. Der Vergleich mit
	// `ErlaubteNutzerwerte` ist dabei die **schärfere** Zusage: er sieht auch
	// eine Liste, die etwas anderes enthält als die Domäne erlaubt.
	for _, eine := range antwort.Einstellungen {
		erlaubt := domain.ErlaubteNutzerwerte(eine.Schluessel)
		if !slices.Equal(eine.Werte, erlaubt) {
			t.Errorf("%q nennt %v statt %v", eine.Schluessel, eine.Werte, erlaubt)
		}

		if eine.Wert != domain.NutzerVorgabe(eine.Schluessel) {
			t.Errorf("%q steht auf %q statt auf der Vorgabe",
				eine.Schluessel, eine.Wert)
		}
	}
}

// **Eine gesetzte Einstellung kommt in derselben Antwort zurück.**
//
// **In einer und nicht in zweien** (§7, Regel 2): wer nach dem Setzen noch
// einmal läse, hätte zwei Stände und im Zwischenraum einen, den niemand
// gewählt hat.
func TestEineNutzereinstellungKommtGesetztZurueck(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	antwort := anwendung.NutzereinstellungSetzen(
		t.Context(), "1", domain.NutzerKontrast, domain.KontrastHoch)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if wertVonSchluessel(antwort, domain.NutzerKontrast) != domain.KontrastHoch {
		t.Errorf("%+v", antwort.Einstellungen)
	}

	// **Und sie bleibt** — die Antwort ist keine Behauptung über den Bestand.
	gelesen := anwendung.Nutzereinstellungen(t.Context(), "1")
	if wertVonSchluessel(gelesen, domain.NutzerKontrast) != domain.KontrastHoch {
		t.Errorf("%+v", gelesen.Einstellungen)
	}
}

// **Ein unbekannter Schlüssel und ein unerlaubter Wert werden abgewiesen.**
func TestDieNutzereinstellungenWeisenUnsinnAb(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)

	faelle := map[string]struct{ nutzerID, schluessel, wert string }{
		"keine Kennung":         {"keine Zahl", domain.NutzerKontrast, domain.KontrastHoch},
		"unbekannter Schlüssel": {"1", "farbe.des.himmels", "blau"},
		"unerlaubter Wert":      {"1", domain.NutzerKontrast, "sehr hoch"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			antwort := anwendung.NutzereinstellungSetzen(
				t.Context(), fall.nutzerID, fall.schluessel, fall.wert)
			if antwort.Fehler == nil {
				t.Error("es kam durch")
			}
		})
	}

	// **Und das Lesen prüft die Kennung ebenso.**
	if antwort := anwendung.Nutzereinstellungen(
		t.Context(), "keine Zahl"); antwort.Fehler == nil {
		t.Error("die Kennung kam durch")
	}
}

// wertVonSchluessel liest einen Wert aus einer Antwort.
func wertVonSchluessel(antwort app.NutzereinstellungenAntwort, schluessel string) string {
	for _, eine := range antwort.Einstellungen {
		if eine.Schluessel == schluessel {
			return eine.Wert
		}
	}

	return ""
}

// **Eine schweigende Platte erreicht beide Wege.**
func TestDieNutzereinstellungenReichenEinenFehlerDurch(t *testing.T) {
	anwendung := angelaufeneAnwendung(t)
	anwendung.MitNutzereinstellungen(stummeEinstellungen{})

	if antwort := anwendung.Nutzereinstellungen(t.Context(), "1"); antwort.Fehler == nil {
		t.Error("das Lesen glückte trotzdem")
	}

	// **Und auch der Weg nach dem geglückten Schreiben** — das Setzen gibt
	// die ganze Liste zurück, und die wird gelesen.
	if antwort := anwendung.NutzereinstellungSetzen(
		t.Context(), "1", domain.NutzerKontrast, domain.KontrastHoch,
	); antwort.Fehler == nil {
		t.Error("das Setzen glückte trotzdem")
	}
}

// stummeEinstellungen antwortet auf jede Frage mit demselben Fehler.
//
// **Das Setzen glückt, das Lesen nicht.** So wird der Weg **nach** dem
// Schreiben betreten — ein Doppel, das schon beim Setzen abbräche, käme dort
// nie an.
type stummeEinstellungen struct{}

func (stummeEinstellungen) Alle(context.Context, int64) (map[string]string, error) {
	return nil, errors.New("die Platte schweigt")
}

func (stummeEinstellungen) Setzen(context.Context, int64, string, string) error {
	return nil
}
