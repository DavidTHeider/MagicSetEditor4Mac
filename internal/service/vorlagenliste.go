package service

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Vorlagenliste findet die Exportvorlagen, die zu einem Set passen.
//
// # Die Naht aus #209
//
// `Vorlagenausfuehrer` stand seit #51 und wurde außerhalb von Tests **nirgends**
// gerufen: alles, was er kann, war von der Oberfläche aus unerreichbar
// (#209).
// Zwischen ihm und dem Fenster fehlte die Frage „welche Vorlagen gibt es hier?".
//
// # Warum nach Spiel gefiltert wird
//
// Eine Exportvorlage nennt im Kopf ihr `game:`. Eine Vorlage für `magic` auf ein
// Set eines anderen Spiels loszulassen, ergäbe Skriptfehler über Feldnamen, die
// es dort nicht gibt — und der Anwendende bekäme eine Fehlerliste statt einer
// Auskunft, dass die Vorlage nicht passt.
type Vorlagenliste struct {
	koepfe  Kopfquelle
	wurzeln []string
}

// NeueVorlagenliste baut die Liste über einer Reihe von Paketwurzeln.
//
// **Die Wurzeln liegen unmittelbar über den Paketen** — derselbe Vertrag wie
// bei [msepaket.NeueBildquelle] und [NeuerStilspeicher].
func NeueVorlagenliste(koepfe Kopfquelle, wurzeln ...string) *Vorlagenliste {
	return &Vorlagenliste{koepfe: koepfe, wurzeln: wurzeln}
}

// Vorlagen gibt die Exportvorlagen, die zum Spiel eines Sets gehören.
//
// **Nicht gemerkt**, und das ist Absicht: die Liste entsteht, wenn jemand das
// Menü öffnet — also selten. Ein Zwischenspeicher machte einen frisch bezogenen
// Satz Vorlagen unsichtbar, ohne etwas zu sparen, das jemand merkt.
func (v *Vorlagenliste) Vorlagen(ctx context.Context, setID int64) (
	[]msepaket.Exportvorlage, []domain.Befund, error,
) {
	koepfe, err := v.koepfe.Setkoepfe(ctx)
	if err != nil {
		return nil, nil, err
	}

	kopf, da := kopfMit(koepfe, setID)
	if !da {
		return nil, befundliste("das Set %d führt der Bestand nicht", setID), nil
	}

	gefunden, befunde := v.ausWurzeln(kopf.Game)

	sort.Slice(gefunden, func(a, b int) bool {
		return gefunden[a].VollerName < gefunden[b].VollerName
	})

	return gefunden, befunde, nil
}

// ausWurzeln liest jede Exportvorlage unter den Wurzeln.
func (v *Vorlagenliste) ausWurzeln(spiel string) ([]msepaket.Exportvorlage, []domain.Befund) {
	var (
		gefunden []msepaket.Exportvorlage
		befunde  []domain.Befund
		gesehen  = map[string]bool{}

		// **Ein Spiel wird je Ruf einmal gelesen** (E03, #2240): im vollen
		// Pack gehören die meisten Vorlagen zu `magic`, und jede las
		// `magic.mse-game` samt Modulen neu — rund 350 MB je Ausfuhr.
		spiele = map[string]spielteil{}
	)

	for _, wurzel := range v.wurzeln {
		eintraege, err := os.ReadDir(wurzel)
		if err != nil {
			continue
		}

		for _, eintrag := range eintraege {
			if msepaket.ArtAusName(eintrag.Name()) != msepaket.ArtExportvorlage ||
				gesehen[eintrag.Name()] {
				continue
			}

			gesehen[eintrag.Name()] = true

			vorlage, befund := v.vorlageAus(filepath.Join(wurzel, eintrag.Name()), spiel, spiele)
			if befund != nil {
				befunde = append(befunde, *befund)

				continue
			}

			if vorlage != nil {
				gefunden = append(gefunden, *vorlage)
			}
		}
	}

	return gefunden, befunde
}

