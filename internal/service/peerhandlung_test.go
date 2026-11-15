package service_test

// **Was die Identität ändert** — Werkstück N1 von Phase 5.4
// (#1454).
//
// **Eine zweite Datei**, weil `peerwerk_test.go` mit ihr über die
// Zeilengrenze aus §2 ging. Der Schnitt folgt der Frage: dort steht, **wie
// es steht**, hier, **was es ändert** — anlegen, ablösen, und die Fehlerwege
// dazu.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Sicherstellen legt eine an, wenn es keine gibt.**
func TestSicherstellenLegtDieErsteIdentitaetAn(t *testing.T) {
	t.Parallel()

	vermerke := &vermerkprobe{}
	bund := &bundprobe{}

	lage, fehler := einPeerwerk(t, vermerke, bund).Sicherstellen(t.Context())
	if fehler != nil {
		t.Fatalf("Sicherstellen: %v", fehler)
	}

	if lage.Stand != domain.PeerstandBereit || lage.Kennung == "" {
		t.Fatalf("die Lage ist %+v", lage)
	}

	if bund.geschrieben != 1 {
		t.Errorf("in den Bund wurde %d mal geschrieben", bund.geschrieben)
	}

	if vermerke.vermerk.Kennung != string(lage.Kennung) {
		t.Errorf("vermerkt wurde %q", vermerke.vermerk.Kennung)
	}

	// **Und ein zweiter Ruf legt nichts Neues an.**
	zweite, fehler := einPeerwerk(t, vermerke, bund).Sicherstellen(t.Context())
	if fehler != nil {
		t.Fatalf("zweites Sicherstellen: %v", fehler)
	}

	if zweite.Kennung != lage.Kennung {
		t.Errorf("der zweite Ruf gab %q", zweite.Kennung)
	}

	if bund.geschrieben != 1 {
		t.Errorf("in den Bund wurde %d mal geschrieben", bund.geschrieben)
	}
}

// **Sicherstellen legt bei Verlust und Widerspruch nichts an.**
//
// **Das ist die Zusage des Issues**, und sie ist der Grund für die vier
// Stände: wer hier eine neue Identität anlegte, täte genau das, was „nie
// still" heißt.
func TestSicherstellenLegtBeiVerlustUndWiderspruchNichtsAn(t *testing.T) {
	t.Parallel()

	kennung, _ := eineKennung(t, 1)
	_, fremdes := eineKennung(t, 99)

	faelle := map[string]*bundprobe{
		"verloren": {},
		"fremd":    {geheimnis: fremdes},
	}

	for name, bund := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			vermerke := &vermerkprobe{vermerk: domain.Peervermerk{
				Kennung: string(kennung), Seit: derPeerlauf,
			}}

			if _, fehler := einPeerwerk(t, vermerke, bund).
				Sicherstellen(t.Context()); !errors.Is(
				fehler, domain.ErrConflict) {
				t.Fatalf("der Fehler ist %v", fehler)
			}

			if vermerke.vermerk.Kennung != string(kennung) {
				t.Errorf("der Vermerk steht auf %q", vermerke.vermerk.Kennung)
			}
		})
	}
}

// **Rotieren legt eine neue an und behält die alte als Vorgängerin.**
func TestRotierenBehaeltDieAlteAlsVorgaengerin(t *testing.T) {
	t.Parallel()

	alt, geheimnis := eineKennung(t, 1)
	vermerke := &vermerkprobe{vermerk: domain.Peervermerk{
		Kennung: string(alt), Seit: derPeerlauf,
	}}

	lage, fehler := einPeerwerk(t, vermerke, &bundprobe{geheimnis: geheimnis}).
		Rotieren(t.Context())
	if fehler != nil {
		t.Fatalf("Rotieren: %v", fehler)
	}

	if lage.Kennung == alt {
		t.Fatal("nach der Rotation steht dieselbe Kennung da")
	}

	if lage.Vorher != alt {
		t.Errorf("die Vorgängerin ist %q", lage.Vorher)
	}

	if vermerke.rotiert != 1 {
		t.Errorf("rotiert wurde %d mal", vermerke.rotiert)
	}

	// **Und die Antwort sagt, was das für Kontakte heißt** — eine Rotation
	// ohne diesen Satz sähe aus wie eine Einstellung.
	if len(lage.Befunde) != 1 {
		t.Fatalf("die Befunde sind %v", lage.Befunde)
	}
}

