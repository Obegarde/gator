package main
import(
	"fmt"
	"github.com/obegarde/gator/internal/config"
	"os"
)

type state struct{
	config *config.Config
}

type command struct{
	name string
	args []string
}

type commands struct{
	handlers map[string]func(*state, command) error
}


func main(){
	configData, err := config.Read()
	if err != nil{
	fmt.Printf("Read Error: %v\n", err)	
	}
	theState := state{}
	theState.config = &configData
	theCommands := newCommands()
	theCommands.register("login", handlerLogin)
	if len(os.Args) < 2{
		os.Exit(1)
	}
	theCommand := new(command)
	theCommand.name = os.Args[1]
	theCommand.args = os.Args[2:]
	err = theCommands.run(&theState,*theCommand)
	if err != nil{
	fmt.Printf("run Error: %v\n", err)
	os.Exit(1)
	}
}

func handlerLogin(s *state, cmd command) error{
	if len(cmd.args) == 0{
		return fmt.Errorf("Error: Expected 1 arg, a username") 
	}
	err := config.SetUser(cmd.args[0],*s.config)
	if err != nil{
	return fmt.Errorf("handlerLogin Error: %w\n", err)
	}
	fmt.Println("Username set successfully")
	return nil
}

func newCommands() *commands{
	return &commands{
		handlers: make(map[string]func(*state, command)error)} 
}

func (c *commands) register(name string, handler func(*state,command)error){
	c.handlers[name] = handler 
}

func (c *commands) run(s *state, cmd command) error{
	handler, exists := c.handlers[cmd.name]
	if !exists{
	return fmt.Errorf("command %s not found", cmd.name)
	}
	return handler(s, cmd)
}


	
	

	


