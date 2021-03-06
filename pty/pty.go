package pty

import (
	"fmt"
	"github.com/gliderlabs/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"github.com/FrelDX/kcos/cluster"
	"github.com/FrelDX/kcos/common"
	interrupt "github.com/FrelDX/kcos/util"
	"log"
	"strconv"
)

const (
	// Fill in spaces when less than 60 characters
	DisplayLengthPod = 60
	DisplayLengthNameSpace = 20
)

// Global pod information storage, used to connect to pod shell
var podIndex []cluster.PodList
func newPty(namespace string,podName string,container string,stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	// to exec,but /bin/bash first then /bin/sh
	exec := func()  error{
		config := common.Config()
		client := common.NewClient()
		err := Remotepty(client, config, namespace, podName, "/bin/bash", container, stdin, stdout, stderr)
		if err != nil {
			err = Remotepty(client, config, namespace, podName, "/bin/sh", container, stdin, stdout, stderr)
			if err != nil {
				log.Print(err)
			}
		}
		return  nil
	}
	// Processing signal
	return interrupt.Chain(nil, func() {
		log.Print("go to interface")
	}).Run(exec)
	}

func MainInterface(s ssh.Session){
	WelcomePage(s)
	term := terminal.NewTerminal(s, s.User() + "# ")
	line := ""
	// get user input
	for {
		line, _ = term.ReadLine()
		if line == "quit" {
			break
		}else if line == "p" {
			DisplayAllPod(s)
		}else if line =="m"{
			WelcomePage(s)
		} else if line =="n"{
			namespace :=DisplayNameSpace(s)
			// Operations corresponding to namespace
			n:=""
			for{
				n, _ = term.ReadLine()
				if n == "m"{
					WelcomePage(s)
					break
				}
				if n == "quit"{
					break
				}
				number, err := strconv.Atoi(n)
				if err ==nil{
					// Prevent index out of range
					if number < len((namespace)){
						DisplayNamespacePod(s,(namespace)[number])
						break
					}
					log.Println(err)
				}
				break
			}
		} else if line == "d"{
			DisplayDeploy(s)
		}
		number, err := strconv.Atoi(line)
		if err == nil{
			if number < len((podIndex)){
				// Multiple containers
				if len(podIndex[number].Containers) >1{
					io.WriteString(s, fmt.Sprint(SetColorBlue("Please select a container "),"\n"))
					for i,c:= range   podIndex[number].Containers{
						io.WriteString(s,fmt.Sprint(SetColorBlue(strconv.Itoa(i)),"\t",SetColorRed(c),"\n"))
					}
					// Get user selected container
					container, _ := term.ReadLine()
					containerNumber, err :=strconv.Atoi(container)
					if err ==nil{
						if containerNumber < len((podIndex[number].Containers)) {
							newPty(podIndex[number].Namespaces,podIndex[number].Name,podIndex[number].Containers[containerNumber],s,s,s)
						}
					}
				}
				newPty(podIndex[number].Namespaces,podIndex[number].Name,"",s,s,s)
			}
			log.Println(err)
			continue
		}
	}
}
func DisplayAllPod(s ssh.Session)  {
	pod :=cluster.GetPodList("")
	// to DisplayPod
	DisplayPod(pod,s)
}
func DisplayNamespacePod(s ssh.Session,namespace string)  {
	pod :=cluster.GetPodList(namespace)
	// to DisplayPod
	DisplayPod(pod,s)
}
func SetColorGreen(msg string) string {
	return  fmt.Sprintf("\033[32;1m%s\033[0m",msg)
}
func SetColorRed(msg string) string {
	return  fmt.Sprintf("\033[31;1m%s\033[0m",msg)
}
func SetColorBlue(msg string) string {
	return  fmt.Sprintf("\033[34;1m%s\033[0m",msg)
}
func SetColorYellow(msg string) string {
	return  fmt.Sprintf("\033[33;1m%s\033[0m",msg)
}
func WelcomePage(s ssh.Session)  {
	io.WriteString(s, SetColorGreen("\t\t\t欢迎登陆kcos (kube-console-on-ssh)\n"))
	io.WriteString(s, SetColorGreen("\t\t\t输入quit退出当前终端\n"))
	io.WriteString(s, SetColorGreen("\t\t\t当前登陆的用户:" + s.User())+"\n")
	io.WriteString(s, SetColorGreen("\t\t\t选择对应的数字连接到对应的pod shell\n"))
	io.WriteString(s, SetColorGreen("\t\t\t输入p查看所有可以登陆的pod列表\n"))
	io.WriteString(s, SetColorGreen("\t\t\t输入n 查看namespace下所有的pod\n"))
	io.WriteString(s, SetColorGreen("\t\t\t输入m 返回到主菜单\n"))
}
func DisplayPod(pod []cluster.PodList,s ssh.Session)  {
	for i,k:=range pod{
		// Unify the length of the name to prevent garbled characters when displaying
		namespace :=k.Namespaces
		if len(k.Namespaces)<DisplayLengthNameSpace{
			for i:=0;i<DisplayLengthNameSpace-len(k.Namespaces);i++{
				namespace = namespace + " "
			}
		}
		pod := k.Name
		if len(k.Name) < DisplayLengthPod{
			for i:=0;i<DisplayLengthPod-len(k.Name);i++{
				pod = pod +" "
			}
		}
		io.WriteString(s, fmt.Sprint(SetColorBlue(strconv.Itoa(i)),"\t",SetColorRed(namespace),"\t",SetColorGreen(pod),"\t",SetColorYellow(k.Ip),"\n"))
	}
	// 最新的信息刷新到全局pod信息保存处
	podIndex = pod
}
func DisplayNameSpace(s ssh.Session) []string  {
	namespace :=cluster.GetNameSpaces()
	for i,k:=range namespace{
		io.WriteString(s, fmt.Sprint(SetColorBlue(strconv.Itoa(i)),"\t",SetColorGreen(k),"\n"))
	}
	return namespace
}
func DisplayDeploy(s ssh.Session)  {

}