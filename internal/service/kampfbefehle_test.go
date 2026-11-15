package service_test

// **Angreifen und Blocken über den Sandkasten** — Werkstück M6 der
// Regelmaschine
// (#1608).
//
// # Was hier geprüft wird
//
// **Dass beide Erklärungen durch denselben Weg gehen** und damit an allen
// vier Flächen zugleich dastehen (§16) — samt ihrer Nutzlast, und die ist bei
// beiden eine **Liste von Paaren**.
//
// **Der Kampfschaden steht nicht dabei**: er ist eine zugbasierte Handlung
// (CR 703.4m) und geschieht von selbst. Ein Befehl dafür wäre eine Wahl, wo
// die Regeln keine lassen.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Angreifer wird über den Sandkasten erklärt.**
func TestAngreiferWerdenUeberDenSandkastenErklaert(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := einKampfstand(t)

	ereignisse, danach, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
		Art: service.BefehlAngreifer,
		Nutzlast: map[string]any{"angriffe": []any{
			map[string]any{"kreatur": "sitz-1/riese/0", "ziel": "sitz-2"},
		}},
	})
	if fehler != nil {
		t.Fatalf("angreifer: %v", fehler)
	}

	switch {
	case len(danach.Zug.Kampf.Angriffe) != 1:
		t.Errorf("erklärt sind %v", danach.Zug.Kampf.Angriffe)
	case len(ereignisse) != 1:
		t.Fatalf("%d Ereignisse", len(ereignisse))
	case ereignisse[0].Oeffentlich["zahl"] != float64(1):
		t.Errorf("das Ereignis nennt %v", ereignisse[0].Oeffentlich)
	}
}

// **Ein Blocker wird über denselben Weg erklärt.**
func TestBlockerWerdenUeberDenSandkastenErklaert(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := mitAngriffImDienst(t)

	_, danach, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-2", Kennung: "b1",
		Art: service.BefehlBlocker,
		Nutzlast: map[string]any{"bloecke": []any{
			map[string]any{
				"kreatur": "sitz-2/baer/0", "angreifer": "sitz-1/riese/0",
			},
		}},
	})
	if fehler != nil {
		t.Fatalf("blocker: %v", fehler)
	}

	if len(danach.Zug.Kampf.Bloecke) != 1 {
		t.Errorf("erklärt sind %v", danach.Zug.Kampf.Bloecke)
	}
}

// **Eine Nutzlast, die keine Liste von Paaren ist, wird abgewiesen** — §8.
func TestEineKampfnutzlastMussPaareTragen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := einKampfstand(t)

	faelle := map[string]any{
		"keine Liste":   map[string]any{"angriffe": "sitz-1/riese/0"},
		"kein Paar":     map[string]any{"angriffe": []any{"sitz-1/riese/0"}},
		"leeres Feld":   map[string]any{"angriffe": []any{map[string]any{}}},
		"falsche Sorte": map[string]any{"angriffe": []any{map[string]any{"kreatur": float64(3)}}},
	}

	for name, nutzlast := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			nutz, _ := nutzlast.(map[string]any) //nolint:errcheck // die Tabelle steht oben

			if _, _, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
				PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
				Art: service.BefehlAngreifer, Nutzlast: nutz,
			}); fehler == nil {
				t.Error("die Nutzlast kam durch")
			}
		})
	}
}

// **Auch keine Angreifer ist eine Erklärung** — CR 508.1a.
func TestKeineAngreiferGehenAuchUeberDenSandkasten(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, danach, fehler := kasten.Auslegen(
		einKampfstand(t), domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art: service.BefehlAngreifer, Nutzlast: map[string]any{},
		})
	if fehler != nil {
		t.Fatalf("angreifer: %v", fehler)
	}

	if !danach.Zug.Kampf.Erklaert {
		t.Error("die leere Erklärung ist nicht vermerkt")
	}
}

// **Wer nicht der aktive Spieler ist, erklärt keine Angreifer** — CR 506.2.
func TestNurDerAktiveSpielerErklaertAngreifer(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, _, fehler := kasten.Auslegen(
		einKampfstand(t), domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-2", Kennung: "b1",
			Art: service.BefehlAngreifer, Nutzlast: map[string]any{},
		})

	if fehler == nil || !strings.Contains(fehler.Error(), "CR 506.2") {
		t.Errorf("der Fehler lautet %v", fehler)
	}
}

