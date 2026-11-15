package service_test

// **Was für einen Handel festgehalten wird** —
// #1643.
//
// # Wogegen hier geprüft wird
//
// **Dass die Revision aus dem Stand kommt und nicht aus dem Ruf.** Eine
// Zusage gilt einer Angebotsrevision; wer sie mitschickte, könnte eine
// nennen, die es nicht mehr gibt.
//
// **Und dass ein unbekannter Vorgang nichts festhält.** Ein Exemplar, das
// für einen Handel gebunden wäre, den es nicht gibt, bliebe für immer
// gebunden — niemand könnte ihn beenden.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// zusagenhaltprobe hält Zusagen im Arbeitsspeicher.
type zusagenhaltprobe struct {
	zusagen []domain.Reservierung
	klagt   error

	// festhaltefehler klagt **nur** beim Eintragen — damit ein Lauf die
	// Absage des Bestands prüfen kann, ohne dass schon das Nachschlagen
	// davor scheitert.
	festhaltefehler error

	// vertrag merkt sich, welchen Vertrag der Dienst durchgereicht hat
	// (Befund 1 des Nachlaufs zum Zweitblick über Phase 5.6; seit #2492 die
	// Zeilen und nicht mehr ihre Summe).
	vertrag domain.Zusagevertrag

	freigegeben int
	uebergeben  int
}

func (z *zusagenhaltprobe) Festhalten(
	_ context.Context, zusage domain.Reservierung, vertrag domain.Zusagevertrag,
) error {
	z.vertrag = vertrag

	if z.festhaltefehler != nil {
		return z.festhaltefehler
	}

	if z.klagt != nil {
		return z.klagt
	}

	z.zusagen = append(z.zusagen, zusage)

	return nil
}

func (z *zusagenhaltprobe) FuerVorgang(
	_ context.Context, vorgang string,
) ([]domain.Reservierung, error) {
	if z.klagt != nil {
		return nil, z.klagt
	}

	gefunden := []domain.Reservierung{}

	for _, zusage := range z.zusagen {
		if zusage.Vorgang == vorgang {
			gefunden = append(gefunden, zusage)
		}
	}

	return gefunden, nil
}

func (z *zusagenhaltprobe) Uebergeben(
	_ context.Context, vorgang string,
) error {
	if z.klagt != nil {
		return z.klagt
	}

	vermerkt := 0

	for stelle, zusage := range z.zusagen {
		if zusage.Vorgang == vorgang {
			z.zusagen[stelle].Uebergeben = true
			vermerkt++
		}
	}

	z.uebergeben += vermerkt

	return nil
}

func (z *zusagenhaltprobe) Freigeben(_ context.Context, vorgang string) error {
	if z.klagt != nil {
		return z.klagt
	}

	for _, zusage := range z.zusagen {
		if zusage.Vorgang == vorgang && zusage.Uebergeben {
			return domain.MitGrund(domain.ErrConflict, nil,
				"eine bestätigte Übergabe gibt nichts mehr frei")
		}
	}

	gehalten := z.zusagen[:0]

	for _, zusage := range z.zusagen {
		if zusage.Vorgang != vorgang {
			gehalten = append(gehalten, zusage)
		}
	}

	z.zusagen = gehalten
	z.freigegeben++

	return nil
}

// **Eine benannte Besitzzeile wird für den Vorgang festgehalten.**
func TestEineBenannteBesitzzeileWirdFestgehalten(t *testing.T) {
	t.Parallel()

	ablage := &zusagenhaltprobe{}

	stand := derHeutigeStand()
	stand.Revision = 3

	werk := service.NeuesZusagenwerk(ablage, standprobe{stand: stand})

	zusage, fehler := werk.Festhalten(context.Background(),
		service.Zusagenwunsch{
			Vorgang: derPostvorgang, Wer: "bich",
			ExemplarID: 7, Menge: 1, Fassung: 4,
		})
	if fehler != nil {
		t.Fatalf("Festhalten: %v", fehler)
	}

	switch {
	case zusage.Revision != 3:
		t.Errorf("die Zusage gilt der Revision %d", zusage.Revision)
	case zusage.ExemplarID != 7 || zusage.Menge != 1 || zusage.Fassung != 4:
		t.Errorf("die Zusage ist %+v", zusage)
	case len(ablage.zusagen) != 1:
		t.Errorf("eingetragen wurden %d", len(ablage.zusagen))
	}
}

// **Ein unbekannter Vorgang hält nichts fest.**
//
// **Sonst bliebe ein Exemplar für immer gebunden** — an einen Handel, den
// niemand beenden kann, weil es ihn nicht gibt.
func TestEinUnbekannterVorgangHaeltNichtsFest(t *testing.T) {
	t.Parallel()

	ablage := &zusagenhaltprobe{}

	werk := service.NeuesZusagenwerk(ablage,
		standprobe{klagt: domain.MitGrund(domain.ErrNotFound, nil,
			"den Handelsvorgang gibt es nicht")})

	_, fehler := werk.Festhalten(context.Background(),
		service.Zusagenwunsch{
			Vorgang: "v-gibtesnicht", Wer: "bich", ExemplarID: 7, Menge: 1,
		})
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("es kam %v", fehler)
	}

	if len(ablage.zusagen) != 0 {
		t.Errorf("eingetragen wurden %d", len(ablage.zusagen))
	}
}

