package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var BinaryURL string
var PrinterAddress string
var scannedFilePath string
func main(){
	PrinterAddress = os.Args[1]
	

	for {
		ss := scanStatus()
		if ss == "Idle" {
			break
		}
		fmt.Println("waiting for scanner, status is", ss)
		time.Sleep(5 * time.Second)
	}

	jobLocation := startScanA4()
	fmt.Println("Job info file is located at", jobLocation)
	for {
		js := jobStatus(jobLocation)
		fmt.Print(".")
		if js == "Completed" {
			break
		}
		time.Sleep(5 * time.Second)

	}
	fmt.Println()
	fmt.Println("pdf file is ready")
	//return scannedFilePath
}

func filePath() string {
	u, err := user.Current()
	checkErr(err)
	dir := u.HomeDir + "/Desktop"
	//dir := fileDir
	dir = "."
	i := 0
	for {
		i += 1
		path := dir + "/Scanned File " + strconv.Itoa(i) + ".pdf"
		if _, err = os.Stat(path); os.IsNotExist(err) {
			return path
		}
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func scannerAddress(path string) string {
	if string(path[0]) != "/" {
		path = "/" + path
	}
	return ("http://" + PrinterAddress + path)
}

func startScan() string {
	payload := `<scan:ScanJob xmlns:scan="http://www.hp.com/schemas/imaging/con/cnx/scan/2008/08/19" xmlns:dd="http://www.hp.com/schemas/imaging/con/dictionaries/1.0/" xmlns:fw="http://www.hp.com/schemas/imaging/con/firewall/2011/01/05"><scan:XResolution>300</scan:XResolution><scan:YResolution>300</scan:YResolution><scan:XStart>0</scan:XStart><scan:YStart>0</scan:YStart><scan:Width>2550</scan:Width><scan:Height>3300</scan:Height><scan:Format>Pdf</scan:Format><scan:CompressionQFactor>25</scan:CompressionQFactor><scan:ColorSpace>Color</scan:ColorSpace><scan:BitDepth>8</scan:BitDepth><scan:InputSource>Platen</scan:InputSource><scan:GrayRendering>NTSC</scan:GrayRendering><scan:ToneMap><scan:Gamma>1000</scan:Gamma><scan:Brightness>1000</scan:Brightness><scan:Contrast>1000</scan:Contrast><scan:Highlite>179</scan:Highlite><scan:Shadow>25</scan:Shadow></scan:ToneMap><scan:ContentType>Document</scan:ContentType></scan:ScanJob>`
	resp, err := http.Post(scannerAddress("/Scan/Jobs"), "application/xml", strings.NewReader(payload))
	checkErr(err)
	location := resp.Header.Get("Location")
	return location
}

func startScanA4() string {
	payload := `<scan:ScanJob xmlns:scan="http://www.hp.com/schemas/imaging/con/cnx/scan/2008/08/19" xmlns:dd="http://www.hp.com/schemas/imaging/con/dictionaries/1.0/" xmlns:fw="http://www.hp.com/schemas/imaging/con/firewall/2011/01/05"><scan:XResolution>300</scan:XResolution><scan:YResolution>300</scan:YResolution><scan:XStart>0</scan:XStart><scan:YStart>0</scan:YStart><scan:Width>2480</scan:Width><scan:Height>3508</scan:Height><scan:Format>Pdf</scan:Format><scan:CompressionQFactor>25</scan:CompressionQFactor><scan:ColorSpace>Color</scan:ColorSpace><scan:BitDepth>8</scan:BitDepth><scan:InputSource>Platen</scan:InputSource><scan:GrayRendering>NTSC</scan:GrayRendering><scan:ToneMap><scan:Gamma>1000</scan:Gamma><scan:Brightness>1000</scan:Brightness><scan:Contrast>1000</scan:Contrast><scan:Highlite>179</scan:Highlite><scan:Shadow>25</scan:Shadow></scan:ToneMap><scan:ContentType>Document</scan:ContentType></scan:ScanJob>`
	resp, err := http.Post(scannerAddress("/Scan/Jobs"), "application/xml", strings.NewReader(payload))
	checkErr(err)
	location := resp.Header.Get("Location")
	return location
}

func downloadFile() {
	path := filePath()
	file, err := os.Create(path)
	checkErr(err)
	defer file.Close()
	resp, err := http.Get(scannerAddress(BinaryURL))
	checkErr(err)
	defer resp.Body.Close()
	n, err := io.Copy(file, resp.Body)
	checkErr(err)
	fmt.Println()
	fmt.Println(n, "bytes written")
	fmt.Println(path)
	scannedFilePath = path
	//openFile(path)
}

func openFile(path string) {
	env := os.Environ()
	binary, err := exec.LookPath("open")
	err = syscall.Exec(binary, []string{"open", path}, env)
	checkErr(err)
}

func extractBinaryUrl(data []byte) {
	var resp struct {
		ScanJob struct {
			PreScanPage struct {
				BinaryURL string
			}
		}
	}
	_ = xml.Unmarshal(data, &resp)
	BinaryURL = resp.ScanJob.PreScanPage.BinaryURL
	fmt.Println("Downloading file from", scannerAddress(BinaryURL))
	go downloadFile()
}

func jobStatus(url string) string {
	res, _ := http.Get(url)
	body, _ := ioutil.ReadAll(res.Body)
	if len(BinaryURL) == 0 {
		extractBinaryUrl(body)
	}
	var resp struct {
		JobState string
	}
	_ = xml.Unmarshal(body, &resp)
	return resp.JobState
}

func scanStatus() string {
	fmt.Println("checking scanner status")
	var body []byte
	respBody, err := http.Get(scannerAddress("/Scan/Status"))
	checkErr(err)
	body, err = ioutil.ReadAll(respBody.Body)
	checkErr(err)

	var resp struct {
		ScannerState string
	}
	err = xml.Unmarshal(body, &resp)
	checkErr(err)
	return resp.ScannerState
}