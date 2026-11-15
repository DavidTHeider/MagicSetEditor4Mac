package app

// **An einem fremden Tisch mitspielen** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
// **Dass ein Wunsch ohne Tisch nicht ins Leere geht, sondern abgewiesen
// wird** — und dass die Wirkungsklasse `externSenden` ist: dieser Rechner
// wird nicht erreichbar, er greift hinaus.
//
// # Was hier nicht geprüft wird
//
// **Der Beitritt selbst.** Er braucht eine echte Sitzungsverwaltung mit
// Wähler, Ausweis und Gegenstelle; was dabei entsteht, hält
// `internal/service/peersitzungen_test.go` über zwei echte Identitäten. Hier
// stehen die Wege **um** ihn herum, und sie stehen hier, weil sie sonst nur
// über ein Fenster zu betreten wären.

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ohne Tisch sitzt dieses Programm an keinem.**
func TestOhneBeitrittSitztNiemandAmFremdenTisch(t *testing.T) {
	t.Parallel()

	antwort := freigabeanwendung(t).PartieBeitrittStand(t.Context())

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %v", antwort.Fehler)
	case antwort.Laeuft:
		t.Error("es sitzt jemand")
	}
}

// **Ein Wunsch ohne Tisch wird abgewiesen** — und zwar als Konflikt.
//
// **Kein Validierungsfehler**: der Wunsch ist in Ordnung, es fehlt der Tisch.
func TestEinWunschOhneTischWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	antwort := freigabeanwendung(t).PartieWuenschen(
		t.Context(), domain.Spielbefehl{
			PartieID: "p1", Absender: "gast", Kennung: "b1",
			Art: "objekt.tappen",
		})

	if antwort.Fehler == nil {
		t.Fatal("der Wunsch kam durch")
	}

	if antwort.Fehler.Code != "konflikt" {
		t.Errorf("Code %q, erwartet konflikt", antwort.Fehler.Code)
	}
}

// **Aufstehen ohne Tisch ist keine Klage.**
//
// **Wer nicht sitzt, kann nicht aufstehen — und das ist kein Fehlschlag.** Ein
// Fehler zwänge das Fenster, den Knopf zu verstecken, statt ihn wirkungslos
// stehen zu lassen.
func TestAufstehenOhneTischIstKeineKlage(t *testing.T) {
	t.Parallel()

	antwort := freigabeanwendung(t).PartieBeitrittBeenden(t.Context())

	if antwort.Fehler != nil || antwort.Laeuft {
		t.Errorf("es kam %+v", antwort)
	}
}

// **Ein Beitritt ohne Gegenüber wird abgewiesen** — ohne es gibt es
// niemanden, den man fragen könnte.
//
// **Ohne Partie nicht mehr** (UX-08.AK-04, #1680): die nennt der Wirt, und
// dieser Weg steht in `partienennung_intern_test.go`.
func TestEinBeitrittOhneGegenueberWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, partie := range map[string]string{
		"mit Partie":  "p1",
		"ohne Partie": "",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := freigabeanwendung(t).PartieBeitreten(
				t.Context(), "", partie, PartiemitbringDTO{})

			switch {
			case antwort.Fehler == nil || antwort.Laeuft:
				t.Errorf("es kam %+v", antwort)
			case antwort.Fehler.Code != "validierung":
				t.Errorf("Code %q", antwort.Fehler.Code)
			}
		})
	}
}

