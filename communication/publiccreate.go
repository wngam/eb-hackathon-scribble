package communication

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/scribble-rs/scribble.rs/game"
)

//This file contains the API methods for the public API

func enterLobby(w http.ResponseWriter, r *http.Request) {
	lobby, err := getLobby(r)
	if err != nil {
		if err == noLobbyIdSuppliedError {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if err == lobbyNotExistentError {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	player := getPlayer(lobby, r)

	if player == nil {
		if len(lobby.Players) >= lobby.MaxPlayers {
			http.Error(w, "lobby already full", http.StatusUnauthorized)
			return
		}

		var clientsWithSameIP int
		requestAddress := getIPAddressFromRequest(r)
		for _, otherPlayer := range lobby.Players {
			if otherPlayer.GetLastKnownAddress() == requestAddress {
				clientsWithSameIP++
				if clientsWithSameIP >= lobby.ClientsPerIPLimit {
					http.Error(w, "maximum amount of newPlayer per IP reached", http.StatusUnauthorized)
					return
				}
			}
		}

		newPlayer := lobby.JoinPlayer(getPlayername(r))
		newPlayer.SetLastKnownAddress(getIPAddressFromRequest(r))

		// Use the players generated usersession and pass it as a cookie.
		http.SetCookie(w, &http.Cookie{
			Name:     "usersession",
			Value:    newPlayer.GetUserSession(),
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
		})
	} else {
		player.SetLastKnownAddress(getIPAddressFromRequest(r))
	}

	lobbyData := &LobbyData{
		LobbyID:                lobby.ID,
		DrawingBoardBaseWidth:  DrawingBoardBaseWidth,
		DrawingBoardBaseHeight: DrawingBoardBaseHeight,
	}

	encodingError := json.NewEncoder(w).Encode(lobbyData)
	if encodingError != nil {
		http.Error(w, encodingError.Error(), http.StatusInternalServerError)
	}
}

func createLobby(w http.ResponseWriter, r *http.Request) {
	formParseError := r.ParseForm()
	if formParseError != nil {
		http.Error(w, formParseError.Error(), http.StatusBadRequest)
		return
	}

	lobbyId, lobbyIdInvalid := parseLobbyId(r.Form.Get("lobby_id"))
	language, languageInvalid := parseLanguage(r.Form.Get("language"))
	drawingTime, drawingTimeInvalid := parseDrawingTime(r.Form.Get("drawing_time"))
	rounds, roundsInvalid := parseRounds(r.Form.Get("rounds"))
	maxPlayers, maxPlayersInvalid := parseMaxPlayers(r.Form.Get("max_players"))
	customWords, customWordsInvalid := parseCustomWords(r.Form.Get("custom_words"))
	customWordChance, customWordChanceInvalid := parseCustomWordsChance(r.Form.Get("custom_words_chance"))
	clientsPerIPLimit, clientsPerIPLimitInvalid := parseClientsPerIPLimit(r.Form.Get("clients_per_ip_limit"))
	enableVotekick := r.Form.Get("enable_votekick") == "true"

	//Prevent resetting the form, since that would be annoying as hell.
	pageData := CreatePageData{
		SettingBounds:     game.LobbySettingBounds,
		Languages:         game.SupportedLanguages,
		LobbyId:           r.Form.Get("lobby_id"),
		DrawingTime:       r.Form.Get("drawing_time"),
		Rounds:            r.Form.Get("rounds"),
		MaxPlayers:        r.Form.Get("max_players"),
		CustomWords:       r.Form.Get("custom_words"),
		CustomWordsChance: r.Form.Get("custom_words_chance"),
		ClientsPerIPLimit: r.Form.Get("clients_per_ip_limit"),
		EnableVotekick:    r.Form.Get("enable_votekick"),
		Language:          r.Form.Get("language"),
	}

	if lobbyIdInvalid != nil {
		pageData.Errors = append(pageData.Errors, lobbyIdInvalid.Error())
	}
	if languageInvalid != nil {
		pageData.Errors = append(pageData.Errors, languageInvalid.Error())
	}
	if drawingTimeInvalid != nil {
		pageData.Errors = append(pageData.Errors, drawingTimeInvalid.Error())
	}
	if roundsInvalid != nil {
		pageData.Errors = append(pageData.Errors, roundsInvalid.Error())
	}
	if maxPlayersInvalid != nil {
		pageData.Errors = append(pageData.Errors, maxPlayersInvalid.Error())
	}
	if customWordsInvalid != nil {
		pageData.Errors = append(pageData.Errors, customWordsInvalid.Error())
	}
	if customWordChanceInvalid != nil {
		pageData.Errors = append(pageData.Errors, customWordChanceInvalid.Error())
	}
	if clientsPerIPLimitInvalid != nil {
		pageData.Errors = append(pageData.Errors, clientsPerIPLimitInvalid.Error())
	}

	if len(pageData.Errors) != 0 {
		http.Error(w, strings.Join(pageData.Errors, ";"), http.StatusBadRequest)
		return
	}

	var playerName = getPlayername(r)
	player, lobby, createError := game.CreateLobby(lobbyId, playerName, language, drawingTime, rounds, maxPlayers, customWordChance, clientsPerIPLimit, customWords, enableVotekick)
	if createError != nil {
		http.Error(w, createError.Error(), http.StatusBadRequest)
		return
	}

	player.SetLastKnownAddress(getIPAddressFromRequest(r))

	// Use the players generated usersession and pass it as a cookie.
	http.SetCookie(w, &http.Cookie{
		Name:     "usersession",
		Value:    player.GetUserSession(),
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	lobbyData := &LobbyData{
		LobbyID:                lobby.ID,
		DrawingBoardBaseWidth:  DrawingBoardBaseWidth,
		DrawingBoardBaseHeight: DrawingBoardBaseHeight,
	}

	encodingError := json.NewEncoder(w).Encode(lobbyData)
	if encodingError != nil {
		//If the encoding / transmitting fails, the creator will never know the
		//ID, therefore we can directly kill the lobby.
		game.RemoveLobby(lobby.ID)
		http.Error(w, encodingError.Error(), http.StatusInternalServerError)
	}
}
