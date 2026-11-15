package service_test

// **Begrenztes Rückgängig** — P8
// (#2376),
// Befund C-07, Vermerk 0197.
//
// Ein Sitz nimmt seinen **letzten** Zug zurück, wenn er auf der Liste steht
// und seither niemand sonst gehandelt hat. Angehängt, nicht gelöscht; eine
// Abweisung ändert weder Strom noch Fassung.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einTischZumZuruecknehmen hat je Sitz ein Objekt auf dem Schlachtfeld.
func einTischZumZuruecknehmen(t *testing.T) (*service.Partieleitung, *partieprobe) {
	t.Helper()

	leitung, speicher := eineLeitung(t)
	partie := einePartie()
	partie.Objekte = []domain.Spielobjekt{
		{
			Kennung: "o1", Art: domain.ObjektKarte, Beherrscher: "sitz-1",
			Eigner: "sitz-1", Zone: domain.ZoneTisch, Karte: "k1",
		},
		{
			Kennung: "o2", Art: domain.ObjektKarte, Beherrscher: "sitz-2",
			Eigner: "sitz-2", Zone: domain.ZoneTisch, Karte: "k1",
		},
	}

	if _, fehler := leitung.Anlegen(context.Background(), partie); fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	return leitung, speicher
}

func stelle(
	t *testing.T, leitung *service.Partieleitung,
	kennung, sitz, art string, nutzlast map[string]any,
) (service.Befehlsausgang, error) {
	t.Helper()

	return leitung.Befehl(context.Background(), domain.Spielbefehl{
		Kennung: kennung, PartieID: "p1", Absender: sitz, Art: art,
		Nutzlast: nutzlast,
	})
}

// gelingt stellt einen Befehl, der wirken muss.
func gelingt(
	t *testing.T, leitung *service.Partieleitung,
	kennung, sitz, art string, nutzlast map[string]any,
) {
	t.Helper()

	if _, fehler := stelle(t, leitung, kennung, sitz, art, nutzlast); fehler != nil {
		t.Fatalf("der Befehl %s: %v", kennung, fehler)
	}
}

func TestRuecknahmeNimmtDasEigeneTappenZurueck(t *testing.T) {
	t.Parallel()

	leitung, speicher := einTischZumZuruecknehmen(t)

	gelingt(t, leitung, "b1", "sitz-1", service.BefehlTappen, map[string]any{"objekt": "o1"})
	ausgang, fehler := stelle(t, leitung, "z1", "sitz-1", service.BefehlZuruecknehmen, nil)
	if fehler != nil {
		t.Fatalf("die Rücknahme: %v", fehler)
	}

	partie := speicher.partien["p1"]
	if partie.Objekte[0].Getappt {
		t.Error("o1 ist noch getappt")
	}

	if ausgang.Fassung != 3 || partie.Fassung != 3 {
		t.Errorf("Fassung %d/%d, erwartet 3", ausgang.Fassung, partie.Fassung)
	}

	// **Angehängt, nicht gelöscht**: der Tappzug bleibt im Strom, danach
	// stehen die Rücknahme mit der Folge, die sie meint, und die Umkehr.
	strom := speicher.ereignisse["p1"]
	arten := make([]string, 0, len(strom))
	for _, ereignis := range strom {
		arten = append(arten, ereignis.Art)
	}

	if strings.Join(arten, " ") != "objekt.tappen spieler.zuruecknehmen objekt.enttappen" {
		t.Fatalf("der Strom: %v", arten)
	}

	if strom[1].Oeffentlich["folge"] != float64(1) || strom[1].Oeffentlich["sitz"] != "sitz-1" {
		t.Errorf("die Rücknahme nennt %v", strom[1].Oeffentlich)
	}
}

// **Ein fremder Folgezug baute auf dem Stand auf** — und eine Abweisung
// ändert nichts.
func TestRuecknahmeNachFremdemZugWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	leitung, speicher := einTischZumZuruecknehmen(t)

	gelingt(t, leitung, "b1", "sitz-1", service.BefehlTappen, map[string]any{"objekt": "o1"})
	gelingt(t, leitung, "b2", "sitz-2", service.BefehlTappen, map[string]any{"objekt": "o2"})

	_, fehler := stelle(t, leitung, "z1", "sitz-1", service.BefehlZuruecknehmen, nil)
	if !errors.Is(fehler, domain.ErrConflict) || !strings.Contains(fehler.Error(), "anderer Sitz") {
		t.Fatalf("es kam %v", fehler)
	}

	if speicher.partien["p1"].Fassung != 3 || len(speicher.ereignisse["p1"]) != 2 {
		t.Errorf("die Abweisung änderte etwas: Fassung %d, %d Ereignisse",
			speicher.partien["p1"].Fassung, len(speicher.ereignisse["p1"]))
	}

	if !speicher.partien["p1"].Objekte[0].Getappt {
		t.Error("o1 wurde trotzdem enttappt")
	}
}

