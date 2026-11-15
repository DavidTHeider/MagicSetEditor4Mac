package service_test

// **Die acht Handlungen über den Sandkasten** — Werkstück M7 der
// Regelmaschine
// (#1609).
//
// # Was hier geprüft wird
//
// **Dass jede über denselben Weg geht** und damit an allen vier Flächen
// zugleich dasteht (§16) — und dass die Nutzlast gelesen wird.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Jede der fünf Objekthandlungen geht durch.**
func TestJedeObjekthandlungGehtDurch(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		art      string
		nutzlast map[string]any
		objekt   string
		zone     domain.Zone
	}{
		"zerstören": {
			art:      service.BefehlZerstoeren,
			nutzlast: map[string]any{"objekt": "sitz-1/baer/0"},
			objekt:   "sitz-1/baer/0", zone: domain.ZoneFriedhof,
		},
		"opfern": {
			art:      service.BefehlOpfern,
			nutzlast: map[string]any{"objekt": "sitz-1/baer/0"},
			objekt:   "sitz-1/baer/0", zone: domain.ZoneFriedhof,
		},
		"exilieren": {
			art:      service.BefehlExilieren,
			nutzlast: map[string]any{"objekt": "sitz-1/hand/0"},
			objekt:   "sitz-1/hand/0", zone: domain.ZoneExil,
		},
		"neutralisieren": {
			art:      service.BefehlNeutralisieren,
			nutzlast: map[string]any{"objekt": "sitz-1/zauber/0"},
			objekt:   "sitz-1/zauber/0", zone: domain.ZoneFriedhof,
		},
	}

	kasten := service.NeuerSandkasten(festeMischung{})

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, danach, fehler := kasten.Auslegen(einHandlungsstand(t),
				domain.Spielbefehl{
					PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
					Art: fall.art, Nutzlast: fall.nutzlast,
				})
			if fehler != nil {
				t.Fatalf("%s: %v", name, fehler)
			}

			for _, objekt := range danach.Objekte {
				if objekt.Kennung == fall.objekt && objekt.Zone != fall.zone {
					t.Errorf("%q liegt in %q statt in %q",
						fall.objekt, objekt.Zone, fall.zone)
				}
			}
		})
	}
}

// **Kämpfen braucht beide Kreaturen** — CR 701.14a.
func TestKaempfenBrauchtBeideKreaturen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, danach, fehler := kasten.Auslegen(einHandlungsstand(t),
		domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art: service.BefehlKaempfen,
			Nutzlast: map[string]any{
				"objekt": "sitz-1/baer/0", "gegen": "sitz-2/baer/0",
			},
		})
	if fehler != nil {
		t.Fatalf("kämpfen: %v", fehler)
	}

	// **Beide sterben** — zwei Schaden auf zwei Widerstandskraft sind
	// tödlich (CR 704.5g), und die Zustandsprüfung läuft nach jedem Befehl
	// (`M5`). Das ist der sichtbare Ausgang eines Kampfes zwischen zwei
	// 2/2, und er steht hier statt der Schadenszahl: die trägt eine Karte
	// im Friedhof nicht mehr.
	for _, objekt := range danach.Objekte {
		if objekt.Kennung == "sitz-1/baer/0" &&
			objekt.Zone != domain.ZoneFriedhof {
			t.Errorf("der erste liegt in %q", objekt.Zone)
		}
	}

	// **Ohne den zweiten geht es nicht.**
	if _, _, fehler := kasten.Auslegen(einHandlungsstand(t),
		domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art:      service.BefehlKaempfen,
			Nutzlast: map[string]any{"objekt": "sitz-1/baer/0"},
		}); fehler == nil ||
		!strings.Contains(fehler.Error(), `"gegen"`) {
		t.Errorf("der Fehler lautet %v", fehler)
	}
}

// **Die drei Bibliothekshandlungen gehen durch** — CR 701.17a, 701.22a,
// 701.25a.
func TestJedeBibliothekshandlungGehtDurch(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	t.Run("mahlen", func(t *testing.T) {
		t.Parallel()

		_, danach, fehler := kasten.Auslegen(einHandlungsstand(t),
			domain.Spielbefehl{
				PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
				Art:      service.BefehlMahlen,
				Nutzlast: map[string]any{"wie": float64(2)},
			})
		if fehler != nil {
			t.Fatalf("mahlen: %v", fehler)
		}

		if oben := danach.Obenauf("sitz-1", 1); len(oben) != 1 ||
			oben[0] != "sitz-1/bib/c" {
			t.Errorf("obenauf liegen %v", oben)
		}
	})

	t.Run("hellsehen", func(t *testing.T) {
		t.Parallel()

		_, danach, fehler := kasten.Auslegen(einHandlungsstand(t),
			domain.Spielbefehl{
				PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
				Art: service.BefehlHellsehen,
				Nutzlast: map[string]any{
					"wie": float64(2), "karten": []any{"sitz-1/bib/a"},
				},
			})
		if fehler != nil {
			t.Fatalf("hellsehen: %v", fehler)
		}

		if oben := danach.Obenauf("sitz-1", 1); len(oben) != 1 ||
			oben[0] != "sitz-1/bib/b" {
			t.Errorf("obenauf liegen %v", oben)
		}
	})

	t.Run("überwachen", func(t *testing.T) {
		t.Parallel()

		ereignisse, _, fehler := kasten.Auslegen(einHandlungsstand(t),
			domain.Spielbefehl{
				PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
				Art: service.BefehlUeberwachen,
				Nutzlast: map[string]any{
					"wie": float64(2), "karten": []any{"sitz-1/bib/a"},
				},
			})
		if fehler != nil {
			t.Fatalf("überwachen: %v", fehler)
		}

		// **Das Ereignis nennt die Zahl und nicht die Karten** — die
		// Bibliothek ist auch für ihren Eigner verdeckt (G1).
		switch {
		case len(ereignisse) != 1:
			t.Fatalf("%d Ereignisse", len(ereignisse))
		case ereignisse[0].Oeffentlich["gewaehlt"] != float64(1):
			t.Errorf("das Ereignis nennt %v", ereignisse[0].Oeffentlich)
		}

		for name := range ereignisse[0].Oeffentlich {
			if strings.Contains(name, "karten") {
				t.Errorf("das Ereignis nennt die Karten: %v",
					ereignisse[0].Oeffentlich)
			}
		}
	})
}

