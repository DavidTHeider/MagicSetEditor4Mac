package app

// **Mehr als ein Anbieter** — Werkstück A8 von Phase 5.3
// (#1447).
//
// # Was sich gegenüber A7 ändert
//
// **A7 kannte eine Laufzeit.** `kilaufzeit` war ein Bauplan, `frageLaufzeit`
// gab **einen** Stand, und der Katalog trug Einrichtung, Fassung und Adresse
// als **flache** Felder. Mit dem zweiten Anbieter trägt das nicht mehr: die
// Einrichtung ist eine Frage **je Anbieter**, und eine Zahl darüber wäre eine
// Zusammenfassung, die das Wichtige weglässt.
//
// **Also eine Liste.** Jeder Anbieter sagt für sich, ob er eingeschaltet ist,
// ob er antwortet und was er kann; der Katalog führt die Modelle aller
// zusammen, und `nurLokal` gilt über alle.
//
// # Warum das eine brechende Änderung ist und trotzdem richtig
//
// **`lesen.ki.modelle` verliert drei Felder und bekommt eines.** Die
// Einordnung steht im PR (§16): **brechend**, und zwar innerhalb derselben
// unveröffentlichten Fassung — A7 ist in `main` und in keiner Freigabe. Wer
// sie stehen ließe, hätte beim zweiten Anbieter drei Felder, die von einem
// von zweien sprechen und so aussehen, als sprächen sie von beiden.

