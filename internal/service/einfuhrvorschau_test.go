package service_test

// **Eine Einfuhr zeigt vorher, was sie zusammenführt** — DOM-03.AK-01
// (#1698).
//
// # Der Befund
//
// Eine Zeile, die es im Bestand schon gibt, wurde beim Einlesen **still
// erhöht**; der Bericht nannte danach eine Zahl. Wer die eigene Ausfuhr ein
// zweites Mal einlas, hatte danach jede Menge doppelt — und erfuhr es erst,
// als es geschehen war.
//
// # Was jetzt gilt
//
// **Die Vorschau schreibt nichts** und sagt je Zeile: neu, trifft eine
// vorhandene Zeile (mit vorher und nachher) oder wiederholt eine Zeile
// derselben Einfuhr. **Wer einliest, wählt**: addieren — das bisherige
// Verhalten und die Vorgabe — oder Vorhandene auslassen. Eine Zeile, die
// dieselbe Einfuhr eben erst angelegt hat, ist keine vorhandene: sie wird
// auch beim Auslassen addiert.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

const zweitprinting = "a1b2c3d4-0000-4000-8000-0000000000b2"

// bestandsliste liest die Zeilen des Doppels — und bricht ab, wenn es nicht
// antwortet.
func bestandsliste(t *testing.T, ablage *sammlungsdoppel) []*domain.Exemplar {
	t.Helper()

	zeilen, fehler := ablage.Liste(t.Context())
	if fehler != nil {
		t.Fatalf("den Bestand lesen: %v", fehler)
	}

	return zeilen
}

// vorhandenerBestand legt zwei Stück des Probedrucks an.
func vorhandenerBestand(t *testing.T) *sammlungsdoppel {
	t.Helper()

	ablage := &sammlungsdoppel{}

	if _, fehler := ablage.Anlegen(t.Context(), &domain.Exemplar{
		Printing: probeprinting, Menge: 2, Zustand: domain.ZustandNearMint,
		Sprache: domain.Vorgabesprache,
	}); fehler != nil {
		t.Fatalf("der Bestand: %v", fehler)
	}

	return ablage
}

func einfuhrzeilen() []service.Listenzeile {
	return []service.Listenzeile{
		listenzeile(),
		listenzeile(func(z *service.Listenzeile) { z.Printing, z.Menge, z.Name = zweitprinting, 1, "Glut" }),
		listenzeile(func(z *service.Listenzeile) { z.Printing, z.Menge, z.Name = zweitprinting, 2, "Glut" }),
	}
}

func TestDieVorschauSagtJeZeileWasGeschieht(t *testing.T) {
	t.Parallel()

	ablage := vorhandenerBestand(t)

	vorschau, fehler := einfuhr(ablage).Vorschauen(t.Context(), einfuhrzeilen())
	if fehler != nil {
		t.Fatalf("vorschauen: %v", fehler)
	}

	if len(vorschau.Posten) != 3 {
		t.Fatalf("%d Posten, erwartet 3", len(vorschau.Posten))
	}

	for stelle, erwartet := range []struct {
		art             service.Einfuhrart
		vorher, nachher int
	}{
		{service.EinfuhrZusammen, 2, 6},
		{service.EinfuhrNeu, 0, 1},
		{service.EinfuhrWiederholt, 1, 3},
	} {
		posten := vorschau.Posten[stelle]
		if posten.Art != erwartet.art || posten.Vorher != erwartet.vorher || posten.Nachher != erwartet.nachher {
			t.Errorf("Posten %d ist %s %d→%d, erwartet %s %d→%d", stelle,
				posten.Art, posten.Vorher, posten.Nachher, erwartet.art, erwartet.vorher, erwartet.nachher)
		}
	}

	// **Geschrieben ist nichts.**
	if zeilen := bestandsliste(t, ablage); len(zeilen) != 1 || zeilen[0].Menge != 2 {
		t.Errorf("im Bestand steht %+v", zeilen)
	}
}

