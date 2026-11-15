package service

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Vermerk sagt, was über die Pakete an einem Ort bekannt ist.
//
// **Über einen Ort und nicht über einen Pfad je Paket**: ein Bezug holt ein
// ganzes Repositorium, und seine Prüfsumme gilt für alles darin (#56). Beim
// Installieren ist der Ort das Paket selbst — derselbe Vergleich, ein Sonderfall
// weniger.
type Vermerk struct {
	Ort        string `json:"-"`
	Herkunft   string `json:"herkunft"`
	Pruefsumme string `json:"pruefsumme"`

	// Abdruck ist ein festgehaltener Stand — gesetzt nur, wenn eine vorige
	// Fassung zurückkommt (DOM-01.AK-02). Leer heißt: jetzt rechnen.
	Abdruck string `json:"abdruck"`
}

// lesen geht die Wurzeln ab und liest jedes Paket, das darin liegt.
//
// `vermerke` sagt, was über einen Ort bekannt ist — der Fall „gerade bezogen"
// oder „gerade installiert". Für alles andere bleibt es leer, und der Abgleich
// behält, was der Katalog schon weiß (siehe [storage.Paketkatalog.Abgleichen]).
//
// **Nur die oberste Ebene je Wurzel.** Die Wurzeln kommen aus
// [msepaket.Paketwurzeln] und liegen bereits **unmittelbar** über den Paketen;
// noch einmal zu suchen hieße, dieselbe Arbeit zweimal zu tun — und der
// Composition Root bliebe dabei ungeprüft, weil dieser Verwalter sie
// stillschweigend nachholte (der Fall steht bei [NeuerStilspeicher]).
func (v *Paketverwalter) lesen(vermerke []Vermerk) ([]domain.Paket, []domain.Befund) {
	var (
		pakete  []domain.Paket
		befunde []domain.Befund
	)

	gesehen := make(map[string]bool)

	for _, wurzel := range v.abzugehende() {
		if strings.TrimSpace(wurzel) == "" {
			continue
		}

		eintraege, err := os.ReadDir(wurzel)
		if err != nil {
			// **Ein unlesbares Verzeichnis ist ein Befund.** Es kann ein
			// entferntes Laufwerk sein, ein fehlendes Leserecht — und die
			// Pakete daneben sollen deswegen nicht verlorengehen.
			befunde = append(befunde, domain.Befund{
				Datei:   wurzel,
				Meldung: "das Paketverzeichnis ließ sich nicht lesen: " + err.Error(),
			})

			continue
		}

		for _, eintrag := range eintraege {
			paket, befund := lies(filepath.Join(wurzel, eintrag.Name()), eintrag.Name(), vermerke)

			switch {
			case befund != nil:
				befunde = append(befunde, *befund)

			case paket != nil && !gesehen[paket.Art+"/"+paket.Name]:
				// **Das erste Vorkommen gewinnt**, und die Reihenfolge der
				// Wurzeln ist die Aussage: wer ein Paket bezogen hat, meint das
				// bezogene; die mitgelieferte Saat kommt danach (#58). Zwei
				// Zeilen mit demselben (art, name) ließe der Katalog ohnehin
				// nicht zu.
				gesehen[paket.Art+"/"+paket.Name] = true

				pakete = append(pakete, *paket)
			}
		}
	}

	return pakete, befunde
}

