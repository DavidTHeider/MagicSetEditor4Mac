// **The “Users and profiles” page** — phase 4.12
// (#1351).
import { faktor, konto } from './seitenkonto'
import { sitzung } from './seitensitzung'

export const nutzer = {
  titel: 'Utilisateurs et profils',

  laschen: {
    allgemein: 'Général',
    konto: 'Compte',
    profil: 'Profils',
    persoenlich: 'Personnel',
  },

  allgemeinErklaerung:
    'Un utilisateur est la personne. Un profil est un environnement de travail de cette personne — ses propres sets, sa propre collection, ses propres réglages. Qui travaille où en ce moment figure dans la barre latérale, sous le nom de l’application ; on change de profil ici.',

  kontoErklaerung:
    'Ce avec quoi cet utilisateur se connecte. Tant qu’il n’y a qu’un seul utilisateur avec un seul profil, rien n’est demandé.',

  profilErklaerung:
    'Qui veut collectionner séparément pour un tournoi n’a pas besoin d’un deuxième compte, mais d’un deuxième profil.',

  niemand:
    'Aucun autre utilisateur pour l’instant. Le premier a été créé au premier démarrage.',
  keineProfile: 'Cet utilisateur n’a pas encore de profil.',

  neuerNutzer: 'Créer un utilisateur',
  neuerNutzerName: 'Nom d’affichage du nouvel utilisateur',
  neuerNutzerNameHinweis:
    'Unique et sans espaces aux extrémités — il distingue deux utilisateurs',
  neuerNutzerKnopf: 'Créer l’utilisateur',
  nutzerAngelegt: 'Créé : {name}',

  neuesProfil: 'Créer un autre profil',
  neuesProfilNutzer: 'Pour quel utilisateur',
  neuesProfilName: 'Nom du profil',
  neuesProfilNameHinweis:
    'Unique par utilisateur — deux personnes peuvent chacune avoir un profil « Tournoi »',
  neuesProfilKnopf: 'Créer le profil',
  profilAnlegenOeffnen: 'Créer un profil …',
  profilAngelegt: 'Créé : {name}',

  wahl: 'Dans quel profil tu travailles',
  wahlErklaerung:
    'Les sets, decks, la collection, les tournois et les réglages appartiennent au profil choisi. Changer de profil ne supprime rien — ce qui se trouve ici sera de retour lorsque tu reviendras.',
  hier: 'Tu travailles ici',
  hierhin: 'Passer ici',
  hierhinFuer: 'Passer à {nutzer} · {profil}',

  nutzerwahl: 'Qui travaille',
  nutzerwahlErklaerung:
    'Regarder dans le profil de quelqu’un d’autre ne change pas qui tu es — tes propres autorisations décident où tu peux aller. Ce n’est pas la même chose que le profil ci-dessus. Pour devenir quelqu’un d’autre, tu prouves son deuxième facteur ; tu travailles ensuite dans son premier profil.',
  dasBinIch: 'C’est toi en ce moment',
  werden: 'Devenir cette personne',
  werdenFuer: 'Devenir {nutzer}',

  anmeldung: 'Connexion via un fournisseur',
  anmeldungErklaerung:
    'Pour te connecter via un fournisseur, saisis son `client_id` sous **Réglages › Utilisateur et profil › Connexion via un fournisseur**. MagicSetEditor4Mac ne fournit aucun identifiant ; tu crées ta propre application OAuth chez le fournisseur, et un secret, s’il est exigé, va dans le trousseau et non dans la bibliothèque.',
  anmelden: 'Se connecter',
  anbieterstand: '{eingerichtet}, {geprueft}',
  eingerichtet: 'configuré',
  nichtEingerichtet: 'non configuré',
  geprueft: {
    simuliert: 'testé uniquement face à une simulation',
    realGeprueft: 'testé face au vrai fournisseur',
  },
  anmeldenBei: 'Se connecter avec {anbieter}',
  anmeldungAbbrechen: 'Annuler',
  anmeldungAbbrechenBei: 'Annuler la connexion avec {anbieter}',
  verbunden: 'Connecté',
  verbundenAls: 'Connecté en tant que {name}',
  loesen: 'Déconnecter',
  loesenVon: 'Se déconnecter de {anbieter}',

  anbieternamen: {
    github: 'GitHub',
    google: 'Google',
    microsoft: 'Microsoft',
    x: 'X',
  } as Record<string, string>,

  pflege: 'Modifier les utilisateurs et les profils',
  pflegeErklaerung:
    'Un nom peut être modifié. Pour supprimer, le nom doit rester tel quel — il confirme de qui il s’agit.',
  bearbeiten: 'Modifier {name}',
  bearbeitenKurz: 'Modifier',
  neuerNameFuer: 'Nom',
  umbenennen: 'Renommer',
  loeschen: 'Supprimer',
  loeschenBestaetigt: 'Supprimer {name} définitivement',
  loeschenFrage: 'Supprimer {name} définitivement ?',
  loeschenJa: 'Supprimer définitivement',
  abbrechen: 'Annuler',
  nutzerLoeschenHinweis:
    'Un utilisateur supprimé emporte ses profils avec lui. Le dernier utilisateur ne peut pas être supprimé.',
  profilLoeschenHinweis:
    'Le dernier profil d’un utilisateur ne peut pas être supprimé. Ce qui se trouve dans le profil reste dans la base de données, mais n’est plus accessible.',

  angaben: 'Informations personnelles',
  angabenErklaerung:
    'Tout est facultatif, sauf le nom d’affichage. La **civilité** est un champ libre : toute liste laisserait quelqu’un de côté, et une application qui force à choisir entre de mauvaises entrées s’est adressée à la personne de travers. L’adresse, le numéro de téléphone et le mot de passe se trouvent sous **Compte**.',
  angabenSichern: 'Enregistrer les informations',
  felder: {
    name: 'Nom d’affichage',
    anrede: 'Civilité',
    namenszusatz: 'Complément de nom',
    vorname: 'Prénom',
    weiterenamen: 'Autres prénoms',
    nachname: 'Nom de famille',
    slogan: 'Slogan',
    kurzebio: 'Biographie courte',
    langebio: 'Biographie longue',
  },

  // **Die Profil-Einstellungen** — Werkstück 9 von Phase 4.12 (#1351).
  badge: 'Badge du profil',
  badgeVon: 'Badge du profil de {name}',
  badgeWaehlen: 'Choisir un badge …',
  badgeEntfernen: 'Retirer le badge',
  profilangaben: 'Informations du profil',
  profilangabenErklaerung:
    'Elles appartiennent à **ce profil**, et non à toi — qui en a plusieurs les distingue ici. Le **slogan** se place sous le tien et ne le remplace pas. Le **badge** se place en bas à droite de ta photo de profil : il indique le rôle dans lequel tu travailles actuellement.',
  profilangabenSichern: 'Enregistrer le profil',
  profilfelder: {
    name: 'Nom du profil',
    subslogan: 'Slogan',
    kurzebeschreibung: 'Description courte',
    langebeschreibung: 'Description longue',
  },

  // **Technische Profile und ihre Zugriffstoken** — Werkstück 10 von
  // Phase 4.12 (#1351), Punkte 12 und 13.
  technisch: 'Profil technique — personne n’est derrière',
  technischErklaerung:
    'Un profil technique sépare ce qu’un service écrit de lui-même de ce que tu fais toi-même. Qui laisse un service travailler à sa place ne veut pas qu’il écrive dans **son** profil : ce qui y apparaît n’a pas été fait de sa main. Il t’appartient malgré tout, car il faut que quelqu’un puisse le désactiver.',
  ablauf: 'Expire le (vide : jamais)',
  ablaufErklaerung:
    'Un instant à partir duquel ce profil est expiré — **pas un ordre de suppression**. Rien ne disparaît de soi-même.',

  token: 'Jetons d’accès',
  tokenErklaerung:
    'C’est ainsi qu’un appelant automatisé s’identifie auprès du point d’accès REST — dans **ce** profil, avec l’autorisation définie ici. La fenêtre, la ligne de commande et MCP ne demandent aucun jeton : on y sait déjà qui travaille.',
  tokenLeer: 'Ce profil n’a pas encore de jeton d’accès.',
  tokenBezeichnung: 'À quoi il sert',
  tokenScope: 'Autorisation',
  tokenLesen: 'Lecture seule',
  tokenSchreiben: 'Lecture et écriture',
  tokenAblauf: 'Expire le (vide : jamais)',
  tokenAnlegen: 'Créer un jeton',
  tokenAendern: 'Appliquer',
  tokenLoeschen: 'Révoquer',
  tokenLoeschenFrage: 'Révoquer le jeton « {name} » ?',
  tokenLoeschenErklaerung:
    'Quiconque l’utilise est ensuite bloqué. Un nouveau jeton doit être saisi de nouveau à cet endroit.',
  tokenLoeschenJa: 'Révoquer',
  tokenAbgelaufen: 'expiré',
  tokenUnbefristet: 'n’expire jamais',

  tokenEinmalig:
    'La valeur en clair apparaît **exactement une fois**, juste après la création. Ensuite, elle n’existe plus nulle part ; qui la perd crée un nouveau jeton.',
  tokenKlartext: 'Copie-la maintenant — cette valeur ne reviendra pas :',

  // **Rotation, bulk revocation, expiry presets and MCP client set-up** —
  // piece O2 of phase 5.2 (#1426).
  tokenPrincipal: 'Pour qui il travaille',
  tokenAblaufwahl: 'Expiration',
  tokenAblaufwahlen: {
    tag: 'Un jour',
    monat: 'Trente jours',
    jahr: 'Un an',
    nie: 'Jamais',
    eigen: 'Date personnalisée',
  },
  tokenRotieren: 'Nouveau secret',
  tokenRotierenFrage: 'Générer un nouveau secret ?',
  tokenRotierenErklaerung:
    'L’ancien reste valable encore une heure afin qu’une tâche en cours puisse se terminer. La nouvelle valeur en clair apparaît exactement une fois.',
  tokenAlleWiderrufen: 'Tout révoquer',
  tokenAlleWiderrufenFrage: 'Révoquer tous les jetons d’accès de ce profil ?',
  tokenAlleWiderrufenErklaerung:
    'Tout appelant automatisé qui en utilise un est ensuite bloqué — même en pleine session. Ce qui a été créé avec eux reste.',
  tokenGezogen: '{zahl} jetons révoqués.',
  tokenLetzteNutzung: 'dernière utilisation : {zeit}',
  tokenNieBenutzt: 'jamais utilisé',
  tokenVorgaengerBis: 'ancien secret valable jusqu’au {zeit}',
  tokenOnboarding: 'Comment configurer un client MCP',
  tokenOnboardingSchritt1:
    'Crée un jeton et copie la valeur en clair — elle n’apparaît qu’une fois.',
  tokenOnboardingSchritt2:
    'Écris-la dans un fichier lisible par toi seul (mode 0600). Un fichier lisible par d’autres est rejeté par « mse-mcp ».',
  tokenOnboardingSchritt3:
    'Lance « mse-mcp » avec MSE_TOKEN_DATEI pointant vers ce fichier — comme variable d’environnement, jamais comme argument : cela apparaîtrait dans la liste des processus.',
  tokenOnboardingSchritt4:
    'Le jeton est vérifié à chaque appel d’outil : une révocation, une expiration ou une autorisation réduite prend effet dès l’appel suivant. Sans jeton, « mse-mcp » fonctionne comme avant, sans authentification, dans le profil affiché.',
  tokenOnboardingBeispiel:
    '{ "mcpServers": { "mse": { "command": "/Applications/MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp", "env": { "MSE_TOKEN_DATEI": "/Users/ich/.mse/mcp-token" } } } }',
  tokenOnboardingBeispielTitel: 'Exemple de configuration',

  // **Vorgaben, die auf Zeilen zeigen** — Werkstück 15 von Phase 4.12
  // (#1351), Punkte 4 bis 7.
  vorgaben: 'Valeurs par défaut de ce profil',
  vorgabenErklaerung:
    'Ce avec quoi commence un nouvel enregistrement dans **ce profil** — un format, un jeu, un modèle de la base de données. Contrairement aux réglages, les valeurs par défaut pointent vers des **lignes de la base de données** ; ce qui peut être choisi dépend de ce qui s’y trouve.',
  vorgabenamen: {
    format: 'Format par défaut des nouveaux decks',
    spiel: 'Jeu par défaut des nouveaux sets',
    vorlage: 'Modèle par défaut des nouveaux sets',
    deck: 'Decks préférés',
    lagerort: 'Emplacements de stockage attribués',
  } as Record<string, string>,
  vorgabenOhne: 'sans valeur par défaut',
  vorgabenLeer: 'Il n’y a encore rien pour cela dans la base de données.',
  vorgabenVerwaist: 'Ne pointe vers rien, car cela n’existe plus :',

  // **Rechte, Rollen und Sichtbarkeit** — Werkstück 11 von Phase 4.12
  // (#1351).
  freigaben: 'Qui peut travailler dans ce profil',
  freigabenErklaerung:
    'Un **lecteur** voit seulement que ce profil existe — pas ce qu’il contient. Un **collaborateur** y passe et y travaille. Un **gestionnaire** accorde lui-même des accès. Ce profil est le tien si tu l’as créé ; tu le gères alors de toute façon.',
  freigabenNurLesen: 'Seul celui qui gère ce profil peut accorder des accès.',
  freigabenLeer:
    'Jusqu’ici, personne d’autre que le propriétaire ne peut entrer dans ce profil.',
  freigabeNutzer: 'À qui',
  freigabeNiemand: 'personne',
  freigabeRolle: 'En tant que',
  freigabeVergeben: 'Accorder',
  freigabeZurueckziehen: 'Retirer',
  rollennamen: {
    leser: 'Lecteur — voit que le profil existe',
    mitarbeiter: 'Collaborateur — y passe et y travaille',
    verwalter: 'Gestionnaire — accorde lui-même des accès',
  } as Record<string, string>,

  bild: 'Photo de profil',
  bildVon: 'Photo de profil de {name}',
  bildWaehlen: 'Choisir une image …',
  bildEntfernen: 'Retirer l’image',
  ohneBild: '{name} n’a pas de photo de profil',

  nochNicht:
    'Les données sont séparées entre les profils — tant que personne n’ouvre le fichier de la base de données avec un outil étranger. Le chiffrement de la base de données sous « Réglages › Données et sauvegarde » aide contre cela.',

  // **Die vier Kontoarten** — Werkstück 6 (#1351), in einer eigenen Datei.
  konto,
  sitzung,

  // **Der zweite Faktor** — Werkstück 7 (#1351).
  faktor,

  // **Was dem Menschen gehört** — Werkstück 8 von Phase 4.12 (#1351),
  // Punkte 9, 10 und 11.
  //
  // **Am 2026-09-06 aus `seiteneinstellungen` hierher gezogen**, nach dem
  // Zweitblick über die Phase: die Einstellungen einer Person lagen an zwei
  // Orten, und wer sie suchte, suchte an beiden.

  persoenlich: 'Mon compte',

  persoenlichErklaerung:
    'Ces réglages t’appartiennent, à **toi**, et non à ce profil — tu les emportes dans chacun de tes profils. Les deux du haut priment sur le système d’exploitation, et **dans un seul sens** : ils activent, ils ne désactivent jamais.',

  auskunft: 'Mes données',

  auskunftErklaerung:
    'Ce que ce programme a enregistré sur toi — informations, modes de connexion, **types** de deuxième facteur, réglages et noms de profils. **Sans mot de passe et sans secret** : ni l’un ni l’autre n’est une donnée sur toi, mais un moyen d’être toi. Les sets, decks et la collection n’y figurent pas — ils appartiennent à un profil et sortent par leurs propres voies.',

  auskunftSichern: 'Enregistrer mes données …',

  datenschutz: 'Ce que cette application sait de toi',

  datenschutzHinweis:
    'Quelles informations sont saisies, où elles se trouvent, ce qui quitte cet ordinateur — et ce qui ne le quitte expressément pas. Décrit sous **Aide → Ce que cette application sait de toi**.',
} as const
