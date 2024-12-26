#清理之前的网络
docker-compose -f explorer/docker-compose.yaml down -v
docker-compose -f docker-compose-byfn.yaml down -v
rm -rf channel-artifacts
rm -rf crypto-config
#生成证书材料
./bin/cryptogen generate --config=./crypto-config.yaml
#创建通道相关资料文件夹
mkdir channel-artifacts
#生成创世区块
./bin/configtxgen -profile TwoOrgsOrdererGenesis -channelID byfn-sys-channel -outputBlock ./channel-artifacts/genesis.block
#生成通道配置
./bin/configtxgen -profile TwoOrgsChannel -outputCreateChannelTx ./channel-artifacts/channel.tx -channelID mychannel
#生成锚节点配置
./bin/configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org1MSPanchors.tx -channelID mychannel -asOrg Org1MSP
#生成锚节点配置
./bin/configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org2MSPanchors.tx -channelID mychannel -asOrg Org2MSP

docker-compose -f docker-compose-byfn.yaml up -d 

peer0org1="CORE_PEER_LOCALMSPID="Org1MSP" CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp CORE_PEER_ADDRESS=peer0.org1.example.com:7051"
peer1org1="CORE_PEER_LOCALMSPID="Org1MSP" CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp CORE_PEER_ADDRESS=peer1.org1.example.com:8051"
peer0org2="CORE_PEER_LOCALMSPID="Org2MSP" CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp CORE_PEER_ADDRESS=peer0.org2.example.com:9051"
peer1org2="CORE_PEER_LOCALMSPID="Org2MSP" CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp CORE_PEER_ADDRESS=peer1.org2.example.com:10051"

docker exec cli bash -c "$peer0org1 peer channel create -o orderer.example.com:7050 -c mychannel -f ./channel-artifacts/channel.tx"

#将所有节点加入通道
docker exec cli bash -c "$peer0org1 peer channel join -b mychannel.block"
docker exec cli bash -c "$peer1org1 peer channel join -b mychannel.block"
docker exec cli bash -c "$peer0org2 peer channel join -b mychannel.block"
docker exec cli bash -c "$peer1org2 peer channel join -b mychannel.block"

#更新锚节点
docker exec cli bash -c "$peer0org1 peer channel update -o orderer.example.com:7050 -c mychannel -f ./channel-artifacts/Org1MSPanchors.tx"
docker exec cli bash -c "$peer0org2 peer channel update -o orderer.example.com:7050 -c mychannel -f ./channel-artifacts/Org2MSPanchors.tx"

#安装链码，只需要在背书节点安装
docker exec cli bash -c "$peer0org1 peer chaincode install -n datashare -v 1.0 -l golang -p github.com/chaincode"
docker exec cli bash -c "$peer0org2 peer chaincode install -n datashare -v 1.0 -l golang -p github.com/chaincode"

#实例化链码，链码必须实例化后才可以使用
docker exec cli bash -c "$peer0org1 peer chaincode instantiate -o orderer.example.com:7050  -C mychannel -n datashare -l golang -v 1.0 -c '{\"Args\":[\"test\",\"链码实例化成功\"]}' -P 'AND ('\''Org1MSP.peer'\'','\''Org2MSP.peer'\'')'"
sleep 5
#在peer0org1上query测试
docker exec cli bash -c "peer chaincode query -C mychannel -n datashare -c '{\"Args\":[\"queryRecord\",\"test\"]}'"

#在peer0org2上invoke测试
docker exec cli bash -c "peer chaincode invoke -o orderer.example.com:7050 -C mychannel -n datashare --peerAddresses peer0.org1.example.com:7051 --peerAddresses peer0.org2.example.com:9051 -c '{\"Args\":[\"sendData\",\"a\",\"b\",\"c\",\"d\",\"e\",\"f\"]}'"
sleep 5
#在peer1org1上安装并query测试
docker exec cli bash -c "peer chaincode query -C mychannel -n datashare -c '{\"Args\":[\"queryRecord\",\"a\"]}'"

#启动区块链浏览器
#替换密钥
priv_sk=$(ls crypto-config/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore)
cp -rf ./explorer/connection-profile/test-network-temp.json ./explorer/connection-profile/test-network.json
sed -i "s/priv_sk/$priv_sk/" ./explorer/connection-profile/test-network.json
docker-compose -f explorer/docker-compose.yaml up -d

#替换tape配置文件的私钥
#测试合约tps的命令：./tape --config=config.yaml --number=100
cp -rf ./tape/config-temp.yaml ./tape/config.yaml
sed -i "s/priv_sk/$priv_sk/" ./tape/config.yaml

#创建后端下载文件的路径
rm -rf ../application/server/files
mkdir -p ../application/server/files/downloadfiles
mkdir -p ../application/server/files/uploadfiles
