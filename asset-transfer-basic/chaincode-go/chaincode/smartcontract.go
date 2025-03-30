package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

type Asset struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Owner    string `json:"owner"`
	Location string `json:"location"`
}

type TransferRequest struct {
	AssetID  string `json:"assetId"`
	NewOwner string `json:"newOwner"`
	Approved bool   `json:"approved"`
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{ID: "item1", Name: "Producto A", Owner: "FabricanteMSP", Location: "Fábrica"},
		{ID: "item2", Name: "Producto B", Owner: "FabricanteMSP", Location: "Fábrica"},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("fallo al agregar al world state: %v", err)
		}
	}

	return nil
}

func getClientOrg(ctx contractapi.TransactionContextInterface) (string, error) {
	clientID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("error getting MSPID: %v", err)
	}
	return clientID, nil
}

func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, name string, location string) error {
	clientOrg, err := getClientOrg(ctx)
	if err != nil {
		return err
	}
	if clientOrg != "Org1MSP" {
		return fmt.Errorf("only owners from Org1MSP can create assets")
	}

	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset with ID %s already exists", id)
	}

	asset := Asset{
		ID:       id,
		Name:     name,
		Owner:    clientOrg,
		Location: location,
	}

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, name string, location string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s doesn't exist", id)
	}

	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	clientOrg, err := getClientOrg(ctx)
	if err != nil {
		return err
	}

	if asset.Owner != clientOrg {
		return fmt.Errorf("only the current owner can update the asset")
	}

	asset.Name = name
	asset.Location = location

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, id string, newOwner string) error {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	clientOrg, err := getClientOrg(ctx)
	if err != nil {
		return err
	}

	// Verificar si la transferencia es válida
	if !isValidTransfer(asset.Owner, clientOrg, newOwner) {
		return fmt.Errorf("transfer not allowed from %s to %s by %s", asset.Owner, newOwner, clientOrg)
	}

	// Realizar la transferencia
	asset.Owner = newOwner

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

func isValidTransfer(currentOwner, clientOrg, newOwner string) bool {
	// Solo el Fabricante puede transferir al Transportista
	if currentOwner == "Org1MSP" && clientOrg == "Org1MSP" && newOwner == "Org2MSP" {
		return true
	}
	// Solo el Transportista puede transferir al Distribuidor
	if currentOwner == "Org2MSP" && clientOrg == "Org2MSP" && newOwner == "Org3MSP" {
		return true
	}
	return false
}
