package service_test

// **Der Blick nennt die Karten beim Namen** —
// #1836.
//
// # Wogegen hier geprüft wird
//
// **Dass der Name aus dem Abzug kommt.** `Spielobjekt.Name` steht nur, wo
// keine Karte dahintersteht — bei einem Spielstein; bei einer Karte ist er
// leer, und bis zum 2026-09-24 stand im privaten Ergebnis `name: ""` für
// jede Karte. Gesehen mit dem kanonischen Set „Green" am Tisch: „Hineinsehen"
// nannte keine einzige beim Namen.
//
// **Warum `TestEinBlickZeigtDieZoneNurDemAbsender` es nie sah**: seine
// Objekte verweisen auf eine Karte, die ihr Abzug nicht führt — und dann gibt
// es keinen Namen, den man vermissen könnte.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// tischMitWald ist [tischMitBibliothek], dessen Karten ihr Abzug führt.
func tischMitWald() domain.Partie {
	partie := tischMitBibliothek()

	for stelle := range partie.Sitze {
		partie.Sitze[stelle].Deck = domain.Partiedeck{
			Abdruck: "abc",
			Karten: []domain.Abzugskarte{{
				Kennung: "k", Name: "Wald", Menge: 3,
				Seiten: []domain.Abzugsseite{{
					Name: "Wald", Felder: map[string]string{"name": "Wald"},
				}},
			}},
		}
	}

	return partie
}

func TestEinBlickNenntDieKartenBeimNamen(t *testing.T) {
	t.Parallel()

	ereignisse, _, fehler := service.NeuerSandkasten(festeMischung{}).Auslegen(
		tischMitWald(), domain.Spielbefehl{
			PartieID: "p1", Absender: "anna", Kennung: "b1",
			Art:      service.BefehlBlick,
			Nutzlast: map[string]any{"zone": "bibliothek"},
		})
	if fehler != nil {
		t.Fatalf("blicken: %v", fehler)
	}

	karten, istListe := ereignisse[0].Privat[0].Inhalt["karten"].([]any)
	if !istListe || len(karten) != 3 {
		t.Fatalf("privat karten: %+v", ereignisse[0].Privat[0].Inhalt)
	}

	for _, karte := range karten {
		eintrag, istVerbund := karte.(map[string]any)
		if !istVerbund || eintrag["name"] != "Wald" {
			t.Errorf("eine Karte ohne ihren Namen: %+v", karte)
		}
	}
}

// **Ein Spielstein behält den Namen, den er am Objekt trägt** — er kommt aus
// keinem Deck (CR 111.1), und einen Abzug, der ihn nennte, gibt es nicht.
func TestEinBlickNenntEinenSpielsteinBeiSeinemNamen(t *testing.T) {
	t.Parallel()

	partie := tischMitWald()
	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "anna/stein", Art: domain.ObjektSpielstein,
		Beherrscher: "anna", Eigner: "anna",
		Zone: domain.ZoneBibliothek, Name: "Soldat", Ordnung: 4,
	})

	ereignisse, _, fehler := service.NeuerSandkasten(festeMischung{}).Auslegen(
		partie, domain.Spielbefehl{
			PartieID: "p1", Absender: "anna", Kennung: "b1",
			Art:      service.BefehlBlick,
			Nutzlast: map[string]any{"zone": "bibliothek"},
		})
	if fehler != nil {
		t.Fatalf("blicken: %v", fehler)
	}

	// **Die Gestalt hält der Lauf darüber**; hier zählen nur die Namen.
	karten, _ := ereignisse[0].Privat[0].Inhalt["karten"].([]any) //nolint:errcheck // siehe oben

	namen := map[string]bool{}

	for _, karte := range karten {
		eintrag, _ := karte.(map[string]any) //nolint:errcheck // dito
		name, _ := eintrag["name"].(string)  //nolint:errcheck // dito
		namen[name] = true
	}

	if !namen["Soldat"] || !namen["Wald"] {
		t.Errorf("Namen im Blick: %v, erwartet Wald und Soldat", namen)
	}
}
