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

func GetLastIndexInClass(className string, lines *[]string) int {
	classIndex := 0
	searchedIndex := 0
	blockCount := 0
	checkBlock := false
	for index, value := range *lines {

		if strings.Contains(value, className) {
			classIndex = index
		}
		if classIndex != 0 {
			if strings.Contains(value, "{") {
				if checkBlock == false {
					checkBlock = true
				}
				blockCount++
			}
			if strings.Contains(value, "}") {
				blockCount--
			}

			if checkBlock && blockCount == 0 {
				searchedIndex = index
				break
			}
		}

	}
	return searchedIndex
}
func MakeHandler(packet string) {

	splitString := strings.Split(packet, "_")
	if len(splitString) < 2 {
		fmt.Printf("[%s]Packet format is invalid\n", packet)
	}
	if splitString[1] != "QRY" && splitString[1] != "REP" && splitString[1] != "CMD" {
		fmt.Printf("[%s]Packet format is invalid\n", packet)
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
		fmt.Printf("[%s] json Data is invalid\n", packet)
		return
	}

	registerPath = projectPath + registerPath
	headerPath = projectPath + headerPath
	cppPath = projectPath + cppPath

	checkIndexs1 := strings.LastIndex(headerPath, "\\")
	checkIndexs2 := strings.LastIndex(headerPath, ".")

	className := "class " + headerPath[checkIndexs1+1:checkIndexs2]

	insertString := fmt.Sprintf(handler, packet, packet)
	lines := ReadFileToString(registerPath)
	lines = insert(lines, len(lines), insertString)
	WriteFile(registerPath, lines)

	///////////////////////

	var searchedIndex int
	insertString = fmt.Sprintf(declare, packet)
	lines = ReadFileToString(headerPath)

	searchedIndex = GetLastIndexInClass(className, &lines)

	lines = insert(lines, searchedIndex, insertString)
	WriteFile(headerPath, lines)

	insertString = fmt.Sprintf(definition, packet)

	lines = ReadFileToString(cppPath)
	lines = insert(lines, len(lines), insertString)
	WriteFile(cppPath, lines)

}

var projectPath string

func main() {

	filePath := "protocol.json"

	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	// JSON 데이터를 구조체에 언마샬링합니다.
	err = json.Unmarshal(data, &jsonMap)
	if err != nil {
		fmt.Println("Error unmarshalling JSON data:", err)
		return
	}

	protocolMap = jsonMap["PROTOCOL"].(map[string]interface{})
	projectPath = jsonMap["ProjectPath"].(string)
	protocolDirPath := projectPath + "\\Servers\\ServerShare\\Server\\"

	result, resultStr := ExecCmd(protocolDirPath, "git", "diff", "--", "./ServerProtocol.h")

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
