package flaechen_test

// **Der Dienst tauscht seinen Bestand nicht — auch nicht beim Neuanfang** —
// G09 (#2273).
//
// Der Dienst hält eine Verbindung über seine ganze Laufzeit, und jede seiner
// Nähte hat sie sich beim Aufbau geholt — deshalb spielt er nicht zurück und
// schaltet die Verschlüsselung nicht um (`internal/rest/bestandssicherung.go`).
// **Bis G09 nahm `POST /ruecksetzung/anwendung` trotzdem den Neuanfang**: er
// schloss diese Verbindung, legte den Bestand beiseite und öffnete einen
// frischen, den nur die Rücksetzung selbst kannte. Jede Anfrage danach traf
// eine geschlossene Verbindung.
//
// # Was er hält
//
//   - Die Vorschau sagt vorher, dass es hier nicht geht — mit dem Hindernis
//     `ohneTausch`.
//   - Zwei gleichzeitige Neuanfänge werden **beide** abgewiesen, bevor etwas
//     angefasst ist: der Bestand liegt, wo er lag, und **kein Nachweis ist
//     verbraucht** — ein Mensch hat ihn am Fenster für etwas gegeben, das
//     hier nicht geschah.
//   - Der Dienst antwortet danach wie davor.
//
// # Was er nicht hält
//
// **Einen Tausch durch ein anderes Programm**, während der Dienst läuft —
// den weist `AlleinAmBestand` ab, und das hält
// `TestGeschaltetWirdNichtWaehrendEinAndererDenBestandHaelt`.

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// antwortAm ist, was ein Ruf aus einer eigenen Goroutine zurückbringt — dort
// darf ein Lauf nicht `t.Fatal` rufen.
type antwortAm struct {
	status int
	rumpf  string
	fehler error
}

// rufNebenher ruft den Dienst ohne `t`, damit mehrere Rufe zugleich gehen.
func rufNebenher(ctx context.Context, an dienst, methode, weg, rumpf string) antwortAm {
	anfrage, fehler := http.NewRequestWithContext(ctx, methode, an.adresse+weg, strings.NewReader(rumpf))
	if fehler != nil {
		return antwortAm{fehler: fehler}
	}

	anfrage.Header.Set("Authorization", "Bearer "+an.marke)

	kam, fehler := an.klient.Do(anfrage)
	if fehler != nil {
		return antwortAm{fehler: fehler}
	}

	defer func() { _ = kam.Body.Close() }() //nolint:errcheck // nur gelesen

	roh, fehler := io.ReadAll(kam.Body)

	return antwortAm{status: kam.StatusCode, rumpf: string(roh), fehler: fehler}
}

// zweiNachweiseFuerDenNeuanfang stellt sie aus, **wie das Fenster es täte**:
// über den Bestand, im gezeigten Profil, als Mensch.
func zweiNachweiseFuerDenNeuanfang(t *testing.T, ort string) [2]string {
	t.Helper()

	db, fehler := storage.Oeffnen(t.Context(), ort)
	if fehler != nil {
		t.Fatalf("den Bestand öffnen: %v", fehler)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur zum Aufräumen

	ctx, profilfehler := storage.MitGezeigtemProfil(t.Context(), db)
	if profilfehler != nil {
		t.Fatalf("das Profil lesen: %v", profilfehler)
	}

	var kennungen [2]string

	for stelle := range kennungen {
		nachweis, ausstellfehler := app.AutorisierungAus(db).Ausstellen(ctx, service.Nachweisantrag{
			Zweck: domain.RuecksetzungsZweck,
			Ziel:  string(domain.RuecksetzungAnwendung),
			Frist: time.Hour,
		})
		if ausstellfehler != nil {
			t.Fatalf("den Nachweis ausstellen: %v", ausstellfehler)
		}

		kennungen[stelle] = nachweis.Kennung
	}

	return kennungen
}

// unverbraucht sagt, ob kein Nachweis eingelöst ist.
func unverbraucht(t *testing.T, ort string, kennungen [2]string) {
	t.Helper()

	db, fehler := storage.OeffnenNurLesend(t.Context(), ort)
	if fehler != nil {
		t.Fatalf("den Bestand lesen: %v", fehler)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur zum Aufräumen

	for _, kennung := range kennungen {
		nachweis, lesefehler := storage.NeueNachweisablage(db).Nachweis(t.Context(), kennung)
		if lesefehler != nil {
			t.Errorf("den Nachweis %s lesen: %v", kennung, lesefehler)

			continue
		}

		if nachweis.Eingeloest {
			t.Errorf("der Nachweis %s ist verbraucht, obwohl nichts geschah", kennung)
		}
	}
}

func TestDerDienstWeistDenNeuanfangAbBevorErEtwasAnfasst(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	an := starteDienst(t, umgebung)
	ort := bestandsortAus(t, umgebung)

	// **Die Gegenprobe zum Nein, vorher**: der Dienst liest.
	mussREST(t, an, http.MethodGet, "/sets", "")

	kennungen := zweiNachweiseFuerDenNeuanfang(t, ort)

	// **Die Vorschau sagt es vorher.**
	roh := mussREST(t, an, http.MethodGet, "/ruecksetzung/anwendung/vorschau", "")

	var vorschau struct {
		Bereit      bool   `json:"bereit"`
		Revision    string `json:"revision"`
		Hindernisse []struct {
			Art string `json:"art"`
		} `json:"hindernisse"`
	}

	if fehler := json.Unmarshal([]byte(roh), &vorschau); fehler != nil {
		t.Fatalf("die Vorschau lesen: %v (%s)", fehler, roh)
	}

	if vorschau.Bereit || len(vorschau.Hindernisse) != 1 ||
		vorschau.Hindernisse[0].Art != domain.RuecksetzungOhneTausch {
		t.Errorf("die Vorschau sagt nicht, dass der Dienst nicht tauscht: %s", roh)
	}

	// **Zwei zugleich**, jede mit ihrem eigenen Nachweis.
	var (
		warte     sync.WaitGroup
		antworten [2]antwortAm
	)

	for stelle, kennung := range kennungen {
		warte.Go(func() {
			antworten[stelle] = rufNebenher(t.Context(), an, http.MethodPost, "/ruecksetzung/anwendung",
				`{"nachweis":"`+kennung+`","revision":"`+vorschau.Revision+`"}`)
		})
	}

	warte.Wait()

	for stelle, antwort := range antworten {
		switch {
		case antwort.fehler != nil:
			t.Errorf("Neuanfang %d: %v", stelle, antwort.fehler)
		case antwort.status != http.StatusConflict ||
			!strings.Contains(antwort.rumpf, `"code":"konflikt"`) ||
			!strings.Contains(antwort.rumpf, "Kommandozeile"):
			t.Errorf("Neuanfang %d: %d %s — erwartet die Abweisung, die sagt, wo es geht",
				stelle, antwort.status, antwort.rumpf)
		}
	}

	// **Nichts angefasst**: kein beiseitegelegter Bestand, kein Nachweis
	// verbraucht.
	if _, fehler := os.Stat(filepath.Join(filepath.Dir(ort),
		service.Rueckspielsicherung)); !errors.Is(fehler, os.ErrNotExist) {
		t.Errorf("neben dem Bestand liegt ein beiseitegelegter: %v", fehler)
	}

	unverbraucht(t, ort, kennungen)

	// **Und der Dienst liest danach wie davor.**
	mussREST(t, an, http.MethodGet, "/sets", "")
}
