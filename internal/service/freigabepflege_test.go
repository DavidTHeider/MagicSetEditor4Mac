package service_test

// **Was ein Gegenüber zu sehen bekommt** — Werkstück `T3` von Phase 5.6
// (#1468).
//
// # Warum die Prüfung hier steht und nicht an der Fläche
//
// **Weil das Issue sie vor **jeder** Projektion verlangt** — vor der Antwort,
// vor dem Zwischenspeicher, vor der Wiederaufnahme, vor Ausfuhr und Beleg.
// Eine Fläche, die sie selbst führte, wäre die erste von vieren, und die
// zweite vergäße sie.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// freigabeprobezeit ist der feste Zeitpunkt dieser Läufe.
var freigabeprobezeit = time.Date(2026, time.September, 18, 12, 0, 0, 0, time.UTC)

// kontaktnachschlagprobe gibt einen Kontakt zurück — oder klagt.
type kontaktnachschlagprobe struct {
	kontakt domain.Peerkontakt
	klagt   error
}

func (k kontaktnachschlagprobe) Lesen(
	context.Context, string,
) (domain.Peerkontakt, error) {
	return k.kontakt, k.klagt
}

// freigabeprobe hält Freigaben im Arbeitsspeicher.
type freigabeprobe struct {
	freigaben []domain.Sammlungsfreigabe
	klagt     error
}

func (f *freigabeprobe) Erteilen(
	_ context.Context, freigabe domain.Sammlungsfreigabe,
) error {
	if f.klagt != nil {
		return f.klagt
	}

	f.freigaben = append(f.freigaben, freigabe)

	return nil
}

func (f *freigabeprobe) Widerrufen(
	_ context.Context, _ string, _ domain.Zweck, umfang domain.Freigabeumfang,
) error {
	if f.klagt != nil {
		return f.klagt
	}

	uebrig := []domain.Sammlungsfreigabe{}

	for _, freigabe := range f.freigaben {
		if freigabe.Umfang != umfang {
			uebrig = append(uebrig, freigabe)
		}
	}

	f.freigaben = uebrig

	return nil
}

func (f *freigabeprobe) Fuer(
	context.Context, string,
) ([]domain.Sammlungsfreigabe, error) {
	return f.freigaben, f.klagt
}

// eineFreigabepflege baut die Pflege über drei Doppeln.
func eineFreigabepflege(
	kontakt domain.Peerkontakt, freigaben []domain.Sammlungsfreigabe,
	exemplare []*domain.Exemplar, festgehalten map[int64]int,
) *service.Freigabepflege {
	return service.NeueFreigabepflege(
		&freigabeprobe{freigaben: freigaben},
		kontaktnachschlagprobe{kontakt: kontakt},
		&sammlungsdoppel{zeilen: exemplare},
		func() time.Time { return freigabeprobezeit },
	).MitReservierungen(zusagenprobe{fest: festgehalten})
}

// handelnderKontakt ist ein bestätigter Kontakt mit dem Recht zu handeln.
func handelnderKontakt() domain.Peerkontakt {
	return domain.Peerkontakt{
		Kennung: "peer-1", Stand: domain.KontaktBestaetigt,
		Rechte: []domain.Zweck{domain.ZweckHandel},
	}
}

// dreiFreigabezeilen sind ein Besitz, ein Wunsch und ein zweiter Besitz.
func dreiFreigabezeilen() []*domain.Exemplar {
	return []*domain.Exemplar{
		{
			ID: 1, Printing: "abc", Menge: 4, Zustand: domain.ZustandNearMint,
			Sprache: "en", Erwerbspreis: 417, Notiz: "geheim",
			Lagerort: "Kiste 4",
		},
		{
			ID: 2, Printing: "def", Menge: 1, Zustand: domain.ZustandNearMint,
			Sprache: "en", Wunsch: true,
		},
		{
			ID: 3, Printing: "ghi", Menge: 2, Zustand: domain.ZustandPlayed,
			Sprache: "de",
		},
	}
}

