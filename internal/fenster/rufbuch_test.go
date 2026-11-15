package fenster

// **Das Rufbuch zählt, welche Bindungen das Fenster ruft** — B01
// (#2220).
//
// Geprüft wird die Mitschrift selbst, ohne Wails: eine Anfrage, wie die
// Laufzeit sie an `/wails/runtime` schickt, eine Antwort, wie die Bindungen
// sie geben — und dass die Anfrage dabei unverändert ankommt.

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// bindungsruf baut eine Anfrage, wie die Laufzeit sie für einen Bindungsruf schickt.
func bindungsruf(t *testing.T, methode string) *http.Request {
	t.Helper()

	rumpf := `{"object":0,"method":0,"args":{"methodName":"` + bindungsvorsatz + methode + `","args":[]}}`

	return httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/wails/runtime", strings.NewReader(rumpf))
}

// antwortend ist ein Handler, der den Rumpf liest und mit `status` und
// `rumpf` antwortet — wie die Laufzeit nach einem Ruf.
func antwortend(t *testing.T, status int, rumpf string) http.Handler {
	t.Helper()

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		gelesen, fehler := io.ReadAll(req.Body)
		if fehler != nil || !bytes.Contains(gelesen, []byte("methodName")) {
			t.Errorf("der Rumpf kam nicht unverändert an: %q, %v", gelesen, fehler)
		}

		rw.WriteHeader(status)
		_, _ = rw.Write([]byte(rumpf)) //nolint:errcheck // ein Recorder scheitert nicht
	})
}

// langeAntwort ist eine Antwort, deren `fehler` erst hinter der Grenze steht,
// bis zu der das Rufbuch den Anfang liest.
func langeAntwort(fehler string) string {
	return `{"daten":"` + strings.Repeat("x", rufgrenze+100) + `","fehler":` + fehler + `}`
}

