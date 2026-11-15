package app

// **Eine Partie für Gäste öffnen** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wozu es diesen Schritt gibt
//
// **Weil erreichbar sein und mitspielen lassen zwei Entscheidungen sind.** Der
// Horchposten aus N2 macht diesen Rechner ansprechbar; er sagt nicht, **was**
// ein Gegenüber dann tun darf. Ohne diesen Schritt liefe jede angenommene
// Sitzung ins Leere — und genau so war es bis hierher: sie stand in der
// Übersicht, und niemand las von ihr.
//
// # Warum nur am Fenster
//
// **Dieselben drei Gründe wie beim Horchposten** (N2, `docs/FAEHIGKEITEN.md`),
// und sie wiegen hier schwerer, weil es nicht um einen Port geht, sondern um
// den **Tisch**:
//
//   - **Kein MCP-Werkzeug**: eine Freigabe, die ein Modell erteilt, überlebt
//     seinen Ruf — und was danach über sie hereinkommt, ändert eine Partie
//     ([0106](../../docs/entscheidungen/0106-drei-vertrauensgrenzen-und-nur-eine-wird-gebaut.md)).
//   - **Kein REST-Endpunkt**: wer über einen Port anfragt, kann die
//     Zustimmung an **diesem** Rechner nicht geben
//     ([0115](../../docs/entscheidungen/0115-die-schleife-ist-eine-vorgabe.md)).
//   - **Kein CLI-Befehl**: er müsste bleiben, solange gespielt wird, und ein
//     Prozess, der bis ⌃C am Tisch sitzt, ist etwas anderes als ein Befehl,
//     der etwas tut und endet.
//
// # Und warum die Sitze hier genannt werden
//
// **Weil eine Sitzung einen Rechner belegt und keinen Sitz.** Wer sie
// gleichsetzte, ließe einen Gast im Namen eines anderen ziehen — der Grund
// steht ausführlich bei [service.Partiefreigabe].

import (
	"context"
	"fmt"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// freigabefelder ist, was eine Anwendung über die offene Partie hält.
type freigabefelder struct {
	schloss sync.Mutex

	// wirt ist der Sitz, für den **dieser** Rechner spielt.
	//
	// **Er steht neben der Freigabe und nicht darin**: jene vergibt
	// **Gast**sitze, dieser ist der eigene, und er wird nie vergeben.
	wirt string

	offen *service.Partiefreigabe
}

// SitzvergabeDTO ist ein Sitz und der Fingerabdruck seines Gegenübers.
//
// **Der Fingerabdruck und nicht die Kennung** — er ist die Gestalt, die ein
// Mensch vergleichen kann; die Kennung trägt den Schlüssel und ist nichts zum
// Ablesen (N1).
type SitzvergabeDTO struct {
	Sitz          string `json:"sitz"`
	Fingerabdruck string `json:"fingerabdruck"`
}

// PartiefreigabeAntwort sagt, welche Partie offensteht.
type PartiefreigabeAntwort struct {
	// Laeuft sagt, ob überhaupt eine offensteht.
	Laeuft bool `json:"laeuft"`

	Partie string `json:"partie"`
	Wirt   string `json:"wirt"`

	// Vergeben ist, wer worauf sitzt.
	Vergeben []SitzvergabeDTO `json:"vergeben"`

	Fehler *Fehler `json:"fehler"`
}

// PartieFreigeben öffnet eine Partie für Gäste.
//
// **Eine zur Zeit.** Zwei offene Partien hießen: eine hereinkommende Sitzung
// müsste sagen, welche sie meint — und das wäre wieder eine Angabe des
// Gastes. Wer eine zweite öffnen will, schließt die erste.
func (a *App) PartieFreigeben(
	ctx context.Context, kennung, fuer string, gastsitze []string,
) PartiefreigabeAntwort {
	ctx = a.imProfil(ctx)

	// **Die Marke vor dem Start** (Nachlauf über Phase 6.12, N2).
	marke := a.laufmarke(domain.LaufPartiefreigabe)

	// **`externEmpfangen` und nicht `lokalSchreiben`** — dieselbe
	// Wirkungsklasse wie der Horchposten (N7, #1568): von hier an ändert ein
	// fremder Rechner den Stand dieser Partie.
	if fehler := a.darf(ctx, domain.WirkungExternEmpfangen); fehler != nil {
		return leerePartiefreigabe(ctx, fehler)
	}

	// **Die Partie wird gelesen, bevor sie freigegeben wird.** Wer eine
	// erfundene Kennung freigäbe, öffnete den Tisch für eine Partie, die es
	// nicht gibt — und der erste Gast bekäme eine Absage, die wie ein
	// Netzfehler aussieht (§8).
	sicht, fehler := a.partienaht().Partie(ctx, kennung, fuer)
	if fehler != nil {
		return leerePartiefreigabe(ctx, fehler)
	}

	if len(gastsitze) == 0 {
		return leerePartiefreigabe(ctx, fmt.Errorf(
			"%w: ohne Gastsitz gibt es nichts freizugeben",
			domain.ErrValidation))
	}

	a.partie.freigabe.schloss.Lock()
	a.partie.freigabe.wirt = fuer
	// **Ein eingeladener Sitz geht an den Eingeladenen** (P9b, #2377).
	a.partie.freigabe.offen = service.NeuePartiefreigabe(kennung, gastsitze).
		MitEinladungen(service.Einladungen(sicht))
	a.partie.freigabe.schloss.Unlock()

	a.nachDemStart(ctx, domain.LaufPartiefreigabe, marke, a.laufende()[domain.LaufPartiefreigabe])

	// **Das Fenster erfährt es** (P9c, #2377): eine Partie mit Einladung
	// öffnet ihren Tisch selbst, bevor ihr Freigabefeld fragt.
	a.besetzungMelden(kennung)

	return a.PartieFreigabeStand(ctx)
}

// PartieFreigabeBeenden schließt sie wieder.
//
// **Die offenen Sitzungen bleiben.** Sie zu schließen ist Sache des
// Horchpostens und des Profilwechsels; was hier endet, ist die **Zusage**,
// dass ein neuer Gast einen Sitz bekommt.
func (a *App) PartieFreigabeBeenden(ctx context.Context) PartiefreigabeAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungSchreiben); fehler != nil {
		return leerePartiefreigabe(ctx, fehler)
	}

	a.partiefreigabeBeenden()

	return a.PartieFreigabeStand(ctx)
}

