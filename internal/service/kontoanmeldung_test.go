package service_test

// **Die vier Kontoarten** — Werkstück 6 von Phase 4.12
// (#1351).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/kennwort"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Wer sich mit einer Adresse einträgt, kommt danach mit ihr herein.**
//
// **Das ist die Abnahme des Werkstücks** und nicht ein grüner Lauf: die
// Anwendung schließen, öffnen, Adresse und Kennwort eingeben — derselbe
// Nutzer. Der Test fährt genau diesen Weg zweimal über dieselbe Ablage.
func TestWerSichEintraegtKommtDamitHerein(t *testing.T) {
	t.Parallel()

	ablage := &kennungsdoppel{}
	konto := service.NeueKontoanmeldung(ablage)

	eingetragen, fehler := konto.Eintragen(t.Context(), service.Kontowunsch{
		Kennung: "wer@example.org", Kennwort: "einlangesgeheimnis",
	})
	if fehler != nil {
		t.Fatalf("eintragen: %v", fehler)
	}

	herein, fehler := konto.Anmelden(t.Context(),
		"wer@example.org", "einlangesgeheimnis", "")
	if fehler != nil {
		t.Fatalf("anmelden: %v", fehler)
	}

	if herein.ID != eingetragen.ID {
		t.Errorf("herein kam %d, eingetragen war %d", herein.ID, eingetragen.ID)
	}
}

// **Eine Rufnummer tut dasselbe wie eine Adresse.**
//
// **Ein Weg für beide**, weil beide dasselbe sagen: wer sich anmeldet. Zwei
// Wege hätten zwei Fehlermeldungen für denselben Fall.
func TestEineRufnummerTutDasselbe(t *testing.T) {
	t.Parallel()

	ablage := &kennungsdoppel{}
	konto := service.NeueKontoanmeldung(ablage)

	if _, fehler := konto.Eintragen(t.Context(), service.Kontowunsch{
		Name: "Wer", Kennung: "+4915112345678", Kennwort: "einlangesgeheimnis",
	}); fehler != nil {
		t.Fatalf("eintragen: %v", fehler)
	}

	if ablage.angelegt.Telefon != "+4915112345678" {
		t.Errorf("angelegt wurde %+v", ablage.angelegt)
	}

	if ablage.angelegt.Email != "" {
		t.Errorf("die Nummer landete in der Adresse: %+v", ablage.angelegt)
	}

	if _, fehler := konto.Anmelden(t.Context(),
		"+4915112345678", "einlangesgeheimnis", ""); fehler != nil {
		t.Fatalf("anmelden: %v", fehler)
	}
}

// **Das Kennwort steht nie im Klartext im Bestand.**
//
// **Die Zusage des Werkstücks**, und sie wird hier geprüft und nicht
// behauptet: was die Ablage zu sehen bekommt, ist eine PHC-Zeichenkette.
func TestDasKennwortStehtNieImKlartext(t *testing.T) {
	t.Parallel()

	ablage := &kennungsdoppel{}

	if _, fehler := service.NeueKontoanmeldung(ablage).Eintragen(t.Context(),
		service.Kontowunsch{
			Kennung: "wer@example.org", Kennwort: "einlangesgeheimnis",
		}); fehler != nil {
		t.Fatalf("eintragen: %v", fehler)
	}

	if strings.Contains(ablage.angelegt.Kennwort, "einlangesgeheimnis") {
		t.Fatalf("das Kennwort steht im Bestand: %q", ablage.angelegt.Kennwort)
	}

	if !strings.HasPrefix(ablage.angelegt.Kennwort, "$argon2id$") {
		t.Errorf("gehütet wurde %q", ablage.angelegt.Kennwort)
	}

	if fehler := kennwort.Stimmt(
		ablage.angelegt.Kennwort, "einlangesgeheimnis"); fehler != nil {
		t.Errorf("das Gehütete passt nicht zum Kennwort: %v", fehler)
	}
}

// abweisung ist der **eine** Satz, den jede gescheiterte Anmeldung gibt.
//
// **Er steht als Zeichenkette da und nicht als Aufruf des Dienstes**: eine
// Erwartung, die sich aus dem Geprüften ableitet, ist immer erfüllt.
const abweisung = "Kennung oder Kennwort passen nicht zusammen."

// **Ein falsches Kennwort kommt nicht herein — und ein leeres auch nicht.**
//
// **Das Leere steht daneben, weil es der gefährlichere Fall ist**: ein Nutzer
// ohne Kennwort (anonym, oder der erste) hat ein leeres Feld im Bestand, und
// ein Vergleich, der Leeres gegen Leeres hält, ließe jeden herein.
//
// # Warum die Meldung Wort für Wort verglichen wird
//
// **Weil die Fehlerklasse allein nichts unterscheidet.** Der erste Anlauf
// prüfte `errors.Is(…, domain.ErrValidation)`, und der Fall „Nutzer ohne
// Kennwort" war grün, **ohne den Riegel im Dienst**: `kennwort.Stimmt` gibt
// für ein leeres Feld „die gehütete Form ist keine von argon2id" — auch ein
// `ErrValidation`. Der Satz stand an der Oberfläche und sagte einem Fremden,
// dass es dieses Konto gibt und es kein Kennwort hat. Die Gegenprobe blieb
// grün und zeigte damit nicht den Riegel, sondern die Lücke im Test (§10).
func TestEinFalschesKennwortKommtNichtHerein(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		nutzer   domain.Nutzer
		kennwort string
	}{
		"falsches Kennwort": {
			nutzer:   mitKennwort(t, "wer@example.org", "einlangesgeheimnis"),
			kennwort: "einanderesgeheimnis",
		},
		"leeres Kennwort gegen ein gesetztes": {
			nutzer:   mitKennwort(t, "wer@example.org", "einlangesgeheimnis"),
			kennwort: "",
		},
		"leeres Kennwort gegen einen Nutzer ohne": {
			nutzer:   domain.Nutzer{ID: 1, Name: "Ich", Email: "wer@example.org"},
			kennwort: "",
		},
		"irgendein Kennwort gegen einen Nutzer ohne": {
			nutzer:   domain.Nutzer{ID: 1, Name: "Ich", Email: "wer@example.org"},
			kennwort: "einlangesgeheimnis",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ablage := &kennungsdoppel{bekannt: fall.nutzer}

			_, fehler := service.NeueKontoanmeldung(ablage).Anmelden(t.Context(),
				"wer@example.org", fall.kennwort, "")
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("herein kam: %v", fehler)
			}

			// **Die Meldung nennt nicht, welches der beiden Felder falsch
			// war** — die Auskunft hülfe niemandem außer dem, der rät.
			if fehler.Error() != abweisung {
				t.Errorf("die Meldung lautet %q", fehler.Error())
			}
		})
	}
}

