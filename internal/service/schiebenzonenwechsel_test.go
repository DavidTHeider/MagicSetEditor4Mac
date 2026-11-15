package service_test

// **Auch ein Schieben von Hand macht ein neues Objekt** — CR 400.7,
// Werkstück M18
// (#1619).
//
// # Warum das hier steht und nicht nur in der Domäne
//
// **Weil `objekt.schieben` der Weg ist, den ein Mensch nimmt.** Die
// Regelmaschine legt Karten über `Zerstoeren`, `Wirken` und den
// Aufräumschritt um; am Tisch legt jemand sie mit der Hand um, und das ist
// derselbe Zonenwechsel. Der Kopf von `sandkastenschieben.go` nannte das bis
// zum 2026-09-17 „eine Regelfrage für die Matrix von G4".

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Marken und Zustand bleiben beim Schieben zurück** — CR 400.7.
func TestEinGeschobenesObjektIstEinNeues(t *testing.T) {
	t.Parallel()

	partie := einHandlungsstand(t)
	partie.Objekte[0].Zaehler = map[string]int{domain.MarkePlusEins: 2}
	partie.Objekte[0].Getappt = true
	partie.Objekte[0].Schaden = 1

	_, danach, fehler := service.NeuerSandkasten(festeMischung{}).Auslegen(
		partie, domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art: "objekt.schieben",
			Nutzlast: map[string]any{
				"objekt": "sitz-1/baer/0", "zone": "friedhof",
			},
		})
	if fehler != nil {
		t.Fatalf("schieben: %v", fehler)
	}

	baer := objektAus(t, danach, "sitz-1/baer/0")

	if anzahl := baer.Zaehler[domain.MarkePlusEins]; anzahl != 0 {
		t.Errorf("auf dem Bären liegen noch %d Marken", anzahl)
	}

	if baer.Getappt {
		t.Error("der Bär liegt getappt im Friedhof")
	}

	if baer.Schaden != 0 {
		t.Errorf("der Bär trägt %d markierten Schaden", baer.Schaden)
	}
}

// **Und was der Befehl sagt, gilt danach.**
//
// **Die Gegenprobe zur Zeile darüber**: „lege sie getappt hin" ist eine
// Angabe über das **neue** Objekt. Ohne diesen Lauf sagte der andere nur,
// dass irgendetwas geräumt wird — auch der Wunsch des Aufrufers.
func TestEinGeschobenesObjektNimmtDenGewuenschtenZustandAn(t *testing.T) {
	t.Parallel()

	partie := einHandlungsstand(t)
	partie.Objekte[0].Getappt = false

	_, danach, fehler := service.NeuerSandkasten(festeMischung{}).Auslegen(
		partie, domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art: "objekt.schieben",
			Nutzlast: map[string]any{
				"objekt": "sitz-1/baer/0", "zone": "exil", "getappt": true,
			},
		})
	if fehler != nil {
		t.Fatalf("schieben: %v", fehler)
	}

	if !objektAus(t, danach, "sitz-1/baer/0").Getappt {
		t.Error("der gewünschte Zustand ist nicht angekommen")
	}
}

// **Und was daran hing, hängt an nichts mehr** — CR 400.7 aus der anderen
// Richtung.
//
// # An einer alten Partie gemessen, und das ist der ganze Punkt
//
// **Eine Partie mit `zustandsbasiert: automatisch` beantwortet die Frage
// nicht.** Dort räumt CR 704.5m die Aura ohnehin ab — sie hängt an nichts
// und geht in den Friedhof —, und ein Lauf darüber wäre grün, auch wenn das
// Ablösen hier fehlte. **Gemessen am 2026-09-17**: die Gegenprobe blieb
// grün, bis dieser Lauf das Profil zurückstellte (§10).
//
// **Die Partien aus G1 bis G8 sind solche alten** (#1461):
// ihr Profil ist eingefroren, sie laufen unverändert weiter, und in ihnen
// ist das Ablösen hier die **einzige** Stelle, die es tut.
func TestEinGeschobenesObjektLaesstSeineAnhaengselLos(t *testing.T) {
	t.Parallel()

	partie := einHandlungsstand(t)

	for stelle := range partie.Profil.Matrix.Eintraege {
		if partie.Profil.Matrix.Eintraege[stelle].Bereich ==
			domain.BereichZustandsbasiert {
			partie.Profil.Matrix.Eintraege[stelle].Grad = domain.UnterstuetzungManuell
		}
	}

	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-1/aura/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneTisch, Karte: "baer",
		HaengtAn: "sitz-1/baer/0",
	})

	_, danach, fehler := service.NeuerSandkasten(festeMischung{}).Auslegen(
		partie, domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art: "objekt.schieben",
			Nutzlast: map[string]any{
				"objekt": "sitz-1/baer/0", "zone": "hand",
			},
		})
	if fehler != nil {
		t.Fatalf("schieben: %v", fehler)
	}

	// **Sie liegt noch da** — in einer alten Partie setzt nichts CR 704.5m
	// durch. Geprüft wird, dass sie **an nichts mehr hängt**.
	if aura := objektAus(t, danach,
		"sitz-1/aura/0"); aura.HaengtAn != "" {
		t.Errorf("die Aura hängt weiter an %q", aura.HaengtAn)
	}
}
