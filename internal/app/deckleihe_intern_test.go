package app

// **Die digitale Deckleihe an der Fassade** — Werkstück G6 von Phase 5.5
// (#1463).
//
// # Wogegen hier geprüft wird
//
//   - Der ganze Weg: **anbieten → Umschlag → annehmen** über zwei Profile.
//   - Ein **verweigerter Import lässt nichts zurück** — weder eine Leihe noch
//     eine halbe Datei.
//   - Jeder Weg fragt nach **seiner** Wirkung.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/austausch"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineLeihanwendung gibt eine Anwendung mit Bestand, Profil und Peeridentität.
//
// **Mit einem Probebund und nicht mit dem Anmeldebund dieses Rechners** (§3):
// der weist ein Testbinärprogramm ab, und täte er es nicht, öffnete er einen
// Dialog.
func eineLeihanwendung(t *testing.T) *App {
	t.Helper()

	anwendung := NeueApp("")
	anwendung.db = bestandFuerDenRiegel(t)
	// **Mit den Einstellungen, wie `verdrahtung.go` sie setzt** — die
	// Handelsrichtlinie liest sie, und ohne sie verleiht niemand (#2435).
	anwendung.einstellungen = einstellungsverwalterUeber(anwendung.db)
	anwendung.profil.Store(1)
	anwendung.MitPeerbund(&peerprobebund{})

	if antwort := anwendung.PeeridentitaetSicherstellen(
		riegelkontext(t)); antwort.Fehler != nil {
		t.Fatalf("die Peeridentität: %+v", antwort.Fehler)
	}

	return anwendung
}

// **Ohne Leihe ist die Liste leer und kein `nil`.**
func TestOhneLeiheIstDieListeLeer(t *testing.T) {
	t.Parallel()

	antwort := eineLeihanwendung(t).Deckleihen(riegelkontext(t))

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	case antwort.Leihen == nil:
		t.Error("die Liste ist nil statt leer")
	case len(antwort.Leihen) != 0:
		t.Errorf("erwartet nichts, bekam %v", antwort.Leihen)
	}
}

// **Jeder Weg fragt nach seiner Wirkung.**
func TestJederLeihwegFragtNachSeinerWirkung(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		fahre   func(*App) DeckleiheAntwort
		wirkung domain.Wirkung
	}{
		"lesen": {
			fahre: func(a *App) DeckleiheAntwort {
				return a.Deckleihen(riegelkontext(t))
			},
			wirkung: domain.WirkungLesen,
		},
		"anbieten": {
			fahre: func(a *App) DeckleiheAntwort {
				return a.DeckleiheAnbieten(riegelkontext(t), LeihwunschDTO{}, "")
			},
			wirkung: domain.WirkungExternSenden,
		},
		"annehmen": {
			fahre: func(a *App) DeckleiheAntwort {
				return a.DeckleiheAnnehmen(riegelkontext(t), "")
			},
			wirkung: domain.WirkungExternEmpfangen,
		},
		"widerrufen": {
			fahre: func(a *App) DeckleiheAntwort {
				return a.DeckleiheWiderrufen(riegelkontext(t), "l1")
			},
			wirkung: domain.WirkungSchreiben,
		},
		"zurückgeben": {
			fahre: func(a *App) DeckleiheAntwort {
				return a.DeckleiheZurueckgeben(riegelkontext(t), "l1")
			},
			wirkung: domain.WirkungSchreiben,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wache := &abweisendeWache{}
			anwendung := eineLeihanwendung(t)

			anwendung.MitWache(wache)

			antwort := fall.fahre(anwendung)

			switch {
			case antwort.Fehler == nil:
				t.Fatal("die Abweisung kam nicht an")
			case antwort.Leihen == nil:
				t.Error("die Liste ist nil statt leer")
			case wache.gefragt != fall.wirkung:
				t.Errorf("gefragt wurde nach %q", wache.gefragt)
			}
		})
	}
}

