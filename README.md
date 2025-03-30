# Práctica de despliegue de red Hyperledger Fabric

Este repositorio es un clon del oficial [fabric-samples](https://github.com/hyperledger/fabric-samples).

Ha sido modificado para que se desplieguen 3 organizaciones en lugar de dos.

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
# Volver al directorio de la test-network y lanzar el despliegue de la red
cd ..
./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go -ccl go

export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=$PWD/../config/
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org3MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp
export CORE_PEER_ADDRESS=localhost:11051

peer lifecycle chaincode package basic.tar.gz --path ../asset-transfer-basic/chaincode-go/ --lang golang --label basic_1.0
peer lifecycle chaincode install basic.tar.gz
peer lifecycle chaincode queryinstalled
export CC_PACKAGE_ID= Aqui pegamos en CC_PACKAGE_ID el resultado del comando queryinstalled

peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" --channelID mychannel --name basic --version 1.0 --package-id $CC_PACKAGE_ID --sequence 1
peer lifecycle chaincode querycommitted --channelID mychannel --name basic --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem"

peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C mychannel -n basic --peerAddresses localhost:9051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt" --peerAddresses localhost:11051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt" -c '{"function":"InitLedger","Args":[]}'

```

### Una vez se ha acabado de experimentar con la red, tirarla para liberar recursos
```sh
./network.sh down 
```
