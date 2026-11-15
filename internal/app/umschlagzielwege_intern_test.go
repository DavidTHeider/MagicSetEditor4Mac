package app

// **Die Fehlerwege von Karte und Vorlage** —
// #1639.
//
// # Warum sie einen eigenen Lauf bekommen
//
// **Weil sie sonst niemand betritt.** Der Abnahmelauf nebenan fährt die Wege,
// die glücken; §10 sagt dazu: „wer nur den Erfolgsweg prüft, hat die Hälfte
// ausgelassen." Und jeder dieser Wege endet in einem Satz, den ein Mensch
// liest.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/austausch"
)

// **Jeder Weg, der nicht geht, sagt es.**
func TestJederZielwegDerNichtGehtSagtEs(t *testing.T) {
	anwendung := mitIdentitaet(t)
	ctx := t.Context()

	fehlt := filepath.Join(t.TempDir(), "gibtesnicht"+austausch.Endung)

	if antwort := anwendung.UmschlagKarte(
		ctx, fehlt, "1"); antwort.Fehler == nil {
		t.Error("der fehlende Umschlag kam als Karte durch")
	}

	if antwort := anwendung.UmschlagVorlage(
		ctx, fehlt, "angenommen.mse-game"); antwort.Fehler == nil {
		t.Error("der fehlende Umschlag kam als Vorlage durch")
	}

	// **Ein Umschlag mit einem Deck trägt weder Karte noch Vorlage.**
	ich, fehler := anwendung.peerwerk().Ausweis(anwendung.imProfil(ctx))
	if fehler != nil {
		t.Fatalf("Ausweis: %v", fehler)
	}

	mitDeck := einUmschlagsziel(t, "nurdeck")

	if _, fehler := austausch.Schreibe(mitDeck, []austausch.Inhalt{{
		Art: austausch.ArtDeck, Name: Deckdatei, Bezeichnung: Deckdatei,
		Bytes: []byte("{}"),
	}}, ich); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if antwort := anwendung.UmschlagKarte(
		ctx, mitDeck, "1"); antwort.Fehler == nil {
		t.Error("das Deck kam als Karte durch")
	}

	if antwort := anwendung.UmschlagVorlage(
		ctx, mitDeck, "angenommen.mse-game"); antwort.Fehler == nil {
		t.Error("das Deck kam als Vorlage durch")
	}

	// **Eine Karte, die keine ist** — `austausch.Lies` prüft Prüfsummen und
	// nicht den Sinn.
	krumm := einUmschlagsziel(t, "krummekarte")

	if _, fehler := austausch.Schreibe(krumm, []austausch.Inhalt{{
		Art: austausch.ArtKarte, Name: Kartendatei, Bezeichnung: Kartendatei,
		Bytes: []byte("kein JSON"),
	}}, ich); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if antwort := anwendung.UmschlagKarte(
		ctx, krumm, "1"); antwort.Fehler == nil {
		t.Error("die krumme Karte kam durch")
	}
}

// **Ein Set, das es nicht gibt, nimmt keine Karte auf.**
//
// **Und ein Feld ohne Namen ist ein Befund** (§8): die Karte steht danach da,
// und ihre Kennung kommt mit zurück — dieselbe Gestalt wie bei einem
// Deckplatz, den niemand legen kann.
//
// **Ein Feld, das die Vorlage nicht kennt, ist übrigens keiner**: ein
// Kartenfeld ist ein Name mit einem Wert, und was davon eine Vorlage zeigt,
// entscheidet sie beim Zeichnen. Gemessen am 2026-09-19 — `gibtesnicht` geht
// durch.
func TestEineKarteMitEinemUnmoeglichenFeldMeldetSich(t *testing.T) {
	anwendung := mitIdentitaet(t)
	ctx := t.Context()

	if antwort := anwendung.karteAnnehmen(
		ctx, "gibtesnicht", KarteDTO{}); antwort.Fehler == nil {
		t.Error("das fehlende Set nahm eine Karte auf")
	}

	geoeffnet := anwendung.SetOeffnen(ctx, saatset(t))
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	antwort := anwendung.karteAnnehmen(ctx, geoeffnet.SetID, KarteDTO{
		Werte: []WertDTO{{Feld: "  ", Wert: "x"}},
	})
	if antwort.Fehler == nil {
		t.Fatal("das namenlose Feld kam durch")
	}

	if antwort.KarteID == "" {
		t.Error("die Karte kam nicht mit zurück — sie steht im Set")
	}
}

