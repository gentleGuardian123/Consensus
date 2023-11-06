package PoS

import (
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddressesToJson(t *testing.T) {
	addresses := []string{
		"abcd",
		"efgh",
	}

	ret, err := json.MarshalIndent(addresses, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	t.Log("The json format of addresses is:")
	t.Log(ret)

	if err := os.WriteFile("addresses_test.json", ret, 0644); err != nil {
		log.Fatal(err)
	}


	var _addresses []string

	if err := json.Unmarshal(ret, &_addresses); err != nil {
		log.Fatal(err)
	}

	t.Log("The resolved result from json is:")
	t.Log(_addresses)

	assert.Equal(t, addresses, _addresses, "The origin addresses should be equal to the resolved ones.")

	if _ret, err := os.ReadFile("addresses_test.json"); err != nil {
		log.Fatal(err)
	} else {
		if err := json.Unmarshal(_ret, &_addresses); err != nil {
			log.Fatal(err)
		} else {
			t.Log("The resolved result from json file is:")
			t.Log(_addresses)
		}
	}

	assert.Equal(t, addresses, _addresses, "The origin addresses should be equal to the resolved ones.")

}
