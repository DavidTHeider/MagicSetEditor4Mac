// **The script console** — `seiten.karten.konsole`, split out because
// `seitenkarten.ts` reached the line limit (§2, U13 of phase 5.1, #1424).
export const konsole = {
  ueberschrift: 'Console de script',
  erklaerung:
    'Une expression MSE Script dans le contexte de la carte ouverte. L’environnement contient les mêmes noms que lors du dessin : « card », « set », « stylesheet » et les fonctions du paquet de jeu. Elle ne modifie rien.',
  ausdruck: 'Expression',
  platzhalter: 'card.name',
  auswerten: 'Évaluer',
  ergebnis: 'Résultat',
  art: 'Type : {art}',
  keineKarte: 'Une carte doit être ouverte pour cela.',
  zuDenKarten: 'Aller aux cartes',
  vorbehalte: 'Réserves lors de la construction de l’environnement',
  stelle: '(ligne {zeile}, colonne {spalte})',
  namen: 'Noms disponibles',
  namenErklaerung:
    'Ce qu’une expression peut nommer sur cette carte — la bibliothèque et ce que le modèle ajoute. Un clic insère le nom.',
  beispiele: 'Exemples',
  beispieleImmer: 'Fonctionnent partout',
  beispieleImmerErklaerung:
    'Un clic insère l’expression. Celles-ci n’ont besoin de rien d’autre que du langage lui-même.',
  beispieleVorlage: 'Nécessitent un modèle',
  beispieleVorlageErklaerung:
    '« card » et « set » sont placés dans l’environnement par le paquet de style. Si aucun n’est installé pour cette carte, la réponse l’indique — les exemples ci-dessus fonctionnent quand même.',
  beispielListe: {
    rechnung: 'Calcul',
    grossSchreiben: 'Une chaîne en majuscules',
    laenge: 'Combien de caractères compte quelque chose',
    ausschnitt: 'Un morceau de chaîne',
    bedingung: 'Une condition',
    name: 'Le nom de cette carte',
    grossName: 'Son nom en majuscules',
    regeltext: 'La longueur de son texte de règles',
    setangabe: 'Le titre du set',
  },
  vervollstaendigung: 'Suggestions',
  uebernehmen: 'Accepter',
  namenFiltern: 'Filtrer les noms',
  namenGezeigt: '{gezeigt} sur {ganz} affichés — saisis du texte pour affiner.',
  namenLeer: 'Aucun nom ne correspond.',
  kuerzel: '⌘↩ évalue.',
  abbrechen: 'Annuler',
  laeuft: 'En cours …',
  dauer: '{ms} ms',
  abgebrochen:
    'Annulé — l’exécution s’est terminée avant qu’un résultat soit disponible. L’expression est toujours dans le champ.',
  baum: 'Résultat sous forme d’arbre',
  verlauf: 'Historique de cette session',
  verlaufErklaerung:
    'Les vingt dernières expressions, uniquement dans cette session. Un clic recharge l’expression ; tout ce qui ressemble à un jeton ou à un mot de passe est masqué ici et ne peut pas être rechargé.',
  verlaufLeer: 'Rien n’a encore été évalué.',
  verlaufLoeschen: 'Effacer l’historique',
  ausgangGelungen: 'réussie',
  ausgangGescheitert: 'échouée',
  ausgangAbgebrochen: 'annulée',
}