// **Jeder der vier Wege fragt nach der Rolle — und die beiden, die hinausgehen,
// nach `externSenden`.**
//
// **Nicht `externEmpfangen`**: dieser Rechner wird dabei nicht erreichbar, er
// greift hinaus. Die beiden Klassen sind nicht dasselbe, und eine Tabelle, die
// die falsche nennt, verspricht etwas anderes, als geschieht (§10).
func TestJederBeitrittswegFragtNachDerRolle(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		wirkung domain.Wirkung
		fahre   func(*App) PartiebeitrittAntwort
	}{
		"PartieBeitreten": {domain.WirkungExternSenden, func(a *App) PartiebeitrittAntwort {
			return a.PartieBeitreten(t.Context(), "wer", "p1", PartiemitbringDTO{})
		}},
		"PartieWuenschen": {domain.WirkungExternSenden, func(a *App) PartiebeitrittAntwort {
			return a.PartieWuenschen(t.Context(), domain.Spielbefehl{})
		}},
		"PartieBeitrittBeenden": {domain.WirkungSchreiben, func(a *App) PartiebeitrittAntwort {
			return a.PartieBeitrittBeenden(t.Context())
		}},
		"PartieBeitrittStand": {domain.WirkungLesen, func(a *App) PartiebeitrittAntwort {
			return a.PartieBeitrittStand(t.Context())
		}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wache := &zaehlendeWache{}
			anwendung := freigabeanwendung(t).MitWache(wache)

			antwort := fall.fahre(anwendung)
			if antwort.Fehler == nil {
				t.Fatal("der Weg kam durch")
			}

			if len(wache.gefragt) == 0 || wache.gefragt[0] != fall.wirkung {
				t.Errorf("gefragt wurde nach %v statt nach %q",
					wache.gefragt, fall.wirkung)
			}
		})
	}
}

// **Ein Beitritt ohne Sitzungsverwaltung klagt und merkt sich nichts.**
//
// **Die Gegenprobe zum Erfolgsweg**, den ein Lauf ohne Gegenstelle nicht
// erreicht: wenn das Öffnen scheitert, darf danach kein halber Tisch
// dastehen (§8, kein halber Stand).
func TestEinGescheiterterBeitrittMerktSichNichts(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)

	antwort := anwendung.PartieBeitreten(t.Context(), "ein-gegenueber", "p1", PartiemitbringDTO{})
	if antwort.Fehler == nil {
		t.Fatal("der Beitritt kam durch")
	}

	if stand := anwendung.PartieBeitrittStand(t.Context()); stand.Laeuft {
		t.Errorf("danach steht ein Tisch: %+v", stand)
	}
}

// **Und was der Melder nicht hat, hält ihn nicht auf.**
//
// **Ohne Fenster geschieht nichts, und das ist kein Fehler**: an der
// Kommandozeile gibt es keines, und ein Tisch, den niemand zeichnet, ist
// trotzdem ein Tisch.
func TestOhneFensterMeldetDerTischmelderNichtsUndKlagtNicht(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)
	anwendung.melden = nil

	melder := anwendung.tischmelder(t.Context(), "p1")

	switch {
	case melder.Ereignis(t.Context(), domain.Ereignis{Folge: 1}) != nil:
		t.Error("das Ereignis klagte")
	case melder.Tischbild(t.Context(), domain.Tischbild{}) != nil:
		t.Error("das Bild klagte")
	case melder.Nebenbei(t.Context(), service.Nebenbefund{}) != nil:
		t.Error("die Nebensache klagte")
	}
}

// **Was gemeldet wird, trägt die Partie und die abgebildete Gestalt.**
func TestDerTischmelderMeldetPartieUndGestalt(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)

	var gemeldet []string

	anwendung.melden = func(_ context.Context, ereignis string, daten ...any) {
		gemeldet = append(gemeldet, ereignis)

		if len(daten) < 2 || daten[0] != "p1" {
			t.Errorf("%s meldete %v", ereignis, daten)
		}
	}

	melder := anwendung.tischmelder(t.Context(), "p1")

	mussMelden(t, melder.Ereignis(t.Context(), domain.Ereignis{Folge: 1}))
	mussMelden(t, melder.Tischbild(t.Context(), domain.Tischbild{}))
	mussMelden(t, melder.Nebenbei(t.Context(), service.Nebenbefund{
		Nachricht: domain.Partienachricht{Art: domain.ArtChat},
		Chat:      domain.Chat{Von: "anna", Text: "hallo"},
		Gilt:      true,
	}))

	if len(gemeldet) != 3 {
		t.Errorf("%d Meldungen: %v", len(gemeldet), gemeldet)
	}
}

