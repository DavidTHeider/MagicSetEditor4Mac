// **Die willigen Doubles des Bestands** — Setköpfe, Karten, Felder, Ausfuhr.
//
// **Eine eigene Datei**, weil `vorrichtung_test.go` über die Zeilengrenze
// wuchs, als das Anlegen eines Sets dazukam (§2). Der Schnitt folgt der Frage:
// dort die **Vorrichtung** — Dienste bauen, rufen, prüfen —, hier die
// **Antworten**, die die Doubles auf Bestandsfragen geben.
package rest_test

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (williger) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return []domain.SetHead{{ID: 1, Name: "Probe", Game: "magic", Style: "m15"}}, nil
}

func (williger) Karten(context.Context, int64) ([]*domain.CardRow, error) {
	return []*domain.CardRow{{
		ID: 2, SetID: 1, Number: 1,
		Values: []domain.Value{{Field: "name", Text: "Feuerball"}},
	}}, nil
}

func (williger) FeldSetzenMitVorwert(
	context.Context, int64, int64, string, string,
) (vorher string, hatte bool, fehler error) {
	return "vorher", true, nil
}

func (williger) Einfahren(
	context.Context, string,
) (setID int64, karten int, befunde []string, fehler error) {
	return 1, 3, []string{}, nil
}

func (williger) Sichern(context.Context, int64, string) ([]string, error) {
	return []string{}, nil
}

func (williger) Zeichnen(
	context.Context, int64, string, string,
) (breite, hoehe int, befunde []string, fehler error) {
	return 375, 523, nil, nil
}

func (williger) Bogen(
	context.Context, int64, string,
) (seiten, karten int, befunde []string, fehler error) {
	return 1, 3, []string{}, nil
}

func (williger) MTGJSONLesen(context.Context, string, string) (
	angelegt []int64, befunde []string, fehler error,
) {
	return []int64{7}, []string{}, nil
}

func (williger) MTGJSONSchreiben(context.Context, int64, string, string) (
	karten int, befunde []string, fehler error,
) {
	return 3, []string{}, nil
}

func (williger) MTGJSONSpiele() []string { return []string{"magic"} }

func (williger) Vorschlag(
	context.Context, int64, int64, string,
) (rest.Vorschlag, error) {
	return rest.Vorschlag{
		Aenderungen: []rest.Aenderung{{Feld: "rule text", Alt: "Alt", Neu: "Neu"}},
		Herkunft:    "https://beispiel.invalid/karte",
	}, nil
}

func (williger) Uebernehmen(
	context.Context, int64, int64, string, []string,
) (vorher []rest.Aenderung, fehler error) {
	return []rest.Aenderung{{Feld: "rule text", Alt: "Alt", Neu: "Neu"}}, nil
}

// **Die Antworten für den Volltext und den Lebenszyklus**
// (#1107).
func (williger) Felder(context.Context, int64) ([]string, error) {
	return []string{"name", "rule text"}, nil
}

func (williger) Suche(context.Context, domain.SearchQuery) (domain.SearchResult, error) {
	return domain.SearchResult{
		Hits: []domain.SearchHit{
			{CardID: 2, SetID: 1, Number: 1, Field: "name", Value: "Feuerball"},
		},
	}, nil
}

func (williger) KarteAnlegen(context.Context, int64) (int64, error) { return 7, nil }

func (williger) KarteKopieren(context.Context, int64, int64) (int64, error) { return 8, nil }

func (williger) KarteLoeschen(context.Context, int64, int64, bool) error { return nil }

func (williger) KartenTauschen(context.Context, int64, int64, int64) error { return nil }

// **Die Antworten für die Setangaben** (#1107).
func (williger) Angaben(context.Context, int64) ([]rest.Feldbeschreibung, []rest.Setwert, error) {
	return []rest.Feldbeschreibung{{Name: "title", Aufschrift: "Titel", Art: "text"}},
		[]rest.Setwert{{Feld: "title", Wert: "Probe"}}, nil
}

func (williger) Setzen(context.Context, int64, string, string) (
	vorher string, hatte bool, fehler error,
) {
	return "vorher", true, nil
}

