package app

// **Den Beleg lesen, ohne einen Schritt zu gehen** — gefunden mit zwei
// Fenstern (Hüllenprobe `nat04handel`, S12.NAT-04,
// #1672).
//
// Bert bestätigte zuerst, Anna danach, und ihr Beleg kam bei ihm an. Sein
// Fenster sagte weiter „Nur du hast unterschrieben": wer unterschrieben hat,
// erfuhr jede Fläche nur als **Antwort auf den eigenen Klick**, und nach
// Annas Beleg klickte Bert nichts mehr.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// eineBelegflaeche baut eine Anwendung mit Identität und hingelegtem Handel.
func eineBelegflaeche(t *testing.T) *App {
	t.Helper()

	anwendung := PeerflaecheUeber(
		bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})

	ctx := anwendung.imProfil(riegelkontext(t))

	if antwort := anwendung.PeeridentitaetSicherstellen(
		ctx); antwort.Fehler != nil {
		t.Fatalf("Identität anlegen: %+v", antwort.Fehler)
	}

	ich, fehler := anwendung.Ausweis(ctx)
	if fehler != nil {
		t.Fatalf("Ausweis: %v", fehler)
	}

	legeHandelHin(t, anwendung.db, ich.Kennung())

	return anwendung
}

func TestEinAngekommenerGegenbelegLaesstSichLesen(t *testing.T) {
	t.Parallel()

	anwendung := eineBelegflaeche(t)
	ctx := anwendung.imProfil(riegelkontext(t))

	freigabe := anwendung.HandelFreigabeErteilen(
		ctx, derNahtvorgang, "verbuchen", "")
	if antwort := anwendung.Handelsverbuchung(
		ctx, derNahtvorgang, freigabe.Kennung); antwort.Fehler != nil {
		t.Fatalf("Handelsverbuchung: %+v", antwort.Fehler)
	}

	// **Der Gegenbeleg kommt an, wie über die Leitung** — ohne dass hier
	// jemand einen Schritt geht.
	if fehler := storage.NeueAbschlussablage(anwendung.db).Gegenzeichnen(
		ctx, domain.Fremdbeleg{
			Vorgang: derNahtvorgang, Von: "peer-bert", Abdruck: "a1",
			Unterschrift: "deadbeef",
		}); fehler != nil {
		t.Fatalf("Gegenzeichnen: %v", fehler)
	}

	antwort := anwendung.Handelsbeleg(ctx, derNahtvorgang)

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Handelsbeleg: %+v", antwort.Fehler)
	case antwort.Zustand != string(domain.HandelLokalVerbucht):
		t.Errorf("der Zustand ist %q", antwort.Zustand)
	case !antwort.HatBeleg || !antwort.Beleg.Beidseitig:
		t.Errorf("der Beleg ist %+v", antwort.Beleg)
	case antwort.Beleg.Gegenzeichner != "peer-bert":
		t.Errorf("gegengezeichnet hat %q", antwort.Beleg.Gegenzeichner)
	}
}

// **Vor dem Buchen gibt es keinen** — und das ist eine Antwort, kein Fehler.
func TestVorDemBuchenGibtEsKeinenBeleg(t *testing.T) {
	t.Parallel()

	anwendung := eineBelegflaeche(t)

	antwort := anwendung.Handelsbeleg(riegelkontext(t), derNahtvorgang)

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Handelsbeleg: %+v", antwort.Fehler)
	case antwort.HatBeleg:
		t.Errorf("vor dem Buchen steht ein Beleg: %+v", antwort.Beleg)
	case antwort.Zustand != string(domain.HandelUebergabeOffen):
		t.Errorf("der Zustand ist %q", antwort.Zustand)
	}
}

// **Ein Handel, den es nicht gibt, ist ein Fehler** — und keine leere
// Antwort, die wie „noch nicht gebucht" aussähe.
func TestEinUnbekannterHandelHatKeinenBeleg(t *testing.T) {
	t.Parallel()

	anwendung := eineBelegflaeche(t)

	antwort := anwendung.Handelsbeleg(riegelkontext(t), "v-gibt-es-nicht")

	if antwort.Fehler == nil || antwort.Fehler.Code != CodeNichtGefunden {
		t.Fatalf("es kam %+v", antwort.Fehler)
	}
}
