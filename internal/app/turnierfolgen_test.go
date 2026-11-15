package app_test

// **Die Folgen einer Berichtigung an der Naht** — DOM-04.AK-02
// (#1699).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieFolgenEinerBerichtigungKommenAlsDTO(t *testing.T) {
	anwendung, naht := turnieranwendung(t)

	antwort := anwendung.TurnierKorrekturfolgen(t.Context(), "7", "12",
		&app.ErgebnisDTO{SpieleB: 2})
	if antwort.Fehler != nil || antwort.Folgen == nil {
		t.Fatalf("die Antwort: %+v", antwort)
	}

	folgen := antwort.Folgen
	if folgen.Runde != 1 || len(folgen.SpaetereRunden) != 1 || folgen.SpaetereRunden[0] != 2 ||
		len(folgen.Verschoben) != 1 {
		t.Fatalf("die Folgen: %+v", folgen)
	}

	bert := folgen.Verschoben[0]
	if bert.TeilnehmerID != "3" || bert.Name != "Bert" || bert.PlatzVorher != 2 ||
		bert.PlatzNachher != 1 || bert.PunkteVorher != 3 || bert.PunkteNachher != 6 {
		t.Errorf("Bert: %+v", bert)
	}

	if naht.gefragt == nil || naht.gefragt.SpieleB != 2 {
		t.Errorf("an die Naht ging %+v", naht.gefragt)
	}

	// **Ohne Stand heißt: zurücknehmen** — die Naht bekommt `nil`.
	if zurueck := anwendung.TurnierKorrekturfolgen(t.Context(), "7", "12", nil); zurueck.Fehler != nil ||
		naht.gefragt != nil {
		t.Errorf("zurücknehmen: %+v, gefragt %+v", zurueck, naht.gefragt)
	}
}

func TestDieFolgenWeisenUnbrauchbaresAb(t *testing.T) {
	anwendung, naht := turnieranwendung(t)

	for name, kennungen := range map[string][2]string{
		"turnierID": {"sieben", "12"},
		"paarungID": {"7", "zwoelf"},
	} {
		if antwort := anwendung.TurnierKorrekturfolgen(t.Context(), kennungen[0], kennungen[1], nil); antwort.Fehler == nil ||
			antwort.Fehler.Feld != name {
			t.Errorf("%s: %+v", name, antwort.Fehler)
		}
	}

	naht.fehler = domain.ErrNotFound
	if antwort := anwendung.TurnierKorrekturfolgen(t.Context(), "7", "12", nil); antwort.Fehler == nil ||
		antwort.Fehler.Code != app.CodeNichtGefunden || antwort.Folgen != nil {
		t.Errorf("ein Nahtfehler: %+v", antwort)
	}
}
