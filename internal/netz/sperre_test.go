// **Ein 429 sperrt wirklich, und die Drossel ist im Prozess geteilt** —
// Befund 2 des Zweitblicks über Phase V, 2026-08-24.
//
// # Der Befund
//
// Zwei Hälften, und beide waren belegt:
//
//   - **Vier Klienten im Fenster**, jeder mit eigenem Drosselzustand; MCP und
//     REST bauten die Naht **je Aufruf** neu. Weil der erste Aufruf einer
//     Drossel absichtlich nicht wartet, sah eine solche Drossel praktisch nie
//     einen zweiten — die Grenze war damit wirkungslos.
//   - **Ein 429 erzeugte keinen Sperrzustand.** `Retry-After` wurde in einen
//     Meldungstext umgerechnet und sonst nichts; der nächste Aufruf ging
//     sofort wieder hinaus. `DRITTANBIETER.md` behauptete dagegen „429 sperrt
//     30 Sekunden" — eine Zusage, die der Code nicht hielt.
//
// # Warum das schwer wiegt
//
// Scryfall schreibt: „It is not acceptable to ignore HTTP 429 responses", und
// wiederholte 429 können zur dauerhaften Sperre führen. Ein Budget, das für
// **Host und IP** gilt, als kurzlebigen Zustand eines Adapters zu führen ist
// dieselbe Klasse wie eine Drossel, die nie greift.
package netz_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// uhrwerk ist eine Uhr, die man stellen kann.
type uhrwerk struct{ jetzt time.Time }

func (u *uhrwerk) lies() time.Time { return u.jetzt }

func (u *uhrwerk) weiter(dauer time.Duration) { u.jetzt = u.jetzt.Add(dauer) }

// **Ein Schlaf, der die Uhr weiterstellt** — sonst wartete jeder Test wirklich.
func (u *uhrwerk) schlafe(dauer time.Duration) { u.weiter(dauer) }

func TestEineSperreHaeltDieDrosselAn(t *testing.T) {
	t.Parallel()

	uhr := &uhrwerk{jetzt: time.Date(2026, time.August, 24, 12, 0, 0, 0, time.UTC)}
	drossel := netz.NeueDrossel(100*time.Millisecond, uhr.lies, uhr.schlafe)

	// **Der erste Aufruf geht sofort durch.**
	if fehler := drossel.Warte(t.Context()); fehler != nil {
		t.Fatalf("Warte: %v", fehler)
	}

	begonnen := uhr.jetzt

	// **Jetzt kommt ein 429 mit 30 Sekunden.**
	drossel.Sperre(uhr.jetzt.Add(30 * time.Second))

	if fehler := drossel.Warte(t.Context()); fehler != nil {
		t.Fatalf("Warte: %v", fehler)
	}

	// **Die Uhr ist um 30 Sekunden weiter, nicht um 100 Millisekunden.**
	if gewartet := uhr.jetzt.Sub(begonnen); gewartet < 30*time.Second {
		t.Errorf("gewartet wurde %v, erwartet mindestens 30s — ein 429 sperrt",
			gewartet)
	}
}

// **Eine Sperre, deren Zeit um ist, hält nichts mehr an.**
//
// Ohne diesen Fall wäre nur belegt, dass die Sperre **wirkt** — nicht, dass
// sie endet. Eine Sperre ohne Ende machte die Anwendung nach einem 429
// dauerhaft stumm.
func TestEineAbgelaufeneSperreHaeltNichtsAn(t *testing.T) {
	t.Parallel()

	uhr := &uhrwerk{jetzt: time.Date(2026, time.August, 24, 12, 0, 0, 0, time.UTC)}
	drossel := netz.NeueDrossel(0, uhr.lies, uhr.schlafe)

	drossel.Sperre(uhr.jetzt.Add(30 * time.Second))
	uhr.weiter(31 * time.Second)

	begonnen := uhr.jetzt

	if fehler := drossel.Warte(t.Context()); fehler != nil {
		t.Fatalf("Warte: %v", fehler)
	}

	if gewartet := uhr.jetzt.Sub(begonnen); gewartet != 0 {
		t.Errorf("gewartet wurde %v, erwartet 0", gewartet)
	}
}

