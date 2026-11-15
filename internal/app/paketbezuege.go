package app

// **Die Stände der bezogenen Bäume** — G04
// (#2251).
//
// Übersetzt wird hier und nicht entschieden (§2): was „neuer" heißt und wann
// eine Bestätigung nötig ist, steht in `service.Paketaktualisierung`.

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Paketaktualisierer ist die Naht zu `service.Paketaktualisierung`.
type Paketaktualisierer interface {
	Staende(ctx context.Context, fragen bool) ([]service.Bezugsstand, error)
	Vorschau(ctx context.Context, name, stand string) (service.Aktualisierungsvorschau, error)
	Aktualisieren(ctx context.Context, name, stand string, bestaetigt bool) (*service.Bezugsergebnis, error)
	Zurueck(ctx context.Context, name string) (*service.Rueckweg, error)
}

// PaketbezugDTO ist ein bezogener Baum, wie das Fenster ihn sieht.
type PaketbezugDTO struct {
	Name    string `json:"name"`
	Adresse string `json:"adresse"`

	// Stand ist der bezogene Commit, Geprueft der, gegen den dieses Projekt
	// gelaufen ist — leer, wo die Quelle nicht in der Liste steht.
	Stand    string `json:"stand"`
	Geprueft string `json:"geprueft"`

	Pakete       int    `json:"pakete"`
	Vorige       bool   `json:"vorige"`
	VorigerStand string `json:"vorigerStand"`

	// Lage ist `ungefragt`, `aktuell`, `neuer`, `abweichend` oder
	// `unbekannt`.
	Lage string `json:"lage"`

	// Spitze ist der Commit, auf den der Hauptzweig zeigt, SpitzeZeit sein
	// Datum (RFC 3339) — beides leer, solange nicht gefragt wurde.
	Spitze     string `json:"spitze"`
	SpitzeZeit string `json:"spitzeZeit"`
	Vorsprung  int    `json:"vorsprung"`

	// Fehler ist der Grund, wenn die Lage `unbekannt` ist.
	Fehler *Fehler `json:"fehler"`
}

// PaketbezuegeAntwort sind die bezogenen Bäume.
type PaketbezuegeAntwort struct {
	Bezuege []PaketbezugDTO `json:"bezuege"`
	Fehler  *Fehler         `json:"fehler"`
}

// PaketaenderungDTO ist ein Paket, das eine Aktualisierung berührt.
type PaketaenderungDTO struct {
	Paket     string `json:"paket"`
	Neu       int    `json:"neu"`
	Geaendert int    `json:"geaendert"`
	Entfernt  int    `json:"entfernt"`
}

// BezugsvorschauAntwort ist, was eine Aktualisierung täte.
type BezugsvorschauAntwort struct {
	Name       string              `json:"name"`
	Adresse    string              `json:"adresse"`
	Von        string              `json:"von"`
	Zu         string              `json:"zu"`
	Lage       string              `json:"lage"`
	Vorsprung  int                 `json:"vorsprung"`
	Rueckstand int                 `json:"rueckstand"`
	Pakete     []PaketaenderungDTO `json:"pakete"`
	Uebrige    int                 `json:"uebrige"`
	Gekuerzt   bool                `json:"gekuerzt"`
	Sets       []string            `json:"sets"`

	// BestaetigungNoetig sagt, dass ohne `bestaetigt` nicht aktualisiert
	// wird; Gruende nennt die Schlüssel dafür (`setVerliert`,
	// `nichtVoraus`, `gekuerzt`).
	BestaetigungNoetig bool     `json:"bestaetigungNoetig"`
	Gruende            []string `json:"gruende"`
	VorigeErsetzt      bool     `json:"vorigeErsetzt"`

	Fehler *Fehler `json:"fehler"`
}

// MitPaketaktualisierer setzt die Naht zu den Bezugsständen — wie
// [App.MitPaketbezieher], und aus demselben Grund: wer fragt, fragt das Netz.
func (a *App) MitPaketaktualisierer(aktualisierer Paketaktualisierer) { a.paketbezuege = aktualisierer }

