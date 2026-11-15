package storage_test

// **Anbieterkonten im Bestand** — Werkstück 4 von Phase 4.12
// (#1351).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein Konto lässt sich verbinden und wiederfinden.**
func TestEinAnbieterkontoLaesstSichVerbindenUndFinden(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	verbunden, fehler := ablage.KontoVerbinden(t.Context(), domain.Anbieterkonto{
		NutzerID: 1, Anbieter: "github", Subjekt: "4711", Anzeigename: "Wer",
	})
	if fehler != nil {
		t.Fatalf("verbinden: %v", fehler)
	}

	if verbunden.ID == 0 {
		t.Error("das Konto hat keine Kennung")
	}

	gefunden, lesefehler := ablage.NutzerZuKonto(t.Context(), "github", "4711")
	if lesefehler != nil {
		t.Fatalf("suchen: %v", lesefehler)
	}

	if gefunden != 1 {
		t.Errorf("Nutzer %d, erwartet war 1", gefunden)
	}
}

// **Ein Subjekt, das niemandem gehört, ist „nicht gefunden".**
//
// **Und das ist keine Panne, sondern der Normalfall**: wer sich zum ersten Mal
// über einen Anbieter anmeldet, hat noch kein Konto. Der Aufrufer legt dann
// eines an — deshalb muss er die beiden Fälle unterscheiden können (§8).
func TestEinUnbekanntesSubjektIstNichtGefunden(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	_, fehler := ablage.NutzerZuKonto(t.Context(), "github", "gibtsnicht")
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet war ErrNotFound", fehler)
	}
}

// **Dieselbe fremde Kennung an zwei Nutzern ist ein Konflikt.**
//
// **Ohne diese Zusage landete eine Anmeldung bei irgendeinem von beiden.**
func TestDieselbeFremdeKennungAnZweiNutzernIstEinKonflikt(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueNutzerablage(db)

	zweiter, fehler := ablage.NutzerAnlegen(t.Context(), domain.Nutzer{Name: "Zweite"})
	if fehler != nil {
		t.Fatalf("den zweiten Nutzer anlegen: %v", fehler)
	}

	if _, verbindfehler := ablage.KontoVerbinden(t.Context(), domain.Anbieterkonto{
		NutzerID: 1, Anbieter: "github", Subjekt: "4711",
	}); verbindfehler != nil {
		t.Fatalf("das erste Konto verbinden: %v", verbindfehler)
	}

	_, zweitfehler := ablage.KontoVerbinden(t.Context(), domain.Anbieterkonto{
		NutzerID: zweiter.ID, Anbieter: "github", Subjekt: "4711",
	})

	if !errors.Is(zweitfehler, domain.ErrConflict) {
		t.Errorf("Fehler %v, erwartet war ErrConflict", zweitfehler)
	}
}

// **Ein Nutzer hat je Anbieter höchstens ein Konto.**
func TestEinNutzerHatJeAnbieterHoechstensEinKonto(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	if _, fehler := ablage.KontoVerbinden(t.Context(), domain.Anbieterkonto{
		NutzerID: 1, Anbieter: "github", Subjekt: "4711",
	}); fehler != nil {
		t.Fatalf("verbinden: %v", fehler)
	}

	_, zweitfehler := ablage.KontoVerbinden(t.Context(), domain.Anbieterkonto{
		NutzerID: 1, Anbieter: "github", Subjekt: "0815",
	})

	if !errors.Is(zweitfehler, domain.ErrConflict) {
		t.Errorf("Fehler %v, erwartet war ErrConflict", zweitfehler)
	}
}

// **Ein Konto zu einem Nutzer, den es nicht gibt, wird abgewiesen.**
func TestEinKontoZuEinemFehlendenNutzerWirdAbgewiesen(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	_, fehler := ablage.KontoVerbinden(t.Context(), domain.Anbieterkonto{
		NutzerID: 4711, Anbieter: "github", Subjekt: "1",
	})

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet war ErrNotFound", fehler)
	}
}

// **Ein unbrauchbares Konto erreicht den Bestand gar nicht.**
func TestEinUnbrauchbaresKontoErreichtDenBestandNicht(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	_, fehler := ablage.KontoVerbinden(t.Context(), domain.Anbieterkonto{
		NutzerID: 1, Anbieter: "apple", Subjekt: "1",
	})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war ErrValidation", fehler)
	}
}

// **Die Konten eines Nutzers lassen sich lesen und lösen.**
//
// # Der Rückweg gehört dazu
//
// Das Issue sagt es ausdrücklich: „ein Nutzer, dessen Anbieterkonto wegfällt,
// darf nicht ausgesperrt sein." Ein Konto zu lösen ist deshalb keine
// Nebensache, sondern die Bedingung dafür, dass die Anmeldung überhaupt
// vertretbar ist.
func TestDieKontenEinesNutzersLassenSichLesenUndLoesen(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	if _, fehler := ablage.KontoVerbinden(t.Context(), domain.Anbieterkonto{
		NutzerID: 1, Anbieter: "github", Subjekt: "4711", Anzeigename: "Wer",
	}); fehler != nil {
		t.Fatalf("verbinden: %v", fehler)
	}

	konten, lesefehler := ablage.Konten(t.Context(), 1)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(konten) != 1 || konten[0].Anbieter != "github" {
		t.Fatalf("Konten %+v", konten)
	}

	if fehler := ablage.KontoLoesen(t.Context(), 1, "github"); fehler != nil {
		t.Fatalf("lösen: %v", fehler)
	}

	nachher, nachfehler := ablage.Konten(t.Context(), 1)
	if nachfehler != nil {
		t.Fatalf("erneut lesen: %v", nachfehler)
	}

	if len(nachher) != 0 {
		t.Errorf("nach dem Lösen: %+v", nachher)
	}
}

// **Ein Konto zu lösen, das es nicht gibt, ist „nicht gefunden".**
func TestEinFehlendesKontoZuLoesenIstNichtGefunden(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	if fehler := ablage.KontoLoesen(t.Context(), 1, "github"); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet war ErrNotFound", fehler)
	}
}
