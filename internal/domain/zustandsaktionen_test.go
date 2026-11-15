package domain_test

// **Die drei zustandsbasierten Aktionen, die einen Sitz kosten** — Werkstück
// M17 der Regelmaschine
// (#1618).
//
// # Wogegen hier geprüft wird
//
//   - CR 704.5a — null oder weniger Leben.
//   - CR 704.5b — wer aus einer leeren Bibliothek ziehen wollte.
//   - CR 704.5c — zehn oder mehr Vergiftungsmarken.
//   - CR 704.4 — sie laufen **gleichzeitig**.
//
// **Und warum sie hier stehen und nicht erst in `M5`.** `M17` führt drei
// Ausscheidungsgründe ein — `leben`, `bibliothek`, `gift`. Ein Grund, den
// nichts erzeugt, ist ein Versprechen, das keiner einlöst (§10); `M5` baut
// die übrigen zweiundzwanzig Aktionen dazu.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestBeiNullLebenScheidetEinSitzAus(t *testing.T) {
	t.Parallel()

	partie := eineEndpartie()
	partie.Sitze[1].Leben = 0

	danach := partie.Zustandspruefung()

	if danach.Ausgang.Sieger != "sitz-1" {
		t.Fatalf("Sieger = %q, erwartet sitz-1", danach.Ausgang.Sieger)
	}

	if len(danach.Ausgang.Ausgeschieden) != 1 ||
		danach.Ausgang.Ausgeschieden[0].Grund != domain.GrundLeben {
		t.Fatalf("ausgeschieden = %+v", danach.Ausgang.Ausgeschieden)
	}
}

func TestBeiZehnVergiftungsmarkenScheidetEinSitzAus(t *testing.T) {
	t.Parallel()

	partie := eineEndpartie()
	partie.Sitze[1].Vergiftung = 10

	danach := partie.Zustandspruefung()

	if len(danach.Ausgang.Ausgeschieden) != 1 ||
		danach.Ausgang.Ausgeschieden[0].Grund != domain.GrundGift {
		t.Fatalf("ausgeschieden = %+v", danach.Ausgang.Ausgeschieden)
	}
}

// **Neun sind neun** — die Grenze steht bei zehn und nicht darunter.
func TestBeiNeunVergiftungsmarkenGeschiehtNichts(t *testing.T) {
	t.Parallel()

	partie := eineEndpartie()
	partie.Sitze[1].Vergiftung = 9

	if danach := partie.Zustandspruefung(); len(danach.Ausgang.Ausgeschieden) != 0 {
		t.Fatalf("ausgeschieden = %+v", danach.Ausgang.Ausgeschieden)
	}
}

// **Wer aus einer leeren Bibliothek ziehen wollte, verliert** — CR 704.5b.
//
// **Und zwar erst bei der Prüfung**: das Ziehen selbst hält nichts an, es
// vermerkt nur, dass es ins Leere ging.
func TestWerAusLeererBibliothekZiehenWollteScheidetAus(t *testing.T) {
	t.Parallel()

	partie := eineEndpartie()
	partie.Sitze[1].Leergezogen = true

	danach := partie.Zustandspruefung()

	if len(danach.Ausgang.Ausgeschieden) != 1 ||
		danach.Ausgang.Ausgeschieden[0].Grund != domain.GrundBibliothek {
		t.Fatalf("ausgeschieden = %+v", danach.Ausgang.Ausgeschieden)
	}
}

// **Zwei bei null Leben sind unentschieden** — CR 704.4 mit CR 104.4a.
func TestZweiBeiNullLebenGebenUnentschieden(t *testing.T) {
	t.Parallel()

	partie := eineEndpartie()
	partie.Sitze[0].Leben = 0
	partie.Sitze[1].Leben = -3

	danach := partie.Zustandspruefung()

	if !danach.Ausgang.Unentschieden || danach.Ausgang.Sieger != "" {
		t.Fatalf("ausgang = %+v", danach.Ausgang)
	}
}