// einKampfstand gibt eine Partie mit zwei Kreaturen im **Angreiferschritt**.
//
// **Ohne Schrittangabe**: der Blockerschritt wird über `mitAngriffImDienst`
// erreicht, und ein Parameter, den nur ein Wert je erreicht, ist eine Zusage
// ohne Einlösung (§10).
func einKampfstand(t *testing.T) domain.Partie {
	t.Helper()

	partie := eineZugpartie(t)

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)

	for stelle := range sitze {
		sitze[stelle].Deck = domain.Partiedeck{
			Abdruck: "abcd",
			Karten: []domain.Abzugskarte{{
				Kennung: "kreatur", Name: "Bär", Menge: 1,
				Seiten: []domain.Abzugsseite{{Felder: map[string]string{
					"super type": "Creature", "sub type": "Bear",
					"power": "2", "toughness": "2",
				}}},
			}},
		}
	}

	partie.Sitze = sitze

	partie.Objekte = append(partie.Objekte,
		domain.Spielobjekt{
			Kennung: "sitz-1/riese/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneTisch, Karte: "kreatur",
		},
		domain.Spielobjekt{
			Kennung: "sitz-2/baer/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-2", Eigner: "sitz-2",
			Zone: domain.ZoneTisch, Karte: "kreatur",
		})

	for range 24 {
		if partie.Zug.Abschnitt.Schritt == domain.SchrittAngreifer {
			return partie
		}

		danach, fehler := partie.Passen(partie.Zug.Prioritaet)
		if fehler != nil {
			t.Fatalf("passen bei %+v: %v", partie.Zug.Abschnitt, fehler)
		}

		partie = danach
	}

	t.Fatal("der Angreiferschritt kam nicht")

	return partie
}

// mitAngriffImDienst gibt eine Partie im Blockerschritt mit einem Angreifer.
func mitAngriffImDienst(t *testing.T) domain.Partie {
	t.Helper()

	partie, fehler := einKampfstand(t).
		AngreiferErklaeren("sitz-1",
			[]domain.Angriff{{Kreatur: "sitz-1/riese/0", Ziel: "sitz-2"}})
	if fehler != nil {
		t.Fatalf("angreifer erklären: %v", fehler)
	}

	for range 6 {
		if partie.Zug.Abschnitt.Schritt == domain.SchrittBlocker {
			return partie
		}

		danach, fehler := partie.Passen(partie.Zug.Prioritaet)
		if fehler != nil {
			t.Fatalf("passen: %v", fehler)
		}

		partie = danach
	}

	t.Fatal("der Blockerschritt kam nicht")

	return partie
}

// **Auch die Blocknutzlast muss Paare tragen** — §8.
//
// **Der Zweig, den dieser Lauf betritt**: derselbe Ableser, aber über den
// **anderen** Befehl. Ohne ihn wäre der Blockweg zur Nutzlast unbetreten —
// und ein Ableser, der dort etwas anderes täte, käme durch (§10).
func TestEineBlocknutzlastMussPaareTragen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := mitAngriffImDienst(t)

	faelle := map[string]map[string]any{
		"keine Liste": {"bloecke": "sitz-2/baer/0"},
		"kein Angreifer am Paar": {"bloecke": []any{
			map[string]any{"kreatur": "sitz-2/baer/0"},
		}},
	}

	for name, nutzlast := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, _, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
				PartieID: "p1", Absender: "sitz-2", Kennung: "b1",
				Art: service.BefehlBlocker, Nutzlast: nutzlast,
			}); fehler == nil {
				t.Error("die Nutzlast kam durch")
			}
		})
	}
}

// **Ein Block, den die Domäne abweist, kommt als Fehler zurück.**
func TestEinAbgewiesenerBlockKommtAlsFehlerZurueck(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	ereignisse, _, fehler := kasten.Auslegen(mitAngriffImDienst(t),
		domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-2", Kennung: "b1",
			Art: service.BefehlBlocker,
			Nutzlast: map[string]any{"bloecke": []any{
				map[string]any{
					"kreatur": "gibtsnicht", "angreifer": "sitz-1/riese/0",
				},
			}},
		})

	switch {
	case fehler == nil:
		t.Fatal("ein Blocker, den es nicht gibt, kam durch")
	case len(ereignisse) != 0:
		t.Errorf("%d Ereignisse trotz Fehler", len(ereignisse))
	}
}
