// **Einstellungen ausführen und einlesen** — Werkstück O3 von Phase 5.2
// (#1427),
// nach [Vermerk 0079](../../docs/entscheidungen/0079-einstellungsebenen-und-portabler-import.md).
//
// # Drei Schritte, und der mittlere ändert nichts
//
// **Ausfuhr → Vorschau → Übernahme.** Der mittlere Schritt ist der, für den es
// dieses Werkstück gibt: eine Datei von irgendwoher wird **gezeigt**, bevor
// sie wirkt — Zeile für Zeile, mit dem alten Wert daneben und der Auskunft,
// wessen Entscheidung der alte war. §16, Regel 6: nichts wird still
// übernommen.
//
// # Warum die Übernahme die Datei noch einmal liest
//
// **Weil ein Diff, den der Aufrufer geschickt hat, seine Behauptung ist.** Wer
// ihn übernähme, ließe die Fläche entscheiden, was in den Bestand kommt — und
// die Prüfung in `domain` wäre eine Höflichkeit. Was der Aufrufer wirklich
// beiträgt, ist die **Auswahl**: welche Zeilen er will, und gegen welchen
// Stand er sie gesehen hat.

package service

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Austauschablage schreibt beide Ebenen in einem Zug — implementiert in
// `storage`.
//
// **In einem Zug, weil ein halber Import schlimmer ist als keiner** (§8): wer
// die Sprache übernimmt und die Barrierefreiheit nicht, hinterlässt einen
// Stand, den niemand gewählt hat und den keine Vorschau je gezeigt hat.
type Austauschablage interface {
	// EinstellungenUebernehmen schreibt beide Ebenen — und prüft `erwartet`
	// **im selben Geschäft** noch einmal (Befund 11 des Zweitblicks über
	// Phase 5.2): die Prüfung hier oben sieht einen Stand von vorhin.
	EinstellungenUebernehmen(
		ctx context.Context, profil, nutzer map[string]string, erwartet string) error
}

// Austauschwache sagt, wer arbeitet und ob er darf.
//
// **Zwei Fragen an einer Naht**, weil sie aus einem Stand kommen müssen: wer
// den Akteur einmal und die Erlaubnis ein zweites Mal holte, könnte zwischen
// beiden einen Wechsel erwischen (§7, Regel 2). `*Autorisierung` erfüllt sie.
type Austauschwache interface {
	Pruefe(ctx context.Context, wirkung domain.Wirkung) error
	Akteur(ctx context.Context) (domain.Akteur, error)
}

// Einstellungsaustausch führt Einstellungen aus und ein.
type Einstellungsaustausch struct {
	ablage      Einstellungsablage
	nutzerwerte Nutzerwerteleser
	schreiber   Austauschablage
	wache       Austauschwache
}

// NeuerEinstellungsaustausch baut ihn über den vier Nähten.
//
// **Die Wache ist Pflicht und nicht `MitPruefer`** — anders als beim
// Setschreiber und der Bildausfuhr (0078). Dort ist ein fehlender Prüfer ein
// Dienst, der seine Sache prüft und nicht die Erlaubnis; hier wäre er ein
// **Import ohne Riegel**, und die Marke über den Stand käme aus dem Nichts.
// Was Pflicht sein kann, wird Pflicht.
func NeuerEinstellungsaustausch(
	ablage Einstellungsablage, nutzerwerte Nutzerwerteleser,
	schreiber Austauschablage, wache Austauschwache,
) *Einstellungsaustausch {
	return &Einstellungsaustausch{
		ablage: ablage, nutzerwerte: nutzerwerte,
		schreiber: schreiber, wache: wache,
	}
}

// lage ist der Stand beider Ebenen, wie er gesetzt ist.
type lage struct {
	profil, nutzer     map[string]string
	profilID, nutzerID int64
}

// Marke gibt die Marke über diesen Stand.
func (l lage) Marke() string {
	return domain.Standmarke(l.profilID, l.nutzerID, l.profil, l.nutzer)
}

// lesen holt beide Ebenen und den Akteur.
func (a *Einstellungsaustausch) lesen(ctx context.Context) (lage, error) {
	akteur, fehler := a.wache.Akteur(ctx)
	if fehler != nil {
		return lage{}, fehler
	}

	imProfil, fehler := a.ablage.Alle(ctx)
	if fehler != nil {
		return lage{}, fehler
	}

	beimMenschen, fehler := a.nutzerwerte.NutzerwerteZumProfil(ctx)
	if fehler != nil {
		return lage{}, fehler
	}

	return lage{
		profil: imProfil, nutzer: beimMenschen,
		profilID: akteur.ProfilID, nutzerID: akteur.NutzerID,
	}, nil
}

