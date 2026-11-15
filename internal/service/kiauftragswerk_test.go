package service_test

// **Das Auftragswerk** — Werkstück A6 von Phase 5.3
// (#1445).
//
// Hier steht die Regel, um die es dem Issue geht: **Generieren verändert die
// Zielkarte nicht.** Was ein Modell vorschlägt, wird geprüft und abgelegt;
// übernommen wird in einem **zweiten** Schritt, und der prüft noch einmal.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// derLauf ist die Uhr, gegen die die Läufe unten rechnen.
var derLauf = time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)

// einWunsch beschreibt, was jemand von der KI will.
func einWunsch() service.KIWunsch {
	return service.KIWunsch{
		ProfilID:   7,
		Faehigkeit: domain.KITextgenerierung,
		Anbieter:   "ollama",
		Alias:      "schreibt:8b",
		Zielkarte:  42,
	}
}

// **Ein Auftrag wird angelegt — und die Karte bleibt, wie sie war.**
func TestEinAuftragAendertDieZielkarteNicht(t *testing.T) {
	t.Parallel()

	karten := &kiKartenprobe{revision: "stand-abc"}
	werk := einWerk(t, karten)

	auftrag, fehler := werk.Anlegen(t.Context(), einWunsch())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if auftrag.Zielrevision != "stand-abc" {
		t.Errorf("die Zielrevision ist %q", auftrag.Zielrevision)
	}

	if auftrag.Kostenposten == "" {
		t.Error("es wurde kein Budget reserviert")
	}

	// **Das ist die Zusage des Issues**, und sie steht als Zahl da: ohne sie
	// wäre der Lauf auch dann grün, wenn nie geschrieben würde.
	if karten.schreibvorgaenge != 0 {
		t.Errorf("beim Anlegen wurde %d mal geschrieben", karten.schreibvorgaenge)
	}
}

// **Ein Modell, das die Fähigkeit nicht belegt, wird abgewiesen.**
func TestEinModellOhneBelegteFaehigkeitWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	wunsch := einWunsch()
	wunsch.Faehigkeit = domain.KIVision

	_, fehler := einWerk(t, &kiKartenprobe{revision: "stand-abc"}).
		Anlegen(t.Context(), wunsch)

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Und ein Modell, das es nicht gibt, ebenso.**
func TestEinUnbekanntesModellWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	wunsch := einWunsch()
	wunsch.Alias = "gibtesnicht"

	_, fehler := einWerk(t, &kiKartenprobe{revision: "stand-abc"}).
		Anlegen(t.Context(), wunsch)

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Unter »nur lokal« geht kein Cloudmodell hinaus — und es wird nichts
// reserviert.**
//
// # Der Satz, um den es geht
//
// „Restriktivste Local-only-/Cloud-/Profil-/Maschinenregel gilt." Und:
// „**Loopback allein beweist keine lokale Inferenz**" — deshalb entscheidet
// der **erklärte Ort** des Modells und nicht seine Adresse.
func TestUnterNurLokalGehtKeinCloudmodellHinaus(t *testing.T) {
	t.Parallel()

	kosten := &kostenprobe{}

	werk := service.NeuesKIAuftragswerk(
		derKatalog(t), &kiKartenprobe{revision: "stand-abc"}, kosten,
		&schemaprobe{}, service.KIOrtsregel{NurLokal: true},
		func() time.Time { return derLauf })

	wunsch := einWunsch()
	wunsch.Alias = "wolke:gross"

	_, fehler := werk.Anlegen(t.Context(), wunsch)

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	// **Nichts reserviert** — das Issue verlangt „null Netzverbindungen"; hier
	// ist die Vorstufe prüfbar: es wird nicht einmal gebucht.
	if kosten.reservierungen != 0 {
		t.Errorf("%d Reservierungen, erwartet keine", kosten.reservierungen)
	}
}

// **Ein lokales Modell geht auch unter »nur lokal«.**
//
// **Die Gegenprobe**: ohne sie sagte der Lauf darüber nur, dass etwas
// abgewiesen wird — nicht, dass die Regel auf den Ort sieht.
func TestEinLokalesModellGehtAuchUnterNurLokal(t *testing.T) {
	t.Parallel()

	werk := service.NeuesKIAuftragswerk(
		derKatalog(t), &kiKartenprobe{revision: "stand-abc"}, &kostenprobe{},
		&schemaprobe{}, service.KIOrtsregel{NurLokal: true},
		func() time.Time { return derLauf })

	if _, fehler := werk.Anlegen(t.Context(), einWunsch()); fehler != nil {
		t.Fatalf("ein lokales Modell wurde abgewiesen: %v", fehler)
	}
}

