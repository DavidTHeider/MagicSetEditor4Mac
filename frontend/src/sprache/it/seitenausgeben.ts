// The texts of the **export assistant** — where to, and what each
// destination needs.
//
// Split out of `seiten.ts` when the assistant brought the file close to the
// 400-line limit (§2, UXR-09, #1917).

export const ausgeben = {
  titel: 'Esporta',
  erklaerung:
    'Cinque modi per portare un set fuori da questa applicazione — e a cosa serve ciascuno.',
  wohin: 'Dove?',
  keinZiel: 'Nessuna destinazione scelta',
  nichtGemerkt:
    'La destinazione vale, ma non è stata memorizzata — la prossima volta non sarà preselezionata:',
  keinZielErklaerung:
    'Scegli in «Dove?» in quale forma deve uscire il set — qui comparirà ciò che c’è da impostare.',
  ohneKarten:
    'Immagini e foglio di stampa non avrebbero nulla da mostrare. Le carte nascono in «Carte»: create, incollate o importate.',
  ziele: {
    setdatei: {
      name: 'File di set',
      kurz: 'Il set stesso, da riaprire — un file `.mse-set` che legge anche Magic Set Editor 2.',
      lang: 'La scorciatoia è ⌘S. **Non inclusi** sono mazzi, copie, prezzi, etichette e la provenienza dei singoli valori — restano nella libreria, perché MSE2 scarta al salvataggio ciò che non conosce.',
    },
    bilder: {
      kurz: 'Un file immagine per ogni carta, da usare altrove — un forum, un foglio di calcolo, un programma di gioco.',
    },
    druckbogen: {
      kurz: 'Le carte affiancate su carta, con i segni di taglio.',
    },
    vorlage: {
      kurz: 'Testo invece di immagini: un elenco, una tabella, un file per un altro programma.',
      lang: 'Quali modelli esistono dipende dal pacchetto installato.',
    },
    mtgjson: {
      kurz: 'Il formato delle carte di mtgjson.com — per gli strumenti che lo leggono.',
      lang: 'Non per continuare a lavorare in questa applicazione: ciò che MTG JSON non può contenere viene indicato dopo l’esportazione.',
    },
  },
  vorlagenEingrenzen: 'Filtra i modelli',
  weitereOptionen: 'Altre opzioni ({zahl})',
  // **„Come verrà"** — der Schritt vor dem Schreiben (UXR-F7, #2015).
  probelauf: {
    ueberschrift: 'Come verrà',
    rechnet: 'Calcolo in corso…',
    hinweis: 'Non viene scritto nulla finché non premi il pulsante qui sotto.',
    bogen: {
      jeSeite: '{zahl} per pagina · {seiten}',
      masse: '{karten} · carta {breite} × {hoehe} mm · {dpi} dpi',
      letzte: 'Nell’ultima pagina: {karten}.',
    },
    vorlage: {
      datei: 'File: {name} ({groesse})',
      karten: 'Valutato su {karten}.',
      beigaben: 'File accanto: {zahl} (cartella «{ordner}», {groesse})',
      mindestens: 'almeno {groesse}',
      befunde: 'Rilevato durante la valutazione:',
    },
  },
}
