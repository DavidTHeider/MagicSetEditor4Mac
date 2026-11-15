package service_test

// **Ein Kartenskript aus einem fremden Spielpaket** — QA-06c
// (#1976),
// Fläche 19 aus `docs/FUZZZIELE.md`.
//
// `add cards script:` steht im Spielpaket, und wer ein Paket bezieht, führt
// es mit einem Klick aus. Drei Zusagen, die bis hierher fehlten:
//
//   - **Nur `new_card` macht eine Karte.** MSE2 wandelt jedes Glied mit
//     `from_script<CardP>` und weist alles andere ab
//     (`AddCardsScript::perform`, `src/data/add_cards_script.cpp`, Zeile 33,
//     Fassung f9d9356 vom 2021-07-06). Hier nahm `alsKarte` jede Abbildung —
//     auch eine mit einem Bildfeld, das `new_card` verbietet, und eine mit
//     Feldern, die es nicht gibt.
//   - **Eine Frist.** Der Kontext einer Bindung hat keine; eine Schleife im
//     Skript hielte sie fest.
//   - **Eine Grenze der Kartenzahl.** Sonst gingen Millionen Karten in einem
//     Zug in den Bestand.

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func spielMit(name, skript string) *spielprobe {
	return &spielprobe{
		skripte: []service.Kartenskriptangabe{{Name: name, Skript: skript}},
		felder:  probefelder(),
	}
}

func TestEineAbbildungOhneNewCardIstKeineKarte(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ name, skript string }{
		{"ein Bildfeld, das new_card verbietet", `[ [kunst: "/fremd/bild.png"] ]`},
		{"ein Feld, das es nicht gibt", `[ [gibtsnicht: "x"] ]`},
		{"eine Karte und eine Abbildung", `[ new_card([name: "Plains"]), [name: "Island"] ]`},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			anleger := &anlegerprobe{}

			_, fehler := service.NeuerKartenskriptdienst(anleger, spielMit("fremd", fall.skript)).
				Ausfuehren(t.Context(), 1, "fremd")
			if !errors.Is(fehler, domain.ErrValidation) || anleger.rufe != 0 {
				t.Errorf("Fehler %v, %d Rufe an die Ablage", fehler, anleger.rufe)
			}
		})
	}
}

func TestEinKartenskriptHatEineFrist(t *testing.T) {
	t.Parallel()

	anleger := &anlegerprobe{}
	dienst := service.NeuerKartenskriptdienst(anleger,
		spielMit("endlos", `for x from 1 to 100000000 do [new_card()]`)).
		MitFrist(50 * time.Millisecond)

	ergebnis := make(chan error, 1)

	go func() {
		_, fehler := dienst.Ausfuehren(context.Background(), 1, "endlos")
		ergebnis <- fehler
	}()

	select {
	case fehler := <-ergebnis:
		if fehler == nil || anleger.rufe != 0 {
			t.Errorf("Fehler %v, %d Rufe an die Ablage", fehler, anleger.rufe)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("das Kartenskript lief über zehn Sekunden — ohne Frist")
	}
}

// grenzfrist ist die Frist der beiden Läufe über die Grenze.
//
// **Großzügig, und das ist die Aussage**: unter `-race` und voller Last
// brauchten 10.000 `new_card` am 2026-10-01 mehr als die 30 s der Vorgabe.
// Der Fristablauf ist **auch** eine Validierung — mit der Vorgabe wäre der
// Lauf über die Grenze an der Frist gescheitert und grün gewesen, ohne die
// Grenze je zu betreten.
const grenzfrist = 10 * time.Minute

func TestEinKartenskriptHatEineGrenze(t *testing.T) {
	t.Parallel()

	anleger := &anlegerprobe{}

	_, fehler := service.NeuerKartenskriptdienst(anleger,
		spielMit("viele", `for x from 0 to `+strconv.Itoa(service.Kartenskriptgrenze)+` do [new_card()]`)).
		MitFrist(grenzfrist).
		Ausfuehren(t.Context(), 1, "viele")
	if !errors.Is(fehler, domain.ErrValidation) || errors.Is(fehler, context.DeadlineExceeded) ||
		anleger.rufe != 0 {
		t.Errorf("Fehler %v, %d Rufe an die Ablage", fehler, anleger.rufe)
	}
}

// **Und an der Grenze selbst geht es noch.**
func TestGenauDieGrenzeGehtNoch(t *testing.T) {
	t.Parallel()

	anleger := &anlegerprobe{}

	lauf, fehler := service.NeuerKartenskriptdienst(anleger,
		spielMit("genug", `for x from 1 to `+strconv.Itoa(service.Kartenskriptgrenze)+` do [new_card()]`)).
		MitFrist(grenzfrist).
		Ausfuehren(t.Context(), 1, "genug")
	if fehler != nil || len(lauf.Karten) != service.Kartenskriptgrenze {
		t.Errorf("Fehler %v", fehler)
	}
}
