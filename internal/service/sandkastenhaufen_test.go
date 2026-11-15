package service_test

// **Der Haufen als Befehl** — CR 400.5,
// #1756.
//
// # Was hier geprüft wird und was daneben
//
// **Hier: dass der Befehl ankommt.** Ob zwei Objekte einen Haufen bilden
// *dürfen*, ist eine Aussage über die Partie und steht in
// `internal/domain/haufen_test.go` — vierzehn Fälle, darunter jeder Zustand,
// den ein Haufen verschleiern könnte.
//
// **Und dass die Ablehnung ankommt.** Eine Prüfung, die nur den Erfolgsweg
// betritt, hat die Hälfte ausgelassen (§10).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// tischMitZweiWaeldern gibt eine Partie mit zwei gleichen Ländern.
func tischMitZweiWaeldern() domain.Partie {
	partie := tischMitBibliothek()

	partie.Objekte = append(partie.Objekte,
		domain.Spielobjekt{
			Kennung: "w1", Art: domain.ObjektKarte, Beherrscher: "anna",
			Eigner: "anna", Zone: domain.ZoneTisch, Karte: "wald",
		},
		domain.Spielobjekt{
			Kennung: "w2", Art: domain.ObjektKarte, Beherrscher: "anna",
			Eigner: "anna", Zone: domain.ZoneTisch, Karte: "wald",
		},
	)

	return partie
}

// haufenVon gibt den Haufen eines Objekts.
func haufenVon(t *testing.T, partie domain.Partie, kennung string) string {
	t.Helper()

	for _, objekt := range partie.Objekte {
		if objekt.Kennung == kennung {
			return objekt.Haufen
		}
	}

	t.Fatalf("%q liegt nicht am Tisch", kennung)

	return ""
}

func TestDerBefehlLegtEineKarteAufEinenHaufen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, danach, fehler := kasten.Auslegen(tischMitZweiWaeldern(),
		domain.Spielbefehl{
			PartieID: "p1", Absender: "anna", Kennung: "w1",
			Art:      service.BefehlHaeufen,
			Nutzlast: map[string]any{"objekt": "w1", "worauf": "w2"},
		})
	if fehler != nil {
		t.Fatalf("häufen: %v", fehler)
	}

	if haufenVon(t, danach, "w1") != "w2" {
		t.Errorf("w1 liegt im Haufen %q", haufenVon(t, danach, "w1"))
	}

	// **Auch das Ziel trägt den Haufen.** Sonst läge nur eine Karte darin,
	// und die andere daneben — ein Haufen aus einem.
	if haufenVon(t, danach, "w2") != "w2" {
		t.Errorf("w2 trägt den Haufen %q", haufenVon(t, danach, "w2"))
	}
}

// **Ohne Ziel gibt es keinen Haufen.** Ein Befehl, der die leere Zeichenkette
// als Ziel nähme, suchte ein Objekt ohne Kennung — und die Meldung käme aus
// der Tiefe statt vom Feld (§8).
func TestEinHaufenBefehlOhneZielWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, _, fehler := kasten.Auslegen(tischMitZweiWaeldern(),
		domain.Spielbefehl{
			PartieID: "p1", Absender: "anna", Kennung: "w1",
			Art:      service.BefehlHaeufen,
			Nutzlast: map[string]any{"objekt": "w1"},
		})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("der Befehl ohne Ziel ging durch: %v", fehler)
	}
}

// **Und die Regel der Domäne kommt als Fehler an.** Ohne diesen Lauf stünde
// der Weg dahin ungeprüft da: der Dienst könnte den Fehler schlucken, und
// der Tisch zeigte einen Haufen, den es nicht geben darf.
func TestEineAbgelehnteAnordnungKommtAlsFehlerAn(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := tischMitZweiWaeldern()

	for stelle := range partie.Objekte {
		if partie.Objekte[stelle].Kennung == "w1" {
			partie.Objekte[stelle].Getappt = true
		}
	}

	_, _, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "anna", Kennung: "w1",
		Art:      service.BefehlHaeufen,
		Nutzlast: map[string]any{"objekt": "w1", "worauf": "w2"},
	})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("die getappte Karte kam auf den Haufen: %v", fehler)
	}
}

func TestDerBefehlNimmtEineKarteAusDemHaufen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, gehaeuft, fehler := kasten.Auslegen(tischMitZweiWaeldern(),
		domain.Spielbefehl{
			PartieID: "p1", Absender: "anna", Kennung: "w1",
			Art:      service.BefehlHaeufen,
			Nutzlast: map[string]any{"objekt": "w1", "worauf": "w2"},
		})
	if fehler != nil {
		t.Fatalf("häufen: %v", fehler)
	}

	_, danach, fehler := kasten.Auslegen(gehaeuft, domain.Spielbefehl{
		PartieID: "p1", Absender: "anna", Kennung: "w1",
		Art:      service.BefehlEnthaeufen,
		Nutzlast: map[string]any{"objekt": "w1"},
	})
	if fehler != nil {
		t.Fatalf("enthäufen: %v", fehler)
	}

	if haufenVon(t, danach, "w1") != "" {
		t.Errorf("w1 liegt noch im Haufen %q", haufenVon(t, danach, "w1"))
	}
}

// **Auch hier die Ablehnung**: was in keinem Haufen liegt, lässt sich nicht
// herausnehmen.
func TestEnthaeufenOhneHaufenWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, _, fehler := kasten.Auslegen(tischMitZweiWaeldern(),
		domain.Spielbefehl{
			PartieID: "p1", Absender: "anna", Kennung: "w1",
			Art:      service.BefehlEnthaeufen,
			Nutzlast: map[string]any{"objekt": "w1"},
		})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("w1 ließ sich aus keinem Haufen nehmen: %v", fehler)
	}
}
