package chaincode

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions
type SmartContract struct {
	contractapi.Contract
}

// CTIData represents the data structure for CTI data entries
type CTIData struct {
	ID         string `json:"ID"`
	Name       string `json:"Name"`
	Uploader   string `json:"Uploader"`
	Timestamp  int    `json:"Timestamp"`
	CID        string `json:"CID"`
	EncryptKey string `json:"encryptKey"`
	Points     int    `json:"Points"`
	Level      int    `json:"Level"`
}

// UserData represents the data structure for user entries
type UserData struct {
	ID          string `json:"ID"`
	UserLevel   int    `json:"UserLevel"`
	UploadCount int    `json:"UploadCount"`
	Points      int    `json:"Points"`
	Subscribed  int    `json:"Subscribed"`
	Balance     int    `json:"Balance"`
}

// ReviewData represents the data structure for review entries
type ReviewData struct {
	ID           string `json:"ID"`
	UserDataID   string `json:"UserDataID"`
	CTIDataID    string `json:"CTIDataID"`
	Accuracy     int    `json:"Accuracy"`
	Timeliness   int    `json:"Timeliness"`
	Completeness int    `json:"Completeness"`
	Consistency  int    `json:"Consistency"`
	ReviewText   string `json:"ReviewText"`
}

// AddCTIItem adds a new CTI item to the ledger
func (cc *SmartContract) AddCTIItem(ctx contractapi.TransactionContextInterface, name string, timestamp int, cid string, encryptKey string, points int, level int) error {
	// Get the current peer ID
	uploader, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get uploader ID: %v", err)
	}

	// Get the current ID from the ledger
	idBytes, err := ctx.GetStub().GetState("latestID")
	if err != nil {
		//return fmt.Errorf("failed to read latest ID from ledger: %v", err)
	}
	var latestID int
	if idBytes == nil {
		latestID = 1 // Start with ID = 1 if it's the first entry
	} else {
		latestID, err = strconv.Atoi(string(idBytes))
		if err != nil {
			return fmt.Errorf("failed to convert latest ID to integer: %v", err)
		}
		latestID++ // Increment the ID
	}

	// Create the CTIData instance
	ctiItem := CTIData{
		ID:         strconv.Itoa(latestID),
		Name:       name,
		Uploader:   uploader,
		Timestamp:  timestamp,
		CID:        cid,
		EncryptKey: encryptKey,
		Points:     points,
		Level:      level,
	}

	// Convert CTIData to JSON
	ctiItemJSON, err := json.Marshal(ctiItem)
	if err != nil {
		return fmt.Errorf("failed to marshal CTIData to JSON: %v", err)
	}

	// Put the CTIData on the ledger
	if err := ctx.GetStub().PutState(fmt.Sprintf("CTI_%d", latestID), ctiItemJSON); err != nil {
		return fmt.Errorf("failed to put CTI data on ledger: %v", err)
	}

	// Update the latest ID on the ledger
	if err := ctx.GetStub().PutState("latestID", []byte(strconv.Itoa(latestID))); err != nil {
		return fmt.Errorf("failed to update latest ID on ledger: %v", err)
	}

	return nil
}

func (cc *SmartContract) UpdateCTIItem(ctx contractapi.TransactionContextInterface, id string, name string, timestamp int, cid string, encryptKey string, points, level int) error {
	// Get the current peer ID
	uploader, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get uploader ID: %v", err)
	}

	// Check if the CTI item exists
	ctiItemJSON, err := ctx.GetStub().GetState(fmt.Sprintf("CTI_%s", id))
	if err != nil {
		return fmt.Errorf("failed to read CTI item from ledger: %v", err)
	}
	if ctiItemJSON == nil {
		return fmt.Errorf("CTI item with ID %s does not exist", id)
	}

	// Update the CTI item
	ctiItem := CTIData{
		ID:         id,
		Name:       name,
		Uploader:   uploader,
		Timestamp:  timestamp,
		CID:        cid,
		EncryptKey: encryptKey,
		Points:     points,
		Level:      level,
	}

	// Convert CTI data to JSON
	ctiItemJSON, err = json.Marshal(ctiItem)
	if err != nil {
		return fmt.Errorf("failed to marshal CTI item to JSON: %v", err)
	}

	// Put the updated CTI item on the ledger
	if err := ctx.GetStub().PutState(fmt.Sprintf("CTI_%s", id), ctiItemJSON); err != nil {
		return fmt.Errorf("failed to put updated CTI item on ledger: %v", err)
	}

	return nil
}

