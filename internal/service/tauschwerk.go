package service

// **Der Abgleich zweier Sammlungen** — Werkstück `T4` von Phase 5.6
// (#1469).
//
// # Warum die fremde Seite ein Schnappschuss ist
//
// **Weil sie freigegeben wurde und nicht abgefragt wird.** Was ein Gegenüber
// zeigt, entscheidet seine Freigabe (T3, [#1468]); hier kommt an, was es
// gezeigt hat. Ein Abgleich, der selbst nachschlüge, ginge an der Freigabe
// vorbei — und er ginge ins Netz, mitten in einer Rechnung, die auf
// Tastendruck läuft.
//
// # Warum das Konzept aus dem eigenen Verzeichnis kommt
//
// **Weil die fremde Zeile nur ihren Druck nennt.** Sie trägt sechs Felder und
// keines mehr; welche Karte hinter einem Druck steht, weiß dieses Programm
// selbst — aus `drucke` (A2, [#1441]). Damit trifft ein druckunabhängiger
// Wunsch, ohne dass drüben mehr hinausgehen muss.
//
// # Und warum das Verzeichnis **einmal** gelesen wird
//
// **Weil ein Nachschlag je Zeile die Rechnung quadriert.** Das Issue verlangt
// ausdrücklich „keine Netzabrufe pro Vergleichszeile und keine
// O(n*m)-Vollpaarung"; gelesen wird deshalb einmal, und danach steht ein
// Verzeichnis im Arbeitsspeicher.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Druckverzeichnis ist, was der Abgleich von der Druckablage braucht.
//
// **Alle auf einmal und nicht einzeln** — derselbe Grund wie bei
// [Preisablage.NeuesteAlle].
type Druckverzeichnis interface {
	Alle(ctx context.Context) ([]domain.Druck, error)
}

// Fremdstand ist, was von einem Gegenüber angekommen ist.
//
// **Zwei Listen und nicht eine.** Was es hergeben kann und was es sucht sind
// zwei Freigaben mit zwei Umfängen; eine gemeinsame Liste verlöre, welche
// Zeile welche war.
type Fremdstand struct {
	// Kennung ist die Peerkennung — sie steht hier, damit ein Ergebnis sagen
	// kann, zu wem es gehört.
	Kennung string

	// Angebot ist, was das Gegenüber hergeben kann.
	Angebot []domain.Sichtkarte

	// Wuensche ist, was es sucht.
	Wuensche []domain.Sichtkarte
}

// Tauschabgleich ist das Ergebnis — beide Richtungen und ihr Vorschlag.
type Tauschabgleich struct {
	// Kennung ist das Gegenüber, zu dem dieser Abgleich gehört.
	Kennung string

	// Gebe sind meine Dubletten für seine Wünsche.
	Gebe []domain.Tauschkandidat

	// Bekomme sind seine Zeilen für meine Wünsche.
	Bekomme []domain.Tauschkandidat

	// Vorschlag gilt nur mit [Tauschabgleich.HatVorschlag].
	Vorschlag domain.Tauschvorschlag

	// HatVorschlag sagt, ob es einen gibt — **ohne beide Seiten gibt es
	// keinen** (der leere Tausch ist keiner).
	HatVorschlag bool

	// Befunde sind die Stellen, an denen eine Kennung gefehlt hat.
	Befunde []domain.Befund
}

// Tauschwerk gleicht die eigene Sammlung gegen einen fremden Stand ab.
type Tauschwerk struct {
	sichten *Sammlungssichten
	drucke  Druckverzeichnis

	// preise ist wahlfrei — **ohne sie wiegt kein Vorschlag**, und das ist
	// besser, als mit einer geratenen Reihe zu wiegen.
	preise Preisverzeichnis

	// reihe ist die **ausdrücklich gewählte** Preisreihe (A3).
	reihe domain.Preisstand
}

// NeuesTauschwerk stellt den Abgleich auf.
func NeuesTauschwerk(sichten *Sammlungssichten, drucke Druckverzeichnis) *Tauschwerk {
	return &Tauschwerk{sichten: sichten, drucke: drucke}
}

