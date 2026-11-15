package storage_test

// **Vier Kontoarten im Bestand** — Werkstück 6 von Phase 4.12
// (#1351).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein Nutzer mit Adresse, Nummer und Kennwort kommt zurück, wie er hinging.**
func TestEinNutzerMitKennungenKommtZurueck(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	angelegt, fehler := ablage.NutzerAnlegen(t.Context(), domain.Nutzer{
		Name:       "Zweite",
		Email:      "wer@example.com",
		Zweitemail: "sicher@example.com",
		Telefon:    "+4915112345678",
		Kennwort:   "$argon2id$…",
	})
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	gelesen, lesefehler := ablage.Nutzer(t.Context())
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	var gefunden domain.Nutzer

	for _, einer := range gelesen {
		if einer.ID == angelegt.ID {
			gefunden = einer
		}
	}

	if gefunden.Email != "wer@example.com" || gefunden.Telefon != "+4915112345678" ||
		gefunden.Zweitemail != "sicher@example.com" || gefunden.Kennwort != "$argon2id$…" {
		t.Errorf("gelesen %+v", gefunden)
	}
}

// **Zwei Nutzer ohne Adresse gehen** — und zwei mit derselben nicht.
//
// # Der Teilindex ist der ganze Punkt
//
// **Ohne `WHERE email <> ”` gäbe es genau einen Nutzer ohne Adresse** — und
// „ohne Konto loslegen" wäre nach dem ersten weiteren Nutzer vorbei.
func TestZweiNutzerOhneAdresseGehen(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	if _, fehler := ablage.NutzerAnlegen(t.Context(),
		domain.Nutzer{Name: "Zweite"}); fehler != nil {
		t.Fatalf("der zweite ohne Adresse: %v", fehler)
	}

	if _, fehler := ablage.NutzerAnlegen(t.Context(),
		domain.Nutzer{Name: "Dritte"}); fehler != nil {
		t.Errorf("der dritte ohne Adresse: %v", fehler)
	}
}

// **Dieselbe Adresse zweimal ist ein Konflikt.**
func TestDieselbeAdresseZweimalIstEinKonflikt(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	if _, fehler := ablage.NutzerAnlegen(t.Context(), domain.Nutzer{
		Name: "Zweite", Email: "wer@example.com",
	}); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	_, zweitfehler := ablage.NutzerAnlegen(t.Context(), domain.Nutzer{
		Name: "Dritte", Email: "wer@example.com",
	})

	if !errors.Is(zweitfehler, domain.ErrConflict) {
		t.Errorf("Fehler %v, erwartet war ErrConflict", zweitfehler)
	}
}

// **Und dieselbe Nummer ebenso.**
func TestDieselbeNummerZweimalIstEinKonflikt(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	if _, fehler := ablage.NutzerAnlegen(t.Context(), domain.Nutzer{
		Name: "Zweite", Telefon: "+4915112345678",
	}); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	_, zweitfehler := ablage.NutzerAnlegen(t.Context(), domain.Nutzer{
		Name: "Dritte", Telefon: "+4915112345678",
	})

	if !errors.Is(zweitfehler, domain.ErrConflict) {
		t.Errorf("Fehler %v, erwartet war ErrConflict", zweitfehler)
	}
}

// **Ein Nutzer lässt sich über seine Adresse finden.**
func TestEinNutzerLaesstSichUeberSeineAdresseFinden(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	angelegt, fehler := ablage.NutzerAnlegen(t.Context(), domain.Nutzer{
		Name: "Zweite", Email: "wer@example.com", Kennwort: "$argon2id$…",
	})
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	gefunden, lesefehler := ablage.NutzerZuKennung(t.Context(), "wer@example.com")
	if lesefehler != nil {
		t.Fatalf("suchen: %v", lesefehler)
	}

	if gefunden.ID != angelegt.ID || gefunden.Kennwort != "$argon2id$…" {
		t.Errorf("gefunden %+v", gefunden)
	}
}

// **Und über seine Nummer.**
func TestEinNutzerLaesstSichUeberSeineNummerFinden(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	angelegt, fehler := ablage.NutzerAnlegen(t.Context(), domain.Nutzer{
		Name: "Zweite", Telefon: "+4915112345678",
	})
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	gefunden, lesefehler := ablage.NutzerZuKennung(t.Context(), "+4915112345678")
	if lesefehler != nil {
		t.Fatalf("suchen: %v", lesefehler)
	}

	if gefunden.ID != angelegt.ID {
		t.Errorf("gefunden %+v", gefunden)
	}
}

// **Eine Kennung, die niemandem gehört, ist „nicht gefunden".**
//
// **Und eine leere ebenso** — sonst fände sie jeden Nutzer ohne Adresse.
func TestEineUnbekannteKennungIstNichtGefunden(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	for _, kennung := range []string{"gibtsnicht@example.com", ""} {
		_, fehler := ablage.NutzerZuKennung(t.Context(), kennung)
		if !errors.Is(fehler, domain.ErrNotFound) {
			t.Errorf("%q: %v, erwartet war ErrNotFound", kennung, fehler)
		}
	}
}

// **Ein anonymer Nutzer trägt seinen Vermerk.**
func TestEinAnonymerNutzerTraegtSeinenVermerk(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	angelegt, fehler := ablage.NutzerAnlegen(t.Context(),
		domain.Nutzer{Name: "Gast-4711", Anonym: true})
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	gelesen, lesefehler := ablage.Nutzer(t.Context())
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	for _, einer := range gelesen {
		if einer.ID == angelegt.ID && !einer.Anonym {
			t.Error("der Vermerk fehlt")
		}
	}
}

// **Der erste Nutzer hat keine Kennungen — „ohne Konto loslegen".**
//
// # Das ist die Abnahme aus der Kommentierung
//
// „Es muss auch möglich [sein], ohne Konto loszulegen (s. Standard-Nutzer)."
// Sie ist seit Wanderung 0034 wahr, und dieser Lauf hält sie: wer nichts tut,
// hat einen Nutzer, ein Profil und keine einzige Angabe über sich.
func TestDerErsteNutzerHatKeineKennungen(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	gelesen, fehler := ablage.Nutzer(t.Context())
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(gelesen) != 1 {
		t.Fatalf("%d Nutzer", len(gelesen))
	}

	erster := gelesen[0]

	if erster.Email != "" || erster.Telefon != "" || erster.Kennwort != "" ||
		erster.Anonym {
		t.Errorf("der erste Nutzer trägt Angaben: %+v", erster)
	}

	if strings.TrimSpace(erster.Name) == "" {
		t.Error("der erste Nutzer hat keinen Namen")
	}
}
