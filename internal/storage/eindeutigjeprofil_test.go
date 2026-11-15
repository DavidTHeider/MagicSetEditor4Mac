package storage_test

// **Eindeutig heißt: eindeutig je Profil** — Werkstück 12 von Phase 4.12
// (#1351).
//
// # Woran das aufgefallen ist
//
// **An der Deckübernahme.** Der Flächenlauf legte in einem fremden Profil ein
// Deck „Nachtlauf" an und holte es herüber — und der Bestand wies es ab:
// `UNIQUE constraint failed: decks.spiel, decks.name`. Die Übernahme wäre
// damit nur für Namen möglich gewesen, die im eigenen Profil noch frei sind.
//
// # Und es ist eine Klasse, kein Fall
//
// Wanderung 0035 hat zehn Wurzeln auf ein Profil geklemmt. Die
// **Eindeutigkeit** hat sie dabei an **einer** Stelle mitgezogen — `sammlung`
// wurde neu gebaut und trägt `UNIQUE (profil_id, printing, …)`. Die drei
// Indizes, die als `CREATE UNIQUE INDEX` danebenstanden, blieben stehen:
// `decks`, `turniere`, `formate` je über `(spiel, name)`, und `sets` über
// `(herkunft)`.
//
// **Was das hieß**: das Profil eines anderen Menschen entschied mit, welche
// Namen man selbst noch vergeben darf — und die Meldung sagte „Konflikt mit
// dem gegenwärtigen Stand", was auf einen Stand zeigte, den man gar nicht
// sehen kann (§8).

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Zwei Profile dürfen ein Deck desselben Namens führen.**
func TestZweiProfileDuerfenDenselbenDecknamen(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	if _, fehler := ablage.Anlegen(imProfil(t, 1),
		domain.Deck{Name: "Nachtlauf", Spiel: "probe"}); fehler != nil {
		t.Fatalf("im ersten Profil: %v", fehler)
	}

	if _, fehler := ablage.Anlegen(imProfil(t, 2),
		domain.Deck{Name: "Nachtlauf", Spiel: "probe"}); fehler != nil {
		t.Fatalf("im zweiten Profil: %v", fehler)
	}
}

// **Und im selben Profil zweimal derselbe Name bleibt ein Konflikt.**
//
// **Die Gegenprobe** (§10): ohne sie wäre ein Bestand ohne jede Eindeutigkeit
// ebenso grün.
func TestImselbenProfilBleibtDerDecknameEindeutig(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	if _, fehler := ablage.Anlegen(imProfil(t, 1),
		domain.Deck{Name: "Nachtlauf", Spiel: "probe"}); fehler != nil {
		t.Fatalf("das erste: %v", fehler)
	}

	if _, fehler := ablage.Anlegen(imProfil(t, 1),
		domain.Deck{Name: "Nachtlauf", Spiel: "probe"}); fehler == nil {
		t.Error("derselbe Name kam im selben Profil zweimal durch")
	}
}

// **Dasselbe für Turniere.**
func TestZweiProfileDuerfenDenselbenTurniernamen(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueTurnierablage(db)

	for _, profil := range []int64{1, 2} {
		if _, fehler := ablage.Anlegen(imProfil(t, profil), domain.Turnier{
			Name: "Freitagsrunde", Spiel: "probe",
		}); fehler != nil {
			t.Fatalf("im Profil %d: %v", profil, fehler)
		}
	}
}

// **Und für Formate.**
func TestZweiProfileDuerfenDenselbenFormatnamen(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueFormatablage(db)

	for _, profil := range []int64{1, 2} {
		if _, fehler := ablage.Anlegen(imProfil(t, profil), domain.Format{
			Name: "Küchentisch", Spiel: "probe", Mindestgroesse: 60,
		}); fehler != nil {
			t.Fatalf("im Profil %d: %v", profil, fehler)
		}
	}
}

// imProfil gibt einen Kontext mit dem genannten Profil.
func imProfil(t *testing.T, profilID int64) context.Context {
	t.Helper()

	return storage.MitProfil(t.Context(), profilID)
}

// **Und zwei Profile dürfen dasselbe Scryfall-Set führen.**
//
// **„Eine Quelle, ein Set" bleibt** (0015) — es gilt jetzt je Profil. Ohne
// diesen Lauf bliebe der vierte Index unbemerkt: er steht in einer anderen
// Wanderung als die drei darüber, und ein Wächter über „(spiel, name)" fände
// ihn nicht.
func TestZweiProfileDuerfenDasselbeScryfallSet(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSetSpeicher(db)
	leser := storage.NeuerSetLeser(db)

	for _, profil := range []int64{1, 2} {
		set := &domain.Set{Name: "Probe", Game: "probe", Style: "probe-schlicht"}

		kennung, fehler := speicher.SetAnlegen(imProfil(t, profil), set, nil)
		if fehler != nil {
			t.Fatalf("im Profil %d: %v", profil, fehler)
		}

		if setzfehler := leser.HerkunftSetzen(imProfil(t, profil),
			kennung, "scryfall:prf"); setzfehler != nil {
			t.Fatalf("die Herkunft im Profil %d: %v", profil, setzfehler)
		}
	}
}
