// The texts of the **technical page** — operating state, effect counter and
// the three tools without a window.
//
// Split out of `seiten.ts` when the file grew past 400 lines with the effect
// counter (work item O9, #1433).

import { betrieb, leistungsprofil, zwischenspeicher } from './seitenbetrieb'
import { dienstzeugnis, fernwege, schnittstellen } from './seitentechnikdienst'
import { einladung } from './einladung'
import { gegenueber } from './gegenueber'

export const technik = {
  betrieb,
  dienstzeugnis,
  einladung,
  gegenueber,
  fernwege,
  schnittstellen,
  leistungsprofil,
  zwischenspeicher,
  gesundheit: '動作状態',
  gesundheitErklaerung:
    'このプロセスが動作しているか、作業できる状態かを示します。外部サービスが応答しなくても、作業不能にはなりません。ネットワークがなくても、編集は続行できます。',
  lebt: 'プロセスは応答しています。',
  bereit: '作業できます。',
  nichtBereit: '現在は作業できません。',
  grundBestandZu: 'データベースが開かれていません。',
  grundWanderungLaeuft: 'データベースは新しいスキーマに移行中です。',
  grundBestandUnlesbar: 'データベースが応答しません。',
  fremdeDienste: '外部サービス',
  erreichbar: '接続可能',
  nichtErreichbar: '接続不可',
  metriken: '起動後の外部への作用',
  metrikenErklaerung:
    'このプロセスが外部に対して行ったことを、インターフェース、作用の種類、結果ごとに示します。確認された作用のみが数えられ、単なる読み取りはここには表示されません。',
  metrikenGrenze:
    '最大{grenze}行です。カード、セット、プロファイルの数が増えても増えません。',
  metrikenLeer: 'まだ外部に作用したものはありません。',
  spalteFlaeche: 'インターフェース',
  spalteWirkung: '作用',
  spalteAusgang: '結果',
  spalteAnzahl: '件数',
  flaechen: {
    fenster: 'ウインドウ',
    kommandozeile: 'コマンドライン',
    mcp: 'MCP',
    dienst: 'スクリプト用サービスを起動',
  },
  wirkungen: {
    lokalLesen: 'ローカルで読み取り',
    lokalSchreiben: 'ローカルで書き込み',
    externSenden: '外部へ送信',
    externSchreiben: '外部を変更',
    verbindlich: '確定操作',
  },
  ausgaenge: { gewaehrt: '許可', abgewiesen: '拒否' },
  paket: {
    ueberschrift: '診断バンドル',
    erklaerung:
      '同じ情報を部分ごとに、チェックサム付きのファイルにまとめたものです。テキストだけでは足りない報告用です。指定した場所に作成され、自動的にどこにも送信されません。',
    inhalt:
      'バージョン、プラットフォーム、データベース、パッケージ、プロセスのカウンタ、ログ行、クラッシュの記録が含まれます。カード、画像、コレクション、認証情報は含まれません。',
    zusammenstellen: 'まとめる',
    sichern: 'バンドルを保存',
    loeschen: 'バンドルを削除',
    groesse: '最大{grenze}のうち{bytes}。',
    zeitraum: 'ログ行は{von}から{bis}までです。',
    ausgelassen: '省略 — {teil}:',
    teile: {
      fassung: 'バージョンとプラットフォーム',
      bestand: 'データベースの場所とスキーマ',
      pakete: 'インストール済みのテンプレートパッケージ',
      metriken: 'このプロセスのカウンタ',
      protokoll: 'ログ行',
    },
    spalteTeil: '部分',
    spalteDatei: 'ファイル',
    spalteZeilen: '行数',
    spalteBytes: 'サイズ',
    spalteMitnehmen: '含める',
    mitnehmen: '{teil}をバンドルに含める',
  },
  diagnose: {
    ueberschrift: 'バグ報告用の診断情報',
    erklaerung:
      '何かがうまく動作しないとき、このテキストが原因の調査に役立ちます。ここに表示されるだけで、他の人に渡すかどうかはあなた次第です。',
    inhalt:
      'バージョン、マシン、データベース、ログ、テンプレートパッケージが含まれます。カードのテキスト、フィールドの値、セットの内容は含まれません。具体的には、バージョン、マシンの種類、プログラムのビルド元、ウインドウとプログラムの間のバインディング規約、データベースの場所、スキーマの状態、孤立した行の数、ログの場所、テンプレートパッケージ、最新のログ200行です。ホームフォルダは「~」と表示されます。',
    anzeigen: '診断情報を表示',
    kopieren: 'コピー',
    kopiert: 'コピーしました',
    kopierfehler:
      'コピーに失敗しました。ウインドウにクリップボードの権限がありません。テキストを選択して、手動でコピーしてください。',
    sichern: 'ファイルとして保存',
    // Filter über die Zeilen (NET-02, S13.NAT-04, #1706).
    filter: '行を検索',
    zeilenGezeigt: '{alle} 行中 {gezeigt} 行',
  },
  // **Die Aufschrift der Lasche — und nicht die des Abschnitts darin.**
  //
  // Sie hießen bis zum 2026-09-14 gleich, und die Lasche trug elf Kästen,
  // von denen einer „Werkzeuge ohne Fenster" war. Wer die
  // Netzeinstellungen suchte, suchte überall sonst.
  bereichUeberschrift: '技術情報と診断',

  // **Das Gebiet, das dabei entstanden ist.**
  netzUeberschrift: 'ネットワークと到達性',

  ueberschrift: 'ウインドウのないツール',
  erklaerung:
    'このアプリケーションには、ウインドウなしで動作する3つのプログラムが同梱されています。コマンドライン用の mse、言語モデル用の mse-mcp、スクリプト用サービスの mse-rest です。これらはアプリケーションバンドルの中にあり、下で設定したときにだけ検索パスに入ります。ここにある行には完全な場所が含まれています。',
  keineShell:
    'コピーするためのもので、実行はしません。ここにある内容は、お好みのターミナルで使ってください。',
  ort: 'これらは {ort} にあります。',
  vorhanden: '利用可能: {namen}。',
  keineWerkzeuge:
    '3つとも見つかりません。アプリケーションをソースから実行している間は正常です。リリース手順で、バンドルに組み込まれます。',
  kopieren: 'コマンドをコピー',
  kopiert: 'コピーしました。',
  kopierfehler:
    'コピーに失敗しました。ウインドウにクリップボードの権限がありません。テキストを選択して、手動でコピーしてください。',
  befehle: {
    hilfe: 'コマンドラインでできること',
    sets: 'データベース内のセットを一覧表示',
    verknuepfen: 'mse を検索パスで使えるようにする (パスワードを求められます)',
    dienst: 'ループバックインターフェースでサービスを起動',
    mcp: 'MCPサーバを起動 (言語モデル向け、標準入出力経由)',
    karten: '開いているセットのカードを一覧表示',
    statistik: '開いているセットを集計',
  },

  // **The local search index** — work item A15 of phase 5.3
  // (#1514).
  suchindex: 'ローカル検索インデックス',
  suchindexErklaerung:
    'カードのテキストとルールの該当箇所を、ローカルで検索できます。全文検索とベクトル検索を統合して、1つのリストにします。各ヒットには、その出所と、情報源のリビジョンが示されます。',
  suchindexKeiner:
    '検索インデックスが作成されていません。作成されるまでは、ここでの検索では何も見つかりません。',
  suchindexModell: 'モデル',
  suchindexLaufzeit: 'ランタイム',
  suchindexVerfahren: '方式',
  suchindexStaende: '情報源ごとのリビジョン',
  suchindexSemantisch: 'セマンティック検索',
  suchindexJa: 'はい',
  suchindexNein: 'いいえ',
  suchindexFrage: '何を探しますか？',
  suchindexSuchen: '検索',
  suchindexLeer: '見つかりませんでした。',
  suchindexModus: '検索方法',
  suchindexModusHybrid: '全文検索とベクトル検索',
  suchindexModusLexikalisch: '全文検索のみ',
  suchindexModusOhneIndex: '検索していません (インデックスがありません)',
  suchindexAufbau: '再構築',
  suchindexAufbauErklaerung:
    '実行中の世代の横で新しい世代を作成し、最後にだけ切り替えます。数分かかり、設定済みの埋め込みモデルが必要です。',
  suchindexModellfeld: '埋め込みモデル',
  suchindexQuellfeld: '一括データファイル',
  suchindexRevisionfeld: '情報源のリビジョン',
  suchindexDelta: '変更された部分のみ読み込む',
  suchindexAufbauen: '作成',
  suchindexAbbrechen: 'キャンセル',
  suchindexLaeuft: '作成中',
  suchindexGeschrieben: '件のセクションを書き込みました',
  suchindexFertig: '世代を作成しました',
  // **Configuration for agents** — A14, #1453.
  agenten: 'エージェント向けの準備',
  agentenErklaerung:
    'Claude Code と Codex は MCP に対応しています。下のスニペットには、このMac上の mse-mcp のフルパスが含まれています。コピーして貼り付けてください。このアプリケーションは、これらのファイルには何も書き込まず、エージェントも起動しません。',
  agentenOhneGeheimnis:
    'アクセスキーは含まれていません。mse-mcp が、キーチェーンから自分で読み取ります。ここに書くと、共有されがちなファイルに平文で残ってしまいます。',
  agentenNamen: {
    claude: 'Claude Code',
    codex: 'Codex',
  } as Record<string, string>,
  befehlszeileTitel: 'コマンドラインを設定',
  befehlszeileSatz:
    'mse と mse-mcp へのリンクを、あなたが所有するフォルダに配置します。管理者権限は不要で、シェルのプロファイルにも触れません。すでにそこにあるもので、ここから作られたものではないものは、そのまま残ります。',
  befehlszeileVerzeichnis: 'フォルダ: {verzeichnis}',
  befehlszeileNichtImPfad:
    'このフォルダはPATHに含まれていません。次の行を、ご自身でシェルのプロファイルに追加してください:',
  befehlszeileStaende: {
    frei: '未設定',
    eigen: '設定済み',
    veraltet: '古い — すでに存在しないプログラムを指しています',
    fremd: '別のコマンドがあるため、そのまま残します',
  } as Record<string, string>,
  befehlszeileVerdeckt:
    '{pfad} に隠されています。名前を入力すると、そちらが実行されます。',
  befehlszeilePruefen: '確認コマンド',
  befehlszeileEinrichten: '設定',
  befehlszeileHerausnehmen: '元に戻す',
  befehlszeileKopierenPfad: 'シェルのプロファイル用の行をコピー',
  befehlszeileKopierenPruefung: '確認コマンドをコピー',
  agentenDatei: '{datei} に記述します',
  agentenInhalt: '{datei} の内容',
  agentenBefehl: 'またはエージェントのコマンドラインで',
  agentenNachgelesen: '確認元: {quelle}',
  agentenKopierenDatei: 'スニペットをコピー',
  agentenKopierenBefehl: 'エージェントのコマンドをコピー',

  peerTitel: 'このプロファイルの身元の示し方',

  peerZeigen: '身元を表示',
  peerErklaerung:
    '相手が、確かにあなたと通信していると確認するための鍵ペアです。秘密鍵はこのMacのキーチェーンに残り、どこにも送信されません。外に出るのは識別子だけです。',
  // Was diese Identität nicht ist (NET-02, S13.NAT-05, #1706).
  peerNichtKonto:
    'これはこのプロファイルのものです。Apple アカウントでも、このコンピュータの名前でもありません。',
  peerStandOhne: 'このプロファイルには、まだ身元がありません。',
  peerStandBereit: '準備完了です。',
  peerStandVerloren: '鍵がありません。',
  peerStandFremd: 'キーチェーンとデータベースが一致しません。',
  peerKennung: '識別子',
  peerFingerabdruck: 'フィンガープリント',
  peerFingerabdruckHinweis:
    '相手に読み上げて伝えるのはこれです。表示名は証明になりません。',
  peerSeit: '{zeitpunkt}から',
  peerVorher: '以前: {kennung}',
  peerAnlegen: '身元を作成',
  peerRotieren: '身元を入れ替える',
  peerRotierenFrage:
    '新しい身元を作成しますか？ 現在のフィンガープリントを確認済みの相手は、全員がもう一度確認する必要があります。',
  peerRotierenJa: '入れ替える',
  peerRotierenNein: 'キャンセル',

  // **The known counterparts** — work item N1 of the same issue.
  kontakteTitel: 'このプロファイルが知っている相手',
  kontakteErklaerung:
    'このプロファイルがすでに会った相手です。会っただけでは何も許可されません。フィンガープリントを照合してはじめて検証可能になり、目的ごとの許可があってはじめて何かが許可されます。',
  kontakteLeer:
    'まだ誰とも会っていません。連絡してきた相手は、あとでここに表示されます (未確認として)。',
  kontakteZeigen: '連絡先を表示',
  kontakteAufnehmen: '連絡先を追加',
  kontakteAufnehmenErklaerung:
    '接続がない間は、識別子を手動で交換します。ここに追加した相手は未確認で、何もできません。',
  kontakteAufnehmenKennung: '相手の識別子',
  kontakteAufnehmenName: '相手の呼び名',
  kontakteStandUnbestaetigt: '未確認',
  kontakteStandBestaetigt: '確認済み',
  kontakteStandGesperrt: 'ブロック中',
  kontakteAbdruck: 'フィンガープリント',
  kontakteNennt: '自称: {name}',
  kontakteNenntHinweis: '自分で付けた名前は、身元の証明になりません。',
  kontakteGesehen: '最終確認: {zeitpunkt}',
  kontakteVergleichen: '照合したフィンガープリントを入力',
  kontakteBestaetigen: '確認',
  kontakteSperren: 'ブロック',
  kontakteEntsperren: 'ブロックを解除',
  kontakteRechte: '許可済み',
  kontakteKeineRechte: '許可されているものはありません。',
  zweckPaarung: '知り合い',
  zweckTeilen: 'オブジェクトの転送',
  zweckHandel: 'カードの取引',
  zweckPartie: 'ゲームをプレイ',
  kontakteAdresse: '接続先',
  kontakteAdresseBeispiel: '192.168.1.7:41234',
  kontakteAdresseHinweis:
    '同じネットワーク上のアドレス、または相手から届いた招待のどちらかです。空欄の場合は指定なしです。',
  kontakteAdresseMerken: '記憶',

  // **Die lokale Notiz und die freigegebenen Sichten** — Werkstück T3
  // von Phase 5.6 (#1468).
  kontakteNotiz: 'メモ',
  kontakteNotizBeispiel: '公正に取引する、返信が遅い…',
  kontakteNotizHinweis:
    'あなただけのメモです。ビューにも共有にも含まれず、外に出ることはありません。',
  kontakteNotizMerken: 'メモを保存',
  freigabenTitel: 'この相手に見えるもの',
  freigabenErklaerung:
    '共有がなければ、何も見えません。複数の範囲は並行して適用され、広い方が優先されます。',
  freigabenLeer: '許可されたビューはありません。',
  freigabenOhneZweck:
    'この相手は、まだ何もできません。まず目的を許可してください。',
  freigabenZweck: '目的',
  freigabenUmfang: '対象',
  freigabenErteilen: 'ビューを許可',
  freigabenWiderrufen: '取り消す',
  freigabenVorschau: '見える内容を確認',
  freigabenVorschauLeer: '何も見えません。',
  fernTitel: '他のネットワークからのアクセスを受け付ける',
  fernErklaerung:
    'これにより、同じネットワークにいなくても、相手があなたに接続できます。ポート転送もアカウントも不要です。招待が発行されるので、好きな方法で相手に送ってください。',
  fernStill: 'このプロファイルは、他のネットワークからは接続できません。',
  fernHinweis:
    '接続は、まず直接の経路を探し、他に方法がない場合にだけリレーを使います。リレーを通るのは暗号化された通信だけで、カードも名前も含まれません。',
  fernStarten: '招待を作成',
  fernBeenden: 'アクセスの受け付けを停止',
  fernWeitergeben:
    'これが相手に必要な情報です。これを傍受されても、得られるのは1行のテキストであって、身元ではありません。確認は接続時に、フィンガープリントで行われます。',
  fernStirbt:
    '招待は、このプロファイルが接続可能な間だけ有効です。停止すると招待は無効になり、次回は新しい招待が発行されます。',
  fernKopieren: '招待をコピー',
  fernKopiert: 'コピーしました。',
  horchenTitel: '他の相手からのアクセスを受け付ける',
  horchenErklaerung:
    'このプロファイルが待ち受けている間、既知のピアが接続を開始できます。' +
    'これは自動では始まらず、再起動後も始まりません。' +
    'プロファイルを切り替えると終了します。',
  horchenAdresse: '待ち受けるアドレス',
  horchenAdresseBeispiel: ':41234',
  horchenAdresseHinweis:
    'ポートだけ (:41234) を指定すると、このマシンのすべてのインターフェースで待ち受けます。' +
    '前に 127.0.0.1 を付けると、このマシン自身でのみ待ち受けます。',
  horchenStarten: 'アクセスを受け付ける',
  horchenBeenden: 'アクセスの受け付けを停止',
  horchenLaeuft: '{adresse} で待ち受け中',
  horchenNamen: '名前 (アドレスが変わっても保たれます)',
  horchenAdressen: 'アドレス',
  horchenNurHier:
    'このバインドでは、外部からは接続できません。127.0.0.1 はこのマシンでのみ待ち受けます。このMacをネットワーク上で接続可能にするには、ポートだけ (:41234) を指定してください。',
  horchenNurDiesesNetz:
    'どちらも、このネットワークでは有効です。割り当てられたアドレスは変わることがありますが、名前は通常保たれます。どこでも有効なアドレスは、ネットワークをまたぐ接続で初めて得られます。',
  horchenStill: 'このプロファイルは、現在待ち受けていません。',
  horchenWeitergeben:
    'これが相手に必要な情報です。このプロファイルの鍵も一緒に必要です。',
  sichtbarTitel: '近くで見つけてもらう',
  sichtbarErklaerung:
    'アナウンスすると、同じネットワーク上のすべての相手に、' +
    'このプロファイルの存在がピアIDとともに伝わります。' +
    '接続可能にすることと、見つけてもらうことは別々の判断で、これは2番目の判断です。',
  sichtbarName: 'アナウンスする名前',
  sichtbarNameHinweis:
    '自分で付けた名前は何も証明しません。このMacのホスト名は送信されません。',
  sichtbarWerden: 'アナウンス',
  sichtbarBeenden: 'アナウンスを停止',
  sichtbarLaeuft: 'このプロファイルはアナウンス中です。',
  sichtbarStill: 'このプロファイルはアナウンスしていません。',
  sichtbarBrauchtHorchen:
    '先に、アクセスを受け付ける設定にしてください。' +
    'ポートなしでアナウンスしても、誰もいないドアへの招待になるだけです。',
  nachbarnTitel: '近くにいる相手',
  nachbarnErklaerung:
    '今、同じネットワーク上でアナウンスしているピアです。' +
    '表示されているのは相手の申告であり、確認されるのは接続したときです。',
  nachbarnSuchen: '探す',
  nachbarnSucht: '探しています…',
  nachbarnLeer: '今、アナウンスしている相手はいません。',
  nachbarnUnbekannt: '不明',
  nachbarnAufnehmen: '連絡先に追加',
  nachbarnGemerkt: '記憶した名前:',
  nachbarnSchonBekannt: 'すでに一覧にあります',
  // Ein bekannter Peer an einer neuen Adresse (NET-02.NAT-02, #1706).
  nachbarnAdresseNeu: 'このピアは新しいアドレスにいます。',
  nachbarnAdresseBisher: 'これまで記憶していたアドレス',
  nachbarnAdresseKeine: 'なし',
  nachbarnAdresseUebernehmen: '新しいアドレスを使う',
  nachbarnAdresseHinweis:
    '変わるのは接続先だけです。このピアへの信頼が以前より増えることはなく、鍵は接続のたびに確認されます。',
  nachbarnVerworfen:
    '{zahl}件の応答にピアIDがなかったため、一覧には含まれていません。',

  // **How this program goes online** — N5
  // (#1571).
  netzwegTitel: 'このプロセスの環境から',
  netzwegErklaerung:
    'これらの値は、プログラムが起動された環境から取得されます。下のフィールドより優先され、ウインドウからは変更できません。',
  netzwegProxyHTTP: 'httpのプロキシ',
  netzwegProxyHTTPS: 'httpsのプロキシ',
  netzwegProxyAus: 'プロキシを使わない対象',
  netzwegAusUmgebung: '{name}: {wert}',

  // **The two switchable services** — N6
  // (#1572).
  dienstRESTTitel: 'スクリプト用のインターフェース',
  dienstRESTErklaerung:
    '同じデータベースを使うWebサービスです。スクリプトから読み取ったり変更したりできます。許可している間だけ動作し、このウインドウとともに終了します。',
  dienstMCPTitel: 'HTTP経由のツールサーバ',
  dienstMCPErklaerung:
    '標準入力経由と同じツールを、アドレスで公開します。他のデバイスからも接続できます。',
  dienstRESTAdresse: 'スクリプト用インターフェースのアドレス',
  dienstRESTPort: 'スクリプト用インターフェースのポート',
  dienstMCPAdresse: 'ツールサーバのアドレス',
  dienstMCPPort: 'ツールサーバのポート',
  dienstMarke: 'アクセストークン',
  dienstMarkeKopieren: 'アクセストークンをコピー',
  dienstFingerabdruck: 'フィンガープリント',
  dienstFingerabdruckKopieren: 'フィンガープリントをコピー',
  dienstKopiert: 'コピーしました',
  dienstStarten: '開始',
  dienstBeenden: '停止',
  dienstErreichbar: '接続先:',
  dienstMitTLS:
    '暗号化されています。証明書は自己発行で、このプロファイルのIDを含みます。接続する側は、そのフィンガープリントを照合します。',
  dienstOhneTLS:
    '暗号化されていませんが、ここでは問題ありません。このアドレスはこのマシン上でのみ有効です。',
  dienstHinweis:
    '127.0.0.1 はこのマシンのみを意味します。0.0.0.0 はネットワーク上のすべてのデバイスを意味し、その場合は暗号化とアクセス制御が必須で、どちらも自動で設定されます。ポート0は、空いているポートを探すことを意味します。',
}
