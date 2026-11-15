// Ein ganzer Satz von Scryfall
// (#1033).
//
// **Eine eigene Datei, weil es ein eigener Anwendungsfall ist**: `import.go`
// liest Dateien, `mtgjson.go` ein fremdes Dateiformat — hier kommt ein Set aus
// dem Netz.

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// Fremdsaetze ist, was der Satzimport von einem fremden Dienst braucht.
//
// **Die Schnittstelle steht hier, wo sie gebraucht wird** (§2): `service`
// bindet die Fähigkeit, nicht den Klienten — und ein Test braucht kein Netz.
type Fremdsaetze interface {
	// Setkarten gibt alle Karten eines Satzes, über alle Seiten.
	Setkarten(ctx context.Context, code string) (scryfall.Satz, []domain.Befund, error)

	// Setangaben gibt, was Scryfall über den Satz selbst sagt — vor allem
	// seinen **Namen**. Ohne ihn hieße das Set wie sein Kürzel.
	Setangaben(ctx context.Context, code string) (
		*scryfall.Setangabe, time.Duration, []domain.Befund, error)
}

// Satzherkunft ist die Herkunft, unter der ein eingelesener Satz im Bestand
// steht.
//
// **Kein Dateipfad, und das ist der Punkt**: die Regel „eine Quelle, ein Set"
// (#886) vergleicht Herkünfte. Ein Satz von Scryfall hat keine Datei — er
// braucht trotzdem eine Kennung, unter der ein zweiter Aufruf ihn wiederfindet.
func Satzherkunft(code string) string { return "scryfall:" + code }

// Satzimport liest einen ganzen Satz von Scryfall in den Bestand.
//
// # Warum das Spiel von außen kommt
//
// **Weil Scryfall es nicht sagt.** Es beschreibt Magic-Karten, nicht
// MSE-Kartenfelder; zwischen beidem liegt eine Zuordnung, und die gehört zum
// **Spiel**. Ein Spiel ohne Zuordnung wird abgewiesen und nicht geraten —
// dieselbe Regel wie beim MTG-JSON-Import.
//
// # Was er **nicht** tut
//
// **Er holt keine Bilder.** Ein Satz mit 300 Karten wären 300 Abrufe und 300
// Nennungen; der Bildweg je Karte steht daneben und wird einzeln gerufen. Das
// gehört gesagt, nicht angenommen.
type Satzimport struct {
	fremd    Fremdsaetze
	speicher SetSpeicher
	koepfe   Kopfliste
}

// NeuerSatzimport baut ihn.
func NeuerSatzimport(
	fremd Fremdsaetze, speicher SetSpeicher, koepfe Kopfliste,
) *Satzimport {
	return &Satzimport{fremd: fremd, speicher: speicher, koepfe: koepfe}
}

// Einlesen holt den Satz und legt ihn als Set ab.
//
// # Die Reihenfolge
//
// **Erst nachsehen, dann holen, dann anlegen.** Ein Satz, der schon im Bestand
// liegt, kostet damit keinen Netzgang; ein Satzcode, den es nicht gibt,
// scheitert, bevor etwas angelegt ist.
func (s *Satzimport) Einlesen(ctx context.Context, code, spiel string) (*Ergebnis, error) {
	// **Ein leerer Satzcode wird abgewiesen und nicht gesucht.** Scryfall
	// beantwortete `e:` mit allem, was es hat — und das Set hieße danach wie
	// sein Code, also gar nicht. `ValidateBasics` fängt das nicht: es verlangt
	// ein Spiel, keinen Namen.
	if strings.TrimSpace(code) == "" {
		return nil, domain.NewFieldError("code", domain.SchluesselPflichtfeld,
			"Ohne Satzkürzel gibt es nichts einzulesen.")
	}

	if schon, err := s.schonEingelesen(ctx, code); err != nil || schon != nil {
		return schon, err
	}

	zuordnung, _, bekannt := scryfall.ZuordnungFuer(spiel)
	if !bekannt {
		return nil, domain.NewFieldError("spiel", domain.SchluesselValidierung,
			"Für das Spiel %q gibt es keine Zuordnung auf Scryfall-Felder.", spiel)
	}

	satz, befunde, holfehler := s.fremd.Setkarten(ctx, code)
	if holfehler != nil {
		return nil, holfehler
	}

	// **Ein abgebrochener Satz wird nicht halb angelegt**
	// (#1042).
	//
	// Bis zum 2026-08-24 legte dieser Weg die Teilmenge als gewöhnliches Set
	// ab. Ein zweiter Aufruf wurde danach wegen derselben Herkunft als „schon
	// offen" abgewiesen — **derselbe Weg konnte das Set nie
	// vervollständigen**, und im Änderungsprotokoll stand „mit allen Karten".
	//
	// Ein halbes Set, das als vollständig gilt, ist schlimmer als keines: es
	// sieht aus wie eine erledigte Arbeit.
	if satz.Abgebrochen {
		return nil, domain.NewFieldError("code", domain.SchluesselValidierung,
			"Der Satz %q führt %d Karten — mehr, als in einem Zug geholt werden "+
				"(%d Seiten à 175). Er wird nicht halb angelegt.",
			code, satz.Gesamt, scryfall.Seitengrenze)
	}

	// **Der Name ist Beiwerk, kein Grund zum Abbruch** (§8): ein Satz, dessen
	// Angaben nicht zu holen sind, heißt eben wie sein Kürzel — und das ist
	// besser als kein Set.
	name := code
	if angabe, _, _, err := s.fremd.Setangaben(ctx, code); err == nil && angabe != nil {
		name = angabe.Name
	}

	set := &domain.Set{
		Name: name, Game: spiel, Origin: Satzherkunft(code),
		Values: []domain.Value{{Field: "code", Text: code}},
	}

	for _, karte := range satz.Karten {
		set.Cards = append(set.Cards, karteAusSatz(zuordnung, karte))
	}

	ergebnis := &Ergebnis{Befunde: befunde, Angelegt: []int64{}}

	// **Ohne `ValidateBasics`, und das ist kein Versehen.** Sie weist
	// namenlose und doppelte Felder ab — beides kann [karteAusSatz] nicht
	// erzeugen: es geht über die Zuordnung, und die nennt jedes Feld genau
	// einmal und keines namenlos. Der Zweig darauf wäre eine Zeile, die kein
	// Aufruf betreten kann (§3).
	//
	// **Was ihn zurückbrächte**: eine Zuordnung mit zwei Paaren auf dasselbe
	// Feld, oder Werte, die nicht aus ihr stammen — so wie der MTG-JSON-Import
	// seinen `Rest` an die Karte hängt. Dann gehört sie wieder davor.
	kennung, err := s.speicher.SetAnlegen(ctx, set, nil)
	if err != nil {
		return s.verlorenesRennen(ctx, code, ergebnis, err)
	}

	ergebnis.SetID = kennung
	ergebnis.Angelegt = append(ergebnis.Angelegt, kennung)
	ergebnis.Karten = len(set.Cards)

	return ergebnis, nil
}