// Ausfuhr gibt die gewählten Abschnitte — ohne Wahl das Bündel.
//
// **Das Bündel ist `nutzer` und `profil`, nicht alles.** Der Arbeitsplatz
// hängt an einem Bildschirm und reist nur, wer ihn nennt; die Ebene `geraet`
// reist gar nicht.
func (a *Einstellungsaustausch) Ausfuhr(
	ctx context.Context, abschnitte []string,
) (domain.Austausch, error) {
	if fehler := a.wache.Pruefe(ctx, domain.WirkungLesen); fehler != nil {
		return domain.Austausch{}, fehler
	}

	gewaehlt, fehler := gewaehlteAbschnitte(abschnitte)
	if fehler != nil {
		return domain.Austausch{}, fehler
	}

	stand, fehler := a.lesen(ctx)
	if fehler != nil {
		return domain.Austausch{}, fehler
	}

	austausch := domain.Austausch{
		Fassung:  domain.AustauschFassung,
		Bereiche: map[string]map[string]string{},
	}

	for _, ebene := range gewaehlt {
		austausch.Bereiche[string(ebene)] = map[string]string{}
	}

	for schluessel, wert := range stand.profil {
		abschnitt, reist := domain.AbschnittImProfil(schluessel)
		if reist && slices.Contains(gewaehlt, abschnitt) {
			austausch.Bereiche[string(abschnitt)][schluessel] = wert
		}
	}

	if slices.Contains(gewaehlt, domain.EbeneNutzer) {
		for schluessel, wert := range stand.nutzer {
			if domain.ReistBeimNutzer(schluessel) {
				austausch.Bereiche[string(domain.EbeneNutzer)][schluessel] = wert
			}
		}
	}

	return austausch, nil
}

// gewaehlteAbschnitte urteilt über die Wahl des Aufrufers.
func gewaehlteAbschnitte(abschnitte []string) ([]domain.Ebene, error) {
	if len(abschnitte) == 0 {
		return []domain.Ebene{domain.EbeneNutzer, domain.EbeneProfil}, nil
	}

	gewaehlt := make([]domain.Ebene, 0, len(abschnitte))

	for _, name := range abschnitte {
		if !domain.KenntEbene(name) {
			return nil, domain.NewFieldError("abschnitt", domain.SchluesselValidierung,
				"%q ist kein Abschnitt der Einstellungen.", name)
		}

		if domain.Ebene(name) == domain.EbeneGeraet {
			return nil, domain.NewFieldError("abschnitt", domain.SchluesselValidierung,
				"Der Abschnitt »geraet« beschreibt diesen Rechner; er reist nicht mit.")
		}

		gewaehlt = append(gewaehlt, domain.Ebene(name))
	}

	return gewaehlt, nil
}

// Einfuhrzeile ist eine Änderung, die eine Datei mitbringt.
type Einfuhrzeile struct {
	Ebene      string `json:"ebene"`
	Schluessel string `json:"schluessel"`
	Vorher     string `json:"vorher"`
	Nachher    string `json:"nachher"`

	// Herkunft sagt, wessen Entscheidung der **alte** Wert war — `profil`,
	// `nutzer` oder `vorgabe`.
	//
	// **Der alte und nicht der neue**: die Frage, die vor einer Übernahme
	// steht, ist „überschreibe ich hier etwas, das jemand gewählt hat?" Beim
	// neuen wäre die Antwort immer dieselbe.
	Herkunft string `json:"herkunft"`

	// Wirkung sagt, wann der neue Wert gilt — OPS-01.AK-02
	// (#1718):
	// der Import zeigt den Neustartbedarf. **Dieselbe Auskunft wie an der
	// Einstellung** (`domain.EinstellungsWirkung`); leer für eine Einstellung
	// des Menschen, die keinen Zeitpunkt nennt.
	Wirkung string `json:"wirkung"`
}

// Einfuhrvorschau ist, was eine Datei ändern würde.
type Einfuhrvorschau struct {
	Fassung int `json:"fassung"`

	// Revision ist die Marke des Stands, gegen den verglichen wurde.
	Revision string `json:"revision"`

	Zeilen  []Einfuhrzeile         `json:"zeilen"`
	Befunde []domain.Einfuhrbefund `json:"befunde"`

	// Unveraendert zählt die Zeilen, die schon so dastehen.
	//
	// **Als Zahl und nicht als Zeile**: sie gehören nicht in einen Diff, und
	// sie zu verschweigen hieße, eine Datei mit dreißig Zeilen als „ändert
	// nichts" auszugeben, ohne zu sagen, dass dreißig geprüft wurden.
	Unveraendert int `json:"unveraendert"`
}

// Vorschau sagt, was eine Datei ändern würde — und ändert nichts.
func (a *Einstellungsaustausch) Vorschau(
	ctx context.Context, roh []byte,
) (Einfuhrvorschau, error) {
	if fehler := a.wache.Pruefe(ctx, domain.WirkungLesen); fehler != nil {
		return Einfuhrvorschau{}, fehler
	}

	stand, fehler := a.lesen(ctx)
	if fehler != nil {
		return Einfuhrvorschau{}, fehler
	}

	return a.vorschauGegen(roh, stand)
}

