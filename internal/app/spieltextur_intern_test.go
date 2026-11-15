package app

// **Die Texturnaht** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Wogegen hier geprüft wird
//
// **Dass eine Adresse hinausgeht und keine Bytes.** Das Issue sagt es
// wörtlich: „keine Bild-Base64-Rückkehr durch die Wails-Naht" — und der Grund
// ist gemessen: 182 KB kosten als Base64 **554 ms** über die Brücke
// (#1270),
// während der ganze Go-Weg 53 ms braucht. Ein Spieltisch zeigt nicht eine
// Karte, sondern sechzig.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// texturprobe gibt eine Textur — oder klagt.
type texturprobe struct {
	klagt error

	gesehen struct {
		seite service.Kartenseite
		guete service.Texturguete
		spiel string
	}

	ruecken int
}

func (t *texturprobe) Textur(
	_ context.Context, _ *domain.CardRow,
	seite service.Kartenseite, guete service.Texturguete,
) (service.Textur, error) {
	t.gesehen.seite, t.gesehen.guete = seite, guete

	if t.klagt != nil {
		return service.Textur{}, t.klagt
	}

	return service.Textur{
		Schluessel: "abc123",
		Adresse:    Spielbildpfad + "abc123.png",
		Guete:      guete,
		Breite:     512, Hoehe: 714,
		Befunde: []string{"die Schrift fehlt"},
	}, nil
}

func (t *texturprobe) Ruecken(
	_ context.Context, spiel string, guete service.Texturguete,
) (service.Textur, error) {
	t.ruecken++
	t.gesehen.spiel = spiel
	t.gesehen.guete = guete

	if t.klagt != nil {
		return service.Textur{}, t.klagt
	}

	return service.Textur{
		Schluessel: "ruecken1",
		Adresse:    Spielbildpfad + "ruecken1.png",
		Guete:      guete,
		Breite:     512, Hoehe: 714,
		Befunde: []string{},
	}, nil
}

// texturkartenleser gibt immer dieselbe Karte — oder klagt.
//
// **Und er zählt mit.** „Der Rücken fragt keine Karte" ist eine Aussage über
// das, was **nicht** geschieht; ohne einen Zähler ließe sie sich nur daran
// ablesen, dass nichts schiefging — und das täte sie auch, wenn er gefragt
// worden wäre.
type texturkartenleser struct {
	klagt error

	gefragt int
}

func (t *texturkartenleser) KarteLesen(context.Context, int64) (
	*domain.CardRow, error,
) {
	t.gefragt++

	if t.klagt != nil {
		return nil, t.klagt
	}

	return &domain.CardRow{ID: 7}, nil
}

func (*texturkartenleser) Herkuenfte(context.Context, int64) (
	map[string]string, error,
) {
	return map[string]string{}, nil
}

// texturanwendung baut eine Anwendung mit Texturwerk und Kartenleser.
func texturanwendung(t *testing.T, werk Texturwerk, leser Kartenleser) *App {
	t.Helper()

	anwendung := NeueApp("")
	anwendung.db = bestandFuerDenRiegel(t)
	anwendung.profil.Store(1)
	anwendung.kartenleser = leser

	return anwendung.MitTexturwerk(werk)
}

// **Eine Adresse geht hinaus, keine Bytes.**
func TestDieTexturnahtGibtEineAdresse(t *testing.T) {
	t.Parallel()

	werk := &texturprobe{}

	antwort := texturanwendung(t, werk, &texturkartenleser{}).
		Spieltextur(t.Context(), "7", "vorne", "tisch")
	if antwort.Fehler != nil {
		t.Fatalf("Spieltextur: %+v", antwort.Fehler)
	}

	switch {
	case !strings.HasPrefix(antwort.Textur.Adresse, Spielbildpfad):
		t.Errorf("Adresse %q", antwort.Textur.Adresse)
	case antwort.Textur.Schluessel == "":
		t.Error("die Antwort trägt keinen Schlüssel")
	case antwort.Textur.Breite != 512 || antwort.Textur.Hoehe != 714:
		t.Errorf("Maße %d×%d", antwort.Textur.Breite, antwort.Textur.Hoehe)
	case antwort.Textur.Seite != "vorne":
		t.Errorf("Seite %q", antwort.Textur.Seite)
	case len(antwort.Textur.Befunde) != 1:
		t.Errorf("Befunde %v", antwort.Textur.Befunde)
	}

	// **Und Seite und Güte kamen unverändert an.**
	if werk.gesehen.seite != service.SeiteVorne ||
		werk.gesehen.guete != service.GueteTisch {
		t.Errorf("gerufen mit %+v", werk.gesehen)
	}
}

// **Was auf dem Weg klagt, kommt als Fehler zurück — mit leeren Listen.**
func TestWasAnDerTexturnahtKlagtKommtDurch(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("der Bestand ist zu")

	for name, bauen := range map[string]func(*testing.T) *App{
		"der Kartenleser": func(t *testing.T) *App {
			return texturanwendung(t, &texturprobe{},
				&texturkartenleser{klagt: kaputt})
		},
		"das Texturwerk": func(t *testing.T) *App {
			return texturanwendung(t, &texturprobe{klagt: kaputt},
				&texturkartenleser{})
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := bauen(t).Spieltextur(t.Context(), "7", "vorne", "tisch")

			if antwort.Fehler == nil {
				t.Fatal("die Klage kam nicht durch")
			}

			// **Leere Listen und kein `nil`** — dieselbe Zusage wie auf dem
			// Erfolgsweg. Genau hier ist sie in Phase 5.4 zweimal gebrochen.
			if antwort.Textur.Befunde == nil {
				t.Error("der Fehlerweg gibt nil")
			}
		})
	}
}