// lies liest ein einzelnes Paket; nil und nil heißt „kein Paket".
func lies(pfad, name string, vermerke []Vermerk) (*domain.Paket, *domain.Befund) {
	art := msepaket.ArtAusName(name)
	if art == msepaket.ArtUnbekannt {
		return nil, nil
	}

	// **Sets zählen nicht zum Katalog.** Ein `.mse-set` ist ein Dokument der
	// Nutzenden, kein installiertes Vorlagenpaket; der Bestand führt es in
	// `sets`. Beides in einer Liste zu zeigen wäre die Vermischung, die die
	// Setliste unbrauchbar machte.
	if art == msepaket.ArtSet {
		return nil, nil
	}

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		return nil, &domain.Befund{
			Datei:   name,
			Meldung: "das Paket ließ sich nicht öffnen: " + err.Error(),
		}
	}

	defer func() { _ = paket.Schliessen() }() //nolint:errcheck // gelesen ist gelesen

	gelesen := domain.Paket{
		Art:     string(art),
		Name:    strings.TrimSuffix(name, "."+string(art)),
		Version: version(paket.Version),
		Pfad:    pfad,

		// **Der Kurzname für die Paketwahl** (#632). Er steht im Kopf als
		// `short name:`; leer heißt „keiner angegeben", und der Rückfall auf
		// den Dateinamen geschieht dort, wo die Liste gebaut wird.
		Kurzname: paket.KurzerName,
		Braucht:  bedarf(paket.Abhaengigkeiten),

		// **Einordnung, wie sie in der Datei steht** (#244). Sie wird nicht
		// erschlossen: `installer group:` und `position hint:` sind Angaben
		// der Paketautoren, und MSE2 baut daraus seinen Paketbaum.
		Gruppe: paket.Gruppe,
		Rang:   paket.Rang,
	}

	for _, vermerk := range vermerke {
		if imVerzeichnis(pfad, vermerk.Ort) {
			gelesen.Herkunft = vermerk.Herkunft
			gelesen.Pruefsumme = vermerk.Pruefsumme

			// **Der Abdruck entsteht mit dem Vermerk** — beim Einbinden und
			// Beziehen, also in dem Augenblick, der als Stand gilt
			// (DOM-01.AK-01). Lässt er sich nicht rechnen, bleibt er leer:
			// ohne Stand gibt es nichts, wovon das Paket abweichen könnte.
			gelesen.Abdruck = vermerk.Abdruck
			if gelesen.Abdruck == "" {
				gelesen.Abdruck, _ = abdruckVon(pfad) //nolint:errcheck // siehe darüber
			}
		}
	}

	return &gelesen, nil
}

// version gibt eine Paketversion als Zeichenkette — leer, wenn keine dasteht.
func version(datum msepaket.Datum) string {
	if datum.Leer() {
		return ""
	}

	return datum.String()
}

// bedarf übersetzt die Abhängigkeiten eines Pakets in die Domäne.
func bedarf(abhaengigkeiten []msepaket.Abhaengigkeit) []domain.Bedarf {
	if len(abhaengigkeiten) == 0 {
		return nil
	}

	braucht := make([]domain.Bedarf, 0, len(abhaengigkeiten))

	for _, abhaengigkeit := range abhaengigkeiten {
		braucht = append(braucht, domain.Bedarf{
			Paket:   abhaengigkeit.Paket,
			Version: version(abhaengigkeit.Version),
		})
	}

	return braucht
}

// paketLesen öffnet ein Paket und wirft es gleich wieder weg.
//
// **Vor dem Kopieren**, damit ein unlesbares Paket gar nicht erst im
// Verwaltungsverzeichnis landet.
func (v *Paketverwalter) paketLesen(pfad string) (*domain.Paket, error) {
	paket, befund := lies(pfad, filepath.Base(filepath.Clean(pfad)), nil)
	if befund != nil {
		return nil, domain.NewFieldError("quelle", domain.SchluesselValidierung, "%s", befund.Meldung)
	}

	if paket == nil {
		return nil, domain.NewFieldError("quelle", domain.SchluesselValidierung,
			"%q ist kein Vorlagenpaket.", pfad)
	}

	return paket, nil
}

