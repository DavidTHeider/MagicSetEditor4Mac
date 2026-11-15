package app

// **Zurück ist nicht eine Handlung, sondern vier** — Werkstück O7 von Phase
// 5.2 (#1431).
//
// Die Fassade übersetzt vier Fragen:
//
//  1. **Was kann überhaupt zurück?** — und was davon kann diese Anwendung
//     selbst.
//  2. **Was kostet eine bestimmte Sicherung?** — Zahl für Zahl, und was danach
//     unklar dastünde.
//  3. **Worüber steht noch nichts fest?** — die offenen Rückkehren.
//  4. **Ich habe nachgesehen.** — die Quittung.

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Ruecksprungfuehrer ist die Naht zum Dienst.
type Ruecksprungfuehrer interface {
	Ruecksprunglage(ctx context.Context, pfad string) (domain.Ruecksprunglage, error)
	OffeneRueckkehren(ctx context.Context) ([]domain.Wiederherstellungsepoche, error)
	Rueckkehren(ctx context.Context) ([]domain.Wiederherstellungsepoche, error)
	RueckkehrQuittieren(ctx context.Context, kennung int64) error
}

// RuecksprungartDTO ist eine Art samt dem, was ihr im Weg steht.
//
// **Die Marke und nicht der Satz** — vier Flächen bauen daraus ihren eigenen;
// dieselbe Entscheidung wie beim Betriebszustand (O9) und beim
// Verträglichkeitsurteil (O6).
type RuecksprungartDTO struct {
	Art string `json:"art"`

	// Hindernis ist leer, wenn diese Anwendung die Art selbst ausführt.
	Hindernis string `json:"hindernis"`
}

// RuecksprungartenAntwort trägt die vier Arten.
type RuecksprungartenAntwort struct {
	Arten  []RuecksprungartDTO `json:"arten"`
	Fehler *Fehler             `json:"fehler"`
}

// Ruecksprungarten sagt, was zurückgehen kann und was davon hier geht.
//
// **Ohne Bestand und ohne Datei.** Es ist eine Auskunft über diese Anwendung,
// nicht über einen Stand — und deshalb steht sie auch dann da, wenn nichts
// aufgeht.
func (a *App) Ruecksprungarten(context.Context) RuecksprungartenAntwort {
	antwort := RuecksprungartenAntwort{Arten: []RuecksprungartDTO{}}

	for _, art := range domain.Ruecksprungarten() {
		antwort.Arten = append(antwort.Arten, RuecksprungartDTO{
			Art:       string(art),
			Hindernis: domain.RuecksprungHindernis(art),
		})
	}

	return antwort
}

// VergleichszeileDTO ist ein Zählstand nebeneinander.
//
// **Der Verlust steht als eigene Zahl da**, obwohl er sich ausrechnen ließe:
// vier Flächen rechneten ihn sonst viermal, und die vierte anders (§10).
type VergleichszeileDTO struct {
	Gegenstand  string `json:"gegenstand"`
	InSicherung int    `json:"inSicherung"`
	Jetzt       int    `json:"jetzt"`
	Verlust     int    `json:"verlust"`
	Unbekannt   bool   `json:"unbekannt"`
}

// MengeDTO ist eine Stückzahl eines Gegenstands der Sammlung.
type MengeDTO struct {
	Schluessel string `json:"schluessel"`
	Zustand    string `json:"zustand"`
	Sprache    string `json:"sprache"`

	// Foil ist der **alte** Wahrheitswert, abgeleitet aus
	// [MengeDTO.Finish] und seit T2 abgekündigt (0119).
	//
	// **Er steht hier, obwohl die Wails-Naht innen ist** (0033, Regel 5):
	// diese DTO geht an **zwei** Stellen unverändert nach draußen —
	// `mse --json ruecksprung-lage` und `GET /ruecksprung/lage` reichen
	// [RuecksprunglageAntwort] durch. Dass eine innere Gestalt zugleich
	// die äußere ist, ist ein Befund und hat sein eigenes Issue; solange
	// er steht, wird die äußere Zusage hier gehalten.
	Foil bool `json:"foil"`

	// Finish ist die Ausführung: `nonfoil`, `foil` oder `etched`.
	Finish  string `json:"finish"`
	Menge   int    `json:"menge"`
	Name    string `json:"name"`
	Setcode string `json:"setcode"`
	Nummer  string `json:"nummer"`
}

