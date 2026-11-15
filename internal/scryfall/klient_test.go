package scryfall_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// **Die vier Fälle der Phasenabnahme**
// (#888):
//
// > Bei 403, 429, Zeitüberschreitung **und ohne Netz** bleibt das Bearbeiten
// > unverändert möglich — gefahren, je Fall ein Test.
//
// # Was „unverändert möglich" hier heißt
//
// Der Aufruf **scheitert** — und er scheitert so, dass der Aufrufer
// weiterarbeiten kann: mit einem Fehler, den er einordnen kann, und ohne dass
// etwas im Bestand kaputtgeht. Liegt ein alter Stand vor, kommt der; das ist
// der Fall darunter.
func TestDieVierAusfaelleHaltenDasBearbeitenNichtAn(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(http.ResponseWriter, *http.Request){
		"403 — die Bedingungen verletzt": func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusForbidden)
		},
		"429 — zu viele Anfragen": func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.Header().Set("Retry-After", "30")
			schreiben.WriteHeader(http.StatusTooManyRequests)
		},
		"500 — die Gegenstelle hat ein Problem": func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusInternalServerError)
		},
	}

	for name, verhalten := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gegenstelle := httptest.NewServer(http.HandlerFunc(verhalten))
			t.Cleanup(gegenstelle.Close)

			klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

			_, _, _, fehler := klient.KarteNachName(t.Context(), "Prüfstein")
			if fehler == nil {
				t.Fatal("der Ausfall kam als Erfolg zurück")
			}
		})
	}
}

// **Eine Zeitüberschreitung ebenso — und schnell.**
//
// # Warum das ein eigener Fall ist
//
// Ein Server, der **schweigt**, ist kein Statuscode. Ohne die Stillstandsfrist
// aus `internal/netz` bliebe der Aufruf hängen, und das Fenster mit ihm.
func TestEineZeitueberschreitungHaeltDasBearbeitenNichtAn(t *testing.T) {
	t.Parallel()

	losgelassen := make(chan struct{})
	t.Cleanup(func() { close(losgelassen) })

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			schreiben.Header().Set("Content-Length", "1000")
			schreiben.WriteHeader(http.StatusOK)

			if flusher, geht := schreiben.(http.Flusher); geht {
				flusher.Flush()
			}

			select {
			case <-losgelassen:
			case <-anfrage.Context().Done():
			}
		}))
	t.Cleanup(gegenstelle.Close)

	kurz := netz.NeuerKlient(netz.Fristen{
		Verbindung: time.Second, Kopf: time.Second, Stillstand: 200 * time.Millisecond,
	})

	klient := klientFuer(t, gegenstelle.URL, kurz)

	los := time.Now()

	if _, _, _, fehler := klient.KarteNachName(t.Context(), "Prüfstein"); fehler == nil {
		t.Fatal("der schweigende Server kam als Erfolg zurück")
	}

	if gedauert := time.Since(los); gedauert > 5*time.Second {
		t.Errorf("es dauerte %v — die Frist greift nicht", gedauert)
	}
}

// **Und ohne Netz kommt der letzte gute Stand.**
//
// # Die eigentliche Zusage von §16, Regel 5
//
// „Ohne Netz arbeitet die Anwendung weiter. Ein Dienst, dessen Ausfall das
// Bearbeiten anhält, ist falsch eingebaut. Der letzte gute Stand bleibt, und
// er sagt, wie alt er ist."
func TestOhneNetzKommtDerLetzteGuteStand(t *testing.T) {
	t.Parallel()

	erreichbar := true

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if !erreichbar {
				schreiben.WriteHeader(http.StatusInternalServerError)

				return
			}

			schreibe(t, schreiben, beispiel(t, "einfach.json"))
		}))
	t.Cleanup(gegenstelle.Close)

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	speicher := netz.NeuerSpeicher(t.TempDir(), scryfall.Speicherfrist, func() time.Time { return jetzt })

	klient := scryfall.NeuerKlient(gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()), speicher,
		func() time.Time { return jetzt }, "prüfung")

	if _, _, _, fehler := klient.KarteNachName(t.Context(), "Prüfstein"); fehler != nil {
		t.Fatalf("das erste Holen: %v", fehler)
	}

	// Zwei Tage später, und das Netz ist weg.
	jetzt = jetzt.Add(48 * time.Hour)
	erreichbar = false

	karte, alter, _, fehler := klient.KarteNachName(t.Context(), "Prüfstein")
	if fehler != nil {
		t.Fatalf("ohne Netz scheiterte der Aufruf: %v", fehler)
	}

	if karte.Name != "Prüfstein" {
		t.Errorf("es kam %q", karte.Name)
	}

	// **Und er sagt, wie alt er ist.**
	if alter != 48*time.Hour {
		t.Errorf("Alter %v, erwartet waren 48h", alter)
	}
}

