import { dienstkatalog } from './einstellungenkatalog'
import { einstellungsbelegung } from './einstellungsbelegung'
import { einstellungsgruppennamen } from './einstellungsgruppennamen'
import {
  schalternamen,
  schalterwerte,
  seitenschalter,
} from './einstellungenschalter'

// **Die Texte der Einstellungsseite** — ausgelagert für Phase 4.8
// (#1321).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als
// einer ankommt** (§9) — dieselbe Gestalt wie `seitenkarten.ts`. Mit den
// Dienstschaltern ging `seiten.ts` über die Zeilengrenze aus §2.
import { ki } from './seitenki'
import { netzbereich, netzhilfe, netznamen } from './einstellungennetz'
import { handelshilfe, handelsnamen, handelswerte } from './einstellungenhandel'
import {
  gestaltungsbeiwerk,
  gestaltungshinweise,
  gestaltungsnamen,
  gestaltungswerte,
} from './einstellungengestaltung'
import { zugaenge } from './einstellungenzugaenge'
import { ansichtsnamen, ansichtswerte } from './einstellungenansicht'
import { vorgabennamen } from './einstellungenvorgaben'
import {
  protokollhinweise,
  protokollnamen,
  protokollwerte,
} from './einstellungenprotokoll'
import { austausch } from './einstellungenaustausch'
import { eigenuebersetzung } from './eigenuebersetzung'
import { kihinweise, kinamen, kiwerte } from './einstellungenki'

