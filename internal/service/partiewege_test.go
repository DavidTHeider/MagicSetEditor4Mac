package service_test

// **Die übrigen Wege der Partieleitung** — Werkstück G1 von Phase 5.5
// (#1458).
//
// **Herausgelöst aus `partieleitung_test.go`**, damit der Schnitt der Frage
// folgt: dort die Annahme eines Befehls, hier der Lebenszyklus, die
// Fehlerwege des Speichers und die Befehle, die der Sandkasten kennt.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Pausieren, fortsetzen, beenden — und was danach nicht mehr geht.**
func TestDerLebenszyklusEinerPartie(t *testing.T) {
	t.Parallel()

	leitung, speicher := eineLeitung(t)
	ctx := context.Background()

	if _, fehler := leitung.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if fehler := leitung.Pausieren(ctx, "p1"); fehler != nil {
		t.Fatalf("Pausieren: %v", fehler)
	}

	if speicher.partien["p1"].Stand != domain.StandPausiert {
		t.Errorf("der Stand ist %q", speicher.partien["p1"].Stand)
	}

	// **Eine angehaltene Partie bietet nichts an** (P3, #2371) — eine leere
	// Liste und kein `nil`: sie ist eine Antwort.
	if handlungen, fehler := handlungenVon(leitung,
		ctx, "p1", "sitz-1"); fehler != nil || handlungen == nil || len(handlungen) != 0 {
		t.Errorf("angehalten: %v, %v", handlungen, fehler)
	}

	if fehler := leitung.Fortsetzen(ctx, "p1"); fehler != nil {
		t.Fatalf("Fortsetzen: %v", fehler)
	}

	if speicher.partien["p1"].Stand != domain.StandLaeuft {
		t.Errorf("der Stand ist %q", speicher.partien["p1"].Stand)
	}

	if fehler := leitung.Beenden(ctx, "p1"); fehler != nil {
		t.Fatalf("Beenden: %v", fehler)
	}

	// **Eine beendete Partie kommt nicht zurück.** „Ein Abbruch, den man
	// zurücknehmen kann, ist keiner" — und ohne diese Zeile stünde am Ende
	// ein erfundenes Ergebnis (#1458).
	if fehler := leitung.Fortsetzen(
		ctx, "p1"); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("das Fortsetzen gab %v", fehler)
	}
}

// **Was der Speicher klagt, klagt die Leitung weiter.**
//
// **Jeder Weg einzeln**, denn ein Fehler auf dem einen sagt nichts über den
// anderen — und ein Weg, der ihn verschluckte, gäbe eine leere Antwort
// zurück, die wie eine gültige aussieht.
func TestWasDerSpeicherKlagtKlagtDieLeitungWeiter(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("der Bestand ist zu")

	faelle := map[string]func(*service.Partieleitung, context.Context) error{
		"Anlegen": func(l *service.Partieleitung, ctx context.Context) error {
			_, fehler := l.Anlegen(ctx, einePartie())

			return fehler
		},
		"Partie": func(l *service.Partieleitung, ctx context.Context) error {
			_, fehler := l.Partie(ctx, "p1", "sitz-1")

			return fehler
		},
		"Tischbild": func(l *service.Partieleitung, ctx context.Context) error {
			_, fehler := l.Tischbild(ctx, "p1", "sitz-1")

			return fehler
		},
		"Befehl": func(l *service.Partieleitung, ctx context.Context) error {
			_, fehler := l.Befehl(ctx, domain.Spielbefehl{
				Kennung: "b1", PartieID: "p1", Absender: "sitz-1",
				Art:      service.BefehlTappen,
				Nutzlast: map[string]any{"objekt": "o1"},
			})

			return fehler
		},
		"Pausieren": func(l *service.Partieleitung, ctx context.Context) error {
			return l.Pausieren(ctx, "p1")
		},
		"Fortsetzen": func(l *service.Partieleitung, ctx context.Context) error {
			return l.Fortsetzen(ctx, "p1")
		},
		"Beenden": func(l *service.Partieleitung, ctx context.Context) error {
			return l.Beenden(ctx, "p1")
		},
	}

	for name, weg := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			leitung, speicher := eineLeitung(t)
			speicher.klagt = kaputt

			if fehler := weg(leitung, context.Background()); !errors.Is(
				fehler, kaputt) {
				t.Fatalf("es kam %v", fehler)
			}
		})
	}
}