// **Eine unbrauchbare Zeile ist ein Befund der Vorschau** — mit ihrer Stelle,
// wie beim Einlesen, und die anderen stehen trotzdem da.
func TestEineUnbrauchbareZeileIstEinBefundDerVorschau(t *testing.T) {
	t.Parallel()

	zeilen := []service.Listenzeile{
		listenzeile(func(z *service.Listenzeile) { z.Menge = 0 }),
		listenzeile(),
	}

	vorschau, fehler := einfuhr(&sammlungsdoppel{}).Vorschauen(t.Context(), zeilen)
	if fehler != nil {
		t.Fatalf("vorschauen: %v", fehler)
	}

	if len(vorschau.Posten) != 1 || vorschau.Posten[0].Zeile != 2 ||
		len(vorschau.Befunde) != 1 || vorschau.Befunde[0].Ort != "Zeile 1" {
		t.Errorf("Vorschau %+v", vorschau)
	}
}

// **Addieren ist die Vorgabe** — das bisherige Verhalten bleibt.
func TestAddierenErhoehtDieVorhandeneZeile(t *testing.T) {
	t.Parallel()

	ablage := vorhandenerBestand(t)

	bericht, fehler := einfuhr(ablage).EinlesenMit(t.Context(), einfuhrzeilen(), service.ZusammenAddieren)
	if fehler != nil {
		t.Fatalf("einlesen: %v", fehler)
	}

	if bericht.Erhoeht != 2 || len(bericht.Angelegt) != 1 || bericht.Ausgelassen != 0 || bericht.Stueck != 7 {
		t.Errorf("Bericht %+v", bericht)
	}
}

// **Auslassen lässt die vorhandene Zeile, wie sie ist** — und addiert, was
// dieselbe Einfuhr eben erst angelegt hat.
func TestAuslassenLaesstVorhandeneStehen(t *testing.T) {
	t.Parallel()

	ablage := vorhandenerBestand(t)

	bericht, fehler := einfuhr(ablage).EinlesenMit(t.Context(), einfuhrzeilen(), service.ZusammenAuslassen)
	if fehler != nil {
		t.Fatalf("einlesen: %v", fehler)
	}

	if bericht.Ausgelassen != 1 || len(bericht.Angelegt) != 1 || bericht.Erhoeht != 1 || bericht.Stueck != 3 {
		t.Errorf("Bericht %+v", bericht)
	}

	zeilen := bestandsliste(t, ablage)
	for _, zeile := range zeilen {
		switch zeile.Printing {
		case probeprinting:
			if zeile.Menge != 2 {
				t.Errorf("die vorhandene Zeile hat %d Stück, erwartet unverändert 2", zeile.Menge)
			}
		case zweitprinting:
			if zeile.Menge != 3 {
				t.Errorf("die neue Zeile hat %d Stück, erwartet 3", zeile.Menge)
			}
		}
	}
}

// **Ein Modus, den es nicht gibt, ist ein Feldfehler** — vor dem Schreiben.
func TestEinUnbekannterModusWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	ablage := vorhandenerBestand(t)

	_, fehler := einfuhr(ablage).EinlesenMit(t.Context(), einfuhrzeilen(), "verdoppeln")

	var feld *domain.FieldError
	if !errors.As(fehler, &feld) || feld.Field != "zusammenfuehrung" {
		t.Fatalf("Fehler = %v, erwartet einen Feldfehler an der Zusammenführung", fehler)
	}

	if zeilen := bestandsliste(t, ablage); len(zeilen) != 1 || zeilen[0].Menge != 2 {
		t.Errorf("im Bestand steht %+v", zeilen)
	}
}

// **Ein Bestand, der nicht antwortet, hält die Vorschau an** — „ich konnte
// nicht nachsehen" ist nicht „alles neu".
func TestEineSchweigendeSammlungHaeltDieVorschauAn(t *testing.T) {
	t.Parallel()

	ablage := &sammlungsdoppel{fehler: errKaputt}

	if _, fehler := einfuhr(ablage).Vorschauen(t.Context(), einfuhrzeilen()); !errors.Is(fehler, errKaputt) {
		t.Errorf("Fehler = %v, erwartet errKaputt", fehler)
	}
}

// **Leer heißt addieren** — wer nichts sagt, bekommt das Verhalten von
// bisher, und das ist an jeder Fläche der Fall ohne das neue Feld.
func TestEineLeereZusammenfuehrungAddiert(t *testing.T) {
	t.Parallel()

	bericht, fehler := einfuhr(vorhandenerBestand(t)).EinlesenMit(t.Context(), einfuhrzeilen(), "")
	if fehler != nil || bericht.Erhoeht != 2 || bericht.Ausgelassen != 0 {
		t.Errorf("Bericht %+v, Fehler %v", bericht, fehler)
	}
}
