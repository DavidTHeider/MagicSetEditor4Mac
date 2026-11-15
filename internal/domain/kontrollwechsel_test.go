package domain_test

// **Kontrollwechsel und Umwandeln** — Werkstück M12 der Regelmaschine
// (#1614).
//
// # Beides ändert ein Objekt, das es schon gibt
//
// **Und beides ist deshalb an dem zu prüfen, was sich *nicht* ändert.** Ein
// Kontrollwechsel, der ein Permanent nebenbei enttappt oder in eine andere
// Zone legt, sähe von außen aus wie einer — und wäre keiner.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Kontrolle wechselt** — CR 110.2 mit CR 613.1b (Fassung 2026-08-19).
func TestBeherrschenWechseltDieKontrolle(t *testing.T) {
	t.Parallel()

	danach, fehler := eineGestaltenpartie(t).Beherrschen(
		"sitz-1/baer/0", "sitz-2")
	if fehler != nil {
		t.Fatalf("beherrschen: %v", fehler)
	}

	baer := objektMit(t, danach, "sitz-1/baer/0")

	if baer.Beherrscher != "sitz-2" {
		t.Errorf("der Bär wird von %q beherrscht", baer.Beherrscher)
	}

	// **Der Eigner bleibt** — CR 110.2 trennt Eigner und Beherrscher; beim
	// Zerstören geht der Bär in *seinen* Friedhof.
	if baer.Eigner != "sitz-1" {
		t.Errorf("der Bär gehört %q", baer.Eigner)
	}
}

// **Drei Dinge, die dabei nicht geschehen** — CR 110.2.
func TestBeherrschenBewegtUndEnttapptNicht(t *testing.T) {
	t.Parallel()

	vorher := eineGestaltenpartie(t)

	danach, fehler := vorher.Beherrschen("sitz-1/baer/0", "sitz-2")
	if fehler != nil {
		t.Fatalf("beherrschen: %v", fehler)
	}

	baer := objektMit(t, danach, "sitz-1/baer/0")

	if baer.Zone != domain.ZoneTisch {
		t.Errorf("der Bär liegt in %q", baer.Zone)
	}

	// **Getappt bleibt getappt** — CR 110.5: das ist ein Zustand des
	// Permanents, und ein neuer Beherrscher ändert daran nichts.
	if !baer.Getappt {
		t.Error("der Bär ist enttappt worden")
	}

	// **Die Partie des Aufrufers bleibt stehen** — sie ist ein Wert, und
	// CR 733 hängt daran.
	if alt := objektMit(t, vorher,
		"sitz-1/baer/0"); alt.Beherrscher != "sitz-1" {
		t.Errorf("die alte Partie sagt %q", alt.Beherrscher)
	}
}

// **Der Zug wird neu gestempelt** — CR 302.6 verlangt durchgehende
// Kontrolle seit dem Beginn des jüngsten eigenen Zuges. Nach einem Wechsel
// ist das nicht mehr wahr, und die Kreatur greift in diesem Zug nicht an.
func TestBeherrschenStempeltDenZugNeu(t *testing.T) {
	t.Parallel()

	partie := eineGestaltenpartie(t)
	partie.Zug.Nummer = 5
	partie.Objekte[1].Seit = 1

	danach, fehler := partie.Beherrschen("sitz-1/baer/0", "sitz-2")
	if fehler != nil {
		t.Fatalf("beherrschen: %v", fehler)
	}

	if baer := objektMit(t, danach, "sitz-1/baer/0"); baer.Seit != 5 {
		t.Errorf("der Bär ist seit Zug %d dabei", baer.Seit)
	}
}