func (williger) Vorbehalte(context.Context, int64) ([]rest.Vorbehalt, error) {
	return []rest.Vorbehalt{{Datei: "set", Zeile: 12, Inhalt: "unbekannt: x"}}, nil
}

func (williger) Beispielsets() ([]rest.Beispielset, error) {
	return []rest.Beispielset{{Name: "Probe", Pfad: "/bündel/saat/Probe.mse-set"}}, nil
}

// **Die Antworten für Stile, Pakete, Kennzahlen, Verlauf und Einstellungen**
// (#1107).
func (williger) Stile(context.Context, int64) ([]rest.Stil, error) {
	return []rest.Stil{{Paket: "probe-schlicht", Spiel: "probe", Gewaehlt: true}}, nil
}

func (williger) Folgen(context.Context, int64, string) (rest.Wechselfolgen, error) {
	return rest.Wechselfolgen{UngedeuteteFelder: []string{"rarity"}}, nil
}

func (williger) Wechseln(context.Context, int64, string) error { return nil }

func (williger) Optionen(context.Context, int64) ([]rest.Stiloption, error) {
	return []rest.Stiloption{{Name: "rand", Art: "choice", Wert: "schmal"}}, nil
}

func (williger) OptionSetzen(context.Context, int64, string, string) error { return nil }

func (williger) OptionZuruecksetzen(context.Context, int64, string) error { return nil }

func (williger) Liste(context.Context) ([]rest.Paket, []string, error) {
	// **Mit Art, Name und Pfad des Katalogs** (#1291): das Entfernen bestätigt
	// über den Pfad, und ein Double ohne ihn prüfte den Fehlerweg statt der
	// Fähigkeit.
	return []rest.Paket{{
		Art: "mse-style", Name: "probe", Version: "1", Pfad: "/pakete/probe.mse-style",
	}}, []string{}, nil
}

func (williger) Beziehen(context.Context, string, string, string, string) (rest.Bezug, error) {
	return rest.Bezug{Ort: "/pakete/probe", Stand: "abc1234", Pruefsumme: "sha256:…"}, nil
}

func (williger) Statistik(context.Context, int64) (rest.Setzahlen, error) {
	return rest.Setzahlen{
		Karten: 1, Felder: 1, Belegung: []rest.Zahl{{Name: "name", Karten: 1}},
	}, nil
}

func (williger) Verteilung(context.Context, int64, string) ([]rest.Wertzahl, error) {
	return []rest.Wertzahl{{Name: "Feuerball", Karten: 1, KartenIDs: []string{"4"}}}, nil
}

func (williger) Zuruecknehmen(context.Context, int64, string) (rest.Bewegung, error) {
	return rest.Bewegung{
		Bewegt: true, KarteID: "2", Feld: "name", Umfang: 1,
		Urheber: "kommandozeile",
	}, nil
}

func (williger) Wiederherstellen(context.Context, int64, string) (rest.Bewegung, error) {
	return rest.Bewegung{
		Bewegt: true, KarteID: "2", Feld: "name", Umfang: 1,
		Urheber: "kommandozeile",
	}, nil
}

func (williger) Alle(context.Context) ([]rest.Einstellung, error) {
	return []rest.Einstellung{{
		Schluessel: "sprache", Wert: "de", Werte: []string{"de", "en"}, Bereich: "anwendung",
	}}, nil
}

func (williger) EinstellungSetzen(context.Context, string, string) error { return nil }

// Die Zwischenablage (#1168).
func (williger) AblageKopieren(context.Context, int64, []int64) (string, error) {
	return "game_name: probe\ncard:\n", nil
}

func (williger) AblageAusschneiden(context.Context, int64, []int64) (string, error) {
	return "game_name: probe\ncard:\n", nil
}

func (williger) AblageEinfuegen(context.Context, int64, string) (
	karten, befunde []string, fehler error,
) {
	return []string{"100"}, nil, nil
}

