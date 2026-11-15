package aktualisierung

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Adresse ist die REST-Schnittstelle von GitHub.
const Adresse = "https://api.github.com"

// Repositorium ist das Repositorium, dessen Freigaben gelten.
//
// **Heute privat, später öffentlich** — Entscheidung D01
// ([0190](../../docs/entscheidungen/0190-das-repositorium-wird-oeffentlich.md)).
// Bis dahin gilt die Funktion für Menschen mit Lesezugriff, und ohne Marke
// antwortet GitHub mit 404; danach fragt jeder ohne Marke, und es gilt die
// Grenze für Anfragen ohne Anmeldung ([grenzeErschoepft]).
const Repositorium = "DavidTHeider/MagicSetEditor4Mac"

// Wirt ist der Name, unter dem eine Drosselung den Dienst nennt — der Wirt
// von [Adresse], auch wenn ein Test eine andere Grundadresse einsetzt.
const Wirt = "api.github.com"

// Kenner ist der Name, unter dem sich dieser Klient meldet.
//
// **Pflicht bei GitHub**, nachgelesen am 2026-08-25: „Requests with no
// `User-Agent` header will be rejected."
var Kenner = netz.Kenner{
	Name:    "MagicSetEditor4Mac (+https://github.com/" + Repositorium + ")",
	Annahme: "application/vnd.github+json",
}

// Beigabenkenner ist der Kenner für den **Inhalt** einer Beigabe.
//
// **Dieselbe Adresse, zwei Antworten.** Nachgelesen am 2026-08-25
// (`https://docs.github.com/en/rest/releases/assets`, nachgemessen in
// #1100):
// mit `application/vnd.github+json` gibt `…/releases/assets/{id}` die
// **Beschreibung** der Beigabe, erst mit `application/octet-stream` ihren
// **Inhalt**. Wer das verwechselt, lädt ein Stück JSON und rechnet dessen
// Prüfsumme.
var Beigabenkenner = netz.Kenner{
	Name:    Kenner.Name,
	Annahme: "application/octet-stream",
}

// Drosselabstand ist der Mindestabstand zweier Anfragen.
//
// **Weit unter der Grenze** (§16, Satz 4: eingehalten, bevor sie greift).
// GitHub nennt am 2026-08-25 60 Anfragen je Stunde ohne Marke und 5.000 mit
// einer. Diese Anwendung fragt höchstens beim Start und auf Knopfdruck; ein
// Abstand von zwei Sekunden liegt bei jeder der beiden Grenzen weit darunter
// und fängt zugleich den Fall ab, dass jemand den Knopf drückt und hält.
const Drosselabstand = 2 * time.Second

// Hoechstens ist die Grenze für die Antwort auf eine Freigabeanfrage.
//
// **Eine Freigabeantwort ist ein Kilobyte groß**; 256 KB sind das Vielfache
// und trotzdem eine Grenze. Ohne sie hinge der Speicherbedarf an dem, was ein
// fremder Rechner schickt.
const Hoechstens = 256 << 10

// Klient fragt GitHub nach der neuesten Freigabe.
//
// **Klient, Uhr und Drossel werden eingesetzt, nicht erzeugt** (§16): sonst
// braucht jeder Test das Netz.
type Klient struct {
	grundadresse string
	unten        *netz.Klient
	drossel      *netz.Drossel

	// marke ist die Zugangsmarke — leer, wenn es keine gibt.
	//
	// **Eingesetzt und nicht selbst gelesen**: ein Klient, der sich seine
	// Marke aus der Umgebung nähme, hätte in jedem Test die dieses Rechners
	// (§3).
	marke string
}

// NeuerKlient baut einen Klienten.
//
// **Die Marke ist der vierte Parameter und kein Nachtrag.** Ohne sie
// antwortet GitHub auf ein privates Repositorium mit 404 — und das ist die
// Vorgabe, wenn niemand eine setzt.
func NeuerKlient(
	grundadresse string, unten *netz.Klient, werk *netz.Drosselwerk, marke string,
) *Klient {
	return &Klient{
		grundadresse: strings.TrimSuffix(grundadresse, "/"),
		unten:        unten,
		drossel:      werk.Fuer("github", Drosselabstand),
		marke:        marke,
	}
}

