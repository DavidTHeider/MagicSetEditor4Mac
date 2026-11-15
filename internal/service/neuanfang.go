package service

// **Ganz von vorn** — Werkstück O4 von Phase 5.2
// (#1428).
//
// # Warum das ein Tausch ist und keine Löschung
//
// „Die ganze Anwendung zurücksetzen" hieße naheliegenderweise: aus jeder
// Tabelle alles löschen. Das wären dreißig Anweisungen, deren Reihenfolge
// niemand nachvollzieht, und beim ersten neuen Schema wäre eine vergessen.
//
// **Was stattdessen geschieht**: der Bestand geht **beiseite**, und der
// nächste Start legt einen frischen an — dieselbe Mechanik, die das
// Zurückspielen benutzt, nur ohne eingehende Datei. Das hat drei
// Eigenschaften, die eine Löschschleife nicht hat:
//
//  1. **Es ist umkehrbar.** Der alte Bestand liegt als
//     `bestand-vor-rueckspielung.db` daneben, und `SicherungZurueckspielen`
//     holt ihn zurück. Die Abnahme verlangt genau das: „Abbruch während
//     Betrieb hinterlässt wiederaufnehmbaren, verständlich gemeldeten
//     Zustand."
//  2. **Es ist vollständig, ohne aufgezählt zu sein.** Was in einer neuen
//     Fassung dazukommt, ist automatisch mit weg.
//  3. **Es fasst nichts an, was der Anwendung nicht gehört.** Die Abnahme
//     sagt: „Unbekannte Dateien und Daten anderer Programme bleiben
//     unangetastet." Bewegt werden zwei Dinge — die Bestandsdatei und das
//     Bilderverzeichnis.
//
// **Die Vorlagenpakete bleiben.** Sie gehören dem **Rechner** und nicht dem
// Bestand (`docs/AUTORISIERUNG.md`), sie sind hunderte Megabyte, und wer sie
// loswerden will, nimmt sie auf der Paketseite heraus. Ein Zurücksetzen, das
// sie stillschweigend mitnähme, kostete einen Nachmittag Download.

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
)

// Neuanfang legt den Bestand beiseite und öffnet einen frischen.
//
// **Die Reihenfolge ist dieselbe wie beim Zurückspielen** — schließen, allein
// sein, Marke, tauschen, Bilder, wieder öffnen —, und aus denselben Gründen.
// Was fehlt, ist die eingehende Datei: es kommt nichts herein.
func (b *Bestandsweg) Neuanfang(ctx context.Context) (string, error) {
	if fehler := b.schliessen(); fehler != nil {
		return "", fmt.Errorf("den Bestand schließen: %w", fehler)
	}

	// **Erst schließen, dann fragen, ob wir allein sind** — der Grund steht
	// bei [Bestandsweg.Zurueckspielen].
	anspruch, fehler := b.allein(ctx, b.orte.Datenbank)
	if fehler != nil {
		return "", errors.Join(fehler, b.oeffnen(context.WithoutCancel(ctx)))
	}

	zurueckgeben := einmalig(anspruch)
	defer zurueckgeben()

	// **Die Marke steht vor dem ersten Schritt** (#1162): zwei Umbenennungen
	// in zwei Verzeichnissen, und kein Dateisystem führt sie zusammen aus.
	// **Sie beschreibt hier ein Leeren** — es kommen keine Bilder herein.
	if markenfehler := b.markeSchreiben(rueckspielstand{
		Quelle: "Neuanfang", Bilder: BilderGeleert,
	}); markenfehler != nil {
		return "", errors.Join(markenfehler, b.wiederOeffnen(ctx, zurueckgeben))
	}

	defer func() { b.markeAufraeumen(ctx) }()

	zurueck, tauschfehler := bestandBeiseite(b.orte.Datenbank)
	if tauschfehler != nil {
		return "", errors.Join(tauschfehler, b.wiederOeffnen(ctx, zurueckgeben))
	}

	bilderZurueck, bilderfehler := b.beiwerkErsetzen("")
	if bilderfehler != nil {
		return "", b.zurueckDrehen(ctx, zurueck, zurueckgeben, bilderfehler)
	}

	// **Wiederöffnen legt den frischen Bestand an** — der Start migriert, und
	// die Migration schreibt das Schema in eine Datei, die es noch nicht gibt.
	//
	// **Scheitert es, kommen erst die Bilder zurück und dann der Bestand** —
	// dieselbe Reihenfolge wie beim Zurückspielen (Befund 3 des Zweitblicks
	// über Phase A); hier stand sie bis zum Nachlauf umgekehrt, und der alte
	// Bestand ging ohne seine Bilder auf.
	if oeffnungsfehler := b.wiederOeffnen(ctx, zurueckgeben); oeffnungsfehler != nil {
		bilderfehler := rueckweg(bilderZurueck)

		return "", errors.Join(b.zurueckDrehen(ctx, zurueck, zurueckgeben, oeffnungsfehler), bilderfehler)
	}

	return filepath.Join(filepath.Dir(b.orte.Datenbank), Rueckspielsicherung), nil
}
