package domain_test

// **Die Fähigkeit der Radioaktivitätsmarken** — CR 728.1, Werkstück M9
// (#1611).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, wie eine Fähigkeit auslöst und auf den Stapel kommt** —
// CR 603.2 und CR 603.3. Hier steht, **was eine bestimmte davon tut**, und
// `ausloeser_test.go` stand mit beidem bei 476 Zeilen (§2).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Radioaktivitätsmarken mahlen in der ersten Hauptphase** — CR 728.1.
func TestStrahlungMahltInDerHauptphase(t *testing.T) {
	t.Parallel()

	partie := eineAusloeserpartie(t)
	partie.Sitze[0].Strahlung = 2

	vorher := partie.Sitze[0].Leben

	danach := bisZurHauptphase(t, partie).Verrechnen()

	// **Zwei Karten gemahlen** — beide Bären, also beides Nichtländer.
	friedhof := 0

	for _, objekt := range danach.Objekte {
		if objekt.Zone == domain.ZoneFriedhof &&
			objekt.Beherrscher == "sitz-1" {
			friedhof++
		}
	}

	if friedhof != 2 {
		t.Errorf("im Friedhof liegen %d Karten", friedhof)
	}

	sitz := danach.SitzVon("sitz-1")
	if sitz == nil {
		t.Fatal("den Sitz gibt es nicht")
	}

	// **Zwei Leben und zwei Marken herunter** — CR 728.1 nennt beide Zahlen
	// als dieselbe.
	if sitz.Leben != vorher-2 {
		t.Errorf("das Leben steht bei %d statt %d", sitz.Leben, vorher-2)
	}

	if sitz.Strahlung != 0 {
		t.Errorf("es liegen noch %d Marken da", sitz.Strahlung)
	}
}

// **Ohne Marken löst nichts aus** — CR 728.1 fragt danach.
func TestOhneMarkenLoestDieStrahlungNichtAus(t *testing.T) {
	t.Parallel()

	danach := bisZurHauptphase(t, eineAusloeserpartie(t))

	for _, objekt := range danach.Objekte {
		if objekt.Art == domain.ObjektFaehigkeit {
			t.Errorf("auf dem Stapel liegt %+v", objekt)
		}
	}
}

// **Ein Land kostet kein Leben** — CR 728.1 zählt die **Nichtländer**.
func TestEinGemahlenesLandKostetKeinLeben(t *testing.T) {
	t.Parallel()

	partie := eineAusloeserpartie(t)
	partie.Sitze[0].Strahlung = 1
	partie.Sitze[0].Deck.Karten = append(partie.Sitze[0].Deck.Karten,
		domain.Abzugskarte{
			Kennung: "wald", Name: "Wald", Menge: 4,
			Seiten: []domain.Abzugsseite{{Felder: map[string]string{
				"name": "Wald", "super type": "Basic Land",
				"sub type": "Forest",
			}}},
		})

	for stelle := range partie.Objekte {
		if partie.Objekte[stelle].Kennung == "sitz-1/bib/a" {
			partie.Objekte[stelle].Karte = "wald"
		}
	}

	vorher := partie.Sitze[0].Leben

	danach := bisZurHauptphase(t, partie).Verrechnen()

	sitz := danach.SitzVon("sitz-1")
	if sitz == nil {
		t.Fatal("den Sitz gibt es nicht")
	}

	if sitz.Leben != vorher {
		t.Errorf("das Leben steht bei %d statt %d", sitz.Leben, vorher)
	}

	// **Und die Marke bleibt liegen** — CR 728.1 nimmt eine je verlorenem
	// Leben.
	if sitz.Strahlung != 1 {
		t.Errorf("es liegen %d Marken da", sitz.Strahlung)
	}
}

// **Eine leere Bibliothek mahlt, was da ist** — CR 701.17a.
func TestEineLeereBibliothekMahltWasDaIst(t *testing.T) {
	t.Parallel()

	partie := eineAusloeserpartie(t)
	partie.Sitze[0].Strahlung = 9

	danach := bisZurHauptphase(t, partie).Verrechnen()

	if sitz := danach.SitzVon("sitz-1"); sitz == nil ||
		sitz.Leben != 20-4 {
		t.Errorf("das Leben steht bei %+v", sitz)
	}
}

