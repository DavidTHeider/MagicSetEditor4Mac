package app

// **Das Kontextpaket am Fenster** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Zusammenstellen ist nicht senden
//
// **Das Paket entsteht und bleibt liegen.** Was es enthielte, steht in seiner
// Freigabe; ob es hinausgeht, entscheidet ein zweiter Schritt — und der
// gehört A6.
//
// **Der Kontext selbst geht nicht ans Fenster.** Was zurückkommt, sind
// Manifest und Freigabe: wie viele Belege, wie viel Umfang, welche
// Kategorien. Den Volltext braucht die Oberfläche nicht, und ein
// schutzbedürftiger Textschnappschuss gehört nicht in eine Antwort, die
// jemand weitergibt (Issue: „Kurzmanifest für A11 von schutzbedürftigem
// Textsnapshot trennen").

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// BelegDTO ist ein Ausschnitt, auf den sich eine Aussage stützen darf.
type BelegDTO struct {
	Kennung string `json:"kennung"`

	Quellkennung string `json:"quellkennung"`
	Quelle       string `json:"quelle"`

	Kennungssystem string `json:"kennungssystem"`
	Fachkennung    string `json:"fachkennung"`
	Semantik       string `json:"semantik"`

	Titel string `json:"titel"`

	// Ausschnitt ist **der Text, der wirklich mitgegeben wurde**.
	Ausschnitt string `json:"ausschnitt"`

	Inhaltshash string `json:"inhaltshash"`

	Sprache     string `json:"sprache"`
	Regelnummer string `json:"regelnummer"`
	Urheber     string `json:"urheber"`

	Bezugsadresse string `json:"bezugsadresse"`

	Quellrevision string `json:"quellrevision"`
	Quelldatum    string `json:"quelldatum"`
	Abgerufen     string `json:"abgerufen"`

	// Veraltet sagt, dass die Quelle nicht dem aktiven Stand entspricht.
	Veraltet bool `json:"veraltet"`
}

// KontextfreigabeDTO ist, was vor dem Senden gezeigt wird.
type KontextfreigabeDTO struct {
	Zielanbieter string `json:"zielanbieter"`
	Zielmodell   string `json:"zielmodell"`

	// Ort sagt, **wo** gerechnet wird.
	Ort string `json:"ort"`

	// Zieladresse sagt, **wohin** es ginge — leer heißt: die Vorgabe des
	// Anbieters. Seit A18 kann hinter einem Anbieternamen ein Gateway
	// stehen, und das ist ein Datenexport an einen Dritten (#1533).
	Zieladresse string `json:"zieladresse"`

	Kategorien []string `json:"kategorien"`

	Belege  int `json:"belege"`
	Zeichen int `json:"zeichen"`
	Tokens  int `json:"tokens"`

	// Privat sagt, ob eigener Inhalt dabei ist.
	Privat bool `json:"privat"`

	// Satz ist die Zusammenfassung — **an einer Stelle gebildet**, damit
	// vier Flächen nicht vier Sätze haben.
	Satz string `json:"satz"`
}

// KontextpaketAntwort ist das zusammengestellte Paket.
type KontextpaketAntwort struct {
	PaketID string `json:"paketID"`

	Aufgabe      string `json:"aufgabe"`
	Zielrevision string `json:"zielrevision"`

	// Zaehlweise nennt, wie Tokens geschätzt wurden — ohne sie wäre jede
	// Zahl hier bedeutungslos.
	Zaehlweise string `json:"zaehlweise"`

	Kontextgrenze  int `json:"kontextgrenze"`
	Ausgabereserve int `json:"ausgabereserve"`
	Verbraucht     int `json:"verbraucht"`

	Kontextpolicy string `json:"kontextpolicy"`

	Korpusstaende []KorpusstandDTO `json:"korpusstaende"`

	Belege []BelegDTO `json:"belege"`

	Freigabe KontextfreigabeDTO `json:"freigabe"`

	// Laeuftab ist, wann es nicht mehr gilt.
	Laeuftab string `json:"laeuftab"`

	// Ergebnis ist die Recherche, aus der die Belege stammen — **damit
	// sichtbar ist, was ausgewählt wurde und was nicht**.
	Ergebnis RechercheAntwort `json:"ergebnis"`

	Fehler *Fehler `json:"fehler"`
}