// **Auch aus einem Verlust heraus lässt sich rotieren.**
//
// **Das ist der Weg heraus**: `Sicherstellen` weigert sich, `Rotieren` ist
// die ausdrückliche Handlung — und danach steht wieder eine Identität da.
func TestAusEinemVerlustLaesstSichRotieren(t *testing.T) {
	t.Parallel()

	alt, _ := eineKennung(t, 1)
	vermerke := &vermerkprobe{vermerk: domain.Peervermerk{
		Kennung: string(alt), Seit: derPeerlauf,
	}}

	lage, fehler := einPeerwerk(t, vermerke, &bundprobe{}).Rotieren(t.Context())
	if fehler != nil {
		t.Fatalf("Rotieren: %v", fehler)
	}

	if lage.Stand != domain.PeerstandBereit || lage.Vorher != alt {
		t.Fatalf("die Lage ist %+v", lage)
	}
}

// **Jeder Fehlerweg reicht seinen Fehler durch.**
//
// **Und keiner davon gibt eine Lage.** Ein Werk, das bei einem unlesbaren
// Bestand „es gibt noch keine" antwortete, führte geradewegs in das stille
// Übernehmen: die nächste Handlung legte eine neue Identität an.
func TestJederFehlerwegDerPeerlageReichtDurch(t *testing.T) {
	t.Parallel()

	kennung, geheimnis := eineKennung(t, 1)
	eigen := errors.New("die Platte ist voll")

	faelle := map[string]struct {
		vermerke *vermerkprobe
		bund     *bundprobe
		fahre    func(*service.Peerwerk) error
	}{
		"der Vermerk lässt sich nicht lesen": {
			&vermerkprobe{lesefehler: eigen}, &bundprobe{},
			func(w *service.Peerwerk) error {
				_, fehler := w.Lage(t.Context())

				return fehler
			},
		},
		"der Bund lässt sich nicht lesen": {
			&vermerkprobe{}, &bundprobe{lesefehler: eigen},
			func(w *service.Peerwerk) error {
				_, fehler := w.Lage(t.Context())

				return fehler
			},
		},
		"das Geheimnis im Bund ist krumm": {
			&vermerkprobe{}, &bundprobe{geheimnis: "keinhex"},
			func(w *service.Peerwerk) error {
				_, fehler := w.Lage(t.Context())

				return fehler
			},
		},
		"das Nachtragen scheitert": {
			&vermerkprobe{merkfehler: eigen}, &bundprobe{geheimnis: geheimnis},
			func(w *service.Peerwerk) error {
				_, fehler := w.Lage(t.Context())

				return fehler
			},
		},
		"der Bund lässt sich nicht beschreiben": {
			&vermerkprobe{}, &bundprobe{schreibfehler: eigen},
			func(w *service.Peerwerk) error {
				_, fehler := w.Sicherstellen(t.Context())

				return fehler
			},
		},
		"das Merken nach dem Anlegen scheitert": {
			&vermerkprobe{merkfehler: eigen}, &bundprobe{},
			func(w *service.Peerwerk) error {
				_, fehler := w.Sicherstellen(t.Context())

				return fehler
			},
		},
		"der Vermerk vor der Rotation lässt sich nicht lesen": {
			&vermerkprobe{lesefehler: eigen}, &bundprobe{geheimnis: geheimnis},
			func(w *service.Peerwerk) error {
				_, fehler := w.Rotieren(t.Context())

				return fehler
			},
		},
		"das Sicherstellen über einem unlesbaren Vermerk": {
			&vermerkprobe{lesefehler: eigen}, &bundprobe{},
			func(w *service.Peerwerk) error {
				_, fehler := w.Sicherstellen(t.Context())

				return fehler
			},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := fall.fahre(
				einPeerwerk(t, fall.vermerke, fall.bund)); fehler == nil {
				t.Fatal("der Fehler kam nicht durch")
			}
		})
	}

	// **Die Kennung aus dem Vorspann wird gebraucht** — sonst stünde sie
	// hier als toter Wert.
	if kennung == "" {
		t.Fatal("die Probekennung ist leer")
	}
}

