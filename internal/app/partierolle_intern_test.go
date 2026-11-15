package app

// **Ohne die Rolle geht kein Weg** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Warum jeder Weg einzeln geprüft wird
//
// **Weil eine Abweisung auf dem einen nichts über den anderen sagt.** Genau
// diese Gestalt hat N7 in Phase 5.4 aufgedeckt: `docs/AUTORISIERUNG.md`
// nannte für den Horchposten eine Mindestrolle, und **kein Ruf prüfte sie**.
// Eine Tabelle, die eine Rolle nennt, die nichts durchsetzt, verspricht mehr,
// als ihr Rumpf hält (§10).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// zaehlendeWache merkt sich, wonach gefragt wurde, und sagt Nein.
type zaehlendeWache struct{ gefragt []domain.Wirkung }

func (w *zaehlendeWache) Pruefe(_ context.Context, wirkung domain.Wirkung) error {
	w.gefragt = append(w.gefragt, wirkung)

	return domain.MitGrund(domain.ErrValidation, errors.New("keine Rolle"),
		"diese Rolle darf das nicht")
}

// **Jeder der sechs Wege fragt — und keiner schreibt, wenn die Antwort Nein
// ist.**
func TestJederPartiewegFragtNachDerRolle(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		wirkung domain.Wirkung
		fahre   func(*Bindungen) PartieAntwort
	}{
		"PartieAnlegen": {domain.WirkungSchreiben, func(b *Bindungen) PartieAntwort {
			return b.PartieAnlegen(t.Context(), PartieentwurfDTO{Spiel: "magic"})
		}},
		"Partie": {domain.WirkungLesen, func(b *Bindungen) PartieAntwort {
			return b.Partie(t.Context(), "p1", "sitz-1")
		}},
		"PartiePausieren": {domain.WirkungSchreiben, func(b *Bindungen) PartieAntwort {
			return b.PartiePausieren(t.Context(), "p1", "sitz-1")
		}},
		"PartieFortsetzen": {domain.WirkungSchreiben, func(b *Bindungen) PartieAntwort {
			return b.PartieFortsetzen(t.Context(), "p1", "sitz-1")
		}},
		"PartieBeenden": {domain.WirkungSchreiben, func(b *Bindungen) PartieAntwort {
			return b.PartieBeenden(t.Context(), "p1", "sitz-1")
		}},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wache := &zaehlendeWache{}
			gebunden := NeueBindungen(
				MitPartienleitung(NeueApp(""), bestandFuerDenRiegel(t)).
					MitWache(wache))

			antwort := fall.fahre(gebunden)
			if antwort.Fehler == nil {
				t.Fatal("der Weg kam durch")
			}

			if len(wache.gefragt) == 0 || wache.gefragt[0] != fall.wirkung {
				t.Errorf("gefragt wurde nach %v statt nach %q",
					wache.gefragt, fall.wirkung)
			}
		})
	}

	// **Und der Befehl ebenso** — er gibt eine andere Antwortgestalt und
	// steht deshalb daneben.
	t.Run("PartieBefehl", func(t *testing.T) {
		t.Parallel()

		wache := &zaehlendeWache{}
		gebunden := NeueBindungen(
			MitPartienleitung(NeueApp(""), bestandFuerDenRiegel(t)).
				MitWache(wache))

		antwort := gebunden.PartieBefehl(t.Context(), "p1", "b1", "sitz-1",
			"objekt.tappen", map[string]any{"objekt": "o1"}, "")
		if antwort.Fehler == nil {
			t.Fatal("der Befehl kam durch")
		}

		// **Eine leere Liste und kein `nil`** — auch auf dem Fehlerweg.
		if antwort.Ereignisse == nil {
			t.Error("die Ereignisliste ist nil")
		}

		if len(wache.gefragt) == 0 ||
			wache.gefragt[0] != domain.WirkungSchreiben {
			t.Errorf("gefragt wurde nach %v", wache.gefragt)
		}
	})
}

