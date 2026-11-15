package app

// **Zurücksetzen benennt genau, was betroffen ist** — Werkstück O4 von Phase
// 5.2 (#1428).
//
// # Zwei Rufe und kein dritter
//
// Nachsehen und zurücksetzen. Der zweite trägt, was der erste gezeigt hat:
// den **Namen** als Bestätigung und — für die ganze Anwendung — den
// **Nachweis**. Ein `bestaetigt: true` wäre keins von beidem.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Ruecksetzungsfuehrer ist die Naht zum Dienst.
type Ruecksetzungsfuehrer interface {
	Vorschau(ctx context.Context, art domain.Ruecksetzungsart,
		ziel service.Ruecksetzungsziel) (domain.Ruecksetzungsvorschau, error)
	Zuruecksetzen(ctx context.Context, art domain.Ruecksetzungsart,
		ziel service.Ruecksetzungsziel, nachweis, revision string) (
		service.Ruecksetzungsbilanz, error)
}

// BetroffenesDTO ist eine Gruppe, die eine Rücksetzung anfasst.
type BetroffenesDTO struct {
	// Was ist eine **Marke** — vier Flächen bauen daraus ihren eigenen Satz.
	Was string `json:"was"`

	Anzahl int `json:"anzahl"`

	// Namen nennen ein paar davon — höchstens fünf.
	Namen []string `json:"namen"`
}

// RuecksetzungsvorschauAntwort ist, was eine Rücksetzung anfassen würde.
type RuecksetzungsvorschauAntwort struct {
	Art string `json:"art"`

	Betroffen   []BetroffenesDTO `json:"betroffen"`
	Hindernisse []HindernisDTO   `json:"hindernisse"`

	// Bereit sagt, ob es so, wie es steht, ginge.
	//
	// **Gerechnet und nicht der Oberfläche überlassen**: „keine Hindernisse"
	// ist dieselbe Aussage, und zwei Stellen, die sie bilden, laufen
	// auseinander.
	Bereit bool `json:"bereit"`

	// NameNoetig sagt, dass der Name des Ziels als Bestätigung dazugehört.
	NameNoetig bool `json:"nameNoetig"`

	NachweisNoetig     bool `json:"nachweisNoetig"`
	SicherungEmpfohlen bool `json:"sicherungEmpfohlen"`

	// Revision ist die Marke über das, was hier steht — sie geht beim
	// Zurücksetzen zurück und bindet die Ausführung an diese Vorschau
	// (Befund 12 des Zweitblicks über Phase 5.2).
	Revision string `json:"revision"`

	Fehler *Fehler `json:"fehler"`
}

// RuecksetzungsbilanzAntwort ist, was aus einer Rücksetzung wurde.
type RuecksetzungsbilanzAntwort struct {
	Art string `json:"art"`

	Zurueckgesetzt int `json:"zurueckgesetzt"`

	// Beiseite ist der Ort des alten Bestands — **nur beim Neuanfang**, und
	// dort ist er die Auskunft, die aus einer Löschung einen Tausch macht.
	Beiseite string `json:"beiseite"`

	Fehler *Fehler `json:"fehler"`

	// Bestandsepoche ist die Epoche danach (J05, #2269) — siehe
	// [Rueckspielantwort.Bestandsepoche].
	Bestandsepoche uint64 `json:"bestandsepoche"`
}

// Ruecksetzungsarten nennt die sechs — **die Reihe ist die Schwere**.
func (a *App) Ruecksetzungsarten() []string {
	arten := make([]string, 0, len(domain.Ruecksetzungsarten()))
	for _, art := range domain.Ruecksetzungsarten() {
		arten = append(arten, string(art))
	}

	return arten
}

// Ruecksetzungsvorschau sieht nach — **und ändert nichts**.
func (a *App) Ruecksetzungsvorschau(
	ctx context.Context, art, profilID, nutzerID string,
) RuecksetzungsvorschauAntwort {
	ctx = a.imProfil(ctx)

	antwort := RuecksetzungsvorschauAntwort{
		Art: art, Betroffen: []BetroffenesDTO{}, Hindernisse: []HindernisDTO{},
	}

	ziel, befund := ruecksetzungsziel(profilID, nutzerID, "")
	if befund != nil {
		antwort.Fehler = befund

		return antwort
	}

	vorschau, fehler := a.ruecksetzung.Vorschau(ctx, domain.Ruecksetzungsart(art), ziel)
	if fehler != nil {
		antwort.Fehler = Uebersetze(ctx, fehler)

		return antwort
	}

	return vorschauAntwort(vorschau)
}

