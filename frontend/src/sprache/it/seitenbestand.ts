// The texts for **database maintenance** — back up and restore
// (#940).
//
// Split out of `seiten.ts` the same way the card page was: one catalogue may
// live in several files as long as it arrives as **one** (§9).

export const bestandspflege = {
  titel: 'Archivio',
  hinweis:
    'Tutto ciò che questa applicazione gestisce: set, carte, valori dei campi, la cronologia e le immagini inserite. Un backup contiene entrambi in un unico file — l’archivio e le immagini.',
  sichern: 'Esegui il backup dell’archivio …',
  waehlen: 'Ripristina un backup …',
  rueckfrage: 'Ripristinare questo backup?',
  datei: 'File',
  angelegtAm: 'Archivio creato il',
  sets: 'Set',
  karten: 'Carte',
  unbekannt: '—',
  warnung:
    'L’archivio attuale verrà sostituito. Non va perso: resta accanto come bestand-vor-rueckspielung.db — fino al prossimo ripristino.',
  bilderLage: 'Immagini inserite',
  bilderErsetzt:
    'Questo backup contiene immagini. Prendono il posto di quelle esistenti; quelle precedenti restano accanto come beiwerk-vorher.',
  bilderGeleert:
    'Questo backup non contiene immagini — proviene da uno stato in cui non ce n’erano di inserite. Quelle esistenti vengono rimosse; restano accanto come beiwerk-vorher.',
  bilderUnberuehrt:
    'Questo backup è un singolo database della 0.1.0 o della 0.2.0 e non contiene immagini. Quelle esistenti restano invariate — su un altro computer mancherebbero del tutto.',
  bestaetigen: 'Ripristina',
  abbrechen: 'Annulla',

  // **The password of a backup** (OPS-06, #1723).
  kennwort: 'Password (facoltativa)',
  kennwortWiederholen: 'Ripeti la password',
  kennwortHinweis:
    'Con una password il backup è cifrato e può essere aperto anche su un altro computer. Una password dimenticata non può essere recuperata — il backup andrebbe allora perso.',
  kennwortVerschieden: 'Le due password non coincidono.',
  klartext: 'Esegui il backup senza password — il file sarà in chiaro',
  klartextHinweis:
    'L’archivio è cifrato. Senza password il backup finirebbe senza protezione sul supporto su cui viene salvato.',
  gesperrt: 'Questo backup è cifrato.',
  gesperrtKennwort: 'Password del backup',
  entsperren: 'Apri',
  verschluesselt: 'Cifrato',
  ja: 'sì',
  nein: 'no',
  paketeEnthalten: 'Pacchetti propri inclusi',
  paketeFehlen: 'Da scaricare di nuovo',
  keine: 'nessuno',
  zurueckgespielt:
    'Ripristinato. L’archivio precedente si trova ora in {beiseite}.',
  rueckkehrUnveraendert: 'Il tuo archivio precedente è invariato.',
  rueckkehrUnvollstaendig:
    'Anche il ripristino all’indietro non è riuscito: parti dell’archivio precedente si trovano accanto come bestand-vor-rueckspielung.db e beiwerk-vorher. Ripristina un backup o rimettile al loro posto a mano.',
  wiederherstellungTitel: 'Impossibile aprire l’archivio',
  wiederherstellungHinweis:
    'L’applicazione tornerà a funzionare solo quando sarà presente un file di archivio utilizzabile. Ripristinare un backup è la strada per arrivarci — sostituisce il file che sta dando errore.',
  wiederherstellungGrund: 'Che cosa è stato segnalato',
  wiederherstellungHinweisHeil:
    'Il file della libreria probabilmente è a posto — il problema è in ciò che lo circonda. Risolvilo come descritto sotto e riprova; non c’è nulla da sostituire.',
  erneutVersuchen: 'Riprova',
  weiterGescheitert: 'Nuovo tentativo — la libreria non si apre ancora.',

  wegZurueckTitel: 'La strada per tornare indietro',
  wegZurueck:
    'Per tornare indietro usa “{datei}” nella cartella del tuo archivio — {sets} set, {karten} carte, versione dello schema {stand}. Selezionalo qui sotto.',

  chiffreTitel: 'Cifratura dell’archivio',
  chiffreAnHinweis:
    'L’archivio è salvato su disco in forma cifrata. Senza la chiave non può essere aperto — nemmeno da te.',
  chiffreAusHinweis:
    'L’archivio è salvato su disco in chiaro. Chi ha il file può leggere tutto ciò che contiene: nomi, date di nascita, indirizzi, foto e metodi di accesso.',
  chiffreAnschalten: 'Attiva la cifratura …',
  chiffreAusschalten: 'Disattiva la cifratura …',
  chiffreFrageAn: 'Cifrare ora l’archivio?',
  chiffreFrageAus: 'Disattivare di nuovo la cifratura?',
  chiffreWarnungAn:
    'La chiave viene creata adesso e finisce nel portachiavi di questo Mac. Non può essere ricalcolata e non esiste da nessun’altra parte: se va persa, vanno persi anche i dati. Prima fai un backup.',
  chiffreWarnungAus:
    'Poi tutto tornerà in chiaro e la chiave verrà rimossa dal portachiavi.',
  chiffreGrenzen:
    'La cifratura **non** protegge dalle manomissioni. Chi confronta più copie dello stesso file vede quali parti sono cambiate. **Un backup ha una password propria** — qui sopra, al momento del backup; senza password sarebbe in chiaro, e questo richiede allora un esplicito “Esegui il backup senza password”.',
  chiffreLaeuft: 'Riscrittura dell’archivio in corso …',

  platzUeberschrift: 'Spazio su disco',
  platzErklaerung:
    'Ciò che questa applicazione salva e quanto pesa — l’intera cartella, perché ciò che non appartiene a nessuna voce viene contato sotto **Altro**. I **backup prima delle migrazioni** vengono creati a ogni avvio che ricostruisce l’archivio; resta la **strada per tornare indietro** — il backup più recente che questa versione può aprire. Del **registro** resta il file in uso.',
  platzDateien: '{anzahl} file',
  platzFreigeben: 'Elimina i più vecchi',
  platzFreigebenFuer: '{posten}: elimina i più vecchi',
  platzFreigebenFrage: '{posten}: eliminare davvero i più vecchi?',
  platzFreigebenErklaerung:
    'I file più vecchi vengono eliminati e non possono essere recuperati. Ciò che serve come strada per tornare indietro resta.',
  platzLoeschenJa: 'Elimina',
  platzNamen: {
    bestand: 'Archivio',
    bilder: 'Immagini inserite',
    pakete: 'Pacchetti scaricati',
    vorsicherungen: 'Backup prima delle migrazioni',
    protokoll: 'Registro',
    sonstiges: 'Altri file nella cartella dell’applicazione',
  },
  alter: {
    ueberschrift: 'Ultimo backup',
    unbekannt: 'sconosciuto',
    unbekanntSatz:
      'Questa versione non ha mai fatto un backup della libreria. Sconosciuto significa sconosciuto — qui non compare mai «oggi» senza che un file sia stato riletto.',
    heute: 'oggi',
    gestern: 'un giorno fa',
    vorTagen: '{tage} giorni fa',
    alt: '{alter} · più di 14 giorni',
    unstimmig: 'ora dopo l’orologio',
    satz: 'Ultimo backup il {zeitpunkt} tramite {flaeche} in {datei}, riletto e verificato.',
    satzPlan:
      'Ultimo backup il {zeitpunkt} secondo il piano in {datei}, riletto e verificato.',
    unstimmigSatz:
      'L’ultimo backup porta la data {zeitpunkt}, successiva all’orologio di questo Mac. La sua età non si può dire; fai un backup per sicurezza.',
    flaechen: {
      fenster: 'la finestra',
      cli: 'la riga di comando',
      mcp: 'MCP',
      rest: 'il servizio',
    },
    planSchalter: 'Backup settimanale in una cartella, conservarne sette',
    planHilfe:
      'Solo dopo aver scelto una cartella. Vengono eliminati solo i suoi backup, mai altri file. Salva finché la finestra è aperta.',
    planOrdner: 'Cartella: {ordner}',
    planZuletzt: 'Ultimo backup secondo il piano il {zeitpunkt}.',
    planNie: 'Il piano non ha ancora fatto backup.',
    planFehler: 'Il piano non è riuscito a fare l’ultimo backup: {grund}',
  },
} as const