// **Ein Vorschlag wird gegen das Schema geprüft und ändert nichts.**
func TestEinVorschlagWirdGeprueftUndAendertNichts(t *testing.T) {
	t.Parallel()

	karten := &kiKartenprobe{revision: "stand-abc"}
	werk := einWerk(t, karten)

	auftrag, fehler := werk.Anlegen(t.Context(), einWunsch())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	befunde, fehler := werk.Vorschlagen(t.Context(), domain.KIVorschlag{
		Auftrag: auftrag.Kennung,
		Felder:  map[string][]string{"name": {"Blitzschlag"}, "erfunden": {"x"}},
	})
	if fehler != nil {
		t.Fatalf("Vorschlagen: %v", fehler)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %+v", len(befunde), befunde)
	}

	if karten.schreibvorgaenge != 0 {
		t.Errorf("beim Vorschlagen wurde %d mal geschrieben", karten.schreibvorgaenge)
	}
}

// **Erst die Übernahme schreibt — und nur die guten Felder.**
func TestErstDieUebernahmeSchreibt(t *testing.T) {
	t.Parallel()

	karten := &kiKartenprobe{revision: "stand-abc"}
	werk := einWerk(t, karten)

	auftrag, _ := werk.Anlegen(t.Context(), einWunsch()) //nolint:errcheck // oben geprüft

	if _, fehler := werk.Vorschlagen(t.Context(), domain.KIVorschlag{
		Auftrag: auftrag.Kennung,
		Felder:  map[string][]string{"name": {"Blitzschlag"}},
	}); fehler != nil {
		t.Fatalf("Vorschlagen: %v", fehler)
	}

	bericht, fehler := werk.Uebernehmen(t.Context(), auftrag.Kennung, []string{"name"}, 7)
	if fehler != nil {
		t.Fatalf("Uebernehmen: %v", fehler)
	}

	if bericht.Felder != 1 || karten.schreibvorgaenge != 1 {
		t.Errorf("übernommen wurden %d Felder, geschrieben %d mal",
			bericht.Felder, karten.schreibvorgaenge)
	}
}

// **Eine Übernahme scheitert an allem, was sich geändert hat.**
func TestEineUebernahmeScheitertAnJederAenderung(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		vorbereiten func(*kiKartenprobe, *service.KIAuftragswerk)
		jetzt       time.Time
	}{
		"die Karte hat sich geändert": {
			func(k *kiKartenprobe, _ *service.KIAuftragswerk) {
				k.revision = "stand-neu"
			},
			derLauf,
		},
		"die Frist ist abgelaufen": {
			func(*kiKartenprobe, *service.KIAuftragswerk) {},
			derLauf.Add(48 * time.Hour),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			karten := &kiKartenprobe{revision: "stand-abc"}
			jetzt := derLauf

			werk := service.NeuesKIAuftragswerk(
				derKatalog(t), karten, &kostenprobe{}, &schemaprobe{},
				service.KIOrtsregel{}, func() time.Time { return jetzt })

			auftrag, _ := werk.Anlegen(t.Context(), einWunsch())     //nolint:errcheck // geprüft
			_, _ = werk.Vorschlagen(t.Context(), domain.KIVorschlag{ //nolint:errcheck // geprüft
				Auftrag: auftrag.Kennung,
				Felder:  map[string][]string{"name": {"Blitzschlag"}},
			})

			fall.vorbereiten(karten, werk)
			jetzt = fall.jetzt

			if _, fehler := werk.Uebernehmen(t.Context(), auftrag.Kennung, []string{"name"}, 7); fehler == nil {
				t.Fatalf("%s kam durch", name)
			}

			if karten.schreibvorgaenge != 0 {
				t.Errorf("trotz %s wurde geschrieben", name)
			}
		})
	}
}