func TestDasRufbuchUnterscheidetDieAusgaenge(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name    string
		status  int
		rumpf   string
		ausgang string
	}{
		{"gelungen", http.StatusOK, `{"sets":[],"fehler":null}`, "gelungen"},
		{"ein fachlicher Fehler in der Antwort", http.StatusOK, `{"fehler":{"code":"validierung"}}`, "fachfehler"},
		{"eine Abweisung des Tors", http.StatusInternalServerError, `Sets: panic: mse4mac-fehler:{}`, "abgewiesen"},
		// **Was sich nicht lesen lässt, ist unbekannt** und kein Erfolg —
		// Zweitblick über Phase B, Befund 4. Jeder Rückgabewert kommt bei
		// Wails als JSON; ein Text ist eine Antwort, die niemand erwartet.
		{"eine Antwort ohne JSON", http.StatusOK, `ein Text`, "unbekannt"},
		{"eine Zeichenkette als Antwort", http.StatusOK, `"ein Name"`, "gelungen"},
		// **Und eine lange Antwort trägt ihren Fehler am Ende** — hinter der
		// Grenze, bis zu der der Anfang gelesen wird.
		{"eine lange Antwort mit Fehler am Ende", http.StatusOK, langeAntwort(`{"code":"validierung","schluessel":"v"}`), "fachfehler"},
		{"eine lange Antwort mit Abbruch am Ende", http.StatusOK, langeAntwort(`{"code":"abgebrochen","schluessel":"a"}`), "abgebrochen"},
		{"eine lange gelungene Antwort", http.StatusOK, langeAntwort(`null`), "gelungen"},
		// **Ein vollständiges Objekt ohne Fehlerfeld ist gelungen** — nur
		// eines, das nie schließt, ist unbekannt.
		{"eine lange Antwort ohne Fehlerfeld", http.StatusOK, `{"daten":"` + strings.Repeat("x", rufgrenze) + `"}`, "gelungen"},
		{"eine abgebrochene lange Antwort", http.StatusOK, `{"daten":"` + strings.Repeat("x", rufgrenze), "unbekannt"},
		// **Der Nachlauf über Phase B fand zwei Wege daran vorbei**: ein
		// verschachteltes Fehlerfeld nach dem echten, und ein Anfang, der
		// genau nach einem Feld endete.
		{"ein verschachteltes Fehlerfeld hinter dem echten", http.StatusOK, `{"daten":"` + strings.Repeat("x", rufgrenze) + `","fehler":{"code":"validierung","schluessel":"v"},"weitere":{"fehler":null}}`, "fachfehler"},
		{"ein Fehler hinter einem Feld, das genau an der Grenze endet", http.StatusOK, `{"daten":"` + strings.Repeat("x", rufgrenze-len(`{"daten":""`)) + `","fehler":{"code":"voll","schluessel":"voll"}}`, "fachfehler"},
		// **Ein Fehlerwert über 4 KiB wird nicht behalten** — und dann nicht
		// geraten, sondern unbekannt.
		{"ein überlanger Fehlerwert", http.StatusOK, `{"daten":"` + strings.Repeat("x", rufgrenze) + `","fehler":{"code":"v","schluessel":"v","meldung":"` + strings.Repeat("m", endgrenze) + `"}}`, "unbekannt"},
		{"ein Fehlerfeld in einer Zeichenkette", http.StatusOK, `{"daten":"` + strings.Repeat("x", rufgrenze) + `\"fehler\":{\"code\":\"x\"}","fehler":null}`, "gelungen"},
		{"ein Fehler vorn in einer langen Antwort", http.StatusOK, `{"fehler":{"code":"voll","schluessel":"voll"},"daten":"` + strings.Repeat("x", rufgrenze) + `"}`, "fachfehler"},
		{"eine lange Liste", http.StatusOK, `[` + strings.Repeat(`1,`, rufgrenze) + `1]`, "gelungen"},
		{"ein langer Schlüssel vor dem Fehler", http.StatusOK, `{"` + strings.Repeat("k", rufgrenze) + `":1,"fehler":null}`, "gelungen"},
		{"ein Objekt ohne Fehler und viel Leerraum dahinter", http.StatusOK, `{"sets":[]}` + strings.Repeat(" ", rufgrenze), "gelungen"},
		{"ein kaputtes Fehlerfeld am Ende", http.StatusOK, `{"daten":"` + strings.Repeat("x", rufgrenze) + `","fehler":}`, "unbekannt"},
		{"ein Fehlerfeld ohne Fehlerform", http.StatusOK, `{"daten":"` + strings.Repeat("x", rufgrenze) + `","fehler":"kaputt"}`, "unbekannt"},
		// **Ein Abbruch ist kein Fehler**: eine Vorschau, die beim Tippen
		// von der nächsten überholt wird, endet so.
		{"ein Abbruch", http.StatusOK, `{"fehler":{"code":"abgebrochen"}}`, "abgebrochen"},
		// **Eine Bindung, die nur einen Fehler zurückgibt** (`*Fehler`, etwa
		// `ExemplarAendern`), gibt ihn ohne Umschlag — und `null`, wenn es
		// gelang.
		{"ein Fehler ohne Umschlag", http.StatusOK, `{"code":"konflikt","schluessel":"konflikt","meldung":"x"}`, "fachfehler"},
		{"kein Fehler ohne Umschlag", http.StatusOK, `null`, "gelungen"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			buch := neuesRufbuch()
			buch.Huelle(antwortend(t, fall.status, fall.rumpf)).ServeHTTP(httptest.NewRecorder(), bindungsruf(t, "Sets"))

			eintrag := buch.Abzug(false)["Sets"]
			if eintrag.Anzahl != 1 || eintrag.Ausgaenge[fall.ausgang] != 1 {
				t.Errorf("Eintrag %+v, erwartet war ein Ruf mit Ausgang %q", eintrag, fall.ausgang)
			}

			if eintrag.DauerMs < 0 {
				t.Errorf("Dauer %v", eintrag.DauerMs)
			}
		})
	}
}

