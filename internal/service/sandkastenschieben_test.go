package service_test

// **Stelle und Zustand beim Schieben** — Werkstück G1 von Phase 5.5
// (#1458).
//
// **Aus `sandkastenzonen_test.go` herausgelöst**, als die Datei über 400
// Zeilen ging (§2). Der Schnitt folgt der Frage: dort, was mit einer **Zone**
// geschieht — ansehen und mischen —, hier, wohin **eine Karte** kommt und in
// welchem Zustand.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Karte kommt an die Stelle, die genannt ist.**
//
// # Warum dieser Lauf hier steht und nicht bei den Objektbefehlen
//
// **Weil er eine Frage über eine Zone beantwortet.** „Oben" und „unten"
// bedeuten nur dort etwas, wo es eine Reihenfolge gibt — und was das für
// welche Zone heißt, steht in [domain.Zone.Geordnet].
func TestEineKarteKommtAnDieGenannteStelle(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		stelle   string
		erwartet int
	}{
		"oben":  {"oben", 1},
		"unten": {"unten", 4},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			vorher := tischMitBibliothek()

			// **Eine Karte, die von der Hand zurückgeht** — sie hat noch
			// keine Ordnung, und genau darum geht es.
			vorher.Objekte = append(vorher.Objekte, domain.Spielobjekt{
				Kennung: "anna/h1", Art: domain.ObjektKarte,
				Beherrscher: "anna", Eigner: "anna",
				Zone: domain.ZoneHand, Karte: "k",
			})

			kasten := service.NeuerSandkasten(festeMischung{})

			_, danach, fehler := kasten.Auslegen(vorher, domain.Spielbefehl{
				PartieID: "p1", Absender: "anna", Kennung: "b1",
				Art: service.BefehlSchieben,
				Nutzlast: map[string]any{
					"objekt": "anna/h1", "zone": "bibliothek",
					"stelle": fall.stelle,
				},
			})
			if fehler != nil {
				t.Fatalf("schieben: %v", fehler)
			}

			pruefeOrdnung(t, danach, "anna/h1", fall.erwartet)

			// **Und die anderen sind mitgerückt statt verdrängt worden.**
			pruefeLueckenloseOrdnung(t, danach, "anna", 4)
		})
	}
}

// **Eine Karte kommt getappt und verdeckt auf den Tisch — in einem Zug.**
//
// **Ein Ereignis und nicht zwei.** Wer erst schöbe und dann tappte, hinterließe
// dazwischen ein Brett, auf dem die Karte ungetappt steht — und jeder, der in
// diesem Augenblick hinsähe, sähe etwas, das nie gemeint war.
func TestEineKarteKommtImGenanntenZustandAn(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	ereignisse, danach, fehler := kasten.Auslegen(tischMitBibliothek(),
		domain.Spielbefehl{
			PartieID: "p1", Absender: "anna", Kennung: "b1",
			Art: service.BefehlSchieben,
			Nutzlast: map[string]any{
				"objekt": "anna/k1", "zone": "tisch",
				"getappt": true, "verdeckt": true,
			},
		})
	if fehler != nil {
		t.Fatalf("schieben: %v", fehler)
	}

	if len(ereignisse) != 1 {
		t.Fatalf("%d Ereignisse, erwartet eines", len(ereignisse))
	}

	for _, objekt := range danach.Objekte {
		if objekt.Kennung != "anna/k1" {
			continue
		}

		if !objekt.Getappt || !objekt.Verdeckt ||
			objekt.Zone != domain.ZoneTisch {
			t.Errorf("%+v", objekt)
		}

		return
	}

	t.Fatal("die Karte ist verschwunden")
}

// **Eine Stelle, die es nicht gibt, wird abgewiesen.**
func TestEineUnbekannteStelleWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, _, fehler := kasten.Auslegen(tischMitBibliothek(), domain.Spielbefehl{
		PartieID: "p1", Absender: "anna", Kennung: "b1",
		Art: service.BefehlSchieben,
		Nutzlast: map[string]any{
			"objekt": "anna/k1", "zone": "bibliothek", "stelle": "mitte",
		},
	})

	if fehler == nil {
		t.Fatal("die Stelle »mitte« kam durch")
	}
}

