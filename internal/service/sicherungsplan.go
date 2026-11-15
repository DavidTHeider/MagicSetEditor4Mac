package service

// **Der freiwillige Plan: wöchentlich in einen Ordner, sieben eigene** —
// OPS01 (#2347),
// Befund A-06 des Audits 0.9.0. Marke `new`: MSE2 sichert keinen Bestand.
//
// # Was „eigen" heißt
//
// **Eine Datei, die ein Vermerk dieses Bestands wiedererkennt** — Name des
// Plans, eine gewöhnliche Datei (kein Verweis), Größe und SHA-256 wie
// vermerkt. Was davon abweicht, bleibt liegen: eine Datei unter dem Namen des
// Plans, die niemand vermerkt hat, eine, die sich seither geändert hat, ein
// Verweis, alles mit anderem Namen. Das Issue nennt es als Risiko — „nie
// beliebige Ordnerinhalte" —, und ein Ordner, in dem jemand eine eigene
// Kopie ablegt, ist genau der Fall.
//
// **Was das kostet:** nach einem Zurückspielen kennt der Bestand die
// jüngeren Vermerke nicht mehr, und ihre Dateien bleiben liegen, bis jemand
// sie wegräumt. Das ist die sichere Richtung; die andere hieße, nach dem
// Namen zu löschen.
//
// # Wann weggeräumt wird
//
// **Erst nach einer geglückten, geprüften und vermerkten Sicherung**, und
// nie davor: der letzte gute Stand bleibt, auch auf einer vollen Platte und
// nach einem Abbruch. Ein Lauf, der nicht sichert, räumt nichts.
//
// # Zwei Prozesse
//
// Die Sperre stellt der Aufrufer — eine Nachbardatei des Bestands
// (`internal/app`). **Ist sie belegt, sichert dieser Lauf nicht** und wartet
// auch nicht: der andere sichert gerade, und danach ist der Plan nicht mehr
// fällig.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// planvorsatz und planzeit bilden den Namen einer Sicherung des Plans.
const (
	planvorsatz = "bestand-plan-"
	planzeit    = "20060102-150405"
)

// Sicherungsplan sichert wöchentlich in einen Ordner.
type Sicherungsplan struct {
	weg *Bestandsweg

	// ordner gibt den Ordner des Plans — leer heißt: kein Plan.
	ordner func(context.Context) (string, error)

	// sperre nimmt die Sperre des Plans, **ohne zu warten**.
	sperre func() (frei func(), belegt bool, fehler error)

	flaeche domain.Flaeche
}

// Planlauf ist, was ein Lauf des Plans getan hat.
type Planlauf struct {
	// Gesichert ist der Name der neuen Sicherung — leer, wenn keine entstand.
	Gesichert string

	// Entfernt sind die eigenen Sicherungen, die dabei gingen.
	Entfernt []string

	// Belegt sagt, dass ein anderer Prozess gerade nach Plan sichert.
	Belegt bool
}

// NeuerSicherungsplan baut ihn über einem Bestandsweg mit Buch.
func NeuerSicherungsplan(
	weg *Bestandsweg, ordner func(context.Context) (string, error),
	sperre func() (func(), bool, error), flaeche domain.Flaeche,
) *Sicherungsplan {
	return &Sicherungsplan{weg: weg, ordner: ordner, sperre: sperre, flaeche: flaeche}
}

// PlanPruefen sagt, ob sich dieser Bestand nach Plan sichern lässt — vor dem
// Einrichten und vor jedem Lauf.
//
// **Ein verschlüsselter nicht** (Zweitblick über Phase 4, Befund 4): der Plan
// hat weder ein Kennwort noch ein ausdrückliches „im Klartext", und
// [Bestandsweg.SicherungswunschPruefen] wies jeden Lauf ab — mit dem Rat der
// Sicherung von Hand, die beides anbietet. Ein eingeschalteter Plan blieb so
// wirkungslos, ohne es zu sagen. **Ein Plan mit Kennwort** wäre ein Kennwort,
// das die Anwendung aufbewahrt, und damit eine eigene Entscheidung.
//
// **Ohne Chiffre gilt der Bestand als Klartext** — wie beim Sichern: eine
// Fläche, die die Verschlüsselung nicht kennt, sichert nicht anders.
func (b *Bestandsweg) PlanPruefen() error {
	if verschluesselt, _ := b.Verschluesselungsstand(); verschluesselt { //nolint:errcheck // siehe oben
		return planVerschluesselt()
	}

	return nil
}

// planVerschluesselt ist die Ablehnung eines Plans über einem verschlüsselten
// Bestand — **für den Lauf und für jedes Einrichten**, auch über die
// Einstellung (Nachlauf über Phase 4, Befund 4).
func planVerschluesselt() error {
	return domain.NewFieldError("planordner", domain.SchluesselPlanVerschluesselt,
		"Der Bestand ist verschlüsselt, und der Plan sichert ohne Kennwort — nach Plan wird er deshalb nicht "+
			"gesichert. Sichere ihn von Hand mit Kennwort.")
}

