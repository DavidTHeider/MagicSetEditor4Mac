package service

// **Mana kommt in den Vorrat** — Werkstück M3 der Regelmaschine
// (#1605).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **`zugbefehle.go` beantwortet, was zum Zug gehört**; hier steht, **wie Mana
// entsteht**. Und `zugbefehle.go` liegt bei 340 Zeilen (§2).

import (
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// manabefehl trägt Mana nach — CR 106.4.
func (s *Sandkasten) manabefehl(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	roh, steht := befehl.Nutzlast["art"].(string)
	if !steht {
		return nil, domain.Partie{}, fmt.Errorf(
			"%w: %q braucht die Manaart in \"art\"",
			domain.ErrValidation, befehl.Art)
	}

	art, bekannt := domain.ManaartAus(roh)
	if !bekannt {
		return nil, domain.Partie{}, fmt.Errorf(
			"%w: %q ist keine der sechs Manaarten (CR 106.1b)",
			domain.ErrValidation, roh)
	}

	um, fehler := ganzeZahl(befehl, "um")
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	danach, fehler := partie.ManaHinzu(befehl.Absender, art, um)

	return manaereignis(danach, fehler, befehl)
}

// landmanabefehl tappt ein Grundland — CR 305.6, CR 605.3b.
func (s *Sandkasten) landmanabefehl(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	objekt, steht := befehl.Nutzlast["objekt"].(string)
	if !steht {
		return nil, domain.Partie{}, fmt.Errorf(
			"%w: %q braucht die Objektkennung in \"objekt\"",
			domain.ErrValidation, befehl.Art)
	}

	danach, fehler := partie.LandFuerMana(befehl.Absender, objekt)

	return manaereignis(danach, fehler, befehl)
}

// manaereignis macht aus dem Ausgang ein Ereignis.
//
// # Warum der Vorrat öffentlich ist
//
// **Weil CR 106.4a und CR 106.4b es verlangen**: wer Mana übrig hat, sagt
// beim Bezahlen und beim Passen an, was noch dasteht. Ein Vorrat, den nur
// sein Besitzer sieht, machte diese Ansage unmöglich — und sie ist eine
// Regel und keine Höflichkeit.
func manaereignis(
	danach domain.Partie, fehler error, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	var vorrat domain.Manavorrat

	for _, sitz := range danach.Sitze {
		if sitz.Kennung == befehl.Absender {
			vorrat = sitz.Manavorrat
		}
	}

	return []domain.Ereignis{{
		Art: befehl.Art,
		Oeffentlich: map[string]any{
			"sitz": befehl.Absender, "vorrat": Vorratstext(vorrat),
		},
	}}, danach, nil
}

// Vorratstext gibt einen Vorrat als Buchstabenfolge — `RRG`, leer heißt leer.
//
// # Warum eine Zeichenkette und keine Tabelle
//
// **Weil MCP, REST und die Kommandozeile Text sind.** `RRG` ist die Auskunft,
// die jemand liest; eine Tabelle aus sechs Zahlen wäre dieselbe Aussage in
// einer Form, die ein Mensch erst zusammensetzen muss. Derselbe Grund wie bei
// [app.EigenschaftenDTO.Farbtext] (M2).
//
// **In der Reihenfolge des Regelwerks** (CR 106.1b), farblos zuletzt — nicht
// in der, in der die Tabelle ihre Schlüssel hergibt: eine Tabelle in Go hat
// keine Reihenfolge, und zwei Läufe nacheinander gäben zwei Texte.
func Vorratstext(vorrat domain.Manavorrat) string {
	var text string

	for _, art := range domain.Manaarten() {
		for range vorrat.Zahl(art) {
			text += string(art)
		}
	}

	return text
}
