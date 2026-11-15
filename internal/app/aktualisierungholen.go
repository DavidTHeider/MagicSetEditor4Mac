package app

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync/atomic"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/aktualisierung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Stufe 2: holen** —
// #1100.
//
// # Die Stufe ist die Grenze, und sie wird hier gezogen
//
// `melden` fragt und schweigt danach; `holen` lädt. Wer die Grenze im Fenster
// zöge, hätte sie in der Schicht, die keine fachlichen Entscheidungen trifft
// (§2) — und eine zweite Fläche käme ohne sie aus.

// EreignisLadefortschritt meldet dem Fenster, wie weit ein Abbild ist.
//
// **Ein Ereignis und keine Rückgabe**, weil ein Ladevorgang von 188 MB länger
// dauert als ein Aufruf. Es trägt zwei Zahlen: gelesen und ganz — und `ganz`
// ist −1, wenn die Gegenstelle keine Länge nennt.
const EreignisLadefortschritt = "aktualisierung:fortschritt"

// LadeAntwort ist, was das Fenster über einen Ladevorgang erfährt.
type LadeAntwort struct {
	// Neuer sagt, ob überhaupt etwas zu holen war.
	Neuer bool `json:"neuer"`

	// Fassung ist die geholte Fassung — leer, wenn nichts geholt wurde.
	Fassung string `json:"fassung"`

	// Pfad ist der Ort der geprüften Datei.
	//
	// **Er steht erst da, wenn die Prüfsumme stimmt.** Eine Antwort mit Pfad
	// ist die Zusage, dass diese Datei geöffnet werden darf.
	Pfad string `json:"pfad"`

	Fehler *Fehler `json:"fehler"`
}

// NeueFassungLaden holt die neueste Freigabe — wenn die Stufe es erlaubt.
func (a *App) NeueFassungLaden(ctx context.Context) LadeAntwort {
	ctx = a.imProfil(ctx)

	freigabe, antwort, weiter := a.zuHolen(ctx)
	if !weiter {
		return antwort
	}

	// **Auch das Fenster geht durch den Riegel** (§7, Regel 4) — „ein Vorgang
	// je Prozess" gilt für alle drei Flächen. Der Wächter im Haken ist **je
	// Hakeninstanz**: zwei Bauteile oder zwei Aufrufe im selben Fenster kämen
	// aneinander vorbei.
	if !laeuftGerade.CompareAndSwap(false, true) {
		return LadeAntwort{
			Neuer:   true,
			Fassung: freigabe.Fassung,
			Fehler: Uebersetze(ctx, fmt.Errorf(
				"%w: es läuft schon ein Ladevorgang — es wird nicht zweimal geholt",
				domain.ErrConflict)),
		}
	}

	defer laeuftGerade.Store(false)

	beenden := a.auftragBeginnen("fassung", freigabe.Fassung)
	antwort = a.fassungHolen(ctx, freigabe)
	beenden(fehlerAus(antwort.Fehler))

	return antwort
}

// fehlerAus macht aus einem Umschlagfehler wieder einen Fehler — fürs Buch.
//
// **Ein Abbruch bleibt einer** (FMT-03) — er wird im Buch kein Scheitern.
func fehlerAus(fehler *Fehler) error {
	switch {
	case fehler == nil:
		return nil
	case fehler.Code == CodeAbgebrochen:
		return context.Canceled
	}

	return errors.New(fehler.Meldung)
}