// **Eine Stelle in einer ungeordneten Zone ist ein Irrtum und kein Wunsch.**
//
// **Abgewiesen statt stillschweigend übergangen**: wer „oben auf den
// Hand" sagt, meint etwas, das es nicht gibt — und eine Antwort, die einfach
// ablegt, ließe ihn glauben, es sei geschehen.
//
// **Hier stand bis zum 2026-09-17 der Friedhof**, und das war falsch:
// CR 400.5 (Fassung 2026-08-19) nennt ihn unter den **geordneten** Zonen,
// und seit G9 (#1592)
// sagt `Zone.Geordnet` dasselbe. Der Lauf darunter prüft jetzt beides.
func TestEineStelleInEinerUngeordnetenZoneWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, _, fehler := kasten.Auslegen(tischMitBibliothek(), domain.Spielbefehl{
		PartieID: "p1", Absender: "anna", Kennung: "b1",
		Art: service.BefehlSchieben,
		Nutzlast: map[string]any{
			"objekt": "anna/k1", "zone": "hand", "stelle": "oben",
		},
	})

	if fehler == nil {
		t.Fatal("»oben auf die Hand« kam durch")
	}
}

// **Zufällig eingesteckt heißt: irgendwohin, und die anderen rücken.**
func TestEineKarteLaesstSichZufaelligEinstecken(t *testing.T) {
	t.Parallel()

	vorher := tischMitBibliothek()

	// **Die eingesetzte Quelle zieht die zweite Stelle von vier.** Gefragt
	// wird nach einer Mischung über alle vier Plätze; genommen wird ihr
	// erster — so braucht diese Naht kein zweites Verfahren für „eine Zahl".
	kasten := service.NeuerSandkasten(festeMischung{reihe: []int{1, 0, 2, 3}})

	vorher.Objekte = append(vorher.Objekte, domain.Spielobjekt{
		Kennung: "anna/h1", Art: domain.ObjektKarte,
		Beherrscher: "anna", Eigner: "anna",
		Zone: domain.ZoneHand, Karte: "k",
	})

	_, danach, fehler := kasten.Auslegen(vorher, domain.Spielbefehl{
		PartieID: "p1", Absender: "anna", Kennung: "b1",
		Art: service.BefehlSchieben,
		Nutzlast: map[string]any{
			"objekt": "anna/h1", "zone": "bibliothek", "stelle": "zufaellig",
		},
	})
	if fehler != nil {
		t.Fatalf("einstecken: %v", fehler)
	}

	pruefeOrdnung(t, danach, "anna/h1", 2)
	pruefeLueckenloseOrdnung(t, danach, "anna", 4)
}

// pruefeOrdnung sichert zu, wo eine Karte gelandet ist.
func pruefeOrdnung(
	t *testing.T, partie domain.Partie, kennung string, erwartet int,
) {
	t.Helper()

	for _, objekt := range partie.Objekte {
		if objekt.Kennung != kennung {
			continue
		}

		if objekt.Ordnung != erwartet {
			t.Errorf("%q liegt an Stelle %d, erwartet %d",
				kennung, objekt.Ordnung, erwartet)
		}

		return
	}

	t.Fatalf("%q ist verschwunden", kennung)
}

// pruefeLueckenloseOrdnung zählt die Bibliothek eines Sitzes durch.
//
// **Lückenlos und ohne Dublette** — sonst hätte ein Einschieben zwei Karten
// auf dieselbe Stelle gelegt, und niemand wüsste, welche oben liegt.
func pruefeLueckenloseOrdnung(
	t *testing.T, partie domain.Partie, sitz string, anzahl int,
) {
	t.Helper()

	gesehen := make([]bool, anzahl+1)

	for _, objekt := range partie.Objekte {
		if objekt.Beherrscher != sitz || objekt.Zone != domain.ZoneBibliothek {
			continue
		}

		if objekt.Ordnung < 1 || objekt.Ordnung > anzahl {
			t.Fatalf("%q liegt an Stelle %d von %d",
				objekt.Kennung, objekt.Ordnung, anzahl)
		}

		if gesehen[objekt.Ordnung] {
			t.Errorf("die Stelle %d ist doppelt belegt", objekt.Ordnung)
		}

		gesehen[objekt.Ordnung] = true
	}

	for stelle := 1; stelle <= anzahl; stelle++ {
		if !gesehen[stelle] {
			t.Errorf("die Stelle %d ist leer", stelle)
		}
	}
}

