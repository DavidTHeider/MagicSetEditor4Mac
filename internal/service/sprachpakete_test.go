package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// feldbefund prüft, dass ein Fehler an einem Feld mit einem Schlüssel steht.
func feldbefund(t *testing.T, fehler error, feld string, schluessel domain.Schluessel) {
	t.Helper()

	var befund *domain.FieldError
	if !errors.As(fehler, &befund) || befund.Field != feld || befund.Key != schluessel {
		t.Fatalf("%v statt %s an %q", fehler, schluessel, feld)
	}
}

// **Ein Assistent legt eine Übersetzung über einer Basis an** — LOC-02.AK-01
// (#1688).
//
// **Angelegt heißt: alles fehlt.** Die Basis wird verwiesen und nicht
// kopiert — eine Kopie machte einen unübersetzten Satz ununterscheidbar von
// einem übersetzten.
func TestEineUebersetzungEntstehtUeberIhrerBasis(t *testing.T) {
	t.Parallel()

	ablage := neuesSprachpaketdoppel()
	dienst := sprachdienst(ablage)
	ctx := t.Context()

	auskunft, err := dienst.Anlegen(ctx, "pt-BR", " Português ", "en")
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if auskunft.Paket.Name != "Português" || auskunft.Paket.Stand != domain.StandEntwurf ||
		len(auskunft.Abgleich.Fehlend) != 4 || auskunft.Abgleich.Uebersetzt != 0 {
		t.Errorf("angelegt: %+v", auskunft)
	}

	feldbefund(t, func() error { _, e := dienst.Anlegen(ctx, "pt_BR", "P", "en"); return e }(),
		"kennung", domain.SchluesselSprachkennungUngueltig)
	feldbefund(t, func() error { _, e := dienst.Anlegen(ctx, "ja", "J", "en"); return e }(),
		"kennung", domain.SchluesselSprachkennungEingebaut)
	feldbefund(t, func() error { _, e := dienst.Anlegen(ctx, "da", "Dansk", "da"); return e }(),
		"basis", domain.SchluesselValidierung)

	if _, err := dienst.Anlegen(ctx, "pt-BR", "P", "en"); !errors.Is(err, domain.ErrDoppelt) {
		t.Errorf("zweimal: %v", err)
	}
}

// **Ein Satz wird geprüft, bevor er gilt**, und seine Quelle ist der
// deutsche Satz von heute — AK-02 und AK-03.
func TestEineUebersetzungWirdSatzFuerSatzGeaendert(t *testing.T) {
	t.Parallel()

	ablage := neuesSprachpaketdoppel()
	dienst := sprachdienst(ablage)
	ctx := t.Context()

	if _, err := dienst.Anlegen(ctx, "ru-x-meine", "Мой", "ru"); err != nil {
		t.Fatal(err)
	}

	name, stand, lizenz := "Мой русский", string(domain.StandGeprueft), "CC0-1.0"

	auskunft, err := dienst.Aendern(ctx, "ru-x-meine",
		service.Sprachpaketangaben{Name: &name, Stand: &stand, Lizenz: &lizenz},
		map[string]string{
			"allgemein.titel":                      "Сеты",
			"allgemein.zaehlwoerter.karten.wenige": "карты",
		})
	if err != nil {
		t.Fatalf("ändern: %v", err)
	}

	paket := auskunft.Paket
	if paket.Name != name || paket.Stand != domain.StandGeprueft || paket.Lizenz != lizenz ||
		paket.Texte["allgemein.titel"] != (domain.Uebersetzungstext{Text: "Сеты", Quelle: "Sets"}) ||
		paket.Texte["allgemein.zaehlwoerter.karten.wenige"].Quelle != "Karten" {
		t.Errorf("geändert: %+v", paket)
	}

	// **Leer heißt: zurück zur Basis.**
	auskunft, err = dienst.Aendern(ctx, "ru-x-meine", service.Sprachpaketangaben{},
		map[string]string{"allgemein.titel": ""})
	if _, da := auskunft.Paket.Texte["allgemein.titel"]; err != nil || da {
		t.Errorf("zurückgenommen: %+v, %v", auskunft.Paket.Texte, err)
	}

	// **Ganz oder gar nicht**: ein falscher Satz unter zwei guten schreibt
	// keinen davon.
	vorher := ablage.geschrieben

	_, err = dienst.Aendern(ctx, "ru-x-meine", service.Sprachpaketangaben{}, map[string]string{
		"allgemein.titel": "Сеты",
		"fehler.symbol":   "Файл слишком большой.",
	})
	feldbefund(t, err, "fehler.symbol", domain.SchluesselPlatzhalterAbweichend)

	if ablage.geschrieben != vorher {
		t.Error("ein halber Stand wurde geschrieben")
	}

	falsch := "fertig"
	_, err = dienst.Aendern(ctx, "ru-x-meine", service.Sprachpaketangaben{Stand: &falsch}, nil)
	feldbefund(t, err, "stand", domain.SchluesselValidierung)

	herkunft := "a\x00"
	_, err = dienst.Aendern(ctx, "ru-x-meine", service.Sprachpaketangaben{Herkunft: &herkunft}, nil)
	feldbefund(t, err, "herkunft", domain.SchluesselValidierung)

	if _, err := dienst.Aendern(ctx, "xx", service.Sprachpaketangaben{}, nil); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ein fremdes Paket: %v", err)
	}
}

// **Die Liste nennt, was fehlt, ohne die Sätze mitzuschicken** — bei acht
// Paketen wären es sonst dreißigtausend.
func TestDieListeDerUebersetzungen(t *testing.T) {
	t.Parallel()

	ablage := neuesSprachpaketdoppel()
	dienst := sprachdienst(ablage)
	ctx := t.Context()

	if _, err := dienst.Anlegen(ctx, "da", "Dansk", "en"); err != nil {
		t.Fatal(err)
	}

	if _, err := dienst.Aendern(ctx, "da", service.Sprachpaketangaben{},
		map[string]string{"allgemein.titel": "Sæt"}); err != nil {
		t.Fatal(err)
	}

	liste, err := dienst.Liste(ctx)
	if err != nil || len(liste) != 1 || liste[0].Paket.Texte != nil ||
		liste[0].Abgleich.Uebersetzt != 1 || len(liste[0].Abgleich.Fehlend) != 3 {
		t.Errorf("die Liste: %+v, %v", liste, err)
	}

	auskunft, err := dienst.Zeige(ctx, "da")
	if err != nil || auskunft.Paket.Texte["allgemein.titel"].Text != "Sæt" {
		t.Errorf("gezeigt: %+v, %v", auskunft, err)
	}

	if err := dienst.Loeschen(ctx, "da"); err != nil {
		t.Fatal(err)
	}

	if _, err := dienst.Zeige(ctx, "da"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("nach dem Löschen: %v", err)
	}

	ablage.fehler = errors.New("der Bestand schweigt")

	if _, err := dienst.Liste(ctx); err == nil {
		t.Error("der Fehler der Liste ging verloren")
	}
}
