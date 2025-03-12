package models

type UserStats struct {
	Rating int `json:"rating" bson:"rating"`
	Games  int `json:"games" bson:"games"`
	Wins   int `json:"wins" bson:"wins"`
	Losses int `json:"losses" bson:"losses"`
}

type User struct {
	ID       string    `json:"id"`
	Username string    `json:"username"`
	Stats    UserStats `json:"stats"`
}

type GameSettings struct {
	GameMode  string `json:"gamemode"`
	Type      string `json:"game_type"`
	BestOf    int    `json:"bestOf"`
	PlayerOne string `json:"playerOne"`
	PlayerTwo string `json:"playerTwo"`
}

type Score struct {
	PlayerOne int `json:"playerOne"`
	PlayerTwo int `json:"playerTwo"`
	Draw      int `json:"draw"`
}
type GameResult struct {
	NumRounds int    `json:"numRounds"`
	Winner    string `json:"winner"`
	Score     Score  `json:"score"`
}

type GameState struct {
	ID              int    `json:"id"`
	PlayerOneChoice string `json:"playerOneChoice"`
	PlayerTwoChoice string `json:"playerTwoChoice"`
	Winner          string `json:"winner"`
	Score           Score  `json:"score,omitempty"`
}

type GameData struct {
	ID         string       `json:"id"`
	StartTime  string       `json:"startTime,omitempty"`
	EndTime    string       `json:"endTime,omitempty"`
	Settings   GameSettings `json:"settings"`
	Result     GameResult   `json:"result"`
	GameStates []GameState  `json:"gameStates"`
}