// **Der Stapel ist eine Reihe für den ganzen Tisch.**
//
// # Der Satz, für den es diesen Lauf gibt
//
// **Zwei Bibliotheken sind zwei Reihen, ein Stapel ist eine.** Wer dort je
// Sitz nummerierte, hätte zwei „oberste" Zauber, und keiner löste sich
// zuerst auf. Das ist die Art Fehler, die still ausfällt: das Brett sieht
// richtig aus, bis zwei Objekte dieselbe Stelle tragen.
func TestDerStapelIstEineReiheFuerDenGanzenTisch(t *testing.T) {
	t.Parallel()

	vorher := tischMitBibliothek()

	// **Berts Zauber liegt schon auf dem Stapel.**
	vorher.Objekte = append(vorher.Objekte, domain.Spielobjekt{
		Kennung: "bert/z1", Art: domain.ObjektKarte,
		Beherrscher: "bert", Eigner: "bert",
		Zone: domain.ZoneStapel, Karte: "z", Ordnung: 1,
	})

	kasten := service.NeuerSandkasten(festeMischung{})

	// **Anna legt ihren darauf** — oben, also über Berts.
	_, danach, fehler := kasten.Auslegen(vorher, domain.Spielbefehl{
		PartieID: "p1", Absender: "anna", Kennung: "b1",
		Art: service.BefehlSchieben,
		Nutzlast: map[string]any{
			"objekt": "anna/k1", "zone": "stapel", "stelle": "oben",
		},
	})
	if fehler != nil {
		t.Fatalf("auf den Stapel legen: %v", fehler)
	}

	pruefeOrdnung(t, danach, "anna/k1", 1)

	// **Und Berts ist gerückt** — nicht verdrängt, nicht doppelt belegt.
	pruefeOrdnung(t, danach, "bert/z1", 2)
}

// **Eine offene geordnete Zone nennt die Stelle, eine verborgene nicht.**
//
// **Der Unterschied ist das Geheimnis selbst**: in einer Bibliothek reicht
// eine Zahl, um den nächsten Zug zu kennen.
func TestDieStelleStehtNurInEinerOffenenZoneImEreignis(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	aufDenStapel, _, fehler := kasten.Auslegen(tischMitBibliothek(),
		domain.Spielbefehl{
			PartieID: "p1", Absender: "anna", Kennung: "b1",
			Art: service.BefehlSchieben,
			Nutzlast: map[string]any{
				"objekt": "anna/k1", "zone": "stapel", "stelle": "oben",
			},
		})
	if fehler != nil {
		t.Fatalf("auf den Stapel: %v", fehler)
	}

	if aufDenStapel[0].Oeffentlich["stelle"] != float64(1) {
		t.Errorf("der Stapel nennt die Stelle nicht: %+v",
			aufDenStapel[0].Oeffentlich)
	}

	inDieBibliothek, _, fehler := kasten.Auslegen(tischMitBibliothek(),
		domain.Spielbefehl{
			PartieID: "p1", Absender: "anna", Kennung: "b1",
			Art: service.BefehlSchieben,
			Nutzlast: map[string]any{
				"objekt": "anna/k1", "zone": "bibliothek", "stelle": "unten",
			},
		})
	if fehler != nil {
		t.Fatalf("in die Bibliothek: %v", fehler)
	}

	if _, dabei := inDieBibliothek[0].Oeffentlich["stelle"]; dabei {
		t.Errorf("die Bibliothek nennt die Stelle: %+v",
			inDieBibliothek[0].Oeffentlich)
	}
}

// **Eine klagende Quelle steckt nichts zufällig ein.**
func TestEineKlagendeQuelleStecktNichtEin(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{klagt: errZufallWeg})

	_, _, fehler := kasten.Auslegen(tischMitBibliothek(), domain.Spielbefehl{
		PartieID: "p1", Absender: "anna", Kennung: "b1",
		Art: service.BefehlSchieben,
		Nutzlast: map[string]any{
			"objekt": "anna/k1", "zone": "bibliothek", "stelle": "zufaellig",
		},
	})

	if fehler == nil {
		t.Fatal("die klagende Quelle kam durch")
	}
}

// **Und eine Mischung, die keine ist, steckt auch nichts ein.**
func TestEineUnbrauchbareMischungStecktNichtEin(t *testing.T) {
	t.Parallel()

	// **Drei Stellen für drei Plätze** — gefragt wird nach dreien, denn das
	// Objekt kommt aus derselben Zone und die anderen zwei bleiben.
	kasten := service.NeuerSandkasten(festeMischung{reihe: []int{9, 9, 9}})

	_, _, fehler := kasten.Auslegen(tischMitBibliothek(), domain.Spielbefehl{
		PartieID: "p1", Absender: "anna", Kennung: "b1",
		Art: service.BefehlSchieben,
		Nutzlast: map[string]any{
			"objekt": "anna/k1", "zone": "bibliothek", "stelle": "zufaellig",
		},
	})

	if fehler == nil {
		t.Fatal("eine Mischung mit dreimal derselben Stelle kam durch")
	}
}
