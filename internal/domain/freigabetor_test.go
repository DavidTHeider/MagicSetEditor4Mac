package domain_test

// **Wer nichts sieht, und warum** — Werkstück `T3` von Phase 5.6
// (#1468).
//
// # Warum das Nein fünf Gründe hat und einen Weg
//
// **Weil die Abnahme fünf davon nennt**: Paarung ohne Freigabe, falsches
// Profil, anderer Empfänger, abgelaufener Umfang, gesperrter Kontakt. Vier
// davon stehen hier; das falsche Profil steht im Bestand, weil eine Freigabe
// einem Profil gehört und eine Abfrage darauf klemmt (0066).
//
// **Und einen Weg**, weil ein zweiter irgendwann anders entscheidet. Wer die
// Sicht baut, fragt **hier** — und bekommt entweder eine Freigabe oder einen
// Grund.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// jetzt ist der feste Zeitpunkt dieser Läufe.
func jetztFuerFreigaben() time.Time {
	return time.Date(2026, time.September, 18, 12, 0, 0, 0, time.UTC)
}

// bestaetigt ist ein Kontakt, der handeln darf.
func bestaetigt() domain.Peerkontakt {
	return domain.Peerkontakt{
		Kennung: "peer-1", Stand: domain.KontaktBestaetigt,
		Rechte: []domain.Zweck{domain.ZweckHandel},
	}
}

// fuerHandel ist eine unbefristete Freigabe der ganzen Sammlung.
func fuerHandel() domain.Sammlungsfreigabe {
	return domain.Sammlungsfreigabe{
		Empfaenger: "peer-1", Umfang: domain.UmfangSammlung,
		Zweck: domain.ZweckHandel, Erteilt: jetztFuerFreigaben(),
	}
}

// **Mit Freigabe und bestätigtem Kontakt gibt es eine Sicht.**
func TestMitFreigabeGibtEsEineSicht(t *testing.T) {
	t.Parallel()

	gefunden, fehler := domain.FreigabenFuer(bestaetigt(),
		[]domain.Sammlungsfreigabe{fuerHandel()}, domain.ZweckHandel,
		jetztFuerFreigaben())
	if fehler != nil {
		t.Fatalf("es kam %v", fehler)
	}

	if len(gefunden) != 1 || gefunden[0].Umfang != domain.UmfangSammlung {
		t.Errorf("gefunden wurde %+v", gefunden)
	}
}

// **Zwei Umfänge gelten nebeneinander, und der weitere gewinnt.**
//
// **Sonst wäre eine zweite Freigabe eine Einschränkung der ersten** — und
// niemand rechnete damit, dass das Freigeben etwas wegnimmt.
func TestZweiUmfaengeGeltenNebeneinander(t *testing.T) {
	t.Parallel()

	dubletten := fuerHandel()
	dubletten.Umfang, dubletten.Behalten = domain.UmfangDubletten, 3

	// **Die weitere steht vorn, die engere hinten.** Der erste Anlauf hatte
	// sie umgekehrt, und die Gegenprobe „nimm einfach die letzte" blieb grün —
	// sie maß die Reihenfolge und nicht die Regel (§10).
	gefunden, fehler := domain.FreigabenFuer(bestaetigt(),
		[]domain.Sammlungsfreigabe{fuerHandel(), dubletten},
		domain.ZweckHandel, jetztFuerFreigaben())
	if fehler != nil {
		t.Fatalf("es kam %v", fehler)
	}

	// **Vier Stück, drei behalten**: die Dublettenfreigabe zeigt eines, die
	// Sammlungsfreigabe vier — und vier ist, was gilt.
	menge, zeigt := domain.Sichtmenge(gefunden, besitz("abc", 4))
	if !zeigt || menge != 4 {
		t.Errorf("es kam (%d, %v)", menge, zeigt)
	}
}

// **Und was keine Freigabe zeigt, zeigt auch das Bündel nicht.**
func TestEinBuendelZeigtNichtMehrAlsSeineTeile(t *testing.T) {
	t.Parallel()

	nurWuensche := fuerHandel()
	nurWuensche.Umfang = domain.UmfangWuensche

	if menge, zeigt := domain.Sichtmenge(
		[]domain.Sammlungsfreigabe{nurWuensche}, besitz("abc", 4),
	); zeigt || menge != 0 {
		t.Errorf("es kam (%d, %v)", menge, zeigt)
	}
}

