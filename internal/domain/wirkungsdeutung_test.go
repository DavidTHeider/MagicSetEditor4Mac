package domain_test

// **Gelesen wird beim Verrechnen** — CR 608.2c (Fassung 2026-08-19),
// #1625.
//
// # Wogegen hier geprüft wird
//
// **Gegen eine Domäne, die einen Auswerter voraussetzt.** Sie hat keinen:
// `internal/domain` importiert nichts aus dem Projekt (§2), und eine Partie
// ohne eingesetzten Deuter führt kein Kartenskript aus. Das ist der Stand
// jeder Partie vor `M8` und der Normalfall in einem Lauf, der keinen stellt.
//
// **Und gegen einen Deuter, der den Stand von vorhin bekommt.** Was er sieht,
// muss aus dem Augenblick der Verrechnung stammen.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// deuterprobe merkt sich, wonach gefragt wurde, und antwortet fest.
type deuterprobe struct {
	gefragt []string
	lagen   []domain.Skriptlage

	gibt   []domain.Spielwirkung
	befund string
}

func (d *deuterprobe) Wirkungen(
	skript string, lage domain.Skriptlage,
) (wirkungen []domain.Spielwirkung, befund string) {
	d.gefragt = append(d.gefragt, skript)
	d.lagen = append(d.lagen, lage)

	return d.gibt, d.befund
}

// eineSkriptpartie legt einen Zauber mit Skript auf den Stapel.
//
// **Die Felder stehen am Objekt** — [domain.Partie.Eigenschaften] nimmt sie
// dann unmittelbar, und der Lauf braucht kein Deck.
func eineSkriptpartie(skript string, ziele []domain.Ziel) domain.Partie {
	partie := eineKartenpartie()

	felder := map[string]string{"name": "Blitz", "super type": "Instant"}
	if skript != "" {
		felder["game script"] = skript
	}

	partie.Objekte = []domain.Spielobjekt{{
		Kennung: "blitz", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneStapel, Ordnung: 1,
		Felder: felder, Ziele: ziele,
	}}

	return partie.ZugBeginnen("sitz-1")
}

// **Ohne Deuter geschieht nichts — und es ist kein Fehler.**
func TestEinePartieOhneDeuterFuehrtKeinSkriptAus(t *testing.T) {
	t.Parallel()

	partie := eineSkriptpartie("draw(amount: 2)", nil)

	danach := partie.Verrechnen()

	if danach.Skriptbefund() != "" {
		t.Errorf("der Befund ist %q", danach.Skriptbefund())
	}

	// **Der Zauber ist trotzdem verrechnet** — CR 608.2n.
	if zone := zoneVon(danach, "blitz"); zone != domain.ZoneFriedhof {
		t.Errorf("der Blitz liegt in %q", zone)
	}
}

// **Mit Deuter wird beim Verrechnen gefragt — und die Lage stimmt.**
func TestDerDeuterWirdBeimVerrechnenGefragt(t *testing.T) {
	t.Parallel()

	probe := &deuterprobe{}

	partie := eineSkriptpartie("draw(amount: 2)", []domain.Ziel{
		{Kennung: "sitz-2", Zone: domain.ZoneTisch},
	}).MitDeuter(probe)

	// **Vor dem Verrechnen wurde nichts gefragt.**
	if len(probe.gefragt) != 0 {
		t.Fatalf("gefragt wurde schon %dmal", len(probe.gefragt))
	}

	danach := partie.Verrechnen()

	if len(probe.gefragt) != 1 || probe.gefragt[0] != "draw(amount: 2)" {
		t.Fatalf("gefragt wurde nach %v", probe.gefragt)
	}

	lage := probe.lagen[0]
	if lage.Beherrscher != "sitz-1" || lage.Objekt != "blitz" {
		t.Errorf("die Lage ist %+v", lage)
	}

	// **Die Ziele kommen als Kennungen** — wo sie liegen, hat CR 608.2b
	// schon entschieden.
	if len(lage.Ziele) != 1 || lage.Ziele[0] != "sitz-2" {
		t.Errorf("die Ziele sind %v", lage.Ziele)
	}

	if danach.Skriptbefund() != "" {
		t.Errorf("der Befund ist %q", danach.Skriptbefund())
	}
}

// **Eine Karte ohne Skript wird nicht gefragt.**
func TestEineKarteOhneSkriptFragtDenDeuterNicht(t *testing.T) {
	t.Parallel()

	probe := &deuterprobe{}

	danach := eineSkriptpartie("", nil).MitDeuter(probe).Verrechnen()

	if len(probe.gefragt) != 0 {
		t.Errorf("gefragt wurde %dmal", len(probe.gefragt))
	}

	if danach.Skriptbefund() != "" {
		t.Errorf("der Befund ist %q", danach.Skriptbefund())
	}
}

// **Ein Befund steht an der Partie und hält sie nicht an** (§8).
func TestEinBefundStehtAnDerPartie(t *testing.T) {
	t.Parallel()

	probe := &deuterprobe{befund: "das Kartenskript ist nicht durchgelaufen"}

	danach := eineSkriptpartie("a()", nil).MitDeuter(probe).Verrechnen()

	if danach.Skriptbefund() != probe.befund {
		t.Errorf("der Befund ist %q", danach.Skriptbefund())
	}

	// **Und der Zauber ist verrechnet** — CR 608.2n.
	if zone := zoneVon(danach, "blitz"); zone != domain.ZoneFriedhof {
		t.Errorf("der Blitz liegt in %q", zone)
	}
}

// zoneVon sagt, wo ein Objekt liegt — leer, wenn es keines gibt.
func zoneVon(partie domain.Partie, kennung string) domain.Zone {
	for _, objekt := range partie.Objekte {
		if objekt.Kennung == kennung {
			return objekt.Zone
		}
	}

	return ""
}

// **Ein Bereich, der einer Variante gehört, ist keine Lücke** —
// [0123](../../docs/entscheidungen/0123-streng-zaehlt-nur-was-zur-variante-gehoert.md),
// #1626.
//
// **Der Lauf zeigt den Unterschied an zwei Profilen**: eine Sandkastenpartie
// führt `commander` nicht als offenen Rückstand, eine Commanderpartie schon
// — dort steht er auf `automatisch` und zählt.
func TestCommanderIstNurInSeinerVarianteEineLuecke(t *testing.T) {
	t.Parallel()

	sandkasten := domain.Grundprofil("magic").Matrix
	for _, bereich := range sandkasten.Luecken() {
		if bereich == domain.BereichCommander {
			t.Error("die Sandkastenpartie führt commander als Lücke")
		}
	}

	// **Und die zwei, die wirklich fehlen, stehen weiter da.**
	if len(sandkasten.Luecken()) != 2 {
		t.Errorf("die Lücken sind %v", sandkasten.Luecken())
	}

	// **In der Variante zählt er** — und er ist dort keine Lücke, weil er
	// automatisch ist.
	variante := domain.Commanderprofil("magic").Matrix
	if variante.Grad(domain.BereichCommander) !=
		domain.UnterstuetzungAutomatisch {
		t.Errorf("in der Variante steht commander auf %q",
			variante.Grad(domain.BereichCommander))
	}
}