// Zuruecksetzen führt sie aus.
//
// **Der Name ist die Bestätigung**, der Nachweis die Freigabe für die ganze
// Anwendung — leer, wo keiner verlangt wird.
func (a *App) Zuruecksetzen(
	ctx context.Context, art, profilID, nutzerID, name, nachweis, revision string,
) RuecksetzungsbilanzAntwort {
	ctx = a.imProfil(ctx)

	antwort := RuecksetzungsbilanzAntwort{Art: art}

	ziel, befund := ruecksetzungsziel(profilID, nutzerID, name)
	if befund != nil {
		antwort.Fehler = befund

		return antwort
	}

	bilanz, fehler := a.ruecksetzung.Zuruecksetzen(
		ctx, domain.Ruecksetzungsart(art), ziel, nachweis, revision)

	// **Die Bilanz kommt auch mit einem Fehler zurück** (§7, Regel 3): was
	// zurückgesetzt ist, bleibt es, und der Ort des beiseitegelegten Bestands
	// ist die Auskunft, ohne die niemand ihn wiederfände.
	antwort.Zurueckgesetzt = bilanz.Zurueckgesetzt
	antwort.Beiseite = bilanz.Beiseite
	antwort.Fehler = Uebersetze(ctx, fehler)

	// **Der Zeiger kann gewandert sein** — dieselbe Nacharbeit wie bei
	// [App.ProfilLoeschen], und aus demselben Grund: der Fachkern stellt ihn
	// um, wenn das gelöschte Profil das gezeigte war.
	if fehler == nil && domain.Ruecksetzungsart(art).LoeschtDaten() {
		a.profilNachziehen(ctx)
	}

	// **Gezählt hat der Tausch, nicht diese Antwort** (Befund 3 des Nachlaufs
	// in J05, #2269): ein zweites Hochzählen außerhalb der Sperre verwarf
	// eine Kennung, die der Finder zwischen Tausch und Antwort schon im
	// **neuen** Bestand eingelesen hatte. Eine Rücksetzung, die nur löscht,
	// zählt nicht — dass ihre Kennungen danach nichts mehr meinen, trägt
	// #2306.
	antwort.Bestandsepoche = a.lauf.epoche.Load()

	return antwort
}

// ruecksetzungsziel liest die beiden Kennungen — **leer heißt „meines"**.
func ruecksetzungsziel(profilID, nutzerID, name string) (service.Ruecksetzungsziel, *Fehler) {
	ziel := service.Ruecksetzungsziel{Name: name}

	if profilID != "" {
		kennung, befund := benannteKennungAus("profilID", profilID)
		if befund != nil {
			return ziel, befund
		}

		ziel.ProfilID = kennung
	}

	if nutzerID != "" {
		kennung, befund := benannteKennungAus("nutzerID", nutzerID)
		if befund != nil {
			return ziel, befund
		}

		ziel.NutzerID = kennung
	}

	return ziel, nil
}

// vorschauAntwort übersetzt die Vorschau.
func vorschauAntwort(vorschau domain.Ruecksetzungsvorschau) RuecksetzungsvorschauAntwort {
	antwort := RuecksetzungsvorschauAntwort{
		Art:                string(vorschau.Art),
		Betroffen:          make([]BetroffenesDTO, 0, len(vorschau.Betroffen)),
		Hindernisse:        make([]HindernisDTO, 0, len(vorschau.Hindernisse)),
		Bereit:             vorschau.Bereit(),
		NameNoetig:         vorschau.NameNoetig,
		NachweisNoetig:     vorschau.NachweisNoetig,
		SicherungEmpfohlen: vorschau.SicherungEmpfohlen,
		Revision:           vorschau.Revision,
	}

	for _, gruppe := range vorschau.Betroffen {
		namen := gruppe.Namen
		if namen == nil {
			// **Eine leere Liste und kein `null`** (§3): „keine Namen" ist
			// eine Aussage, `null` ist keine.
			namen = []string{}
		}

		antwort.Betroffen = append(antwort.Betroffen, BetroffenesDTO{
			Was: gruppe.Was, Anzahl: gruppe.Anzahl, Namen: namen,
		})
	}

	for _, hindernis := range vorschau.Hindernisse {
		antwort.Hindernisse = append(antwort.Hindernisse, HindernisDTO{
			Art: string(hindernis.Art), Stelle: hindernis.Stelle,
		})
	}

	return antwort
}

// MitRuecksetzung setzt die Naht.
func (a *App) MitRuecksetzung(fuehrer Ruecksetzungsfuehrer) { a.ruecksetzung = fuehrer }

// Ruecksetzungsarten ist die Fensterbindung für die Reihe.
func (b *Bindungen) Ruecksetzungsarten(ctx context.Context) []string {
	return b.app(ctx).Ruecksetzungsarten()
}

// Ruecksetzungsvorschau ist die Fensterbindung für die Vorschau.
func (b *Bindungen) Ruecksetzungsvorschau(
	ctx context.Context,
	art, profilID, nutzerID string,
) RuecksetzungsvorschauAntwort {
	return b.app(ctx).Ruecksetzungsvorschau(b.ruf(ctx), art, profilID, nutzerID)
}

// Zuruecksetzen ist die Fensterbindung für die Ausführung.
func (b *Bindungen) Zuruecksetzen(
	ctx context.Context,
	art, profilID, nutzerID, name, nachweis, revision string,
) RuecksetzungsbilanzAntwort {
	return b.app(ctx).Zuruecksetzen(b.ruf(ctx), art, profilID, nutzerID, name,
		nachweis, revision)
}