// vorlageAus liest eine Vorlage — nil heißt: sie gehört zu einem anderen Spiel.
//
// `spiele` merkt die Spielangaben, die dieser Ruf schon gelesen hat.
func (v *Vorlagenliste) vorlageAus(
	pfad, spiel string, spiele map[string]spielteil,
) (*msepaket.Exportvorlage, *domain.Befund) {
	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		return nil, &domain.Befund{
			Datei:   filepath.Base(pfad),
			Meldung: "die Exportvorlage ließ sich nicht öffnen: " + err.Error(),
		}
	}

	// **Ein Fehler beim Schließen wird verworfen, und das gehört gesagt.** Das
	// Paket ist gelesen; ein sperriges Archiv danach änderte am Ergebnis
	// nichts, und ein Befund darüber stünde in einer Liste, die von Vorlagen
	// handelt. Der Fall ist in `msepaket` geprüft, wo er hingehört.
	//
	//nolint:errcheck // der Grund steht darüber
	defer func() { _ = paket.Schliessen() }()

	// **Erst die Module, dann das Modell** (#216).
	//
	// Eine Exportvorlage darf `include file: script` schreiben, genau wie ein
	// Spiel- oder Stilpaket — `magic-lackeybot` und `magic-namecheck` beginnen
	// ihr Skript so. Hier stand der Griff nach `Exportmodell()` unmittelbar
	// hinter dem Öffnen; die Zeile blieb dann unaufgelöst stehen, und der
	// Zerteiler las `include` als Namen: „Zeile 1, Spalte 9: hier hört der
	// Ausdruck auf, aber der Text geht weiter".
	//
	// **Ein fehlendes Modul ist ein Befund, kein Abbruch** — dieselbe Antwort
	// wie in `stilwahl.go`: eine Vorlage, der ein Modul fehlt, scheitert
	// nachher an einem Namen, und *diese* Meldung nennt den Namen. Eine Vorlage
	// hier ganz zu verschweigen, nähme dem Anwendenden die Auskunft, dass es
	// sie gibt.
	_ = paket.ModuleEinbinden(msepaket.NeuerPaketleser(v.wurzeln...)) //nolint:errcheck // der Grund steht darüber

	vorlage := paket.Exportmodell()

	if !strings.EqualFold(vorlage.Spiel, spiel) {
		return nil, nil
	}

	// **Eine Vorlage ohne Skript ist keine.** Zwei Pakete des bezogenen Packs
	// tragen die Endung und kein `script:`; sie anzubieten hieße, einen Knopf
	// zu zeigen, der nur scheitern kann.
	if strings.TrimSpace(vorlage.Skript) == "" {
		return nil, &domain.Befund{
			Datei:   filepath.Base(pfad),
			Meldung: "die Exportvorlage führt kein Skript",
		}
	}

	// **Das Spielskript gehört dazu** (#216). Eine Vorlage ruft die Funktionen
	// ihres Spiels; ohne sie scheitert sie an einem unbekannten Namen.
	//
	// **Und die Kartenfelder ebenso** (#609): in MSE trägt jede Karte jedes
	// Feld ihres Spiels, und eine Vorlage, die ein ausgelassenes liest, bekommt
	// dort einen leeren Text statt eines Fehlers.
	teil, gelesen := spiele[vorlage.Spiel]
	if !gelesen {
		teil = v.spielangaben(vorlage.Spiel)
		spiele[vorlage.Spiel] = teil
	}

	vorlage.Spielskript, vorlage.Kartenfelder, vorlage.Setfelder,
		vorlage.Symbolfassungen = teil.skript, teil.kartenfelder, teil.setfelder, teil.fassungen

	return &vorlage, nil
}