// **Ein Paketname, der ein Verzeichnis nennt, wird abgewiesen.**
//
// **Weil er von einem Menschen kommt**: ein `../` darin legte ein Paket
// irgendwohin.
func TestEinPaketnameDerEinVerzeichnisNenntWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	sendung := geleseneSendung{ort: t.TempDir()}

	for _, name := range []string{"", "..", "ordner/paket.mse-game"} {
		if _, fehler := vorlagePacken(sendung, name, nil); fehler == nil {
			t.Errorf("der Name %q kam durch", name)
		}
	}
}

// **Eine Vorlagendatei, die nicht dasteht, hält die Annahme auf.**
//
// **Über eine Sendung von Hand**: `austausch.Lies` prüft jede Prüfsumme, und
// ein gültiger Umschlag hat seine Dateien. Der Fall ist der Wettlauf mit dem
// Aufräumer — und die Antwort darauf ist eine Klage und keine halbe Vorlage.
func TestEineFehlendeVorlagendateiHaeltDieAnnahmeAuf(t *testing.T) {
	t.Parallel()

	sendung := geleseneSendung{ort: t.TempDir()}
	sendung.Verzeichnis = filepath.Join(sendung.ort, "gibtesnicht")
	sendung.Manifest = austausch.Manifest{Objekte: []austausch.Objekt{{
		Art: austausch.ArtVorlage, Name: "vorlagen/spiel",
	}}}

	_, fehler := vorlagenAusSendung(sendung)
	if fehler == nil {
		t.Fatal("die fehlende Datei kam durch")
	}

	// **Und liegt das Verzeichnis da, fehlt nur die Datei** — der zweite
	// Zweig, und er nennt sie: an ihr lässt sie sich suchen.
	sendung.Verzeichnis = sendung.ort

	zweiter, fehler := vorlagenAusSendung(sendung)
	if fehler == nil {
		t.Fatalf("die fehlende Datei kam beim zweiten Weg durch: %+v", zweiter)
	}

	if !strings.Contains(fehler.Error(), "vorlagen/spiel") {
		t.Errorf("die Klage nennt die Datei nicht: %v", fehler)
	}
}

// **Und die vier Bindungen halten ihre drei Ausgänge.**
//
// **Abgebrochen, klagend, ins Leere** — dieselben drei wie bei `UmschlagDeck`
// und `UmschlagSicht`, und aus demselben Grund: jeder endet in einem Satz.
func TestDieZielbindungenHaltenIhreDreiAusgaenge(t *testing.T) {
	anwendung := mitIdentitaet(t)
	bindungen := NeueBindungen(anwendung)

	anwendung.MitDateiwahl(func(context.Context, Dateiauswahl) (string, error) {
		return "", nil
	})

	if antwort := bindungen.UmschlagKarte(t.Context(), "1"); antwort.Fehler != nil {
		t.Errorf("der Abbruch gab einen Fehler: %+v", antwort.Fehler)
	}

	if antwort := bindungen.UmschlagVorlage(t.Context(),
		"angenommen.mse-game"); antwort.Fehler != nil {
		t.Errorf("der Abbruch gab einen Fehler: %+v", antwort.Fehler)
	}

	anwendung.MitDateiwahl(func(context.Context, Dateiauswahl) (string, error) {
		return "", errProbedialog
	})

	if antwort := bindungen.UmschlagKarte(t.Context(), "1"); antwort.Fehler == nil {
		t.Error("die Klage des Dialogs ging verloren")
	}

	if antwort := bindungen.UmschlagVorlage(t.Context(),
		"angenommen.mse-game"); antwort.Fehler == nil {
		t.Error("die Klage des Dialogs ging verloren")
	}

	anwendung.MitDateiwahl(func(context.Context, Dateiauswahl) (string, error) {
		return filepath.Join(t.TempDir(), "gibtesnicht"+austausch.Endung), nil
	})

	if antwort := bindungen.UmschlagKarte(t.Context(), "1"); antwort.Fehler == nil {
		t.Error("der fehlende Umschlag kam durch")
	}

	if antwort := bindungen.UmschlagVorlage(t.Context(),
		"angenommen.mse-game"); antwort.Fehler == nil {
		t.Error("der fehlende Umschlag kam durch")
	}
}

