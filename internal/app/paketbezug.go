package app

import (
	"context"
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/bezug"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Paketbezieher holt ein Vorlagenpaket — die Naht zu `internal/service`.
type Paketbezieher interface {
	Herkunft(adresse, stand string) (quelle, stand2 string)
	Beziehen(ctx context.Context, adresse, stand, unterverzeichnis, name string) (*service.Bezugsergebnis, error)
}

// BezugAntwort ist der Ausgang eines Paketbezugs.
type BezugAntwort struct {
	Ort        string `json:"ort"`
	Stand      string `json:"stand"`
	Pruefsumme string `json:"pruefsumme"`

	// Uebergangen sind die Dateien, die nicht ausgepackt wurden.
	//
	// **Genannt und nicht verschwiegen** (#139): wer nicht erfährt, dass die
	// Windows-Programme fehlen, sucht später den Fehler im Paket.
	Uebergangen []string `json:"uebergangen"`

	Fehler *Fehler `json:"fehler"`
}

// HerkunftAntwort sagt, wessen Inhalte ein Bezug holen würde.
//
// **Angaben, kein Satz** (§9): der Satz steht in den Sprachkatalogen des
// Frontends, hier stehen die Werte, die er nennt. Ein deutscher Fließtext von
// hier wäre in der englischen Oberfläche deutsch.
type HerkunftAntwort struct {
	Quelle string `json:"quelle"`
	Stand  string `json:"stand"`
}

// PaketHerkunft sagt, wessen Inhalte ein Bezug holen würde.
//
// **Vor dem Bezug, nicht danach.** Die Oberfläche zeigt es und lädt erst, wenn
// jemand zustimmt — das ist die Zusage aus #139, und sie steht hier, weil sie
// sonst nur im Kopf der Oberfläche stünde.
func (a *App) PaketHerkunft(adresse, stand string) HerkunftAntwort {
	quelle, gestand := a.paketbezieher.Herkunft(adresse, stand)

	return HerkunftAntwort{Quelle: quelle, Stand: gestand}
}

// PaketBeziehen holt ein Paket auf diesen Rechner.
func (a *App) PaketBeziehen(ctx context.Context, adresse, stand, unterverzeichnis, name string) BezugAntwort {
	ctx = a.imProfil(ctx)

	leer := BezugAntwort{Uebergangen: []string{}}

	// **Der Fehler vom Start kommt hier ans Licht.** Wer nie ein Paket bezieht,
	// hat ihn nie gebraucht; wer eines bezieht, bekommt ihn statt eines
	// Bezuges an einen Ort, den niemand bestimmt hat.
	if a.paketfehler != nil {
		leer.Fehler = Uebersetze(ctx, a.paketfehler)

		return leer
	}

	ergebnis, err := a.paketbezieher.Beziehen(ctx, adresse, stand, unterverzeichnis, name)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	leer.Ort = ergebnis.Ort
	leer.Stand = ergebnis.Stand
	leer.Pruefsumme = ergebnis.Pruefsumme
	leer.Uebergangen = append(leer.Uebergangen, ergebnis.Uebergangen...)

	// **Und vermerken, was dabei ankam** (#56). Ohne diesen Schritt stünde die
	// Prüfsumme in der Antwort und sonst nirgends — nach dem nächsten Start
	// wäre sie fort, und niemand könnte sagen, woher die 324 Pakete stammen.
	//
	// **Ein Fehler dabei nimmt den Bezug nicht zurück** (§7, Regel 3):
	// geschrieben ist geschrieben, das Paket liegt auf der Platte. Er wird
	// gemeldet — als Fehler an der Antwort, denn der Katalog ist danach nicht
	// auf dem Stand, den er behauptet.
	if _, err := a.pakete.Vermerken(ctx, service.Vermerk{
		Ort:        ergebnis.Ort,
		Herkunft:   herkunftstext(adresse, ergebnis.Stand),
		Pruefsumme: ergebnis.Pruefsumme,
	}); err != nil {
		leer.Fehler = Uebersetze(ctx, err)
	}

	return leer
}

// herkunftstext setzt Adresse und Stand zu einer Herkunft zusammen.
//
// **Angaben, kein Satz** (§9): hier steht kein Fließtext, sondern die beiden
// Werte in einer Form, die sich wiedererkennen lässt — `adresse@stand`.
func herkunftstext(adresse, stand string) string {
	if stand == "" {
		return adresse
	}

	return adresse + "@" + stand
}

// MitPaketbezieher setzt die Naht zum Paketbezug.
func (a *App) MitPaketbezieher(bezieher Paketbezieher) { a.paketbezieher = bezieher }

// Paketverzeichnis gibt den Ort, an dem bezogene Pakete liegen.
//
// Neben der Datenbank unter `~/Library/Application Support/`: es ist Bestand der
// Anwendung, kein Dokument der Nutzenden.
//
// **Am Vorgabeort und nicht neben einem umgelenkten Bestand** — QA-10.NAT-01
// (#1741). Pakete gehören dem **Rechner** und keinem Bestand
// (`docs/AUTORISIERUNG.md`): ein Skript, das mit `MSE_BESTAND` an einem
// Wegwerfbestand übt, zeichnet mit den bezogenen Paketen dieses Rechners. Wer
// sie trennen will, trennt das Zuhause — so baut das Labor jede Instanz.
func Paketverzeichnis() (string, error) {
	return paketablage(true)
}

// paketablage gibt das Paketverzeichnis — und legt es an, wenn gewünscht.
//
// **Eine Stelle für zwei Fragen**: der Bezug braucht das Verzeichnis, die
// Schriftsuche fragt nur ([paketschriftwurzel]). Zwei Wege dorthin waren
// schon einmal zwei Antworten.
func paketablage(anlegen bool) (string, error) {
	pfad, err := Datenbankpfad()
	if err != nil {
		return "", err
	}

	ort := filepath.Join(filepath.Dir(pfad), "pakete")
	if !anlegen {
		return ort, nil
	}

	if err := os.MkdirAll(ort, 0o750); err != nil {
		return "", err
	}

	return ort, nil
}

// QuelleDTO ist eine angebotene Bezugsquelle, wie das Fenster sie sieht.
//
// **Angaben, kein Satz** (§9): `schluessel` benennt den erklärenden Satz in den
// Sprachkatalogen. Ein deutscher Fließtext von hier stünde in der englischen
// Oberfläche deutsch — derselbe Fehler, den der Zweitblick über Phase 9 an
// `Herkunft` angestrichen hat.
type QuelleDTO struct {
	Name             string `json:"name"`
	Adresse          string `json:"adresse"`
	Stand            string `json:"stand"`
	Unterverzeichnis string `json:"unterverzeichnis"`
	Schluessel       string `json:"schluessel"`
	Megabyte         int    `json:"megabyte"`
}

// BezugsquellenAntwort trägt die angebotenen Quellen.
type BezugsquellenAntwort struct {
	Quellen []QuelleDTO `json:"quellen"`
	Fehler  *Fehler     `json:"fehler"`
}

// Bezugsquellen gibt die Quellen, die die Anwendung anbietet (#263).
//
// **Ohne Naht und ohne Bestand.** Sie ist genau das, was jemand vor seinem
// ersten Paket sieht; an einen Katalog gehängt, gäbe sie auf einem frischen
// Rechner eine leere Seite. Die Liste steht in `internal/bezug`, weil sie
// dorthin gehört — sie beschreibt, **woher** bezogen wird.
func (a *App) Bezugsquellen() BezugsquellenAntwort {
	antwort := BezugsquellenAntwort{Quellen: []QuelleDTO{}}

	for _, quelle := range bezug.Quellen() {
		antwort.Quellen = append(antwort.Quellen, QuelleDTO{
			Name:             quelle.Name,
			Adresse:          quelle.Adresse,
			Stand:            quelle.Stand,
			Unterverzeichnis: quelle.Unterverzeichnis,
			Schluessel:       quelle.Schluessel,
			Megabyte:         quelle.Megabyte,
		})
	}

	return antwort
}