// MitPreisreihe hängt die monetäre Auskunft an — **eine Reihe, ausdrücklich**.
//
// **Und keine Vorgabe.** Welche Quelle, welcher Markt und welche Richtung
// gelten, ist eine Entscheidung dessen, der tauscht; eine hier gewählte
// Vorgabe stünde später in einer Zahl, die niemand gewollt hat (A3).
func (t *Tauschwerk) MitPreisreihe(
	preise Preisverzeichnis, reihe domain.Preisstand,
) *Tauschwerk {
	t.preise, t.reihe = preise, reihe

	return t
}

// Abgleichen sagt, wer was von wem bekommen könnte.
//
// **`behalten` ist die Zahl, die bleiben soll** — dieselbe Angabe wie bei
// [Sammlungssichten.Dubletten]: eine Dublette ist kein Zustand, sondern ein
// Verhältnis, und die Zahl bringt der Rufende mit (0057).
func (t *Tauschwerk) Abgleichen(
	ctx context.Context, fremd Fremdstand, behalten int, umfang domain.Tauschumfang,
) (Tauschabgleich, error) {
	if fehler := ctx.Err(); fehler != nil {
		return Tauschabgleich{}, fehler
	}

	verzeichnis, fehler := t.verzeichnisLesen(ctx)
	if fehler != nil {
		return Tauschabgleich{}, fehler
	}

	preise, fehler := t.preiseLesen(ctx)
	if fehler != nil {
		return Tauschabgleich{}, fehler
	}

	meinAngebot, fehler := t.eigenesAngebot(ctx, behalten, verzeichnis, preise)
	if fehler != nil {
		return Tauschabgleich{}, fehler
	}

	meineWuensche, fehler := t.eigeneWuensche(ctx, verzeichnis)
	if fehler != nil {
		return Tauschabgleich{}, fehler
	}

	if fehler := ctx.Err(); fehler != nil {
		return Tauschabgleich{}, fehler
	}

	gebe, gebeBefunde := domain.Tauschkandidaten(
		meinAngebot, alsWunschzeilen(fremd.Wuensche, verzeichnis))

	bekomme, bekommeBefunde := domain.Tauschkandidaten(
		bepreist(alsAngebotszeilen(fremd.Angebot, verzeichnis), preise), meineWuensche)

	abgleich := Tauschabgleich{
		Kennung: fremd.Kennung, Gebe: gebe, Bekomme: bekomme,
		Befunde: append(gebeBefunde, bekommeBefunde...),
	}

	abgleich.Vorschlag, abgleich.HatVorschlag = domain.Tauschvorschlagen(gebe, bekomme, umfang)

	return abgleich, nil
}

// eigenesAngebot macht aus den eigenen Dubletten Angebotszeilen.
func (t *Tauschwerk) eigenesAngebot(
	ctx context.Context, behalten int, verzeichnis map[string]string,
	preise map[string]domain.Preisstand,
) ([]domain.Angebotszeile, error) {
	dubletten, fehler := t.sichten.Dubletten(ctx, behalten)
	if fehler != nil {
		return nil, fehler
	}

	zeilen := make([]domain.Angebotszeile, 0, len(dubletten))

	for _, dublette := range dubletten {
		zeile := alsAngebotszeile(dublette.Exemplar, verzeichnis)
		zeile.Menge = dublette.Ueberzaehlig
		zeilen = append(zeilen, zeile)
	}

	return bepreist(zeilen, preise), nil
}

// eigeneWuensche macht aus den eigenen Wünschen Wunschzeilen.
func (t *Tauschwerk) eigeneWuensche(
	ctx context.Context, verzeichnis map[string]string,
) ([]domain.Wunschzeile, error) {
	wuensche, fehler := t.sichten.Wuensche(ctx)
	if fehler != nil {
		return nil, fehler
	}

	zeilen := make([]domain.Wunschzeile, 0, len(wuensche))
	for _, wunsch := range wuensche {
		zeilen = append(zeilen, alsWunschzeile(wunsch, verzeichnis))
	}

	return zeilen, nil
}
