package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/bezug"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Paketbezieher holt Vorlagenpakete in ein Verzeichnis auf diesem Rechner.
type Paketbezieher struct {
	wurzel string

	// kenner ist die Selbstnennung im `User-Agent`.
	//
	// **Gereicht und nicht gebaut** (#899): `internal/bezug` ist ein Blatt und
	// kennt die Fassung der Anwendung nicht (§2). Sie hier ein zweites Mal zu
	// schreiben wäre der Fehler aus #896 — eine Zahl an einem Ort, den kein
	// Wächter kennt.
	kenner string

	// api ist die Wurzel der Auskunft über Stände — leer heißt GitHub
	// ([githubauskunft]). Ein Lauf setzt sie auf einen Testserver.
	api string
}

// NeuerPaketbezieher baut den Bezieher über einem Zielverzeichnis.
func NeuerPaketbezieher(wurzel, kenner string) *Paketbezieher {
	return &Paketbezieher{wurzel: wurzel, kenner: kenner}
}

// Bezugsergebnis ist der Ausgang eines Bezugs.
type Bezugsergebnis struct {
	// Adresse ist die Quelle, aus der bezogen wurde — eine Aktualisierung
	// nennt sie nicht, sie kommt aus dem Vermerk des Baums (G04).
	Adresse     string
	Ort         string
	Stand       string
	Pruefsumme  string
	Uebergangen []string
}

// Herkunft sagt, wessen Inhalte ein Bezug holen würde.
//
// **Vor dem Bezug, nicht danach** (#139, §11): wer nicht weiß, woher eine
// Vorlage kommt, kann ihre Rechtslage nicht einschätzen. Die Anwendung zeigt
// die Angaben in ihrer Sprache und lädt erst, wenn jemand zustimmt.
func (p *Paketbezieher) Herkunft(adresse, stand string) (quelle, gestand string) {
	woher := bezug.Woher(bezug.Repositorium{Adresse: adresse, Stand: stand})

	return woher.Quelle, woher.Stand
}

// Beziehen holt ein Paket und legt es unter seinem Namen ab.
//
// **Bezug nur auf ausdrückliche Anweisung** (#56): ein Programm, das ungefragt
// über ein Gigabyte lädt, überrascht seine Nutzenden. Dass die Anweisung
// vorliegt, entscheidet der Aufrufer — hier wird sie ausgeführt.
func (p *Paketbezieher) Beziehen(
	ctx context.Context, adresse, stand, unterverzeichnis, name string,
) (*Bezugsergebnis, error) {
	return p.beziehen(ctx, adresse, stand, unterverzeichnis, name, "")
}

// beziehen ist [Paketbezieher.Beziehen] — und hebt, wenn `aufheben` einen
// Ort nennt, den bisherigen Baum dort auf, statt ihn zu räumen (G04,
// `paketaktualisierungstausch.go`).
func (p *Paketbezieher) beziehen(
	ctx context.Context, adresse, stand, unterverzeichnis, name, aufheben string,
) (*Bezugsergebnis, error) {
	if err := p.wurzelPruefen(); err != nil {
		return nil, err
	}

	if err := namePruefen(name); err != nil {
		return nil, err
	}

	ort := filepath.Join(p.wurzel, name)

	// **Erst ins Leere, dann an den Platz.** Ein Bezug, der auf halber Strecke
	// scheitert, hinterlässt sonst ein halbes Paket unter dem Namen eines
	// ganzen — und das nächste Öffnen findet es und hält es für vollständig.
	vorlaeufig, err := os.MkdirTemp(p.wurzel, "."+name+"-")
	if err != nil {
		return nil, fmt.Errorf("Paket beziehen: %w", err)
	}

	defer func() { _ = os.RemoveAll(vorlaeufig) }() //nolint:errcheck // nach einem geglückten Umbenennen gibt es ihn nicht mehr

	ergebnis, err := bezug.Hole(ctx, bezug.Repositorium{
		Adresse:          archivadresse(adresse),
		Marken:           markenadresse(adresse),
		Stand:            stand,
		Unterverzeichnis: unterverzeichnis,
	}, vorlaeufig, p.kenner)
	if err != nil {
		return nil, err
	}

	// **Der Baum sagt, wie er bezogen wurde** (G04) — sonst wüsste eine
	// spätere Aktualisierung das Unterverzeichnis nicht, und ein Zurück nicht,
	// was es zurückholt.
	if err := vermerkSchreiben(vorlaeufig, Bezugsvermerk{
		Adresse: adresse, Stand: ergebnis.Stand, Unterverzeichnis: unterverzeichnis, Pruefsumme: ergebnis.Pruefsumme,
	}); err != nil {
		return nil, err
	}

	// **Niemand schreibt gerade darunter** (#1190). Die Schreibsperre der
	// Paketwerkstatt liegt an einer **Datei**; dieser Weg tauscht den **Baum**
	// aus, und die beiden Ebenen sahen einander nicht.
	//
	// **Erst hier und nicht um den ganzen Bezug.** Was zu schützen ist, ist
	// der **Tausch**; das Herunterladen dauert bei `M15-Magic-Pack` Minuten,
	// und in dieser Zeit die Werkstatt zu sperren hieße, sie für die Dauer
	// eines Bezugs zu schließen. Ein Schreiber, der währenddessen anfängt,
	// hält seine geteilte Sperre — und der Tausch wartet auf ihn.
	fertig, sperrfehler := alleinAmBaum(ort)
	if sperrfehler != nil {
		return nil, sperrfehler
	}

	defer fertig()

	var einsetzen error
	if aufheben == "" {
		einsetzen = paketEinsetzen(ort, vorlaeufig, os.Rename)
	} else {
		einsetzen = paketEinsetzenUndAufheben(ort, vorlaeufig, aufheben, os.Rename)
	}

	if einsetzen != nil {
		return nil, fmt.Errorf("Paket an seinen Platz bringen: %w", einsetzen)
	}

	return &Bezugsergebnis{
		Adresse:     adresse,
		Ort:         ort,
		Stand:       ergebnis.Stand,
		Pruefsumme:  ergebnis.Pruefsumme,
		Uebergangen: ergebnis.Uebergangen,
	}, nil
}