// **Der ganze Weg: anbieten, ablegen, annehmen.**
func TestEineLeiheReistUeberDenUmschlag(t *testing.T) {
	t.Parallel()

	verleiher := eineLeihanwendung(t)
	entleiher := eineLeihanwendung(t)

	deckID := einDeckImBestand(t, verleiher.db)

	bert := entleiher.Peeridentitaet(riegelkontext(t))
	if bert.Fehler != nil {
		t.Fatalf("die Kennung des Entleihers: %+v", bert.Fehler)
	}

	ziel := filepath.Join(t.TempDir(), "leihe.msepackage")

	angeboten := verleiher.DeckleiheAnbieten(riegelkontext(t), LeihwunschDTO{
		Entleiher: bert.Kennung, DeckID: deckID,
		Bindungsart: "partie", Woran: "p1",
		Rechte: []string{"spielen"}, Stunden: 24, Modus: "zuende",
	}, ziel)

	if angeboten.Fehler != nil {
		t.Fatalf("anbieten: %+v", angeboten.Fehler)
	}

	if len(angeboten.Leihen) != 1 || !angeboten.Leihen[0].Verliehen {
		t.Fatalf("die Leihe beim Verleiher ist %+v", angeboten.Leihen)
	}

	if angeboten.Leihen[0].Karten == 0 {
		t.Error("der Abzug ist leer")
	}

	genommen := entleiher.DeckleiheAnnehmen(riegelkontext(t), ziel)
	if genommen.Fehler != nil {
		t.Fatalf("annehmen: %+v", genommen.Fehler)
	}

	switch {
	case len(genommen.Leihen) != 1:
		t.Fatalf("die Leihe beim Entleiher ist %+v", genommen.Leihen)
	case genommen.Leihen[0].Verliehen:
		t.Error("die angenommene Leihe steht als verliehen")
	case genommen.Leihen[0].Abdruck != angeboten.Leihen[0].Abdruck:
		t.Errorf("die Abdrücke sind %q und %q",
			genommen.Leihen[0].Abdruck, angeboten.Leihen[0].Abdruck)
	}
}

// **Ein fremdes Profil nimmt sie nicht an — und es bleibt nichts liegen.**
func TestEineFremdeLeiheLaesstNichtsZurueck(t *testing.T) {
	t.Parallel()

	verleiher := eineLeihanwendung(t)
	entleiher := eineLeihanwendung(t)
	dritter := eineLeihanwendung(t)

	deckID := einDeckImBestand(t, verleiher.db)

	bert := entleiher.Peeridentitaet(riegelkontext(t))
	ziel := filepath.Join(t.TempDir(), "leihe.msepackage")

	if antwort := verleiher.DeckleiheAnbieten(riegelkontext(t), LeihwunschDTO{
		Entleiher: bert.Kennung, DeckID: deckID,
		Bindungsart: "partie", Woran: "p1",
		Rechte: []string{"spielen"}, Stunden: 24, Modus: "zuende",
	}, ziel); antwort.Fehler != nil {
		t.Fatalf("anbieten: %+v", antwort.Fehler)
	}

	antwort := dritter.DeckleiheAnnehmen(riegelkontext(t), ziel)

	if antwort.Fehler == nil {
		t.Fatal("die fremde Leihe wurde angenommen")
	}

	stand := dritter.Deckleihen(riegelkontext(t))
	if len(stand.Leihen) != 0 {
		t.Fatalf("es blieb eine Leihe liegen: %+v", stand.Leihen)
	}
}

