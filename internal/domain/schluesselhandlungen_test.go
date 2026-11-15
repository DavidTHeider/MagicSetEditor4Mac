package domain_test

// **Handlungen mit Schlüsselwort** — Werkstück M7 der Regelmaschine
// (#1609).
//
// # Je Handlung ihre Regelnummer
//
// **So verlangt es die Abnahme des Issues**: „Je umgesetzter Handlung ein
// Lauf mit ihrer Regelnummer im Kopf."

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einHandlungstisch gibt eine Partie mit einer Kreatur, einem Zauber auf dem
// Stapel und vier Karten in der Bibliothek.
func einHandlungstisch(t *testing.T) domain.Partie {
	t.Helper()

	partie := einePartie()

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)
	sitze[0].Deck = domain.Partiedeck{
		Abdruck: "abcd",
		Karten: []domain.Abzugskarte{{
			Kennung: "baer", Name: "Bär", Menge: 1,
			Seiten: []domain.Abzugsseite{{Felder: map[string]string{
				"super type": "Creature", "sub type": "Bear",
				"power": "2", "toughness": "2",
			}}},
		}},
	}
	partie.Sitze = sitze

	partie.Objekte = append(partie.Objekte,
		domain.Spielobjekt{
			Kennung: "sitz-1/baer/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneTisch, Karte: "baer",
		},
		domain.Spielobjekt{
			Kennung: "sitz-1/zauber/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneStapel, Karte: "baer",
		},
		domain.Spielobjekt{
			Kennung: "sitz-1/hand/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneHand, Karte: "baer",
		})

	for stelle := range 4 {
		partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
			Kennung:     "sitz-1/bib/" + string(rune('a'+stelle)),
			Art:         domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneBibliothek, Karte: "baer",
			Ordnung: stelle,
		})
	}

	return partie
}

// **Zerstören legt ein Permanent in den Friedhof seines Eigners** —
// CR 701.8a.
func TestZerstoerenLegtInDenFriedhof(t *testing.T) {
	t.Parallel()

	danach, fehler := einHandlungstisch(t).Zerstoeren("sitz-1/baer/0")
	if fehler != nil {
		t.Fatalf("zerstören: %v", fehler)
	}

	if baer := objektMit(t, danach,
		"sitz-1/baer/0"); baer.Zone != domain.ZoneFriedhof {
		t.Errorf("der Bär liegt in %q", baer.Zone)
	}
}

// **Opfern verlangt den Beherrscher** — CR 701.21a.
func TestOpfernVerlangtDenBeherrscher(t *testing.T) {
	t.Parallel()

	partie := einHandlungstisch(t)

	danach, fehler := partie.Opfern("sitz-1", "sitz-1/baer/0")
	if fehler != nil {
		t.Fatalf("opfern: %v", fehler)
	}

	if baer := objektMit(t, danach,
		"sitz-1/baer/0"); baer.Zone != domain.ZoneFriedhof {
		t.Errorf("der Bär liegt in %q", baer.Zone)
	}

	// **Die Gegenprobe** (§10): ein fremdes Permanent opfert niemand.
	if _, fehler := partie.Opfern("sitz-2",
		"sitz-1/baer/0"); fehler == nil ||
		!strings.Contains(fehler.Error(), "CR 701.21a") {
		t.Errorf("der Fehler lautet %v", fehler)
	}
}

// **Exilieren geht von überall** — CR 701.13a.
func TestExilierenGehtVonUeberall(t *testing.T) {
	t.Parallel()

	partie := einHandlungstisch(t)

	for _, objekt := range []string{
		"sitz-1/baer/0", "sitz-1/hand/0", "sitz-1/zauber/0", "sitz-1/bib/a",
	} {
		t.Run(objekt, func(t *testing.T) {
			t.Parallel()

			danach, fehler := partie.Exilieren(objekt)
			if fehler != nil {
				t.Fatalf("exilieren: %v", fehler)
			}

			if eines := objektMit(t, danach,
				objekt); eines.Zone != domain.ZoneExil {
				t.Errorf("%q liegt in %q", objekt, eines.Zone)
			}
		})
	}
}

