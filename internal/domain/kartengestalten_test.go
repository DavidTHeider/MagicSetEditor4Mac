package domain_test

// **Spielsteine, Embleme und die verdeckte Seite** — Werkstück M12 der
// Regelmaschine
// (#1614).
//
// # Je Gestalt ihre Regelnummer
//
// **Der Kontrollwechsel und das Umwandeln stehen nebenan**
// (`kontrollwechsel_test.go`) — sie ändern etwas an einem Objekt, das es
// schon gibt, und diese Datei stand mit ihnen über der Grenze aus §2.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineGestaltenpartie gibt eine Partie mit einer zweiseitigen Karte im Deck
// des ersten Sitzes und einem Permanent daraus auf dem Tisch.
func eineGestaltenpartie(t *testing.T) domain.Partie {
	t.Helper()

	partie := einePartie()

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)
	sitze[0].Deck = domain.Partiedeck{
		Abdruck: "abcd",
		Karten: []domain.Abzugskarte{
			{
				Kennung: "wandler", Name: "Wandler", Menge: 1,
				Seiten: []domain.Abzugsseite{
					{Felder: map[string]string{
						"name": "Wandler", "super type": "Creature",
						"power": "1", "toughness": "1",
					}},
					{Felder: map[string]string{
						"name": "Gewandelt", "super type": "Creature",
						"power": "3", "toughness": "3",
					}},
				},
			},
			{
				Kennung: "baer", Name: "Bär", Menge: 1,
				Seiten: []domain.Abzugsseite{{Felder: map[string]string{
					"name": "Bär", "super type": "Creature",
					"power": "2", "toughness": "2",
				}}},
			},
		},
	}
	partie.Sitze = sitze

	partie.Objekte = []domain.Spielobjekt{
		{
			Kennung: "sitz-1/wandler/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneTisch, Karte: "wandler",
		},
		{
			Kennung: "sitz-1/baer/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneTisch, Karte: "baer", Getappt: true,
		},
	}

	return partie
}

// einSpielsteinfeld gibt die Felder eines 1/1-Spielsteins.
func einSpielsteinfeld() map[string]string {
	return map[string]string{
		"name": "Soldat", "super type": "Creature", "sub type": "Soldier",
		"power": "1", "toughness": "1",
	}
}

// **Ein Spielstein entsteht auf dem Schlachtfeld** — CR 701.7a
// (Fassung 2026-08-19).
func TestSpielsteinEntstehtAufDemSchlachtfeld(t *testing.T) {
	t.Parallel()

	danach, fehler := eineGestaltenpartie(t).Spielstein(
		"sitz-1", "sitz-1/soldat/0", einSpielsteinfeld())
	if fehler != nil {
		t.Fatalf("spielstein: %v", fehler)
	}

	stein := objektMit(t, danach, "sitz-1/soldat/0")

	if stein.Zone != domain.ZoneTisch {
		t.Errorf("der Spielstein liegt in %q", stein.Zone)
	}

	if stein.Art != domain.ObjektSpielstein {
		t.Errorf("der Spielstein ist ein %q", stein.Art)
	}

	if stein.Karte != "" {
		t.Errorf("der Spielstein verweist auf die Karte %q", stein.Karte)
	}
}

// **Ein Spielstein bringt seine Eigenschaften mit** — CR 111.1: er vertritt
// keine Karte, und sein Erzeuger nennt sie.
func TestSpielsteinBringtSeineEigenschaftenMit(t *testing.T) {
	t.Parallel()

	danach, fehler := eineGestaltenpartie(t).Spielstein(
		"sitz-1", "sitz-1/soldat/0", einSpielsteinfeld())
	if fehler != nil {
		t.Fatalf("spielstein: %v", fehler)
	}

	eigen, da := danach.Eigenschaften(objektMit(t, danach, "sitz-1/soldat/0"))
	if !da {
		t.Fatal("der Spielstein hat keine Eigenschaften")
	}

	if eigen.Name != "Soldat" {
		t.Errorf("der Spielstein heißt %q", eigen.Name)
	}

	if eigen.Staerke.Zahl != 1 || eigen.Widerstandskraft.Zahl != 1 {
		t.Errorf("der Spielstein ist %d/%d",
			eigen.Staerke.Zahl, eigen.Widerstandskraft.Zahl)
	}
}

// **Die Felder des Aufrufers bleiben draußen.**
//
// **Die Gegenprobe zum halben Kopieren**: `Sandkasten.kopie` hat denselben
// Fehler schon einmal gehabt, und er fiel erst an einer geänderten
// Lebenszahl auf. Wer eine Karte in die Partie reicht, darf sie danach
// nicht mehr von außen ändern.
func TestSpielsteinKopiertSeineFelder(t *testing.T) {
	t.Parallel()

	felder := einSpielsteinfeld()

	danach, fehler := eineGestaltenpartie(t).Spielstein(
		"sitz-1", "sitz-1/soldat/0", felder)
	if fehler != nil {
		t.Fatalf("spielstein: %v", fehler)
	}

	felder["power"] = "9"

	eigen, _ := danach.Eigenschaften(objektMit(t, danach, "sitz-1/soldat/0"))
	if eigen.Staerke.Zahl != 1 {
		t.Errorf("die Stärke des Spielsteins ist %d", eigen.Staerke.Zahl)
	}
}

