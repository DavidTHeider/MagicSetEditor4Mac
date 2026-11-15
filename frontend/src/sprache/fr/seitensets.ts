// The texts of the **set list** — open, create, import.
//
// Split out of `seiten.ts` when the file grew past 400 lines (§2).

export const sets = {
  auswahlLoesen: 'Afficher tous les sets',
  erklaerung:
    'Chaque set ouvert se trouve dans la base de données. D’ici, tu accèdes à ses cartes — ou tu le réécris dans un fichier.',
  titel: 'Sets',
  leer: 'Aucun set ouvert pour l’instant. Ouvre un fichier .mse-set pour commencer.',
  setOeffnen: 'Ouvrir un set …',
  neuTitel: 'Un nouveau set',
  neuKopf: 'Nouveau set',
  einlesenTitel: 'Importer un set d’ailleurs',
  neuHinweis:
    'Un set vide — le jeu et le style déterminent l’apparence de ses cartes. Le style peut être changé plus tard.',
  neuName: 'Nom du set',
  neuStil: 'Style du nouveau set',
  neuAnlegen: 'Créer le set',
  neuSchritte: 'Étapes vers un nouveau set',
  neuSchrittArt: 'Type',
  neuSchrittAussehen: 'Apparence',
  neuSchrittName: 'Nom',
  neuArtFrage:
    'À quoi servent les cartes ? Nous cherchons le nom du paquet pour toi.',
  neuArtZusatz: '· {stile} disponibles',
  neuEigenes: 'Quelque chose de tout à fait personnel — créer un paquet de jeu',
  neuWeiter: 'Suivant',
  neuSchrittZurueck: 'Retour',
  neuUeberspringen: 'Passer',
  neuWahl: '{spiel} · {stil}',
  neuOhneSpiele:
    'Aucun paquet de jeu n’est installé. Sans paquet, il n’y a pas de champs de carte — les paquets viennent de la page des paquets.',
  neuOhneStile:
    'Aucun paquet de style n’est installé pour {spiel}. Sans style, aucun set ne peut être créé — les paquets viennent de la page des paquets.',
  neuZuDenPaketen: 'Aller aux paquets',
  neuZurueck:
    'Appuie sur ⌘[ pour revenir ici ; ce que tu as saisi y est toujours.',
  beispielsets: 'Fournis pour essayer',
  geoeffnet: 'ouvert',
  schonOffen: 'Ce fichier est déjà ouvert — le voici.',
  teilerfolg:
    'Le fichier n’a pas été lu jusqu’au bout. {anzahl} sets se trouvent malgré tout dans la base de données — ils figurent dans la liste ci-dessous.',
  beimOeffnenBefunde:
    'Illisible à l’ouverture. Le set est quand même présent ; ces éléments y manquent :',
  // **The preview before reading in** — FMT-03.AK-01 (#1692). It appears
  // only when there is something to say.
  vorschau: {
    ueberschrift: 'Avant l’importation : {name}',
    umfang: '{karten} · jeu {spiel} · style {stil}',
    umfangOhneStil: '{karten} · jeu {spiel}',
    schonOffen:
      'Ce fichier est déjà dans ta bibliothèque — Ouvrir affiche le set existant au lieu d’en ajouter un second.',
    befunde: 'Ces éléments sont illisibles ; le set arriverait sans eux :',
    bedarf:
      'Des paquets manquent. Tu peux quand même l’importer — les cartes sont dessinées avec un plan de remplacement tant que les paquets ne sont pas là :',
    einlesen: 'Importer',
    aufschlagen: 'Ouvrir',
    verwerfen: 'Ne pas importer',
  },
  // **Deletion** (#1218) — final, confirmed by name.
  loeschen: 'Supprimer …',
  loeschenTitel: 'Supprimer ce set',
  loeschenWarnung:
    '{name} disparaît de la base de données, avec {karten}. Il n’y a ni corbeille ni annulation possible. Ce qui se trouve dans un fichier y reste — c’est la copie dans la base de données qui est supprimée.',
  loeschenSammlung:
    'Ta collection reste : un exemplaire est rattaché à une impression, pas à un set.',
  loeschenFeld: 'Saisis le nom du set pour confirmer',
  loeschenPasstNicht: 'Le nom ne correspond pas encore à {name}.',
  loeschenJa: 'Supprimer définitivement',
  loeschenNein: 'Annuler',
  loeschenBefunde: 'Le set a disparu. Ces restes n’ont pas pu être nettoyés :',

  spalten: {
    name: 'Nom',
    spiel: 'Jeu',
    karten: 'Cartes',
    herkunft: 'Origine',
  },
  mtgjsonUeberschrift: 'MTG JSON',
  mtgjsonErklaerung:
    'Définitions de cartes de mtgjson.com — un fichier unique ou un dossier. Les champs de carte auxquels elles correspondent dépendent du jeu.',
  mtgjsonSpiel: 'Jeu',
  mtgjsonDatei: 'Ouvrir un fichier …',
  mtgjsonVerzeichnis: 'Ouvrir un dossier …',
  mtgjsonKeinSpiel:
    'Aucun jeu n’a de correspondance pour les champs MTG JSON. Sans elle, rien ne peut être importé.',
  scryfallsatzUeberschrift: 'Un set depuis Scryfall',
  scryfallsatzErklaerung:
    'Le code du set détermine quelles cartes arrivent — avec leurs textes, sans images. Le nom, le coût de mana, les types, le texte de règles, le texte d’ambiance, la force, l’endurance, la rareté et l’illustrateur sont repris ; les images se récupèrent carte par carte. Choisis le style ensuite.',
  scryfallsatzKeinSpiel:
    'Sans jeu disposant d’une correspondance, aucun set ne peut être importé.',
  scryfallsatzSpiel: 'Jeu pour la correspondance',
  scryfallsatzCode: 'Code du set sur Scryfall',
  scryfallsatzPlatzhalter: 'p. ex. lea',
  scryfallsatzEinlesen: 'Importer le set',
  vorlageAusfuhr: 'Via un modèle d’export …',
  vorlageWaehlen: 'Modèle d’export',
  vorlageAusfuehren: 'Exécuter',
  vorlageSchliessen: 'Fermer',
  vorlageKeine:
    'Aucun modèle d’export n’est installé pour ce jeu. Les modèles proviennent des paquets téléchargés.',
  vorlageGeschrieben: 'Écrit dans',
  vorlageBefunde:
    'Quelque chose a été relevé pendant l’exécution — le fichier a quand même été écrit :',
  vorlageGescheitert:
    'Ce modèle ne s’exécute pas jusqu’au bout. Ce qui lui manque figure ci-dessous :',
  mtgjsonSichern: 'Enregistrer en MTG JSON …',
  mtgjsonGesichert: 'cartes enregistrées dans',
  mtgjsonVerluste:
    'Non exprimable à l’export. Le fichier existe ; ces informations y manquent :',
  mtgjsonSpalte: 'Export',
  speichern: 'Enregistrer',
  speichernUnter: 'Enregistrer sous …',
  speicherSpalte: 'Enregistrer',
  gespeichert: 'Enregistré dans',
  beimSpeichernBefunde:
    'Non repris lors de l’enregistrement. Le fichier du set a été écrit ; ces fichiers à côté manquent :',
  oeffnen: 'Afficher les cartes de ce set',
  handlungenAn: 'Actions — {name}',
  handlungenKurz: '…',
  vorbehalte: {
    mitZahl: 'Non interprétées ({anzahl})',
    ueberschrift: 'Non interprété',
    erklaerung:
      'L’import n’a pas compris ces lignes. Elles sont conservées telles quelles et ne seront pas perdues lors de l’enregistrement.',
    keine: 'Rien — ce set a été lu intégralement.',
  },

  // **The portable envelope** — work item N4 of phase 5.4
  // (#1457).
  umschlagAblegen: 'Enregistrer comme enveloppe',
  umschlagGeschnuert:
    'Enregistré. L’enveloppe est signée — quiconque la reçoit peut vérifier qu’elle vient d’ici.',
  umschlagTitel: 'Accueillir une enveloppe',
  umschlagErklaerung:
    'Une enveloppe est un set enregistré par quelqu’un d’autre. Elle est d’abord examinée puis intégrée — rien ne change dans la base de données avant.',
  umschlagWaehlen: 'Examiner une enveloppe …',
  umschlagUrheber: 'Signé par',
  umschlagAbdruckHinweis:
    'L’empreinte indique QUI a signé — pas si tu dois lui faire confiance. Si tu connais la personne, compare-la avec elle.',
  umschlagInhalt: 'Elle contient',
  umschlagUebernehmen: 'Intégrer',
  umschlagVerwerfen: 'Laisser tomber',
  umschlagUebernommen:
    'Intégré. Le set est créé à neuf — il a ses propres identifiants et n’écrase rien.',
}
