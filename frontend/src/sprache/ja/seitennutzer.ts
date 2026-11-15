// **The “Users and profiles” page** — phase 4.12
// (#1351).
import { faktor, konto } from './seitenkonto'
import { sitzung } from './seitensitzung'

export const nutzer = {
  titel: 'ユーザとプロファイル',

  laschen: {
    allgemein: '一般',
    konto: 'アカウント',
    profil: 'プロファイル',
    persoenlich: '個人',
  },

  allgemeinErklaerung:
    'ユーザは人を表します。プロファイルは、その人の1つの作業環境で、専用のセット、専用のコレクション、専用の設定を持ちます。今誰がどこで作業しているかは、サイドバーのアプリケーション名の下に表示されます。切り替えはここで行います。',

  kontoErklaerung:
    'このユーザがサインインに使うものです。ユーザとプロファイルがそれぞれ1つだけの間は、何も尋ねられません。',

  profilErklaerung:
    'トーナメント用に別にコレクションを管理したい場合は、2つ目のアカウントではなく、2つ目のプロファイルが必要です。',

  niemand:
    'ほかのユーザはまだいません。最初のユーザは初回の起動時に作成されました。',
  keineProfile: 'このユーザにはまだプロファイルがありません。',

  neuerNutzer: 'ユーザを作成',
  neuerNutzerName: '新しいユーザの表示名',
  neuerNutzerNameHinweis:
    '一意で、前後に空白を含まないこと。2人のユーザを区別するために使われます',
  neuerNutzerKnopf: 'ユーザを作成',
  nutzerAngelegt: '作成しました：{name}',

  neuesProfil: '別のプロファイルを作成',
  neuesProfilNutzer: 'どのユーザ用か',
  neuesProfilName: 'プロファイル名',
  neuesProfilNameHinweis:
    'ユーザごとに一意です。2人がどちらも「トーナメント」を持つことはできます',
  neuesProfilKnopf: 'プロファイルを作成',
  profilAnlegenOeffnen: 'プロファイルを作成…',
  profilAngelegt: '作成しました：{name}',

  wahl: '作業するプロファイル',
  wahlErklaerung:
    'セット、デッキ、コレクション、トーナメント、設定は、選択したプロファイルに属します。切り替えても何も削除されません。元に戻せば、ここにあるものがまた表示されます。',
  hier: 'ここで作業中です',
  hierhin: 'ここに切り替える',
  hierhinFuer: '{nutzer} · {profil} に切り替える',

  nutzerwahl: '作業している人',
  nutzerwahlErklaerung:
    '他の人のプロファイルを見ても、あなたが誰であるかは変わりません。どこに入れるかは、あなた自身が共有されている範囲で決まります。これは上のプロファイルとは別のものです。別の人になるには、その人の第2要素を証明する必要があります。そのうえで、その人の最初のプロファイルで作業します。',
  dasBinIch: '現在のあなたです',
  werden: 'この人になる',
  werdenFuer: '{nutzer} になる',

  anmeldung: 'プロバイダでサインイン',
  anmeldungErklaerung:
    'プロバイダでサインインするには、「設定」＞「ユーザとプロファイル」＞「プロバイダを使ったサインイン」に、そのプロバイダの `client_id` を入力します。MagicSetEditor4Mac には認証情報は同梱されていません。プロバイダ側で自分の OAuth アプリケーションを作成してください。シークレットが必要な場合は、ライブラリではなくキーチェーンに保存します。',
  anmelden: 'サインイン',
  anbieterstand: '{eingerichtet}、{geprueft}',
  eingerichtet: '設定済み',
  nichtEingerichtet: '未設定',
  geprueft: {
    simuliert: 'シミュレーションでのみ検証済み',
    realGeprueft: '実際のプロバイダで検証済み',
  },
  anmeldenBei: '{anbieter} でサインイン',
  anmeldungAbbrechen: 'キャンセル',
  anmeldungAbbrechenBei: '{anbieter} でのサインインをキャンセル',
  verbunden: '接続済み',
  verbundenAls: '{name} として接続済み',
  loesen: '接続を解除',
  loesenVon: '{anbieter} との接続を解除',

  anbieternamen: {
    github: 'GitHub',
    google: 'Google',
    microsoft: 'Microsoft',
    x: 'X',
  } as Record<string, string>,

  pflege: 'ユーザとプロファイルを編集',
  pflegeErklaerung:
    '名前は変更できます。削除するには、誰のことかを確認するため、名前をそのまま入力する必要があります。',
  bearbeiten: '{name} を編集',
  bearbeitenKurz: '編集',
  neuerNameFuer: '名前',
  umbenennen: '名前を変更',
  loeschen: '削除',
  loeschenBestaetigt: '{name} を完全に削除',
  loeschenFrage: '{name} を完全に削除しますか？',
  loeschenJa: '完全に削除',
  abbrechen: 'キャンセル',
  nutzerLoeschenHinweis:
    'ユーザを削除すると、そのユーザのプロファイルも一緒に削除されます。最後のユーザは削除できません。',
  profilLoeschenHinweis:
    'ユーザの最後のプロファイルは削除できません。プロファイル内のデータはストアに残りますが、アクセスできなくなります。',

  angaben: '個人情報',
  angabenErklaerung:
    '表示名以外はすべて任意です。**敬称**は自由入力です。どんな選択肢のリストでも、誰かを漏らしてしまいます。誤った選択肢のどれかを強制するアプリケーションは、その人に間違った呼びかけをしていることになります。住所、電話番号、パスワードは**アカウント**にあります。',
  angabenSichern: '個人情報を保存',
  felder: {
    name: '表示名',
    anrede: '敬称',
    namenszusatz: '名前の付記',
    vorname: '名',
    weiterenamen: 'ミドルネーム',
    nachname: '姓',
    slogan: 'キャッチフレーズ',
    kurzebio: '短い経歴',
    langebio: '詳しい経歴',
  },

  // **Die Profil-Einstellungen** — Werkstück 9 von Phase 4.12 (#1351).
  badge: 'プロファイルのバッジ',
  badgeVon: '{name} のプロファイルのバッジ',
  badgeWaehlen: 'バッジを選択…',
  badgeEntfernen: 'バッジを削除',
  profilangaben: 'プロファイルの詳細',
  profilangabenErklaerung:
    'これらは**このプロファイル**に属し、あなた自身には属しません。複数のプロファイルを持つ場合、ここで見分けられます。**キャッチフレーズ**はあなた自身のものの下に表示され、置き換えるものではありません。**バッジ**はプロフィール画像の右下に表示され、現在作業している役割を示します。',
  profilangabenSichern: 'プロファイルを保存',
  profilfelder: {
    name: 'プロファイル名',
    subslogan: 'キャッチフレーズ',
    kurzebeschreibung: '短い説明',
    langebeschreibung: '詳しい説明',
  },

  // **Technische Profile und ihre Zugriffstoken** — Werkstück 10 von
  // Phase 4.12 (#1351), Punkte 12 und 13.
  technisch: '技術プロファイル — 背後に人はいません',
  technischErklaerung:
    '技術プロファイルは、サービスが自動的に書き込む内容を、あなた自身の操作とは分けて保持します。サービスに代わりに作業させる場合、**自分の**プロファイルに書き込まれたくはないはずです。そこに現れるものは、自分の手で作ったものではないからです。それでも、これはあなたのものです。誰かがオフにできる必要があるためです。',
  ablauf: '有効期限（空欄：無期限）',
  ablaufErklaerung:
    'このプロファイルが失効する日時です。**削除の指示ではありません**。何も自動的には消えません。',

  token: 'アクセストークン',
  tokenErklaerung:
    '自動化された呼び出し元が REST エンドポイントで自分を識別するためのものです。**このプロファイル**で、ここで設定した権限で動作します。ウインドウ、コマンドライン、MCP ではトークンは求められません。誰が作業しているかが、そこではすでに決まっているためです。',
  tokenLeer: 'このプロファイルには、アクセストークンがまだありません。',
  tokenBezeichnung: '用途',
  tokenScope: '権限',
  tokenLesen: '読み取り専用',
  tokenSchreiben: '読み取りと書き込み',
  tokenAblauf: '有効期限（空欄：無期限）',
  tokenAnlegen: 'トークンを作成',
  tokenAendern: '適用',
  tokenLoeschen: '無効にする',
  tokenLoeschenFrage: 'トークン「{name}」を無効にしますか？',
  tokenLoeschenErklaerung:
    'このトークンを使っているものは、その後アクセスできなくなります。新しいトークンは、そちらに入力し直す必要があります。',
  tokenLoeschenJa: '無効にする',
  tokenAbgelaufen: '失効済み',
  tokenUnbefristet: '無期限',

  tokenEinmalig:
    '平文の値が表示されるのは、作成した直後の**1回だけ**です。その後はどこにも存在しません。失くした場合は、新しいトークンを作成します。',
  tokenKlartext: '今すぐコピーしてください。この値は二度と表示されません：',

  // **Rotation, bulk revocation, expiry presets and MCP client set-up** —
  // piece O2 of phase 5.2 (#1426).
  tokenPrincipal: '誰のために動作するか',
  tokenAblaufwahl: '有効期限',
  tokenAblaufwahlen: {
    tag: '1日',
    monat: '30日',
    jahr: '1年',
    nie: '無期限',
    eigen: '日時を指定',
  },
  tokenRotieren: '新しいシークレット',
  tokenRotierenFrage: '新しいシークレットを生成しますか？',
  tokenRotierenErklaerung:
    '実行中のジョブが終了できるように、古いシークレットはあと1時間有効です。新しい平文の値は1回だけ表示されます。',
  tokenAlleWiderrufen: 'すべて無効にする',
  tokenAlleWiderrufenFrage:
    'このプロファイルのすべてのアクセストークンを無効にしますか？',
  tokenAlleWiderrufenErklaerung:
    'それらを使っている自動化された呼び出し元は、セッションの途中でもアクセスできなくなります。すでに作成されたものは残ります。',
  tokenGezogen: '{zahl} 個のトークンを無効にしました。',
  tokenLetzteNutzung: '最終使用：{zeit}',
  tokenNieBenutzt: '未使用',
  tokenVorgaengerBis: '古いシークレットは {zeit} まで有効',
  tokenOnboarding: 'MCP クライアントの設定方法',
  tokenOnboardingSchritt1:
    'トークンを作成し、平文の値をコピーします。表示されるのは1回だけです。',
  tokenOnboardingSchritt2:
    '自分だけが読めるファイル（モード 0600）に書き込みます。他の人が読めるファイルは、「mse-mcp」に拒否されます。',
  tokenOnboardingSchritt3:
    'MSE_TOKEN_DATEI にそのファイルを指定して、「mse-mcp」を起動します。環境変数として指定し、引数には決して指定しないでください。引数はプロセス一覧に表示されます。',
  tokenOnboardingSchritt4:
    'トークンはツールの呼び出しのたびに確認されます。無効化、失効、権限の縮小は、次の呼び出しから有効になります。トークンがない場合、「mse-mcp」はこれまでどおり、表示中のプロファイルで認証なしで動作します。',
  tokenOnboardingBeispiel:
    '{ "mcpServers": { "mse": { "command": "/Applications/MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp", "env": { "MSE_TOKEN_DATEI": "/Users/ich/.mse/mcp-token" } } } }',
  tokenOnboardingBeispielTitel: '設定例',

  // **Vorgaben, die auf Zeilen zeigen** — Werkstück 15 von Phase 4.12
  // (#1351), Punkte 4 bis 7.
  vorgaben: 'このプロファイルの初期値',
  vorgabenErklaerung:
    '**このプロファイル**で、新しいレコードが最初に持つ内容です（フォーマット、ゲーム、ストアにあるテンプレートなど）。設定とは異なり、初期値は**ストア内の行**を指します。選択できる内容は、そこにあるものによって決まります。',
  vorgabenamen: {
    format: '新しいデッキの初期フォーマット',
    spiel: '新しいセットの初期ゲーム',
    vorlage: '新しいセットの初期テンプレート',
    deck: '優先するデッキ',
    lagerort: '割り当てた保管場所',
  } as Record<string, string>,
  vorgabenOhne: '初期値なし',
  vorgabenLeer: 'これに該当するものは、ストアにまだありません。',
  vorgabenVerwaist: '存在しないため、どこも指していません：',

  // **Rechte, Rollen und Sichtbarkeit** — Werkstück 11 von Phase 4.12
  // (#1351).
  freigaben: 'このプロファイルで作業できる人',
  freigabenErklaerung:
    '**閲覧者**は、このプロファイルがあることだけがわかり、中身は見えません。**共同作業者**はこのプロファイルに切り替えて、そこで作業します。**管理者**は、自分でアクセスを許可できます。あなたが作成したプロファイルは、あなたのものです。その場合、あなたは常に管理者でもあります。',
  freigabenNurLesen:
    'アクセスを許可できるのは、このプロファイルの管理者だけです。',
  freigabenLeer: 'これまで、所有者以外はこのプロファイルに入れません。',
  freigabeNutzer: '誰に',
  freigabeNiemand: '誰もいません',
  freigabeRolle: '役割',
  freigabeVergeben: '許可',
  freigabeZurueckziehen: '取り消す',
  rollennamen: {
    leser: '閲覧者 — プロファイルがあることがわかる',
    mitarbeiter: '共同作業者 — 切り替えて、そこで作業する',
    verwalter: '管理者 — 自分でアクセスを許可する',
  } as Record<string, string>,

  bild: 'プロフィール画像',
  bildVon: '{name} のプロフィール画像',
  bildWaehlen: '画像を選択…',
  bildEntfernen: '画像を削除',
  ohneBild: '{name} にはプロフィール画像がありません',

  nochNicht:
    'データはプロファイルごとに分けられています。ただし、誰かが別のツールでストアのファイルを開かない限りです。それに対しては、「設定」＞「データとバックアップ」でストアを暗号化すると効果があります。',

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

  persoenlich: '自分のアカウント',

  persoenlichErklaerung:
    'これらの設定は、このプロファイルではなく**あなた**に属し、あなたのどのプロファイルにも引き継がれます。上の2つはオペレーティングシステムの設定を上書きしますが、**一方向のみ**です。オンにはできても、オフにはできません。',

  auskunft: '自分のデータ',

  auskunftErklaerung:
    'このプログラムが保存しているあなたに関する情報です：個人情報、サインイン方法、第2要素の**種類**、設定、プロファイル名。**パスワードとシークレットは含まれません**。どちらもあなたに関するデータではなく、あなたになりすます手段だからです。セット、デッキ、コレクションは含まれません。それらはプロファイルに属し、別の方法で持ち出します。',

  auskunftSichern: '自分のデータを保存…',

  datenschutz: 'このアプリケーションが把握しているあなたの情報',

  datenschutzHinweis:
    'どの情報が記録され、どこに保存され、何がこのコンピュータの外に出るのか、そして何が出ないのかを説明します。**ヘルプ → このアプリケーションが把握しているあなたの情報**に説明があります。',
} as const
