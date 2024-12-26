package api

import (
	"fmt"
	"os"
	"server/api/rsa"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	shell "github.com/ipfs/go-ipfs-api"
)

var (
	sdk           *fabsdk.FabricSDK                                              // Fabric SDK
	configPath    = "./api/config.yaml"                                          // 配置文件路径
	channelName   = "mychannel"                                                  // 通道名称
	user          = "Admin"                                                      // 用户
	chainCodeName = "datashare"                                                  // 链码名称
	endpoints     = []string{"peer0.org1.example.com", "peer0.org2.example.com"} // 要发送交易的节点
	sh            = shell.NewShell("127.0.0.1:6001")                             //ipfs api
)

// Init 初始化
func Init() {

	var err error
	// 通过配置文件初始化SDK
	sdk, err = fabsdk.New(config.FromFile(configPath))
	if err != nil {
		panic(err)
	}

}

func GenerateKeyPair() (string, string) {
	rsakey, _ := rsa.GenerateRsaKeyBase64(1024)
	return rsakey.PrivateKey, rsakey.PublicKey
}
func EncryptCid(cid string, publicKey string) string {
	//ecid:encrypted cid
	ecid, err := rsa.RsaEncryptToBase64([]byte(cid), publicKey)
	if err != nil {
		return err.Error()
	}
	return ecid
}

func DecryptCid(ecid string, privatekey string) string {
	// fmt.Printf("================ecid:%v,sk:%v", ecid, privatekey)
	cid, err := rsa.RsaDecryptByBase64(ecid, privatekey)
	if err != nil {
		return err.Error()
	}
	return string(cid)
}

func IpfsAdd(filename string) string {
	ipfsfile, _ := os.Open(fmt.Sprintf("./files/uploadfiles/%v", filename))
	defer ipfsfile.Close()
	cid, err := sh.Add(ipfsfile)
	if err != nil {
		return fmt.Sprintf("file upload failed,err:%v", err)
	}
	return cid
}

func IpfsGet(cid string, filename string) error {
	// fmt.Printf("cid:%v,filename:%v", cid, filename)
	err := sh.Get(cid, fmt.Sprintf("./files/downloadfiles/%v", filename))
	if err != nil {
		fmt.Printf("err:%v", err)
		return fmt.Errorf("ipfs get file faile,err:%v", err)
	}
	return nil

}

func IpfsUpdate(oldcid string, filename string) string {
	//读取原文件，错误判断
	ipfsfile, err := os.Open(fmt.Sprintf("./files/uploadfiles/%v", filename))
	if err != nil {
		return fmt.Sprintf("ipfs get file faile,err:%v", err)
	}
	defer ipfsfile.Close()

	//上传新文件到IPFS网络，并返回新的CID
	newcid, err := sh.Add(ipfsfile)
	return newcid

}

func IpfsDelete(cid string) error {
	fmt.Printf(cid)
	err := sh.Unpin(cid)
	if err != nil {
		return err
	}
	return nil
}

// ChannelExecute 区块链交互
func ChannelExecute(fcn string, args [][]byte) (channel.Response, error) {
	// 创建客户端，表明在通道的身份
	ctx := sdk.ChannelContext(channelName, fabsdk.WithUser(user))
	cli, err := channel.New(ctx)
	if err != nil {
		return channel.Response{}, err
	}
	// 对区块链账本的写操作（调用了链码的invoke）
	resp, err := cli.Execute(channel.Request{
		ChaincodeID: chainCodeName,
		Fcn:         fcn,
		Args:        args,
	}, channel.WithTargetEndpoints(endpoints...))
	if err != nil {
		return channel.Response{}, err
	}
	//返回链码执行后的结果
	return resp, nil
}

// ChannelQuery 区块链查询
func ChannelQuery(fcn string, args [][]byte) (channel.Response, error) {
	// 创建客户端，表明在通道的身份
	ctx := sdk.ChannelContext(channelName, fabsdk.WithUser(user))
	cli, err := channel.New(ctx)
	if err != nil {
		return channel.Response{}, err
	}
	// 对区块链账本查询的操作（调用了链码的invoke），只返回结果
	resp, err := cli.Query(channel.Request{
		ChaincodeID: chainCodeName,
		Fcn:         fcn,
		Args:        args,
	}, channel.WithTargetEndpoints(endpoints...))
	if err != nil {
		return channel.Response{}, err
	}
	//返回链码执行后的结果
	return resp, nil
}
