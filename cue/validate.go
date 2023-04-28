package cue

import (
	_ "embed"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

// Have a struct or something to embed that accepts the cue file as an argument.
// The struct will then have a method that will validate it against the input that is passed
// in which is a yaml file.
//
// This yaml file will represent either correct or incorrect configuration for features
// and the method will spit out an error or nil.
//
// Validate should accept an argument of an slice of byte slices that represent files passed in as yaml
// files for them to be validated against the cue file.
//
// This file and package therefore will have majority of the cue logic for checking
// the file and returning any errors if any.
var (
	//go:embed flipt.cue
	cueFile []byte

	//go:embed features.yaml
	yamlFile []byte
)

// Look at the cue vet example in the source code to see exactly how the validating files
// logic works.
func Validate() error {
	fmt.Println(string(yamlFile))
	ctx := cuecontext.New()

	val := ctx.CompileBytes(cueFile)
	if err := val.Err(); err != nil {
		return err
	}

	v := ctx.CompileBytes(yamlFile, cue.Filename("features.yaml"))
	if err := v.Err(); err != nil {
		fmt.Println("error here is: ", err)
		return err
	}

	va := val.Unify(v)

	err := va.Validate()

	fmt.Println("The error here is: ", err)

	return nil
}