// **Neutralisieren nimmt vom Stapel** — CR 701.6a.
func TestNeutralisierenNimmtVomStapel(t *testing.T) {
	t.Parallel()

	partie := einHandlungstisch(t)

	danach, fehler := partie.Neutralisieren("sitz-1/zauber/0")
	if fehler != nil {
		t.Fatalf("neutralisieren: %v", fehler)
	}

	switch {
	case objektMit(t, danach, "sitz-1/zauber/0").Zone != domain.ZoneFriedhof:
		t.Errorf("der Zauber liegt in %q",
			objektMit(t, danach, "sitz-1/zauber/0").Zone)
	case danach.StapelOben() != nil:
		t.Error("der Stapel ist nicht leer")
	}

	// **Die Gegenprobe**: was nicht auf dem Stapel liegt, wird nicht
	// neutralisiert.
	if _, fehler := partie.Neutralisieren(
		"sitz-1/baer/0"); fehler == nil ||
		!strings.Contains(fehler.Error(), "CR 701.6a") {
		t.Errorf("der Fehler lautet %v", fehler)
	}
}

// **Mahlen legt die obersten Karten in den Friedhof** — CR 701.17a.
func TestMahlenLegtDieOberstenInDenFriedhof(t *testing.T) {
	t.Parallel()

	danach, fehler := einHandlungstisch(t).Mahlen("sitz-1", 2)
	if fehler != nil {
		t.Fatalf("mahlen: %v", fehler)
	}

	switch {
	case objektMit(t, danach, "sitz-1/bib/a").Zone != domain.ZoneFriedhof:
		t.Errorf("die oberste liegt in %q",
			objektMit(t, danach, "sitz-1/bib/a").Zone)
	case objektMit(t, danach, "sitz-1/bib/b").Zone != domain.ZoneFriedhof:
		t.Errorf("die zweite liegt in %q",
			objektMit(t, danach, "sitz-1/bib/b").Zone)

	// **Und die dritte bleibt liegen** — die Gegenprobe (§10): ein Weg, der
	// die ganze Bibliothek mahlte, käme sonst durch.
	case objektMit(t, danach, "sitz-1/bib/c").Zone != domain.ZoneBibliothek:
		t.Errorf("die dritte liegt in %q",
			objektMit(t, danach, "sitz-1/bib/c").Zone)
	}
}

// **Mehr mahlen als da ist, ist kein Fehler** — CR 701.17a.
//
// **Es geht, so weit es geht.** Ob daraus ein Verlust wird, entscheidet
// CR 704.5b beim nächsten Ziehen — nicht diese Handlung.
func TestMehrMahlenAlsDaIstIstKeinFehler(t *testing.T) {
	t.Parallel()

	danach, fehler := einHandlungstisch(t).Mahlen("sitz-1", 99)
	if fehler != nil {
		t.Fatalf("mahlen: %v", fehler)
	}

	for _, kennung := range []string{
		"sitz-1/bib/a", "sitz-1/bib/b", "sitz-1/bib/c", "sitz-1/bib/d",
	} {
		if objektMit(t, danach, kennung).Zone != domain.ZoneFriedhof {
			t.Errorf("%q liegt in %q", kennung,
				objektMit(t, danach, kennung).Zone)
		}
	}
}