// GetCTIItem retrieves a CTI item from the ledger by its ID
func (cc *SmartContract) GetCTIItem(ctx contractapi.TransactionContextInterface, id int) (*CTIData, error) {
	ctiItemJSON, err := ctx.GetStub().GetState(fmt.Sprintf("CTI_%d", id))
	if err != nil {
		return nil, err
	}
	if ctiItemJSON == nil {
		return nil, fmt.Errorf("CTI item with ID %d does not exist", id)
	}

	var ctiItem CTIData
	err = json.Unmarshal(ctiItemJSON, &ctiItem)
	if err != nil {
		return nil, err
	}

	return &ctiItem, nil
}

// GetAllCTIItems retrieves all CTI data entries from the ledger
func (cc *SmartContract) GetAllCTIItems(ctx contractapi.TransactionContextInterface) ([]*CTIData, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("CTI_0", "CTI_999999")
	if err != nil {
		return nil, fmt.Errorf("failed to get CTI data range: %v", err)
	}
	defer resultsIterator.Close()

	var ctiItems []*CTIData
	for resultsIterator.HasNext() {
		item, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over CTI data range: %v", err)
		}

		var ctiItem CTIData
		if err := json.Unmarshal(item.Value, &ctiItem); err != nil {
			return nil, fmt.Errorf("failed to unmarshal CTI data: %v", err)
		}
		ctiItems = append(ctiItems, &ctiItem)
	}

	return ctiItems, nil
}

// AddUserData adds user statistics data to the ledger
func (cc *SmartContract) AddUserData(ctx contractapi.TransactionContextInterface, uploadCount int, points int, subscribed int, balance int) error {
	user, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get client identity: %v", err)
	}

	userData := UserData{
		ID:          user,
		UploadCount: uploadCount,
		Points:      points,
		Subscribed:  subscribed,
		Balance:     balance,
	}

	userDataJSON, err := json.Marshal(userData)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(fmt.Sprintf("UserData_%s", user), userDataJSON)
}

// GetUserData retrieves user statistics data from the ledger by user ID
func (cc *SmartContract) GetUserDataOld(ctx contractapi.TransactionContextInterface, user string) (*UserData, error) {
	userDataJSON, err := ctx.GetStub().GetState(fmt.Sprintf("UserData_%s", user))
	if err != nil {
		return nil, err
	}
	if userDataJSON == nil {
		return nil, fmt.Errorf("User data for user %s does not exist", user)
	}

	var userData UserData
	err = json.Unmarshal(userDataJSON, &userData)
	if err != nil {
		return nil, err
	}

	return &userData, nil
}

/*
// GetUserData retrieves user statistics data from the ledger by peer ID
func (cc *SmartContract) GetUserData(ctx contractapi.TransactionContextInterface) (*UserData, error) {
	// Retrieve the current peer ID
	peerID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return nil, fmt.Errorf("failed to get current peer ID: %v", err)
	}

	userDataJSON, err := ctx.GetStub().GetState(fmt.Sprintf("UserData_%s", peerID))
	if err != nil {
		return nil, err
	}
	if userDataJSON == nil {
		return nil, fmt.Errorf("User data for peer %s does not exist", peerID)
	}

	var userData UserData
	err = json.Unmarshal(userDataJSON, &userData)
	if err != nil {
		return nil, err
	}

	return &userData, nil
}
*/