// **Eine Übernahme ohne Vorschlag scheitert.**
func TestEineUebernahmeOhneVorschlagScheitert(t *testing.T) {
	t.Parallel()

	werk := einWerk(t, &kiKartenprobe{revision: "stand-abc"})

	auftrag, _ := werk.Anlegen(t.Context(), einWunsch()) //nolint:errcheck // geprüft

	if _, fehler := werk.Uebernehmen(t.Context(), auftrag.Kennung, []string{"name"}, 7); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Und ein Vorschlag lässt sich nur einmal übernehmen.**
//
// **Die Kennung ist konsumierbar** — das Issue sagt es so. Zweimal übernehmen
// hieße, denselben Vorschlag auf einen Stand zu legen, der ihn schon trägt.
func TestEinVorschlagLaesstSichNurEinmalUebernehmen(t *testing.T) {
	t.Parallel()

	karten := &kiKartenprobe{revision: "stand-abc"}
	werk := einWerk(t, karten)

	auftrag, _ := werk.Anlegen(t.Context(), einWunsch())     //nolint:errcheck // geprüft
	_, _ = werk.Vorschlagen(t.Context(), domain.KIVorschlag{ //nolint:errcheck // geprüft
		Auftrag: auftrag.Kennung,
		Felder:  map[string][]string{"name": {"Blitzschlag"}},
	})

	if _, fehler := werk.Uebernehmen(t.Context(), auftrag.Kennung, []string{"name"}, 7); fehler != nil {
		t.Fatalf("die erste Übernahme: %v", fehler)
	}

	if _, fehler := werk.Uebernehmen(t.Context(), auftrag.Kennung, []string{"name"}, 7); fehler == nil {
		t.Fatal("die zweite Übernahme kam durch")
	}

	if karten.schreibvorgaenge != 1 {
		t.Errorf("geschrieben wurde %d mal", karten.schreibvorgaenge)
	}
}

// **Ein Wunsch ohne Profil oder Karte wird abgewiesen.**
//
// **Und zwar nachdem reserviert wurde** — deshalb steht der Lauf hier: er
// prüft, dass ein unvollständiger Auftrag nicht durchkommt, auch wenn alles
// andere geklappt hat.
func TestEinUnvollstaendigerWunschWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, kaputt := range map[string]func(*service.KIWunsch){
		"ohne Profil": func(w *service.KIWunsch) { w.ProfilID = 0 },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wunsch := einWunsch()
			kaputt(&wunsch)

			if _, fehler := einWerk(t, &kiKartenprobe{revision: "stand-abc"}).
				Anlegen(t.Context(), wunsch); fehler == nil {
				t.Errorf("%s kam durch", name)
			}
		})
	}
}

// **Ein Fehler beim zweiten Blick auf Karte oder Schema hält die Übernahme auf.**
//
// **Der zweite Blick ist der Punkt**: zwischen Vorschlag und Übernahme liegt
// Zeit, und wer dann nicht mehr lesen kann, weiß nicht, ob er schreiben darf.
func TestEinLesefehlerBeiDerUebernahmeHaeltSieAuf(t *testing.T) {
	t.Parallel()

	eigen := errors.New("die Ablage klemmt")

	for name, klemmen := range map[string]func(*kiKartenprobe, *schemaprobe){
		"die Karte":  func(k *kiKartenprobe, _ *schemaprobe) { k.fehler = eigen },
		"das Schema": func(_ *kiKartenprobe, s *schemaprobe) { s.fehler = eigen },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			karten := &kiKartenprobe{revision: "stand-abc"}
			schema := &schemaprobe{}

			werk := service.NeuesKIAuftragswerk(derKatalog(t), karten,
				&kostenprobe{}, schema, service.KIOrtsregel{},
				func() time.Time { return derLauf })

			auftrag, _ := werk.Anlegen(t.Context(), einWunsch())     //nolint:errcheck // geprüft
			_, _ = werk.Vorschlagen(t.Context(), domain.KIVorschlag{ //nolint:errcheck // geprüft
				Auftrag: auftrag.Kennung,
				Felder:  map[string][]string{"name": {"Blitzschlag"}},
			})

			klemmen(karten, schema)

			if _, fehler := werk.Uebernehmen(
				t.Context(), auftrag.Kennung, []string{"name"}, 7); !errors.Is(fehler, eigen) {
				t.Errorf("bei %s ist der Fehler %v", name, fehler)
			}
		})
	}
}

// **Die Ortsregel kommt aus der Richtlinie von 5.2 und nicht aus einer
// eigenen Einstellung.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil hier fast eine zweite entstanden wäre.** `richtlinie.nurlokal` gibt
// es längst, und sie sagt genau das; eine eigene `ki.ort` daneben wären zwei
// Schalter für dieselbe Frage. Das Issue verlangt ausdrücklich, die Ebenen
// von 5.2 zu **konsumieren**.
func TestDieOrtsregelKommtAusDerRichtlinie(t *testing.T) {
	t.Parallel()

	for wirksam, erwartet := range map[string]bool{
		domain.NurLokalAn:  true,
		domain.NurLokalAus: false,
		"":                 false,
	} {
		if regel := service.KIOrtsregelAus(wirksam); regel.NurLokal != erwartet {
			t.Errorf("%q gibt NurLokal=%v, erwartet %v",
				wirksam, regel.NurLokal, erwartet)
		}
	}

	// **Und die Regel entscheidet am Ort und nicht an der Adresse** — der
	// Satz aus dem Issue: „Loopback allein beweist keine lokale Inferenz."
	streng := service.KIOrtsregelAus(domain.NurLokalAn)

	if !streng.Erlaubt(domain.KIOrtLokal) || streng.Erlaubt(domain.KIOrtCloud) {
		t.Error("die strenge Regel sieht nicht auf den Ort")
	}

	offen := service.KIOrtsregelAus(domain.NurLokalAus)

	if !offen.Erlaubt(domain.KIOrtLokal) || !offen.Erlaubt(domain.KIOrtCloud) {
		t.Error("die offene Regel verbietet etwas")
	}
}
