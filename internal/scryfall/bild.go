package scryfall

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Bildart sagt, welches der angebotenen Bilder gemeint ist.
//
// **Drei von sieben.** Scryfall liefert `small`, `normal`, `large`, `png`,
// `art_crop`, `border_crop`; hier stehen die, für die es einen Verwender gibt.
type Bildart string

const (
	// BildKlein ist die Vorschau — für eine Trefferliste.
	BildKlein Bildart = "klein"
	// BildNormal ist die **ganze Karte** mit Rahmen, Copyright und Künstlername.
	BildNormal Bildart = "normal"
	// BildAusschnitt ist `art_crop` — die Illustration ohne Rahmen.
	//
	// **Sie hat eine Bedingung**, und die ist der Grund für [Bildfund.Nennung].
	// Nachgelesen am 2026-08-23 unter `https://scryfall.com/docs/api`:
	//
	//	When using the art_crop, list the artist name and copyright elsewhere
	//	in the same interface presenting the art crop, or use the full card
	//	image elsewhere in the same interface. Users should be able to identify
	//	the artist and source of the image somehow.
	BildAusschnitt Bildart = "ausschnitt"
)

// Copyright ist der Vermerk, der zu jedem Kartenbild gehört.
//
// **Er steht hier als Konstante und nicht in einer Oberfläche**, damit alle
// vier Flächen denselben zeigen — und damit ein Wächter ihn finden kann.
const Copyright = "™ & © Wizards of the Coast"

// Bildfund ist ein geholtes Bild samt dem, was daneben stehen muss.
type Bildfund struct {
	// Daten sind die Bildbytes, unverändert.
	//
	// **Unverändert ist eine Bedingung**: „Do not distort, skew, or stretch
	// card images. Do not blur, sharpen, desaturate, or color-shift card
	// images. Do not add your own watermarks, stamps, or logos."
	Daten []byte

	Art       Bildart
	Kuenstler string

	// Nennung ist der Satz, der in derselben Oberfläche stehen muss.
	//
	// **Er kommt mit dem Bild und wird nicht daneben gebaut.** Wer ein Bild
	// hat, hat damit auch seine Nennung — und kann nicht vergessen, sie zu
	// holen.
	Nennung string

	// Adresse ist die Seite bei Scryfall — die „source" aus der Bedingung.
	Adresse string
}

// bildgrenze ist die Grenze für eine Bilddatei.
//
// **16 MiB.** Ein `png` von Scryfall wiegt rund ein Megabyte, ein `art_crop`
// weniger. Die Grenze ist großzügig und deckelt trotzdem, was ein Fehler oder
// eine fremde Adresse anrichten könnte.
const bildgrenze = 16 << 20

// Bild holt ein Bild zu einer Karte.
//
// # Die Wache, und warum sie eine ist
//
// **Ein `art_crop` ohne Künstlernamen wird abgewiesen.** Die Bedingung
// verlangt, dass Künstler und Copyright in derselben Oberfläche stehen; ohne
// den Namen lässt sie sich nicht einhalten, und ein Bild auszuliefern, dessen
// Bedingung man nicht erfüllen kann, ist keine Auslegungsfrage.
//
// **Die ganze Karte hat sie nicht** — dort stehen Künstler und Copyright im
// Bild selbst, und die Bedingung lautet nur: nicht abschneiden.
//
// # Warum keine Drossel
//
// Die Bilddateien liegen auf `*.scryfall.io`, und dazu steht ausdrücklich:
// „The direct file origins located at *.scryfall.io" haben keine der oben
// genannten Grenzen. Der `User-Agent` geht trotzdem mit.
func (k *Klient) Bild(ctx context.Context, karte *Karte, art Bildart) (Bildfund, error) {
	adresse, fehler := bildadresse(karte, art)
	if fehler != nil {
		return Bildfund{}, fehler
	}

	if art == BildAusschnitt && karte.Kuenstler == "" {
		return Bildfund{}, fmt.Errorf(
			"%w: für den Bildausschnitt fehlt der Künstlername — ohne ihn lässt sich "+
				"die Bedingung von Scryfall nicht einhalten (Künstler und Copyright "+
				"in derselben Oberfläche)", domain.ErrValidation)
	}

	antwort, holfehler := netz.Hole(ctx, k.klient, nil, k.kenner, adresse, bildgrenze)
	if holfehler != nil {
		return Bildfund{}, holfehler
	}

	if antwort.Status != netz.StatusGut {
		return Bildfund{}, fmt.Errorf("%w: das Bild antwortet mit %d",
			domain.ErrValidation, antwort.Status)
	}

	return Bildfund{
		Daten:     antwort.Rumpf,
		Art:       art,
		Kuenstler: karte.Kuenstler,
		Nennung:   NennungFuer(karte.Kuenstler),
		Adresse:   karte.Adresse,
	}, nil
}

