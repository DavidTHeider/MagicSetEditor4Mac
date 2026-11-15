package service

// **Der Kostenwächter** — Werkstück O12 von Phase 5.2
// (#1439).
//
// # Drei Fragen vor jedem Auftrag, der Geld kostet
//
//  1. **Darf dieser Akteur etwas Verbindliches?** Das ist die Rolle
//     ([domain.WirkungVerbindlich]) — ein Token reicht nie so weit (0078).
//  2. **Hat ein Mensch **diesen** Auftrag gewollt?** Das ist der
//     [domain.Nachweis], und er trägt die Kosten als Bindung: `Kostengrenze`
//     hält den Höchstbetrag gegen das, was er freigegeben hat.
//  3. **Trägt das Budget ihn?** Das entscheidet das Journal, in einer
//     Transaktion — und damit über alle Prozesse hinweg.
//
// # Was hier **nicht** entschieden wird
//
// **Ob der Anbieter das Geld wirklich nimmt.** Das Journal ist eine lokale
// Buchführung; wer denselben Schlüssel woanders benutzt, verbraucht Kontingent,
// von dem es nichts weiß. Die Grenze heißt deshalb **lokale** Grenze, und der
// Satz steht an jeder Fläche.

import (
	"context"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// KostenjournalZweck ist der Zweck des Nachweises für einen bezahlten Auftrag.
const KostenjournalZweck = "kosten.auftrag"

// Kostenpostenliste ist, wie viele Posten eine Auskunft nennt.
//
// **Endlich, wie die Metrikreihen aus O9**: eine Liste, die mit dem Journal
// wächst, ist keine Auskunft mehr, sondern ein Abzug.
const Kostenpostenliste = 50

// Kostenablage ist die Naht zum Journal — erfüllt von `internal/storage`.
type Kostenablage interface {
	BudgetSetzen(ctx context.Context, profilID int64, konto string,
		zeitraum domain.Zeitraum, betrag domain.Betrag, jetzt time.Time) error
	Reservieren(ctx context.Context, posten domain.Kostenposten,
		zeitraum domain.Zeitraum, jetzt time.Time) (domain.Kostenposten, error)
	Abschliessen(ctx context.Context, profilID int64, kennung string,
		betrag domain.Betrag, jetzt time.Time) (domain.Kostenposten, error)
	Freigeben(ctx context.Context, profilID int64, kennung string, jetzt time.Time) (
		domain.Kostenposten, error)
	AlsUnklarMerken(ctx context.Context, profilID int64, kennung, grund string,
		jetzt time.Time) (domain.Kostenposten, error)
	Stand(ctx context.Context, profilID int64, konto string,
		zeitraum domain.Zeitraum, jetzt time.Time) (domain.Haushaltsstand, error)
	Posten(ctx context.Context, profilID int64, konto string,
		zeitraum domain.Zeitraum, jetzt time.Time, hoechstens int) (
		[]domain.Kostenposten, error)
}

// Kostenwache führt das Journal.
type Kostenwache struct {
	ablage Kostenablage
	wache  Ruecksetzungswache
	uhr    func() time.Time
}

// NeueKostenwache baut sie.
//
// **Die Wache steht im Bauplan** — dieselbe Entscheidung wie bei der
// Rücksetzung: eine Buchführung, die niemanden fragt, ist keine.
func NeueKostenwache(
	ablage Kostenablage, wache Ruecksetzungswache, uhr func() time.Time,
) *Kostenwache {
	return &Kostenwache{ablage: ablage, wache: wache, uhr: uhr}
}

// Kostenauftrag ist, was ein Aufrufer vor dem Ruf nach draußen anmeldet.
type Kostenauftrag struct {
	// Kennung ist die **Auftragsidentität** und kommt vom Aufrufer — sie ist
	// der Schutz gegen den zweiten kostenpflichtigen Auftrag.
	Kennung string

	Konto   string
	Aufgabe string

	Hoechstbetrag domain.Betrag

	// Zeitraum ist, worüber gerechnet wird — leer heißt **Tag**, die engere
	// der beiden Grenzen.
	Zeitraum domain.Zeitraum
}

// Reservieren meldet einen Auftrag an — **vor** dem Ruf nach draußen.
func (k *Kostenwache) Reservieren(
	ctx context.Context, auftrag Kostenauftrag, nachweis string,
) (domain.Kostenposten, error) {
	zeitraum, zeitfehler := gewaehlterZeitraum(auftrag.Zeitraum)
	if zeitfehler != nil {
		return domain.Kostenposten{}, zeitfehler
	}

	if nachweis == "" {
		return domain.Kostenposten{}, fmt.Errorf(
			"%w: ein Auftrag, der Geld kostet, braucht die ausdrückliche "+
				"Freigabe eines Menschen", domain.ErrErlaubnis)
	}

	akteur, fehler := k.wache.Akteur(ctx)
	if fehler != nil {
		return domain.Kostenposten{}, fehler
	}

	// **Die Kosten reisen in der Bindung mit** — der Nachweis trägt eine
	// Grenze, und sie wird gegen den Höchstbetrag gehalten (O1).
	if fehler := k.wache.PruefeMitNachweis(ctx, domain.WirkungVerbindlich, nachweis,
		Nachweisbindung{
			Zweck: KostenjournalZweck, Ziel: auftrag.Konto,
			Kosten: auftrag.Hoechstbetrag.Wert, Waehrung: auftrag.Hoechstbetrag.Waehrung,
		}); fehler != nil {
		return domain.Kostenposten{}, fehler
	}

	return k.ablage.Reservieren(ctx, domain.Kostenposten{
		Kennung: auftrag.Kennung, ProfilID: akteur.ProfilID, Konto: auftrag.Konto,
		Aufgabe: auftrag.Aufgabe, Hoechstbetrag: auftrag.Hoechstbetrag,
	}, zeitraum, k.uhr())
}

// Abschliessen trägt den wirklichen Betrag ein.
//
// **Ohne zweiten Nachweis.** Er galt dem Auftrag, nicht der Buchhaltung; ihn
// hier noch einmal zu verlangen hieße, einen Menschen zu etwas zustimmen zu
// lassen, das schon geschehen ist.
func (k *Kostenwache) Abschliessen(
	ctx context.Context, kennung string, betrag domain.Betrag,
) (domain.Kostenposten, error) {
	akteur, fehler := k.wache.Akteur(ctx)
	if fehler != nil {
		return domain.Kostenposten{}, fehler
	}

	return k.ablage.Abschliessen(ctx, akteur.ProfilID, kennung, betrag, k.uhr())
}

// Freigeben sagt, dass nichts angefallen ist.
//
// **Im Profil dessen, der fragt** — dieselbe Frage wie beim Reservieren
// (Nachlauf über Phase 6.1, Befund 4): die Kennung gibt der Aufrufer vor, und
// in zwei Profilen kann es dieselbe sein.
func (k *Kostenwache) Freigeben(ctx context.Context, kennung string) (
	domain.Kostenposten, error,
) {
	akteur, fehler := k.wache.Akteur(ctx)
	if fehler != nil {
		return domain.Kostenposten{}, fehler
	}

	return k.ablage.Freigeben(ctx, akteur.ProfilID, kennung, k.uhr())
}

// AlsUnklarMerken hält die Reservierung und sagt, warum.
func (k *Kostenwache) AlsUnklarMerken(ctx context.Context, kennung, grund string) (
	domain.Kostenposten, error,
) {
	akteur, fehler := k.wache.Akteur(ctx)
	if fehler != nil {
		return domain.Kostenposten{}, fehler
	}

	return k.ablage.AlsUnklarMerken(ctx, akteur.ProfilID, kennung, grund, k.uhr())
}

// Stand sagt, was an einem Konto steht.
func (k *Kostenwache) Stand(
	ctx context.Context, konto string, zeitraum domain.Zeitraum,
) (domain.Haushaltsstand, error) {
	gewaehlt, zeitfehler := gewaehlterZeitraum(zeitraum)
	if zeitfehler != nil {
		return domain.Haushaltsstand{}, zeitfehler
	}

	akteur, fehler := k.wache.Akteur(ctx)
	if fehler != nil {
		return domain.Haushaltsstand{}, fehler
	}

	return k.ablage.Stand(ctx, akteur.ProfilID, konto, gewaehlt, k.uhr())
}

// Posten gibt die letzten Aufträge eines Kontos.
func (k *Kostenwache) Posten(
	ctx context.Context, konto string, zeitraum domain.Zeitraum,
) ([]domain.Kostenposten, error) {
	gewaehlt, zeitfehler := gewaehlterZeitraum(zeitraum)
	if zeitfehler != nil {
		return nil, zeitfehler
	}

	akteur, fehler := k.wache.Akteur(ctx)
	if fehler != nil {
		return nil, fehler
	}

	return k.ablage.Posten(ctx, akteur.ProfilID, konto, gewaehlt, k.uhr(), Kostenpostenliste)
}

// BudgetSetzen legt die Grenze eines Kontos fest.
//
// **Ohne Nachweis, und das ist die Aussage**: eine Grenze gibt nichts aus, sie
// begrenzt. Sie zu setzen ist ein Schreiben wie jedes andere — was **darüber
// hinaus** geht, verlangt danach ohnehin je Auftrag eine Freigabe.
func (k *Kostenwache) BudgetSetzen(
	ctx context.Context, konto string, zeitraum domain.Zeitraum, betrag domain.Betrag,
) error {
	gewaehlt, zeitfehler := gewaehlterZeitraum(zeitraum)
	if zeitfehler != nil {
		return zeitfehler
	}

	akteur, fehler := k.wache.Akteur(ctx)
	if fehler != nil {
		return fehler
	}

	if fehler := k.wache.PruefeMitNachweis(ctx, domain.WirkungSchreiben, "",
		Nachweisbindung{}); fehler != nil {
		return fehler
	}

	return k.ablage.BudgetSetzen(ctx, akteur.ProfilID, konto, gewaehlt, betrag, k.uhr())
}

// gewaehlterZeitraum gibt den Zeitraum — leer heißt **Tag**.
//
// **Die engere der beiden**: wer nichts nennt, bekommt die Grenze, die
// häufiger greift. Eine Vorgabe, die den Monat meint, ließe an einem Tag
// dreißigmal so viel durch.
func gewaehlterZeitraum(gewuenscht domain.Zeitraum) (domain.Zeitraum, error) {
	if gewuenscht == "" {
		return domain.ZeitraumTag, nil
	}

	if !domain.IstZeitraum(string(gewuenscht)) {
		return "", fmt.Errorf("%w: %q ist kein Zeitraum, über den gerechnet wird",
			domain.ErrValidation, gewuenscht)
	}

	return gewuenscht, nil
}
