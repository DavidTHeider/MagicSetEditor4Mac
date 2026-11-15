package app_test

// **Kontakte aus einer vCard, durch die Bindungen des Fensters** — ID-04
// (#1703).
//
// Gegen den echten Bestand: ein Dateidialog (gestopft) wählt eine Datei, die
// Vorschau nennt jede Karte, die Übernahme schreibt, und ein zweiter Import
// derselben Datei findet alles als „gleich“.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

const adressbuchdatei = "BEGIN:VCARD\r\nVERSION:3.0\r\nFN:Ada Lovelace\r\nUID:ada\r\n" +
	"EMAIL;type=INTERNET:ada@example.org\r\nTEL;type=CELL:+49 170 0\r\n" +
	"PHOTO;ENCODING=b;TYPE=PNG:iVBORw0KGgo=\r\nEND:VCARD\r\n" +
	"BEGIN:VCARD\r\nFN:Bo\r\nUID:bo\r\nEND:VCARD\r\n"

func mitAdressbuchdatei(t *testing.T, anwendung *app.App) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "auswahl.vcf")
	if err := os.WriteFile(pfad, []byte(adressbuchdatei), 0o600); err != nil {
		t.Fatal(err)
	}

	anwendung.MitDateiwahl(func(_ context.Context, auswahl app.Dateiauswahl) (string, error) {
		if auswahl.Muster != "*.vcf" {
			t.Errorf("der Dialog fragt nach %q", auswahl.Muster)
		}

		return pfad, nil
	})

	return pfad
}

func TestKontakteAusEinerVCardUebernehmen(t *testing.T) {
	anwendung := gestartet(t)
	bindungen := app.NeueBindungen(anwendung)
	mitAdressbuchdatei(t, anwendung)

	vorschau := bindungen.KontakteEinlesen(t.Context())
	if vorschau.Fehler != nil || vorschau.Abgebrochen || len(vorschau.Karten) != 2 {
		t.Fatalf("%+v", vorschau)
	}

	ada := vorschau.Karten[0]
	if ada.Abgleich != "neu" || ada.Kontakt.Name != "Ada Lovelace" || len(ada.Kontakt.Wege) != 2 ||
		ada.Kontakt.Bildart != "image/png" || len(ada.Kontakt.Bild) == 0 {
		t.Errorf("%+v", ada)
	}

	// **Nur Ada, und nur ihre Adresse** (AK-01): die Auswahl ist die des
	// Menschen, nicht die der Datei.
	gewaehlt := ada.Kontakt
	gewaehlt.Wege = gewaehlt.Wege[:1]

	uebernahme := bindungen.KontakteUebernehmen(t.Context(), []app.KontaktDTO{gewaehlt})
	if uebernahme.Fehler != nil || uebernahme.Angelegt != 1 || uebernahme.Ersetzt != 0 {
		t.Fatalf("%+v", uebernahme)
	}

	liste := bindungen.Kontakte(t.Context())
	if liste.Fehler != nil || len(liste.Kontakte) != 1 || len(liste.Kontakte[0].Wege) != 1 ||
		liste.Kontakte[0].Uebernommen == "" {
		t.Fatalf("%+v", liste)
	}

	// **Ein zweiter Import derselben Datei legt nichts doppelt an** (AK-02).
	zweite := bindungen.KontakteEinlesen(t.Context())
	if zweite.Karten[0].Abgleich != "geaendert" || zweite.Karten[0].Vorhanden == nil ||
		zweite.Karten[0].Vorhanden.ID != liste.Kontakte[0].ID || zweite.Karten[1].Abgleich != "neu" {
		t.Errorf("%+v", zweite.Karten)
	}

	nochmal := bindungen.KontakteUebernehmen(t.Context(), []app.KontaktDTO{zweite.Karten[0].Kontakt})
	if nochmal.Fehler == nil || nochmal.Fehler.Code != app.CodeKonflikt {
		t.Errorf("eine Dublette ging durch: %+v", nochmal)
	}

	if antwort := bindungen.KontaktLoeschen(t.Context(), liste.Kontakte[0].ID, liste.Kontakte[0].Fassung); antwort.Fehler != nil {
		t.Fatal(antwort.Fehler)
	}

	if leer := bindungen.Kontakte(t.Context()); len(leer.Kontakte) != 0 {
		t.Errorf("nach dem Löschen: %+v", leer)
	}
}

