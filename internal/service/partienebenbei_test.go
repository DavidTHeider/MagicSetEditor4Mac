package service_test

// **Wer geredet hat, sagt der Empfänger** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
// **Gegen dieselbe Lücke, die der Absender eines Befehls schon hatte.** Der
// Wirt setzt `Spielbefehl.Absender` aus der Sitzung, weil ein Gast sonst im
// Namen eines anderen zöge — bei Chat und Präsenz stand sie offen. Der erste
// Lauf hier ist deshalb der, in dem ein Gast sich „bert" nennt und „anna"
// dabei herauskommt.

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// eineNebennachricht baut eine, die ihre Gestalt hält.
func eineNebennachricht(
	t *testing.T, art domain.Nachrichtenart, nutzlast any,
) domain.Partienachricht {
	t.Helper()

	inhalt, fehler := json.Marshal(nutzlast)
	if fehler != nil {
		t.Fatalf("die Nutzlast ließ sich nicht verpacken: %v", fehler)
	}

	nachricht := domain.Partienachricht{
		Art: art, Partie: "p1", Von: "gast", Epoche: 1, Inhalt: inhalt,
	}

	if art.WirdAufbewahrt() {
		nachricht.Folge = 1
	}

	return nachricht
}

// **Der Sprecher kommt aus der Sitzung und nicht aus der Nachricht.**
//
// # Warum das der wichtigste Lauf dieser Datei ist
//
// **Weil es dieselbe Anmaßung ist wie ein fremder Absender.** Ohne diesen
// Schritt könnte ein Gast im Namen eines anderen Sitzes reden und dessen
// Zeiger bewegen — und niemand am Tisch könnte es unterscheiden.
func TestDerWirtSetztDenSprecherAusDerSitzung(t *testing.T) {
	t.Parallel()

	pruefer := service.NeuerNebenpruefer()

	faelle := []struct {
		name     string
		art      domain.Nachrichtenart
		nutzlast any
		von      func(service.Nebenbefund) string
	}{
		{
			"ein Satz",
			domain.ArtChat,
			domain.Chat{Von: "bert", Text: "gut gezogen"},
			func(befund service.Nebenbefund) string { return befund.Chat.Von },
		},
		{
			"ein Zeiger",
			domain.ArtPraesenz,
			domain.Praesenz{
				Von: "bert", Art: domain.PraesenzZeiger, Stand: 1,
			},
			func(befund service.Nebenbefund) string {
				return befund.Praesenz.Von
			},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			nachricht := eineNebennachricht(t, fall.art, fall.nutzlast)

			befund, fehler := pruefer.Beglaubige(nachricht, "anna")
			if fehler != nil {
				t.Fatalf("beglaubigen: %v", fehler)
			}

			if !befund.Gilt {
				t.Fatal("die Nachricht galt nicht")
			}

			if von := fall.von(befund); von != "anna" {
				t.Errorf("Sprecher %q — die Behauptung des Gastes hat gewonnen",
					von)
			}

			// **Auch die Nachricht trägt den beglaubigten Sprecher.** Sie
			// geht so ins Fenster; wer nur den Befund richtigstellte, hätte
			// zwei Wahrheiten in einem Ergebnis.
			if !strings.Contains(string(befund.Nachricht.Inhalt), `"von":"anna"`) {
				t.Errorf("die Nutzlast der Nachricht sagt etwas anderes: %s",
					befund.Nachricht.Inhalt)
			}
		})
	}
}

// **Ohne Sitz gilt, was in der Nutzlast steht** — das ist die Gastseite.
func TestDerGastGlaubtDenSprecherDesWirts(t *testing.T) {
	t.Parallel()

	pruefer := service.NeuerNebenpruefer()

	nachricht := eineNebennachricht(t, domain.ArtChat,
		domain.Chat{Von: "anna", Text: "hallo"})
	nachricht.Von = "wirt"

	befund, fehler := pruefer.Beglaubige(nachricht, "")
	if fehler != nil || !befund.Gilt {
		t.Fatalf("beglaubigen: %v (gilt: %v)", fehler, befund.Gilt)
	}

	if gesagt := befund.Chat; gesagt.Von != "anna" {
		t.Errorf("Sprecher %q — der Wirt ist die Autorität, und er sagte anna",
			gesagt.Von)
	}
}