// **Ein Befehl, der nicht einmal zugestellt werden kann.**
func TestEinUnzustellbarerBefehl(t *testing.T) {
	t.Parallel()

	leitung, _ := eineLeitung(t)

	if _, fehler := leitung.Befehl(
		context.Background(), domain.Spielbefehl{PartieID: "p1"},
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Und einer, dessen Ereignis sich nicht ketten lässt.**
//
// **Ein echter Fall**: der Ausleger legt die Nutzlast in das Ereignis, und
// die kommt von außen. Wer einen Kanal hineinlegt, hat etwas hineingelegt,
// das kein JSON wird — und dann gibt es keinen Abdruck.
func TestEinEreignisOhneAbdruckWirdNichtGeschrieben(t *testing.T) {
	t.Parallel()

	speicher := &partieprobe{partien: map[string]domain.Partie{}}
	leitung := service.NeuePartieleitung(speicher, &krummerausleger{})
	ctx := context.Background()

	if _, fehler := leitung.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	_, fehler := leitung.Befehl(ctx, domain.Spielbefehl{
		Kennung: "b1", PartieID: "p1", Absender: "sitz-1",
		Art: service.BefehlTappen,
	})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}

	if speicher.angewendet != 0 {
		t.Error("das kettenlose Ereignis wurde geschrieben")
	}
}

// krummerausleger gibt ein Ereignis, das sich nicht kodieren lässt.
type krummerausleger struct{}

func (k *krummerausleger) Auslegen(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	return []domain.Ereignis{{
		Art:         befehl.Art,
		Oeffentlich: map[string]any{"kanal": make(chan int)},
	}}, partie, nil
}

func (k *krummerausleger) Erlaubt(domain.Partie, string) []string { return nil }

// **Und ein Ausleger, der gar nichts erlaubt, gibt eine leere Liste.**
func TestEinAuslegerOhneErlaubnisGibtEineLeereListe(t *testing.T) {
	t.Parallel()

	speicher := &partieprobe{partien: map[string]domain.Partie{}}
	leitung := service.NeuePartieleitung(speicher, &krummerausleger{})
	ctx := context.Background()

	if _, fehler := leitung.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	erlaubt, fehler := erlaubtVon(leitung, ctx, "p1", "sitz-1")
	if fehler != nil {
		t.Fatalf("Erlaubnis: %v", fehler)
	}

	if erlaubt == nil || len(erlaubt) != 0 {
		t.Errorf("es kam %v statt einer leeren Liste", erlaubt)
	}
}

// **Die sieben Befehle des Sandkastens, jeder mit seiner Wirkung.**
func TestDieSiebenBefehleDesSandkastens(t *testing.T) {
	t.Parallel()

	leitung, speicher := eineLeitung(t)
	ctx := context.Background()

	if _, fehler := leitung.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	folge := []domain.Spielbefehl{
		{
			Kennung: "b1", Art: service.BefehlTappen,
			Nutzlast: map[string]any{"objekt": "o1"},
		},
		{
			Kennung: "b2", Art: service.BefehlEnttappen,
			Nutzlast: map[string]any{"objekt": "o1"},
		},
		{
			Kennung: "b3", Art: service.BefehlUmdrehen,
			Nutzlast: map[string]any{"objekt": "o1"},
		},
		{
			Kennung: "b4", Art: service.BefehlAnhaengen,
			Nutzlast: map[string]any{"objekt": "o1", "an": "o2"},
		},
		{
			Kennung: "b5", Art: service.BefehlLoesen,
			Nutzlast: map[string]any{"objekt": "o1"},
		},
		{
			Kennung: "b6", Art: service.BefehlZaehler,
			Nutzlast: map[string]any{
				"objekt": "o1", "art": "gift", "menge": float64(3),
			},
		},
		// **Mehr abnehmen, als daliegt, nimmt alles** — und der Zähler
		// verschwindet, statt auf null stehen zu bleiben.
		{
			Kennung: "b7", Art: service.BefehlZaehler,
			Nutzlast: map[string]any{
				"objekt": "o1", "art": "gift", "menge": float64(-9),
			},
		},
	}

	for _, befehl := range folge {
		befehl.PartieID = "p1"
		befehl.Absender = "sitz-1"

		if _, fehler := leitung.Befehl(ctx, befehl); fehler != nil {
			t.Fatalf("%s: %v", befehl.Kennung, fehler)
		}
	}

	objekt := speicher.partien["p1"].Objekte[0]

	if objekt.Getappt {
		t.Error("das Objekt ist getappt geblieben")
	}

	if !objekt.Verdeckt {
		t.Error("das Umdrehen hat nicht gewirkt")
	}

	if objekt.HaengtAn != "" {
		t.Errorf("es hängt noch an %q", objekt.HaengtAn)
	}

	if _, noch := objekt.Zaehler["gift"]; noch {
		t.Errorf("der Zähler steht noch da: %v", objekt.Zaehler)
	}
}

// **Jeder Speicherweg einzeln, und keiner verschluckt seinen Fehler.**
//
// # Warum das ein zweiter Lauf ist
//
// **Weil der erste sie nie erreicht.** Ein Speicher, der auf **jedem** Weg
// klagt, klagt schon beim Lesen der Partie — und die drei Wege danach sind
// dann ungemessen. Das ist die Gestalt, vor der §10 warnt: der Lauf sieht
// vollständig aus, und drei Fehlerwege sind nie betreten worden.
func TestJederSpeicherwegKlagtFuerSich(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("gerade dieser Weg ist zu")

	for _, weg := range []string{"Befehlsergebnis", "LetzteKette", "Anwenden"} {
		t.Run(weg, func(t *testing.T) {
			t.Parallel()

			speicher := &partieprobe{
				partien:  map[string]domain.Partie{},
				stolpert: map[string]error{weg: kaputt},
			}

			leitung := service.NeuePartieleitung(speicher, service.NeuerSandkasten(service.NeuerZufall()))
			ctx := context.Background()

			if _, fehler := leitung.Anlegen(ctx, einePartie()); fehler != nil {
				t.Fatalf("Anlegen: %v", fehler)
			}

			_, fehler := leitung.Befehl(ctx, domain.Spielbefehl{
				Kennung: "b1", PartieID: "p1", Absender: "sitz-1",
				Art:      service.BefehlTappen,
				Nutzlast: map[string]any{"objekt": "o1"},
			})

			if !errors.Is(fehler, kaputt) {
				t.Fatalf("es kam %v", fehler)
			}
		})
	}
}

// **Ein Objekt hängt nicht an sich selbst.**
//
// **Der Weg, auf dem eine Auslegung einen ungültigen Zustand baut** — und
// der Grund, warum der Sandkasten nach dem Wirken noch einmal prüft. Ohne
// ihn stünde im Strom ein Ereignis über etwas, das es nicht geben darf.
func TestEinObjektHaengtNichtAnSichSelbst(t *testing.T) {
	t.Parallel()

	leitung, speicher := eineLeitung(t)
	ctx := context.Background()

	if _, fehler := leitung.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	_, fehler := leitung.Befehl(ctx, domain.Spielbefehl{
		Kennung: "b1", PartieID: "p1", Absender: "sitz-1",
		Art:      service.BefehlAnhaengen,
		Nutzlast: map[string]any{"objekt": "o1", "an": "o1"},
	})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}

	if speicher.angewendet != 0 {
		t.Error("der ungültige Zustand wurde geschrieben")
	}

	// **Und die Partie ist unberührt** — die Auslegung arbeitet auf einer
	// Kopie, und ein Fehlschlag hinterlässt keinen halben Stand.
	if speicher.partien["p1"].Objekte[0].HaengtAn != "" {
		t.Errorf("die Partie hat sich geändert: %+v",
			speicher.partien["p1"].Objekte[0])
	}
}