// PaketBezuege nennt die bezogenen Bäume — und fragt ihre Quellen nur, wenn
// `fragen` gesetzt ist.
func (a *App) PaketBezuege(ctx context.Context, fragen bool) PaketbezuegeAntwort {
	ctx = a.imProfil(ctx)

	staende, err := a.paketbezuege.Staende(ctx, fragen)
	if err != nil {
		return PaketbezuegeAntwort{Bezuege: []PaketbezugDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	return PaketbezuegeAntwort{Bezuege: BezuegeAlsDTO(ctx, staende)}
}

// PaketBezugVorschau sagt, was eine Aktualisierung auf `stand` täte.
func (a *App) PaketBezugVorschau(ctx context.Context, name, stand string) BezugsvorschauAntwort {
	ctx = a.imProfil(ctx)

	vorschau, err := a.paketbezuege.Vorschau(ctx, name, stand)
	if err != nil {
		return BezugsvorschauAntwort{Pakete: []PaketaenderungDTO{}, Sets: []string{}, Gruende: []string{}, Fehler: Uebersetze(ctx, err)}
	}

	return BezugsvorschauAlsDTO(vorschau)
}

// PaketBezugAktualisieren bezieht `stand` an die Stelle des Baums `name`.
func (a *App) PaketBezugAktualisieren(ctx context.Context, name, stand string, bestaetigt bool) BezugAntwort {
	ctx = a.imProfil(ctx)

	ergebnis, err := a.paketbezuege.Aktualisieren(ctx, name, stand, bestaetigt)
	if err != nil {
		return BezugAntwort{Uebergangen: []string{}, Fehler: Uebersetze(ctx, err)}
	}

	return a.bezugVermerkt(ctx, ergebnis.Adresse, *ergebnis)
}

// PaketBezugZurueck tauscht den Baum `name` mit seinem aufgehobenen Stand.
func (a *App) PaketBezugZurueck(ctx context.Context, name string) BezugAntwort {
	ctx = a.imProfil(ctx)

	rueckweg, err := a.paketbezuege.Zurueck(ctx, name)
	if err != nil {
		return BezugAntwort{Uebergangen: []string{}, Fehler: Uebersetze(ctx, err)}
	}

	return a.bezugVermerkt(ctx, rueckweg.Adresse, service.Bezugsergebnis{
		Ort: rueckweg.Ort, Stand: rueckweg.Stand, Pruefsumme: rueckweg.Pruefsumme,
	})
}

// bezugVermerkt trägt Herkunft und Prüfsumme in den Katalog ein.
//
// **Ein Fehler dabei nimmt den Tausch nicht zurück** (§7, Regel 3) —
// derselbe Weg wie nach [App.PaketBeziehen].
func (a *App) bezugVermerkt(ctx context.Context, adresse string, ergebnis service.Bezugsergebnis) BezugAntwort {
	antwort := BezugAntwort{
		Ort: ergebnis.Ort, Stand: ergebnis.Stand, Pruefsumme: ergebnis.Pruefsumme,
		Uebergangen: append([]string{}, ergebnis.Uebergangen...),
	}

	if _, err := a.pakete.Vermerken(ctx, service.Vermerk{
		Ort: ergebnis.Ort, Herkunft: herkunftstext(adresse, ergebnis.Stand), Pruefsumme: ergebnis.Pruefsumme,
	}); err != nil {
		antwort.Fehler = Uebersetze(ctx, err)
	}

	return antwort
}

// BezuegeAlsDTO übersetzt die Stände — für jede Fläche dieselbe Gestalt.
func BezuegeAlsDTO(ctx context.Context, staende []service.Bezugsstand) []PaketbezugDTO {
	bezuege := make([]PaketbezugDTO, 0, len(staende))

	for _, stand := range staende {
		dto := PaketbezugDTO{
			Name: stand.Name, Adresse: stand.Adresse, Stand: stand.Stand, Geprueft: stand.Geprueft,
			Pakete: stand.Pakete, Vorige: stand.Vorige, VorigerStand: stand.VorigerStand,
			Lage: string(stand.Lage), Spitze: stand.Spitze.Commit, Vorsprung: stand.Vorsprung,
		}

		if !stand.Spitze.Zeit.IsZero() {
			dto.SpitzeZeit = stand.Spitze.Zeit.UTC().Format(time.RFC3339)
		}

		if stand.Fehler != nil {
			dto.Fehler = Uebersetze(ctx, stand.Fehler)
		}

		bezuege = append(bezuege, dto)
	}

	return bezuege
}

// BezugsvorschauAlsDTO übersetzt eine Vorschau.
func BezugsvorschauAlsDTO(vorschau service.Aktualisierungsvorschau) BezugsvorschauAntwort {
	antwort := BezugsvorschauAntwort{
		Name: vorschau.Name, Adresse: vorschau.Adresse, Von: vorschau.Von, Zu: vorschau.Zu,
		Lage: vorschau.Lage, Vorsprung: vorschau.Vorsprung, Rueckstand: vorschau.Rueckstand,
		Pakete: make([]PaketaenderungDTO, 0, len(vorschau.Pakete)), Uebrige: vorschau.Uebrige,
		Gekuerzt: vorschau.Gekuerzt, Sets: append([]string{}, vorschau.Sets...),
		BestaetigungNoetig: vorschau.BestaetigungNoetig, Gruende: append([]string{}, vorschau.Gruende...),
		VorigeErsetzt: vorschau.VorigeErsetzt,
	}

	for _, paket := range vorschau.Pakete {
		antwort.Pakete = append(antwort.Pakete, PaketaenderungDTO{
			Paket: paket.Paket, Neu: paket.Neu, Geaendert: paket.Geaendert, Entfernt: paket.Entfernt,
		})
	}

	return antwort
}