// **Eine überholte Präsenz kommt nicht an, und das ist kein Fehler.**
func TestEineUeberholtePraesenzGiltNicht(t *testing.T) {
	t.Parallel()

	pruefer := service.NeuerNebenpruefer()

	zeige := func(stand uint64) (bool, error) {
		befund, fehler := pruefer.Beglaubige(
			eineNebennachricht(t, domain.ArtPraesenz, domain.Praesenz{
				Von: "anna", Art: domain.PraesenzZeiger, Stand: stand,
			}), "anna")

		return befund.Gilt, fehler
	}

	for _, fall := range []struct {
		name  string
		stand uint64
		gilt  bool
	}{
		{"die erste gilt", 5, true},
		{"dieselbe nicht", 5, false},
		{"eine ältere nicht", 3, false},
		{"eine neuere schon", 6, true},
	} {
		gilt, fehler := zeige(fall.stand)

		if fehler != nil {
			t.Fatalf("%s: %v", fall.name, fehler)
		}

		if gilt != fall.gilt {
			t.Errorf("%s: erwartet %v, bekommen %v",
				fall.name, fall.gilt, gilt)
		}
	}
}

// **Eine ungültige Präsenz schiebt den Stand nicht vor.**
//
// # Warum dieser Lauf dasteht
//
// **Weil die Reihenfolge der beiden Schritte eine Entscheidung ist.** Wer
// erst merkte und dann prüfte, ließe mit einer einzigen unsinnigen Nachricht
// den Stand auf eine Zahl springen, hinter der alles Gültige abgewiesen wird
// — ein Zeiger, den niemand mehr loswird.
func TestEineUngueltigePraesenzSchiebtDenStandNichtVor(t *testing.T) {
	t.Parallel()

	pruefer := service.NeuerNebenpruefer()

	kaputt := eineNebennachricht(t, domain.ArtPraesenz, domain.Praesenz{
		Von: "anna", Art: domain.Praesenzart("winken"), Stand: 999,
	})

	if _, fehler := pruefer.Beglaubige(kaputt, "anna"); fehler == nil {
		t.Fatal("eine erfundene Präsenzart ging durch")
	}

	gut := eineNebennachricht(t, domain.ArtPraesenz, domain.Praesenz{
		Von: "anna", Art: domain.PraesenzZeiger, Stand: 1,
	})

	befund, fehler := pruefer.Beglaubige(gut, "anna")
	if fehler != nil {
		t.Fatalf("beglaubigen: %v", fehler)
	}

	if !befund.Gilt {
		t.Error("die erste gültige Präsenz galt nicht — der Stand ist " +
			"durch die ungültige vorgeschoben worden")
	}
}

// **Was weder Chat noch Präsenz ist, hält an.**
func TestDerNebenprueferWeistFremdeArtenAb(t *testing.T) {
	t.Parallel()

	pruefer := service.NeuerNebenpruefer()

	faelle := map[string]domain.Partienachricht{
		"ein Ereignis": eineNebennachricht(
			t, domain.ArtEreignis, domain.Ereignis{}),
		"ein Satz, der keiner ist": {
			Art: domain.ArtChat, Partie: "p1", Von: "gast",
			Epoche: 1, Folge: 1, Inhalt: []byte(`"kein Verbund"`),
		},
		"eine Präsenz, die keine ist": {
			Art: domain.ArtPraesenz, Partie: "p1", Von: "gast",
			Epoche: 1, Inhalt: []byte(`42`),
		},
		"ein leerer Satz": eineNebennachricht(
			t, domain.ArtChat, domain.Chat{Von: "anna"}),
	}

	for name, nachricht := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := pruefer.Beglaubige(nachricht, "anna"); fehler == nil {
				t.Fatal("es ging durch, obwohl es hätte anhalten sollen")
			} else if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("Fehler %v — erwartet war eine Validierung", fehler)
			}
		})
	}
}

// **Ein Sitz, der aufsteht, wird vergessen.**
func TestDerNebenprueferVergisstEinenSitz(t *testing.T) {
	t.Parallel()

	pruefer := service.NeuerNebenpruefer()

	spaet := eineNebennachricht(t, domain.ArtPraesenz, domain.Praesenz{
		Von: "anna", Art: domain.PraesenzZeiger, Stand: 50,
	})

	befund, fehler := pruefer.Beglaubige(spaet, "anna")
	if fehler != nil || !befund.Gilt {
		t.Fatalf("die erste Präsenz galt nicht: %v", fehler)
	}

	frisch := eineNebennachricht(t, domain.ArtPraesenz, domain.Praesenz{
		Von: "anna", Art: domain.PraesenzZeiger, Stand: 1,
	})

	zurueck, fehler := pruefer.Beglaubige(frisch, "anna")
	if fehler != nil || zurueck.Gilt {
		t.Fatalf("ein Rückschritt galt, obwohl anna noch am Tisch saß: %v",
			fehler)
	}

	pruefer.Vergiss("anna")

	wieder, fehler := pruefer.Beglaubige(frisch, "anna")
	if fehler != nil || !wieder.Gilt {
		t.Errorf("nach dem Aufstehen galt der frische Anfang nicht: %v",
			fehler)
	}
}
