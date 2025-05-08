package main

import (
	"errors"
	"testing"
)

func TestLoadConfigPanicWhenFileMissing(t *testing.T){

	config_file_path := "broker/testdata/noconfig.yaml"
	expected := "Couldnt find the config file"
	defer func(){
		if actual := recover(); actual == nil{
			t.Errorf("Expected panic when config is missing")
		}else if actual != expected{
			t.Errorf("Expected panic %q but got %q", expected, actual)
		}else{
			t.Logf("recovered from a expected panic %q",expected)
		}
	}()

	//load config should panic with noconfig.yaml file
	loadConfig(config_file_path)
	t.Logf("recovered from a expected panic %q",expected)
}

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

