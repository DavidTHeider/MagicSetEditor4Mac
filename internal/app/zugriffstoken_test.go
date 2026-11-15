// **Zugriffstoken über die Fassade** — Werkstück 10 von Phase 4.12
// (#1351),
// Punkt 13.
package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// tokenfuehrer ist eine Verwaltung im Arbeitsspeicher.
type tokenfuehrer struct {
	zeilen []domain.Zugriffstoken

	// klartext ist der, den [tokenfuehrer.Anlegen] herausgibt.
	klartext string

	// gesehen ist, was zuletzt hineingereicht wurde.
	gesehen domain.Zugriffstoken

	// geloescht ist die Kennung des letzten Löschens.
	geloescht int64

	// **Werkstück O2** (#1426): was rotiert und widerrufen wurde.
	rotiert    int64
	widerrufen int64

	fehler error
}

func (f *tokenfuehrer) Alle(context.Context, int64) ([]domain.Zugriffstoken, error) {
	return f.zeilen, f.fehler
}

func (f *tokenfuehrer) Anlegen(_ context.Context, neu domain.Zugriffstoken) (
	domain.Zugriffstoken, string, error,
) {
	f.gesehen = neu

	if f.fehler != nil {
		return domain.Zugriffstoken{}, "", f.fehler
	}

	neu.ID = 7

	return neu, f.klartext, nil
}

func (f *tokenfuehrer) Aendern(_ context.Context, geaendert domain.Zugriffstoken) (
	domain.Zugriffstoken, error,
) {
	f.gesehen = geaendert

	return geaendert, f.fehler
}

func (f *tokenfuehrer) Loeschen(_ context.Context, tokenID int64) error {
	f.geloescht = tokenID

	return f.fehler
}

// ProfilDes gibt das Profil 1 — die Läufe hier stehen ohne Nutzerverwaltung,
// und dort fragt niemand danach (#1817).
func (f *tokenfuehrer) ProfilDes(context.Context, int64) (int64, error) {
	return 1, f.fehler
}

// **Rotation, Sammelwiderruf und Ablaufwahlen** — Werkstück O2 von Phase 5.2
// (#1426).
func (f *tokenfuehrer) Rotieren(_ context.Context, tokenID int64) (
	domain.Zugriffstoken, string, error,
) {
	f.rotiert = tokenID

	if f.fehler != nil {
		return domain.Zugriffstoken{}, "", f.fehler
	}

	return domain.Zugriffstoken{
		ID: tokenID, ProfilID: 1, Bezeichnung: "Nachtlauf", Principal: "Baurechner",
		Scope: domain.ScopeLesen, VorgaengerBis: "2026-09-08T13:00:00Z", Fassung: 2,
	}, f.klartext, nil
}

func (f *tokenfuehrer) AlleWiderrufen(_ context.Context, profilID int64) (int64, error) {
	f.widerrufen = profilID

	return 2, f.fehler
}

func (f *tokenfuehrer) Ablaufwahlen() []domain.Ablaufwahl {
	return []domain.Ablaufwahl{{Name: "tag", Ablauf: "2026-09-09T12:00:00Z"}, {Name: "nie"}}
}

// mitToken baut eine Anwendung über einer Verwaltung.
func mitToken(fuehrer *tokenfuehrer) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitToken(fuehrer)

	return anwendung
}

// **Jedes Feld kommt an, und die Kennungen reisen als Text.**
func TestEinTokenKommtVollstaendigAn(t *testing.T) {
	anwendung := mitToken(&tokenfuehrer{zeilen: []domain.Zugriffstoken{{
		ID: 7, ProfilID: 1, Bezeichnung: "Nachtlauf",
		Scope: domain.ScopeLesen, Ablauf: "2099-01-01T00:00:00Z",
		Angelegt: "2026-09-05T12:00:00Z", Fassung: 2,
	}}})

	antwort := anwendung.Token(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Token) != 1 {
		t.Fatalf("%+v", antwort.Token)
	}

	eines := antwort.Token[0]

	faelle := map[string]struct{ ist, soll string }{
		"ID":          {eines.ID, "7"},
		"ProfilID":    {eines.ProfilID, "1"},
		"Bezeichnung": {eines.Bezeichnung, "Nachtlauf"},
		"Scope":       {eines.Scope, "lesen"},
		"Ablauf":      {eines.Ablauf, "2099-01-01T00:00:00Z"},
		"Angelegt":    {eines.Angelegt, "2026-09-05T12:00:00Z"},
		"Fassung":     {eines.Fassung, "2"},
	}

	for name, fall := range faelle {
		if fall.ist != fall.soll {
			t.Errorf("%s = %q, erwartet %q", name, fall.ist, fall.soll)
		}
	}

	if eines.Abgelaufen {
		t.Error("ein Token mit Ablauf 2099 gilt als abgelaufen")
	}
}

