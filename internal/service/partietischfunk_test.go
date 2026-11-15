package service_test

// **Der Tisch bekommt Funk** —
// #1627.
//
// # Wogegen hier geprüft wird
//
// **Gegen einen Tisch, der aus Zwiegesprächen bestand.** Ein Zug ging bis
// zum 2026-09-19 nur an den Sitz, der ihn gestellt hat: `Partiewirt` schickt
// über **seine** Sitzung, und ein Wirt gehört zu **einem** Gastsitz. Zog der
// Wirt selbst oder ein anderer Gast, ging **nichts** hinaus.
//
// **Und gegen eine zweite Lücke daneben**: was hinausging, war das **rohe**
// Ereignis — samt der privaten Ergänzungen **aller** Sitze. Die Projektion
// [domain.Ereignis.Fuer] gab es, `internal/app` benutzte sie, der Weg über
// die Leitung nicht. Mit dem Funk wäre daraus ein systematischer Verrat
// geworden: jeder Sitz bekäme, was jeder andere gezogen hat.

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einZugMitGeheimnis ist ein Ereignis, das für **beide** Sitze etwas
// Privates trägt — anders ließe sich die Projektion nicht prüfen.
func einZugMitGeheimnis() []domain.Ereignis {
	return []domain.Ereignis{{
		Folge: 7, Art: "zone.blick",
		Oeffentlich: map[string]any{"sitz": "anna"},
		Privat: []domain.Ergaenzung{
			{Empfaenger: "anna", Inhalt: map[string]any{"karten": "annas"}},
			{Empfaenger: "bert", Inhalt: map[string]any{"karten": "berts"}},
		},
	}}
}

// gelesen packt aus, was auf einer Leitung ankam.
func gelesen(t *testing.T, nachricht domain.Partienachricht) domain.Ereignis {
	t.Helper()

	var ereignis domain.Ereignis
	if fehler := json.Unmarshal(nachricht.Inhalt, &ereignis); fehler != nil {
		t.Fatalf("das Ereignis ließ sich nicht auspacken: %v", fehler)
	}

	return ereignis
}

// TestEinZugErreichtDenGanzenTisch ist die Zusage aus #1627.
func TestEinZugErreichtDenGanzenTisch(t *testing.T) {
	t.Parallel()

	runde := &service.Partierunde{}

	anna, annasLeitung := einPlatz("anna")
	bert, bertsLeitung := einPlatz("bert")
	runde.Setze(anna)
	runde.Setze(bert)

	nehmer := &nehmerdoppel{gibt: einZugMitGeheimnis()}
	funk := service.NeuerTischfunk("p1", nehmer, runde)

	if fehler := funk.Verkuende(context.Background(), "anna",
		service.Befehlsausgang{Fassung: 2, Ereignisse: nehmer.gibt},
	); fehler != nil {
		t.Fatalf("Verkuende: %v", fehler)
	}

	// **Bert bekommt den Zug, den er nicht gestellt hat** — Ereignis und
	// Tischbild, in dieser Reihenfolge.
	beiBert := bertsLeitung.alle()
	if len(beiBert) != 2 {
		t.Fatalf("bei Bert kamen %d Nachrichten an", len(beiBert))
	}

	if beiBert[0].Art != domain.ArtEreignis ||
		beiBert[1].Art != domain.ArtKontrolle {
		t.Errorf("angekommen sind %v und %v", beiBert[0].Art, beiBert[1].Art)
	}

	// **Und der Sitz, der gestellt hat, bekommt ihn hier nicht** — er hat
	// ihn über seinen eigenen Wirt schon.
	if len(annasLeitung.alle()) != 0 {
		t.Errorf("bei Anna kamen %d Nachrichten an", len(annasLeitung.alle()))
	}
}

// TestJederBekommtNurSeinGeheimnis hält die zweite Lücke.
func TestJederBekommtNurSeinGeheimnis(t *testing.T) {
	t.Parallel()

	runde := &service.Partierunde{}

	bert, bertsLeitung := einPlatz("bert")
	runde.Setze(bert)

	nehmer := &nehmerdoppel{gibt: einZugMitGeheimnis()}
	funk := service.NeuerTischfunk("p1", nehmer, runde)

	if fehler := funk.Verkuende(context.Background(), "anna",
		service.Befehlsausgang{Ereignisse: nehmer.gibt},
	); fehler != nil {
		t.Fatalf("Verkuende: %v", fehler)
	}

	ereignis := gelesen(t, bertsLeitung.alle()[0])

	if len(ereignis.Privat) != 1 {
		t.Fatalf("bei Bert stehen %d Ergänzungen", len(ereignis.Privat))
	}

	if ereignis.Privat[0].Empfaenger != "bert" {
		t.Errorf("die Ergänzung gehört %q", ereignis.Privat[0].Empfaenger)
	}
}

