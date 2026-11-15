package app

// **An einem fremden Tisch Platz nehmen, ohne die Partie zu kennen** —
// UX-08.AK-04 (#1680).
//
// # Was hier gemessen wird
//
// **Dass ein Gast nur sagt, bei wem er mitspielt.** Die Partie nennt der
// Wirt mit seinem ersten Wort; der Lauf über beide Enden belegt, dass danach
// derselbe Tisch dasteht wie mit genannter Partie — ein Zug wirkt beim Wirt.
//
// **Und dass eine alte Schleife den Tisch nicht mehr anfasst**, sobald jemand
// aufgestanden ist oder sich woanders hingesetzt hat.

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Gast ohne Partie sitzt am offenen Tisch des Wirts — und sein Zug
// wirkt.**
func TestEinGastOhnePartieSitztAmOffenenTischDesWirts(t *testing.T) {
	t.Parallel()

	wirt, gast := zweiAnwendungen(t)

	MitPartienleitung(wirt.anwendung, wirt.anwendung.db)

	partieID, objekt := einTischBeimWirt(t, wirt)

	if antwort := wirt.anwendung.PartieFreigeben(
		wirt.ctx, partieID, "anna", []string{"bert"},
	); antwort.Fehler != nil {
		t.Fatalf("PartieFreigeben: %+v", antwort.Fehler)
	}

	angenommen := make(chan error, 1)

	go func() {
		griff, fehler := wirt.anwendung.sitzungsverwaltung().Annehmen(
			wirt.ctx, wirt.leitung)
		if fehler == nil {
			wirt.anwendung.partiesitzungAufnehmen(wirt.ctx, griff)
		}

		angenommen <- fehler
	}()

	antwort := gast.anwendung.PartieBeitreten(
		gast.ctx, string(wirt.kennung), "", PartiemitbringDTO{})

	if fehler := <-angenommen; fehler != nil {
		t.Fatalf("Annehmen: %v", fehler)
	}

	if antwort.Fehler != nil || !antwort.Laeuft {
		t.Fatalf("PartieBeitreten: %+v", antwort)
	}

	warteAufGenanntePartie(t, gast, partieID)

	if antwort := gast.anwendung.PartieWuenschen(gast.ctx, domain.Spielbefehl{
		Kennung: "b1", Art: "objekt.tappen",
		Nutzlast: map[string]any{"objekt": objekt},
	}); antwort.Fehler != nil {
		t.Fatalf("PartieWuenschen: %+v", antwort.Fehler)
	}

	if !warteAufGetapptesObjekt(t, wirt, partieID, objekt) {
		t.Fatal("der Zug des Gastes hat beim Wirt nichts bewirkt")
	}
}

// warteAufGenanntePartie wartet, bis der Gast die Partie des Wirts kennt.
//
// **Auf die Wirkung und nicht auf den Ruf** (§3): genannt wird sie über die
// Leitung, also später als die Antwort auf den Beitritt.
func warteAufGenanntePartie(t *testing.T, gast *seite, partieID string) {
	t.Helper()

	frist := time.After(5 * time.Second)

	for {
		if gast.anwendung.PartieBeitrittStand(gast.ctx).Partie == partieID {
			return
		}

		select {
		case <-frist:
			t.Fatalf("der Gast kennt die Partie nicht: %+v",
				gast.anwendung.PartieBeitrittStand(gast.ctx))
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// **Bevor der Wirt die Partie nennt, geht kein Wunsch hinaus** — als
// Konflikt, wie ein Wunsch ohne Tisch.
func TestEinWunschVorDerPartieWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)
	anwendung.beitrittMerken(eineAngenommene(t, domain.ZweckPartie), "")

	antwort := anwendung.PartieWuenschen(t.Context(), domain.Spielbefehl{
		Kennung: "b1", Art: "objekt.tappen",
	})

	switch {
	case antwort.Fehler == nil:
		t.Fatal("der Wunsch kam durch")
	case antwort.Fehler.Code != "konflikt":
		t.Errorf("Code %q, erwartet konflikt", antwort.Fehler.Code)
	}
}

// **Das erste Wort des Wirts setzt den Gast an seinen Tisch.**
func TestDasErsteWortSetztDenGastAnDenTisch(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)

	griff := eineAngenommene(t, domain.ZweckPartie)
	griff.gibt = [][]byte{eineBegruessung(t, "p9")}
	griff.lesefehler = errors.New("danach ist Schluss")

	_, lauf := anwendung.beitrittMerken(griff, "")

	anwendung.partieErfragen(t.Context(), griff, lauf)

	// **Danach riss die Leitung, und der Tisch ist verloren** — also saß der
	// Gast an ihm: ohne genannte Partie liefe keine Schleife, und der
	// Beitritt stünde noch offen da. **Und die Begrüßung las die Schleife
	// aus dem Vorzimmer**, nicht ein zweites Mal von der Leitung.
	stand := anwendung.PartieBeitrittStand(t.Context())

	switch {
	case stand.Laeuft || !stand.Pausiert:
		t.Errorf("es kam %+v", stand)
	case griff.gelesen != 1:
		t.Errorf("gelesen %dmal — die Begrüßung ging verloren", griff.gelesen)
	}
}

