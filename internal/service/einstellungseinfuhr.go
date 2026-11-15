// **Die Übernahme** — Werkstück O3 von Phase 5.2
// (#1427).
//
// # Was sie vom Aufrufer nimmt und was nicht
//
// **Sie nimmt die Datei, die Auswahl und die Revision.** Die Datei wird noch
// einmal geprüft — ein Diff aus fremder Hand ist eine Behauptung. Die Auswahl
// sagt, welche Zeilen gewollt sind; wer nichts sagt, will alle. Die Revision
// sagt, **gegen welchen Stand** jemand hingesehen hat: weicht er ab, bleibt
// alles stehen und die Antwort sagt es (§7, Regel 2).
//
// # Warum die Revision Pflicht ist
//
// **Weil „ich habe nicht hingesehen" kein Zustand ist, den man schreiben
// lässt.** Eine Übernahme ohne Revision wäre ein Import ohne Vorschau, und
// die Vorschau ist der Grund, warum es dieses Werkstück gibt.

package service

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Einfuhrergebnis ist, was eine Übernahme bewirkt hat.
type Einfuhrergebnis struct {
	// Uebernommen ist die Zahl der geschriebenen Zeilen.
	Uebernommen int `json:"uebernommen"`

	// Revision ist die Marke **nach** der Übernahme — sie ist die, gegen die
	// ein nächster Schritt geht.
	Revision string `json:"revision"`

	Zeilen  []Einfuhrzeile         `json:"zeilen"`
	Befunde []domain.Einfuhrbefund `json:"befunde"`
}

// Uebernehmen schreibt die gewählten Zeilen einer Datei — atomar.
//
// **`nur` nennt Zeilen als `<ebene>:<schluessel>`**, weil ein Schlüssel in
// zwei Abschnitten stehen kann: `sprache` gehört dem Menschen **und** steht
// als Übersteuerung im Profil. Eine Auswahl über den Schlüssel allein träfe
// beide oder die falsche.
func (a *Einstellungsaustausch) Uebernehmen(
	ctx context.Context, roh []byte, revision string, nur []string,
) (Einfuhrergebnis, error) {
	// **Das Nein kommt vor jeder anderen Antwort** (0078) — wer nicht
	// schreiben darf, bekommt auch keine Auskunft darüber, was er geschrieben
	// hätte.
	if fehler := a.wache.Pruefe(ctx, domain.WirkungSchreiben); fehler != nil {
		return Einfuhrergebnis{}, fehler
	}

	if revision == "" {
		return Einfuhrergebnis{}, domain.NewFieldError(
			"revision", domain.SchluesselValidierung,
			"Zum Übernehmen gehört die Revision aus der Vorschau.")
	}

	stand, fehler := a.lesen(ctx)
	if fehler != nil {
		return Einfuhrergebnis{}, fehler
	}

	vorschau, fehler := a.vorschauGegen(roh, stand)
	if fehler != nil {
		return Einfuhrergebnis{}, fehler
	}

	if vorschau.Revision != revision {
		return Einfuhrergebnis{}, fmt.Errorf(
			"%w: die Einstellungen haben sich geändert, seit du hingesehen "+
				"hast. Es wurde nichts übernommen — sieh sie dir noch einmal an",
			domain.ErrConflict)
	}

	gewaehlt, befunde := ausgewaehlteZeilen(vorschau.Zeilen, nur)
	vorschau.Befunde = append(vorschau.Befunde, befunde...)

	// **Eine Einfuhr ist kein Umweg um den Riegel der Netzwege** (#2000):
	// wer eine Adresse aus einer Datei setzt, bestimmt so wie der, der sie
	// eintippt, wohin gesendet wird. Gefragt wird **vor** dem Schreiben —
	// ganz oder gar nicht, wie die Übernahme selbst.
	if traegtNetzweg(gewaehlt) {
		if fehler := a.wache.Pruefe(ctx, domain.WirkungExternSenden); fehler != nil {
			return Einfuhrergebnis{}, fehler
		}
	}

	return a.schreiben(ctx, stand, gewaehlt, vorschau.Befunde)
}

// traegtNetzweg sagt, ob eine der Zeilen bestimmt, wohin gesendet wird.
func traegtNetzweg(zeilen []Einfuhrzeile) bool {
	return slices.ContainsFunc(zeilen, func(zeile Einfuhrzeile) bool {
		return domain.IstNetzweg(zeile.Schluessel)
	})
}

