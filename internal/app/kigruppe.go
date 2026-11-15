package app

// **Die atomare Gruppe an der Naht** — Werkstück A13.1b von Phase 5.3
// (#1452).
//
// # Was hier geschieht und was nicht
//
// **Hier wird umgerechnet, nicht entschieden.** Welche Vorschläge eine Gruppe
// bilden, ob sie noch gelten und ob die Auswahl dazu passt, entscheidet
// [service.KIAuftragswerk.GruppeUebernehmen]; diese Datei macht aus
// `map[string][]string` die Werte des Bestands und legt das Set dazu, das die
// Aufsicht braucht.
//
// # Und warum das Set hier geholt wird
//
// **Weil nur hier jemand die Karte lesen kann.** `service` kennt den
// Kartenleser nicht, und die Aufsicht will das Set — sonst prüfte sie gegen
// eine Zahl, die der Aufrufer selbst gewählt hat.

import (
	"context"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// FelderGruppeSetzen schreibt eine Gruppe von Übernahmen in **einem** Zug.
func (k kikarten) FelderGruppeSetzen(
	ctx context.Context, schritte []service.KIGruppenschritt,
) error {
	if k.app.kartenleser == nil || k.app.schreiber == nil {
		return fehlenderBestand()
	}

	// **Erst alle lesen, dann schreiben.** Eine Karte, die es nicht mehr
	// gibt, soll auffallen, bevor die erste geschrieben ist — dieselbe
	// Reihenfolge wie in `service`, und aus demselben Grund.
	umgerechnet := make([]domain.Kartenschritt, 0, len(schritte))

	for _, schritt := range schritte {
		zeile, fehler := k.app.kartenleser.KarteLesen(ctx, schritt.Karte)
		if fehler != nil {
			return fehler
		}

		werte := alsWerte(schritt.Felder)

		umgerechnet = append(umgerechnet, domain.Kartenschritt{
			KarteID: schritt.Karte, SetID: zeile.SetID, Werte: werte,
			Herkunft: kiherkunft,

			// **Der Abdruck des Geschriebenen gehört dazu** (A11, #1450):
			// die Herkunft sagt nicht nur, wer geantwortet hat, sondern
			// **worauf** — und das ist der Wert, der wirklich hinging.
			Herkuenfte: mitAbdruckDesGeschriebenen(schritt.Herkunft, werte),

			// **Und der erwartete Stand reist mit** (Befund 2 des
			// Zweitblicks über Phase 5.3) — geprüft wird er in der
			// Schreibtransaktion, nicht hier.
			Erwartet: schritt.Erwartet,
		})
	}

	return k.app.schreiber.GruppeVonAussenMitHerkunft(ctx, umgerechnet)
}

// KIGruppenwahlDTO ist ein Vorschlag mit den Feldern, die gelten sollen.
type KIGruppenwahlDTO struct {
	Kennung string   `json:"kennung"`
	Felder  []string `json:"felder"`
}

// KIGruppenantwort sagt, was aus der Gruppe wurde.
type KIGruppenantwort struct {
	// Karten ist die Zahl der geschriebenen Karten.
	Karten int `json:"karten"`

	// Felder ist die Zahl der geschriebenen Felder über alle Karten.
	Felder int `json:"felder"`

	// Namen sind die geschriebenen Feldnamen je Karte — leer bei einem
	// Fehler, und zwar als **leere Liste**: „nichts geschrieben" und „die
	// Frage ging schief" sind zwei Auskünfte.
	Namen []KIGruppenkarteDTO `json:"namen"`

	Fehler *Fehler `json:"fehler"`
}

// KIGruppenkarteDTO nennt eine Karte und ihre geschriebenen Felder.
type KIGruppenkarteDTO struct {
	KarteID int64    `json:"karteId"`
	Felder  []string `json:"felder"`
}

// KIGruppeUebernehmen schreibt mehrere Vorschläge als **eine** Gruppe.
//
// **Ganz oder gar nicht** ([service.KIAuftragswerk.GruppeUebernehmen]):
// scheitert ein Schritt, ist nichts geschrieben und kein Schlüssel
// verbraucht.
func (a *App) KIGruppeUebernehmen(
	ctx context.Context, wahl []KIGruppenwahlDTO,
) KIGruppenantwort {
	ctx = a.imProfil(ctx)

	werk, fehler := a.kiwerkOhneQuelle(ctx)
	if fehler != nil {
		return gruppenfehler(ctx, fehler)
	}

	gewaehlt := make([]service.KIGruppenwahl, 0, len(wahl))
	for _, eine := range wahl {
		gewaehlt = append(gewaehlt, service.KIGruppenwahl{
			Kennung: eine.Kennung, Felder: eine.Felder,
		})
	}

	profil, fehler := storage.ProfilAus(ctx)
	if fehler != nil {
		return gruppenfehler(ctx, fehler)
	}

	// **Dieselbe Profilgrenze wie bei der einzelnen Übernahme** (Befund 1 des
	// Nachlaufs über Phase 5.3).
	uebernahme, fehler := werk.GruppeUebernehmen(ctx, gewaehlt, profil)
	if fehler != nil {
		return gruppenfehler(ctx, fehler)
	}

	return KIGruppenantwort{
		Karten: uebernahme.Karten, Felder: uebernahme.Felder,
		Namen: alsGruppenkarten(uebernahme.Namen),
	}
}

// gruppenfehler gibt eine Antwort mit Fehler — **und leeren Listen**.
func gruppenfehler(ctx context.Context, fehler error) KIGruppenantwort {
	return KIGruppenantwort{
		Namen: []KIGruppenkarteDTO{}, Fehler: Uebersetze(ctx, fehler),
	}
}

// alsGruppenkarten macht aus der Abbildung eine geordnete Liste.
//
// **Geordnet nach Kartenkennung** — eine Abbildung hat keine Reihenfolge, und
// eine Antwort, deren Zeilen bei jedem Lauf anders stehen, ist nicht zu
// vergleichen.
func alsGruppenkarten(namen map[int64][]string) []KIGruppenkarteDTO {
	karten := make([]int64, 0, len(namen))
	for karte := range namen {
		karten = append(karten, karte)
	}

	slices.Sort(karten)

	zeilen := make([]KIGruppenkarteDTO, 0, len(karten))
	for _, karte := range karten {
		zeilen = append(zeilen, KIGruppenkarteDTO{
			KarteID: karte, Felder: namen[karte],
		})
	}

	return zeilen
}

// KIGruppeUebernehmen ist die Fensterbindung.
func (b *Bindungen) KIGruppeUebernehmen(ctx context.Context, wahl []KIGruppenwahlDTO) KIGruppenantwort {
	return b.app(ctx).KIGruppeUebernehmen(b.ruf(ctx), wahl)
}

// KIBatchVorschlagen ist die Fensterbindung des Batches (A13.1b, #1452).
//
// # Warum sie erst jetzt entsteht
//
// **Weil es den Kasten vorher nicht gab.** A13.1 hat den Batch an drei
// Flächen gebaut und am Fenster ausgelassen — mit Grund: welche Karten
// gemeint sind, steht in einer Auswahl, und eine gebundene Methode ohne sie
// wäre eine Zusage ohne Einlösung (§2). Mit [KIBatchkasten] gibt es die
// Auswahl, und damit gibt es die Bindung.
//
// **Die Kennungen kommen als Zeichenketten**, wie überall am Fenster: eine
// JavaScript-Zahl trägt keine 64 Bit.
func (b *Bindungen) KIBatchVorschlagen(
	ctx context.Context,
	karteIDs []string, anbieter, alias, auftrag, nachweis string,
) KIBatchAntwort {
	b.tor(ctx)

	ziele := make([]int64, 0, len(karteIDs))

	for _, roh := range karteIDs {
		kennung, befund := benannteKennungAus("karteIds", roh)
		if befund != nil {
			return KIBatchAntwort{Ziele: []KIBatchlage{}, Fehler: befund}
		}

		ziele = append(ziele, kennung)
	}

	return b.app(ctx).KIBatchVorschlagen(b.ruf(ctx), ziele, anbieter, alias,
		auftrag, nachweis)
}
