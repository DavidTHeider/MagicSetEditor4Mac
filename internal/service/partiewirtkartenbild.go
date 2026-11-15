package service

// **Der Wirt beantwortet einen Bildwunsch** — P9
// (#2377),
// Befund C-12, Vermerk 0198.
//
// # Was er zeichnet
//
// **Ein Objekt, das der Gast sehen darf, aus einem Deck dieses Bestands.**
// Ob er es sehen darf, sagt die Sicht, die der Wirt **für ihn** rechnet
// (`MitBestandssicht`); mit welcher Zeile gezeichnet wird, die Sicht in
// diesem Bestand. Die Zeile verlässt den Wirt nie — hinaus geht ein Bild.
//
// # Was er nicht zeichnet
//
// **Ein geheimes Objekt, eines aus einem fremden Bestand, eines, das es nicht
// gibt** — dafür steht ein Grund in der Antwort, damit der Gast nicht wartet.
// Und nichts, das größer wäre, als die Leitung trägt.

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// Kartenbild ist ein gezeichnetes Bild samt Maßen.
type Kartenbild struct {
	Png        []byte
	Breite     int
	Hoehe      int
	Zweiseitig bool
}

// Kartenbildmaler zeichnet eine Kartenseite aus einer Zeile dieses Bestands.
type Kartenbildmaler interface {
	Kartenbild(ctx context.Context, karte int64, seite, guete string) (
		Kartenbild, error)
}

// bildreserve lässt Platz für den Umschlag um das Bild — Base64 macht aus
// drei Bytes vier.
const bildreserve = 4096

// MitMaler setzt, womit der Wirt Bilder für seinen Gast zeichnet.
func (w *Partiewirt) MitMaler(maler Kartenbildmaler) *Partiewirt {
	w.maler = maler

	return w
}

// bildAntworten zeichnet, was erfragt ist — oder sagt, warum nicht.
//
// **Was ein anderer Gast zeichnet, wartet nicht in dieser Schleife**
// (Zweitblick über Phase 7, Befund 5): fragten A nach einer Karte von B und
// B nach einer von A zugleich, warteten beide Wirtsschleifen — und jede
// Antwort muss durch die andere. Das Weiterreichen läuft nebenher und gibt
// seine Antwort selbst hinaus — mit eigener Frist; scheitert das, legt es die
// Leitung nieder, und der Fragende liest nach seiner Frist einen Grund.
func (w *Partiewirt) bildAntworten(ctx context.Context, roh []byte) error {
	var wunsch domain.Bildwunsch
	if fehler := json.Unmarshal(roh, &wunsch); fehler != nil {
		return fmt.Errorf("%w: der Bildwunsch ist unlesbar: %s",
			domain.ErrValidation, fehler.Error())
	}

	if fehler := wunsch.Pruefe(); fehler != nil {
		return fehler
	}

	antwort, weiter := w.bildFuer(ctx, wunsch)
	if weiter != "" {
		go w.bildWeiterreichen(ctx, weiter, wunsch)

		return nil
	}

	return w.bildHinaus(ctx, antwort)
}

// bildHinaus legt eine Bildantwort in die Schlange.
func (w *Partiewirt) bildHinaus(ctx context.Context, antwort domain.Bildantwort) error {
	inhalt, _ := domain.VerpackeKontrolle( //nolint:errcheck // Bytes, Zahlen und Zeichenketten klagen nicht
		domain.KontrolleBild, antwort)

	return w.sitzung.Sende(ctx, domain.ArtKontrolle, inhalt)
}

// bildFuer gibt das Bild oder den Grund — oder den Sitz, dessen Gast es
// zeichnet (`weiter`).
func (w *Partiewirt) bildFuer(
	ctx context.Context, wunsch domain.Bildwunsch,
) (antwort domain.Bildantwort, weiter string) {
	nein := func(grund string) domain.Bildantwort {
		return domain.Bildantwort{Wunsch: wunsch, Grund: grund}
	}

	zeile, eigner, grund := w.zeileFuer(ctx, wunsch)

	switch {
	case grund != "":
		return nein(grund), ""
	case zeile == 0:
		if grund := w.weiterreichbar(ctx, eigner, wunsch); grund != "" {
			return nein(grund), ""
		}

		return domain.Bildantwort{}, eigner
	case w.maler == nil:
		return nein("dieser Tisch zeichnet keine Bilder für Gäste"), ""
	}

	return gemalt(ctx, w.maler, zeile, wunsch), ""
}

// weiterreichfrist ist, wie lange der Wirt auf das Bild eines anderen Gastes
// wartet (P9b) — **kürzer als die Frist des Fragenden**, damit der einen Grund
// liest und nicht selbst aufgibt.
const weiterreichfrist = 8 * time.Second

// weiterreichbar sagt, warum der Gast am Sitz `eigner` nicht gefragt wird —
// leer, wenn er es wird (P9b, #2377).
//
// **Nie den Fragenden selbst**: seine Antwort käme über die Schleife, die
// gerade hier liest — er wartete auf sich, bis die Frist ablief. **Und nur
// einen, der das Bild aus seinem Bestand zeichnet** (Befund 6).
func (w *Partiewirt) weiterreichbar(
	ctx context.Context, eigner string, wunsch domain.Bildwunsch,
) string {
	if eigner == w.gast || w.runde == nil {
		return "dieses Bild zeichnet nicht der Wirt — das Deck stammt " +
			"aus einem anderen Bestand"
	}

	if !w.runde.zeichnetSelbst(ctx, w.nehmer, w.partie, eigner, wunsch.Objekt) {
		return drittquelle
	}

	return ""
}