func TestDasRufbuchZaehltNurBindungsrufe(t *testing.T) {
	t.Parallel()

	buch := neuesRufbuch()
	huelle := buch.Huelle(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))

	huelle.ServeHTTP(httptest.NewRecorder(),
		httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/index.html", http.NoBody))
	huelle.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(),
		http.MethodPost, "/wails/runtime", strings.NewReader(`{"object":7,"method":1,"args":{}}`)))
	huelle.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(),
		http.MethodPost, "/wails/runtime", strings.NewReader(`kein JSON`)))

	if abzug := buch.Abzug(false); len(abzug) != 0 {
		t.Errorf("gezählt wurde %+v", abzug)
	}
}

func TestDasRufbuchLeertSichAufWunsch(t *testing.T) {
	t.Parallel()

	buch := neuesRufbuch()
	huelle := buch.Huelle(antwortend(t, http.StatusOK, `{}`))

	for range 3 {
		huelle.ServeHTTP(httptest.NewRecorder(), bindungsruf(t, "Datenstand"))
	}

	if abzug := buch.Abzug(true); abzug["Datenstand"].Anzahl != 3 {
		t.Errorf("Abzug %+v", abzug)
	}

	if abzug := buch.Abzug(false); len(abzug) != 0 {
		t.Errorf("nach dem Leeren steht noch %+v", abzug)
	}
}

// **Die Antwort geht unverändert hinaus** — das Rufbuch schreibt mit und
// greift nicht ein.
func TestDasRufbuchReichtDieAntwortDurch(t *testing.T) {
	t.Parallel()

	aufnahme := httptest.NewRecorder()
	neuesRufbuch().Huelle(antwortend(t, http.StatusAccepted, `{"x":1}`)).ServeHTTP(aufnahme, bindungsruf(t, "Sets"))

	if aufnahme.Code != http.StatusAccepted || aufnahme.Body.String() != `{"x":1}` {
		t.Errorf("hinaus ging %d %q", aufnahme.Code, aufnahme.Body.String())
	}
}

// **Ein zerteilter Ruf fehlt nicht still** — sein Name steht im ganzen
// Rumpf, nicht im Teil.
func TestEinZerteilterRufWirdAlsSolcherGezaehlt(t *testing.T) {
	t.Parallel()

	buch := neuesRufbuch()
	anfrage := bindungsruf(t, "Sets")
	anfrage.Header.Set("x-wails-chunk-id", "1")
	buch.Huelle(antwortend(t, http.StatusOK, `{}`)).ServeHTTP(httptest.NewRecorder(), anfrage)

	if abzug := buch.Abzug(false); abzug["zerteilt"].Anzahl != 1 {
		t.Errorf("Abzug %+v", abzug)
	}
}

// **Das Ende bleibt auch bei kleinen Stücken das Ende** — eine Antwort, die
// in vielen kleinen Schreibvorgängen kommt, behält ihre letzten Bytes.
func TestDasRufbuchBehaeltDasEndeAuchStueckweise(t *testing.T) {
	t.Parallel()

	rumpf := langeAntwort(`{"code":"validierung","schluessel":"v"}`)
	stueckweise := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, _ = io.ReadAll(req.Body) //nolint:errcheck // gelesen wird nur, damit der Rumpf weg ist

		for stelle := 0; stelle < len(rumpf); stelle += 100 {
			_, _ = rw.Write([]byte(rumpf[stelle:min(stelle+100, len(rumpf))])) //nolint:errcheck // ein Recorder scheitert nicht
		}
	})

	buch := neuesRufbuch()
	buch.Huelle(stueckweise).ServeHTTP(httptest.NewRecorder(), bindungsruf(t, "Sets"))

	if eintrag := buch.Abzug(false)["Sets"]; eintrag.Ausgaenge["fachfehler"] != 1 {
		t.Errorf("Eintrag %+v, erwartet war ein fachlicher Fehler", eintrag)
	}
}
