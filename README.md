# data_system
基于区块链的数据管理系统
### 系统简介：
本系统使用RSA算法生成密钥对， RSA私钥用于用户身份认证；用户发送的数据将存储于IPFS， IPFS返回的CID（IPFS Hash）使用用户的RSA私钥加密后存储于区块链; 区块链部分使用Hyperledger Fabric,并用Hyperledger Explorer追踪交易

### 包含功能
1. 基于Fabric v1.4.4 first-network，四个peer一个orderer节点，使用docker部署
2. IPFS使用的是ipfs/kubo镜像，负责用户数据文件的存储，IPFS返回的CID存储于Fabric
3. 项目包含了Hyperledger Explorer（区块链浏览器），默认跟随脚本启动
4. 项目包含了tape对链码压测
5. 使用RSA公私钥鉴别用户身份（1024位）
6. 链码对传输记录进行存储，包含：发送者公钥、接收者公钥、文件在IPFS的加密CID（由发送者或接收者的私钥加密）、文件名、时间戳、Fabric交易id
7. 后端使用gin框架实现，
   使用go fabric sdk调用智能合约；使用go-IPfs-api上传与下载用户文件；使用uuid对用户的文件名（下载时）进行加密

### 安装步骤
1. 安装ubuntu 20.04
   
2. 向/etc/hosts 写入：  
   ````bash
    IP orderer.example.com
    IP peer0.org1.example.com
    IP peer1.org1.example.com
    IP peer0.org2.example.com
    IP peer1.org2.example.com
   ````
   这里的IP填写分两种情况：
- 区块链和后端都在虚拟机内或云服务器上跑此项目，IP设置为127.0.0.1即可
- 云服务器跑区块链，本地跑后端，则修改本地hosts文件，IP设置为云服务器IP

3. 全局替换127.0.0.1为自己的IP
- 这里的IP是指后端服务器的IP，因为涉及到前端HTTP请求。如果是在全在虚拟机内运行此项目，则设置成127.0.0.1，如果是云服务器则设置为服务器的公网IP

4. 启动区块链部分
   ````bash
   cd blockchain
   ./start.sh
   ````
5. 启动前后端
   ````bash
   cd application/server
   go run main.go
   ````
6. 如果是服务器
   在防火墙放行9090和8080TCP端口
7. 打开网页
   ip：9090/web

# 注意：
1. 如果全部是在虚拟机内操作，需要修改的IP改为127.0.0.1即可
2. 提示密钥不对请检查是否修改好hosts