// partiefreigabeBeenden nimmt die Zusage zurück — der Weg hinter
// [App.PartieFreigabeBeenden] und hinter dem Schalter (OPS-10b, #1978).
func (a *App) partiefreigabeBeenden() {
	a.partie.freigabe.schloss.Lock()
	defer a.partie.freigabe.schloss.Unlock()

	a.partie.freigabe.wirt = ""
	a.partie.freigabe.offen = nil
}

// PartieFreigabeStand sagt, was gerade offensteht.
func (a *App) PartieFreigabeStand(ctx context.Context) PartiefreigabeAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungLesen); fehler != nil {
		return leerePartiefreigabe(ctx, fehler)
	}

	offen, wirt := a.offeneFreigabe()
	if offen == nil {
		return PartiefreigabeAntwort{Vergeben: []SitzvergabeDTO{}}
	}

	vergeben := offen.Vergeben()
	antwort := PartiefreigabeAntwort{
		Laeuft:   true,
		Partie:   offen.Partie(),
		Wirt:     wirt,
		Vergeben: make([]SitzvergabeDTO, 0, len(vergeben)),
	}

	for _, eine := range vergeben {
		antwort.Vergeben = append(antwort.Vergeben, SitzvergabeDTO{
			Sitz: eine.Sitz, Fingerabdruck: fingerabdruckVon(eine.Gegenueber),
		})
	}

	return antwort
}

// offeneFreigabe gibt die offene Partie und den eigenen Sitz.
func (a *App) offeneFreigabe() (offen *service.Partiefreigabe, wirt string) {
	a.partie.freigabe.schloss.Lock()
	defer a.partie.freigabe.schloss.Unlock()

	return a.partie.freigabe.offen, a.partie.freigabe.wirt
}

// leerePartiefreigabe gibt eine Antwort mit leerer Liste und übersetztem
// Fehler.
//
// **Nicht `leereFreigabe`** — die gibt es schon für die **Regel**freigabe
// (`regelfreigabe.go`), und zwei gleichnamige Helfer an derselben Naht wären
// zwei verschiedene Dinge unter einem Namen. Derselbe Fall wie bei
// `PartieverlaufAntwort` gegen `VerlaufAntwort`.
//
// **Eine leere Liste und kein `nil`** — dieselbe Zusage wie auf dem
// Erfolgsweg; genau hier ist sie in Phase 5.4 zweimal gebrochen.
func leerePartiefreigabe(ctx context.Context, fehler error) PartiefreigabeAntwort {
	return PartiefreigabeAntwort{
		Vergeben: []SitzvergabeDTO{},
		Fehler:   Uebersetze(ctx, fehler),
	}
}
