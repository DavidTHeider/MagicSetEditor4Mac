package service_test

// **Die Fehlerwege der vier Kontoarten** — Werkstück 6 von Phase 4.12
// (#1351).
//
// **Sie stehen eigens hier**, weil sie eine eigene Verhaltensklasse sind: was
// geschieht, wenn nicht der Nutzer sich vertippt, sondern etwas darunter
// schiefgeht. Wer sie zwischen die Erfolgswege mischte, hätte die eine Datei,
// an der jede Änderung stattfindet (§2).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/kennwort"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// errStummerBestand ist ein Fehler, der **kein** [domain.ErrNotFound] ist.
var errStummerBestand = errors.New("der Bestand antwortet nicht")

// **Ein kaputter Bestand reißt ab und wird nicht für „gibt es nicht"
// gehalten.**
//
// **Beide Wege, weil beide dieselbe Unterscheidung treffen**: das Eintragen
// fragt, ob die Kennung frei ist, das Anmelden, wem sie gehört — und ein
// Lesefehler sähe an beiden Stellen aus wie ein freier Platz. Beim Eintragen
// entstünde ein zweites Konto unter derselben Kennung, beim Anmelden käme die
// Abweisung für einen Fehler, den der Nutzer nicht beheben kann (§8).
func TestEinKaputterBestandReisstAb(t *testing.T) {
	t.Parallel()

	wege := map[string]func(*service.Kontoanmeldung) error{
		"beim Eintragen": func(konto *service.Kontoanmeldung) error {
			_, fehler := konto.Eintragen(t.Context(), service.Kontowunsch{
				Kennung: "wer@example.org", Kennwort: "einlangesgeheimnis",
			})

			return fehler
		},
		"beim Anmelden": func(konto *service.Kontoanmeldung) error {
			_, fehler := konto.Anmelden(t.Context(),
				"wer@example.org", "einlangesgeheimnis", "")

			return fehler
		},
		"beim anonymen Anlegen": func(konto *service.Kontoanmeldung) error {
			_, fehler := konto.Anonym(t.Context())

			return fehler
		},
	}

	for name, weg := range wege {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := weg(service.NeueKontoanmeldung(&kennungsdoppel{kaputt: errStummerBestand}))
			if !errors.Is(fehler, errStummerBestand) {
				t.Fatalf("durchgelassen wurde: %v", fehler)
			}
		})
	}
}

// **Ein krummes Kennwort im Bestand ist ein Fehler und keine Abweisung.**
//
// **Weil es keiner ist, den der Nutzer gemacht hat.** Wer die Bestandsdatei
// von Hand geändert hat, soll das erfahren — „Kennung oder Kennwort passen
// nicht zusammen" schickte ihn auf die Suche nach seinem Tippfehler.
//
// **Und der Riegel davor greift nicht**: das Feld ist nicht leer, es ist
// unbrauchbar. Beides gehört unterschieden.
func TestEinKrummesKennwortImBestandIstEinFehler(t *testing.T) {
	t.Parallel()

	ablage := &kennungsdoppel{bekannt: domain.Nutzer{
		ID: 1, Name: "Ich", Email: "wer@example.org",
		Kennwort: "$argon2id$kaputt",
	}}

	_, fehler := service.NeueKontoanmeldung(ablage).Anmelden(t.Context(),
		"wer@example.org", "einlangesgeheimnis", "")
	if fehler == nil || fehler.Error() == abweisung {
		t.Fatalf("gemeldet wurde: %v", fehler)
	}
}