// **Aber es verlässt den Kampf** — CR 506.4 nennt den Kontrollwechsel
// ausdrücklich.
func TestBeherrschenNimmtAusDemKampf(t *testing.T) {
	t.Parallel()

	partie := eineGestaltenpartie(t)
	partie.Zug.Kampf = domain.Kampfstand{
		Erklaert: true,
		Angriffe: []domain.Angriff{
			{Kreatur: "sitz-1/baer/0", Ziel: "sitz-2"},
			{Kreatur: "sitz-1/wandler/0", Ziel: "sitz-2"},
		},
		Bloecke: []domain.Block{
			{Kreatur: "sitz-2/block/0", Angreifer: "sitz-1/baer/0"},
			{Kreatur: "sitz-2/block/1", Angreifer: "sitz-1/wandler/0"},
		},
	}

	danach, fehler := partie.Beherrschen("sitz-1/baer/0", "sitz-2")
	if fehler != nil {
		t.Fatalf("beherrschen: %v", fehler)
	}

	if anzahl := len(danach.Zug.Kampf.Angriffe); anzahl != 1 {
		t.Errorf("es greifen noch %d an", anzahl)
	}

	// **Der Block auf ihn geht mit** — ein Blocker ohne Angreifer wäre ein
	// halber Stand, und genau danach fragt §10.
	if anzahl := len(danach.Zug.Kampf.Bloecke); anzahl != 1 {
		t.Errorf("es blocken noch %d", anzahl)
	}

	if danach.Zug.Kampf.Bloecke[0].Angreifer != "sitz-1/wandler/0" {
		t.Errorf("geblieben ist der Block auf %q",
			danach.Zug.Kampf.Bloecke[0].Angreifer)
	}
}

// **Und ein blockender Verteidiger geht ebenso** — CR 506.4 unterscheidet
// nicht zwischen den beiden Rollen.
func TestBeherrschenNimmtAuchDenBlockerAusDemKampf(t *testing.T) {
	t.Parallel()

	partie := eineGestaltenpartie(t)
	partie.Zug.Kampf = domain.Kampfstand{
		Geblockt: true,
		Bloecke: []domain.Block{
			{Kreatur: "sitz-1/baer/0", Angreifer: "sitz-2/greifer/0"},
		},
	}

	danach, fehler := partie.Beherrschen("sitz-1/baer/0", "sitz-2")
	if fehler != nil {
		t.Fatalf("beherrschen: %v", fehler)
	}

	if anzahl := len(danach.Zug.Kampf.Bloecke); anzahl != 0 {
		t.Errorf("es blocken noch %d", anzahl)
	}
}

// **Was kein Kontrollwechsel ist.**
func TestWasKeinKontrollwechselIst(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		objekt, wer string
		erwartet    error
	}{
		"ein Sitz, den es nicht gibt": {
			objekt: "sitz-1/baer/0", wer: "sitz-9",
			erwartet: domain.ErrNotFound,
		},
		"ein Objekt, das es nicht gibt": {
			objekt: "sitz-1/greif/0", wer: "sitz-2",
			erwartet: domain.ErrNotFound,
		},
		"wer schon beherrscht": {
			objekt: "sitz-1/baer/0", wer: "sitz-1",
			erwartet: domain.ErrValidation,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := eineGestaltenpartie(t).Beherrschen(
				fall.objekt, fall.wer)
			if !errors.Is(fehler, fall.erwartet) {
				t.Errorf("beherrschen: %v", fehler)
			}
		})
	}
}

// **Und nur auf dem Schlachtfeld** — CR 110.2 spricht vom Permanent.
func TestBeherrschenNurAufDemSchlachtfeld(t *testing.T) {
	t.Parallel()

	partie := eineGestaltenpartie(t)
	partie.Objekte[1].Zone = domain.ZoneFriedhof

	_, fehler := partie.Beherrschen("sitz-1/baer/0", "sitz-2")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("beherrschen im Friedhof: %v", fehler)
	}
}