// **Eine Prüfung ohne Anlass ändert nichts** — und gibt dieselbe Partie.
func TestOhneAnlassBleibtDiePartieWieSieWar(t *testing.T) {
	t.Parallel()

	partie := eineEndpartie()

	danach := partie.Zustandspruefung()

	if danach.Stand != partie.Stand || len(danach.Ausgang.Ausgeschieden) != 0 {
		t.Fatalf("stand = %q, ausgang = %+v", danach.Stand, danach.Ausgang)
	}
}

// **Ein Sitz, der schon draußen ist, scheidet nicht noch einmal aus.**
//
// Sein Leben bleibt bei null stehen — ohne diese Bedingung meldete jede
// weitere Prüfung denselben Sitz erneut.
func TestWerSchonDraussenIstScheidetNichtNochEinmalAus(t *testing.T) {
	t.Parallel()

	partie := eineEndpartie()
	partie.Sitze = append(partie.Sitze, domain.Sitz{Kennung: "sitz-3", Leben: 20})
	partie.Sitze[1].Leben = 0

	einmal := partie.Zustandspruefung()
	zweimal := einmal.Zustandspruefung()

	if len(zweimal.Ausgang.Ausgeschieden) != 1 {
		t.Fatalf("ausgeschieden = %+v", zweimal.Ausgang.Ausgeschieden)
	}
}

// **Das Ziehen ins Leere vermerkt sich** — CR 121.1 mit CR 704.5b.
func TestZiehenAusLeererBibliothekWirdVermerkt(t *testing.T) {
	t.Parallel()

	partie := eineEndpartie()
	partie.Objekte = nil
	partie.Zug = domain.Zugstand{
		Nummer: 2, Aktiv: "sitz-2", Starter: "sitz-1",
		Abschnitt: domain.Abschnitt{
			Phase: domain.PhaseAnfang, Schritt: domain.SchrittVersorgung,
		},
	}

	// **Aus der Versorgung heraus**, damit der Ziehschritt **betreten** wird
	// — die zugbasierten Handlungen laufen beim Betreten (CR 703.3).
	danach := partie.Weiter()

	if !danach.Sitze[1].Leergezogen {
		t.Fatalf("sitz-2 hat nicht vermerkt, dass die Bibliothek leer war")
	}
}

// **Ein Zug für jemanden, der nicht am Tisch sitzt, vermerkt nichts.**
//
// Er entsteht nur aus einem von Hand geänderten Bestand — `Zug.Aktiv` nennt
// dann eine Kennung, die es nicht gibt. **Der Zweig steht trotzdem da**, und
// ein Zweig ohne Lauf ist eine Zusage ohne Prüfung (§3).
func TestEinZugOhneSitzVermerktNichts(t *testing.T) {
	t.Parallel()

	partie := eineEndpartie()
	partie.Objekte = nil
	partie.Zug = domain.Zugstand{
		Nummer: 2, Aktiv: "geist", Starter: "sitz-1",
		Abschnitt: domain.Abschnitt{
			Phase: domain.PhaseAnfang, Schritt: domain.SchrittVersorgung,
		},
	}

	danach := partie.Weiter()

	for _, sitz := range danach.Sitze {
		if sitz.Leergezogen {
			t.Fatalf("%q hat einen Vermerk bekommen", sitz.Kennung)
		}
	}
}

// **Und zweimal ins Leere zu greifen vermerkt es nur einmal.**
//
// Die Gegenprobe zum Lauf darüber: hier **gibt** es den Sitz, und der zweite
// Griff läuft durch denselben Zweig wie der erste — er findet den Vermerk
// schon vor.
func TestZweimalInsLeereVermerktEinmal(t *testing.T) {
	t.Parallel()

	partie := eineEndpartie()
	partie.Objekte = nil
	partie.Sitze[0].Leergezogen = true
	partie.Zug = domain.Zugstand{
		Nummer: 3, Aktiv: "sitz-1", Starter: "sitz-1",
		Abschnitt: domain.Abschnitt{
			Phase: domain.PhaseAnfang, Schritt: domain.SchrittVersorgung,
		},
	}

	if danach := partie.Weiter(); !danach.Sitze[0].Leergezogen {
		t.Fatal("der Vermerk ist verlorengegangen")
	}
}
