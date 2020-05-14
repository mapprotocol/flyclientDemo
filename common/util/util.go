package util

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"encoding/json"
	"reflect"
)

// 获得时间戳相关的唯一id
func GetUniqueId() string {
	curNano := time.Now().UnixNano()
	r := rand.New(rand.NewSource(curNano))
	return fmt.Sprintf("%d%06v", curNano, r.Int31n(1000000))
}

// 解析json字符串
func ParseJson(data string, result interface{}) error {
	//var json = jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Unmarshal([]byte(data), result)
}

// json转字符串
func StringifyJson(data interface{}) string {
	//var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b, e := json.Marshal(&data)
	if e != nil {
		//cslog.Debug().Err(e).Msg("stringify json报错")
	}
	return string(b)
}

// 解析json bytes
func ParseJsonFromBytes(data []byte, result interface{}) error {
	return json.Unmarshal(data, result)
}

// json bytes转字符串
func StringifyJsonToBytes(data interface{}) []byte {
	b, _ := json.Marshal(&data)
	return b
}

func StringifyJsonToBytesWithErr(data interface{}) ([]byte, error) {
	b, err := json.Marshal(&data)
	return b, err
}

// LoadJSON reads the given file and unmarshals its content.
func LoadJSON(file string, val interface{}) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(content, val); err != nil {
		if syntaxerr, ok := err.(*json.SyntaxError); ok {
			line := findLine(content, syntaxerr.Offset)
			return fmt.Errorf("JSON syntax error at %v:%v: %v", file, line, err)
		}
		return fmt.Errorf("JSON unmarshal error in %v: %v", file, err)
	}
	return nil
}

// findLine returns the line number for the given offset into data.
func findLine(data []byte, offset int64) (line int) {
	line = 1
	for i, r := range string(data) {
		if int64(i) >= offset {
			return
		}
		if r == '\n' {
			line++
		}
	}
	return
}

func FileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

func AbsolutePath(Datadir string, filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(Datadir, filename)
}

// 获取Home dir
func HomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return "~/"
}

// 杀掉某个应用
func KillProcess(appName string) {
	exec.Command("/bin/bash", "-c", fmt.Sprintf("ps aux|grep %v|awk '{print $2}'|xargs kill ", appName)).Output()
	//cslog.Info().Err(err).Str("result", string(rb)).Str("程序名称", appName).Msg("执行kill程序的命令")
}

// 杀掉远程的某个应用
func KillRemoteProcess(sshConnStr, appName string) {
	ExecCmdBySSH(sshConnStr, fmt.Sprintf("ps aux|grep %v|awk '{print $2}'|xargs kill ", appName))
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// 根据匹配条件获取本机ip
func GetCurPcIp(matcher string) string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		if strings.Contains(a.String(), matcher) {
			return strings.Split(a.String(), "/")[0]
		}
	}
	return ""
}

// 判断是否是远端的host
func IsRemoteHost(host string, ipMatcher string) bool {
	curIp := GetCurPcIp(ipMatcher)
	// 不是任何本地ip配置，那么就是远端的
	if host != "" && host != "localhost" && host != "127.0.0.1" && host != curIp {
		return true
	}
	return false
}

//// 要确保不用输入用户名
//func CpFileToTargetHost(filePath string, loginUser string, targetHost string, targetPath string) {
//	scpCmdStr := fmt.Sprintf("scp %v %v@%v:%v", filePath, loginUser, targetHost, targetPath)
//	log.Debug("执行copy file to remote", "cmd", scpCmdStr)
//	rb, err := exec.Command("/bin/bash", "-c", scpCmdSr).Output()
//	if err != nil {
//		log.Warn("执行copy file to remote报错", "err", err, "result", string(rb))
//	}
//}

// 用SSH执行命令
func ExecCmdBySSH(sshConnStr string, cmdStr string) {
	//rb, err := exec.Command("ssh", sshConnStr, `"` + cmdStr + `"`).Output()
	tmpShFile := filepath.Join("/tmp", GetUniqueId()+"_tmp_cs.sh")
	ioutil.WriteFile(tmpShFile, []byte(cmdStr), 0755)

	cmdExecStr := fmt.Sprintf(`ssh %v bash -s < %v`, sshConnStr, tmpShFile)
	log.Debug("执行远端ssh命令", "cmd exec str", cmdExecStr, "cmdStr", cmdStr)
	rb, err := exec.Command("/bin/bash", "-c", cmdExecStr).Output()
	if err != nil {
		log.Warn("执行远端ssh命令报错", "err", err, "result", string(rb))
	}
	os.RemoveAll(tmpShFile)
}

// 判断是否是测试环境
func IsTestEnv() bool {
	return flag.Lookup("test.v") != nil
}

// 检查slice是否包含str
func StrContainsInSlice(ss []string, str string) bool {
	for _, s := range ss {
		if s == str {
			return true
		}
	}
	return false
}

func StopChanClosed(stop chan struct{}) bool {
	if stop == nil {
		return true
	}
	select {
	case _, ok := <-stop:
		return !ok
	default:
		return false
	}
}

func ExecuteFuncWithTimeout(f func(), t time.Duration) {
	finish := make(chan struct{})
	go func() {
		f()
		close(finish)
	}()

	timer := time.NewTimer(t)
	// 调用者会阻塞在这里直到f执行完或是定时器触发超时
	select {
	case <-finish:
		timer.Stop()
		return
	case <-timer.C:
		panic("exec func timeout")
	}
}

// 如果外边没有调用finishFunc则必然会触发timeout
func SetTimeout(timeoutFunc func(), dur time.Duration) (finishFunc func()) {
	finishChan := make(chan struct{})

	timer := time.NewTimer(dur)
	// 这个协程就是个计时器，如果外边不做任何操作，它也会在固定时间
	go func() {
		select {
		case <-finishChan:
			timer.Stop()
			return
		case <-timer.C:
			timeoutFunc()
		}
	}()

	return func() {
		close(finishChan)
	}
}

// like: []interface{} -> []*Block, or []*Block -> []interface{}
func InterfaceSliceCopy(to, from interface{}) {
	toV := reflect.ValueOf(to)
	fromV := reflect.ValueOf(from)
	fLen := fromV.Len()
	for i := 0; i < fLen; i++ {
		// support bothway copy
		toV.Index(i).Set(reflect.ValueOf(fromV.Index(i).Interface()))
	}
}