// **Was der Bestand ablehnt, kommt als Absage zurück.**
func TestEineAbgelehnteZusageKommtZurueck(t *testing.T) {
	t.Parallel()

	werk := service.NeuesZusagenwerk(
		&zusagenhaltprobe{festhaltefehler: domain.MitGrund(
			domain.ErrConflict, nil, "frei ist weniger")},
		standprobe{stand: derHeutigeStand()})

	if _, fehler := werk.Festhalten(context.Background(),
		service.Zusagenwunsch{
			Vorgang: derPostvorgang, Wer: "bich", ExemplarID: 7, Menge: 1,
		}); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Und die Zusagen eines Vorgangs lassen sich lesen.**
func TestDieZusagenEinesVorgangsLassenSichLesen(t *testing.T) {
	t.Parallel()

	ablage := &zusagenhaltprobe{zusagen: []domain.Reservierung{
		{Vorgang: derPostvorgang, ExemplarID: 7, Menge: 1},
		{Vorgang: "v-anderer", ExemplarID: 8, Menge: 1},
	}}

	werk := service.NeuesZusagenwerk(ablage,
		standprobe{stand: derHeutigeStand()})

	gefunden, fehler := werk.Zusagen(context.Background(), derPostvorgang)
	if fehler != nil {
		t.Fatalf("Zusagen: %v", fehler)
	}

	if len(gefunden) != 1 || gefunden[0].ExemplarID != 7 {
		t.Errorf("gefunden wurde %+v", gefunden)
	}
}

// **Ein Lesefehler wird weitergereicht und nicht zu einer leeren Liste.**
func TestEinLesefehlerDerZusagenKommtDurch(t *testing.T) {
	t.Parallel()

	werk := service.NeuesZusagenwerk(
		&zusagenhaltprobe{klagt: errors.New("der Bestand klemmt")},
		standprobe{stand: derHeutigeStand()})

	if _, fehler := werk.Zusagen(
		context.Background(), derPostvorgang); fehler == nil {
		t.Error("der Lesefehler kam als leere Liste")
	}
}

// **Die Grenze des Vertrags wird durchgereicht** — Befund 1 des Zweitblicks
// über Phase 5.6, und Befund 1 seines Nachlaufs.
//
// # Warum sie hier gerechnet und dort geprüft wird
//
// **Weil eine Prüfung vor der Transaktion eine Vorschau ist.** Der Dienst
// weiß, wie viel dieser Vertrag von dieser Partei nimmt; ob es hineinpasst,
// entscheidet der Bestand — in **derselben** Transaktion wie das Eintragen.
// Zwei Rufe für zwei verschiedene Besitzzeilen lasen sonst beide „null von
// eins" und schrieben beide.
func TestDieGrenzeDesVertragsWirdDurchgereicht(t *testing.T) {
	t.Parallel()

	ablage := &zusagenhaltprobe{}

	werk := service.NeuesZusagenwerk(ablage,
		standprobe{stand: derHeutigeStand()})

	if _, fehler := werk.Festhalten(context.Background(),
		service.Zusagenwunsch{
			Vorgang: derPostvorgang, Wer: "bich", ExemplarID: 7, Menge: 1,
		}); fehler != nil {
		t.Fatalf("Festhalten: %v", fehler)
	}

	// **Der Stand nimmt ein Stück von `bich`** — mehr bindet er nicht, und
	// welches, sagt die Zeile (#2492).
	if len(ablage.vertrag) != 1 || ablage.vertrag[0].Von != "bich" || ablage.vertrag[0].Menge != 1 {
		t.Errorf("durchgereicht wurde %+v", ablage.vertrag)
	}
}

// **Wer nichts weggibt, reicht einen leeren Vertrag durch.**
//
// **Und der Bestand weist sie ab**: ohne diese Zahl hielte jemand Exemplare
// für einen Handel fest, bei dem alles von drüben kommt — gebunden, ohne je
// wegzugehen.
func TestWerNichtsWeggibtReichtDieGrenzeNullDurch(t *testing.T) {
	t.Parallel()

	ablage := &zusagenhaltprobe{}

	werk := service.NeuesZusagenwerk(ablage,
		standprobe{stand: derHeutigeStand()})

	if _, fehler := werk.Festhalten(context.Background(),
		service.Zusagenwunsch{
			Vorgang: derPostvorgang, Wer: dasGegenueber, ExemplarID: 8,
			Menge: 1,
		}); fehler != nil {
		t.Fatalf("Festhalten: %v", fehler)
	}

	if len(ablage.vertrag) != 0 {
		t.Errorf("durchgereicht wurde %+v", ablage.vertrag)
	}
}