// **Eine unbekannte Kennung kommt nicht herein.**
func TestEineUnbekannteKennungKommtNichtHerein(t *testing.T) {
	t.Parallel()

	ablage := &kennungsdoppel{fehlt: true}

	_, fehler := service.NeueKontoanmeldung(ablage).Anmelden(t.Context(),
		"wer@example.org", "einlangesgeheimnis", "")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("herein kam: %v", fehler)
	}

	// **Wort für Wort derselbe Satz wie beim falschen Kennwort** — sonst sagt
	// der Unterschied, welche Kennungen es gibt.
	if fehler.Error() != abweisung {
		t.Errorf("die Meldung lautet %q", fehler.Error())
	}
}

// **Ein anonymes Konto bekommt eine Zufallskennung und kein Kennwort.**
func TestEinAnonymesKontoBekommtEineZufallskennung(t *testing.T) {
	t.Parallel()

	gesehen := map[string]bool{}

	for range 32 {
		ablage := &kennungsdoppel{}

		angelegt, fehler := service.NeueKontoanmeldung(ablage).Anonym(t.Context())
		if fehler != nil {
			t.Fatalf("anonym: %v", fehler)
		}

		if !angelegt.Anonym {
			t.Errorf("der Vermerk fehlt: %+v", angelegt)
		}

		if angelegt.Email != "" || angelegt.Telefon != "" ||
			ablage.angelegt.Kennwort != "" {
			t.Errorf("anonym ist er nicht: %+v", ablage.angelegt)
		}

		if gesehen[angelegt.Name] {
			t.Fatalf("zweimal derselbe Name: %q", angelegt.Name)
		}

		gesehen[angelegt.Name] = true
	}
}

// **Eine krumme Kennung wird abgewiesen, und die Meldung nennt das Feld.**
func TestEineKrummeKennungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]service.Kontowunsch{
		"gar keine Kennung": {Kennwort: "einlangesgeheimnis"},
		"weder Adresse noch Nummer": {
			Kennung: "kein Postfach", Kennwort: "einlangesgeheimnis",
		},
		"Adresse ohne Kennwort": {Kennung: "wer@example.org"},
		"zu kurzes Kennwort":    {Kennung: "wer@example.org", Kennwort: "kurz"},
	}

	for name, wunsch := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ablage := &kennungsdoppel{}

			_, fehler := service.NeueKontoanmeldung(ablage).Eintragen(
				t.Context(), wunsch)
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("angenommen wurde: %v", fehler)
			}

			if ablage.angelegt.Name != "" {
				t.Errorf("trotzdem angelegt: %+v", ablage.angelegt)
			}
		})
	}
}

// **Eine belegte Kennung wird abgewiesen, bevor irgendetwas entsteht.**
func TestEineBelegteKennungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	ablage := &kennungsdoppel{
		bekannt: domain.Nutzer{ID: 3, Name: "Schon da", Email: "wer@example.org"},
	}

	_, fehler := service.NeueKontoanmeldung(ablage).Eintragen(t.Context(),
		service.Kontowunsch{
			Kennung: "wer@example.org", Kennwort: "einlangesgeheimnis",
		})
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("angenommen wurde: %v", fehler)
	}

	if ablage.angelegt.Name != "" {
		t.Errorf("trotzdem angelegt: %+v", ablage.angelegt)
	}
}

// **Ohne Namen bekommt ein Konto einen aus seiner Kennung.**
//
// **Weil ein Nutzer einen Namen haben muss** (`domain.Nutzer.Validate`) und
// niemand zweimal dasselbe eintippen soll.
func TestOhneNamenBekommtEinKontoEinenAusSeinerKennung(t *testing.T) {
	t.Parallel()

	ablage := &kennungsdoppel{}

	if _, fehler := service.NeueKontoanmeldung(ablage).Eintragen(t.Context(),
		service.Kontowunsch{
			Kennung: "wer@example.org", Kennwort: "einlangesgeheimnis",
		}); fehler != nil {
		t.Fatalf("eintragen: %v", fehler)
	}

	if ablage.angelegt.Name != "wer" {
		t.Errorf("der Name wurde %q", ablage.angelegt.Name)
	}
}

// mitKennwort baut einen Nutzer mit gehütetem Kennwort.
func mitKennwort(t *testing.T, adresse, klartext string) domain.Nutzer {
	t.Helper()

	gehuetet, fehler := kennwort.Huete(klartext)
	if fehler != nil {
		t.Fatalf("hüten: %v", fehler)
	}

	return domain.Nutzer{ID: 1, Name: "Ich", Email: adresse, Kennwort: gehuetet}
}
