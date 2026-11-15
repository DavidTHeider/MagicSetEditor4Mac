// **Blick und Mischung** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Zwei Befehle, die keinem Objekt gelten
//
// **Herausgelöst aus `sandkasten.go`**, und der Schnitt folgt der Frage (§2):
// dort steht, was mit **einem** Ding geschieht, hier, was mit einer **Zone**
// geschieht. Die sieben Befehle nebenan zeigen alle auf eine Objektkennung;
// diese zwei tun es nicht, und ein `objektstelle` davor suchte ein Objekt,
// das gar nicht gemeint ist.
//
// # Warum sie ein Paar sind
//
// **Ein Blick zerstört das Geheimnis einer Zone, ein Mischen stellt es
// wieder her.** Das ist die ganze Abmachung, und sie ist der Grund, warum
// eine Suche im Deck überhaupt möglich sein darf: ohne das Mischen danach
// wäre sie ein einseitiger Gewinn.
//
// # Und was hier ausdrücklich nicht steht
//
//   - **Kein Zeigen.** „Offenbare deinem Gegner deine Hand" ist etwas anderes
//     als ein Blick in die eigene: es gibt jemand anderem eine Sicht, die er
//     sonst nicht hätte. Das ist eine Regel und keine Invariante, und es
//     gehört zur Regelmatrix in G4
//     (#1461).
//   - **Kein Zwang.** Dass nach einer Suche gemischt werden **muss**, ist
//     ebenfalls G4. Der Sandkasten merkt die Pflicht und zeigt sie jedem —
//     er erzwingt sie nicht.

package service