// KontextZusammenstellen baut ein Paket — und sendet es nicht.
func (a *App) KontextZusammenstellen(
	ctx context.Context, aufgabe, zielrevision, fall, frage string,
	grenze int, anbieter, modell string, fenster, reserve, hoechstens int,
	freigegeben []string,
) KontextpaketAntwort {
	ctx = a.imProfil(ctx)

	antwort := leeresKontextpaket()

	if a.recherche == nil {
		antwort.Fehler = Uebersetze(ctx, fehlenderBestand())

		return antwort
	}

	paket, fehler := a.recherche.Zusammenstellen(ctx, service.Kontextwunsch{
		Aufgabe: aufgabe, Zielrevision: zielrevision,
		Anfrage: domain.Rechercheanfrage{
			Fall: domain.Recherchefall(fall), Frage: frage, Grenze: grenze,

			// **Die Freigabe reist mit der Anfrage** (A13.4, #1452): das
			// Paket zeigt danach, was **wirklich** darin steht — und ohne
			// Freigabe steht der eigene Bestand eben nicht darin.
			Freigegeben: freigegebeneQuellen(freigegeben),
		},
		Zielanbieter: anbieter, Zielmodell: modell,
		Kontextgrenze: fenster, Ausgabereserve: reserve,
		HoechstensBelege: hoechstens,
	})
	if fehler != nil {
		antwort.Fehler = Uebersetze(ctx, fehler)

		return antwort
	}

	ort, adresse, zielfehler := a.zielFuer(ctx, anbieter)
	if zielfehler != nil {
		antwort.Fehler = Uebersetze(ctx, zielfehler)

		return antwort
	}

	return kontextpaketAntwort(paket, ort, adresse)
}

// zielFuer sagt, wo ein Anbieter rechnet und an welcher Adresse.
//
// **Einmal gelesen und nicht zweimal** — Ort und Adresse stammen aus
// derselben Einstellung, und ein zweiter Griff könnte zwischen beiden etwas
// anderes lesen. **Scheitert das Lesen, gibt es keine Freigabe**: wer nicht
// weiß, wohin es geht, sagt es nicht.
func (a *App) zielFuer(ctx context.Context, anbieter string) (domain.KIOrt, string, error) {
	adresse, fehler := a.dienstadresse(ctx, domain.Dienst(anbieter))
	if fehler != nil {
		return "", "", fehler
	}

	return a.ortFuer(ctx, anbieter, adresse), adresse, nil
}

// ortFuer sagt, wo ein Anbieter rechnet.
//
// # Warum das hier steht und nicht am Modell
//
// **Weil die Freigabe die Frage stellt, bevor ein Modell gefragt wird.** Ein
// Katalogruf ginge nach draußen — und genau das soll die Freigabe ja
// entscheiden.
//
// **Nur die lokale Laufzeit kann lokal sein**, und auch sie nur, wenn ihre
// Adresse auf der Schleife liegt: „Loopback allein beweist keine lokale
// Inferenz" gilt in die andere Richtung genauso — eine fremde Adresse beweist
// „nicht lokal" (0098).
func (a *App) ortFuer(ctx context.Context, anbieter, adresse string) domain.KIOrt {
	if anbieter == ollamaanbieter {
		return domain.KIOrtFuerAdresse(a.kiadresse(ctx))
	}

	// **Seit A18 entscheidet die eingestellte Adresse mit** (#1533): hinter
	// „openai" kann ein Firmengateway stehen, und eines auf der Schleife
	// rechnet auf diesem Rechner.
	return domain.OrtFuerDienstadresse(adresse)
}

// leeresKontextpaket ist die Antwort, wenn nichts entstand.
//
// **Mit leeren Listen und nicht mit Nullwerten**: ein `null` statt `[]` kommt
// im Fenster als „gibt es nicht" an (§3, die Naht zwischen den Prüfebenen).
func leeresKontextpaket() KontextpaketAntwort {
	return KontextpaketAntwort{
		Korpusstaende: []KorpusstandDTO{}, Belege: []BelegDTO{},
		Freigabe: KontextfreigabeDTO{Kategorien: []string{}},
		Ergebnis: RechercheAntwort{
			Treffer:       []RecherchetrefferDTO{},
			Korpusstaende: []KorpusstandDTO{}, Einschraenkungen: []string{},
		},
	}
}

