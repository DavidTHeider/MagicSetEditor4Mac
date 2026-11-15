package app

// **Welche Frage ein Preis beantwortet** — Werkstück A3 von Phase 5.3
// (#1442).
//
// # Warum als Verbund und nicht als sieben Zeichenketten
//
// Eine Reihe hat seit A3 sieben Teile. Sieben Zeichenketten nebeneinander an
// einer Bindung verwechselt irgendwann jemand — und die Verwechslung fiele
// nicht auf, weil alle sieben Zeichenketten sind.

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

// PreisreiheDTO benennt eine Preisreihe eines Drucks.
type PreisreiheDTO struct {
	Printing string `json:"printing"`

	Quelle   string `json:"quelle"`
	Markt    string `json:"markt"`
	Art      string `json:"art"`
	Waehrung string `json:"waehrung"`
	Finish   string `json:"finish"`

	// Zustand und Sprache sind heute leer — siehe [domain.Preisstand].
	Zustand string `json:"zustand"`
	Sprache string `json:"sprache"`
}

// alsReihe macht daraus die Angabe, die der Dienst liest.
//
// **Ohne Vorgaben.** Eine Bindung, die „wenn leer, dann Scryfall und Papier"
// ergänzte, beantwortete eine Frage, die niemand gestellt hat — und genau das
// hat A3 an der alten Naht angestrichen.
func (r PreisreiheDTO) alsReihe() domain.Preisstand {
	return domain.Preisstand{
		Quelle:   domain.Preisquelle(r.Quelle),
		Markt:    domain.Preismarkt(r.Markt),
		Art:      domain.Preisart(r.Art),
		Waehrung: r.Waehrung,
		Finish:   domain.Finish(r.Finish),
		Zustand:  r.Zustand,
		Sprache:  r.Sprache,
	}
}