func TestRuecknahmeKehrtLebenUndVergiftungUm(t *testing.T) {
	t.Parallel()

	leitung, speicher := einTischZumZuruecknehmen(t)
	vorher := speicher.partien["p1"].Sitze[0]

	gelingt(t, leitung, "b1", "sitz-1", service.BefehlLeben, map[string]any{"um": float64(-3)})
	gelingt(t, leitung, "z1", "sitz-1", service.BefehlZuruecknehmen, nil)
	gelingt(t, leitung, "b2", "sitz-1", service.BefehlVergiftung, map[string]any{"um": float64(2)})
	gelingt(t, leitung, "z2", "sitz-1", service.BefehlZuruecknehmen, nil)

	nachher := speicher.partien["p1"].Sitze[0]
	if nachher.Leben != vorher.Leben || nachher.Vergiftung != vorher.Vergiftung {
		t.Errorf("Leben %d→%d, Gift %d→%d", vorher.Leben, nachher.Leben,
			vorher.Vergiftung, nachher.Vergiftung)
	}
}

// **Dieselbe Kennung ist derselbe Wunsch**; eine neue nach einer Rücknahme
// träfe die Rücknahme selbst — und die nimmt man nicht zurück.
func TestRuecknahmeWirktEinmal(t *testing.T) {
	t.Parallel()

	leitung, speicher := einTischZumZuruecknehmen(t)

	gelingt(t, leitung, "b1", "sitz-1", service.BefehlTappen, map[string]any{"objekt": "o1"})
	gelingt(t, leitung, "z1", "sitz-1", service.BefehlZuruecknehmen, nil)
	gelingt(t, leitung, "z1", "sitz-1", service.BefehlZuruecknehmen, nil)

	if speicher.partien["p1"].Fassung != 3 {
		t.Errorf("die Wiederholung wirkte: Fassung %d", speicher.partien["p1"].Fassung)
	}

	_, fehler := stelle(t, leitung, "z2", "sitz-1", service.BefehlZuruecknehmen, nil)
	if !errors.Is(fehler, domain.ErrValidation) || !strings.Contains(fehler.Error(), "Rücknahme") {
		t.Fatalf("es kam %v", fehler)
	}

	if speicher.partien["p1"].Fassung != 3 || len(speicher.ereignisse["p1"]) != 3 {
		t.Error("die zweite Rücknahme änderte etwas")
	}
}

func TestRuecknahmeOhneVorherigenZug(t *testing.T) {
	t.Parallel()

	leitung, speicher := einTischZumZuruecknehmen(t)

	_, fehler := stelle(t, leitung, "z1", "sitz-1", service.BefehlZuruecknehmen, nil)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}

	if speicher.partien["p1"].Fassung != 1 {
		t.Error("die Abweisung änderte die Fassung")
	}
}

// **Ein Lesefehler des Speichers kommt als er selbst heraus.**
func TestRuecknahmeMeldetDenLesefehler(t *testing.T) {
	t.Parallel()

	leitung, speicher := einTischZumZuruecknehmen(t)
	gelingt(t, leitung, "b1", "sitz-1", service.BefehlTappen, map[string]any{"objekt": "o1"})

	klage := errors.New("die Platte ist voll")
	speicher.stolpert = map[string]error{"LetzterBefehl": klage}

	if _, fehler := stelle(t, leitung, "z1", "sitz-1", service.BefehlZuruecknehmen, nil); !errors.Is(fehler, klage) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Die Umkehr geht durch dieselbe Auslegung** — weist der Kern sie ab, kommt
// sein Grund heraus.
func TestRuecknahmeTraegtDieAbweisungDerUmkehr(t *testing.T) {
	t.Parallel()

	leitung, speicher := einTischZumZuruecknehmen(t)
	gelingt(t, leitung, "b1", "sitz-1", service.BefehlTappen, map[string]any{"objekt": "o1"})

	// o1 verschwindet aus dem Stand — die Umkehr findet es nicht mehr.
	partie := speicher.partien["p1"]
	partie.Objekte = partie.Objekte[1:]
	speicher.partien["p1"] = partie

	if _, fehler := stelle(t, leitung, "z1", "sitz-1", service.BefehlZuruecknehmen, nil); fehler == nil {
		t.Fatal("die Umkehr ging durch")
	}

	if len(speicher.ereignisse["p1"]) != 1 {
		t.Error("die Abweisung schrieb in den Strom")
	}
}

// **Jeder darf sie stellen, solange die Partie läuft** — ob sie wirkt,
// entscheidet der Kern am Befehl.
func TestRuecknahmeStehtInDerErlaubnis(t *testing.T) {
	t.Parallel()

	leitung, _ := einTischZumZuruecknehmen(t)

	erlaubt, fehler := erlaubtVon(leitung, context.Background(), "p1", "sitz-2")
	if fehler != nil {
		t.Fatalf("Erlaubnis: %v", fehler)
	}

	for _, art := range erlaubt {
		if art == service.BefehlZuruecknehmen {
			return
		}
	}

	t.Errorf("die Rücknahme fehlt in %v", erlaubt)
}
