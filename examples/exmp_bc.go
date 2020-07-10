package main

import (
	"os"
	bc "./blockchain" 
)

const (
	BLOCKCHAIN_DB = "blockchain.db"
)

const (
	purse1 = "MIIBPAIBAAJBAO+M9NcRJAoZucR2PoenzfTN3bpfpJuuq4bXwbzabAh+XhmY4I5LfNN016SNhjzHqp8y6uD/pdHmp6tVcXZ5Gk0CAwEAAQJBAJdiNLhVFLltWIBXWVzRJs616pGFL7lSwQMOBlkaj3stZ7koGsIcT3PaQRMf23FKq1gUzMPyey0njP4+WJ7ZTuECIQD5k7Nn8HFYAHjnadoJFTzxHzbWKEpLcULNmtFoGGL65wIhAPW3MtXdWdy5zcWFad4Rf8SCmsip4NMWAQBaIBsvyi6rAiEA9e3AVISO+7uDZ+1bZ6Xg5mzuuTr8crjJrhaHkL6vEjMCIAllSfJXlnoUOMQWx+eP77zJ6bVGmaw4qS0kRRoFB8wlAiEA8kYB4ODHHgP/4kn74gVOqDWKGOZZk2uxNTWzgQnF4f0="
	purse2 = "MIIBPAIBAAJBAOJhasuqmGs/YSTkQOsrXmiFLvdqwRVTUJCJ3LG64pietxsAZKiGs57uGHjs4nDTmMiHkeo6OxoLBlUJvGB4RV8CAwEAAQJAD72wXFsPYB23FoackP2iAeiI3IFaza3OY06CvwT8rpnxHcinpeAfxeanEraHsWNpPmzqhsx066Wuup3idJQxkQIhAP96uD/G+taZ1owsDid+aCzNjSDNFzxLqCibygv3Ghm5AiEA4teEV4XvF5yFqDMpwqXYFuGFXvwMmKuFuyhjo3b7g9cCIQCHMvByf+Ca1MqfX2kWKWUD4LuA0kgrlnYlF1yzPE9JsQIhAJNJroOJ0lG5bstk1ROuzV0l94xSCffDPyTjt7Z21h6nAiEA5atlLYR1LaId1rz1wJFMi5O2iMFJjLC1g6USqO0s5ZA="
)

func main() {
	user1 := bc.LoadUser(purse1)
	user2 := bc.LoadUser(purse2)

	err := bc.NewChain(BLOCKCHAIN_DB, user1.Address())
	if err != nil {
		panic(err)
	}
	chain := bc.LoadChain(BLOCKCHAIN_DB)

	for i := 0; i < 3; i++ {
		block := bc.NewBlock(user2.Address(), chain.LastHash())
		block.AddTransaction(chain, block.NewTransaction(user1, "aaa", 11))
		block.AddTransaction(chain, block.NewTransaction(user1, "bbb", 11))
		chain.PushBlock(user2, block)
	}

	chain.PrintChain()
}

func fileIsExist(filename string) bool {
    if _, err := os.Stat(filename); os.IsNotExist(err) {
        return false
    }
    return true
}
