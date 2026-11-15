package service_test

// **Nach außen eine Sichtkennung** — #2480, Vermerk 0201, im Dienst: das
// Geheimnis, der Wechsel nach jedem Befehl, die Namen der Ereignisse und die
// Rückrechnung beim Eingang.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestJedePartieTraegtIhrEigenesGeheimnis(t *testing.T) {
	t.Parallel()

	leitung, speicher := eineLeitung(t)

	erste, fehler := leitung.Anlegen(context.Background(), einePartie())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	zweite := einePartie()
	zweite.ID = "p2"

	if _, fehler := leitung.Anlegen(context.Background(), zweite); fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if erste.Geheimnis == "" || erste.Geheimnis == speicher.partien["p2"].Geheimnis {
		t.Errorf("die Geheimnisse sind %q und %q", erste.Geheimnis, speicher.partien["p2"].Geheimnis)
	}
}

// **Eine Partie aus der Zeit davor bekommt es mit dem nächsten Befehl.**
func TestEinBefehlGibtEinerAltenPartieIhrGeheimnis(t *testing.T) {
	t.Parallel()

	leitung, speicher := eineLeitung(t)
	speicher.partien["p1"] = einePartie()

	if _, fehler := leitung.Befehl(context.Background(), domain.Spielbefehl{
		Kennung: "b1", PartieID: "p1", Absender: "sitz-1",
		Art: service.BefehlTappen, Nutzlast: map[string]any{"objekt": "o1"},
	}); fehler != nil {
		t.Fatalf("Befehl: %v", fehler)
	}

	if speicher.partien["p1"].Geheimnis == "" {
		t.Error("die Partie hat nach dem Befehl kein Geheimnis")
	}
}

// **Ein Befehl nennt das Objekt unter seiner Sichtkennung**, und das
// Ereignis darüber geht unter ihr hinaus.
func TestEinBefehlMitSichtkennungMeintSeinObjekt(t *testing.T) {
	t.Parallel()

	leitung, speicher := eineLeitung(t)
	ctx := context.Background()

	angelegt, fehler := leitung.Anlegen(ctx, einePartie())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	name := angelegt.Sichtkennung(angelegt.Objekte[0])

	ausgang, fehler := leitung.Befehl(ctx, domain.Spielbefehl{
		Kennung: "b1", PartieID: "p1", Absender: "sitz-1",
		Art: service.BefehlTappen, Nutzlast: map[string]any{"objekt": name},
	})
	if fehler != nil {
		t.Fatalf("Befehl: %v", fehler)
	}

	if !speicher.partien["p1"].Objekte[0].Getappt {
		t.Error("die Sichtkennung meinte nicht o1")
	}

	gezeigt := ausgang.Ereignisse[0].Fuer("sitz-2")
	if gezeigt.Oeffentlich["objekt"] != name {
		t.Errorf("das Ereignis nennt %v, erwartet %q", gezeigt.Oeffentlich["objekt"], name)
	}
}

// **Auch die Handlungen nennen die Sichtkennung** — sie gehen im Tischbild
// an den Gast.
func TestDieHandlungsauskunftNenntDieSichtkennung(t *testing.T) {
	t.Parallel()

	partie := einePartie()
	partie.Geheimnis = "geheim"
	partie.Objekte[0].Zone = domain.ZoneTisch

	auskunft := service.Handlungsauskunft(partie, "sitz-1", []string{service.BefehlTappen})

	if len(auskunft) != 1 || auskunft[0].Objekt != partie.Sichtkennung(partie.Objekte[0]) {
		t.Errorf("die Auskunft nennt %+v", auskunft)
	}
}

// **Ein Gast über die Leitung nennt keine Kennung** — Vermerk 0201,
// Punkt 6: er kennt nur Sichtkennungen, und eine erratene Kennung ist der
// Griff in ein verborgenes Objekt.
func TestEinGastNenntKeineKennung(t *testing.T) {
	t.Parallel()

	leitung, _ := eineLeitung(t)
	ctx := service.MitLeitungsherkunft(context.Background())

	angelegt, fehler := leitung.Anlegen(context.Background(), einePartie())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if _, fehler := leitung.Befehl(ctx, domain.Spielbefehl{
		Kennung: "b1", PartieID: "p1", Absender: "sitz-1",
		Art: service.BefehlTappen, Nutzlast: map[string]any{"objekt": "o1"},
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("eine Kennung von der Leitung: %v", fehler)
	}

	if _, fehler := leitung.Befehl(ctx, domain.Spielbefehl{
		Kennung: "b2", PartieID: "p1", Absender: "sitz-1",
		Art: service.BefehlTappen, Nutzlast: map[string]any{"objekt": angelegt.Sichtkennung(angelegt.Objekte[0])},
	}); fehler != nil {
		t.Errorf("eine Sichtkennung von der Leitung: %v", fehler)
	}
}

// klagenderAusleger weist jeden Befehl ab und nennt dabei die Kennung — wie
// die Sätze der Domäne es tun.
type klagenderAusleger struct{}

func (klagenderAusleger) Auslegen(
	_ domain.Partie, _ domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	return nil, domain.Partie{}, fmt.Errorf("%w: das Objekt %q liegt woanders", domain.ErrValidation, "o1")
}

func (klagenderAusleger) Erlaubt(domain.Partie, string) []string { return nil }

// **Eine Absage nennt keine Kennung** — Vermerk 0201, Punkt 7; die Klasse
// bleibt, damit der Wirt sie weiter als Urteil erkennt.
func TestEineAbsageNenntKeineKennung(t *testing.T) {
	t.Parallel()

	speicher := &partieprobe{partien: map[string]domain.Partie{}}
	leitung := service.NeuePartieleitung(speicher, klagenderAusleger{})

	angelegt, fehler := leitung.Anlegen(context.Background(), einePartie())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	_, fehler = leitung.Befehl(context.Background(), domain.Spielbefehl{
		Kennung: "b1", PartieID: "p1", Absender: "sitz-1",
		Art: service.BefehlTappen, Nutzlast: map[string]any{},
	})

	switch {
	case !errors.Is(fehler, domain.ErrValidation):
		t.Fatalf("die Klasse ging verloren: %v", fehler)
	case strings.Contains(fehler.Error(), `"o1"`):
		t.Errorf("die Absage nennt die Kennung: %v", fehler)
	case !strings.Contains(fehler.Error(), angelegt.Sichtkennung(angelegt.Objekte[0])):
		t.Errorf("die Absage nennt das Objekt gar nicht: %v", fehler)
	}
}