// bildadresse sucht die Adresse zur gewünschten Art.
//
// **Auch auf der ersten Seite.** Eine zweiseitige Karte hat kein `image_uris`
// auf der obersten Ebene; ihre Bilder hängen an den Seiten. Wer nur oben
// nachsähe, fände für jede zweiseitige Karte nichts.
func bildadresse(karte *Karte, art Bildart) (string, error) {
	if karte == nil {
		return "", fmt.Errorf("%w: ohne Karte gibt es kein Bild", domain.ErrValidation)
	}

	if adresse := ausBildern(karte.Bilder, art); adresse != "" {
		return adresse, nil
	}

	if len(karte.Seiten) > 0 {
		if adresse := ausBildern(karte.Seiten[0].Bilder, art); adresse != "" {
			return adresse, nil
		}
	}

	return "", fmt.Errorf("%w: Scryfall bietet für %q kein Bild der Art %q",
		domain.ErrNotFound, karte.Name, art)
}

// ausBildern gibt die Adresse einer Art.
func ausBildern(bilder Bilder, art Bildart) string {
	switch art {
	case BildKlein:
		return bilder.Klein
	case BildNormal:
		return bilder.Normal
	case BildAusschnitt:
		return bilder.Ausschnitt
	default:
		return ""
	}
}

// NennungFuer baut den Satz, der neben einem übernommenen Wert stehen muss.
//
// **Exportiert, weil ihn zwei brauchen** (#1025):
// der Bildabruf stellt ihn dem Bild bei, und die Karte braucht ihn später
// wieder — dann liegt kein `Bildfund` mehr vor, sondern nur die Herkunft und
// der Künstlername aus dem Kartenfeld. Zwei Fassungen liefen auseinander,
// sobald jemand eine davon nachzöge.
//
// **Ohne Künstlernamen bleibt das Copyright.** Ein Wert von einer Kartenseite
// hat einen Rechteinhaber, auch wenn die Karte keinen Illustrator führt; der
// Satz sagt dann, was bekannt ist, statt gar nichts.
func NennungFuer(kuenstler string) string {
	if kuenstler == "" {
		return Copyright
	}

	return "Illustration: " + kuenstler + ". " + Copyright
}

// IstKartenseite sagt, ob eine Adresse auf eine Kartenseite bei Scryfall zeigt.
//
// # Warum das eine Frage an dieses Paket ist
//
// **Weil hier steht, was in eine Herkunft geschrieben wird.** Der Feldabgleich
// und der Bildweg legen `Karte.Adresse` ab — das `scryfall_uri` der Antwort,
// etwa `https://scryfall.com/card/lea/232/black-lotus?utm_source=api`. Wer die
// Prüfung anderswo nachbaute, prüfte eine Gestalt, die er nicht setzt.
//
// **Der Wirt wird verglichen, nicht gesucht.** Ein `strings.Contains` auf
// „scryfall.com" hielte `https://scryfall.com.example.org/…` für eine
// Kartenseite — und die Anwendung schriebe fremdes Copyright unter ein Bild,
// das von irgendwoher kam.
func IstKartenseite(adresse string) bool {
	zerlegt, fehler := url.Parse(adresse)
	if fehler != nil || zerlegt.Scheme != "https" {
		return false
	}

	if zerlegt.Host != "scryfall.com" && zerlegt.Host != "www.scryfall.com" {
		return false
	}

	return strings.HasPrefix(zerlegt.Path, "/card/")
}