// Die Verweise auf echte Drucke (#1167).
func (williger) VerweisSetzen(_ context.Context, _ int64, printing string) (
	rest.Verweissicht, error,
) {
	return rest.Verweissicht{
		KarteID: "2", Printing: printing, Name: "Lightning Bolt",
	}, nil
}

func (williger) Verweis(context.Context, int64) (*rest.Verweissicht, error) {
	return &rest.Verweissicht{KarteID: "2", Printing: "probe-druck", Name: "Lightning Bolt"}, nil
}

func (williger) VerweisLoesen(context.Context, int64) error { return nil }

// **Und die Marken** (#1208).
func (williger) MarkeSetzen(context.Context, int64, int64, string) error { return nil }

func (williger) MarkeEntfernen(context.Context, int64, int64, string) error { return nil }

// **`nil` und nicht `[]string{}`** — die Antwort muss daraus eine leere Liste
// machen, sonst käme über die Fläche `null` an (#1208).
func (williger) KartenMarken(context.Context, int64, int64) ([]string, error) {
	return nil, nil
}

func (williger) Setmarken(context.Context, int64) ([]string, error) {
	return []string{"fertig"}, nil
}

func (williger) Verweise(context.Context, int64) ([]rest.Verweissicht, error) {
	return []rest.Verweissicht{{KarteID: "2", Printing: "probe-druck"}}, nil
}

// Kartensuche gibt einen festen Treffer (#1167).
func (williger) Kartensuche(context.Context, string) (rest.Suchergebnis, error) {
	return rest.Suchergebnis{
		Treffer: []rest.Suchtreffersicht{{Kennung: "abc", Name: "Lightning Bolt"}},
	}, nil
}

// Dienstkatalog gibt eine feste Fähigkeit (#1440).
func (williger) Grenzen(context.Context) (rest.Grenzen, error) {
	return rest.GrenzenVon(domain.Vorgabegrenzen()), nil
}

func (williger) Dienstkatalog(context.Context) ([]rest.Dienstfaehigkeit, error) {
	return []rest.Dienstfaehigkeit{{Dienst: "scryfall", Frage: "suche", Status: "bereit"}}, nil
}

// Das Ersetzen glückt und findet eine Stelle (#1168).
func (williger) ErsetzungVorschau(
	context.Context, int64, domain.Ersetzungsfrage,
) (domain.Ersetzungsvorschau, error) {
	return domain.Ersetzungsvorschau{
		Stellen: []domain.Ersetzungsstelle{
			{KarteID: 2, Nummer: 0, Feld: "name", Vorher: "a", Nachher: "b", Treffer: 1},
		},
		Karten: 1, Treffer: 1,
	}, nil
}

func (w williger) ErsetzungAusfuehren(
	ctx context.Context, setID int64, frage domain.Ersetzungsfrage,
) (domain.Ersetzungsvorschau, error) {
	return w.ErsetzungVorschau(ctx, setID, frage)
}

// Die Kartenliste glückt und nennt eine Spalte (#1168).

func (williger) Listenaufbau(context.Context, int64) (
	domain.Kartenlistenaufbau, error,
) {
	return domain.Kartenlistenaufbau{
		Spalten: []domain.Listenspalte{
			{Feld: "name", Aufschrift: "Name", Sichtbar: true},
		},
		Zeilen: []domain.Listenzeile{
			{KarteID: 2, Werte: map[string]string{"name": "Feuerball"}},
		},
		Ordnung: domain.Kartenordnung{Feld: "name", Aufsteigend: true},
	}, nil
}

func (williger) ListenspaltenSetzen(
	context.Context, int64, []domain.Spaltenwahl,
) error {
	return nil
}

func (williger) ListensortierungSetzen(
	context.Context, int64, domain.Kartenordnung,
) error {
	return nil
}

func (williger) PaketInstallieren(context.Context, string) (rest.Paket, error) {
	return rest.Paket{Art: "mse-style", Name: "neu", Version: "2026-08-30"}, nil
}

func (williger) PaketAnlegen(_ context.Context, art, name, _, _ string) (rest.Paket, error) {
	return rest.Paket{Art: art, Name: name, Version: "2026-09-04"}, nil
}