// TestDasFensterDiesesRechnersHoertMit — der Wirt spielt selbst mit.
func TestDasFensterDiesesRechnersHoertMit(t *testing.T) {
	t.Parallel()

	runde := &service.Partierunde{}

	bert, _ := einPlatz("bert")
	runde.Setze(bert)

	gehoert := &mitschrift{}
	nehmer := &nehmerdoppel{gibt: einZugMitGeheimnis()}

	funk := service.NeuerTischfunk("p1", nehmer, runde).
		MitFenster("anna", gehoert)

	if fehler := funk.Verkuende(context.Background(), "bert",
		service.Befehlsausgang{Ereignisse: nehmer.gibt},
	); fehler != nil {
		t.Fatalf("Verkuende: %v", fehler)
	}

	if len(gehoert.ereignisse) != 1 || len(gehoert.bilder) != 1 {
		t.Fatalf("am Fenster kamen %d Ereignisse und %d Bilder an",
			len(gehoert.ereignisse), len(gehoert.bilder))
	}

	if len(gehoert.ereignisse[0].Privat) != 1 ||
		gehoert.ereignisse[0].Privat[0].Empfaenger != "anna" {
		t.Errorf("am Fenster steht %+v", gehoert.ereignisse[0].Privat)
	}
}

// TestDerEigeneZugKommtNichtZweimalAn — wer stellt, bekommt ihn über den
// Weg, über den er gestellt hat.
func TestDerEigeneZugKommtNichtZweimalAn(t *testing.T) {
	t.Parallel()

	runde := &service.Partierunde{}

	gehoert := &mitschrift{}
	nehmer := &nehmerdoppel{gibt: einZugMitGeheimnis()}

	funk := service.NeuerTischfunk("p1", nehmer, runde).
		MitFenster("anna", gehoert)

	if fehler := funk.Verkuende(context.Background(), "anna",
		service.Befehlsausgang{Ereignisse: nehmer.gibt},
	); fehler != nil {
		t.Fatalf("Verkuende: %v", fehler)
	}

	if len(gehoert.ereignisse) != 0 || len(gehoert.bilder) != 0 {
		t.Errorf("am Fenster kamen %d Ereignisse und %d Bilder an",
			len(gehoert.ereignisse), len(gehoert.bilder))
	}
}

// klagendesFenster ist ein Fenster, das an einer benannten Stelle klagt.
//
// **Zwei Schalter und nicht einer**: klagt es schon beim Ereignis, bliebe
// der Weg zum Bild ungemessen (§10).
type klagendesFenster struct {
	mitschrift

	ereignisklage error
	bildklage     error
}

func (f *klagendesFenster) Ereignis(
	ctx context.Context, ereignis domain.Ereignis,
) error {
	if f.ereignisklage != nil {
		return f.ereignisklage
	}

	return f.mitschrift.Ereignis(ctx, ereignis)
}

func (f *klagendesFenster) Tischbild(
	ctx context.Context, bild domain.Tischbild,
) error {
	if f.bildklage != nil {
		return f.bildklage
	}

	return f.mitschrift.Tischbild(ctx, bild)
}

// TestEinFunkOhneTischGehtInsLeere — ein Wirt ohne offene Gastsitzung führt
// die Partie trotzdem, und dann gibt es keine Runde.
func TestEinFunkOhneTischGehtInsLeere(t *testing.T) {
	t.Parallel()

	funk := service.NeuerTischfunk("p1", &nehmerdoppel{}, nil)

	if fehler := funk.Verkuende(context.Background(), "anna",
		service.Befehlsausgang{Ereignisse: einZugMitGeheimnis()},
	); fehler != nil {
		t.Errorf("ohne Tisch kam %v", fehler)
	}
}

// TestEinGegangenerSitzHaeltDenTischNichtAuf — zwischen dem Aufzählen und
// dem Schicken kann eine Leitung zugegangen sein.
func TestEinGegangenerSitzHaeltDenTischNichtAuf(t *testing.T) {
	t.Parallel()

	runde := &service.Partierunde{}

	if fehler := runde.Schicke(context.Background(), "wegewesen",
		domain.ArtEreignis, []byte("{}")); fehler != nil {
		t.Errorf("an einen leeren Sitz kam %v", fehler)
	}
}