// **Eine Kennung, die keine Zahl ist, wird abgewiesen — am Feld.**
func TestEineUnlesbareKartenkennungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	antwort := texturanwendung(t, &texturprobe{}, &texturkartenleser{}).
		Spieltextur(t.Context(), "keine Zahl", "vorne", "tisch")

	if antwort.Fehler == nil {
		t.Fatal("die unlesbare Kennung kam durch")
	}

	if antwort.Fehler.Feld != "karteID" {
		t.Errorf("der Fehler hängt an %q", antwort.Fehler.Feld)
	}
}

// **Der Rücken kommt ohne Kartenkennung durch die Naht.**
//
// # Wofür dieser Lauf da ist
//
// **Weil ein verdecktes Objekt keine Kennung trägt.** Die Projektion aus G1
// hält sie zurück — „eine Bibliothek ist auch für ihren Eigner verdeckt" —,
// und ein Weg, der eine verlangte, gäbe genau dort nichts her.
//
// **Und der Kartenleser wird dafür nicht gefragt.** Das ist die zweite
// Hälfte: wer ihn fragte, bände Maße und Schlüssel des Rückens an die Karte
// darunter — und verriete sie über die Metadaten (#1459).
func TestDerRueckenBrauchtKeineKartenkennung(t *testing.T) {
	t.Parallel()

	leser := &texturkartenleser{}
	werk := &texturprobe{}

	antwort := texturanwendung(t, werk, leser).
		Rueckentextur(t.Context(), "magic", "tisch")

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %v", antwort.Fehler)
	case leser.gefragt != 0:
		t.Errorf("der Kartenleser wurde %dmal gefragt, erwartet keinmal",
			leser.gefragt)
	case werk.gesehen.spiel != "magic":
		t.Errorf("Spiel %q", werk.gesehen.spiel)
	case antwort.Textur.Seite != "ruecken":
		t.Errorf("Seite %q", antwort.Textur.Seite)
	case antwort.Textur.Adresse == "":
		t.Error("keine Adresse")
	}
}

// **Und was beim Rücken klagt, kommt übersetzt zurück.**
//
// **Der Fehlerweg gehört dazu** (§10): er geht nicht durch
// [App.Spieltextur] und hätte sonst keinen Lauf.
func TestEinFehlerAmRueckenKommtUebersetztZurueck(t *testing.T) {
	t.Parallel()

	werk := &texturprobe{klagt: errors.New("kein Rasterer")}

	antwort := texturanwendung(t, werk, &texturkartenleser{}).
		Rueckentextur(t.Context(), "magic", "tisch")

	if antwort.Fehler == nil {
		t.Fatal("kein Fehler")
	}

	if len(antwort.Textur.Befunde) != 0 || antwort.Textur.Befunde == nil {
		t.Errorf("Befunde %v — erwartet eine leere Liste", antwort.Textur.Befunde)
	}
}

// **Eine Kartenseite braucht weiterhin eine Kennung.**
//
// **Die Gegenprobe zum Weg daneben**: dass der Rücken ohne auskommt, heißt
// nicht, dass eine Karte es täte — dort ist die leere Kennung ein Fehler im
// Aufrufer und keine Lage.
func TestEineKartenseiteBrauchtEineKennung(t *testing.T) {
	t.Parallel()

	leser := &texturkartenleser{}

	antwort := texturanwendung(t, &texturprobe{}, leser).
		Spieltextur(t.Context(), "", "vorne", "tisch")

	if antwort.Fehler == nil {
		t.Fatal("kein Fehler")
	}

	if leser.gefragt != 0 {
		t.Errorf("der Kartenleser wurde %dmal gefragt", leser.gefragt)
	}
}

// **Beide Texturwege fragen nach der Rolle — und zeichnen nicht, wenn die
// Antwort Nein ist.**
//
// # Warum das hier steht und nicht nur in der Matrix
//
// **Weil `docs/AUTORISIERUNG.md` eine Zeile ist und keine Prüfung.** Sie
// sagt, dass beide `lokalLesen` sind und einen Leser voraussetzen; dass der
// Riegel **gefragt** wird, sagt sie nicht — und ein Weg, der ihn vergäße,
// sähe in der Matrix genauso aus.
func TestBeideTexturwegeFragenNachDerRolle(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(*App) TexturAntwort{
		"Spieltextur": func(a *App) TexturAntwort {
			return a.Spieltextur(t.Context(), "7", "vorne", "tisch")
		},
		"Rueckentextur": func(a *App) TexturAntwort {
			return a.Rueckentextur(t.Context(), "magic", "tisch")
		},
		// **Und das fremde Bild** (P9, #2377) — es liest, wie die beiden.
		"Fremdkartenbild": func(a *App) TexturAntwort {
			return a.Fremdkartenbild(t.Context(), "p1", "o1", "b1", "vorne", "tisch")
		},
	}

	for name, fahre := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wache := &zaehlendeWache{}
			werk := &texturprobe{}
			leser := &texturkartenleser{}

			antwort := fahre(texturanwendung(t, werk, leser).MitWache(wache))

			switch {
			case antwort.Fehler == nil:
				t.Fatal("der Weg kam durch")
			case len(wache.gefragt) != 1 ||
				wache.gefragt[0] != domain.WirkungLesen:
				t.Errorf("gefragt wurde nach %v", wache.gefragt)
			case werk.ruecken != 0:
				t.Errorf("der Rücken wurde %dmal gezeichnet", werk.ruecken)
			case leser.gefragt != 0:
				t.Errorf("die Karte wurde %dmal gelesen", leser.gefragt)
			}

			// **Und die leere Antwort trägt trotzdem eine leere Liste.**
			if antwort.Textur.Befunde == nil {
				t.Error("Befunde sind `nil`")
			}
		})
	}
}