// eineSammlungsfreigabe ist die Freigabe der ganzen Sammlung zum Handel.
func eineSammlungsfreigabe() domain.Sammlungsfreigabe {
	return domain.Sammlungsfreigabe{
		Empfaenger: "peer-1", Umfang: domain.UmfangSammlung,
		Zweck: domain.ZweckHandel, Erteilt: freigabeprobezeit,
	}
}

// **Mit Freigabe kommt eine Sicht — und sie zeigt nur den Besitz.**
func TestDieSichtZeigtDenFreigegebenenBesitz(t *testing.T) {
	t.Parallel()

	pflege := eineFreigabepflege(handelnderKontakt(),
		[]domain.Sammlungsfreigabe{eineSammlungsfreigabe()}, dreiFreigabezeilen(), nil)

	sicht, fehler := pflege.Sicht(
		context.Background(), "peer-1", "s-1", domain.ZweckHandel)
	if fehler != nil {
		t.Fatalf("Sicht: %v", fehler)
	}

	// **Zwei Besitzzeilen, kein Wunsch.**
	if len(sicht) != 2 {
		t.Fatalf("die Sicht hat %d Zeilen: %+v", len(sicht), sicht)
	}

	if sicht[0].Printing != "abc" || sicht[0].Menge != 4 {
		t.Errorf("die erste Zeile ist %+v", sicht[0])
	}
}

// **Was zugesagt ist, wird nicht ein zweites Mal gezeigt.**
//
// # Warum das hier steht und nicht im Issue
//
// **Weil es sonst zweimal angeboten wäre.** `T2` hält fest, was einem Handel
// zugesagt ist; eine Sicht, die es trotzdem zeigte, lüde einen zweiten
// Partner ein, danach zu fragen — und die Absage käme erst beim Zusagen. Das
// Issue nennt den Fall nicht; er folgt aus `T2` (0120), und die Entscheidung
// steht hier statt als Annahme im Kopf.
func TestDieSichtZeigtNichtWasZugesagtIst(t *testing.T) {
	t.Parallel()

	pflege := eineFreigabepflege(handelnderKontakt(),
		[]domain.Sammlungsfreigabe{eineSammlungsfreigabe()}, dreiFreigabezeilen(),
		map[int64]int{1: 3, 3: 2})

	sicht, fehler := pflege.Sicht(
		context.Background(), "peer-1", "s-1", domain.ZweckHandel)
	if fehler != nil {
		t.Fatalf("Sicht: %v", fehler)
	}

	// **Vier minus drei ist eines; zwei minus zwei ist keines** — und eine
	// Zeile ohne Stück steht nicht in der Sicht.
	if len(sicht) != 1 || sicht[0].Menge != 1 {
		t.Errorf("die Sicht ist %+v", sicht)
	}
}

// **Ohne Freigabe kommt keine Sicht, sondern ein Nein.**
func TestOhneFreigabeKommtKeineSicht(t *testing.T) {
	t.Parallel()

	pflege := eineFreigabepflege(handelnderKontakt(), nil, dreiFreigabezeilen(), nil)

	_, fehler := pflege.Sicht(
		context.Background(), "peer-1", "s-1", domain.ZweckHandel)
	if !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Und ein Kontakt, den es nicht gibt, bekommt dasselbe Nein.**
//
// **Dasselbe, und das ist Absicht** (§8): „gibt es nicht" und „darfst du
// nicht" sind für ein fremdes Gegenüber dieselbe Antwort. Wer sie
// unterschiede, ließe die Kontaktliste abfragen.
func TestEinUnbekannterKontaktBekommtDasselbeNein(t *testing.T) {
	t.Parallel()

	pflege := service.NeueFreigabepflege(
		&freigabeprobe{freigaben: []domain.Sammlungsfreigabe{eineSammlungsfreigabe()}},
		kontaktnachschlagprobe{klagt: domain.ErrNotFound},
		&sammlungsdoppel{zeilen: dreiFreigabezeilen()},
		func() time.Time { return freigabeprobezeit },
	)

	_, fehler := pflege.Sicht(
		context.Background(), "peer-1", "s-1", domain.ZweckHandel)
	if !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("es kam %v", fehler)
	}

	// **Und der wirkliche Grund geht ins Protokoll** (§8, Vermerk 0075).
	//
	// **Dieser Satz ist der Unterschied, den der Lauf sonst nicht sähe**: ein
	// Kontakt, der gar nicht geladen wird, ist der Nullwert — und der darf
	// ohnehin nichts. Ohne die Frage nach dem Grund bliebe der Lauf grün,
	// auch wenn niemand den Fehler des Nachschlags ansähe (§10).
	if grund := domain.GrundVon(fehler); !errors.Is(grund, domain.ErrNotFound) {
		t.Errorf("der Grund ist %v", grund)
	}
}

