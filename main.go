package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Server struct {
	ServerName    string
	ServerUrl     string
	tempoExecucao float64
	status        int
	dataFalha     string
}

func criarListaServidores(serverList *os.File) []Server {
	csvReader := csv.NewReader(serverList)
	data, err := csvReader.ReadAll()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var servidores []Server
	for i, line := range data {
		if i > 0 {
			servidor := Server{
				ServerName: line[0],
				ServerUrl:  line[1],
			}
			servidores = append(servidores, servidor)
		}
	}
	return servidores
}

func checkServer(servidores []Server) []Server {
	var downServers []Server
	for _, servidor := range servidores {
		agora := time.Now() //registrar momento da execução da verif
		get, err := http.Get(servidor.ServerUrl)
		if err != nil {
			fmt.Printf("Server %s is down [%s]\n", servidor.ServerName, err.Error())
			servidor.status = 0
			servidor.dataFalha = agora.Format("02-01-2006 15:04")
			downServers = append(downServers, servidor)
			continue
		}
		servidor.status = get.StatusCode
		if servidor.status != 200 {
			servidor.dataFalha = agora.Format("02-01-2006 15:04")
			downServers = append(downServers, servidor)
		}
		servidor.tempoExecucao = time.Since(agora).Seconds()
		fmt.Printf("Status: [%d] Tempo de carga: [%f] URL: [%s]\n", servidor.status, servidor.tempoExecucao, servidor.ServerUrl)
	}

	return downServers

}

func openFiles(serverListFile string, downtimeFile string) (*os.File, *os.File) {
	//retorna 2 arquivos os.File
	serverList, err := os.OpenFile(serverListFile, os.O_RDONLY, 0666) //SOMENTE LEITURA O_RDONLY
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	downtimeList, err := os.OpenFile(downtimeFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return serverList, downtimeList
}

func generateDowntime(downtimeList *os.File, downServers []Server) {
	csvWriter := csv.NewWriter(downtimeList)
	for _, servidor := range downServers {
		line := []string{servidor.ServerName, servidor.ServerUrl, servidor.dataFalha, fmt.Sprintf("%f", servidor.tempoExecucao), fmt.Sprintf("%d", servidor.status)}
		csvWriter.Write(line) //GRAVA A LINHA
	}
	csvWriter.Flush() //PASSA OS DADOS DE MEMORIA PARA O ARQ
}

func main() {
	serverList, downtimeList := openFiles(os.Args[1], os.Args[2])
	defer serverList.Close() //gatante q qnd fechar a func fica td bem
	defer downtimeList.Close()
	servidores := criarListaServidores(serverList)

	downServers := checkServer(servidores)
	generateDowntime(downtimeList, downServers)

}