// kopierePaket legt ein Paketverzeichnis an einem zweiten Ort ab.
//
// **Kopieren und nicht verschieben.** Die Quelle gehört jemand anderem: eine
// MSE2-Installation, ein Download, ein Verzeichnis, das die Nutzenden pflegen.
// Wer sie verschöbe, nähme sie ihnen weg.
//
// **Über `os.CopyFS` und nicht über einen eigenen Baumläufer.** Der erste
// Anlauf schrieb beides selbst — Verzeichnis wie Datei — und hatte damit sechs
// Fehlerwege, von denen zwei (ein scheiterndes `io.Copy`, ein scheiterndes
// `Close`) sich nicht nachstellen lassen. Ein Fehlerpfad, den man beseitigen
// kann, ist besser als einer, den man nachstellt (Phase 2, Punkt 1).
//
// Dass die Quelle ein Verzeichnis ist, steht beim Aufrufer fest — dort ist der
// Fehlerweg erreichbar, hier wäre er es nicht.
//
// # Erst ins Leere, dann an den Platz
//
// **Dieselbe Regel wie beim Bezug** ([Paketbezieher.Beziehen]), und sie stand
// hier bis zum Nachlauf über Phase G nicht: `os.CopyFS` legte das
// Zielverzeichnis **zuerst** an und schrieb die Dateien danach hinein.
//
// Damit änderte sich die Änderungszeit der Paketwurzel beim **Anlegen** — und
// `internal/schriften` bindet seinen Namensindex an genau diese Zeit. Eine
// Schriftsuche im Fenster dazwischen merkte sich einen halben Bestand und
// behielt ihn, bis die Wurzel sich wieder änderte. Über das Umbenennen ändert
// sie sich **zuletzt**, also dann, wenn das Paket ganz dasteht.
//
// Das zweite, das dabei abfällt: ein Abbruch auf halber Strecke hinterlässt
// kein halbes Paket unter dem Namen eines ganzen — auch dann nicht, wenn der
// Prozess dabei stirbt und niemand mehr aufräumen kann.
func kopierePaket(quelle, ziel string) error {
	vorlaeufig, err := os.MkdirTemp(filepath.Dir(ziel), "."+filepath.Base(ziel)+"-")
	if err != nil {
		return fmt.Errorf("%q nach %q kopieren: %w", quelle, ziel, err)
	}

	defer func() { _ = os.RemoveAll(vorlaeufig) }() //nolint:errcheck // nach einem geglückten Umbenennen gibt es ihn nicht mehr

	gebaut := filepath.Join(vorlaeufig, filepath.Base(ziel))

	if fehler := os.CopyFS(gebaut, os.DirFS(quelle)); fehler != nil {
		return fmt.Errorf("%q nach %q kopieren: %w", quelle, ziel, fehler)
	}

	if fehler := os.Rename(gebaut, ziel); fehler != nil {
		return fmt.Errorf("%q an seinen Platz bringen: %w", quelle, fehler)
	}

	return nil
}

// abzugehende sind die Wurzeln **und** das verwaltete Verzeichnis.
//
// # Warum das verwaltete Verzeichnis extra dazugehört
//
// Die Wurzeln kommen aus [msepaket.Paketwurzeln], und das meldet ein
// Verzeichnis nur, wenn **schon ein Paket darin liegt**. Auf einem frischen
// Rechner ist das verwaltete Verzeichnis leer und damit keine Wurzel — und
// dann las dieser Verwalter das, wohin er gerade kopiert hatte, **nicht**.
//
// **Was das gekostet hat** (#1291):
// `Installieren` und `Uebernehmen` gaben eine leere Liste zurück und meldeten
// Erfolg. Das Paket lag da; gesehen hat es erst der nächste Prozess, weil das
// Verzeichnis dann nicht mehr leer war. Gemessen an `mse mse2-uebernehmen`.
//
// **Und nicht in `Paketwurzeln` behoben**: dort ist „ein Verzeichnis ohne
// Paket ist keine Wurzel" eine ausdrückliche Zusage mit Grund — sonst liefe
// `Bildquelle` für jedes fehlende Bild durch jeden leeren Ordner. Wer
// **verwaltet**, liest sein Verzeichnis; wer **sucht**, sucht, wo etwas liegt.
//
// **Unter dem Paketordner wird bei jedem Lesen neu gesucht** (G04, #2251).
// Die Wurzeln standen fest, seit der Verwalter gebaut war, und ein Bezug legt
// seinen Baum eine Ebene tiefer, als im Paketordner selbst gelesen wird: der
// Vermerk nach dem Bezug fand die neuen Pakete nicht, und ihre Herkunft kam
// nie in den Katalog. **Die Reihenfolge bleibt die Aussage** — was bezogen
// ist, vor dem, was mitkommt.
func (v *Paketverwalter) abzugehende() []string {
	if v.verwaltet == "" {
		return v.wurzeln
	}

	wurzeln := []string{v.verwaltet}

	for _, wurzel := range slices.Concat(msepaket.Paketwurzeln(v.verwaltet), v.wurzeln) {
		if !slices.Contains(wurzeln, wurzel) {
			wurzeln = append(wurzeln, wurzel)
		}
	}

	return wurzeln
}