// **Was der Bestand klagt, kommt zurück — aus jedem der vier Wege.**
//
// **Jeder einzeln**, weil ein Lauf über alle zusammen grün bliebe, sobald
// **einer** von ihnen klagt (§10).
func TestWasDerBestandBeiDenSichtenKlagtKommtZurueck(t *testing.T) {
	t.Parallel()

	klage := errors.New("die Platte ist voll")

	for name, bauen := range map[string]func() *service.Freigabepflege{
		"beim Lesen der Freigaben": func() *service.Freigabepflege {
			return service.NeueFreigabepflege(
				&freigabeprobe{klagt: klage},
				kontaktnachschlagprobe{kontakt: handelnderKontakt()},
				&sammlungsdoppel{zeilen: dreiFreigabezeilen()},
				func() time.Time { return freigabeprobezeit },
			)
		},
		"beim Lesen der Sammlung": func() *service.Freigabepflege {
			return service.NeueFreigabepflege(
				&freigabeprobe{
					freigaben: []domain.Sammlungsfreigabe{eineSammlungsfreigabe()},
				},
				kontaktnachschlagprobe{kontakt: handelnderKontakt()},
				&sammlungsdoppel{fehler: klage},
				func() time.Time { return freigabeprobezeit },
			)
		},
		"beim Zählen der Zusagen": func() *service.Freigabepflege {
			return service.NeueFreigabepflege(
				&freigabeprobe{
					freigaben: []domain.Sammlungsfreigabe{eineSammlungsfreigabe()},
				},
				kontaktnachschlagprobe{kontakt: handelnderKontakt()},
				&sammlungsdoppel{zeilen: dreiFreigabezeilen()},
				func() time.Time { return freigabeprobezeit },
			).MitReservierungen(zusagenprobe{klagt: klage})
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := bauen().Sicht(
				context.Background(), "peer-1", "s-1", domain.ZweckHandel)
			if !errors.Is(fehler, klage) {
				t.Errorf("es kam %v", fehler)
			}
		})
	}
}

// **Und die drei Wege, die nur weiterreichen, reichen wirklich weiter.**
//
// **Sie stehen hier, weil sie nichts entscheiden** — und genau deshalb wäre
// eine Naht, die sie vergäße, still: „nichts freigegeben" sähe aus wie eine
// Antwort.
func TestDieDreiSchreibwegeDerPflegeReichenDurch(t *testing.T) {
	t.Parallel()

	ablage := &freigabeprobe{}
	pflege := service.NeueFreigabepflege(ablage,
		kontaktnachschlagprobe{kontakt: handelnderKontakt()},
		&sammlungsdoppel{zeilen: dreiFreigabezeilen()},
		func() time.Time { return freigabeprobezeit })

	ctx := context.Background()

	// **Ohne Zeitpunkt trägt die Pflege ihn ein** — eine Freigabe ohne
	// Erteilungszeitpunkt hätte keine Spur.
	if fehler := pflege.Erteilen(ctx, domain.Sammlungsfreigabe{
		Empfaenger: "peer-1", Umfang: domain.UmfangSammlung,
		Zweck: domain.ZweckHandel,
	}); fehler != nil {
		t.Fatalf("erteilen: %v", fehler)
	}

	if len(ablage.freigaben) != 1 ||
		!ablage.freigaben[0].Erteilt.Equal(freigabeprobezeit) {
		t.Errorf("erteilt wurde %+v", ablage.freigaben)
	}

	gelesen, fehler := pflege.Freigaben(ctx, "peer-1")
	if fehler != nil || len(gelesen) != 1 {
		t.Errorf("gelesen wurde %+v (%v)", gelesen, fehler)
	}

	if fehler := pflege.Widerrufen(ctx, "peer-1", domain.ZweckHandel,
		domain.UmfangSammlung); fehler != nil {
		t.Errorf("widerrufen: %v", fehler)
	}

	if len(ablage.freigaben) != 0 {
		t.Errorf("es stehen noch %+v da", ablage.freigaben)
	}
}

