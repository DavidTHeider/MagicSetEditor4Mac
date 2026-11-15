package mcp_test

// **Die Zusage des KI-Vorschlags** — Werkstück A7 von Phase 5.3
// (#1446).
//
// **Eine eigene Datei**, weil `zusagenliste_test.go` mit ihr über die
// Zeilengrenze aus §2 ginge.

// kiVorschlagszusage gibt die Zusage des KI-Vorschlags.
//
// # Warum er zu „eine Karte je Aufruf" gehört
//
// **Er ändert die Karte, die er nennt — und nur, wenn er es soll.** Ohne
// `uebernehmen` schreibt er **gar nichts**; mit ihm setzt er Felder
// **einer** Karte, deren Kennung in seinen Argumenten steht.
//
// **Er überschreibt dabei** und steht deshalb nicht in der Liste
// darunter. Umkehrbar ist es aus demselben Grund wie bei `schreiben.feld`:
// der Schreibweg legt den vorigen Wert im Verlauf ab, und
// `schreiben.zuruecknehmen` holt ihn.
//
// **Was das Schema nicht durchlässt, wird nicht geschrieben** — es steht
// als Befund in der Antwort. Der Umfang ist damit nicht „was ein Modell
// sagt", sondern „was die Vorlage kennt und zulässt".
//
// # Und die zwei Wege der Messung stehen daneben
//
// **Sie fassen keine Karte an** (A9, #1448). Eine Messung schreibt eine
// **Aussage über ein Modell** — je Profil, Anbieter, Alias und Fähigkeit
// eine —, und `vergessen` nimmt sie zurück. Nach 0044 ist der Maßstab die
// Umkehrbarkeit, und hier ist sie vollständig: die eine Richtung hebt die
// andere auf.
func kiVorschlagszusage() map[string]bool {
	return map[string]bool{
		"schreiben.ki.vorschlag":        true,
		"schreiben.ki.modell.pruefen":   true,
		"schreiben.ki.modell.vergessen": true,

		// **Der Sitzungsschalter fasst den Bestand gar nicht an** (AI-05,
		// #2042): er lebt im Arbeitsspeicher des Servers, und sein Gegenteil
		// ist derselbe Ruf mit dem anderen Wert — umkehrbar nach 0044.
		"schreiben.ki.geraetemodell.sitzung": true,

		// **Der Batch fasst keine Karte an** (A13.1, #1452). Er schlägt für
		// mehrere vor und **schreibt an keiner**; übernommen wird je
		// Vorschlag über `schreiben.ki.vorschlag`, und dort gilt die Zusage
		// darüber.
		//
		// **Was er schreibt, ist ein Kostenposten** — genau **einer** für
		// den ganzen Batch ([0105](../../docs/entscheidungen/0105-eine-freigabe-fuer-einen-batch.md)),
		// und er schließt ihn selbst wieder. Nach 0044 ist der Maßstab die
		// Umkehrbarkeit, und die ist hier vollständig: der Posten ist beim
		// Zurückkommen entweder als unklar gemerkt oder freigegeben, und
		// offen bleibt keiner (#1542).
		"schreiben.ki.batch": true,

		// **Die Einschätzung fasst noch weniger an** (A13.2a, #1452). Sie
		// ändert **keine** Karte, **keine** Sammlung und **keinen**
		// Spielstand — das Issue führt es als eigene Abnahme. Was sie
		// schreibt, ist derselbe Kostenposten wie beim Batch, und sie
		// schließt ihn ebenso selbst.
		//
		// **Und sie wird nicht übernommen.** Eine Einschätzung ist eine
		// Auskunft; es gibt kein zweites Werkzeug, das sie in eine Karte
		// trüge. Nach 0044 ist der Maßstab die Umkehrbarkeit — hier gibt es
		// nichts umzukehren.
		"schreiben.ki.einschaetzung": true,

		// **Und die Kartenerkennung ebenso** (A13.3, #1452). Sie liest ein
		// Foto und nennt Drucke, die dazu passen — sie **bindet keinen** und
		// legt **kein Exemplar** an. Das Issue führt es als eigene Abnahme:
		// „Bild einer mehrdeutigen/falschen/keiner Karte erzeugt keine
		// falsche Printing-Bindung oder Sammlungsmengen."
		//
		// **Was sie schreibt, ist derselbe Kostenposten** — eine Sichtanfrage
		// kostet Geld —, und sie schließt ihn selbst. Umzukehren gibt es auch
		// hier nichts.
		"schreiben.ki.erkennung": true,
	}
}