// **Jede Stelle, an der das Packen klemmen kann, klemmt.**
//
// **Tabellengetrieben, weil es dieselbe Zusage dreimal ist** (§3): was nicht
// ganz in ein Paket kommt, kommt gar nicht hinein — eine halbe Vorlage sähe
// aus wie eine ganze.
func TestJedeStelleDesPackensKannKlemmen(t *testing.T) {
	t.Parallel()

	// **Der Wegwerfort ist weg** — der Wettlauf mit dem Aufräumer.
	ohneOrt := geleseneSendung{ort: filepath.Join(t.TempDir(), "gibtesnicht")}

	if _, fehler := vorlagePacken(
		ohneOrt, "angenommen.mse-game", nil); fehler == nil {
		t.Error("der fehlende Ort kam durch")
	}

	// **Der Name liegt schon als Datei da** — dann wird daraus kein
	// Verzeichnis.
	belegt := geleseneSendung{ort: t.TempDir()}

	if fehler := os.WriteFile(
		filepath.Join(belegt.ort, "angenommen.mse-game"),
		[]byte("keine Vorlage"), 0o600); fehler != nil {
		t.Fatalf("die Datei legen: %v", fehler)
	}

	if _, fehler := vorlagePacken(
		belegt, "angenommen.mse-game", nil); fehler == nil {
		t.Error("der belegte Name kam durch")
	}

	// **Und eine Datei, die in ein Unterverzeichnis will, das es nicht
	// gibt.** Aus dem Manifest kommt sie so nicht — `path.Base` nimmt den
	// Weg weg —, und genau deshalb steht der Zweig hier: er hält die Zusage,
	// dass nichts halb hineinkommt.
	frisch := geleseneSendung{ort: t.TempDir()}

	if _, fehler := vorlagePacken(frisch, "angenommen.mse-game",
		[]Vorlagendatei{{Name: "unter/datei", Inhalt: "x"}}); fehler == nil {
		t.Error("die Datei im fehlenden Unterverzeichnis kam durch")
	}
}

// **Und ein krummer Name endet vor der Installation.**
//
// **Über den ganzen Weg und nicht nur über [vorlagePacken]**: gemessen wird,
// dass die Absage bis zur Antwort durchkommt.
func TestEinKrummerPaketnameEndetVorDerInstallation(t *testing.T) {
	anwendung := mitIdentitaet(t)
	ctx := t.Context()

	anwendung.paketdateien = werkstattUeber(t, saatpaket(t))

	ziel := einUmschlagsziel(t, "vorlagekrumm")

	if geschnuert := anwendung.UmschlagSchnueren(
		ctx, "vorlage", "probe.mse-game", ziel); geschnuert.Fehler != nil {
		t.Fatalf("schnüren: %+v", geschnuert.Fehler)
	}

	if antwort := anwendung.UmschlagVorlage(
		ctx, ziel, "../woanders"); antwort.Fehler == nil {
		t.Error("der krumme Name kam durch")
	}
}
