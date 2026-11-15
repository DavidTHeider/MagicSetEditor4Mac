package app

// **Die Liste der eigenen Partien** —
// #1600.
//
// **Eine eigene Datei**, weil `partieprofil_intern_test.go` mit ihr über die
// Zeilengrenze aus §2 ginge — und weil der Schnitt der Frage folgt: dort geht
// es um den Sitz und sein Profil, hier darum, **welche** Partien es gibt.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die eigenen Partien stehen in einer Liste** —
// #1600.
//
// # Wogegen hier geprüft wird
//
// **Dass es den Weg zurück überhaupt gibt.** `storage.Partieablage.Partien`
// stand seit G1 ohne Aufrufer da; der Wirt kam in eine laufende Partie nicht
// zurück, weil sie im Fenster nur im Arbeitsspeicher lebte.
//
// **Und dass die Liste keine Sicht ist**: sie nennt Sitze, Stand und die Zahl
// der Objekte — keinen Kartennamen.
func TestDieEigenenPartienStehenInEinerListe(t *testing.T) {
	t.Parallel()

	gebunden, deckID := eineSpielende(t)

	if leer := gebunden.Partien(t.Context()); len(leer.Partien) != 0 {
		t.Fatalf("ohne Partie kamen %d Zeilen", len(leer.Partien))
	}

	angelegt := gebunden.PartieAnlegen(t.Context(), einEntwurf(deckID))
	if angelegt.Fehler != nil {
		t.Fatalf("PartieAnlegen: %+v", angelegt.Fehler)
	}

	liste := gebunden.Partien(t.Context())
	if liste.Fehler != nil {
		t.Fatalf("Partien: %+v", liste.Fehler)
	}

	if len(liste.Partien) != 1 {
		t.Fatalf("es kamen %d Zeilen", len(liste.Partien))
	}

	zeile := liste.Partien[0]

	switch {
	case zeile.ID != angelegt.Partie.ID:
		t.Errorf("die Kennung ist %q", zeile.ID)
	case zeile.Stand != "laeuft":
		t.Errorf("der Stand ist %q", zeile.Stand)
	case len(zeile.Sitze) != 2:
		t.Errorf("die Sitze sind %+v", zeile.Sitze)
	case zeile.Sitze[0].Kennung != "sitz-1" || zeile.Sitze[0].Name != "David":
		t.Errorf("der erste Sitz ist %+v", zeile.Sitze[0])
	// **Zehn je Sitz** — so viele legt [einDeckImBestand] an.
	case zeile.Objekte != 20:
		t.Errorf("gezählt sind %d Objekte", zeile.Objekte)
	}
}

// **Beide Fehlerwege der Partienliste geben eine leere Liste** —
// #1600.
//
// # Warum zwei und nicht einer
//
// **Weil sie aus zwei Richtungen kommen.** Der eine ist die Rolle: wer nicht
// lesen darf, fragt gar nicht erst nach. Der andere ist der Bestand: er
// antwortet nicht. Beide enden gleich — **leere Liste und kein `nil`** (§8,
// und dieselbe Zusage wie bei `leerePartie`).
func TestBeideFehlerwegeDerPartienlisteGebenEineLeereListe(t *testing.T) {
	t.Parallel()

	t.Run("ohne Erlaubnis", func(t *testing.T) {
		t.Parallel()

		wache := &abweisendeWache{}
		anwendung := NeueApp("")
		anwendung.MitWache(wache)

		antwort := anwendung.Partien(t.Context())

		switch {
		case antwort.Fehler == nil:
			t.Fatal("die Abweisung kam nicht an")
		case antwort.Partien == nil || len(antwort.Partien) != 0:
			t.Errorf("die Liste ist %#v", antwort.Partien)
		case wache.gefragt != domain.WirkungLesen:
			t.Errorf("gefragt wurde nach %q", wache.gefragt)
		}
	})

	t.Run("wenn der Bestand klagt", func(t *testing.T) {
		t.Parallel()

		anwendung := NeueApp("")
		anwendung.profil.Store(1)
		anwendung.MitPartienaht(&klagendePartieleitung{
			weg: "Partien", klagt: errors.New("der Bestand ist zu"),
		})

		antwort := anwendung.Partien(t.Context())

		switch {
		case antwort.Fehler == nil:
			t.Fatal("der Fehler wurde verschluckt")
		case antwort.Partien == nil || len(antwort.Partien) != 0:
			t.Errorf("die Liste ist %#v", antwort.Partien)
		}
	})

	// **Ohne gezeigtes Profil gibt es nichts zu fragen** — und die Frage geht
	// nicht an den Dienst: „dran" ist eine Antwort für ein Profil (UXR-F16,
	// #2031). Das Doppel klagt hier nicht; kämen wir bei ihm an, wäre die
	// Liste nicht leer.
	t.Run("ohne gezeigtes Profil", func(t *testing.T) {
		t.Parallel()

		anwendung := NeueApp("")
		anwendung.MitPartienaht(&klagendePartieleitung{})

		antwort := anwendung.Partien(t.Context())

		switch {
		case antwort.Fehler == nil:
			t.Fatal("ohne Profil kam eine Liste")
		case antwort.Partien == nil || len(antwort.Partien) != 0:
			t.Errorf("die Liste ist %#v", antwort.Partien)
		}
	})
}

// **Und die Fläche ohne Fenster nennt dieselben Partien** (§16).
//
// **Hier und nicht nur in `internal/flaechen`**: jener Lauf fährt die
// **gebauten** Programme und zählt für dieses Paket zu keiner Abdeckung — der
// Weg wäre ungemessen, und ein Tippfehler darin fiele erst an der
// Kommandozeile auf.
func TestDieKopfloseFlaecheNenntDieselbenPartien(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)
	flaeche := PartieflaecheUeber(db)
	deckID := einDeckImBestand(t, db)

	angelegt := flaeche.PartieAusWuenschen(riegelkontext(t), "magic", "anna",
		[]Sitzwunsch{
			{Kennung: "anna", Name: "Anna", DeckID: deckID},
			{Kennung: "bert", Name: "Bert", DeckID: deckID},
		})
	if angelegt.Fehler != nil {
		t.Fatalf("PartieAusWuenschen: %+v", angelegt.Fehler)
	}

	liste := flaeche.Partien(riegelkontext(t))

	switch {
	case liste.Fehler != nil:
		t.Fatalf("Partien: %+v", liste.Fehler)
	case len(liste.Partien) != 1:
		t.Fatalf("es kamen %d Zeilen", len(liste.Partien))
	case liste.Partien[0].ID != angelegt.Partie.ID:
		t.Errorf("die Kennung ist %q", liste.Partien[0].ID)
	}
}