// Laufen sichert, wenn der Plan fällig ist, und dünnt danach aus.
//
// **Ein Fehler beim Ausdünnen nimmt die Sicherung nicht zurück** (§7,
// Regel 3): der Lauf nennt sie, und der Fehler steht daneben.
func (p *Sicherungsplan) Laufen(ctx context.Context) (Planlauf, error) {
	ordner, fehler := p.ordner(ctx)
	if fehler != nil || ordner == "" || p.weg.vermerke == nil {
		return Planlauf{}, fehler
	}

	if abgewiesen := p.weg.PlanPruefen(); abgewiesen != nil {
		return Planlauf{}, abgewiesen
	}

	frei, belegt, fehler := p.sperre()
	if fehler != nil || belegt {
		return Planlauf{Belegt: belegt}, fehler
	}

	defer frei()

	// **Erst, ob es den Ordner gibt — auch, wenn der Plan nicht fällig wäre**
	// (Nachlauf über Phase 4, Befund 3): ein abgezogener Stick sagt es, statt
	// hinter dem Vermerk eines früheren Ordners still zu bleiben.
	if lage, fehlt := os.Stat(ordner); fehlt != nil || !lage.IsDir() {
		return Planlauf{}, domain.MitGrund(domain.ErrNotFound, errors.Join(fehlt, fmt.Errorf("%s ist kein Ordner", ordner)),
			"der Ordner des Plans ist nicht erreichbar — stecke den Datenträger an oder wähle einen anderen Ordner")
	}

	// **Unter der Sperre gefragt**: wer eben vor diesem Lauf sicherte, hat
	// den Plan schon erfüllt.
	letzter, bekannt, fehler := p.weg.Planvermerk(ctx, ordner)
	if fehler != nil {
		return Planlauf{}, fehler
	}

	// **Und die letzte muss noch taugen** (Befund 3): eine gleich groß
	// beschädigte hielten Name und Größe für frisch. Gerechnet wird nur
	// über diese eine Datei, und nur, wenn der Plan sonst warten würde.
	if !domain.SicherungsplanFaellig(letzter, bekannt, p.weg.uhr()) {
		if gleich, _ := gleichGeblieben(filepath.Join(ordner, letzter.Datei), letzter); gleich { //nolint:errcheck // unlesbar heißt: sichern
			return Planlauf{}, nil
		}
	}

	name := Sicherungsplanname(p.weg.uhr())
	wunsch := Sicherungswunsch{Flaeche: p.flaeche, Anlass: domain.SicherungNachPlan}

	if gescheitert := p.weg.Sichern(MitSicherungswunsch(ctx, wunsch), filepath.Join(ordner, name)); gescheitert != nil {
		return Planlauf{}, gescheitert
	}

	entfernt, fehler := p.ausduennen(ctx, ordner)

	// **Und die halben Archive früherer Läufe** — erst nach einer gelungenen
	// Sicherung, wie das Ausdünnen.
	_, halb := halbeArchiveRaeumen(ordner, p.weg.uhr())

	return Planlauf{Gesichert: name, Entfernt: entfernt}, errors.Join(fehler, halb)
}

// ausduennen nimmt die eigenen Sicherungen jenseits der sieben jüngsten weg.
func (p *Sicherungsplan) ausduennen(ctx context.Context, ordner string) ([]string, error) {
	vermerke, fehler := p.weg.vermerke.Vermerke(ctx, domain.SicherungNachPlan)
	if fehler != nil {
		return nil, fehler
	}

	eintraege, fehler := os.ReadDir(ordner)
	if fehler != nil {
		return nil, fmt.Errorf("den Ordner des Plans lesen: %w", fehler)
	}

	var (
		gefunden []domain.Sicherungsvermerk
		gelesen  error
	)

	for _, eintrag := range eintraege {
		vermerk, eigen, lesefehler := wiedererkannt(ordner, eintrag, vermerke)
		gelesen = errors.Join(gelesen, lesefehler)

		if eigen {
			gefunden = append(gefunden, vermerk)
		}
	}

	if nachDemErkennen != nil {
		nachDemErkennen()
	}

	weg := domain.SicherungsplanAusduennen(gefunden)

	// **Die Grundlage der Entscheidung noch einmal** (Nachlauf über Phase 4,
	// Befund 2): änderte sich eine der behaltenen seither, beruhte die Liste
	// auf ihr, und eine gültige ältere ginge für sie. Dann wird nichts
	// entfernt — der nächste Lauf entscheidet neu. **Zwischen dieser Prüfung
	// und dem Löschen bleibt ein Fenster**; es zu schließen hieße, jede
	// Änderung am Ordner mitzuschreiben.
	if fehler := behalteneUnveraendert(ordner, gefunden, weg); fehler != nil {
		return nil, errors.Join(gelesen, fehler)
	}

	var entfernt []string

	for _, vermerk := range weg {
		weg, fehler := wegnehmen(filepath.Join(ordner, vermerk.Datei), vermerk)
		if fehler != nil {
			return entfernt, errors.Join(gelesen, fehler)
		}

		if weg {
			entfernt = append(entfernt, vermerk.Datei)
		}
	}

	return entfernt, gelesen
}

