package domain

// **Ob ein Objekt so dastehen darf** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, *was* ein Spielobjekt ist** — die Zonen, die Arten und
// jedes Feld mit seinem Grund. Hier steht, **was daran unmöglich ist**: ein
// Spielstein mit einer Karte, ein Zähler unter null, ein Anhängsel an sich
// selbst. `spielobjekt.go` stand mit beidem bei 401 Zeilen (§2) — die
// Wirkungen aus `M8`
// (#1610)
// waren die Zeile, die darüber hinausging.

import (
	"fmt"
	"strings"
)

// Pruefe sagt, ob dieses Objekt so auf dem Tisch liegen darf.
func (o Spielobjekt) Pruefe() error {
	if strings.TrimSpace(o.Kennung) == "" {
		return fmt.Errorf(
			"%w: ein Spielobjekt ohne Kennung lässt sich nicht ansprechen — "+
				"und zwei gleiche Karten wären dann dasselbe Ding",
			ErrValidation)
	}

	if fehler := o.pruefeArtUndZone(); fehler != nil {
		return fehler
	}

	// **Weniger als kein Schaden gibt es nicht** — CR 120.3 (Fassung
	// 2026-08-19), Werkstück G9 von Phase 5.5 (#1592).
	//
	// **Hier und nicht nur in [Partie.Pruefe]**: der Sandkasten prüft nach
	// jedem Befehl das **Objekt**, und wer den Schaden dort abzöge, käme an
	// einer Prüfung der ganzen Partie vorbei.
	if o.Schaden < 0 {
		return fmt.Errorf(
			"%w: auf %q sind %d Schaden markiert — weniger als keiner gibt "+
				"es nicht", ErrValidation, o.Kennung, o.Schaden)
	}

	if strings.TrimSpace(o.Beherrscher) == "" {
		return fmt.Errorf(
			"%w: das Spielobjekt %q beherrscht niemand", ErrValidation,
			o.Kennung)
	}

	if o.HaengtAn == o.Kennung {
		return fmt.Errorf(
			"%w: das Spielobjekt %q hängt an sich selbst", ErrValidation,
			o.Kennung)
	}

	return o.pruefeZaehler()
}

// pruefeArtUndZone hält beides gegen das, was darin stehen muss.
func (o Spielobjekt) pruefeArtUndZone() error {
	bekannt := false

	for _, art := range Objektarten() {
		if art == o.Art {
			bekannt = true
		}
	}

	if !bekannt {
		return fmt.Errorf(
			"%w: %q ist keine Objektart", ErrValidation, o.Art)
	}

	if o.Art == ObjektKarte && strings.TrimSpace(o.Karte) == "" {
		return fmt.Errorf(
			"%w: das Spielobjekt %q ist eine Karte und nennt keine — dann "+
				"weiß niemand, was dort liegt", ErrValidation, o.Kennung)
	}

	if o.Art != ObjektKarte && strings.TrimSpace(o.Name) == "" {
		return fmt.Errorf(
			"%w: das Spielobjekt %q hat weder Karte noch Namen",
			ErrValidation, o.Kennung)
	}

	for _, zone := range Zonen() {
		if zone == o.Zone {
			return nil
		}
	}

	return fmt.Errorf("%w: %q ist keine Zone", ErrValidation, o.Zone)
}

// pruefeZaehler weist ab, was sich nicht zählen lässt.
func (o Spielobjekt) pruefeZaehler() error {
	for art, menge := range o.Zaehler {
		if strings.TrimSpace(art) == "" {
			return fmt.Errorf(
				"%w: das Spielobjekt %q trägt einen namenlosen Zähler",
				ErrValidation, o.Kennung)
		}

		if menge < 0 {
			return fmt.Errorf(
				"%w: das Spielobjekt %q trägt %d Zähler der Art %q — "+
					"negative gibt es nicht", ErrValidation, o.Kennung,
				menge, art)
		}
	}

	return nil
}
