package service_test

// **Blick und Mischung** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Warum diese zwei Befehle zusammen geprüft werden
//
// **Weil sie ein Paar sind.** Ein Blick in eine verborgene Zone zerstört ihr
// Geheimnis; nur ein Mischen stellt es wieder her. Jeden für sich zu prüfen
// hieße, die Zusage zu prüfen und nicht die Abmachung.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// errZufallWeg ist die Klage einer Quelle, die nichts hergibt.
var errZufallWeg = errors.New("die Quelle ist zu")

// festeMischung gibt immer dieselbe Permutation — die eingesetzte Quelle
// aus §3, ohne die kein Lauf zweimal dasselbe Brett sähe.
type festeMischung struct {
	reihe []int
	klagt error

	// zahl ist, was [festeMischung.Zahl] gibt — für Münzwurf und Würfel
	// (M15, #1617).
	//
	// **Null ist der Vorgabewert**, und das ist Absicht: ein Lauf, der sich
	// für den Wurf nicht interessiert, bekommt immer dieselbe Seite und
	// muss nichts einsetzen.
	zahl int
}

func (f festeMischung) Mische(anzahl int) ([]int, error) {
	if f.klagt != nil {
		return nil, f.klagt
	}

	if len(f.reihe) != anzahl {
		// **Der Lauf sagt, wenn seine Vorrichtung nicht passt** — sonst
		// erklärte `pruefeMischung` einen Fehler des Laufs zum Befund.
		return nil, errZufallWeg
	}

	return f.reihe, nil
}

// Zahl gibt die eingesetzte Zahl — modulo der Grenze, damit ein Lauf sie
// nicht je Grenze neu setzen muss.
func (f festeMischung) Zahl(grenze int) (int, error) {
	if f.klagt != nil {
		return 0, f.klagt
	}

	if grenze <= 0 {
		return 0, errZufallWeg
	}

	return f.zahl % grenze, nil
}

// tischMitBibliothek gibt eine Partie mit drei Karten je Sitz.
func tischMitBibliothek() domain.Partie {
	partie := domain.Partie{
		ID: "p1", Fassung: 1, Spiel: "magic", Stand: domain.StandLaeuft,
		Regelwerk: domain.Regelwerk{
			Kennung: "sandkasten", Fassung: domain.Grundprofilfassung,
			Strenge: domain.StrengeSandkasten,
		},
		Wirt: "anna",
		Sitze: []domain.Sitz{
			{Kennung: "anna", Name: "Anna", Deck: einDreierdeck()},
			{Kennung: "bert", Name: "Bert", Deck: einDreierdeck()},
		},
	}

	for _, sitz := range []string{"anna", "bert"} {
		for stelle := 1; stelle <= 3; stelle++ {
			partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
				Kennung:     sitz + "/k" + string(rune('0'+stelle)),
				Art:         domain.ObjektKarte,
				Beherrscher: sitz, Eigner: sitz,
				Zone: domain.ZoneBibliothek, Karte: "k", Ordnung: stelle,
			})
		}
	}

	return partie
}

func einDreierdeck() domain.Partiedeck {
	return domain.Partiedeck{
		Abdruck: "abc",
		Karten: []domain.Abzugskarte{
			{Kennung: "k1", Name: "Wald", Menge: 3},
		},
	}
}

// **Ein Blick in die eigene Bibliothek zeigt sie — nur dem, der hinsieht.**
func TestEinBlickZeigtDieZoneNurDemAbsender(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	ereignisse, danach, fehler := kasten.Auslegen(tischMitBibliothek(),
		domain.Spielbefehl{
			PartieID: "p1", Absender: "anna", Kennung: "b1",
			Art:      service.BefehlBlick,
			Nutzlast: map[string]any{"zone": "bibliothek"},
		})
	if fehler != nil {
		t.Fatalf("blicken: %v", fehler)
	}

	if len(ereignisse) != 1 {
		t.Fatalf("%d Ereignisse, erwartet eines", len(ereignisse))
	}

	// **Öffentlich steht, dass geschaut wurde — und wie viel.**
	oeffentlich := ereignisse[0].Oeffentlich
	if oeffentlich["sitz"] != "anna" || oeffentlich["zone"] != "bibliothek" {
		t.Errorf("öffentlich: %+v", oeffentlich)
	}

	if oeffentlich["anzahl"] != float64(3) {
		t.Errorf("anzahl %v, erwartet 3", oeffentlich["anzahl"])
	}

	// **Und nirgends steht, was darin liegt.**
	for schluessel := range oeffentlich {
		if schluessel == "karten" {
			t.Error("die Karten stehen im öffentlichen Teil")
		}
	}

	// **Privat steht der Inhalt — für genau einen Empfänger.**
	if len(ereignisse[0].Privat) != 1 ||
		ereignisse[0].Privat[0].Empfaenger != "anna" {
		t.Fatalf("privat: %+v", ereignisse[0].Privat)
	}

	karten, istListe := ereignisse[0].Privat[0].Inhalt["karten"].([]any)
	if !istListe || len(karten) != 3 {
		t.Fatalf("privat karten: %+v", ereignisse[0].Privat[0].Inhalt)
	}

	// **Und danach steht die Mischpflicht.**
	if !danach.SchuldetMischen("anna", domain.ZoneBibliothek) {
		t.Error("nach dem Blick steht keine Mischpflicht")
	}
}

