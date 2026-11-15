package domain_test

// **Wer zwei Fähigkeiten zugleich beherrscht, ordnet sie** — CR 603.3b,
// Werkstück M9 (#1611).
//
// # Der Fall ist echt und nicht gestellt
//
// **Ein Sitz kann Monarch **und** Träger der Initiative sein.** Fügt ihm dann
// eine Kreatur Kampfschaden zu, lösen **zwei** eingebaute Fähigkeiten aus —
// CR 725.2 und CR 726.2 —, und beide gehören **ihm**, weil er die
// Bezeichnungen zur Zeit des Auslösens hatte.
//
// **Und die Reihenfolge ist seine Wahl** (CR 603.3b). Sie ist sichtbar: was
// zuerst gelegt wird, liegt unten (CR 405.2) und wird zuletzt verrechnet.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Zwei Fähigkeiten desselben Sitzes warten auf seine Reihung** —
// CR 603.3b, und bis dahin bekommt niemand Priorität (CR 603.3).
func TestZweiFaehigkeitenDesselbenSitzesWerdenGeordnet(t *testing.T) {
	t.Parallel()

	danach := nachDemSchlagAufDenTraeger(t, eineTraegerpartie(t))

	if danach.Ausloeserreihung != "sitz-2" {
		t.Errorf("geordnet wird von %q", danach.Ausloeserreihung)
	}

	if len(danach.Ausgeloest) != 2 {
		t.Fatalf("es warten %d Fähigkeiten", len(danach.Ausgeloest))
	}

	// **Noch liegt nichts** — CR 603.3: gelegt wird, wenn geordnet ist.
	if danach.StapelOben() != nil {
		t.Error("auf dem Stapel liegt schon etwas")
	}

	// **Und niemand handelt** — CR 603.3 stellt die Fähigkeiten **vor** die
	// Priorität.
	if danach.Zug.Prioritaet != "" {
		t.Errorf("Priorität hat %q", danach.Zug.Prioritaet)
	}

	// **Jede Wartende ist ansprechbar** — ohne eigene Kennung ließe sich
	// die eine nicht von der anderen unterscheiden.
	if danach.Ausgeloest[0].Kennung == danach.Ausgeloest[1].Kennung {
		t.Errorf("beide heißen %q", danach.Ausgeloest[0].Kennung)
	}
}

// **Die gewählte kommt zuerst auf den Stapel** — CR 603.3b mit CR 405.2:
// unten liegt, wer zuerst gelegt wird, und wird zuletzt verrechnet.
func TestDieGewaehlteFaehigkeitKommtZuerstAufDenStapel(t *testing.T) {
	t.Parallel()

	danach := nachDemSchlagAufDenTraeger(t, eineTraegerpartie(t))

	zuerst := wartendeArt(t, danach, domain.AusloeserInitiative)

	geordnet, fehler := danach.Reihung("sitz-2", zuerst)
	if fehler != nil {
		t.Fatalf("reihung: %v", fehler)
	}

	if geordnet.Ausloeserreihung != "" {
		t.Errorf("es wird noch von %q geordnet", geordnet.Ausloeserreihung)
	}

	if len(geordnet.Ausgeloest) != 0 {
		t.Errorf("%d warten noch", len(geordnet.Ausgeloest))
	}

	// **Oben liegt die zweite** — die gewählte ging zuerst und liegt unten.
	oben := geordnet.StapelOben()
	if oben == nil {
		t.Fatal("auf dem Stapel liegt nichts")
	}

	if oben.Name != string(domain.AusloeserMonarchwechsel) {
		t.Errorf("oben liegt %q", oben.Name)
	}

	// **Und danach handelt wieder jemand** — CR 117.3a.
	if geordnet.Zug.Prioritaet != "sitz-1" {
		t.Errorf("Priorität hat %q", geordnet.Zug.Prioritaet)
	}

	// **Beide wirken** — der Angreifer nimmt beide Bezeichnungen.
	fertig := geordnet.Verrechnen().Verrechnen()

	if fertig.Monarch != "sitz-1" || fertig.Initiative != "sitz-1" {
		t.Errorf("Monarch ist %q, die Initiative hat %q",
			fertig.Monarch, fertig.Initiative)
	}
}

// **Mehrere Kreaturen geben eine Initiative und zwei Monarchwechsel** —
// CR 726.2 zählt „eine oder mehr Kreaturen" als **einen** Auslöser, CR 725.2
// jede Kreatur einzeln.
func TestMehrereKreaturenGebenNurEineInitiative(t *testing.T) {
	t.Parallel()

	partie := eineTraegerpartie(t)
	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-1/baer/1", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneTisch, Karte: "baer",
	})

	danach := nachDemSchlagAufDenTraeger(t, partie, "sitz-1/baer/1")

	var fehler error

	gezaehlt := map[domain.Ausloeserart]int{}
	for _, einer := range danach.Ausgeloest {
		gezaehlt[einer.Art]++
	}

	if gezaehlt[domain.AusloeserInitiative] != 1 {
		t.Errorf("%d Initiativauslöser", gezaehlt[domain.AusloeserInitiative])
	}

	if gezaehlt[domain.AusloeserMonarchwechsel] != 2 {
		t.Errorf("%d Monarchwechsel", gezaehlt[domain.AusloeserMonarchwechsel])
	}

	// **Und auf dem Stapel bleiben es drei verschiedene Objekte** — zwei
	// Monarchwechsel desselben Sitzes im selben Zug tragen dieselbe
	// Vorsilbe, und zwei Objekte mit derselben Kennung wären eines
	// ([Spielobjekt.Kennung]).
	gelegt := danach

	for range 3 {
		if gelegt.Ausloeserreihung == "" {
			break
		}

		gelegt, fehler = gelegt.Reihung(
			gelegt.Ausloeserreihung, gelegt.Ausgeloest[0].Kennung)
		if fehler != nil {
			t.Fatalf("reihung: %v", fehler)
		}
	}

	gesehen := map[string]bool{}

	for _, objekt := range gelegt.Objekte {
		if objekt.Zone != domain.ZoneStapel {
			continue
		}

		if gesehen[objekt.Kennung] {
			t.Errorf("die Kennung %q liegt zweimal da", objekt.Kennung)
		}

		gesehen[objekt.Kennung] = true
	}

	if len(gesehen) != 3 {
		t.Errorf("auf dem Stapel liegen %d Fähigkeiten", len(gesehen))
	}
}