// verlorenesRennen deutet einen Konflikt beim Ablegen.
//
// # Wofür
//
// Seit Migration 0015 hält die **Datenbank** die Regel „ein Satz von Scryfall,
// ein Set" (#1042).
// Zwei gleichzeitige Importe desselben Satzes sehen beim Nachsehen beide
// „nicht vorhanden"; einer gewinnt, der andere bekommt einen Konflikt.
//
// **Und der wollte dasselbe, was der Gewinner jetzt hat.** Ein Rennen soll
// sich nicht anders anfühlen als eine Wiederholung: hier kommt dieselbe
// Antwort heraus wie beim zweiten Aufruf hintereinander — `SchonOffen` mit der
// Kennung des vorhandenen Sets.
//
// **Nur wenn es wirklich dasteht.** Ein Konflikt aus einem anderen Grund
// bleibt ein Fehler; ihn zu verschlucken hieße, einen Befund in einen Erfolg
// zu verwandeln.
func (s *Satzimport) verlorenesRennen(
	ctx context.Context, code string, ergebnis *Ergebnis, ablagefehler error,
) (*Ergebnis, error) {
	if !errors.Is(ablagefehler, domain.ErrConflict) {
		return ergebnis, fmt.Errorf("Set ablegen: %w", ablagefehler)
	}

	schon, nachsehfehler := s.schonEingelesen(ctx, code)
	if nachsehfehler != nil || schon == nil {
		return ergebnis, fmt.Errorf("Set ablegen: %w", ablagefehler)
	}

	return schon, nil
}

// karteAusSatz macht aus einer Scryfall-Karte eine Karte der Domäne.
//
// **Über dieselbe Zuordnung wie der Einzelabgleich** — eine zweite liefe beim
// ersten Nachziehen auseinander, und dann hinge es am Weg, welche Felder eine
// Karte bekommt.
func karteAusSatz(zuordnung scryfall.Zuordnung, karte *scryfall.Karte) *domain.Card {
	gefunden := zuordnung.Werte(karte)

	// **In der Reihenfolge der Zuordnung, nicht der Map.** Eine Map gibt ihre
	// Schlüssel in wechselnder Reihenfolge heraus; zwei Läufe schrieben dann
	// dieselbe Karte mit verschieden geordneten Feldern, und der Rundlauf
	// gegen `.mse-set` verglich zwei gleiche Sets als verschieden.
	werte := make([]domain.Value, 0, len(gefunden))

	for _, paar := range zuordnung.Felder {
		if wert := gefunden[paar.Feld]; paar.Feld != "" && wert != "" {
			werte = append(werte, domain.Value{Field: paar.Feld, Text: wert})
		}
	}

	return &domain.Card{Values: werte}
}

// schonEingelesen sagt, ob dieser Satz schon im Bestand liegt.
//
// **Dieselbe Regel wie bei Dateien** (#886), anderer Schlüssel: verglichen wird
// die Herkunft, und die ist hier `scryfall:<code>` statt eines Pfades.
func (s *Satzimport) schonEingelesen(ctx context.Context, code string) (*Ergebnis, error) {
	if s.koepfe == nil {
		return nil, nil //nolint:nilnil // „keine Regel, kein Fehler" — der Aufrufer prüft auf nil
	}

	koepfe, err := s.koepfe.Setkoepfe(ctx)
	if err != nil {
		return nil, fmt.Errorf("den Bestand nachsehen: %w", err)
	}

	for _, kopf := range koepfe {
		if kopf.Origin == Satzherkunft(code) {
			return &Ergebnis{
				SetID: kopf.ID, Karten: kopf.Cards,
				Befunde: []domain.Befund{}, Angelegt: []int64{}, SchonOffen: true,
			}, nil
		}
	}

	return nil, nil //nolint:nilnil // siehe oben
}