// MarkeAusUmgebung liest die Marke aus `GH_TOKEN`, sonst `GITHUB_TOKEN`.
//
// **Dieselben zwei Namen und dieselbe Vorrangfolge wie bei `gh`**,
// nachgelesen am 2026-08-25 (`gh help environment`): „GH_TOKEN,
// GITHUB_TOKEN … the first of these environment variables to be set is used."
//
// **Nötig ist die Berechtigung `Contents` in der Stufe `read`** — nachgelesen
// am 2026-08-25 unter
// `https://docs.github.com/en/rest/authentication/permissions-required-for-fine-grained-personal-access-tokens`
// für `GET /repos/{owner}/{repo}/releases/latest` und
// `GET /repos/{owner}/{repo}/releases/assets/{asset_id}`; beide brauchen keine
// weitere. Mit einer klassischen Marke ist es der Bereich `repo`.
func MarkeAusUmgebung() string {
	if marke := os.Getenv("GH_TOKEN"); marke != "" {
		return marke
	}

	return os.Getenv("GITHUB_TOKEN")
}

// kenner gibt den Kenner für Auskünfte — mit Marke, wenn eine da ist.
func (k *Klient) kenner() netz.Kenner {
	mit := Kenner
	mit.Marke = k.marke

	return mit
}

// beigabenkenner gibt den Kenner für Inhalte — mit Marke, wenn eine da ist.
func (k *Klient) beigabenkenner() netz.Kenner {
	mit := Beigabenkenner
	mit.Marke = k.marke

	return mit
}

// Freigabe ist eine Freigabe, wie diese Anwendung sie braucht.
//
// **Die Zuordnung und nicht das fremde DTO** (§16, dritte Schicht): was
// GitHub schickt, steht in [freigabeDTO] und bleibt dort. Ändert der Anbieter
// seine Feldnamen, ändert sich eine Stelle.
type Freigabe struct {
	// Fassung ist die Marke ohne führendes `v` — `0.9.0`.
	Fassung string

	// Seite ist die Freigabeseite, zu der eine Meldung führen kann.
	Seite string

	// Erschienen ist der Zeitpunkt der Freigabe — Nullwert, wenn unbekannt.
	Erschienen time.Time

	// Abbild ist das `.dmg` — Nullwert, wenn die Freigabe keines trägt.
	Abbild Beigabe

	// WeitereAbbilder nennt jedes weitere `.dmg` derselben Freigabe — I06
	// (#2264).
	// **Dann wird nichts geladen** ([Klient.Laden]): welches gälte, entschiede
	// sonst die Reihenfolge, in der die Gegenstelle die Beigaben nennt.
	WeitereAbbilder []string

	// Pruefsummen ist `pruefsummen.txt` — Nullwert, wenn sie fehlt.
	//
	// **Sie liegt jeder Freigabe bei**, und damit lässt sich das Geladene
	// prüfen, bevor irgendjemand es öffnet.
	Pruefsummen Beigabe

	// Manifest ist `manifest.json` — Nullwert, wenn es fehlt.
	//
	// **Eine Freigabe ohne Manifest ist keine ungültige**, sondern eine von
	// vor Werkstück O6 (#1430).
	// Was daraus folgt, entscheidet `domain.Vertraeglichkeit`: **ungeklärt**,
	// nicht unverträglich.
	Manifest Beigabe

	// Quelltext ist `MagicSetEditor4Mac-<fassung>-quelltext.tar.gz` —
	// Nullwert, wenn er fehlt. **Er liegt seit D01 jeder Freigabe bei**
	// (Vermerk 0190, `scripts/releasequelltext.sh`): GPL-2.0 §3 a, der
	// vollständige Quelltext neben dem Abbild.
	Quelltext Beigabe

	// Signatur ist `pruefsummen.txt.asc` — Nullwert, wenn sie fehlt.
	//
	// **Ohne sie sagt die Prüfsumme nur, dass zwei Dateien zueinander
	// passen** — nicht, woher sie kommen. Der Ladeweg weist eine Freigabe
	// ohne Signatur ab; der Grund steht in `signatur.go`.
	Signatur Beigabe
}

// Beigabe ist eine Datei, die zu einer Freigabe gehört.
type Beigabe struct {
	Name    string
	Groesse int64

	// Adresse ist der Weg zum Inhalt — mit `Accept: application/octet-stream`.
	Adresse string
}

// freigabeDTO ist die Gestalt, die GitHub schickt.
//
// **Pflichtfelder streng, Zusatzfelder tolerant** (§16). `tag_name` ist das
// einzige Feld, ohne das nichts geht; alles andere darf fehlen, und was
// GitHub zusätzlich schickt, wird übergangen.
type freigabeDTO struct {
	Marke      string `json:"tag_name"`
	Seite      string `json:"html_url"`
	Erschienen string `json:"published_at"`

	Beigaben []struct {
		Name    string `json:"name"`
		Groesse int64  `json:"size"`
		Adresse string `json:"url"`
	} `json:"assets"`
}

