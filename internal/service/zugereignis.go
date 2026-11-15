package service

// **Das Ereignis eines Zugbefehls** — Werkstück `G9` von Phase 5.5
// (#1592).
//
// **Herausgezogen aus `zugbefehle.go`**, weil die Datei an der Zeilengrenze
// stand (§2). Der Schnitt folgt der Frage: dort steht, **was** ein Befehl
// tut, hier, **was der Strom davon erfährt**.

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

// ereignisAus macht aus einem Ausgang das Ereignis, das ihn beschreibt.
func ereignisAus(
	danach domain.Partie, fehler error, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	oeffentlich := map[string]any{
		"sitz":       befehl.Absender,
		"zug":        float64(danach.Zug.Nummer),
		"aktiv":      danach.Zug.Aktiv,
		"phase":      string(danach.Zug.Abschnitt.Phase),
		"schritt":    string(danach.Zug.Abschnitt.Schritt),
		"prioritaet": danach.Zug.Prioritaet,
	}

	// **Was ein Kartenskript beim Verrechnen zu melden hatte, steht hier**
	// (#1625).
	//
	// **An diesem Ereignis und nicht an dem des Ankündigens**: verrechnet
	// wird beim Passen (CR 117.3b), und ein Befund gehört an den
	// Augenblick, in dem er entstanden ist. Ein Feld, das immer dastünde,
	// wäre bei jedem Passen eine leere Auskunft — deshalb nur, wenn es eine
	// gibt.
	if befund := danach.Skriptbefund(); befund != "" {
		oeffentlich["befund"] = befund
	}

	return []domain.Ereignis{{
		Art: befehl.Art, Oeffentlich: oeffentlich,
	}}, danach, nil
}
