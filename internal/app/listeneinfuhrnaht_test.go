// **Die Listeneinfuhrnaht** — Werkstück 2 von Phase 4.5
// (#1292).
//
// # Warum gegen echtes Storage und eine Gegenstelle
//
// Die Naht ist der **Composition Root** dieser Fähigkeit: sie baut Klient,
// Zwischenspeicher, Drossel und Speicher. Gegen Doubles geprüft bliebe genau
// das ungeprüft, was sie tut.
//
// **Zwei Dinge stehen an jedem Lauf** (§3): ein eigenes `HOME` **und** eine
// umgelenkte Adresse. Das eine ohne das andere hat am 2026-08-24 zweimal
// wirklich 25 MB von Scryfall geladen
// (#888).
package app_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// listenkarte ist eine Kartenantwort, wie Scryfall sie schickt — selbst
// geschrieben (§11), auf das Nötige gekürzt.
const listenkarte = `{"object":"card","id":"00000000-0000-4000-8000-00000000000%d",` +
	`"name":"%s","lang":"en","layout":"normal","set":"probe","set_name":"Probesatz",` +
	`"collector_number":"%d","rarity":"common","type_line":"Instant"}`

// einfuhrnaht baut die Naht über einem frischen Bestand und einer Gegenstelle.
func einfuhrnaht(t *testing.T) *app.Listeneinfuhrnaht {
	t.Helper()

	t.Setenv("HOME", t.TempDir())

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			frage := anfrage.URL.RawQuery

			// **Ein Ausfall der Gegenstelle** und **mehr als eine Seite** —
			// G01 (`listeneinfuhrbefund_test.go`).
			if strings.Contains(frage, "Ausfall") {
				schreiben.WriteHeader(http.StatusInternalServerError)

				return
			}

			if strings.Contains(frage, "Viele") {
				schreibeRoh(t, schreiben, `{"object":"list","has_more":true,"total_cards":200,"data":[`+
					fmt.Sprintf(listenkarte, 1, "Viele", 1)+`,`+fmt.Sprintf(listenkarte, 2, "Viele", 2)+`]}`)

				return
			}

			if strings.Contains(frage, "Gibtesnicht") {
				schreiben.WriteHeader(http.StatusNotFound)
				schreibeRoh(t, schreiben, `{"object":"error","code":"not_found",`+
					`"status":404,"details":"No cards found."}`)

				return
			}

			// **Ein genauer Name samt Setangabe gibt genau einen** — so
			// verhält sich der echte Dienst (nachgemessen am 2026-08-30).
			if strings.Contains(frage, "%21%22") && strings.Contains(frage, "e%3A") {
				schreibeRoh(t, schreiben, `{"object":"list","has_more":false,`+
					`"total_cards":1,"data":[`+
					fmt.Sprintf(listenkarte, 1, "Erste Probe", 1)+`]}`)

				return
			}

			schreibeRoh(t, schreiben, `{"object":"list","has_more":false,`+
				`"total_cards":2,"data":[`+
				fmt.Sprintf(listenkarte, 1, "Erste Probe", 1)+`,`+
				fmt.Sprintf(listenkarte, 2, "Erste Probe", 2)+`]}`)
		}))

	t.Cleanup(dienst.Close)

	vorher := app.MitScryfalladresse(dienst.URL)
	t.Cleanup(func() { app.MitScryfalladresse(vorher) })

	naht, err := app.ListeneinfuhrUeber(frischerBestand(t))
	if err != nil {
		t.Fatalf("ListeneinfuhrUeber: %v", err)
	}

	return naht
}

// schreibeRoh schreibt und meldet, wenn es nicht geht.
func schreibeRoh(t *testing.T, schreiben http.ResponseWriter, roh string) {
	t.Helper()

	if _, fehler := schreiben.Write([]byte(roh)); fehler != nil {
		t.Errorf("schreiben: %v", fehler)
	}
}