// **Ein 404 ist kein Ausfall, sondern eine Antwort.**
//
// # Warum das der wichtigste Unterschied dieses Klienten ist
//
// Ein Ausfall darf durch einen alten Stand ersetzt werden; ein fachliches Nein
// nicht. Wer „diese Karte gibt es nicht" mit dem Stand einer **anderen** Karte
// beantwortet, gibt eine falsche Auskunft — und der Aufrufer merkt es nicht.
func TestEin404IstKeinAusfall(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusNotFound)
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	_, _, _, fehler := klient.KarteNachName(t.Context(), "GibtsNicht")
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler %v ist kein %v", fehler, domain.ErrNotFound)
	}
}

// **Innerhalb der Frist wird nicht gefragt.**
//
// Das ist die Zusage an die Gegenstelle: „cache […] at least for 24 hours".
func TestInnerhalbDerFristWirdNichtGefragt(t *testing.T) {
	t.Parallel()

	gefragt := 0

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gefragt++
			schreibe(t, schreiben, beispiel(t, "einfach.json"))
		}))
	t.Cleanup(gegenstelle.Close)

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)
	speicher := netz.NeuerSpeicher(t.TempDir(), scryfall.Speicherfrist, func() time.Time { return jetzt })

	klient := scryfall.NeuerKlient(gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()), speicher,
		func() time.Time { return jetzt }, "prüfung")

	for range 5 {
		if _, _, _, fehler := klient.KarteNachName(t.Context(), "Prüfstein"); fehler != nil {
			t.Fatalf("holen: %v", fehler)
		}
	}

	if gefragt != 1 {
		t.Errorf("Scryfall wurde %dmal gefragt, erwartet war einmal", gefragt)
	}
}

// **Die Anfrage nennt sich, und der Weg stimmt.**
func TestDieAnfrageNenntSichUndTrifftDenWeg(t *testing.T) {
	t.Parallel()

	var gesehen *http.Request

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			gesehen = anfrage.Clone(anfrage.Context())
			schreibe(t, schreiben, beispiel(t, "einfach.json"))
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	if _, _, _, fehler := klient.KarteNachName(t.Context(), "Lightning Bolt"); fehler != nil {
		t.Fatalf("holen: %v", fehler)
	}

	if gesehen.URL.Path != "/cards/named" {
		t.Errorf("Weg %q", gesehen.URL.Path)
	}

	// **Der Name geht entwertet mit.** „Lightning Bolt" hat ein Leerzeichen.
	if genau := gesehen.URL.Query().Get("exact"); genau != "Lightning Bolt" {
		t.Errorf("exact = %q", genau)
	}

	if kenner := gesehen.Header.Get("User-Agent"); kenner == "" ||
		!containsAll(kenner, "MagicSetEditor4Mac", "github.com") {
		t.Errorf("der User-Agent nennt die Anwendung nicht: %q", kenner)
	}

	if gesehen.Header.Get("Accept") == "" {
		t.Error("der Accept fehlt — er ist Pflicht")
	}
}

// klientFuer baut einen Klienten über einer Gegenstelle.
func klientFuer(t *testing.T, adresse string, ueber *netz.Klient) *scryfall.Klient {
	t.Helper()

	jetzt := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)

	return scryfall.NeuerKlient(adresse, ueber,
		netz.NeuerSpeicher(t.TempDir(), scryfall.Speicherfrist, func() time.Time { return jetzt }),
		func() time.Time { return jetzt }, "prüfung")
}