// **Ohne Naht zum Handeln ist festgehalten nichts.**
//
// **Und das ist keine Annahme, sondern der Stand**: wer das Handeln nicht
// verdrahtet hat, hat keine Zusagen. Ohne diesen Lauf bliebe der Zweig
// unbetreten — und eine Pflege ohne Naht fiele im Betrieb auf eine `nil`-Map.
func TestOhneNahtZumHandelnIstNichtsFestgehalten(t *testing.T) {
	t.Parallel()

	pflege := service.NeueFreigabepflege(
		&freigabeprobe{
			freigaben: []domain.Sammlungsfreigabe{eineSammlungsfreigabe()},
		},
		kontaktnachschlagprobe{kontakt: handelnderKontakt()},
		&sammlungsdoppel{zeilen: dreiFreigabezeilen()},
		func() time.Time { return freigabeprobezeit },
	)

	sicht, fehler := pflege.Sicht(
		context.Background(), "peer-1", "s-1", domain.ZweckHandel)
	if fehler != nil {
		t.Fatalf("Sicht: %v", fehler)
	}

	// **Zwei Besitzzeilen, und keine davon ist beschnitten.**
	if len(sicht) != 2 || sicht[0].Menge != 4 {
		t.Errorf("die Sicht ist %+v", sicht)
	}
}

// **Nichts Privates kommt in der Sicht an.**
//
// **Der Lauf über die Domäne prüft die Gestalt** (`sichtkarte_test.go`);
// dieser hier prüft, dass der **Weg** sie benutzt. Zwischen beidem liegt jede
// Abkürzung, die jemand später nimmt.
func TestDerWegLaesstNichtsPrivatesDurch(t *testing.T) {
	t.Parallel()

	pflege := eineFreigabepflege(handelnderKontakt(),
		[]domain.Sammlungsfreigabe{eineSammlungsfreigabe()}, dreiFreigabezeilen(), nil)

	sicht, fehler := pflege.Sicht(
		context.Background(), "peer-1", "s-1", domain.ZweckHandel)
	if fehler != nil {
		t.Fatalf("Sicht: %v", fehler)
	}

	for _, zeile := range sicht {
		if zeile.Referenz == "" || zeile.Referenz == "1" {
			t.Errorf("die Referenz ist %q", zeile.Referenz)
		}
	}
}

// **Die Pflege setzt die Vorgabe der Behaltemenge, bevor sie ablegt** — UXR-F18
// (#2033).
// Ohne sie gäbe „Dubletten" ohne Angabe die ganze Sammlung frei.
func TestDiePflegeSetztDieVorgabeDerBehaltemenge(t *testing.T) {
	t.Parallel()

	ablage := &freigabeprobe{}
	pflege := service.NeueFreigabepflege(ablage,
		kontaktnachschlagprobe{kontakt: handelnderKontakt()},
		&sammlungsdoppel{zeilen: dreiFreigabezeilen()},
		func() time.Time { return freigabeprobezeit })

	if fehler := pflege.Erteilen(context.Background(), domain.Sammlungsfreigabe{
		Empfaenger: "peer-1", Umfang: domain.UmfangDubletten,
		Zweck: domain.ZweckHandel,
	}); fehler != nil {
		t.Fatalf("erteilen: %v", fehler)
	}

	if len(ablage.freigaben) != 1 || ablage.freigaben[0].Behalten != 1 {
		t.Errorf("abgelegt wurde %+v — erwartet Behalten 1", ablage.freigaben)
	}
}