// **Der ganze Weg: deuten, dann einlesen.**
func TestDieEinfuhrnahtFuehrtDenGanzenWeg(t *testing.T) {
	naht := einfuhrnaht(t)

	deutung, err := naht.Deuten(probekontext(t), "2 Erste Probe (probe) 1\n3 Zweite Probe")
	if err != nil {
		t.Fatalf("Deuten: %v", err)
	}

	if len(deutung.Zeilen) != 2 {
		t.Fatalf("%d Zeilen, erwartet 2", len(deutung.Zeilen))
	}

	if !deutung.Zeilen[0].Eindeutig {
		t.Error("die Zeile mit Setangabe gilt als nicht eindeutig")
	}

	if deutung.Zeilen[1].Eindeutig || deutung.Zeilen[1].Drucke != 2 {
		t.Errorf("die offene Zeile: eindeutig=%v, Drucke=%d",
			deutung.Zeilen[1].Eindeutig, deutung.Zeilen[1].Drucke)
	}

	// **Der lesbare Bezug ist mitgekommen**, nicht nur die Kennung.
	if deutung.Zeilen[0].Treffer[0].Setname != "Probesatz" {
		t.Errorf("Setname %q", deutung.Zeilen[0].Treffer[0].Setname)
	}

	bericht, err := naht.Einlesen(probekontext(t), []app.Listeneinfuhrzeile{{
		Printing: deutung.Zeilen[0].Treffer[0].Printing,
		Menge:    deutung.Zeilen[0].Menge,
		Zustand:  string(domain.ZustandNearMint),
		Sprache:  domain.Vorgabesprache,
		Name:     "Erste Probe", Setcode: "probe", Nummer: "1",
	}}, "")
	if err != nil {
		t.Fatalf("Einlesen: %v", err)
	}

	if len(bericht.Angelegt) != 1 || bericht.Stueck != 2 {
		t.Errorf("%d angelegt, %d Stück — erwartet 1 und 2",
			len(bericht.Angelegt), bericht.Stueck)
	}
}

// **Ein Name, den es nicht gibt, ist ein Befund und kein Fehler.**
func TestDieEinfuhrnahtMitUnbekanntemNamen(t *testing.T) {
	deutung, err := einfuhrnaht(t).Deuten(probekontext(t), "1 Gibtesnicht")
	if err != nil {
		t.Fatalf("Deuten: %v", err)
	}

	if deutung.Zeilen[0].Befund == "" {
		t.Error("kein Befund zu einem Namen, den es nicht gibt")
	}

	// **Und die Trefferliste ist leer und nicht `nil`** — über die Naht käme
	// sonst ein `null` heraus, wo die Oberfläche eine Liste zählt (§3).
	if deutung.Zeilen[0].Treffer == nil {
		t.Error("die Trefferliste ist nil statt leer")
	}
}

// **Eine leere Liste ist ein fachliches Nein.**
func TestDieEinfuhrnahtMitLeererListe(t *testing.T) {
	_, err := einfuhrnaht(t).Deuten(probekontext(t), "   \n ")

	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet ErrValidation", err)
	}
}

// **Ohne eine einzige Zeile ist das Einlesen ein fachliches Nein.**
func TestDieEinfuhrnahtOhneZeilen(t *testing.T) {
	_, err := einfuhrnaht(t).Einlesen(probekontext(t), nil, "")

	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet ErrValidation", err)
	}
}

// **Ohne Zuhause gibt es keine Naht.**
//
// Der Zweig, der ohne diesen Lauf unbetreten bliebe: `Scryfallverzeichnis`
// scheitert, und die Naht wird gar nicht erst gebaut — statt eine halbe
// herauszugeben (§8).
func TestDieEinfuhrnahtOhneZuhause(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	if _, err := app.ListeneinfuhrUeber(frischerBestand(t)); err == nil {
		t.Error("ohne Zuhause kam eine Naht heraus")
	}
}

// **Ohne eingerichtete Naht sagt die Bindung es selbst.**
//
// Sie kann fehlen, und das ist kein Programmfehler: `ListeneinfuhrUeber`
// braucht ein Zuhause für den Zwischenspeicher. Ohne diesen Lauf bliebe der
// Zweig unbetreten — und im Betrieb bliebe die Seite leer, statt zu sagen,
// was ihr fehlt (§8).
func TestOhneEingerichteteListeneinfuhr(t *testing.T) {
	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	gedeutet := anwendung.ListeDeuten(t.Context(), "1 Probe")
	if gedeutet.Fehler == nil {
		t.Error("ListeDeuten ohne Naht gab keinen Fehler")
	}

	eingelesen := anwendung.ListeEinlesen(t.Context(), nil, "")
	if eingelesen.Fehler == nil {
		t.Error("ListeEinlesen ohne Naht gab keinen Fehler")
	}
}

