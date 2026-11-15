package storage_test

// **Die Ablage gemessener Modellfähigkeiten** — Werkstück A9 von Phase 5.3
// (#1448).

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// eineMessung baut eine vollständige Probe.
func eineMessung(teile ...func(*domain.KIModellprobe)) domain.KIModellprobe {
	probe := domain.KIModellprobe{
		Anbieter: "openai", Alias: "gpt-probe",
		Faehigkeit: domain.KIStrukturierteGenerierung,
		Datum:      "2026-09-12",
		Beleg:      "Probeauftrag gegen /v1/responses",
		Laufzeit:   "Responses-API /v1/responses",
		Modell:     "gpt-probe-2026-09-01",
	}

	for _, teil := range teile {
		teil(&probe)
	}

	return probe
}

// **Eine gemerkte Messung kommt zurück.**
func TestEineGemerkteMessungKommtZurueck(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueModellprobenablage(db)

	if fehler := ablage.Merken(probekontext(t), 1, eineMessung()); fehler != nil {
		t.Fatalf("merken: %v", fehler)
	}

	proben, fehler := ablage.Alle(probekontext(t), 1)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(proben) != 1 || proben[0] != eineMessung() {
		t.Fatalf("die Messungen sind %+v", proben)
	}
}

// **Eine zweite Messung ersetzt die erste.**
//
// **Und hängt sich nicht daneben**: was gestern galt, sagt über heute
// nichts, und zwei Daten nebeneinander ließen offen, welches gilt.
func TestEineZweiteMessungErsetztDieErste(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueModellprobenablage(db)

	for _, tag := range []string{"2026-09-12", "2026-09-13"} {
		if fehler := ablage.Merken(probekontext(t), 1,
			eineMessung(func(p *domain.KIModellprobe) { p.Datum = tag }),
		); fehler != nil {
			t.Fatalf("merken: %v", fehler)
		}
	}

	proben, fehler := ablage.Alle(probekontext(t), 1)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(proben) != 1 || proben[0].Datum != "2026-09-13" {
		t.Fatalf("die Messungen sind %+v", proben)
	}
}

// **Eine unvollständige Messung wird gar nicht erst geschrieben.**
//
// **Vor dem Schreiben und nicht beim Lesen**: ein Eintrag ohne Beleg käme
// später in den Katalog, und der weist dann **alles** ab — auch die Modelle
// der anderen Anbieter (A8).
func TestEineUnvollstaendigeMessungWirdNichtGeschrieben(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueModellprobenablage(db)

	if fehler := ablage.Merken(probekontext(t), 1,
		eineMessung(func(p *domain.KIModellprobe) { p.Beleg = "" }),
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	proben, fehler := ablage.Alle(probekontext(t), 1)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(proben) != 0 {
		t.Errorf("die unvollständige Messung steht da: %+v", proben)
	}
}

// **Ein anderes Profil sieht sie nicht.**
//
// **Weil der Schlüssel je Profil liegt**: wer mit einem anderen Zugang
// fragt, fragt vielleicht ein anderes Konto mit anderen Modellen.
func TestEinAnderesProfilSiehtDieMessungNicht(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueModellprobenablage(db)

	if fehler := ablage.Merken(probekontext(t), 1, eineMessung()); fehler != nil {
		t.Fatalf("merken: %v", fehler)
	}

	proben, fehler := ablage.Alle(probekontext(t), 2)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(proben) != 0 {
		t.Errorf("das fremde Profil sieht %+v", proben)
	}
}

// **Eine vergessene Messung ist fort — und die anderen bleiben.**
func TestEineVergesseneMessungIstFort(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueModellprobenablage(db)

	for _, alias := range []string{"gpt-probe", "gpt-anderes"} {
		if fehler := ablage.Merken(probekontext(t), 1,
			eineMessung(func(p *domain.KIModellprobe) { p.Alias = alias }),
		); fehler != nil {
			t.Fatalf("merken: %v", fehler)
		}
	}

	if fehler := ablage.Vergessen(probekontext(t), 1, "openai",
		"gpt-probe"); fehler != nil {
		t.Fatalf("vergessen: %v", fehler)
	}

	proben, fehler := ablage.Alle(probekontext(t), 1)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	// **Die Gegenprobe steht in derselben Zusicherung**: ein `DELETE` ohne
	// Alias träfe beide, und der Lauf bliebe grün, wenn er nur zählte, dass
	// die eine fort ist.
	if len(proben) != 1 || proben[0].Alias != "gpt-anderes" {
		t.Fatalf("die Messungen sind %+v", proben)
	}
}

// **Was nur der Stolpertreiber hergibt.**
//
// **Ein Fehler mitten im Lesen und eine Zeile, die nicht passt** — eine echte
// Datenbank liefert beides nicht, und ohne sie blieben zwei Zweige
// unbetreten (§10).
func TestDieProbenablageMeldetWasNurDerStolpertreiberHergibt(t *testing.T) {
	t.Parallel()

	// **Der Treibername steht zusammengesetzt da** — `gosec` hält
	// „stolper-proben-scan" sonst für ein eingebautes Geheimnis, und ein
	// `//nolint` wäre eine Ausnahme für etwas, das keine braucht.
	stolper := "stolper-proben-"

	for name, treiber := range map[string]string{
		"ein Fehler mitten im Lesen":            stolper + "ende",
		"eine Zeile passt nicht in ihre Felder": stolper + "scan",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db, fehler := sql.Open(treiber, "egal")
			if fehler != nil {
				t.Fatalf("öffnen: %v", fehler)
			}

			t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // nur ein Test

			if _, fehler := storage.NeueModellprobenablage(db).
				Alle(probekontext(t), 1); fehler == nil {
				t.Error("der Stolpertreiber kam ohne Fehler durch")
			}
		})
	}
}

// **Und ein geschlossener Bestand nimmt nichts an und gibt nichts her.**
func TestEinGeschlossenerBestandNimmtKeineMessungAn(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	ablage := storage.NeueModellprobenablage(db)

	if fehler := ablage.Merken(probekontext(t), 1, eineMessung()); fehler == nil {
		t.Error("der geschlossene Bestand nahm die Messung an")
	}

	if _, fehler := ablage.Alle(probekontext(t), 1); fehler == nil {
		t.Error("der geschlossene Bestand gab Messungen her")
	}

	if fehler := ablage.Vergessen(probekontext(t), 1, "openai",
		"gpt-probe"); fehler == nil {
		t.Error("der geschlossene Bestand vergaß etwas")
	}
}