// vorschauGegen baut die Vorschau gegen einen bereits gelesenen Stand.
//
// **Eine Stelle für beide Wege**: die Übernahme liest dieselbe Datei noch
// einmal gegen denselben Stand, und zwei Rechnungen liefen auseinander.
func (a *Einstellungsaustausch) vorschauGegen(roh []byte, stand lage) (
	Einfuhrvorschau, error,
) {
	austausch, befunde, fehler := domain.Zerlege(roh)
	if fehler != nil {
		return Einfuhrvorschau{}, fehler
	}

	paare, ausDenZeilen := domain.PruefeAustausch(austausch)

	vorschau := Einfuhrvorschau{
		Fassung:  austausch.Fassung,
		Revision: stand.Marke(),
		Zeilen:   []Einfuhrzeile{},
		Befunde:  append(befunde, ausDenZeilen...),
	}

	geltung := domain.WirksamerStand(stand.profil, stand.nutzer)

	for _, paar := range paare {
		zeile, befund := zeileFuer(paar, stand, geltung)
		if befund != nil {
			vorschau.Befunde = append(vorschau.Befunde, *befund)

			continue
		}

		if zeile.Vorher == zeile.Nachher {
			vorschau.Unveraendert++

			continue
		}

		vorschau.Zeilen = append(vorschau.Zeilen, zeile)
	}

	return vorschau, nil
}

// zeileFuer baut eine Diffzeile — oder den Befund, warum es keine gibt.
func zeileFuer(paar domain.Einfuhrpaar, stand lage, geltung map[string]domain.Geltung) (
	Einfuhrzeile, *domain.Einfuhrbefund,
) {
	vorher, herkunft := vorherigerWert(paar, stand, geltung)

	// **Eine Schranke lässt sich nicht durch eine Datei aufweichen.**
	//
	// Verglichen wird gegen das, was **ohne diese Zeile** gälte. Wer als
	// Mensch „nur lokal" gesagt hat, hat es für jedes Profil gesagt; eine
	// Datei, die in einem Profil einen Dienst anschaltet, hebt diese
	// Entscheidung auf, und genau das soll sie nicht können.
	//
	// **Die eigene Wahl ist dabei keine Schranke gegen sich selbst.** Ein
	// Profil, das seinen Dienst abgeschaltet hat, darf ihn per Datei wieder
	// anschalten — sonst wäre jede einmal getroffene Wahl endgültig.
	if schranke := wirksameSchranke(paar, stand); domain.Lockert(
		paar.Schluessel, paar.Wert, schranke) {
		return Einfuhrzeile{}, &domain.Einfuhrbefund{
			Ebene: string(paar.Ebene), Schluessel: paar.Schluessel,
			Grund: fmt.Sprintf("Hier gilt %q, und %q wäre lockerer. Eine "+
				"Schranke lässt sich nicht durch eine Datei aufweichen.",
				schranke, paar.Wert),
		}
	}

	return Einfuhrzeile{
		Ebene: string(paar.Ebene), Schluessel: paar.Schluessel,
		Vorher: vorher, Nachher: paar.Wert, Herkunft: herkunft,
		Wirkung: domain.EinstellungsWirkung(paar.Schluessel),
	}, nil
}

// vorherigerWert gibt, was heute an dieser Stelle steht, und wessen
// Entscheidung das war.
func vorherigerWert(
	paar domain.Einfuhrpaar, stand lage, geltung map[string]domain.Geltung,
) (wert, herkunft string) {
	if paar.Ebene == domain.EbeneNutzer {
		if wert, gesetzt := stand.nutzer[paar.Schluessel]; gesetzt {
			return wert, domain.GeltungNutzer
		}

		return domain.NutzerVorgabe(paar.Schluessel), domain.GeltungVorgabe
	}

	if wert, gesetzt := stand.profil[paar.Schluessel]; gesetzt {
		return wert, geltung[paar.Schluessel].Herkunft
	}

	return domain.EinstellungsVorgabe(paar.Schluessel),
		geltung[paar.Schluessel].Herkunft
}

// wirksameSchranke gibt den Wert, gegen den eine Richtlinie geprüft wird —
// leer, wenn keine besteht.
//
// # Was hier gerechnet wird
//
// **Was gälte, wenn diese Zeile nichts sagte.** Für einen Profilwert heißt
// das: derselbe Stand ohne ihn — dann bleibt stehen, was der Mensch gesetzt
// hat, samt der übergreifenden Schranke „nur lokal". Für einen Nutzerwert gibt
// es keine Ebene darüber; dort ist die Schranke, was heute dort steht.
//
// # Warum eine Vorgabe keine Schranke ist
//
// **Weil sie niemand gesetzt hat.** `dienst.mtgapi` steht ohne Wahl auf „aus";
// wer das für eine Schranke hielte, ließe keinen einzigen Dienst je durch eine
// Datei anschalten. Die Herkunft sagt den Unterschied — und dafür gibt es sie.
func wirksameSchranke(paar domain.Einfuhrpaar, stand lage) string {
	if paar.Ebene == domain.EbeneNutzer {
		return stand.nutzer[paar.Schluessel]
	}

	ohneDieseZeile := maps.Clone(stand.profil)
	delete(ohneDieseZeile, paar.Schluessel)

	geltung := domain.WirksamerStand(ohneDieseZeile, stand.nutzer)[paar.Schluessel]
	if geltung.Herkunft == domain.GeltungVorgabe {
		return ""
	}

	return geltung.Wirksam
}
