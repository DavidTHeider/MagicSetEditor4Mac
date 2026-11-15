package domain

// **Was einen Sitz zum Spieler macht** — Werkstück G9 von Phase 5.5
// (#1592).
//
// # Warum diese drei Zahlen und nicht mehr
//
// **Weil G1 bis G8 einen Tisch gebaut haben und keinen Spieler.** Ein Sitz
// trug Kennung, Name und Deck — kein Leben, keine Marken, kein Schaden. Wer
// in diesem Programm Magic spielte, führte sie auf Papier daneben.
//
// # Und warum die Handgrenze **keine** davon ist
//
// **Weil niemand sie ändern kann.** CR 402.2 sagt „normally seven" und lässt
// Effekten Raum; Effekte gibt es hier nicht (sie sind Phase 5.8). Ein Feld am
// Sitz, das nur eine Zahl trägt und das niemand verstellt, wäre ein
// Versprechen, das keiner einlöst (§10) — die Zahl steht deshalb hier, und
// wenn es Effekte gibt, zieht sie an den Sitz.

import "fmt"

const (
	// Handgrenze ist die Höchstzahl auf der Hand — CR 402.2 (Fassung
	// 2026-08-19).
	//
	// **Sie greift im Aufräumschritt** (CR 514.1) und sonst nie: dazwischen
	// darf jeder beliebig viele Karten halten.
	Handgrenze = 7

	// Starthand ist die Zahl, die jeder zu Beginn zieht — CR 103.5 (Fassung
	// 2026-08-19).
	//
	// **Dieselbe Zahl und trotzdem eine eigene.** CR 103.5 spricht von der
	// „starting hand size", CR 402.2 von der „maximum hand size"; ein Effekt
	// kann die eine ändern, ohne die andere anzufassen. Eine Konstante für
	// beides sagte, es sei dieselbe Regel.
	Starthand = 7
)

// MitStartleben gibt die Partie mit dem Startleben aus ihrem Regelprofil.
//
// # Warum aus dem Profil und nicht aus einer Zahl hier
//
// **Weil es eingefroren mitreist** (G4,
// #1461):
// dieselbe Partie gibt nach einem Neustart dieselben Antworten, auch wenn
// draußen ein neues Profil gilt. CR 119.1 und CR 103.4 (Fassung 2026-08-19)
// nennen zwanzig als das Gewöhnliche — und dass Varianten davon abweichen
// dürfen, ist genau der Grund, warum die Zahl im Profil steht und nicht hier.
//
// **Und eine Null bleibt eine Null.** Ein Regelwerk ohne Lebenspunkte — ein
// anderes Spiel, das dieses Programm trägt — bekommt keine erfunden
// ([Regelprofil.Startleben]: „0 heißt: das Spiel kennt keine").
//
// **Sie setzt und addiert nicht**: der Aufbau ruft sie einmal, und ein
// zweiter Ruf gäbe denselben Stand. Was danach geschieht, sind Befehle.
func (p Partie) MitStartleben() Partie {
	sitze := make([]Sitz, len(p.Sitze))
	copy(sitze, p.Sitze)

	for stelle := range sitze {
		sitze[stelle].Leben = p.Profil.Startleben
	}

	p.Sitze = sitze

	return p
}

// pruefeMarken weist negative Marken und negativen Schaden ab.
//
// # Warum Leben davon ausgenommen ist
//
// **Weil ein Spieler unter null ein Spieler ist, der verloren hat** —
// CR 704.5a, und das ist eine zustandsbasierte Aktion. Die gibt es hier nicht
// (Phase 5.8); bis dahin steht die Zahl da, wie sie ist, und wer sie auf −3
// stellt, hat etwas gesagt und keinen Fehler gemacht.
//
// **Marken sind etwas anderes**: CR 122.1 (Fassung 2026-08-19) kennt keine
// negative Anzahl — eine Marke liegt da oder nicht.
//
// **Und der markierte Schaden steht bei [Spielobjekt.Pruefe]** und nicht
// hier: der Sandkasten prüft nach jedem Befehl das Objekt, und eine zweite
// Prüfung an dieser Stelle wäre dieselbe Aussage an zwei Orten (§2).
func (p Partie) pruefeMarken() error {
	for _, sitz := range p.Sitze {
		if sitz.Vergiftung < 0 {
			return fmt.Errorf(
				"%w: der Sitz %q hat %d Vergiftungsmarken — weniger als keine "+
					"gibt es nicht", ErrValidation, sitz.Kennung, sitz.Vergiftung)
		}
	}

	return nil
}
