package app

// **Bestandsprobleme werden erklärt und nur nach überprüfbarer Vorschau
// korrigiert** — Werkstück O8 von Phase 5.2
// (#1432).

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Pflegefuehrer ist die Naht zum Dienst.
type Pflegefuehrer interface {
	Vorschau(ctx context.Context) (domain.Massnahmenplan, error)
	Pflegen(ctx context.Context, erstellt time.Time, wahl []domain.Pflegewahl) (
		service.Pflegebilanz, error)
}

// PflegebefundDTO ist eine Auffälligkeit im Bestand.
type PflegebefundDTO struct {
	// Art und Klasse sind **Marken** — vier Flächen bauen daraus ihren
	// eigenen Satz.
	Art    string `json:"art"`
	Klasse string `json:"klasse"`

	Anzahl int `json:"anzahl"`

	// Revision ist die Prüfsumme über die betroffenen Zeilen. **Sie reist mit
	// zurück**: wer sie mitschickt, sagt, welche Zeilen er gesehen hat.
	Revision string `json:"revision"`

	Beispiele []string `json:"beispiele"`
}

// PflegevorschauAntwort ist die Vorschau samt ihrer Frist.
type PflegevorschauAntwort struct {
	Erstellt   string            `json:"erstellt"`
	Gueltigbis string            `json:"gueltigbis"`
	Befunde    []PflegebefundDTO `json:"befunde"`
	Fehler     *Fehler           `json:"fehler"`
}

// Pflegevorschau sieht nach — **und ändert nichts**.
func (a *App) Pflegevorschau(ctx context.Context) PflegevorschauAntwort {
	ctx = a.imProfil(ctx)

	antwort := PflegevorschauAntwort{Befunde: []PflegebefundDTO{}}

	plan, fehler := a.pflege.Vorschau(ctx)
	if fehler != nil {
		antwort.Fehler = Uebersetze(ctx, fehler)

		return antwort
	}

	antwort.Erstellt = alsZeitpunkt(plan.Erstellt)
	antwort.Gueltigbis = alsZeitpunkt(plan.Gueltigbis)
	antwort.Befunde = befundeDTO(plan.Befunde)

	return antwort
}

// PflegewahlDTO ist eine gewählte Maßnahme samt gesehener Revision.
type PflegewahlDTO struct {
	Art      string `json:"art"`
	Revision string `json:"revision"`
}

// PflegezeileDTO ist eine Maßnahme mit ihrem Ausgang.
type PflegezeileDTO struct {
	Art     string `json:"art"`
	Ausgang string `json:"ausgang"`
	Anzahl  int    `json:"anzahl"`
	Grund   string `json:"grund"`
}

// PflegebilanzAntwort ist, was aus einer Pflege wurde.
type PflegebilanzAntwort struct {
	// Gesichert ist der Ort der Sicherung, die vor dem Schreiben entstand —
	// leer, wenn nichts zu tun war.
	Gesichert string `json:"gesichert"`

	Zeilen []PflegezeileDTO `json:"zeilen"`

	// Vorher und Nachher sind die Befunde vor und nach der Pflege.
	//
	// **Beide, denn „drei erledigt" ist keine Auskunft über den Bestand.**
	Vorher  []PflegebefundDTO `json:"vorher"`
	Nachher []PflegebefundDTO `json:"nachher"`

	Fehler *Fehler `json:"fehler"`
}

// Pflegen führt die gewählten Maßnahmen aus.
//
// **`erstellt` kommt aus der Vorschau** und sagt, wann jemand hingesehen hat.
// Ein unlesbarer Zeitpunkt ist ein Fehler der Anfrage und keine frische Frist.
func (a *App) Pflegen(
	ctx context.Context, erstellt string, wahl []PflegewahlDTO,
) PflegebilanzAntwort {
	ctx = a.imProfil(ctx)

	antwort := PflegebilanzAntwort{
		Zeilen: []PflegezeileDTO{}, Vorher: []PflegebefundDTO{}, Nachher: []PflegebefundDTO{},
	}

	wann, zeitfehler := time.Parse(time.RFC3339, erstellt)
	if zeitfehler != nil {
		antwort.Fehler = Uebersetze(ctx, domain.MitGrund(domain.ErrValidation, zeitfehler,
			"Diese Anfrage nennt keinen brauchbaren Zeitpunkt der Vorschau."))

		return antwort
	}

	gewaehlt := make([]domain.Pflegewahl, 0, len(wahl))
	for _, eine := range wahl {
		gewaehlt = append(gewaehlt,
			domain.Pflegewahl{Art: eine.Art, Revision: eine.Revision})
	}

	bilanz, fehler := a.pflege.Pflegen(ctx, wann, gewaehlt)

	// **Die Bilanz kommt auch mit einem Fehler zurück** (§7, Regel 3):
	// gepflegt ist gepflegt, und ein Fehler beim Nachsehen darf die Zeilen
	// nicht verschlucken.
	antwort.Gesichert = bilanz.Gesichert
	antwort.Zeilen = pflegezeilenDTO(bilanz.Zeilen)
	antwort.Vorher = befundeDTO(bilanz.Vorher)
	antwort.Nachher = befundeDTO(bilanz.Nachher)
	antwort.Fehler = Uebersetze(ctx, fehler)

	return antwort
}

// MitPflege setzt die Naht.
func (a *App) MitPflege(fuehrer Pflegefuehrer) { a.pflege = fuehrer }

func befundeDTO(befunde []domain.Pflegebefund) []PflegebefundDTO {
	uebersetzt := make([]PflegebefundDTO, 0, len(befunde))

	for _, befund := range befunde {
		beispiele := befund.Beispiele
		if beispiele == nil {
			// **Eine leere Liste und kein `null`** (§3): „keine Beispiele"
			// ist eine Aussage, `null` ist keine.
			beispiele = []string{}
		}

		uebersetzt = append(uebersetzt, PflegebefundDTO{
			Art: befund.Art, Klasse: string(befund.Klasse), Anzahl: befund.Anzahl,
			Revision: befund.Revision, Beispiele: beispiele,
		})
	}

	return uebersetzt
}

// pflegezeilenDTO übersetzt die Bilanzzeilen.
//
// **Nicht `zeilenDTO`**: den Namen hat die Kartenliste (`listenspalten.go`).
func pflegezeilenDTO(zeilen []domain.Pflegezeile) []PflegezeileDTO {
	uebersetzt := make([]PflegezeileDTO, 0, len(zeilen))

	for _, zeile := range zeilen {
		uebersetzt = append(uebersetzt, PflegezeileDTO{
			Art: zeile.Art, Ausgang: string(zeile.Ausgang),
			Anzahl: zeile.Anzahl, Grund: zeile.Grund,
		})
	}

	return uebersetzt
}

// Pflegevorschau ist die Fensterbindung für die Vorschau.
func (b *Bindungen) Pflegevorschau(ctx context.Context) PflegevorschauAntwort {
	return b.app(ctx).Pflegevorschau(b.ruf(ctx))
}

// Pflegen ist die Fensterbindung für die Ausführung.
func (b *Bindungen) Pflegen(ctx context.Context, erstellt string, wahl []PflegewahlDTO) PflegebilanzAntwort {
	return b.app(ctx).Pflegen(b.ruf(ctx), erstellt, wahl)
}
