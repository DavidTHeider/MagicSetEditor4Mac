// **Keine leere Liste wird zu `null`** — und zwar zur **Laufzeit**
// (#1288).
//
// # Die andere Hälfte der Naht
//
// `internal/werkzeuge/leerelisten_test.go` hält, dass kein Listenfeld in
// `internal/app` ein `omitempty` trägt. Er sagt selbst, was er **nicht** hält:
//
//	Ein `nil`-Slice wird zu `null`, auch ohne `omitempty`. Dagegen hilft nur
//	`make([]T, 0)` an jeder Stelle, die die Antwort baut — und ob das
//	geschehen ist, sieht kein Muster.
//
// Das war richtig, solange man es als **Muster** dachte. Als **Lauf** sieht man
// es: jede Bindung wird gerufen, ihre Antwort wird zu JSON gemacht, und jedes
// Feld, das im Go-Typ eine Liste oder eine Tabelle ist, muss `[]` oder `{}`
// sein — nie `null`.
//
// # Was am 2026-08-29 dabei herauskam
//
// **Am 2026-08-29: 132 Fundstellen in 112 der 142 Bindungen.** **112** davon lagen an
// `Fehler.Werte` — 110 als `fehler.werte` einer Antwort und zwei als `werte`
// einer Antwort, die **selbst** ein Fehler ist (`ExemplarAendern`,
// `ExemplarLoeschen`). Das ist das Feld, das der Zweitblick über die Wails-Naht
// an der laufenden Brücke gemessen hat: `{"fehler":{…,"werte":null}}`, während
// der erzeugte Typ `werte: Record<string, string>` sagt, **nicht** optional.
//
// Die übrigen **20** waren andere Listen in Antworten: `befunde`, `teile`,
// `angelegt`, `felder`, `hinweise`.
//
// **Die Zerlegung stand an zwei Stellen verschieden da** — hier 110 + 22, in
// `leerelisten.go` 112 + 20. Dieselbe Messung, zwei Zahlen: angestrichen vom
// Zweitblick über Phase 4.1, Befund 4.
//
// # Was dieser Lauf hält
//
// Jede exportierte Methode auf `*Bindungen` wird mit **Nullwerten** gerufen.
// Die Antwort wird zu JSON gemacht und gegen ihren Go-Typ gelesen: wo der Typ
// Slice oder Map sagt, darf das JSON nicht `null` sein.
//
// # Was er nicht hält, und das ist die Hälfte
//
//   - **Nur Fehlerwege.** Nullwerte als Argumente heißt: eine Kennung, die
//     keine ist, ein Pfad, den es nicht gibt. Was eine Bindung mit **echten**
//     Daten zurückgibt, sieht er nicht. Genau dafür gibt es den Nahtlauf durch
//     die laufende Brücke — und der ist die Prüfung der Grenze, nicht diese.
//   - **Nicht, dass die Antwort im Fenster ankommt.** Zwei grüne Seiten einer
//     Grenze sind kein Nachweis für die Grenze. Dieser Lauf steht auf der
//     Go-Seite.
//   - **Zeiger dürfen `null` sein.** `*Fehler` ist genau so gebaut: „kein
//     Fehler" heißt `null`, und der erzeugte TypeScript-Typ sagt es auch. Wer
//     das mitprüfte, prüfte eine Zusage, die niemand gegeben hat.
package app_test

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// gerueststeteBindungen baut eine gestartete Anwendung mit gestopften Nähten.
//
// **Gestartet, nicht bloß gebaut.** Ohne `Starten` steht kein Bestand, und
// sechzehn Bindungen laufen in eine Nullzeiger-Ausnahme statt in eine Antwort
// — gemessen am 2026-08-29. Eine Messung über 126 von 142 Bindungen wäre eine
// Aussage über 126 (§10).
func gerueststeteBindungen(t *testing.T) *app.Bindungen {
	t.Helper()

	// **Zwei Dinge, und einzeln reicht keines** (§3): ein eigenes Zuhause,
	// damit der Lauf nicht den Stand dieses Rechners sieht, **und** eine
	// umgelenkte Adresse, damit er nicht nach draußen greift.
	//
	// Gemessen am 2026-08-29, bevor die zweite Zeile hier stand:
	// `ScryfallRegeln` brauchte **3,24 s** und war damit der einzige Aufruf
	// dieses Laufs, der wirklich funktioniert hat — nach Scryfall.
	t.Setenv("HOME", t.TempDir())

	vorher := app.MitScryfalladresse("http://127.0.0.1:1/")
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	// **Der Dateidialog wird gestopft, nicht bedient.** Dreizehn Fähigkeiten
	// hängen an ihm; ein Lauf, der ihn öffnete, bliebe stehen.
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return "", errors.New("in diesem Lauf gibt es kein Fenster")
	})

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("die Anwendung startet nicht: %v", err)
	}

	anwendung.Bereit(t.Context())

	return app.NeueBindungen(anwendung)
}

