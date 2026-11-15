package rest_test

// **Dateien hin und her über den Dienst** — FMT-03.AK-03
// (#1692).
//
// Ein Aufrufer auf einem anderen Rechner hat keinen Pfad auf diesem. Er
// schickt die Datei selbst (`Content-Type: application/zip`), und eine
// Ausgabe ohne `ziel` holt er über `Location: /ausgang/{name}` ab.

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// ablageprobe hält Hochgeladenes und Ausgaben im Speicher.
type ablageprobe struct {
	sperre     sync.Mutex
	angenommen map[string][]byte
	ausgaenge  map[string][]byte
	zwischen   []string
	weg        []string
	scheitern  error
}

func neueAblageprobe() *ablageprobe {
	return &ablageprobe{angenommen: map[string][]byte{}, ausgaenge: map[string][]byte{}}
}

func (a *ablageprobe) Ablegen(_ context.Context, endung string, inhalt io.Reader) (string, error) {
	roh, err := io.ReadAll(inhalt)
	if err != nil {
		return "", err
	}

	if a.scheitern != nil {
		return "", a.scheitern
	}

	summe := sha256.Sum256(roh)
	pfad := "/eingang/" + hex.EncodeToString(summe[:4]) + endung

	a.sperre.Lock()
	defer a.sperre.Unlock()

	a.angenommen[pfad] = roh

	return pfad, nil
}

// Zwischenlegen merkt sich, was nur für eine Anfrage lag, und ob es danach
// wieder wegkam.
func (a *ablageprobe) Zwischenlegen(
	_ context.Context, endung string, inhalt io.Reader,
) (pfad string, weg func(), err error) {
	if _, err := io.ReadAll(inhalt); err != nil {
		return "", nil, err
	}

	if a.scheitern != nil {
		return "", nil, a.scheitern
	}

	pfad = "/vorschau/zwischen" + endung

	a.sperre.Lock()
	defer a.sperre.Unlock()

	a.zwischen = append(a.zwischen, pfad)

	return pfad, func() {
		a.sperre.Lock()
		defer a.sperre.Unlock()

		a.weg = append(a.weg, pfad)
	}, nil
}

func (a *ablageprobe) Ausgangsort(_ context.Context, endung string) (name, pfad string, err error) {
	if a.scheitern != nil {
		return "", "", a.scheitern
	}

	a.sperre.Lock()
	defer a.sperre.Unlock()

	name = "a" + strconv.Itoa(len(a.ausgaenge)) + endung
	a.ausgaenge[name] = []byte("%PDF-inhalt von " + name)

	return name, "/ausgang/" + name, nil
}

func (a *ablageprobe) Abholen(_ context.Context, name string) (io.ReadCloser, error) {
	a.sperre.Lock()
	defer a.sperre.Unlock()

	roh, da := a.ausgaenge[name]
	if !da {
		return nil, domain.ErrNotFound
	}

	return io.NopCloser(bytes.NewReader(roh)), nil
}

func (a *ablageprobe) Vergessen(_ context.Context, name string) error {
	a.sperre.Lock()
	defer a.sperre.Unlock()

	if _, da := a.ausgaenge[name]; !da {
		return domain.ErrNotFound
	}

	delete(a.ausgaenge, name)

	return nil
}

// merkendeAusfuhr merkt sich, womit Einfahren und Bogen gerufen wurden.
type merkendeAusfuhr struct {
	williger

	sperre              sync.Mutex
	eingefahren, gelegt string
}

func (m *merkendeAusfuhr) Einfahren(
	ctx context.Context, pfad string,
) (setID int64, karten int, befunde []string, fehler error) {
	m.sperre.Lock()
	m.eingefahren = pfad
	m.sperre.Unlock()

	return m.williger.Einfahren(ctx, pfad)
}

func (m *merkendeAusfuhr) Bogen(
	ctx context.Context, setID int64, ziel string,
) (seiten, karten int, befunde []string, fehler error) {
	m.sperre.Lock()
	m.gelegt = ziel
	m.sperre.Unlock()

	return m.williger.Bogen(ctx, setID, ziel)
}

func uebertragungsdienst(ablage *ablageprobe, ausfuhr *merkendeAusfuhr, grenzen domain.Haushaltsgrenzen) http.Handler {
	return rest.NeuerDienst(probemarke, time.Now, rest.Naehte{
		Leser: williger{}, Ausfuhr: ausfuhr, Setvorschau: williger{}, Dateiablage: ablage,
		Auftraege: rest.NeuesAuftragsbuch(time.Now),
	}).MitGrenzen(grenzen).Behandler()
}

