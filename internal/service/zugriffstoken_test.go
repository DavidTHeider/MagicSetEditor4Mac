package service_test

// **Was ein Zugriffstoken zusagt** — Werkstück 10 von Phase 4.12
// (#1351).
//
// Die Fehlerwege stehen in `zugriffstokenfehler_test.go`; hier steht, was
// gelingt.

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einToken ist die Vorlage, aus der die Fälle ihre Abweichung bauen.
func einToken() domain.Zugriffstoken {
	return domain.Zugriffstoken{
		ProfilID:    7,
		Bezeichnung: "Nachtlauf",
		Scope:       domain.ScopeLesen,
	}
}

func TestEinTokenKommtEinmalHeraus(t *testing.T) {
	t.Parallel()

	ablage := neuesTokendoppel()
	verwaltung := service.NeueTokenverwaltung(ablage)

	angelegt, klartext, fehler := verwaltung.Anlegen(context.Background(), einToken())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	t.Run("der Klartext trägt Marke und Kennung", func(t *testing.T) {
		erwartet := "mse_1_"
		if !strings.HasPrefix(klartext, erwartet) {
			t.Errorf("Klartext %q beginnt nicht mit %q", klartext, erwartet)
		}
	})

	t.Run("der Bestand sieht ihn nie", func(t *testing.T) {
		if ablage.gesehen == "" {
			t.Fatal("die Ablage hat gar nichts zu sehen bekommen")
		}

		if strings.Contains(klartext, ablage.gesehen) ||
			strings.Contains(ablage.gesehen, klartext) {
			t.Errorf("die gehütete Form %q trägt den Klartext %q",
				ablage.gesehen, klartext)
		}
	})

	t.Run("das Angelegte trägt kein Geheimnis", func(t *testing.T) {
		// **Die Zusage steht am Typ und nicht an dieser Zeile**:
		// [domain.Zugriffstoken] hat kein Feld dafür. Was hier geprüft wird,
		// ist, dass der Dienst keines nachträglich unterschiebt — etwa in der
		// Bezeichnung.
		if strings.Contains(angelegt.Bezeichnung, klartext) {
			t.Errorf("die Bezeichnung %q trägt den Klartext", angelegt.Bezeichnung)
		}
	})

	t.Run("der Zeitpunkt des Anlegens steht drin", func(t *testing.T) {
		if _, err := time.Parse(time.RFC3339, angelegt.Angelegt); err != nil {
			t.Errorf("Angelegt = %q, das ist kein RFC 3339: %v",
				angelegt.Angelegt, err)
		}
	})
}

func TestZweiTokenSindNieDasselbe(t *testing.T) {
	t.Parallel()

	verwaltung := service.NeueTokenverwaltung(neuesTokendoppel())

	_, erstes, fehler := verwaltung.Anlegen(context.Background(), einToken())
	if fehler != nil {
		t.Fatalf("erstes Anlegen: %v", fehler)
	}

	_, zweites, fehler := verwaltung.Anlegen(context.Background(), einToken())
	if fehler != nil {
		t.Fatalf("zweites Anlegen: %v", fehler)
	}

	if erstes == zweites {
		t.Fatalf("beide Male %q", erstes)
	}
}

func TestEinTokenWirdWiedererkannt(t *testing.T) {
	t.Parallel()

	verwaltung := service.NeueTokenverwaltung(neuesTokendoppel())

	angelegt, klartext, fehler := verwaltung.Anlegen(context.Background(), einToken())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	erkannt, fehler := verwaltung.Pruefe(
		context.Background(), klartext, domain.ScopeLesen)
	if fehler != nil {
		t.Fatalf("Pruefe: %v", fehler)
	}

	if erkannt.ID != angelegt.ID || erkannt.ProfilID != angelegt.ProfilID {
		t.Errorf("erkannt = %d/%d, angelegt = %d/%d",
			erkannt.ID, erkannt.ProfilID, angelegt.ID, angelegt.ProfilID)
	}
}

func TestUmschliessendeLeerzeichenStoerenNicht(t *testing.T) {
	t.Parallel()

	// **Weil ein Token durch die Zwischenablage reist.** Wer es aus einer
	// Kopfzeile oder einer Datei nimmt, nimmt oft einen Zeilenumbruch mit —
	// und eine Ablehnung dafür wäre eine, die niemand sich erklären kann.
	verwaltung := service.NeueTokenverwaltung(neuesTokendoppel())

	_, klartext, fehler := verwaltung.Anlegen(context.Background(), einToken())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if _, fehler := verwaltung.Pruefe(
		context.Background(), " "+klartext+"\n", domain.ScopeLesen); fehler != nil {
		t.Fatalf("Pruefe: %v", fehler)
	}
}

func TestSchreibenReichtZumLesen(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		hat      domain.Scope
		verlangt domain.Scope
		reicht   bool
	}{
		{"lesen liest", domain.ScopeLesen, domain.ScopeLesen, true},
		{"lesen schreibt nicht", domain.ScopeLesen, domain.ScopeSchreiben, false},
		{"schreiben schreibt", domain.ScopeSchreiben, domain.ScopeSchreiben, true},
		{"schreiben liest auch", domain.ScopeSchreiben, domain.ScopeLesen, true},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			verwaltung := service.NeueTokenverwaltung(neuesTokendoppel())

			neu := einToken()
			neu.Scope = fall.hat

			_, klartext, fehler := verwaltung.Anlegen(context.Background(), neu)
			if fehler != nil {
				t.Fatalf("Anlegen: %v", fehler)
			}

			_, fehler = verwaltung.Pruefe(
				context.Background(), klartext, fall.verlangt)
			if fall.reicht != (fehler == nil) {
				t.Fatalf("reicht = %v, Fehler = %v", fall.reicht, fehler)
			}
		})
	}
}