// **In eine fremde verborgene Zone sieht niemand.**
//
// **Das ist eine Sichtbarkeitsinvariante und keine Regel** — und der
// Sandkasten lockert sie ausdrücklich nicht (#1458).
func TestEinBlickInEineFremdeZoneWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, _, fehler := kasten.Auslegen(tischMitBibliothek(), domain.Spielbefehl{
		PartieID: "p1", Absender: "anna", Kennung: "b1",
		Art:      service.BefehlBlick,
		Nutzlast: map[string]any{"zone": "bibliothek", "sitz": "bert"},
	})

	if fehler == nil {
		t.Fatal("Anna hat in Berts Bibliothek gesehen")
	}
}

// **Ein Blick in eine offene Zone ist keiner.**
func TestEinBlickInEineOffeneZoneWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, _, fehler := kasten.Auslegen(tischMitBibliothek(), domain.Spielbefehl{
		PartieID: "p1", Absender: "anna", Kennung: "b1",
		Art:      service.BefehlBlick,
		Nutzlast: map[string]any{"zone": "tisch"},
	})

	if fehler == nil {
		t.Fatal("ein Blick auf den Tisch kam durch")
	}
}

// **Mischen ordnet die Zone neu und löst die Pflicht.**
func TestMischenOrdnetNeuUndLoestDiePflicht(t *testing.T) {
	t.Parallel()

	vorher := tischMitBibliothek()
	vorher.MerkeMischpflicht("anna", domain.ZoneBibliothek)

	// **Eine feste Mischung**: Stelle 0 bekommt die dritte Karte, Stelle 1
	// die erste, Stelle 2 die zweite.
	kasten := service.NeuerSandkasten(festeMischung{reihe: []int{2, 0, 1}})

	ereignisse, danach, fehler := kasten.Auslegen(vorher, domain.Spielbefehl{
		PartieID: "p1", Absender: "anna", Kennung: "b1",
		Art:      service.BefehlMischen,
		Nutzlast: map[string]any{"zone": "bibliothek"},
	})
	if fehler != nil {
		t.Fatalf("mischen: %v", fehler)
	}

	if danach.SchuldetMischen("anna", domain.ZoneBibliothek) {
		t.Error("die Mischpflicht steht nach dem Mischen noch")
	}

	// **Die Ordnungen sind eine Permutation von 1 bis 3** — und Berts
	// Bibliothek ist unberührt.
	gesehen := map[int]string{}

	for _, objekt := range danach.Objekte {
		if objekt.Beherrscher != "anna" {
			if objekt.Ordnung == 0 {
				t.Errorf("Berts Karte %q hat ihre Ordnung verloren", objekt.Kennung)
			}

			continue
		}

		if frueher, doppelt := gesehen[objekt.Ordnung]; doppelt {
			t.Errorf("die Ordnung %d liegt bei %q und bei %q",
				objekt.Ordnung, frueher, objekt.Kennung)
		}

		gesehen[objekt.Ordnung] = objekt.Kennung
	}

	if len(gesehen) != 3 {
		t.Errorf("%d verschiedene Ordnungen, erwartet drei: %v", len(gesehen), gesehen)
	}

	// **Und die Reihenfolge ist wirklich die eingesetzte.**
	if gesehen[1] != "anna/k3" || gesehen[2] != "anna/k1" || gesehen[3] != "anna/k2" {
		t.Errorf("die eingesetzte Mischung ist nicht angekommen: %v", gesehen)
	}

	// **Öffentlich steht nur, dass gemischt wurde.**
	if ereignisse[0].Oeffentlich["anzahl"] != float64(3) ||
		len(ereignisse[0].Privat) != 0 {
		t.Errorf("%+v", ereignisse[0])
	}
}