func TestDasEinlesenKenntAbbruchUndFehler(t *testing.T) {
	anwendung := gestartet(t)
	bindungen := app.NeueBindungen(anwendung)

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) { return "", nil })

	if antwort := bindungen.KontakteEinlesen(t.Context()); !antwort.Abgebrochen || antwort.Fehler != nil || antwort.Karten == nil {
		t.Errorf("Abbruch: %+v", antwort)
	}

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return "", errors.New("der Dialog klemmt")
	})

	if antwort := bindungen.KontakteEinlesen(t.Context()); antwort.Fehler == nil {
		t.Errorf("ein klemmender Dialog: %+v", antwort)
	}

	fehlt := filepath.Join(t.TempDir(), "gibt-es-nicht.vcf")
	if antwort := anwendung.KontakteAusDatei(t.Context(), fehlt); antwort.Fehler == nil || antwort.Fehler.Code != app.CodeNichtGefunden {
		t.Errorf("eine fehlende Datei: %+v", antwort)
	}

	// **Eine zu große Datei wird nicht gelesen** — ein Adressbuch mit
	// Bildern ist ein paar Megabyte, keine Hunderte.
	gross := filepath.Join(t.TempDir(), "gross.vcf")
	if err := os.WriteFile(gross, make([]byte, app.Adressbuchgrenze+1), 0o600); err != nil {
		t.Fatal(err)
	}

	if antwort := anwendung.KontakteAusDatei(t.Context(), gross); antwort.Fehler == nil || antwort.Fehler.Code != app.CodeValidierung {
		t.Errorf("eine zu große Datei: %+v", antwort)
	}
}

func TestOhneAdressbuchSagtDieAnwendungEs(t *testing.T) {
	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	if antwort := anwendung.Kontakte(t.Context()); antwort.Fehler == nil || antwort.Kontakte == nil {
		t.Errorf("Liste: %+v", antwort)
	}

	if antwort := anwendung.KontakteAusDatei(t.Context(), "x.vcf"); antwort.Fehler == nil {
		t.Errorf("Einlesen: %+v", antwort)
	}

	if antwort := anwendung.KontakteUebernehmen(t.Context(), nil); antwort.Fehler == nil {
		t.Errorf("Übernehmen: %+v", antwort)
	}

	if antwort := anwendung.KontaktLoeschen(t.Context(), 1, 1); antwort.Fehler == nil {
		t.Errorf("Löschen: %+v", antwort)
	}
}

// klemmendesAdressbuch scheitert an jedem Weg.
type klemmendesAdressbuch struct{}

var errAdressbuchKlemmt = errors.New("das Adressbuch klemmt")

func (klemmendesAdressbuch) Vorschau(context.Context, []byte) (service.Einlesebericht, error) {
	return service.Einlesebericht{}, errAdressbuchKlemmt
}

func (klemmendesAdressbuch) Uebernehmen(context.Context, []domain.Kontakt) (angelegt, ersetzt int, fehler error) {
	return 0, 0, errAdressbuchKlemmt
}

func (klemmendesAdressbuch) Liste(context.Context) ([]domain.Kontakt, error) {
	return nil, errAdressbuchKlemmt
}

func (klemmendesAdressbuch) Loeschen(context.Context, int64, int64) error { return errAdressbuchKlemmt }

func TestJederFehlerwegDesAdressbuchsKommtAn(t *testing.T) {
	anwendung := gestartet(t)
	pfad := mitAdressbuchdatei(t, anwendung)
	anwendung.MitAdressbuch(klemmendesAdressbuch{})

	if antwort := anwendung.KontakteAusDatei(t.Context(), pfad); antwort.Fehler == nil {
		t.Errorf("Vorschau: %+v", antwort)
	}

	if antwort := anwendung.Kontakte(t.Context()); antwort.Fehler == nil || antwort.Kontakte == nil {
		t.Errorf("Liste: %+v", antwort)
	}

	// **Ein Verzeichnis und eine Datei ohne Leserecht** — beides ist keine
	// vCard, und beides sagt es.
	gesperrt := filepath.Join(t.TempDir(), "gesperrt.vcf")
	if err := os.WriteFile(gesperrt, []byte(adressbuchdatei), 0o000); err != nil {
		t.Fatal(err)
	}

	for name, weg := range map[string]string{"ein Verzeichnis": t.TempDir(), "ohne Leserecht": gesperrt} {
		if antwort := gestartet(t).KontakteAusDatei(t.Context(), weg); antwort.Fehler == nil {
			t.Errorf("%s: %+v", name, antwort)
		}
	}
}
