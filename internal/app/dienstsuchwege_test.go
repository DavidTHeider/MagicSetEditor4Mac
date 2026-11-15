package app_test

// **Die übrigen Wege der Dienstsuche** — Werkstück 8 von Phase 4.8.
//
// **Eine eigene Datei**, damit `dienstsuchnaht_test.go` unter der
// Zeilengrenze aus §2 bleibt. Der Schnitt folgt der Frage: dort steht die
// **Abnahme**, hier die Fehlerwege und die beiden Dienste, die sich nur mit
// Zugang bauen lassen.

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine Suche über MTGJSON läuft, wenn die Marke hinterlegt ist.**
func TestEineSucheUeberMTGJSONLaeuft(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(schreiber http.ResponseWriter, _ *http.Request) {
			if _, err := schreiber.Write([]byte(
				`{"data":{"cards":[{"uuid":"u1","name":"Blitzschlag",` +
					`"setCode":"PRB","manaCost":"{R}","type":"Instant"}]}}`)); err != nil {
				t.Errorf("antworten: %v", err)
			}
		}))
	defer server.Close()

	t.Setenv("HOME", t.TempDir())

	zurueck := app.MitDienstadressen(server.URL, server.URL, server.URL)
	defer zurueck()

	quelle := app.Suchnaht(
		dienstwahlprobe{an: map[domain.Dienst]bool{domain.DienstMTGJSON: true}},
		bundprobe{werte: map[string]string{"mtgjson.marke": "geheim"}},
		fragewahlprobe{dienst: domain.DienstMTGJSON})

	auskunft, suchfehler := quelle.Suche(probekontext(t), "Blitzschlag")
	if suchfehler != nil {
		t.Fatalf("suchen: %v", suchfehler)
	}

	if len(auskunft.Treffer) != 1 || auskunft.Treffer[0].Kennung != "u1" {
		t.Errorf("gelesen: %+v", auskunft.Treffer)
	}
}

// **Und eine über Scrydex, wenn beide Angaben da sind.**
func TestEineSucheUeberScrydexLaeuft(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(schreiber http.ResponseWriter, _ *http.Request) {
			if _, err := schreiber.Write([]byte(
				`{"status":200,"data":[{"id":"s1","name":"Blitzschlag",` +
					`"mana_cost":"{R}","type":"Instant","rarity":"common",` +
					`"expansion":{"code":"PRB","name":"Probeset"}}]}`)); err != nil {
				t.Errorf("antworten: %v", err)
			}
		}))
	defer server.Close()

	t.Setenv("HOME", t.TempDir())

	zurueck := app.MitDienstadressen(server.URL, server.URL, server.URL)
	defer zurueck()

	quelle := app.Suchnaht(
		dienstwahlprobe{an: map[domain.Dienst]bool{domain.DienstScrydex: true}},
		bundprobe{werte: map[string]string{
			"scrydex.schluessel": "geheim",
			"scrydex.team":       "mannschaft",
		}},
		fragewahlprobe{dienst: domain.DienstScrydex})

	auskunft, suchfehler := quelle.Suche(probekontext(t), "Blitzschlag")
	if suchfehler != nil {
		t.Fatalf("suchen: %v", suchfehler)
	}

	if len(auskunft.Treffer) != 1 || auskunft.Treffer[0].Setname != "Probeset" {
		t.Errorf("gelesen: %+v", auskunft.Treffer)
	}
}

// **Ein Lesefehler am Schalter hält die Naht an.**
//
// Wer nicht weiß, was erlaubt ist, baut keine Karte von Diensten — eine ohne
// den betroffenen Dienst sähe aus wie eine bewusste Abschaltung.
type kaputtewahl struct{}

func (kaputtewahl) IstAn(context.Context, domain.Dienst) (bool, error) {
	return false, domain.ErrNotFound
}