// RuecksprunglageAntwort ist der Vergleich vor der Rückkehr.
type RuecksprunglageAntwort struct {
	Datenstand         string               `json:"datenstand"`
	Zeilen             []VergleichszeileDTO `json:"zeilen"`
	Unklar             []MengeDTO           `json:"unklar"`
	HandpruefungNoetig bool                 `json:"handpruefungNoetig"`
	Fehler             *Fehler              `json:"fehler"`
}

// Ruecksprunglage stellt eine Sicherung neben den laufenden Bestand — eine
// verschlüsselte mit ihrem Kennwort (OPS-06), sonst leer.
func (a *App) Ruecksprunglage(ctx context.Context, pfad, kennwort string) RuecksprunglageAntwort {
	ctx = a.imProfil(ctx)
	ctx = service.MitSicherungswunsch(ctx, service.Sicherungswunsch{Kennwort: kennwort})

	antwort := RuecksprunglageAntwort{
		Zeilen: []VergleichszeileDTO{}, Unklar: []MengeDTO{},
	}

	lage, fehler := a.ruecksprung.Ruecksprunglage(ctx, pfad)
	if fehler != nil {
		antwort.Fehler = Uebersetze(ctx, fehler)

		return antwort
	}

	antwort.Datenstand = alsZeitpunkt(lage.Datenstand)
	antwort.HandpruefungNoetig = lage.HandpruefungNoetig

	for _, zeile := range lage.Zeilen {
		antwort.Zeilen = append(antwort.Zeilen, VergleichszeileDTO{
			Gegenstand:  zeile.Gegenstand,
			InSicherung: zeile.InSicherung,
			Jetzt:       zeile.Jetzt,
			Verlust:     zeile.Verlust(),
			Unbekannt:   zeile.Unbekannt(),
		})
	}

	antwort.Unklar = mengenDTO(lage.Unklar)

	return antwort
}

// RueckkehrDTO ist eine verbuchte Rückkehr.
type RueckkehrDTO struct {
	ID              int64      `json:"id"`
	Zurueckgespielt string     `json:"zurueckgespielt"`
	Datenstand      string     `json:"datenstand"`
	Quelle          string     `json:"quelle"`
	Klaerung        string     `json:"klaerung"`
	GeklaertAm      string     `json:"geklaertAm"`
	Mengen          []MengeDTO `json:"mengen"`
}

// RueckkehrenAntwort trägt die Rückkehren.
type RueckkehrenAntwort struct {
	Rueckkehren []RueckkehrDTO `json:"rueckkehren"`
	Fehler      *Fehler        `json:"fehler"`
}

// OffeneRueckkehren gibt die Rückkehren, über die noch nichts feststeht.
func (a *App) OffeneRueckkehren(ctx context.Context) RueckkehrenAntwort {
	ctx = a.imProfil(ctx)

	return a.rueckkehren(ctx, a.ruecksprung.OffeneRueckkehren)
}

// Rueckkehren gibt jede Rückkehr — auch die quittierten.
//
// **Neben [App.OffeneRueckkehren] und nicht an ihrer Stelle**: die Sperre
// hängt an den offenen, der **Nachweis** an allen. Wer nur die offenen hätte,
// könnte nach der Quittung nicht mehr sagen, dass es eine Rückkehr gab.
func (a *App) Rueckkehren(ctx context.Context) RueckkehrenAntwort {
	ctx = a.imProfil(ctx)

	return a.rueckkehren(ctx, a.ruecksprung.Rueckkehren)
}

