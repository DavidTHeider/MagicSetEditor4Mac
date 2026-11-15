package rest_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// endpunkt beschreibt einen Weg samt einem Aufruf, der gelingt.
type endpunkt struct {
	name    string
	methode string
	weg     string
	rumpf   string

	// erfolg ist der Status, den ein gültiger Aufruf gibt.
	//
	// **Genau einer und nicht „unter 400".** Ein Endpunkt, der ein 204 gäbe
	// oder umleitete, wäre sonst grün — obwohl er den Anwendungsfall nicht
	// erreicht hat (Nachlauf zum Zweitblick über Phase U).
	erfolg int

	// ohneFehlerweg heißt: dieser Endpunkt **kann** nicht scheitern.
	//
	// **Genau einer, und er ist es aus einem sichtbaren Grund** (#1291):
	// `GET /mtgjson/spiele` gibt eine Liste, die im Programm steht — die
	// Naht dahinter (`Ausfuhr.MTGJSONSpiele() []string`) hat kein
	// Fehlerergebnis, und das sieht der Übersetzer.
	//
	// **Und die Zusage wird nicht weggelassen, sondern umgedreht**: statt
	// „hier kommt ein 404" prüft der Lauf „hier kommt trotz feindlicher Naht
	// die Antwort". Ein Endpunkt, der doch einen Fehlerweg bekommt, fällt
	// damit auf — er gäbe dann kein 200 mehr (§10).
	ohneFehlerweg bool

	// vorher richtet den Dienst ein, bevor der Weg gerufen wird — `nil`
	// heißt: nichts. **Für den Ausgang** (#2002): eine Ausgabe holt nur ab,
	// wer sie anstieß, und ein Weg ohne angelegte Ausgabe endete sonst vor
	// seinem Anwendungsfall.
	vorher func(*rest.Dienst)
}

// alleEndpunkte ist die Liste, aus der die vier Klassen entstehen.
//
// **Sie ist die einzige Stelle, an der ein Endpunkt genannt wird**, und ein
// Wächter hält sie gegen `endpunkte.go`. Ein siebter Weg ohne Zeile hier macht
// die Pipeline rot — das ist der Sinn: die vier Klassen sollen nicht deshalb
// fehlen, weil jemand sie beim Hinzufügen vergessen hat.
// Die drei Gestalten eines Aufrufs — sie tragen die drei Abweisungsklassen.
type gestalt func(*http.Request)

func gueltig(anfrage *http.Request) {
	anfrage.Host = "127.0.0.1:7777"
	anfrage.Header.Set("Authorization", "Bearer "+probemarke)
}

func ohneMarke(anfrage *http.Request) {
	anfrage.Host = "127.0.0.1:7777"
}

func fremdeHerkunft(anfrage *http.Request) {
	anfrage.Host = "127.0.0.1:7777"
	anfrage.Header.Set("Origin", "https://angreifer.example")
	anfrage.Header.Set("Authorization", "Bearer "+probemarke)
}

// ruf fährt einen Endpunkt in einer Gestalt.
func ruf(t *testing.T, dienst *rest.Dienst, ziel endpunkt, wie gestalt) *httptest.ResponseRecorder {
	t.Helper()

	var koerper *strings.Reader
	if ziel.rumpf != "" {
		koerper = strings.NewReader(ziel.rumpf)
	}

	var anfrage *http.Request
	if koerper == nil {
		anfrage = httptest.NewRequestWithContext(t.Context(), ziel.methode, ziel.weg, http.NoBody)
	} else {
		anfrage = httptest.NewRequestWithContext(t.Context(), ziel.methode, ziel.weg, koerper)
	}

	wie(anfrage)

	if ziel.vorher != nil {
		ziel.vorher(dienst)
	}

	antwort := httptest.NewRecorder()
	dienst.Behandler().ServeHTTP(antwort, anfrage)

	return antwort
}

// alsBefund liest eine Fehlerantwort.
func alsBefund(t *testing.T, antwort *httptest.ResponseRecorder) rest.Befund {
	t.Helper()

	var befund rest.Befund
	if err := json.Unmarshal(antwort.Body.Bytes(), &befund); err != nil {
		t.Fatalf("die Antwort ist kein Befund: %v\n%s", err, antwort.Body.String())
	}

	return befund
}