// **Eine ungeordnete Zone lässt sich nicht mischen.**
func TestEineHandLaesstSichNichtMischen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, _, fehler := kasten.Auslegen(tischMitBibliothek(), domain.Spielbefehl{
		PartieID: "p1", Absender: "anna", Kennung: "b1",
		Art:      service.BefehlMischen,
		Nutzlast: map[string]any{"zone": "hand"},
	})

	if fehler == nil {
		t.Fatal("eine Hand wurde gemischt")
	}
}

// **Eine Quelle, die klagt, mischt nicht halb.**
func TestEineKlagendeQuelleMischtNicht(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{klagt: errZufallWeg})

	_, danach, fehler := kasten.Auslegen(tischMitBibliothek(), domain.Spielbefehl{
		PartieID: "p1", Absender: "anna", Kennung: "b1",
		Art:      service.BefehlMischen,
		Nutzlast: map[string]any{"zone": "bibliothek"},
	})

	if fehler == nil {
		t.Fatal("die klagende Quelle kam durch")
	}

	if len(danach.Objekte) != 0 {
		t.Error("ein halber Stand ist zurückgeblieben")
	}
}

// **Eine Mischung, die keine ist, wird abgewiesen.**
//
// **Der Fall, auf den es ankommt**: `[0, 0, 0]` legte drei Karten auf
// dieselbe Stelle, und die Bibliothek verlöre zwei — still.
func TestEineUnvollstaendigeMischungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{reihe: []int{0, 0, 0}})

	_, _, fehler := kasten.Auslegen(tischMitBibliothek(), domain.Spielbefehl{
		PartieID: "p1", Absender: "anna", Kennung: "b1",
		Art:      service.BefehlMischen,
		Nutzlast: map[string]any{"zone": "bibliothek"},
	})

	if fehler == nil {
		t.Fatal("eine Mischung mit dreimal derselben Stelle kam durch")
	}
}

// **Eine Zone, die es nicht gibt, lässt sich weder ansehen noch mischen.**
func TestEineUnbekannteZoneWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	for _, art := range []string{service.BefehlBlick, service.BefehlMischen} {
		t.Run(art, func(t *testing.T) {
			t.Parallel()

			_, _, fehler := kasten.Auslegen(tischMitBibliothek(),
				domain.Spielbefehl{
					PartieID: "p1", Absender: "anna", Kennung: "b1",
					Art:      art,
					Nutzlast: map[string]any{"zone": "kiste"},
				})

			if fehler == nil {
				t.Fatal("die Zone »kiste« kam durch")
			}
		})
	}
}

// **Ein Blick mit dem eigenen Sitz in der Nutzlast geht durch.**
//
// **Der Zweig, den die Abweisung daneben nicht betritt**: `sitz` steht da und
// ist der eigene — das ist kein Fremdblick, sondern eine ausdrückliche
// Angabe.
func TestEinBlickMitDemEigenenSitzGehtDurch(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, danach, fehler := kasten.Auslegen(tischMitBibliothek(),
		domain.Spielbefehl{
			PartieID: "p1", Absender: "anna", Kennung: "b1",
			Art:      service.BefehlBlick,
			Nutzlast: map[string]any{"zone": "bibliothek", "sitz": "anna"},
		})
	if fehler != nil {
		t.Fatalf("blicken: %v", fehler)
	}

	if !danach.SchuldetMischen("anna", domain.ZoneBibliothek) {
		t.Error("nach dem Blick steht keine Mischpflicht")
	}
}

// **Ein fremder Sitz darf gemischt werden — ein Mischen zeigt nichts.**
func TestEineFremdeBibliothekLaesstSichMischen(t *testing.T) {
	t.Parallel()

	vorher := tischMitBibliothek()
	vorher.MerkeMischpflicht("bert", domain.ZoneBibliothek)

	kasten := service.NeuerSandkasten(festeMischung{reihe: []int{2, 0, 1}})

	_, danach, fehler := kasten.Auslegen(vorher, domain.Spielbefehl{
		PartieID: "p1", Absender: "anna", Kennung: "b1",
		Art:      service.BefehlMischen,
		Nutzlast: map[string]any{"zone": "bibliothek", "sitz": "bert"},
	})
	if fehler != nil {
		t.Fatalf("mischen: %v", fehler)
	}

	if danach.SchuldetMischen("bert", domain.ZoneBibliothek) {
		t.Error("Berts Mischpflicht steht noch")
	}
}
