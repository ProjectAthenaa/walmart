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

type CreditCardForm struct {
	EncryptedPan   string `json:"encryptedPan"`
	EncryptedCvv   string `json:"encryptedCvv"`
	IntegrityCheck string `json:"integrityCheck"`
	KeyID          string `json:"keyId"`
	Phase          string `json:"phase"`
	State          string `json:"state"`
	City           string `json:"city"`
	AddressType    string `json:"addressType"`
	PostalCode     string `json:"postalCode"`
	AddressLineOne string `json:"addressLineOne"`
	AddressLineTwo string `json:"addressLineTwo"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	ExpiryMonth    string `json:"expiryMonth"`
	ExpiryYear     string `json:"expiryYear"`
	Phone          string `json:"phone"`
	CardType       string `json:"cardType"`
	IsGuest        bool   `json:"isGuest"`
}

type Payment struct {
	Payments []struct {
		PaymentType    string `json:"paymentType"`
		CardType       string `json:"cardType"`
		FirstName      string `json:"firstName"`
		LastName       string `json:"lastName"`
		AddressLineOne string `json:"addressLineOne"`
		AddressLineTwo string `json:"addressLineTwo"`
		City           string `json:"city"`
		State          string `json:"state"`
		PostalCode     string `json:"postalCode"`
		ExpiryMonth    string `json:"expiryMonth"`
		ExpiryYear     string `json:"expiryYear"`
		Email          string `json:"email"`
		Phone          string `json:"phone"`
		EncryptedPan   string `json:"encryptedPan"`
		EncryptedCvv   string `json:"encryptedCvv"`
		IntegrityCheck string `json:"integrityCheck"`
		KeyID          string `json:"keyId"`
		Phase          string `json:"phase"`
		PiHash         string `json:"piHash"`
	} `json:"payments"`
	CvvInSession bool `json:"cvvInSession"`
}

type OrderConfirm struct {
	CvvInSession    bool `json:"cvvInSession"`
	VoltagePayments []map[string]string `json:"voltagePayments"`
}