// GetUserData retrieves user statistics data from the ledger by peer ID.
// If user data doesn't exist, it creates an empty user data entry with the current peer ID.
func (cc *SmartContract) GetUserData(ctx contractapi.TransactionContextInterface) (*UserData, error) {
	// Retrieve the current peer ID
	peerID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return nil, fmt.Errorf("failed to get current peer ID: %v", err)
	}

	userDataJSON, err := ctx.GetStub().GetState(fmt.Sprintf("UserData_%s", peerID))
	if err != nil {
		return nil, err
	}

	if userDataJSON == nil {
		// Create empty user data
		userData := &UserData{
			ID:          peerID,
			UploadCount: 0,
			Points:      0,
			Subscribed:  0,
			Balance:     0,
		}

		// Marshal the user data to JSON
		userDataJSON, err := json.Marshal(userData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal user data: %v", err)
		}

		// Save the empty user data on the ledger
		err = ctx.GetStub().PutState(fmt.Sprintf("UserData_%s", peerID), userDataJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to put user data on ledger: %v", err)
		}

		return userData, nil
	}

	var userData UserData
	err = json.Unmarshal(userDataJSON, &userData)
	if err != nil {
		return nil, err
	}

	return &userData, nil
}

// UpdateUserData updates the user data for the current peer with the provided fields
func (cc *SmartContract) UpdateUserData(ctx contractapi.TransactionContextInterface, uploadCount, points, subscribed, balance int) error {
	// Retrieve the current peer ID
	peerID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get current peer ID: %v", err)
	}

	// Check if user data exists
	existingUserDataJSON, err := ctx.GetStub().GetState(fmt.Sprintf("UserData_%s", peerID))
	if err != nil {
		return err
	}

	if existingUserDataJSON == nil {
		return fmt.Errorf("User data for peer %s does not exist", peerID)
	}

	// Retrieve existing user data
	var existingUserData UserData
	err = json.Unmarshal(existingUserDataJSON, &existingUserData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal existing user data: %v", err)
	}

	// Update user data fields
	existingUserData.UploadCount = uploadCount
	existingUserData.Points = points
	existingUserData.Subscribed = subscribed
	existingUserData.Balance = balance

	// Marshal the updated user data
	userDataJSON, err := json.Marshal(existingUserData)
	if err != nil {
		return fmt.Errorf("failed to marshal updated user data: %v", err)
	}

	// Put the updated user data on the ledger
	err = ctx.GetStub().PutState(fmt.Sprintf("UserData_%s", peerID), userDataJSON)
	if err != nil {
		return fmt.Errorf("failed to put updated user data on ledger: %v", err)
	}

	return nil
}

// AddReviewDataByCTIDataID adds review data for a specific CTI data ID
func (cc *SmartContract) AddReviewData(ctx contractapi.TransactionContextInterface, ctiDataID string, accuracy, timeliness, completeness, consistency int, reviewText string) error {
	// Retrieve the current peer ID
	peerID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get current peer ID: %v", err)
	}

	// Check if the CTI item exists
	ctiItemJSON, err := ctx.GetStub().GetState(fmt.Sprintf("CTI_%s", ctiDataID))
	if err != nil {
		return fmt.Errorf("failed to read CTI item from ledger: %v", err)
	}
	if ctiItemJSON == nil {
		return fmt.Errorf("CTI item with ID %s does not exist", ctiDataID)
	}

	// Generate a unique ID for the review data
	reviewID, err := generateUniqueID(ctx, "Review")
	if err != nil {
		return fmt.Errorf("failed to generate review ID: %v", err)
	}

	// Create the review data instance
	review := ReviewData{
		ID:           reviewID,
		UserDataID:   peerID,
		CTIDataID:    ctiDataID,
		Accuracy:     accuracy,
		Timeliness:   timeliness,
		Completeness: completeness,
		Consistency:  consistency,
		ReviewText:   reviewText,
	}

	// Convert review data to JSON
	reviewJSON, err := json.Marshal(review)
	if err != nil {
		return fmt.Errorf("failed to marshal review data to JSON: %v", err)
	}

	// Put the review data on the ledger
	if err := ctx.GetStub().PutState(fmt.Sprintf("Review_%s", reviewID), reviewJSON); err != nil {
		return fmt.Errorf("failed to put review data on ledger: %v", err)
	}

	return nil
}