// kontextpaketAntwort übersetzt das Paket.
func kontextpaketAntwort(
	paket service.Kontextpaket, ort domain.KIOrt, adresse string,
) KontextpaketAntwort {
	manifest := paket.Manifest

	belege := alsBelegDTOs(manifest.Belege)
	staende := alsKorpusstandDTOs(manifest.Korpusstaende)

	return KontextpaketAntwort{
		PaketID: manifest.PaketID, Aufgabe: manifest.Aufgabe,
		Zielrevision: manifest.Zielrevision, Zaehlweise: manifest.Zaehlweise,
		Kontextgrenze:  manifest.Kontextgrenze,
		Ausgabereserve: manifest.Ausgabereserve,
		Verbraucht:     manifest.Verbraucht,
		Kontextpolicy:  manifest.Kontextpolicy,
		Korpusstaende:  staende, Belege: belege,
		Freigabe: alsFreigabeDTO(service.FreigabeFuer(paket, ort, adresse)),
		Laeuftab: manifest.Laeuftab.UTC().Format(time.RFC3339),
		Ergebnis: rechercheAntwort(paket.Ergebnis),
	}
}

// alsBelegDTOs übersetzt die Belege eines Pakets.
//
// **Seit A13.2a zwei Aufrufer** (#1452): das Kontextpaket zeigt sie, und eine
// Einschätzung bringt sie mit. Zweimal dieselbe Schleife liefe beim nächsten
// Feld auseinander.
func alsBelegDTOs(belege []domain.Beleg) []BelegDTO {
	zeilen := make([]BelegDTO, 0, len(belege))
	for _, beleg := range belege {
		zeilen = append(zeilen, alsBelegDTO(beleg))
	}

	return zeilen
}

// alsKorpusstandDTOs übersetzt die Stände — ebenso zweimal gebraucht.
func alsKorpusstandDTOs(staende []domain.Korpusstand) []KorpusstandDTO {
	zeilen := make([]KorpusstandDTO, 0, len(staende))
	for _, stand := range staende {
		zeilen = append(zeilen, KorpusstandDTO{
			Quelle: string(stand.Quelle), Revision: stand.Revision,
			Datum: stand.Datum,
		})
	}

	return zeilen
}

// alsBelegDTO übersetzt einen Beleg.
func alsBelegDTO(beleg domain.Beleg) BelegDTO {
	return BelegDTO{
		Kennung: string(beleg.Kennung), Quellkennung: beleg.Quellkennung,
		Quelle:         string(beleg.Quelle),
		Kennungssystem: string(beleg.Kennungssystem),
		Fachkennung:    beleg.Fachkennung,
		Semantik:       string(beleg.Semantik),
		Titel:          beleg.Titel, Ausschnitt: beleg.Ausschnitt,
		Inhaltshash: beleg.Inhaltshash, Sprache: beleg.Sprache,
		Regelnummer: beleg.Regelnummer, Urheber: string(beleg.Urheber),
		Bezugsadresse: beleg.Bezugsadresse,
		Quellrevision: beleg.Quellrevision, Quelldatum: beleg.Quelldatum,
		Abgerufen: beleg.Abgerufen, Veraltet: beleg.Veraltet,
	}
}

// alsFreigabeDTO übersetzt die Freigabe.
func alsFreigabeDTO(freigabe service.Kontextfreigabe) KontextfreigabeDTO {
	kategorien := make([]string, 0, len(freigabe.Kategorien))
	for _, kategorie := range freigabe.Kategorien {
		kategorien = append(kategorien, string(kategorie))
	}

	return KontextfreigabeDTO{
		Zielanbieter: freigabe.Zielanbieter, Zielmodell: freigabe.Zielmodell,
		Ort: string(freigabe.Ort), Zieladresse: freigabe.Zieladresse,
		Kategorien: kategorien,
		Belege:     freigabe.Belege, Zeichen: freigabe.Zeichen,
		Tokens: freigabe.Tokens, Privat: freigabe.Privat,
		Satz: freigabe.Zusammenfassung(),
	}
}

// KontextZusammenstellen ist die Fensterbindung für das Paket.
func (b *Bindungen) KontextZusammenstellen(
	ctx context.Context,
	aufgabe, zielrevision, fall, frage string, grenze int,
	anbieter, modell string, fenster, reserve, hoechstens int,
	freigegeben []string,
) KontextpaketAntwort {
	return b.app(ctx).KontextZusammenstellen(b.ruf(ctx), aufgabe, zielrevision,
		fall, frage, grenze, anbieter, modell, fenster, reserve, hoechstens,
		freigegeben)
}
