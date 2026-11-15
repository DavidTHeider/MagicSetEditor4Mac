package app

// **Wann der Bestand zuletzt gesichert wurde, für jede Fläche** — OPS01
// (#2347),
// Befund A-06 des Audits 0.9.0.
//
// # Was hier nicht entschieden wird
//
// **Nichts.** Wie alt „alt" ist, steht in `domain.SicherungsstandBei`, wann
// der Plan fällig ist, in `domain.SicherungsplanFaellig` — an **einer**
// Stelle, damit vier Flächen dieselbe Antwort geben (§16).

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// SicherungsstandDTO ist das Alter der letzten Sicherung.
//
// **`gesichert` ist leer, wenn es keinen Vermerk gibt** — unbekannt heißt
// unbekannt, und eine Fläche, die daraus eine Zeit machte, erfände eine.
type SicherungsstandDTO struct {
	// Stufe ist `unbekannt`, `frisch`, `alt` (über 14 Tage) oder
	// `unstimmig` (der Vermerk liegt nach der Uhr).
	Stufe string `json:"stufe"`

	// Gesichert ist der Augenblick der letzten geprüften Sicherung, in
	// Weltzeit nach RFC 3339.
	Gesichert string `json:"gesichert"`

	// Tage sind die vollen Tage seither.
	Tage int `json:"tage"`

	Anlass         string `json:"anlass"`
	Flaeche        string `json:"flaeche"`
	Datei          string `json:"datei"`
	Verschluesselt bool   `json:"verschluesselt"`

	Plan SicherungsplanDTO `json:"plan"`
}

// SicherungsplanDTO ist der freiwillige Plan.
type SicherungsplanDTO struct {
	// Ordner ist der Ordner des Plans — leer heißt: kein Plan.
	Ordner string `json:"ordner"`

	// Letzter ist der Augenblick der letzten Sicherung nach Plan — leer, wenn
	// es keine gibt.
	Letzter string `json:"letzter"`

	// Faellig sagt, ob der Plan beim nächsten Lauf sichert.
	Faellig bool `json:"faellig"`
}

// zeitfuerFlaechen ist die Form einer Zeit an den Flächen.
const zeitfuerFlaechen = time.RFC3339

// sicherungsstandAlsDTO übersetzt den Stand, den letzten Planvermerk und den
// Ordner des Plans.
func sicherungsstandAlsDTO(
	stand domain.Sicherungsstand, plan domain.Sicherungsvermerk, planBekannt bool, ordner string, jetzt time.Time,
) SicherungsstandDTO {
	dto := SicherungsstandDTO{Stufe: string(stand.Stufe), Tage: stand.Tage}

	if stand.Stufe != domain.SicherungsstufeUnbekannt {
		letzter := stand.Letzter
		dto.Gesichert = letzter.Gesichert.UTC().Format(zeitfuerFlaechen)
		dto.Anlass, dto.Flaeche, dto.Datei = string(letzter.Anlass), string(letzter.Flaeche), letzter.Datei
		dto.Verschluesselt = letzter.Verschluesselt
	}

	if ordner != "" {
		dto.Plan.Ordner = ordner
		dto.Plan.Faellig = domain.SicherungsplanFaellig(plan, planBekannt, jetzt)
	}

	if planBekannt {
		dto.Plan.Letzter = plan.Gesichert.UTC().Format(zeitfuerFlaechen)
	}

	return dto
}

// sicherungsstandUeber liest den Stand über einem Bestandsweg — für jede
// Fläche dieselbe Folge. **Jeder Fehler wird genannt**: die drei Lesungen
// hängen nicht voneinander ab, und eine Meldung über die erste verschwiege
// die übrigen.
func sicherungsstandUeber(
	ctx context.Context, weg *service.Bestandsweg, einstellungen Einstellungsfuehrer, jetzt time.Time,
) (SicherungsstandDTO, error) {
	// **Erst der Plan, dann der Stand**: Vermerke kommen nur hinzu, und wer
	// den Plan zuerst liest, sieht im Stand danach jede Sicherung, die der
	// Plan nennt. Andersherum nannte der Plan eine, die der Stand noch nicht
	// kannte — sicherte er genau zwischen beiden Lesungen (gesehen unter Last
	// an `TestDerPlanSichertGleichNachDerOrdnerwahl`).
	ordner, ordnerfehler := planordnerAus(ctx, einstellungen)
	plan, bekannt, planfehler := weg.Planvermerk(ctx, ordner)
	stand, standfehler := weg.Sicherungsstand(ctx)

	if fehler := errors.Join(standfehler, planfehler, ordnerfehler); fehler != nil {
		return SicherungsstandDTO{}, fehler
	}

	return sicherungsstandAlsDTO(stand, plan, bekannt, ordner, jetzt), nil
}

