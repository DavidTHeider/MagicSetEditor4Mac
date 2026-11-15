// **The script console** — `seiten.karten.konsole`, split out because
// `seitenkarten.ts` reached the line limit (§2, U13 of phase 5.1, #1424).
export const konsole = {
  ueberschrift: 'Console degli script',
  erklaerung:
    'Un’espressione MSE script nel contesto della carta aperta. L’ambiente contiene gli stessi nomi usati durante il disegno: “card”, “set”, “stylesheet” e le funzioni del pacchetto di gioco. Non modifica nulla.',
  ausdruck: 'Espressione',
  platzhalter: 'card.name',
  auswerten: 'Valuta',
  ergebnis: 'Risultato',
  art: 'Tipo: {art}',
  keineKarte: 'Per questo deve essere aperta una carta.',
  zuDenKarten: 'Vai alle carte',
  vorbehalte: 'Riserve durante la creazione dell’ambiente',
  stelle: '(riga {zeile}, colonna {spalte})',
  namen: 'Nomi disponibili',
  namenErklaerung:
    'Ciò che un’espressione può nominare su questa carta — la libreria e ciò che aggiunge il modello. Un clic inserisce il nome.',
  beispiele: 'Esempi',
  beispieleImmer: 'Funzionano ovunque',
  beispieleImmerErklaerung:
    'Un clic inserisce l’espressione. Questi non richiedono altro che il linguaggio stesso.',
  beispieleVorlage: 'Richiedono un modello',
  beispieleVorlageErklaerung:
    '“card” e “set” vengono inseriti nell’ambiente dal pacchetto di stile. Se per questa carta non ne è installato nessuno, la risposta lo dice — gli esempi qui sopra funzionano comunque.',
  beispielListe: {
    rechnung: 'Aritmetica',
    grossSchreiben: 'Una stringa in maiuscolo',
    laenge: 'Quanti caratteri ha qualcosa',
    ausschnitt: 'Un pezzo di una stringa',
    bedingung: 'Una condizione',
    name: 'Il nome di questa carta',
    grossName: 'Il suo nome in maiuscolo',
    regeltext: 'La lunghezza del suo testo delle regole',
    setangabe: 'Il titolo del set',
  },
  vervollstaendigung: 'Suggerimenti',
  uebernehmen: 'Accetta',
  namenFiltern: 'Restringi i nomi',
  namenGezeigt:
    'Ne sono mostrati {gezeigt} su {ganz} — scrivi per restringere.',
  namenLeer: 'Nessun nome corrisponde.',
  kuerzel: '⌘↩ valuta.',
  abbrechen: 'Annulla',
  laeuft: 'In corso …',
  dauer: '{ms} ms',
  abgebrochen:
    'Annullato — l’esecuzione è terminata prima che ci fosse un risultato. L’espressione è ancora nel campo.',
  baum: 'Risultato come albero',
  verlauf: 'Cronologia di questa sessione',
  verlaufErklaerung:
    'Le ultime venti espressioni, solo in questa sessione. Un clic ricarica l’espressione; ciò che suona come un token o una password qui è oscurato e non si può ricaricare.',
  verlaufLeer: 'Ancora nulla valutato.',
  verlaufLoeschen: 'Cancella la cronologia',
  ausgangGelungen: 'riuscita',
  ausgangGescheitert: 'non riuscita',
  ausgangAbgebrochen: 'annullata',
}
