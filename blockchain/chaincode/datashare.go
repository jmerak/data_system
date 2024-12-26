package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Define the Smart Contract structure
type SmartContract struct {
}

// Define record structure
type Record struct {
	Sender                string `json:"sender"`
	Receiver              string `json:"receiver"`
	SenderEncrypted_cid   string `json:"secid"`
	ReceiverEncrypted_cid string `json:"recid"`
	Filename              string `json:"filename"`
	Activity              string `json:"activity"`
	Txid                  string `json:"txid"`
	Timestamp             string `json:"timestamp"`
}

type Records []Record

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Get the args from the transaction proposal
	args := APIstub.GetStringArgs()
	if len(args) != 2 {
		return shim.Error("Incorrect arguments. Expecting a key and a value")
	}
	// We store the key and the value on the ledger
	err := APIstub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to create asset: %s", args[0]))
	}
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "queryRecord" {
		return s.queryRecord(APIstub, args)
	} else if function == "sendData" {
		return s.sendData(APIstub, args)
	}
	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) queryRecord(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	recordsAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(recordsAsBytes)
}

func (s *SmartContract) sendData(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	//验证参数个数
	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 6")
	}
	//生成时间戳
	txtime, err := APIstub.GetTxTimestamp()
	if err != nil {
		return shim.Error(err.Error())
	}
	timesecond := txtime.Seconds
	time := time.Unix(timesecond+28800, 0).Format("2006-01-02 15:04:05")

	//实例化Record结构体并使用传入的参数赋值
	var record = Record{Sender: args[0], Receiver: args[1], SenderEncrypted_cid: args[2], ReceiverEncrypted_cid: args[3], Filename: args[4], Activity: args[5], Timestamp: time, Txid: APIstub.GetTxID()}
	//将数据传输记录保存到发送方的记录中
	recordsAsBytes_sender, err := APIstub.GetState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	var records_sender Records
	json.Unmarshal(recordsAsBytes_sender, &records_sender)
	records_sender = append(records_sender, record)
	recordsAsBytes_sender, err = json.Marshal(records_sender)
	if err != nil {
		return shim.Error(err.Error())
	}
	//record to fabric
	APIstub.PutState(args[0], recordsAsBytes_sender)

	//将数据传输记录保存到接受者的记录中
	recordsAsBytes_receiver, err := APIstub.GetState(args[1])
	if err != nil {
		return shim.Error(err.Error())
	}
	var records_receiver Records
	json.Unmarshal(recordsAsBytes_receiver, &records_receiver)
	records_receiver = append(records_receiver, record)
	recordsAsBytes_receiver, err = json.Marshal(records_receiver)
	if err != nil {
		return shim.Error(err.Error())
	}
	//record to fabric
	APIstub.PutState(args[1], recordsAsBytes_receiver)

	return shim.Success(nil)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {
	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
