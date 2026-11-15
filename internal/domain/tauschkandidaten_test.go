package domain_test

// **Wer was von wem bekommen könnte** — Werkstück `T4` von Phase 5.6
// (#1469).
//
// # Was hier gemessen wird
//
// **Dass ein genauer Wunsch genau bleibt.** Zwei Zeilen mit demselben
// Kartennamen sind nicht dasselbe Angebot; wer aus Namensgleichheit einen
// Treffer machte, böte jemandem etwas an, das er nie gewollt hat.
//
// **Dass ein Angebot nicht zweimal vergeben wird.** Ein Stück Papier liegt
// einmal da. Zwei Wünsche, die beide darauf zeigen, ergeben **einen** Treffer
// und einen Rest — sonst verspräche ein Vorschlag mehr, als der Tisch hergibt.
//
// **Und dass eine fehlende Kennung einen Befund gibt statt einer Vermutung.**

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineAngebotszeile baut eine Zeile, die jemand hergeben kann.
func eineAngebotszeile(anders func(*domain.Angebotszeile)) domain.Angebotszeile {
	zeile := domain.Angebotszeile{
		Druck: "lea-232", Konzept: "konzept-shivan", Menge: 2,
		Zustand: domain.ZustandNearMint, Sprache: "de",
		Finish: domain.FinishNonfoil, Bezeichnung: "Shivan Dragon · LEA #232",
	}

	if anders != nil {
		anders(&zeile)
	}

	return zeile
}

// eineWunschzeile baut einen Wunsch auf denselben Druck.
func eineWunschzeile(anders func(*domain.Wunschzeile)) domain.Wunschzeile {
	zeile := domain.Wunschzeile{
		Druck: "lea-232", Konzept: "konzept-shivan", Menge: 1,
		Zustand: domain.ZustandNearMint, Sprache: "de",
		Finish: domain.FinishNonfoil, Art: domain.WunschGenau,
		Rang: domain.RangNormal, Bezeichnung: "Shivan Dragon · LEA #232",
	}

	if anders != nil {
		anders(&zeile)
	}

	return zeile
}

func TestEinGenauerWunschTrifftGenauDiesenDruck(t *testing.T) {
	t.Parallel()

	treffer, befunde := domain.Tauschkandidaten(
		[]domain.Angebotszeile{eineAngebotszeile(nil)},
		[]domain.Wunschzeile{eineWunschzeile(nil)},
	)

	if len(befunde) != 0 {
		t.Fatalf("es kamen Befunde: %v", befunde)
	}

	if len(treffer) != 1 {
		t.Fatalf("es kamen %d Treffer", len(treffer))
	}

	if treffer[0].Menge != 1 || treffer[0].Fehlend != 0 {
		t.Errorf("es sind %d Stück, %d fehlen", treffer[0].Menge, treffer[0].Fehlend)
	}
}

// **Was nicht übereinstimmt, trifft nicht** — und zwar einzeln geprüft, damit
// ein Treffer nicht aus einem Feld kommt, das gar nicht verglichen wird.
func TestWasNichtUebereinstimmtTrifftNicht(t *testing.T) {
	t.Parallel()

	for name, anders := range map[string]func(*domain.Angebotszeile){
		"ein anderer Druck":      func(a *domain.Angebotszeile) { a.Druck = "m10-123" },
		"eine andere Sprache":    func(a *domain.Angebotszeile) { a.Sprache = "en" },
		"ein anderer Zustand":    func(a *domain.Angebotszeile) { a.Zustand = domain.ZustandPlayed },
		"eine andere Ausführung": func(a *domain.Angebotszeile) { a.Finish = domain.FinishFoil },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			treffer, _ := domain.Tauschkandidaten(
				[]domain.Angebotszeile{eineAngebotszeile(anders)},
				[]domain.Wunschzeile{eineWunschzeile(nil)},
			)

			if len(treffer) != 0 {
				t.Errorf("es kam ein Treffer: %+v", treffer[0])
			}
		})
	}
}

