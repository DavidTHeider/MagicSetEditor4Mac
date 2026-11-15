package storage_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Schlüsselwörter eines Sets überleben den Bestand** —
// #1062.
func TestSchluesselwoerterKommenMitDemSet(t *testing.T) {
	db := frischeDatenbank(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probe", Game: "probe",
		Keywords: []*domain.Keyword{
			{
				Word: "Flügelschlag", Match: "Flügelschlag", Mode: "core",
				Reminder: "Nur von Fliegern zu blocken.", Rules: "Eigen.",
			},
			{Word: "Beben", Match: "Beben <atom-param>zahl</atom-param>"},
		},
	}, nil)
	if err != nil {
		t.Fatalf("SetAnlegen: %v", err)
	}

	woerter, lesefehler := storage.NeuerSetLeser(db).Schluesselwoerter(probekontext(t), setID)
	if lesefehler != nil {
		t.Fatalf("Schluesselwoerter: %v", lesefehler)
	}

	if len(woerter) != 2 {
		t.Fatalf("%d Schlüsselwörter: %+v", len(woerter), woerter)
	}

	// **Die Reihenfolge der Datei ist die der Liste** — sie entscheidet,
	// welches zuerst auf einen Text passt.
	if woerter[0].Word != "Flügelschlag" || woerter[1].Word != "Beben" {
		t.Errorf("die Reihenfolge ist %q, %q", woerter[0].Word, woerter[1].Word)
	}

	if woerter[0].Reminder != "Nur von Fliegern zu blocken." ||
		woerter[0].Rules != "Eigen." || woerter[0].Mode != "core" {
		t.Errorf("das erste ist %+v", woerter[0])
	}

	// **Und jedes trägt eine Kennung** — ohne sie ließe sich keines ändern.
	if woerter[0].ID == 0 || woerter[0].ID == woerter[1].ID {
		t.Errorf("die Kennungen sind %d und %d", woerter[0].ID, woerter[1].ID)
	}
}

// **Ein Set ohne Schlüsselwörter gibt eine leere Liste.**
func TestEinSetOhneSchluesselwoerterImBestand(t *testing.T) {
	db := frischeDatenbank(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Probe", Game: "probe"}, nil)
	if err != nil {
		t.Fatalf("SetAnlegen: %v", err)
	}

	woerter, lesefehler := storage.NeuerSetLeser(db).Schluesselwoerter(probekontext(t), setID)
	if lesefehler != nil || len(woerter) != 0 {
		t.Errorf("%+v / %v", woerter, lesefehler)
	}
}
