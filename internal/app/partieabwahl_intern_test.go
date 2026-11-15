package app

// **Die gewählte Abwahl reist bis in die eingefrorene Partie** — P1a,
// Befund C-01 aus dem Audit zur Produktionsreife 0.9.0
// (#2335).
//
// # Was hier geprüft wird
//
// **Dass dieselbe Wahl über beide Bauwege dieselbe Partie gibt** — das
// Fenster über [Bindungen.PartieAnlegen] und die kopflosen Flächen über
// [Partieflaeche.PartieUnterRegeln] —, und zwar **im Bestand**, nicht nur in
// der Antwort: eine Antwort, die die Wahl wiederholt, belegte nichts über
// das, was eingefroren wurde. Der Befund war genau so ein Fall — die
// Vorschau sagte manuell, die gespeicherte Matrix automatisch.
//
// **Und dass eine unbekannte Abwahl nichts anlegt.** Eine Absage, nach der
// trotzdem eine Partie dasteht, wäre ein halber Stand.

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// eingefroren liest die Partie so, wie sie im Bestand steht.
func eingefroren(t *testing.T, gebunden *Bindungen, kennung string) domain.Partie {
	t.Helper()

	partie, fehler := storage.NeuePartieablage(gebunden.anwendung.db).Partie(
		riegelkontext(t), kennung)
	if fehler != nil {
		t.Fatalf("die Partie %q steht nicht im Bestand: %v", kennung, fehler)
	}

	return partie
}

// abwahlPruefen hält fest, was eine Partie mit abgewählter Zugstruktur
// tragen muss: die Wahl im Regelwerk **und** in der wirksamen Matrix.
func abwahlPruefen(t *testing.T, partie domain.Partie) {
	t.Helper()

	if !slices.Equal(partie.Regelwerk.Abgewaehlt,
		[]domain.Regelbereich{domain.BereichZugstruktur}) {
		t.Errorf("das Regelwerk führt %v", partie.Regelwerk.Abgewaehlt)
	}

	if grad := partie.Profil.Matrix.Grad(domain.BereichZugstruktur); grad !=
		domain.UnterstuetzungManuell {
		t.Errorf("die Zugstruktur steht eingefroren auf %q", grad)
	}
}

func TestDasFensterFriertDieGewaehlteAbwahlEin(t *testing.T) {
	t.Parallel()

	gebunden, deckID := eineSpielende(t)

	entwurf := einEntwurf(deckID)
	entwurf.Regelwerk.Abgewaehlt = []string{"zugstruktur"}

	antwort := gebunden.PartieAnlegen(riegelkontext(t), entwurf)
	if antwort.Fehler != nil {
		t.Fatalf("PartieAnlegen: %+v", antwort.Fehler)
	}

	if !slices.Equal(antwort.Partie.Regelwerk.Abgewaehlt,
		[]string{"zugstruktur"}) {
		t.Errorf("die Antwort nennt %v", antwort.Partie.Regelwerk.Abgewaehlt)
	}

	abwahlPruefen(t, eingefroren(t, gebunden, antwort.Partie.ID))
}

func TestDieKopfloseFlaecheFriertDieGewaehlteAbwahlEin(t *testing.T) {
	t.Parallel()

	gebunden, deckID := eineSpielende(t)

	antwort := PartieflaecheUeber(gebunden.anwendung.db).PartieUnterRegeln(
		riegelkontext(t), "magic", "", []Sitzwunsch{
			{Kennung: "sitz-1", Name: "David", DeckID: deckID},
			{Kennung: "sitz-2", Name: "Gegenüber", DeckID: deckID},
		}, RegelwerkDTO{Abgewaehlt: []string{"zugstruktur"}})
	if antwort.Fehler != nil {
		t.Fatalf("PartieUnterRegeln: %+v", antwort.Fehler)
	}

	abwahlPruefen(t, eingefroren(t, gebunden, antwort.Partie.ID))
}

// **Die Gegenprobe**: ohne Abwahl bleibt die Zugstruktur automatisch —
// sonst prüfte [abwahlPruefen] eine Matrix, die immer manuell sagt (§10).
func TestOhneAbwahlBleibtDieZugstrukturAutomatisch(t *testing.T) {
	t.Parallel()

	gebunden, deckID := eineSpielende(t)

	antwort := gebunden.PartieAnlegen(riegelkontext(t), einEntwurf(deckID))
	if antwort.Fehler != nil {
		t.Fatalf("PartieAnlegen: %+v", antwort.Fehler)
	}

	partie := eingefroren(t, gebunden, antwort.Partie.ID)

	if len(partie.Regelwerk.Abgewaehlt) != 0 {
		t.Errorf("ohne Wahl führt das Regelwerk %v", partie.Regelwerk.Abgewaehlt)
	}

	if grad := partie.Profil.Matrix.Grad(domain.BereichZugstruktur); grad !=
		domain.UnterstuetzungAutomatisch {
		t.Errorf("ohne Wahl steht die Zugstruktur auf %q", grad)
	}
}

// **Eine unbekannte Abwahl legt an keinem Weg etwas an.**
func TestEineUnbekannteAbwahlLegtNichtsAn(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name    string
		anlegen func(*Bindungen, string) PartieAntwort
	}{
		{"Fenster", func(gebunden *Bindungen, deckID string) PartieAntwort {
			entwurf := einEntwurf(deckID)
			entwurf.Regelwerk.Abgewaehlt = []string{"zugstruktur", "quatsch"}

			return gebunden.PartieAnlegen(riegelkontext(t), entwurf)
		}},
		{"kopflos", func(gebunden *Bindungen, deckID string) PartieAntwort {
			return PartieflaecheUeber(gebunden.anwendung.db).PartieUnterRegeln(
				riegelkontext(t), "magic", "", []Sitzwunsch{
					{Kennung: "sitz-1", Name: "David", DeckID: deckID},
					{Kennung: "sitz-2", Name: "Gegenüber", DeckID: deckID},
				}, RegelwerkDTO{Abgewaehlt: []string{"quatsch"}})
		}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			gebunden, deckID := eineSpielende(t)

			antwort := fall.anlegen(gebunden, deckID)

			if antwort.Fehler == nil || antwort.Fehler.Code != CodeValidierung {
				t.Fatalf("die unbekannte Abwahl kam durch: %+v", antwort.Fehler)
			}

			partien, fehler := storage.NeuePartieablage(
				gebunden.anwendung.db).Partien(riegelkontext(t))
			if fehler != nil {
				t.Fatalf("Partien: %v", fehler)
			}

			if len(partien) != 0 {
				t.Errorf("nach der Absage stehen %d Partien da", len(partien))
			}
		})
	}
}
