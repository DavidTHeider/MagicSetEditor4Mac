package app

// **Ein Vorschlag darf sich auf Belege stützen** — Werkstück A13.4 von Phase
// 5.3 (#1452).
//
// # Der Satz, um den herum das gebaut ist
//
// **„A11 erhält dieselben Manifest-/Belegdaten wie A16; keine zweite
// Prompt-/Retrievalpipeline."**
//
// Bis hierher bekam ein Vorschlag den getippten Auftrag und das Schema —
// sonst nichts. Jetzt darf er denselben Wissensbestand sehen wie die
// Einschätzung, und zwar über **denselben** [service.Kontextdienst]: dasselbe
// Manifest, dieselbe Zählweise, dieselbe Freigabe.
//
// # Was das Werkstück trennt
//
// **Beobachtung und Vermutung.** Die Belege sind, was **dasteht** — mit
// Quelle, Revision und Abrufdatum. Die Feldwerte sind, was ein Modell daraus
// **macht**, und sie werden bei der Übernahme als `ki` verbucht (A11). Die
// Antwort führt beides nebeneinander und behauptet **nicht**, dass ein Wert
// aus einem Beleg folgt.
//
// # Und warum die Belegfrage wahlfrei ist
//
// **Weil der gewöhnliche Vorschlag keine braucht.** „Schreibe einen
// Regeltext" ist eine Aufgabe an das Modell und keine Frage an einen Index;
// wer trotzdem suchte, zahlte Zeit für nichts. Leer heißt: kein Ruf an den
// Kontextdienst.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// KIVorschlagsbezug ist, was für einen Vorschlag gebraucht wird.
//
// **Ein Verbund und keine sieben Stellungsparameter** — dieselbe Begründung
// wie bei [KIEinschaetzungsbezug]: mit der Belegfrage und der Freigabe wären
// es sieben, und wer zwei davon vertauscht, merkt es an keiner Stelle.
type KIVorschlagsbezug struct {
	KarteID int64

	Anbieter string
	Alias    string

	// Auftrag ist, was das Modell tun soll.
	Auftrag string

	// Nachweis ist die Freigabe eines Menschen — leer heißt: keine.
	Nachweis string

	// Belegfrage geht an den Suchindex — **leer heißt: gar nicht fragen**.
	Belegfrage string

	// Freigegeben nennt die **privaten** Korpora, die mitgesucht werden
	// dürfen. Leer heißt: keine.
	Freigegeben []string

	// Nur nennt die Felder, nach denen gefragt wird — **leer heißt: alle**
	// des Stils. Nur Name und Flavortext machen eine kleine Textaufgabe, und
	// die geht mit dem Vorrang zuerst an das Gerätemodell (AI-05, #2042).
	Nur []string
}

// KIVorschlagenMit fragt ein Modell — wahlweise mit Belegen.
func (a *App) KIVorschlagenMit(
	ctx context.Context, bezug KIVorschlagsbezug,
) KIVorschlagAntwort {
	ctx = a.imProfil(ctx)

	// **Die Rolle vor dem Modell** — dieselbe Frage wie bei
	// [App.KIVorschlagen] (ID-03, #1702).
	if verweigert := a.darf(ctx, domain.WirkungSchreiben); verweigert != nil {
		return KIVorschlagAntwort{Fehler: Uebersetze(ctx, verweigert)}
	}

	profil, fehler := a.profilDiesesRufs(ctx)
	if fehler != nil {
		return KIVorschlagAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	paket, fehler := a.vorschlagspaket(ctx, bezug)
	if fehler != nil {
		return KIVorschlagAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	frage := kifrage{
		profil: profil, karte: bezug.KarteID, anbieter: bezug.Anbieter,
		alias: bezug.Alias, auftrag: mitBelegen(bezug.Auftrag, paket), nur: bezug.Nur,
	}

	var antwort KIVorschlagAntwort

	// **Ohne Anbieter wählt die Regel** — AI-05.AK-05 (#2042): eine kleine
	// Textaufgabe zuerst an das Gerätemodell, dann lokal, nie die Cloud.
	if bezug.Anbieter == domain.KIVorrang {
		antwort = a.vorschlagMitVorrang(ctx, frage, bezug)
	} else {
		antwort = a.vorschlagGewaehlt(ctx, frage, bezug)
	}

	if antwort.Fehler != nil {
		return antwort
	}

	// **Die Belege stehen an der Antwort und nicht im Auftrag** — das ist
	// die Trennung, für die es das Werkstück gibt: was hineinging, ist
	// nachzulesen, ohne den Prompt zu rekonstruieren.
	antwort.Belege = alsBelegDTOs(paket.Manifest.Belege)
	antwort.Korpusstaende = alsKorpusstandDTOs(paket.Manifest.Korpusstaende)
	antwort.Einschraenkungen = paket.Ergebnis.Einschraenkungen

	return antwort
}

// vorschlagspaket stellt die Belege zusammen — über den Dienst aus A16.
//
// **Ohne Belegfrage gar nicht.** Das leere Paket ist kein Fehler: es ist der
// gewöhnliche Vorschlag, und der hat keine Belege, weil niemand welche
// verlangt hat.
func (a *App) vorschlagspaket(
	ctx context.Context, bezug KIVorschlagsbezug,
) (service.Kontextpaket, error) {
	if bezug.Belegfrage == "" {
		return service.Kontextpaket{}, nil
	}

	if a.recherche == nil {
		return service.Kontextpaket{}, fehlenderBestand()
	}

	return a.recherche.Zusammenstellen(ctx, service.Kontextwunsch{
		Aufgabe: bezug.Auftrag,
		Anfrage: domain.Rechercheanfrage{
			Fall: domain.FallKarten, Frage: bezug.Belegfrage,
			Grenze:      KIEinschaetzungstreffer,
			Freigegeben: freigegebeneQuellen(bezug.Freigegeben),
		},
		Zielanbieter: bezug.Anbieter, Zielmodell: bezug.Alias,

		// **Dieselben Zahlen wie bei der Einschätzung.** Zwei Budgets für
		// dieselbe Pipeline wären zwei Antworten auf dieselbe Frage —
		// und die zweite bliebe beim nächsten Modell stehen.
		Kontextgrenze:    KIEinschaetzungsfenster,
		Ausgabereserve:   KIEinschaetzungsreserve,
		HoechstensBelege: KIEinschaetzungsbelege,
	})
}

// mitBelegen stellt den Belegtext vor den Auftrag.
//
// **Davor und nicht dahinter**: ein Modell liest von oben, und der Auftrag
// ist das, was es zuletzt gelesen haben soll.
//
// **Und ohne Belege bleibt der Auftrag unverändert** — kein leerer Vorspann,
// der wie ein abgeschnittener Kontext aussieht.
func mitBelegen(auftrag string, paket service.Kontextpaket) string {
	if paket.Kontext == "" {
		return auftrag
	}

	return "Diese Stellen stehen dir zur Verfügung. Sie sind das, was " +
		"DASTEHT — was du daraus machst, ist dein Vorschlag:\n\n" +
		paket.Kontext + "\n\n" + auftrag
}
