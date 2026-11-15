package domain

// **Ein Regelwerk als versioniertes Datenprofil** — Werkstück G4 von Phase
// 5.5 (#1461).
//
// # Warum das Profil in die Partie eingefroren wird
//
// **Weil dasselbe Profil nach einem Neustart dieselben Entscheidungen geben
// muss** — auch dann, wenn draußen inzwischen eine neue Fassung gilt. Das ist
// der zweite Abnahmepunkt des Issues, wörtlich: „Gleicher Ruleset-/Format-
// Snapshot liefert nach Neustart dieselben Entscheidungen, auch wenn externe
// Legality-Daten inzwischen aktualisiert wurden."
//
// **Dieselbe Überlegung wie beim Deck** (G1): der Abzug wird eingefroren,
// damit eine spätere Änderung die laufende Partie nicht anfasst. Hier ist es
// das Regelwerk.
//
// # Und deshalb ist es Daten und kein `switch`
//
// **Dieselbe Entscheidung wie bei [Format]** ([0061](../../docs/entscheidungen/0061-was-ein-deck-hier-ist.md):
// „Daten, nicht Code"): wer ein eigenes Regelwerk spielt, soll es anlegen
// können, ohne dass jemand das Programm ändert.

import (
	"fmt"
	"slices"
	"strings"
)

// Spielerzahl ist, wie viele an einen Tisch passen.
type Spielerzahl struct {
	Mindestens int
	Hoechstens int
}

// Regelprofil ist ein Regelwerk in einer Fassung.
type Regelprofil struct {
	// Kennung nennt das Regelpaket — `sandkasten`, `commander`.
	Kennung string

	// Fassung steigt, wenn sich am Profil etwas ändert.
	Fassung int

	// Spiel ist `magic`, `space` — ein Regelwerk gehört zu einem Spiel.
	Spiel string

	// Wirksam ist das Datum, ab dem diese Fassung gilt — `2026-08-19`.
	//
	// **Das Issue verlangt es**: „Regelquelle/Fassung/Wirksamkeitsdatum".
	// Eine Fassungsnummer sagt, **welche** es ist; das Datum sagt, **ab
	// wann** — und erst beides zusammen erklärt, warum zwei Partien
	// verschieden entschieden haben.
	Wirksam string

	// Quelle belegt das Profil als Ganzes.
	Quelle Regelquelle

	Spieler Spielerzahl

	// Startleben ist die Zahl, mit der jeder anfängt; 0 heißt: das Spiel
	// kennt keine.
	Startleben int

	// Matrix sagt, was dieses Programm davon wirklich durchsetzt.
	Matrix Regelmatrix

	// Zusatzzonen sind Zonen über die gewöhnlichen hinaus.
	Zusatzzonen []Zone

	// ErlaubteHausregeln sind die Abweichungen, die eine Runde vereinbaren
	// darf. **Leer heißt: keine.**
	ErlaubteHausregeln []Hausregelart
}

// Pruefe sagt, ob dieses Profil eine Partie tragen kann.
func (p Regelprofil) Pruefe() error {
	if strings.TrimSpace(p.Kennung) == "" {
		return fmt.Errorf(
			"%w: ein Regelprofil ohne Kennung lässt sich nicht einfrieren",
			ErrValidation)
	}

	if p.Fassung < 1 {
		return fmt.Errorf(
			"%w: das Regelprofil %q hat die Fassung %d — eine Partie friert "+
				"eine Fassung ein, und die Null ist keine",
			ErrValidation, p.Kennung, p.Fassung)
	}

	if strings.TrimSpace(p.Wirksam) == "" {
		return fmt.Errorf(
			"%w: das Regelprofil %q nennt kein Wirksamkeitsdatum",
			ErrValidation, p.Kennung)
	}

	if fehler := p.Quelle.Pruefe(); fehler != nil {
		return fmt.Errorf("das Regelprofil %q: %w", p.Kennung, fehler)
	}

	if fehler := p.pruefeSpieler(); fehler != nil {
		return fehler
	}

	if fehler := p.Matrix.Pruefe(); fehler != nil {
		return fmt.Errorf("das Regelprofil %q: %w", p.Kennung, fehler)
	}

	for _, art := range p.ErlaubteHausregeln {
		if !IstHausregelart(art) {
			return fmt.Errorf("%w: %q ist keine Hausregel", ErrValidation, art)
		}
	}

	for _, zone := range p.Zusatzzonen {
		if !IstZone(zone) {
			return fmt.Errorf("%w: %q ist keine Zone", ErrValidation, zone)
		}
	}

	return nil
}

// pruefeSpieler hält die Spielerzahl beisammen.
func (p Regelprofil) pruefeSpieler() error {
	if p.Spieler.Mindestens < 1 {
		return fmt.Errorf(
			"%w: das Regelprofil %q lässt %d Spieler zu",
			ErrValidation, p.Kennung, p.Spieler.Mindestens)
	}

	if p.Spieler.Hoechstens < p.Spieler.Mindestens {
		return fmt.Errorf(
			"%w: das Regelprofil %q lässt höchstens %d und mindestens %d "+
				"Spieler zu", ErrValidation, p.Kennung,
			p.Spieler.Hoechstens, p.Spieler.Mindestens)
	}

	return nil
}

// ErlaubtHausregel sagt, ob eine Abweichung vereinbart werden darf.
//
// **Und wo nicht, sagt sie warum** — als [Regelkonflikt] zwischen der
// Hausregel und dem Profil. „Das geht nicht" nennt keine Stufe.
func (p Regelprofil) ErlaubtHausregel(regel Hausregel) error {
	if fehler := regel.Pruefe(); fehler != nil {
		return fehler
	}

	if slices.Contains(p.ErlaubteHausregeln, regel.Art) {
		return nil
	}

	return Regelkonflikt{
		Staerker:   RangBasis,
		Schwaecher: RangHausregel,
		Gegenstand: string(regel.Art),
	}.Fehler()
}

// TraegtSpieler sagt, ob so viele an diesen Tisch passen.
func (p Regelprofil) TraegtSpieler(anzahl int) bool {
	return anzahl >= p.Spieler.Mindestens && anzahl <= p.Spieler.Hoechstens
}
