// **Was an den Tokenbindungen schiefgehen kann** — Werkstück 10 von
// Phase 4.12
// (#1351).
//
// **Eine eigene Datei**, weil `zugriffstoken_test.go` mit ihnen über 400
// Zeilen ginge (§2). Der Schnitt folgt der Frage: dort steht, **was eine
// Bindung kann**, hier, **was sie sagt, wenn sie es nicht kann**.
package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Jede krumme Kennung und jeder stolpernde Dienst wird gemeldet.**
//
// **Ohne diese Fälle bliebe die Hälfte jeder Bindung unbetreten** (§3) — und
// eine Bindung, die nur den Erfolgsweg kennt, meldet einen Fehlschlag als
// Erfolg.
func TestJedeTokenbindungMeldetIhrNein(t *testing.T) {
	kaputt := &tokenfuehrer{fehler: errors.New("die Platte ist voll")}

	faelle := map[string]func(*app.App) *app.Fehler{
		"Token, krumme Profilkennung": func(a *app.App) *app.Fehler {
			return a.Token(context.Background(), "keine-zahl").Fehler
		},
		"TokenAnlegen, krumme Profilkennung": func(a *app.App) *app.Fehler {
			return a.TokenAnlegen(context.Background(),
				app.TokenangabenDTO{ProfilID: "keine-zahl"}).Fehler
		},
		"TokenAendern, krumme Kennung": func(a *app.App) *app.Fehler {
			return a.TokenAendern(context.Background(),
				app.TokenangabenDTO{ID: "keine-zahl", Fassung: "1"}).Fehler
		},
		"TokenAendern, krumme Fassung": func(a *app.App) *app.Fehler {
			return a.TokenAendern(context.Background(),
				app.TokenangabenDTO{ID: "7", Fassung: "keine-zahl"}).Fehler
		},
		"TokenLoeschen, krumme Kennung": func(a *app.App) *app.Fehler {
			return a.TokenLoeschen(context.Background(), "keine-zahl").Fehler
		},
	}

	for name, ruf := range faelle {
		t.Run(name, func(t *testing.T) {
			if befund := ruf(mitToken(&tokenfuehrer{})); befund == nil {
				t.Error("die krumme Eingabe kam durch")
			}
		})
	}

	stolpernd := map[string]func(*app.App) *app.Fehler{
		"Token": func(a *app.App) *app.Fehler {
			return a.Token(context.Background(), "1").Fehler
		},
		"TokenAnlegen": func(a *app.App) *app.Fehler {
			return a.TokenAnlegen(context.Background(),
				app.TokenangabenDTO{ProfilID: "1"}).Fehler
		},
		"TokenAendern": func(a *app.App) *app.Fehler {
			return a.TokenAendern(context.Background(),
				app.TokenangabenDTO{ID: "7", Fassung: "1"}).Fehler
		},
		"TokenLoeschen": func(a *app.App) *app.Fehler {
			return a.TokenLoeschen(context.Background(), "7").Fehler
		},
	}

	for name, ruf := range stolpernd {
		t.Run(name+", stolpernder Dienst", func(t *testing.T) {
			if befund := ruf(mitToken(kaputt)); befund == nil {
				t.Error("der stolpernde Dienst kam durch")
			}
		})
	}
}

// **Und die Bindungen des Fensters reichen durch.**
//
// **Ohne diesen Lauf bliebe `tokenbindungen.go` unbetreten** — vier Zeilen,
// die jede für sich richtig aussehen und von denen eine die falsche Methode
// rufen könnte.
func TestDieTokenbindungenReichenDurch(t *testing.T) {
	fuehrer := &tokenfuehrer{
		klartext: "mse_7_GEHEIM",
		zeilen:   []domain.Zugriffstoken{{ID: 7, ProfilID: 1}},
	}

	bindungen := app.NeueBindungen(mitToken(fuehrer))

	if antwort := bindungen.Token(t.Context(), "1"); len(antwort.Token) != 1 {
		t.Errorf("Token: %+v", antwort)
	}

	if antwort := bindungen.TokenAnlegen(t.Context(), app.TokenangabenDTO{
		ProfilID: "1", Bezeichnung: "Nachtlauf", Principal: "Nachtrechner", Scope: "lesen",
	}); antwort.Klartext != "mse_7_GEHEIM" {
		t.Errorf("TokenAnlegen: %+v", antwort)
	}

	if antwort := bindungen.TokenAendern(t.Context(), app.TokenangabenDTO{
		ID: "7", Bezeichnung: "Tagesbau", Principal: "Nachtrechner", Scope: "lesen", Fassung: "1",
	}); antwort.Token == nil || antwort.Token.Bezeichnung != "Tagesbau" {
		t.Errorf("TokenAendern: %+v", antwort)
	}

	if antwort := bindungen.TokenLoeschen(t.Context(), "7"); antwort.Fehler != nil {
		t.Errorf("TokenLoeschen: %+v", antwort.Fehler)
	}

	if fuehrer.geloescht != 7 {
		t.Errorf("gelöscht wurde %d", fuehrer.geloescht)
	}
}

