package scryfall

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Grenzen und Grundadresse, nachgelesen am 2026-08-23.
const (
	// Adresse ist die Grundadresse des Dienstes.
	Adresse = "https://api.scryfall.com"

	// kartenabstand ist der Mindestabstand für die Kartenwege.
	//
	// **600 ms statt der erlaubten 500.** §16, Regel 4: gedrosselt wird
	// **unter** der genannten Zahl. `/cards/search`, `/cards/named`,
	// `/cards/random` und `/cards/collection` dürfen 2/s — wer genau 500 ms
	// wartet, liegt bei jeder Uhrungenauigkeit darüber.
	kartenabstand = 600 * time.Millisecond

	// allgemeinabstand ist der Mindestabstand für alles Übrige.
	//
	// **125 ms sind 8/s, erlaubt sind 10.** Dieselbe Regel wie darüber, und
	// dieselbe Reserve — §16, Regel 4: gedrosselt wird **unter** der genannten
	// Zahl.
	//
	// **Warum nicht dieselbe Drossel wie die Kartenwege**: die Grenzen sind
	// verschieden. Eine gemeinsame müsste die strengere nehmen und bremste die
	// Rulings, Setangaben und die Symbologie auf ein Fünftel dessen, was
	// erlaubt ist — ohne dass jemand etwas davon hätte.
	allgemeinabstand = 125 * time.Millisecond

	// hoechstens ist die Grenze für eine Antwort.
	//
	// **Ein Mebibyte.** Eine Kartenantwort wiegt wenige Kilobyte; eine Suche
	// mit 175 Treffern bleibt weit darunter. Die Bulk-Dateien sind ein eigener
	// Weg — sie wiegen 74 MB und gehen nicht über diesen Klienten.
	hoechstens = 1 << 20

	// Speicherfrist ist, wie lange ein Stand als frisch gilt.
	//
	// **24 Stunden**, das Mindestmaß aus den Bedingungen: „We encourage you to
	// cache the data you download from Scryfall […] at least for 24 hours."
	//
	// **Exportiert, weil der Zwischenspeicher eingesetzt wird** (§16). Wer ihn
	// baut, soll die Zahl nicht erfinden müssen — und wer sie hier ändert,
	// ändert sie für alle, die sie benutzen.
	Speicherfrist = 24 * time.Hour
)

// Klient fragt Scryfall.
//
// # Die drei Schichten, und wo sie liegen
//
// **Transport** in `internal/netz` — Fristen, Kopfzeilen, Drossel, 429.
// **Fremdes DTO** in `karte.go`. **Zuordnung** in `zuordnung.go`. Hier steht,
// was sie verbindet: welche Adresse, welcher Abstand, welcher Schlüssel im
// Speicher.
type Klient struct {
	grundadresse string
	klient       *netz.Klient
	drossel      *netz.Drossel
	allgemein    *netz.Drossel
	speicher     *netz.Speicher
	kenner       netz.Kenner
}

// NeuerKlient baut einen Klienten.
//
// **Klient, Uhr und Zwischenspeicher werden eingesetzt** (§16). Sonst brauchte
// jeder Test das Netz — und was das wert ist, sagt §3.
func NeuerKlient(
	grundadresse string, klient *netz.Klient, speicher *netz.Speicher,
	jetzt func() time.Time, fassung string,
) *Klient {
	return &Klient{
		grundadresse: grundadresse,
		klient:       klient,
		drossel:      netz.NeueDrossel(kartenabstand, jetzt, nil),
		allgemein:    netz.NeueDrossel(allgemeinabstand, jetzt, nil),
		speicher:     speicher,
		kenner: netz.Kenner{
			// **Der `User-Agent` nennt diese Anwendung.** Pflicht, und die
			// Bedingungen sagen auch, warum: „Your User-Agent header must be
			// accurate to your usage context."
			Name: "MagicSetEditor4Mac/" + fassung +
				" (https://github.com/DavidTHeider/MagicSetEditor4Mac)",
			Annahme: "application/json;q=0.9,*/*;q=0.8",
		},
	}
}

// KarteNachName holt eine Karte über `/cards/named`.
//
// # Was bei einem Ausfall geschieht
//
// **Der letzte gute Stand kommt** — mit seinem Alter, ohne Fehler. Das ist
// §16, Regel 5: „Ohne Netz arbeitet die Anwendung weiter." Erst wenn auch
// kein Stand da ist, scheitert der Aufruf.
//
// **Und ein fachliches Nein ist kein Ausfall.** Eine Karte, die es nicht gibt,
// gibt 404 — daraus soll kein alter Stand einer anderen Karte werden. Deshalb
// wird die Statusprüfung **vor** dem Speicher gemacht.
func (k *Klient) KarteNachName(
	ctx context.Context, name string,
) (karte *Karte, alter time.Duration, befunde []domain.Befund, fehler error) {
	adresse := k.grundadresse + "/cards/named?exact=" + url.QueryEscape(name)

	return k.karte(ctx, "cards/named:"+name, adresse)
}