// wurzelPruefen weist eine Wurzel ab, die kein fester Ort ist.
//
// # Warum das hier steht und nicht nur im Composition Root
//
// **Weil hier gelöscht wird.** `Beziehen` räumt den Platz des Pakets, bevor es
// das neue dorthin umbenennt. Ist die Wurzel leer, wird daraus ein **relativer**
// Pfad — und geräumt würde ein gleichnamiges Verzeichnis dort, wo der Prozess
// gerade steht.
//
// Der Weg dahin war real: der Composition Root verwarf den Fehler von
// `app.Paketverzeichnis()`, und eine Datei namens `pakete` an der Stelle des
// Verzeichnisses genügt, damit er auftritt. Der Zweitblick über Phase 9 hat es
// gefunden.
//
// Eine Prüfung im Aufrufer allein wäre die halbe Zusage: sie gälte für den
// einen Aufrufer, den es heute gibt.
func (p *Paketbezieher) wurzelPruefen() error {
	if !filepath.IsAbs(p.wurzel) {
		return domain.NewFieldError("wurzel", domain.SchluesselValidierung,
			"%q ist kein fester Ort für Pakete. Ohne einen wüsste niemand, wo geräumt wird.",
			p.wurzel)
	}

	return nil
}

// namePruefen weist einen Namen ab, der aus dem Verzeichnis führt.
func namePruefen(name string) error {
	if name == "" || strings.ContainsRune(name, filepath.Separator) || name == ".." || name == "." {
		return domain.NewFieldError("name", domain.SchluesselValidierung,
			"%q ist kein Paketname: er muss ein einzelner Verzeichnisname sein.", name)
	}

	return nil
}

// archivadresse macht aus einer Repositoriumsadresse die des Archivs.
//
// **GitHub ist der Regelfall, nicht die Ausnahme**: MSE-Vorlagen liegen heute
// als Repositorien dort. `codeload` liefert einen Stand als `tar.gz`, ohne dass
// wir einen Zugang brauchen.
func archivadresse(adresse string) string {
	kern := strings.TrimSuffix(strings.TrimSuffix(adresse, "/"), ".git")

	if rest, ist := strings.CutPrefix(kern, "https://github.com/"); ist {
		return "https://codeload.github.com/" + rest + "/tar.gz"
	}

	return kern + "/tar.gz"
}

// markenadresse gibt den Ort, an dem sich eine Freigabe nachschlagen lässt.
func markenadresse(adresse string) string {
	kern := strings.TrimSuffix(strings.TrimSuffix(adresse, "/"), ".git")

	if rest, ist := strings.CutPrefix(kern, "https://github.com/"); ist {
		return "https://api.github.com/repos/" + rest + "/git/ref/tags/"
	}

	return kern + "/tags/"
}

