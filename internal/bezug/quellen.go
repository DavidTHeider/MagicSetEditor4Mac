package bezug

// Die bekannten Bezugsquellen — #263.
//
// # Wofür
//
// Der Paketverwalter kann seit Phase 7 beziehen; was fehlte, ist die letzte
// Meile: bis hierher musste jemand eine Adresse abtippen und einen
// vierzigstelligen Commit dazu. Diese Liste macht daraus einen Klick.
//
// # Was sie **nicht** ist
//
// **Kein Mitliefern.** Die Dateien landen beim Beziehen im Benutzerverzeichnis,
// nicht im Repository und nicht im `.dmg` (§11). Die Anwendung zeigt vor jedem
// Bezug, wessen Inhalte das sind, und lädt erst nach einer Zustimmung (#139).
//
// # Die Lizenzlage, nachgemessen
//
// Am 2026-08-19 über `gh api users/MagicSetEditorPacks/repos`: **zehn
// Repositorien, keines mit Lizenz.** Ohne Lizenz gilt das Urheberrecht in
// seiner Vollform. Deshalb wird hier nichts übernommen, nichts bearbeitet und
// nichts weitergegeben — es wird **verwiesen**, und geholt wird von der Quelle
// selbst. `DRITTANBIETER.md` führt jede Zeile dieser Liste samt dieser Lage;
// `internal/werkzeuge` hält beide gegeneinander.

// Quelle ist eine Bezugsquelle, die die Anwendung anbietet.
type Quelle struct {
	// Name ist der Name des Repositoriums und zugleich der Vorschlag für das
	// Verzeichnis, in das bezogen wird.
	Name string

	// Adresse und Stand sind, was [Repositorium] braucht.
	Adresse string

	// Stand ist ein **Commit**, kein Zweig (#56) — und einer, gegen den dieses
	// Projekt tatsächlich gelaufen ist, nicht der neueste. Ein Zweig ändert
	// sich, und dieselbe Angabe ergäbe morgen ein anderes Paket.
	Stand string

	// Unterverzeichnis ist die Paketwurzel im Archiv — **leer**, und das ist
	// der Kern von #263.
	//
	// Die Packs tragen ihre Vorlagen unter `data` und ihre **Schriften**
	// daneben (`Magic - Fonts`; im installierten `M15-Magic-Pack` 131 Dateien,
	// 26 MB gegen 779 MB Vorlagen). Wer nur `data` holt — so stand es bis
	// hierher im Formular —, bekommt Vorlagen ohne ihre Schriften, und jede
	// Karte zeigt einen Befund „Schrift fehlt", ohne dass jemand sähe, warum.
	//
	// [msepaket.Paketwurzeln] findet die Pakete auch eine Ebene tiefer, und
	// `internal/schriften` durchsucht die **ganze** Paketwurzel — beides ist
	// schon da; was fehlte, war das Herholen.
	Unterverzeichnis string

	// Schluessel benennt den erklärenden Satz in den Sprachkatalogen.
	//
	// **Ein Schlüssel und kein Satz** (§9): ein deutscher Fließtext von hier
	// stünde in der englischen Oberfläche deutsch. Ein Vertragstest hält
	// beide Kataloge gegen diese Liste.
	Schluessel string

	// Megabyte ist die Größe des **Repositoriums**, wie GitHub sie meldet.
	//
	// **Nicht die des Downloads** — den meldet `codeload` nicht: die Antwort
	// kommt ohne `Content-Length`, gemessen am 2026-08-19. Die Zahl ist damit
	// eine Größenordnung und keine Zusage; für `M15-Magic-Pack` meldet GitHub
	// 875 und auf der Platte liegen 805, also 9 % daneben.
	//
	// Sie steht trotzdem da, weil zwischen der kleinsten und der größten
	// Quelle der Faktor fünf liegt: wer ihn nicht sieht, startet einen Bezug
	// über Stunden, weil er den Namen für einen Hinweis hielt.
	Megabyte int
}

// quellen ist die Liste, geordnet von klein nach groß.
//
// **Nach Größe und nicht alphabetisch.** Die Reihenfolge ist die Empfehlung:
// wer zum ersten Mal hier steht, nimmt das kleinste, das seine Karten zeichnet.
// Alphabetisch stünde `Full-Magic-Pack` mit 1,8 GB an zweiter Stelle.
//
// **Vier und nicht zehn.** Von den zehn Repositorien des Kontos bleiben sechs
// draußen, und jedes aus einem genannten Grund:
//
//   - `Beta-Update-Pack` — „may not be ready for full packs yet", sagt es
//     selbst.
//   - `Installer-Pack` — Installationsdateien. `bezug` packt Ausführbares nicht
//     aus (#139); der Bezug ergäbe ein leeres Verzeichnis.
//   - `Outdated-Templates` — „kept on hand to help transition very old files".
//   - `Magic-Mini-Packs` — Einzelstücke für DownGit, keine Sammlung.
//   - `TypesLocalizer` — ein Skript, kein Paket.
//   - `Font-Pack` — **hier ist die Ablehnung begründungspflichtig**, denn #263
//     nennt die Schriften im Titel. Er wird nicht gebraucht: die vier Packs
//     bringen ihre Schriften selbst mit (`Magic - Fonts`), und genau die holt
//     das leere `Unterverzeichnis` mit. Ein zweiter Bezug für dieselben
//     Dateien wäre eine zweite Fassung derselben Schrift im selben Baum.
var quellen = []Quelle{
	{
		Name:       "Basic-M15-Magic-Pack",
		Adresse:    "https://github.com/MagicSetEditorPacks/Basic-M15-Magic-Pack",
		Stand:      "df1bf728440c3b3cfb0ec38a3d5b484d39c35cad",
		Schluessel: "basicM15",
		Megabyte:   316,
	},
	{
		Name:       "Full-Non-Magic-Pack",
		Adresse:    "https://github.com/MagicSetEditorPacks/Full-Non-Magic-Pack",
		Stand:      "84c6722f18e0838ff4a842ad13eefb71d9a39f24",
		Schluessel: "fullNonMagic",
		Megabyte:   575,
	},
	{
		Name:       "M15-Magic-Pack",
		Adresse:    "https://github.com/MagicSetEditorPacks/M15-Magic-Pack",
		Stand:      "3f7f8ceed0c1fcc0d193d752c646aaaea7bece85",
		Schluessel: "m15",
		Megabyte:   855,
	},
	{
		Name:       "Full-Magic-Pack",
		Adresse:    "https://github.com/MagicSetEditorPacks/Full-Magic-Pack",
		Stand:      "622160f9c8b9a134fe95e6a7ec5ff2a0596d553c",
		Schluessel: "fullMagic",
		Megabyte:   1734,
	},
}

// Quellen gibt die bekannten Bezugsquellen.
//
// **Eine Kopie und nicht die Liste selbst.** Ein Aufrufer, der einen Eintrag
// änderte, änderte ihn für jeden weiteren — und der Fehler fiele erst beim
// zweiten Bezug auf.
func Quellen() []Quelle {
	return append([]Quelle(nil), quellen...)
}
