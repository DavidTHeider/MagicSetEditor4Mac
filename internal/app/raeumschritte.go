package app

import (
	"context"
	"fmt"
	"log/slog"
)

// **Jeder Schritt des Räumens für sich** — Befund 2 des Nachlaufs über
// Phase 1 von 0.9.0 (F02, #2341).
//
// Bis dahin liefen die Schritte hintereinander in **einem** Schritt von
// [innerhalb]: warf einer, fing `innerhalb` die Panik, und die übrigen
// entfielen — nach einer Panik am fremden Tisch blieben Gerätemodell,
// Peersitzungen, Horchen und Dienstposten offen. Der Schluss schloss den
// Bestand trotzdem, aber nicht, was neben ihm lief. Jetzt fängt jeder seine
// eigene, und der Schluss nennt, welcher es war.

// raeumschritt ist ein Schritt des Räumens mit dem, was er räumt.
type raeumschritt struct {
	was string
	tue func()
}

// schritteFahren fährt jeden Schritt, auch hinter einem, der in Panik endet,
// und gibt die Namen derer, die es taten.
func schritteFahren(schritte []raeumschritt) []string {
	var verloren []string

	for _, schritt := range schritte {
		if !gefangen(schritt) {
			verloren = append(verloren, schritt.was)
		}
	}

	return verloren
}

// gefangen fährt einen Schritt und fängt seine Panik.
func gefangen(schritt raeumschritt) (erledigt bool) {
	defer func() {
		if wert := recover(); wert != nil {
			slog.Default().Error("ein Schritt des Räumens brach mit einer Panik ab",
				"art", schritt.was, "panik", fmt.Sprint(wert))
		}
	}()

	schritt.tue()

	return true
}

// eigeneLaeufeBeenden beendet, was die Anwendung selbst laufen lässt, und
// gibt, welcher Schritt dabei in Panik endete.
//
// **Dieselben Schritte wie beim Wechsel des Profils** ([App.profilgebundenesRaeumen])
// und davor der Tisch als Gast — er gehört keinem Profil, sondern diesem
// Programm, und beim Wechsel bleibt er.
func (a *App) eigeneLaeufeBeenden(ctx context.Context) []string {
	verloren := schritteFahren([]raeumschritt{
		{"der Tisch als Gast", func() { a.partiebeitrittBeenden(ctx) }},

		// **Und der Plan der Sicherung** (OPS01, #2347): er gehört dem
		// Bestand, nicht dem Profil, und vor dem Tausch muss er fort sein.
		{"der Plan der Sicherung", a.sicherungsplanBeenden},
	})

	return append(verloren, a.profilgebundenesRaeumen()...)
}

// profilgebundenesRaeumen beendet, was einem Profil gehört — beim Wechsel,
// und seit ID-02 auch beim Sperren und Abmelden: gesperrt darf nichts mehr
// für das Profil arbeiten, das eben verschlossen wurde (Zweitblick über
// ID-02, Befund 6).
//
// **Jeder Schritt für sich** ([schritteFahren]): wirft einer, laufen die
// übrigen trotzdem, und zurück kommt, welche es waren.
func (a *App) profilgebundenesRaeumen() []string {
	return schritteFahren([]raeumschritt{
		// **Und die offenen KI-Vorschläge gehen mit** (§7, Regel 1): sie gehören
		// dem Profil, das sie angelegt hat — [App.kiwerkRaeumen] sagt, warum das
		// Räumen neben der Grenze am Auftrag steht und nicht statt ihrer.
		{"die KI-Vorschläge", a.kiwerkRaeumen},

		// **Und `fm serve` endet** (AI-05.AK-07, #2042): das nächste Profil hat
		// das Gerätemodell vielleicht aus. Wer es braucht, startet es neu — bei
		// Bedarf, wie beim ersten Mal.
		{"das Gerätemodell", a.geraetBeenden},

		// **Und die offenen Peersitzungen ebenso** (N1, #1454). Das Issue
		// verlangt es wörtlich: „Bei Nutzer-/Profilwechsel Sessions pausieren
		// und Schlüsselkontext schließen, verspätete Antworten verwerfen."
		{"die Peersitzungen", a.sitzungenPausieren},

		// **Und das Horchen hört auf** (N2, #1455).
		//
		// **Weil der Lauscher dem Profil gehört, das ihn geöffnet hat.** Er
		// trägt dessen Identität in seinem Zeugnis und fragt dessen
		// Kontaktliste; wer ihn über den Wechsel hinweg offen ließe, machte das
		// eine Profil unter dem Schlüssel des anderen erreichbar. Wer weiter
		// erreichbar sein will, sagt es noch einmal — das ist die „bewusste
		// Auswahl" aus dem Issue.
		{"das Horchen", a.horchenBeenden},

		// **Und das Fernhorchen ebenso** (Nachlauf über Phase 6.7, N2) — aus
		// demselben Grund: es nimmt mit dem Schlüssel und der Kontaktliste des
		// Profils an, das es geöffnet hat.
		{"das Fernhorchen", a.fernhorchenBeenden},

		// **Und die Spielbilder gehen mit** (G2,
		// #1459).
		//
		// **Weil sie dem Profil gehören, das sie gezeichnet hat.** Das Issue
		// nennt den Profilwechsel neben dem Speicherbudget, und beides hängt
		// zusammen: ein Fach, das über den Wechsel hinweg bliebe, hielte bis zu
		// 128 MB für Karten, die niemand mehr sieht.
		//
		// **Ein falsches Bild käme dabei nicht heraus**, denn der Schlüssel
		// steht über dem Inhalt — dieselbe Karte in einem anderen Profil ist
		// ein anderer Schlüssel. Das ist der Grund, warum hier **Speicher** der
		// Grund ist und nicht Verwechslung; wer beides in einen Satz zöge,
		// hätte eine Begründung, die nicht trägt (§10).
		{"die Spielbilder", a.spielbilderRaeumen},

		// **Und die Vorschauen ebenso** — OPS-08.AK-02 (#1725): dieselbe Regel.
		// Ihr Schlüssel ist der Inhalt, ein falsches Bild käme also nicht heraus;
		// aber eine Adresse, die im vorigen Profil entstand, soll im nächsten
		// nichts mehr liefern.
		{"die Vorschauen", a.bildfach.Leere},

		// **Und die Dienstposten ebenso** (N6, #1572).
		//
		// **Aus demselben Grund**: ein REST-Dienst, der über den Wechsel hinweg
		// liefe, bediente Anfragen für einen Bestandsausschnitt, den niemand
		// mehr sieht — und unter einem Zugang, den jemand für das **andere**
		// Profil bekommen hat. Wer weiter erreichbar sein will, sagt es noch
		// einmal.
		{"die Dienstposten", a.diensteBeenden},

		// **Und die vorbereiteten Teilungen** (Zweitblick über Phase 6.7, Z8):
		// ihre Kopien tragen Einladung, Sammlung oder Set des vorigen Profils.
		{"die Teilungen", a.teilungenRaeumen},
	})
}
