package service

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Kartenanleger legt eine leere Karte in einem Set an.
//
// **Die Schnittstelle steht hier, wo sie gebraucht wird** (§2) — dieselbe
// Gestalt wie in [Lebenszyklus], damit dieselbe Naht beides bedienen kann.
type Kartenanleger interface {
	KarteAnlegen(ctx context.Context, setID int64) (int64, error)
}

// Scryfallanlage macht aus einem Kartennamen eine Karte.
//
// # Warum das ein eigener Anwendungsfall ist
//
// Bis #1032
// arbeitete der Abgleich nur an einer **vorhandenen** Karte: erst eine leere
// anlegen, dann abgleichen. Zwei Schritte für einen Vorgang — und den ersten
// muss man kennen. Eine Kapselung, die den naheliegendsten Vorgang nicht
// anbietet, ist gekapselt und nicht benutzbar; dieselbe Klasse wie
// #852.
//
// # Warum sie neben dem Abgleich steht und nicht in ihm
//
// **Weil sie mehr braucht.** Der Abgleich kennt den fremden Dienst und den
// Schreibweg; hier kommt der Lebenszyklus dazu. Ein Abgleich mit einem
// optionalen Anleger hätte eine Naht, die bei jedem Aufrufer fehlen kann, der
// sie nicht setzt — und dann legte er still nichts an.
type Scryfallanlage struct {
	abgleich *Scryfallabgleich
	anleger  Kartenanleger
}

// NeueScryfallanlage baut sie.
func NeueScryfallanlage(abgleich *Scryfallabgleich, anleger Kartenanleger) *Scryfallanlage {
	return &Scryfallanlage{abgleich: abgleich, anleger: anleger}
}

// Anlegen legt eine Karte an und füllt sie aus Scryfall.
//
// # Die Reihenfolge ist die Entscheidung
//
// **Zuerst holen, dann anlegen, dann schreiben.** Ein falscher Kartenname
// scheitert damit, *bevor* etwas im Bestand steht. Legte man zuerst an, bliebe
// bei jedem Tippfehler eine leere Karte zurück, und niemand räumt sie weg.
//
// **Die Kennung kommt auch im Fehlerfall zurück** (#155): scheitert das Füllen
// nach dem Anlegen, ist die Karte da — §7 Regel 3, „geschrieben ist
// geschrieben". Wer sie nicht **nennt**, lässt sie im Bestand zurück, ohne
// dass jemand davon weiß.
//
// # Was sie nicht tut
//
// **Kein Bild.** Das ist ein eigener Weg mit einer eigenen Bedingung — der
// Künstlername muss neben dem Ausschnitt stehen —, und er entscheidet
// zusätzlich über den Zuschnitt. Die Fassade verkettet beides, wenn eine
// Bildart genannt ist.
func (s *Scryfallanlage) Anlegen(
	ctx context.Context, setID int64, spiel, name string,
) (ergebnis Anlageergebnis, fehler error) {
	// **Ohne jetzige Werte**: die Karte gibt es noch nicht, also ist jedes
	// abgebildete Feld eine Änderung.
	vorschlag, holfehler := s.abgleich.Vorschlag(ctx, spiel, map[string]string{}, name)
	if holfehler != nil {
		return Anlageergebnis{}, holfehler
	}

	karteID, anlegefehler := s.anleger.KarteAnlegen(ctx, setID)
	if anlegefehler != nil {
		return Anlageergebnis{}, anlegefehler
	}

	ergebnis = Anlageergebnis{
		KarteID: karteID, Herkunft: vorschlag.Herkunft,
		Alter: vorschlag.Alter, Befunde: vorschlag.Befunde,
	}

	felder := make([]string, 0, len(vorschlag.Aenderungen))
	for _, aenderung := range vorschlag.Aenderungen {
		felder = append(felder, aenderung.Feld)
	}

	uebernommen, schreibfehler := s.abgleich.Uebernehmen(
		ctx, setID, karteID, vorschlag, felder)

	ergebnis.Uebernommen = uebernommen

	return ergebnis, schreibfehler
}

// Anlageergebnis ist, was beim Anlegen herauskommt.
//
// **Ein Verbund und nicht vier Rückgabewerte**: die Kennung allein sagt nicht,
// was geschrieben wurde, und die Befunde gehören zum Ergebnis und nicht zum
// Fehler (§8) — eine Karte, deren Seltenheit Scryfall in einem unbekannten
// Wort nennt, ist brauchbar, und die Auffälligkeit gehört daneben.
type Anlageergebnis struct {
	// KarteID ist die angelegte Karte — **auch wenn das Füllen scheitert**.
	KarteID int64

	// Uebernommen sind die geschriebenen Felder samt ihrem vorigen Wert.
	Uebernommen []Feldvorschlag

	// Befunde sind die Auffälligkeiten der Antwort.
	Befunde []domain.Befund

	// Herkunft ist die Kartenseite bei Scryfall.
	Herkunft string

	// Alter sagt, wie alt die Antwort war (§16, Regel 5).
	Alter time.Duration
}