// karte holt und liest eine Kartenantwort.
func (k *Klient) karte(
	ctx context.Context, schluessel, adresse string,
) (karte *Karte, alter time.Duration, befunde []domain.Befund, fehler error) {
	roh, alter, fehler := k.speicher.Hole(ctx, schluessel, func() ([]byte, error) {
		return k.hole(ctx, adresse, "diese Karte")
	})
	if fehler != nil {
		return nil, 0, nil, fehler
	}

	gelesen, befunde, lesefehler := LiesKarte(roh)
	if lesefehler != nil {
		return nil, 0, nil, lesefehler
	}

	return gelesen, alter, befunde, nil
}

// hole stellt die Anfrage und deutet den Status.
//
// # Warum die Statusprüfung hier steht und nicht im Transport
//
// Weil erst dieses Paket weiß, was ein Status **bedeutet**. Ein 404 von
// Scryfall heißt „die Karte gibt es nicht" — ein fachliches Nein, das an den
// Aufrufer gehört und nicht durch einen alten Stand ersetzt werden darf.
// `internal/netz` kennt keine Karten.
func (k *Klient) hole(
	ctx context.Context, adresse, gegenstand string,
) ([]byte, error) {
	return k.holeMit(ctx, k.drossel, adresse, gegenstand)
}

// holeMit ist dasselbe mit gewählter Drossel.
//
// **Die Drossel steht im Aufruf und nicht im Klienten**, weil sie zur
// **Endpunktklasse** gehört und nicht zur Verbindung: derselbe Klient fragt
// Kartenwege mit 2/s und alles Übrige mit 10/s.
// **`gegenstand` sagt, wonach gefragt wurde** —
// #1054.
//
// Dieselbe Funktion bedient Karten, den Satz, die Satzangaben, das
// Verzeichnis der Sammeldateien und die Symbologie — ein fester Satz über
// eine Karte schickte jemanden auf die Suche nach einer, die niemand gesucht
// hat.
//
// **Der Aufrufer nennt ihn**, denn nur er weiß es. Diese Funktion kennt die
// Adresse und sonst nichts.
func (k *Klient) holeMit(
	ctx context.Context, drossel *netz.Drossel, adresse, gegenstand string,
) ([]byte, error) {
	antwort, fehler := netz.Hole(ctx, k.klient, drossel, k.kenner, adresse, hoechstens)
	if fehler != nil {
		return nil, fehler
	}

	if antwort.Status == netz.StatusNichtGefunden {
		return nil, fmt.Errorf("%w: Scryfall kennt %s nicht",
			domain.ErrNotFound, gegenstand)
	}

	if antwort.Status != netz.StatusGut {
		return nil, fmt.Errorf("%w: Scryfall antwortet mit %d",
			domain.ErrValidation, antwort.Status)
	}

	return antwort.Rumpf, nil
}

// Treffer ist ein Ergebnis einer Suche.
type Treffer struct {
	Karten []*Karte
	// Weitere sagt, ob Scryfall mehr hat, als hier steht.
	//
	// **Es wird nicht nachgeladen.** Eine Suche, die von selbst durch alle
	// Seiten liefe, stellte ohne Zutun beliebig viele Anfragen — und die
	// Bedingungen sagen, dass für viele Karten die Bulk-Datei zu nehmen ist.
	Weitere bool
}

// Suche holt Karten über `/cards/search` — **eine Zeile je Karte**.
//
// **Das ist Scryfalls Vorgabe `unique=cards`**, und für eine Kartensuche ist
// sie richtig: wer „bolt" tippt, will die Karten sehen und nicht 72 Ausgaben
// derselben.
func (k *Klient) Suche(
	ctx context.Context, frage string,
) (treffer Treffer, alter time.Duration, befunde []domain.Befund, fehler error) {
	return k.suchen(ctx, frage, "")
}

// Druckesuche holt Karten über `/cards/search` — **eine Zeile je Druck**.
//
// # Warum es sie neben [Klient.Suche] gibt
//
// Weil die Listeneinfuhr die andere Frage stellt: *welche Ausgaben dieser
// Karte gibt es?* Ohne `unique=prints` gibt Scryfall **eine** — und damit sähe
// jede Zeile einer eingefügten Liste eindeutig aus.
//
// **Nachgemessen am 2026-08-30** gegen `api.scryfall.com`, mit dem Kopf aus
// §16 Regel 2:
//
//	!"Counterspell"                     → total_cards  1
//	!"Counterspell" &unique=prints      → total_cards 72
//
// **Gefunden hat es die laufende Brücke**, nicht ein Test: das Doppel in
// `service`, die Gegenstelle in `internal/flaechen` und die Vorrichtung im
// Frontend gaben alle mehrere Treffer, weil sie so geschrieben waren. Erst der
// Ruf durch `window.go.app.Bindungen.ListeDeuten` gegen den echten Dienst gab
// für „Counterspell" **eindeutig: true** zurück
// (#1292).
//
// **`order=released&dir=asc`**, damit die Auswahl eine Reihenfolge hat, die
// ein Mensch versteht: die älteste Ausgabe zuerst. Ohne Angabe ordnet Scryfall
// nach Relevanz, und die ist bei gleichem Namen keine.
func (k *Klient) Druckesuche(
	ctx context.Context, frage string,
) (treffer Treffer, alter time.Duration, befunde []domain.Befund, fehler error) {
	return k.suchen(ctx, frage, "&unique=prints&order=released&dir=asc")
}