// bildWeiterreichen fragt den Gast, dem das Deck gehört, und gibt seine Antwort
// hinaus — **nebenher**, siehe [Partiewirt.bildAntworten].
func (w *Partiewirt) bildWeiterreichen(
	ctx context.Context, eigner string, wunsch domain.Bildwunsch,
) {
	frist, fertig := context.WithTimeout(ctx, weiterreichfrist)
	defer fertig()

	antwort, fehler := w.runde.Bild(frist, eigner, wunsch)
	if fehler != nil {
		antwort = domain.Bildantwort{Grund: fehler.Error()}
	}

	antwort.Wunsch = wunsch

	// **Auch der Versand hat eine Frist** (Nachlauf, R3) — ein Gast, der
	// nicht mehr liest und die Verbindung offen hält, hielte sonst diesen
	// Lauf und jeden anderen, der leeren will. **Und ein Fehler legt die
	// Leitung nieder**: zurückgeben kann ihn hier niemand.
	versand, aus := context.WithTimeout(ctx, versandfrist)
	defer aus()

	fehler = w.bildHinaus(versand, antwort)
	if fehler == nil {
		fehler = w.sitzung.Leere(versand)
	}

	if fehler != nil {
		_ = w.sitzung.Schliesse() //nolint:errcheck // die Leitung ist ohnehin verloren
	}
}

// versandfrist ist, wie lange das Hinausgeben einer weitergereichten Antwort
// dauern darf.
const versandfrist = 5 * time.Second

// gemalt zeichnet eine Zeile des eigenen Bestands für die Leitung — oder sagt,
// warum nicht. **Derselbe Schritt für Wirt und Gast** (P9b): wer ein Bild
// hinausgibt, gibt nur eines, das die Leitung trägt.
func gemalt(
	ctx context.Context, maler Kartenbildmaler, zeile int64,
	wunsch domain.Bildwunsch,
) domain.Bildantwort {
	bild, fehler := maler.Kartenbild(ctx, zeile, wunsch.Seite, wunsch.Guete)
	if fehler != nil {
		return domain.Bildantwort{
			Wunsch: wunsch,
			Grund:  "das Bild ließ sich nicht zeichnen: " + fehler.Error(),
		}
	}

	if len(bild.Png)*4/3+bildreserve > peer.Hoechstlaenge {
		return domain.Bildantwort{
			Wunsch: wunsch,
			Grund:  "das Bild ist größer, als die Leitung trägt",
		}
	}

	return domain.Bildantwort{
		Wunsch: wunsch, Png: bild.Png, Breite: bild.Breite, Hoehe: bild.Hoehe,
		Zweiseitig: bild.Zweiseitig,
	}
}

// zeileFuer gibt die Zeile, mit der ein Objekt gezeichnet wird — oder den
// Grund, aus dem der Gast kein Bild bekommt.
//
// **Null ohne Grund heißt: ein anderer Bestand** — dann nennt `eigner` den
// Sitz, dem das Deck gehört.
func (w *Partiewirt) zeileFuer(
	ctx context.Context, wunsch domain.Bildwunsch,
) (zeile int64, eigner, grund string) {
	objekt := wunsch.Objekt

	fuerGast, fehler := w.nehmer.Partie(MitBestandssicht(ctx, w.gegenueber), w.partie, w.gast)
	if fehler != nil {
		return 0, "", "der Tisch ist gerade nicht zu lesen"
	}

	gesehen, da := sichtobjekt(fuerGast, objekt)
	switch {
	case !da:
		return 0, "", "dieses Objekt liegt nicht auf dem Tisch"
	case gesehen.Geheim || gesehen.Bild == "":
		return 0, "", "dieses Objekt ist für dich nicht zu sehen"
	case !bildPasst(gesehen, wunsch):
		return 0, "", "diese Bildkennung gehört nicht zu diesem Objekt"
	}

	hier, fehler := w.nehmer.Partie(ctx, w.partie, w.gast)
	if fehler != nil {
		return 0, "", "der Tisch ist gerade nicht zu lesen"
	}

	beimWirt, _ := sichtobjekt(hier, objekt)

	return beimWirt.KarteID, beimWirt.Eigner, ""
}

// bildPasst sagt, ob die Bildkennung eines Wunsches die des Objekts ist —
// Nachlauf zum Zweitblick über Phase 7, R2. **Ohne Kennung prüft niemand**:
// der Fragende legt dann unter keiner ab.
func bildPasst(gesehen domain.Sichtobjekt, wunsch domain.Bildwunsch) bool {
	return wunsch.Bild == "" || wunsch.Bild == gesehen.Bild
}

// sichtobjekt findet ein Objekt in einer Sicht.
func sichtobjekt(sicht domain.Partiesicht, kennung string) (domain.Sichtobjekt, bool) {
	for _, objekt := range sicht.Objekte {
		if objekt.Kennung == kennung {
			return objekt, true
		}
	}

	return domain.Sichtobjekt{}, false
}