// **Ein Spielstein ohne Eigenschaften ist keiner** — CR 701.7a nennt sie.
func TestSpielsteinOhneEigenschaftenWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	_, fehler := eineGestaltenpartie(t).Spielstein(
		"sitz-1", "sitz-1/soldat/0", nil)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("ein Spielstein ohne Felder: %v", fehler)
	}
}

// **Ein Emblem liegt in der Kommandozone** — CR 114.1, und es gehört dem,
// der es bekommt (CR 114.2).
func TestEmblemLiegtInDerKommandozone(t *testing.T) {
	t.Parallel()

	danach, fehler := eineGestaltenpartie(t).Emblem(
		"sitz-2", "sitz-2/emblem/0")
	if fehler != nil {
		t.Fatalf("emblem: %v", fehler)
	}

	emblem := objektMit(t, danach, "sitz-2/emblem/0")

	if emblem.Zone != domain.ZoneKommando {
		t.Errorf("das Emblem liegt in %q", emblem.Zone)
	}

	if emblem.Art != domain.ObjektEmblem {
		t.Errorf("das Emblem ist ein %q", emblem.Art)
	}

	if emblem.Eigner != "sitz-2" || emblem.Beherrscher != "sitz-2" {
		t.Errorf("das Emblem gehört %q und wird von %q beherrscht",
			emblem.Eigner, emblem.Beherrscher)
	}
}

// **Was ein neues Objekt nicht sein darf.**
//
// **Der Erfolgsweg steht oben**, damit jede Abweisung hier eine Aussage ist
// (§10): ohne ihn sagte „wird abgewiesen" nur, dass irgendetwas fehlt.
func TestWasKeinNeuesObjektIst(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		wer, kennung string
		erwartet     error
	}{
		"ein Sitz, den es nicht gibt": {
			wer: "sitz-9", kennung: "neu", erwartet: domain.ErrNotFound,
		},
		"eine leere Kennung": {
			wer: "sitz-1", kennung: "", erwartet: domain.ErrValidation,
		},
		"eine Kennung, die es schon gibt": {
			wer: "sitz-1", kennung: "sitz-1/baer/0",
			erwartet: domain.ErrConflict,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := eineGestaltenpartie(t).Spielstein(
				fall.wer, fall.kennung, einSpielsteinfeld())
			if !errors.Is(fehler, fall.erwartet) {
				t.Errorf("als Spielstein: %v", fehler)
			}

			_, fehler = eineGestaltenpartie(t).Emblem(
				fall.wer, fall.kennung)
			if !errors.Is(fehler, fall.erwartet) {
				t.Errorf("als Emblem: %v", fehler)
			}
		})
	}
}

// **Ein verdecktes Permanent ist eine 2/2 ohne alles** — CR 708.2a.
func TestVerdecktesPermanentIstEineZweiZwei(t *testing.T) {
	t.Parallel()

	partie := eineGestaltenpartie(t)
	partie.Objekte[0].Verdeckt = true

	eigen, da := partie.Eigenschaften(partie.Objekte[0])
	if !da {
		t.Fatal("das verdeckte Permanent hat keine Eigenschaften")
	}

	if eigen.Name != "" {
		t.Errorf("das verdeckte Permanent heißt %q", eigen.Name)
	}

	if eigen.Staerke.Zahl != 2 || eigen.Widerstandskraft.Zahl != 2 {
		t.Errorf("das verdeckte Permanent ist %d/%d",
			eigen.Staerke.Zahl, eigen.Widerstandskraft.Zahl)
	}
}

// **Und nur auf dem Schlachtfeld** — CR 708.2a spricht vom Permanent.
//
// **Die Gegenprobe zur Zeile darüber**: eine verdeckte Karte in der Hand ist
// die Karte, die sie ist; wer sie zur 2/2 machte, hätte jede Handkarte
// vergessen.
func TestVerdecktInDerHandBleibtDieKarte(t *testing.T) {
	t.Parallel()

	partie := eineGestaltenpartie(t)
	partie.Objekte[0].Verdeckt = true
	partie.Objekte[0].Zone = domain.ZoneHand

	eigen, da := partie.Eigenschaften(partie.Objekte[0])
	if !da {
		t.Fatal("die verdeckte Handkarte hat keine Eigenschaften")
	}

	if eigen.Name != "Wandler" {
		t.Errorf("die verdeckte Handkarte heißt %q", eigen.Name)
	}
}

// **Eine Seite, die es nicht gibt, ist die Vorderseite** — CR 712.8a.
//
// **Der Fall kommt aus dem Bestand und nicht aus einem Befehl**: [Umwandeln]
// lässt eine einseitige Karte gar nicht erst drehen. Wer eine Partie von
// gestern öffnet, deren Karte inzwischen nur noch eine Seite hat, bekommt
// diese — und keinen Absturz (§8).
func TestEineSeiteDieEsNichtGibtIstDieVorderseite(t *testing.T) {
	t.Parallel()

	partie := eineGestaltenpartie(t)
	partie.Objekte[1].Seite = 7

	eigen, da := partie.Eigenschaften(partie.Objekte[1])
	if !da {
		t.Fatal("der Bär hat keine Eigenschaften")
	}

	if eigen.Name != "Bär" {
		t.Errorf("oben liegt %q", eigen.Name)
	}
}