// planordnerAus liest den Ordner des Plans — mit `~` gedehnt zum Zuhause
// dessen, der die Anwendung ausführt (`domain.OrdnerAus`).
func planordnerAus(ctx context.Context, einstellungen Einstellungsfuehrer) (string, error) {
	werte, fehler := einstellungen.Alle(ctx)
	if fehler != nil {
		return "", fehler
	}

	zuhause, _ := os.UserHomeDir() //nolint:errcheck // ohne Zuhause bleibt `~/…` ungedehnt und wird nicht gefunden

	return domain.OrdnerAus(werte[domain.SicherungsplanOrdner], zuhause), nil
}

// Sicherungsstand nennt das Alter der letzten Sicherung — für die Flächen
// ohne Fenster.
func (n *Bestandsnaht) Sicherungsstand(ctx context.Context) (SicherungsstandDTO, error) {
	return sicherungsstandUeber(ctx, n.weg, einstellungsverwalterUeber(*n.offen), time.Now())
}

// SicherungsstandAntwort ist der Stand am Fenster — dazu, woran der Plan
// zuletzt scheiterte, denn nur das Fenster führt ihn aus.
type SicherungsstandAntwort struct {
	Stand      SicherungsstandDTO `json:"stand"`
	Planfehler *Fehler            `json:"planfehler"`
	Fehler     *Fehler            `json:"fehler"`
}

// Sicherungsstand nennt das Alter der letzten Sicherung am Fenster.
func (a *App) Sicherungsstand(ctx context.Context) SicherungsstandAntwort {
	ctx = a.imProfil(ctx)

	stand, fehler := sicherungsstandUeber(ctx, a.bestandsweg(), a.einstellungen, time.Now())
	if fehler != nil {
		return SicherungsstandAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return SicherungsstandAntwort{Stand: stand, Planfehler: a.sicherungsplanfehler(ctx)}
}

// Sicherungsstand ist die Fensterbindung.
func (b *Bindungen) Sicherungsstand(ctx context.Context) SicherungsstandAntwort {
	return b.app(ctx).Sicherungsstand(b.ruf(ctx))
}

// SicherungsplanEinrichten fragt nach einem Ordner und richtet den Plan dort
// ein — **nur nach einer Ordnerwahl**: ein abgebrochener Dialog richtet
// nichts ein, und der Plan ist freiwillig. **Gleich danach fragt er**, und
// weil er ohne eigenen Vermerk fällig ist, sichert er.
//
// **Ob es den Ordner gibt, fragt das Setzen** (`IstOrdnerangabe`) — ein Stick,
// der eben abgezogen wurde, wird abgewiesen statt angelegt.
func (a *App) SicherungsplanEinrichten(ctx context.Context) SicherungsstandAntwort {
	ctx = a.imProfil(ctx)

	// **Gefragt, bevor jemand einen Ordner wählt** (Zweitblick über Phase 4,
	// Befund 4) — ob es geht, entscheidet `service`.
	if fehler := a.bestandsweg().PlanPruefen(); fehler != nil {
		return SicherungsstandAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	ordner, fehler := a.waehleDatei(ctx, Dateiauswahl{Titel: DialogtexteFuer(a.sprache).Planordner, Verzeichnis: true})
	if fehler != nil {
		return SicherungsstandAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	if ordner == "" {
		return a.Sicherungsstand(ctx)
	}

	if fehler := a.einstellungen.Setzen(ctx, domain.SicherungsplanOrdner, ordner); fehler != nil {
		return SicherungsstandAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	a.planfehlerMerken(nil)
	a.sicherungsplanAnstossen()

	return a.Sicherungsstand(ctx)
}

// SicherungsplanBeenden nimmt den Ordner des Plans weg — **die Sicherungen
// darin bleiben**: sie sind Dateien am Ort, den jemand gewählt hat.
func (a *App) SicherungsplanBeenden(ctx context.Context) SicherungsstandAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.einstellungen.Setzen(ctx, domain.SicherungsplanOrdner, ""); fehler != nil {
		return SicherungsstandAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	a.planfehlerMerken(nil)

	return a.Sicherungsstand(ctx)
}

// SicherungsplanEinrichten ist die Fensterbindung.
func (b *Bindungen) SicherungsplanEinrichten(ctx context.Context) SicherungsstandAntwort {
	return b.app(ctx).SicherungsplanEinrichten(b.ruf(ctx))
}

// SicherungsplanBeenden ist die Fensterbindung.
func (b *Bindungen) SicherungsplanBeenden(ctx context.Context) SicherungsstandAntwort {
	return b.app(ctx).SicherungsplanBeenden(b.ruf(ctx))
}