// TestKeineBindungGibtEineListeAlsNull ist der Lauf selbst.
func TestKeineBindungGibtEineListeAlsNull(t *testing.T) {
	bindungen := gerueststeteBindungen(t)

	typ := reflect.TypeOf(bindungen)
	wert := reflect.ValueOf(bindungen)

	gerufen := 0

	for i := range typ.NumMethod() {
		methode := typ.Method(i)

		// **Die Sitzung bleibt offen** (ID-02): `Abmelden` schließt sie, und
		// danach hielte das Tor jede weitere Bindung an. Der Nutzer dieses
		// Laufs hat weder Kennwort noch Faktor — eine lokale Anmeldung genügt.
		if bindungen.Sitzungsstand(t.Context()).Zustand == "abgemeldet" {
			bindungen.LokalAnmelden(t.Context(), "1", "", "")
		}

		t.Run(methode.Name, func(t *testing.T) {
			antwort, hat := mitNullwertenRufen(t, wert.Method(i), methode.Type)
			if !hat {
				return
			}

			gerufen++

			roh, err := json.Marshal(antwort)
			if err != nil {
				t.Fatalf("die Antwort lässt sich nicht zu JSON machen: %v", err)
			}

			// **Ein Einstieg für jede Gestalt.**
			//
			// `MTGJSONSpiele() []string` gibt keinen Verbund zurück, und der
			// Leser stieg für alles, was kein Verbund ist, sofort wieder aus:
			// eine `return nil` dort blieb grün (Zweitblick über Phase 4.1,
			// Befund 2). Die Zusage lautet **jede** Bindung — und seit dem
			// Nachlauf auch **jede Ebene** darin.
			listenwertPruefen(t, methode.Type.Out(0), roh, "die Antwort")
		})
	}

	// **Sonst wäre ein Lauf ohne Bindungen grün.** Die Zahl steht nicht hier,
	// weil sie mit jeder neuen Bindung stiege; was sie hält, ist der Zähler in
	// `internal/werkzeuge`.
	if gerufen == 0 {
		t.Fatal("keine einzige Bindung gab eine Antwort — der Lauf greift daneben")
	}
}

// mitNullwertenRufen ruft eine Methode mit Nullwerten und gibt ihr erstes
// Ergebnis.
func mitNullwertenRufen(
	t *testing.T, methode reflect.Value, art reflect.Type,
) (any, bool) {
	t.Helper()

	if art.NumOut() == 0 {
		return nil, false
	}

	// **Der Rufkontext ist kein Nullwert** (BK-01, #2109): Wails füllt ihn
	// immer, und ein `nil` darin wäre ein Ruf, den es nicht gibt.
	kontexttyp := reflect.TypeFor[context.Context]()

	argumente := make([]reflect.Value, 0, art.NumIn()-1)
	for stelle := 1; stelle < art.NumIn(); stelle++ {
		if art.In(stelle) == kontexttyp {
			argumente = append(argumente, reflect.ValueOf(t.Context()))

			continue
		}

		argumente = append(argumente, reflect.Zero(art.In(stelle)))
	}

	return methode.Call(argumente)[0].Interface(), true
}

