package service

// **Der Beginn einer Partie als Befehle** — Mischen, Starthand und Mulligan,
// CR 103.3 und CR 103.5 (Fassung 2026-09-25, auch für jede weitere Nummer
// hier), #1839.
//
// # Warum vier Befehle und kein Knopf „Partie beginnen"
//
// **Weil jeder Sitz für sich entscheidet**: ob er behält, wie oft er neu
// zieht, welche Karten er unter die Bibliothek legt. Und weil ein Befehl an
// allen vier Flächen zugleich dasteht (§16) — ein Knopf nur im Fenster.
//
// # Warum ohne Priorität
//
// **Weil der Beginn vor dem ersten Zug liegt** — CR 103.8 steht nach
// CR 103.5. Niemand hat dabei Priorität, und ein Befehl, der sie verlangte,
// wäre für jeden außer dem Startspieler nicht zu stellen.

import (
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

const (
	// BefehlStarthand mischt die Bibliothek und zieht die Starthand —
	// CR 103.3, dann CR 103.5.
	BefehlStarthand = "spieler.starthand"

	// BefehlMulligan mischt die Hand zurück und zieht eine neue — CR 103.5.
	BefehlMulligan = "spieler.mulligan"

	// BefehlUnterlegen legt die geschuldeten Karten unter die Bibliothek —
	// CR 103.5. **Welche, sagt die Nutzlast** (`objekte`), in der Reihenfolge,
	// in der sie unten liegen sollen.
	BefehlUnterlegen = "spieler.unterlegen"

	// BefehlBehalten beendet die Mulligans dieses Sitzes — CR 103.5.
	BefehlBehalten = "spieler.behalten"
)

// Eroeffnungsbefehle gibt die vier.
func Eroeffnungsbefehle() []string {
	return []string{
		BefehlStarthand, BefehlMulligan, BefehlUnterlegen, BefehlBehalten,
	}
}

// waehrendDesBeginns gibt, was vor dem ersten Zug gestellt werden darf —
// Befund 2 des Zweitblicks über Phase 6.9.
//
// **Die vier des Beginns, und drei, die ihn nicht berühren**: aufgeben geht
// jederzeit (CR 104.3a), und Münze und Würfel sind die Mittel, mit denen
// CR 103.1 den Startspieler bestimmen lässt. Alles andere ist eine Handlung
// im Spiel, und das Spiel fängt erst mit dem ersten Zug an (CR 103.8).
func waehrendDesBeginns() []string {
	return append(Eroeffnungsbefehle(),
		BefehlAufgeben, BefehlMuenzwurf, BefehlWuerfeln)
}

// beginnHaelt sagt, ob der Beginn gerade alles andere zurückhält.
//
// **Nur, wo die Zugstruktur gilt.** Wer sie für seine Runde abwählt, hat
// einen Sandkasten, in dem jeder jederzeit alles darf (§15, Punkt 4) — und
// eine Partie von vorher trägt keinen Stand des Beginns
// ([domain.Partie.BeginnLaeuft]).
func beginnHaelt(partie domain.Partie) bool {
	return partie.BeginnLaeuft() &&
		partie.Profil.Matrix.Grad(domain.BereichZugstruktur) ==
			domain.UnterstuetzungAutomatisch
}

// istEroeffnungsbefehl sagt, ob dieser Befehl zum Beginn gehört.
func istEroeffnungsbefehl(art string) bool {
	return slices.Contains(Eroeffnungsbefehle(), art)
}

// eroeffnungsschritt sagt, ob der Kern diesen Schritt gerade annähme.
//
// **Die Bedingungen stehen im Kern** ([domain.Partie.Eroeffnungsschritte]) —
// hier steht nur, welcher Befehl welchen Schritt meint.
func eroeffnungsschritt(schritte domain.Eroeffnungsschritte, art string) bool {
	switch art {
	case BefehlStarthand:
		return schritte.Starthand
	case BefehlMulligan:
		return schritte.Mulligan
	case BefehlUnterlegen:
		return schritte.Unterlegen
	}

	return schritte.Behalten
}

// eroeffnungsbefehl legt die vier aus.
//
// **Die Mischung kommt aus der eingesetzten Quelle**, wie bei `zone.mischen`
// — und zwar über die ganze Bibliothek, beim Mulligan samt der Hand, die in
// sie zurückgeht. Geprüft wird sie im Kern.
func (s *Sandkasten) eroeffnungsbefehl(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	wer := befehl.Absender

	var (
		danach domain.Partie
		fehler error
	)

	switch befehl.Art {
	case BefehlStarthand, BefehlMulligan:
		anzahl := len(objekteIn(partie, wer, domain.ZoneBibliothek))
		if befehl.Art == BefehlMulligan {
			anzahl += len(objekteIn(partie, wer, domain.ZoneHand))
		}

		mischung, geklagt := s.zufall.Mische(anzahl)
		if geklagt != nil {
			return nil, domain.Partie{}, geklagt
		}

		if befehl.Art == BefehlStarthand {
			danach, fehler = partie.ZieheStarthand(wer, mischung)
		} else {
			danach, fehler = partie.NimmMulligan(wer, mischung)
		}

	case BefehlUnterlegen:
		objekte, gelesen := kennungsliste(befehl, "objekte")
		if gelesen != nil {
			return nil, domain.Partie{}, gelesen
		}

		danach, fehler = partie.Unterlege(wer, objekte)

	default:
		danach, fehler = partie.Behalte(wer)
	}

	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	sitz := danach.SitzVon(wer)

	// **Wie viele, nicht welche** — die Hand ist verborgen (CR 402.3), und
	// ein Ereignis mit Kennungen verriete sie allen.
	return []domain.Ereignis{{
		Art: befehl.Art,
		Oeffentlich: map[string]any{
			"sitz":       wer,
			"stand":      string(sitz.Eroeffnung),
			"hand":       float64(len(objekteIn(danach, wer, domain.ZoneHand))),
			"mulligans":  float64(sitz.Mulligans),
			"unterlegen": float64(sitz.Unterlegen),
		},
	}}, danach, nil
}
