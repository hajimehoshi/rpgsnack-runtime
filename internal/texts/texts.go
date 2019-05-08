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
	TextIDResetGame
	TextIDResumeGame
	TextIDNewGameWarning
	TextIDYes
	TextIDNo
	TextIDOK
	TextIDSettings
	TextIDLanguage
	TextIDCredits
	TextIDCreditsEntry
	TextIDRemoveAds
	TextIDRemoveAdsDesc
	TextIDReviewThisApp
	TextIDRestorePurchases
	TextIDMoreGames
	TextIDPrivacyPolicy
	TextIDMenu
	TextIDBGMVolume
	TextIDSEVolume
	TextIDAdvancedSettings
	TextIDBackToTitle
	TextIDBack
	TextIDClose
	TextIDItemCheck
	TextIDQuitGame
	TextIDStoreError
	TextIDShop
	TextIDMinigameWatchAds
	TextIDMinigameProgress
	TextIDVibration
)

func Text(lang language.Tag, id TextID) string {
	if lang == language.Und {
		lang = language.English
	}
	return texts[lang][id]
}

var texts = map[language.Tag]map[TextID]string{
	language.English: {
		TextIDNewGame:      "New Game",
		TextIDResetGame:    "Reset Game",
		TextIDResumeGame:   "Resume Game",
		TextIDYes:          "Yes",
		TextIDNo:           "No",
		TextIDOK:           "OK",
		TextIDSettings:     "Settings",
		TextIDLanguage:     "Language",
		TextIDCredits:      "Credits",
		TextIDCreditsEntry: "Entry",
		TextIDShop:         "Shop",
		TextIDRemoveAds:    "Remove Ads",
		TextIDRemoveAdsDesc: `Would you like to remove ads
from the game for %s?`,
		TextIDReviewThisApp:    "Review this App",
		TextIDRestorePurchases: "Restore Purchases",
		TextIDMoreGames:        "More Games",
		TextIDPrivacyPolicy:    "Privacy Policy",
		TextIDAdvancedSettings: "Advanced Settings",
		TextIDBack:             "Back",
		TextIDClose:            "Close",
		TextIDItemCheck:        "Check",
		TextIDMenu:             "Menu",
		TextIDBGMVolume:        "BGM",
		TextIDSEVolume:         "SE",
		TextIDNewGameWarning: `You have on-going game data.
Do you want to reset your
progress and start a new game?`,
		TextIDBackToTitle: `Are you sure you want to
go back to the title screen?`,
		TextIDQuitGame: `Are you sure you want to
go quit the game?`,
		TextIDStoreError: `Failed to connect to the store.
Please make sure to sign in
and connect to the network.`,
		TextIDMinigameWatchAds: "Watch ads to boost",
		TextIDMinigameProgress: "Progress %d / %d",
		TextIDVibration:        "Vibration",
	},
	language.German: {
		TextIDNewGame:      "Neues Spiel",
		TextIDResetGame:    "Spiel zurücksetzen",
		TextIDResumeGame:   "Spiel forsetzen",
		TextIDYes:          "Ja",
		TextIDNo:           "Nein",
		TextIDOK:           "OK",
		TextIDSettings:     "Einstellungen",
		TextIDLanguage:     "Sprache",
		TextIDCredits:      "Danksagungen",
		TextIDCreditsEntry: "Eintrag",
		TextIDShop:         "Geschäft",
		TextIDRemoveAds:    "Anzeigen entfernen",
		TextIDRemoveAdsDesc: `Willst du Anzeigen für %s
enfernen?`,
		TextIDReviewThisApp:    "Rezension schreiben",
		TextIDRestorePurchases: "Einkäufe wiederherstellen",
		TextIDMoreGames:        "Mehr Spiele",
		TextIDPrivacyPolicy:    "Privacy Policy",
		TextIDBack:             "Zurück",
		TextIDClose:            "Zurück",
		TextIDItemCheck:        "Info",
		TextIDMenu:             "Menü",

		TextIDNewGameWarning: `Willst du wirklich deinen
Spielfortschritt löschen und 
nochmal von Vorne anfangen?`,
		TextIDBackToTitle: `Willst du wirklich 
zurück zum Menü?`,
		TextIDQuitGame: `Willst du wirklich 
das Spiel verlassen?`,
		TextIDStoreError: `Verbindung mit dem Store nicht möglich.
Stelle sicher, dass du angemeldet
und mit dem Internet verbindet bist.`,
		TextIDMinigameWatchAds: "Beschleunigen mit Ads",
		TextIDMinigameProgress: "Fortschritt %d / %d",
	},
	language.Spanish: {
		TextIDNewGame:      "Nuevo Juego",
		TextIDResetGame:    "Reiniciar el juego", // TODO Review
		TextIDResumeGame:   "Reanudar Juego",
		TextIDYes:          "Sí",
		TextIDNo:           "No",
		TextIDOK:           "OK",
		TextIDSettings:     "Configuraciones",
		TextIDLanguage:     "Idioma",
		TextIDCredits:      "Créditos",
		TextIDCreditsEntry: "Entrada",
		TextIDShop:         "Tienda",
		TextIDRemoveAds:    "Remover anuncios",
		TextIDRemoveAdsDesc: `¿Te gustaría pagar %s
para quitar los anuncios del juego?`,
		TextIDReviewThisApp:    "Puntúa esta App",
		TextIDRestorePurchases: "Restaurar Compra",
		TextIDMoreGames:        "Más Juegos",
		TextIDPrivacyPolicy:    "Política de privacidad",
		TextIDBack:             "Cerrar",
		TextIDClose:            "Cerrar",
		TextIDItemCheck:        "Revisar",
		TextIDMenu:             "Menú",

		TextIDNewGameWarning: `Tienes datos del juego en curso.
¿Quieres eliminar el progreso 
e iniciar un nuevo juego?`,
		TextIDBackToTitle: "¿Quieres volver al título?",
		TextIDQuitGame:    "¿Quieres salir del juego?",
		TextIDStoreError: `Fallo al conectarse con la tienda. 
Por favor asegúrate de iniciar 
sesión y conectarse a internet`,
		TextIDMinigameWatchAds: "¡Acelera viendo anuncios!",
		TextIDMinigameProgress: "Progreso %d / %d",
	},
	language.Portuguese: {
		TextIDNewGame:      "Novo Jogo",
		TextIDResetGame:    "Redefinir Jogo", // TODO Review
		TextIDResumeGame:   "Resume Game",
		TextIDYes:          "Sim",
		TextIDNo:           "não",
		TextIDOK:           "OK",
		TextIDSettings:     "Configurações",
		TextIDLanguage:     "Language",
		TextIDCredits:      "Créditos",
		TextIDCreditsEntry: "Input",
		TextIDShop:         "Loja",
		TextIDRemoveAds:    "Remover anúncios",
		TextIDRemoveAdsDesc: `Você gostaria de pagar% s
remover anúncios do jogo?`,
		TextIDReviewThisApp:    "Avalie este aplicativo",
		TextIDRestorePurchases: "Restaurar Compra",
		TextIDMoreGames:        "Mais jogos",
		TextIDPrivacyPolicy:    "Política de Privacidade",
		TextIDBack:             "Fechar",
		TextIDClose:            "Fechar",
		TextIDItemCheck:        "Revisão",
		TextIDMenu:             "Menu",
		TextIDNewGameWarning: `Você tem dados do jogo em 
andamento. 
Você deseja excluir o progresso 
e começar um novo jogo?`,
		TextIDBackToTitle: "Você quer retornar ao título?",
		TextIDQuitGame:    "Você quer sair do jogo?",
		TextIDStoreError: `Falha ao conectar-se ao 
armazenamento. 
Por favor, certifique-se de fazer 
o login e se conectar à internet`,
		TextIDMinigameWatchAds: "Acelerar assistindo anúncios!",
		TextIDMinigameProgress: "Progresso %d / %d",
	},
	language.Japanese: {
		TextIDNewGame:      "はじめから",
		TextIDResetGame:    "ゲームのリセット",
		TextIDResumeGame:   "つづきから",
		TextIDYes:          "はい",
		TextIDNo:           "いいえ",
		TextIDOK:           "OK",
		TextIDSettings:     "設定",
		TextIDLanguage:     "言語",
		TextIDCredits:      "クレジット",
		TextIDCreditsEntry: "登録",
		TextIDShop:         "ショップ",
		TextIDRemoveAds:    "広告を消す",
		TextIDRemoveAdsDesc: `%sを支払って、
広告を消去しますか？`,
		TextIDReviewThisApp:    "このアプリをレビューする",
		TextIDRestorePurchases: "購入情報のリストア",
		TextIDMoreGames:        "ほかのゲーム",
		TextIDPrivacyPolicy:    "プライバシーポリシー",
		TextIDAdvancedSettings: "ゲームオプション",
		TextIDBack:             "もどる",
		TextIDClose:            "閉じる",
		TextIDItemCheck:        "チェック",
		TextIDMenu:             "タイトル",
		TextIDBGMVolume:        "BGM音量",
		TextIDSEVolume:         "SE音量",
		TextIDNewGameWarning: `進行中のゲームデータがあります。
進行中のゲームデータを消して、
新しいゲームを開始しますか?`,
		TextIDBackToTitle: "タイトル画面にもどりますか？",
		TextIDQuitGame:    "ゲームを終了しますか？",
		TextIDStoreError: `ストアへの接続に失敗しました。
ネットワークに接続しているか
確認してください`,
		TextIDMinigameWatchAds: "広告を見てブースト！",
		TextIDMinigameProgress: "進捗 %d / %d",
	},
	language.SimplifiedChinese: {
		TextIDNewGame:      "新游戏",
		TextIDResetGame:    "重置游戏", // TODO Review
		TextIDResumeGame:   "继续游戏",
		TextIDYes:          "确定",
		TextIDNo:           "取消",
		TextIDOK:           "OK",
		TextIDSettings:     "设定",
		TextIDLanguage:     "语言",
		TextIDCredits:      "制作人员",
		TextIDCreditsEntry: "注册",
		TextIDShop:         "商店",
		TextIDRemoveAds:    "移除广告",
		TextIDRemoveAdsDesc: `你希望支付%s
来移除游戏里的广告吗?`,
		TextIDReviewThisApp:    "点评我们的游戏",
		TextIDRestorePurchases: "恢复购买",
		TextIDMoreGames:        "更多游戏",
		TextIDPrivacyPolicy:    "隐私政策",
		TextIDBack:             "返回",
		TextIDClose:            "关闭",
		TextIDItemCheck:        "查看",
		TextIDMenu:             "主选单",

		TextIDNewGameWarning: `系统已经存在一个中断存档。
开始新游戏会导致中断存档被清除。
你确定要重新开始新游戏吗?`,
		TextIDBackToTitle: "返回主选单?",
		TextIDQuitGame:    "退出游戏?",
		TextIDStoreError: `无法连接商店。
请确定你已经登录并已连上网络`,
		TextIDMinigameWatchAds: "通过观看广告加速！",
		TextIDMinigameProgress: "进展 %d / %d",
	},
	language.TraditionalChinese: {
		TextIDNewGame:      "新遊戲",
		TextIDResetGame:    "重置遊戲", // TODO Review
		TextIDResumeGame:   "繼續遊戲",
		TextIDYes:          "確定",
		TextIDNo:           "取消",
		TextIDOK:           "OK",
		TextIDSettings:     "設定",
		TextIDLanguage:     "語言",
		TextIDCredits:      "製作人員",
		TextIDCreditsEntry: "註冊",
		TextIDShop:         "商店",
		TextIDRemoveAds:    "移除廣告",
		TextIDRemoveAdsDesc: `你希望支付%s
來移除遊戲裡的廣告嗎?`,
		TextIDReviewThisApp:    "點評我們的遊戲",
		TextIDRestorePurchases: "恢復購買",
		TextIDMoreGames:        "更多遊戲",
		TextIDPrivacyPolicy:    "隱私政策",
		TextIDBack:             "返回",
		TextIDClose:            "關閉",
		TextIDItemCheck:        "查看",
		TextIDMenu:             "主選單",

		TextIDNewGameWarning: `系統已經存在一個中斷存檔。
開始新遊戲會導致中斷存檔被清除。
你確定要重新開始新遊戲嗎？`,
		TextIDBackToTitle: "返回主選單?",
		TextIDQuitGame:    "退出遊戲?",
		TextIDStoreError: `無法連接商店。
請確定你已經登錄並已連上網絡`,
		TextIDMinigameWatchAds: "通過觀看廣告加速！",
		TextIDMinigameProgress: "進展 %d / %d",
	},
	language.Korean: {
		TextIDNewGame:      "처음부터",
		TextIDResetGame:    "게임리셋",
		TextIDResumeGame:   "이어하기",
		TextIDYes:          "네",
		TextIDNo:           "아니오",
		TextIDOK:           "OK",
		TextIDSettings:     "설정",
		TextIDLanguage:     "언어",
		TextIDCredits:      "크레딧",
		TextIDCreditsEntry: "등록",
		TextIDShop:         "상점",
		TextIDRemoveAds:    "광고 제거",
		TextIDRemoveAdsDesc: `%s를 지불하고, 
광고를 제거하시겠습니까?`,
		TextIDReviewThisApp:    "리뷰하기",
		TextIDRestorePurchases: "구입정보 복구",
		TextIDMoreGames:        "다른 게임",
		TextIDPrivacyPolicy:    "개인 정보 정책",
		TextIDBack:             "뒤로",
		TextIDClose:            "닫기",
		TextIDItemCheck:        "체크",
		TextIDMenu:             "타이틀",

		TextIDNewGameWarning: `진행중인 게임 데이터가 있습니다.
진행중인 게임 데이터를 지우고,
새로 게임을 시작하시겠습니까?`,
		TextIDBackToTitle: "타이틀 화면으로 돌아가시겠습니까?",
		TextIDQuitGame:    "게임을 종료하시겠습니까?",
		TextIDStoreError: `스토어에 접속 실패했습니다.
네트워크 접속이 되어있는지
확인해주세요`,
		TextIDMinigameWatchAds: "광고를 보고 부스트!",
		TextIDMinigameProgress: "진행 %d / %d",
	},
}
