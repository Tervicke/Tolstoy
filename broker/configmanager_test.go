package broker

import (
	"errors"
	"testing"
)

func TestLoadConfig(t *testing.T) {

	type testCase struct{
		name string 
		configpath string
		expected error
	}
	
	tests := []testCase{
		{
			name : "test no port defined in the config file",
			configpath: "testdata/noport.yaml",
			expected: errors.New("no port defined in the config file"),
		},
		{
			name : "test no host in config",
			configpath: "testdata/nohost.yaml",
			expected: errors.New("no host defined in the config file"),
		},
		{
			name : "test correct config file",
			configpath: "testdata/correctconfig1.yaml",
			expected: nil,
		},
		{
			name : "test config file doesnt exist",
			configpath: "",
			expected: errors.New("could not find the config file"),
		},
		{
			name : "test error parsing the config yaml",
			configpath: "testdata/incorrectyaml.yaml",
			expected: errors.New("yaml: line 2: could not find expected ':'"),
		},
	}
	
	for _,tt := range tests{
		t.Run(tt.name , func(t *testing.T){
			//reset the default data
			brokerSettings = configdata{
				Port: -1,
				Host:"",
			}
			actual := loadConfig(tt.configpath)
			if tt.expected == nil{
				if tt.expected != actual {
					t.Errorf("Expected Nil got %v",actual)
				}
			}else{
				if actual.Error() != tt.expected.Error() {
					t.Errorf("Expected Error %v got %v",tt.expected.Error(),actual.Error())
				}
			}
		})
	}
	
}