// rueckkehren übersetzt eine Liste von Epochen.
func (a *App) rueckkehren(
	ctx context.Context,
	lesen func(context.Context) ([]domain.Wiederherstellungsepoche, error),
) RueckkehrenAntwort {
	ctx = a.imProfil(ctx)

	antwort := RueckkehrenAntwort{Rueckkehren: []RueckkehrDTO{}}

	epochen, fehler := lesen(ctx)
	if fehler != nil {
		antwort.Fehler = Uebersetze(ctx, fehler)

		return antwort
	}

	for _, epoche := range epochen {
		antwort.Rueckkehren = append(antwort.Rueckkehren, RueckkehrDTO{
			ID:              epoche.ID,
			Zurueckgespielt: alsZeitpunkt(epoche.Zurueckgespielt),
			Datenstand:      alsZeitpunkt(epoche.Datenstand),
			Quelle:          epoche.Quelle,
			Klaerung:        string(epoche.Klaerung),
			GeklaertAm:      alsZeitpunkt(epoche.GeklaertAm),
			Mengen:          mengenDTO(epoche.Mengen),
		})
	}

	return antwort
}

// RueckkehrQuittieren erklärt eine Rückkehr für nachgesehen.
func (a *App) RueckkehrQuittieren(ctx context.Context, kennung int64) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return SchreibAntwort{
		Fehler: Uebersetze(ctx, a.ruecksprung.RueckkehrQuittieren(ctx, kennung)),
	}
}

// MitRuecksprung setzt die Naht.
func (a *App) MitRuecksprung(fuehrer Ruecksprungfuehrer) { a.ruecksprung = fuehrer }

// mengenDTO übersetzt die Mengen.
func mengenDTO(mengen []domain.Sammlungsmenge) []MengeDTO {
	uebersetzt := make([]MengeDTO, 0, len(mengen))

	for _, menge := range mengen {
		uebersetzt = append(uebersetzt, MengeDTO{
			Schluessel: menge.Schluessel,
			Zustand:    menge.Zustand,
			Sprache:    menge.Sprache,
			Foil:       menge.Ausfuehrung() == domain.FinishFoil,
			Finish:     string(menge.Ausfuehrung()),
			Menge:      menge.Menge,
			Name:       menge.Name,
			Setcode:    menge.Setcode,
			Nummer:     menge.Nummer,
		})
	}

	return uebersetzt
}

// alsZeitpunkt gibt eine Zeit als Text — **leer bleibt leer**.
//
// **Die Nullzeit ist kein Zeitpunkt.** Sie als `0001-01-01T00:00:00Z` an die
// Oberfläche zu geben hieße, aus „steht nicht fest" ein Datum zu machen, das
// jemand liest.
func alsZeitpunkt(wann time.Time) string {
	if wann.IsZero() {
		return ""
	}

	return wann.UTC().Format(time.RFC3339)
}

// Ruecksprungarten ist die Fensterbindung für die vier Arten.
func (b *Bindungen) Ruecksprungarten(ctx context.Context) RuecksprungartenAntwort {
	return b.app(ctx).Ruecksprungarten(b.ruf(ctx))
}

// Ruecksprunglage ist die Fensterbindung für den Vergleich.
func (b *Bindungen) Ruecksprunglage(ctx context.Context, pfad, kennwort string) RuecksprunglageAntwort {
	return b.app(ctx).Ruecksprunglage(b.ruf(ctx), pfad, kennwort)
}

// Rueckkehren ist die Fensterbindung für die verbuchten Rückkehren.
//
// **Alle und nicht nur die offenen** — dieselbe Antwort wie an den drei
// Flächen daneben (§16). Der erste Anlauf gab hier die offenen und dort alle:
// dieselbe Fähigkeit unter demselben Namen mit zwei Bedeutungen, und genau das
// findet der Flächenlauf nicht, weil er die Gestalt prüft und nicht den Sinn.
// Welche davon oben stehen, entscheidet die Seite.
func (b *Bindungen) Rueckkehren(ctx context.Context) RueckkehrenAntwort {
	return b.app(ctx).Rueckkehren(b.ruf(ctx))
}

// RueckkehrQuittieren ist die Fensterbindung für die Quittung.
func (b *Bindungen) RueckkehrQuittieren(ctx context.Context, kennung int64) SchreibAntwort {
	return b.app(ctx).RueckkehrQuittieren(b.ruf(ctx), kennung)
}
