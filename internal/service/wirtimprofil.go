package service

// **Der Wirt sitzt im gezeigten Profil, wenn niemand anderes gewählt ist** —
// UXR-F21 (#2041),
// Vermerk [0164](../../docs/entscheidungen/0164-der-wirt-sitzt-im-gezeigten-profil.md).
//
// # Warum es eine Vorgabe braucht
//
// **Weil „dran" am eigenen Sitz hängt** ([eigenerSitz]), und einen eigenen
// Sitz hatte nur, wer ihn gebunden hat. Binden lässt sich am Fenster nur mit
// zwei Profilen; mit einem — dem Normalfall — erschien die Marke nie.
//
// # Warum sie keine Sicht ändert
//
// **Der Tisch zeigt einem ungebundenen Profil ohnehin den Wirt**
// ([domain.Partie.SichtFuerProfil]). Wer den Wirt an das gezeigte Profil
// bindet, gibt diesem Profil denselben Sitz, den es schon sah — und einem
// zweiten Profil weiter den Wirt. Neu ist nur, dass die Liste ihn als den
// eigenen kennt.
//
// # Wann sie nichts tut
//
//   - **Sobald ein Sitz gebunden ist.** Wer wählt, hat gewählt; eine Vorgabe
//     daneben wäre eine zweite Antwort.
//   - **Ohne gezeigtes Profil** (null): dann gibt es niemanden, der dort säße.
//   - **Wenn der Wirt nicht am Tisch sitzt**: dann weist der Aufbau ihn ab,
//     und eine Bindung an einen Sitz, den es nicht gibt, gäbe es nicht.

// WirtImGezeigtenProfil gibt die Wünsche mit der Vorgabe — als Abschrift.
func WirtImGezeigtenProfil(wuensche []Sitzwunsch, wirt string, gezeigt int64) []Sitzwunsch {
	gebunden := append([]Sitzwunsch(nil), wuensche...)

	if gezeigt == 0 {
		return gebunden
	}

	for _, wunsch := range gebunden {
		if wunsch.ProfilID != 0 {
			return gebunden
		}
	}

	for stelle := range gebunden {
		if gebunden[stelle].Kennung == wirt {
			gebunden[stelle].ProfilID = gezeigt
		}
	}

	return gebunden
}
