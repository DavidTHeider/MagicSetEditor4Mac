// The texts of the **set list** — open, create, import.
//
// Split out of `seiten.ts` when the file grew past 400 lines (§2).

export const sets = {
  auswahlLoesen: 'Mostra tutti i set',
  erklaerung:
    'Ogni set aperto si trova nell’archivio. Da qui raggiungi le sue carte — oppure lo riscrivi in un file.',
  titel: 'Set',
  leer: 'Nessun set aperto finora. Apri un file .mse-set per iniziare.',
  setOeffnen: 'Apri set …',
  neuTitel: 'Un nuovo set',
  neuKopf: 'Nuovo set',
  einlesenTitel: 'Importa un set da un’altra fonte',
  neuHinweis:
    'Un set vuoto — gioco e stile decidono l’aspetto delle sue carte. Lo stile si può cambiare in seguito.',
  neuName: 'Nome del set',
  neuStil: 'Stile per il nuovo set',
  neuAnlegen: 'Crea set',
  neuSchritte: 'Passaggi per un nuovo set',
  neuSchrittArt: 'Tipo',
  neuSchrittAussehen: 'Aspetto',
  neuSchrittName: 'Nome',
  neuArtFrage:
    'A cosa servono le carte? Il nome del pacchetto lo cerchiamo noi per te.',
  neuArtZusatz: '· {stile} disponibili',
  neuEigenes: 'Qualcosa di tutto tuo — crea un pacchetto di gioco',
  neuWeiter: 'Avanti',
  neuSchrittZurueck: 'Indietro',
  neuUeberspringen: 'Salta',
  neuWahl: '{spiel} · {stil}',
  neuOhneSpiele:
    'Non è installato alcun pacchetto di gioco. Senza non ci sono campi per le carte — i pacchetti arrivano dalla pagina dei pacchetti.',
  neuOhneStile:
    'Per {spiel} non è installato alcun pacchetto di stile. Senza uno stile non si può creare alcun set — i pacchetti arrivano dalla pagina dei pacchetti.',
  neuZuDenPaketen: 'Vai ai pacchetti',
  neuZurueck: 'Premi ⌘[ per tornare qui; ciò che hai inserito è ancora lì.',
  beispielsets: 'Inclusi per provare',
  geoeffnet: 'aperto',
  schonOffen: 'Questo file è già aperto — eccolo.',
  teilerfolg:
    'Il file non è stato letto fino in fondo. {anzahl} set sono comunque nell’archivio — si trovano nell’elenco qui sotto.',
  beimOeffnenBefunde:
    'Non leggibile durante l’apertura. Il set c’è comunque; mancano questi punti:',
  // **The preview before reading in** — FMT-03.AK-01 (#1692). It appears
  // only when there is something to say.
  vorschau: {
    ueberschrift: 'Prima dell’importazione: {name}',
    umfang: '{karten} · gioco {spiel} · stile {stil}',
    umfangOhneStil: '{karten} · gioco {spiel}',
    schonOffen:
      'Questo file è già nel tuo archivio — Apri mostra il set esistente invece di aggiungerne un secondo.',
    befunde:
      'Questi punti non si possono leggere; il set arriverebbe senza di essi:',
    bedarf:
      'Mancano dei pacchetti. Puoi importarlo comunque — le carte vengono disegnate con un sostituto finché i pacchetti non ci sono:',
    einlesen: 'Importa',
    aufschlagen: 'Apri',
    verwerfen: 'Non importare',
  },
  // **Deletion** (#1218) — final, confirmed by name.
  loeschen: 'Elimina …',
  loeschenTitel: 'Elimina questo set',
  loeschenWarnung:
    '{name} scompare dall’archivio, insieme a {karten}. Non c’è cestino né modo di annullare. Ciò che si trova in un file resta lì — viene eliminata la copia nell’archivio.',
  loeschenSammlung:
    'La tua collezione resta: un esemplare appartiene a un’edizione, non a un set.',
  loeschenFeld: 'Scrivi il nome del set per confermare',
  loeschenPasstNicht: 'Il nome non corrisponde ancora a {name}.',
  loeschenJa: 'Elimina definitivamente',
  loeschenNein: 'Annulla',
  loeschenBefunde:
    'Il set non c’è più. Questi residui non si sono potuti rimuovere:',

  spalten: {
    name: 'Nome',
    spiel: 'Gioco',
    karten: 'Carte',
    herkunft: 'Origine',
  },
  mtgjsonUeberschrift: 'MTG JSON',
  mtgjsonErklaerung:
    'Definizioni di carte da mtgjson.com — un singolo file o una cartella. A quali campi delle carte corrispondano dipende dal gioco.',
  mtgjsonSpiel: 'Gioco',
  mtgjsonDatei: 'Apri file …',
  mtgjsonVerzeichnis: 'Apri cartella …',
  mtgjsonKeinSpiel:
    'Nessun gioco ha una corrispondenza per i campi MTG JSON. Senza, non si può importare nulla.',
  scryfallsatzUeberschrift: 'Un set da Scryfall',
  scryfallsatzErklaerung:
    'Il codice del set decide quali carte arrivano — con i loro testi, senza immagini. Vengono letti nome, costo di mana, tipi, testo delle regole, flavor text, forza, costituzione, rarità e illustratore; le immagini si scaricano carta per carta. Lo stile si sceglie dopo.',
  scryfallsatzKeinSpiel:
    'Senza un gioco con una corrispondenza non si può importare alcun set.',
  scryfallsatzSpiel: 'Gioco per la corrispondenza',
  scryfallsatzCode: 'Codice del set su Scryfall',
  scryfallsatzPlatzhalter: 'ad es. lea',
  scryfallsatzEinlesen: 'Importa set',
  vorlageAusfuhr: 'Tramite un modello di esportazione …',
  vorlageWaehlen: 'Modello di esportazione',
  vorlageAusfuehren: 'Esegui',
  vorlageSchliessen: 'Chiudi',
  vorlageKeine:
    'Per questo gioco non è installato alcun modello di esportazione. I modelli arrivano dai pacchetti scaricati.',
  vorlageGeschrieben: 'Scritto in',
  vorlageBefunde:
    'Durante l’esecuzione è emerso qualcosa — il file è stato comunque scritto:',
  vorlageGescheitert:
    'Questo modello non viene eseguito fino in fondo. Ciò che gli manca è indicato qui sotto:',
  mtgjsonSichern: 'Salva come MTG JSON …',
  mtgjsonGesichert: 'carte salvate in',
  mtgjsonVerluste:
    'Non esprimibile in uscita. Il file c’è; mancano questi dettagli:',
  mtgjsonSpalte: 'Esportazione',
  speichern: 'Salva',
  speichernUnter: 'Salva come …',
  speicherSpalte: 'Salva',
  gespeichert: 'Salvato in',
  beimSpeichernBefunde:
    'Non incluso durante il salvataggio. Il file del set è stato scritto; mancano questi file accanto ad esso:',
  oeffnen: 'Mostra le carte di questo set',
  handlungenAn: 'Azioni — {name}',
  handlungenKurz: '…',
  vorbehalte: {
    mitZahl: 'Non interpretato ({anzahl})',
    ueberschrift: 'Non interpretato',
    erklaerung:
      'L’importazione non ha capito queste righe. Vengono mantenute invariate e non andranno perse al salvataggio.',
    keine: 'Nulla — questo set è stato letto per intero.',
  },

  // **The portable envelope** — work item N4 of phase 5.4
  // (#1457).
  umschlagAblegen: 'Salva come busta',
  umschlagGeschnuert:
    'Salvato. La busta è firmata — chi la riceve può verificare che provenga da qui.',
  umschlagTitel: 'Accogli una busta',
  umschlagErklaerung:
    'Una busta è un set salvato da qualcun altro. Prima viene esaminata e poi accolta — prima di allora nell’archivio non cambia nulla.',
  umschlagWaehlen: 'Esamina una busta …',
  umschlagUrheber: 'Firmata da',
  umschlagAbdruckHinweis:
    'L’impronta dice CHI ha firmato — non se ti puoi fidare di lui. Se conosci la persona, confrontala con lei.',
  umschlagInhalt: 'Contiene',
  umschlagUebernehmen: 'Accogli',
  umschlagVerwerfen: 'Lascia stare',
  umschlagUebernommen:
    'Accolta. Il set è stato creato da zero — ha identificativi propri e non sovrascrive nulla.',
}