// **Und die Wache prüft ein vorgezeigtes Token.**
//
// **Gegen einen echten Bestand**, weil sie zwei Schichten zusammenführt:
// `service` prüft das Geheimnis, `storage` hängt das Profil an den Kontext.
// Ein Doppel dazwischen prüfte den Nachbau.
func TestDieTokenwachePruefteinVorgezeigtesToken(t *testing.T) {
	db := frischerBestand(t)

	verwaltung := service.NeueTokenverwaltung(storage.NeueZugriffstokenablage(db))

	angelegt, klartext, fehler := verwaltung.Anlegen(t.Context(),
		domain.Zugriffstoken{
			ProfilID: 1, Bezeichnung: "Nachtlauf", Principal: "Nachtrechner", Scope: domain.ScopeLesen,
		})
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	wache := app.TokenwacheUeber(db)

	ausgestattet, identitaet, fehler := wache.Pruefe(t.Context(), klartext, false)
	if fehler != nil {
		t.Fatalf("prüfen: %v", fehler)
	}

	profilID, lesefehler := storage.ProfilAus(ausgestattet)
	if lesefehler != nil {
		t.Fatalf("das Profil lesen: %v", lesefehler)
	}

	if profilID != angelegt.ProfilID {
		t.Errorf("Profil = %d, erwartet %d", profilID, angelegt.ProfilID)
	}

	// **Die Maschinenidentität kommt heraus** — Befund 5 des Zweitblicks über
	// Phase 5.2: der Dienst budgetiert danach, und ohne sie zählte er nach dem
	// Geheimnis, das eine Rotation wechselt.
	if erwartet := fmt.Sprintf("%d:%s", angelegt.ProfilID, angelegt.Principal); identitaet != erwartet {
		t.Errorf("die Identität ist %q, erwartet %q", identitaet, erwartet)
	}

	t.Run("und weist ab, was nicht reicht", func(t *testing.T) {
		if _, _, fehler := wache.Pruefe(t.Context(), klartext, true); fehler == nil {
			t.Error("ein Lesetoken durfte schreiben")
		}
	})

	t.Run("und was es gar nicht gibt", func(t *testing.T) {
		if _, _, fehler := wache.Pruefe(
			t.Context(), "mse_99_ERFUNDEN", false); fehler == nil {
			t.Error("ein erfundenes Token kam durch")
		}
	})

	t.Run("und ein abgelaufenes", func(t *testing.T) {
		abgelaufen, klartext, fehler := verwaltung.Anlegen(t.Context(),
			domain.Zugriffstoken{
				ProfilID: 1, Bezeichnung: "Vorbei", Principal: "Nachtrechner", Scope: domain.ScopeLesen,
				Ablauf: time.Now().Add(-time.Hour).UTC().Format(time.RFC3339),
			})
		if fehler != nil {
			t.Fatalf("anlegen: %v", fehler)
		}

		if _, _, fehler := wache.Pruefe(t.Context(), klartext, false); fehler == nil {
			t.Errorf("das abgelaufene Token %d kam durch", abgelaufen.ID)
		}
	})
}

// **Und der Flächenweg reicht alle vier durch.**
//
// **Über einen echten Bestand**, denn das ist die Zusage dieser Naht: die drei
// Flächen ohne Fenster bekommen dieselbe Anwendung (§16). Ein Doppel dahinter
// prüfte, dass vier Zeilen einander gleichen.
func TestDerTokenwegUeberDieFlaeche(t *testing.T) {
	flaeche := app.NutzerflaecheUeber(frischerBestand(t))

	angelegt := flaeche.TokenAnlegen(probekontext(t), app.TokenangabenDTO{
		ProfilID: "1", Bezeichnung: "Nachtlauf", Principal: "Nachtrechner", Scope: "lesen",
	})
	if angelegt.Fehler != nil {
		t.Fatalf("%+v", angelegt.Fehler)
	}

	if angelegt.Klartext == "" || angelegt.Token == nil {
		t.Fatalf("angekommen ist %+v", angelegt)
	}

	gelesen := flaeche.Token(probekontext(t), "1")
	if gelesen.Fehler != nil {
		t.Fatalf("%+v", gelesen.Fehler)
	}

	if len(gelesen.Token) != 1 {
		t.Fatalf("gelesen = %+v", gelesen.Token)
	}

	geaendert := flaeche.TokenAendern(probekontext(t), app.TokenangabenDTO{
		ID: angelegt.Token.ID, Bezeichnung: "Tagesbau", Principal: "Nachtrechner", Scope: "schreiben",
		Fassung: angelegt.Token.Fassung,
	})
	if geaendert.Fehler != nil {
		t.Fatalf("%+v", geaendert.Fehler)
	}

	if antwort := flaeche.TokenLoeschen(
		probekontext(t), angelegt.Token.ID); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	nachher := flaeche.Token(probekontext(t), "1")
	if len(nachher.Token) != 0 {
		t.Errorf("nach dem Löschen stehen noch %d da", len(nachher.Token))
	}
}