// fassungHolen ist der Weg nach der Eintragung — herausgelöst, damit das
// Buch Anfang und Ende sieht (#1422).
func (a *App) fassungHolen(ctx context.Context, freigabe aktualisierung.Freigabe) LadeAntwort {
	ort, err := a.downloadort()
	if err != nil {
		return LadeAntwort{Fehler: Uebersetze(ctx, domain.MitGrund(
			domain.ErrNotFound, err,
			"das Downloadverzeichnis ist nicht zu finden — richte in den "+
				"Einstellungen des Systems einen Ordner für Downloads ein"))}
	}

	pfad, err := a.fassungen.Laden(ctx, freigabe, ort,
		func(gelesen, ganz int64) {
			a.melden(a.fenster, EreignisLadefortschritt, gelesen, ganz)
		})
	if err != nil {
		return LadeAntwort{Neuer: true, Fassung: freigabe.Fassung, Fehler: Uebersetze(ctx, err)}
	}

	return LadeAntwort{Neuer: true, Fassung: freigabe.Fassung, Pfad: pfad}
}

// zuHolen beantwortet, ob geladen werden darf und was.
//
// **Getrennt, weil hier fünf Gründe zusammenkommen, nichts zu tun** — und
// keiner davon ist „es ging schief": abgeschaltet, nur melden, kein Zugang,
// kein Bestand, nichts Neues.
func (a *App) zuHolen(ctx context.Context) (aktualisierung.Freigabe, LadeAntwort, bool) {
	werte, err := a.einstellungen.Alle(ctx)
	if err != nil {
		return aktualisierung.Freigabe{}, LadeAntwort{Fehler: Uebersetze(ctx, err)}, false
	}

	// **`melden` ist keine halbe Erlaubnis.** Wer die Stufe 1 gewählt hat, hat
	// gesagt, dass er Bescheid wissen will — nicht, dass 188 MB kommen dürfen.
	if werte[domain.EinstellungAktualisierung] != domain.AktualisierungHolen {
		return aktualisierung.Freigabe{}, LadeAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: die Aktualisierung steht nicht auf %q",
			domain.ErrValidation, domain.AktualisierungHolen))}, false
	}

	if a.fassungen == nil {
		return aktualisierung.Freigabe{}, LadeAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: der Zugang zu den Freigaben ist nicht eingerichtet",
			domain.ErrValidation))}, false
	}

	freigabe, err := a.fassungen.NeuesteFreigabe(ctx)
	if err != nil {
		return aktualisierung.Freigabe{}, LadeAntwort{Fehler: Uebersetze(ctx, err)}, false
	}

	neuer, err := domain.FassungIstNeuer(freigabe.Fassung, Fassung)
	if err != nil {
		return aktualisierung.Freigabe{}, LadeAntwort{Fehler: Uebersetze(ctx, err)}, false
	}

	// **Die laufende Fassung wird nicht geholt** — das wäre die teuerste Art,
	// nichts zu tun.
	if !neuer {
		return aktualisierung.Freigabe{}, LadeAntwort{Fassung: freigabe.Fassung}, false
	}

	return freigabe, LadeAntwort{}, true
}

// MitDownloadort setzt den Ort, an den geladen wird.
//
// **Als Naht und aus demselben Grund wie `bilderort`**: er entsteht aus
// `os.UserHomeDir`, und ein Test hat kein Benutzerverzeichnis — der Zweig, der
// sein Fehlen meldet, wäre sonst eine Behauptung (§3).
func (a *App) MitDownloadort(ort func() (string, error)) { a.downloadort = ort }

// downloadortEcht gibt das Standard-Downloadverzeichnis dieses Rechners.
//
// **Der Name steht so in der Vorgabe und wird nicht übersetzt**: macOS legt
// `~/Downloads` an und zeigt ihn je nach Sprache anders an — der Pfad selbst
// bleibt englisch.
func downloadortEcht() (string, error) {
	zuhause, err := heimort()
	if err != nil {
		return "", fmt.Errorf("das Heimatverzeichnis finden: %w", err)
	}

	return filepath.Join(zuhause, "Downloads"), nil
}

