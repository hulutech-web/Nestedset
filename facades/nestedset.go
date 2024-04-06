package facades

import (
	"log"

	"goravel/packages/nestedset"
	"goravel/packages/nestedset/contracts"
)

func Nestedset() contracts.Nestedset {
	instance, err := nestedset.App.Make(nestedset.Binding)
	if err != nil {
		log.Println(err)
		return nil
	}

	return instance.(contracts.Nestedset)
}