// **Wer den Träger nicht trifft, löst nichts aus** — CR 726.2 nennt
// Kampfschaden **am Spieler mit der Initiative**.
//
// **Die Gegenprobe zur Regel darüber** (§10): ohne sie sagte der Lauf nur,
// dass ein Treffer auslöst — nicht, dass ein anderer Schaden es nicht tut.
func TestSchadenAnEinerKreaturNimmtKeineInitiative(t *testing.T) {
	t.Parallel()

	partie := eineTraegerpartie(t)
	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-2/wand/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-2", Eigner: "sitz-2",
		Zone: domain.ZoneTisch, Karte: "baer",
	})

	partie.Zug = domain.Zugstand{
		Nummer: 3, Aktiv: "sitz-1", Prioritaet: "sitz-1",
		Abschnitt: domain.Abschnitt{
			Phase: domain.PhaseKampf, Schritt: domain.SchrittBlocker,
		},
		Kampf: domain.Kampfstand{
			Erklaert: true, Geblockt: true,
			Angriffe: []domain.Angriff{
				{Kreatur: "sitz-1/baer/0", Ziel: "sitz-2"},
			},
			Bloecke: []domain.Block{
				{Kreatur: "sitz-2/wand/0", Angreifer: "sitz-1/baer/0"},
			},
		},
	}

	danach, fehler := partie.Passen("sitz-1")
	if fehler != nil {
		t.Fatalf("passen: %v", fehler)
	}

	danach, fehler = danach.Passen(danach.Zug.Prioritaet)
	if fehler != nil {
		t.Fatalf("zweites passen: %v", fehler)
	}

	for _, einer := range danach.Ausgeloest {
		t.Errorf("es hat %q ausgelöst", einer.Art)
	}

	if danach.Initiative != "sitz-2" {
		t.Errorf("die Initiative hat %q", danach.Initiative)
	}
}

// **Wer nicht gefragt ist, ordnet nicht** — CR 603.3b nennt den
// Beherrscher.
func TestNurDerBeherrscherOrdnetSeineFaehigkeiten(t *testing.T) {
	t.Parallel()

	danach := nachDemSchlagAufDenTraeger(t, eineTraegerpartie(t))

	_, fehler := danach.Reihung("sitz-1", danach.Ausgeloest[0].Kennung)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("sitz-1 durfte ordnen: %v", fehler)
	}

	_, fehler = danach.Reihung("sitz-2", "gibtesnicht")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("eine unbekannte Kennung ging durch: %v", fehler)
	}

	// **Und ohne offene Frage gibt es nichts zu ordnen.**
	_, fehler = eineTraegerpartie(t).Reihung("sitz-2", "egal")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("ohne Frage ging es durch: %v", fehler)
	}
}

// eineTraegerpartie gibt eine Partie, in der sitz-2 Monarch ist **und** die
// Initiative hat.
func eineTraegerpartie(t *testing.T) domain.Partie {
	t.Helper()

	partie, fehler := eineAusloeserpartie(t).Monarchwerden("sitz-2")
	if fehler != nil {
		t.Fatalf("monarch: %v", fehler)
	}

	partie, fehler = partie.Initiativenehmen("sitz-2")
	if fehler != nil {
		t.Fatalf("initiative: %v", fehler)
	}

	return partie
}

// nachDemSchlagAufDenTraeger lässt die genannten Kreaturen sitz-2 treffen.
func nachDemSchlagAufDenTraeger(
	t *testing.T, partie domain.Partie, weitere ...string,
) domain.Partie {
	t.Helper()

	angriffe := make([]domain.Angriff, 0, 1+len(weitere))
	angriffe = append(angriffe, domain.Angriff{
		Kreatur: "sitz-1/baer/0", Ziel: "sitz-2",
	})

	for _, kreatur := range weitere {
		angriffe = append(angriffe, domain.Angriff{
			Kreatur: kreatur, Ziel: "sitz-2",
		})
	}

	partie.Zug = domain.Zugstand{
		Nummer: 3, Aktiv: "sitz-1", Prioritaet: "sitz-1",
		Abschnitt: domain.Abschnitt{
			Phase: domain.PhaseKampf, Schritt: domain.SchrittBlocker,
		},
		Kampf: domain.Kampfstand{
			Erklaert: true, Geblockt: true, Angriffe: angriffe,
		},
	}

	danach, fehler := partie.Passen("sitz-1")
	if fehler != nil {
		t.Fatalf("passen: %v", fehler)
	}

	danach, fehler = danach.Passen(danach.Zug.Prioritaet)
	if fehler != nil {
		t.Fatalf("zweites passen: %v", fehler)
	}

	return danach
}

// wartendeArt gibt die Kennung der wartenden Fähigkeit dieser Art.
func wartendeArt(
	t *testing.T, partie domain.Partie, art domain.Ausloeserart,
) string {
	t.Helper()

	for _, einer := range partie.Ausgeloest {
		if einer.Art == art {
			return einer.Kennung
		}
	}

	t.Fatalf("keine wartende Fähigkeit der Art %q", art)

	return ""
}
