package service_test

// **Ein Partiespeicher im Arbeitsspeicher** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Was er hält und was er nicht ist
//
// **Er ist ein Doppel und kein zweiter Bestand** (§3): er hält Partien,
// Ereignisse und die Kennungen der Befehle, die schon gewirkt haben — mehr
// braucht die Leitung nicht.
//
// **Was er ausdrücklich nicht prüft, ist die Atomarität.** Er schreibt in
// einem Rutsch, weil er in einem Prozess lebt; dass zwei Tabellen wirklich
// zusammen geschrieben werden, misst `internal/storage` gegen eine echte
// Datenbank — und dort gehört es hin.

import (
	"context"
	"fmt"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

type partieprobe struct {
	partien map[string]domain.Partie

	// ereignisse hält den Strom je Partie.
	ereignisse map[string][]domain.Ereignis

	// befehle merkt, was ein Befehl bewirkt hat.
	befehle map[string][]domain.Ereignis

	// letzter merkt je Partie, was der zuletzt angewendete Befehl schrieb.
	letzter map[string][]domain.Ereignis

	angelegt   int
	angewendet int

	klagt error

	// stolpert klagt je Weg — für die Fehler, die ein `klagt` nie erreicht.
	//
	// **Ein Speicher, der auf jedem Weg klagt, klagt schon beim ersten**,
	// und alles danach bleibt ungemessen (§10).
	stolpert map[string]error
	// letzteOhneStrom lässt `LetzteKette` eine Folge nennen, die `Ereignisse`
	// nicht mehr hat — der Widerspruch, den ein Wirt benennen muss (§8).
	letzteOhneStrom bool

	// freigaben hält die Einwilligungen je Partie (G5, #1462).
	freigaben map[string][]domain.Offenlegung

	// geloescht merkt, was wirklich verschwunden ist (PLAY-03.AK-05).
	//
	// **Eine Liste und kein Schalter**: „gelöscht wurde etwas" und
	// „gelöscht wurde *diese*" sind zwei Aussagen, und die zweite ist die,
	// die ein Lauf braucht.
	geloescht []string
}

// neueProbe gibt eine Probe, deren Abbildungen alle stehen.
//
// **Die meisten Läufe bauen sie von Hand und lassen weg, was sie nicht
// brauchen** — das geht, solange jeder Weg eine fehlende Abbildung erträgt.
// Wer alle drei braucht, nimmt diese hier statt drei Zeilen zu wiederholen.
func neueProbe() *partieprobe {
	return &partieprobe{
		partien:    map[string]domain.Partie{},
		ereignisse: map[string][]domain.Ereignis{},
		befehle:    map[string][]domain.Ereignis{},
		freigaben:  map[string][]domain.Offenlegung{},
	}
}

func (p *partieprobe) Freigaben(
	_ context.Context, partieID string,
) ([]domain.Offenlegung, error) {
	if fehler := p.stolpernAuf("freigaben"); fehler != nil {
		return nil, fehler
	}

	return p.freigaben[partieID], nil
}

func (p *partieprobe) Freigeben(
	ctx context.Context, partieID string, gabe domain.Offenlegung,
) error {
	if fehler := p.stolpernAuf("freigeben"); fehler != nil {
		return fehler
	}

	if fehler := p.Freigabeloeschen(ctx, partieID, gabe.Von); fehler != nil {
		return fehler
	}

	p.freigaben[partieID] = append(p.freigaben[partieID], gabe)

	return nil
}

func (p *partieprobe) Freigabeloeschen(
	_ context.Context, partieID, von string,
) error {
	if fehler := p.stolpernAuf("freigabeloeschen"); fehler != nil {
		return fehler
	}

	if p.freigaben == nil {
		p.freigaben = map[string][]domain.Offenlegung{}
	}

	geblieben := make([]domain.Offenlegung, 0, len(p.freigaben[partieID]))

	for _, gabe := range p.freigaben[partieID] {
		if gabe.Von != von {
			geblieben = append(geblieben, gabe)
		}
	}

	p.freigaben[partieID] = geblieben

	return nil
}

func (p *partieprobe) Vergessen(
	_ context.Context, partieID string,
) error {
	if fehler := p.stolpernAuf("vergessen"); fehler != nil {
		return fehler
	}

	delete(p.partien, partieID)
	delete(p.ereignisse, partieID)
	delete(p.befehle, partieID)
	delete(p.freigaben, partieID)

	return nil
}

// stolpert sagt, ob dieser Weg klagen soll.
func (p *partieprobe) stolpernAuf(weg string) error {
	if p.klagt != nil {
		return p.klagt
	}

	return p.stolpert[weg]
}

func (p *partieprobe) Anlegen(
	_ context.Context, partie domain.Partie,
) (domain.Partie, error) {
	// **Über `stolpernAuf` und nicht über `klagt` allein**: ein Speicher,
	// der auf jedem Weg klagt, klagt schon beim ersten, und alles danach
	// bleibt ungemessen (§10). `klagt` gilt weiterhin — es steht in
	// [partieprobe.stolpernAuf] an erster Stelle.
	if fehler := p.stolpernAuf("anlegen"); fehler != nil {
		return domain.Partie{}, fehler
	}

	p.angelegt++

	partie.Fassung = 1
	p.partien[partie.ID] = partie

	return partie, nil
}

func (p *partieprobe) Partie(
	_ context.Context, kennung string,
) (domain.Partie, error) {
	if p.klagt != nil {
		return domain.Partie{}, p.klagt
	}

	partie, da := p.partien[kennung]
	if !da {
		return domain.Partie{}, fmt.Errorf(
			"%w: die Partie %q gibt es nicht", domain.ErrNotFound, kennung)
	}

	return partie, nil
}

// Partien gibt die Partien der Probe — in stabiler Reihenfolge.
//
// **Nach der Kennung sortiert und nicht in Kartenreihenfolge**: eine `map`
// gibt sie zufällig, und ein Lauf über zwei Partien wäre dann mal grün und
// mal rot.
func (p *partieprobe) Partien(_ context.Context) ([]domain.Partie, error) {
	if p.klagt != nil {
		return nil, p.klagt
	}

	kennungen := make([]string, 0, len(p.partien))
	for kennung := range p.partien {
		kennungen = append(kennungen, kennung)
	}

	slices.Sort(kennungen)

	gefunden := make([]domain.Partie, 0, len(kennungen))
	for _, kennung := range kennungen {
		gefunden = append(gefunden, p.partien[kennung])
	}

	return gefunden, nil
}

func (p *partieprobe) Anwenden(
	_ context.Context, erwartet int64, befehlID string,
	zustand domain.Partie, ereignisse []domain.Ereignis,
) error {
	if fehler := p.stolpernAuf("Anwenden"); fehler != nil {
		return fehler
	}

	// **Die erwartete Fassung wird auch hier geprüft**, und das ist keine
	// Doppelung: die Leitung fragt, damit der Aufrufer einen Satz bekommt;
	// der Speicher fragt, damit zwei gleichzeitige Schreiber nicht beide
	// gewinnen. Im echten Bestand ist es ein `WHERE fassung = ?`.
	if p.partien[zustand.ID].Fassung != erwartet {
		return fmt.Errorf(
			"%w: die Fassung hat sich geändert", domain.ErrConflict)
	}

	p.angewendet++

	p.partien[zustand.ID] = zustand

	if p.ereignisse == nil {
		p.ereignisse = map[string][]domain.Ereignis{}
	}

	if p.befehle == nil {
		p.befehle = map[string][]domain.Ereignis{}
	}

	p.ereignisse[zustand.ID] = append(p.ereignisse[zustand.ID], ereignisse...)
	p.befehle[zustand.ID+"/"+befehlID] = ereignisse

	if p.letzter == nil {
		p.letzter = map[string][]domain.Ereignis{}
	}

	p.letzter[zustand.ID] = ereignisse

	return nil
}

func (p *partieprobe) Befehlsergebnis(
	_ context.Context, partieID, befehlID string,
) ([]domain.Ereignis, bool, error) {
	if fehler := p.stolpernAuf("Befehlsergebnis"); fehler != nil {
		return nil, false, fehler
	}

	gefunden, da := p.befehle[partieID+"/"+befehlID]

	return gefunden, da, nil
}

func (p *partieprobe) LetzterBefehl(
	_ context.Context, partieID string,
) ([]domain.Ereignis, error) {
	if fehler := p.stolpernAuf("LetzterBefehl"); fehler != nil {
		return nil, fehler
	}

	return p.letzter[partieID], nil
}

func (p *partieprobe) LetzteKette(
	_ context.Context, partieID string,
) (kette string, folge int64, fehler error) {
	if gestolpert := p.stolpernAuf("LetzteKette"); gestolpert != nil {
		return "", 0, gestolpert
	}

	strom := p.ereignisse[partieID]
	if len(strom) == 0 {
		return "", 0, nil
	}

	letztes := strom[len(strom)-1]

	if p.letzteOhneStrom {
		p.ereignisse[partieID] = nil
	}

	return letztes.Kette, letztes.Folge, nil
}

func (p *partieprobe) Standsetzen(
	_ context.Context, partieID string, stand domain.Partiestand,
) (domain.Partie, error) {
	if p.klagt != nil {
		return domain.Partie{}, p.klagt
	}

	partie, da := p.partien[partieID]
	if !da {
		return domain.Partie{}, fmt.Errorf(
			"%w: die Partie %q gibt es nicht", domain.ErrNotFound, partieID)
	}

	partie.Stand = stand
	partie.Fassung++
	p.partien[partieID] = partie

	return partie, nil
}

// Ereignisse gibt den Strom — **ungefiltert**, wie die Schnittstelle es
// verlangt (G5, #1462).
//
// **Die Probe hält alle Ereignisse in ihrer Reihenfolge**, und `von`/`bis`
// schneiden daraus; null und null bedeuten alles.
func (p *partieprobe) Ereignisse(
	_ context.Context, partieID string, von, bis int64,
) ([]domain.Ereignis, error) {
	if fehler := p.stolpernAuf("ereignisse"); fehler != nil {
		return nil, fehler
	}

	gefunden := make([]domain.Ereignis, 0, len(p.ereignisse[partieID]))

	for _, ereignis := range p.ereignisse[partieID] {
		if von != 0 && ereignis.Folge < von {
			continue
		}

		if bis != 0 && ereignis.Folge > bis {
			continue
		}

		gefunden = append(gefunden, ereignis)
	}

	return gefunden, nil
}
