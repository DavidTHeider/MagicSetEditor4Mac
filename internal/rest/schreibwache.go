package rest

import (
	"context"
	"net/http"
)

// MitSchreibwache setzt die Frage, die jede **schreibende** Anfrage vor ihrem
// Weg stellt — ID-03.AK-02
// (#1702).
//
// # Warum es sie gibt
//
// **Der Dienst hält seine Verbindung über seine ganze Laufzeit**, aber jede
// Anfrage arbeitet im Profil, das der Bestand **gerade** zeigt — oder in dem
// ihres Tokens. Wer nach dem Start als Leser in ein fremdes Profil wechselte
// oder dem eine Freigabe entzogen wurde, schrieb über die schreibende
// Verbindung vom Start weiter.
//
// **Die Frage stellt die Wurzel**: `cmd/mse-rest` baut den Akteur je Anfrage
// frisch und fragt, ob er schreiben darf.
//
// # Was sie nicht fragt
//
// **Lesende Anfragen** — `GET` und `HEAD`. **Und die zwei Wechsel**, `PUT
// /profil` und `PUT /nutzer`: wer nur noch lesen darf, muss hinauskommen.
func (d *Dienst) MitSchreibwache(wache func(ctx context.Context) error) *Dienst {
	d.schreibwache = wache

	return d
}

// schreibenVerwehrt sagt, ob diese Anfrage hier nicht laufen darf.
func (d *Dienst) schreibenVerwehrt(anfrage *http.Request) error {
	if d.schreibwache == nil || !schreibend(anfrage) || zeigerwechsel(anfrage) {
		return nil
	}

	return d.schreibwache(anfrage.Context())
}

// zeigerwechsel sagt, ob die Anfrage den Zeiger bewegt.
func zeigerwechsel(anfrage *http.Request) bool {
	return anfrage.Method == http.MethodPut &&
		(anfrage.URL.Path == "/profil" || anfrage.URL.Path == "/nutzer")
}

// MitSchemafreigabe setzt die Freigabe, die **jede** Anfrage vor allem
// anderen holt und bis zu ihrem Ende hält — Zweitblick über Phase 2 von
// 0.9.0, Befunde 1 und 2 (#2323).
//
// # Warum nicht in der Schreibwache
//
// **Dort stand sie, und drei Wege liefen an ihr vorbei**: der Nachweis, der
// `letzte_nutzung` schreibt, bevor die Schreibwache fragt; die zwei Wechsel,
// die sie ausnimmt; und der gelöste Auftrag, der läuft, wenn die Anfrage
// vorbei ist. **Und sie fragte nur**: zwischen ihrer Antwort und dem Schreiben
// konnte eine Wanderung festschreiben. Die Freigabe hält die Antwort
// (`storage.HalteSchemafreigabe`), bis die Anfrage fertig ist.
func (d *Dienst) MitSchemafreigabe(freigabe func(ctx context.Context) (func(), error)) *Dienst {
	d.schemafreigabe = freigabe

	return d
}

// freigabeHolen holt die Freigabe — ohne eingesetzte gibt es nichts zu
// halten, und eine Anfrage unter [Dienst.UnterFreigabe] hält sie schon.
func (d *Dienst) freigabeHolen(ctx context.Context) (frei func(), err error) {
	if ctx.Value(freigegeben{}) != nil {
		return func() {}, nil
	}

	return d.freigabeFuerLauf(ctx)
}

// freigabeFuerLauf holt sie **immer** — für den Lauf eines gelösten
// Auftrags: er trägt den Kontext der Anfrage weiter, die ihn annahm, und
// deren Freigabe ist längst zurückgegeben.
func (d *Dienst) freigabeFuerLauf(ctx context.Context) (frei func(), err error) {
	if d.schemafreigabe == nil {
		return func() {}, nil
	}

	return d.schemafreigabe(ctx)
}

// freigegeben markiert eine Anfrage, die ihre Freigabe schon hält.
type freigegeben struct{}

// UnterFreigabe setzt die Freigabe **vor** einen Behandler — für die Wurzel,
// die vor dem Dienst schon den Bestand liest (Nachlauf über Phase 2 von
// 0.9.0, Befund 3: `cmd/mse-rest` las das gezeigte Profil vor ihr). Der
// Dienst innen holt sie dann nicht ein zweites Mal.
func (d *Dienst) UnterFreigabe(weiter http.Handler) http.Handler {
	return http.HandlerFunc(func(schreiben http.ResponseWriter, anfrage *http.Request) {
		frei, verwehrt := d.freigabeHolen(anfrage.Context())
		if verwehrt != nil {
			fehlerhaft(schreiben, verwehrt)

			return
		}

		defer frei()

		weiter.ServeHTTP(schreiben, anfrage.WithContext(context.WithValue(anfrage.Context(), freigegeben{}, true)))
	})
}