// **Die längere von zwei Sperren gilt.**
//
// Zwei Anfragen können gleichzeitig ein 429 bekommen, mit verschiedenen
// `Retry-After`. Die kürzere zuletzt zu nehmen hieße, die längere zu
// unterlaufen.
func TestDieLaengereSperreGilt(t *testing.T) {
	t.Parallel()

	uhr := &uhrwerk{jetzt: time.Date(2026, time.August, 24, 12, 0, 0, 0, time.UTC)}
	drossel := netz.NeueDrossel(0, uhr.lies, uhr.schlafe)

	drossel.Sperre(uhr.jetzt.Add(60 * time.Second))
	drossel.Sperre(uhr.jetzt.Add(10 * time.Second))

	begonnen := uhr.jetzt

	if fehler := drossel.Warte(t.Context()); fehler != nil {
		t.Fatalf("Warte: %v", fehler)
	}

	if gewartet := uhr.jetzt.Sub(begonnen); gewartet < 60*time.Second {
		t.Errorf("gewartet wurde %v, erwartet 60s — die kürzere Sperre hat "+
			"die längere unterlaufen", gewartet)
	}
}

// **Ein Drosselwerk gibt für denselben Schlüssel dieselbe Drossel.**
//
// # Warum es das gibt
//
// Das Budget gilt für **Host und IP**, nicht für eine Instanz in diesem
// Programm. Wer vier Klienten baut, hat trotzdem ein Budget — und vier
// Drosseln, die einander nicht kennen, halten es nicht ein.
func TestEinDrosselwerkTeiltDieDrossel(t *testing.T) {
	t.Parallel()

	uhr := &uhrwerk{jetzt: time.Date(2026, time.August, 24, 12, 0, 0, 0, time.UTC)}
	werk := netz.NeuesDrosselwerk(uhr.lies, uhr.schlafe)

	erste := werk.Fuer("api.scryfall.com/karten", 500*time.Millisecond)
	zweite := werk.Fuer("api.scryfall.com/karten", 500*time.Millisecond)

	if erste != zweite {
		t.Fatal("zwei Aufrufe gaben zwei Drosseln — dann ist das Budget geteilt")
	}

	// **Und ein anderer Schlüssel gibt eine andere.**
	andere := werk.Fuer("api.scryfall.com/allgemein", 100*time.Millisecond)
	if andere == erste {
		t.Error("zwei Endpunktklassen bekamen dieselbe Drossel")
	}
}

// **Zwei Klienten aus demselben Werk halten den Abstand gemeinsam ein.**
//
// Das ist die Zusage, um die es geht: nicht dass eine Drossel geteilt **ist**,
// sondern dass der Abstand dadurch wirklich eingehalten wird.
func TestZweiHalterHaltenDenAbstandGemeinsam(t *testing.T) {
	t.Parallel()

	uhr := &uhrwerk{jetzt: time.Date(2026, time.August, 24, 12, 0, 0, 0, time.UTC)}
	werk := netz.NeuesDrosselwerk(uhr.lies, uhr.schlafe)

	erste := werk.Fuer("gemeinsam", 500*time.Millisecond)
	zweite := werk.Fuer("gemeinsam", 500*time.Millisecond)

	if fehler := erste.Warte(t.Context()); fehler != nil {
		t.Fatalf("Warte: %v", fehler)
	}

	begonnen := uhr.jetzt

	// **Der zweite Halter wartet**, obwohl er selbst noch nie gefragt hat.
	if fehler := zweite.Warte(t.Context()); fehler != nil {
		t.Fatalf("Warte: %v", fehler)
	}

	if gewartet := uhr.jetzt.Sub(begonnen); gewartet < 500*time.Millisecond {
		t.Errorf("gewartet wurde %v, erwartet 500ms — der zweite Halter kannte "+
			"den ersten nicht", gewartet)
	}
}

