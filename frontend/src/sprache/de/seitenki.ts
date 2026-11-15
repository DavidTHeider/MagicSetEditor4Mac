// **Die Texte der KI** — Werkstück A6 von Phase 5.3
// (#1445),
// erweitert für A7
// (#1446).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als einer
// ankommt** (§9): `seiteneinstellungen.ts` stand mit ihnen an der
// Zeilengrenze aus §2.
export const ki = {
  // **Was über KI-Modelle feststeht** — A6, #1445.
  kimodelleTitel: 'KI-Modelle',
  kimodelleKeine:
    'Es steht kein KI-Modell fest. Eine Modellliste eines Anbieters ist noch keine belegte Fähigkeit — hier stehen nur Modelle, die geprüft wurden.',
  kiNurLokal:
    'Es gilt „nur lokal“: kein Modell außerhalb dieses Rechners wird gefragt.',
  kiAuchCloud:
    'Modelle bei fremden Anbietern sind erlaubt. Was an sie geht, verlässt diesen Rechner.',
  kiBeleg: '{faehigkeit} — geprüft am {datum}: {beleg}',
  kiKontextfenster: 'Kontextfenster: {tokens} Tokens — {beleg}',
  kiKontextfensterUnbekannt: 'Kontextfenster: nicht angegeben',
  kiHerkunftUnbekannt:
    'Herkunft unbekannt — es ist nicht belegt, welche Datei hier wirklich läuft.',
  // **Die Einrichtung** — A7, #1446. Fünf Stände, und sie sind nicht
  // dasselbe: eine leere Liste sagt nicht, **warum** sie leer ist.
  kiEinrichtungAus:
    'Nicht eingeschaltet. Nichts wird gefragt, und nichts verlässt diesen Rechner.',
  kiEinrichtungUnerreichbar:
    'Unter der eingestellten Adresse antwortet niemand. Diese Anwendung startet dort nichts und installiert nichts.',
  kiEinrichtungOhneModell:
    'Es antwortet, hat aber kein Modell. Diese Anwendung lädt keines nach.',
  kiEinrichtungOhneFaehigkeit:
    'Es gibt Modelle, aber keines hat eine Fähigkeit belegt, mit der hier etwas anzufangen wäre.',
  kiEinrichtungBereit:
    'Mindestens ein Modell steht mit einer belegten Fähigkeit bereit.',
  // **Der Hinweis auf Ollama steht bei der lokalen Laufzeit** und nicht mehr
  // in den fünf Ständen: die gelten seit A8 (#1447) für **jeden** Anbieter,
  // und „Läuft Ollama?“ traf bei Anthropic auf nichts zu (Fund der
  // Sichtprüfung am 2026-09-12).
  kiLaufzeitOllama:
    'Antwortet hier nichts, läuft Ollama nicht. Diese Anwendung startet es nicht und installiert es nicht — das geschieht in Ollama selbst.',
  kiLaufzeitTitel: 'Lokale KI',
  kiLaufzeitFassung: 'Laufzeit: {fassung}',
  kiAdresse: 'Gefragt wird {adresse}.',
  kiAdresseVorgabe: 'die Vorgabeadresse auf diesem Rechner',
  kiCloudAus: 'Die Auslagerung in Ollamas Cloud ist abgeschaltet.',
  kiCloudAn: 'Die Auslagerung in Ollamas Cloud ist nicht abgeschaltet.',
  kiCloudUngewiss:
    'Ob die Auslagerung in Ollamas Cloud abgeschaltet ist, steht von hier aus nicht fest.',
  kiCloudBeleg: 'Beleg: {beleg}',
  // **Der Vorschlag** — A7, #1446.
  kiVorschlagTitel: 'Vorschlag von einem Modell',
  kiVorschlagModell: 'Modell',
  kiVorschlagAuftrag: 'Auftrag',
  kiVorschlagAuftragHinweis: 'Was soll das Modell tun?',
  kiVorschlagen: 'Vorschlagen',
  kiVorschlagLaeuft:
    'Das Modell arbeitet — das kann beim ersten Mal dauern, weil es geladen wird.',
  kiVorschlagAbbrechen: 'Abbrechen',
  kiVorschlagUebernehmen: 'Übernehmen',
  kiVorschlagNichts:
    'Das Modell hat kein Feld vorgeschlagen, das die Vorlage kennt.',
  kiVorschlagAendertNichts:
    'Ein Vorschlag ändert die Karte nicht. Erst „Übernehmen“ schreibt.',
  kiVorschlagUebernommen: '{felder} Felder wurden geschrieben.',
  kiVorschlagAblauf: 'Übernehmbar bis {ablauf}.',
  // **Der zweite Anbieter** — A8, #1447.
  kiAnbieterZeile: 'Anbieter {anbieter}: {stand}',
  kiVorschlagAnbieter: 'Anbieter und Modell',
  kiVorschlagModellWaehlen: 'Bitte wählen',
  kiVorschlagOhneModell:
    'Es steht kein Modell bereit. Was fehlt, steht unter „KI-Modelle“.',
  kiVorschlagAntwortVon: 'Geantwortet hat {anbieter}/{modell}.',
  // **Messen, was ein Modell kann** — A9, #1448.
  kiMessbarTitel: 'Nicht beschriebene Modelle',
  kiAdresseSocket:
    'Gefragt wird über einen eigenen Unix-Socket dieses Programms — kein Port.',
  kiGeraetSitzung: 'In dieser Sitzung fragen',
  kiGeraetSitzungHinweis:
    'Gilt bis zum Neustart; danach gilt wieder die Einstellung „Apples Gerätemodell“.',
  kiFeldVorschlagen: 'Vorschlagen',
  kiFeldVorschlagenFuer: '{feld} — Vorschlag',
  kiFeldAuftragName:
    'Schlage einen passenden Eigennamen für diese Karte vor — nicht den Kartentyp und keine Beschreibung der Regel. Antworte auf Deutsch, nur mit dem Namen.',
  kiFeldAuftragFlavor:
    'Schlage einen kurzen Flavortext für diese Karte vor, ein oder zwei Sätze: Stimmung oder Geschichte, die den Regeltext nicht wiederholt. Antworte auf Deutsch.',
  kiFeldKarte: 'Die Karte bisher:',
  kiFeldUebernehmen: 'Übernehmen',
  kiFeldVon: 'von {anbieter} · {weg}',
  kiFeldWegGewaehlt: 'gewählt',
  kiFeldWegVorrang: 'Vorrang',
  kiFeldVorrangFragt: 'Vorrang fragt: {reihe}',
  kiFeldWegRueckfall: 'Rückfall',
  kiFeldAbgelehnt: 'Das Modell hat abgelehnt: {grund}',
  kiFeldNichts: 'Für dieses Feld kam kein Vorschlag.',
  kiMessbarSatz:
    '{anbieter} nennt diese Modelle, sagt aber nichts über sie. Eine Messung schickt einen kleinen Auftrag mit festem Schema — sie kostet, was dieser eine Auftrag kostet.',
  kiMessen: 'Messen',
  kiMessungVergessen: 'Messung zurücknehmen',
  kiMessungLaeuft: 'Es wird gemessen …',
  kiMessungGeglueckt:
    '{alias} hat geantwortet: {faehigkeiten} sind jetzt belegt.',
  // **Auswahl, Vergleich und Herkunft** — A11, #1450.
  kiVorschlagAuswahl: 'Übernehmen: {feld}',
  kiVorschlagBisher: 'Bisher: {wert}',
  kiVorschlagBisherLeer: 'Bisher leer.',
  kiVorschlagNichtsGewaehlt:
    'Es ist kein Feld angehakt. Übernommen wird, was angehakt ist.',
  kiHerkunftTitel: 'Woher die Werte stammen',
  kiHerkunftZeile: '{feld}: {anbieter}/{modell} am {zeitpunkt}',
  kiHerkunftAufgabe: 'Auftrag war: {aufgabe}',
  kiHerkunftVeraendert: 'Seither von Hand geändert.',
  kiHerkunftKeine: 'Kein Feld dieser Karte stammt aus einem Vorschlag.',
  // **Der Bildvorschlag** — A12, #1451.
  kiBildTitel: 'Bild von einem Modell',
  kiBildAendertNichts:
    'Erzeugen ändert die Karte nicht. Was entsteht, ist ein Vorschlag mit Frist — er wird angesehen, zugeschnitten und erst dann eingesetzt.',
  kiBildFeld: 'Bildfeld',
  kiBildAufgabe: 'Was soll darauf zu sehen sein?',
  kiBildAufgabeHinweis: 'Ein Blitz über einer weiten Ebene, Ölgemälde',
  kiBildGroesse: 'Größe',
  kiBildGuete: 'Güte',
  kiBildBearbeiten: 'Das vorhandene Bild bearbeiten statt ein neues erzeugen',
  kiBildErzeugen: 'Bild erzeugen',
  kiBildLaeuft:
    'Das Bild entsteht bei {anbieter}/{modell} — das dauert Sekunden bis Minuten.',
  kiBildAbbrechen: 'Abbrechen',
  kiBildAbgebrochen:
    'Abgebrochen. Es wurde nichts übernommen; was bis dahin an Kosten entstand, steht im Kostenjournal.',
  kiBildVorschau: 'Vorschlag für {feld}',
  kiBildMasse: '{breite} × {hoehe} Bildpunkte, {kilobytes} kB',
  kiBildAblauf: 'Gilt bis {ablauf}.',
  kiBildUmschrift: 'Gezeichnet wurde: {umschrift}',
  kiBildNachweisTitel: 'Woher dieses Bild stammt',
  kiBildAngabeGroesse: 'Größe',
  kiBildAngabeGuete: 'Güte',
  kiBildAngabeUmschrift: 'Umschrift',
  kiBildAngabeMedientyp: 'Dateiformat',
  kiBildUebernehmen: 'Zuschneiden und einsetzen',
  kiBildVerwerfen: 'Verwerfen',
  kiBildVerworfen: 'Der Vorschlag ist weg.',
  kiBildKosten:
    'Ein Bildauftrag wird gesondert freigegeben — die Freigabe für Text gilt dafür nicht.',
  // **Die Freigabe eines bezahlten Auftrags** —
  // #1539.
  //
  // **Sie stand nirgends**, und deshalb ging am Fenster kein einziger
  // kostenpflichtiger Auftrag hinaus: die Kostenwache verlangt die
  // ausdrückliche Freigabe eines Menschen, und keine Bindung stellte eine aus.
  kiFreigabeTitel: 'Dieser Auftrag kostet Geld',
  kiFreigabeSatz:
    'Das gewählte Modell rechnet bei {anbieter}. Gib einen Höchstbetrag frei — er gilt für diesen einen Auftrag und läuft nach fünf Minuten ab.',
  kiFreigabeBetrag: 'Höchstbetrag in der kleinsten Einheit (bei EUR: Cent)',
  kiFreigabeVerfuegbar: 'Heute verfügbar: {betrag}',
  kiFreigabeOhneBudget:
    'Für {anbieter} ist nichts freigegeben. Ohne Freigabe geht kein kostenpflichtiger Auftrag hinaus — die Grenze wird unter „Kosten“ gesetzt.',
  kiFreigabeErteilen: 'Freigeben und fragen',
  kiFreigabeVerzichten: 'Nicht freigeben',
  kiFreigabeLaeuft: 'Die Freigabe wird ausgestellt …',
  // **Die belegte Einschätzung** — A13.2a
  // (#1452).
  //
  // **Der Vorbehalt steht im Katalog und nicht im Bauteil**, weil er in
  // beiden Sprachen gebraucht wird — der Satz aus `domain` ist deutsch und
  // reist mit der Antwort; hier steht, was das Fenster dazu sagt.
  kiEinschaetzungTitel: 'Beurteilen lassen',
  kiEinschaetzungErklaerung:
    'Ein Modell beurteilt diese Karte anhand von Stellen aus dem lokalen Suchindex. Es ändert nichts — weder an der Karte noch an der Sammlung.',
  kiEinschaetzungArt: 'Worum geht es',
  kiEinschaetzungArtRegel: 'Nach den Regeln',
  kiEinschaetzungArtBalance: 'Im Vergleich',
  kiEinschaetzungFrage: 'Was soll beurteilt werden',
  // **Nicht derselbe Text wie der Titel darüber.** Er stand einmal
  // zweimal da — auf der Aufklappzeile und auf dem Knopf —, und dann
  // sagen zwei Bedienelemente dasselbe, ohne dasselbe zu tun (§6).
  kiEinschaetzungFragen: 'Einschätzung holen',
  kiEinschaetzungLaeuft: 'Das Modell liest die Belege …',
  kiEinschaetzungLeer: 'Noch nichts beurteilt.',
  kiEinschaetzungBelegt: 'Alles Genannte steht in den Belegen.',
  kiEinschaetzungUnbelegt:
    'Diese Antwort nennt etwas, das in keinem Beleg steht. Was das ist, steht darunter.',
  kiEinschaetzungBefunde: 'Was nicht belegt ist',
  kiEinschaetzungEinschraenkungen: 'Was nicht betrachtet wurde',
  kiEinschaetzungBelege: 'Die mitgegebenen Stellen',
  kiEinschaetzungStaende: 'Stand der Quellen',

  // **Die Balanceanalyse** — A13.2b
  // (#1452).
  kiEinschaetzungFormat: 'Format',
  kiEinschaetzungOhneFormat: 'keines genannt',
  kiEinschaetzungVergleiche: 'Womit verglichen wurde',
  kiEinschaetzungStrategie: 'Gefunden über',
  kiEinschaetzungOhneVergleich:
    'Diese Antwort nennt keine Vergleichskarte. Eine Aussage darüber, wie eine Karte dasteht, steht im Vergleich.',
  kiEinschaetzungSynergien: 'Welche Synergien außen vor blieben',

  // **Der Batch und die atomare Gruppe** — A13.1b
  // (#1452).
  kiBatchTitel: 'Mehrere Karten auf einmal',
  kiBatchSatz:
    'Ein Modell beantwortet denselben Auftrag für jede gewählte Karte. Geschrieben wird dabei nichts — das entscheidet der zweite Schritt.',
  kiBatchZiele: '{gewaehlt} von {gesamt} Karten gewählt',
  kiBatchOhneKarten: 'Dieses Set hat noch keine Karte.',
  kiBatchOhneName: '(ohne Namen)',
  kiBatchAuftrag: 'Auftrag für jede Karte',
  kiBatchFahren: 'Vorschläge holen',
  kiBatchLaeuft: 'Läuft …',
  kiBatchLagen: 'Was daraus wurde',
  kiBatchGanzOderGar:
    'Übernommen wird alles auf einmal oder nichts. Scheitert ein Ziel, bleibt jede Karte, wie sie war — und die Vorschläge bleiben offen.',
  kiBatchGruppe: 'Als Gruppe übernehmen ({karten})',
  kiBatchGruppeLaeuft: 'Wird übernommen …',
  kiBatchGeschrieben: '{karten} Karten geschrieben.',

  // **Die Kartenerkennung auf einem Foto** — A13.3
  // (#1452).
  kiErkennungTitel: 'Karte auf einem Foto erkennen',
  kiErkennungSatz:
    'Ein Modell liest, was auf dem Foto steht. Welche Karte das ist, entscheidet dieselbe Drucksuche wie bei einer getippten Liste — und aufgenommen wird dabei nichts.',
  kiErkennungLesen: 'Foto wählen und lesen',
  kiErkennungLaeuft: 'Das Modell liest das Foto …',
  kiErkennungLeer: 'Noch kein Foto gelesen.',
  kiErkennungGelesen: 'Was gelesen wurde',
  kiErkennungNichtGelesen: 'Wozu nichts zu sagen war',
  kiErkennungUnlesbar: 'nicht lesbar',
  kiErkennungOhneWert: 'nichts gelesen',
  kiErkennungArten: {
    name: 'Kartenname',
    setcode: 'Set',
    nummer: 'Sammlernummer',
    typzeile: 'Typzeile',
    layout: 'Bauart',
    kunstwerk: 'Illustration',
  } as Record<string, string>,
  kiErkennungKandidaten: 'Welche Drucke passen',
  kiErkennungDrucke: 'Es passen {drucke} Drucke.',
  kiErkennungEinDruck: 'Genau ein Druck passt.',
  kiErkennungWeitere:
    'Und dahinter geht es weiter — gezählt ist nur die erste Seite.',
  kiErkennungOhneKandidat:
    'Zu dem, was gelesen wurde, passt kein Druck. Das kann am Foto liegen oder daran, dass die Karte hier nicht geführt wird.',
  kiErkennungGestuetzt: 'Stützt sich auf: {arten}',
  kiErkennungOhneStuetze:
    'Nichts an diesem Druck ist auf dem Foto wiederzufinden — er steht nur da, weil die Suche ihn zum Begriff gefunden hat.',
  kiErkennungNichtsAngelegt:
    'Aufgenommen wird nichts. Wer ein Exemplar eintragen will, tut das oben im Formular — dann steht fest, welcher Druck gemeint ist.',
  kiErkennungStand: 'Stand der Drucksuche: {alter}',
  kiErkennungModell: 'Gelesen von {modell} bei {anbieter}',

  // **Die Belegfrage** — A13.4
  // (#1452).
  kiVorschlagBelegfrage: 'Stellen aus dem Wissensbestand mitgeben (wahlfrei)',
  kiVorschlagBelegfrageHinweis:
    'wonach gesucht werden soll — leer heißt: gar nicht suchen',
} as const