// **Die andere Seite kommt nach oben** — CR 701.27a (Fassung 2026-08-19).
func TestUmwandelnDrehtAufDieZweiteSeite(t *testing.T) {
	t.Parallel()

	danach, fehler := eineGestaltenpartie(t).Umwandeln("sitz-1/wandler/0")
	if fehler != nil {
		t.Fatalf("umwandeln: %v", fehler)
	}

	wandler := objektMit(t, danach, "sitz-1/wandler/0")

	if wandler.Seite != 1 {
		t.Errorf("oben liegt Seite %d", wandler.Seite)
	}

	eigen, da := danach.Eigenschaften(wandler)
	if !da {
		t.Fatal("der Wandler hat keine Eigenschaften")
	}

	if eigen.Name != "Gewandelt" || eigen.Staerke.Zahl != 3 {
		t.Errorf("oben liegt %q mit Stärke %d",
			eigen.Name, eigen.Staerke.Zahl)
	}
}

// **Und zurück** — CR 701.27a kennt beide Richtungen, und CR 712.8a sagt,
// dass außerhalb von Schlachtfeld und Stapel die Vorderseite gilt.
func TestUmwandelnDrehtAuchZurueck(t *testing.T) {
	t.Parallel()

	einmal, fehler := eineGestaltenpartie(t).Umwandeln("sitz-1/wandler/0")
	if fehler != nil {
		t.Fatalf("erstes umwandeln: %v", fehler)
	}

	zweimal, fehler := einmal.Umwandeln("sitz-1/wandler/0")
	if fehler != nil {
		t.Fatalf("zweites umwandeln: %v", fehler)
	}

	if seite := objektMit(t, zweimal,
		"sitz-1/wandler/0").Seite; seite != 0 {
		t.Errorf("oben liegt Seite %d", seite)
	}
}

// **Es ist kein Zonenwechsel** — CR 701.27a lässt das Permanent umdrehen
// und nicht die Zone wechseln: es bleibt dasselbe Objekt und behält seinen
// Zustand.
func TestUmwandelnIstKeinZonenwechsel(t *testing.T) {
	t.Parallel()

	partie := eineGestaltenpartie(t)
	partie.Objekte[0].Getappt = true
	partie.Objekte[0].Zaehler = map[string]int{domain.MarkePlusEins: 2}

	danach, fehler := partie.Umwandeln("sitz-1/wandler/0")
	if fehler != nil {
		t.Fatalf("umwandeln: %v", fehler)
	}

	wandler := objektMit(t, danach, "sitz-1/wandler/0")

	if !wandler.Getappt {
		t.Error("der Wandler ist enttappt worden")
	}

	if wandler.Zaehler[domain.MarkePlusEins] != 2 {
		t.Errorf("auf dem Wandler liegen %d Marken",
			wandler.Zaehler[domain.MarkePlusEins])
	}

	// **Und die Marken gelten weiter** — CR 613.4c legt sie über die Werte
	// der Seite, die oben liegt: 3/3 plus zwei ist 5/5.
	staerke, _ := danach.Werte(wandler)
	if staerke.Zahl != 5 {
		t.Errorf("der Wandler hat Stärke %d", staerke.Zahl)
	}
}

// **Was sich nicht umwandeln lässt.**
func TestWasSichNichtUmwandelnLaesst(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		objekt   string
		vorher   func(*domain.Partie)
		erwartet error
	}{
		"eine einseitige Karte": {
			objekt: "sitz-1/baer/0", erwartet: domain.ErrValidation,
		},
		"ein Objekt, das es nicht gibt": {
			objekt: "sitz-1/greif/0", erwartet: domain.ErrNotFound,
		},
		"eine Karte in der Hand": {
			objekt: "sitz-1/wandler/0",
			vorher: func(p *domain.Partie) {
				p.Objekte[0].Zone = domain.ZoneHand
			},
			erwartet: domain.ErrValidation,
		},
		"ein Spielstein ohne Karte": {
			objekt: "sitz-1/wandler/0",
			vorher: func(p *domain.Partie) {
				p.Objekte[0].Karte = ""
			},
			erwartet: domain.ErrValidation,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			partie := eineGestaltenpartie(t)
			if fall.vorher != nil {
				fall.vorher(&partie)
			}

			_, fehler := partie.Umwandeln(fall.objekt)
			if !errors.Is(fehler, fall.erwartet) {
				t.Errorf("umwandeln: %v", fehler)
			}
		})
	}
}