// TestEineKlagendeLeitungHaeltDieAnderenNichtAuf — wer am Tisch sitzt, soll
// nicht davon abhängen, ob ein anderer noch eine Leitung hat.
func TestEineKlagendeLeitungHaeltDieAnderenNichtAuf(t *testing.T) {
	t.Parallel()

	runde := &service.Partierunde{}

	bert, bertsLeitung := einPlatz("bert")
	clara, clarasLeitung := einPlatz("clara")
	bertsLeitung.klage = errors.New("die Leitung ist zu")
	runde.Setze(bert)
	runde.Setze(clara)

	nehmer := &nehmerdoppel{gibt: einZugMitGeheimnis()}
	funk := service.NeuerTischfunk("p1", nehmer, runde)

	fehler := funk.Verkuende(context.Background(), "anna",
		service.Befehlsausgang{Ereignisse: nehmer.gibt})
	if fehler == nil {
		t.Fatal("die tote Leitung kam nicht zurück")
	}

	if !strings.Contains(fehler.Error(), "bert") {
		t.Errorf("der Fehler nennt nicht, wen es traf: %v", fehler)
	}

	// **Und Clara hat trotzdem bekommen, was ihr zusteht.**
	if len(clarasLeitung.alle()) != 2 {
		t.Errorf("bei Clara kamen %d Nachrichten an",
			len(clarasLeitung.alle()))
	}
}

// TestEinUnverpackbaresEreignisKommtAlsFehler — `Oeffentlich` ist eine
// offene Karte, und was darin nicht durch `encoding/json` geht, ist ein
// Fehler und kein Absturz.
func TestEinUnverpackbaresEreignisKommtAlsFehler(t *testing.T) {
	t.Parallel()

	runde := &service.Partierunde{}
	bert, _ := einPlatz("bert")
	runde.Setze(bert)

	nehmer := &nehmerdoppel{gibt: []domain.Ereignis{{
		Art:         "unverpackbar",
		Oeffentlich: map[string]any{"kanal": make(chan int)},
	}}}

	if fehler := service.NeuerTischfunk("p1", nehmer, runde).Verkuende(
		context.Background(), "anna",
		service.Befehlsausgang{Ereignisse: nehmer.gibt},
	); fehler == nil {
		t.Error("das unverpackbare Ereignis ging durch")
	}
}

// TestEinGescheitertesBildKommtAlsFehler — an beiden Stellen, an denen es
// gerechnet wird.
func TestEinGescheitertesBildKommtAlsFehler(t *testing.T) {
	t.Parallel()

	for name, nehmer := range map[string]*nehmerdoppel{
		// **Ein Lesen, eine Klage** (#2481): Erlaubnis und Handlungen
		// haben keinen eigenen Weg mehr, der scheitern könnte.
		"das Bild klemmt": {bildklage: errors.New("das Bild klemmt")},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			runde := &service.Partierunde{}
			bert, _ := einPlatz("bert")
			runde.Setze(bert)

			gehoert := &klagendesFenster{}

			funk := service.NeuerTischfunk("p1", nehmer, runde).
				MitFenster("anna", gehoert)

			if fehler := funk.Verkuende(context.Background(), "dora",
				service.Befehlsausgang{},
			); fehler == nil {
				t.Error("das gescheiterte Bild ging durch")
			}
		})
	}
}

// TestEinKlagendesFensterKommtZurueck — an beiden seiner Wege.
func TestEinKlagendesFensterKommtZurueck(t *testing.T) {
	t.Parallel()

	for name, fenster := range map[string]*klagendesFenster{
		"beim Ereignis": {ereignisklage: errors.New("das Fenster ist weg")},
		"beim Bild":     {bildklage: errors.New("das Fenster ist weg")},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			nehmer := &nehmerdoppel{gibt: einZugMitGeheimnis()}

			funk := service.NeuerTischfunk("p1", nehmer, nil).
				MitFenster("anna", fenster)

			if fehler := funk.Verkuende(context.Background(), "bert",
				service.Befehlsausgang{Ereignisse: nehmer.gibt},
			); fehler == nil {
				t.Error("das klagende Fenster ging durch")
			}
		})
	}
}

// TestDerWirtGibtSeinFensterWeiter — `MitFenster` am Wirt reicht bis in den
// Funk durch.
func TestDerWirtGibtSeinFensterWeiter(t *testing.T) {
	t.Parallel()

	runde := &service.Partierunde{}
	gehoert := &mitschrift{}

	nehmer := &nehmerdoppel{gibt: einZugMitGeheimnis()}
	wirt, _ := einWirt(nehmer)
	wirt.MitRunde(runde).MitFenster("anna", gehoert)

	if fehler := wirt.Angekommen(context.Background(), einSpielwunsch(t,
		domain.Spielbefehl{Kennung: "b1", Art: "objekt.tappen"},
	)); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	if len(gehoert.ereignisse) != 1 || len(gehoert.bilder) != 1 {
		t.Errorf("am Fenster des Wirts kamen %d Ereignisse und %d Bilder an",
			len(gehoert.ereignisse), len(gehoert.bilder))
	}
}