// **Ein 429 setzt die Sperre — und die nächste Anfrage wartet sie ab.**
//
// # Warum das die eigentliche Zusage ist
//
// Die Läufe darüber prüfen den Mechanismus. Dieser prüft, dass er einen
// **Anlass** hat: vor dem 2026-08-24 wurde `Retry-After` in einen Meldungstext
// umgerechnet und sonst nichts, und der nächste Aufruf ging sofort wieder
// hinaus. `DRITTANBIETER.md` behauptete dagegen „429 sperrt 30 Sekunden".
func TestEin429SetztDieSperre(t *testing.T) {
	t.Parallel()

	var zeiten []time.Time

	uhr := &uhrwerk{jetzt: time.Date(2026, time.August, 24, 12, 0, 0, 0, time.UTC)}

	erste := true

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			zeiten = append(zeiten, uhr.jetzt)

			if erste {
				erste = false

				schreiben.Header().Set("Retry-After", "30")
				schreiben.WriteHeader(http.StatusTooManyRequests)

				return
			}

			if _, fehler := schreiben.Write([]byte("gut")); fehler != nil {
				t.Errorf("schreiben: %v", fehler)
			}
		}))
	t.Cleanup(server.Close)

	drossel := netz.NeueDrossel(0, uhr.lies, uhr.schlafe)
	klient := netz.KlientUeber(server.Client())
	kenner := netz.Kenner{Name: "prüfung", Annahme: "*/*"}

	// **Der erste Aufruf bekommt das 429.**
	_, fehler := netz.Hole(t.Context(), klient, drossel, kenner, server.URL, 1<<20)
	if !errors.Is(fehler, netz.ErrGedrosselt) {
		t.Fatalf("Fehler = %v, erwartet ErrGedrosselt", fehler)
	}

	// **Der zweite wartet die Sperre ab.**
	if _, zweiter := netz.Hole(
		t.Context(), klient, drossel, kenner, server.URL, 1<<20); zweiter != nil {
		t.Fatalf("zweiter Aufruf: %v", zweiter)
	}

	if len(zeiten) != 2 {
		t.Fatalf("%d Anfragen", len(zeiten))
	}

	if abstand := zeiten[1].Sub(zeiten[0]); abstand < 30*time.Second {
		t.Errorf("Abstand = %v, erwartet mindestens 30s — das 429 hat nichts "+
			"gesperrt", abstand)
	}
}

// **Ein 429 ohne `Retry-After` sperrt trotzdem.**
//
// Die Kopfzeile ist nicht Pflicht. Ohne Sperre ginge der nächste Aufruf sofort
// wieder hinaus — und das ist genau das Übergehen, das die Bedingungen
// verbieten. Genommen wird dann die Zahl, die Scryfall selbst nennt.
func TestEin429OhneKopfzeileSperrtTrotzdem(t *testing.T) {
	t.Parallel()

	var zeiten []time.Time

	uhr := &uhrwerk{jetzt: time.Date(2026, time.August, 24, 12, 0, 0, 0, time.UTC)}

	erste := true

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			zeiten = append(zeiten, uhr.jetzt)

			if erste {
				erste = false

				schreiben.WriteHeader(http.StatusTooManyRequests)

				return
			}

			if _, fehler := schreiben.Write([]byte("gut")); fehler != nil {
				t.Errorf("schreiben: %v", fehler)
			}
		}))
	t.Cleanup(server.Close)

	drossel := netz.NeueDrossel(0, uhr.lies, uhr.schlafe)
	klient := netz.KlientUeber(server.Client())
	kenner := netz.Kenner{Name: "prüfung", Annahme: "*/*"}

	// **Der erste Aufruf bekommt das 429** — erwartet, deshalb geprüft und
	// nicht verschluckt.
	if _, fehler := netz.Hole(
		t.Context(), klient, drossel, kenner, server.URL, 1<<20,
	); !errors.Is(fehler, netz.ErrGedrosselt) {
		t.Fatalf("Fehler = %v, erwartet ErrGedrosselt", fehler)
	}

	if _, fehler := netz.Hole(
		t.Context(), klient, drossel, kenner, server.URL, 1<<20,
	); fehler != nil {
		t.Fatalf("zweiter Aufruf: %v", fehler)
	}

	if len(zeiten) != 2 {
		t.Fatalf("%d Anfragen", len(zeiten))
	}

	if abstand := zeiten[1].Sub(zeiten[0]); abstand < netz.SperreOhneAngabe {
		t.Errorf("Abstand = %v, erwartet mindestens %v", abstand, netz.SperreOhneAngabe)
	}
}

