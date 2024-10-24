package config

import(
	"os"
	"encoding/json"

)
const configFileName = "/.gatorconfig.json"

type Config struct{
	DbURL string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config,error){
	newConfig := Config{}
	configPath, err := getConfigFilePath()
	if err != nil{
	return Config{},err 
	}
	data, err := os.ReadFile(configPath)
	if err != nil{
	return Config{}, err
	}
	err = json.Unmarshal(data,&newConfig)
	if err != nil {
	return Config{},err
	}
	return newConfig, nil
}

func SetUser(userName string,cfg Config) error{
	cfg.CurrentUserName = userName
	err := write(cfg)
	if err != nil{
	return err
	}
	return nil
}


func getConfigFilePath() (string,error){	
	homedirstring,err := os.UserHomeDir()
	if err != nil{
	return "",err
	}
	configFilePath := homedirstring + configFileName
	return configFilePath, nil
}

func write(cfg Config) error{
	marshalledConfig,err := json.Marshal(cfg)
	if err != nil{
	return err
	}	
	configPath, err := getConfigFilePath()
	if err != nil{
	return err 
	}
	err = os.WriteFile(configPath,marshalledConfig,0600)
	if err != nil{
	return err
	}
	return nil
}