// suchen ist beides — die Zusätze entscheiden, was gezählt wird.
func (k *Klient) suchen(
	ctx context.Context, frage, zusatz string,
) (treffer Treffer, alter time.Duration, befunde []domain.Befund, fehler error) {
	adresse := k.grundadresse + "/cards/search?q=" + url.QueryEscape(frage) + zusatz

	// **Der Zusatz gehört in den Schlüssel des Zwischenspeichers.** Ohne ihn
	// lägen beide Suchen unter demselben Namen, und die zweite bekäme die
	// Antwort der ersten — mit einer Zeile statt zweiundsiebzig.
	roh, alter, fehler := k.speicher.Hole(ctx, "cards/search:"+frage+zusatz, func() ([]byte, error) {
		// **Eine Suche ist keine Karte** (#1054): wer nichts findet, hat nichts
		// gefunden — und nicht „diese Karte" nicht.
		return k.hole(ctx, adresse, "eine Karte zu dieser Suche")
	})
	if fehler != nil {
		return Treffer{}, 0, nil, fehler
	}

	var liste struct {
		Objekt  string            `json:"object"`
		Weitere bool              `json:"has_more"`
		Daten   []json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(roh, &liste); err != nil {
		return Treffer{}, 0, nil, domain.MitGrund(domain.ErrValidation, err,
			"die Suchantwort von Scryfall ist kein gültiges JSON")
	}

	if liste.Objekt != "list" {
		return Treffer{}, 0, nil, fmt.Errorf(
			"%w: die Suchantwort von Scryfall hat kein Feld %q — das ist ein Pflichtfeld",
			domain.ErrValidation, "object")
	}

	karten := make([]*Karte, 0, len(liste.Daten))

	for _, eine := range liste.Daten {
		gelesen, weitere, lesefehler := LiesKarte(eine)
		if lesefehler != nil {
			return Treffer{}, 0, nil, lesefehler
		}

		karten = append(karten, gelesen)
		befunde = append(befunde, weitere...)
	}

	// **Was eine frische Suche gezeigt hat, ist Vorrat** — erst, wenn jede
	// Karte darin lesbar war, und nur, was gerade geholt wurde.
	if alter == 0 {
		k.vorratAblegen(liste.Daten, karten)
	}

	return Treffer{Karten: karten, Weitere: liste.Weitere}, alter, befunde, nil
}

// Drosselschluessel sind die Namen, unter denen die beiden Klassen im
// Drosselwerk liegen.
//
// **Sie sind exportiert**, weil das Werk beim Composition Root liegt und die
// Schlüssel dort gebraucht werden — und weil zwei Schreibweisen desselben
// Schlüssels zwei Drosseln ergäben, die einander nicht kennen.
const (
	// DrosselKarten ist die Klasse der Kartenwege — 2/s bei Scryfall.
	DrosselKarten = "scryfall/karten"

	// DrosselAllgemein ist alles Übrige — 10/s.
	DrosselAllgemein = "scryfall/allgemein"
)

// MitDrosselwerk lässt diesen Klienten seine Drosseln aus einem Werk nehmen.
//
// # Warum das nötig ist
//
// **Das Budget gilt für Host und IP, nicht für eine Instanz.** Vor dem
// 2026-08-24 baute das Fenster vier Klienten, jeden mit eigener Drossel; MCP
// und REST bauten ihre Naht je Aufruf neu. Weil der erste Aufruf einer Drossel
// absichtlich nicht wartet, sah eine solche Drossel praktisch nie einen
// zweiten — die Grenze war wirkungslos. Angestrichen vom Zweitblick über
// Phase V.
//
// # Warum eine Naht und kein Pflichtargument
//
// Weil ein Klient ohne Werk richtig bleibt: er hält die Grenze für sich ein.
// Ein Pflichtargument zwänge jeden Test, ein Werk zu bauen, und die Grenze
// wäre dadurch nicht besser eingehalten.
//
// **Sie gibt den Klienten zurück**, damit sie sich an den Bau anhängen lässt.
func (k *Klient) MitDrosselwerk(werk *netz.Drosselwerk) *Klient {
	if werk == nil {
		return k
	}

	k.drossel = werk.Fuer(DrosselKarten, kartenabstand)
	k.allgemein = werk.Fuer(DrosselAllgemein, allgemeinabstand)

	return k
}