func TestEineZuKnappeErlaubnisSagtWarum(t *testing.T) {
	t.Parallel()

	// **Der eine Fall, in dem die Meldung mehr sagt als „gilt nicht"** —
	// und der Grund steht am Dienst: wer ein gültiges Token hat, ist kein
	// Fremder, und ohne diesen Satz suchte er den Fehler bei seiner Anfrage.
	verwaltung := service.NeueTokenverwaltung(neuesTokendoppel())

	_, klartext, fehler := verwaltung.Anlegen(context.Background(), einToken())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	_, fehler = verwaltung.Pruefe(
		context.Background(), klartext, domain.ScopeSchreiben)
	if !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("Fehler = %v, erwartet ErrErlaubnis", fehler)
	}

	// **Und ausdrücklich **kein** [domain.ErrValidation].** Wären beide
	// dasselbe, unterschiede eine Fläche, die nur auf jenen prüft, den
	// Berechtigten nicht vom Fremden — und die ganze Ausnahme wäre keine.
	if errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler = %v, ist zugleich ErrValidation", fehler)
	}

	if !strings.Contains(fehler.Error(), "lesen") ||
		!strings.Contains(fehler.Error(), "schreiben") {
		t.Errorf("die Meldung %q nennt nicht beide Erlaubnisse", fehler)
	}
}

func TestNurDieTokenDiesesProfils(t *testing.T) {
	t.Parallel()

	verwaltung := service.NeueTokenverwaltung(neuesTokendoppel())

	meines := einToken()
	if _, _, fehler := verwaltung.Anlegen(
		context.Background(), meines); fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	fremdes := einToken()
	fremdes.ProfilID = 8

	if _, _, fehler := verwaltung.Anlegen(
		context.Background(), fremdes); fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	gefunden, fehler := verwaltung.Alle(context.Background(), 7)
	if fehler != nil {
		t.Fatalf("Alle: %v", fehler)
	}

	if len(gefunden) != 1 || gefunden[0].ProfilID != 7 {
		t.Fatalf("Alle(7) = %+v", gefunden)
	}
}

func TestBezeichnungUndAblaufLassenSichAendern(t *testing.T) {
	t.Parallel()

	verwaltung := service.NeueTokenverwaltung(neuesTokendoppel())

	angelegt, klartext, fehler := verwaltung.Anlegen(
		context.Background(), einToken())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	angelegt.Bezeichnung = "Tageslauf"

	geaendert, fehler := verwaltung.Aendern(context.Background(), angelegt)
	if fehler != nil {
		t.Fatalf("Aendern: %v", fehler)
	}

	if geaendert.Bezeichnung != "Tageslauf" {
		t.Errorf("Bezeichnung = %q", geaendert.Bezeichnung)
	}

	t.Run("und das Geheimnis bleibt dasselbe", func(t *testing.T) {
		// **Die Zusage, für die es diesen Fall gibt**: eine Änderung an der
		// Bezeichnung darf niemandem sein Token entziehen.
		if _, fehler := verwaltung.Pruefe(
			context.Background(), klartext, domain.ScopeLesen); fehler != nil {
			t.Fatalf("Pruefe nach Aendern: %v", fehler)
		}
	})
}

// **Ein Token wechselt seinen Sitz nicht** (#1819) — auch nicht, wenn eine
// Änderung ohne Bindung hereinkommt. Das Fenster liest sie beim Ändern gar
// nicht; ohne diesen Fall stünde ein gebundenes Token nach dem Umbenennen
// dort als ungebunden, während der Bestand es weiter band.
func TestEineAenderungBehaeltDieSitzbindung(t *testing.T) {
	t.Parallel()

	ablage := neuesTokendoppel()
	verwaltung := service.NeueTokenverwaltung(ablage)
	bindung := domain.Sitzbindung{Partie: "p1", Sitz: "anna"}

	gebunden := einToken()
	gebunden.Sitzbindung = bindung

	angelegt, _, fehler := verwaltung.Anlegen(context.Background(), gebunden)
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	angelegt.Bezeichnung = "Umbenannt"
	angelegt.Sitzbindung = domain.Sitzbindung{}

	geaendert, fehler := verwaltung.Aendern(context.Background(), angelegt)
	if fehler != nil || geaendert.Sitzbindung != bindung ||
		ablage.abgelegt[angelegt.ID].Sitzbindung != bindung {
		t.Errorf("zurück %+v, abgelegt %+v (%v)", geaendert.Sitzbindung,
			ablage.abgelegt[angelegt.ID].Sitzbindung, fehler)
	}
}

func TestEinGeloeschtesTokenGiltNichtMehr(t *testing.T) {
	t.Parallel()

	verwaltung := service.NeueTokenverwaltung(neuesTokendoppel())

	angelegt, klartext, fehler := verwaltung.Anlegen(
		context.Background(), einToken())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if fehler := verwaltung.Loeschen(
		context.Background(), angelegt.ID); fehler != nil {
		t.Fatalf("Loeschen: %v", fehler)
	}

	if _, fehler := verwaltung.Pruefe(
		context.Background(), klartext, domain.ScopeLesen); fehler == nil {
		t.Fatal("das gelöschte Token gilt noch")
	}
}
