package service

// **Kleine Regeln der Turnierleitung** — ausgelagert aus `turnierleitung.go`,
// als DOM-04.AK-04 (#1699) die Abschlussregel dazubrachte und die Datei über
// die Zeilengrenze aus §2 ging.

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

// gehoertDazu sagt, ob eine Paarung zu diesem Turnier gehört.
func gehoertDazu(turnier domain.Turnier, paarungID int64) bool {
	for _, runde := range turnier.Runden {
		for _, paarung := range runde.Paarungen {
			if paarung.ID == paarungID {
				return true
			}
		}
	}

	return false
}