func schicke(t *testing.T, dienst http.Handler, methode, weg, art string, rumpf []byte, kopf ...string) *httptest.ResponseRecorder {
	t.Helper()

	anfrage := httptest.NewRequestWithContext(t.Context(), methode, weg, bytes.NewReader(rumpf))
	gueltig(anfrage)

	if art != "" {
		anfrage.Header.Set("Content-Type", art)
	}

	for stelle := 0; stelle+1 < len(kopf); stelle += 2 {
		anfrage.Header.Set(kopf[stelle], kopf[stelle+1])
	}

	antwort := httptest.NewRecorder()
	dienst.ServeHTTP(antwort, anfrage)

	return antwort
}

func TestEinHochgeladenesSetWirdEingefahren(t *testing.T) {
	ablage, ausfuhr := neueAblageprobe(), &merkendeAusfuhr{}
	dienst := uebertragungsdienst(ablage, ausfuhr, domain.Vorgabegrenzen())
	inhalt := []byte("PK\x03\x04 ein Set")

	antwort := schicke(t, dienst, http.MethodPost, "/sets", "application/zip", inhalt)
	if antwort.Code != http.StatusCreated || !strings.HasPrefix(ausfuhr.eingefahren, "/eingang/") ||
		!bytes.Equal(ablage.angenommen[ausfuhr.eingefahren], inhalt) {
		t.Fatalf("%d %s — eingefahren %q", antwort.Code, antwort.Body.String(), ausfuhr.eingefahren)
	}

	// **Derselbe Inhalt, derselbe Ort** — und damit gilt „eine Quelle, ein
	// Set" auch für eine zweite Übertragung.
	erster := ausfuhr.eingefahren
	schicke(t, dienst, http.MethodPost, "/sets", "application/octet-stream", inhalt)

	if ausfuhr.eingefahren != erster {
		t.Errorf("zweiter Ort %q, erster %q", ausfuhr.eingefahren, erster)
	}

	// **Auch die Vorschau nimmt die Datei selbst — aber nur zwischen** (#1944):
	// sie liegt für diese eine Anfrage und geht danach, und in den Eingang
	// kommt nichts dazu.
	vorher := len(ablage.angenommen)
	if vorschau := schicke(t, dienst, http.MethodPost, "/sets/vorschau", "application/zip", inhalt); vorschau.Code != http.StatusOK ||
		!strings.Contains(vorschau.Body.String(), "/vorschau/zwischen.mse-set") {
		t.Errorf("Vorschau: %d %s", vorschau.Code, vorschau.Body.String())
	}

	if len(ablage.angenommen) != vorher || len(ablage.zwischen) != 1 || len(ablage.weg) != 1 {
		t.Errorf("Eingang %d → %d, zwischen %v, weg %v", vorher, len(ablage.angenommen), ablage.zwischen, ablage.weg)
	}

	// **Und eine, die nicht zwischengelegt werden kann, sagt es.**
	ablage.scheitern = domain.ErrVoll
	if vorschau := schicke(t, dienst, http.MethodPost, "/sets/vorschau", "application/zip", inhalt); vorschau.Code != http.StatusInsufficientStorage {
		t.Errorf("Vorschau ohne Platz: %d %s", vorschau.Code, vorschau.Body.String())
	}
}

func TestEinScheiterndesAnnehmenWirdGemeldet(t *testing.T) {
	ablage, ausfuhr := neueAblageprobe(), &merkendeAusfuhr{}
	ablage.scheitern = domain.ErrVoll
	dienst := uebertragungsdienst(ablage, ausfuhr, domain.Vorgabegrenzen())

	if antwort := schicke(t, dienst, http.MethodPost, "/sets", "application/zip", []byte("PK")); antwort.Code != http.StatusInsufficientStorage {
		t.Errorf("%d %s", antwort.Code, antwort.Body.String())
	}

	if antwort := schicke(t, dienst, http.MethodPost, "/sets/3/bogen", "application/json", []byte(`{}`)); antwort.Code != http.StatusInsufficientStorage {
		t.Errorf("Ausgang: %d %s", antwort.Code, antwort.Body.String())
	}
}