func TestEinLesefehlerAmSchalterHaeltAn(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	quelle := app.Suchnaht(kaputtewahl{}, bundprobe{},
		fragewahlprobe{dienst: domain.DienstScryfall})

	// **Beim Suchen und nicht beim Verdrahten** — die Schalter werden je
	// Anfrage gelesen (Befund 1 des Zweitblicks über Phase 4.8).
	_, suchfehler := quelle.Suche(probekontext(t), "Blitz")

	if !errors.Is(suchfehler, domain.ErrNotFound) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", suchfehler)
	}
}

// **Ein kaputter Schlüsselbund hält sie ebenso an.**
//
// „Da steht nichts" und „der Schlüsselbund antwortet nicht" sind zwei
// Auskünfte — und nur die zweite ist ein Fehler.
type kaputterbund struct{}

func (kaputterbund) Lies(context.Context, string, string) (string, error) {
	return "", domain.ErrConflict
}

func TestEinKaputterSchluesselbundHaeltAn(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	quelle := app.Suchnaht(
		dienstwahlprobe{an: map[domain.Dienst]bool{domain.DienstMTGJSON: true}},
		kaputterbund{},
		fragewahlprobe{dienst: domain.DienstMTGJSON})

	_, suchfehler := quelle.Suche(probekontext(t), "Blitz")

	if !errors.Is(suchfehler, domain.ErrConflict) {
		t.Fatalf("Fehler ist %v, erwartet ErrConflict", suchfehler)
	}

	if !strings.Contains(suchfehler.Error(), "Schlüsselbund") {
		t.Errorf("die Meldung %q sagt nicht, woran es lag", suchfehler)
	}
}

// **Und derselbe Weg für Scrydex' zweite Angabe.**
func TestEinKaputterSchluesselbundHaeltAuchScrydexAn(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	quelle := app.Suchnaht(
		dienstwahlprobe{an: map[domain.Dienst]bool{domain.DienstScrydex: true}},
		kaputterbund{},
		fragewahlprobe{dienst: domain.DienstScrydex})

	_, suchfehler := quelle.Suche(probekontext(t), "Blitz")

	if !errors.Is(suchfehler, domain.ErrConflict) {
		t.Errorf("Fehler ist %v, erwartet ErrConflict", suchfehler)
	}
}

// **Die Wahl je Frage kommt aus den Einstellungen.**
//
// Ohne Eintrag gilt die Vorgabe — der erste Dienst, der die Frage beantworten
// kann.
func TestDieFragewahlKommtAusDenEinstellungen(t *testing.T) {
	db := frischerBestand(t)

	wahl := app.NeueEinstellungswahl(storage.NeueEinstellungen(db))

	dienst, err := wahl.FuerFrage(probekontext(t), domain.FrageSuche)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if dienst != domain.DienstScryfall {
		t.Errorf("ohne Eintrag ist %q gewählt, erwartet scryfall", dienst)
	}

	if err := storage.NeueEinstellungen(db).Setzen(probekontext(t),
		domain.FrageEinstellung(domain.FrageSuche),
		string(domain.DienstMTGAPI)); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	gesetzt, lesefehler := wahl.FuerFrage(probekontext(t), domain.FrageSuche)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if gesetzt != domain.DienstMTGAPI {
		t.Errorf("gewählt ist %q, erwartet mtgapi", gesetzt)
	}
}

// **`ImmerAn` nimmt den ersten, der kann** — und sagt es, wenn es keinen gibt.
func TestImmerAnNimmtDenErsten(t *testing.T) {
	t.Parallel()

	immer := app.ImmerAn{}

	dienst, err := immer.FuerFrage(probekontext(t), domain.FrageSuche)
	if err != nil {
		t.Fatalf("ImmerAn: %v", err)
	}

	if dienst != domain.DienstScryfall {
		t.Errorf("ImmerAn nimmt %q, erwartet scryfall", dienst)
	}

	_, unbekannt := immer.FuerFrage(probekontext(t), domain.Frage("gibtsnicht"))
	if unbekannt == nil {
		t.Error("eine unbekannte Frage kam ohne Fehler durch")
	}
}

