package service_test

// **Die zweite Seite** — Werkstück A16 von Phase 5.3
// (#1515).
//
// **Der Satz, gegen den hier geprüft wird**: „bei Wechsel konsistent
// fortsetzen oder explizit ungültig, **keine still gemischten Seiten**."

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// vieleTreffer baut so viele, dass geblättert werden muss.
func vieleTreffer(zahl int) []domain.Retrievaltreffer {
	treffer := make([]domain.Retrievaltreffer, 0, zahl)

	for i := range zahl {
		kennung := "o" + string(rune('a'+i)) + ":0"
		treffer = append(treffer, einIndextreffer(kennung,
			"o"+string(rune('a'+i)), "Karte "+string(rune('a'+i)), i+1))
	}

	return treffer
}

// **Eine zweite Seite setzt fort, wo die erste aufhörte.**
func TestEineZweiteSeiteSetztFort(t *testing.T) {
	t.Parallel()

	quelle := &recherchequellendoppel{
		stand: einAktiverStand(), treffer: vieleTreffer(7),
	}

	dienst := einRecherchedienst(quelle, nil)
	anfrage := domain.Rechercheanfrage{
		Fall: domain.FallKarten, Frage: "karte", Grenze: 3,
	}

	erste, fehler := dienst.Recherchieren(t.Context(), anfrage, "")
	if fehler != nil {
		t.Fatalf("die erste Seite: %v", fehler)
	}

	if len(erste.Treffer) != 3 || erste.Fortsetzung == "" {
		t.Fatalf("die erste Seite ist %+v", erste)
	}

	zweite, fehler := dienst.Recherchieren(t.Context(), anfrage, erste.Fortsetzung)
	if fehler != nil {
		t.Fatalf("die zweite Seite: %v", fehler)
	}

	if len(zweite.Treffer) != 3 || zweite.Treffer[0].Kennung == erste.Treffer[0].Kennung {
		t.Fatalf("die zweite Seite ist %+v", zweite.Treffer)
	}

	// **Die Ränge zählen über die ganze Liste** — Platz vier heißt vier, auch
	// als erster der zweiten Seite.
	if zweite.Treffer[0].Rang != 4 {
		t.Errorf("der erste Treffer der zweiten Seite trägt Rang %d",
			zweite.Treffer[0].Rang)
	}

	// **Die letzte Seite trägt keinen Zeiger.**
	dritte, fehler := dienst.Recherchieren(t.Context(), anfrage, zweite.Fortsetzung)
	if fehler != nil {
		t.Fatalf("die dritte Seite: %v", fehler)
	}

	if len(dritte.Treffer) != 1 || dritte.Fortsetzung != "" {
		t.Fatalf("die dritte Seite ist %+v", dritte)
	}
}

// **Ein Zeiger zu einer anderen Frage wird abgewiesen.**
func TestEinZeigerZuEinerAnderenFrageWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	quelle := &recherchequellendoppel{
		stand: einAktiverStand(), treffer: vieleTreffer(7),
	}

	dienst := einRecherchedienst(quelle, nil)

	erste, fehler := dienst.Recherchieren(t.Context(), domain.Rechercheanfrage{
		Fall: domain.FallKarten, Frage: "karte", Grenze: 3,
	}, "")
	if fehler != nil {
		t.Fatalf("die erste Seite: %v", fehler)
	}

	if _, fehler := dienst.Recherchieren(t.Context(), domain.Rechercheanfrage{
		Fall: domain.FallKarten, Frage: "etwas anderes", Grenze: 3,
	}, erste.Fortsetzung); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("der fremde Zeiger kam durch: %v", fehler)
	}
}

// **Ein Zeiger zu einem anderen Indexstand ebenso.**
//
// **Das ist der Fall, für den es die Bindung gibt**: nach einem Neuaufbau ist
// die Liste eine andere, und die zweite Seite daraus wäre eine Mischung.
func TestEinZeigerZuEinemAnderenIndexstandWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	quelle := &recherchequellendoppel{
		stand: einAktiverStand(), treffer: vieleTreffer(7),
	}

	dienst := einRecherchedienst(quelle, nil)
	anfrage := domain.Rechercheanfrage{
		Fall: domain.FallKarten, Frage: "karte", Grenze: 3,
	}

	erste, fehler := dienst.Recherchieren(t.Context(), anfrage, "")
	if fehler != nil {
		t.Fatalf("die erste Seite: %v", fehler)
	}

	// **Der Index wird neu gebaut** — anderer Korpusstand, anderer Abdruck.
	neuerStand := einAktiverStand()
	neuerStand.Manifest.Korpusstaende = []domain.Korpusstand{{
		Quelle: domain.QuelleOracle, Revision: "2026-09-12",
	}}
	quelle.stand = neuerStand

	if _, fehler := dienst.Recherchieren(t.Context(), anfrage,
		erste.Fortsetzung); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("der Zeiger aus dem alten Index kam durch: %v", fehler)
	}
}

// **Ein abgelaufener Zeiger wird abgewiesen.**
func TestEinAbgelaufenerZeigerWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	quelle := &recherchequellendoppel{
		stand: einAktiverStand(), treffer: vieleTreffer(7),
	}

	jetzt := dieRechercheuhr()
	dienst := service.NeuerRecherchedienst(quelle, nil, profildoppel{}).
		MitUhr(func() time.Time { return jetzt })

	anfrage := domain.Rechercheanfrage{
		Fall: domain.FallKarten, Frage: "karte", Grenze: 3,
	}

	erste, fehler := dienst.Recherchieren(t.Context(), anfrage, "")
	if fehler != nil {
		t.Fatalf("die erste Seite: %v", fehler)
	}

	jetzt = jetzt.Add(domain.Zeigerfrist)

	if _, fehler := dienst.Recherchieren(t.Context(), anfrage,
		erste.Fortsetzung); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("der abgelaufene Zeiger kam durch: %v", fehler)
	}
}

// **Ein Zeiger hinter das Ende gibt eine leere Seite und keinen Fehler.**
//
// **Weil die Liste kürzer geworden sein kann**, ohne dass sich der Abdruck
// ändert — ein Widerruf nimmt Treffer weg, und dann zeigt ein gültiger Zeiger
// ins Leere. Das ist eine Auskunft und kein Fehlschlag.
func TestEinZeigerHinterDasEndeGibtEineLeereSeite(t *testing.T) {
	t.Parallel()

	quelle := &recherchequellendoppel{
		stand: einAktiverStand(), treffer: vieleTreffer(7),
	}

	dienst := einRecherchedienst(quelle, nil)
	anfrage := domain.Rechercheanfrage{
		Fall: domain.FallKarten, Frage: "karte", Grenze: 3,
	}

	erste, fehler := dienst.Recherchieren(t.Context(), anfrage, "")
	if fehler != nil {
		t.Fatalf("die erste Seite: %v", fehler)
	}

	// **Die Liste schrumpft** — zwei Treffer sind widerrufen.
	quelle.treffer = vieleTreffer(2)

	zweite, fehler := dienst.Recherchieren(t.Context(), anfrage, erste.Fortsetzung)
	if fehler != nil {
		t.Fatalf("die zweite Seite: %v", fehler)
	}

	if len(zweite.Treffer) != 0 || zweite.Leere != domain.LeereNichtsGefunden {
		t.Fatalf("die zweite Seite ist %+v", zweite)
	}
}
