package domain

// **Was ein Gast bei seiner Rückkehr bekommt** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum die Antwort eine eigene Gestalt braucht
//
// **Weil der Gast danach an einer anderen Stelle steht.** [Wiederaufnahme]
// entscheidet, **wie** er zurückkommt; diese Gestalt ist, was der Wirt ihm
// dafür schickt — und bei einem Abzug springt sein Stand. Ein Tischbild allein
// sagte nicht, auf welche Folge; er spielte weiter, als hätte er nichts
// verpasst.
//
// # Und warum die Folge der Partie zählt und nicht die der Leitung
//
// **Weil eine Leitung bei jedem Abriss neu anfängt.** `Partienachricht.Folge`
// gehört der **Sitzung**: nach einem Neuverbinden steht sie wieder bei eins.
// Was den Spielstand beschreibt, ist [Ereignis.Folge] — sie gehört der Partie
// und überlebt jede Leitung.
//
// **Der erste Entwurf hat das verwechselt** und hätte einen Gast mit den
// Zahlen seiner frischen Sitzung nach seinem Spielstand fragen lassen.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Partieepoche ist die Epoche, unter der eine Partie wiederaufgenommen wird.
//
// # Warum sie eine Konstante ist und keine Zahl, die steigt
//
// **Weil eine Partie keine Epoche hat.** Die Epoche in
// [Wiederaufnahmewunsch] schützt gegen einen Strom, der neu angefangen hat —
// und bei einer Partie fängt keiner neu an: ihre Ereignisse liegen im Bestand
// und tragen ihre Folge für immer.
//
// **Was die Epoche sonst leistete, leistet hier der Folgenvergleich**: ein
// Wirt, der seinen Stand verloren hätte, stünde **hinter** dem Gast, und
// [Wiederaufnahme] weist genau das ab („das ist nicht dieselbe Partie").
//
// **Auf der Leitung gibt es die Epoche weiterhin**, und dort trägt sie: jede
// frische Sitzung bekommt eine, und [Empfaengerstand] wirft damit weg, was
// aus einer alten Verbindung nachkommt. Das sind zwei verschiedene Fragen,
// und sie tragen denselben Namen — deshalb steht dieser Absatz hier.
const Partieepoche = 1

// Rueckkehr ist die Antwort des Wirts auf einen [Wiederaufnahmewunsch].
type Rueckkehr struct {
	Art Wiederaufnahmeart `json:"art"`

	// Grund sagt, warum — **auch beim Weiterspielen**.
	Grund string `json:"grund"`

	// AbFolge ist die erste nachgereichte Folge — nur bei [WegEreignisse].
	AbFolge int64 `json:"abFolge"`

	// Stand ist die Folge, auf der der Gast danach steht — bei [WegAbzug] die
	// des Wirts, sonst seine eigene.
	Stand int64 `json:"stand"`

	// Abdruck ist der sichtbezogene Nachweis zu [Rueckkehr.Stand].
	//
	// **Ohne ihn käme der Gast genau einmal zurück.** Nach einem Abzug hat er
	// die Ereignisse nicht, aus denen er ihn selbst rechnen könnte; schickte
	// er beim nächsten Mal eine leere Kette, wiese der Wirt ihn ab. Gefunden
	// beim Entwerfen und nicht im Betrieb — dort hätte es wie ein
	// Netzfehler ausgesehen.
	Abdruck string `json:"abdruck"`

	// Bild ist der Abzug — nur bei [WegAbzug] gefüllt.
	//
	// **Und es ist dasselbe Tischbild wie nach jedem Zug.** Ein zweiter
	// Abzugstyp wäre eine zweite Wahrheit über denselben Tisch; was ein Sitz
	// sehen darf, rechnet die Projektion aus G1 — einmal.
	Bild Tischbild `json:"bild"`
}

// Sichtabdruck rechnet den **sichtbezogenen** Nachweis eines Ereignisses.
//
// # Warum die Kette allein nicht reicht
//
// **Weil sie für alle dieselbe ist.** [Ereignis.Kette] schließt die privaten
// Ergänzungen ausdrücklich aus — zwei Gäste bei derselben Folge haben
// dieselbe Kette und **verschiedene** Karten auf der Hand. Ein Gast, der mit
// der Kette allein zurückkäme, könnte den privaten Strom eines anderen
// bekommen, ohne dass es auffiele.
//
// **Das Issue verlangt genau das**: „Resume mit Session-ID, Epoch, Sequence
// und **sichtbezogenem letzten Hash**."
//
// # Was eingeht und was nicht
//
// **Die Kette und die Ergänzung für diesen Sitz.** Ergänzungen an andere
// bleiben draußen — sonst wüsste ein Gast an seinem eigenen Abdruck, dass
// jemand anderes etwas bekommen hat.
func (e Ereignis) Sichtabdruck(fuer string) (string, error) {
	// **Über die Fassung, die der Gast bekommt** (#2480): er rechnet über
	// das, was er empfing, und das trägt Sichtkennungen.
	e = e.Fuer(fuer)

	meine := ""

	// **Nur noch die eigenen** — [Ereignis.Fuer] hat die übrigen genommen.
	for _, ergaenzung := range e.Privat {
		abdruck, fehler := ergaenzung.Abdruck()
		if fehler != nil {
			return "", fmt.Errorf(
				"der Sichtabdruck für %q ließ sich nicht rechnen: %w",
				fuer, fehler)
		}

		meine += abdruck
	}

	summe := sha256.Sum256([]byte(e.Kette + "\x00" + meine))

	return hex.EncodeToString(summe[:]), nil
}
