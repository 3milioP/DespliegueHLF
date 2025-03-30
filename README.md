# Práctica de despliegue de red Hyperledger Fabric

Este repositorio es un clon del oficial [fabric-samples](https://github.com/hyperledger/fabric-samples).

Ha sido modificado para que se desplieguen 3 organizaciones en lugar de dos y dos de éstas tendrán 2 peers en lugar de 1.

Organización 1 (Fabricante, 2 Peers)
Organización 2 (Transportista, 2 Peers)
Organización 3 (Distribuidor, 1 Peer)

## Pasos realizados para añadir una tercera organización y un peer más para cada una de las organizaciones iniciales

### 1. Implementar un nuevo peer para cada una de las organizaciones iniciales (Org1 y Org2)

```sh
# Tener los pre-requisitos listos
# Más información en la documentación oficial:
# https://hyperledger-fabric.readthedocs.io/en/release-2.5/prereqs.html

# Implementar un nuevo peer en Org1 y Org2 mediante la modificación de los siguientes archivos
test-network/compose/compose-couch.yaml
test-network/compose/compose-test-net.yaml
test-network/compose/docker/docker-compose-test-net.yaml
test-network/organizations/ccp-generate.sh
test-network/organizations/ccp-template.json
test-network/organizations/ccp-template.yaml
test-network/organizations/cryptogen/crypto-config-org1.yaml
test-network/organizations/cryptogen/crypto-config-org2.yaml

# Levantar la red y crear un canal
./network.sh up createChannel -s couchdb
```

### 2. Conectar los nuevos peers a la red

```sh
#Organización 1
export PATH=${PWD}/../bin:$PATH

export FABRIC_CFG_PATH=${PWD}/../config/
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7055

peer channel join -b channel-artifacts/mychannel.block

#Organización 2
export PATH=${PWD}/../bin:$PATH

export FABRIC_CFG_PATH=${PWD}/../config/
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer1.org2.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_ADDRESS=localhost:9055

peer channel join -b channel-artifacts/mychannel.block
```

### 3. Añadir una tercera organización (Org3)

```sh
# Entrar al directorio addOrg3
cd addOrg3

# Lanzar script para añadir el peer de Org3
./addOrg3.sh up -s couchdb
```
### 4. Despliegue de Chaincode

```sh
# Volver al directorio de la test-network
cd ..
# Conectar el peer de la nueva organización
export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=$PWD/../config/
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org3MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp
export CORE_PEER_ADDRESS=localhost:11051

# Desplegar el chaincode
peer lifecycle chaincode package basic.tar.gz --path ../asset-transfer-basic/chaincode-go/ --lang golang --label basic_1.0
peer lifecycle chaincode install basic.tar.gz
peer lifecycle chaincode queryinstalled
export CC_PACKAGE_ID= Aqui pegamos en CC_PACKAGE_ID el resultado del comando queryinstalled
./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go -ccl go

# Activar la validación para la Org3
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" --channelID mychannel --name basic --version 1.0 --package-id $CC_PACKAGE_ID --sequence 1
peer lifecycle chaincode querycommitted --channelID mychannel --name basic --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem"

# Comprobar que todo ha ido correcto
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n basic --peerAddresses localhost:9051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt" --peerAddresses localhost:11051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt" -c '{"function":"InitLedger","Args":[]}'
# Visualizar la respuesta en JSON
peer chaincode query -C mychannel -n basic -c '{"Args":["GetAllAssets"]}' | jq 
```

### 5. Desplegar servicios REST
```sh
cd ..
cd 12-despliegue-fabric-2-tx-rest-api/
go run cmd/main.go
```
Una vez desplegado el servicio REST se podrá interactuar con el contrato mediante la colección de Postman ubicada en "fabric-samples/12-despliegue-fabric-2-tx-rest-api/EmilioPractica - Transactions API REST Local.postman_collection.json" (No me dio tiempo a hacer un front en condiciones). En la descripción del contrato se describe la lógica de negocio.

# Descripción del Chaincode

Este es un contrato inteligente de Hyperledger Fabric que gestiona la transferencia de activos entre tres organizaciones en una cadena de suministro. Está basado en "asset-transfer-basic" de los fabric-samples de la Tesnet pero se han modificado sus funciones.

## Descripción

Este contrato permite la creación de activos y su transferencia entre las siguientes organizaciones:

1. **Fabricante**: El propietario inicial de los activos.
2. **Transportista**: Puede recibir activos del Fabricante.
3. **Distribuidor**: Puede recibir activos del Transportista.

### Funciones

El contrato inteligente tiene las siguientes funciones principales:

- **Crear activos**: Solo el Fabricante puede crear nuevos activos.
- **Actualizar activos**: Los activos pueden ser actualizados por el propietario actual del activo.
- **Transferir activos**: Los activos pueden ser transferidos entre las organizaciones de acuerdo con reglas predefinidas.

## Funciones del contrato

### `InitLedger`

- Inicializa un conjunto básico de activos en el libro mayor (ledger).
- Asocia dos activos iniciales con el Fabricante como propietario.

### `CreateAsset`

- Crea un nuevo activo con un ID, nombre y ubicación proporcionados.
- Solo el Fabricante puede crear nuevos activos.

### `TransferAsset`

- Permite la transferencia de un activo de una organización a otra.
- La transferencia es válida solo si:
  - El **Fabricante** puede transferir al **Transportista**.
  - El **Transportista** puede transferir al **Distribuidor**.
- Si la transferencia no es válida, se lanza un error.

### `isValidTransfer`

- Verifica si la transferencia de un activo entre organizaciones es válida.
- Solo se permite la transferencia entre:
  - **Fabricante -> Transportista**
  - **Transportista -> Distribuidor**

### `ReadAsset`

- Lee los detalles de un activo dado su ID.
- Devuelve el activo almacenado en el libro mayor.

### `getClientOrg`

- Obtiene el MSP (Membership Service Provider) de la organización que está ejecutando la transacción.
- Usado para determinar qué organización está intentando realizar una acción.


# Una vez se ha acabado de experimentar con la red, tirarla para liberar recursos
```sh
./network.sh down 
```
