// **Die Stufen, die Seiten und der Schlüssel** — Werkstück G2 von Phase 5.5
// (#1459).
//
// **Aus `spieltextur.go` herausgelöst**, als es über die Zeilengrenze aus §2
// ging. Der Schnitt folgt der Frage: dort steht, **wie** ein Bild entsteht,
// hier, **woran** man es wiedererkennt.

package service

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Texturguete ist eine der fünf benannten Stufen.
type Texturguete string

const (
	// GueteMiniatur ist die Kartenliste neben dem Tisch.
	GueteMiniatur Texturguete = "miniatur"

	// GueteTischNiedrig ist der Tisch, solange noch geladen wird.
	GueteTischNiedrig Texturguete = "tisch-niedrig"

	// GueteTisch ist die Karte, wie sie am Tisch liegt.
	GueteTisch Texturguete = "tisch"

	// GueteBetrachten ist die Karte, die jemand herangezoomt hat.
	GueteBetrachten Texturguete = "betrachten"

	// GueteAusfuhr ist die Größe, in der ein Bild die Anwendung verlässt.
	GueteAusfuhr Texturguete = "ausfuhr"
)

// Texturgueten gibt alle fünf, von klein nach groß.
func Texturgueten() []Texturguete {
	return []Texturguete{
		GueteMiniatur, GueteTischNiedrig, GueteTisch, GueteBetrachten,
		GueteAusfuhr,
	}
}

// IstTexturguete sagt, ob dieses Wort eine Stufe benennt.
func IstTexturguete(guete Texturguete) bool {
	_, bekannt := texturbreiten[guete]

	return bekannt
}

// texturbreiten sind die Zielbreiten in Bildpunkten.
//
// **Messstartwerte und keine feste Schnittstelle** (#1459). Die Spanne 128 bis
// 1536 stammt aus dem Entwurf; gepinnt wird sie erst nach einer Messung am
// echten Tisch, und bis dahin sagt dieser Kommentar, dass sie es nicht ist.
var texturbreiten = map[Texturguete]int{
	GueteMiniatur:     128,
	GueteTischNiedrig: 256,
	GueteTisch:        512,
	GueteBetrachten:   1024,
	GueteAusfuhr:      1536,
}

// Breite gibt die Zielbreite dieser Stufe in Bildpunkten.
func (g Texturguete) Breite() int { return texturbreiten[g] }

// rasterstufe gibt die Güte, mit der der Rasterer zeichnet.
//
// **Klein wird schnell gezeichnet, groß sorgfältig.** Eine Miniatur mit der
// Sorgfalt eines Ausfuhrbildes zu zeichnen kostet Zeit für Pixel, die niemand
// sieht; umgekehrt sieht man jede gesparte Kante.
func (g Texturguete) rasterstufe() render.Guete {
	switch g {
	case GueteMiniatur, GueteTischNiedrig:
		return render.GueteSchnell

	case GueteBetrachten, GueteAusfuhr:
		return render.GueteHoch

	case GueteTisch:
	}

	return render.GueteMittel
}

// Kartenseite sagt, welche **öffentliche** Seite einer Karte gemeint ist.
//
// # Warum der Rücken hier nicht steht
//
// **Weil er keine Seite einer Karte ist.** Das Issue verlangt die Trennung
// ausdrücklich: „Kartenrückseite ist kein Synonym für die zweite öffentliche
// Kartenseite." Der erste Anlauf führte ihn als dritten Wert dieses Typs, und
// damit war er eine Seite unter Seiten — mit einer Kartenkennung, die er
// nicht braucht, und einer Karte, die er nicht kennen darf.
//
// **Jetzt sind es zwei Fragen und zwei Wege**: [Spieltexturen.Textur] fragt
// nach einer Seite **dieser** Karte, [Spieltexturen.Ruecken] nach dem Rücken
// **dieses Spiels**. Wer sie verwechselt, bekommt einen Übersetzungsfehler
// statt eines falschen Bildes.
type Kartenseite string

const (
	// SeiteVorne ist die erste Seite — bei einer einseitigen Karte die
	// ganze.
	SeiteVorne Kartenseite = "vorne"

	// SeiteZweite ist die **zweite öffentliche** Seite.
	//
	// **Sie gibt es nur, wenn die Karte sie führt** — [domain.CardRow.HatZweiteSeite].
	// Ein Stil zeichnet beide **nebeneinander** auf eine Vorlage doppelter
	// Breite (`magic-m15-mainframe-dfc.mse-style`: 752 × 523 statt
	// 375 × 523, `name` bei 58, `name 2` bei 435 — gemessen am 2026-09-15);
	// diese Seite ist die rechte Hälfte davon.
	SeiteZweite Kartenseite = "zweite"
)

// MarkeRuecken ist, was für den Rücken im Schlüssel und in der Antwort steht.
//
// **Ein Wort und keine [Kartenseite]** — der Rücken gehört einem Spiel und
// nicht einer Karte. Im Schlüssel braucht er trotzdem eine Stelle, sonst
// hinge er allein am Plan.
//
// **Exportiert, weil zwei andere Stellen dasselbe Wort schreiben**: die Naht
// legt es in die Antwort, und `internal/werkzeuge/texturwoerter_test.go`
// hält die Wörter des Fensters dagegen. Drei Fassungen liefen beim ersten
// Umbau auseinander.
const MarkeRuecken = "ruecken"