// klientOhneWarten gibt einen Klienten, dessen Uhr bei jedem Blick um eine
// Sekunde springt.
//
// # Wofür
//
// Die Drossel wartet **echt** (`schlafe` ist im Klienten `nil`), und sie liest
// die Zeit über die eingesetzte Uhr. Mit einer stehenden Uhr ist jeder Abstand
// immer unerfüllt: ein Lauf über vierzig Seiten kostete vierzig mal 600 ms.
//
// **Die Drossel bleibt an.** Sie wird nicht abgeschaltet, sondern mit einer
// Uhr bedient, für die genug Zeit vergangen ist — der Weg durch `Warte` wird
// also gegangen, er kostet nur nichts.
func klientOhneWarten(t *testing.T, adresse string, ueber *netz.Klient) *scryfall.Klient {
	t.Helper()

	beginn := time.Date(2026, time.August, 23, 12, 0, 0, 0, time.UTC)

	var schritte atomic.Int64

	uhr := func() time.Time {
		return beginn.Add(time.Duration(schritte.Add(1)) * time.Second)
	}

	return scryfall.NeuerKlient(adresse, ueber,
		netz.NeuerSpeicher(t.TempDir(), scryfall.Speicherfrist, uhr), uhr, "prüfung")
}

// containsAll sagt, ob alle Teile vorkommen.
func containsAll(text string, teile ...string) bool {
	for _, teil := range teile {
		if !strings.Contains(text, teil) {
			return false
		}
	}

	return true
}

// **Eine Suche, die scheitert, gibt den Fehler weiter.**
//
// Ohne diesen Fall bliebe der Zweig unbetreten, der bei einer Suche einen
// Ausfall meldet — und die vier Fälle der Abnahme gälten nur für das Holen
// einer einzelnen Karte.
func TestEineSucheDieScheitertGibtDenFehlerWeiter(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(http.StatusForbidden)
		}))
	t.Cleanup(gegenstelle.Close)

	klient := klientFuer(t, gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()))

	if _, _, _, fehler := klient.Suche(t.Context(), "t:creature"); fehler == nil {
		t.Error("die gescheiterte Suche kam als Erfolg zurück")
	}
}

// **Zwischen zwei Anfragen liegt der Mindestabstand.**
//
// # Warum dieser Test sein muss
//
// Ohne ihn belegte **nichts**, dass die Drossel überhaupt eingehängt ist.
// Nachgestellt am 2026-08-23: sie durch `nil` zu ersetzen ließ jeden Test
// grün — die Zusage „gedrosselt wird unter der genannten Zahl" hing an einer
// Zeile, die niemand prüfte.
//
// # Was er misst, und was das kostet
//
// **Echte Zeit**, denn die Drossel dieses Klienten wartet echt: eine
// eingesetzte Uhr allein genügt nicht, weil `Warte` ohne eingesetzten Schlaf
// einen Zeitgeber benutzt. Der Lauf kostet dadurch gut eine halbe Sekunde —
// das ist der Preis dafür, dass die Zusage geprüft ist statt behauptet.
//
// **Zwei verschiedene Namen**, damit der Zwischenspeicher die zweite Anfrage
// nicht beantwortet.
func TestZwischenZweiAnfragenLiegtDerMindestabstand(t *testing.T) {
	t.Parallel()

	gegenstelle := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreibe(t, schreiben, beispiel(t, "einfach.json"))
		}))
	t.Cleanup(gegenstelle.Close)

	// **Mit echter Uhr**, denn gemessen wird echte Zeit.
	klient := scryfall.NeuerKlient(gegenstelle.URL, netz.KlientUeber(gegenstelle.Client()),
		netz.NeuerSpeicher(t.TempDir(), scryfall.Speicherfrist, time.Now),
		time.Now, "prüfung")

	los := time.Now()

	for _, name := range []string{"Erster", "Zweiter"} {
		if _, _, _, fehler := klient.KarteNachName(t.Context(), name); fehler != nil {
			t.Fatalf("holen (%s): %v", name, fehler)
		}
	}

	// **Die erlaubte Zahl ist 2/s, also 500 ms.** Gewartet wird darunter — bei
	// 600 ms. Geprüft wird gegen 400, damit eine ungenaue Uhr den Lauf nicht
	// flackern lässt; wer gar nicht drosselt, liegt weit darunter.
	if gedauert := time.Since(los); gedauert < 400*time.Millisecond {
		t.Errorf("zwei Anfragen dauerten %v — die Drossel greift nicht", gedauert)
	}
}
