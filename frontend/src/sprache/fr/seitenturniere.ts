// Die Texte der **Turniere** — Werkstück 5 von Phase 4.7
// (#1294),
// nach Vermerk 0062.
//
// **Ein eigener Baustein**, wie `seitendecks.ts` und aus demselben Grund:
// `seiten.ts` steht an der Zeilengrenze aus §2.
export const turniere = {
  alleTurniere: 'Tous les tournois ({anzahl})',
  titel: 'Tournois',
  erklaerung:
    'Un tournoi est un déroulement dans le temps : rondes, appariements, résultats. Le classement en est calculé et n’est enregistré nulle part — corrige un résultat et tu vois aussitôt le nouveau classement.',

  leer: 'Aucun tournoi pour l’instant. « Créer un tournoi … », en haut, en commence un ; « Importer un tournoi … » en reprend un depuis un fichier.',

  neu: 'Créer un tournoi',
  neuOeffnen: 'Créer un tournoi …',
  name: 'Nom',
  spiel: 'Jeu',
  notiz: 'Note',
  einzelspiel: 'Une partie par match',
  einzelspielHinweis:
    'Alors le pourcentage de victoires en partie et celui des adversaires ne disent rien — ils recopient les chiffres du match. Ils sont tout de même calculés.',
  anlegen: 'Créer',

  // **The wizard** — DOM-04.AK-01 (#1699).
  assistent: 'Un nouveau tournoi en quatre étapes',
  schrittTurnier: 'Tournoi',
  schrittRegeln: 'Règles',
  schrittTeilnehmende: 'Participants',
  schrittErsteRunde: 'Première ronde',
  weiter: 'Suivant',
  zurueck: 'Retour',
  verlassen: 'Continuer plus tard',
  vorlage: 'Modèle',
  ohneVorlage: 'Sans modèle',
  vorlageHinweis:
    'Un modèle apporte jeu, format, règles et note. Les noms aussi si tu le souhaites, jamais les decks : une liste soumise vaut pour un seul tournoi.',
  mitTeilnehmenden: 'Reprendre les noms des participants',
  nameWarum:
    'Obligatoire — pour le retrouver dans la liste. Deux tournois d’un même jeu portent des noms différents.',
  spielWarum:
    'Obligatoire — il détermine quels formats et quels decks conviennent.',
  format: 'Format',
  ohneFormat: 'Aucun format',
  formatWarum: 'Facultatif — tu peux choisir parmi les formats de ce jeu.',
  regelnAusVorlage: 'Les règles proviennent du modèle.',
  schweizerSystem:
    'Les appariements suivent le système suisse : les joueurs ayant le même nombre de points s’affrontent, et personne ne rencontre deux fois le même adversaire. Le programme propose le nombre de rondes d’après le nombre de participants.',
  zuWenige:
    'Une ronde nécessite deux personnes qui s’affrontent — inscris-en au moins deux.',
  ersteRundeHinweis:
    'Voici à quoi ressemblerait la première ronde. Rien n’est apparié tant que tu n’appuies pas sur le bouton.',
  ersteRundeAuslosen: 'Apparier la ronde 1',

  waehlen: 'Ouvrir',
  loeschen: 'Supprimer le tournoi',
  handlungenAn: 'Actions — {name}',
  handlungenKurz: '…',
  turniermenue: 'Tournoi',
  loeschenFrage:
    'Supprimer ce tournoi ? Les rondes, appariements et résultats disparaissent avec lui. Les decks des participants sont conservés.',
  loeschenJa: 'Supprimer',
  loeschenNein: 'Annuler',

  teilnehmer: 'Participants',
  teilnehmerLeer: 'Personne d’inscrit pour l’instant.',
  anmelden: 'Inscrire',
  gebiete: 'Rubriques de ce tournoi',
  gebietTeilnehmer: 'Participants',
  gebietRunden: 'Rondes',
  gebietRangliste: 'Classement',

  // **Closing and log** — DOM-04.AK-04 (#1699).
  abschliessen: 'Clore',
  abgeschlossenAm:
    'Clos le {datum}. Les participants, rondes et résultats sont définitifs.',
  wiedereroeffnen: 'Rouvrir',
  nichtWiedereroeffnen: 'Ne pas rouvrir',
  wiedereroeffnenFrage: 'Rouvrir ce tournoi ?',
  wiedereroeffnenErklaerung:
    'Ensuite, les participants, rondes et résultats pourront de nouveau être modifiés. Le journal consigne que cela a eu lieu.',
  gebietProtokoll: 'Journal',
  protokollLeer: 'Il ne s’est encore rien passé.',
  protokollAngelegt: 'Créé sous le nom « {angaben} »',
  protokollAngemeldet: '{angaben} inscrit',
  protokollAusgestiegen: '{angaben} a abandonné',
  protokollEingestiegen: '{angaben} de nouveau inscrit',
  protokollRunde: 'Ronde {angaben} appariée',
  protokollErgebnis: 'Résultat {paarung} : {angaben}',
  protokollZurueckgenommen: 'Résultat {paarung} annulé',
  protokollAbgeschlossen: 'Clos',
  protokollWiedereroeffnet: 'Rouvert',
  protokollEingelesen: 'Importé sous le nom « {angaben} »',

  // **File and report** — DOM-04.AK-03 (#1699).
  ausgeben: 'Exporter …',
  berichtSichern: 'Enregistrer le rapport …',
  einlesen: 'Importer un tournoi …',
  gesichertUnter: 'Enregistré dans {pfad}.',
  eingelesen: 'Importé et ouvert.',
  einlesebefunde: 'À signaler lors de l’importation :',

  // **Header and links after the artboard** — DOM-04.AK-02 (#1699).
  kopfzeile: 'Système suisse · Ronde {runde} sur {von} · {anzahl} participants',
  kopfzeileOhneRunde: 'Système suisse · {anzahl} participants',
  kopfzeileOhneEmpfehlung:
    'Système suisse · Ronde {runde} · {anzahl} participants',
  gebietDecklisten: 'Listes de deck',
  decklistenLeer:
    'Personne n’a encore soumis de liste — choisir un deck à l’inscription soumet sa liste.',
  deckAufschlagen: 'Ouvrir le deck',
  partieOeffnen: 'Ouvrir la partie',
  deck: 'Deck',
  ohneDeck: 'Sans deck',
  deckHinweis:
    'Si un deck est indiqué, sa liste est copiée et non référencée : une liste de deck dans un tournoi est un document soumis.',
  deckabzug: 'Liste soumise',
  deckabzugLeer: 'Aucun deck soumis.',
  deckabzugUnbenannt: 'Carte',
  aussteigen: 'Abandonner',
  einsteigen: 'Réinscrire',
  ausgestiegen: 'a abandonné',
  ausgestiegenHinweis:
    'Il n’est plus apparié et reste dans le classement : son pourcentage continue de compter pour celui de ses adversaires.',

  runden: 'Rondes',
  rundenLeer: 'Aucune ronde appariée pour l’instant.',
  runde: 'Ronde {nummer}',
  empfehlung:
    '{gelegt} rondes sur {empfohlen} — les règles de tournoi en recommandent {empfohlen} pour {spieler} participants. Une suggestion, pas une obligation.',
  vorschlag: 'Voir la ronde suivante',
  vorschlagTitel: 'Voici à quoi ressemblerait la ronde suivante',
  vorschlagHinweis:
    'Aperçu uniquement, rien n’est créé. Seul le bouton ci-dessous la crée.',
  rundeLegen: 'Créer la ronde',
  bye: 'Bye (2–0)',
  gegen: 'contre',
  offen: 'pas encore jouée',
  spiele: 'Parties',
  spieleA: 'Parties du premier',
  spieleB: 'Parties du second',
  unentschiedenSpiele: 'Égalités',
  eintragen: 'Enregistrer',
  zuruecknehmen: 'Annuler',
  ergebnis: 'Résultat',

  // **The effects of a correction** — DOM-04.AK-02 (#1699).
  folgenFrage: 'Cette correction modifie le classement.',
  folgenPlatz:
    '{name} : place {vorher} → {nachher}, {punkteVorher} → {punkteNachher} points',
  folgenSpaeter:
    'La ronde {runden} a été appariée d’après l’ancien classement et reste telle quelle.',
  folgenEintragen: 'Enregistrer quand même',
  folgenNicht: 'Ne pas enregistrer',
  ergebnisZahlen:
    'Les parties se comptent en nombres entiers à partir de zéro.',
  ergebnisOhneSpiel:
    'Un résultat sans aucune partie n’est pas un résultat — qui n’a pas joué n’en a pas encore.',
  ergebnisZuViele: 'Un match comporte ici au plus {hoechstens} parties.',

  rangliste: 'Classement',
  ranglisteLeer: 'Rien n’a encore été joué.',
  platz: 'Place',
  bilanz: 'Bilan',
  matchpunkte: 'Points',
  gegnermatchquote: 'OMW',
  spielquote: 'GW',
  gegnerspielquote: 'OGW',
  quotenHinweis:
    'Quatre critères de départage dans l’ordre des Magic Tournament Rules : points, pourcentage de victoires en match des adversaires, pourcentage de victoires en partie, pourcentage de victoires en partie des adversaires. Chaque pourcentage a un plancher de 0,33.',

  // **Eine Paarung, an eine Partie gebunden** (G7, #1464).
  match: 'Joué ici',
  matchErklaerung:
    'Si cet appariement est joué dans ce programme, lie-le à la partie. Le résultat est alors déclaré par les deux côtés et n’est enregistré que lorsque les deux disent la même chose.',
  matchNichtGebunden: 'Aucune partie n’est rattachée à cet appariement.',
  matchAnsehen: 'Consulter',
  matchPartie: 'Identifiant de la partie',
  matchSitzA: 'Place du premier joueur',
  matchSitzB: 'Place du second joueur',
  matchBinden: 'Lier à cette partie',
  matchGebunden: 'Lié à la partie',
  matchSitz: 'Place',
  matchDeckabdruck: 'Empreinte du deck',
  matchKeineMeldung: 'Personne n’a encore déclaré.',
  matchMeldungVon: 'déclaré par',
  matchMelden: 'Déclarer pour cette place',
  matchAusDerPartie: 'Déclarer le résultat depuis la partie',
  matchAusDerPartieHilfe:
    'Le score et la preuve proviennent de la partie terminée — tu indiques seulement qui déclare.',
  matchNachweis: 'Preuve (dernier maillon de la chaîne)',
  matchUebernehmen: 'Enregistrer le résultat',
  matchUebernommen: 'Le résultat figure dans le tournoi.',
  matchGrenze:
    'La preuve indique que les deux côtés ont vu le même déroulement. Elle n’indique pas que la partie a été jouée correctement — et une ronde privée n’est pas une homologation officielle de tournoi.',
}