// **Ein zu langes Kennwort wird abgewiesen, und die Meldung nennt das Feld.**
//
// **Die Obergrenze ist keine Schikane**: Argon2id rechnet über die ganze
// Eingabe, und wer ein Megabyte schickt, beschäftigt den Rechner damit.
func TestEinZuLangesKennwortWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	ablage := &kennungsdoppel{}

	_, fehler := service.NeueKontoanmeldung(ablage).Eintragen(t.Context(),
		service.Kontowunsch{
			Kennung:  "wer@example.org",
			Kennwort: strings.Repeat("x", kennwort.Hoechstlaenge+1),
		})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("angenommen wurde: %v", fehler)
	}

	var feld *domain.FieldError
	if !errors.As(fehler, &feld) || feld.Field != "kennwort" {
		t.Errorf("gemeldet wurde %v", fehler)
	}

	if ablage.angelegt.Name != "" {
		t.Errorf("trotzdem angelegt: %+v", ablage.angelegt)
	}
}

// **Ein krummer Name wird abgewiesen, obwohl die Kennung taugt.**
//
// **Der Weg, den niemand vermutet**: die Kennung ist geprüft, das Kennwort
// gehütet — und dann sagt [domain.Nutzer.Validate] nein. Ohne diesen Fall
// bliebe der letzte Riegel vor dem Bestand unbetreten.
func TestEinKrummerKontonameWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	ablage := &kennungsdoppel{}

	_, fehler := service.NeueKontoanmeldung(ablage).Eintragen(t.Context(),
		service.Kontowunsch{
			Name:     "Wer ",
			Kennung:  "wer@example.org",
			Kennwort: "einlangesgeheimnis",
		})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("angenommen wurde: %v", fehler)
	}

	if ablage.angelegt.Name != "" {
		t.Errorf("trotzdem angelegt: %+v", ablage.angelegt)
	}
}

// **Eine krumme Kennung wird nach dem gemeldet, was sie sein wollte.**
//
// **Weil die Meldung sonst in die Irre schickt**: wer `+4915` eintippt, soll
// „eine Rufnummer hat 6 bis 15 Ziffern" lesen und nicht „eine Adresse hat
// genau ein @" — und wer `wer@` eintippt, umgekehrt.
//
// **Und weil die Alternative eine Reihenfolge von Versuchen wäre**: erst die
// Adresse, bei einem Fehler die Nummer — und am Ende stünde der Fehler der
// zweiten Prüfung unter einer vertippten Adresse.
func TestEineKrummeKennungWirdNachIhrerArtGemeldet(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct{ kennung, erwartet string }{
		"zu kurze Rufnummer":  {kennung: "+4915", erwartet: "Rufnummer"},
		"Adresse ohne Domäne": {kennung: "wer@", erwartet: "Adresse"},
		"Adresse ohne Postfach": {
			kennung: "@example.org", erwartet: "Adresse",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := service.NeueKontoanmeldung(&kennungsdoppel{}).Eintragen(
				t.Context(), service.Kontowunsch{
					Kennung: fall.kennung, Kennwort: "einlangesgeheimnis",
				})
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("angenommen wurde: %v", fehler)
			}

			if !strings.Contains(fehler.Error(), fall.erwartet) {
				t.Errorf("gemeldet wurde %q", fehler.Error())
			}
		})
	}
}

// **Ein Konto auf eine Rufnummer ohne Namen heißt wie seine Nummer.**
//
// **Und nicht wie ihr vorderer Teil**: eine Nummer hat keinen, und wer den
// Namen aus der Adresse auch auf sie anwendete, bekäme die ganze Nummer —
// oder, hätte sie je ein `@`, ein Bruchstück.
func TestEinKontoAufEineRufnummerHeisstWieSeineNummer(t *testing.T) {
	t.Parallel()

	ablage := &kennungsdoppel{}

	if _, fehler := service.NeueKontoanmeldung(ablage).Eintragen(t.Context(),
		service.Kontowunsch{
			Kennung: "+4915112345678", Kennwort: "einlangesgeheimnis",
		}); fehler != nil {
		t.Fatalf("eintragen: %v", fehler)
	}

	if ablage.angelegt.Name != "+4915112345678" {
		t.Errorf("der Name wurde %q", ablage.angelegt.Name)
	}
}