// **Der Ablauf wird hier ausgerechnet und nicht an der Fläche.**
//
// **Sonst rechnete jede der vier dieselbe Zeitspanne**, und die Ergebnisse
// liefen an der Zeitzone auseinander.
func TestDerAblaufWirdAusgerechnet(t *testing.T) {
	anwendung := mitToken(&tokenfuehrer{zeilen: []domain.Zugriffstoken{
		{ID: 1, Ablauf: "2020-01-01T00:00:00Z"},
		{ID: 2, Ablauf: ""},
	}})

	antwort := anwendung.Token(t.Context(), "1")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if !antwort.Token[0].Abgelaufen {
		t.Error("das Token von 2020 gilt nicht als abgelaufen")
	}

	// **Die Gegenrichtung**: ohne sie stünde der Fall darüber auch dann grün
	// da, wenn jedes Token als abgelaufen gälte (§10).
	if antwort.Token[1].Abgelaufen {
		t.Error("ein Token ohne Ablauf gilt als abgelaufen")
	}
}

// **Eine leere Liste ist eine Liste** (§3) — auch auf jedem Fehlerweg.
func TestEineLeereTokenlisteIstEineListe(t *testing.T) {
	faelle := map[string]struct {
		profilID string
		fuehrer  *tokenfuehrer
	}{
		"nichts da":          {"1", &tokenfuehrer{}},
		"krumme Kennung":     {"keine-zahl", &tokenfuehrer{}},
		"stolpernder Dienst": {"1", &tokenfuehrer{fehler: errors.New("kaputt")}},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			antwort := mitToken(fall.fuehrer).Token(t.Context(), fall.profilID)
			if antwort.Token == nil {
				t.Error("die Liste ist nil")
			}
		})
	}
}

// **Der Klartext kommt einmal heraus — und nur beim Anlegen.**
func TestDerKlartextKommtNurBeimAnlegen(t *testing.T) {
	fuehrer := &tokenfuehrer{klartext: "mse_7_GEHEIM"}
	anwendung := mitToken(fuehrer)

	angelegt := anwendung.TokenAnlegen(t.Context(), app.TokenangabenDTO{
		ProfilID: "1", Bezeichnung: "Nachtlauf", Principal: "Nachtrechner", Scope: "lesen",
		Ablauf: "2099-01-01T00:00:00Z",
	})
	if angelegt.Fehler != nil {
		t.Fatalf("%+v", angelegt.Fehler)
	}

	if angelegt.Klartext != "mse_7_GEHEIM" {
		t.Errorf("Klartext = %q", angelegt.Klartext)
	}

	if angelegt.Token == nil || angelegt.Token.ID != "7" {
		t.Fatalf("%+v", angelegt.Token)
	}

	// **Die Angaben kommen unten an**, und die Profilkennung als Zahl.
	if fuehrer.gesehen.ProfilID != 1 ||
		fuehrer.gesehen.Bezeichnung != "Nachtlauf" ||
		fuehrer.gesehen.Scope != domain.ScopeLesen {
		t.Errorf("gesehen = %+v", fuehrer.gesehen)
	}

	t.Run("und beim Ändern nicht", func(t *testing.T) {
		geaendert := anwendung.TokenAendern(t.Context(), app.TokenangabenDTO{
			ID: "7", Bezeichnung: "Tagesbau", Principal: "Nachtrechner", Scope: "schreiben", Fassung: "1",
		})
		if geaendert.Fehler != nil {
			t.Fatalf("%+v", geaendert.Fehler)
		}

		if geaendert.Klartext != "" {
			t.Errorf("das Ändern gab einen Klartext: %q", geaendert.Klartext)
		}

		if geaendert.Token == nil || geaendert.Token.Bezeichnung != "Tagesbau" {
			t.Fatalf("%+v", geaendert.Token)
		}
	})
}

// **Ein Token wird zurückgezogen.**
func TestEinTokenWirdZurueckgezogen(t *testing.T) {
	fuehrer := &tokenfuehrer{}

	if antwort := mitToken(fuehrer).TokenLoeschen(
		t.Context(), "7"); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if fuehrer.geloescht != 7 {
		t.Errorf("gelöscht wurde %d", fuehrer.geloescht)
	}
}