// **Ein Ausleger ohne Erlaubnis gibt eine leere Liste, kein `nil`.**
func TestEinAuslegerOhneErlaubnisGibtUeberDerNahtEineLeereListe(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)

	anwendung := NeueApp("")
	anwendung.db = db
	anwendung.profil.Store(1)
	anwendung.MitPartieaufbau(service.NeuerPartieaufbau(nil))
	anwendung.MitPartienaht(&schweigsameLeitung{})

	antwort := NeueBindungen(anwendung).Partie(t.Context(), "p1", "sitz-1")
	if antwort.Fehler != nil {
		t.Fatalf("Partie: %+v", antwort.Fehler)
	}

	if antwort.Erlaubt == nil || len(antwort.Erlaubt) != 0 {
		t.Errorf("es kam %v statt einer leeren Liste", antwort.Erlaubt)
	}

	// **Und ein Objekt ohne Zähler bekommt eine leere Abbildung** — sonst
	// steht im Fenster `zaehler: null`.
	if len(antwort.Partie.Objekte) != 1 ||
		antwort.Partie.Objekte[0].Zaehler == nil {
		t.Errorf("die Objekte: %+v", antwort.Partie.Objekte)
	}
}

// **Ein Ereignis ohne Inhalt bekommt leere Abbildungen.**
func TestEinEreignisOhneInhaltBekommtLeereAbbildungen(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.db = bestandFuerDenRiegel(t)
	anwendung.profil.Store(1)
	anwendung.MitPartienaht(&schweigsameLeitung{})

	antwort := NeueBindungen(anwendung).PartieBefehl(t.Context(),
		"p1", "b1", "sitz-1", "nichts", nil, "")
	if antwort.Fehler != nil {
		t.Fatalf("PartieBefehl: %+v", antwort.Fehler)
	}

	if len(antwort.Ereignisse) != 1 {
		t.Fatalf("es kamen %d Ereignisse", len(antwort.Ereignisse))
	}

	if antwort.Ereignisse[0].Oeffentlich == nil ||
		antwort.Ereignisse[0].Privat == nil {
		t.Errorf("das Ereignis gibt nil: %+v", antwort.Ereignisse[0])
	}
}

// **Was die Leitung klagt, klagt die Naht weiter — Weg für Weg.**
//
// **Jeder für sich**, denn eine Leitung, die auf **jedem** Weg klagt, klagt
// schon beim ersten — und alles danach bliebe ungemessen. Dieselbe Falle wie
// beim Partiespeicher in G1, und dort hat sie drei Fehlerwege verdeckt (§10).
func TestWasDieLeitungKlagtKlagtDieNahtWeiter(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("gerade dieser Weg ist zu")

	for _, weg := range []string{
		"Anlegen", "Tischbild", "Pausieren",
	} {
		t.Run(weg, func(t *testing.T) {
			t.Parallel()

			anwendung := NeueApp("")
			anwendung.db = bestandFuerDenRiegel(t)
			anwendung.profil.Store(1)
			anwendung.MitPartienaht(&klagendePartieleitung{weg: weg, klagt: kaputt})

			gebunden := NeueBindungen(anwendung)

			var antwort PartieAntwort

			switch weg {
			case "Anlegen":
				antwort = anwendung.PartieAnlegen(
					riegelkontext(t), domain.Partie{ID: "p1"})
			case "Pausieren":
				antwort = gebunden.PartiePausieren(t.Context(), "p1", "sitz-1")
			default:
				antwort = gebunden.Partie(t.Context(), "p1", "sitz-1")
			}

			if antwort.Fehler == nil {
				t.Fatal("der Weg kam durch")
			}
		})
	}
}

// **Ein Zuschauer kommt als DTO mit.**
//
// **Ohne diesen Lauf stünde die Schleife darüber ungeprüft da** — und ein
// Fenster, das Zuschauer zeigen will, bekäme eine leere Liste, obwohl welche
// da sind.
func TestEinZuschauerKommtAlsDTOMit(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.db = bestandFuerDenRiegel(t)
	anwendung.profil.Store(1)
	anwendung.MitPartienaht(&klagendePartieleitung{})

	antwort := NeueBindungen(anwendung).Partie(t.Context(), "p1", "sitz-1")
	if antwort.Fehler != nil {
		t.Fatalf("Partie: %+v", antwort.Fehler)
	}

	if len(antwort.Partie.Zuschauer) != 1 ||
		antwort.Partie.Zuschauer[0].Name != "Nachbar" {
		t.Errorf("die Zuschauer: %+v", antwort.Partie.Zuschauer)
	}
}