// **Eine Übertragung über der Grenze wird abgewiesen** — und die Grenze der
// Übertragung ist eine andere als die einer Anfrage mit JSON.
func TestEineZuGrosseUebertragungWirdAbgewiesen(t *testing.T) {
	ablage, ausfuhr := neueAblageprobe(), &merkendeAusfuhr{}
	grenzen := domain.Vorgabegrenzen()
	grenzen.Antwort = 64

	dienst := uebertragungsdienst(ablage, ausfuhr, grenzen)

	if antwort := schicke(t, dienst, http.MethodPost, "/sets", "application/zip", bytes.Repeat([]byte("x"), 200)); antwort.Code != http.StatusCreated {
		t.Errorf("über der Grenze einer Antwort, unter der einer Übertragung: %d", antwort.Code)
	}

	eng := rest.NeuerDienst(probemarke, time.Now, rest.Naehte{
		Leser: williger{}, Ausfuhr: ausfuhr, Dateiablage: ablage,
	}).MitHochladegrenze(100).Behandler()

	if antwort := schicke(t, eng, http.MethodPost, "/sets", "application/zip", bytes.Repeat([]byte("x"), 200)); antwort.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("über der Grenze einer Übertragung: %d %s", antwort.Code, antwort.Body.String())
	}
}

// **Eine Setdatei hat keine eigene Art** — sie geht als Bytes hinaus.
func TestEineAbgeholteSetdateiIstBytes(t *testing.T) {
	ablage, ausfuhr := neueAblageprobe(), &merkendeAusfuhr{}
	dienst := uebertragungsdienst(ablage, ausfuhr, domain.Vorgabegrenzen())

	gesichert := schicke(t, dienst, http.MethodPost, "/sets/3/sicherung", "application/json", []byte(`{}`))
	abgeholt := schicke(t, dienst, http.MethodGet, gesichert.Header().Get("Location"), "", nil)

	if abgeholt.Code != http.StatusOK || abgeholt.Header().Get("Content-Type") != "application/octet-stream" {
		t.Errorf("%d %v", abgeholt.Code, abgeholt.Header())
	}
}

func TestEineAusgabeOhneZielWirdAbgeholt(t *testing.T) {
	ablage, ausfuhr := neueAblageprobe(), &merkendeAusfuhr{}
	dienst := uebertragungsdienst(ablage, ausfuhr, domain.Vorgabegrenzen())

	gelegt := schicke(t, dienst, http.MethodPost, "/sets/3/bogen", "application/json", []byte(`{}`))
	ort := gelegt.Header().Get("Location")

	if gelegt.Code != http.StatusCreated || ort != "/ausgang/a0.pdf" || ausfuhr.gelegt != "/ausgang/a0.pdf" {
		t.Fatalf("%d %s — Ort %q, gelegt nach %q", gelegt.Code, gelegt.Body.String(), ort, ausfuhr.gelegt)
	}

	abgeholt := schicke(t, dienst, http.MethodGet, ort, "", nil)
	if abgeholt.Code != http.StatusOK || abgeholt.Header().Get("Content-Type") != "application/pdf" ||
		!strings.Contains(abgeholt.Header().Get("Content-Disposition"), `filename=a0.pdf`) ||
		!strings.HasPrefix(abgeholt.Body.String(), "%PDF-") {
		t.Errorf("abgeholt: %d %v %q", abgeholt.Code, abgeholt.Header(), abgeholt.Body.String())
	}

	if weg := schicke(t, dienst, http.MethodDelete, ort, "", nil); weg.Code != http.StatusNoContent {
		t.Errorf("vergessen: %d", weg.Code)
	}

	if danach := schicke(t, dienst, http.MethodGet, ort, "", nil); danach.Code != http.StatusNotFound {
		t.Errorf("danach: %d", danach.Code)
	}

	if weg := schicke(t, dienst, http.MethodDelete, ort, "", nil); weg.Code != http.StatusNotFound {
		t.Errorf("zweimal vergessen: %d", weg.Code)
	}
}