// **Wer aufgestanden ist, bekommt keine Partie nachgetragen.**
func TestNachDemAufstehenKommtKeinePartieMehr(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)

	griff := eineAngenommene(t, domain.ZweckPartie)
	griff.gibt = [][]byte{eineBegruessung(t, "p9")}

	_, lauf := anwendung.beitrittMerken(griff, "")

	anwendung.PartieBeitrittBeenden(t.Context())
	anwendung.partieErfragen(t.Context(), griff, lauf)

	if stand := anwendung.PartieBeitrittStand(t.Context()); stand.Laeuft ||
		stand.Partie != "" || stand.Pausiert {
		t.Errorf("es kam %+v", stand)
	}
}

// **Wer inzwischen an einem anderen Tisch sitzt, behält dessen Partie.**
//
// **Der Fall, den das Aufstehen allein nicht zeigt**: dort ist der Griff
// schon weg, und jede Prüfung ließe die Partie fallen. Hier steht ein neuer
// Griff da — und nur der Lauf sagt, dass die Begrüßung zum alten gehört.
func TestEinNeuerTischBehaeltSeinePartie(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)

	alt := eineAngenommene(t, domain.ZweckPartie)
	alt.gibt = [][]byte{eineBegruessung(t, "p9")}
	alt.lesefehler = errors.New("danach ist Schluss")

	_, lauf := anwendung.beitrittMerken(alt, "")
	anwendung.beitrittMerken(eineAngenommene(t, domain.ZweckPartie), "p2")

	anwendung.partieErfragen(t.Context(), alt, lauf)

	if stand := anwendung.PartieBeitrittStand(t.Context()); !stand.Laeuft ||
		stand.Partie != "p2" {
		t.Errorf("es kam %+v", stand)
	}
}

// **Reißt die Leitung vor dem ersten Wort, ist der Platz pausiert.**
func TestOhneErstesWortIstDerPlatzPausiert(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)

	griff := eineAngenommene(t, domain.ZweckPartie)
	griff.lesefehler = errors.New("die Leitung ist weg")

	_, lauf := anwendung.beitrittMerken(griff, "")

	anwendung.partieErfragen(t.Context(), griff, lauf)

	if stand := anwendung.PartieBeitrittStand(t.Context()); stand.Laeuft ||
		!stand.Pausiert {
		t.Errorf("es kam %+v", stand)
	}
}

// **Eine alte Schleife räumt einen Tisch nicht ab, den es nicht mehr gibt.**
//
// **Beides stand vorher unbedingt da**: wer aufgestanden war, hieß beim
// nächsten Nachsehen „pausiert", sobald die alte Schleife auslief — und wer
// sich inzwischen an einen anderen Tisch gesetzt hatte, verlor ihn.
func TestEinVerlorenerTischGehoertDerSchleifeDieIhnHielt(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)

	_, alt := anwendung.beitrittMerken(
		eineAngenommene(t, domain.ZweckPartie), "p1")
	anwendung.beitrittMerken(eineAngenommene(t, domain.ZweckPartie), "p2")

	anwendung.tischVerloren(alt, nil)

	if stand := anwendung.PartieBeitrittStand(t.Context()); !stand.Laeuft ||
		stand.Partie != "p2" {
		t.Fatalf("der neue Tisch ist weg: %+v", stand)
	}

	_, lauf := anwendung.beitrittMerken(
		eineAngenommene(t, domain.ZweckPartie), "p3")

	anwendung.tischVerloren(lauf, nil)

	if stand := anwendung.PartieBeitrittStand(t.Context()); stand.Laeuft ||
		!stand.Pausiert {
		t.Errorf("es kam %+v", stand)
	}
}

