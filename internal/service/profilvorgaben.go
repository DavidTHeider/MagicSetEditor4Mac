// **Vorgaben, die auf Zeilen zeigen** — Werkstück 15 von Phase 4.12
// (#1351),
// Punkte 4 bis 7.
//
// # Die eine Liste, die prüft und anbietet
//
// **`Moeglich` ist beides.** Sie sagt, was ein Profil wählen kann — und
// dieselbe Liste entscheidet, ob eine Wahl gilt. Zwei Listen liefen
// auseinander, sobald jemand ein Format löscht: die Auswahl zeigte es nicht
// mehr, die Prüfung ließe es noch durch.
//
// **Und deshalb steht hier keine Aufzählung.** Was es gibt, weiß der Bestand;
// dieser Dienst fragt ihn und urteilt danach.
//
// # Was eine Vorgabe überlebt
//
// **Das Löschen dessen, worauf sie zeigt.** Wer ein Deck löscht, das ein
// Präferenzdeck war, bekommt kein Nein — die Vorgabe bleibt stehen und zeigt
// ins Leere.
//
// **Gemeldet wird das trotzdem**: [Vorgabenstand.Verwaist] nennt die Werte,
// die es nicht mehr gibt. Sie stillschweigend wegzuräumen hieße, jemandem
// seine Wahl zu nehmen, während er nicht hinsieht (§8) — und sie zu
// verschweigen hieße, eine Vorgabe zu führen, die nichts tut.

package service

