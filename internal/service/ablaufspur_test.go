package service_test

// **Die Ablaufspur** — OPS-05.AK-02/03
// (#1722):
// je Ruf eine Zeile mit Vorgangskennung, Fläche, Weg, Dauer und Ausgang —
// **aus, bis jemand sie einschaltet**, abgetastet, mit einer Höchstzahl von
// Einträgen und einer Aufbewahrung. Kein Rumpf, keine Parameter.

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func ruf(weg string) service.Laufzeitpunkt {
	return service.Laufzeitpunkt{Flaeche: domain.ModusDienst, Weg: weg, Dauer: 2 * time.Millisecond}
}

func TestDieSpurIstAusBisJemandSieEinschaltet(t *testing.T) {
	service.LaufzeitwerkZuruecksetzen()
	t.Cleanup(service.LaufzeitwerkZuruecksetzen)

	ctx := service.MitVorgang(t.Context(), "0a1b2c3d")
	service.ErfasseRuf(ctx, ruf("GET /sets"))

	stand := service.Spurstand()
	if stand.An || len(stand.Punkte) != 0 || stand.Gesehen != 0 {
		t.Errorf("ohne Einschalten: %+v", stand)
	}

	// **Die Verteilung zählt trotzdem** — sie ist begrenzt und kostet nichts.
	if len(service.Laufzeitstand()) != 1 {
		t.Errorf("Laufzeiten: %+v", service.Laufzeitstand())
	}
}

func TestDieSpurNenntVorgangUndWeg(t *testing.T) {
	service.LaufzeitwerkZuruecksetzen()
	t.Cleanup(service.LaufzeitwerkZuruecksetzen)

	service.SpurEinrichten(domain.MessregelAus(map[string]string{domain.MessungSpur: "an"}))

	ctx := service.MitVorgang(t.Context(), "0a1b2c3d")
	punkt := ruf("GET /sets/{kennung}")
	punkt.Ausgang = domain.RufAbgewiesen
	service.ErfasseRuf(ctx, punkt)

	stand := service.Spurstand()
	if !stand.An || len(stand.Punkte) != 1 {
		t.Fatalf("%+v", stand)
	}

	eintrag := stand.Punkte[0]
	if eintrag.Vorgang != "0a1b2c3d" || eintrag.Weg != "GET /sets/{kennung}" ||
		eintrag.Flaeche != domain.ModusDienst || eintrag.Ausgang != domain.RufAbgewiesen ||
		eintrag.Art != service.ArtLesend || eintrag.DauerMs != 2 || eintrag.Zeit.IsZero() {
		t.Errorf("%+v", eintrag)
	}
}

// **Abgetastet wird genau, nicht zufällig** — bei zehn Prozent jeder zehnte
// Ruf, damit ein Lauf wiederholbar dieselbe Zahl gibt.
func TestDieSpurTastetAb(t *testing.T) {
	service.LaufzeitwerkZuruecksetzen()
	t.Cleanup(service.LaufzeitwerkZuruecksetzen)

	service.SpurEinrichten(domain.MessregelAus(map[string]string{
		domain.MessungSpur: "an", domain.MessungAbtastung: "10",
	}))

	for range 100 {
		service.ErfasseRuf(t.Context(), ruf("GET /sets"))
	}

	stand := service.Spurstand()
	if stand.Gesehen != 100 || stand.Aufgenommen != 10 || len(stand.Punkte) != 10 || stand.Abtastung != 10 {
		t.Errorf("gesehen %d, aufgenommen %d, Punkte %d", stand.Gesehen, stand.Aufgenommen, len(stand.Punkte))
	}
}

// **Die Grenze ist ein Ring**: der älteste Eintrag geht, und die Zahl der
// verdrängten steht in der Auskunft — sonst sähe eine volle Spur aus wie eine
// vollständige.
func TestDieSpurHaeltIhreGrenze(t *testing.T) {
	service.LaufzeitwerkZuruecksetzen()
	t.Cleanup(service.LaufzeitwerkZuruecksetzen)

	service.SpurEinrichten(domain.MessregelAus(map[string]string{
		domain.MessungSpur: "an", domain.MessungSpurgrenze: "10",
	}))

	for stelle := range 25 {
		service.ErfasseRuf(t.Context(), ruf(string(rune('a'+stelle))))
	}

	stand := service.Spurstand()
	if len(stand.Punkte) != 10 || stand.Verdraengt != 15 || stand.Grenze != 10 {
		t.Fatalf("Punkte %d, verdrängt %d", len(stand.Punkte), stand.Verdraengt)
	}

	if stand.Punkte[0].Weg != "p" || stand.Punkte[9].Weg != "y" {
		t.Errorf("erster %q, letzter %q", stand.Punkte[0].Weg, stand.Punkte[9].Weg)
	}
}

// **Was älter ist als die Aufbewahrung, fällt heraus** — beim Lesen und beim
// nächsten Eintragen.
func TestDieSpurVergisstNachDerAufbewahrung(t *testing.T) {
	service.LaufzeitwerkZuruecksetzen()
	t.Cleanup(service.LaufzeitwerkZuruecksetzen)

	jetzt := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	service.SpurUhrSetzen(func() time.Time { return jetzt })

	service.SpurEinrichten(domain.MessregelAus(map[string]string{
		domain.MessungSpur: "an", domain.MessungAufbewahrung: "5",
	}))

	service.ErfasseRuf(t.Context(), ruf("alt"))

	jetzt = jetzt.Add(6 * time.Minute)
	service.ErfasseRuf(t.Context(), ruf("neu"))

	stand := service.Spurstand()
	if len(stand.Punkte) != 1 || stand.Punkte[0].Weg != "neu" || stand.Abgelaufen != 1 ||
		stand.AufbewahrungS != 300 {
		t.Errorf("%+v", stand)
	}
}

// **Die Art steht dabei** — ein schreibender Ruf heißt so.
func TestDieSpurNenntDieArt(t *testing.T) {
	service.LaufzeitwerkZuruecksetzen()
	t.Cleanup(service.LaufzeitwerkZuruecksetzen)

	service.SpurEinrichten(domain.MessregelAus(map[string]string{domain.MessungSpur: "an"}))

	punkt := ruf("POST /pflege")
	punkt.Schreibend = true
	service.ErfasseRuf(t.Context(), punkt)

	if stand := service.Spurstand(); len(stand.Punkte) != 1 || stand.Punkte[0].Art != service.ArtSchreibend {
		t.Errorf("%+v", stand)
	}
}
