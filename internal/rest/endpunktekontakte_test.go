// **Die Wege der bekannten Gegenüber** — Werkstück N1 von Phase 5.4
// (#1454).
//
// **Eine eigene Datei neben `endpunktepeer_test.go`**, aus demselben Grund
// und demselben Schnitt: dort, wer **dieses** Programm ist — hier, wer die
// **anderen** sind.
package rest_test

import (
	"net/http"
	"testing"
)

// endpunkteKontakte sind die sieben Wege zu den Kontakten.
//
// **Die Vorrichtung antwortet auf jeden mit der ganzen Liste**, und deshalb
// steht überall `StatusOK`: die Klassen, die dieser Lauf prüft — Marke,
// Herkunft, Drossel, Durchreichen —, hängen am Weg und nicht am Ergebnis.
func endpunkteKontakte() []endpunkt {
	return []endpunkt{
		{
			name: "Kontakte lesen", methode: http.MethodGet,
			weg: "/peer/kontakte", erfolg: http.StatusOK,
		},
		{
			name: "Kontakt aufnehmen", methode: http.MethodPost,
			weg:    "/peer/kontakte",
			rumpf:  `{"kennung":"bwer","name":"Wer"}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Kontakt bestätigen", methode: http.MethodPost,
			weg:    "/peer/kontakte/bwer/bestaetigung",
			rumpf:  `{"fingerabdruck":"abcd-ef01"}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Kontakt sperren", methode: http.MethodPut,
			weg: "/peer/kontakte/bwer/sperre", erfolg: http.StatusOK,
		},
		{
			name: "Kontakt entsperren", methode: http.MethodDelete,
			weg: "/peer/kontakte/bwer/sperre", erfolg: http.StatusOK,
		},
		{
			name: "Zweck freigeben", methode: http.MethodPut,
			weg: "/peer/kontakte/bwer/zwecke/teilen", erfolg: http.StatusOK,
		},
		{
			name: "Zweck zurücknehmen", methode: http.MethodDelete,
			weg: "/peer/kontakte/bwer/zwecke/teilen", erfolg: http.StatusOK,
		},
		{
			name: "Adresse setzen", methode: http.MethodPut,
			weg:    "/peer/kontakte/bwer/adresse",
			rumpf:  `{"adresse":"192.168.1.7:41234"}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Notiz setzen", methode: http.MethodPut,
			weg:    "/peer/kontakte/bwer/notiz",
			rumpf:  `{"notiz":"tauscht fair"}`,
			erfolg: http.StatusOK,
		},

		// **Einladen, prüfen, annehmen** (NET-01, #1705) — der Code im
		// Rumpf und nie im Pfad.
		{
			name: "Einladung ausstellen", methode: http.MethodPost,
			weg:    "/peer/einladungen",
			rumpf:  `{"name":"Anna","zweck":"handel"}`,
			erfolg: http.StatusCreated,
		},
		{
			name: "Einladung prüfen", methode: http.MethodPost,
			weg:    "/peer/einladungen/pruefung",
			rumpf:  `{"code":"mse4mac-kontakt:1:probe.probe"}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Einladung annehmen", methode: http.MethodPost,
			weg:    "/peer/einladungen/annahme",
			rumpf:  `{"code":"mse4mac-kontakt:1:probe.probe","name":"Anna","vergleichscode":"","freigeben":false}`,
			erfolg: http.StatusOK,
		},

		// **Die freigegebenen Sichten** (T3, #1468) — eine Freigabe ist eine
		// **Ressource**: `POST` erteilt, `DELETE` nimmt zurück.
		{
			name: "Freigaben zeigen", methode: http.MethodGet,
			weg: "/peer/kontakte/bwer/freigaben", erfolg: http.StatusOK,
		},
		{
			name: "Sicht freigeben", methode: http.MethodPost,
			weg:    "/peer/kontakte/bwer/freigaben",
			rumpf:  `{"umfang":"sammlung","zweck":"handel"}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Freigabe widerrufen", methode: http.MethodDelete,
			weg:    "/peer/kontakte/bwer/freigaben/handel/sammlung",
			erfolg: http.StatusOK,
		},
		{
			name: "Sicht ansehen", methode: http.MethodGet,
			weg: "/peer/kontakte/bwer/sicht?zweck=handel", erfolg: http.StatusOK,
		},
		{
			name: "Tausch abgleichen", methode: http.MethodPost,
			weg:    "/peer/kontakte/bwer/abgleich",
			rumpf:  `{"angebot":[],"wuensche":[]}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Poststand zeigen", methode: http.MethodGet,
			weg: "/peer/kontakte/bwer/post", erfolg: http.StatusOK,
		},
		{
			name: "Nachbarn zeigen", methode: http.MethodGet,
			weg: "/peer/nachbarn", erfolg: http.StatusOK,
		},

		// **Der Handel ist die Ressource, der Schritt sein Zustand** (T6,
		// #1471).
		{
			name: "Handelsschritt gehen", methode: http.MethodPost,
			weg:    "/handel/v-1/schritt",
			rumpf:  `{"schritt":"verbuchen","wer":""}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Journal zeigen", methode: http.MethodGet,
			weg: "/handel/v-1/journal", erfolg: http.StatusOK,
		},
		// **Und wer unterschrieben hat — ohne einen Schritt** (Hüllenprobe
		// `nat04handel`, #1672).
		{
			name: "Beleg zeigen", methode: http.MethodGet,
			weg: "/handel/v-1/beleg", erfolg: http.StatusOK,
		},

		// **Und die Schritte davor** (T1, #1466, angeschlossen mit #1637).
		{
			name: "Handel eröffnen", methode: http.MethodPost,
			weg:    "/handel",
			rumpf:  `{"vorgang":"v-1","art":"tausch","gegenueber":["bgegenueber"]}`,
			erfolg: http.StatusCreated,
		},
		{
			name: "Handelsstand zeigen", methode: http.MethodGet,
			weg: "/handel/v-1", erfolg: http.StatusOK,
		},
		// **Und wie er dorthin kam** (TRADE-03, #1717).
		{
			name: "Handelsprotokoll zeigen", methode: http.MethodGet,
			weg: "/handel/v-1/protokoll", erfolg: http.StatusOK,
		},
		// **Und was die Seiten ungefähr wert sind** (TRADE-02b, #1951).
		{
			name: "Preisanhalt zeigen", methode: http.MethodGet,
			weg: "/handel/v-1/anhalt", erfolg: http.StatusOK,
		},
		// **Die Liste und die Entwürfe** (TRADE-01, #1715).
		{
			name: "Handelsliste zeigen", methode: http.MethodGet,
			weg: "/handel?lage=dran&lage=wartet&archiv=true", erfolg: http.StatusOK,
		},
		{
			name: "Entwurf ändern", methode: http.MethodPut, weg: "/handel/v-1/entwurf",
			rumpf: `{"art":"tausch","gegenueber":["bgegenueber"]}`, erfolg: http.StatusOK,
		},
		{name: "Entwurf löschen", methode: http.MethodDelete, weg: "/handel/v-1/entwurf", erfolg: http.StatusNoContent},
		{name: "Handel kopieren", methode: http.MethodPost, weg: "/handel/v-1/kopie", erfolg: http.StatusCreated},
		{name: "Handel archivieren", methode: http.MethodPut, weg: "/handel/v-1/archiv", erfolg: http.StatusNoContent},
		{name: "Handel aus dem Archiv", methode: http.MethodDelete, weg: "/handel/v-1/archiv", erfolg: http.StatusNoContent},
		// **Die Handel mit einem Gegenüber** (UX-08.AK-04, #1680) — damit
		// niemand ihre Kennung tippen muss.
		{
			name: "Handel mit einem Gegenüber zeigen", methode: http.MethodGet,
			weg: "/peer/kontakte/bwer/handel", erfolg: http.StatusOK,
		},
		{
			name: "Handel vorschlagen", methode: http.MethodPost,
			weg:    "/handel/v-1/vorschlag",
			rumpf:  `{"positionen":[]}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Handel versenden", methode: http.MethodPost,
			weg:    "/handel/v-1/versand",
			rumpf:  `{"empfaenger":"bgegenueber"}`,
			erfolg: http.StatusCreated,
		},
		{
			name: "Handel wiederaufnehmen", methode: http.MethodPost,
			weg:    "/handel/v-1/wiederaufnahme",
			rumpf:  `{"empfaenger":"bgegenueber","revision":0,"abdruck":"","folge":0}`,
			erfolg: http.StatusOK,
		},
		// **Die Zusagen** (#1643) — welche Besitzzeilen weggehen.
		{
			name: "Handel festhalten", methode: http.MethodPost,
			weg:    "/handel/v-1/zusagen",
			rumpf:  `{"exemplarId":7,"menge":1,"fassung":1}`,
			erfolg: http.StatusCreated,
		},
		{
			name: "Handelszusagen zeigen", methode: http.MethodGet,
			weg: "/handel/v-1/zusagen", erfolg: http.StatusOK,
		},
		{
			name: "Handelspost bestätigen", methode: http.MethodPost,
			weg: "/handelspost/n-1/bestaetigung", rumpf: `{}`,
			erfolg: http.StatusOK,
		},

		// **Wohin eine Karte außerhalb führt** (T8, #1473).
		{
			name: "Marktlinks zeigen", methode: http.MethodGet,
			weg: "/drucke/bolt-m10/links", erfolg: http.StatusOK,
		},

		// **Das eigene Konto bei einem Marktplatz** (T11, #1476) — ein
		// `GET`, und es schreibt nirgends.
		{
			name: "Marktkonto zeigen", methode: http.MethodGet,
			weg: "/markt/cardtrader/konto", erfolg: http.StatusOK,
		},
	}
}

