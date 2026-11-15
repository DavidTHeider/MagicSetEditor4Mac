package service

// **Die Suche über den gewählten Dienst** — Werkstück 8 von Phase 4.8
// (#1321),
// entschieden in [0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md).

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Suchdienst ist ein fremder Dienst, der Karten suchen kann.
//
// **Die Gestalt ist die von [Kartensuche.Suche]** — dieselbe Antwort, gleich
// wer sie gibt. Das ist §16, „dieselbe Fähigkeit, überall dieselbe Antwort",
// eine Ebene tiefer: nicht dieselbe Fläche, sondern dieselbe **Auskunft**.
type Suchdienst interface {
	Suche(ctx context.Context, frage string) (*Suchauskunft, error)
}

// Quellenkarte gibt die Dienste, die **jetzt** bereitstehen.
//
// # Warum eine Karte je Anfrage und nicht eine feste
//
// **Weil sich beides ändert, was sie füllt**: der Schalter eines Dienstes und
// seine Zugangsdaten. Das Fenster baut diese Naht **einmal**, beim Start; eine
// feste Karte wäre damit der Stand von damals.
//
// **Befund 1 des Zweitblicks über Phase 4.8**
// (#1321):
// die Sperre am Transport löste `an` → `aus` — ein gebauter Klient wird vor
// jeder Anfrage gefragt. Für `aus` → `an` gab es den Klienten gar nicht, und
// wer nach dem Start einen Schlüssel eintrug, sah ihn bis zum Neustart nicht
// wirken. Der Fall war behoben, die Klasse nicht (§10).
type Quellenkarte interface {
	Fuer(ctx context.Context) (Quellen, error)
}

// Quellen sind die bereitstehenden Dienste — **und warum die anderen fehlen**.
//
// # Warum der Grund mitreist
//
// **Weil „steht nicht bereit" zwei Sachverhalte sind** (§8): abgeschaltet, und
// ohne hinterlegten Zugang. Wer beides in einen Satz packt, rät — „fehlt sein
// Zugang, oder ist er abgeschaltet?" ist keine Auskunft, sondern eine Frage
// zurück an den Lesenden.
//
// Genau das war der Preis der Korrektur zu Befund 1: solange die Karte fest
// war, sagte die Sperre am Transport `ErrDienstAus` mit Namen und Weg. Seit
// die Karte je Anfrage entsteht, fehlt ein abgeschalteter Dienst schlicht —
// und ohne diesen Grund daneben wäre die Meldung schlechter geworden.
type Quellen struct {
	// Bereit sind die Dienste, die jetzt antworten können.
	Bereit map[domain.Dienst]Suchdienst

	// Grund sagt je fehlendem Dienst, warum er fehlt — leer, wenn es keinen
	// benannten Grund gibt.
	Grund map[domain.Dienst]error
}

// Fragewahl sagt, welcher Dienst eine Frage beantwortet.
//
// **Sie steht hier als Schnittstelle**, weil `service` den Bestand nicht
// kennt (§2). Wer sie erfüllt, liest die Einstellung.
type Fragewahl interface {
	FuerFrage(ctx context.Context, frage domain.Frage) (domain.Dienst, error)
}

// Dienstsuche fragt den Dienst, der für die Suche gewählt ist.
//
// # Warum sie nicht ausweicht
//
// **Ein Ausweichweg ist still** (0063, Entscheidung 1). Wer
// „magicthegathering.io" einstellt und eine Antwort von Scryfall bekommt, hat
// eine Antwort aus einer Quelle, die er nicht gewählt hat — und sieht es
// nicht. Was bei einem Bild harmlos aussieht, ist bei einem **Preis** eine
// falsche Zahl mit einem falschen Namen daneben.
//
// **Ein Ausweichweg ist damit nicht verboten, sondern eine Handlung**: der
// gewählte Dienst schweigt, die Fläche sagt es und bietet den anderen an.
type Dienstsuche struct {
	quellen Quellenkarte
	wahl    Fragewahl
}

// NeueDienstsuche baut sie über einer Quellenkarte.
func NeueDienstsuche(quellen Quellenkarte, wahl Fragewahl) *Dienstsuche {
	return &Dienstsuche{quellen: quellen, wahl: wahl}
}

// Suche gibt die Treffer des gewählten Dienstes.
func (d *Dienstsuche) Suche(ctx context.Context, frage string) (
	*Suchauskunft, error,
) {
	dienst, err := d.wahl.FuerFrage(ctx, domain.FrageSuche)
	if err != nil {
		return nil, err
	}

	// **Jetzt gelesen und nicht beim Bauen** — siehe [Quellenkarte].
	karte, kartenfehler := d.quellen.Fuer(ctx)
	if kartenfehler != nil {
		return nil, kartenfehler
	}

	quelle, gibts := karte.Bereit[dienst]
	if !gibts {
		// **„Nichts gefunden" wäre die falsche Antwort.** Gefunden wurde
		// nichts, weil gar nicht gefragt wurde — und der Grund ist eine
		// Einstellung, keine Karte (§8).
		//
		// **Der genannte Grund geht vor.** Er nennt den Sachverhalt; der Satz
		// darunter nennt beide Möglichkeiten und ist damit die schlechtere
		// Auskunft — er bleibt für den Fall, dass niemand einen Grund kennt.
		if grund := karte.Grund[dienst]; grund != nil {
			return nil, grund
		}

		return nil, fmt.Errorf(
			"%w: für die Suche ist %s gewählt, aber dieser Dienst steht nicht "+
				"bereit — fehlt sein Zugang, oder ist er abgeschaltet?",
			domain.ErrValidation, dienst)
	}

	return quelle.Suche(ctx, frage)
}