// listenwertPruefen hält einen Wert, der **selbst** eine Liste ist — und alles
// darin.
//
// **Getrennt von [listenfelderPruefen]**, weil es eine andere Frage ist: dort
// geht es um die Felder eines Verbunds, hier um den Wert selbst. Wer beides in
// eine Schleife legte, bräuchte einen Zweig für „ist kein Verbund" — und genau
// den hat der erste Anlauf benutzt, um auszusteigen.
//
// **Und er steigt in die Liste hinein.** `[][]T` mit einem `nil` darin,
// `map[K][]T` mit einem `nil`-Wert und eine Liste von Verbünden, deren
// Listenfeld `nil` ist, gingen bis zum Nachlauf über Phase 4.1 durch (Befund
// 2): der Wächter sah die äußere Liste und hörte dort auf.
func listenwertPruefen(t *testing.T, art reflect.Type, roh json.RawMessage, was string) {
	t.Helper()

	for art.Kind() == reflect.Pointer {
		art = art.Elem()
	}

	switch art.Kind() {
	case reflect.Slice, reflect.Map:
		if string(roh) == "null" {
			t.Errorf("%s ist `null` — der Typ sagt %s, und das Fenster bekommt "+
				"keine leere Liste, sondern gar keine", was, art)

			return
		}

		hineinPruefen(t, art, roh, was)
	case reflect.Struct:
		listenfelderPruefen(t, art, roh, was+".")

	default:
	}
}

// hineinPruefen geht durch die Glieder einer Liste oder Tabelle.
func hineinPruefen(t *testing.T, art reflect.Type, roh json.RawMessage, was string) {
	t.Helper()

	if art.Kind() == reflect.Map {
		var glieder map[string]json.RawMessage
		if err := json.Unmarshal(roh, &glieder); err != nil {
			return
		}

		for name, glied := range glieder {
			listenwertPruefen(t, art.Elem(), glied, was+"["+name+"]")
		}

		return
	}

	var glieder []json.RawMessage
	if err := json.Unmarshal(roh, &glieder); err != nil {
		return
	}

	for stelle, glied := range glieder {
		listenwertPruefen(t, art.Elem(), glied,
			was+"["+strconv.Itoa(stelle)+"]")
	}
}

// listenfelderPruefen liest eine Antwort gegen ihren Go-Typ.
//
// **Über den Typ und nicht über das JSON.** Ob `[]` richtig ist, sagt die
// Erklärung des Feldes; ein JSON-Wert `null` allein sagt nur, dass etwas fehlt
// — bei einem Zeiger ist das die Zusage, bei einer Liste ihr Bruch.
func listenfelderPruefen(t *testing.T, art reflect.Type, roh json.RawMessage, pfad string) {
	t.Helper()

	for art.Kind() == reflect.Pointer {
		art = art.Elem()
	}

	if art.Kind() != reflect.Struct {
		return
	}

	var felder map[string]json.RawMessage
	if err := json.Unmarshal(roh, &felder); err != nil {
		return
	}

	for stelle := range art.NumField() {
		feld := art.Field(stelle)

		// **`jsonname` steht in `leerelisten_test.go`** — dieselbe Frage,
		// dieselbe Antwort, und zwei Abschriften liefen auseinander.
		name := jsonname(feld)
		if name == "" || name == "-" {
			continue
		}

		wert, dabei := felder[name]
		if !dabei {
			continue
		}

		switch feld.Type.Kind() {
		case reflect.Slice, reflect.Map:
			listenwertPruefen(t, feld.Type, wert, pfad+name)
		default:
			listenfelderPruefen(t, feld.Type, wert, pfad+name+".")
		}
	}
}