// spielangaben holt `init script:` und Kartenfelder des Spielpakets.
//
// **Dieselbe Suche wie in `stilquelle.go`**, und bewusst nicht dieselbe
// Funktion: dort hängt sie an einem `Stilspeicher` mit eigenen Wurzeln. Was
// beide teilen, ist der Weg über `msepaket` — und der steht dort, wo er
// hingehört.
//
// **Ohne Sonderfall für den leeren Namen.** Ein `".mse-game"` lässt sich nicht
// öffnen, und die Schleife gibt dann dasselbe zurück wie eine ausdrückliche
// Prüfung — nur ohne einen Zweig, den kein Aufrufer je betritt.
func (v *Vorlagenliste) spielangaben(spiel string) spielteil {
	name := spiel + "." + string(msepaket.ArtSpiel)

	for _, wurzel := range v.wurzeln {
		paket, err := msepaket.Oeffne(filepath.Join(wurzel, name))
		if err != nil {
			continue
		}

		// **Ein fehlendes Modul ist hier kein Befund.** Was es kostet, zeigt
		// sich beim Ausführen als unbekannter Name — mit der Stelle, an der er
		// gebraucht wird. Ein zweiter Befund an dieser Stelle nennte dieselbe
		// Sache ohne die Stelle.
		_ = paket.ModuleEinbinden(msepaket.NeuerPaketleser(v.wurzeln...)) //nolint:errcheck // der Grund steht darüber
		gelesen := paket.Initskript()

		// **Die Befunde des Spielmodells bleiben hier liegen.** Ein Feld ohne
		// Namen meldet der Zeichenpfad; ihn hier ein zweites Mal zu melden
		// nennte dieselbe Sache an einer Stelle, an der niemand sie erwartet.
		spielmodell, _ := paket.Spielmodell()

		_ = paket.Schliessen() //nolint:errcheck // das Skript ist vollständig kopiert

		// **Und die Fassungen des Setsymbols** (#660). Sie stehen am
		// `set field:` des Spiels; beim Zeichnen holt sie der Stilspeicher,
		// hier die Vorlagenliste — dieselbe Auskunft, zwei Wege zu ihr, weil
		// beide Seiten eigene Wurzeln haben.
		fassungen := paket.Spielsymbolfassungen()

		// **Gekappt**, denn die Vorlagen eines Rufs teilen sie: ein `append`
		// an einer schriebe sonst in die Felder der anderen.
		return spielteil{
			skript: gelesen, kartenfelder: slices.Clip(spielmodell.Fields),
			setfelder: slices.Clip(spielmodell.SetFields), fassungen: slices.Clip(fassungen),
		}
	}

	// **Kein Befund, wenn es das Paket nicht gibt.** Eine Vorlage ohne
	// installiertes Spielpaket ist der Normalfall einer frischen Einrichtung;
	// wer sie ausführt, bekommt den unbekannten Namen genannt, und *der* sagt,
	// was fehlt.
	return spielteil{}
}

// spielteil ist, was eine Vorlage von ihrem Spielpaket braucht.
type spielteil struct {
	skript                  string
	kartenfelder, setfelder []domain.Field
	fassungen               []msepaket.Symbolfassung
}

// Vorlagendienst führt Liste und Ausführer zusammen.
//
// **Eine Naht und nicht zwei.** `internal/app` soll eine Frage stellen können —
// „welche Vorlagen gibt es, und führe diese aus" —, ohne zwei Teile zu kennen,
// die ohnehin nur zusammen einen Sinn ergeben.
type Vorlagendienst struct {
	*Vorlagenliste
	*Vorlagenausfuehrer
}

// NeuerVorlagendienst baut den Dienst über einer Reihe von Paketwurzeln.
func NeuerVorlagendienst(koepfe Kopfquelle, wurzeln ...string) *Vorlagendienst {
	return &Vorlagendienst{
		Vorlagenliste:      NeueVorlagenliste(koepfe, wurzeln...),
		Vorlagenausfuehrer: NeuerVorlagenausfuehrer(),
	}
}
