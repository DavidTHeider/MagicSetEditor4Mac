package domain_test

// **Eine Einladung wird beim Einladenden eingelöst** — NET-01b
// (#1954).

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func eineNonce() string {
	return base64.RawURLEncoding.EncodeToString([]byte("0123456789abcdef"))
}

func TestEineEinloesungGehtHinUndZurueck(t *testing.T) {
	t.Parallel()

	roh, fehler := domain.Einloesung{Nonce: eineNonce(), Adresse: "192.0.2.7:41234", Name: "Bert"}.Kodiere()
	if fehler != nil {
		t.Fatalf("Kodiere: %v", fehler)
	}

	gelesen, fehler := domain.ZerlegeEinloesung(roh)
	if fehler != nil {
		t.Fatalf("ZerlegeEinloesung: %v", fehler)
	}

	if gelesen.Fassung != domain.Einloesefassung || gelesen.Nonce != eineNonce() ||
		gelesen.Adresse != "192.0.2.7:41234" || gelesen.Name != "Bert" {
		t.Errorf("gelesen: %+v", gelesen)
	}
}

func TestEineKrummeEinloesungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, roh := range map[string]string{
		"kein JSON":        "{",
		"ohne Fassung":     `{"nonce":"` + eineNonce() + `"}`,
		"ohne Nonce":       `{"fassung":1}`,
		"kurze Nonce":      `{"fassung":1,"nonce":"abc"}`,
		"krumme Adresse":   `{"fassung":1,"nonce":"` + eineNonce() + `","adresse":"` + strings.Repeat("x", 300) + `"}`,
		"übergroß":         `{"fassung":1,"nonce":"` + strings.Repeat("a", 5000) + `"}`,
		"negative Fassung": `{"fassung":-1,"nonce":"` + eineNonce() + `"}`,
	} {
		if _, fehler := domain.ZerlegeEinloesung([]byte(roh)); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("%s: %v", name, fehler)
		}
	}

	if _, fehler := (domain.Einloesung{Nonce: "abc"}).Kodiere(); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Kodiere mit krummer Nonce: %v", fehler)
	}
}

// **Eine neuere Fassung wird benannt**, nicht als krumm abgewiesen — sonst
// sähe ein Mensch „ungültig", wo „zu neu" gemeint ist (AK-03).
func TestEineNeuereEinloesefassungWirdBenannt(t *testing.T) {
	t.Parallel()

	_, fehler := domain.ZerlegeEinloesung([]byte(`{"fassung":2,"nonce":"` + eineNonce() + `"}`))
	if !errors.Is(fehler, domain.ErrEinloesefassung) {
		t.Errorf("Fassung 2: %v", fehler)
	}
}

// **Jeder Befund des Einladenden kommt als dasselbe Ergebnis an** — und
// wird drüben wieder zu demselben benannten Fehler.
func TestJederBefundKommtAlsSeinErgebnisAn(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		fehler   error
		ergebnis domain.Einloeseergebnis
	}{
		{nil, domain.EinloesungAufgenommen},
		{domain.ErrEinladungFremd, domain.EinloesungFremd},
		{domain.ErrEinladungAbgelaufen, domain.EinloesungAbgelaufen},
		{domain.ErrEinladungVerbraucht, domain.EinloesungVerbraucht},
		{domain.ErrEinloesungGesperrt, domain.EinloesungGesperrt},
		{domain.ErrEinloesenGedrosselt, domain.EinloesungGedrosselt},
		{domain.ErrEinloesefassung, domain.EinloesungZuNeu},
		{domain.ErrEinloesungUnbeantwortet, domain.EinloesungUnbeantwortet},
		{errors.New("die Platte klemmt"), domain.EinloesungGescheitert},
	} {
		antwort := domain.AntwortAuf(fall.fehler)
		if antwort.Ergebnis != fall.ergebnis || antwort.Fassung != domain.Einloesefassung {
			t.Errorf("%v → %+v", fall.fehler, antwort)

			continue
		}

		gelesen, lesefehler := domain.ZerlegeEinloeseantwort(antwort.Kodiere())
		if lesefehler != nil {
			t.Fatalf("ZerlegeEinloeseantwort: %v", lesefehler)
		}

		zurueck := gelesen.Fehler()

		switch {
		case fall.fehler == nil && zurueck != nil:
			t.Errorf("aufgenommen wurde zu %v", zurueck)
		case fall.ergebnis == domain.EinloesungGescheitert:
			if !errors.Is(zurueck, domain.ErrEinloesungGescheitert) {
				t.Errorf("gescheitert wurde zu %v", zurueck)
			}
		case fall.fehler != nil && !errors.Is(zurueck, fall.fehler):
			t.Errorf("%v wurde zu %v", fall.fehler, zurueck)
		}
	}
}