// NeuesteFreigabe holt die jüngste Freigabe.
//
// **Ein 404 ist ein fachliches Nein und kein Ausfall** (§16, Satz 4):
// solange das Repositorium privat ist, gibt es ohne Lesezugriff keine
// Freigabe zu sehen. Wiederholt wird das nicht. **Eine erschöpfte Grenze ist
// keines von beiden**, sondern eine Drossel mit Uhrzeit ([grenzeErschoepft]).
func (k *Klient) NeuesteFreigabe(ctx context.Context) (Freigabe, error) {
	adresse := k.grundadresse + "/repos/" + Repositorium + "/releases/latest"

	antwort, err := netz.Hole(ctx, k.unten, k.drossel, k.kenner(), adresse, Hoechstens)

	// **Ein 429 kommt als Fehler, ein 403 als Antwort** — gedeutet werden
	// beide an derselben Stelle ([grenzeErschoepft]).
	if bis, erschoepft := grenzeErschoepft(antwort, k.drossel.Jetzt()); erschoepft {
		// **Gelesen und nicht durchbrochen** (§16, Satz 4): bis dahin fragt
		// dieser Klient nicht wieder — und **keine Drossel dieses Werks**,
		// denn die Grenze gilt der Gegenstelle und nicht einem Weg zu ihr.
		k.drossel.SperreGegenstelle(bis)

		// **Gemeldet wird das Ende, das gilt** (Nachlauf über Phase 0,
		// Befund 2): beim 429 hat [netz.Hole] schon gesperrt, und die
		// längere Sperre bleibt.
		return Freigabe{}, domain.NeueDrosselung(
			fmt.Errorf("GitHub antwortete mit %d", antwort.Status), Wirt,
			k.drossel.GesperrtBis())
	}

	if err != nil {
		return Freigabe{}, fmt.Errorf("nach der neuesten Fassung fragen: %w", err)
	}

	if antwort.Status == netz.StatusNichtGefunden {
		return Freigabe{}, fmt.Errorf(
			"%w: keine Freigabe zu sehen — solange das Repositorium privat "+
				"ist, antwortet GitHub ohne Lesezugriff so", domain.ErrNotFound)
	}

	if antwort.Status != netz.StatusGut {
		return Freigabe{}, fmt.Errorf(
			"%w: GitHub antwortete mit %d", domain.ErrConflict, antwort.Status)
	}

	var roh freigabeDTO
	if err := json.Unmarshal(antwort.Rumpf, &roh); err != nil {
		return Freigabe{}, fmt.Errorf("die Antwort von GitHub lesen: %w", err)
	}

	return zuFreigabe(roh)
}

// zuFreigabe ordnet das fremde DTO auf das zu, was diese Anwendung kennt.
func zuFreigabe(roh freigabeDTO) (Freigabe, error) {
	if strings.TrimSpace(roh.Marke) == "" {
		return Freigabe{}, fmt.Errorf(
			"%w: die Freigabe trägt keine Marke (`tag_name`) — ohne sie ist "+
				"nicht zu sagen, welche Fassung sie ist", domain.ErrValidation)
	}

	freigabe := Freigabe{
		Fassung: strings.TrimPrefix(strings.TrimSpace(roh.Marke), "v"),
		Seite:   roh.Seite,
	}

	// **Ein unlesbarer Zeitpunkt ist kein Fehlschlag.** Er steht neben der
	// Meldung und trägt sie nicht; ohne ihn fehlt ein Satz, nicht die Auskunft.
	if erschienen, err := time.Parse(time.RFC3339, roh.Erschienen); err == nil {
		freigabe.Erschienen = erschienen
	}

	for _, beigabe := range roh.Beigaben {
		nach := Beigabe{
			Name: beigabe.Name, Groesse: beigabe.Groesse, Adresse: beigabe.Adresse,
		}

		switch {
		case strings.HasSuffix(strings.ToLower(beigabe.Name), ".dmg") && freigabe.Abbild.Name != "":
			freigabe.WeitereAbbilder = append(freigabe.WeitereAbbilder, beigabe.Name)
		case strings.HasSuffix(strings.ToLower(beigabe.Name), ".dmg"):
			freigabe.Abbild = nach
		case strings.EqualFold(beigabe.Name, "pruefsummen.txt"):
			freigabe.Pruefsummen = nach
		case strings.EqualFold(beigabe.Name, "pruefsummen.txt.asc"):
			freigabe.Signatur = nach
		case strings.EqualFold(beigabe.Name, Manifestname):
			freigabe.Manifest = nach
		case strings.HasSuffix(strings.ToLower(beigabe.Name), "-quelltext.tar.gz"):
			freigabe.Quelltext = nach
		}
	}

	return freigabe, nil
}
