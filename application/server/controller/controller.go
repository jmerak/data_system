package controller

import (
	"fmt"
	"net/http"
	"os"
	"server/api"
	"server/api/rsa"
	"server/model"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

func GenerateKeyPair(c *gin.Context) {
	sk, pk := api.GenerateKeyPair()
	//sk:private key pk:public key
	c.JSON(http.StatusOK, gin.H{"sk": sk, "pk": pk})

}
func RestoreKey(c *gin.Context) {
	//convert sk
	pk, err := rsa.SktoPub(c.PostForm("sk"))
	if err != nil {
		c.String(http.StatusOK, "no sk input")
		return
	}
	c.String(http.StatusOK, pk)

}
func Upload(c *gin.Context) {
	//save file
	file, _ := c.FormFile("file")
	fmt.Printf("file:%v", fmt.Sprintf("./files/uploadfiles/%v", file.Filename))
	err := c.SaveUploadedFile(file, fmt.Sprintf("./files/uploadfiles/%v", file.Filename))
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("file upload failed,err:%v", err))
		return
	}
	// upload to ipfs
	cid := api.IpfsAdd(file.Filename)

	//assignment record
	var record model.Record
	c.ShouldBind(&record)
	// get activity
	record.Activity = c.PostForm("activity")
	// fmt.Printf(record.Activity)
	// fmt.Printf("%v", record.Sender)
	senderpk, err := rsa.SktoPub(record.Sender)
	// fmt.Printf("1111111111111111111111111111111111111111111111111111111111111111111111111111111111")
	if err != nil {
		c.String(http.StatusBadRequest, "no sk input")
		return
	}
	// fmt.Printf("22222222222222222222222222")
	record.SenderEncryptedCid = api.EncryptCid(cid, senderpk)
	record.RecevierEncryptedCid = api.EncryptCid(cid, record.Recevier)
	record.Filename = file.Filename
	// fmt.Printf("333333333333333333333333333333333333333333333333333333333")
	//upload to fabric
	var args [][]byte
	args = append(args, []byte(senderpk))
	args = append(args, []byte(record.Recevier))
	args = append(args, []byte(record.SenderEncryptedCid))
	args = append(args, []byte(record.RecevierEncryptedCid))
	args = append(args, []byte(record.Filename))
	args = append(args, []byte(record.Activity))
	for i:= 0; i < len(args); i++ {
		fmt.Printf("\n %d: \n" + string(args[i]) + "\n", i)
	}
	res, err := api.ChannelExecute("sendData", args)
	if err != nil {
		fmt.Printf("444444444444444444444444444444444444444")
		c.String(http.StatusBadRequest, fmt.Sprintf("execute chaincode failed err:%v", err))
		return
	}

	//return txid
	c.JSON(http.StatusOK, gin.H{
		"txid": res.TransactionID,
	})

}

func GetRecords(c *gin.Context) {
	//convert sk
	pk, err := rsa.SktoPub(c.PostForm("sk"))
	if err != nil {
		c.String(http.StatusBadRequest, "no sk input")
		return
	}
	//query chaincode
	var args [][]byte
	args = append(args, []byte(pk))
	res, err := api.ChannelQuery("queryRecord", args)
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("execute chaincode failed err:%v", err))
		return
	}
	//retun res
	fmt.Printf("/nres.Payload:" + string(res.Payload) + "/n")
	c.String(http.StatusOK, string(res.Payload))
}

func CreateFile(c *gin.Context) {
	//get pk
	pk, err := rsa.SktoPub(c.PostForm("sk"))
	if err != nil {
		c.String(http.StatusBadRequest, "no sk input")
		return
	}
	//ecid:encrypted cid
	ecid := c.PostForm("ecid")
	filename := c.PostForm("filename")
	if pk == "" {
		c.String(http.StatusBadRequest, "no sk input")
		return
	}
	//get old cid
	oldcid := api.DecryptCid(ecid, c.PostForm("sk"))
	//remove old file
	os.Remove(fmt.Sprintf("./files/uploadfiles/%v", filename))
	//get newfile
	file, _ := c.FormFile("file")
	fmt.Printf("file:%v", fmt.Sprintf("./files/uploadfiles/%v", file.Filename))
	err = c.SaveUploadedFile(file, fmt.Sprintf("./files/uploadfiles/%v", file.Filename))
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("file upload failed,err:%v", err))
		return
	}
	_ = api.IpfsUpdate(oldcid, file.Filename)
	c.Status(http.StatusOK)
}

func DeleteFile(c *gin.Context) {
	pk, err := rsa.SktoPub(c.PostForm("sk"))
	if err != nil {
		c.String(http.StatusBadRequest, "no sk input")
		return
	}
	//ecid:encrypted cid
	ecid := c.PostForm("ecid")
	filename := c.PostForm("filename")

	if pk == "" {
		c.String(http.StatusBadRequest, "no sk input")
		return
	}
	fmt.Printf("\necid:" + ecid + "\n")
	cid := api.DecryptCid(ecid, c.PostForm("sk"))
	err = api.IpfsDelete(cid)
	if err != nil {
		fmt.Printf("delete file fail!")
		c.String(http.StatusBadGateway, fmt.Sprintf("%v", err))
		return
	}
	os.Remove(fmt.Sprintf("./files/uploadfiles/%v", filename))
	c.Status(http.StatusOK)
}

func GetFile(c *gin.Context) {
	//convert sk
	pk, err := rsa.SktoPub(c.PostForm("sk"))
	if err != nil {
		c.String(http.StatusBadRequest, "no sk input")
		return
	}
	//ecid:encrypted cid
	ecid := c.PostForm("ecid")
	filename := c.PostForm("filename")
	if pk == "" {
		c.String(http.StatusBadRequest, "no sk input")
		return
	}
	cid := api.DecryptCid(ecid, c.PostForm("sk"))
	//generate uuid
	u1, _ := uuid.NewV4()
	//encryptedfilename
	enFilename := fmt.Sprintf("%v%v", u1, filename)
	fmt.Printf("sk:%v\n,ecid:%v\n,cid:%v\n,pk:%v\n", c.PostForm("sk"), ecid, cid, pk)
	//get file from ipfs
	err = api.IpfsGet(cid, enFilename)
	if err != nil {
		fmt.Printf("ccccccccccccccccccccccccccccccccccccccccccccc")
		c.String(http.StatusBadGateway, fmt.Sprintf("%v", err))
		return
	}
	//return filepath
	c.String(http.StatusOK, fmt.Sprintf("http://127.0.0.1:9090/downloadfile?filepath=%v", enFilename))
}

func DownloadFile(c *gin.Context) {
	enFilename := c.Query("filepath")
	//uuid 36
	filename := enFilename[36:]
	c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	_, err := os.Stat(fmt.Sprintf("./files/downloadfiles/%v", enFilename))
	if err != nil {
		c.String(http.StatusBadRequest, "文件已删除！请重新获取链接！")
		return
	}
	//return file
	c.File(fmt.Sprintf("./files/downloadfiles/%v", enFilename))
	//delete file
	// os.Remove(fmt.Sprintf("./files/downloadfiles/%v", enFilename))
}
