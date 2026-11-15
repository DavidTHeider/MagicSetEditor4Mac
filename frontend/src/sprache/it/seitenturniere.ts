// Die Texte der **Turniere** — Werkstück 5 von Phase 4.7
// (#1294),
// nach Vermerk 0062.
//
// **Ein eigener Baustein**, wie `seitendecks.ts` und aus demselben Grund:
// `seiten.ts` steht an der Zeilengrenze aus §2.
export const turniere = {
  alleTurniere: 'Tutti i tornei ({anzahl})',
  titel: 'Tornei',
  erklaerung:
    'Un torneo è un percorso nel tempo: turni, abbinamenti, risultati. La classifica viene calcolata da questi dati e non è memorizzata da nessuna parte — se correggi un risultato, vedi subito la nuova classifica.',

  leer: 'Ancora nessun torneo. «Crea torneo …», in alto, ne inizia uno; «Importa torneo …» ne porta uno da un file.',

  neu: 'Crea torneo',
  neuOeffnen: 'Crea torneo …',
  name: 'Nome',
  spiel: 'Gioco',
  notiz: 'Nota',
  einzelspiel: 'Una sola partita per match',
  einzelspielHinweis:
    'In tal caso la percentuale di partite vinte e quella degli avversari non dicono nulla — sono copie dei valori dei match. Vengono comunque calcolate.',
  anlegen: 'Crea',

  // **The wizard** — DOM-04.AK-01 (#1699).
  assistent: 'Un nuovo torneo in quattro passaggi',
  schrittTurnier: 'Torneo',
  schrittRegeln: 'Regole',
  schrittTeilnehmende: 'Partecipanti',
  schrittErsteRunde: 'Primo turno',
  weiter: 'Avanti',
  zurueck: 'Indietro',
  verlassen: 'Continua più tardi',
  vorlage: 'Modello',
  ohneVorlage: 'Nessun modello',
  vorlageHinweis:
    'Un modello porta con sé gioco, formato, regole e nota. Se vuoi anche i nomi, mai i mazzi: una lista presentata vale per un solo torneo.',
  mitTeilnehmenden: 'Riprendi i nomi dei giocatori',
  nameWarum:
    'Obbligatorio — così lo ritrovi nell’elenco. Due tornei dello stesso gioco hanno nomi diversi.',
  spielWarum: 'Obbligatorio — decide quali formati e mazzi sono adatti.',
  format: 'Formato',
  ohneFormat: 'Nessun formato',
  formatWarum: 'Facoltativo — si può scegliere tra i formati di questo gioco.',
  regelnAusVorlage: 'Le regole provengono dal modello.',
  schweizerSystem:
    'Gli abbinamenti seguono il sistema svizzero: chi ha lo stesso numero di punti gioca contro, e nessuno incontra due volte lo stesso avversario. Il programma propone quanti turni giocare in base al numero di giocatori.',
  zuWenige:
    'Un turno richiede due persone che giochino l’una contro l’altra — iscrivine almeno due.',
  ersteRundeHinweis:
    'Così sarebbe il primo turno. Gli abbinamenti vengono creati solo con il pulsante.',
  ersteRundeAuslosen: 'Abbina il turno 1',

  waehlen: 'Apri',
  loeschen: 'Elimina torneo',
  handlungenAn: 'Azioni — {name}',
  handlungenKurz: '…',
  turniermenue: 'Torneo',
  loeschenFrage:
    'Eliminare questo torneo? Turni, abbinamenti e risultati vanno con esso. I mazzi dei giocatori restano.',
  loeschenJa: 'Elimina',
  loeschenNein: 'Annulla',

  teilnehmer: 'Partecipanti',
  teilnehmerLeer: 'Nessuno ancora iscritto.',
  anmelden: 'Iscrivi',
  gebiete: 'Aree di questo torneo',
  gebietTeilnehmer: 'Partecipanti',
  gebietRunden: 'Turni',
  gebietRangliste: 'Classifica',

  // **Closing and log** — DOM-04.AK-04 (#1699).
  abschliessen: 'Concludi',
  abgeschlossenAm:
    'Concluso il {datum}. Giocatori, turni e risultati sono definitivi.',
  wiedereroeffnen: 'Riapri',
  nichtWiedereroeffnen: 'Non riaprire',
  wiedereroeffnenFrage: 'Riaprire questo torneo?',
  wiedereroeffnenErklaerung:
    'Dopo, giocatori, turni e risultati potranno di nuovo cambiare. Il registro annota che è successo.',
  gebietProtokoll: 'Registro',
  protokollLeer: 'Non è ancora successo nulla.',
  protokollAngelegt: 'Creato come “{angaben}”',
  protokollAngemeldet: '{angaben} iscritto',
  protokollAusgestiegen: '{angaben} ritirato',
  protokollEingestiegen: '{angaben} iscritto di nuovo',
  protokollRunde: 'Turno {angaben} abbinato',
  protokollErgebnis: 'Risultato {paarung}: {angaben}',
  protokollZurueckgenommen: 'Risultato {paarung} ritirato',
  protokollAbgeschlossen: 'Concluso',
  protokollWiedereroeffnet: 'Riaperto',
  protokollEingelesen: 'Importato come “{angaben}”',

  // **File and report** — DOM-04.AK-03 (#1699).
  ausgeben: 'Esporta …',
  berichtSichern: 'Salva rapporto …',
  einlesen: 'Importa torneo …',
  gesichertUnter: 'Salvato in {pfad}.',
  eingelesen: 'Importato e aperto.',
  einlesebefunde: 'Durante l’importazione è emerso:',

  // **Header and links after the artboard** — DOM-04.AK-02 (#1699).
  kopfzeile: 'Sistema svizzero · Turno {runde} di {von} · {anzahl} giocatori',
  kopfzeileOhneRunde: 'Sistema svizzero · {anzahl} giocatori',
  kopfzeileOhneEmpfehlung:
    'Sistema svizzero · Turno {runde} · {anzahl} giocatori',
  gebietDecklisten: 'Decklist',
  decklistenLeer:
    'Nessuno ha ancora presentato una lista — scegliere un mazzo durante l’iscrizione ne presenta la lista.',
  deckAufschlagen: 'Apri mazzo',
  partieOeffnen: 'Apri partita',
  deck: 'Mazzo',
  ohneDeck: 'Nessun mazzo',
  deckHinweis:
    'Se viene indicato un mazzo, la sua lista viene copiata e non collegata: una decklist in un torneo è un documento presentato.',
  deckabzug: 'Lista presentata',
  deckabzugLeer: 'Nessun mazzo presentato.',
  deckabzugUnbenannt: 'Carta',
  aussteigen: 'Ritira',
  einsteigen: 'Reiscrivi',
  ausgestiegen: 'ritirato',
  ausgestiegenHinweis:
    'Non viene più abbinato e resta in classifica: la sua percentuale continua a contare per quella dei suoi avversari.',

  runden: 'Turni',
  rundenLeer: 'Ancora nessun turno abbinato.',
  runde: 'Turno {nummer}',
  empfehlung:
    '{gelegt} turni su {empfohlen} — il regolamento del torneo raccomanda {empfohlen} per {spieler} giocatori. Un suggerimento, non una prescrizione.',
  vorschlag: 'Anteprima del turno successivo',
  vorschlagTitel: 'Così sarebbe il turno successivo',
  vorschlagHinweis:
    'Solo un’anteprima, non creato. Solo il pulsante qui sotto lo crea.',
  rundeLegen: 'Crea turno',
  bye: 'Bye (2–0)',
  gegen: 'contro',
  offen: 'non ancora giocato',
  spiele: 'Partite',
  spieleA: 'Partite per il primo',
  spieleB: 'Partite per il secondo',
  unentschiedenSpiele: 'Pareggi',
  eintragen: 'Registra',
  zuruecknehmen: 'Annulla',
  ergebnis: 'Risultato',

  // **The effects of a correction** — DOM-04.AK-02 (#1699).
  folgenFrage: 'Questa correzione cambia la classifica.',
  folgenPlatz:
    '{name}: posizione {vorher} → {nachher}, {punkteVorher} → {punkteNachher} punti',
  folgenSpaeter:
    'Il turno {runden} è stato abbinato sulla vecchia classifica e resta com’è.',
  folgenEintragen: 'Registra comunque',
  folgenNicht: 'Non registrare',
  ergebnisZahlen: 'Le partite si contano in numeri interi da zero in su.',
  ergebnisOhneSpiel:
    'Un risultato senza nemmeno una partita non è un risultato — chi non ha giocato non ne ha ancora uno.',
  ergebnisZuViele: 'Un match qui si svolge al massimo su {hoechstens} partite.',

  rangliste: 'Classifica',
  ranglisteLeer: 'Ancora nulla giocato.',
  platz: 'Posizione',
  bilanz: 'Bilancio',
  matchpunkte: 'Punti',
  gegnermatchquote: 'OMW',
  spielquote: 'GW',
  gegnerspielquote: 'OGW',
  quotenHinweis:
    'Quattro tiebreaker nell’ordine delle Magic Tournament Rules: punti, percentuale di match vinti dagli avversari, percentuale di partite vinte, percentuale di partite vinte dagli avversari. Ogni percentuale ha un minimo di 0,33.',

  // **Eine Paarung, an eine Partie gebunden** (G7, #1464).
  match: 'Giocato qui',
  matchErklaerung:
    'Se questo abbinamento viene giocato in questo programma, collegalo alla partita. Il risultato viene allora comunicato da entrambe le parti e registrato solo quando entrambe dicono la stessa cosa.',
  matchNichtGebunden: 'A questo abbinamento non è collegata alcuna partita.',
  matchAnsehen: 'Verifica',
  matchPartie: 'ID della partita',
  matchSitzA: 'Posto del primo giocatore',
  matchSitzB: 'Posto del secondo giocatore',
  matchBinden: 'Collega a questa partita',
  matchGebunden: 'Collegato alla partita',
  matchSitz: 'Posto',
  matchDeckabdruck: 'Impronta del mazzo',
  matchKeineMeldung: 'Nessuno ha ancora comunicato.',
  matchMeldungVon: 'comunicato da',
  matchMelden: 'Comunica per questo posto',
  matchAusDerPartie: 'Comunica il risultato dalla partita',
  matchAusDerPartieHilfe:
    'Punteggio e prova provengono dalla partita conclusa — devi solo indicare chi comunica.',
  matchNachweis: 'Prova (ultimo anello della catena)',
  matchUebernehmen: 'Registra il risultato',
  matchUebernommen: 'Il risultato è nel torneo.',
  matchGrenze:
    'La prova dice che entrambe le parti hanno visto lo stesso svolgimento. Non dice che si sia giocato correttamente — e un turno privato non è un’omologazione ufficiale di torneo.',
}