// **Jeder der drei Klienten reicht seinen Fehler durch.**
//
// Gegen eine tote Adresse: die Suche gibt einen Fehler, keine leere Liste.
// Ohne diesen Lauf bliebe offen, ob ein Ausfall wie „nichts gefunden"
// aussieht (§8).
func TestJederKlientReichtSeinenFehlerDurch(t *testing.T) {
	for _, fall := range []struct {
		name   string
		dienst domain.Dienst
		bund   bundprobe
	}{
		{"magicthegathering.io", domain.DienstMTGAPI, bundprobe{}},
		{
			"MTGJSON", domain.DienstMTGJSON,
			bundprobe{werte: map[string]string{"mtgjson.marke": "geheim"}},
		},
		{
			"Scrydex", domain.DienstScrydex,
			bundprobe{werte: map[string]string{
				"scrydex.schluessel": "geheim",
				"scrydex.team":       "mannschaft",
			}},
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())

			tot := "http://127.0.0.1:1"

			zurueck := app.MitDienstadressen(tot, tot, tot)
			defer zurueck()

			quelle := app.Suchnaht(
				dienstwahlprobe{an: map[domain.Dienst]bool{fall.dienst: true}},
				fall.bund, fragewahlprobe{dienst: fall.dienst})

			if _, suchfehler := quelle.Suche(probekontext(t), "X"); suchfehler == nil {
				t.Error("gegen eine tote Adresse kam kein Fehler")
			}
		})
	}
}

// **Ein Lesefehler bei der Fragewahl kommt an.**
func TestEinLesefehlerBeiDerFragewahlKommtAn(t *testing.T) {
	t.Parallel()

	wahl := app.NeueEinstellungswahl(leserprobe{fehler: domain.ErrNotFound})

	if _, err := wahl.FuerFrage(probekontext(t), domain.FrageSuche); !errors.Is(
		err, domain.ErrNotFound) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", err)
	}
}

// **Eine gesetzte Umgebungsangabe lenkt die Adresse um.**
//
// Ohne diesen Lauf bliebe der Zweig unbetreten, und ein Testlauf ginge
// irgendwann an den echten Dienst, ohne dass es jemandem auffiele (§3).
func TestEineUmgebungsangabeLenktUm(t *testing.T) {
	t.Setenv(app.MTGAPIumgebung, "http://127.0.0.1:12345")

	if gelesen := app.AdresseAus(app.MTGAPIumgebung, "echt"); gelesen != "http://127.0.0.1:12345" {
		t.Errorf("die Adresse ist %q, erwartet die umgelenkte", gelesen)
	}

	// **Und ohne Angabe bleibt die echte** — sonst prüfte der Lauf oben nur,
	// dass immer umgelenkt wird.
	t.Setenv(app.MTGAPIumgebung, "")

	if gelesen := app.AdresseAus(app.MTGAPIumgebung, "echt"); gelesen != "echt" {
		t.Errorf("ohne Angabe ist die Adresse %q, erwartet »echt«", gelesen)
	}
}

// **Auch die zweite Angabe von Scrydex wird geprüft.**
//
// `kaputterbund` scheitert schon am Schlüssel — dieser hier gibt ihn her und
// scheitert erst an der Team-Kennung. Ohne den Unterschied bliebe der zweite
// Zweig unbetreten, und ein Fehler dort verschwände still.
type halbkaputterbund struct{}

func (halbkaputterbund) Lies(_ context.Context, _, konto string) (string, error) {
	if strings.HasSuffix(konto, domain.ZugangSchluessel) {
		return "geheim", nil
	}

	return "", domain.ErrConflict
}

func TestAuchDieZweiteAngabeVonScrydexWirdGeprueft(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	quelle := app.Suchnaht(
		dienstwahlprobe{an: map[domain.Dienst]bool{domain.DienstScrydex: true}},
		halbkaputterbund{},
		fragewahlprobe{dienst: domain.DienstScrydex})

	_, suchfehler := quelle.Suche(probekontext(t), "Blitz")

	if !errors.Is(suchfehler, domain.ErrConflict) {
		t.Errorf("Fehler ist %v, erwartet ErrConflict", suchfehler)
	}
}
