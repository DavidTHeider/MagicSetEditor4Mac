package service

// Der **Umgebungsbauer** — die setweite Hälfte der Skriptumgebung einmal statt
// je Karte (#1304).

import (
	"context"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// umgebungsbauer hält, was für **alle** Karten eines Sets gleich ist.
//
// # Was gemessen war
//
// `Packarten` über `Vanguard.mse-set` (107 Karten, Stil `m15-altered`, volles
// Pack) brauchte am 2026-08-29 **4.051 ms** in der Umgebung und **1 ms** in
// allen zwölf Filtern zusammen. Zerlegt:
//
//	Initskripte   1905 ms
//	Wertskripte   1823 ms
//	card_style     157 ms
//	mitVorgaben     36 ms
//	belegeVorlage    2 ms
//	Bibliothek       1 ms
//
// **Die Initskripte sind der eine Posten, der nicht je Karte anfällt.** Sie
// erklären Namen — `mana_sort`, `module_card`, `template` —, und ein Name
// bekommt seine Umgebung beim **Aufruf** gereicht, nicht bei der Erklärung
// (`msescript.Umgebung`, „eine geschweifte Klammer schließt nichts ein").
//
// # Und so macht MSE2 es auch
//
// `SetScriptContext::getContext` hält den Kontext je Stilbogen in einer
// Abbildung, lässt `game->init_script` und `stylesheet->init_script` beim
// **ersten** Zugriff laufen und setzt je Karte nur `card` und `styling`
// (`src/script/script_manager.cpp`, Zeilen 30–75). Die Attrappe für den
// ersten Lauf ist dort `set.cards.front()` — hier ist es dieselbe Karte, mit
// der der Bauer entsteht.
//
// Fassung f9d9356 vom 2021-07-06.
//
// # Was das kostet
//
// **Für eine einzelne Karte kostet er nichts.** Der erste Anlauf ließ die
// Vorbereitung zweimal laufen — einmal für die Attrappe, einmal für die Karte
// —, und der Kanonlauf über 680 Karten wurde davon messbar langsamer.
// [umgebungsbauer.fuerKarte] erkennt die Attrappe und nimmt, was schon
// dasteht. Wer hundert Karten filtert, spart 99 Läufe der Initskripte; wer
// eine zeichnet, zahlt dasselbe wie vorher.
//
// **Ein zweiter Weg für den Sonderfall wäre zwei Wege** — und zwei Wege durch
// dieselbe teuer erkaufte Reihenfolge (#240, #270, #298, #345) wären der
// teurere Preis.
type umgebungsbauer struct {
	vorlage   Vorlage
	schriften render.Schriftquelle
	bilder    render.Bildquelle

	// setweit trägt die Bibliothek, die Vorlage und die Namen der
	// Initskripte. Je Karte entsteht darüber eine eigene Ebene.
	setweit *msescript.Umgebung

	// initbefunde sind die Befunde der beiden Initskripte — einmal erhoben
	// und je Karte gemeldet, weil sie je Karte gälten.
	initbefunde []domain.Befund

	// attrappe ist die Karte, mit der die Initskripte gelaufen sind, und
	// attrappenfelder ihre vorbereiteten Felder.
	//
	// **Für sie ist die Vorbereitung schon getan.** Ohne diese beiden Zeilen
	// zahlte der häufigste Fall — **eine** Karte zeichnen — die Vorbereitung
	// zweimal, und der Rasterer wurde langsamer, statt gleich zu bleiben.
	attrappe        *domain.CardRow
	attrappenfelder []msescript.Eintrag

	// vermerk ist der Schlüsselwortvermerk dieses Bauers (#1354).
	//
	// **Einer für alle Karten, geleert vor jeder.** Er hängt an der Umgebung
	// des Spiels, weil `text_filter` die Funktion beim Initialisieren einfriert
	// (Befund 5 des Zweitblicks über Phase 4.11); Karten werden nacheinander
	// gezeichnet.
	vermerk *Schluesselwortverwendung
}

// neuerUmgebungsbauer lässt die Initskripte **einmal** laufen.
func neuerUmgebungsbauer(
	ctx context.Context, vorlage Vorlage,
	schriften render.Schriftquelle, bilder render.Bildquelle,
	attrappe *domain.CardRow,
) *umgebungsbauer {
	setweit := msescript.Bibliothek().Darueber()

	// **Die Schlüsselwortnaht steht vor dem Spielskript** — sonst friert
	// `text_filter` den Stummel ein, und keine Karte bekommt einen
	// Erinnerungstext (Befund 5 des Zweitblicks über Phase 4.11).
	vermerk := &Schluesselwortverwendung{}
	belegeSchluesselwortnaht(setweit, vorlage.Schluesselwoerter, vermerk)

	felder := kartenumgebungVorbereiten(
		ctx, setweit, attrappe, vorlage, schriften, bilder)

	// **Das Spiel vor dem Stil.** Ein Stil darf eine Funktion des Spiels
	// überschreiben — das ist der Sinn einer Vorlage; umgekehrt machte das
	// Spiel jede Anpassung eines Stils wirkungslos.
	befunde := vorbelegt(ctx, vorlage.Spielskript, "init script des Spiels", setweit)
	befunde = append(befunde,
		vorbelegt(ctx, vorlage.Stil.Initskript, "init script", setweit)...)

	return &umgebungsbauer{
		vorlage: vorlage, schriften: schriften, bilder: bilder,
		setweit: setweit, initbefunde: befunde,
		attrappe: attrappe, attrappenfelder: felder,
		vermerk: vermerk,
	}
}

// fuerKarte baut die Umgebung einer Karte über der setweiten.
//
// **Eine eigene Ebene je Karte und nicht dieselbe.** `Setze` legt einen Namen
// auf **dieser** Ebene an; würde jede Karte auf der setweiten schreiben, sähe
// die nächste, was die vorige angelegt hat. Über eine Ebene darüber schattet
// sie `card` und `styling`, und die Namen der Initskripte bleiben erreichbar.
func (b *umgebungsbauer) fuerKarte(ctx context.Context, karte *domain.CardRow) (
	*msescript.Umgebung, []msescript.Eintrag, []domain.Befund,
) {
	umgebung := b.setweit.Darueber()

	// **Für die Attrappe steht die Vorbereitung schon da.** Sie ist genau die
	// Karte, mit der die Initskripte gelaufen sind; ihr `card`, `card_style`
	// und `styling` stehen eine Ebene tiefer und werden von hier aus gesehen.
	// Ein zweiter Durchgang gäbe dieselben Werte und kostete den Rasterer
	// einen Aufschlag, den es nicht geben muss.
	felder := b.attrappenfelder

	// **Der Vermerk gehört zur Karte, auch wenn er am Spiel hängt** (Befund 5
	// des Zweitblicks über Phase 4.11).
	b.vermerk.Leere()

	if karte != b.attrappe {
		felder = kartenumgebungVorbereiten(
			ctx, umgebung, karte, b.vorlage, b.schriften, b.bilder)
	}

	felder, befunde := nachbereiten(
		ctx, umgebung, felder, b.vorlage, b.schriften, b.bilder, b.vermerk)

	// **Die Befunde der Initskripte reisen je Karte mit.** Sie sind einmal
	// erhoben, gelten aber für jede Karte, die mit dieser Vorlage gezeichnet
	// wird — wer sie nur an der ersten meldete, verlöre sie beim Blättern.
	return umgebung, felder, append(slices.Clone(b.initbefunde), befunde...)
}