// **Ein druckunabhängiger Wunsch trifft über das Konzept** — und nur darüber.
func TestEinDruckunabhaengigerWunschTrifftUeberDasKonzept(t *testing.T) {
	t.Parallel()

	jederDruck := func(w *domain.Wunschzeile) {
		w.Art, w.Druck = domain.WunschJederDruck, "irgendein-anderer"
	}

	treffer, befunde := domain.Tauschkandidaten(
		[]domain.Angebotszeile{eineAngebotszeile(nil)},
		[]domain.Wunschzeile{eineWunschzeile(jederDruck)},
	)

	if len(befunde) != 0 || len(treffer) != 1 {
		t.Fatalf("es kamen %d Treffer und %d Befunde", len(treffer), len(befunde))
	}

	// **Der Name allein trifft nie.** Dieselbe Bezeichnung, ein anderes
	// Konzept — das ist die stille Namensgleichheit, die dieses Werkstück
	// ausdrücklich nicht will.
	nurDerName := func(a *domain.Angebotszeile) { a.Konzept = "konzept-anderes" }

	ohne, _ := domain.Tauschkandidaten(
		[]domain.Angebotszeile{eineAngebotszeile(nurDerName)},
		[]domain.Wunschzeile{eineWunschzeile(jederDruck)},
	)

	if len(ohne) != 0 {
		t.Errorf("der Name allein hat getroffen: %+v", ohne[0])
	}
}

// **Eine fehlende Kennung ist ein Befund und keine Vermutung.**
func TestOhneKonzeptGibtEsEinenBefund(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		angebot domain.Angebotszeile
		wunsch  domain.Wunschzeile
		ort     string
	}{
		"das Angebot kennt sein Konzept nicht": {
			eineAngebotszeile(func(a *domain.Angebotszeile) { a.Konzept = "" }),
			eineWunschzeile(func(w *domain.Wunschzeile) {
				w.Art, w.Druck = domain.WunschJederDruck, "irgendein-anderer"
			}),
			"Shivan Dragon · LEA #232",
		},
		"der Wunsch kennt seines nicht": {
			eineAngebotszeile(nil),
			eineWunschzeile(func(w *domain.Wunschzeile) {
				w.Art, w.Konzept, w.Bezeichnung = domain.WunschJederDruck, "", "Ein Wunsch ohne Konzept"
			}),
			"Ein Wunsch ohne Konzept",
		},
		"der genaue Wunsch nennt keinen Druck": {
			eineAngebotszeile(nil),
			eineWunschzeile(func(w *domain.Wunschzeile) {
				w.Druck, w.Bezeichnung = "", "Ein Wunsch ohne Druck"
			}),
			"Ein Wunsch ohne Druck",
		},
		"eine Angebotszeile ohne Druck": {
			eineAngebotszeile(func(a *domain.Angebotszeile) { a.Druck, a.Bezeichnung = "", "Zeile ohne Druck" }),
			eineWunschzeile(nil),
			"Zeile ohne Druck",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			treffer, befunde := domain.Tauschkandidaten(
				[]domain.Angebotszeile{fall.angebot}, []domain.Wunschzeile{fall.wunsch})

			if len(treffer) != 0 {
				t.Errorf("es kam ein geratener Treffer: %+v", treffer[0])
			}

			if len(befunde) != 1 {
				t.Fatalf("es kamen %d Befunde: %v", len(befunde), befunde)
			}

			if befunde[0].Ort != fall.ort {
				t.Errorf("der Befund nennt %q", befunde[0].Ort)
			}

			// **Er gehört der Eingabe** und nicht der Portierung — sonst
			// stünde in der Buchführung, dieses Programm könne etwas nicht,
			// was MSE kann.
			if befunde[0].Art != domain.ArtEingabe {
				t.Errorf("die Art ist %q", befunde[0].Art)
			}
		})
	}
}

// **Ein Befund entsteht nur, wo er zählt.** Eine Zeile ohne Konzept neben
// lauter genauen Wünschen ist kein Befund: niemand hat danach gefragt.
func TestEineUnbenutzteLueckeIstKeinBefund(t *testing.T) {
	t.Parallel()

	_, befunde := domain.Tauschkandidaten(
		[]domain.Angebotszeile{eineAngebotszeile(func(a *domain.Angebotszeile) { a.Konzept = "" })},
		[]domain.Wunschzeile{eineWunschzeile(nil)},
	)

	if len(befunde) != 0 {
		t.Errorf("es kamen Befunde: %v", befunde)
	}
}

// **Ein Stück Papier liegt einmal da.**
func TestEinAngebotWirdNichtZweimalVergeben(t *testing.T) {
	t.Parallel()

	eins := func(a *domain.Angebotszeile) { a.Menge = 1 }

	dringend := eineWunschzeile(func(w *domain.Wunschzeile) {
		w.Rang, w.Bezeichnung = domain.RangSehrHoch, "Der dringende"
	})

	beilaeufig := eineWunschzeile(func(w *domain.Wunschzeile) {
		w.Rang, w.Bezeichnung = domain.RangNiedrig, "Der beiläufige"
	})

	treffer, _ := domain.Tauschkandidaten(
		[]domain.Angebotszeile{eineAngebotszeile(eins)},
		// **Der beiläufige steht zuerst** — sonst prüfte der Lauf die
		// Eingabereihenfolge und nicht den Rang.
		[]domain.Wunschzeile{beilaeufig, dringend},
	)

	if len(treffer) != 1 {
		t.Fatalf("es kamen %d Treffer: %+v", len(treffer), treffer)
	}

	if treffer[0].Wunsch.Bezeichnung != "Der dringende" {
		t.Errorf("bedient wurde %q", treffer[0].Wunsch.Bezeichnung)
	}
}