// **Jede Handlung braucht ihre Nutzlast** — §8.
func TestJedeHandlungBrauchtIhreNutzlast(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	faelle := map[string]struct {
		art      string
		nutzlast map[string]any
		wort     string
	}{
		"zerstören ohne Objekt": {
			art: service.BefehlZerstoeren, nutzlast: map[string]any{},
			wort: `"objekt"`,
		},
		"mahlen ohne Zahl": {
			art: service.BefehlMahlen, nutzlast: map[string]any{},
			wort: "wie",
		},
		"hellsehen mit falscher Kartenliste": {
			art: service.BefehlHellsehen,
			nutzlast: map[string]any{
				"wie": float64(2), "karten": "sitz-1/bib/a",
			},
			wort: "keine Liste",
		},
		"hellsehen mit einer Zahl als Karte": {
			art: service.BefehlHellsehen,
			nutzlast: map[string]any{
				"wie": float64(2), "karten": []any{float64(3)},
			},
			wort: "statt einer Kennung",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, _, fehler := kasten.Auslegen(einHandlungsstand(t),
				domain.Spielbefehl{
					PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
					Art: fall.art, Nutzlast: fall.nutzlast,
				}); fehler == nil ||
				!strings.Contains(fehler.Error(), fall.wort) {
				t.Errorf("der Fehler lautet %v", fehler)
			}
		})
	}
}

// einHandlungsstand gibt eine Partie mit allem, was die acht brauchen.
func einHandlungsstand(t *testing.T) domain.Partie {
	t.Helper()

	partie := eineZugpartie(t)

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)

	for stelle := range sitze {
		sitze[stelle].Deck = domain.Partiedeck{
			Abdruck: "abcd",
			Karten: []domain.Abzugskarte{{
				Kennung: "baer", Name: "Bär", Menge: 1,
				Seiten: []domain.Abzugsseite{{Felder: map[string]string{
					"super type": "Creature", "sub type": "Bear",
					"power": "2", "toughness": "2",
				}}},
			}},
		}
	}

	partie.Sitze = sitze

	// **Die Objekte des Aufbaus fallen weg.** Sie liegen in derselben
	// Bibliothek und tragen dieselben Ordnungszahlen; ein Lauf über „die
	// obersten zwei" prüfte sonst, welche von zwei gleichen zuerst kommt —
	// und das ist keine Regel, sondern eine Reihenfolge im Speicher (§3).
	partie.Objekte = nil

	partie.Objekte = append(partie.Objekte,
		domain.Spielobjekt{
			Kennung: "sitz-1/baer/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneTisch, Karte: "baer",
		},
		domain.Spielobjekt{
			Kennung: "sitz-2/baer/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-2", Eigner: "sitz-2",
			Zone: domain.ZoneTisch, Karte: "baer",
		},
		domain.Spielobjekt{
			Kennung: "sitz-1/zauber/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneStapel, Karte: "baer",
		},
		domain.Spielobjekt{
			Kennung: "sitz-1/hand/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneHand, Karte: "baer",
		})

	for stelle := range 4 {
		partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
			Kennung:     "sitz-1/bib/" + string(rune('a'+stelle)),
			Art:         domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneBibliothek, Karte: "baer",
			Ordnung: stelle,
		})
	}

	return partie
}

// **Ohne Kartenliste sortiert Hellsehen nichts um** — CR 701.22a.
//
// **Der Zweig, den dieser Lauf betritt**: das fehlende Feld in
// `kennungsliste`. Wer hellsieht und **nichts** wegsortiert, schickt gar
// nichts mit — und das ist der häufigste Fall, nicht der seltene (§10).
func TestHellsehenOhneKartenlisteSortiertNichts(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, danach, fehler := kasten.Auslegen(einHandlungsstand(t),
		domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art:      service.BefehlHellsehen,
			Nutzlast: map[string]any{"wie": float64(2)},
		})
	if fehler != nil {
		t.Fatalf("hellsehen: %v", fehler)
	}

	if oben := danach.Obenauf("sitz-1", 2); len(oben) != 2 ||
		oben[0] != "sitz-1/bib/a" || oben[1] != "sitz-1/bib/b" {
		t.Errorf("obenauf liegen %v", oben)
	}
}