// **Ein krummer Rumpf endet vor der Naht** — Werkstück N1 von Phase 5.4.
//
// # Warum das hier steht und nicht in der Liste oben
//
// **Weil es kein Weg ist, sondern ein Nein.** Die Liste führt Wege, die
// glücken; dieser Lauf prüft die drei, die einen Rumpf entgegennehmen, gegen
// einen, der keiner ist — und dass sie ihn abweisen, **bevor** irgendjemand
// eine Kennung zu sehen bekommt.
func TestEinKrummerKontaktrumpfEndetVorDerNaht(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct{ methode, weg string }{
		"aufnehmen":   {http.MethodPost, "/peer/kontakte"},
		"bestaetigen": {http.MethodPost, "/peer/kontakte/bwer/bestaetigung"},
		"adresse":     {http.MethodPut, "/peer/kontakte/bwer/adresse"},
		"notiz":       {http.MethodPut, "/peer/kontakte/bwer/notiz"},

		// **Und die drei Wege der Einladung** (NET-01, #1705).
		"einladen":  {http.MethodPost, "/peer/einladungen"},
		"prüfen":    {http.MethodPost, "/peer/einladungen/pruefung"},
		"annehmen":  {http.MethodPost, "/peer/einladungen/annahme"},
		"freigeben": {http.MethodPost, "/peer/kontakte/bwer/freigaben"},

		// **Und der Abgleich** (T4, #1469): sein Rumpf ist der fremde Stand,
		// und einer, der keiner ist, endet vor der Naht.
		"abgleichen": {http.MethodPost, "/peer/kontakte/bwer/abgleich"},

		// **Und die vier Wege des Handels** (T1, #1466, angeschlossen mit
		// #1637): ihr Rumpf trägt Positionen, einen Empfänger oder einen
		// fremden Fortschritt — und keiner davon darf halb gedeutet werden.
		"eröffnen":        {http.MethodPost, "/handel"},
		"vorschlagen":     {http.MethodPost, "/handel/v-1/vorschlag"},
		"versenden":       {http.MethodPost, "/handel/v-1/versand"},
		"wiederaufnehmen": {http.MethodPost, "/handel/v-1/wiederaufnahme"},
		"festhalten":      {http.MethodPost, "/handel/v-1/zusagen"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := ruf(t, neuerDienst(t), endpunkt{
				methode: fall.methode, weg: fall.weg, rumpf: "kein JSON",
			}, gueltig)

			if antwort.Code != http.StatusBadRequest {
				t.Fatalf("der krumme Rumpf gab %d: %s",
					antwort.Code, antwort.Body.String())
			}
		})
	}
}