// **Was der Einladende nicht sagt, sagt er nicht** — eine Antwort ohne
// bekanntes Ergebnis ist eine krumme, und keine stille Zusage.
func TestEineKrummeAntwortIstKeineZusage(t *testing.T) {
	t.Parallel()

	for name, roh := range map[string]string{
		"kein JSON":        "{",
		"fremdes Ergebnis": `{"fassung":1,"ergebnis":"vielleicht"}`,
		"ohne Ergebnis":    `{"fassung":1}`,
		"übergroß":         `{"fassung":1,"ergebnis":"` + strings.Repeat("a", 5000) + `"}`,
	} {
		if _, fehler := domain.ZerlegeEinloeseantwort([]byte(roh)); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("%s: %v", name, fehler)
		}
	}

	if _, fehler := domain.ZerlegeEinloeseantwort([]byte(`{"fassung":2,"ergebnis":"aufgenommen"}`)); !errors.Is(
		fehler, domain.ErrEinloesefassung) {
		t.Errorf("Fassung 2: %v", fehler)
	}
}

// **Die Klassen tragen, was ein Aufrufer tun kann** (§8).
func TestDieEinloesefehlerHabenIhreKlasse(t *testing.T) {
	t.Parallel()

	for fehler, klasse := range map[error]error{
		domain.ErrEinladungFremd:      domain.ErrNotFound,
		domain.ErrEinladungAbgelaufen: domain.ErrValidation,
		domain.ErrEinladungVerbraucht: domain.ErrConflict,
		domain.ErrEinloesungGesperrt:  domain.ErrConflict,
		domain.ErrEinloesenGedrosselt: domain.ErrGedrosselt,
		domain.ErrEinloesefassung:     domain.ErrValidation,
	} {
		if !errors.Is(fehler, klasse) {
			t.Errorf("%v ist kein %v", fehler, klasse)
		}
	}
}

// **Die Einladung sagt selbst, ob sie eingelöst werden kann** — in den
// unterschriebenen Daten. Eine Gegenstelle, die das Einlösen nicht kennt,
// hat nie eine einlösbare ausgestellt; wer annimmt, versucht dann nichts,
// was drüben niemand beantwortet (AK-03).
func TestDieEinladungSagtObSieEinloesbarIst(t *testing.T) {
	t.Parallel()

	einladung := eineKontakteinladung()
	einladung.Einloesbar = true

	code, fehler := einladung.Kodiere(probeUnterschrift)
	if fehler != nil {
		t.Fatalf("Kodiere: %v", fehler)
	}

	gelesen, _, fehler := domain.ZerlegeKontakteinladung(code)
	if fehler != nil || !gelesen.Einloesbar {
		t.Fatalf("gelesen %+v, %v", gelesen, fehler)
	}

	// **Ohne Adresse gibt es keinen Ort**, an dem eingelöst würde.
	einladung.Adresse = ""
	if _, ohneAdresse := einladung.Kodiere(probeUnterschrift); !errors.Is(ohneAdresse, domain.ErrValidation) {
		t.Errorf("einlösbar ohne Adresse: %v", ohneAdresse)
	}

	// **Und wer das Feld nicht setzt, schreibt es nicht** — so sah jede
	// Einladung bis NET-01b aus, und so liest sie eine ältere Fassung.
	alt, fehler := eineKontakteinladung().Kodiere(probeUnterschrift)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if daten, _, _ := strings.Cut(strings.SplitN(alt, ":", 3)[2], "."); strings.Contains(
		string(mussBase64(t, daten)), "einloesbar") {
		t.Errorf("eine gewöhnliche Einladung nennt das Einlösen: %s", mussBase64(t, daten))
	}
}

func mussBase64(t *testing.T, text string) []byte {
	t.Helper()

	roh, fehler := base64.RawURLEncoding.DecodeString(text)
	if fehler != nil {
		t.Fatal(fehler)
	}

	return roh
}

// **Ein Name aus dem Handschlag ist eine Behauptung** — was ein Mensch
// nicht lesen kann, wird nicht übernommen, sondern leer gelassen.
func TestNurEinLesbarerNameWirdUebernommen(t *testing.T) {
	t.Parallel()

	for roh, erwartet := range map[string]string{
		"  Bert  ":              "Bert",
		"":                      "",
		"Be\u202ert":            "",
		strings.Repeat("x", 65): "",
		"\xff":                  "",
	} {
		if gelesen := domain.BrauchbarerName(roh); gelesen != erwartet {
			t.Errorf("%q → %q", roh, gelesen)
		}
	}
}

// **Eine Antwort schreibt die Fassung dieses Programms** — wie die Einlösung,
// auch wenn sie von Hand gebaut wurde (D05, #2236). Ohne das wiese die
// Gegenseite eine Antwort ab, die nur `AntwortAuf` richtig baute.
func TestEineAntwortSchreibtDieEigeneFassung(t *testing.T) {
	t.Parallel()

	roh := domain.Einloeseantwort{Ergebnis: domain.EinloesungAufgenommen}.Kodiere()

	gelesen, fehler := domain.ZerlegeEinloeseantwort(roh)
	if fehler != nil || gelesen.Fassung != domain.Einloesefassung {
		t.Errorf("die Antwort ohne Fassung: %s (%v)", roh, fehler)
	}
}