// **Eine Sperre gibt auf, wenn der Kontext endet.**
//
// # Warum das dazugehört
//
// Eine Sperre von dreißig Sekunden hält den ganzen Vorgang an. Wer ⌃C drückt,
// hat gesagt, dass er nicht mehr will — dieselbe Überlegung wie beim Abstand
// (siehe [netz.Drossel.Warte]). Ohne diesen Zweig blieb er ungeprüft, und
// niemand wüsste, ob ein Abbruch dreißig Sekunden braucht.
func TestEineSperreGibtBeimAbbruchAuf(t *testing.T) {
	t.Parallel()

	uhr := &uhrwerk{jetzt: time.Date(2026, time.August, 24, 12, 0, 0, 0, time.UTC)}

	// **Echtes Warten**, damit der Abbruch etwas zu unterbrechen hat: mit
	// gestellter Uhr wäre die Sperre sofort um, und der Zweig bliebe ungeprüft.
	drossel := netz.NeueDrossel(0, uhr.lies, nil)
	drossel.Sperre(uhr.jetzt.Add(time.Hour))

	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	if fehler := drossel.Warte(ctx); fehler == nil {
		t.Error("der Abbruch während der Sperre blieb ohne Fehler")
	}
}

// **Ein 429 sperrt das ganze Werk, nicht nur seine Klasse** —
// #1037.
//
// # Was hier stand und was fehlte
//
// Der Lauf darüber sichert zu, dass zwei Endpunktklassen **verschiedene**
// Drosseln bekommen — und das ist richtig: ihre Abstände sind verschieden.
// Daraus folgte aber nicht, was `DRITTANBIETER.md` zusagt:
//
// > Bis dahin geht aus diesem Prozess **keine** Anfrage an Scryfall hinaus —
// > auch nicht aus einem anderen Zugang.
//
// **Ein 429 sperrte nur die Drossel des gescheiterten Aufrufs.** Ein
// unmittelbar folgender Rulings- oder Symbologieabruf ging weiter hinaus.
// Angestrichen vom Nachlauf über Phase V.
//
// **Zwei Zusagen, nicht eine**: die Abstände bleiben getrennt, die Sperre wird
// gemeinsam. Wer beides zusammenwürfe, bremste die allgemeinen Wege auf ein
// Fünftel dessen, was erlaubt ist.
func TestEineSperreGiltFuerDasGanzeWerk(t *testing.T) {
	t.Parallel()

	uhr := &uhrwerk{jetzt: time.Date(2026, time.August, 24, 12, 0, 0, 0, time.UTC)}
	werk := netz.NeuesDrosselwerk(uhr.lies, uhr.schlafe)

	karten := werk.Fuer("scryfall/karten", 600*time.Millisecond)
	allgemein := werk.Fuer("scryfall/allgemein", 125*time.Millisecond)

	// **Ein 429 auf einem Kartenweg.**
	karten.SperreFuer(30 * time.Second)

	if !allgemein.Gesperrt() {
		t.Error("die allgemeine Klasse ist nicht gesperrt — aus diesem Prozess " +
			"ginge weiter eine Anfrage an Scryfall hinaus")
	}

	// **Und umgekehrt.**
	zweites := netz.NeuesDrosselwerk(uhr.lies, uhr.schlafe)
	zweiteKarten := zweites.Fuer("scryfall/karten", 600*time.Millisecond)
	zweitesAllgemein := zweites.Fuer("scryfall/allgemein", 125*time.Millisecond)

	zweitesAllgemein.SperreFuer(30 * time.Second)

	if !zweiteKarten.Gesperrt() {
		t.Error("die Kartenklasse ist nicht gesperrt")
	}
}

// **Eine Drossel ohne Werk sperrt nur sich selbst.**
//
// Sie hat keinen Ort für mehr — und ihr einen unterzuschieben hieße, eine
// Zusage zu erfinden, die niemand geben kann.
func TestEineDrosselOhneWerkSperrtNurSichSelbst(t *testing.T) {
	t.Parallel()

	uhr := &uhrwerk{jetzt: time.Date(2026, time.August, 24, 12, 0, 0, 0, time.UTC)}

	einzeln := netz.NeueDrossel(600*time.Millisecond, uhr.lies, uhr.schlafe)
	einzeln.SperreFuer(30 * time.Second)

	if !einzeln.Gesperrt() {
		t.Error("die eigene Drossel ist nicht gesperrt")
	}
}