// **Ein Umschlag ohne Leihe wird gemeldet und nicht angenommen.**
func TestEinUmschlagOhneLeiheWirdGemeldet(t *testing.T) {
	t.Parallel()

	anwendung := eineLeihanwendung(t)

	leer := filepath.Join(t.TempDir(), "leer.msepackage")

	if fehler := os.WriteFile(leer, []byte("kein Umschlag"), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	antwort := anwendung.DeckleiheAnnehmen(riegelkontext(t), leer)

	switch {
	case antwort.Fehler == nil:
		t.Fatal("der falsche Umschlag kam durch")
	case antwort.Leihen == nil:
		t.Error("die Liste ist nil statt leer")
	}
}

// **Jeder Fehlerweg des Leihens wird betreten** (§3).
func TestJederFehlerwegDerLeiheWirdBetreten(t *testing.T) {
	t.Parallel()

	anwendung := eineLeihanwendung(t)
	ctx := riegelkontext(t)

	// **Eine Deckkennung, die keine ist.**
	if antwort := anwendung.DeckleiheAnbieten(ctx, LeihwunschDTO{
		Entleiher: "bert", DeckID: "keine Zahl",
	}, ""); antwort.Fehler == nil {
		t.Error("die krumme Deckkennung kam durch")
	}

	// **Ein Wunsch, aus dem keine gültige Leihe wird.**
	deckID := einDeckImBestand(t, anwendung.db)

	if antwort := anwendung.DeckleiheAnbieten(ctx, LeihwunschDTO{
		Entleiher: "bert", DeckID: deckID, Bindungsart: "sammlung",
	}, ""); antwort.Fehler == nil {
		t.Error("der ungültige Wunsch kam durch")
	}

	// **Eine Dauer, die keine Dauer fasst** (J05, #2269): 5.124.096 Stunden
	// liefen in Go zu 25 Minuten um, und die verkürzte Leihe wurde
	// unterschrieben. Abgewiesen wird am Feld, das sie verursacht.
	antwort := anwendung.DeckleiheAnbieten(ctx, LeihwunschDTO{
		Entleiher: "12D3KooWBert", DeckID: deckID,
		Bindungsart: "partie", Rechte: []string{"spielen"},
		Stunden: 5124096, Modus: "zuende",
	}, filepath.Join(t.TempDir(), "leihe.msepackage"))
	if antwort.Fehler == nil || antwort.Fehler.Feld != "stunden" {
		t.Errorf("die übergroße Dauer kam durch oder steht nicht am Feld: %+v", antwort.Fehler)
	}

	// **Ein Ziel, an das sich nicht schreiben lässt.**
	if antwort := anwendung.DeckleiheAnbieten(ctx, LeihwunschDTO{
		Entleiher: "12D3KooWBert", DeckID: deckID,
		Bindungsart: "partie", Rechte: []string{"spielen"},
		Stunden: 24, Modus: "zuende",
	}, "/gibt/es/nicht/leihe.msepackage"); antwort.Fehler == nil {
		t.Error("das unschreibbare Ziel kam durch")
	}

	// **Ein Umschlag, den es nicht gibt.**
	if antwort := anwendung.DeckleiheAnnehmen(
		ctx, "/gibt/es/nicht.msepackage"); antwort.Fehler == nil {
		t.Error("der fehlende Umschlag kam durch")
	}

	// **Eine Leihe, die es nicht gibt.**
	for name, antwort := range map[string]DeckleiheAntwort{
		"widerrufen":  anwendung.DeckleiheWiderrufen(ctx, "gibtesnicht"),
		"zurückgeben": anwendung.DeckleiheZurueckgeben(ctx, "gibtesnicht"),
	} {
		if antwort.Fehler == nil {
			t.Errorf("%s: die fehlende Leihe kam durch", name)
		}
	}
}

// **Ohne Peeridentität geht keine Leihe hinaus.**
func TestOhneIdentitaetGehtKeineLeiheHinaus(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.db = bestandFuerDenRiegel(t)
	anwendung.profil.Store(1)
	anwendung.MitPeerbund(&peerprobebund{})

	deckID := einDeckImBestand(t, anwendung.db)

	antwort := anwendung.DeckleiheAnbieten(riegelkontext(t), LeihwunschDTO{
		Entleiher: "12D3KooWBert", DeckID: deckID,
		Bindungsart: "partie", Rechte: []string{"spielen"},
		Stunden: 24, Modus: "zuende",
	}, filepath.Join(t.TempDir(), "l.msepackage"))

	if antwort.Fehler == nil {
		t.Fatal("ohne Identität wurde verliehen")
	}
}

// **Ohne Wegwerfort meldet sich das Annehmen.**
//
// **Gemessen über `TMPDIR`**, denn [os.MkdirTemp] mit leerem ersten Argument
// fragt [os.TempDir], und das liest die Umgebung — dieselbe Form wie beim
// Umschlag.
func TestOhneWegwerfortMeldetSichDasAnnehmen(t *testing.T) {
	anwendung := eineLeihanwendung(t)

	t.Setenv("TMPDIR", filepath.Join(t.TempDir(), "gibtesnicht"))

	if antwort := anwendung.DeckleiheAnnehmen(
		riegelkontext(t), "egal"); antwort.Fehler == nil {
		t.Fatal("ohne Wegwerfort wurde angenommen")
	}
}

// **Ein Umschlag ohne die beiden Objekte wird gemeldet.**
func TestEinUmschlagOhneDieBeidenObjekteWirdGemeldet(t *testing.T) {
	t.Parallel()

	anwendung := eineLeihanwendung(t)
	ctx := riegelkontext(t)

	faelle := map[string][]austausch.Inhalt{
		"ohne Abzug": {{
			Art: austausch.ArtLeihe, Name: Leihdokumentname,
			Bezeichnung: "Leihe", Bytes: []byte(`{}`),
		}},
		"mit unlesbarem Dokument": {
			{
				Art: austausch.ArtLeihe, Name: Leihdokumentname,
				Bezeichnung: "Leihe", Bytes: []byte(`{`),
			},
			{
				Art: austausch.ArtLeihe, Name: Leihabzugname,
				Bezeichnung: "Abzug", Bytes: []byte(`{}`),
			},
		},
	}

	ich, fehler := anwendung.peerwerk().Ausweis(ctx)
	if fehler != nil {
		t.Fatalf("Ausweis: %v", fehler)
	}

	for name, inhalte := range faelle {
		t.Run(name, func(t *testing.T) {
			ziel := filepath.Join(t.TempDir(), "halb.msepackage")

			if _, fehler := austausch.Schreibe(
				ziel, inhalte, ich); fehler != nil {
				t.Fatalf("schreiben: %v", fehler)
			}

			if antwort := anwendung.DeckleiheAnnehmen(
				ctx, ziel); antwort.Fehler == nil {
				t.Fatal("der halbe Umschlag kam durch")
			}
		})
	}
}
