package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Verlaufsspeicher ist, was der Verlaufsdienst vom Bestand braucht.
//
// Die Schnittstelle steht hier und wird in `internal/storage` erfüllt — nie
// umgekehrt (AGENTS.md §2). Dass sie `domain.Verlaufseintrag` nennt und keinen
// eigenen Typ, ist die Bedingung dafür: `service` darf `storage` nicht kennen,
// also muss der Typ in einer Schicht liegen, die beide sehen dürfen.
type Verlaufsspeicher interface {
	Verlauf(ctx context.Context, setID int64) ([]domain.Verlaufseintrag, error)
	// Zuruecknehmen nimmt den jüngsten geltenden Schritt des Sets zurück,
	// oder mit `vorgang` **genau diesen** (#1393).
	VorgangZuruecknehmen(ctx context.Context, setID int64, vorgang string) (
		domain.Verlaufsschritt, bool, error)
	VorgangWiederherstellen(ctx context.Context, setID int64, vorgang string) (
		domain.Verlaufsschritt, bool, error)
}

// Bewegung ist der Ausgang einer Rücknahme oder Wiederherstellung.
//
// # Warum sie mehr sagt als „es hat geklappt"
//
// **Die Karte**, weil eine Rücknahme genau **einen** Datensatz wiederherstellt
// (§7): wer nur erfährt, dass etwas geschehen ist, lädt den ganzen Stand nach
// und überschreibt damit, woran gerade jemand tippt.
//
// **Den Urheber**, weil zwei Prozesse schreiben
// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)). Eine
// Rücknahme, die nicht sagt, wessen Änderung sie umkehrt, ist im
// Zwei-Schreiber-Betrieb eine Vermutung (#52) — die Oberfläche soll es sagen
// können.
type Bewegung struct {
	// Bewegt ist falsch, wenn es nichts zu bewegen gab. **Kein Fehler**: ⌘Z am
	// Anfang des Verlaufs ist eine gewöhnliche Lage und keine Meldung wert.
	Bewegt bool

	KarteID int64
	Feld    string

	Urheber domain.Urheber

	// Umfang ist die Zahl der Karten, die der Schritt berührt hat — eins bei
	// einer Einzeländerung, viele bei einer Massenänderung.
	Umfang int
}

// Verlaufsdienst nimmt Änderungen zurück und stellt sie wieder her.
//
// # Was er trägt und was nicht
//
// Er ist dünn, und das gehört gesagt: heute übersetzt er im Wesentlichen einen
// Eintrag in eine Bewegung und macht aus „nichts da" eine Antwort statt eines
// Fehlers. Er steht trotzdem als eigene Schicht da, weil **hier** hingehört, was
// dieses Werkstück noch erwartet — eine Rücknahme, die nach Urheber unterscheidet,
// und eine Massenänderung, die als **ein** Schritt zurückgeht statt als
// dreihundert. Beides ist eine fachliche Regel über mehreren Entitäten und
// gehört weder in `app` noch in `storage` (§2).
type Verlaufsdienst struct {
	speicher Verlaufsspeicher

	// sets sagt, ob es ein Set gibt (#1336) — wahlfrei, wie bei den Verweisen.
	sets Setkopfleser
}

// MitSetwache lässt den Verlauf eine fehlende Setkennung abweisen.
func (d *Verlaufsdienst) MitSetwache(sets Setkopfleser) *Verlaufsdienst {
	d.sets = sets

	return d
}

// NeuerVerlaufsdienst baut den Dienst über einem Speicher.
func NeuerVerlaufsdienst(speicher Verlaufsspeicher) *Verlaufsdienst {
	return &Verlaufsdienst{speicher: speicher}
}

// Verlauf gibt die Einträge eines Sets, älteste zuerst.
func (d *Verlaufsdienst) Verlauf(
	ctx context.Context, setID int64,
) ([]domain.Verlaufseintrag, error) {
	return d.speicher.Verlauf(ctx, setID)
}

// Zuruecknehmen macht eine Änderung rückgängig.
//
// **„Nichts zu bewegen" ist für ein Set ohne Verlauf richtig — für eines ohne
// Dasein nicht** (#1336). Das Issue stellt die Frage ausdrücklich, und die
// Antwort ist: ein Set, das es nicht gibt, hat keinen Verlauf, den man
// zurücknehmen könnte. Es ist ein anderer Sachverhalt.
//
// # Welche Änderung
//
// **Ohne `vorgang` die jüngste noch geltende des Sets** — das ist ⌘Z, und am
// Fenster ist es genau richtig. **Mit `vorgang` genau diese**
// (#1393):
// über MCP, Kommandozeile und REST schreibt womöglich ein zweiter Prozess in
// dasselbe Set, und „die jüngste" träfe dann dessen Änderung.
//
// Die Kennung kommt aus [Kartenaufsicht.FelderSetzen].
func (d *Verlaufsdienst) Zuruecknehmen(
	ctx context.Context, setID int64, vorgang string,
) (Bewegung, error) {
	if fehler := d.setDaSein(ctx, setID); fehler != nil {
		return Bewegung{}, fehler
	}

	return bewegung(d.speicher.VorgangZuruecknehmen(ctx, setID, vorgang))
}

// setDaSein weist eine Kennung ab, die auf nichts zeigt — wenn eine Wache da ist.
func (d *Verlaufsdienst) setDaSein(ctx context.Context, setID int64) error {
	if d.sets == nil {
		return nil
	}

	return SetDaSein(ctx, d.sets, setID)
}

// Wiederherstellen führt eine zurückgenommene Änderung wieder aus.
//
// **Ohne `vorgang` die jüngste zurückgenommene, mit ihm genau diese** — sonst
// wäre die gezielte Rücknahme eine Einbahnstraße (#1393).
func (d *Verlaufsdienst) Wiederherstellen(
	ctx context.Context, setID int64, vorgang string,
) (Bewegung, error) {
	if fehler := d.setDaSein(ctx, setID); fehler != nil {
		return Bewegung{}, fehler
	}

	return bewegung(d.speicher.VorgangWiederherstellen(ctx, setID, vorgang))
}

// bewegung macht aus einem Eintrag den Ausgang.
//
// **Eine Bewegung, die es nicht gab, trägt nichts** — auch keinen Urheber. Ein
// leerer Eintrag mit gesetztem Namen wäre eine Aussage über eine Änderung, die
// niemand gemacht hat.
func bewegung(schritt domain.Verlaufsschritt, gab bool, err error) (Bewegung, error) {
	if err != nil || !gab {
		return Bewegung{}, err
	}

	return Bewegung{
		Bewegt:  true,
		KarteID: schritt.KarteID,
		Feld:    schritt.Feld,
		Urheber: schritt.Urheber,
		Umfang:  schritt.Umfang,
	}, nil
}