// **Wer dem Monarchen mit einem Nichtkreatur-Objekt schadet, löst nicht
// aus** — CR 725.2 nennt eine **Kreatur**.
func TestNurEineKreaturMachtDenMonarchenzug(t *testing.T) {
	t.Parallel()

	partie, fehler := eineAusloeserpartie(t).Monarchwerden("sitz-1")
	if fehler != nil {
		t.Fatalf("monarch: %v", fehler)
	}

	// **Der eigene Bär schadet dem eigenen Monarchen nicht** — und selbst
	// wenn, wäre der Beherrscher schon Monarch (CR 725.2).
	partie.Zug = domain.Zugstand{
		Nummer: 3, Aktiv: "sitz-1", Prioritaet: "sitz-1",
		Abschnitt: domain.Abschnitt{
			Phase: domain.PhaseKampf, Schritt: domain.SchrittBlocker,
		},
		Kampf: domain.Kampfstand{
			Erklaert: true, Geblockt: true,
			Angriffe: []domain.Angriff{
				{Kreatur: "sitz-1/baer/0", Ziel: "sitz-1"},
			},
		},
	}

	danach, fehler := partie.Passen("sitz-1")
	if fehler != nil {
		t.Fatalf("passen: %v", fehler)
	}

	danach, fehler = danach.Passen(danach.Zug.Prioritaet)
	if fehler != nil {
		t.Fatalf("zweites passen: %v", fehler)
	}

	for _, objekt := range danach.Objekte {
		if objekt.Art == domain.ObjektFaehigkeit {
			t.Errorf("auf dem Stapel liegt %+v", objekt)
		}
	}
}

// bisZurHauptphase passt bis in die erste Hauptphase.
func bisZurHauptphase(t *testing.T, partie domain.Partie) domain.Partie {
	t.Helper()

	for range 40 {
		if partie.Zug.Abschnitt.Phase == domain.PhaseHauptVorKampf {
			return partie
		}

		if partie.Zug.Prioritaet == "" {
			partie = partie.Weiter()

			continue
		}

		danach, fehler := partie.Passen(partie.Zug.Prioritaet)
		if fehler != nil {
			t.Fatalf("passen: %v", fehler)
		}

		partie = danach
	}

	t.Fatal("die Hauptphase kam nicht")

	return partie
}

// **Eine Fähigkeit ohne Marken tut nichts** — §8.
//
// **Der Fall entsteht zwischen Auslösen und Verrechnen**: CR 728.1 löst aus,
// wenn Marken daliegen, und bis zum Verrechnen kann sie jemand wegnehmen.
// Eine Fähigkeit, die dann noch mahlte, mahlte ohne Grund.
func TestEineStrahlungOhneMarkenTutNichts(t *testing.T) {
	t.Parallel()

	partie := eineAusloeserpartie(t)
	partie.Sitze[0].Strahlung = 1

	gestapelt := bisZurHauptphase(t, partie)

	// **Die Marken gehen weg, während die Fähigkeit auf dem Stapel liegt.**
	ohne, fehler := gestapelt.Strahlung("sitz-1", -1)
	if fehler != nil {
		t.Fatalf("strahlung: %v", fehler)
	}

	danach := ohne.Verrechnen()

	for _, objekt := range danach.Objekte {
		if objekt.Zone == domain.ZoneFriedhof {
			t.Errorf("gemahlen wurde trotzdem: %+v", objekt)
		}
	}
}

// **Eine Fähigkeit eines Sitzes, den es nicht gibt, kommt nicht auf den
// Stapel** — CR 603.3 mit CR 800.4.
//
// **Die APNAP-Reihenfolge zählt die Sitze**, und wer keiner ist, kommt darin
// nicht vor. Die Fähigkeit fällt still weg — das ist kein Verlust, sondern
// die einzige Antwort auf eine Fähigkeit ohne Beherrscher (§8).
func TestEineFaehigkeitOhneSitzKommtNichtAufDenStapel(t *testing.T) {
	t.Parallel()

	partie := eineAusloeserpartie(t)
	partie.Ausgeloest = []domain.Ausloeser{
		{Art: domain.AusloeserStrahlung, Beherrscher: "weg"},
	}

	danach := partie.Weiter().Verrechnen()

	for _, objekt := range danach.Objekte {
		if objekt.Zone == domain.ZoneFriedhof {
			t.Errorf("gemahlen wurde trotzdem: %+v", objekt)
		}
	}
}