// schreiben legt die gewählten Zeilen ab und rechnet die neue Marke.
func (a *Einstellungsaustausch) schreiben(
	ctx context.Context, stand lage, zeilen []Einfuhrzeile,
	befunde []domain.Einfuhrbefund,
) (Einfuhrergebnis, error) {
	imProfil := map[string]string{}
	beimMenschen := map[string]string{}

	for _, zeile := range zeilen {
		if zeile.Ebene == string(domain.EbeneNutzer) {
			beimMenschen[zeile.Schluessel] = zeile.Nachher

			continue
		}

		imProfil[zeile.Schluessel] = zeile.Nachher
	}

	// **Kein leerer Schreibvorgang.** Eine Transaktion, die nichts ändert,
	// wäre keine Wirkung und trotzdem eine Sperre auf der Datei.
	if len(zeilen) > 0 {
		// **Die Revision geht mit** (Befund 11 des Zweitblicks über Phase
		// 5.2): der Vergleich oben sah den Stand von vorhin, und zwischen ihm
		// und diesem Schreiben kann ein anderer Prozess geschrieben haben.
		// Der Bestand rechnet sie im selben Geschäft noch einmal.
		if fehler := a.schreiber.EinstellungenUebernehmen(
			ctx, imProfil, beimMenschen, stand.Marke()); fehler != nil {
			return Einfuhrergebnis{}, fehler
		}
	}

	// **Die neue Marke wird gerechnet und nicht nachgelesen.** Ein zweites
	// Lesen wäre ein zweiter Stand (§7, Regel 2) — und ein Lesefehler danach
	// machte eine geglückte Übernahme zu einer gescheiterten (§7, Regel 3).
	nachherImProfil := maps.Clone(stand.profil)
	maps.Copy(nachherImProfil, imProfil)

	nachherBeimMenschen := maps.Clone(stand.nutzer)
	maps.Copy(nachherBeimMenschen, beimMenschen)

	return Einfuhrergebnis{
		Uebernommen: len(zeilen),
		Revision: domain.Standmarke(stand.profilID, stand.nutzerID,
			nachherImProfil, nachherBeimMenschen),
		Zeilen:  zeilen,
		Befunde: befunde,
	}, nil
}

// ausgewaehlteZeilen filtert den Diff auf das, was gewollt ist.
//
// **Eine leere Auswahl heißt „alles"** und nicht „nichts": wer eine Datei
// einliest, ohne etwas abzuwählen, will sie — und ein Aufruf ohne Wahl, der
// nichts täte, wäre eine Falle an der Kommandozeile.
func ausgewaehlteZeilen(zeilen []Einfuhrzeile, nur []string) (
	[]Einfuhrzeile, []domain.Einfuhrbefund,
) {
	if len(nur) == 0 {
		return zeilen, nil
	}

	gewaehlt := []Einfuhrzeile{}
	getroffen := map[string]bool{}

	for _, zeile := range zeilen {
		name := zeile.Ebene + ":" + zeile.Schluessel
		if slices.Contains(nur, name) {
			getroffen[name] = true

			gewaehlt = append(gewaehlt, zeile)
		}
	}

	return gewaehlt, verfehlteAuswahl(nur, getroffen)
}

// verfehlteAuswahl meldet, was gewählt wurde und im Diff nicht steht.
//
// **Ein Befund und kein Fehler**: die übrigen Zeilen sind in Ordnung, und wer
// eine Zeile nennt, die schon so dasteht, soll das erfahren statt eine
// Ablehnung zu bekommen.
func verfehlteAuswahl(nur []string, getroffen map[string]bool) []domain.Einfuhrbefund {
	verfehlt := []string{}

	for _, name := range nur {
		if !getroffen[name] && !slices.Contains(verfehlt, name) {
			verfehlt = append(verfehlt, name)
		}
	}

	slices.Sort(verfehlt)

	befunde := make([]domain.Einfuhrbefund, 0, len(verfehlt))

	for _, name := range verfehlt {
		ebene, schluessel, _ := strings.Cut(name, ":")
		befunde = append(befunde, domain.Einfuhrbefund{
			Ebene: ebene, Schluessel: schluessel,
			Grund: "Diese Zeile steht nicht in der Vorschau; sie ändert " +
				"nichts oder wurde als Befund geführt.",
		})
	}

	return befunde
}
