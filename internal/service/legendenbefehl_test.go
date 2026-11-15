package service_test

// **Die Legendenwahl über den Sandkasten** — CR 704.5j, Werkstück M12
// (#1614).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan stehen die Befehle, die eine Gestalt ändern** — Spielstein,
// Emblem, Kontrolle, Umwandeln, Kopieren. Hier steht der eine, der eine
// **zustandsbasierte Aktion abschließt**, und `gestaltbefehle_test.go` stand
// mit ihm bei 422 Zeilen (§2). Die Schwester dazu — die Reihung ausgelöster
// Fähigkeiten (CR 603.3b) — liegt in `reihungsbefehl_test.go`.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Legendenwahl geht über den Sandkasten** — CR 704.5j.
//
// **Und sie braucht keine Priorität** (wie der Abwurf, CR 514.1): eine
// zustandsbasierte Aktion wird abgearbeitet, **bevor** jemand Priorität
// bekommt (CR 704.3). Wer auf seinen Zug warten müsste, hielte die Partie
// an.
func TestLegendenbefehlEntscheidetDieWahl(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := einLegendenstand(t)

	_, danach, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-2", Kennung: "b1",
		Art:      service.BefehlLegende,
		Nutzlast: map[string]any{"objekt": "sitz-2/held/0"},
	})
	if fehler != nil {
		t.Fatalf("legende: %v", fehler)
	}

	if zone := objektAus(t, danach,
		"sitz-2/held/1").Zone; zone != domain.ZoneFriedhof {
		t.Errorf("die übrige Legende liegt in %q", zone)
	}

	if anzahl := len(danach.Legendenwahlen); anzahl != 0 {
		t.Errorf("es stehen noch %d Wahlen an", anzahl)
	}
}

// **Und sie steht dem an, der nicht am Zug ist.**
func TestDieLegendenwahlBrauchtKeinePrioritaet(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := einLegendenstand(t)

	erlaubt := kasten.Erlaubt(partie, "sitz-2")

	for _, art := range erlaubt {
		if art == service.BefehlLegende {
			return
		}
	}

	t.Errorf("%q steht dem Wartenden nicht offen: %v",
		service.BefehlLegende, erlaubt)
}

// einLegendenstand gibt eine Partie, in der sitz-2 wählen muss.
//
// **Sitz 2 und nicht Sitz 1**: Sitz 1 ist am Zug, und eine Wahl, die nur
// dem aktiven Spieler offenstünde, sagte über CR 704.3 nichts aus.
func einLegendenstand(t *testing.T) domain.Partie {
	t.Helper()

	partie := einHandlungsstand(t)

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)

	for stelle := range sitze {
		sitze[stelle].Deck.Karten = []domain.Abzugskarte{{
			Kennung: "held", Name: "Held", Menge: 2,
			Seiten: []domain.Abzugsseite{{Felder: map[string]string{
				"name": "Held", "super type": "Legendary Creature",
				"power": "2", "toughness": "2",
			}}},
		}}
	}

	partie.Sitze = sitze
	partie.Objekte = []domain.Spielobjekt{
		{
			Kennung: "sitz-2/held/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-2", Eigner: "sitz-2",
			Zone: domain.ZoneTisch, Karte: "held",
		},
		{
			Kennung: "sitz-2/held/1", Art: domain.ObjektKarte,
			Beherrscher: "sitz-2", Eigner: "sitz-2",
			Zone: domain.ZoneTisch, Karte: "held",
		},
	}

	return partie.Zustandspruefung()
}

// **Das Heimholen geht über denselben Weg** — CR 903.9a, Werkstück M19
// (#1620).
//
// # Warum es hier steht und nicht bei den Gestalten
//
// **Weil es dieselbe Gestalt hat wie die Legendenwahl nebenan**: ein Befehl,
// der eine **Erlaubnis** einlöst, statt etwas am Tisch zu ändern. CR 903.9a
// sagt „darf", und der Eigner entscheidet — deshalb nennt der Befehl kein
// zweites Objekt und keinen zweiten Sitz.
//
// **Und weil der Zweig bis zum 2026-09-18 unbetreten war** (§10): er stand
// da, seit `M19` ihn gebaut hat, und kein Lauf des Sandkastens ist je durch
// ihn gegangen.
func TestDerHeimholbefehlNimmtDenCommanderZurueck(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := einCommanderstand(t)

	_, danach, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
		Art:      service.BefehlHeimholen,
		Nutzlast: map[string]any{"objekt": "sitz-1/held/0"},
	})
	if fehler != nil {
		t.Fatalf("heimholen: %v", fehler)
	}

	if zone := objektAus(t, danach, "sitz-1/held/0").Zone; zone !=
		domain.ZoneKommando {
		t.Errorf("der Commander liegt in %q", zone)
	}
}

// **Und ein fremder Commander bleibt liegen** — CR 903.9a nennt den Eigner.
func TestEinFremderCommanderWirdNichtHeimgeholt(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, _, fehler := kasten.Auslegen(einCommanderstand(t), domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-2", Kennung: "b1",
		Art:      service.BefehlHeimholen,
		Nutzlast: map[string]any{"objekt": "sitz-1/held/0"},
	})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("sitz-2 durfte heimholen: %v", fehler)
	}
}

// einCommanderstand gibt eine Partie, in der sitz-1 seinen Commander im
// Friedhof liegen hat.
func einCommanderstand(t *testing.T) domain.Partie {
	t.Helper()

	partie := einHandlungsstand(t)

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)
	sitze[0].Commander = "held"
	partie.Sitze = sitze

	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-1/held/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneFriedhof, Karte: "held",
	})

	return partie
}