// Kartenseiten gibt beide.
func Kartenseiten() []Kartenseite {
	return []Kartenseite{SeiteVorne, SeiteZweite}
}

// IstKartenseite sagt, ob dieses Wort eine Seite benennt.
func IstKartenseite(seite Kartenseite) bool {
	for _, bekannt := range Kartenseiten() {
		if seite == bekannt {
			return true
		}
	}

	return false
}

// Texturwunsch ist, was außer dem Plan in den Schlüssel geht.
type Texturwunsch struct {
	// Seite ist eine [Kartenseite] oder [markeRuecken].
	Seite string

	Guete Texturguete

	// Gesichter ist, wie viele Seiten die Vorlage **nebeneinander** trägt —
	// eins bei jeder gewöhnlichen Karte, zwei bei einer zweiseitigen.
	//
	// **Sie geht in den Schlüssel**, weil sie den Ausschnitt bestimmt: aus
	// demselben Plan entstehen bei zwei Gesichtern zwei Bilder.
	Gesichter int
}

// Texturschluessel gibt den Schlüssel zu einem Plan und einem Wunsch.
//
// **Öffentlich, weil er die Frage beantwortet, ohne zu zeichnen** — wer
// vorladen will, braucht ihn vorher.
//
// # Warum er nicht scheitern kann
//
// **Weil eine fehlende Datei ein Wert ist und kein Fehler.** Wer eine Karte
// mit fehlendem Rahmen zeichnet, bekommt ein Bild mit einem Befund daneben —
// und dieses Bild soll wiedererkannt werden. Ein Fehlerweg hier machte aus
// einem Befund einen Abbruch, und die Karte verschwände ganz. Die Quellen
// geben deshalb für „fehlt" einen **anderen** Wert als für jede Datei: taucht
// sie auf, ändert sich der Schlüssel.
func Texturschluessel(
	plan *render.Plan, ressourcen Ressourcenquelle, wunsch Texturwunsch,
) string {
	summe := sha256.New()

	schreibeAbdruckteil(summe, "fassung", render.Fassung)
	schreibeAbdruckteil(summe, "seite", wunsch.Seite)
	schreibeAbdruckteil(summe, "guete", string(wunsch.Guete))
	schreibeAbdruckteil(summe, "breite", strconv.Itoa(wunsch.Guete.Breite()))
	schreibeAbdruckteil(summe, "gesichter", strconv.Itoa(wunsch.Gesichter))
	schreibeAbdruckteil(summe, "plan", render.Planabdruck(*plan))

	// **Die Dateien, die der Plan nur nennt.** Ohne sie wäre ein
	// ausgetauschter Rahmen unsichtbar — und genau das ist die Klasse „Bild"
	// aus dem Issue.
	for _, pfad := range render.Bildpfade(*plan) {
		schreibeAbdruckteil(summe, "bild:"+pfad, ressourcen.Bildabdruck(pfad))
	}

	for _, name := range render.Schriftnamen(*plan) {
		schreibeAbdruckteil(summe, "schrift:"+name,
			ressourcen.Schriftabdruck(name))
	}

	return hex.EncodeToString(summe.Sum(nil))
}

// schreibeAbdruckteil hängt ein benanntes Stück an den Abdruck.
//
// **Mit Namen und Länge** — sonst gäbe `ab` + `c` denselben Abdruck wie `a` +
// `bc`, und zwei verschiedene Schlüssel wären einer.
//
// **`schreibe` heißt es nicht**: den Namen trägt schon der Paketschreiber in
// `msepaket`-Nähe, und zwei gleiche Namen wären in jedem Review eine
// Rückfrage.
func schreibeAbdruckteil(summe hash.Hash, name, wert string) {
	//nolint:errcheck // ein Hash klagt nicht (`hash.Hash`: „never returns an error")
	_, _ = summe.Write([]byte(
		name + "\x1f" + strconv.Itoa(len(wert)) + "\x1f" + wert + "\x1e"))
}

// vergroesserungFuer rechnet aus der Zielbreite den Faktor des Rasterers.
//
// **Der Rasterer kennt keine Bildpunkte**, er kennt Vorlagenpunkte und einen
// Faktor. Ein Plan ohne Breite gäbe eine Division durch null — er bekommt den
// Faktor eins und damit die Vorlage in ihren Punkten.
func vergroesserungFuer(plan *render.Plan, breite, gesichter int) float64 {
	if plan.Breite <= 0 {
		return 1
	}

	// **Die Zielbreite gilt für ein Gesicht und nicht für die Vorlage.** Wer
	// `tisch` verlangt, will eine 512 Punkte breite **Karte**; eine
	// zweiseitige Vorlage ist doppelt so breit, und ohne diesen Faktor käme
	// jede ihrer Seiten halb so groß daher wie jede andere Karte am Tisch.
	if gesichter > 1 {
		breite *= gesichter
	}

	return float64(breite) / plan.Breite
}

// benannt zählt die erlaubten Werte auf.
func benannt[T ~string](werte []T) string {
	namen := make([]string, 0, len(werte))
	for _, wert := range werte {
		namen = append(namen, string(wert))
	}

	return strings.Join(namen, ", ")
}