// **Und ohne sie nicht — aus jedem der vier Gründe.**
//
// **Jeder einzeln**, weil ein Lauf über alle vier zusammen grün bliebe,
// sobald **einer** von ihnen greift (§10).
func TestOhneFreigabeGibtEsKeineSicht(t *testing.T) {
	t.Parallel()

	gesperrt := bestaetigt()
	gesperrt.Stand = domain.KontaktGesperrt

	nurGepaart := bestaetigt()
	nurGepaart.Stand, nurGepaart.Rechte = domain.KontaktUnbestaetigt, nil

	abgelaufen := fuerHandel()
	abgelaufen.Ablauf = jetztFuerFreigaben().Add(-time.Minute)

	fuerAndere := fuerHandel()
	fuerAndere.Empfaenger = "peer-2"

	zumTeilen := fuerHandel()
	zumTeilen.Zweck = domain.ZweckTeilen

	for name, fall := range map[string]struct {
		kontakt   domain.Peerkontakt
		freigaben []domain.Sammlungsfreigabe
	}{
		// **Paarung allein gestattet nichts** (N1): sie ist das
		// Kennenlernen, keine Vollmacht.
		"nur gepaart":             {nurGepaart, []domain.Sammlungsfreigabe{fuerHandel()}},
		"gesperrt":                {gesperrt, []domain.Sammlungsfreigabe{fuerHandel()}},
		"gar keine Freigabe":      {bestaetigt(), nil},
		"abgelaufene":             {bestaetigt(), []domain.Sammlungsfreigabe{abgelaufen}},
		"für einen anderen":       {bestaetigt(), []domain.Sammlungsfreigabe{fuerAndere}},
		"für einen anderen Zweck": {bestaetigt(), []domain.Sammlungsfreigabe{zumTeilen}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := domain.FreigabenFuer(fall.kontakt, fall.freigaben,
				domain.ZweckHandel, jetztFuerFreigaben())
			if !errors.Is(fehler, domain.ErrErlaubnis) {
				t.Errorf("es kam %v", fehler)
			}
		})
	}
}

// **Was eine Freigabe zu einer macht.**
func TestWasEineFreigabeScheiternLaesst(t *testing.T) {
	t.Parallel()

	ohneEmpfaenger := fuerHandel()
	ohneEmpfaenger.Empfaenger = ""

	erfundenerUmfang := fuerHandel()
	erfundenerUmfang.Umfang = "alles"

	erfundenerZweck := fuerHandel()
	erfundenerZweck.Zweck = "spaß"

	// **Eine Auswahl ohne Auswahl ist keine.** Sie sähe aus wie eine enge
	// Freigabe und wäre eine, die nichts zeigt — und wer sie erteilt, glaubt,
	// er habe etwas freigegeben.
	leereAuswahl := fuerHandel()
	leereAuswahl.Umfang = domain.UmfangAuswahl

	// **Und eine Behaltemenge ergibt nur bei Dubletten einen Sinn.**
	fremdeBehaltemenge := fuerHandel()
	fremdeBehaltemenge.Behalten = 2

	// **Gewählte Drucke ebenso** — an einer Sammlungsfreigabe sagen sie
	// nichts, und wer sie dort hinschriebe, hätte etwas anderes gemeint.
	fremdeAuswahl := fuerHandel()
	fremdeAuswahl.Auswahl = []string{"abc"}

	// **Eine negative Behaltemenge ist keine.** Sie zeigte **mehr** als die
	// Zeile hat — `Menge - (-2)`.
	krummeBehaltemenge := fuerHandel()
	krummeBehaltemenge.Umfang, krummeBehaltemenge.Behalten =
		domain.UmfangDubletten, -2

	for name, freigabe := range map[string]domain.Sammlungsfreigabe{
		"ohne Empfänger":                    ohneEmpfaenger,
		"erfundener Umfang":                 erfundenerUmfang,
		"erfundener Zweck":                  erfundenerZweck,
		"Auswahl ohne Auswahl":              leereAuswahl,
		"Behaltemenge ohne Dublettenumfang": fremdeBehaltemenge,
		"gewählte Drucke ohne Auswahl":      fremdeAuswahl,
		"negative Behaltemenge":             krummeBehaltemenge,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := freigabe.Validate(); !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("es kam %v", fehler)
			}
		})
	}

	// **Und eine richtige kommt durch** — sonst prüfte der Lauf nur, dass
	// überhaupt etwas abgewiesen wird (§10).
	for name, freigabe := range map[string]domain.Sammlungsfreigabe{
		"die ganze Sammlung": fuerHandel(),
		"eine Auswahl": {
			Empfaenger: "peer-1", Umfang: domain.UmfangAuswahl,
			Zweck: domain.ZweckHandel, Auswahl: []string{"abc"},
		},
		"Dubletten mit Behaltemenge": {
			Empfaenger: "peer-1", Umfang: domain.UmfangDubletten,
			Zweck: domain.ZweckHandel, Behalten: 1,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := freigabe.Validate(); fehler != nil {
				t.Errorf("es kam %v", fehler)
			}
		})
	}
}
