package app

// **Eine Fläche ohne Fenster richtet keine Verbindung** — gefunden vom
// Flächenlauf zu #1819:
// `mse partie-anlegen` stürzte über einen Nullzeiger ab, sobald vorher ein
// Profil gewechselt worden war.
//
// # Warum paketintern
//
// **Weil die Zusage „keine zweite Verbindung" nur von innen zu sehen ist.**
// Von außen bliebe allein „stürzt nicht ab" — und eine Fläche, die ihre
// Verbindung still gegen eine leere tauschte, wäre dann grün.

import "testing"

func TestEineFlaecheOhneFensterRichtetNachEinemWechselNichts(t *testing.T) {
	db := bestandFuerDenRiegel(t)

	// **Die Revision der Rechte bewegt sich** — wie nach `mse profil <id>`:
	// jeder Wechsel des Zeigers zählt sie hoch (Wanderung 0047). In einem
	// frischen Bestand steht sie auf null, und genau deshalb fiel der Absturz
	// keinem Lauf auf, der nicht vorher gewechselt hatte.
	if _, fehler := db.ExecContext(riegelkontext(t),
		`UPDATE freigabenrevision SET revision = revision + 1 WHERE id = 1`); fehler != nil {
		t.Fatalf("die Revision bewegen: %v", fehler)
	}

	for name, anwendung := range map[string]*App{
		"die Partiefläche": PartieflaecheUeber(db).app,
		"die Regelfläche":  RegelflaecheUeber(db),
		"die Netzfläche":   NetzflaecheUeber(db),
		"die Peerfläche":   PeerflaecheUeber(db),
	} {
		t.Run(name, func(t *testing.T) {
			anwendung.imProfil(riegelkontext(t))

			if anwendung.nurLesen != nil {
				t.Error("eine Fläche ohne Fenster öffnete eine zweite Verbindung")
			}
		})
	}
}