import (
	"context"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// kianbieter ist eine Quelle samt dem, was über sie feststeht.
type kianbieter struct {
	// name ist der Anbietername, unter dem seine Modelle im Katalog stehen.
	name string

	// bauen gibt die Quelle — nil heißt: dieser Anbieter ist nicht verdrahtet.
	bauen func() KILaufzeitquelle

	// eingeschaltet sagt, ob überhaupt gefragt werden darf.
	eingeschaltet bool

	// adresse ist, wohin gefragt wird — leer bei einem Anbieter mit fester.
	adresse string

	// gewaehlt sagt, ob **jemand** ihn eingeschaltet hat.
	//
	// # Warum das etwas anderes ist als [kianbieter.eingeschaltet]
	//
	// **Weil eine Richtlinie dazwischenstehen kann.** `richtlinie.nurlokal`
	// schaltet **jeden fremden Dienst** ab (O3, #1427) — der wirksame Wert
	// eines `dienst.`-Schlüssels ist dann `aus`, auch wenn die Wahl `an`
	// sagt. Wer nur den wirksamen läse, sähe einen abgeschalteten Anbieter
	// und könnte nicht sagen, **wer** ihn abgeschaltet hat.
	gewaehlt bool

	// sitzungAus sagt, dass er für **diesen Prozess** abgeschaltet ist —
	// nur das Gerätemodell kennt das (AI-05.AK-04, #2042).
	sitzungAus bool

	// hinweis sagt, warum er aus ist, wenn es **nicht** an einem Schalter
	// liegt — beim Gerätemodell: was auf diesem Mac fehlt (AI-05.AK-03).
	hinweis string
}

// anbieterstand ist, was ein Anbieter über sich hergibt.
type anbieterstand struct {
	name        string
	einrichtung string
	fassung     string
	adresse     string
	modelle     []domain.KIModell
	befunde     []domain.Befund

	// messbar sind die Modelle, die der Anbieter **nennt**, ohne etwas über
	// sie zu sagen (A9, #1448).
	//
	// **Sie stehen nicht im Katalog** — ohne Beleg gehört keines hinein
	// (0097) —, und ohne diese Liste wären sie unerreichbar: wer messen
	// will, braucht einen Namen, und der stünde nur in einem Satz.
	messbar []string

	// sitzungAus kommt aus [kianbieter.sitzungAus].
	sitzungAus bool

	// gemessen sind die Modelle, für die eine Messung im Bestand liegt.
	//
	// **Neben [anbieterstand.messbar] und nicht darin**: das eine ist, was
	// noch niemand gemessen hat, das andere, was jemand zurücknehmen kann.
	gemessen []string
}

// frageAnbieter holt den Stand jedes verdrahteten Anbieters.
//
// **Der Reihe nach und nicht nebenläufig.** Zwei Anbieter sind zwei Rufe;
// nebenläufig gemacht bräuchte es eine Reihenfolge für die Befunde und eine
// Antwort auf die Frage, was gilt, wenn einer hängt. Beides für einen
// Zeitgewinn, den niemand gemessen hat (§10).
func (a *App) frageAnbieter(ctx context.Context) (
	[]anbieterstand, []kianbieter, error,
) {
	anbieter, fehler := a.kianbieter(ctx)
	if fehler != nil {
		return nil, nil, fehler
	}

	// **Die gemessenen Fähigkeiten kommen einmal und nicht je Anbieter**
	// (A9, #1448): sie liegen alle in derselben Ablage, und ein Ruf je
	// Anbieter läse dieselbe Tabelle mehrfach.
	proben := a.gemesseneFaehigkeiten(ctx)

	staende := make([]anbieterstand, 0, len(anbieter))

	for _, einer := range anbieter {
		staende = append(staende, a.standVon(ctx, einer, proben))
	}

	return staende, anbieter, nil
}

// gemesseneFaehigkeiten liest, was an diesem Bestand gemessen wurde.
//
// **Ohne Fehlerrückgabe, mit Absicht.** Eine unlesbare Ablage ist kein Grund,
// den Katalog zu leeren — sie ist ein Grund, weniger zu **belegen**. Ohne
// Messungen steht dort, was die Anbieter selbst sagen; das ist eine engere
// Auskunft, keine falsche (§16, Satz 5).
func (a *App) gemesseneFaehigkeiten(ctx context.Context) []domain.KIModellprobe {
	profil, fehler := storage.ProfilAus(ctx)
	if fehler != nil {
		return nil
	}

	ablage, fehler := a.probenablage()
	if fehler != nil {
		return nil
	}

	proben, fehler := ablage.Alle(ctx, profil)
	if fehler != nil {
		return nil
	}

	return proben
}

// standVon holt den Stand eines Anbieters.
func (a *App) standVon(
	ctx context.Context, einer kianbieter, proben []domain.KIModellprobe,
) anbieterstand {
	if !einer.eingeschaltet || einer.bauen == nil {
		stand := anbieterstand{
			name: einer.name, einrichtung: KIEinrichtungAus,
			adresse: einer.adresse, sitzungAus: einer.sitzungAus,
		}

		if einer.hinweis != "" {
			stand.befunde = []domain.Befund{{Art: domain.ArtUmgebung, Meldung: einer.hinweis}}
		}

		return stand
	}

	stand := a.frageQuelle(ctx, einer.bauen(), einer.adresse)

	// **Die Messungen liegen über der Auskunft des Anbieters**, und sie
	// ersetzen sie nicht: sagt er dasselbe, ist die Fähigkeit zweimal
	// belegt, und [domain.KIModell.Kann] fragt ohnehin nur, ob **ein**
	// gültiger Beleg dasteht.
	modelle := domain.MitProben(
		umgestuft(stand.modelle, einer.name, stand.adresse), proben)

	// **Und ein gemessenes Modell kommt dazu, wenn der Anbieter es nennt** —
	// #1541.
	//
	// **Ohne diese Zeile blieb eine bezahlte Messung bei OpenAI folgenlos**:
	// dort führt `Modelle()` **immer** eine leere Liste, jedes Modell wird zu
	// einem Befund, und `MitProben` ergänzt nur Vorhandenes. Gemessen,
	// gebucht, gemeldet — und zu nichts zu gebrauchen. Gefunden in der
	// Sichtprüfung am laufenden Fenster.
	//
	// **Die Aliasse kommen aus den Befunden dieser Antwort** und nicht aus
	// dem Bestand: „kein Katalogeintrag ohne Anbieterantwort" (0097) bleibt
	// damit stehen.
	modelle = domain.MitGemessenen(modelle, proben, genannteAus(stand.befunde),
		einer.name, ortVon(einer.name, stand.adresse))

	return anbieterstand{
		name: einer.name, einrichtung: einrichtungMit(stand.einrichtung, modelle),
		fassung: stand.fassung, adresse: stand.adresse,
		modelle: modelle, befunde: stand.befunde,
		messbar:  messbareAus(stand.befunde, modelle),
		gemessen: gemesseneAus(einer.name, proben),
	}
}

// genannteAus nennt die Aliasse, die in den Befunden dieser Antwort stehen.
//
// **Ein Befund mit Dateinamen ist die Aussage »dieses Modell gibt es«** — die
// Fähigkeit fehlt, das Modell nicht. Dieselbe Lesart wie in [messbareAus],
// und dieselbe Quelle: eine zweite Liste daneben liefe auseinander.
func genannteAus(befunde []domain.Befund) []string {
	genannt := make([]string, 0, len(befunde))

	for _, befund := range befunde {
		if befund.Datei != "" && !slices.Contains(genannt, befund.Datei) {
			genannt = append(genannt, befund.Datei)
		}
	}

	return genannt
}

// gemesseneAus nennt die Modelle eines Anbieters mit einer Messung.
func gemesseneAus(anbieter string, proben []domain.KIModellprobe) []string {
	gemessen := make([]string, 0, len(proben))

	for _, probe := range proben {
		if probe.Anbieter != anbieter || slices.Contains(gemessen, probe.Alias) {
			continue
		}

		gemessen = append(gemessen, probe.Alias)
	}

	return gemessen
}

// messbareAus nennt die Modelle, über die nichts feststeht.
//
// **Aus den Befunden und nicht aus einer zweiten Liste.** Ein Befund mit
// einem Dateinamen ist genau die Aussage »dieses Modell gibt es, und über
// es steht nichts fest« — bei OpenAI für jedes, bei Anthropic für eines
// ohne `capabilities` oder ohne lesbares Datum.
//
// **Was schon im Katalog steht, fällt weg**: dort ist etwas belegt, und ein
// zweiter Weg dorthin verwirrte nur.
func messbareAus(befunde []domain.Befund, modelle []domain.KIModell) []string {
	imKatalog := make(map[string]bool, len(modelle))
	for _, modell := range modelle {
		imKatalog[modell.Alias] = true
	}

	messbar := make([]string, 0, len(befunde))

	for _, befund := range befunde {
		if befund.Datei == "" || imKatalog[befund.Datei] ||
			slices.Contains(messbar, befund.Datei) {
			continue
		}

		messbar = append(messbar, befund.Datei)
	}

	return messbar
}

// einrichtungMit hebt den Stand an, wenn eine Messung etwas belegt.
//
// # Warum das nötig ist
//
// **Weil der Anbieter den Stand meldet, bevor die Messungen dazukommen.**
// OpenAI meldet `ohnefaehigkeit` — es nennt ja keine —, und wer danach
// zwei Modelle gemessen hat, läse trotzdem „keines kann etwas Belegtes".
// Das wäre eine Auskunft, die ihrem eigenen Katalog widerspricht (§10).
func einrichtungMit(gemeldet string, modelle []domain.KIModell) string {
	// **Zwei Stände sagen dasselbe: nichts Brauchbares** —
	// #1541.
	//
	// Hier stand nur `ohnefaehigkeit`. Bei OpenAI meldet der Klient aber
	// `ohnemodell` (seine Liste ist **immer** leer), und ein gemessenes
	// Modell stand danach im Katalog, während der Anbieter daneben „hat kein
	// Modell" sagte — zwei Auskünfte, von denen eine falsch war.
	//
	// **`unerreichbar` und `aus` bleiben, was sie sind**: eine Messung von
	// gestern macht einen abgeschalteten Anbieter nicht bereit.
	if gemeldet != KIEinrichtungOhneFaehigkeit &&
		gemeldet != KIEinrichtungOhneModell {
		return gemeldet
	}

	for _, modell := range modelle {
		if len(modell.Geprueft) > 0 {
			return KIEinrichtungBereit
		}
	}

	return gemeldet
}

// alleModelle führt die Modelle aller Anbieter zusammen.
func alleModelle(staende []anbieterstand) []domain.KIModell {
	var alle []domain.KIModell

	for _, stand := range staende {
		alle = append(alle, stand.modelle...)
	}

	return alle
}

// besteEinrichtung gibt den Stand, der am weitesten kommt.
//
// # Wofür es sie gibt
//
// **Für die eine Zeile, die über allem steht.** Wer die Seite aufschlägt,
// will zuerst wissen, ob überhaupt etwas geht; **welcher** Anbieter woran
// hängt, steht darunter je Zeile.
//
// **Der beste und nicht der schlechteste**: ein eingerichteter Anbieter macht
// die Anwendung brauchbar, auch wenn ein zweiter aus ist. Andersherum sähe
// eine funktionierende Einrichtung wie eine kaputte aus, sobald jemand einen
// zweiten Anbieter kennt und nicht benutzt.
func besteEinrichtung(staende []anbieterstand) string {
	reihe := []string{
		KIEinrichtungAus, KIEinrichtungUnerreichbar, KIEinrichtungOhneModell,
		KIEinrichtungOhneFaehigkeit, KIEinrichtungBereit,
	}

	beste := KIEinrichtungAus

	for _, stand := range staende {
		if slices.Index(reihe, stand.einrichtung) > slices.Index(reihe, beste) {
			beste = stand.einrichtung
		}
	}

	return beste
}

// umgestuft setzt den Ort nach dem, was über den Anbieter feststeht.
//
// **Die Adresse schlägt die Selbstauskunft — und nur in eine Richtung.** Sie
// kann „nicht lokal" beweisen und „lokal" nicht ([domain.KIOrtFuerAdresse]).
//
// **Ein Anbieter ohne eigene Adresse rechnet dort, wo er steht.** Anthropic
// hat keine Schleifenadresse und keine Einstufung: seine Modelle sind Cloud,
// und das Blatt sagt es schon. Die Umstufung hier greift deshalb nur, wo eine
// Adresse eingestellt werden kann.
func umgestuft(
	modelle []domain.KIModell, anbieter, adresse string,
) []domain.KIModell {
	if anbieter != ollamaanbieter ||
		domain.KIOrtFuerAdresse(adresse) == domain.KIOrtLokal {
		return modelle
	}

	umgestellt := make([]domain.KIModell, 0, len(modelle))

	for _, modell := range modelle {
		modell.Ort = domain.KIOrtCloud
		umgestellt = append(umgestellt, modell)
	}

	return umgestellt
}

// vonDerRichtlinieAbgeschaltet nennt, was jemand eingeschaltet hat und was
// trotzdem nicht gefragt wird.
//
// # Warum es diesen Befund gibt
//
// **Weil die Auskunft sonst »aus« lautet und damit falsch klingt.** Wer den
// Schalter umgelegt hat, sieht ihn auf `an` stehen und keine Modelle; der
// Grund steht in einer **anderen** Einstellung, und ohne diesen Satz müsste
// er ihn suchen (§6).
//
// # Und warum über die Wahl und nicht über die Modelle
//
// **Weil unter »nur lokal« gar nicht erst gefragt wird.** Der Riegel am
// Transport hält schon die Modellliste auf — schon sie wäre eine Auskunft
// nach draußen. Wer hier die Modelle zählte, zählte null und schwiege
// deshalb; gefunden am 2026-09-11 an einem Lauf, der genau das tat.
func vonDerRichtlinieAbgeschaltet(anbieter []kianbieter) []domain.Befund {
	var namen []string

	for _, einer := range anbieter {
		// **Für die Sitzung abgeschaltet ist etwas anderes** (AI-05, #2042):
		// das hat jemand eben getan, und die Richtlinie war es nicht.
		if einer.gewaehlt && !einer.eingeschaltet && !einer.sitzungAus {
			namen = append(namen, einer.name)
		}
	}

	if len(namen) == 0 {
		return nil
	}

	return []domain.Befund{{
		Art: domain.ArtUmgebung,
		Meldung: strings.Join(namen, ", ") + " ist eingeschaltet und wird " +
			"trotzdem nicht gefragt: »nur lokal« schaltet jeden fremden " +
			"Dienst ab, und diese Modelle rechnen nicht auf diesem Rechner.",
	}}
}