func (williger) PaketKopieren(_ context.Context, art, name, neu string) (rest.Paket, error) {
	return rest.Paket{Art: art, Name: name + "-" + neu}, nil
}

func (williger) Nutzer(context.Context) ([]rest.Nutzer, error) {
	return []rest.Nutzer{{ID: "1", Name: "Ich", Profile: []rest.Profil{}}}, nil
}

func (williger) NutzerAnlegen(_ context.Context, name, bild string) (rest.Nutzer, error) {
	return rest.Nutzer{ID: "2", Name: name, Bild: bild, Profile: []rest.Profil{}}, nil
}

func (williger) ProfilAnlegen(_ context.Context, nutzerID, name string) (rest.Profil, error) {
	return rest.Profil{ID: "20", NutzerID: nutzerID, Name: name}, nil
}

func (williger) GezeigtesProfil(context.Context) (rest.Profil, error) {
	return rest.Profil{ID: "10", NutzerID: "1", Name: "Standard"}, nil
}

// WerArbeitet nennt den Menschen, der arbeitet — Werkstück 17 von Phase 4.12.
func (williger) WerArbeitet(context.Context) (string, error) { return "7", nil }

// Arbeitskontext sagt, wer wo als was arbeitet — Werkstück O1 von Phase 5.2.
func (williger) Arbeitskontext(context.Context) (rest.Arbeitskontext, error) {
	return rest.Arbeitskontext{
		NutzerID: "7", ProfilID: "10", EignerID: "7", Rolle: "verwalter",
		Modus: "dienst", Revision: 3,
	}, nil
}

// NutzerWechseln stellt um, wer arbeitet.
func (williger) NutzerWechseln(_ context.Context, nutzerID, _ string) (rest.Profil, error) {
	return rest.Profil{ID: "1", NutzerID: nutzerID, Name: "Erstes"}, nil
}

func (williger) ProfilWechseln(_ context.Context, profilID, _ string) (rest.Profil, error) {
	return rest.Profil{ID: profilID, NutzerID: "1", Name: "Gewechselt"}, nil
}

func (williger) Anbieterkonten(context.Context, string) ([]rest.Konto, error) {
	return []rest.Konto{{Anbieter: "github", Anzeigename: "Wer"}}, nil
}

func (williger) KontoLoesen(context.Context, string, string) error { return nil }

func (williger) NutzerAendern(_ context.Context, angaben rest.Nutzerangaben) (
	rest.Nutzer, error,
) {
	return rest.Nutzer{
		ID: angaben.NutzerID, Name: angaben.Name, Bild: angaben.Bild,
		Vorname: angaben.Vorname, Fassung: "2", Profile: []rest.Profil{},
	}, nil
}

func (williger) ProfilAendern(_ context.Context, angaben rest.Profilangaben) (
	rest.Profil, error,
) {
	return rest.Profil{
		ID: angaben.ProfilID, NutzerID: "1", Name: angaben.Name,

		Subslogan:         angaben.Subslogan,
		Kurzebeschreibung: angaben.Kurzebeschreibung,
		Langebeschreibung: angaben.Langebeschreibung,
		Badge:             angaben.Badge,

		Fassung: "2",
	}, nil
}

func (williger) PaketEntfernen(context.Context, string, string, string) error { return nil }

func (williger) PaketUebernehmen(context.Context, string) (rest.Uebernahme, error) {
	return rest.Uebernahme{
		Uebernommen: []rest.Paket{{Art: "mse-game", Name: "magic"}},
		Uebergangen: []string{},
		Befunde:     []string{},
	}, nil
}

func (williger) MSE2Orte() []string { return []string{"/Anwendungen/Magic Set Editor/data"} }

func (williger) AusDatei(_ context.Context, _, _ int64, feld, _ string) (rest.Bildmasse, error) {
	return rest.Bildmasse{
		Feld: feld, QuelleBreite: 400, QuelleHoehe: 300,
		ZielBreite: 315, ZielHoehe: 230,
	}, nil
}