import (
	"context"
	"fmt"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Vorgabewahl ist ein Wert, den ein Profil wählen kann.
//
// **Mit Beschriftung**, weil eine Kennung allein nichts sagt: `7` ist kein
// Format, und die Fläche baut daraus keine Auswahl, die jemand bedienen kann.
type Vorgabewahl struct {
	// Wert ist, was gespeichert wird — eine Kennung oder ein Paketname.
	Wert string `json:"wert"`

	// Beschriftung ist, was dabeisteht.
	Beschriftung string `json:"beschriftung"`
}

// Vorgabenstand ist, worauf ein Profil bei einer Art zeigt — samt Angebot.
type Vorgabenstand struct {
	Art string `json:"art"`

	// Mehrwertig sagt, ob mehrere Werte gelten dürfen.
	Mehrwertig bool `json:"mehrwertig"`

	// Werte sind die gewählten, in ihrer Reihenfolge.
	Werte []string `json:"werte"`

	// Moeglich ist, was dieses Profil wählen kann.
	Moeglich []Vorgabewahl `json:"moeglich"`

	// Verwaist sind gewählte Werte, die es nicht mehr gibt.
	//
	// **Sie stehen weiter in `Werte`.** Diese Liste sagt, welche davon ins
	// Leere zeigen — sie ist ein Befund und kein zweiter Stand.
	Verwaist []string `json:"verwaist"`
}

// Vorgabenablage ist die Naht zum Bestand.
type Vorgabenablage interface {
	Vorgaben(ctx context.Context, profilID int64) (
		map[domain.Vorgabeart][]string, error)
	VorgabeSetzen(ctx context.Context, profilID int64,
		vorgabe domain.Profilvorgabe) error
}

// Vorgabenangebot sagt, was es für eine Art überhaupt gibt.
//
// **Eine Naht und keine fünf.** Was die fünf Arten unterscheidet, ist die
// Abfrage dahinter — und die gehört zu dem, der sie beantworten kann.
type Vorgabenangebot interface {
	Angebot(ctx context.Context, art domain.Vorgabeart) ([]Vorgabewahl, error)
}

// Vorgabenverwaltung liest und setzt die Vorgaben eines Profils.
type Vorgabenverwaltung struct {
	ablage  Vorgabenablage
	angebot Vorgabenangebot
}

// NeueVorgabenverwaltung baut sie.
func NeueVorgabenverwaltung(
	ablage Vorgabenablage, angebot Vorgabenangebot,
) *Vorgabenverwaltung {
	return &Vorgabenverwaltung{ablage: ablage, angebot: angebot}
}

// Stand gibt alle fünf Arten — **auch die leeren**.
//
// **Alle fünf und nicht nur die gesetzten**: die Seite zeigt sie
// nebeneinander, und wer nur die gesetzten bekäme, könnte keine neue setzen,
// ohne die Liste der Arten selbst zu kennen (§7, Regel 2).
func (v *Vorgabenverwaltung) Stand(ctx context.Context, profilID int64) (
	[]Vorgabenstand, error,
) {
	gesetzt, fehler := v.ablage.Vorgaben(ctx, profilID)
	if fehler != nil {
		return nil, fehler
	}

	staende := make([]Vorgabenstand, 0, len(domain.Vorgabearten()))

	for _, art := range domain.Vorgabearten() {
		moeglich, fehler := v.angebot.Angebot(ctx, art)
		if fehler != nil {
			return nil, fehler
		}

		werte := gesetzt[art]
		if werte == nil {
			werte = []string{}
		}

		staende = append(staende, Vorgabenstand{
			Art:        string(art),
			Mehrwertig: art.Mehrwertig(),
			Werte:      werte,
			Moeglich:   moeglich,
			Verwaist:   verwaiste(werte, moeglich),
		})
	}

	return staende, nil
}

// Setzen ersetzt die Werte einer Art — **gegen das Angebot geprüft**.
func (v *Vorgabenverwaltung) Setzen(
	ctx context.Context, profilID int64, vorgabe domain.Profilvorgabe,
) error {
	if fehler := vorgabe.Validate(); fehler != nil {
		return fehler
	}

	moeglich, fehler := v.angebot.Angebot(ctx, vorgabe.Art)
	if fehler != nil {
		return fehler
	}

	// **Geprüft wird beim Setzen und nicht beim Lesen.** Wer wählt, soll
	// erfahren, dass es das nicht gibt; wer liest, soll seinen Stand sehen —
	// auch einen, der inzwischen ins Leere zeigt.
	if fehlend := verwaiste(vorgabe.Werte, moeglich); len(fehlend) > 0 {
		return domain.NewFieldError("werte", domain.SchluesselNichtGefunden,
			"Für %q gibt es %q nicht.", string(vorgabe.Art), fehlend[0])
	}

	return v.ablage.VorgabeSetzen(ctx, profilID, vorgabe)
}

// verwaiste gibt die Werte, die im Angebot fehlen.
func verwaiste(werte []string, moeglich []Vorgabewahl) []string {
	fehlend := []string{}

	for _, wert := range werte {
		if !slices.ContainsFunc(moeglich, func(w Vorgabewahl) bool {
			return w.Wert == wert
		}) {
			fehlend = append(fehlend, wert)
		}
	}

	return fehlend
}

// Wirksam gibt den **einen** Wert einer einwertigen Art — oder leer.
//
// **Für die Stellen, die eine Vorgabe benutzen** statt sie zu verwalten: ein
// neues Deck fragt nach dem Format, ein neues Set nach Spiel und Vorlage. Sie
// wollen einen Wert und keine Liste.
//
// **Und ein verwaister Wert kommt trotzdem** — wer ihn benutzt, bekommt vom
// Bestand ein Nein mit der richtigen Meldung. Ihn hier zu verschlucken machte
// aus einer falschen Vorgabe eine unsichtbare.
func (v *Vorgabenverwaltung) Wirksam(
	ctx context.Context, profilID int64, art domain.Vorgabeart,
) (string, error) {
	if art.Mehrwertig() {
		return "", fmt.Errorf(
			"%w: %q hält mehrere Werte — hier ist einer verlangt",
			domain.ErrValidation, string(art))
	}

	gesetzt, fehler := v.ablage.Vorgaben(ctx, profilID)
	if fehler != nil {
		return "", fehler
	}

	if werte := gesetzt[art]; len(werte) > 0 {
		return werte[0], nil
	}

	return "", nil
}
