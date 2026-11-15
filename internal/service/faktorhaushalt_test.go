package service_test

// **Ein Code gilt einmal, und Fehlversuche haben ein Budget** — SEC-09 und
// SEC-10 (#2150,
// #2151,
// Vermerk 0178). Über das Doppel; der Durchstich durch den echten Bestand
// steht in `internal/app/faktorhaushalt_test.go`.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/totp"
)

// eingerichtet gibt einen Dienst mit TOTP-Faktor und den Code der Probezeit.
func eingerichtet(t *testing.T) (*faktordoppel, *service.Zweiterfaktor, string) {
	t.Helper()

	ablage := &faktordoppel{}
	dienst := zweiterfaktor(ablage)

	einrichtung, fehler := dienst.TOTPEinrichten(t.Context(), 1)
	if fehler != nil {
		t.Fatal(fehler)
	}

	zaehler := totp.Zaehler{Geheimnis: geheimnisAus(t, einrichtung.Geheimnis)}

	// **Bestätigt mit dem Code des Schritts davor** — der der Probezeit
	// bleibt frei für die Läufe.
	if fehler := dienst.TOTPBestaetigen(t.Context(), 1,
		totp.Code(zaehler, probezeit.Add(-totp.Schritt))); fehler != nil {
		t.Fatal(fehler)
	}

	return ablage, dienst, totp.Code(zaehler, probezeit)
}

func TestEinTOTPCodeGiltEinmal(t *testing.T) {
	t.Parallel()

	_, dienst, code := eingerichtet(t)

	if fehler := dienst.Pruefe(t.Context(), 1, code); fehler != nil {
		t.Fatalf("das erste Mal: %v", fehler)
	}

	if fehler := dienst.Pruefe(t.Context(), 1, code); fehler == nil {
		t.Error("derselbe Code galt zweimal")
	}
}

func TestDerBestaetigungscodeGiltNichtNochEinmal(t *testing.T) {
	t.Parallel()

	ablage, dienst, _ := eingerichtet(t)

	if ablage.schritt == 0 {
		t.Fatal("die Bestätigung hat ihren Schritt nicht verbraucht")
	}

	zaehler := totp.Zaehler{Geheimnis: geheimnisAus(t, ablage.geheimnis)}

	if fehler := dienst.Pruefe(t.Context(), 1,
		totp.Code(zaehler, probezeit.Add(-totp.Schritt))); fehler == nil {
		t.Error("der Code der Bestätigung galt noch einmal")
	}
}

func TestFuenfFalscheCodesSperrenJedenCodeweg(t *testing.T) {
	t.Parallel()

	_, dienst, code := eingerichtet(t)

	for versuch := range domain.Faktorgrenze {
		if fehler := dienst.Pruefe(t.Context(), 1, "000000"); fehler == nil || istGesperrt(fehler) {
			t.Fatalf("Versuch %d: %v", versuch+1, fehler)
		}
	}

	wege := map[string]func() error{
		"der richtige Code": func() error { return dienst.Pruefe(t.Context(), 1, code) },
		"der Zettel": func() error {
			_, fehler := dienst.CodeFinden(t.Context(), 1, "irgendein-zettel")

			return fehler
		},
	}

	for name, weg := range wege {
		if fehler := weg(); !istGesperrt(fehler) {
			t.Errorf("%s: %v statt der Sperre", name, fehler)
		}
	}
}

func TestEinErfolgSetztDasBudgetZurueck(t *testing.T) {
	t.Parallel()

	ablage, dienst, code := eingerichtet(t)

	for range domain.Faktorgrenze - 1 {
		if fehler := dienst.Pruefe(t.Context(), 1, "000000"); fehler == nil {
			t.Fatal("ein falscher Code galt")
		}
	}

	if fehler := dienst.Pruefe(t.Context(), 1, code); fehler != nil {
		t.Fatalf("der richtige Code: %v", fehler)
	}

	if ablage.versuche != 0 {
		t.Errorf("nach dem Erfolg stehen %d Versuche", ablage.versuche)
	}
}

func TestDerHaushaltReichtSeineFehlerDurch(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("planmäßig: der Haushalt antwortet nicht")

	for name, kaputtmachen := range map[string]func(*faktordoppel){
		"beim Anmelden":     func(f *faktordoppel) { f.haushaltfehler = kaputt },
		"beim Zurücksetzen": func(f *faktordoppel) { f.gelungenfehler = kaputt },
		"beim Verbrauchen":  func(f *faktordoppel) { f.schrittfehler = kaputt },
	} {
		t.Run(name, func(t *testing.T) {
			ablage, dienst, code := eingerichtet(t)
			kaputtmachen(ablage)

			if fehler := dienst.Pruefe(t.Context(), 1, code); !errors.Is(fehler, kaputt) {
				t.Errorf("gekommen ist %v", fehler)
			}
		})
	}
}

func istGesperrt(fehler error) bool {
	var befund *domain.FieldError

	return errors.As(fehler, &befund) && befund.Key == domain.SchluesselFaktorGesperrt
}

func TestEineBestaetigungOhneAblageVerbrauchtNichts(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{anlegefehler: errors.New("planmäßig: die Ablage schreibt nicht")}
	dienst := zweiterfaktor(ablage)

	einrichtung, fehler := dienst.TOTPEinrichten(t.Context(), 1)
	if fehler != nil {
		t.Fatal(fehler)
	}

	code := totp.Code(totp.Zaehler{Geheimnis: geheimnisAus(t, einrichtung.Geheimnis)}, probezeit)

	if fehler := dienst.TOTPBestaetigen(t.Context(), 1, code); !errors.Is(fehler, ablage.anlegefehler) {
		t.Errorf("gekommen ist %v", fehler)
	}

	if ablage.schritt != 0 {
		t.Errorf("verbraucht wurde Schritt %d", ablage.schritt)
	}
}

// **Auch die Bestätigung läuft im Budget** (Zweitblick über F, Befund F2).
func TestDieBestaetigungLaeuftImBudget(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{}
	dienst := zweiterfaktor(ablage)

	einrichtung, fehler := dienst.TOTPEinrichten(t.Context(), 1)
	if fehler != nil {
		t.Fatal(fehler)
	}

	for versuch := range domain.Faktorgrenze {
		if fehler := dienst.TOTPBestaetigen(t.Context(), 1, "000000"); fehler == nil || istGesperrt(fehler) {
			t.Fatalf("Versuch %d: %v", versuch+1, fehler)
		}
	}

	code := totp.Code(totp.Zaehler{Geheimnis: geheimnisAus(t, einrichtung.Geheimnis)}, probezeit)
	if fehler := dienst.TOTPBestaetigen(t.Context(), 1, code); !istGesperrt(fehler) {
		t.Errorf("der richtige Code nach fünf falschen: %v", fehler)
	}
}
