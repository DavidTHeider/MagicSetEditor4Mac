package service_test

// **Wer in der Nähe ist** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Was hier gemessen wird
//
// **Dass ein Fund kein Kontakt ist.** Wer sich meldet, sagt eine Kennung —
// mehr nicht. Ob dieses Profil ihn kennt, ob ein Mensch seinen Abdruck
// verglichen hat und was er darf, steht in der Kontaktliste, und das
// zusammenzubringen ist eine **fachliche** Frage und keine des Netzes.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// nachbardoppel gibt vorbereitete Funde.
type nachbardoppel struct {
	funde     []service.Nachbarfund
	verworfen int
	fehler    error
}

func (n nachbardoppel) Suche(context.Context) (service.Nachbarfunde, error) {
	if n.fehler != nil {
		return service.Nachbarfunde{}, n.fehler
	}

	return service.Nachbarfunde{Funde: n.funde, Verworfen: n.verworfen}, nil
}

// **Ein unbekannter Fund steht als unbekannt da.**
func TestEinUnbekannterFundStehtAlsUnbekanntDa(t *testing.T) {
	t.Parallel()

	werk := service.NeuesNachbarwerk(
		nachbardoppel{funde: []service.Nachbarfund{{
			Kennung: "bfremd", Name: "Annas Mac", Adresse: "192.168.1.7:41234",
		}}},
		service.NeuesKontaktwerk(einKontaktdoppel(), einePeeruhr))

	nachbarn, fehler := werk.Liste(t.Context())
	if fehler != nil {
		t.Fatalf("Liste: %v", fehler)
	}

	if len(nachbarn.Nachbarn) != 1 {
		t.Fatalf("es kamen %d", len(nachbarn.Nachbarn))
	}

	nachbar := nachbarn.Nachbarn[0]

	if nachbar.Bekannt || nachbar.Stand != "" {
		t.Fatalf("der Nachbar ist %+v", nachbar)
	}

	if nachbar.Name != "Annas Mac" || nachbar.Adresse != "192.168.1.7:41234" {
		t.Fatalf("der Nachbar ist %+v", nachbar)
	}
}

// **Ein bekannter Fund bringt seinen Stand mit.**
//
// **Das ist der Satz, für den es dieses Werk gibt.** Eine Liste, die jeden
// Fund gleich aussehen ließe, ließe jemanden einen Kontakt zum zweiten Mal
// eintragen — und er stünde dann doppelt da, einmal bestätigt und einmal
// nicht.
func TestEinBekannterFundBringtSeinenStandMit(t *testing.T) {
	t.Parallel()

	kontakte := einKontaktdoppel("bbekannt")
	kontakte.kontakte["bbekannt"].Stand = domain.KontaktBestaetigt

	werk := service.NeuesNachbarwerk(
		nachbardoppel{funde: []service.Nachbarfund{
			{Kennung: "bbekannt", Name: "Kennen wir", Adresse: "a:1"},
			{Kennung: "bfremd", Name: "Kennen wir nicht", Adresse: "b:2"},
		}},
		service.NeuesKontaktwerk(kontakte, einePeeruhr))

	nachbarn, fehler := werk.Liste(t.Context())
	if fehler != nil {
		t.Fatalf("Liste: %v", fehler)
	}

	nach := map[peer.Kennung]service.Nachbar{}
	for _, n := range nachbarn.Nachbarn {
		nach[n.Kennung] = n
	}

	if !nach["bbekannt"].Bekannt ||
		nach["bbekannt"].Stand != domain.KontaktBestaetigt {
		t.Errorf("der bekannte ist %+v", nach["bbekannt"])
	}

	if nach["bfremd"].Bekannt {
		t.Errorf("der fremde ist %+v", nach["bfremd"])
	}
}

// **Der eigene Widerhall steht nicht in der Liste.**
//
// # Warum das kein Schönheitsfehler ist
//
// **Weil mDNS die eigene Ankündigung zurückgibt.** Wer sich meldet und dann
// sucht, findet sich selbst — und stünde dann in seiner eigenen Nachbarschaft,
// mit einem Knopf „als Kontakt aufnehmen", der einen Kontakt zu sich selbst
// erzeugte.
func TestDerEigeneWiderhallStehtNichtInDerListe(t *testing.T) {
	t.Parallel()

	werk := service.NeuesNachbarwerk(
		nachbardoppel{funde: []service.Nachbarfund{
			{Kennung: "bselbst", Name: "Ich", Adresse: "a:1"},
			{Kennung: "bfremd", Name: "Jemand", Adresse: "b:2"},
		}},
		service.NeuesKontaktwerk(einKontaktdoppel(), einePeeruhr))

	nachbarn, fehler := werk.Liste(
		service.MitEigenerKennung(t.Context(), "bselbst"))
	if fehler != nil {
		t.Fatalf("Liste: %v", fehler)
	}

	if len(nachbarn.Nachbarn) != 1 ||
		nachbarn.Nachbarn[0].Kennung != "bfremd" {
		t.Fatalf("es kamen %+v", nachbarn.Nachbarn)
	}
}

// **Die verworfenen Antworten reisen mit.**
func TestDieVerworfenenAntwortenReisenMit(t *testing.T) {
	t.Parallel()

	werk := service.NeuesNachbarwerk(
		nachbardoppel{verworfen: 3},
		service.NeuesKontaktwerk(einKontaktdoppel(), einePeeruhr))

	nachbarn, fehler := werk.Liste(t.Context())
	if fehler != nil {
		t.Fatalf("Liste: %v", fehler)
	}

	if nachbarn.Verworfen != 3 {
		t.Fatalf("verworfen wurden %d", nachbarn.Verworfen)
	}

	// **Und die Liste ist leer und nicht `nil`** — sonst verschwände sie
	// aus der Antwort, und wer sie liest, hielte „keine Nachbarn" für
	// „keine Angabe".
	if nachbarn.Nachbarn == nil {
		t.Fatal("die Liste ist nil")
	}
}

// **Ein Fehler der Suche ist kein leeres Ergebnis.**
func TestEinFehlerDerSucheIstKeinLeeresErgebnis(t *testing.T) {
	t.Parallel()

	werk := service.NeuesNachbarwerk(
		nachbardoppel{fehler: domain.MitGrund(
			domain.ErrValidation, nil, "das Netz nimmt niemanden auf")},
		service.NeuesKontaktwerk(einKontaktdoppel(), einePeeruhr))

	if _, fehler := werk.Liste(t.Context()); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Und ein Fehler der Kontaktliste ebenso.**
//
// **Der Unterschied zählt** (§8): „ich kenne ihn nicht" und „ich weiß nicht,
// ob ich ihn kenne" sind zwei Dinge, und ein Nachbar, der als unbekannt
// dastünde, weil der Bestand schwieg, lüde zum doppelten Eintragen ein.
func TestEinFehlerDerKontaktlisteIstKeinUnbekannt(t *testing.T) {
	t.Parallel()

	kontakte := einKontaktdoppel()
	kontakte.scheitert = errors.New("kein Bestand")

	werk := service.NeuesNachbarwerk(
		nachbardoppel{funde: []service.Nachbarfund{{Kennung: "bfremd"}}},
		service.NeuesKontaktwerk(kontakte, einePeeruhr))

	if _, fehler := werk.Liste(t.Context()); fehler == nil {
		t.Fatal("die Liste kam ohne Bestand durch")
	}
}