// **Was fehlt, steht dran** — ein halb gedeckter Wunsch ist kein erfüllter.
func TestEinHalbGedeckterWunschNenntSeinenRest(t *testing.T) {
	t.Parallel()

	treffer, _ := domain.Tauschkandidaten(
		[]domain.Angebotszeile{eineAngebotszeile(func(a *domain.Angebotszeile) { a.Menge = 1 })},
		[]domain.Wunschzeile{eineWunschzeile(func(w *domain.Wunschzeile) { w.Menge = 4 })},
	)

	if len(treffer) != 1 {
		t.Fatalf("es kamen %d Treffer", len(treffer))
	}

	if treffer[0].Menge != 1 || treffer[0].Fehlend != 3 {
		t.Errorf("es sind %d Stück, %d fehlen", treffer[0].Menge, treffer[0].Fehlend)
	}
}

// **Zweimal dieselbe Eingabe, zweimal dasselbe Ergebnis** (§10).
func TestDieReihenfolgeIstStabil(t *testing.T) {
	t.Parallel()

	angebot := []domain.Angebotszeile{
		eineAngebotszeile(nil),
		eineAngebotszeile(func(a *domain.Angebotszeile) {
			a.Druck, a.Konzept, a.Bezeichnung = "m10-123", "konzept-bolt", "Lightning Bolt · M10 #123"
		}),
	}

	wuensche := []domain.Wunschzeile{
		eineWunschzeile(nil),
		eineWunschzeile(func(w *domain.Wunschzeile) {
			w.Druck, w.Konzept, w.Bezeichnung = "m10-123", "konzept-bolt", "Lightning Bolt · M10 #123"
			w.Rang = domain.RangHoch
		}),
	}

	erster, _ := domain.Tauschkandidaten(angebot, wuensche)
	zweiter, _ := domain.Tauschkandidaten(angebot, wuensche)

	if len(erster) != 2 || len(zweiter) != 2 {
		t.Fatalf("es kamen %d und %d Treffer", len(erster), len(zweiter))
	}

	for stelle := range erster {
		if erster[stelle].Wunsch.Bezeichnung != zweiter[stelle].Wunsch.Bezeichnung {
			t.Errorf("an Stelle %d stehen %q und %q", stelle,
				erster[stelle].Wunsch.Bezeichnung, zweiter[stelle].Wunsch.Bezeichnung)
		}
	}

	// **Und der dringendere steht vorn** — die Reihenfolge ist die Aussage.
	if erster[0].Wunsch.Rang != domain.RangHoch {
		t.Errorf("vorn steht %q", erster[0].Wunsch.Rang)
	}
}

// **Die Vorgabe und ihr ausgeschriebener Wert sind dasselbe** (0119): „leer"
// heißt `nonfoil`, und beim Zustand `nearMint`. Wer die rohen Felder
// vergliche, machte aus derselben Karte zwei — und der Tausch käme nie
// zustande, ohne dass irgendwo stünde, warum.
func TestDieVorgabeTrifftDenAusgeschriebenenWert(t *testing.T) {
	t.Parallel()

	leer := func(a *domain.Angebotszeile) { a.Finish, a.Zustand = "", "" }

	treffer, befunde := domain.Tauschkandidaten(
		[]domain.Angebotszeile{eineAngebotszeile(leer)},
		[]domain.Wunschzeile{eineWunschzeile(nil)},
	)

	if len(befunde) != 0 {
		t.Fatalf("es kamen Befunde: %v", befunde)
	}

	if len(treffer) != 1 {
		t.Fatalf("es kamen %d Treffer — die Vorgabe hat nicht getroffen", len(treffer))
	}

	// **Und andersherum ebenso** — die Leere darf auf beiden Seiten stehen.
	andersherum, _ := domain.Tauschkandidaten(
		[]domain.Angebotszeile{eineAngebotszeile(nil)},
		[]domain.Wunschzeile{eineWunschzeile(func(w *domain.Wunschzeile) {
			w.Finish, w.Zustand = "", ""
		})},
	)

	if len(andersherum) != 1 {
		t.Errorf("es kamen %d Treffer", len(andersherum))
	}
}