// **Kann der Wirt nicht begrüßen, legt er die Sitzung nieder — und liest
// nichts.**
func TestEineMisslungeneBegruessungLegtDieSitzungNieder(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)
	anwendung.PartieFreigeben(t.Context(), "p1", "anna", []string{"bert"})

	sitzung := eineAngenommene(t, domain.ZweckPartie)
	sitzung.leerfehler = errors.New("die Leitung ist weg")
	sitzung.gibt = [][]byte{eineBegruessung(t, "p1")}

	anwendung.partieBedienen(t.Context(), sitzung, "p1", "bert",
		einePartiesitzung(sitzung), einGastplatz(sitzung))

	switch {
	case !sitzung.istZu():
		t.Error("die Sitzung blieb offen")
	case sitzung.gelesen != 0:
		t.Errorf("gelesen %dmal — die Schleife lief trotzdem", sitzung.gelesen)
	}
}

// **Und kann er das Tischbild gar nicht bauen** — eine Partie, die es beim
// Wirt nicht gibt —, gilt dasselbe. Bis zum 2026-09-25 betrat diesen Zweig
// nur ein Lauf, dessen Zeitverhalten ihn zufällig hergab: die Pipeline über
// UX-08.AK-04 fand `begruessen` bei 66,7 %.
func TestEinTischbildOhnePartieLegtDieSitzungNieder(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t).MitPartienaht(&unbekannteLeitung{})
	anwendung.PartieFreigeben(t.Context(), "p1", "anna", []string{"bert"})

	sitzung := eineAngenommene(t, domain.ZweckPartie)
	sitzung.gibt = [][]byte{eineBegruessung(t, "p1")}

	anwendung.partieBedienen(t.Context(), sitzung, "p1", "bert",
		einePartiesitzung(sitzung), einGastplatz(sitzung))

	switch {
	case !sitzung.istZu():
		t.Error("die Sitzung blieb offen")
	case sitzung.gelesen != 0:
		t.Errorf("gelesen %dmal — die Schleife lief trotzdem", sitzung.gelesen)
	}
}

// unbekannteLeitung kennt keine Partie — ihr Tischbild lässt sich nicht
// bauen, und die Begrüßung scheitert, bevor etwas in die Warteschlange geht.
type unbekannteLeitung struct{ schweigsameLeitung }

func (u *unbekannteLeitung) Partie(
	context.Context, string, string,
) (domain.Partiesicht, error) {
	return domain.Partiesicht{}, domain.ErrNotFound
}

func (u *unbekannteLeitung) Tischbild(
	context.Context, string, string,
) (domain.Tischbild, error) {
	return domain.Tischbild{}, domain.ErrNotFound
}

// eineBegruessung ist das Tischbild, mit dem ein Wirt begrüßt.
func eineBegruessung(t *testing.T, partie string) []byte {
	t.Helper()

	inhalt, fehler := domain.VerpackeKontrolle(domain.KontrolleTischbild,
		domain.Tischbild{Sicht: domain.Partiesicht{PartieID: partie, Fuer: "bert"}})
	if fehler != nil {
		t.Fatalf("VerpackeKontrolle: %v", fehler)
	}

	roh, fehler := json.Marshal(domain.Partienachricht{
		Art: domain.ArtKontrolle, Partie: partie, Epoche: 1, Folge: 1,
		Von: "wirt", Inhalt: inhalt,
	})
	if fehler != nil {
		t.Fatalf("Marshal: %v", fehler)
	}

	return roh
}
