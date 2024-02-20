package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/textproto"
	"os"
	"os/exec"
	"strings"
)

var jsonMap map[string]interface{}
var protocolMap map[string]interface{}

func ExecCmd(cmdDir string, cmdStr string, args ...string) (bool, string) {

	fmt.Printf("%s 명령어 실행 시작\n", cmdStr)
	cmd := exec.Command(cmdStr, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if cmdDir != "" {
		cmd.Dir = cmdDir
	}

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
		return false, err.Error()
	}

	err = cmd.Wait()

	if stderr.String() != "" {
		return false, stderr.String()
	}

	var resultString string
	resultString = out.String()
	resultString = textproto.TrimString(resultString)

	fmt.Printf("%s 명령어 실행 종료\n", cmdStr)
	return true, resultString
}

func insert(a []string, index int, value string) []string {
	if len(a) == index { // nil or empty slice or after last element
		return append(a, value)
	}
	a = append(a[:index+1], a[index:]...) // index < len(a)
	a[index] = value
	return a

}
func ReadFileToString(path string) []string {
	file, err := os.ReadFile(path)
	if err != nil {
		return []string{}
	}

	fileText := string(file)

	return strings.Split(fileText, "\n")
}

func WriteFile(path string, dataString []string) bool {
	data := strings.Join(dataString, "\n")
	err := os.WriteFile(path, []byte(data), os.ModePerm)
	if err != nil {
		return false
	}

	return true
}

func MakeHandler(packet string) {

	splitString := strings.Split(packet, "_")
	if len(splitString) < 2 {
		fmt.Println("Packet format is invalid")
	}
	protocol := protocolMap[splitString[0]].(map[string]interface{})

	qryData := protocol[splitString[1]].(map[string]interface{})

	registerPath := qryData["RegisterPath"].(string)
	headerPath := qryData["headerPath"].(string)
	cppPath := qryData["cppPath"].(string)
	handler := qryData["Handler"].(string)
	declare := qryData["Declare"].(string)
	definition := qryData["Definition"].(string)

	if registerPath == "" && headerPath == "" && cppPath == "" && handler == "" && declare == "" && definition == "" {
		fmt.Printf("[%s] json 정보가 없다", packet)
		return
	}

	insertString := fmt.Sprintf(handler, packet, packet)
	lines := ReadFileToString(registerPath)
	lines = insert(lines, len(lines), insertString)
	WriteFile(registerPath, lines)

	///////////////////////

	var searchedIndex int
	insertString = fmt.Sprintf(declare, packet)
	lines = ReadFileToString(headerPath)

	for index, value := range lines {

		if strings.Contains(value, "//Packet Handler") {
			searchedIndex = index
		}
	}

	lines = insert(lines, searchedIndex+1, insertString)
	WriteFile(headerPath, lines)

	insertString = fmt.Sprintf(definition, packet)

	lines = ReadFileToString(cppPath)
	lines = insert(lines, len(lines), insertString)
	WriteFile(cppPath, lines)

}

func main() {

	filePath := "protocol.json"

	// 파일로부터 JSON 데이터를 읽습니다.
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	// JSON 데이터를 저장할 구조체 변수를 선언합니다.

	// JSON 데이터를 구조체에 언마샬링합니다.
	err = json.Unmarshal(data, &jsonMap)
	if err != nil {
		fmt.Println("Error unmarshalling JSON data:", err)
		return
	}

	protocolMap = jsonMap["PROTOCOL"].(map[string]interface{})

	result, resultStr := ExecCmd("C:\\ProjectR\\AvatarGlobal_Server\\Servers\\ServerShare\\Server\\", "git", "diff", "--", "./ServerProtocol.h")

	if result == false {
		fmt.Println(resultStr)
		return
	}

	line := strings.Split(resultStr, "\n")
	diffString := []string{}

	for _, value := range line {

		if strings.Contains(value, "+\t\t") {

			after := strings.ReplaceAll(value, "\t", "")
			after = strings.ReplaceAll(value, "+\t\t", "")
			index := strings.Index(after, ",")
			if index != -1 {
				after = after[:index]
			}
			diffString = append(diffString, after)
		}
	}

	for _, value := range diffString {

		result := strings.Split(value, "_")

		switch result[0] {
		case "LGAD", "LGGW", "GWLG", "GWGS", "GSGW", "GSGD", "GDGS", "TOLG":
			MakeHandler(value)

		default:

		}

	}

}