// **Zwei Kreaturen kämpfen und teilen einander ihre Stärke zu** — CR 701.14a.
func TestZweiKreaturenKaempfen(t *testing.T) {
	t.Parallel()

	partie := einHandlungstisch(t)

	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-2/baer/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-2", Eigner: "sitz-2",
		Zone: domain.ZoneTisch, Karte: "baer",
	})

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)
	sitze[1].Deck = sitze[0].Deck
	partie.Sitze = sitze

	danach, fehler := partie.Kaempfen("sitz-1/baer/0", "sitz-2/baer/0")
	if fehler != nil {
		t.Fatalf("kämpfen: %v", fehler)
	}

	switch {
	case objektMit(t, danach, "sitz-1/baer/0").Schaden != 2:
		t.Errorf("der erste trägt %d Schaden",
			objektMit(t, danach, "sitz-1/baer/0").Schaden)
	case objektMit(t, danach, "sitz-2/baer/0").Schaden != 2:
		t.Errorf("der zweite trägt %d Schaden",
			objektMit(t, danach, "sitz-2/baer/0").Schaden)
	}

	// **Und niemand kämpft gegen sich selbst.**
	if _, fehler := partie.Kaempfen("sitz-1/baer/0",
		"sitz-1/baer/0"); fehler == nil {
		t.Error("eine Kreatur kämpfte gegen sich selbst")
	}
}

// **Hellsehen legt die Gewählten nach unten** — CR 701.22a.
func TestHellsehenLegtNachUnten(t *testing.T) {
	t.Parallel()

	partie := einHandlungstisch(t)

	// **Der Blick vor der Wahl** — CR 701.22a: erst ansehen, dann wählen.
	oben := partie.Obenauf("sitz-1", 2)

	if len(oben) != 2 || oben[0] != "sitz-1/bib/a" {
		t.Fatalf("obenauf liegen %v", oben)
	}

	danach, fehler := partie.Hellsehen("sitz-1", 2, []string{"sitz-1/bib/a"})
	if fehler != nil {
		t.Fatalf("hellsehen: %v", fehler)
	}

	// **Die zweite liegt jetzt oben, die erste ganz unten.**
	nachher := danach.Obenauf("sitz-1", 4)

	erwartet := []string{
		"sitz-1/bib/b", "sitz-1/bib/c", "sitz-1/bib/d", "sitz-1/bib/a",
	}

	if strings.Join(nachher, ",") != strings.Join(erwartet, ",") {
		t.Errorf("die Bibliothek steht als %v", nachher)
	}
}

// **Überwachen legt die Gewählten in den Friedhof** — CR 701.25a.
func TestUeberwachenLegtInDenFriedhof(t *testing.T) {
	t.Parallel()

	danach, fehler := einHandlungstisch(t).
		Ueberwachen("sitz-1", 2, []string{"sitz-1/bib/a"})
	if fehler != nil {
		t.Fatalf("überwachen: %v", fehler)
	}

	switch {
	case objektMit(t, danach, "sitz-1/bib/a").Zone != domain.ZoneFriedhof:
		t.Errorf("die gewählte liegt in %q",
			objektMit(t, danach, "sitz-1/bib/a").Zone)
	case objektMit(t, danach, "sitz-1/bib/b").Zone != domain.ZoneBibliothek:
		t.Errorf("die zweite liegt in %q",
			objektMit(t, danach, "sitz-1/bib/b").Zone)
	}

	// **Und sie liegt wieder oben** — der Rest bleibt in der Reihenfolge.
	if oben := danach.Obenauf("sitz-1", 1); len(oben) != 1 ||
		oben[0] != "sitz-1/bib/b" {
		t.Errorf("obenauf liegen %v", oben)
	}
}

// **Was nicht obenauf liegt, lässt sich nicht wählen** — CR 701.22a.
//
// **Sonst wäre die Wahl ein Griff in die ganze Bibliothek** — und die ist
// auch für ihren Eigner verdeckt (G1).
func TestNurWasObenaufLiegtLaesstSichWaehlen(t *testing.T) {
	t.Parallel()

	_, fehler := einHandlungstisch(t).
		Hellsehen("sitz-1", 2, []string{"sitz-1/bib/d"})

	if fehler == nil || !strings.Contains(fehler.Error(), "CR 701.22a") {
		t.Errorf("der Fehler lautet %v", fehler)
	}
}