// **Ein Fehler der Naht kommt als Fehler an der Bindung an.**
func TestEinFehlerDerListeneinfuhr(t *testing.T) {
	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitListeneinfuhr(stolpernde{})

	if anwendung.ListeDeuten(t.Context(), "1 Probe").Fehler == nil {
		t.Error("ListeDeuten gab keinen Fehler")
	}

	if anwendung.ListeEinlesen(t.Context(), nil, "").Fehler == nil {
		t.Error("ListeEinlesen gab keinen Fehler")
	}
}

// stolpernde scheitert auf beiden Wegen.
type stolpernde struct{}

func (stolpernde) Deuten(context.Context, string) (*app.Listendeutung, error) {
	return nil, errors.New("kein Netz")
}

func (stolpernde) Ausgeben(context.Context, string) (string, error) {
	return "", errors.New("die Platte klemmt")
}

func (stolpernde) CSVEinlesen(context.Context, string, string, string) (*app.Einfuhrbericht, error) {
	return nil, errors.New("die Platte klemmt")
}

func (stolpernde) Einlesen(context.Context, []app.Listeneinfuhrzeile, string) (
	*app.Einfuhrbericht, error,
) {
	return nil, errors.New("kein Netz")
}

// **Die Bindungen gehen denselben Weg wie die App-Methoden.**
//
// Ohne diesen Lauf blieben `Bindungen.ListeDeuten` und `.ListeEinlesen`
// ungerufen — und eine Bindung, die kein Lauf betritt, ist Code mit voller
// Abdeckung, den niemand benutzt (§10, erste Dimension).
func TestDieBindungenDerListeneinfuhr(t *testing.T) {
	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitListeneinfuhr(stolpernde{})

	bindungen := app.NeueBindungen(anwendung)

	if bindungen.ListeDeuten(t.Context(), "1 Probe").Fehler == nil {
		t.Error("Bindungen.ListeDeuten gab keinen Fehler")
	}

	if bindungen.ListeEinlesen(t.Context(), nil, "").Fehler == nil {
		t.Error("Bindungen.ListeEinlesen gab keinen Fehler")
	}
}

// **Der geglückte Weg über die App-Methoden.**
//
// Die Läufe darüber gehen an `Listeneinfuhrnaht` unmittelbar; hier geht es
// durch `App` — die Schicht, die den Umschlag baut. Ohne diesen Lauf war der
// Erfolgszweig beider Methoden unbetreten, und geprüft war nur, was sie im
// Fehlerfall tun.
func TestDerGeglueckteWegUeberDieApp(t *testing.T) {
	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitListeneinfuhr(willige{})

	gedeutet := anwendung.ListeDeuten(t.Context(), "1 Probe")
	if gedeutet.Fehler != nil {
		t.Fatalf("ListeDeuten: %+v", gedeutet.Fehler)
	}

	if gedeutet.Deutung == nil || len(gedeutet.Deutung.Zeilen) != 1 {
		t.Errorf("die Deutung kam als %+v an", gedeutet.Deutung)
	}

	eingelesen := anwendung.ListeEinlesen(t.Context(), []app.Listeneinfuhrzeile{{
		Printing: "probe-druck", Menge: 1,
	}}, "")
	if eingelesen.Fehler != nil {
		t.Fatalf("ListeEinlesen: %+v", eingelesen.Fehler)
	}

	if eingelesen.Bericht == nil || eingelesen.Bericht.Stueck != 1 {
		t.Errorf("der Bericht kam als %+v an", eingelesen.Bericht)
	}
}

// willige gibt auf beiden Wegen ein Ergebnis.
type willige struct{}

func (willige) Ausgeben(context.Context, string) (string, error) {
	return "printing,menge\n", nil
}

func (willige) CSVEinlesen(context.Context, string, string, string) (*app.Einfuhrbericht, error) {
	return &app.Einfuhrbericht{Angelegt: []int64{1}, Stueck: 1, Befunde: []string{}}, nil
}

func (willige) Deuten(context.Context, string) (*app.Listendeutung, error) {
	return &app.Listendeutung{
		Zeilen:  []app.Listendeutungszeile{{Roh: "1 Probe", Menge: 1}},
		Befunde: []string{},
	}, nil
}

func (willige) Einlesen(context.Context, []app.Listeneinfuhrzeile, string) (
	*app.Einfuhrbericht, error,
) {
	return &app.Einfuhrbericht{
		Angelegt: []int64{1}, Stueck: 1, Befunde: []string{},
	}, nil
}
