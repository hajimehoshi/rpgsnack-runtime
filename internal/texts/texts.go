// Copyright 2017 Hajime Hoshi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package texts

import (
	"golang.org/x/text/language"
)

type TextID int

const (
	TextIDNewGame TextID = iota
	TextIDResumeGame
	TextIDNewGameWarning
	TextIDYes
	TextIDNo
	TextIDOK
	TextIDSettings
	TextIDLanguage
	TextIDCredit
	TextIDRemoveAds
	TextIDRemoveAdsDesc
	TextIDReviewThisApp
	TextIDRestorePurchases
	TextIDMoreGames
	TextIDBackToTitle
	TextIDClose
	TextIDItemCheck
	TextIDQuitGame
	TextIDStoreError
)

func Text(lang language.Tag, id TextID) string {
	if lang == language.Und {
		lang = language.English
	}
	return texts[lang][id]
}

var texts = map[language.Tag]map[TextID]string{
	language.English: {
		TextIDNewGame:    "New Game",
		TextIDResumeGame: "Resume Game",
		TextIDYes:        "Yes",
		TextIDNo:         "No",
		TextIDOK:         "OK",
		TextIDSettings:   "Settings",
		TextIDLanguage:   "Language",
		TextIDCredit:     "Credits",
		TextIDRemoveAds:  "Remove Ads",
		TextIDRemoveAdsDesc: `Would you like to remove ads
from the game for %s?`,
		TextIDReviewThisApp:    "Review this App",
		TextIDRestorePurchases: "Restore Purchases",
		TextIDMoreGames:        "More Games",
		TextIDClose:            "Close",
		TextIDItemCheck:        "Check",

		TextIDNewGameWarning: `You have on-going game data.
Do you want to reset your
progress and start a new game?`,
		TextIDBackToTitle: "Are you sure you want to 
go back to the title screen?",
		TextIDQuitGame:    "Are you sure you want to 
go quit the game?",
		TextIDStoreError: `Failed to connect to the store.
Please make sure to sign in
and connect to the network.`,
	},
	language.German: {
		TextIDNewGame:    "Neues Spiel",
		TextIDResumeGame: "Spiel forsetzen",
		TextIDYes:        "Ja",
		TextIDNo:         "Nein",
		TextIDOK:         "OK",
		TextIDSettings:   "Einstellungen",
		TextIDLanguage:   "Sprache",
		TextIDCredit:     "Danksagungen",
		TextIDRemoveAds:  "Anzeigen entfernen",
		TextIDRemoveAdsDesc: `Willst du Anzeigen für %s
enfernen?`,
		TextIDReviewThisApp:    "Rezension schreiben",
		TextIDRestorePurchases: "Einkäufe wiederherstellen",
		TextIDMoreGames:        "Mehr Spiele",
		TextIDClose:            "Zurück",
		TextIDItemCheck:        "Info",

		TextIDNewGameWarning: `Willst du wirklich deinen Spielfortschritt
löschen und nochmal von Vorne anfangen?`,
		TextIDBackToTitle: "Willst du wirklich zurück zum Menü?",
		TextIDQuitGame:    "Willst du wirklich das Spiel verlassen?",
		TextIDStoreError: `Verbindung mit dem Store nicht möglich.
Stelle sicher, dass du angemeldet
und mit dem Internet verbindet bist.`,
	},
	language.Spanish: {
		TextIDNewGame:    "Nuevo Juego",
		TextIDResumeGame: "Reanudar Juego",
		TextIDYes:        "Sí",
		TextIDNo:         "No",
		TextIDOK:         "OK",
		TextIDSettings:   "Configuraciones",
		TextIDLanguage:   "Idioma",
		TextIDCredit:     "Créditos",
		TextIDRemoveAds:  "Remover anuncios",
		TextIDRemoveAdsDesc: `¿Te gustaría pagar %s
para quitar los anuncios del juego?`,
		TextIDReviewThisApp:    "Puntúa esta App",
		TextIDRestorePurchases: "Restaurar Compra",
		TextIDMoreGames:        "Más Juegos",
		TextIDClose:            "Cerrar",
		TextIDItemCheck:        "Revisar",

		TextIDNewGameWarning: `Tienes datos del juego en curso.
¿Quieres eliminar el progreso 
e iniciar un nuevo juego?`,
		TextIDBackToTitle: "¿Quieres volver al título?",
		TextIDQuitGame:    "¿Quieres salir del juego?",
		TextIDStoreError:  `Fallo al conectarse con la tienda. Por favor asegúrate de iniciar sesión y conectarse a internet`,
	},
	language.Japanese: {
		TextIDNewGame:    "はじめから",
		TextIDResumeGame: "つづきから",
		TextIDYes:        "はい",
		TextIDNo:         "いいえ",
		TextIDOK:         "OK",
		TextIDSettings:   "設定",
		TextIDLanguage:   "言語",
		TextIDCredit:     "クレジット",
		TextIDRemoveAds:  "広告を消す",
		TextIDRemoveAdsDesc: `%sを支払って、
広告を消去しますか？`,
		TextIDReviewThisApp:    "このアプリをレビューする",
		TextIDRestorePurchases: "購入情報のリストア",
		TextIDMoreGames:        "ほかのゲーム",
		TextIDClose:            "閉じる",
		TextIDItemCheck:        "チェック",

		TextIDNewGameWarning: `進行中のゲームデータがあります。
進行中のゲームデータを消して、
新しいゲームを開始しますか?`,
		TextIDBackToTitle: "タイトル画面にもどりますか？",
		TextIDQuitGame:    "ゲームを終了しますか？",
		TextIDStoreError: `ストアへの接続に失敗しました。
ネットワークに接続しているか
確認してください`,
	},
	language.SimplifiedChinese: {
		TextIDNewGame:    "新游戏",
		TextIDResumeGame: "继续游戏",
		TextIDYes:        "确定",
		TextIDNo:         "取消",
		TextIDOK:         "OK",
		TextIDSettings:   "设定",
		TextIDLanguage:   "语言",
		TextIDCredit:     "制作人员",
		TextIDRemoveAds:  "移除广告",
		TextIDRemoveAdsDesc: `你希望支付%s
来移除游戏里的广告吗?`,
		TextIDReviewThisApp:    "点评我们的游戏",
		TextIDRestorePurchases: "恢复购买",
		TextIDMoreGames:        "更多游戏",
		TextIDClose:            "关闭",
		TextIDItemCheck:        "查看",

		TextIDNewGameWarning: `系统已经存在一个中断存档。
开始新游戏会导致中断存档被清除。
你确定要重新开始新游戏吗?`,
		TextIDBackToTitle: "返回主菜单?",
		TextIDQuitGame:    "退出游戏?",
		TextIDStoreError: `无法连接商店。
请确定你已经登录并已连上网络`,
	},
	language.TraditionalChinese: {
		// TODO: Translat this
		TextIDNewGame:    "新遊戲",
		TextIDResumeGame: "繼續遊戲",
		TextIDYes:        "確定",
		TextIDNo:         "取消",
		TextIDOK:         "OK",
		TextIDSettings:   "設定",
		TextIDLanguage:   "語言",
		TextIDCredit:     "製作人員",
		TextIDRemoveAds:  "移除廣告",
		TextIDRemoveAdsDesc: `你希望支付%s
來移除遊戲裡的廣告嗎?`,
		TextIDReviewThisApp:    "點評我們的遊戲",
		TextIDRestorePurchases: "恢復購買",
		TextIDMoreGames:        "更多遊戲",
		TextIDClose:            "關閉",
		TextIDItemCheck:        "查看",

		TextIDNewGameWarning: `系統已經存在一個中斷存檔。
開始新遊戲會導致中斷存檔被清除。
你確定要重新開始新遊戲嗎？`,
		TextIDBackToTitle: "返回主菜單?",
		TextIDQuitGame:    "退出遊戲?",
		TextIDStoreError: `無法連接商店。
請確定你已經登錄並已連上網絡`,
	},
}