// **Gemeldet wird, wer geredet hat — nicht, wer geschickt hat.**
//
// # Warum das einen eigenen Lauf bekommt
//
// **Weil es der Fund des Laufs über beide Enden war** (§16,
// `partietischgespraech_intern_test.go`). Der erste Anlauf nahm `Von` aus dem
// **Umschlag**, und der trägt den Absender dieses Sprungs: ein Satz von Anna,
// den der Wirt an Bert weiterreichte, stand bei Bert als Satz des Wirts.
//
// **Hier stand vorher etwas anderes** — ein Lauf über einen unlesbaren
// Inhalt. Den Zweig gibt es nicht mehr: die Nutzlast wird in
// `internal/service` ausgepackt und geprüft, bevor sie hierher kommt, und ein
// zweites Auspacken hätte einen Fehlerweg, den kein Lauf betreten kann (§10).
func TestGemeldetWirdWerGeredetHat(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)

	var gesehen NebenbeiDTO

	anwendung.melden = func(_ context.Context, _ string, daten ...any) {
		if verbund, gut := daten[1].(NebenbeiDTO); gut {
			gesehen = verbund
		}
	}

	melder := anwendung.tischmelder(t.Context(), "p1")

	mussMelden(t, melder.Nebenbei(t.Context(), service.Nebenbefund{
		Nachricht: domain.Partienachricht{Art: domain.ArtChat, Von: "wirt"},
		Chat:      domain.Chat{Von: "anna", Text: "gut gezogen"},
		Gilt:      true,
	}))

	if gesehen.Von != "anna" || gesehen.Text != "gut gezogen" {
		t.Errorf("gemeldet wurde %q von %q", gesehen.Text, gesehen.Von)
	}

	mussMelden(t, melder.Nebenbei(t.Context(), service.Nebenbefund{
		Nachricht: domain.Partienachricht{Art: domain.ArtPraesenz, Von: "wirt"},
		Praesenz: domain.Praesenz{
			Von: "bert", Art: domain.PraesenzGreift, Stand: 3,
			Objekt: "k7", X: 55, Y: -12,
		},
		Gilt: true,
	}))

	switch {
	case gesehen.Art != string(domain.ArtPraesenz):
		t.Errorf("gemeldet als %q", gesehen.Art)
	case gesehen.Von != "bert" || gesehen.Zeigerart != "greift":
		t.Errorf("%q tut %q", gesehen.Von, gesehen.Zeigerart)
	case gesehen.Objekt != "k7" || gesehen.X != 55 || gesehen.Y != -12:
		t.Errorf("er zeigt auf %q bei %v/%v",
			gesehen.Objekt, gesehen.X, gesehen.Y)
	case gesehen.Text != "":
		t.Errorf("eine Präsenz trägt den Text %q", gesehen.Text)
	}
}

// mussMelden bricht ab, wenn eine Meldung klagt.
func mussMelden(t *testing.T, fehler error) {
	t.Helper()

	if fehler != nil {
		t.Fatalf("melden: %v", fehler)
	}
}

// **Eine Absage des Wirts erreicht das Fenster** (#2478) — am Tischbild, mit
// Befehl, Art und Grund; ohne Absage trägt die Antwort keine.
func TestDerTischmelderGibtDieAbsageWeiter(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)

	var bilder []PartieAntwort

	anwendung.melden = func(_ context.Context, _ string, daten ...any) {
		if bild, ist := daten[1].(PartieAntwort); ist {
			bilder = append(bilder, bild)
		}
	}

	melder := anwendung.tischmelder(t.Context(), "p1")
	mussMelden(t, melder.Tischbild(t.Context(), domain.Tischbild{
		Absage: &domain.Wunschabsage{Befehl: "b1", Art: "objekt.tappen", Grund: "nicht erlaubt"},
	}))
	mussMelden(t, melder.Tischbild(t.Context(), domain.Tischbild{}))

	switch {
	case len(bilder) != 2:
		t.Fatalf("%d Bilder", len(bilder))
	case bilder[0].Absage == nil || *bilder[0].Absage != (WunschabsageDTO{Befehl: "b1", Art: "objekt.tappen", Grund: "nicht erlaubt"}):
		t.Errorf("die Absage: %+v", bilder[0].Absage)
	case bilder[1].Absage != nil:
		t.Errorf("ohne Absage: %+v", bilder[1].Absage)
	}
}