// generateUniqueID generates a unique ID for a given prefix
func generateUniqueID(ctx contractapi.TransactionContextInterface, prefix string) (string, error) {
	// Retrieve the current ID for the given prefix
	idBytes, err := ctx.GetStub().GetState("latestID_" + prefix)
	if err != nil {
		return "", fmt.Errorf("failed to read latest ID for prefix %s: %v", prefix, err)
	}

	// Convert the ID to an integer
	var latestID int
	if idBytes == nil {
		latestID = 1 // Start with ID = 1 if it's the first entry
	} else {
		latestID, err = strconv.Atoi(string(idBytes))
		if err != nil {
			return "", fmt.Errorf("failed to convert latest ID to integer: %v", err)
		}
		latestID++ // Increment the ID
	}

	// Update the latest ID on the ledger
	if err := ctx.GetStub().PutState("latestID_"+prefix, []byte(strconv.Itoa(latestID))); err != nil {
		return "", fmt.Errorf("failed to update latest ID for prefix %s on ledger: %v", prefix, err)
	}

	// Return the generated ID
	return fmt.Sprintf("%s_%d", prefix, latestID), nil
}

// GetAllReviewData retrieves all review data entries from the ledger
func (cc *SmartContract) GetAllReviewData(ctx contractapi.TransactionContextInterface) ([]*ReviewData, error) {
	// Construct partial composite key for review data
	startKey := "Review_"
	endKey := "Review_z" // Assumes z as the upper limit for ASCII characters

	// Get iterator for all review data entries
	iterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read all review data entries: %v", err)
	}
	defer iterator.Close()

	// Iterate through the results and unmarshal review data entries
	var reviews []*ReviewData
	for iterator.HasNext() {
		item, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next item in iterator: %v", err)
		}

		// Unmarshal review data
		var review ReviewData
		err = json.Unmarshal(item.Value, &review)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal review data: %v", err)
		}

		// Append review data to result list
		reviews = append(reviews, &review)
	}

	return reviews, nil
}

// GetAllReviewData retrieves all review data entries from the ledger and filters them by the specified CTI data ID
func (cc *SmartContract) GetReviewDataByCTIDataID(ctx contractapi.TransactionContextInterface, ctiDataID string) ([]*ReviewData, error) {
	// Get all review data entries from the ledger
	allReviewData, err := cc.GetAllReviewData(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all review data entries: %v", err)
	}

	// Filter review data by the specified CTI data ID
	var filteredReviews []*ReviewData
	for _, review := range allReviewData {
		if review.CTIDataID == ctiDataID {
			filteredReviews = append(filteredReviews, review)
		}
	}

	return filteredReviews, nil
}

// GetCTIItemsFilteredBySubscriptionLevel retrieves CTI data entries from the ledger filtered by subscription level
func (cc *SmartContract) GetCTIItemsFilteredBySubscriptionLevel(ctx contractapi.TransactionContextInterface) ([]*CTIData, error) {
	// Retrieve all CTI data entries from the ledger
	allCTIItems, err := cc.GetAllCTIItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all CTI data entries: %v", err)
	}

	// Retrieve user data for the current peer
	userData, err := cc.GetUserData(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user data: %v", err)
	}

	// Filter CTI data entries based on subscription level
	var filteredCTIItems []*CTIData
	for _, ctiItem := range allCTIItems {
		if ctiItem.Level <= userData.Subscribed {
			filteredCTIItems = append(filteredCTIItems, ctiItem)
		}
	}

	return filteredCTIItems, nil
}

// DeleteCTIItemByID deletes a CTI data entry from the ledger by its ID
func (cc *SmartContract) DeleteCTIItemByID(ctx contractapi.TransactionContextInterface, id string) error {
	// Check if the CTI data entry exists
	existingItemJSON, err := ctx.GetStub().GetState(fmt.Sprintf("CTI_%s", id))
	if err != nil {
		return fmt.Errorf("failed to read CTI data entry: %v", err)
	}
	if existingItemJSON == nil {
		return fmt.Errorf("CTI data entry with ID %s does not exist", id)
	}

	// Delete the CTI data entry from the ledger
	err = ctx.GetStub().DelState(fmt.Sprintf("CTI_%s", id))
	if err != nil {
		return fmt.Errorf("failed to delete CTI data entry: %v", err)
	}

	return nil
}