// paketEinsetzen setzt das bezogene Paket an den Platz des bisherigen.
//
// # Was vorher galt
//
// `errors.Join(os.RemoveAll(ort), os.Rename(vorlaeufig, ort))` — **erst
// räumen, dann umbenennen**, mit der Begründung, nach einem geglückten Räumen
// könne das Umbenennen im selben Verzeichnis kaum noch scheitern.
//
// „Kaum" ist keine Zusage. Zwischen beiden Zügen liegt ein Augenblick, in dem
// das bisherige Paket **weg** ist und das neue noch nicht steht; scheitert
// der zweite Zug oder bricht der Prozess dazwischen ab, hat jemand ein Paket
// verloren, das er hatte. Ein `RemoveAll`, das auf halber Strecke scheitert,
// lässt sogar einen zerlegten Rest stehen — gemeldet als „Paket an seinen
// Platz bringen".
//
// Angestrichen vom zweiten Zweitblick über die Freigabe 0.3.0
// (#1170,
// Befund 3).
//
// # Was jetzt gilt
//
// **Beiseite, einsetzen, wegräumen** — dieselbe Gestalt wie `austauschen` in
// `internal/msepaket` und aus demselben Grund. Der bisherige Stand ist bis zum
// geglückten zweiten Zug vollständig da; scheitert er, wird zurückgeholt.
//
// **`umbenennen` kommt herein**, damit dieser Weg zu betreten ist: einen
// Fehlschlag genau zwischen zwei Umbenennungen kann kein Dateisystem auf
// Bestellung liefern (§3).
//
// # Und was es nicht hält
//
// Nach einem gescheiterten **dritten** Zug — dem Zurückholen — liegt der
// bisherige Stand unter einem Namen mit führendem Punkt daneben, und die
// Meldung sagt, unter welchem. Weiter kommt man mit Umbenennungen nicht.
func paketEinsetzen(ort, vorlaeufig string, umbenennen func(alt, neu string) error) error {
	beiseite := ""

	if _, err := os.Stat(ort); err == nil {
		beiseite = ort + ".vorher"
		if err := umbenennen(ort, beiseite); err != nil {
			return err
		}
	}

	if err := umbenennen(vorlaeufig, ort); err != nil {
		if beiseite == "" {
			return err
		}

		if zurueck := umbenennen(beiseite, ort); zurueck != nil {
			return fmt.Errorf("%w — und der bisherige Stand liegt jetzt in %q: %w",
				err, beiseite, zurueck)
		}

		return err
	}

	if beiseite != "" {
		// **Der neue Stand steht — ein Rest daneben verwirft ihn nicht** (§7,
		// Regel 3: geschrieben ist geschrieben). Aber er wird aufgeschrieben:
		// ein liegengebliebenes `.vorher` wiegt bei `M15-Magic-Pack` über ein
		// Gigabyte, und still verbraucht es den Platz von jemandem.
		if fehler := os.RemoveAll(beiseite); fehler != nil {
			slog.Default().Warn("der bisherige Paketstand blieb liegen",
				"ort", beiseite, "grund", fehler.Error())
		}
	}

	return nil
}

// paketEinsetzenUndAufheben setzt einen bezogenen Baum ein und hebt den
// bisherigen unter `aufheben` auf — eine Ebene tief.
//
// **Der schon aufgehobene weicht erst, wenn alles gelang**: scheitert das
// Einsetzen, kehren beide an ihren Platz zurück — derselbe Grund wie in
// [paketEinsetzen] und beim Ersetzen eines einzelnen Pakets (Befund Z1 des
// Zweitblicks über Phase 6.5).
func paketEinsetzenUndAufheben(ort, vorlaeufig, aufheben string, umbenennen func(alt, neu string) error) error {
	if err := os.MkdirAll(filepath.Dir(aufheben), 0o750); err != nil {
		return err
	}

	alt := ""

	if _, err := os.Stat(aufheben); err == nil {
		alt = aufheben + ".alt"
		if err := umbenennen(aufheben, alt); err != nil {
			return err
		}
	}

	// **Ein Rückweg, der scheitert, steht im Fehler** (Befund 1 des
	// Zweitblicks über Phase G, zweiter Teil): bis dahin verwarf das
	// Einsetzen ihn, und der Bericht sagte nicht, dass ein Baum nicht an
	// seinem Platz liegt. `errors.Join` übergeht ein nil.
	zurueck := func() error {
		if alt == "" {
			return nil
		}

		return umbenennen(alt, aufheben)
	}

	if err := umbenennen(ort, aufheben); err != nil {
		return errors.Join(err, zurueck())
	}

	if err := umbenennen(vorlaeufig, ort); err != nil {
		return errors.Join(err, umbenennen(aufheben, ort), zurueck())
	}

	if alt != "" {
		// **Geschrieben ist geschrieben** (§7, Regel 3) — ein Rest daneben
		// verwirft den neuen Stand nicht, aber er wird aufgeschrieben.
		if fehler := os.RemoveAll(alt); fehler != nil {
			slog.Default().Warn("ein älterer aufgehobener Stand blieb liegen", "ort", alt, "grund", fehler.Error())
		}
	}

	return nil
}