// **Dieselbe Fähigkeit ohne Fenster** (§16) —
// #1100,
// Werkstück 7.
//
// Kommandozeile und Dienst brauchen dasselbe wie das Fenster, nur ohne
// Einstellung und ohne Ereignis. **Und sie dürfen `internal/aktualisierung`
// nicht kennen** — `internal/schichten` lässt `cmd` nur an `app`, `domain`,
// `rest` und `storage`. Ohne diese beiden Funktionen stünde die Fähigkeit
// dreimal da, jedes Mal mit einem eigenen Vergleich der Fassungsnummern.

// Fassungsauskunft ist, was eine Fläche ohne Fenster über die Fassungen
// erfährt.
type Fassungsauskunft struct {
	Laufend     string
	Freigegeben string
	Seite       string
	Neuer       bool
}

// NeuesteFassung fragt nach der neuesten Freigabe.
//
// **Ohne Rücksicht auf die Einstellung**, und das ist keine Auslassung: die
// Vorgabe „aus" verhindert, dass die Anwendung **von selbst** ins Netz greift
// (§16, Regel 6). Wer diese Funktion ruft, hat gefragt.
func NeuesteFassung(ctx context.Context) (Fassungsauskunft, error) {
	freigabe, err := Fassungsnaht().NeuesteFreigabe(ctx)
	if err != nil {
		return Fassungsauskunft{}, err
	}

	neuer, err := domain.FassungIstNeuer(freigabe.Fassung, Fassung)
	if err != nil {
		return Fassungsauskunft{}, err
	}

	return Fassungsauskunft{
		Laufend: Fassung, Freigegeben: freigabe.Fassung,
		Seite: freigabe.Seite, Neuer: neuer,
	}, nil
}

// laeuftGerade sagt, ob in diesem Prozess schon ein Ladevorgang läuft.
//
// **§7, Regel 4: ein Vorgang je Datensatz** — und der Datensatz ist hier die
// Anwendung selbst. Das Fenster hält seinen eigenen Wächter im Haken; der
// Dienst hatte keinen, und jeder `POST /aktualisierung/laden` fing einen
// weiteren Ladevorgang an. Angestrichen vom Gesamtreview vor 0.2.0.
//
// **Ein Riegel und keine Warteschlange**: wer zweimal drückt, meint einmal.
var laeuftGerade atomic.Bool

// FassungLaden holt das Abbild der neuesten Freigabe und gibt seinen Ort.
//
// **`melden` darf `nil` sein** — eine HTTP-Antwort kommt am Ende, und wer
// keinen Fortschritt zeigen kann, soll keinen erfinden müssen.
func FassungLaden(ctx context.Context, melden func(gelesen, ganz int64)) (string, error) {
	if !laeuftGerade.CompareAndSwap(false, true) {
		return "", fmt.Errorf(
			"%w: es läuft schon ein Ladevorgang — es wird nicht zweimal geholt",
			domain.ErrConflict)
	}

	defer laeuftGerade.Store(false)

	naht := Fassungsnaht()

	freigabe, err := naht.NeuesteFreigabe(ctx)
	if err != nil {
		return "", err
	}

	// **Die laufende Fassung wird nicht geholt** — und die Prüfung steht
	// **hier**, an der gemeinsamen Stelle, nicht an jeder Fläche einzeln.
	//
	// Stünde sie an den Flächen, ginge ein Weg an ihr vorbei — ein
	// `POST /aktualisierung/laden` lüde dann dieselbe oder eine ältere
	// Fassung. Genau die Abweichung, die §16 verbietet.
	neuer, err := domain.FassungIstNeuer(freigabe.Fassung, Fassung)
	if err != nil {
		return "", err
	}

	if !neuer {
		return "", fmt.Errorf(
			"%w: %s ist nicht neuer als die laufende Fassung %s — es gibt nichts zu holen",
			domain.ErrValidation, freigabe.Fassung, Fassung)
	}

	ort, ortfehler := Downloadverzeichnis()
	if ortfehler != nil {
		return "", ortfehler
	}

	return naht.Laden(ctx, freigabe, ort, melden)
}