// **Auch die kopflose Fläche fragt nach der Rolle.**
//
// **Ein eigener Lauf, weil sie einen anderen Weg geht**: das Fenster fragt
// schon beim Aufbauen (dort liest es Decks), die Fläche bekommt eine fertige
// Partie und fragt beim Anlegen. Zwei Wege, zwei Prüfungen — und ohne diesen
// Lauf wäre die zweite ungemessen.
func TestAuchDieKopfloseFlaecheFragtNachDerRolle(t *testing.T) {
	t.Parallel()

	wache := &zaehlendeWache{}
	anwendung := MitPartienleitung(NeueApp(""), bestandFuerDenRiegel(t)).
		MitWache(wache)

	antwort := anwendung.PartieAnlegen(riegelkontext(t), domain.Partie{ID: "p1"})
	if antwort.Fehler == nil {
		t.Fatal("das Anlegen kam durch")
	}

	if len(wache.gefragt) != 1 || wache.gefragt[0] != domain.WirkungSchreiben {
		t.Errorf("gefragt wurde nach %v", wache.gefragt)
	}
}

// **Eine private Ergänzung kommt nur bei ihrem Empfänger an — als DTO.**
//
// **Die Projektion steht in `domain`, ihre Übersetzung hier** — und ohne
// diesen Lauf bliebe offen, ob die Naht sie überhaupt weiterreicht.
func TestEinePrivateErgaenzungKommtAlsDTOAn(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.db = bestandFuerDenRiegel(t)
	anwendung.profil.Store(1)
	anwendung.MitPartienaht(&ergaenzendeLeitung{})

	gebunden := NeueBindungen(anwendung)

	fuerIhn := gebunden.PartieBefehl(t.Context(), "p1", "b1", "sitz-1", "karte.gezogen", nil, "")
	if fuerIhn.Fehler != nil {
		t.Fatalf("PartieBefehl: %+v", fuerIhn.Fehler)
	}

	if len(fuerIhn.Ereignisse) != 1 ||
		fuerIhn.Ereignisse[0].Privat["karte"] != "wald" {
		t.Errorf("die Ergänzung fehlt: %+v", fuerIhn.Ereignisse)
	}

	// **Und beim anderen kommt sie nicht an** — die Gegenprobe, ohne die
	// oben nur stünde, dass irgendetwas durchgereicht wird.
	fuerDenAnderen := gebunden.PartieBefehl(t.Context(),
		"p1", "b2", "sitz-2", "karte.gezogen", nil, "")
	if fuerDenAnderen.Fehler != nil {
		t.Fatalf("PartieBefehl: %+v", fuerDenAnderen.Fehler)
	}

	if len(fuerDenAnderen.Ereignisse[0].Privat) != 0 {
		t.Errorf("die fremde Ergänzung kam an: %+v",
			fuerDenAnderen.Ereignisse[0].Privat)
	}
}

// ergaenzendeLeitung gibt ein Ereignis mit einer privaten Ergänzung.
type ergaenzendeLeitung struct{ schweigsameLeitung }

func (e *ergaenzendeLeitung) BefehlVerkuendet(
	_ context.Context, _ domain.Spielbefehl, verkuende service.Verkuendung,
) (service.Befehlsausgang, error) {
	ausgang := service.Befehlsausgang{
		Fassung: 2,
		Ereignisse: []domain.Ereignis{{
			Folge: 1, Art: "karte.gezogen",
			Oeffentlich: map[string]any{"sitz": "sitz-1"},
			Privat: []domain.Ergaenzung{{
				Empfaenger: "sitz-1",
				Inhalt:     map[string]any{"karte": "wald"},
			}},
		}},
	}
	verkuende(ausgang)

	return ausgang, nil
}
