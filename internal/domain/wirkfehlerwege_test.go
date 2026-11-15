package domain_test

// **Die Fehlerwege des Wirkens** — Werkstück M4 der Regelmaschine
// (#1606).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, was ein Zauber tut, wenn er zustande kommt**; hier steht,
// **woran er scheitert** — und `wirken_test.go` stand mit beidem bei 508
// Zeilen (§2).

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jeder Weg, auf dem ein Zauber nicht zustande kommt, sagt warum** — §8.
//
// **Und jeder lässt die Partie, wie sie war** — CR 733. Das wird hier nicht
// je Fall noch einmal zugesichert; es folgt daraus, dass eine [domain.Partie]
// ein **Wert** ist, und der Lauf darüber prüft es an dem Fall, an dem am
// meisten schiefgehen kann.
func TestJederFehlerwegDesWirkensSagtWarum(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		bauen  func(t *testing.T) domain.Partie
		wer    string
		objekt string
		ziele  []string
		wort   string
	}{
		"ein Objekt, das es nicht gibt": {
			bauen: func(t *testing.T) domain.Partie { return eineWirkpartie(t, "[R]") },
			wer:   "sitz-1", objekt: "gibtsnicht",
			wort: "gibt es in dieser Partie nicht",
		},

		"eine fremde Karte": {
			bauen:  mitFremderKarte,
			wer:    "sitz-1",
			objekt: "sitz-2/fremd/0",
			wort:   "beherrscht das Objekt",
		},

		"etwas, das schon auf dem Stapel liegt": {
			bauen:  aufDemStapel,
			wer:    "sitz-1",
			objekt: "sitz-1/blitz/0",
			wort:   "liegt schon auf dem Stapel",
		},

		"eine Karte ohne Abzug": {
			bauen:  ohneEigenenAbzug,
			wer:    "sitz-1",
			objekt: "sitz-1/blitz/0",
			wort:   "eingefrorenen Abzug",
		},

		"Kosten mit X": {
			bauen: func(t *testing.T) domain.Partie {
				return eineWirkpartie(t, "[X][R]")
			},
			wer: "sitz-1", objekt: "sitz-1/blitz/0",
			wort: "lassen sich hier nicht zahlen",
		},

		"ein Sitz, den es nicht gibt": {
			bauen:  ohneEigenenSitz,
			wer:    "sitz-1",
			objekt: "sitz-1/blitz/0",
			wort:   "gibt es in dieser Partie nicht",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := fall.bauen(t).Wirken(fall.wer, fall.objekt, fall.ziele)

			if fehler == nil {
				t.Fatal("es kam kein Fehler")
			}

			if !strings.Contains(fehler.Error(), fall.wort) {
				t.Errorf("der Fehler lautet %q", fehler)
			}
		})
	}
}

// aufDemStapel legt den Blitz schon einmal hinauf.
func aufDemStapel(t *testing.T) domain.Partie {
	t.Helper()

	partie := eineWirkpartie(t, "[R]")
	partie.Objekte[len(partie.Objekte)-1].Zone = domain.ZoneStapel

	return partie
}

// ohneEigenenAbzug nimmt dem Sitz seinen eingefrorenen Abzug.
func ohneEigenenAbzug(t *testing.T) domain.Partie {
	t.Helper()

	partie := eineWirkpartie(t, "[R]")
	partie.Sitze[0].Deck = domain.Partiedeck{Abdruck: "abcd"}

	return partie
}

// ohneEigenenSitz nimmt den Sitz weg, dem die Karte gehört.
//
// **Ein Zustand aus einem von Hand geänderten Bestand** (§8) — er darf nicht
// stillschweigend durchgehen, und deshalb hat er einen Lauf.
//
// **Der Abzug bleibt an Ort und Stelle**: die Karte muss lesbar sein, damit
// der Lauf am **Bezahlen** scheitert und nicht schon an den Kosten.
func ohneEigenenSitz(t *testing.T) domain.Partie {
	t.Helper()

	partie := eineWirkpartie(t, "")

	// **Die Karte wandert an den zweiten Sitz**, und der erste fällt weg —
	// damit findet `Eigenschaften` sie noch und `manaAbziehen` den Sitz
	// nicht mehr.
	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)
	sitze[1].Deck = sitze[0].Deck
	partie.Sitze = sitze[1:]

	objekte := make([]domain.Spielobjekt, len(partie.Objekte))
	copy(objekte, partie.Objekte)
	objekte[len(objekte)-1].Eigner = "sitz-2"
	partie.Objekte = objekte

	partie.Zug.Prioritaet = "sitz-1"

	return partie
}

// mitFremderKarte legt eine Karte an den zweiten Sitz.
func mitFremderKarte(t *testing.T) domain.Partie {
	t.Helper()

	partie := eineWirkpartie(t, "[R]")

	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-2/fremd/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-2", Eigner: "sitz-2",
		Zone: domain.ZoneHand, Karte: "blitz",
	})

	return partie
}

// **Ein fremder Sitz ist als Ziel zulässig, ein fremdes Objekt auch** —
// CR 115.1. Was zählt, ist die Zone beim Ansagen.
func TestEinZielDasDieZoneHaeltBleibtZulaessig(t *testing.T) {
	t.Parallel()

	partie, fehler := eineWirkpartie(t, "[R]").
		ManaHinzu("sitz-1", domain.ManaRot, 1)
	if fehler != nil {
		t.Fatalf("mana hinzu: %v", fehler)
	}

	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-2/baer/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-2", Eigner: "sitz-2",
		Zone: domain.ZoneTisch, Karte: "baer",
	})

	danach, fehler := partie.Wirken("sitz-1", "sitz-1/blitz/0",
		[]string{"sitz-2/baer/0"})
	if fehler != nil {
		t.Fatalf("wirken: %v", fehler)
	}

	// **Das Ziel bleibt liegen** — der Zauber wird verrechnet.
	nachher := danach.Verrechnen()

	blitz := objektMit(t, nachher, "sitz-1/blitz/0")

	if blitz.Zone != domain.ZoneFriedhof {
		t.Errorf("der Blitz liegt in %q", blitz.Zone)
	}
}

// **Ein Ziel, das ganz verschwunden ist, macht den Zauber verpuffen** —
// CR 608.2b.
func TestEinVerschwundenesZielLaesstVerpuffen(t *testing.T) {
	t.Parallel()

	partie, fehler := eineWirkpartie(t, "[R]").
		ManaHinzu("sitz-1", domain.ManaRot, 1)
	if fehler != nil {
		t.Fatalf("mana hinzu: %v", fehler)
	}

	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-2/baer/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-2", Eigner: "sitz-2",
		Zone: domain.ZoneTisch, Karte: "baer",
	})

	danach, fehler := partie.Wirken("sitz-1", "sitz-1/blitz/0",
		[]string{"sitz-2/baer/0"})
	if fehler != nil {
		t.Fatalf("wirken: %v", fehler)
	}

	// **Das Ziel gibt es nicht mehr** — ein Spielstein, der aufhört zu
	// existieren (CR 111.7), oder ein von Hand geänderter Bestand.
	behalten := make([]domain.Spielobjekt, 0, len(danach.Objekte))

	for _, objekt := range danach.Objekte {
		if objekt.Kennung != "sitz-2/baer/0" {
			behalten = append(behalten, objekt)
		}
	}

	danach.Objekte = behalten

	if blitz := objektMit(t, danach.Verrechnen(),
		"sitz-1/blitz/0"); blitz.Zone != domain.ZoneFriedhof {
		t.Errorf("der Blitz liegt in %q", blitz.Zone)
	}
}