// **Eine Rotation ohne Vorgängerin sagt nichts über Kontakte.**
//
// **Die Gegenprobe** (§10): ein Satz „prüft den Fingerabdruck erneut" ohne
// eine abgelöste Kennung spräche über eine Bestätigung, die es nie gab.
func TestEineRotationOhneVorgaengerinSagtNichtsUeberKontakte(t *testing.T) {
	t.Parallel()

	lage, fehler := einPeerwerk(t, &vermerkprobe{}, &bundprobe{}).
		Rotieren(t.Context())
	if fehler != nil {
		t.Fatalf("Rotieren: %v", fehler)
	}

	if lage.Vorher != "" || len(lage.Befunde) != 0 {
		t.Fatalf("die Lage ist %+v", lage)
	}
}

// **Ohne Zufall entsteht keine Identität.**
//
// **Die gefährliche Richtung wäre die andere**: ein Schlüssel aus einer
// versiegten Quelle wäre vorhersagbar, und eine Identität, die jemand erraten
// kann, ist keine.
func TestOhneZufallEntstehtKeinePeeridentitaet(t *testing.T) {
	t.Parallel()

	werk := einPeerwerk(t, &vermerkprobe{}, &bundprobe{}).
		MitZufall(versiegt{})

	if _, fehler := werk.Sicherstellen(t.Context()); fehler == nil {
		t.Fatal("ohne Zufall kam eine Identität")
	}
}

// versiegt gibt nichts.
type versiegt struct{}

func (versiegt) Read([]byte) (int, error) {
	return 0, errors.New("der Zufall ist zu Ende")
}

// **Eine Rotation stellt jeden Kontakt zurück — und zwar davor.**
//
// # Warum die Reihenfolge geprüft wird und nicht nur der Ruf
//
// **Weil der Schaden nur in einer Richtung heilbar ist.** Scheitert das
// Zurückstellen, ist nichts geschehen. Scheiterte stattdessen die Rotation
// danach, stünde eine unveränderte Identität mit unbestätigten Kontakten da —
// lästig, von Hand zu heilen. Andersherum stünde eine **neue** Identität mit
// Bestätigungen da, die einer alten gelten, und das sieht niemand.
func TestEineRotationStelltDieKontakteZurueck(t *testing.T) {
	t.Parallel()

	vermerk := &vermerkprobe{}
	bund := &bundprobe{}
	kontakte := &kontaktprobe{}

	werk := service.NeuesPeerwerk(vermerk, bund, kontakte,
		func() time.Time { return derPeerlauf })

	if _, fehler := werk.Rotieren(t.Context()); fehler != nil {
		t.Fatalf("Rotieren: %v", fehler)
	}

	if kontakte.gerufen != 1 {
		t.Fatalf("die Kontakte wurden %dmal zurückgestellt", kontakte.gerufen)
	}
}

// **Und scheitert das Zurückstellen, bleibt die Identität, wie sie war.**
func TestScheiterndeRueckstellungLaesstDieIdentitaet(t *testing.T) {
	t.Parallel()

	vermerk := &vermerkprobe{}
	bund := &bundprobe{}

	werk := service.NeuesPeerwerk(vermerk, bund,
		&kontaktprobe{scheitert: errors.New("der Bestand sagt nein")},
		func() time.Time { return derPeerlauf })

	if _, fehler := werk.Rotieren(t.Context()); fehler == nil {
		t.Fatal("die Rotation kam durch")
	}

	if bund.geheimnis != "" {
		t.Errorf("ein Schlüssel wurde geschrieben: %q", bund.geheimnis)
	}
}

// **Ein Bund, der nichts annimmt, ist kein fehlender Schlüssel** —
// S13.NAT-03 (#1706):
// „fehlender Profil-Schlüssel und gescheiterte Schlüsselablage werden
// unterschieden".
func TestEinGescheitertesAblegenHatSeinenEigenenSatz(t *testing.T) {
	t.Parallel()

	eigen := errors.New("der Schlüsselbund ist gesperrt")

	_, fehler := einPeerwerk(t, &vermerkprobe{}, &bundprobe{schreibfehler: eigen}).Sicherstellen(t.Context())

	if !errors.Is(fehler, domain.ErrSchluesselbundAblage) || !errors.Is(fehler, eigen) {
		t.Errorf("es kam %v", fehler)
	}

	if errors.Is(fehler, domain.ErrOhnePeeridentitaet) {
		t.Errorf("als fehlende Identität gedeutet: %v", fehler)
	}
}