export const einstellungen = {
  ...einstellungsgruppennamen,
  ...seitenschalter,

  titel: '設定',

  namen: {
    ...schalternamen,
    sprache: '言語',
    'sprache.eigene': '独自の翻訳',
    darstellung: '地域の形式',
    zeitzone: 'タイムゾーン',
    'filter.gespeichert': '保存済みのフィルタ',
    'kuerzel.ziffern': '⌘0〜⌘9 の割り当て',
    erscheinungsbild: 'テーマ',
    schriftgroesse: '文字サイズ',
    aktualisierung: '新しいバージョンの確認',
    ersteschritte: '起動時の紹介',
    seitenleiste: 'サイドバー',
    leistenlage: 'サイドバーの位置',
    fensterstart: '起動時のウインドウ',
    fensterbreite: '最後に記憶した幅',
    fensterhoehe: '最後に記憶した高さ',
    schriftart: '表示書体',
    feldtiefe: 'カードフォーム',
    'tisch.autoanordnung': 'テーブル上のカードを自動的に配置',

    'barrierefreiheit.bewegung': 'モーション',
    'barrierefreiheit.kontrast': 'コントラスト',
    'sicherheit.faktorbeimwechsel': 'プロファイル切り替え時の第2要素',

    ...vorgabennamen,
    ...protokollnamen,
    // **Die lokale KI-Laufzeit** — A7, #1446.
    // **A cloud provider's address** — piece 3 of A18 (#1533), labelled with
    // #1540.
    'dienst.anthropic.adresse': 'Anthropic のアドレス',
    'dienst.openai.adresse': 'OpenAI のアドレス',
    'dienst.xai.adresse': 'xAI のアドレス',
    ...netznamen,
    ...handelsnamen,
    ...gestaltungsnamen,
    ...ansichtsnamen,
    'konto.schau': '自分のマーケットアカウントを読み取る',
    ...kinamen,
    // **The client id per provider** — Werkstück 4 of Phase 4.12 (#1351).
    'anbieter.github.kennung': 'GitHub のクライアント ID',
    'anbieter.google.kennung': 'Google のクライアント ID',
    'anbieter.microsoft.kennung': 'Microsoft のクライアント ID',
    'anbieter.x.kennung': 'X のクライアント ID',

    // **The workspace and the two limits** — Werkstück O3, Phase 5.2 (#1427).
    'arbeitsplatz.navigator': '左の列',
    'arbeitsplatz.inspektor': '右の列',
    'arbeitsplatz.navigatorbreite': '左の列の幅',
    'arbeitsplatz.inspektorbreite': '右の列の幅',
    'richtlinie.nurlokal': 'ローカルのみで動作',
    'richtlinie.kostengrenze': '費用の上限（最小単位。EURの場合はセント）',
  },

  // Origin, default and search — work item U11 of phase 5.1 (#1422).
  herkunftMarke: '設定元：{herkunft}',
  wirksamKurz: '有効なのは「{wert}」です。',
  wirksamOffenKurz: '有効なのはコンピュータの初期設定です。',
  wirksamGehalten: '有効なのは「{wert}」です。ポリシーで固定されています。',
  herkunftZeile: '設定元：{herkunft}、レベル：{ebene}。',
  herkunftNamen: {
    profil: 'このプロファイル',
    nutzer: 'あなた',
    geraet: 'このデバイス',
    vorgabe: '誰も設定していません（初期設定）',
  },
  ebenenNamen: {
    profil: 'プロファイル',
    nutzer: 'ユーザ',
    geraet: 'デバイス',
    arbeitsplatz: 'ワークスペース',
  },
  ebenenErklaerung:
    'どの設定も、デバイス、ユーザ、プロファイル、ワークスペースのいずれかのレベルで有効になります。**デバイス**はこのコンピュータ自体（パスやウインドウのサイズなど）を表し、設定の書き出しには含まれません。**ユーザ**はあなたのすべてのプロファイルで有効です。**プロファイル**はそのプロファイルでのみ有効で、ユーザとして設定した内容よりも優先されます。**ワークスペース**は、このウインドウの列の配置です。有効になるのは、プロファイルの設定、なければユーザとしての設定、なければ初期設定の順です。ポリシーの場合は、最も厳しい値が有効です。どの設定にも管理者権限は必要ありません。',
  fachleute: '上級者向けのその他の設定',
  fachleuteSatz:
    'タイムアウト、他のネットワーク経由での到達性、外部への経路、他のプログラム向けのサービスなどです。展開しても何も変わりません。表示が増えるだけです。',
  zuruecksetzen: '「{vorgabe}」に戻す',
  erbeMitWert: '{erbe}（{wert}）',
  zuruecksetzenVon: '{name} を初期設定に戻す',
  suche: '設定を検索',
  suchePlatzhalter: '名前、キー、または値',
  sucheTreffer: '{zahl} 件一致',
  sucheKeine: '「{frage}」に一致する設定はありません。',
  zumGebiet: '「{gebiet}」の領域で表示',

  // **Below the appearance settings** — UX-09 (#1681).
  gestaltung: gestaltungsbeiwerk,

  hinweise: {
    ...protokollhinweise,
    'sprache.eigene':
      'ウインドウで使う独自の翻訳です。空欄の場合は使いません。その翻訳がなくなった場合は、選択した言語が再び使われます。',
    feldtiefe:
      'カードフォームにどこまで表示するかを指定します。**標準**は入力する項目です。**詳細**はゲームパッケージのスクリプトが計算する項目を加え、**エキスパート**はゲームパッケージが何も説明していない項目（`stylesheet` など）まで加えます。この3つは入れ子になっていて、深いほど項目が増えるだけで、別のものになることはありません。プロファイルごとに適用されます。',
    sprache:
      'ウインドウ全体とメニューバーに適用されます。MSE パッケージに由来するもの（フィールド名、ゲーム名など）は、ファイルに書かれたままです。 ドイツ語が原文の言語です。その他の言語は機械翻訳で、まだ校閲されていません。',
    darstellung:
      '数値、金額、日付の表示形式です。たとえば **de-CH** や **en-GB** を指定します。空欄の場合はシステムの設定に従います。テキストの言語には影響しません。',
    zeitzone:
      '時刻を表示するタイムゾーンです。たとえば **Europe/Berlin** を指定します。「システムに合わせる」を選ぶと、コンピュータ自身の設定になります。プロファイルごとに保存され、設定の書き出しには含まれません。',
    'filter.gespeichert':
      'すべてのビューの保存済みフィルタを JSON のリストで表します。名前、ビュー、マーク、**and** か **or** のどちらで結ぶかを含みます。フィルタは各ビューのナビゲータで作成します。プロファイルに属し、設定の書き出しに含まれます。',
    'kuerzel.ziffern':
      '⌘ と数字キーで開くページです。空欄の場合は、サイドバーの各領域になります。「一般」のショートカット一覧で設定します。',
    erscheinungsbild:
      'ライト、ダーク、またはコンピュータの設定に合わせます。この選択はこのアプリケーションにのみ適用されます。',
    schriftgroesse:
      '文字だけでなくインターフェイス全体を拡大縮小します。余白や罫線も一緒に大きくなります。カードのプレビューは変わりません。印刷される状態を表示するためです。',
    aktualisierung:
      '初期設定は**オフ**です。ネットワークには何も送信されず、何も問い合わせません。**通知のみ**にすると、起動時に新しいバージョンがあるかを確認し、あればお知らせします。それ以上のことはしません。**通知してダウンロード**にすると、さらにボタンが表示され、押すとバージョンをダウンロードフォルダに取得してチェックサムを照合します。自動的にインストールされることはありません。',
    ersteschritte:
      '最初のカードを作るまでの3つのステップを、起動時のウインドウで表示します。同じスイッチがそのウインドウの中にもあり、そこでオフにしても、ここで再びオンにできます。',
    seitenleiste:
      'アプリケーションの各領域が並ぶ、左側のバーです。折りたたむと、その幅がワークスペースに回ります。⌥⌘S、バーの端のボタン、**表示**メニューの項目で折りたたみ、展開できます。プロファイルごとに保存され、設定の書き出しには含まれません。',
    leistenlage:
      '左側では、コンテンツの隣に縦に表示されます。上側では、コンテンツの上に横1列で表示されます。項目、ショートカット、矢印キーは同じで、向きが横になるだけです。プロファイルごとに保存され、設定の書き出しには含まれません。',
    fensterstart:
      '初期設定は**画面いっぱいに表示**です。ウインドウは使える領域いっぱいに広がりますが、**フルスクリーンにはしません**。フルスクリーンではメニューバーも隠れてしまうためです。**標準サイズ**は組み込みのサイズで、**前回と同じ**は最後に閉じたときの状態で開きます。プロファイルごとに保存され、設定の書き出しには含まれません。',
    fensterbreite:
      'ウインドウを最後に閉じたときに測った値です。**「前回と同じ」の場合にのみ**使われます。空欄の場合はまだ前回がないことを表し、ウインドウは組み込みのサイズで開きます。',
    fensterhoehe:
      '高さについても同様です。どちらもウインドウを閉じるときに書き込まれます。700〜8,192 ポイントの範囲外の値は記憶されません。次の画面に収まらないウインドウは、つかめなくなってしまうためです。',
    schriftart:
      '見出しとセット名の書体です。本文はシステムフォントのままです。段落全体に装飾書体を使うと読みにくくなります。4種類はすべて macOS に付属しており、何もダウンロードされません。',
    'barrierefreiheit.bewegung':
      'この上書きは**一方向のみ**です。モーションをオフにすることはできても、オンにすることはできません。システムで**視差効果を減らす**を設定している場合は、ここでも自動的に適用されます。それを上書きするアプリケーションは、ユーザーの意図に反することになるためです。',
    'barrierefreiheit.kontrast':
      '同じ方向です。**高**にすると、繊維のテクスチャを取り除き、すべての文字が 7 : 1 を満たすまで色を調整します。システムでコントラストを高くするよう設定している場合も同じ結果になります。',
    'sicherheit.faktorbeimwechsel':
      '制限を**緩める**唯一の上書きで、プロファイルを**切り替える**ときにのみ有効です。サインインには影響しません。サインインでは、第2要素が本来の役割を果たします。1人で使うコンピュータで、切り替えのたびにワンタイムコードを求められるのは、利点のない手間です。',
    'vorgabe.zustand':
      'コレクションに新しい行を追加するときの初期値です。**初期値なし**の場合、フィールドはこれまでどおり空のままになります。',
    'vorgabe.waehrung':
      '新しい行の通貨フィールドの初期値です。たとえば **EUR** を指定します。換算は行われません。',
    'vorgabe.einzelspiel':
      '新しいトーナメントで、**1ゲーム**の対戦を行うかどうかを指定します。決まるのは表示だけで、計算は変わりません。ゲーム勝率と対戦相手の勝率は、どちらの場合も保持されます。',
    'vorgabe.hochladepfad':
      'ファイルを開くダイアログが最初に表示するフォルダです。画像、MTG JSON、そのほかディスクから読み込むものすべてに適用されます。フルパス、または **~/** で始まるパスを指定します。空欄の場合は、オペレーティングシステムが開く場所になります。',
    'vorgabe.ausfuhrpfad':
      '操作に固有の場所がない場合に、保存ダイアログが最初に表示するフォルダです。保存時にフォルダが見つからない場合（外付けドライブが接続されていないときなど）は、別の場所に保存せず、アプリケーションがそのことをお知らせします。',
    'vorgabe.sicherungspfad':
      'ライブラリのバックアップ用ダイアログが最初に表示するフォルダです。フォルダが見つからない場合は、メッセージを表示してバックアップを中止します。',
    'vorgabe.diagnosepfad':
      '診断用のテキストやバンドルを保存するフォルダです。フォルダが見つからない場合は、メッセージを表示して中止します。',
    'bildausfuhr.ordner':
      '画像の書き出し先として最後に使ったフォルダです。「画像」ページはこのフォルダを提案し、新しく選んだフォルダをここに記録します — MSE と同じく、すべてのセットで共通です。空の場合、フォルダの選択はファイルを書き出すフォルダから始まります。',
    'ausgabe.ziele':
      '「書き出し」が各セットで最後に表示した出力先です。JSON のリストで、最新の選択が末尾、最大 200 件。「書き出し先」で選ぶと書き込まれます。空の場合、何も事前選択されません。',
    'dienst.anthropic.adresse':
      'リクエストの送信先です。空欄の場合は Anthropic 自身のインターフェイスになります。別のアドレスを指定すると**ゲートウェイ**、つまりリクエストを転送するサービスになります。そのキーは別に保存されます。',
    'dienst.openai.adresse':
      'OpenAI についても同様です。空欄の場合は `https://api.openai.com` になります。ここに何かを入力すると、OpenAI ではなく、そのアドレスで応答するものと通信することになります。',
    'dienst.xai.adresse':
      'xAI についても同様です。空欄の場合は `https://api.x.ai` になります。',
    ...netzhilfe,
    ...handelshilfe,
    ...gestaltungshinweise,
    'konto.schau':
      'このプログラムが、マーケットプレイスにある**自分の**アカウント（出品と注文）を読み取ってよいかどうかを指定します。**オフ**の場合は、一切見ません。マーケットプレイスをオンにするだけでは足りません。そちらは**他の**出品者が問い合わせる内容に答え、こちらは**自分側**にあるものを扱います。出品、カート、購入など、何かがそこに書き込まれることはありません。',
    ...kihinweise,
    'anbieter.github.kennung':
      'GitHub で作成した OAuth アプリケーションの `client_id` です。空欄の場合、GitHub でのサインインは提供されません。',
    'anbieter.google.kennung':
      'Google で作成した OAuth アプリケーションの `client_id` です。通常は `apps.googleusercontent.com` で終わります。',
    'anbieter.microsoft.kennung':
      'Microsoft Entra ポータルのアプリケーション ID（UUID）です。Microsoft は、このフローでシークレットを**必要としません**。',
    'anbieter.x.kennung':
      'X の開発者ポータルで作成したアプリケーションの `client_id` です。',
  },

  werte: {
    ...schalterwerte,
    ...protokollwerte,
    ...handelswerte,
    ...gestaltungswerte,
    ...ansichtswerte,
    // **Die lokale KI-Laufzeit** — A7, #1446.
    // **Die Kontoschau** — T11, #1476. Dieselben Worte wie bei der
    // KI-Laufzeit darüber: zwei Schreibweisen für denselben Zustand
    // lernt niemand gern.
    'konto.schau.aus': 'オフ',
    'konto.schau.an': 'オン',
    ...kiwerte,
    'sprache.erben': 'ユーザの設定に従う',
    'erscheinungsbild.erben': 'ユーザの設定に従う',
    'schriftgroesse.erben': 'ユーザの設定に従う',
    'schriftart.erben': 'ユーザの設定に従う',
    'sprache.de': 'Deutsch',
    'sprache.en': 'English',
    'sprache.es': 'Español',
    'sprache.fr': 'Français',
    'sprache.it': 'Italiano',
    'sprache.ja': '日本語',
    'sprache.ru': 'Русский',
    'sprache.zh': '简体中文',
    'erscheinungsbild.system': 'システムに合わせる',
    'erscheinungsbild.hell': 'ライト',
    'erscheinungsbild.dunkel': 'ダーク',
    'schriftgroesse.klein': '小',
    'schriftgroesse.normal': '標準',
    'schriftgroesse.gross': '大',
    'schriftgroesse.sehrgross': '特大',
    'aktualisierung.aus': 'オフ',
    'aktualisierung.melden': '通知のみ',
    'aktualisierung.holen': '通知してダウンロード',
    'ersteschritte.an': 'オン',
    'ersteschritte.aus': 'オフ',
    'seitenleiste.sichtbar': '表示',
    'seitenleiste.verborgen': '非表示',
    'leistenlage.senkrecht': '左',
    'leistenlage.waagerecht': '上',
    'fensterstart.maximiert': '画面いっぱいに表示',
    'fensterstart.normal': '標準サイズ',
    'fensterstart.zuletzt': '前回と同じ',

    'feldtiefe.normal': '標準',
    'feldtiefe.erweitert': '詳細',
    'feldtiefe.experte': 'エキスパート',
    'tisch.autoanordnung.an': 'はい — テーブルが自動的に並べます',
    'tisch.autoanordnung.aus': 'いいえ — 置いた場所にそのまま残ります',
    'tisch.beschriftung': 'テーブル上のラベル',
    'tisch.beschriftung.aus': 'オフ — 枚数のみ',
    'tisch.beschriftung.klein': '小',
    'tisch.beschriftung.normal': '中',
    'tisch.beschriftung.gross': '大',
    'schriftart.pergament': '羊皮紙 — 落ち着いていて読みやすい',
    'schriftart.fantasy': 'ファンタジー — 遊び心がある',
    'schriftart.antike': '古代 — ローマン・キャピタル',
    'schriftart.kalligrafie': 'カリグラフィ — 手書き風',
    'barrierefreiheit.bewegung.system': 'システムに合わせる',
    'barrierefreiheit.bewegung.reduziert': '常に控えめ',
    'barrierefreiheit.kontrast.system': 'システムに合わせる',
    'barrierefreiheit.kontrast.hoch': '常に高い',
    'sicherheit.faktorbeimwechsel.verlangen': '切り替え時も',
    'sicherheit.faktorbeimwechsel.nurAnmelden': 'サインイン時のみ',
    'vorgabe.zustand.ohne': '初期値なし',
    'vorgabe.zustand.mint': 'Mint',
    'vorgabe.zustand.nearMint': 'Near Mint',
    'vorgabe.zustand.excellent': 'Excellent',
    'vorgabe.zustand.good': 'Good',
    'vorgabe.zustand.lightPlayed': 'Light Played',
    'vorgabe.zustand.played': 'Played',
    'vorgabe.zustand.poor': 'Poor',
    'vorgabe.einzelspiel.aus': 'オフ',
    'vorgabe.einzelspiel.an': 'オン',

    // **The workspace and the limit** — Werkstück O3 of Phase 5.2 (#1427).
    'arbeitsplatz.navigator.erben': 'ユーザの設定に従う',
    'arbeitsplatz.navigatorbreite.erben': 'ユーザの設定に従う',
    'arbeitsplatz.inspektorbreite.erben': 'ユーザの設定に従う',
    'arbeitsplatz.navigator.sichtbar': '表示',
    'arbeitsplatz.navigator.verborgen': '非表示',
    'arbeitsplatz.inspektor.erben': 'ユーザの設定に従う',
    'arbeitsplatz.inspektor.sichtbar': '表示',
    'arbeitsplatz.inspektor.verborgen': '非表示',
    'richtlinie.nurlokal.aus': 'オフ',
    'richtlinie.nurlokal.an': 'オン',
  },

  // **Was dem Menschen gehört** — Werkstück 8 von Phase 4.12 (#1351).
  // **Where a login comes from in the first place** — follow-up to Werkstück 4
  // of Phase 4.12 (#1351).
  anmeldung: 'プロバイダを使ったサインイン',
  anmeldungHinweis:
    'このアプリケーションには**認証情報が同梱されていません**。GitHub、Google、Microsoft、X でサインインするには、プロバイダ側で自分で **OAuth アプリケーション**を作成し、その `client_id` をここに入力します。リダイレクト URI には、サインイン中にブラウザに表示されるアドレスを指定します。このアドレスはこのコンピュータを指し、サインインが行われている間だけ有効です。',

  anmeldungGeheimnis:
    'GitHub では、追加で `client_secret` が必要です。Google では任意で、Microsoft ではこのようなプログラムには使用が禁止されています。シークレットはここではなく**キーチェーン**に保存します。ライブラリは共有されることがありますが、キーチェーンは共有されません。**キーチェーンアクセス**で**アプリケーションパスワード**として、名前を `MagicSetEditor4Mac-Anmeldung`、アカウントを `github` にして作成します。コマンドラインでは：`security add-generic-password -s MagicSetEditor4Mac-Anmeldung -a github -w`',

  geheimnis: 'シークレット',
  geheimnisPflicht: 'は必須ですが、設定されていません',
  geheimnisVerboten:
    'は不要で、指定しても受け付けられません。ここではクライアント ID だけで足ります。',

  anbietername: {
    github: 'GitHub',
    google: 'Google',
    microsoft: 'Microsoft',
    x: 'X',
  },

  wirktNachNeustart: '再起動後に反映されます。',

  trefferInGruppe: '「{frage}」に一致：{zahl} 件',
  gesuchtMarke: '一致',
  sucheSprung: '最初の一致箇所へ移動',

  vorgaben: '初期値',

  // **The pointer to the help** — Werkstück 13 of Phase 4.12 (#1351).

  dienste: '外部のカードサービス',
  diensteHinweis:
    'オフにすると、このサービスにはリクエストが一切送信されません。アプリケーションはサービスがなくても動作し続けます。外部から取得するものだけが取得できなくなります。',

  kiAnbieter: '言語モデル',
  kiAnbieterHinweis:
    '質問を送ってよい送信先です。どのプロバイダも自社のコンピュータで処理を行い、送信した内容を見ることができます。そのため、プロバイダごとに個別にオンにする必要があり、初めからオンになっているものはありません。',
  // See the German catalogue for why these two live here and not under
  // `werte`: they are the same for every service.
  dienstAus: 'オフ',
  dienstAn: 'オン',
  dienstUngeprueft: '実際のサービスに対して実行されたことがありません',
  dienstUngeprueftHinweis:
    'このサービスのアカウントは、ここにはありません。検証されているのは、こちら側の境界、つまりリクエスト、ヘッダ、マッピングです。サービスが想定どおりに応答するかどうかは、計測されていません。',
  fragen: '誰が何に答えるか',
  fragenHinweis:
    '複数のサービスが同じ質問に答えられる場合は、ここで問い合わせ先を決めます。**別のサービスへの切り替えは行われません**。選択したサービスが応答しない場合は、その旨をお知らせします。別の情報源からの答えが、気づかないうちに表示されることはありません。',
  'frage.suche': 'カード検索',

  // The service catalogue — work item A1 of phase 5.3 (#1440).
  ...dienstkatalog,

  dienstSchluesselFehlt: '認証情報が保存されていません',
  dienstSchluesselHinweis:
    'このサービスには認証情報が必要です。認証情報は、コレクションファイルではなく macOS のキーチェーンに保存されます。コレクションファイルは共有されることがあるためです。',
  zugaenge: '認証情報',
  zugaengeHinweis:
    '認証情報は macOS のキーチェーンに保存され、コレクションファイルには保存**されません**。コレクションファイルは共有されることがあるためです。ここに表示されるのは、何かが保存されているか**どうか**だけで、値そのものが返されることはありません。',
  // See the German catalogue: these name the headers each service asks for
  // (documented in decision 0056).

  kuerzelUeberschrift: 'キーボードショートカット',
  ...einstellungsbelegung,
  kuerzelHinweis:
    'ウインドウ全体で有効です。テキストフィールドの中では、⌘Z と ⌘F はそのフィールドの操作になります。',
  kuerzelZweiteReihe:
    '⌥ を使う8つのショートカットは、最上段のキーに左から割り当てられています。AZERTY など、配列の異なるキーボードでは別の文字が印字されていますが、キーは同じです。',
  kuerzelTaste: 'キー',
  kuerzelHandlung: '操作',

  austausch,
  eigenuebersetzung,
  // **An account's remaining quota at a provider** — A5, #1444.
  kontingentTitel: 'プロバイダの利用枠',
  kontingentErklaerung:
    'プランに含まれるクレジット数と、残りのクレジット数です。確認はネットワーク経由で行われ、それ自体が1クレジットを消費します。',
  kontingentHolen: '利用枠を確認',
  kontingentHolenBei: '{dienst} の利用枠を確認',
  kontingentLaeuft: '確認しています…',
  kontingentZahlen: '{gesamt} クレジットのうち、残りは {rest} クレジットです。',
  kontingentFrist:
    'この数値は、プロバイダ側で最大 {minuten} 分遅れることがあります。上限ではありません。ここで実際に支出を制限するのは、このライブラリの予算です。',
  kontingentUeberzug: 'プランを超える1クレジットの費用は {satz} です。',
  // **What is known about AI models** — A6, #1445.
  ...ki,

  // **Die Zugangsdaten** (#1534) — eigene Datei, ein Katalog (§9).
  ...zugaenge,

  ...netzbereich,
}
