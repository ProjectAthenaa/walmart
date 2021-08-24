package module

type Store struct {
	Stores []struct {
		StoreID     string   `json:"storeId"`
		Types       []string `json:"types"`
		Distance    float64  `json:"distance"`
		BuID        string   `json:"buId"`
		ID          int      `json:"id"`
		DisplayName string   `json:"displayName"`
		StoreType   struct {
			ID          int    `json:"id"`
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
		} `json:"storeType"`
		Address struct {
			PostalCode string `json:"postalCode"`
			Address1   string `json:"address1"`
			City       string `json:"city"`
			State      string `json:"state"`
			Country    string `json:"country"`
		} `json:"address"`
	} `json:"stores"`
}