// **Mit einem Ziel bleibt alles, wie es war** — und ein Rumpf, der kein
// Objekt ist, geht unverändert an den Weg, der ihn abweist.
func TestEinGenanntesZielBleibt(t *testing.T) {
	ablage, ausfuhr := neueAblageprobe(), &merkendeAusfuhr{}
	dienst := uebertragungsdienst(ablage, ausfuhr, domain.Vorgabegrenzen())

	mit := schicke(t, dienst, http.MethodPost, "/sets/3/bogen", "application/json", []byte(`{"ziel":"/tmp/eigen.pdf"}`))
	if mit.Code != http.StatusCreated || mit.Header().Get("Location") != "" || ausfuhr.gelegt != "/tmp/eigen.pdf" {
		t.Errorf("%d %v %q", mit.Code, mit.Header(), ausfuhr.gelegt)
	}

	if kaputt := schicke(t, dienst, http.MethodPost, "/sets/3/bogen", "application/json", []byte(`[kaputt`)); kaputt.Code != http.StatusBadRequest {
		t.Errorf("kaputt: %d", kaputt.Code)
	}
}

// **Eine Ausgabe als Auftrag behält ihren Ort** — das Ergebnis nennt ihn wie
// die Antwort, die der Weg gleich gegeben hätte.
func TestEineAusgabeAlsAuftragNenntIhrenOrt(t *testing.T) {
	ablage, ausfuhr := neueAblageprobe(), &merkendeAusfuhr{}
	dienst := uebertragungsdienst(ablage, ausfuhr, domain.Vorgabegrenzen())

	angenommen := schicke(t, dienst, http.MethodPost, "/sets/3/bogen", "application/json", []byte(`{}`), "Prefer", "respond-async")
	auftrag := alsAuftrag(t, angenommen)

	bisZumStand(t, dienst, auftrag.ID, "fertig")

	ergebnis := schicke(t, dienst, http.MethodGet, "/auftraege/"+auftrag.ID+"/ergebnis", "", nil)
	if ergebnis.Code != http.StatusCreated || !strings.HasPrefix(ergebnis.Header().Get("Location"), "/ausgang/") {
		t.Errorf("%d %v", ergebnis.Code, ergebnis.Header())
	}
}

// **Die Liste der Übertragungswege ist vollständig und geordnet** — drei
// nehmen eine Datei, fünf schreiben in den Ausgang; der Wächter über
// `docs/DATEIOPERATIONEN.md` liest sie.
func TestDieUebertragungswegeSindGezaehlt(t *testing.T) {
	wege := rest.Uebertragungswege()

	if len(wege) != 8 || !slices.IsSorted(wege) || !slices.Contains(wege, "POST /sets/vorschau") ||
		!slices.Contains(wege, "POST /pakete/weitergabe") {
		t.Errorf("%v", wege)
	}
}

// fristenschreiber merkt sich die Fristen, die jemand an ihm setzt.
type fristenschreiber struct {
	*httptest.ResponseRecorder

	lesen, schreiben time.Time
}

func (f *fristenschreiber) SetReadDeadline(frist time.Time) error {
	f.lesen = frist

	return nil
}

func (f *fristenschreiber) SetWriteDeadline(frist time.Time) error {
	f.schreiben = frist

	return nil
}

// **Nach einer langen Übertragung darf die Antwort noch geschrieben
// werden.** Die Schreibfrist des Dienstes zählt ab dem Anfragekopf
// (`net/http`, `WriteTimeout`); ein Set, das sechs Minuten lang ankommt,
// hätte bei fünf Minuten Schreibfrist keine Antwort mehr bekommen — auch
// kein `413`.
func TestEineUebertragungVerlaengertBeideFristen(t *testing.T) {
	dienst := uebertragungsdienst(neueAblageprobe(), &merkendeAusfuhr{}, domain.Vorgabegrenzen())
	gleich := time.Now().Add(rest.Hochladefrist - time.Minute)

	anfrage := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/sets", strings.NewReader("PK"))
	gueltig(anfrage)
	anfrage.Header.Set("Content-Type", "application/zip")

	hin := &fristenschreiber{ResponseRecorder: httptest.NewRecorder()}
	dienst.ServeHTTP(hin, anfrage)

	if hin.Code != http.StatusCreated || hin.lesen.Before(gleich) || hin.schreiben.Before(gleich) {
		t.Errorf("%d — Lesefrist %v, Schreibfrist %v", hin.Code, hin.lesen, hin.schreiben)
	}

	abholung := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/ausgang/fehlt.pdf", nil)
	gueltig(abholung)

	zurueck := &fristenschreiber{ResponseRecorder: httptest.NewRecorder()}
	dienst.ServeHTTP(zurueck, abholung)

	if zurueck.schreiben.Before(gleich) {
		t.Errorf("Abholen: Schreibfrist %v", zurueck.schreiben)
	}
}