// nachDemErkennen ist eine Naht für den Lauf — leer im Betrieb: zwischen
// Erkennen und Ausdünnen ändert sich sonst nichts, wenn ein Lauf es will.
var nachDemErkennen func()

// behalteneUnveraendert prüft die Sicherungen, die bleiben sollen, noch
// einmal gegen ihre Prüfsumme.
func behalteneUnveraendert(ordner string, gefunden, weg []domain.Sicherungsvermerk) error {
	for _, vermerk := range gefunden {
		if slices.Contains(weg, vermerk) {
			continue
		}

		if gleich, fehler := gleichGeblieben(filepath.Join(ordner, vermerk.Datei), vermerk); !gleich {
			return errors.Join(fehler, fmt.Errorf("%w: %s änderte sich während des Ausdünnens — nichts entfernt",
				domain.ErrConflict, vermerk.Datei))
		}
	}

	return nil
}

// wiedererkannt sucht den Vermerk eines Eintrags — **nach Name, Art, Größe
// und Prüfsumme**.
//
// **Die Prüfsumme gehört zur Erkennung und nicht erst vor das Löschen**
// (Zweitblick über Phase 4, Befund 2): eine jüngere Sicherung, die sich
// seither gleich groß geändert hat, belegte sonst einen der sieben Plätze,
// und eine gültige ältere ging dafür. Das kostet je Lauf einen Lauf über
// jede eigene Sicherung — einmal in der Woche, und es sind höchstens acht.
func wiedererkannt(
	ordner string, eintrag os.DirEntry, vermerke []domain.Sicherungsvermerk,
) (domain.Sicherungsvermerk, bool, error) {
	if !istPlanname(eintrag.Name()) || !eintrag.Type().IsRegular() {
		return domain.Sicherungsvermerk{}, false, nil
	}

	// **Verschwand sie seit dem Lesen des Ordners, erkennt sie niemand
	// wieder** — und was nicht erkannt ist, bleibt.
	lage, fehler := eintrag.Info()

	for _, vermerk := range vermerke {
		if fehler == nil && vermerk.Datei == eintrag.Name() && vermerk.Groesse == lage.Size() {
			gleich, lesefehler := gleichGeblieben(filepath.Join(ordner, eintrag.Name()), vermerk)

			return vermerk, gleich, lesefehler
		}
	}

	return domain.Sicherungsvermerk{}, false, nil
}

// gleichGeblieben sagt, ob eine Datei noch die vermerkte Prüfsumme trägt.
//
// **Unlesbar und verändert enden gleich: nicht wiedererkannt** — ein Fehler
// beim Lesen kommt nur mit, damit ihn jemand sieht.
func gleichGeblieben(pfad string, vermerk domain.Sicherungsvermerk) (bool, error) {
	datei, fehler := os.Open(pfad) // #nosec G304 -- ein Eintrag des Planordners, wiedererkannt am Vermerk
	if fehler != nil {
		return false, fmt.Errorf("eine Sicherung des Plans lesen: %w", fehler)
	}

	rechner := sha256.New()
	_, lesefehler := io.Copy(rechner, io.LimitReader(datei, vermerk.Groesse+1))

	if fehler := errors.Join(lesefehler, datei.Close()); fehler != nil ||
		hex.EncodeToString(rechner.Sum(nil)) != vermerk.Pruefsumme {
		return false, fehler
	}

	return true, nil
}

// wegnehmen löscht eine eigene Sicherung — **nur, wenn ihre Prüfsumme noch
// die vermerkte ist**. Erkannt war sie schon daran; zwischen Erkennen und
// Löschen kann sie sich geändert haben, und eine, die sich geändert hat,
// gehört nicht mehr diesem Plan.
func wegnehmen(pfad string, vermerk domain.Sicherungsvermerk) (bool, error) {
	if gleich, fehler := gleichGeblieben(pfad, vermerk); fehler != nil || !gleich {
		return false, fehler
	}

	if fehler := os.Remove(pfad); fehler != nil {
		return false, fmt.Errorf("eine alte Sicherung des Plans entfernen: %w", fehler)
	}

	return true, nil
}

// Sicherungsplanname ist der Name einer Sicherung des Plans — **in
// Weltzeit**, damit sich die Namen nach der Zeit ordnen, auch über eine
// Sommerzeit hinweg.
func Sicherungsplanname(zeit time.Time) string {
	return planvorsatz + zeit.UTC().Format(planzeit) + Sicherungsendung
}

// istPlanname sagt, ob ein Name einer des Plans sein kann.
func istPlanname(name string) bool {
	mitte, vorn := strings.CutPrefix(name, planvorsatz)
	mitte, hinten := strings.CutSuffix(mitte, Sicherungsendung)

	if !vorn || !hinten {
		return false
	}

	_, fehler := time.Parse(planzeit, mitte)

	return fehler == nil
}