import (
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die zwei Befehlsarten, die einer Zone gelten.
const (
	// BefehlBlick sieht in eine verborgene Zone des Absenders.
	BefehlBlick = "zone.blick"

	// BefehlMischen ordnet eine geordnete Zone neu.
	BefehlMischen = "zone.mischen"
)

// istZonenbefehl sagt, ob dieser Befehl einer Zone gilt statt einem Objekt.
func istZonenbefehl(art string) bool {
	return art == BefehlBlick || art == BefehlMischen
}

// zonenbefehl legt die beiden aus.
func (s *Sandkasten) zonenbefehl(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	zone, fehler := gemeinteZone(befehl)
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	if befehl.Art == BefehlBlick {
		return s.blicken(partie, befehl, zone)
	}

	return s.mischen(partie, befehl, zone)
}

// gemeinteZone liest die Zone aus der Nutzlast.
func gemeinteZone(befehl domain.Spielbefehl) (domain.Zone, error) {
	genannt, _ := befehl.Nutzlast["zone"].(string) //nolint:errcheck // siehe [Sandkasten.objektstelle]

	for _, zone := range domain.Zonen() {
		if string(zone) == genannt {
			return zone, nil
		}
	}

	return "", fmt.Errorf(
		"%w: %q ist keine Zone, in die sich etwas schieben ließe",
		domain.ErrValidation, genannt)
}

// blicken zeigt dem Absender, was in seiner verborgenen Zone liegt.
//
// # Zwei Bedingungen, und beide sind Invarianten
//
//  1. **Die Zone gehört dem Absender.** In eine fremde verborgene Zone sieht
//     niemand — das ist die Sichtbarkeitsinvariante, die der Sandkasten
//     ausdrücklich **nicht** lockert (#1458).
//  2. **Die Zone ist verborgen.** Ein Blick auf den Tisch ist kein Blick; wer
//     ihn zuließe, erzeugte eine Mischpflicht für etwas, das jeder sieht.
func (s *Sandkasten) blicken(
	partie domain.Partie, befehl domain.Spielbefehl, zone domain.Zone,
) ([]domain.Ereignis, domain.Partie, error) {
	if genannt, _ := befehl.Nutzlast["sitz"].(string); genannt != "" && //nolint:errcheck // dito
		genannt != befehl.Absender {
		return nil, domain.Partie{}, fmt.Errorf(
			"%w: %q sieht nicht in die verborgene Zone von %q — wer etwas "+
				"zeigen will, zeigt es",
			domain.ErrValidation, befehl.Absender, genannt)
	}

	if zone.Sicht() == domain.SichtOffen {
		return nil, domain.Partie{}, fmt.Errorf(
			"%w: %q liegt offen da — darin muss niemand nachsehen",
			domain.ErrValidation, zone)
	}

	darin := objekteIn(partie, befehl.Absender, zone)

	gezeigt := make([]any, 0, len(darin))

	for _, stelle := range darin {
		objekt := partie.Objekte[stelle]

		gezeigt = append(gezeigt, map[string]any{
			"objekt":  objekt.Kennung,
			"karte":   objekt.Karte,
			"name":    gezeigterName(partie, objekt),
			"ordnung": float64(objekt.Ordnung),
		})
	}

	// **Der Blick ändert nichts an den Objekten** — er ändert, was jemand
	// weiß. Was sich ändert, ist die Schuld, die er damit aufnimmt.
	danach := s.kopie(partie)
	danach.MerkeMischpflicht(befehl.Absender, zone)

	return []domain.Ereignis{{
		Art: befehl.Art,
		Oeffentlich: map[string]any{
			"sitz": befehl.Absender, "zone": string(zone),
			"anzahl": float64(len(darin)),
		},
		// **Eine Ergänzung für genau einen** — und sie geht nicht in den
		// Abdruck ein, sonst hinge die Integrität des Stroms am Geheimnis.
		Privat: []domain.Ergaenzung{{
			Empfaenger: befehl.Absender,
			Inhalt:     map[string]any{"karten": gezeigt},
		}},
	}}, danach, nil
}

// gezeigterName nennt ein Objekt so, wie ein Blick es zeigt —
// #1836.
//
// **Aus dem Abzug, wo eine Karte dahintersteht.** `Spielobjekt.Name` trägt
// den Namen nur, wo keine steht — bei einem Spielstein; bei einer Karte ist
// er leer, und im Ergebnis eines Blicks stünde für jede Karte `name: ""`.
// **Derselbe Weg wie die Legendenregel** (`legendenname`), und deshalb
// derselbe Name.
func gezeigterName(partie domain.Partie, objekt domain.Spielobjekt) string {
	if eigen, da := partie.Eigenschaften(objekt); da && eigen.Name != "" {
		return eigen.Name
	}

	return objekt.Name
}

// mischen ordnet eine geordnete Zone neu.
//
// **Ein fremder Sitz ist hier erlaubt**, anders als beim Blick: ein Mischen
// zeigt nichts. Wessen Bibliothek gemischt wird, ist eine Regelfrage — wer
// dabei etwas **sieht**, wäre eine Sichtbarkeitsfrage, und es sieht niemand
// etwas.
func (s *Sandkasten) mischen(
	partie domain.Partie, befehl domain.Spielbefehl, zone domain.Zone,
) ([]domain.Ereignis, domain.Partie, error) {
	if !zone.Geordnet() {
		return nil, domain.Partie{}, fmt.Errorf(
			"%w: in %q gibt es keine Reihenfolge, die sich mischen ließe",
			domain.ErrValidation, zone)
	}

	sitz, _ := befehl.Nutzlast["sitz"].(string) //nolint:errcheck // dito
	if sitz == "" {
		sitz = befehl.Absender
	}

	darin := objekteIn(partie, sitz, zone)

	mischung, fehler := s.zufall.Mische(len(darin))
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	if fehler := pruefeMischung(mischung, len(darin)); fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	danach := s.kopie(partie)

	// **Die Permutation sagt, welche Karte an welche Stelle kommt.**
	// `mischung[i] = j` heißt: an Stelle `i+1` liegt danach, was vorher an
	// Stelle `j` der Zone lag.
	for stelle, gezogen := range mischung {
		danach.Objekte[darin[gezogen]].Ordnung = stelle + 1
	}

	danach.LoeseMischpflicht(sitz, zone)

	return []domain.Ereignis{{
		Art: befehl.Art,
		Oeffentlich: map[string]any{
			"sitz": sitz, "zone": string(zone),
			"anzahl": float64(len(darin)),
		},
	}}, danach, nil
}

// objekteIn gibt die Stellen der Objekte, die dieser Sitz in dieser Zone hat.
//
// **Stellen und keine Objekte**: der Aufrufer schreibt in seine eigene Kopie,
// und eine Liste von Werten daneben liefe auseinander.
//
// **Und die Reihenfolge ist die der Liste, nicht die der Ordnung.** Sie ist
// nur ein Zugriffsweg; welche Karte oben liegt, sagt [domain.Spielobjekt].
func objekteIn(partie domain.Partie, sitz string, zone domain.Zone) []int {
	stellen := []int{}

	for i := range partie.Objekte {
		if partie.Objekte[i].Beherrscher == sitz &&
			partie.Objekte[i].Zone == zone {
			stellen = append(stellen, i)
		}
	}

	return stellen
}
