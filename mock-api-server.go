package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Mock API server to provide the /api/v1/ endpoints expected by the MCP client

type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}

type Player struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Firstname   string `json:"firstname"`
	Club        string `json:"club"`
	ClubID      string `json:"club_id"`
	BirthYear   int    `json:"birth_year"`
	Gender      string `json:"gender"`
	Nation      string `json:"nation"`
	FideID      int    `json:"fide_id"`
	CurrentDWZ  int    `json:"current_dwz"`
	DWZIndex    int    `json:"dwz_index"`
	Status      string `json:"status"`
}

type Club struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	City         string `json:"city"`
	State        string `json:"state"`
	Region       string `json:"region"`
	MemberCount  int    `json:"member_count"`
	ActiveCount  int    `json:"active_count"`
	Founded      string `json:"founded"`
	Status       string `json:"status"`
}

type Tournament struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Location     string `json:"location"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	Status       string `json:"status"`
	Participants int    `json:"participants"`
	Organization string `json:"organization"`
}

type SearchResponse struct {
	Success bool        `json:"success"`
	Data    SearchData  `json:"data"`
}

type SearchData struct {
	Data []interface{} `json:"data"`
	Meta MetaData      `json:"meta"`
}

type MetaData struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
}

// Mock data
var mockPlayers = []Player{
	{
		ID: "C0327-297", Name: "Tran", Firstname: "Minh Cuong", Club: "SC Altbach 1926 e.V.", ClubID: "C0327",
		BirthYear: 1990, Gender: "male", Nation: "GER", FideID: 24663832, CurrentDWZ: 1643, DWZIndex: 60, Status: "active",
	},
	{
		ID: "C0505-1043", Name: "Aab", Firstname: "Manfred", Club: "SC Böblingen 1975 e.V.", ClubID: "C0505",
		BirthYear: 1963, Gender: "male", Nation: "GER", FideID: 24663833, CurrentDWZ: 1643, DWZIndex: 60, Status: "active",
	},
}

var mockClubs = []Club{
	{
		ID: "C0327", Name: "SC Altbach 1926 e.V.", City: "Altbach", State: "Baden-Württemberg", Region: "Württemberg",
		MemberCount: 45, ActiveCount: 38, Founded: "1926", Status: "active",
	},
	{
		ID: "C0505", Name: "SC Böblingen 1975 e.V.", City: "Böblingen", State: "Baden-Württemberg", Region: "Württemberg",
		MemberCount: 62, ActiveCount: 51, Founded: "1975", Status: "active",
	},
}

var mockTournaments = []Tournament{
	{
		ID: "C350-C01-SMU", Name: "Ulm Open 2024", Location: "Ulm", StartDate: "2024-03-15", EndDate: "2024-03-17",
		Status: "completed", Participants: 84, Organization: "SC Ulm 1946 e.V.",
	},
	{
		ID: "B735-705-QCB", Name: "Bezirksliga Württemberg 2024", Location: "Stuttgart", StartDate: "2024-02-10", EndDate: "2024-02-11",
		Status: "completed", Participants: 56, Organization: "Württemberg Chess Federation",
	},
	{
		ID: "T96887", Name: "Kreismeisterschaft 2024", Location: "Esslingen", StartDate: "2024-01-20", EndDate: "2024-01-21",
		Status: "completed", Participants: 32, Organization: "SK Esslingen 1925",
	},
}

func main() {
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/api/v1/health", handleAPIHealth)
	http.HandleFunc("/api/v1/players", handlePlayers)
	http.HandleFunc("/api/v1/players/", handlePlayerDetails)
	http.HandleFunc("/api/v1/clubs", handleClubs)
	http.HandleFunc("/api/v1/clubs/", handleClubDetails)
	http.HandleFunc("/api/v1/tournaments", handleTournaments)
	http.HandleFunc("/api/v1/tournaments/", handleTournamentDetails)
	http.HandleFunc("/api/v1/tournaments/search", handleTournamentSearch)
	http.HandleFunc("/api/v1/tournaments/recent", handleRecentTournaments)
	http.HandleFunc("/api/v1/admin/cache", handleCacheStats)
	http.HandleFunc("/api/v1/addresses/regions", handleRegions)
	http.HandleFunc("/api/v1/addresses/", handleRegionAddresses)

	fmt.Println("Mock DWZ API Server starting on :8080")
	fmt.Println("Endpoints available:")
	fmt.Println("  GET /health")
	fmt.Println("  GET /api/v1/health") 
	fmt.Println("  GET /api/v1/players")
	fmt.Println("  GET /api/v1/players/{id}")
	fmt.Println("  GET /api/v1/players/{id}/history")
	fmt.Println("  GET /api/v1/clubs")
	fmt.Println("  GET /api/v1/clubs/{id}")
	fmt.Println("  GET /api/v1/clubs/{id}/players")
	fmt.Println("  GET /api/v1/clubs/{id}/statistics")
	fmt.Println("  GET /api/v1/tournaments")
	fmt.Println("  GET /api/v1/tournaments/search")
	fmt.Println("  GET /api/v1/tournaments/recent")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "healthy",
		"version": "1.0.0",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleAPIHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":        "healthy",
		"response_time": 15,
		"api_version":   "1.0.0", 
		"timestamp":     time.Now().Format(time.RFC3339),
		"services": map[string]interface{}{
			"database": map[string]interface{}{
				"status":        "healthy",
				"response_time": 5,
				"last_check":    time.Now().Format(time.RFC3339),
			},
			"cache": map[string]interface{}{
				"status":        "healthy", 
				"response_time": 2,
				"last_check":    time.Now().Format(time.RFC3339),
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handlePlayers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	var filteredPlayers []Player
	for _, player := range mockPlayers {
		if query == "" || strings.Contains(strings.ToLower(player.Name+" "+player.Firstname), strings.ToLower(query)) {
			filteredPlayers = append(filteredPlayers, player)
		}
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if start > len(filteredPlayers) {
		start = len(filteredPlayers)
	}
	if end > len(filteredPlayers) {
		end = len(filteredPlayers)
	}

	paginatedPlayers := filteredPlayers[start:end]
	
	// Convert to interface{} slice
	var data []interface{}
	for _, p := range paginatedPlayers {
		data = append(data, p)
	}

	response := SearchResponse{
		Success: true,
		Data: SearchData{
			Data: data,
			Meta: MetaData{
				Total:  len(filteredPlayers),
				Limit:  limit,
				Offset: offset,
				Count:  len(paginatedPlayers),
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handlePlayerDetails(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/players/")
	
	// Check if this is a history request
	if strings.Contains(path, "/history") {
		handlePlayerHistory(w, r)
		return
	}
	
	playerID := path
	
	for _, player := range mockPlayers {
		if player.ID == playerID {
			response := SearchResponse{
				Success: true,
				Data: SearchData{
					Data: []interface{}{player},
					Meta: MetaData{Total: 1, Limit: 1, Offset: 0, Count: 1},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// Player not found
	http.Error(w, "Player not found", http.StatusNotFound)
}

func handleClubs(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	var filteredClubs []Club
	for _, club := range mockClubs {
		if query == "" || strings.Contains(strings.ToLower(club.Name), strings.ToLower(query)) {
			filteredClubs = append(filteredClubs, club)
		}
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if start > len(filteredClubs) {
		start = len(filteredClubs)
	}
	if end > len(filteredClubs) {
		end = len(filteredClubs)
	}

	paginatedClubs := filteredClubs[start:end]
	
	// Convert to interface{} slice
	var data []interface{}
	for _, c := range paginatedClubs {
		data = append(data, c)
	}

	response := SearchResponse{
		Success: true,
		Data: SearchData{
			Data: data,
			Meta: MetaData{
				Total:  len(filteredClubs),
				Limit:  limit,
				Offset: offset,
				Count:  len(paginatedClubs),
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleClubDetails(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/clubs/")
	
	// Check if this is a players request
	if strings.Contains(path, "/players") {
		handleClubPlayers(w, r)
		return
	}
	
	// Check if this is a statistics request
	if strings.Contains(path, "/statistics") {
		handleClubStatistics(w, r)
		return
	}
	
	clubID := path
	
	for _, club := range mockClubs {
		if club.ID == clubID {
			response := SearchResponse{
				Success: true,
				Data: SearchData{
					Data: []interface{}{club},
					Meta: MetaData{Total: 1, Limit: 1, Offset: 0, Count: 1},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// Club not found
	http.Error(w, "Club not found", http.StatusNotFound)
}

func handleTournaments(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	var filteredTournaments []Tournament
	for _, tournament := range mockTournaments {
		if query == "" || strings.Contains(strings.ToLower(tournament.Name+" "+tournament.Location), strings.ToLower(query)) {
			filteredTournaments = append(filteredTournaments, tournament)
		}
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if start > len(filteredTournaments) {
		start = len(filteredTournaments)
	}
	if end > len(filteredTournaments) {
		end = len(filteredTournaments)
	}

	paginatedTournaments := filteredTournaments[start:end]
	
	// Convert to interface{} slice
	var data []interface{}
	for _, t := range paginatedTournaments {
		data = append(data, t)
	}

	response := SearchResponse{
		Success: true,
		Data: SearchData{
			Data: data,
			Meta: MetaData{
				Total:  len(filteredTournaments),
				Limit:  limit,
				Offset: offset,
				Count:  len(paginatedTournaments),
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleTournamentDetails(w http.ResponseWriter, r *http.Request) {
	tournamentID := strings.TrimPrefix(r.URL.Path, "/api/v1/tournaments/")
	
	for _, tournament := range mockTournaments {
		if tournament.ID == tournamentID {
			response := SearchResponse{
				Success: true,
				Data: SearchData{
					Data: []interface{}{tournament},
					Meta: MetaData{Total: 1, Limit: 1, Offset: 0, Count: 1},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// Tournament not found
	http.Error(w, "Tournament not found", http.StatusNotFound)
}

func handleCacheStats(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"hit_ratio": 0.893,
		"operations": map[string]interface{}{
			"hits":    int64(1250),
			"misses":  int64(150),
			"sets":    int64(450),
			"deletes": int64(25),
			"flushes": int64(5),
		},
		"performance": map[string]interface{}{
			"average_get_time": int64(2500000), // 2.5 ms in nanoseconds  
			"average_set_time": int64(3200000), // 3.2 ms in nanoseconds
			"connection_time":  int64(1000000), // 1 ms in nanoseconds
		},
		"usage": map[string]interface{}{
			"used_memory":    int64(2516582), // ~2.4MB in bytes
			"max_memory":     int64(16777216), // 16MB in bytes
			"memory_percent": 15.0,
			"key_count":      int64(450),
			"expired_keys":   int64(12),
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleRegions(w http.ResponseWriter, r *http.Request) {
	regions := []map[string]interface{}{
		{
			"code":          "BW",
			"name":          "Baden-Württemberg",
			"country":       "Germany",
			"address_types": []string{"president", "secretary", "treasurer"},
		},
		{
			"code":          "BY",
			"name":          "Bayern",
			"country":       "Germany",
			"address_types": []string{"president", "secretary", "treasurer"},
		},
		{
			"code":          "BE",
			"name":          "Berlin",
			"country":       "Germany",
			"address_types": []string{"president", "secretary"},
		},
		{
			"code":          "NW",
			"name":          "Nordrhein-Westfalen",
			"country":       "Germany",
			"address_types": []string{"president", "secretary", "treasurer"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(regions)
}

func handleRegionAddresses(w http.ResponseWriter, r *http.Request) {
	region := strings.TrimPrefix(r.URL.Path, "/api/v1/addresses/")
	
	// Mock address data with correct structure
	addresses := []map[string]interface{}{
		{
			"id":          "addr-" + region + "-1",
			"region":      region,
			"type":        "president",
			"name":        "Hans Müller",
			"position":    "President",
			"email":       "president@chess-" + strings.ToLower(region) + ".de",
			"phone":       "+49 123 456789",
			"address":     "Musterstraße 123",
			"city":        region,
			"postal_code": "12345",
			"country":     "Germany",
		},
		{
			"id":          "addr-" + region + "-2",
			"region":      region,
			"type":        "secretary",
			"name":        "Maria Schmidt",
			"position":    "Secretary",
			"email":       "secretary@chess-" + strings.ToLower(region) + ".de",
			"phone":       "+49 123 456790",
			"address":     "Beispielweg 456",
			"city":        region,
			"postal_code": "12346",
			"country":     "Germany",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(addresses)
}

// handlePlayerHistory handles player rating history requests
func handlePlayerHistory(w http.ResponseWriter, r *http.Request) {
	playerID := strings.TrimPrefix(r.URL.Path, "/api/v1/players/")
	playerID = strings.TrimSuffix(playerID, "/history")
	
	// Mock rating history data
	mockHistory := []map[string]interface{}{
		{
			"period": "2024-01",
			"dwz": 1850,
			"index": 25,
			"games": 8,
			"performance": 1900,
		},
		{
			"period": "2023-12", 
			"dwz": 1830,
			"index": 23,
			"games": 6,
			"performance": 1875,
		},
		{
			"period": "2023-11",
			"dwz": 1810,
			"index": 22,
			"games": 9,
			"performance": 1850,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mockHistory)
}

// handleClubPlayers handles club players requests  
func handleClubPlayers(w http.ResponseWriter, r *http.Request) {
	clubID := strings.TrimPrefix(r.URL.Path, "/api/v1/clubs/")
	clubID = strings.TrimSuffix(clubID, "/players")
	
	// Mock club players data
	mockClubPlayers := []Player{
		{
			ID: "C0327-297", Name: "Tran", Firstname: "Minh Cuong", Club: "SC Altbach 1926 e.V.", ClubID: clubID,
			BirthYear: 1990, Gender: "M", Nation: "GER", FideID: 0, CurrentDWZ: 1850, DWZIndex: 25, Status: "active",
		},
		{
			ID: "C0327-298", Name: "Schmidt", Firstname: "Hans", Club: "SC Altbach 1926 e.V.", ClubID: clubID,
			BirthYear: 1985, Gender: "M", Nation: "GER", FideID: 0, CurrentDWZ: 1720, DWZIndex: 18, Status: "active",
		},
	}

	response := SearchResponse{
		Success: true,
		Data: SearchData{
			Data: convertToInterface(mockClubPlayers),
			Meta: MetaData{
				Total:  len(mockClubPlayers),
				Limit:  len(mockClubPlayers),
				Offset: 0,
				Count:  len(mockClubPlayers),
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleClubStatistics handles club statistics requests
func handleClubStatistics(w http.ResponseWriter, r *http.Request) {
	clubID := strings.TrimPrefix(r.URL.Path, "/api/v1/clubs/")
	clubID = strings.TrimSuffix(clubID, "/statistics")
	
	// Mock club statistics data
	mockStats := map[string]interface{}{
		"club_id": clubID,
		"member_count": 45,
		"active_count": 38,
		"average_dwz": 1650,
		"rating_distribution": map[string]interface{}{
			"under_1200": 5,
			"1200_1400": 8,
			"1400_1600": 12,
			"1600_1800": 10,
			"1800_2000": 7,
			"over_2000": 3,
		},
		"tournament_participation": map[string]interface{}{
			"total_tournaments": 24,
			"avg_per_player": 2.1,
			"top_performers": []string{"C0327-297", "C0327-298"},
		},
		"performance_trends": map[string]interface{}{
			"last_6_months": map[string]interface{}{
				"average_change": +15,
				"improving_players": 12,
				"declining_players": 8,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mockStats)
}

// handleTournamentSearch handles tournament search by date range
func handleTournamentSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	// startDate := r.URL.Query().Get("start_date") // TODO: implement date filtering
	// endDate := r.URL.Query().Get("end_date")     // TODO: implement date filtering
	
	// Filter tournaments based on query (ignoring dates for now)
	filteredTournaments := []Tournament{}
	for _, tournament := range mockTournaments {
		// Simple filtering logic
		if query == "" || strings.Contains(strings.ToLower(tournament.Name), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(tournament.Location), strings.ToLower(query)) {
			filteredTournaments = append(filteredTournaments, tournament)
		}
	}

	response := SearchResponse{
		Success: true,
		Data: SearchData{
			Data: convertToInterface(filteredTournaments),
			Meta: MetaData{
				Total:  len(filteredTournaments),
				Limit:  len(filteredTournaments),
				Offset: 0,
				Count:  len(filteredTournaments),
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleRecentTournaments handles recent tournaments request
func handleRecentTournaments(w http.ResponseWriter, r *http.Request) {
	// days := 30 // default - TODO: implement date filtering  
	limit := 25 // default
	
	// if daysStr := r.URL.Query().Get("days"); daysStr != "" {
	// 	if d, err := strconv.Atoi(daysStr); err == nil {
	// 		days = d
	// 	}
	// }
	
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	// For simplicity, return all mock tournaments (in real scenario would filter by date)
	recentTournaments := mockTournaments
	if len(recentTournaments) > limit {
		recentTournaments = recentTournaments[:limit]
	}

	// Return direct array, not wrapped in SearchResponse
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recentTournaments)
}

// Helper function to convert slice to []interface{}
func convertToInterface(data interface{}) []interface{} {
	switch v := data.(type) {
	case []Tournament:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case []Player:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case []Club:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	default:
		return nil
	}
}
